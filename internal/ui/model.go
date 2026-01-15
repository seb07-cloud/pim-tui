package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebsebseb1982/pim-tui/internal/azure"
	"github.com/sebsebseb1982/pim-tui/internal/config"
)

// Re-export azure types for convenience
type ActivationStatus = azure.ActivationStatus

const (
	StatusInactive     = azure.StatusInactive
	StatusActive       = azure.StatusActive
	StatusExpiringSoon = azure.StatusExpiringSoon
	StatusPending      = azure.StatusPending
)

type Tab int

const (
	TabRoles Tab = iota
	TabGroups
)

type ViewMode int

const (
	ViewMain ViewMode = iota
	ViewLighthouse
)

type LogLevel int

const (
	LogError LogLevel = iota
	LogInfo
	LogDebug
)

func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	default:
		return "ERROR"
	}
}

type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Message string
}

type ActivationHistoryEntry struct {
	Time          time.Time
	Type          string // "role" or "group"
	Name          string
	Duration      time.Duration
	Justification string
	Success       bool
}

type State int

const (
	StateLoading State = iota
	StateNormal
	StateConfirm
	StateConfirmDeactivate
	StateJustification
	StateActivating
	StateDeactivating
	StateHelp
	StateSearch
	StateError
)

type Model struct {
	// Azure client
	client *azure.Client
	config config.Config

	// Version
	version string

	// Data
	tenant          *azure.Tenant
	roles           []azure.Role
	groups          []azure.Group
	lighthouse      []azure.LighthouseSubscription
	userDisplayName string
	userEmail       string

	// UI state
	activeTab      Tab
	viewMode       ViewMode
	state          State
	rolesCursor    int
	groupsCursor   int
	lightCursor    int
	selectedRoles  map[int]bool
	selectedGroups map[int]bool
	selectedLight  map[int]bool

	// Loading state
	loading        bool
	loadingMessage string
	rolesLoaded    bool
	groupsLoaded   bool

	// Duration
	duration      time.Duration
	durationIndex int

	// Logs
	logs        []LogEntry
	logLevel    LogLevel
	logsCopied  bool
	copyMessage string

	// Auto-refresh
	autoRefresh bool
	lastRefresh time.Time

	// Input
	justificationInput   textinput.Model
	searchInput          textinput.Model
	pendingActivations   []interface{}
	pendingDeactivations []interface{}

	// Search/filter
	searchActive bool
	searchQuery  string

	// Activation history
	activationHistory []ActivationHistoryEntry

	// Help
	help   help.Model
	width  int
	height int

	// Error
	err error
}

// Messages
type clientReadyMsg struct{ client *azure.Client }
type tenantLoadedMsg struct{ tenant *azure.Tenant }
type userInfoLoadedMsg struct {
	displayName string
	email       string
}
type rolesLoadedMsg struct{ roles []azure.Role }
type groupsLoadedMsg struct{ groups []azure.Group }
type lighthouseLoadedMsg struct{ subs []azure.LighthouseSubscription }
type activationDoneMsg struct{ err error }
type deactivationDoneMsg struct{ err error }
type tickMsg time.Time
type errMsg struct {
	err    error
	source string // "roles", "groups", "tenant", etc.
}

func NewModel(cfg config.Config, version string) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter justification..."
	ti.CharLimit = 500

	si := textinput.New()
	si.Placeholder = "Type to filter..."
	si.CharLimit = 100

	return Model{
		config:             cfg,
		version:            version,
		selectedRoles:      make(map[int]bool),
		selectedGroups:     make(map[int]bool),
		selectedLight:      make(map[int]bool),
		duration:           time.Duration(cfg.DefaultDuration) * time.Hour,
		durationIndex:      indexOf(cfg.DurationPresets, cfg.DefaultDuration),
		logLevel:           parseLogLevel(cfg.LogLevel),
		autoRefresh:        cfg.AutoRefreshEnabled,
		help:               help.New(),
		justificationInput: ti,
		searchInput:        si,
		logs:               make([]LogEntry, 0),
		state:              StateLoading,
		loading:            true,
		loadingMessage:     "Authenticating with Azure...",
	}
}

func indexOf(slice []int, val int) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return 0
}

func parseLogLevel(s string) LogLevel {
	switch s {
	case "debug":
		return LogDebug
	case "error":
		return LogError
	default:
		return LogInfo
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		initClientCmd(),
		tickCmd(),
	)
}

