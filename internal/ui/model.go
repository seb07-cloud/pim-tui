package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/seb07-cloud/pim-tui/internal/azure"
	"github.com/seb07-cloud/pim-tui/internal/config"
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
	TabSubscriptions
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

// SubscriptionRoleActivation wraps subscription info with the role for activation
type SubscriptionRoleActivation struct {
	SubscriptionID   string
	SubscriptionName string
	Role             azure.EligibleAzureRole
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
	StateUnauthenticated  // User needs to authenticate (not an error, a prompt)
	StateAuthenticating   // Device code auth in progress
)

type Model struct {
	// Azure client
	client *azure.Client
	config config.Config

	// Version
	version string

	// Device code authentication
	deviceCodeMessage string           // Device code message to display during auth
	authCancelFunc    context.CancelFunc // Cancel function for auth context

	// Data
	tenant          *azure.Tenant
	roles           []azure.Role
	groups          []azure.Group
	lighthouse      []azure.LighthouseSubscription
	userDisplayName string
	userEmail       string

	// UI state
	activeTab        Tab
	state            State
	rolesCursor      int
	groupsCursor     int
	lightCursor      int
	subRoleCursor    int  // Cursor for navigating roles within a subscription
	subRoleFocus     bool // True when focus is on role list in detail panel
	selectedRoles    map[int]bool
	selectedGroups   map[int]bool
	selectedLight    map[int]bool
	selectedSubRoles map[string]map[int]bool // subscription ID -> role index -> selected

	// Scroll offsets - independent per panel, preserved across tab switches
	rolesScrollOffset  int // Scroll offset for roles list (index of first visible item)
	groupsScrollOffset int // Scroll offset for groups list
	lightScrollOffset  int // Scroll offset for lighthouse/subscriptions list

	// Loading state
	loading          bool
	loadingMessage   string
	rolesLoaded      bool
	groupsLoaded     bool
	lighthouseLoaded bool

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
type lighthouseLoadedMsg struct {
	subs []azure.LighthouseSubscription
}
type activationDoneMsg struct{ err error }
type deactivationDoneMsg struct{ err error }
type delayedRefreshMsg struct{} // Triggers a refresh after a delay
type tickMsg time.Time
type errMsg struct {
	err    error
	source string // "roles", "groups", "tenant", etc.
}

// authRequiredMsg signals that authentication is required (no valid session)
type authRequiredMsg struct{}

// authCodeMsg carries the device code message for display
type authCodeMsg struct {
	message string // The full message with URL and code
}

// authCompleteMsg signals authentication completed
type authCompleteMsg struct {
	client *azure.Client
	err    error
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
		selectedSubRoles:   make(map[string]map[int]bool),
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
		// Scroll offsets initialized to 0 (top of list)
		rolesScrollOffset:  0,
		groupsScrollOffset: 0,
		lightScrollOffset:  0,
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
			// Check if this is an auth-related error that can be resolved with device code login
			if isAuthError(err) {
				return authRequiredMsg{}
			}
			return errMsg{fmt.Errorf("authentication failed: %w", err), "auth"}
		}

		// Test authentication by getting current user
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = client.GetCurrentUser(ctx)
		if err != nil {
			// Check if this is an auth-related error
			if isAuthError(err) {
				return authRequiredMsg{}
			}
			return errMsg{fmt.Errorf("authentication failed: %w", err), "auth"}
		}

		return clientReadyMsg{client}
	}
}

// isAuthError checks if the error is related to authentication/credentials
// that could be resolved with device code login
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	authKeywords := []string{
		"credential", "login", "authenticate", "token",
		"az login", "not logged in", "unauthorized",
		"failed to get token", "no credential",
	}
	for _, keyword := range authKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
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

// delayedRefreshCmd waits for a specified duration then triggers a refresh
func delayedRefreshCmd(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return delayedRefreshMsg{}
	})
}

