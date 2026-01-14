package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// PIM Governance API response types for Entra Roles
type pimRoleResponse struct {
	Value []struct {
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
	} `json:"value"`
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

	data, err := c.pimRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result pimRoleResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	roles := make([]Role, 0, len(result.Value))
	for _, r := range result.Value {
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

	data, err := c.pimRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result pimRoleResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	active := make(map[string]*time.Time)
	for _, r := range result.Value {
		if r.EndDateTime != "" {
			t, err := time.Parse(time.RFC3339, r.EndDateTime)
			if err == nil {
				active[r.RoleDefinition.ID] = &t
			}
		}
	}

	return active, nil
}

func (c *Client) GetRoles(ctx context.Context) ([]Role, error) {
	eligible, err := c.GetEligibleRoles(ctx)
	if err != nil {
		return nil, err
	}

	active, err := c.GetActiveRoles(ctx)
	if err != nil {
		return nil, err
	}

	for i := range eligible {
		if expiry, ok := active[eligible[i].RoleDefinitionID]; ok {
			eligible[i].ExpiresAt = expiry
			eligible[i].Status = StatusFromExpiry(expiry)
		}
	}

	return eligible, nil
}

func (c *Client) ActivateRole(ctx context.Context, roleDefinitionID, directoryScopeID, justification string, duration time.Duration) error {
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

func (c *Client) DeactivateRole(ctx context.Context, roleDefinitionID, directoryScopeID string) error {
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
