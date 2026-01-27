package ui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seb07-cloud/pim-tui/internal/azure"
	"github.com/seb07-cloud/pim-tui/internal/config"
)

// testModel creates a Model with default config and specified initial state
func testModel(initialState State) Model {
	cfg := config.Default()
	m := NewModel(cfg, "test")
	m.state = initialState
	// Set height to reasonable value for scroll calculations
	m.height = 40
	m.width = 120
	return m
}

// TestUpdateStateTransitions verifies key state transitions via messages
func TestUpdateStateTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialState State
		msg          tea.Msg
		wantState    State
	}{
		{
			name:         "error message sets error state from normal",
			initialState: StateNormal,
			msg:          errMsg{err: fmt.Errorf("test error"), source: "auth"},
			wantState:    StateError,
		},
		{
			name:         "error message sets error state from loading on auth",
			initialState: StateLoading,
			msg:          errMsg{err: fmt.Errorf("auth failed"), source: "auth"},
			wantState:    StateError,
		},
		{
			name:         "error message sets error state from loading on tenant",
			initialState: StateLoading,
			msg:          errMsg{err: fmt.Errorf("tenant failed"), source: "tenant"},
			wantState:    StateError,
		},
		{
			name:         "window resize preserves normal state",
			initialState: StateNormal,
			msg:          tea.WindowSizeMsg{Width: 100, Height: 50},
			wantState:    StateNormal,
		},
		{
			name:         "window resize preserves loading state",
			initialState: StateLoading,
			msg:          tea.WindowSizeMsg{Width: 100, Height: 50},
			wantState:    StateLoading,
		},
		{
			name:         "help key transitions normal to help",
			initialState: StateNormal,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			wantState:    StateHelp,
		},
		{
			name:         "esc key transitions help to normal",
			initialState: StateHelp,
			msg:          tea.KeyMsg{Type: tea.KeyEsc},
			wantState:    StateNormal,
		},
		{
			name:         "question key transitions help to normal",
			initialState: StateHelp,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			wantState:    StateNormal,
		},
		{
			name:         "q key transitions help to normal",
			initialState: StateHelp,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantState:    StateNormal,
		},
		{
			name:         "unknown key preserves normal state",
			initialState: StateNormal,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}},
			wantState:    StateNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(tt.initialState)

			newModel, _ := m.Update(tt.msg)
			got := newModel.(Model)

			if got.state != tt.wantState {
				t.Errorf("state = %v, want %v", got.state, tt.wantState)
			}
		})
	}
}

// TestUpdateKeyHandling tests key navigation and tab switching
func TestUpdateKeyHandling(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(m *Model)
		key            tea.KeyMsg
		wantTab        Tab
		wantCursor     int
		wantRolesCursor int
		wantGroupsCursor int
	}{
		{
			name: "tab key cycles from roles to groups",
			setup: func(m *Model) {
				m.activeTab = TabRoles
			},
			key:     tea.KeyMsg{Type: tea.KeyTab},
			wantTab: TabGroups,
		},
		{
			name: "tab key cycles from groups to subscriptions",
			setup: func(m *Model) {
				m.activeTab = TabGroups
			},
			key:     tea.KeyMsg{Type: tea.KeyTab},
			wantTab: TabSubscriptions,
		},
		{
			name: "tab key cycles from subscriptions to roles",
			setup: func(m *Model) {
				m.activeTab = TabSubscriptions
			},
			key:     tea.KeyMsg{Type: tea.KeyTab},
			wantTab: TabRoles,
		},
		{
			name: "left arrow moves from groups to roles",
			setup: func(m *Model) {
				m.activeTab = TabGroups
			},
			key:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
			wantTab: TabRoles,
		},
		{
			name: "right arrow moves from roles to groups",
			setup: func(m *Model) {
				m.activeTab = TabRoles
			},
			key:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
			wantTab: TabGroups,
		},
		{
			name: "left arrow at roles stays at roles",
			setup: func(m *Model) {
				m.activeTab = TabRoles
			},
			key:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}},
			wantTab: TabRoles,
		},
		{
			name: "right arrow at subscriptions stays at subscriptions",
			setup: func(m *Model) {
				m.activeTab = TabSubscriptions
			},
			key:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}},
			wantTab: TabSubscriptions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(StateNormal)
			if tt.setup != nil {
				tt.setup(&m)
			}

			newModel, _ := m.Update(tt.key)
			got := newModel.(Model)

			if got.activeTab != tt.wantTab {
				t.Errorf("activeTab = %v, want %v", got.activeTab, tt.wantTab)
			}
		})
	}
}

