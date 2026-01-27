package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/seb07-cloud/pim-tui/internal/azure"
)

func TestClampCursor(t *testing.T) {
	tests := []struct {
		name     string
		cursor   int
		delta    int
		length   int
		expected int
	}{
		{
			name:     "can't go below 0",
			cursor:   0,
			delta:    -1,
			length:   5,
			expected: 0,
		},
		{
			name:     "can't exceed length-1",
			cursor:   4,
			delta:    1,
			length:   5,
			expected: 4,
		},
		{
			name:     "normal movement up",
			cursor:   2,
			delta:    -1,
			length:   5,
			expected: 1,
		},
		{
			name:     "normal movement down",
			cursor:   2,
			delta:    1,
			length:   5,
			expected: 3,
		},
		{
			name:     "empty list edge case",
			cursor:   0,
			delta:    0,
			length:   0,
			expected: 0,
		},
		{
			name:     "empty list with positive delta",
			cursor:   0,
			delta:    1,
			length:   0,
			expected: 0,
		},
		{
			name:     "empty list with negative delta",
			cursor:   0,
			delta:    -1,
			length:   0,
			expected: 0,
		},
		{
			name:     "large positive delta clamped",
			cursor:   0,
			delta:    100,
			length:   5,
			expected: 4,
		},
		{
			name:     "large negative delta clamped",
			cursor:   4,
			delta:    -100,
			length:   5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampCursor(tt.cursor, tt.delta, tt.length)
			if got != tt.expected {
				t.Errorf("clampCursor(%d, %d, %d) = %d, want %d",
					tt.cursor, tt.delta, tt.length, got, tt.expected)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		val      int
		expected int
	}{
		{
			name:     "find 4 in middle",
			slice:    []int{1, 2, 4, 8},
			val:      4,
			expected: 2,
		},
		{
			name:     "find 1 at start",
			slice:    []int{1, 2, 4, 8},
			val:      1,
			expected: 0,
		},
		{
			name:     "find 8 at end",
			slice:    []int{1, 2, 4, 8},
			val:      8,
			expected: 3,
		},
		{
			name:     "not found returns 0",
			slice:    []int{1, 2, 4, 8},
			val:      5,
			expected: 0,
		},
		{
			name:     "empty slice returns 0",
			slice:    []int{},
			val:      1,
			expected: 0,
		},
		{
			name:     "find 2 in middle",
			slice:    []int{1, 2, 4, 8},
			val:      2,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexOf(tt.slice, tt.val)
			if got != tt.expected {
				t.Errorf("indexOf(%v, %d) = %d, want %d",
					tt.slice, tt.val, got, tt.expected)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{
			name:     "debug returns LogDebug",
			input:    "debug",
			expected: LogDebug,
		},
		{
			name:     "error returns LogError",
			input:    "error",
			expected: LogError,
		},
		{
			name:     "info returns LogInfo",
			input:    "info",
			expected: LogInfo,
		},
		{
			name:     "empty string returns LogInfo (default)",
			input:    "",
			expected: LogInfo,
		},
		{
			name:     "unknown returns LogInfo (default)",
			input:    "unknown",
			expected: LogInfo,
		},
		{
			name:     "uppercase DEBUG returns LogInfo (case-sensitive)",
			input:    "DEBUG",
			expected: LogInfo,
		},
		{
			name:     "warning returns LogInfo (not recognized)",
			input:    "warning",
			expected: LogInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLogLevel(tt.input)
			if got != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateJustification(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid reason passes",
			input:       "valid reason",
			expected:    "valid reason",
			expectError: false,
		},
		{
			name:        "empty string returns error",
			input:       "",
			expected:    "",
			expectError: true,
			errorSubstr: "required",
		},
		{
			name:        "whitespace only returns error",
			input:       "   ",
			expected:    "",
			expectError: true,
			errorSubstr: "required",
		},
		{
			name:        "tabs only returns error",
			input:       "\t\t",
			expected:    "",
			expectError: true,
			errorSubstr: "required",
		},
		{
			name:        "string with 501 chars returns error",
			input:       strings.Repeat("a", 501),
			expected:    "",
			expectError: true,
			errorSubstr: "exceeds 500",
		},
		{
			name:        "string with exactly 500 chars passes",
			input:       strings.Repeat("a", 500),
			expected:    strings.Repeat("a", 500),
			expectError: false,
		},
		{
			name:        "string with NUL char returns error",
			input:       "test\x00string",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
		{
			name:        "string with DEL char returns error",
			input:       "test\x7fstring",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
		{
			name:        "string with BEL char returns error",
			input:       "test\x07string",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
		{
			name:        "string with tabs is allowed",
			input:       "reason\twith\ttabs",
			expected:    "reason\twith\ttabs",
			expectError: false,
		},
		{
			name:        "string with newlines is allowed",
			input:       "reason\nwith\nnewlines",
			expected:    "reason\nwith\nnewlines",
			expectError: false,
		},
		{
			name:        "string with carriage returns is allowed",
			input:       "reason\rwith\rCR",
			expected:    "reason\rwith\rCR",
			expectError: false,
		},
		{
			name:        "leading and trailing whitespace trimmed",
			input:       "  trimmed reason  ",
			expected:    "trimmed reason",
			expectError: false,
		},
		{
			name:        "string with ESC char returns error",
			input:       "test\x1bstring",
			expected:    "",
			expectError: true,
			errorSubstr: "control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateJustification(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("validateJustification(%q) error = nil, want error containing %q",
						tt.input, tt.errorSubstr)
					return
				}
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("validateJustification(%q) error = %q, want error containing %q",
						tt.input, err.Error(), tt.errorSubstr)
				}
				return
			}

			if err != nil {
				t.Errorf("validateJustification(%q) error = %v, want nil", tt.input, err)
				return
			}

			if got != tt.expected {
				t.Errorf("validateJustification(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{
			name:     "LogError returns ERROR",
			level:    LogError,
			expected: "ERROR",
		},
		{
			name:     "LogInfo returns INFO",
			level:    LogInfo,
			expected: "INFO",
		},
		{
			name:     "LogDebug returns DEBUG",
			level:    LogDebug,
			expected: "DEBUG",
		},
		{
			name:     "unknown level returns ERROR (default)",
			level:    LogLevel(99),
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("LogLevel.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestStateTransition_TreeView tests the 't' key transitions from StateNormal to StateTreeView
func TestStateTransition_TreeView(t *testing.T) {
	tests := []struct {
		name          string
		initialState  State
		key           string
		expectedState State
	}{
		{
			name:          "t key in StateNormal opens tree view",
			initialState:  StateNormal,
			key:           "t",
			expectedState: StateTreeView,
		},
		{
			name:          "t key in StateLoading does nothing",
			initialState:  StateLoading,
			key:           "t",
			expectedState: StateLoading,
		},
		{
			name:          "t key in StateHelp stays in help",
			initialState:  StateHelp,
			key:           "t",
			expectedState: StateHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal model
			m := Model{
				state:          tt.initialState,
				width:          100,
				height:         50,
				selectedRoles:  make(map[int]bool),
				selectedGroups: make(map[int]bool),
				selectedLight:  make(map[int]bool),
				selectedSubRoles: make(map[string]map[int]bool),
			}

			// Send key message
			msg := fakeKeyMsg(tt.key)
			result, _ := m.handleKeyPress(msg)
			newModel := result.(Model)

			if newModel.state != tt.expectedState {
				t.Errorf("state = %v, want %v", newModel.state, tt.expectedState)
			}
		})
	}
}

// TestTreeView_EscReturns tests that Esc in StateTreeView returns to StateNormal
func TestTreeView_EscReturns(t *testing.T) {
	m := Model{
		state:          StateTreeView,
		width:          100,
		height:         50,
		selectedRoles:  make(map[int]bool),
		selectedGroups: make(map[int]bool),
		selectedLight:  make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:       NewTreeView(96, 40),
	}

	msg := fakeKeyMsg("esc")
	result, _ := m.handleKeyPress(msg)
	newModel := result.(Model)

	if newModel.state != StateNormal {
		t.Errorf("state = %v, want StateNormal", newModel.state)
	}
}

// TestTreeView_ReceivesData tests that tree view receives active roles from model
func TestTreeView_ReceivesData(t *testing.T) {
	m := Model{
		state:          StateNormal,
		width:          100,
		height:         50,
		selectedRoles:  make(map[int]bool),
		selectedGroups: make(map[int]bool),
		selectedLight:  make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		tenant: &azure.Tenant{ID: "test-id", DisplayName: "Test Tenant"},
		// Add test data - only active roles will show
		roles: []azure.Role{
			{DisplayName: "Active Role 1", Status: azure.StatusActive},
			{DisplayName: "Active Role 2", Status: azure.StatusActive},
			{DisplayName: "Inactive Role", Status: azure.StatusInactive}, // Won't show
		},
	}

	// Open tree view with 't'
	msg := fakeKeyMsg("t")
	result, _ := m.handleKeyPress(msg)
	newModel := result.(Model)

	// Verify state changed
	if newModel.state != StateTreeView {
		t.Fatalf("state = %v, want StateTreeView", newModel.state)
	}

	// Verify tree view has nodes (root + 2 active roles)
	nodes := newModel.treeView.Nodes()
	if len(nodes) != 3 { // root + 2 active roles
		t.Errorf("tree nodes count = %d, want 3 (root + 2 active roles)", len(nodes))
	}

	// Verify root node exists
	if len(nodes) > 0 && nodes[0].Label != "You" {
		t.Errorf("root node label = %q, want %q", nodes[0].Label, "You")
	}
}

// TestTreeView_WindowResize tests that window resize updates tree view dimensions
func TestTreeView_WindowResize(t *testing.T) {
	m := Model{
		state:          StateTreeView,
		width:          100,
		height:         50,
		selectedRoles:  make(map[int]bool),
		selectedGroups: make(map[int]bool),
		selectedLight:  make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:       NewTreeView(96, 40),
	}

	// Send WindowSizeMsg via Update (not handleKeyPress)
	msg := fakeWindowSizeMsg(120, 60)
	result, _ := m.Update(msg)
	newModel := result.(Model)

	// Model dimensions should update
	if newModel.width != 120 {
		t.Errorf("width = %d, want 120", newModel.width)
	}
	if newModel.height != 60 {
		t.Errorf("height = %d, want 60", newModel.height)
	}
}

// fakeKeyMsg creates a tea.KeyMsg for testing
func fakeKeyMsg(key string) tea.KeyMsg {
	// Handle special keys
	switch key {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	default:
		// Single character keys
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

// fakeWindowSizeMsg creates a tea.WindowSizeMsg for testing
func fakeWindowSizeMsg(width, height int) tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: width, Height: height}
}

// TestTreeView_AnimationStartsOnKeyA tests that 'a' key in StateTreeView starts animation
func TestTreeView_AnimationStartsOnKeyA(t *testing.T) {
	m := Model{
		state:            StateTreeView,
		width:            100,
		height:           50,
		selectedRoles:    make(map[int]bool),
		selectedGroups:   make(map[int]bool),
		selectedLight:    make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:         NewTreeView(96, 40),
	}
	// Ensure tree has nodes
	m.treeView.SetData(nil, nil, nil)

	// Verify not animating initially
	if m.treeView.IsAnimating() {
		t.Fatal("tree view should not be animating initially")
	}

	// Press 'a' key
	msg := fakeKeyMsg("a")
	result, cmd := m.handleKeyPress(msg)
	newModel := result.(Model)

	// Animation should have started
	if !newModel.treeView.IsAnimating() {
		t.Error("tree view should be animating after pressing 'a'")
	}

	// A command should be returned (the animation tick)
	if cmd == nil {
		t.Error("expected a command to be returned for animation tick")
	}

	// Animation phase should be UserPulse
	if newModel.treeView.animationPhase != PhaseUserPulse {
		t.Errorf("animation phase = %d, want PhaseUserPulse (%d)",
			newModel.treeView.animationPhase, PhaseUserPulse)
	}
}

// TestTreeView_AnimationTogglesOnKeyA tests that 'a' key toggles animation
func TestTreeView_AnimationTogglesOnKeyA(t *testing.T) {
	m := Model{
		state:            StateTreeView,
		width:            100,
		height:           50,
		selectedRoles:    make(map[int]bool),
		selectedGroups:   make(map[int]bool),
		selectedLight:    make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:         NewTreeView(96, 40),
	}
	m.treeView.SetData(nil, nil, nil)

	// First 'a' - should start animation
	msg := fakeKeyMsg("a")
	result, _ := m.handleKeyPress(msg)
	m = result.(Model)

	if !m.treeView.IsAnimating() {
		t.Fatal("animation should be running after first 'a'")
	}

	// Second 'a' - should stop animation
	result, _ = m.handleKeyPress(msg)
	m = result.(Model)

	if m.treeView.IsAnimating() {
		t.Error("animation should be stopped after second 'a'")
	}
}

// TestTreeView_AnimationViewChanges verifies that the View output changes during animation
func TestTreeView_AnimationViewChanges(t *testing.T) {
	m := Model{
		state:            StateTreeView,
		width:            100,
		height:           50,
		selectedRoles:    make(map[int]bool),
		selectedGroups:   make(map[int]bool),
		selectedLight:    make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:         NewTreeView(96, 40),
	}
	m.treeView.SetData(nil, nil, nil)

	// Get view before animation
	viewBefore := m.View()

	// Start animation by pressing 'a'
	msg := fakeKeyMsg("a")
	result, _ := m.handleKeyPress(msg)
	m = result.(Model)

	// Verify animation started
	if !m.treeView.IsAnimating() {
		t.Fatal("animation should be running")
	}

	// Get view during animation
	viewAfter := m.View()

	// Help text should change from "a: animate" to "a: stop animation"
	if !strings.Contains(viewBefore, "a: animate") {
		t.Logf("View before: %s", viewBefore)
		t.Error("expected 'a: animate' in view before animation")
	}
	if !strings.Contains(viewAfter, "a: stop animation") {
		t.Logf("View after: %s", viewAfter)
		t.Error("expected 'a: stop animation' in view during animation")
	}
}

// TestTreeView_AnimationTickProcessing verifies animation tick messages are processed correctly
func TestTreeView_AnimationTickProcessing(t *testing.T) {
	m := Model{
		state:            StateTreeView,
		width:            100,
		height:           50,
		selectedRoles:    make(map[int]bool),
		selectedGroups:   make(map[int]bool),
		selectedLight:    make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:         NewTreeView(96, 40),
	}
	m.treeView.SetData(nil, nil, nil)

	// Start animation
	msg := fakeKeyMsg("a")
	result, cmd := m.handleKeyPress(msg)
	m = result.(Model)

	if !m.treeView.IsAnimating() {
		t.Fatal("animation should be running after pressing 'a'")
	}
	if cmd == nil {
		t.Fatal("expected tick command after starting animation")
	}

	// Simulate receiving animation tick message
	// The command would execute and return an animationTickMsg
	initialFrame := m.treeView.animationFrame
	tickMsg := animationTickMsg{time: time.Now(), id: m.treeView.animationID}

	// Process tick through Update
	result, cmd = m.Update(tickMsg)
	m = result.(Model)

	// Animation should still be running
	if !m.treeView.IsAnimating() {
		t.Error("animation should still be running after tick")
	}

	// Frame should have advanced
	if m.treeView.animationFrame <= initialFrame {
		t.Errorf("animation frame should advance: was %d, now %d",
			initialFrame, m.treeView.animationFrame)
	}

	// Another tick command should be returned
	if cmd == nil {
		t.Error("expected next tick command")
	}
}

// TestTreeView_AnimationFullCycle tests that animation loops continuously through all phases
func TestTreeView_AnimationFullCycle(t *testing.T) {
	m := Model{
		state:            StateTreeView,
		width:            100,
		height:           50,
		selectedRoles:    make(map[int]bool),
		selectedGroups:   make(map[int]bool),
		selectedLight:    make(map[int]bool),
		selectedSubRoles: make(map[string]map[int]bool),
		treeView:         NewTreeView(96, 40),
	}
	// Add some data for a more interesting animation
	m.treeView.SetData(
		[]azure.Group{{DisplayName: "TestGroup"}},
		[]azure.Role{{DisplayName: "TestRole", Status: azure.StatusActive}},
		&azure.Tenant{ID: "test-tenant-id", DisplayName: "Test Tenant"},
	)

	// Start animation
	msg := fakeKeyMsg("a")
	result, _ := m.handleKeyPress(msg)
	m = result.(Model)

	if m.treeView.animationPhase != PhaseUserPulse {
		t.Fatalf("expected PhaseUserPulse, got %d", m.treeView.animationPhase)
	}

	// Run through enough frames to complete one full cycle
	// UserPulse(30) + GroupsFlow(18) + RolesReveal(12) + FadeOut(18) = 78 frames
	fullCycleFrames := UserPulseFrames + GroupFlowFrames + RoleRevealFrames + FadeOutFrames + 5
	for i := 0; i < fullCycleFrames; i++ {
		tickMsg := animationTickMsg{time: time.Now(), id: m.treeView.animationID}
		result, _ = m.Update(tickMsg)
		m = result.(Model)
	}

	// Animation should still be running (loops continuously)
	if !m.treeView.IsAnimating() {
		t.Error("animation should still be running (loops continuously)")
	}
	// Should have looped back to UserPulse
	if m.treeView.animationPhase != PhaseUserPulse {
		t.Errorf("expected PhaseUserPulse after loop, got %d", m.treeView.animationPhase)
	}
}
