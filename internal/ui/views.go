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
	logoStyle := lipgloss.NewStyle().
		Foreground(colorHighlight).
		Bold(true)

	versionStyle := lipgloss.NewStyle().
		Foreground(colorDim).
		MarginTop(1)

	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := int(time.Now().UnixMilli()/100) % len(spinnerChars)
	spinner := lipgloss.NewStyle().Foreground(colorActive).Render(spinnerChars[spinnerIdx])

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		MarginTop(2)

	// Show tenant info if available
	tenantInfo := ""
	if m.tenant != nil {
		tenantInfo = lipgloss.NewStyle().
			Foreground(colorActive).
			MarginTop(1).
			Render(fmt.Sprintf("Connected to: %s", m.tenant.DisplayName))
	}

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
			icon = "✓"
			style = lipgloss.NewStyle().Foreground(colorActive)
		} else if step.active {
			icon = spinner
			style = lipgloss.NewStyle().Foreground(colorHighlight)
		} else {
			icon = "○"
			style = lipgloss.NewStyle().Foreground(colorDim)
		}
		stepLines = append(stepLines, style.Render(fmt.Sprintf("  %s %s", icon, step.name)))
	}

	stepsContent := lipgloss.NewStyle().
		MarginTop(2).
		Render(strings.Join(stepLines, "\n"))

	contentParts := []string{
		logoStyle.Render(asciiLogo),
		versionStyle.Render(fmt.Sprintf("v%s", m.version)),
	}

	if tenantInfo != "" {
		contentParts = append(contentParts, tenantInfo)
	}

	contentParts = append(contentParts,
		messageStyle.Render(spinner+" "+m.loadingMessage),
		stepsContent,
	)

	content := lipgloss.JoinVertical(lipgloss.Center, contentParts...)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m Model) renderError() string {
	logoStyle := lipgloss.NewStyle().
		Foreground(colorError).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(colorError).
		MarginTop(2).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(colorDim).
		MarginTop(2)

	content := lipgloss.JoinVertical(lipgloss.Center,
		logoStyle.Render(asciiLogo),
		errorStyle.Render("Authentication Failed"),
		messageStyle.Render(m.err.Error()),
		helpStyle.Render("Press 'r' to retry or 'q' to quit"),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
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

	// Right box: Info
	labelStyle := lipgloss.NewStyle().Foreground(colorDim)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	activeValueStyle := lipgloss.NewStyle().Foreground(colorActive).Bold(true)

	var infoLines []string

	// Tenant Name
	if m.tenant != nil {
		infoLines = append(infoLines, labelStyle.Render("Tenant:  ")+activeValueStyle.Render(m.tenant.DisplayName))
	} else {
		infoLines = append(infoLines, labelStyle.Render("Tenant:  ")+valueStyle.Render("Connecting..."))
	}

	// Tenant ID
	if m.tenant != nil {
		infoLines = append(infoLines, labelStyle.Render("ID:      ")+valueStyle.Render(m.tenant.ID))
	} else {
		infoLines = append(infoLines, labelStyle.Render("ID:      ")+valueStyle.Render("-"))
	}

	// User Principal Name
	if m.userEmail != "" {
		infoLines = append(infoLines, labelStyle.Render("User:    ")+valueStyle.Render(m.userEmail))
	} else if m.userDisplayName != "" {
		infoLines = append(infoLines, labelStyle.Render("User:    ")+valueStyle.Render(m.userDisplayName))
	} else {
		infoLines = append(infoLines, labelStyle.Render("User:    ")+valueStyle.Render("-"))
	}

	// Active Roles count
	activeRoles := 0
	activeGroups := 0
	for _, r := range m.roles {
		if r.Status == azure.StatusActive || r.Status == azure.StatusExpiringSoon {
			activeRoles++
		}
	}
	for _, g := range m.groups {
		if g.Status == azure.StatusActive || g.Status == azure.StatusExpiringSoon {
			activeGroups++
		}
	}
	activeTotal := activeRoles + activeGroups
	if activeTotal > 0 {
		infoLines = append(infoLines, labelStyle.Render("Active:  ")+activeValueStyle.Render(fmt.Sprintf("%d roles, %d groups", activeRoles, activeGroups)))
	} else {
		infoLines = append(infoLines, labelStyle.Render("Active:  ")+valueStyle.Render("0"))
	}

	// Refresh state
	var refreshStr string
	if m.autoRefresh {
		if !m.lastRefresh.IsZero() {
			elapsed := time.Since(m.lastRefresh)
			remaining := time.Duration(m.config.AutoRefreshInterval)*time.Second - elapsed
			if remaining > 0 {
				refreshStr = activeValueStyle.Render(fmt.Sprintf("Auto (%ds)", int(remaining.Seconds())))
			} else {
				refreshStr = activeValueStyle.Render("Auto (ON)")
			}
		} else {
			refreshStr = activeValueStyle.Render("Auto (ON)")
		}
	} else {
		if !m.lastRefresh.IsZero() {
			elapsed := time.Since(m.lastRefresh)
			if elapsed < time.Minute {
				refreshStr = valueStyle.Render("just now")
			} else {
				refreshStr = valueStyle.Render(fmt.Sprintf("%dm ago", int(elapsed.Minutes())))
			}
		} else {
			refreshStr = valueStyle.Render("-")
		}
	}
	infoLines = append(infoLines, labelStyle.Render("Refresh: ")+refreshStr)

	// Version
	infoLines = append(infoLines, labelStyle.Render("Version: ")+valueStyle.Render(fmt.Sprintf("v%s", m.version)))

	// Mode/Search indicators (if any)
	if m.viewMode == ViewLighthouse {
		infoLines = append(infoLines, lipgloss.NewStyle().Foreground(colorHighlight).Bold(true).Render("[LIGHTHOUSE]"))
	} else if m.searchActive {
		infoLines = append(infoLines, lipgloss.NewStyle().Foreground(colorPending).Bold(true).Render(fmt.Sprintf("[SEARCH: %s]", m.searchQuery)))
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
	hint := lipgloss.NewStyle().Foreground(colorDim).Render("  (Tab/←→ to switch)")

	tabBarContent := tabs + hint

	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(tabBarContent)
}

func (m Model) renderRoleDetail() string {
	if len(m.roles) == 0 || m.rolesCursor >= len(m.roles) {
		return lipgloss.NewStyle().Foreground(colorDim).Render("No role selected")
	}

	role := m.roles[m.rolesCursor]

	var lines []string

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
	lines = append(lines, titleStyle.Render("Role Details"))
	lines = append(lines, "")

	// Role name
	labelStyle := lipgloss.NewStyle().Foreground(colorPending).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	lines = append(lines, labelStyle.Render("Name: ")+valueStyle.Render(role.DisplayName))

	// Status
	statusStr := statusIcon(role.Status) + " " + role.Status.String()
	lines = append(lines, labelStyle.Render("Status: ")+statusStr)

	// Expiration
	if role.ExpiresAt != nil {
		remaining := time.Until(*role.ExpiresAt)
		if remaining > 0 {
			lines = append(lines, labelStyle.Render("Expires: ")+valueStyle.Render(formatDuration(remaining)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Permissions:"))

	// Show permissions from static data or API
	permissions := GetRolePermissions(role.RoleDefinitionID)
	if len(permissions) > 0 {
		permStyle := lipgloss.NewStyle().Foreground(colorDim)
		for _, perm := range permissions {
			lines = append(lines, permStyle.Render("  • "+perm))
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Italic(true).Render("  (permissions not available)"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderGroupDetail() string {
	if len(m.groups) == 0 || m.groupsCursor >= len(m.groups) {
		return lipgloss.NewStyle().Foreground(colorDim).Render("No group selected")
	}

	group := m.groups[m.groupsCursor]

	var lines []string

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
	lines = append(lines, titleStyle.Render("Group Details"))
	lines = append(lines, "")

	// Group name
	labelStyle := lipgloss.NewStyle().Foreground(colorPending).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	lines = append(lines, labelStyle.Render("Name: ")+valueStyle.Render(group.DisplayName))

	// Description
	if group.Description != "" {
		lines = append(lines, labelStyle.Render("Type: ")+valueStyle.Render(group.Description))
	}

	// Status
	statusStr := statusIcon(group.Status) + " " + group.Status.String()
	lines = append(lines, labelStyle.Render("Status: ")+statusStr)

	// Expiration
	if group.ExpiresAt != nil {
		remaining := time.Until(*group.ExpiresAt)
		if remaining > 0 {
			lines = append(lines, labelStyle.Render("Expires: ")+valueStyle.Render(formatDuration(remaining)))
		}
	}

	// Linked Entra Roles
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Linked Entra Roles:"))
	if len(group.LinkedRoles) > 0 {
		roleStyle := lipgloss.NewStyle().Foreground(colorDim)
		for _, lr := range group.LinkedRoles {
			icon := statusIcon(lr.Status)
			lines = append(lines, roleStyle.Render("  "+icon+" "+lr.DisplayName))
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Italic(true).Render("  (none)"))
	}

	// Linked Azure RBAC Roles
	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Linked Azure RBAC:"))
	if len(group.LinkedAzureRBac) > 0 {
		rbacStyle := lipgloss.NewStyle().Foreground(colorDim)
		for _, ar := range group.LinkedAzureRBac {
			scopeShort := ar.Scope
			if len(scopeShort) > 30 {
				scopeShort = "..." + scopeShort[len(scopeShort)-27:]
			}
			lines = append(lines, rbacStyle.Render("  • "+ar.DisplayName))
			lines = append(lines, rbacStyle.Render("    "+scopeShort))
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Italic(true).Render("  (none)"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderRolesList(height int) string {
	if len(m.roles) == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  No eligible roles found")
	}

	var lines []string
	for i, role := range m.roles {
		if len(lines) >= height {
			break
		}
		// Filter by search query
		if m.searchActive && !strings.Contains(strings.ToLower(role.DisplayName), strings.ToLower(m.searchQuery)) {
			continue
		}
		lines = append(lines, m.renderRoleItem(i, role))
	}

	if len(lines) == 0 && m.searchActive {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  No roles match filter")
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderRoleItem(idx int, role azure.Role) string {
	// Calculate available width for the item
	// Two panels side by side: each has border (2) + padding (2) = 4 chars
	totalWidth := m.width - 8
	listPanelWidth := totalWidth * 9 / 20 // 45% - same as logo box
	itemWidth := listPanelWidth           // content width (padding already accounted for)

	checkbox := lipgloss.NewStyle().Foreground(colorDim).Render(checkboxUnchecked)
	if m.selectedRoles[idx] {
		checkbox = lipgloss.NewStyle().Foreground(colorActive).Render(checkboxChecked)
	}

	icon := statusIcon(role.Status)

	// Calculate space for name (no time info in list anymore)
	// Format: "[x] ● Name"
	// Prefix: checkbox(3) + space(1) + icon(1) + space(1) = 6
	prefixWidth := 6
	nameWidth := itemWidth - prefixWidth
	if nameWidth < 10 {
		nameWidth = 10
	}

	// Truncate name if needed
	name := role.DisplayName
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	line := fmt.Sprintf("%s %s %s", checkbox, icon, name)

	if idx == m.rolesCursor && m.activeTab == TabRoles {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(colorHighlight).
			Bold(true).
			Render(line)
	}
	return lipgloss.NewStyle().Render(line)
}

func (m Model) renderGroupsList(height int) string {
	if len(m.groups) == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  No eligible groups found")
	}

	var lines []string
	for i, group := range m.groups {
		if len(lines) >= height {
			break
		}
		// Filter by search query
		if m.searchActive && !strings.Contains(strings.ToLower(group.DisplayName), strings.ToLower(m.searchQuery)) {
			continue
		}
		lines = append(lines, m.renderGroupItem(i, group))
	}

	if len(lines) == 0 && m.searchActive {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  No groups match filter")
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderGroupItem(idx int, group azure.Group) string {
	// Calculate available width for the item
	// Two panels side by side: each has border (2) + padding (2) = 4 chars
	totalWidth := m.width - 8
	listPanelWidth := totalWidth * 9 / 20 // 45% - same as logo box
	itemWidth := listPanelWidth           // content width (padding already accounted for)

	checkbox := lipgloss.NewStyle().Foreground(colorDim).Render(checkboxUnchecked)
	if m.selectedGroups[idx] {
		checkbox = lipgloss.NewStyle().Foreground(colorActive).Render(checkboxChecked)
	}

	icon := statusIcon(group.Status)

	// Calculate space for name (no time info in list anymore)
	// Format: "[x] ● Name"
	// Prefix: checkbox(3) + space(1) + icon(1) + space(1) = 6
	prefixWidth := 6
	nameWidth := itemWidth - prefixWidth
	if nameWidth < 10 {
		nameWidth = 10
	}

	// Truncate name if needed
	name := group.DisplayName
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	line := fmt.Sprintf("%s %s %s", checkbox, icon, name)

	if idx == m.groupsCursor && m.activeTab == TabGroups {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(colorHighlight).
			Bold(true).
			Render(line)
	}
	return lipgloss.NewStyle().Render(line)
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
		return lipgloss.NewStyle().Foreground(colorDim).Render("  No lighthouse subscriptions found")
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
	checkbox := lipgloss.NewStyle().Foreground(colorDim).Render(checkboxUnchecked)
	if m.selectedLight[idx] {
		checkbox = lipgloss.NewStyle().Foreground(colorActive).Render(checkboxChecked)
	}

	icon := statusIcon(sub.Status)
	name := sub.DisplayName

	groupInfo := ""
	if sub.LinkedGroupName != "" {
		groupInfo = lipgloss.NewStyle().Foreground(colorPending).Render(fmt.Sprintf(" via: %s", sub.LinkedGroupName))
	}

	line := fmt.Sprintf("%s %s %s%s", checkbox, icon, truncate(name, 30), groupInfo)

	if idx == m.lightCursor {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(colorHighlight).
			Bold(true).
			Padding(0, 1).
			Render(line)
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

		var levelStyle lipgloss.Style
		var msgStyle lipgloss.Style
		switch entry.Level {
		case LogDebug:
			levelStyle = lipgloss.NewStyle().Foreground(colorDim)
			msgStyle = logDebugStyle
		case LogError:
			levelStyle = lipgloss.NewStyle().Foreground(colorError).Bold(true)
			msgStyle = logErrorStyle
		default:
			levelStyle = lipgloss.NewStyle().Foreground(colorPending)
			msgStyle = logInfoStyle
		}

		timeStr := lipgloss.NewStyle().Foreground(colorDim).Render(entry.Time.Format("15:04:05"))
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
	// Duration indicator
	durationStyle := lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)
	durationStr := durationStyle.Render(fmt.Sprintf("⏱ %dh", int(m.duration.Hours())))

	// Auto-refresh indicator with countdown
	var autoStr string
	if m.autoRefresh {
		if !m.lastRefresh.IsZero() {
			elapsed := time.Since(m.lastRefresh)
			remaining := time.Duration(m.config.AutoRefreshInterval)*time.Second - elapsed
			if remaining > 0 {
				autoStr = lipgloss.NewStyle().Foreground(colorActive).Render(fmt.Sprintf("↻ %ds", int(remaining.Seconds())))
			} else {
				autoStr = lipgloss.NewStyle().Foreground(colorActive).Render("↻ ON")
			}
		} else {
			autoStr = lipgloss.NewStyle().Foreground(colorActive).Render("↻ ON")
		}
	} else {
		autoStr = lipgloss.NewStyle().Foreground(colorDim).Render("↻ OFF")
	}

	// Eligible count
	eligibleStr := lipgloss.NewStyle().Foreground(colorPending).
		Render(fmt.Sprintf("📋 %d roles, %d groups", len(m.roles), len(m.groups)))

	// Selection count
	selected := len(m.selectedRoles) + len(m.selectedGroups)
	if m.viewMode == ViewLighthouse {
		selected = len(m.selectedLight)
	}
	var selectStr string
	if selected > 0 {
		selectStr = lipgloss.NewStyle().Foreground(colorActive).Bold(true).Render(fmt.Sprintf("✓ %d selected", selected))
	} else {
		selectStr = lipgloss.NewStyle().Foreground(colorDim).Render("✓ 0 selected")
	}

	// Help hints
	helpHints := lipgloss.NewStyle().Foreground(colorDim).Render(
		"↑↓ navigate │ Tab/←→ switch tab │ Space select │ Enter activate │ L lighthouse │ r refresh │ ? help │ q quit",
	)

	left := fmt.Sprintf("%s  │  %s  │  %s  │  %s", durationStr, autoStr, eligibleStr, selectStr)

	return helpStyle.Width(m.width - 2).Render(
		left + "\n" + helpHints,
	)
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
	count := len(m.pendingActivations)
	countStr := lipgloss.NewStyle().Foreground(colorHighlight).Bold(true).Render(fmt.Sprintf("%d", count))
	durationStr := lipgloss.NewStyle().Foreground(colorActive).Bold(true).Render(fmt.Sprintf("%d hours", int(m.duration.Hours())))

	durationHint := lipgloss.NewStyle().Foreground(colorDim).Render("(1-4/Tab to change duration)")

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Confirm Activation") + "\n\n" +
			fmt.Sprintf("Activate %s item(s) for %s?\n%s\n\n", countStr, durationStr, durationHint) +
			lipgloss.NewStyle().Foreground(colorActive).Render("(y)es") + " to continue or " +
			lipgloss.NewStyle().Foreground(colorError).Render("(n)o") + " to cancel",
	)
}

func (m Model) renderJustification() string {
	durationStr := lipgloss.NewStyle().Foreground(colorActive).Bold(true).Render(fmt.Sprintf("%d hours", int(m.duration.Hours())))
	durationHint := lipgloss.NewStyle().Foreground(colorDim).Render("(1-4/Tab to change)")

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Justification Required") + "\n" +
			fmt.Sprintf("Duration: %s %s\n\n", durationStr, durationHint) +
			m.justificationInput.View() + "\n\n" +
			lipgloss.NewStyle().Foreground(colorDim).Render("Press Enter to confirm or Esc to cancel"),
	)
}

func (m Model) renderActivating() string {
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := int(time.Now().UnixMilli()/100) % len(spinnerChars)
	spinner := lipgloss.NewStyle().Foreground(colorActive).Render(spinnerChars[spinnerIdx])

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Activating...") + "\n\n" +
			spinner + " Please wait while activations are processed...",
	)
}

func (m Model) renderConfirmDeactivate() string {
	count := len(m.pendingDeactivations)
	countStr := lipgloss.NewStyle().Foreground(colorError).Bold(true).Render(fmt.Sprintf("%d", count))

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorError).Render("Confirm Deactivation") + "\n\n" +
			fmt.Sprintf("Deactivate %s active item(s)?\n\n", countStr) +
			lipgloss.NewStyle().Foreground(colorActive).Render("(y)es") + " to continue or " +
			lipgloss.NewStyle().Foreground(colorError).Render("(n)o") + " to cancel",
	)
}

func (m Model) renderDeactivating() string {
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := int(time.Now().UnixMilli()/100) % len(spinnerChars)
	spinner := lipgloss.NewStyle().Foreground(colorError).Render(spinnerChars[spinnerIdx])

	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorError).Render("Deactivating...") + "\n\n" +
			spinner + " Please wait while deactivations are processed...",
	)
}

func (m Model) renderSearch() string {
	return confirmStyle.Width(m.width - 10).Render(
		titleStyle.Foreground(colorHighlight).Render("Search / Filter") + "\n\n" +
			m.searchInput.View() + "\n\n" +
			lipgloss.NewStyle().Foreground(colorDim).Render("Press Enter to apply or Esc to cancel"),
	)
}

// Helper functions

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