// TestUpdateCursorMovement tests cursor movement with j/k keys
func TestUpdateCursorMovement(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(m *Model)
		key        tea.KeyMsg
		wantCursor int
	}{
		{
			name: "j moves cursor down in roles",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{{}, {}, {}}
				m.rolesCursor = 0
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantCursor: 1,
		},
		{
			name: "k moves cursor up in roles",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{{}, {}, {}}
				m.rolesCursor = 2
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			wantCursor: 1,
		},
		{
			name: "k at top stays at 0",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{{}, {}, {}}
				m.rolesCursor = 0
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			wantCursor: 0,
		},
		{
			name: "j at bottom stays at last",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{{}, {}, {}}
				m.rolesCursor = 2
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantCursor: 2,
		},
		{
			name: "j in empty list stays at 0",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{}
				m.rolesCursor = 0
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantCursor: 0,
		},
		{
			name: "down arrow moves cursor",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{{}, {}, {}}
				m.rolesCursor = 0
			},
			key:        tea.KeyMsg{Type: tea.KeyDown},
			wantCursor: 1,
		},
		{
			name: "up arrow moves cursor",
			setup: func(m *Model) {
				m.activeTab = TabRoles
				m.roles = []azure.Role{{}, {}, {}}
				m.rolesCursor = 2
			},
			key:        tea.KeyMsg{Type: tea.KeyUp},
			wantCursor: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(StateNormal)
			if tt.setup != nil {
				tt.setup(&m)
			}

			newModel, _ := m.Update(tt.key)
			got := newModel.(Model)

			if got.rolesCursor != tt.wantCursor {
				t.Errorf("rolesCursor = %v, want %v", got.rolesCursor, tt.wantCursor)
			}
		})
	}
}

// TestUpdateGroupsCursorMovement tests cursor movement in groups tab
func TestUpdateGroupsCursorMovement(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(m *Model)
		key        tea.KeyMsg
		wantCursor int
	}{
		{
			name: "j moves cursor down in groups",
			setup: func(m *Model) {
				m.activeTab = TabGroups
				m.groups = []azure.Group{{}, {}, {}}
				m.groupsCursor = 0
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			wantCursor: 1,
		},
		{
			name: "k moves cursor up in groups",
			setup: func(m *Model) {
				m.activeTab = TabGroups
				m.groups = []azure.Group{{}, {}, {}}
				m.groupsCursor = 2
			},
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			wantCursor: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(StateNormal)
			if tt.setup != nil {
				tt.setup(&m)
			}

			newModel, _ := m.Update(tt.key)
			got := newModel.(Model)

			if got.groupsCursor != tt.wantCursor {
				t.Errorf("groupsCursor = %v, want %v", got.groupsCursor, tt.wantCursor)
			}
		})
	}
}

