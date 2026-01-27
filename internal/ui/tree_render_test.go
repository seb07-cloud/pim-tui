package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/seb07-cloud/pim-tui/internal/azure"
)

func TestBuildLipglossTree_Empty(t *testing.T) {
	// Empty model with no data returns simple "You" tree
	m := NewTreeView(80, 24)
	m.rebuildVisibleNodes()

	tree := buildLipglossTree(m)
	output := tree.String()

	if !strings.Contains(output, "You") {
		t.Errorf("expected tree to contain 'You', got: %s", output)
	}
}

func TestBuildLipglossTree_WithActiveRoles(t *testing.T) {
	m := NewTreeView(80, 24)
	// Only active roles are shown in tree
	roles := []azure.Role{
		{DisplayName: "Contributor", Status: azure.StatusActive},
		{DisplayName: "Reader", Status: azure.StatusActive},
	}
	tenant := &azure.Tenant{ID: "test-id", DisplayName: "Test Tenant"}
	m.SetData(nil, roles, tenant)

	tree := buildLipglossTree(m)
	output := tree.String()

	// Active roles should appear with key icon
	if !strings.Contains(output, "Contributor") {
		t.Errorf("expected tree to contain 'Contributor', got: %s", output)
	}
	if !strings.Contains(output, "Reader") {
		t.Errorf("expected tree to contain 'Reader', got: %s", output)
	}
	// Roles should have key icon
	if !strings.Contains(output, "🔑") {
		t.Errorf("expected role key icon '🔑', got: %s", output)
	}
	// Tenant should appear
	if !strings.Contains(output, "Test Tenant") {
		t.Errorf("expected tree to contain tenant name, got: %s", output)
	}
}

func TestBuildLipglossTree_OnlyActiveRolesShown(t *testing.T) {
	m := NewTreeView(80, 24)
	// Mix of active and inactive roles
	roles := []azure.Role{
		{DisplayName: "Active Role", Status: azure.StatusActive},
		{DisplayName: "Inactive Role", Status: azure.StatusInactive},
	}
	m.SetData(nil, roles, nil)

	tree := buildLipglossTree(m)
	output := tree.String()

	// Only active role should appear
	if !strings.Contains(output, "Active Role") {
		t.Errorf("expected tree to contain 'Active Role', got: %s", output)
	}
	// Inactive role should not appear
	if strings.Contains(output, "Inactive Role") {
		t.Errorf("expected tree NOT to contain 'Inactive Role', got: %s", output)
	}
}

func TestBuildLipglossTree_TenantDisplay(t *testing.T) {
	m := NewTreeView(80, 24)
	roles := []azure.Role{
		{DisplayName: "Test Role", Status: azure.StatusActive},
	}
	tenant := &azure.Tenant{ID: "abc-123", DisplayName: "Contoso Corp"}
	m.SetData(nil, roles, tenant)

	tree := buildLipglossTree(m)
	output := tree.String()

	// Tenant name and ID should appear
	if !strings.Contains(output, "Contoso Corp") {
		t.Errorf("expected tree to contain tenant name, got: %s", output)
	}
	if !strings.Contains(output, "abc-123") {
		t.Errorf("expected tree to contain tenant ID, got: %s", output)
	}
}

func TestBuildLipglossTree_BoxCharacters(t *testing.T) {
	m := NewTreeView(80, 24)
	// Need active roles for box characters to appear
	roles := []azure.Role{
		{DisplayName: "Role 1", Status: azure.StatusActive},
		{DisplayName: "Role 2", Status: azure.StatusActive},
	}
	m.SetData(nil, roles, nil)

	tree := buildLipglossTree(m)
	output := tree.String()

	// Should contain Unicode box-drawing characters from RoundedEnumerator
	// These are: ├ (U+251C), ─ (U+2500), │ (U+2502), ╰ (U+2570)
	hasBoxChars := strings.Contains(output, "├") ||
		strings.Contains(output, "─") ||
		strings.Contains(output, "│") ||
		strings.Contains(output, "╰")

	if !hasBoxChars {
		t.Errorf("expected tree to contain box-drawing characters, got: %s", output)
	}
}

func TestView_ContainsBorder(t *testing.T) {
	m := NewTreeView(80, 24)
	roles := []azure.Role{
		{DisplayName: "Test Role", Status: azure.StatusActive},
	}
	m.SetData(nil, roles, nil)

	output := m.View()

	// Rounded border characters: ╭ ╮ ╰ ╯
	hasBorder := strings.Contains(output, "╭") ||
		strings.Contains(output, "╮") ||
		strings.Contains(output, "╰") ||
		strings.Contains(output, "╯")

	if !hasBorder {
		t.Errorf("expected View to contain border characters, got: %s", output)
	}
}

func TestView_ContainsHelpText(t *testing.T) {
	m := NewTreeView(80, 24)
	m.rebuildVisibleNodes()

	output := m.View()

	// Check for navigation help text
	if !strings.Contains(output, "j/k") {
		t.Errorf("expected View to contain 'j/k' help text, got: %s", output)
	}
	if !strings.Contains(output, "details") {
		t.Errorf("expected View to contain 'details' help text, got: %s", output)
	}
	if !strings.Contains(output, "Esc") {
		t.Errorf("expected View to contain 'Esc' help text, got: %s", output)
	}
}

