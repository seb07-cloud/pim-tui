package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
)

// PIM Governance API response types for Groups
type pimGroupResponse struct {
	Value []struct {
		ID         string `json:"id"`
		ResourceID string `json:"resourceId"` // Group ID
		RoleDefinition struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"` // "Member" or "Owner"
		} `json:"roleDefinition"`
		Subject struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"subject"`
		AssignmentState string `json:"assignmentState"`
		EndDateTime     string `json:"endDateTime"`
	} `json:"value"`
}

// pimGroupResourceResponse for getting group details
type pimGroupResourceResponse struct {
	Value []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Type        string `json:"type"`
	} `json:"value"`
}

func (c *Client) GetEligibleGroups(ctx context.Context) ([]Group, error) {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// Use PIM Governance API for Groups
	filter := fmt.Sprintf("(subject/id eq '%s') and (assignmentState eq 'Eligible')", userID)
	expand := "linkedEligibleRoleAssignment,subject,scopedResource,roleDefinition($expand=resource)"
	reqURL := fmt.Sprintf("%s/aadGroups/roleAssignments?$expand=%s&$filter=%s", pimBaseURL, url.QueryEscape(expand), url.QueryEscape(filter))

	data, err := c.pimRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result pimGroupResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Get group details for each unique group (in parallel)
	groupIDs := make(map[string]bool)
	for _, g := range result.Value {
		groupIDs[g.ResourceID] = true
	}

	// Fetch group names in parallel
	groupNames := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for groupID := range groupIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			// Error intentionally ignored - we fall back to using the group ID as display name
			if name, err := c.getGroupName(ctx, id); err == nil && name != "" {
				mu.Lock()
				groupNames[id] = name
				mu.Unlock()
			}
		}(groupID)
	}
	wg.Wait()

	groups := make([]Group, 0, len(result.Value))
	for _, g := range result.Value {
		displayName := groupNames[g.ResourceID]
		if displayName == "" {
			displayName = g.ResourceID // Fallback to ID if name not found
		}

		groups = append(groups, Group{
			ID:          g.ResourceID,
			DisplayName: displayName,
			Description: g.RoleDefinition.DisplayName, // "Member" or "Owner"
			Status:      StatusInactive,
			MaxDuration: 8 * time.Hour,
		})
	}

	return groups, nil
}

func (c *Client) getGroupName(ctx context.Context, groupID string) (string, error) {
	reqURL := fmt.Sprintf("%s/aadGroups/resources/%s", pimBaseURL, groupID)

	data, err := c.pimRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	var result struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	return result.DisplayName, nil
}

func (c *Client) GetActiveGroups(ctx context.Context) (map[string]*time.Time, error) {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf("(subject/id eq '%s') and (assignmentState eq 'Active')", userID)
	expand := "linkedEligibleRoleAssignment,subject,scopedResource,roleDefinition($expand=resource)"
	reqURL := fmt.Sprintf("%s/aadGroups/roleAssignments?$expand=%s&$filter=%s", pimBaseURL, url.QueryEscape(expand), url.QueryEscape(filter))

	data, err := c.pimRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result pimGroupResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	active := make(map[string]*time.Time)
	for _, g := range result.Value {
		if g.EndDateTime != "" {
			t, err := time.Parse(time.RFC3339, g.EndDateTime)
			if err == nil {
				active[g.ResourceID] = &t
			}
		}
	}

	return active, nil
}

func (c *Client) GetGroups(ctx context.Context) ([]Group, error) {
	// Fetch eligible and active groups in parallel
	var eligible []Group
	var active map[string]*time.Time
	var eligibleErr, activeErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		eligible, eligibleErr = c.GetEligibleGroups(ctx)
	}()
	go func() {
		defer wg.Done()
		active, activeErr = c.GetActiveGroups(ctx)
	}()
	wg.Wait()

	if eligibleErr != nil {
		return nil, eligibleErr
	}
	if activeErr != nil {
		return nil, activeErr
	}

	for i := range eligible {
		if expiry, ok := active[eligible[i].ID]; ok {
			eligible[i].ExpiresAt = expiry
			eligible[i].Status = StatusFromExpiry(expiry)
		}
	}

	return eligible, nil
}

func (c *Client) ActivateGroup(ctx context.Context, groupID, justification string, duration time.Duration) error {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	minutes := int(duration.Minutes())

	// First we need to get the roleDefinitionId for "Member" role in this group
	// For now we assume "member" is the standard role
	body := map[string]interface{}{
		"resourceId":       groupID,
		"roleDefinitionId": "member", // This might need to be fetched dynamically
		"subjectId":        userID,
		"assignmentState":  "Active",
		"type":             "UserAdd",
		"reason":           justification,
		"schedule": map[string]interface{}{
			"type":          "Once",
			"startDateTime": time.Now().UTC().Format(time.RFC3339),
			"duration":      fmt.Sprintf("PT%dM", minutes),
		},
	}

	_, err = c.pimRequest(ctx, "POST", pimBaseURL+"/aadGroups/roleAssignmentRequests", body)
	return err
}

func (c *Client) DeactivateGroup(ctx context.Context, groupID string) error {
	userID, err := c.GetCurrentUser(ctx)
	if err != nil {
		return err
	}

	// For deactivation, we use UserRemove type with a minimal schedule
	body := map[string]interface{}{
		"resourceId":       groupID,
		"roleDefinitionId": "member",
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

	_, err = c.pimRequest(ctx, "POST", pimBaseURL+"/aadGroups/roleAssignmentRequests", body)
	return err
}