// TestUpdateAsyncMessages tests handling of async data loading messages
func TestUpdateAsyncMessages(t *testing.T) {
	t.Run("rolesLoadedMsg populates roles and sets flag", func(t *testing.T) {
		m := testModel(StateLoading)
		m.tenant = &azure.Tenant{DisplayName: "Test Tenant"}
		m.groupsLoaded = true
		m.lighthouseLoaded = true

		roles := []azure.Role{
			{DisplayName: "Role1"},
			{DisplayName: "Role2"},
		}

		newModel, _ := m.Update(rolesLoadedMsg{roles: roles})
		got := newModel.(Model)

		if len(got.roles) != 2 {
			t.Errorf("roles length = %d, want 2", len(got.roles))
		}
		if !got.rolesLoaded {
			t.Error("rolesLoaded = false, want true")
		}
		// Should transition to normal since all data is loaded
		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
	})

	t.Run("groupsLoadedMsg populates groups and sets flag", func(t *testing.T) {
		m := testModel(StateLoading)
		m.tenant = &azure.Tenant{DisplayName: "Test Tenant"}
		m.rolesLoaded = true
		m.lighthouseLoaded = true

		groups := []azure.Group{
			{DisplayName: "Group1"},
			{DisplayName: "Group2"},
			{DisplayName: "Group3"},
		}

		newModel, _ := m.Update(groupsLoadedMsg{groups: groups})
		got := newModel.(Model)

		if len(got.groups) != 3 {
			t.Errorf("groups length = %d, want 3", len(got.groups))
		}
		if !got.groupsLoaded {
			t.Error("groupsLoaded = false, want true")
		}
		// Should transition to normal since all data is loaded
		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
	})

	t.Run("lighthouseLoadedMsg populates lighthouse and sets flag", func(t *testing.T) {
		m := testModel(StateLoading)
		m.tenant = &azure.Tenant{DisplayName: "Test Tenant"}
		m.rolesLoaded = true
		m.groupsLoaded = true

		subs := []azure.LighthouseSubscription{
			{DisplayName: "Sub1"},
		}

		newModel, _ := m.Update(lighthouseLoadedMsg{subs: subs})
		got := newModel.(Model)

		if len(got.lighthouse) != 1 {
			t.Errorf("lighthouse length = %d, want 1", len(got.lighthouse))
		}
		if !got.lighthouseLoaded {
			t.Error("lighthouseLoaded = false, want true")
		}
		// Should transition to normal since all data is loaded
		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
	})

	t.Run("stays loading until all data loaded", func(t *testing.T) {
		m := testModel(StateLoading)
		m.tenant = &azure.Tenant{DisplayName: "Test Tenant"}
		// Only roles loaded, groups and lighthouse not yet

		roles := []azure.Role{{DisplayName: "Role1"}}
		newModel, _ := m.Update(rolesLoadedMsg{roles: roles})
		got := newModel.(Model)

		// Should stay loading because groups and lighthouse not loaded yet
		if got.state != StateLoading {
			t.Errorf("state = %v, want StateLoading (groups and lighthouse not loaded)", got.state)
		}
	})

	t.Run("errMsg sets error state and logs", func(t *testing.T) {
		m := testModel(StateNormal)

		newModel, _ := m.Update(errMsg{err: fmt.Errorf("something went wrong"), source: "auth"})
		got := newModel.(Model)

		if got.state != StateError {
			t.Errorf("state = %v, want StateError", got.state)
		}
		if got.err == nil {
			t.Error("err = nil, want error set")
		}
	})

	t.Run("roles error marks rolesLoaded true to allow progress", func(t *testing.T) {
		m := testModel(StateLoading)
		m.tenant = &azure.Tenant{DisplayName: "Test Tenant"}
		m.groupsLoaded = true
		m.lighthouseLoaded = true

		newModel, _ := m.Update(errMsg{err: fmt.Errorf("roles error"), source: "roles"})
		got := newModel.(Model)

		if !got.rolesLoaded {
			t.Error("rolesLoaded = false, want true (to allow UI progress)")
		}
		// With all *Loaded flags true, should transition to normal despite error
		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal (all data sources attempted)", got.state)
		}
	})

	t.Run("userInfoLoadedMsg populates user info", func(t *testing.T) {
		m := testModel(StateLoading)

		newModel, _ := m.Update(userInfoLoadedMsg{displayName: "Test User", email: "test@example.com"})
		got := newModel.(Model)

		if got.userDisplayName != "Test User" {
			t.Errorf("userDisplayName = %q, want %q", got.userDisplayName, "Test User")
		}
		if got.userEmail != "test@example.com" {
			t.Errorf("userEmail = %q, want %q", got.userEmail, "test@example.com")
		}
	})
}

// TestUpdateConfirmStateTransitions tests state transitions in confirm state
func TestUpdateConfirmStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		key       tea.KeyMsg
		wantState State
	}{
		{
			name:      "y transitions to justification",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			wantState: StateJustification,
		},
		{
			name:      "enter transitions to justification",
			key:       tea.KeyMsg{Type: tea.KeyEnter},
			wantState: StateJustification,
		},
		{
			name:      "n cancels back to normal",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			wantState: StateNormal,
		},
		{
			name:      "esc cancels back to normal",
			key:       tea.KeyMsg{Type: tea.KeyEsc},
			wantState: StateNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(StateConfirm)
			m.pendingActivations = []interface{}{"something"}

			newModel, _ := m.Update(tt.key)
			got := newModel.(Model)

			if got.state != tt.wantState {
				t.Errorf("state = %v, want %v", got.state, tt.wantState)
			}
		})
	}
}

