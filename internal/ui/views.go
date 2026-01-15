package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sebsebseb1982/pim-tui/internal/azure"
)

const asciiLogo = ` ██████╗ ██╗███╗   ███╗    ████████╗██╗   ██╗██╗
 ██╔══██╗██║████╗ ████║    ╚══██╔══╝██║   ██║██║
 ██████╔╝██║██╔████╔██║       ██║   ██║   ██║██║
 ██╔═══╝ ██║██║╚██╔╝██║       ██║   ██║   ██║██║
 ██║     ██║██║ ╚═╝ ██║       ██║   ╚██████╔╝██║
 ╚═╝     ╚═╝╚═╝     ╚═╝       ╚═╝    ╚═════╝ ╚═╝`

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Show loading or error state
	if m.state == StateLoading {
		sections = append(sections, m.renderLoading())
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	if m.state == StateError {
		sections = append(sections, m.renderError())
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	// Header with tenant info
	sections = append(sections, m.renderHeader())

	// Main content
	switch m.state {
	case StateHelp:
		sections = append(sections, m.renderHelp())
	case StateConfirm:
		sections = append(sections, m.renderConfirm())
	case StateConfirmDeactivate:
		sections = append(sections, m.renderConfirmDeactivate())
	case StateJustification:
		sections = append(sections, m.renderJustification())
	case StateActivating:
		sections = append(sections, m.renderActivating())
	case StateDeactivating:
		sections = append(sections, m.renderDeactivating())
	case StateSearch:
		sections = append(sections, m.renderSearch())
	default:
		if m.viewMode == ViewLighthouse {
			sections = append(sections, m.renderLighthouseView())
		} else {
			sections = append(sections, m.renderMainView())
		}
	}

	// Log panel
	sections = append(sections, m.renderLogs())

	// Status bar
	sections = append(sections, m.renderStatusBar())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderLoading() string {
	spin := spinner(colorActive)

	// Build step indicators
	steps := []struct {
		name   string
		done   bool
		active bool
	}{
		{"Initializing Azure SDK", m.client != nil, m.client == nil},
		{"Loading tenant info", m.tenant != nil, m.client != nil && m.tenant == nil},
		{"Loading PIM roles", m.rolesLoaded, m.tenant != nil && !m.rolesLoaded},
		{"Loading PIM groups", m.groupsLoaded, m.tenant != nil && !m.groupsLoaded},
	}

	var stepLines []string
	for _, step := range steps {
		var icon string
		var style lipgloss.Style
		if step.done {
			icon, style = "✓", activeStyle
		} else if step.active {
			icon, style = spin, highlightBoldStyle
		} else {
			icon, style = "○", dimStyle
		}
		stepLines = append(stepLines, style.Render(fmt.Sprintf("  %s %s", icon, step.name)))
	}

	contentParts := []string{
		highlightBoldStyle.Render(asciiLogo),
		dimStyle.MarginTop(1).Render(fmt.Sprintf("v%s", m.version)),
	}

	if m.tenant != nil {
		contentParts = append(contentParts, activeStyle.MarginTop(1).Render(fmt.Sprintf("Connected to: %s", m.tenant.DisplayName)))
	}

	contentParts = append(contentParts,
		detailValueStyle.MarginTop(2).Render(spin+" "+m.loadingMessage),
		lipgloss.NewStyle().MarginTop(2).Render(strings.Join(stepLines, "\n")),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, contentParts...))
}

func (m Model) renderError() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		errorBoldStyle.Render(asciiLogo),
		errorBoldStyle.MarginTop(2).Render("Authentication Failed"),
		detailValueStyle.MarginTop(1).Render(m.err.Error()),
		dimStyle.MarginTop(2).Render("Press 'r' to retry or 'q' to quit"),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderHeader() string {
	// The ASCII logo needs ~55 chars width (49 chars + border/padding)
	// Only show logo if we have at least 110 chars total width
	minWidthForLogo := 110
	showLogo := m.width >= minWidthForLogo

	var logoBoxWidth, infoBoxWidth int
	if showLogo {
		// Two panels side by side: each has border (2) + padding (2) = 4 chars
		totalWidth := m.width - 8
		// Use same width as list panel in main view (45%)
		logoBoxWidth = totalWidth * 9 / 20
		infoBoxWidth = totalWidth - logoBoxWidth
	} else {
		// Single panel: border (2) + padding (2) = 4 chars
		infoBoxWidth = m.width - 4
	}

	// Right box: Info - use shared styles
	var infoLines []string

	// Tenant Name and ID
	if m.tenant != nil {
		infoLines = append(infoLines, dimStyle.Render("Tenant:  ")+activeBoldStyle.Render(m.tenant.DisplayName))
		infoLines = append(infoLines, dimStyle.Render("ID:      ")+detailValueStyle.Render(m.tenant.ID))
	} else {
		infoLines = append(infoLines, dimStyle.Render("Tenant:  ")+detailValueStyle.Render("Connecting..."))
		infoLines = append(infoLines, dimStyle.Render("ID:      ")+detailValueStyle.Render("-"))
	}

	// User Principal Name
	user := "-"
	if m.userEmail != "" {
		user = m.userEmail
	} else if m.userDisplayName != "" {
		user = m.userDisplayName
	}
	infoLines = append(infoLines, dimStyle.Render("User:    ")+detailValueStyle.Render(user))

	// Active Roles count
	activeRoles, activeGroups := m.countActiveItems()
	if activeRoles+activeGroups > 0 {
		infoLines = append(infoLines, dimStyle.Render("Active:  ")+activeBoldStyle.Render(fmt.Sprintf("%d roles, %d groups", activeRoles, activeGroups)))
	} else {
		infoLines = append(infoLines, dimStyle.Render("Active:  ")+detailValueStyle.Render("0"))
	}

	// Refresh state
	var refreshStr string
	if m.autoRefresh {
		if secs, ok := m.refreshCountdown(); ok {
			refreshStr = activeBoldStyle.Render(fmt.Sprintf("Auto (%ds)", secs))
		} else {
			refreshStr = activeBoldStyle.Render("Auto (ON)")
		}
	} else if !m.lastRefresh.IsZero() {
		elapsed := time.Since(m.lastRefresh)
		if elapsed < time.Minute {
			refreshStr = detailValueStyle.Render("just now")
		} else {
			refreshStr = detailValueStyle.Render(fmt.Sprintf("%dm ago", int(elapsed.Minutes())))
		}
	} else {
		refreshStr = detailValueStyle.Render("-")
	}
	infoLines = append(infoLines, dimStyle.Render("Refresh: ")+refreshStr)

	// Version
	infoLines = append(infoLines, dimStyle.Render("Version: ")+detailValueStyle.Render(fmt.Sprintf("v%s", m.version)))

	// Mode/Search indicators (if any)
	if m.viewMode == ViewLighthouse {
		infoLines = append(infoLines, highlightBoldStyle.Render("[LIGHTHOUSE]"))
	} else if m.searchActive {
		infoLines = append(infoLines, detailLabelStyle.Render(fmt.Sprintf("[SEARCH: %s]", m.searchQuery)))
	}

	infoContent := strings.Join(infoLines, "\n")

	infoBox := panelStyle.
		Width(infoBoxWidth).
		Height(7).
		Render(infoContent)

	if showLogo {
		logoBox := panelStyle.
			Width(logoBoxWidth).
			Height(7).
			Align(lipgloss.Center, lipgloss.Center).
			Render(highlightBoldStyle.Render(asciiLogo))
		return lipgloss.JoinHorizontal(lipgloss.Top, logoBox, infoBox)
	}

	return infoBox
}

func (m Model) renderMainView() string {
	tabBar := m.renderTabBar()

	// Panel dimensions
	totalWidth := m.width - 8
	listPanelWidth := totalWidth * 9 / 20
	detailPanelWidth := totalWidth - listPanelWidth
	panelHeight := m.height - 25

	// Select content based on active tab
	var title, listContent, detailContent string
	if m.activeTab == TabRoles {
		title, listContent, detailContent = "● PIM Roles", m.renderRolesList(panelHeight-2), m.renderRoleDetail()
	} else {
		title, listContent, detailContent = "● PIM Groups", m.renderGroupsList(panelHeight-2), m.renderGroupDetail()
	}

	listPanel := activePanelStyle.Width(listPanelWidth).Height(panelHeight).Render(
		panelTitleStyle.Foreground(colorHighlight).Render(title) + "\n" + listContent,
	)
	detailPanel := panelStyle.Width(detailPanelWidth).Height(panelHeight).Render(detailContent)

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel))
}

