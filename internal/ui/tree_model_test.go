package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seb07-cloud/pim-tui/internal/azure"
)

// Helper function to create test groups
func createTestGroups(count int) []azure.Group {
	groups := make([]azure.Group, count)
	for i := 0; i < count; i++ {
		groups[i] = azure.Group{
			ID:          "group-" + string(rune('A'+i)),
			DisplayName: "Test Group " + string(rune('A'+i)),
		}
	}
	return groups
}

// Helper function to create test roles
func createTestRoles(count int) []azure.Role {
	roles := make([]azure.Role, count)
	for i := 0; i < count; i++ {
		roles[i] = azure.Role{
			ID:          "role-" + string(rune('1'+i)),
			DisplayName: "Test Role " + string(rune('1'+i)),
		}
	}
	return roles
}

func TestNewTreeView(t *testing.T) {
	tv := NewTreeView(80, 24)

	if tv.width != 80 {
		t.Errorf("expected width 80, got %d", tv.width)
	}
	if tv.height != 24 {
		t.Errorf("expected height 24, got %d", tv.height)
	}
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", tv.cursor)
	}
	if len(tv.nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(tv.nodes))
	}
	if tv.expanded == nil {
		t.Error("expected expanded map to be initialized")
	}
}

func TestSetData(t *testing.T) {
	tv := NewTreeView(80, 24)
	// Create roles with Active status (only active roles are shown)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
		{ID: "r2", DisplayName: "Role 2", Status: azure.StatusActive},
		{ID: "r3", DisplayName: "Role 3", Status: azure.StatusInactive}, // Won't show
	}
	tenant := &azure.Tenant{ID: "test-id", DisplayName: "Test Tenant"}

	tv.SetData(nil, roles, tenant)

	// Only active roles shown: 1 root + 2 active roles = 3 nodes
	if len(tv.nodes) != 3 {
		t.Errorf("expected 3 nodes (root + 2 active roles), got %d", len(tv.nodes))
	}

	// Check root node
	if tv.nodes[0].Label != "You" {
		t.Errorf("expected root label 'You', got '%s'", tv.nodes[0].Label)
	}
	if tv.nodes[0].Level != 0 {
		t.Errorf("expected root level 0, got %d", tv.nodes[0].Level)
	}
	if !tv.nodes[0].HasChildren {
		t.Error("expected root to have children (active roles exist)")
	}

	// Check first role (level 1)
	if tv.nodes[1].Level != 1 {
		t.Errorf("expected role level 1, got %d", tv.nodes[1].Level)
	}
	// Roles have tenant as child
	if !tv.nodes[1].HasChildren {
		t.Error("expected role to have tenant as child")
	}
}

func TestNavigateDown(t *testing.T) {
	tv := NewTreeView(80, 24)
	// Only active roles are shown
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
		{ID: "r2", DisplayName: "Role 2", Status: azure.StatusActive},
		{ID: "r3", DisplayName: "Role 3", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// Structure: 1 root + 3 active roles = 4 nodes (indices 0-3)
	tests := []struct {
		name           string
		initialCursor  int
		expectedCursor int
	}{
		{"from root", 0, 1},
		{"from first role", 1, 2},
		{"from second role", 2, 3},
		{"at last node stays", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv.cursor = tt.initialCursor
			tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyDown})
			if tv.cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, tv.cursor)
			}
		})
	}
}

func TestNavigateUp(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
		{ID: "r2", DisplayName: "Role 2", Status: azure.StatusActive},
		{ID: "r3", DisplayName: "Role 3", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	tests := []struct {
		name           string
		initialCursor  int
		expectedCursor int
	}{
		{"from last role", 3, 2},
		{"from second role", 2, 1},
		{"from first role", 1, 0},
		{"at root stays", 0, 0}, // Can't go above root
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tv.cursor = tt.initialCursor
			tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyUp})
			if tv.cursor != tt.expectedCursor {
				t.Errorf("expected cursor %d, got %d", tt.expectedCursor, tv.cursor)
			}
		})
	}
}

func TestActiveRolesOnlyShown(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Active Role", Status: azure.StatusActive},
		{ID: "r2", DisplayName: "Inactive Role", Status: azure.StatusInactive},
	}
	tenant := &azure.Tenant{ID: "test", DisplayName: "Test Tenant"}
	tv.SetData(nil, roles, tenant)

	// Only 1 active role shown: 1 root + 1 active role = 2 nodes
	if len(tv.nodes) != 2 {
		t.Fatalf("expected 2 nodes (root + 1 active role), got %d", len(tv.nodes))
	}

	// Root should have children (active role exists)
	if !tv.nodes[0].HasChildren {
		t.Error("expected root to have children")
	}

	// Role should have tenant as child
	if !tv.nodes[1].HasChildren {
		t.Error("expected role to have tenant as child")
	}

	// Verify only active role is shown
	if tv.nodes[1].Label != "Active Role" {
		t.Errorf("expected 'Active Role', got '%s'", tv.nodes[1].Label)
	}
}

func TestCursorBounds(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
		{ID: "r2", DisplayName: "Role 2", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// 1 root + 2 active roles = 3 nodes
	if len(tv.nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(tv.nodes))
	}

	// Test cursor stays within bounds when navigating down at end
	tv.cursor = 2 // Last node
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if tv.cursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", tv.cursor)
	}

	// Test cursor stays within bounds when navigating up at start
	tv.cursor = 0 // First node
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if tv.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", tv.cursor)
	}
}