// TestUpdateDurationSetting tests duration preset selection
func TestUpdateDurationSetting(t *testing.T) {
	tests := []struct {
		name         string
		key          tea.KeyMsg
		wantDuration time.Duration
		wantIndex    int
	}{
		{
			name:         "1 key sets 1 hour",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}},
			wantDuration: 1 * time.Hour,
			wantIndex:    0,
		},
		{
			name:         "2 key sets 2 hours",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}},
			wantDuration: 2 * time.Hour,
			wantIndex:    1,
		},
		{
			name:         "3 key sets 4 hours",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}},
			wantDuration: 4 * time.Hour,
			wantIndex:    2,
		},
		{
			name:         "4 key sets 8 hours",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}},
			wantDuration: 8 * time.Hour,
			wantIndex:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel(StateNormal)

			newModel, _ := m.Update(tt.key)
			got := newModel.(Model)

			if got.duration != tt.wantDuration {
				t.Errorf("duration = %v, want %v", got.duration, tt.wantDuration)
			}
			if got.durationIndex != tt.wantIndex {
				t.Errorf("durationIndex = %d, want %d", got.durationIndex, tt.wantIndex)
			}
		})
	}
}

// TestUpdateSelectionToggle tests space key selection toggle
func TestUpdateSelectionToggle(t *testing.T) {
	t.Run("space toggles role selection on", func(t *testing.T) {
		m := testModel(StateNormal)
		m.activeTab = TabRoles
		m.roles = []azure.Role{{DisplayName: "Role1"}}
		m.rolesCursor = 0

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		got := newModel.(Model)

		if !got.selectedRoles[0] {
			t.Error("selectedRoles[0] = false, want true")
		}
	})

	t.Run("space toggles role selection off", func(t *testing.T) {
		m := testModel(StateNormal)
		m.activeTab = TabRoles
		m.roles = []azure.Role{{DisplayName: "Role1"}}
		m.rolesCursor = 0
		m.selectedRoles[0] = true

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		got := newModel.(Model)

		if got.selectedRoles[0] {
			t.Error("selectedRoles[0] = true, want false")
		}
	})

	t.Run("space toggles group selection", func(t *testing.T) {
		m := testModel(StateNormal)
		m.activeTab = TabGroups
		m.groups = []azure.Group{{DisplayName: "Group1"}}
		m.groupsCursor = 0

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		got := newModel.(Model)

		if !got.selectedGroups[0] {
			t.Error("selectedGroups[0] = false, want true")
		}
	})
}

// TestUpdateAutoRefreshToggle tests auto-refresh toggle
func TestUpdateAutoRefreshToggle(t *testing.T) {
	t.Run("a toggles auto-refresh off", func(t *testing.T) {
		m := testModel(StateNormal)
		m.autoRefresh = true

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		got := newModel.(Model)

		if got.autoRefresh {
			t.Error("autoRefresh = true, want false")
		}
	})

	t.Run("a toggles auto-refresh on", func(t *testing.T) {
		m := testModel(StateNormal)
		m.autoRefresh = false

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		got := newModel.(Model)

		if !got.autoRefresh {
			t.Error("autoRefresh = false, want true")
		}
	})
}

// TestUpdateSearchStateTransitions tests search mode transitions
func TestUpdateSearchStateTransitions(t *testing.T) {
	t.Run("slash enters search mode", func(t *testing.T) {
		m := testModel(StateNormal)

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		got := newModel.(Model)

		if got.state != StateSearch {
			t.Errorf("state = %v, want StateSearch", got.state)
		}
	})

	t.Run("enter exits search mode", func(t *testing.T) {
		m := testModel(StateSearch)
		m.searchInput.SetValue("test")

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		got := newModel.(Model)

		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
		if got.searchQuery != "test" {
			t.Errorf("searchQuery = %q, want %q", got.searchQuery, "test")
		}
	})

	t.Run("esc exits search mode", func(t *testing.T) {
		m := testModel(StateSearch)
		m.searchInput.SetValue("test")

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		got := newModel.(Model)

		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
	})
}

// TestUpdateErrorStateHandling tests behavior in error state
func TestUpdateErrorStateHandling(t *testing.T) {
	t.Run("r in error state triggers retry", func(t *testing.T) {
		m := testModel(StateError)
		m.err = fmt.Errorf("previous error")

		newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		got := newModel.(Model)

		if got.state != StateLoading {
			t.Errorf("state = %v, want StateLoading", got.state)
		}
		if got.err != nil {
			t.Errorf("err = %v, want nil", got.err)
		}
		if cmd == nil {
			t.Error("cmd = nil, want retry command")
		}
	})

	t.Run("other keys in error state are ignored except q", func(t *testing.T) {
		m := testModel(StateError)

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
		got := newModel.(Model)

		if got.state != StateError {
			t.Errorf("state = %v, want StateError (should stay)", got.state)
		}
	})
}

