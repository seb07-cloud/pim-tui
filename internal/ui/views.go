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
		// Left box: Full ASCII logo
		logoStyle := lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

		logoContent := logoStyle.Render(asciiLogo)

		logoBox := panelStyle.
			Width(logoBoxWidth).
			Height(7).
			Align(lipgloss.Center, lipgloss.Center).
			Render(logoContent)

		return lipgloss.JoinHorizontal(lipgloss.Top, logoBox, infoBox)
	}

	return infoBox
}

func (m Model) renderMainView() string {
	// Tab bar
	tabBar := m.renderTabBar()

	// Content area
	// Header: 9 lines (7 content + 2 border), Tab bar: 3 lines, Logs: 10 lines, Status: 3 lines
	// Each panel has border (2) + padding (2) = 4 chars, and we have 2 panels side by side
	totalWidth := m.width - 8
	listPanelWidth := totalWidth * 9 / 20          // List panel (45%) - same as logo box in header
	detailPanelWidth := totalWidth - listPanelWidth // Detail panel (55%)
	panelHeight := m.height - 25

	var listPanel, detailPanel string

	if m.activeTab == TabRoles {
		// Roles list panel (narrower)
		rolesContent := m.renderRolesList(panelHeight - 2)
		listPanel = activePanelStyle.Width(listPanelWidth).Height(panelHeight).Render(
			panelTitleStyle.Foreground(colorHighlight).Render("● PIM Roles") + "\n" + rolesContent,
		)
		// Role details panel (wider)
		detailPanel = panelStyle.Width(detailPanelWidth).Height(panelHeight).Render(
			m.renderRoleDetail(),
		)
	} else {
		// Groups list panel (narrower)
		groupsContent := m.renderGroupsList(panelHeight - 2)
		listPanel = activePanelStyle.Width(listPanelWidth).Height(panelHeight).Render(
			panelTitleStyle.Foreground(colorHighlight).Render("● PIM Groups") + "\n" + groupsContent,
		)
		// Group details panel (wider)
		detailPanel = panelStyle.Width(detailPanelWidth).Height(panelHeight).Render(
			m.renderGroupDetail(),
		)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}

func (m Model) renderTabBar() string {
	width := m.width - 4

	// Tab styles
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorHighlight).
		Background(lipgloss.Color("#333333")).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		BorderForeground(colorHighlight)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(colorDim).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		BorderForeground(colorBorder)

	var rolesTab, groupsTab string
	if m.activeTab == TabRoles {
		rolesTab = activeTabStyle.Render("Roles")
		groupsTab = inactiveTabStyle.Render("Groups")
	} else {
		rolesTab = inactiveTabStyle.Render("Roles")
		groupsTab = activeTabStyle.Render("Groups")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, rolesTab, " ", groupsTab)

	// Add hint for tab switching
	hint := dimStyle.Render("  (Tab/←→ to switch)")

	tabBarContent := tabs + hint

	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(tabBarContent)
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

	if role.ExpiresAt != nil {
		if remaining := time.Until(*role.ExpiresAt); remaining > 0 {
			lines = append(lines, detailLabelStyle.Render("Expires: ")+detailValueStyle.Render(formatDuration(remaining)))
		}
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

	if group.ExpiresAt != nil {
		if remaining := time.Until(*group.ExpiresAt); remaining > 0 {
			lines = append(lines, detailLabelStyle.Render("Expires: ")+detailValueStyle.Render(formatDuration(remaining)))
		}
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
	if len(m.roles) == 0 {
		return detailDimStyle.Render("  No eligible roles found")
	}

	var lines []string
	for i, role := range m.roles {
		if len(lines) >= height {
			break
		}
		if m.searchActive && !strings.Contains(strings.ToLower(role.DisplayName), strings.ToLower(m.searchQuery)) {
			continue
		}
		lines = append(lines, m.renderRoleItem(i, role))
	}

	if len(lines) == 0 && m.searchActive {
		return detailDimStyle.Render("  No roles match filter")
	}

	return strings.Join(lines, "\n")
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
	itemWidth := m.listPanelWidth()

	// Format: "[x] ● Name" - Prefix: checkbox(3) + space(1) + icon(1) + space(1) = 6
	nameWidth := itemWidth - 6
	if nameWidth < 10 {
		nameWidth = 10
	}

	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	line := fmt.Sprintf("%s %s %s", renderCheckbox(selected), statusIcon(status), name)

	if isCursor {
		return cursorStyle.Render(line)
	}
	return line
}

func (m Model) renderRoleItem(idx int, role azure.Role) string {
	return m.renderListItem(idx, role.DisplayName, role.Status, m.selectedRoles[idx], idx == m.rolesCursor && m.activeTab == TabRoles)
}

func (m Model) renderGroupsList(height int) string {
	if len(m.groups) == 0 {
		return detailDimStyle.Render("  No eligible groups found")
	}

	var lines []string
	for i, group := range m.groups {
		if len(lines) >= height {
			break
		}
		if m.searchActive && !strings.Contains(strings.ToLower(group.DisplayName), strings.ToLower(m.searchQuery)) {
			continue
		}
		lines = append(lines, m.renderGroupItem(i, group))
	}

	if len(lines) == 0 && m.searchActive {
		return detailDimStyle.Render("  No groups match filter")
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderGroupItem(idx int, group azure.Group) string {
	return m.renderListItem(idx, group.DisplayName, group.Status, m.selectedGroups[idx], idx == m.groupsCursor && m.activeTab == TabGroups)
}

func (m Model) renderLighthouseView() string {
	// Match the total width of two side-by-side panels in main view
	// Two panels use totalWidth = m.width - 8, each adding border(2) + padding(2)
	// For single panel to match: use same totalWidth but only one set of border/padding
	totalWidth := m.width - 8
	panelWidth := totalWidth + 4 // Add back one panel's border/padding since we only have one panel
	panelHeight := m.height - 25 // Same as main view

	if panelHeight < 5 {
		panelHeight = 5
	}
	if panelWidth < 20 {
		panelWidth = 20
	}

	listHeight := panelHeight - 2
	if listHeight < 1 {
		listHeight = 1
	}

	content := m.renderLighthouseList(listHeight)
	panel := activePanelStyle.Width(panelWidth).Height(panelHeight).Render(
		panelTitleStyle.Foreground(colorHighlight).Render("● Lighthouse Subscriptions") + "\n" + content,
	)

	return panel
}

func (m Model) renderLighthouseList(height int) string {
	if len(m.lighthouse) == 0 {
		return detailDimStyle.Render("  No lighthouse subscriptions found")
	}

	if height < 1 {
		height = 1
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

		// Calculate available width for message (prefix takes ~20 chars: "[ERROR] 15:04:05 ")
		prefixWidth := lipgloss.Width(levelStr) + 1 + lipgloss.Width(timeStr) + 1
		msgWidth := width - prefixWidth
		if msgWidth < 20 {
			msgWidth = 20
		}

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

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Help") + helpContent,
	)
}

func (m Model) renderConfirm() string {
	countStr := highlightBoldStyle.Render(fmt.Sprintf("%d", len(m.pendingActivations)))
	durationStr := activeBoldStyle.Render(fmt.Sprintf("%d hours", int(m.duration.Hours())))

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Confirm Activation") + "\n\n" +
			fmt.Sprintf("Activate %s item(s) for %s?\n%s\n\n", countStr, durationStr, dimStyle.Render("(1-4/Tab to change duration)")) +
			activeStyle.Render("(y)es") + " to continue or " + errorBoldStyle.Render("(n)o") + " to cancel",
	)
}

func (m Model) renderJustification() string {
	durationStr := activeBoldStyle.Render(fmt.Sprintf("%d hours", int(m.duration.Hours())))

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Justification Required") + "\n" +
			fmt.Sprintf("Duration: %s %s\n\n", durationStr, dimStyle.Render("(1-4/Tab to change)")) +
			m.justificationInput.View() + "\n\n" +
			dimStyle.Render("Press Enter to confirm or Esc to cancel"),
	)
}

func (m Model) renderActivating() string {
	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Activating...") + "\n\n" +
			spinner(colorActive) + " Please wait while activations are processed...",
	)
}

func (m Model) renderConfirmDeactivate() string {
	countStr := errorBoldStyle.Render(fmt.Sprintf("%d", len(m.pendingDeactivations)))

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorError).Render("Confirm Deactivation") + "\n\n" +
			fmt.Sprintf("Deactivate %s active item(s)?\n\n", countStr) +
			activeStyle.Render("(y)es") + " to continue or " + errorBoldStyle.Render("(n)o") + " to cancel",
	)
}

func (m Model) renderDeactivating() string {
	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorError).Render("Deactivating...") + "\n\n" +
			spinner(colorError) + " Please wait while deactivations are processed...",
	)
}

func (m Model) renderSearch() string {
	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Search / Filter") + "\n\n" +
			m.searchInput.View() + "\n\n" +
			dimStyle.Render("Press Enter to apply or Esc to cancel"),
	)
}

// Helper functions

func spinner(color lipgloss.Color) string {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := int(time.Now().UnixMilli()/100) % len(chars)
	return lipgloss.NewStyle().Foreground(color).Render(chars[idx])
}

func (m Model) countActiveItems() (roles, groups int) {
	for _, r := range m.roles {
		if r.Status == azure.StatusActive || r.Status == azure.StatusExpiringSoon {
			roles++
		}
	}
	for _, g := range m.groups {
		if g.Status == azure.StatusActive || g.Status == azure.StatusExpiringSoon {
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
