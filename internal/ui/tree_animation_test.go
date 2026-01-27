package ui

import (
	"testing"
	"time"

	"github.com/seb07-cloud/pim-tui/internal/azure"
)

func TestAnimationStartStop(t *testing.T) {
	m := NewTreeView(80, 24)

	// Initially not animating
	if m.IsAnimating() {
		t.Error("TreeView should not be animating initially")
	}

	// Start animation
	cmd := m.StartAnimation()
	if cmd == nil {
		t.Error("StartAnimation should return a tick command")
	}
	if !m.IsAnimating() {
		t.Error("TreeView should be animating after StartAnimation")
	}
	if m.animationPhase != PhaseUserPulse {
		t.Errorf("Expected PhaseUserPulse, got %d", m.animationPhase)
	}

	// Stop animation
	m.StopAnimation()
	if m.IsAnimating() {
		t.Error("TreeView should not be animating after StopAnimation")
	}
	if m.animationPhase != PhaseIdle {
		t.Errorf("Expected PhaseIdle, got %d", m.animationPhase)
	}
}

func TestAnimationIDPreventsStale(t *testing.T) {
	m := NewTreeView(80, 24)

	// Start animation
	m.StartAnimation()
	firstID := m.animationID

	// Create a tick with the current ID
	tick := animationTickMsg{time: time.Now(), id: firstID}

	// Stop animation (increments ID)
	m.StopAnimation()

	// Try to process the stale tick
	cmd := m.UpdateAnimation(tick)
	if cmd != nil {
		t.Error("Stale tick should return nil command")
	}
}

func TestAnimationPhaseProgression(t *testing.T) {
	m := NewTreeView(80, 24)

	// Add some test data to have groups and roles
	m.nodes = []TreeNode{
		{ID: 0, Level: 0, Label: "You"},
		{ID: 1, Level: 1, Label: "Group1", HasChildren: true},
		{ID: 2, Level: 2, Label: "Role1"},
	}

	// Start animation
	m.StartAnimation()

	// Verify initial phase
	if m.animationPhase != PhaseUserPulse {
		t.Errorf("Expected PhaseUserPulse, got %d", m.animationPhase)
	}

	// Simulate frames to advance past user pulse (30 frames)
	for i := 0; i < UserPulseFrames; i++ {
		tick := animationTickMsg{time: time.Now(), id: m.animationID}
		m.UpdateAnimation(tick)
	}

	if m.animationPhase != PhaseGroupsFlow {
		t.Errorf("Expected PhaseGroupsFlow after %d frames, got %d", UserPulseFrames, m.animationPhase)
	}
}

func TestAnimationLoopsContinuously(t *testing.T) {
	m := NewTreeView(80, 24)

	// Add minimal test data
	m.nodes = []TreeNode{
		{ID: 0, Level: 0, Label: "You"},
		{ID: 1, Level: 1, Label: "Group1", HasChildren: true},
		{ID: 2, Level: 2, Label: "Role1"},
	}

	// Start animation
	m.StartAnimation()
	initialID := m.animationID

	// Run through enough frames to complete one full cycle and loop back
	// UserPulse + GroupsFlow + RolesReveal + FadeOut = ~78 frames minimum
	totalFrames := UserPulseFrames + GroupFlowFrames + RoleRevealFrames + FadeOutFrames + 10
	for i := 0; i < totalFrames; i++ {
		tick := animationTickMsg{time: time.Now(), id: m.animationID}
		m.UpdateAnimation(tick)
	}

	// Animation should still be running (loops continuously)
	if !m.IsAnimating() {
		t.Error("Animation should still be running (loops continuously)")
	}

	// Animation should have looped back to UserPulse
	if m.animationPhase != PhaseUserPulse {
		t.Errorf("Expected PhaseUserPulse after loop, got %d", m.animationPhase)
	}

	// Animation ID should remain the same (not restarted)
	if m.animationID != initialID {
		t.Errorf("Animation ID should remain %d, got %d", initialID, m.animationID)
	}
}

func TestGetGroupIndex(t *testing.T) {
	m := NewTreeView(80, 24)
	// Now we check by Value type, not Level
	m.nodes = []TreeNode{
		{ID: 0, Level: 0, Label: "You", Value: nil},
		{ID: 1, Level: 1, Label: "Group1", Value: azure.Group{ID: "g1"}},
		{ID: 2, Level: 1, Label: "Group2", Value: azure.Group{ID: "g2"}},
		{ID: 3, Level: 1, Label: "Role1", Value: azure.Role{ID: "r1"}},
	}

	// First group (index 1 in nodes) should be group index 0
	if idx := m.getGroupIndex(1); idx != 0 {
		t.Errorf("Expected group index 0 for node 1, got %d", idx)
	}

	// Second group (index 2 in nodes) should be group index 1
	if idx := m.getGroupIndex(2); idx != 1 {
		t.Errorf("Expected group index 1 for node 2, got %d", idx)
	}

	// Role node should return -1
	if idx := m.getGroupIndex(3); idx != -1 {
		t.Errorf("Expected -1 for role node, got %d", idx)
	}
}

func TestGetRoleIndex(t *testing.T) {
	m := NewTreeView(80, 24)
	// Now we check by Value type, not Level
	m.nodes = []TreeNode{
		{ID: 0, Level: 0, Label: "You", Value: nil},
		{ID: 1, Level: 1, Label: "Group1", Value: azure.Group{ID: "g1"}},
		{ID: 2, Level: 1, Label: "Role1", Value: azure.Role{ID: "r1"}},
		{ID: 3, Level: 1, Label: "Role2", Value: azure.Role{ID: "r2"}},
	}

	// First role (index 2 in nodes) should be role index 0
	if idx := m.getRoleIndex(2); idx != 0 {
		t.Errorf("Expected role index 0 for node 2, got %d", idx)
	}

	// Second role (index 3 in nodes) should be role index 1
	if idx := m.getRoleIndex(3); idx != 1 {
		t.Errorf("Expected role index 1 for node 3, got %d", idx)
	}

	// Group node should return -1
	if idx := m.getRoleIndex(1); idx != -1 {
		t.Errorf("Expected -1 for group node, got %d", idx)
	}
}