func initClientCmd() tea.Cmd {
	return func() tea.Msg {
		client, err := azure.NewClient()
		if err != nil {
			return errMsg{fmt.Errorf("authentication failed: %w", err), "auth"}
		}

		// Test authentication by getting current user
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = client.GetCurrentUser(ctx)
		if err != nil {
			return errMsg{fmt.Errorf("authentication failed: %w", err), "auth"}
		}

		return clientReadyMsg{client}
	}
}

func loadTenantCmd(client *azure.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tenant, err := client.GetTenant(ctx)
		if err != nil {
			return errMsg{fmt.Errorf("failed to get tenant: %w", err), "tenant"}
		}
		return tenantLoadedMsg{tenant}
	}
}

func loadRolesCmd(client *azure.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		roles, err := client.GetRoles(ctx)
		if err != nil {
			return errMsg{fmt.Errorf("failed to load roles: %w", err), "roles"}
		}
		return rolesLoadedMsg{roles}
	}
}

func loadGroupsCmd(client *azure.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		groups, err := client.GetGroups(ctx)
		if err != nil {
			return errMsg{fmt.Errorf("failed to load groups: %w", err), "groups"}
		}
		return groupsLoadedMsg{groups}
	}
}

func loadUserInfoCmd(client *azure.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		displayName, email, err := client.GetCurrentUserInfo(ctx)
		if err != nil {
			// Non-fatal error, just log it
			return userInfoLoadedMsg{"", ""}
		}
		return userInfoLoadedMsg{displayName, email}
	}
}

