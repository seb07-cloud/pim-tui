package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type lighthouseResponse struct {
	Value []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties struct {
			RegistrationDefinition struct {
				Properties struct {
					ManagedByTenantID   string `json:"managedByTenantId"`
					ManagedByTenantName string `json:"managedByTenantName"`
					Description         string `json:"description"`
				} `json:"properties"`
			} `json:"registrationDefinition"`
		} `json:"properties"`
	} `json:"value"`
}

type subscriptionResponse struct {
	Value []struct {
		SubscriptionID string `json:"subscriptionId"`
		DisplayName    string `json:"displayName"`
		TenantID       string `json:"tenantId"`
	} `json:"value"`
}

func (c *Client) armRequest(ctx context.Context, method, url string) ([]byte, error) {
	// Use 'az rest' when useAzRest is true (c.cred is nil in that case)
	if c.useAzRest {
		return c.azRestRequest(method, url, nil)
	}

	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ARM token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ARM API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) GetLighthouseSubscriptions(ctx context.Context, groups []Group) ([]LighthouseSubscription, error) {
	// First get all subscriptions
	subsData, err := c.armRequest(ctx, "GET", "https://management.azure.com/subscriptions?api-version=2022-01-01")
	if err != nil {
		return nil, err
	}

	var subsResult subscriptionResponse
	if err := json.Unmarshal(subsData, &subsResult); err != nil {
		return nil, err
	}

	// Query lighthouse delegations for all subscriptions in parallel
	type subResult struct {
		subs []LighthouseSubscription
	}
	results := make(chan subResult, len(subsResult.Value))

	var wg sync.WaitGroup
	for _, sub := range subsResult.Value {
		wg.Add(1)
		go func(sub struct {
			SubscriptionID string `json:"subscriptionId"`
			DisplayName    string `json:"displayName"`
			TenantID       string `json:"tenantId"`
		}) {
			defer wg.Done()

			lhURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.ManagedServices/registrationAssignments?api-version=2022-10-01&$expandRegistrationDefinition=true", sub.SubscriptionID)

			lhData, err := c.armRequest(ctx, "GET", lhURL)
			if err != nil {
				results <- subResult{} // Skip subscriptions we can't access
				return
			}

			var lhResult lighthouseResponse
			if err := json.Unmarshal(lhData, &lhResult); err != nil {
				results <- subResult{}
				return
			}

			var subs []LighthouseSubscription
			for _, lh := range lhResult.Value {
				subscription := LighthouseSubscription{
					ID:             sub.SubscriptionID,
					DisplayName:    sub.DisplayName,
					CustomerTenant: sub.TenantID,
					Status:         StatusInactive,
				}

				// Try to match with a PIM group based on description or name
				for _, g := range groups {
					if containsGroupReference(lh.Properties.RegistrationDefinition.Properties.Description, g.DisplayName) {
						subscription.LinkedGroupID = g.ID
						subscription.LinkedGroupName = g.DisplayName
						subscription.Status = g.Status
						break
					}
				}

				subs = append(subs, subscription)
			}
			results <- subResult{subs}
		}(sub)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var subscriptions []LighthouseSubscription
	for r := range results {
		subscriptions = append(subscriptions, r.subs...)
	}

	return subscriptions, nil
}

func containsGroupReference(description, groupName string) bool {
	if groupName == "" || description == "" {
		return false
	}
	return strings.Contains(description, groupName) || strings.Contains(groupName, description)
}
