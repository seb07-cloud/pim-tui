package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	var subscriptions []LighthouseSubscription

	// For each subscription, check for lighthouse delegations
	for _, sub := range subsResult.Value {
		lhURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.ManagedServices/registrationAssignments?api-version=2022-10-01&$expandRegistrationDefinition=true", sub.SubscriptionID)

		lhData, err := c.armRequest(ctx, "GET", lhURL)
		if err != nil {
			continue // Skip subscriptions we can't access
		}

		var lhResult lighthouseResponse
		if err := json.Unmarshal(lhData, &lhResult); err != nil {
			continue
		}

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

			subscriptions = append(subscriptions, subscription)
		}
	}

	return subscriptions, nil
}

func containsGroupReference(description, groupName string) bool {
	if groupName == "" || description == "" {
		return false
	}
	return strings.Contains(description, groupName) || strings.Contains(groupName, description)
}