func loadLighthouseCmd(client *azure.Client, groups []azure.Group) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		subs, err := client.GetLighthouseSubscriptions(ctx, groups)
		if err != nil {
			return errMsg{fmt.Errorf("failed to load lighthouse: %w", err), "lighthouse"}
		}
		return lighthouseLoadedMsg{subs}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) log(level LogLevel, format string, args ...interface{}) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: fmt.Sprintf(format, args...),
	}
	m.logs = append(m.logs, entry)
	// Keep last 100 entries
	if len(m.logs) > 100 {
		m.logs = m.logs[1:]
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case clientReadyMsg:
		m.client = msg.client
		m.loading = true
		m.loadingMessage = "Loading tenant info..."
		m.log(LogInfo, "Authentication successful")
		return m, loadTenantCmd(m.client)

	case tenantLoadedMsg:
		m.tenant = msg.tenant
		m.loadingMessage = "Loading PIM roles and groups..."
		m.log(LogInfo, "Connected to tenant: %s", m.tenant.DisplayName)
		return m, tea.Batch(
			loadRolesCmd(m.client),
			loadGroupsCmd(m.client),
			loadUserInfoCmd(m.client),
		)

	case userInfoLoadedMsg:
		m.userDisplayName = msg.displayName
		m.userEmail = msg.email
		m.log(LogDebug, "User: %s", m.userDisplayName)
		return m, nil

	case rolesLoadedMsg:
		m.roles = msg.roles
		m.rolesLoaded = true
		m.log(LogInfo, "Loaded %d eligible roles", len(m.roles))
		m.checkLoadingComplete()
		return m, nil

	case groupsLoadedMsg:
		m.groups = msg.groups
		m.groupsLoaded = true
		m.log(LogInfo, "Loaded %d eligible groups", len(m.groups))
		m.checkLoadingComplete()
		return m, nil

	case lighthouseLoadedMsg:
		m.lighthouse = msg.subs
		m.log(LogInfo, "Loaded %d lighthouse subscriptions", len(m.lighthouse))
		return m, nil

	case activationDoneMsg:
		m.state = StateNormal
		if msg.err != nil {
			m.log(LogError, "Activation failed: %v", msg.err)
		} else {
			m.log(LogInfo, "Activation completed successfully")
			// Clear selections after successful activation
			m.selectedRoles = make(map[int]bool)
			m.selectedGroups = make(map[int]bool)
		}
		// Refresh after activation
		return m, tea.Batch(
			loadRolesCmd(m.client),
			loadGroupsCmd(m.client),
		)

	case deactivationDoneMsg:
		m.state = StateNormal
		if msg.err != nil {
			errStr := msg.err.Error()
			if strings.Contains(errStr, "ActiveDurationTooShort") {
				m.log(LogError, "Cannot deactivate: role must be active for at least 5 minutes")
			} else {
				m.log(LogError, "Deactivation failed: %v", msg.err)
			}
		} else {
			m.log(LogInfo, "Deactivation completed successfully")
			// Clear selections after successful deactivation
			m.selectedRoles = make(map[int]bool)
			m.selectedGroups = make(map[int]bool)
		}
		// Refresh after deactivation
		return m, tea.Batch(
			loadRolesCmd(m.client),
			loadGroupsCmd(m.client),
		)

	case tickMsg:
		// Always continue ticking for animations
		cmds := []tea.Cmd{tickCmd()}

		// Auto-refresh check (only in normal state)
		if m.autoRefresh && m.client != nil && m.state == StateNormal &&
			time.Since(m.lastRefresh) > time.Duration(m.config.AutoRefreshInterval)*time.Second {
			m.lastRefresh = time.Now()
			m.log(LogDebug, "Auto-refreshing...")
			cmds = append(cmds, loadRolesCmd(m.client), loadGroupsCmd(m.client))
		}

		return m, tea.Batch(cmds...)

	case errMsg:
		m.err = msg.err
		m.log(LogError, "%v", msg.err)

		// Handle errors based on source
		switch msg.source {
		case "auth":
			m.loading = false
			m.state = StateError
		case "roles":
			m.rolesLoaded = true // Mark as loaded even on error so UI progresses
			m.checkLoadingComplete()
		case "groups":
			m.groupsLoaded = true // Mark as loaded even on error so UI progresses
			m.checkLoadingComplete()
		case "tenant":
			m.loading = false
			m.state = StateError
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) checkLoadingComplete() {
	// Consider loading complete when we have tenant and both roles/groups have loaded
	if m.tenant != nil && m.rolesLoaded && m.groupsLoaded {
		m.loading = false
		m.state = StateNormal
		m.lastRefresh = time.Now()
	}
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Always allow quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// In error state, only allow quit or retry
	if m.state == StateError {
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "r":
			m.state = StateLoading
			m.loading = true
			m.loadingMessage = "Retrying authentication..."
			m.err = nil
			return m, initClientCmd()
		}
		return m, nil
	}

	// In loading state, only allow quit
	if m.state == StateLoading {
		if msg.String() == "q" {
			return m, tea.Quit
		}
		return m, nil
	}

	// Handle special states
	switch m.state {
	case StateHelp:
		if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
			m.state = StateNormal
		}
		return m, nil

	case StateConfirm:
		switch msg.String() {
		case "y", "enter":
			m.state = StateJustification
			m.justificationInput.Focus()
			return m, textinput.Blink
		case "n", "esc":
			m.state = StateNormal
			m.pendingActivations = nil
		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			m.setDurationByIndex(idx)
		case "tab":
			m.cycleDuration()
		}
		return m, nil

	case StateConfirmDeactivate:
		switch msg.String() {
		case "y", "enter":
			return m.startDeactivation()
		case "n", "esc":
			m.state = StateNormal
			m.pendingDeactivations = nil
		}
		return m, nil

	case StateJustification:
		switch msg.String() {
		case "enter":
			if strings.TrimSpace(m.justificationInput.Value()) == "" {
				m.log(LogError, "Justification is required")
				return m, nil
			}
			return m.startActivation()
		case "esc":
			m.state = StateNormal
			m.pendingActivations = nil
			return m, nil
		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			m.setDurationByIndex(idx)
			return m, nil
		case "tab":
			m.cycleDuration()
			return m, nil
		default:
			var cmd tea.Cmd
			m.justificationInput, cmd = m.justificationInput.Update(msg)
			return m, cmd
		}

	case StateActivating:
		return m, nil

	case StateSearch:
		switch msg.String() {
		case "enter", "esc":
			m.state = StateNormal
			m.searchQuery = m.searchInput.Value()
			m.searchActive = m.searchQuery != ""
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.searchQuery = m.searchInput.Value()
			m.searchActive = m.searchQuery != ""
			return m, cmd
		}
	}

	// Normal state key handling
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "?":
		m.state = StateHelp
		return m, nil

	case "up", "k":
		m.moveCursor(-1)

	case "down", "j":
		m.moveCursor(1)

	case "left", "h":
		if m.viewMode == ViewMain {
			m.activeTab = TabRoles
		}

	case "right", "l":
		if m.viewMode == ViewMain {
			m.activeTab = TabGroups
		}

	case "tab":
		if m.viewMode == ViewMain {
			if m.activeTab == TabRoles {
				m.activeTab = TabGroups
			} else {
				m.activeTab = TabRoles
			}
		}

	case " ":
		m.toggleSelection()

	case "enter":
		return m.initiateActivation()

	case "x", "delete", "backspace":
		return m.initiateDeactivation()

	case "L":
		if m.viewMode == ViewMain {
			m.viewMode = ViewLighthouse
			if len(m.lighthouse) == 0 && m.client != nil {
				m.log(LogInfo, "Loading lighthouse subscriptions...")
				return m, loadLighthouseCmd(m.client, m.groups)
			}
		} else {
			m.viewMode = ViewMain
		}

	case "r", "R":
		if m.client != nil {
			m.log(LogInfo, "Refreshing...")
			m.lastRefresh = time.Now()
			return m, tea.Batch(
				loadRolesCmd(m.client),
				loadGroupsCmd(m.client),
			)
		}

	case "a":
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			m.log(LogInfo, "Auto-refresh enabled")
		} else {
			m.log(LogInfo, "Auto-refresh disabled")
		}

	case "1", "2", "3", "4":
		idx := int(msg.String()[0] - '1')
		m.setDurationByIndex(idx)

	case "d":
		m.cycleDuration()

	case "v":
		m.cycleLogLevel()

	case "c", "C":
		m.copyLogs()

	case "e", "E":
		m.exportHistory()

	case "/":
		m.state = StateSearch
		m.searchInput.SetValue(m.searchQuery)
		m.searchInput.Focus()
		return m, textinput.Blink

	case "escape":
		// Clear search if active
		if m.searchActive {
			m.searchActive = false
			m.searchQuery = ""
			m.searchInput.SetValue("")
		}
	}

	return m, nil
}

