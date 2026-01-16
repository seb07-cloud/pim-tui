package azure

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// newUUID generates a random UUID v4
func newUUID() string {
	uuid := make([]byte, 16)
	rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// roleEligibilityResponse represents the ARM API response for eligible role assignments
type roleEligibilityResponse struct {
	Value []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties struct {
			RoleDefinitionID          string `json:"roleDefinitionId"`
			PrincipalID               string `json:"principalId"`
			Scope                     string `json:"scope"`
			Status                    string `json:"status"`
			RoleEligibilityScheduleID string `json:"roleEligibilityScheduleId"`
			StartDateTime             string `json:"startDateTime"`
			EndDateTime               string `json:"endDateTime"`
			MemberType                string `json:"memberType"`
			ExpandedProperties        *struct {
				Principal *struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					Type        string `json:"type"`
					Email       string `json:"email"`
				} `json:"principal"`
				RoleDefinition *struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					Type        string `json:"type"`
				} `json:"roleDefinition"`
				Scope *struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					Type        string `json:"type"`
				} `json:"scope"`
			} `json:"expandedProperties"`
		} `json:"properties"`
	} `json:"value"`
}

// subscriptionResponse represents the ARM API response for subscription details
type subscriptionResponse struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscriptionId"`
	DisplayName    string `json:"displayName"`
	TenantID       string `json:"tenantId"`     // Home tenant of the subscription
	HomeTenantID   string `json:"homeTenantId"` // For Lighthouse, this is the customer tenant
	State          string `json:"state"`
	// ManagedByTenants shows which tenants have delegated access (Lighthouse)
	ManagedByTenants []struct {
		TenantID string `json:"tenantId"`
	} `json:"managedByTenants"`
}

// roleAssignmentResponse represents the ARM API response for active role assignments
type roleAssignmentResponse struct {
	Value []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties struct {
			RoleDefinitionID string `json:"roleDefinitionId"`
			PrincipalID      string `json:"principalId"`
			Scope            string `json:"scope"`
			Status           string `json:"status"`
			StartDateTime    string `json:"startDateTime"`
			EndDateTime      string `json:"endDateTime"`
			AssignmentType   string `json:"assignmentType"`
			MemberType       string `json:"memberType"`
		} `json:"properties"`
	} `json:"value"`
}

