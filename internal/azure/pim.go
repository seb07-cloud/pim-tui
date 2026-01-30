package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
)

// pimRoleAssignment represents a single role assignment from PIM API
type pimRoleAssignment struct {
	ID         string `json:"id"`
	ResourceID string `json:"resourceId"`
	RoleDefinition struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Resource    struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"resource"`
	} `json:"roleDefinition"`
	Subject struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	} `json:"subject"`
	AssignmentState string `json:"assignmentState"` // "Eligible" or "Active"
	EndDateTime     string `json:"endDateTime"`
}

// PIM Governance API response types for Entra Roles
type pimRoleResponse struct {
	Value    []pimRoleAssignment `json:"value"`
	NextLink string              `json:"@odata.nextLink"`
}

func (c *Client) GetEligibleRoles(ctx context.Context) ([]Role, error) {
	// Get current user ID first
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// Use PIM Governance API for Entra Roles
	// Filter format: (subject/id eq 'xxx') and (assignmentState eq 'Eligible')
	filter := fmt.Sprintf("(subject/id eq '%s') and (assignmentState eq 'Eligible')", userID)
	expand := "linkedEligibleRoleAssignment,subject,scopedResource,roleDefinition($expand=resource)"
	reqURL := fmt.Sprintf("%s/aadroles/roleAssignments?$expand=%s&$filter=%s", pimBaseURL, url.QueryEscape(expand), url.QueryEscape(filter))

	// Collect all results across pages
	var allAssignments []pimRoleAssignment
	for reqURL != "" {
		data, err := c.pimRequest(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, err
		}

		var result pimRoleResponse
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		allAssignments = append(allAssignments, result.Value...)
		reqURL = result.NextLink // Follow pagination until no more pages
	}

	roles := make([]Role, 0, len(allAssignments))
	for _, r := range allAssignments {
		roles = append(roles, Role{
			ID:               r.ID,
			DisplayName:      r.RoleDefinition.DisplayName,
			RoleDefinitionID: r.RoleDefinition.ID,
			DirectoryScopeID: "/", // Entra roles are tenant-scoped
			Status:           StatusInactive,
			MaxDuration:      8 * time.Hour,
		})
	}

	return roles, nil
}

func (c *Client) GetActiveRoles(ctx context.Context) (map[string]*time.Time, error) {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get active role assignments
	filter := fmt.Sprintf("(subject/id eq '%s') and (assignmentState eq 'Active')", userID)
	expand := "linkedEligibleRoleAssignment,subject,scopedResource,roleDefinition($expand=resource)"
	reqURL := fmt.Sprintf("%s/aadroles/roleAssignments?$expand=%s&$filter=%s", pimBaseURL, url.QueryEscape(expand), url.QueryEscape(filter))

	// Collect all results across pages
	var allAssignments []pimRoleAssignment
	for reqURL != "" {
		data, err := c.pimRequest(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, err
		}

		var result pimRoleResponse
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		allAssignments = append(allAssignments, result.Value...)
		reqURL = result.NextLink // Follow pagination until no more pages
	}

	active := make(map[string]*time.Time)
	for _, r := range allAssignments {
		// Always add active assignments to the map
		// Parse EndDateTime if present, otherwise use nil (active but unknown expiry)
		if r.EndDateTime != "" {
			t, err := time.Parse(time.RFC3339, r.EndDateTime)
			if err == nil {
				active[r.RoleDefinition.ID] = &t
			} else {
				// Parse failed but assignment is active - mark as active with no expiry
				active[r.RoleDefinition.ID] = nil
			}
		} else {
			// No EndDateTime but assignment is active (e.g., permanent or pending)
			active[r.RoleDefinition.ID] = nil
		}
	}

	return active, nil
}

func (c *Client) GetRoles(ctx context.Context) ([]Role, error) {
	// Fetch eligible and active roles in parallel
	var eligible []Role
	var active map[string]*time.Time
	var eligibleErr, activeErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		eligible, eligibleErr = c.GetEligibleRoles(ctx)
	}()
	go func() {
		defer wg.Done()
		active, activeErr = c.GetActiveRoles(ctx)
	}()
	wg.Wait()

	if eligibleErr != nil {
		return nil, eligibleErr
	}
	if activeErr != nil {
		return nil, activeErr
	}

	for i := range eligible {
		if expiry, ok := active[eligible[i].RoleDefinitionID]; ok {
			// Found in active map - mark as active
			eligible[i].ExpiresAt = expiry
			if expiry != nil {
				eligible[i].Status = StatusFromExpiry(expiry)
			} else {
				// Active but no expiry info - mark as Active
				eligible[i].Status = StatusActive
			}
		}
	}

	return eligible, nil
}

func (c *Client) ActivateRole(ctx context.Context, roleDefinitionID, justification string, duration time.Duration) error {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	// First get the resource ID (tenant ID) for Entra roles
	tenant, err := c.GetTenant(ctx)
	if err != nil {
		return err
	}

	minutes := int(duration.Minutes())

	body := map[string]interface{}{
		"roleDefinitionId":     roleDefinitionID,
		"resourceId":           tenant.ID,
		"subjectId":            userID,
		"assignmentState":      "Active",
		"type":                 "UserAdd",
		"reason":               justification,
		"schedule": map[string]interface{}{
			"type":          "Once",
			"startDateTime": time.Now().UTC().Format(time.RFC3339),
			"duration":      fmt.Sprintf("PT%dM", minutes),
		},
	}

	_, err = c.pimRequest(ctx, "POST", pimBaseURL+"/aadroles/roleAssignmentRequests", body)
	return err
}

func (c *Client) DeactivateRole(ctx context.Context, roleDefinitionID string) error {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	// Get the resource ID (tenant ID) for Entra roles
	tenant, err := c.GetTenant(ctx)
	if err != nil {
		return err
	}

	// For deactivation, we use UserRemove type with a minimal schedule
	// The API requires a schedule even for removal
	body := map[string]interface{}{
		"roleDefinitionId": roleDefinitionID,
		"resourceId":       tenant.ID,
		"subjectId":        userID,
		"assignmentState":  "Active",
		"type":             "UserRemove",
		"reason":           "Deactivated via PIM-TUI",
		"schedule": map[string]interface{}{
			"type":          "Once",
			"startDateTime": nil,
			"endDateTime":   nil,
		},
	}

	_, err = c.pimRequest(ctx, "POST", pimBaseURL+"/aadroles/roleAssignmentRequests", body)
	return err
}