func (m *Model) copyLogs() {
	if len(m.logs) == 0 {
		return
	}

	var lines []string
	for _, entry := range m.logs {
		line := fmt.Sprintf("[%s] %s %s", entry.Level.String(), entry.Time.Format("15:04:05"), entry.Message)
		lines = append(lines, line)
	}

	text := strings.Join(lines, "\n")
	if err := clipboard.WriteAll(text); err != nil {
		m.log(LogError, "Failed to copy to clipboard: %v", err)
	} else {
		m.logsCopied = true
		m.copyMessage = fmt.Sprintf("Copied %d log entries to clipboard", len(m.logs))
		m.log(LogInfo, "Copied %d log entries to clipboard", len(m.logs))
	}
}

func (m *Model) exportHistory() {
	if len(m.activationHistory) == 0 {
		m.log(LogInfo, "No activation history to export")
		return
	}

	var lines []string
	lines = append(lines, "Activation History Export")
	lines = append(lines, fmt.Sprintf("Generated: %s", time.Now().Format(time.RFC3339)))
	lines = append(lines, "")
	lines = append(lines, "Time\tType\tName\tDuration\tJustification\tSuccess")
	lines = append(lines, strings.Repeat("-", 80))

	for _, entry := range m.activationHistory {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%v",
			entry.Time.Format("2006-01-02 15:04:05"),
			entry.Type,
			entry.Name,
			formatDuration(entry.Duration),
			entry.Justification,
			entry.Success,
		)
		lines = append(lines, line)
	}

	text := strings.Join(lines, "\n")
	if err := clipboard.WriteAll(text); err != nil {
		m.log(LogError, "Failed to export history: %v", err)
	} else {
		m.log(LogInfo, "Exported %d activation history entries to clipboard", len(m.activationHistory))
	}
}

func clampCursor(cursor, delta, length int) int {
	if length == 0 {
		return 0
	}
	cursor += delta
	if cursor < 0 {
		return 0
	}
	if cursor >= length {
		return length - 1
	}
	return cursor
}

func (m *Model) moveCursor(delta int) {
	switch {
	case m.viewMode == ViewLighthouse:
		m.lightCursor = clampCursor(m.lightCursor, delta, len(m.lighthouse))
	case m.activeTab == TabRoles:
		m.rolesCursor = clampCursor(m.rolesCursor, delta, len(m.roles))
	case m.activeTab == TabGroups:
		m.groupsCursor = clampCursor(m.groupsCursor, delta, len(m.groups))
	}
}

func (m *Model) toggleSelection() {
	switch {
	case m.viewMode == ViewLighthouse:
		m.selectedLight[m.lightCursor] = !m.selectedLight[m.lightCursor]
	case m.activeTab == TabRoles:
		m.selectedRoles[m.rolesCursor] = !m.selectedRoles[m.rolesCursor]
	case m.activeTab == TabGroups:
		m.selectedGroups[m.groupsCursor] = !m.selectedGroups[m.groupsCursor]
	}
}