func TestParentNavigation(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// Move to a role (level 1)
	tv.cursor = 1

	// Press left to move to parent (root, level 0)
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyLeft})

	// Cursor should be at the root (level 0)
	if tv.nodes[tv.cursor].Level != 0 {
		t.Errorf("expected to be at level 0 (root), got level %d", tv.nodes[tv.cursor].Level)
	}
	if tv.cursor != 0 {
		t.Errorf("expected cursor at index 0, got %d", tv.cursor)
	}
}

func TestVimStyleNavigation(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
		{ID: "r2", DisplayName: "Role 2", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// Test 'j' for down
	tv.cursor = 0
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tv.cursor != 1 {
		t.Errorf("expected cursor 1 after 'j', got %d", tv.cursor)
	}

	// Test 'k' for up
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0 after 'k', got %d", tv.cursor)
	}

	// Test 'G' for end
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if tv.cursor != len(tv.nodes)-1 {
		t.Errorf("expected cursor at end (%d) after 'G', got %d", len(tv.nodes)-1, tv.cursor)
	}

	// Test 'g' for home
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if tv.cursor != 0 {
		t.Errorf("expected cursor 0 after 'g', got %d", tv.cursor)
	}
}

func TestRightKeyNavigation(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// Roles have tenant as child but it's not expandable in UI
	tv.cursor = 1 // First role

	oldCursor := tv.cursor
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyRight})

	// Right key behavior depends on node having children
	// The role has HasChildren=true but nothing to expand to
	if tv.cursor < oldCursor {
		t.Errorf("cursor moved backwards unexpectedly: %d -> %d", oldCursor, tv.cursor)
	}
}

func TestLeftKeyMovesToParent(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// Move to a role (level 1)
	tv.cursor = 1

	// Press left to move to parent (root)
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyLeft})

	// Should be at root (level 0)
	if tv.nodes[tv.cursor].Level != 0 {
		t.Errorf("expected to be at root (level 0), got level %d", tv.nodes[tv.cursor].Level)
	}
}

func TestSelectedNode(t *testing.T) {
	tv := NewTreeView(80, 24)
	roles := []azure.Role{
		{ID: "r1", DisplayName: "Role 1", Status: azure.StatusActive},
	}
	tv.SetData(nil, roles, nil)

	// Test getting selected node
	node := tv.SelectedNode()
	if node == nil {
		t.Fatal("expected non-nil selected node")
	}
	if node.Label != "You" {
		t.Errorf("expected selected node label 'You', got '%s'", node.Label)
	}

	// Move cursor and check again
	tv.cursor = 1
	node = tv.SelectedNode()
	if node.Level != 1 {
		t.Errorf("expected level 1, got %d", node.Level)
	}
}

func TestEmptyTree(t *testing.T) {
	tv := NewTreeView(80, 24)
	tv.SetData([]azure.Group{}, []azure.Role{}, nil)

	// Should still have root node
	if len(tv.nodes) != 1 {
		t.Errorf("expected 1 node (root) for empty tree, got %d", len(tv.nodes))
	}

	// Root should have no children
	if tv.nodes[0].HasChildren {
		t.Error("expected root to have no children when groups are empty")
	}

	// Navigation should not crash
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyDown})
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyUp})
	tv, _ = tv.Update(tea.KeyMsg{Type: tea.KeyEnter})
}

func TestRoleDetail_EnterOnRoleShowsDetail(t *testing.T) {
	// Setup: tree with active role (only active roles shown)
	m := NewTreeView(80, 24)
	roles := []azure.Role{{
		ID:          "r1",
		DisplayName: "Test Role",
		Description: "A test role",
		Status:      azure.StatusActive, // Must be active to appear
	}}
	tenant := &azure.Tenant{ID: "test", DisplayName: "Test Tenant"}
	m.SetData(nil, roles, tenant)

	// Structure: root (0), role (1)
	// Move to role node
	m.cursor = 1 // should be the role

	// Verify role node properties (level 1, has tenant child)
	node := m.nodes[m.cursor]
	if node.Level != 1 {
		t.Fatalf("Expected role node at level 1, got level=%d", node.Level)
	}

	// Verify it's actually a role
	if _, ok := node.Value.(azure.Role); !ok {
		t.Fatalf("Expected node to be a role, got %T", node.Value)
	}

	// Press Enter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !m.showingDetail {
		t.Error("Expected showingDetail to be true after Enter on role")
	}
	if m.detailRole == nil {
		t.Error("Expected detailRole to be set")
	}
	if m.detailRole.DisplayName != "Test Role" {
		t.Errorf("Expected detail role 'Test Role', got '%s'", m.detailRole.DisplayName)
	}
}

func TestRoleDetail_EnterDismissesDetail(t *testing.T) {
	m := NewTreeView(80, 24)
	m.showingDetail = true
	m.detailRole = &azure.Role{DisplayName: "Test"}

	// Press Enter to dismiss
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.showingDetail {
		t.Error("Expected showingDetail to be false after Enter")
	}
	if m.detailRole != nil {
		t.Error("Expected detailRole to be nil")
	}
}

func TestRoleDetail_EscDismissesDetail(t *testing.T) {
	m := NewTreeView(80, 24)
	m.showingDetail = true
	m.detailRole = &azure.Role{DisplayName: "Test"}

	// Press Esc to dismiss
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if m.showingDetail {
		t.Error("Expected showingDetail to be false after Esc")
	}
}

func TestRoleDetail_ViewRendersDetail(t *testing.T) {
	m := NewTreeView(80, 24)
	m.showingDetail = true
	m.detailRole = &azure.Role{
		DisplayName: "Global Admin",
		Description: "Full admin access",
		Status:      azure.StatusActive,
	}

	output := m.View()

	if !strings.Contains(output, "Global Admin") {
		t.Error("Expected detail view to contain role name")
	}
	if !strings.Contains(output, "Full admin access") {
		t.Error("Expected detail view to contain description")
	}
}