// TestUpdateLoadingStateHandling tests behavior in loading state
func TestUpdateLoadingStateHandling(t *testing.T) {
	t.Run("most keys ignored in loading state", func(t *testing.T) {
		m := testModel(StateLoading)

		// Try various keys that should be ignored
		keys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'j'}},
			{Type: tea.KeyRunes, Runes: []rune{'k'}},
			{Type: tea.KeyRunes, Runes: []rune{'?'}},
			{Type: tea.KeyEnter},
		}

		for _, key := range keys {
			newModel, _ := m.Update(key)
			got := newModel.(Model)

			if got.state != StateLoading {
				t.Errorf("state = %v after %v, want StateLoading", got.state, key)
			}
		}
	})
}

// TestUpdateActivationDone tests activation completion handling
func TestUpdateActivationDone(t *testing.T) {
	t.Run("successful activation returns to normal", func(t *testing.T) {
		m := testModel(StateActivating)
		m.selectedRoles[0] = true

		newModel, _ := m.Update(activationDoneMsg{err: nil})
		got := newModel.(Model)

		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
		// Selections should be cleared
		if len(got.selectedRoles) != 0 {
			t.Errorf("selectedRoles not cleared, got %d entries", len(got.selectedRoles))
		}
	})

	t.Run("failed activation returns to normal with error logged", func(t *testing.T) {
		m := testModel(StateActivating)
		initialLogCount := len(m.logs)

		newModel, _ := m.Update(activationDoneMsg{err: fmt.Errorf("activation failed")})
		got := newModel.(Model)

		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
		// Should have logged the error
		if len(got.logs) <= initialLogCount {
			t.Error("error should have been logged")
		}
	})
}

// TestUpdateDeactivationDone tests deactivation completion handling
func TestUpdateDeactivationDone(t *testing.T) {
	t.Run("successful deactivation returns to normal", func(t *testing.T) {
		m := testModel(StateDeactivating)
		m.selectedRoles[0] = true

		newModel, _ := m.Update(deactivationDoneMsg{err: nil})
		got := newModel.(Model)

		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
	})

	t.Run("ActiveDurationTooShort error logs specific message", func(t *testing.T) {
		m := testModel(StateDeactivating)
		initialLogCount := len(m.logs)

		newModel, _ := m.Update(deactivationDoneMsg{err: fmt.Errorf("ActiveDurationTooShort")})
		got := newModel.(Model)

		if got.state != StateNormal {
			t.Errorf("state = %v, want StateNormal", got.state)
		}
		// Should have logged the specific message
		if len(got.logs) <= initialLogCount {
			t.Error("error should have been logged")
		}
	})
}

// TestUpdateWindowSize tests window size handling
func TestUpdateWindowSize(t *testing.T) {
	m := testModel(StateNormal)

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 100})
	got := newModel.(Model)

	if got.width != 200 {
		t.Errorf("width = %d, want 200", got.width)
	}
	if got.height != 100 {
		t.Errorf("height = %d, want 100", got.height)
	}
}

// TestUpdateDurationCycle tests d key cycling through durations
func TestUpdateDurationCycle(t *testing.T) {
	m := testModel(StateNormal)
	// Default is 4 hours (index 2)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := newModel.(Model)

	// Should cycle to next (8 hours, index 3)
	if got.durationIndex != 3 {
		t.Errorf("durationIndex = %d, want 3", got.durationIndex)
	}

	// Cycle again to wrap to 0
	newModel2, _ := got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got2 := newModel2.(Model)

	if got2.durationIndex != 0 {
		t.Errorf("durationIndex = %d, want 0 (wrapped)", got2.durationIndex)
	}
}

// TestUpdateLogLevelCycle tests v key cycling through log levels
func TestUpdateLogLevelCycle(t *testing.T) {
	m := testModel(StateNormal)
	// Default is LogInfo (1)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	got := newModel.(Model)

	if got.logLevel != LogDebug {
		t.Errorf("logLevel = %v, want LogDebug", got.logLevel)
	}
}