func (m *Model) initiateActivation() (tea.Model, tea.Cmd) {
	// Collect pending activations
	m.pendingActivations = nil

	if m.viewMode == ViewLighthouse {
		for idx := range m.selectedLight {
			if idx < len(m.lighthouse) {
				m.pendingActivations = append(m.pendingActivations, m.lighthouse[idx])
			}
		}
	} else {
		for idx := range m.selectedRoles {
			if idx < len(m.roles) {
				m.pendingActivations = append(m.pendingActivations, m.roles[idx])
			}
		}
		for idx := range m.selectedGroups {
			if idx < len(m.groups) {
				m.pendingActivations = append(m.pendingActivations, m.groups[idx])
			}
		}
	}

	if len(m.pendingActivations) == 0 {
		return m, nil
	}

	m.state = StateConfirm
	return m, nil
}

func (m *Model) startActivation() (tea.Model, tea.Cmd) {
	m.state = StateActivating
	justification := m.justificationInput.Value()
	client := m.client
	duration := m.duration
	pending := m.pendingActivations

	// Track activation history
	for _, item := range pending {
		entry := ActivationHistoryEntry{
			Time:          time.Now(),
			Duration:      duration,
			Justification: justification,
			Success:       true, // Will be updated if failed
		}
		switch v := item.(type) {
		case azure.Role:
			entry.Type = "role"
			entry.Name = v.DisplayName
		case azure.Group:
			entry.Type = "group"
			entry.Name = v.DisplayName
		}
		m.activationHistory = append(m.activationHistory, entry)
	}

	return m, func() tea.Msg {
		ctx := context.Background()
		for _, item := range pending {
			switch v := item.(type) {
			case azure.Role:
				if err := client.ActivateRole(ctx, v.RoleDefinitionID, v.DirectoryScopeID, justification, duration); err != nil {
					return activationDoneMsg{err}
				}
			case azure.Group:
				if err := client.ActivateGroup(ctx, v.ID, justification, duration); err != nil {
					return activationDoneMsg{err}
				}
			}
		}
		return activationDoneMsg{nil}
	}
}

func (m *Model) initiateDeactivation() (tea.Model, tea.Cmd) {
	// Collect active items for deactivation
	m.pendingDeactivations = nil

	if m.viewMode == ViewMain {
		// Only collect active roles (StatusActive or StatusExpiringSoon are both active)
		for idx := range m.selectedRoles {
			if idx < len(m.roles) {
				status := m.roles[idx].Status
				if status == azure.StatusActive || status == azure.StatusExpiringSoon {
					m.pendingDeactivations = append(m.pendingDeactivations, m.roles[idx])
				}
			}
		}
		// Only collect active groups (StatusActive or StatusExpiringSoon are both active)
		for idx := range m.selectedGroups {
			if idx < len(m.groups) {
				status := m.groups[idx].Status
				if status == azure.StatusActive || status == azure.StatusExpiringSoon {
					m.pendingDeactivations = append(m.pendingDeactivations, m.groups[idx])
				}
			}
		}
	}

	if len(m.pendingDeactivations) == 0 {
		m.log(LogInfo, "No active items selected for deactivation")
		return m, nil
	}

	m.state = StateConfirmDeactivate
	return m, nil
}

func (m *Model) startDeactivation() (tea.Model, tea.Cmd) {
	m.state = StateDeactivating
	client := m.client
	pending := m.pendingDeactivations

	return m, func() tea.Msg {
		ctx := context.Background()
		for _, item := range pending {
			switch v := item.(type) {
			case azure.Role:
				if err := client.DeactivateRole(ctx, v.RoleDefinitionID, v.DirectoryScopeID); err != nil {
					return deactivationDoneMsg{err}
				}
			case azure.Group:
				if err := client.DeactivateGroup(ctx, v.ID); err != nil {
					return deactivationDoneMsg{err}
				}
			}
		}
		return deactivationDoneMsg{nil}
	}
}

func (m *Model) setDurationByIndex(idx int) {
	if idx >= len(m.config.DurationPresets) {
		return
	}
	m.durationIndex = idx
	m.duration = time.Duration(m.config.DurationPresets[idx]) * time.Hour
	m.log(LogInfo, "Duration set to %d hours", m.config.DurationPresets[idx])
}

func (m *Model) cycleDuration() {
	m.setDurationByIndex((m.durationIndex + 1) % len(m.config.DurationPresets))
}

func (m *Model) cycleLogLevel() {
	m.logLevel = (m.logLevel + 1) % 3
	m.log(LogInfo, "Log level: %s", m.logLevel.String())
}