func (m Model) renderTabBar() string {
	tabStyle := func(active bool) lipgloss.Style {
		if active {
			return activeTabStyle
		}
		return inactiveTabStyle
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom,
		tabStyle(m.activeTab == TabRoles).Render("Roles"), " ",
		tabStyle(m.activeTab == TabGroups).Render("Groups"),
	)
	return lipgloss.NewStyle().Width(m.width - 4).Padding(0, 1).Render(tabs + dimStyle.Render("  (Tab/←→ to switch)"))
}

func (m Model) renderRoleDetail() string {
	if len(m.roles) == 0 || m.rolesCursor >= len(m.roles) {
		return detailDimStyle.Render("No role selected")
	}

	role := m.roles[m.rolesCursor]
	var lines []string

	lines = append(lines, detailTitleStyle.Render("Role Details"), "")
	lines = append(lines, detailLabelStyle.Render("Name: ")+detailValueStyle.Render(role.DisplayName))
	lines = append(lines, detailLabelStyle.Render("Status: ")+statusIcon(role.Status)+" "+role.Status.String())
	if exp := renderExpiryLine(role.ExpiresAt); exp != "" {
		lines = append(lines, exp)
	}

	lines = append(lines, "", detailLabelStyle.Render("Permissions:"))
	if permissions := GetRolePermissions(role.RoleDefinitionID); len(permissions) > 0 {
		for _, perm := range permissions {
			lines = append(lines, detailDimStyle.Render("  • "+perm))
		}
	} else {
		lines = append(lines, detailDimStyle.Italic(true).Render("  (permissions not available)"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderGroupDetail() string {
	if len(m.groups) == 0 || m.groupsCursor >= len(m.groups) {
		return detailDimStyle.Render("No group selected")
	}

	group := m.groups[m.groupsCursor]
	var lines []string

	lines = append(lines, detailTitleStyle.Render("Group Details"), "")
	lines = append(lines, detailLabelStyle.Render("Name: ")+detailValueStyle.Render(group.DisplayName))

	if group.Description != "" {
		lines = append(lines, detailLabelStyle.Render("Type: ")+detailValueStyle.Render(group.Description))
	}

	lines = append(lines, detailLabelStyle.Render("Status: ")+statusIcon(group.Status)+" "+group.Status.String())
	if exp := renderExpiryLine(group.ExpiresAt); exp != "" {
		lines = append(lines, exp)
	}

	// Linked Entra Roles
	lines = append(lines, "", detailLabelStyle.Render("Linked Entra Roles:"))
	if len(group.LinkedRoles) > 0 {
		for _, lr := range group.LinkedRoles {
			lines = append(lines, detailDimStyle.Render("  "+statusIcon(lr.Status)+" "+lr.DisplayName))
		}
	} else {
		lines = append(lines, detailDimStyle.Italic(true).Render("  (none)"))
	}

	// Linked Azure RBAC Roles
	lines = append(lines, "", detailLabelStyle.Render("Linked Azure RBAC:"))
	if len(group.LinkedAzureRBac) > 0 {
		for _, ar := range group.LinkedAzureRBac {
			scopeShort := ar.Scope
			if len(scopeShort) > 30 {
				scopeShort = "..." + scopeShort[len(scopeShort)-27:]
			}
			lines = append(lines, detailDimStyle.Render("  • "+ar.DisplayName))
			lines = append(lines, detailDimStyle.Render("    "+scopeShort))
		}
	} else {
		lines = append(lines, detailDimStyle.Italic(true).Render("  (none)"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderRolesList(height int) string {
	return m.renderItemList(height, "roles", len(m.roles), func(i int) (string, azure.ActivationStatus, bool, bool) {
		role := m.roles[i]
		return role.DisplayName, role.Status, m.selectedRoles[i], i == m.rolesCursor && m.activeTab == TabRoles
	})
}

func (m Model) listPanelWidth() int {
	return (m.width - 8) * 9 / 20
}

func renderCheckbox(selected bool) string {
	if selected {
		return activeStyle.Render(checkboxChecked)
	}
	return dimStyle.Render(checkboxUnchecked)
}

func (m Model) renderListItem(idx int, name string, status azure.ActivationStatus, selected, isCursor bool) string {
	// Format: "[x] ● Name" - Prefix: checkbox(3) + space(1) + icon(1) + space(1) = 6
	nameWidth := max(m.listPanelWidth()-6, 10)

	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	line := fmt.Sprintf("%s %s %s", renderCheckbox(selected), statusIcon(status), name)

	if isCursor {
		return cursorStyle.Render(line)
	}
	return line
}

func (m Model) renderGroupsList(height int) string {
	return m.renderItemList(height, "groups", len(m.groups), func(i int) (string, azure.ActivationStatus, bool, bool) {
		group := m.groups[i]
		return group.DisplayName, group.Status, m.selectedGroups[i], i == m.groupsCursor && m.activeTab == TabGroups
	})
}

func (m Model) renderLighthouseView() string {
	// Match the total width of two side-by-side panels in main view
	// Two panels use totalWidth = m.width - 8, each adding border(2) + padding(2)
	// For single panel to match: use same totalWidth but only one set of border/padding
	panelWidth := max(m.width-8+4, 20)
	panelHeight := max(m.height-25, 5)
	content := m.renderLighthouseList(max(panelHeight-2, 1))
	panel := activePanelStyle.Width(panelWidth).Height(panelHeight).Render(
		panelTitleStyle.Foreground(colorHighlight).Render("● Lighthouse Subscriptions") + "\n" + content,
	)

	return panel
}

func (m Model) renderLighthouseList(height int) string {
	if len(m.lighthouse) == 0 {
		return detailDimStyle.Render("  No lighthouse subscriptions found")
	}

	var lines []string
	for i, sub := range m.lighthouse {
		if i >= height {
			break
		}
		lines = append(lines, m.renderLighthouseItem(i, sub))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderLighthouseItem(idx int, sub azure.LighthouseSubscription) string {
	groupInfo := ""
	if sub.LinkedGroupName != "" {
		groupInfo = detailLabelStyle.Render(fmt.Sprintf(" via: %s", sub.LinkedGroupName))
	}

	line := fmt.Sprintf("%s %s %s%s", renderCheckbox(m.selectedLight[idx]), statusIcon(sub.Status), truncate(sub.DisplayName, 30), groupInfo)

	if idx == m.lightCursor {
		return cursorStyle.Padding(0, 1).Render(line)
	}
	return itemStyle.Render(line)
}

func (m Model) renderLogs() string {
	logHeight := 8
	// Match the width of two side-by-side panels in header/main view
	// Two panels: totalWidth (m.width - 8) + 2*border(2) + 2*padding(2) = m.width
	// Single panel needs same visual width, so: content + border(2) + padding(2) = m.width
	// But there seems to be extra space, so use m.width - 6 to account for outer margin
	width := m.width - 6

	var lines []string

	// Get last N logs
	start := len(m.logs) - logHeight
	if start < 0 {
		start = 0
	}

	for i := start; i < len(m.logs); i++ {
		entry := m.logs[i]

		var levelStyle, msgStyle lipgloss.Style
		switch entry.Level {
		case LogDebug:
			levelStyle, msgStyle = dimStyle, logDebugStyle
		case LogError:
			levelStyle, msgStyle = errorBoldStyle, logErrorStyle
		default:
			levelStyle, msgStyle = detailLabelStyle, logInfoStyle
		}

		timeStr := dimStyle.Render(entry.Time.Format("15:04:05"))
		levelStr := levelStyle.Render(fmt.Sprintf("[%s]", entry.Level.String()))

		// Calculate available width for message
		msgWidth := max(width-lipgloss.Width(levelStr)-lipgloss.Width(timeStr)-2, 20)

		// Truncate message before styling to avoid cutting ANSI codes
		msg := entry.Message
		if len(msg) > msgWidth {
			msg = msg[:msgWidth-3] + "..."
		}

		line := fmt.Sprintf("%s %s %s", levelStr, timeStr, msgStyle.Render(msg))
		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for len(lines) < logHeight {
		lines = append([]string{""}, lines...)
	}

	return logPanelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) renderStatusBar() string {
	durationStr := highlightBoldStyle.Render(fmt.Sprintf("⏱ %dh", int(m.duration.Hours())))

	var autoStr string
	if m.autoRefresh {
		if secs, ok := m.refreshCountdown(); ok {
			autoStr = activeStyle.Render(fmt.Sprintf("↻ %ds", secs))
		} else {
			autoStr = activeStyle.Render("↻ ON")
		}
	} else {
		autoStr = dimStyle.Render("↻ OFF")
	}

	eligibleStr := detailLabelStyle.Render(fmt.Sprintf("📋 %d roles, %d groups", len(m.roles), len(m.groups)))

	selected := len(m.selectedRoles) + len(m.selectedGroups)
	if m.viewMode == ViewLighthouse {
		selected = len(m.selectedLight)
	}
	var selectStr string
	if selected > 0 {
		selectStr = activeBoldStyle.Render(fmt.Sprintf("✓ %d selected", selected))
	} else {
		selectStr = dimStyle.Render("✓ 0 selected")
	}

	helpHints := dimStyle.Render("↑↓ navigate │ Tab/←→ switch tab │ Space select │ Enter activate │ L lighthouse │ r refresh │ ? help │ q quit")

	left := fmt.Sprintf("%s  │  %s  │  %s  │  %s", durationStr, autoStr, eligibleStr, selectStr)

	return helpStyle.Width(m.width - 2).Render(left + "\n" + helpHints)
}

func (m Model) renderHelp() string {
	// Build dynamic duration help based on config
	durationHelp := ""
	for i, preset := range m.config.DurationPresets {
		if i < 4 {
			durationHelp += fmt.Sprintf("    %d              Set duration to %d hour(s)\n", i+1, preset)
		}
	}

	helpContent := fmt.Sprintf(`
Key Bindings:

  Navigation
    ↑/k, ↓/j       Move cursor up/down
    ←/h, →/l       Switch tabs (Roles / Groups)
    Tab            Switch tabs
    L              Toggle Lighthouse mode (delegated subs)

  Selection & Search
    Space          Select/deselect item under cursor
    /              Search/filter roles and groups
    Esc            Clear search filter

  Actions
    Enter          Activate selected items
    x / Del / BS   Deactivate selected active items
    r / F5         Refresh data from Azure

  Duration (for activation) - Current: %dh
%s
  Display & Settings
    v              Cycle log level (error → info → debug)
    c              Copy logs to clipboard
    e              Export activation history to clipboard
    a              Toggle auto-refresh (60s interval)
    ?              Show/hide this help
    Esc            Cancel current action / close dialogs
    q / Ctrl+C     Quit application

Status Icons:
    ●  Active       ◐  Expiring soon (< 30 min)
    ○  Inactive     ◌  Pending approval
`, int(m.duration.Hours()), durationHelp)

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("Help") + helpContent,
	)
}

func (m Model) renderConfirm() string {
	countStr := highlightBoldStyle.Render(fmt.Sprintf("%d", len(m.pendingActivations)))

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("Confirm Activation") + "\n\n" +
			fmt.Sprintf("Activate %s item(s) for %s?\n%s\n\n", countStr, m.durationStr(), dimStyle.Render("(1-4/Tab to change duration)")) +
			activeStyle.Render("(y)es") + " to continue or " + errorBoldStyle.Render("(n)o") + " to cancel",
	)
}

func (m Model) renderJustification() string {
	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("Justification Required") + "\n" +
			fmt.Sprintf("Duration: %s %s\n\n", m.durationStr(), dimStyle.Render("(1-4/Tab to change)")) +
			m.justificationInput.View() + "\n\n" +
			dimStyle.Render("Press Enter to confirm or Esc to cancel"),
	)
}

func (m Model) renderActivating() string {
	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("Activating...") + "\n\n" +
			spinner(colorActive) + " Please wait while activations are processed...",
	)
}

func (m Model) renderConfirmDeactivate() string {
	countStr := errorBoldStyle.Render(fmt.Sprintf("%d", len(m.pendingDeactivations)))

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorError).Render("Confirm Deactivation") + "\n\n" +
			fmt.Sprintf("Deactivate %s active item(s)?\n\n", countStr) +
			activeStyle.Render("(y)es") + " to continue or " + errorBoldStyle.Render("(n)o") + " to cancel",
	)
}

func (m Model) renderDeactivating() string {
	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorError).Render("Deactivating...") + "\n\n" +
			spinner(colorError) + " Please wait while deactivations are processed...",
	)
}

func (m Model) renderSearch() string {
	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("Search / Filter") + "\n\n" +
			m.searchInput.View() + "\n\n" +
			dimStyle.Render("Press Enter to apply or Esc to cancel"),
	)
}

// Helper functions

func (m Model) renderItemList(height int, itemType string, count int, getItem func(int) (name string, status azure.ActivationStatus, selected, isCursor bool)) string {
	if count == 0 {
		return detailDimStyle.Render(fmt.Sprintf("  No eligible %s found", itemType))
	}

	var lines []string
	for i := 0; i < count; i++ {
		if len(lines) >= height {
			break
		}
		name, status, selected, isCursor := getItem(i)
		if m.searchActive && !strings.Contains(strings.ToLower(name), strings.ToLower(m.searchQuery)) {
			continue
		}
		lines = append(lines, m.renderListItem(i, name, status, selected, isCursor))
	}

	if len(lines) == 0 && m.searchActive {
		return detailDimStyle.Render(fmt.Sprintf("  No %s match filter", itemType))
	}

	return strings.Join(lines, "\n")
}

func spinner(color lipgloss.Color) string {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := int(time.Now().UnixMilli()/100) % len(chars)
	return lipgloss.NewStyle().Foreground(color).Render(chars[idx])
}

func (m Model) countActiveItems() (roles, groups int) {
	for _, r := range m.roles {
		if r.Status.IsActive() {
			roles++
		}
	}
	for _, g := range m.groups {
		if g.Status.IsActive() {
			groups++
		}
	}
	return
}

func (m Model) refreshCountdown() (remaining int, hasCountdown bool) {
	if m.lastRefresh.IsZero() {
		return 0, false
	}
	elapsed := time.Since(m.lastRefresh)
	rem := time.Duration(m.config.AutoRefreshInterval)*time.Second - elapsed
	if rem > 0 {
		return int(rem.Seconds()), true
	}
	return 0, false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func (m Model) dialogWidth() int {
	return m.width - 10
}

func (m Model) durationStr() string {
	return activeBoldStyle.Render(fmt.Sprintf("%d hours", int(m.duration.Hours())))
}

func renderExpiryLine(expiresAt *time.Time) string {
	if expiresAt == nil {
		return ""
	}
	if remaining := time.Until(*expiresAt); remaining > 0 {
		return detailLabelStyle.Render("Expires: ") + detailValueStyle.Render(formatDuration(remaining))
	}
	return ""
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