func TestView_ContainsTitle(t *testing.T) {
	m := NewTreeView(80, 24)
	m.rebuildVisibleNodes()

	output := m.View()

	if !strings.Contains(output, "Active Roles Flow") {
		t.Errorf("expected View to contain 'Active Roles Flow' title, got: %s", output)
	}
}

func TestView_EmptyStateMessage(t *testing.T) {
	// When no active roles, should show helpful empty state message
	m := NewTreeView(80, 24)
	m.SetData(nil, nil, nil)

	output := m.View()

	// Should still show "You" root node
	if !strings.Contains(output, "You") {
		t.Errorf("expected View to contain 'You' even with no data, got: %s", output)
	}

	// Should show empty state message for no active roles (flowchart diagram message)
	if !strings.Contains(output, "No active roles to display") {
		t.Errorf("expected View to contain empty state message, got: %s", output)
	}
}

func TestBuildLipglossTree_RootCursorHighlight(t *testing.T) {
	// When cursor is at 0 (on "You"), root should be highlighted
	// Note: In test environment, lipgloss may disable ANSI codes when not in TTY
	// This test verifies the tree structure is correct regardless of styling
	m := NewTreeView(80, 24)
	m.SetData(nil, nil, nil)
	m.cursor = 0

	tree := buildLipglossTree(m)
	output := tree.String()

	// "You" should still be present in output
	if !strings.Contains(output, "You") {
		t.Errorf("expected tree to contain 'You', got: %s", output)
	}

	// Verify cursor position is valid
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got: %d", m.cursor)
	}

	// Verify there's exactly one node when empty
	if len(m.nodes) != 1 {
		t.Errorf("expected 1 node for empty tree, got: %d", len(m.nodes))
	}
}

func TestView_AnimationHelpTextChanges(t *testing.T) {
	// When animation is running, help text should change
	m := NewTreeView(80, 24)
	m.SetData(nil, nil, nil)

	// Before animation - should show "a: animate"
	outputBefore := m.View()
	if !strings.Contains(outputBefore, "a: animate") {
		t.Errorf("expected 'a: animate' in help text before animation, got: %s", outputBefore)
	}

	// Start animation
	m.StartAnimation()

	// After animation started - should show "a: stop animation"
	outputAfter := m.View()
	if !strings.Contains(outputAfter, "a: stop animation") {
		t.Errorf("expected 'a: stop animation' in help text during animation, got: %s", outputAfter)
	}
	if strings.Contains(outputAfter, "a: animate |") {
		t.Errorf("should NOT contain 'a: animate |' during animation, got: %s", outputAfter)
	}
}

func TestBuildLipglossTree_AnimationStyles(t *testing.T) {
	// Enable lipgloss color output for testing
	lipgloss.SetColorProfile(termenv.TrueColor)

	// When animating in PhaseUserPulse, root node should have pulse styling
	m := NewTreeView(80, 24)
	m.SetData(nil, nil, nil)
	m.rebuildVisibleNodes()

	// Start animation - should be in PhaseUserPulse
	m.StartAnimation()
	if m.animationPhase != PhaseUserPulse {
		t.Errorf("expected PhaseUserPulse, got %d", m.animationPhase)
	}
	if !m.animating {
		t.Error("expected animating to be true")
	}

	// Get style for root node (index 0, level 0)
	style := m.getNodeAnimationStyle(0, m.nodes[0])

	// The style should be either pulseStyleBright or pulseStyleDim
	// We can't easily compare lipgloss.Style objects, but we can check the output
	output := style.Render("Test")
	t.Logf("Animation style output: %q", output)

	// At minimum, the output should be different from plain "Test"
	// (contains ANSI escape codes for styling)
	if output == "Test" {
		t.Error("expected animated style to modify output, but got plain 'Test'")
	}
}

func TestView_AnimatedTreeHasStyling(t *testing.T) {
	// Enable lipgloss color output for testing
	lipgloss.SetColorProfile(termenv.TrueColor)

	m := NewTreeView(80, 24)
	m.SetData(nil, nil, nil)

	// Get view before animation
	viewBefore := m.View()

	// Start animation
	m.StartAnimation()

	// Get view during animation
	viewAfter := m.View()

	t.Logf("View before (first 200 chars): %q", viewBefore[:min(200, len(viewBefore))])
	t.Logf("View after (first 200 chars): %q", viewAfter[:min(200, len(viewAfter))])

	// Views should be different (animation changes help text)
	if viewBefore == viewAfter {
		t.Error("View should be different during animation")
	}

	// During animation, help text should show "stop animation"
	if !strings.Contains(viewAfter, "stop animation") {
		t.Errorf("expected 'stop animation' in animated view, got: %s", viewAfter)
	}

	// Animation should still show "You" box (flowchart diagram doesn't use brackets anymore)
	if !strings.Contains(viewAfter, "You") {
		t.Errorf("expected 'You' in animated view, got: %s", viewAfter)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