// startAuthCmd starts the device code authentication flow.
// It uses a channel to communicate the device code message back to the UI.
func startAuthCmd(ctx context.Context) tea.Cmd {
	// Channel to receive device code message
	codeChan := make(chan string, 1)

	// Start auth in goroutine
	authCmd := func() tea.Msg {
		client, err := azure.AuthenticateWithDeviceCode(ctx, func(message string) error {
			// Send device code message to channel (non-blocking)
			select {
			case codeChan <- message:
			default:
			}
			return nil
		})
		return authCompleteMsg{client: client, err: err}
	}

	// Listen for device code message
	codeListenerCmd := func() tea.Msg {
		select {
		case msg := <-codeChan:
			return authCodeMsg{message: msg}
		case <-ctx.Done():
			return nil
		}
	}

	return tea.Batch(authCmd, codeListenerCmd)
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

	case authRequiredMsg:
		// No valid Azure CLI session, show friendly login prompt
		m.loading = false
		m.state = StateUnauthenticated
		m.err = nil
		m.deviceCodeMessage = ""
		m.log(LogInfo, "Authentication required - press L to login")
		return m, nil

	case authCodeMsg:
		// Device code received, display to user
		m.deviceCodeMessage = msg.message
		m.log(LogInfo, "Device code received - complete login in browser")
		return m, nil

	case authCompleteMsg:
		// Cancel func is no longer needed
		m.authCancelFunc = nil
		if msg.err != nil {
			// Auth failed, go back to unauthenticated state
			m.state = StateUnauthenticated
			m.deviceCodeMessage = ""
			m.log(LogError, "Authentication failed: %v", msg.err)
			return m, nil
		}
		// Auth succeeded, proceed to loading
		m.client = msg.client
		m.state = StateLoading
		m.loading = true
		m.loadingMessage = "Loading tenant info..."
		m.deviceCodeMessage = ""
		m.log(LogInfo, "Authentication successful")
		return m, loadTenantCmd(m.client)

	case tenantLoadedMsg:
		m.tenant = msg.tenant
		m.loadingMessage = "Loading PIM roles and groups..."
		m.log(LogInfo, "Connected to tenant: %s", m.tenant.DisplayName)
		return m, tea.Batch(m.refreshCmd(), loadUserInfoCmd(m.client))

	case userInfoLoadedMsg:
		m.userDisplayName = msg.displayName
		m.userEmail = msg.email
		m.log(LogDebug, "User: %s", m.userDisplayName)
		return m, nil

	case rolesLoadedMsg:
		m.roles = msg.roles
		m.rolesLoaded = true
		// Clamp scroll offset if list got shorter
		if m.rolesScrollOffset >= len(m.roles) && len(m.roles) > 0 {
			m.rolesScrollOffset = len(m.roles) - 1
		} else if len(m.roles) == 0 {
			m.rolesScrollOffset = 0
		}
		m.log(LogInfo, "Loaded %d eligible roles", len(m.roles))
		m.checkLoadingComplete()
		return m, nil

	case groupsLoadedMsg:
		m.groups = msg.groups
		m.groupsLoaded = true
		// Clamp scroll offset if list got shorter
		if m.groupsScrollOffset >= len(m.groups) && len(m.groups) > 0 {
			m.groupsScrollOffset = len(m.groups) - 1
		} else if len(m.groups) == 0 {
			m.groupsScrollOffset = 0
		}
		m.log(LogInfo, "Loaded %d eligible groups", len(m.groups))
		m.checkLoadingComplete()
		return m, nil

	case lighthouseLoadedMsg:
		m.lighthouse = msg.subs
		m.lighthouseLoaded = true
		// Sort by tenant name (already populated during load with cache)
		sort.Slice(m.lighthouse, func(i, j int) bool {
			if m.lighthouse[i].TenantName != m.lighthouse[j].TenantName {
				return m.lighthouse[i].TenantName < m.lighthouse[j].TenantName
			}
			return m.lighthouse[i].DisplayName < m.lighthouse[j].DisplayName
		})
		// Clamp scroll offset if list got shorter
		if m.lightScrollOffset >= len(m.lighthouse) && len(m.lighthouse) > 0 {
			m.lightScrollOffset = len(m.lighthouse) - 1
		} else if len(m.lighthouse) == 0 {
			m.lightScrollOffset = 0
		}
		// Count total eligible roles across all subscriptions
		totalRoles := 0
		for _, sub := range m.lighthouse {
			totalRoles += len(sub.EligibleRoles)
		}
		m.log(LogInfo, "Loaded %d subscriptions with %d eligible roles", len(m.lighthouse), totalRoles)
		m.checkLoadingComplete()
		return m, nil

	case activationDoneMsg:
		m.state = StateNormal
		if msg.err != nil {
			m.log(LogError, "Activation failed: %v", msg.err)
			return m, nil
		}
		m.log(LogInfo, "Activation completed successfully")
		m.clearSelections()
		// Immediate refresh + delayed refresh after 5s for Azure to process
		return m, tea.Batch(m.refreshCmd(), delayedRefreshCmd(5*time.Second))

	case deactivationDoneMsg:
		m.state = StateNormal
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "ActiveDurationTooShort") {
				m.log(LogError, "Cannot deactivate: role must be active for at least 5 minutes")
			} else {
				m.log(LogError, "Deactivation failed: %v", msg.err)
			}
			return m, nil
		}
		m.log(LogInfo, "Deactivation completed successfully")
		m.clearSelections()
		// Immediate refresh + delayed refresh after 5s for Azure to process
		return m, tea.Batch(m.refreshCmd(), delayedRefreshCmd(5*time.Second))

	case delayedRefreshMsg:
		// Delayed refresh triggered after activation/deactivation
		if m.client != nil && m.state == StateNormal {
			m.log(LogDebug, "Post-activation refresh...")
			return m, m.refreshCmd()
		}
		return m, nil

	case tickMsg:
		// Auto-refresh check (only in normal state)
		if m.autoRefresh && m.client != nil && m.state == StateNormal &&
			time.Since(m.lastRefresh) > time.Duration(m.config.AutoRefreshInterval)*time.Second {
			m.lastRefresh = time.Now()
			m.log(LogDebug, "Auto-refreshing...")
			return m, tea.Batch(tickCmd(), m.refreshCmd())
		}
		return m, tickCmd()

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
		case "lighthouse":
			m.lighthouseLoaded = true // Mark as loaded even on error so UI progresses
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
	// Consider loading complete when we have tenant and all data sources have loaded
	if m.tenant != nil && m.rolesLoaded && m.groupsLoaded && m.lighthouseLoaded {
		m.loading = false
		m.state = StateNormal
		m.lastRefresh = time.Now()
	}
}