// getSubscriptionDetails fetches subscription details including tenant ID
func (c *Client) getSubscriptionDetails(ctx context.Context, subscriptionID string) (*subscriptionResponse, error) {
	reqURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s?api-version=2022-12-01", subscriptionID)
	data, err := c.armRequest(ctx, "GET", reqURL)
	if err != nil {
		return nil, err
	}

	var result subscriptionResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getTenantDisplayName tries to get a tenant's display name from cache or returns the ID
func getTenantDisplayName(tenantID string, tenantCache map[string]string) string {
	if name, ok := tenantCache[tenantID]; ok && name != "" {
		return name
	}
	// Shorten tenant ID for display if no name available
	if len(tenantID) > 8 {
		return tenantID[:8] + "..."
	}
	return tenantID
}

// getTenantNameByID fetches a tenant's display name using the Graph API
// This works for any tenant, including Lighthouse customer tenants
func (c *Client) getTenantNameByID(ctx context.Context, tenantID string) (string, error) {
	// Use Graph API to lookup tenant info by ID
	reqURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/tenantRelationships/findTenantInformationByTenantId(tenantId='%s')", tenantID)

	data, err := c.graphRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	var result struct {
		TenantID          string `json:"tenantId"`
		DisplayName       string `json:"displayName"`
		DefaultDomainName string `json:"defaultDomainName"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	if result.DisplayName != "" {
		return result.DisplayName, nil
	}
	if result.DefaultDomainName != "" {
		return result.DefaultDomainName, nil
	}
	return "", nil
}

func (c *Client) armRequest(ctx context.Context, method, reqURL string) ([]byte, error) {
	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ARM token: %w", err)
	}

	// Retry with exponential backoff for rate limiting
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Retry on 429 Too Many Requests
		if resp.StatusCode == 429 && attempt < maxRetries {
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("ARM API error %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}
	return nil, fmt.Errorf("ARM API request failed after %d retries", maxRetries)
}

// GetLighthouseSubscriptions fetches subscriptions where the current user has eligible PIM roles
// Uses the ARM API with $filter=asTarget() to get only the current user's eligible assignments
func (c *Client) GetLighthouseSubscriptions(ctx context.Context, groups []Group) ([]LighthouseSubscription, error) {
	// Query all eligible role assignments for the current user using asTarget() filter
	// This returns ONLY the current user's eligible assignments across all subscriptions
	// Build URL with proper query parameter encoding
	baseURL := "https://management.azure.com/providers/Microsoft.Authorization/roleEligibilityScheduleInstances"
	params := url.Values{}
	params.Set("api-version", "2020-10-01")
	params.Set("$filter", "asTarget()")
	eligibleURL := baseURL + "?" + params.Encode()

	eligibleData, err := c.armRequest(ctx, "GET", eligibleURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get eligible role assignments: %w", err)
	}

	var eligibleResult roleEligibilityResponse
	if err := json.Unmarshal(eligibleData, &eligibleResult); err != nil {
		return nil, fmt.Errorf("failed to parse eligible role assignments: %w", err)
	}

	// Group eligible roles by subscription
	subMap := make(map[string]*LighthouseSubscription)

	for _, e := range eligibleResult.Value {
		// Extract subscription ID from scope
		// Scope format: /subscriptions/{subId} or /subscriptions/{subId}/resourceGroups/...
		scope := e.Properties.Scope
		parts := strings.Split(scope, "/")
		if len(parts) < 3 || parts[1] != "subscriptions" {
			continue // Skip non-subscription scopes
		}
		subID := parts[2]

		// Get or create subscription entry
		sub, exists := subMap[subID]
		if !exists {
			// Get subscription display name from expanded properties or fetch it
			displayName := subID
			if e.Properties.ExpandedProperties != nil && e.Properties.ExpandedProperties.Scope != nil {
				displayName = e.Properties.ExpandedProperties.Scope.DisplayName
			}

			sub = &LighthouseSubscription{
				ID:            subID,
				DisplayName:   displayName,
				Status:        StatusInactive,
				EligibleRoles: make([]EligibleAzureRole, 0),
			}
			subMap[subID] = sub
		}

		// Get role name from expanded properties
		roleName := ""
		if e.Properties.ExpandedProperties != nil && e.Properties.ExpandedProperties.RoleDefinition != nil {
			roleName = e.Properties.ExpandedProperties.RoleDefinition.DisplayName
		}
		if roleName == "" {
			// Extract role name from roleDefinitionId (last segment)
			defParts := strings.Split(e.Properties.RoleDefinitionID, "/")
			if len(defParts) > 0 {
				roleName = defParts[len(defParts)-1]
			}
		}

		role := EligibleAzureRole{
			RoleDefinitionID:   e.Properties.RoleDefinitionID,
			RoleDefinitionName: roleName,
			RoleEligibilityID:  e.Properties.RoleEligibilityScheduleID,
			Scope:              e.Properties.Scope,
			Status:             StatusInactive,
			ExpiresAt:          nil,
		}

		sub.EligibleRoles = append(sub.EligibleRoles, role)
	}

	// Phase 1: Fetch subscription details to get tenant IDs (in parallel)
	subTenantMap := make(map[string]string) // subID -> tenantID
	var wg sync.WaitGroup
	var mu sync.Mutex
	for subID := range subMap {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if details, err := c.getSubscriptionDetails(ctx, id); err == nil {
				tenantID := details.HomeTenantID
				if tenantID == "" {
					tenantID = details.TenantID
				}
				if tenantID != "" {
					mu.Lock()
					subTenantMap[id] = tenantID
					mu.Unlock()
				}
			}
		}(subID)
	}
	wg.Wait()

	// Phase 2: Collect unique tenant IDs
	uniqueTenants := make(map[string]bool)
	for _, tenantID := range subTenantMap {
		uniqueTenants[tenantID] = true
	}

	// Phase 3: Fetch tenant names for unique tenant IDs only (in parallel)
	// This is the optimization: N subscriptions in M tenants = M calls instead of N calls
	tenantCache := make(map[string]string) // tenantID -> tenantName
	for tenantID := range uniqueTenants {
		wg.Add(1)
		go func(tid string) {
			defer wg.Done()
			if name, err := c.getTenantNameByID(ctx, tid); err == nil && name != "" {
				mu.Lock()
				tenantCache[tid] = name
				mu.Unlock()
			}
		}(tenantID)
	}
	wg.Wait()

	// Debug: Log tenant cache efficiency
	log.Printf("[lighthouse] Fetched names for %d unique tenants (from %d subscriptions)", len(uniqueTenants), len(subMap))

	// Phase 4: Apply cached tenant info to subscriptions
	for subID, tenantID := range subTenantMap {
		if sub, ok := subMap[subID]; ok {
			sub.TenantID = tenantID
			sub.TenantName = getTenantDisplayName(tenantID, tenantCache)
		}
	}

	// Query active role assignments to update status
	// This is optional - if it fails, we just don't show which roles are active
	activeBaseURL := "https://management.azure.com/providers/Microsoft.Authorization/roleAssignmentScheduleInstances"
	activeParams := url.Values{}
	activeParams.Set("api-version", "2020-10-01")
	activeParams.Set("$filter", "asTarget()")
	activeURL := activeBaseURL + "?" + activeParams.Encode()

	if activeData, activeErr := c.armRequest(ctx, "GET", activeURL); activeErr == nil {
		var activeResult roleAssignmentResponse
		if jsonErr := json.Unmarshal(activeData, &activeResult); jsonErr == nil {
			// Build a map of active assignments: scope+roleDefinitionId -> endDateTime
			activeMap := make(map[string]time.Time)
			for _, a := range activeResult.Value {
				// Only consider "Activated" assignments (not permanent ones)
				if a.Properties.AssignmentType == "Activated" {
					key := a.Properties.Scope + "|" + a.Properties.RoleDefinitionID
					if a.Properties.EndDateTime != "" {
						if endTime, parseErr := time.Parse(time.RFC3339, a.Properties.EndDateTime); parseErr == nil {
							activeMap[key] = endTime
						}
					}
				}
			}

			// Update status of eligible roles that are active
			for _, sub := range subMap {
				for i := range sub.EligibleRoles {
					role := &sub.EligibleRoles[i]
					key := role.Scope + "|" + role.RoleDefinitionID
					if endTime, exists := activeMap[key]; exists {
						role.ExpiresAt = &endTime
						role.Status = StatusFromExpiry(&endTime)
					}
				}
			}
		}
	}
	// Note: errors from active assignments query are silently ignored
	// because this is an enhancement - eligible roles are still returned

	// Convert map to slice and sort by tenant name, then subscription name
	subscriptions := make([]LighthouseSubscription, 0, len(subMap))
	for _, sub := range subMap {
		subscriptions = append(subscriptions, *sub)
	}

	// Sort by tenant name, then by subscription display name
	sort.Slice(subscriptions, func(i, j int) bool {
		if subscriptions[i].TenantName != subscriptions[j].TenantName {
			return subscriptions[i].TenantName < subscriptions[j].TenantName
		}
		return subscriptions[i].DisplayName < subscriptions[j].DisplayName
	})

	return subscriptions, nil
}

// ActivateAzureRole activates an eligible Azure RBAC role
// scope should be the full scope path (e.g., /subscriptions/{id} or /subscriptions/{id}/resourceGroups/{name})
func (c *Client) ActivateAzureRole(ctx context.Context, scope, roleDefinitionID, roleEligibilityID, justification string, duration time.Duration) error {
	requestID := newUUID()
	activationURL := fmt.Sprintf("https://management.azure.com%s/providers/Microsoft.Authorization/roleAssignmentScheduleRequests/%s?api-version=2020-10-01", scope, requestID)

	// Get current user ID for principalId
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	body := map[string]interface{}{
		"properties": map[string]interface{}{
			"principalId":                    userID,
			"roleDefinitionId":               roleDefinitionID,
			"requestType":                    "SelfActivate",
			"linkedRoleEligibilityScheduleId": roleEligibilityID,
			"justification":                  justification,
			"scheduleInfo": map[string]interface{}{
				"startDateTime": time.Now().UTC().Format(time.RFC3339),
				"expiration": map[string]interface{}{
					"type":     "AfterDuration",
					"duration": fmt.Sprintf("PT%dH", int(duration.Hours())),
				},
			},
		},
	}

	_, err = c.armRequestWithBody(ctx, "PUT", activationURL, body)
	return err
}

// DeactivateAzureRole deactivates an active Azure RBAC role
// scope should be the full scope path (e.g., /subscriptions/{id} or /subscriptions/{id}/resourceGroups/{name})
func (c *Client) DeactivateAzureRole(ctx context.Context, scope, roleDefinitionID string) error {
	requestID := newUUID()
	deactivationURL := fmt.Sprintf("https://management.azure.com%s/providers/Microsoft.Authorization/roleAssignmentScheduleRequests/%s?api-version=2020-10-01", scope, requestID)

	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	body := map[string]interface{}{
		"properties": map[string]interface{}{
			"principalId":      userID,
			"roleDefinitionId": roleDefinitionID,
			"requestType":      "SelfDeactivate",
		},
	}

	_, err = c.armRequestWithBody(ctx, "PUT", deactivationURL, body)
	return err
}

// armRequestWithBody makes an ARM API request with a JSON body
func (c *Client) armRequestWithBody(ctx context.Context, method, reqURL string, body interface{}) ([]byte, error) {
	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ARM token: %w", err)
	}

	// Marshal body once outside the retry loop
	var jsonBody []byte
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Retry with exponential backoff for rate limiting
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}

		var reqBody io.Reader
		if jsonBody != nil {
			reqBody = strings.NewReader(string(jsonBody))
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Retry on 429 Too Many Requests
		if resp.StatusCode == 429 && attempt < maxRetries {
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("ARM API error %d: %s", resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}
	return nil, fmt.Errorf("ARM API request failed after %d retries", maxRetries)
}