func (m *Model) refreshCmd() tea.Cmd {
	return tea.Batch(
		loadRolesCmd(m.client),
		loadGroupsCmd(m.client),
		loadLighthouseCmd(m.client, nil), // groups not needed for lighthouse API
	)
}

func (m *Model) clearSelections() {
	m.selectedRoles = make(map[int]bool)
	m.selectedGroups = make(map[int]bool)
	m.selectedLight = make(map[int]bool)
	m.selectedSubRoles = make(map[string]map[int]bool)
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

	// In unauthenticated state, allow login or quit
	if m.state == StateUnauthenticated {
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "l", "L":
			// Start device code authentication flow
			ctx, cancel := context.WithCancel(context.Background())
			m.authCancelFunc = cancel
			m.state = StateAuthenticating
			m.deviceCodeMessage = ""
			m.log(LogInfo, "Starting device code authentication...")
			return m, startAuthCmd(ctx)
		}
		return m, nil
	}

	// In authenticating state, allow quit (which cancels auth)
	if m.state == StateAuthenticating {
		switch msg.String() {
		case "q":
			// Cancel ongoing auth
			if m.authCancelFunc != nil {
				m.authCancelFunc()
				m.authCancelFunc = nil
			}
			return m, tea.Quit
		case "esc":
			// Cancel auth and go back to unauthenticated
			if m.authCancelFunc != nil {
				m.authCancelFunc()
				m.authCancelFunc = nil
			}
			m.state = StateUnauthenticated
			m.deviceCodeMessage = ""
			m.log(LogInfo, "Authentication cancelled")
			return m, nil
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
			_, err := validateJustification(m.justificationInput.Value())
			if err != nil {
				m.log(LogError, "%v", err)
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
		// If in subscription role focus, exit back to subscription list
		if m.activeTab == TabSubscriptions && m.subRoleFocus {
			m.subRoleFocus = false
			return m, nil
		}
		// Switch tabs - scroll offsets preserved independently per panel
		if m.activeTab > TabRoles {
			m.activeTab--
			m.subRoleFocus = false
		}

	case "right", "l":
		// If on subscriptions tab with roles available, enter role focus mode
		if m.activeTab == TabSubscriptions && !m.subRoleFocus {
			if m.lightCursor < len(m.lighthouse) && len(m.lighthouse[m.lightCursor].EligibleRoles) > 0 {
				m.subRoleFocus = true
				m.subRoleCursor = 0
				return m, nil
			}
		}
		// Switch tabs - scroll offsets preserved independently per panel
		if m.activeTab < TabSubscriptions {
			m.activeTab++
			m.subRoleFocus = false
		}

	case "tab":
		// If on subscriptions tab, toggle between list and role detail
		if m.activeTab == TabSubscriptions {
			if m.lightCursor < len(m.lighthouse) && len(m.lighthouse[m.lightCursor].EligibleRoles) > 0 {
				m.subRoleFocus = !m.subRoleFocus
				if m.subRoleFocus {
					m.subRoleCursor = 0
				}
				return m, nil
			}
		}
		// Cycle tabs - scroll offsets preserved independently per panel
		m.activeTab = (m.activeTab + 1) % 3
		m.subRoleFocus = false

	case " ":
		m.toggleSelection()

	case "enter":
		return m.initiateActivation()

	case "x", "delete":
		return m.initiateDeactivation()

	case "backspace":
		// In subscriptions tab with active search, backspace removes last char
		if m.activeTab == TabSubscriptions && m.searchQuery != "" {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.searchInput.SetValue(m.searchQuery)
			m.searchActive = m.searchQuery != ""
			// Reset cursor to first visible item
			visibleIndices := m.getVisibleSubscriptionIndices()
			if len(visibleIndices) > 0 {
				m.lightCursor = visibleIndices[0]
			}
			return m, nil
		}
		// Otherwise, treat as deactivation
		return m.initiateDeactivation()

	case "r", "R":
		if m.client != nil {
			m.log(LogInfo, "Refreshing...")
			m.lastRefresh = time.Now()
			return m, m.refreshCmd()
		}

	case "a":
		m.autoRefresh = !m.autoRefresh
		m.log(LogInfo, "Auto-refresh %s", map[bool]string{true: "enabled", false: "disabled"}[m.autoRefresh])

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

	default:
		// Inline search for subscriptions tab - typing filters directly
		if m.activeTab == TabSubscriptions && !m.subRoleFocus {
			// Only handle printable characters (letters, numbers, spaces)
			if len(msg.String()) == 1 {
				char := msg.String()[0]
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9') || char == ' ' || char == '-' || char == '_' {
					m.searchQuery += msg.String()
					m.searchInput.SetValue(m.searchQuery)
					m.searchActive = true
					// Reset cursor to first visible item when search changes
					visibleIndices := m.getVisibleSubscriptionIndices()
					if len(visibleIndices) > 0 {
						m.lightCursor = visibleIndices[0]
					}
				}
			}
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

// getVisibleSubscriptionIndices returns indices of subscriptions that match the current search filter
func (m *Model) getVisibleSubscriptionIndices() []int {
	indices := make([]int, 0, len(m.lighthouse))
	for i, sub := range m.lighthouse {
		if m.searchActive && m.searchQuery != "" {
			query := strings.ToLower(m.searchQuery)
			// Search in subscription name, tenant name, and role names
			match := strings.Contains(strings.ToLower(sub.DisplayName), query) ||
				strings.Contains(strings.ToLower(sub.TenantName), query)
			if !match {
				for _, role := range sub.EligibleRoles {
					if strings.Contains(strings.ToLower(role.RoleDefinitionName), query) {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}
		indices = append(indices, i)
	}
	return indices
}

// getCurrentSubscription returns the subscription that should be displayed in the detail pane
func (m *Model) getCurrentSubscription() *azure.LighthouseSubscription {
	if len(m.lighthouse) == 0 {
		return nil
	}
	visibleIndices := m.getVisibleSubscriptionIndices()
	if len(visibleIndices) == 0 {
		return nil
	}
	// Find cursor position in visible list, default to first visible if cursor is on hidden item
	for _, idx := range visibleIndices {
		if idx == m.lightCursor {
			return &m.lighthouse[idx]
		}
	}
	// Cursor is on a hidden item, return first visible
	return &m.lighthouse[visibleIndices[0]]
}

func (m *Model) moveCursor(delta int) {
	// Calculate approximate display height for scroll adjustment
	// Panel height is m.height - 25 (header, tabs, logs, status), minus 3 for panel chrome
	displayHeight := m.height - 28
	if displayHeight < 5 {
		displayHeight = 5 // Minimum reasonable height
	}

	switch m.activeTab {
	case TabSubscriptions:
		if m.subRoleFocus {
			// Navigate roles within current subscription
			sub := m.getCurrentSubscription()
			if sub != nil {
				m.subRoleCursor = clampCursor(m.subRoleCursor, delta, len(sub.EligibleRoles))
			}
		} else {
			// Navigate through visible subscriptions only
			visibleIndices := m.getVisibleSubscriptionIndices()
			if len(visibleIndices) == 0 {
				return
			}
			// Find current position in visible list
			currentVisibleIdx := 0
			for idx, i := range visibleIndices {
				if i == m.lightCursor {
					currentVisibleIdx = idx
					break
				}
			}
			// Move within visible list
			newVisibleIdx := clampCursor(currentVisibleIdx, delta, len(visibleIndices))
			oldCursor := m.lightCursor
			m.lightCursor = visibleIndices[newVisibleIdx]
			// Reset role cursor when changing subscriptions
			if oldCursor != m.lightCursor {
				m.subRoleCursor = 0
			}
			// Adjust scroll offset to keep cursor visible
			m.lightScrollOffset = m.adjustScrollOffset(newVisibleIdx, m.lightScrollOffset, len(visibleIndices), displayHeight)
		}
	case TabRoles:
		m.rolesCursor = clampCursor(m.rolesCursor, delta, len(m.roles))
		// Adjust scroll offset to keep cursor visible
		m.rolesScrollOffset = m.adjustScrollOffset(m.rolesCursor, m.rolesScrollOffset, len(m.roles), displayHeight)
	case TabGroups:
		m.groupsCursor = clampCursor(m.groupsCursor, delta, len(m.groups))
		// Adjust scroll offset to keep cursor visible
		m.groupsScrollOffset = m.adjustScrollOffset(m.groupsCursor, m.groupsScrollOffset, len(m.groups), displayHeight)
	}
}

// adjustScrollOffset adjusts scroll offset to keep cursor visible within display window
func (m *Model) adjustScrollOffset(cursor, currentOffset, listLen, displayHeight int) int {
	if listLen <= displayHeight {
		// List fits in window, no scrolling needed
		return 0
	}

	// Ensure offset is valid
	offset := currentOffset
	maxOffset := listLen - displayHeight
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	// If cursor is above visible window, scroll up
	if cursor < offset {
		offset = cursor
	}

	// If cursor is below visible window, scroll down
	if cursor >= offset+displayHeight {
		offset = cursor - displayHeight + 1
	}

	return offset
}

func (m *Model) toggleSelection() {
	switch m.activeTab {
	case TabSubscriptions:
		if m.subRoleFocus {
			// Toggle selection for a specific role within the subscription
			if m.lightCursor < len(m.lighthouse) {
				sub := m.lighthouse[m.lightCursor]
				if m.subRoleCursor < len(sub.EligibleRoles) {
					if m.selectedSubRoles[sub.ID] == nil {
						m.selectedSubRoles[sub.ID] = make(map[int]bool)
					}
					m.selectedSubRoles[sub.ID][m.subRoleCursor] = !m.selectedSubRoles[sub.ID][m.subRoleCursor]
					if !m.selectedSubRoles[sub.ID][m.subRoleCursor] {
						delete(m.selectedSubRoles[sub.ID], m.subRoleCursor)
					}
				}
			}
		} else {
			// Toggle all roles for the subscription
			if m.lightCursor < len(m.lighthouse) {
				sub := m.lighthouse[m.lightCursor]
				// Check if any roles are currently selected
				anySelected := false
				if m.selectedSubRoles[sub.ID] != nil {
					for range m.selectedSubRoles[sub.ID] {
						anySelected = true
						break
					}
				}
				if anySelected {
					// Deselect all
					delete(m.selectedSubRoles, sub.ID)
				} else {
					// Select all roles
					m.selectedSubRoles[sub.ID] = make(map[int]bool)
					for i := range sub.EligibleRoles {
						m.selectedSubRoles[sub.ID][i] = true
					}
				}
			}
		}
	case TabRoles:
		m.selectedRoles[m.rolesCursor] = !m.selectedRoles[m.rolesCursor]
	case TabGroups:
		m.selectedGroups[m.groupsCursor] = !m.selectedGroups[m.groupsCursor]
	}
}

func (m *Model) initiateActivation() (tea.Model, tea.Cmd) {
	// Collect pending activations
	m.pendingActivations = nil

	switch m.activeTab {
	case TabSubscriptions:
		// Collect selected roles from subscriptions
		for subID, roleSelections := range m.selectedSubRoles {
			// Find the subscription
			for _, sub := range m.lighthouse {
				if sub.ID == subID {
					for roleIdx := range roleSelections {
						if roleIdx < len(sub.EligibleRoles) {
							m.pendingActivations = append(m.pendingActivations, SubscriptionRoleActivation{
								SubscriptionID:   sub.ID,
								SubscriptionName: sub.DisplayName,
								Role:             sub.EligibleRoles[roleIdx],
							})
						}
					}
					break
				}
			}
		}
	case TabRoles:
		for idx := range m.selectedRoles {
			if idx < len(m.roles) {
				m.pendingActivations = append(m.pendingActivations, m.roles[idx])
			}
		}
	case TabGroups:
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
		case SubscriptionRoleActivation:
			entry.Type = "azure-role"
			entry.Name = fmt.Sprintf("%s on %s", v.Role.RoleDefinitionName, v.SubscriptionName)
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
				if err := client.ActivateGroup(ctx, v.ID, v.RoleDefinitionID, justification, duration); err != nil {
					return activationDoneMsg{err}
				}
			case SubscriptionRoleActivation:
				if err := client.ActivateAzureRole(ctx, v.Role.Scope, v.Role.RoleDefinitionID, v.Role.RoleEligibilityID, justification, duration); err != nil {
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

	switch m.activeTab {
	case TabRoles:
		for idx := range m.selectedRoles {
			if idx < len(m.roles) && m.roles[idx].Status.IsActive() {
				m.pendingDeactivations = append(m.pendingDeactivations, m.roles[idx])
			}
		}
	case TabGroups:
		for idx := range m.selectedGroups {
			if idx < len(m.groups) && m.groups[idx].Status.IsActive() {
				m.pendingDeactivations = append(m.pendingDeactivations, m.groups[idx])
			}
		}
	case TabSubscriptions:
		// Collect selected active roles from subscriptions
		for subID, roleSelections := range m.selectedSubRoles {
			for _, sub := range m.lighthouse {
				if sub.ID == subID {
					for roleIdx := range roleSelections {
						if roleIdx < len(sub.EligibleRoles) && sub.EligibleRoles[roleIdx].Status.IsActive() {
							m.pendingDeactivations = append(m.pendingDeactivations, SubscriptionRoleActivation{
								SubscriptionID:   sub.ID,
								SubscriptionName: sub.DisplayName,
								Role:             sub.EligibleRoles[roleIdx],
							})
						}
					}
					break
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
				if err := client.DeactivateGroup(ctx, v.ID, v.RoleDefinitionID); err != nil {
					return deactivationDoneMsg{err}
				}
			case SubscriptionRoleActivation:
				if err := client.DeactivateAzureRole(ctx, v.Role.Scope, v.Role.RoleDefinitionID); err != nil {
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

// validateJustification checks the justification input for validity.
// Returns the cleaned string if valid, or an error describing the issue.
func validateJustification(input string) (string, error) {
	cleaned := strings.TrimSpace(input)

	if cleaned == "" {
		return "", fmt.Errorf("justification is required")
	}

	// Check for control characters (ASCII 0-31 except 9=tab, 10=newline, 13=CR)
	for _, r := range cleaned {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			return "", fmt.Errorf("justification contains invalid control characters")
		}
		if r == 127 { // DEL character
			return "", fmt.Errorf("justification contains invalid control characters")
		}
	}

	if len(cleaned) > 500 {
		return "", fmt.Errorf("justification exceeds 500 character limit (%d chars)", len(cleaned))
	}

	return cleaned, nil
}
