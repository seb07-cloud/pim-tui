package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/seb07-cloud/pim-tui/internal/azure"
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

	// Full-screen states use lipgloss.Place internally to fill terminal
	// Prefix with ANSI clear screen to prevent ghost lines in some terminals
	const clearScreen = "\033[2J\033[H"

	if m.state == StateLoading {
		return clearScreen + m.renderLoading()
	}

	if m.state == StateUnauthenticated || m.state == StateAuthenticating {
		return clearScreen + m.renderUnauthenticated()
	}

	if m.state == StateError {
		return clearScreen + m.renderError()
	}

	var sections []string

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
		sections = append(sections, m.renderMainView())
	}

	// Log panel
	sections = append(sections, m.renderLogs())

	// Status bar
	sections = append(sections, m.renderStatusBar())

	// Join all sections and ensure output fills terminal to prevent ghost lines
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(content)
}

func (m Model) renderLoading() string {
	spin := spinner(colorActive)

	// Build step indicators - sequential display with parallel loading after step 2
	// Steps must complete in order for display, even if parallel loads finish out of order
	authDone := m.client != nil
	tenantDone := authDone && m.tenant != nil

	steps := []struct {
		name string
		done bool
	}{
		{"Authenticating with Graph API...", authDone},
		{"Loading Tenant Information", tenantDone},
		{"Loading PIM roles", tenantDone && m.rolesLoaded},
		{"Loading PIM groups", tenantDone && m.groupsLoaded},
		{"Loading Subscriptions", tenantDone && m.lighthouseLoaded},
	}

	// Determine active step: first incomplete step gets the spinner
	activeIdx := -1
	for i, step := range steps {
		if !step.done {
			activeIdx = i
			break
		}
	}

	// Count completed steps for progress bar
	completed := 0
	for _, step := range steps {
		if step.done {
			completed++
		}
	}

	totalSteps := len(steps)
	var stepLines []string
	for i, step := range steps {
		var icon string
		var style lipgloss.Style
		if step.done {
			icon, style = "✓", activeStyle
		} else if i == activeIdx {
			icon, style = spin, highlightBoldStyle
		} else {
			icon, style = "○", dimStyle
		}
		// Add step number for clarity
		stepLines = append(stepLines, style.Render(fmt.Sprintf("  %s [%d/%d] %s", icon, i+1, totalSteps, step.name)))
	}

	// Build overall progress bar
	progressWidth := 30
	progressBar := renderProgressBar(float64(completed), float64(len(steps)), progressWidth)
	progressPercent := (completed * 100) / len(steps)

	contentParts := []string{
		highlightBoldStyle.Render(asciiLogo),
		dimStyle.MarginTop(1).Render(fmt.Sprintf("v%s", m.version)),
	}

	if m.tenant != nil {
		contentParts = append(contentParts,
			activeStyle.MarginTop(1).Render(fmt.Sprintf("✓ Connected to: %s", m.tenant.DisplayName)))
	}

	contentParts = append(contentParts,
		detailValueStyle.MarginTop(2).Render(spin+" "+m.loadingMessage),
		lipgloss.NewStyle().MarginTop(1).Render(strings.Join(stepLines, "\n")),
		lipgloss.NewStyle().MarginTop(2).Render(
			fmt.Sprintf("%s %s",
				progressBar,
				dimStyle.Render(fmt.Sprintf("%d%%", progressPercent)),
			),
		),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, contentParts...))
}

func (m Model) renderError() string {
	// Build troubleshooting tips based on error type
	var tips string
	errStr := ""
	if m.err != nil {
		errStr = m.err.Error()
	}

	tips = dimStyle.Render("━━━ Troubleshooting Tips ━━━\n")
	if strings.Contains(errStr, "token") || strings.Contains(errStr, "credential") {
		tips += dimStyle.Render("  • Run 'az login' to refresh your Azure credentials\n")
		tips += dimStyle.Render("  • Check if your session has expired\n")
	} else if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		tips += dimStyle.Render("  • Check your internet connection\n")
		tips += dimStyle.Render("  • Verify VPN is connected if required\n")
	} else if strings.Contains(errStr, "permission") || strings.Contains(errStr, "403") {
		tips += dimStyle.Render("  • Verify you have PIM access in this tenant\n")
		tips += dimStyle.Render("  • Contact your administrator\n")
	} else {
		tips += dimStyle.Render("  • Run 'az login' to refresh credentials\n")
		tips += dimStyle.Render("  • Check your network connection\n")
		tips += dimStyle.Render("  • Verify Azure CLI is installed correctly\n")
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		errorBoldStyle.Render(asciiLogo),
		errorBoldStyle.MarginTop(2).Render("⚠ Authentication Failed"),
		"",
		detailLabelStyle.Render("Error: ")+detailValueStyle.Render(truncate(errStr, 60)),
		"",
		tips,
		"",
		activeStyle.Render(" [R] Retry ")+"  "+dimStyle.Render(" [Q] Quit "),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) renderUnauthenticated() string {
	var contentParts []string

	// Logo and version
	contentParts = append(contentParts,
		highlightBoldStyle.Render(asciiLogo),
		dimStyle.MarginTop(1).Render(fmt.Sprintf("v%s", m.version)),
	)

	// Title and instructions based on state
	if m.state == StateAuthenticating {
		spin := spinner(colorActive)
		contentParts = append(contentParts,
			highlightBoldStyle.MarginTop(2).Render("Authenticating..."),
			"",
			dimStyle.Render("Opening browser for authentication..."),
			detailValueStyle.Render(spin+" Waiting for browser sign-in..."),
			"",
			dimStyle.Render("Complete sign-in in your browser window."),
			"",
			dimStyle.Render("[Esc] Cancel")+"    "+dimStyle.Render("[Q] Quit"),
		)
	} else {
		contentParts = append(contentParts,
			highlightBoldStyle.MarginTop(2).Render("Authentication Required"),
			"",
			dimStyle.Render("No Azure CLI session found."),
			"",
			activeStyle.Render("[L] Login with Browser")+"    "+dimStyle.Render("[Q] Quit"),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, contentParts...)
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
		infoLines = append(infoLines, dimStyle.Render("Tenant: ")+activeBoldStyle.Render(m.tenant.DisplayName))
		infoLines = append(infoLines, dimStyle.Render("User:   ")+detailValueStyle.Render(truncate(m.userEmail, 35)))
	} else {
		infoLines = append(infoLines, dimStyle.Render("Tenant: ")+detailValueStyle.Render("Connecting..."))
		infoLines = append(infoLines, dimStyle.Render("User:   ")+detailValueStyle.Render("-"))
	}

	// Quick stats badges
	activeRoles, activeGroups := m.countActiveItems()
	expiringCount := m.countExpiringItems()

	// Build badge line
	var badges []string

	// Active badge
	if activeRoles+activeGroups > 0 {
		activeBadge := lipgloss.NewStyle().
			Background(colorActive).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("● %d ACTIVE", activeRoles+activeGroups))
		badges = append(badges, activeBadge)
	}

	// Expiring badge
	if expiringCount > 0 {
		expiringBadge := lipgloss.NewStyle().
			Background(colorExpiring).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1).
			Render(fmt.Sprintf("◐ %d EXPIRING", expiringCount))
		badges = append(badges, expiringBadge)
	}

	if len(badges) > 0 {
		infoLines = append(infoLines, strings.Join(badges, " "))
	}

	// Roles and Groups counts
	rolesBadge := lipgloss.NewStyle().
		Background(colorBorder).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1).
		Render(fmt.Sprintf("🔐 %d Roles", len(m.roles)))
	groupsBadge := lipgloss.NewStyle().
		Background(colorBorder).
		Foreground(lipgloss.Color("#ffffff")).
		Padding(0, 1).
		Render(fmt.Sprintf("👥 %d Groups", len(m.groups)))
	infoLines = append(infoLines, rolesBadge+" "+groupsBadge)

	// Add search indicator if active
	if m.searchActive {
		infoLines = append(infoLines, detailLabelStyle.Render(fmt.Sprintf("🔍 \"%s\"", m.searchQuery)))
	}

	// Refresh state
	var refreshStr string
	if m.autoRefresh {
		if secs, ok := m.refreshCountdown(); ok {
			refreshStr = activeBoldStyle.Render(fmt.Sprintf("↻ Auto (%ds)", secs))
		} else {
			refreshStr = activeBoldStyle.Render("↻ Auto (ON)")
		}
	} else if !m.lastRefresh.IsZero() {
		elapsed := time.Since(m.lastRefresh)
		if elapsed < time.Minute {
			refreshStr = detailValueStyle.Render("↻ just now")
		} else {
			refreshStr = dimStyle.Render(fmt.Sprintf("↻ %dm ago", int(elapsed.Minutes())))
		}
	} else {
		refreshStr = dimStyle.Render("↻ -")
	}
	infoLines = append(infoLines, refreshStr)
	infoLines = append(infoLines, dimStyle.Render(fmt.Sprintf("v%s", m.version)))

	infoContent := strings.Join(infoLines, "\n")

	if showLogo {
		// Logo has 6 lines - pad info content to match for consistent box heights
		logoLines := 6
		for len(infoLines) < logoLines {
			infoLines = append(infoLines, "")
		}
		infoContent = strings.Join(infoLines, "\n")

		infoBox := panelStyle.
			Width(infoBoxWidth).
			Render(infoContent)

		logoBox := panelStyle.
			Width(logoBoxWidth).
			Align(lipgloss.Center, lipgloss.Center).
			Render(highlightBoldStyle.Render(asciiLogo))
		return lipgloss.JoinHorizontal(lipgloss.Top, logoBox, infoBox)
	}

	infoBox := panelStyle.
		Width(infoBoxWidth).
		Render(infoContent)

	return infoBox
}

func (m Model) renderMainView() string {
	tabBar := m.renderTabBar()

	// Panel dimensions
	totalWidth := m.width - 8
	listPanelWidth := totalWidth * 9 / 20
	detailPanelWidth := totalWidth - listPanelWidth
	panelHeight := m.height - 25

	// Select content based on active tab with enhanced icons
	var title, listContent, detailContent string
	switch m.activeTab {
	case TabRoles:
		title = "🔐 PIM Roles"
		listContent = m.renderRolesList(panelHeight - 2)
		detailContent = m.renderRoleDetail()
	case TabGroups:
		title = "👥 PIM Groups"
		listContent = m.renderGroupsList(panelHeight - 2)
		detailContent = m.renderGroupDetail()
	case TabSubscriptions:
		title = "📑 Subscriptions"
		// Show inline search filter if active
		if m.searchActive && m.searchQuery != "" {
			title = fmt.Sprintf("📑 Subscriptions [🔍 %s]", m.searchQuery)
		}
		listContent = m.renderSubscriptionsList(max(panelHeight-2, 1))
		detailContent = m.renderSubscriptionDetail()
	}

	// Prominent panel title with background
	prominentTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(colorHighlight).
		Padding(0, 2).
		Render(title)

	listPanel := activePanelStyle.Width(listPanelWidth).Height(panelHeight).Render(
		prominentTitle + "\n" + listContent,
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

	// Count active items per tab for badges
	activeRoles := 0
	for _, r := range m.roles {
		if r.Status.IsActive() {
			activeRoles++
		}
	}
	activeGroups := 0
	for _, g := range m.groups {
		if g.Status.IsActive() {
			activeGroups++
		}
	}
	activeSubs := 0
	for _, s := range m.lighthouse {
		// Count subscription as active if any of its roles are active
		for _, role := range s.EligibleRoles {
			if role.Status.IsActive() {
				activeSubs++
				break
			}
		}
	}

	// Build tab labels with counts
	rolesLabel := fmt.Sprintf("🔐 Roles (%d)", len(m.roles))
	if activeRoles > 0 {
		rolesLabel = fmt.Sprintf("🔐 Roles (%d) %s", len(m.roles), activeStyle.Render(fmt.Sprintf("●%d", activeRoles)))
	}

	groupsLabel := fmt.Sprintf("👥 Groups (%d)", len(m.groups))
	if activeGroups > 0 {
		groupsLabel = fmt.Sprintf("👥 Groups (%d) %s", len(m.groups), activeStyle.Render(fmt.Sprintf("●%d", activeGroups)))
	}

	subsLabel := fmt.Sprintf("📑 Subs (%d)", len(m.lighthouse))
	if activeSubs > 0 {
		subsLabel = fmt.Sprintf("📑 Subs (%d) %s", len(m.lighthouse), activeStyle.Render(fmt.Sprintf("●%d", activeSubs)))
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom,
		tabStyle(m.activeTab == TabRoles).Render(rolesLabel), " ",
		tabStyle(m.activeTab == TabGroups).Render(groupsLabel), " ",
		tabStyle(m.activeTab == TabSubscriptions).Render(subsLabel),
	)

	// Add full-width underline indicator for active tab
	tabBarWidth := m.width - 6
	underline := highlightBoldStyle.Render(strings.Repeat("━", tabBarWidth))

	return lipgloss.NewStyle().Width(m.width-4).Padding(0, 1).Render(
		tabs + dimStyle.Render("  ←→/Tab") + "\n" + underline,
	)
}

func (m Model) renderRoleDetail() string {
	if len(m.roles) == 0 || m.rolesCursor >= len(m.roles) {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			"",
			dimStyle.Render("🔐"),
			"",
			dimStyle.Render("No role selected"),
			"",
			dimStyle.Render("Select a role from the list"),
			dimStyle.Render("to view its details"),
		)
	}

	role := m.roles[m.rolesCursor]
	var lines []string

	// Title with decorative line
	lines = append(lines, detailTitleStyle.Render("━━━ 🔐 Role Details ━━━"), "")
	lines = append(lines, detailLabelStyle.Render("Name: ")+detailValueStyle.Render(role.DisplayName))
	lines = append(lines, detailLabelStyle.Render("Status: ")+statusIcon(role.Status)+" "+role.Status.String())

	// Enhanced expiry display with progress bar
	if role.ExpiresAt != nil {
		remaining := time.Until(*role.ExpiresAt)
		if remaining > 0 {
			lines = append(lines, detailLabelStyle.Render("Expires: ")+detailValueStyle.Render(formatDuration(remaining)))
			// Show progress bar for active roles (assuming max 8h activation)
			maxDuration := 8 * time.Hour
			lines = append(lines, detailDimStyle.Render("         ")+renderProgressBar(remaining.Seconds(), maxDuration.Seconds(), 20))
		}
	}

	lines = append(lines, "", detailDimStyle.Render("─────────────────────────────"))
	lines = append(lines, detailLabelStyle.Render("Permissions:"))
	if permissions := GetRolePermissions(role.RoleDefinitionID); len(permissions) > 0 {
		maxWidth := 40 // Reasonable width for detail panel
		for _, perm := range permissions {
			wrapped := wrapPermission(perm, maxWidth)
			for i, line := range wrapped {
				if i == 0 {
					lines = append(lines, detailDimStyle.Render("  • "+line))
				} else {
					lines = append(lines, detailDimStyle.Render("    "+line)) // indent continuation
				}
			}
		}
	} else {
		lines = append(lines, detailDimStyle.Italic(true).Render("  (permissions not available)"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderGroupDetail() string {
	if len(m.groups) == 0 || m.groupsCursor >= len(m.groups) {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			"",
			dimStyle.Render("👥"),
			"",
			dimStyle.Render("No group selected"),
			"",
			dimStyle.Render("Select a group from the list"),
			dimStyle.Render("to view its details"),
		)
	}

	group := m.groups[m.groupsCursor]
	var lines []string

	// Title with decorative line
	lines = append(lines, detailTitleStyle.Render("━━━ 👥 Group Details ━━━"), "")
	lines = append(lines, detailLabelStyle.Render("Name: ")+detailValueStyle.Render(group.DisplayName))

	if group.Description != "" {
		lines = append(lines, detailLabelStyle.Render("Type: ")+detailValueStyle.Render(group.Description))
	}

	lines = append(lines, detailLabelStyle.Render("Status: ")+statusIcon(group.Status)+" "+group.Status.String())

	// Enhanced expiry display with progress bar
	if group.ExpiresAt != nil {
		remaining := time.Until(*group.ExpiresAt)
		if remaining > 0 {
			lines = append(lines, detailLabelStyle.Render("Expires: ")+detailValueStyle.Render(formatDuration(remaining)))
			maxDuration := 8 * time.Hour
			lines = append(lines, detailDimStyle.Render("         ")+renderProgressBar(remaining.Seconds(), maxDuration.Seconds(), 20))
		}
	}

	// Linked Entra Roles
	lines = append(lines, "", detailDimStyle.Render("─────────────────────────────"))
	lines = append(lines, detailLabelStyle.Render("Linked Entra Roles:"))
	if len(group.LinkedRoles) > 0 {
		for _, lr := range group.LinkedRoles {
			lines = append(lines, detailDimStyle.Render("  "+statusIcon(lr.Status)+" "+lr.DisplayName))
		}
	} else {
		lines = append(lines, detailDimStyle.Italic(true).Render("  (none)"))
	}

	// Linked Azure RBAC Roles
	lines = append(lines, "", detailDimStyle.Render("─────────────────────────────"))
	lines = append(lines, detailLabelStyle.Render("Linked Azure RBAC:"))
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
	return m.renderItemListWithExpiry(height, "roles", len(m.roles), m.rolesScrollOffset, func(i int) (string, azure.ActivationStatus, bool, bool, *time.Time) {
		role := m.roles[i]
		return role.DisplayName, role.Status, m.selectedRoles[i], i == m.rolesCursor && m.activeTab == TabRoles, role.ExpiresAt
	})
}

func (m Model) listPanelWidth() int {
	return (m.width - 8) * 9 / 20
}

func renderCheckbox(selected bool) string {
	if selected {
		return highlightBoldStyle.Render(checkboxChecked)
	}
	return dimStyle.Render(checkboxUnchecked)
}

func (m Model) renderListItem(idx int, name string, status azure.ActivationStatus, selected, isCursor bool) string {
	// Format: "[x] ● Name" - Prefix: checkbox(3) + space(1) + icon(1) + space(1) = 6
	nameWidth := max(m.listPanelWidth()-6, 10)

	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// Apply search highlighting if search is active
	displayName := name
	if m.searchActive && m.searchQuery != "" {
		displayName = highlightSearchMatch(name, m.searchQuery)
	}

	line := fmt.Sprintf("%s %s %s", renderCheckbox(selected), statusIcon(status), displayName)

	if isCursor {
		return cursorStyle.Render(line)
	}
	return line
}

func (m Model) renderGroupsList(height int) string {
	return m.renderItemListWithExpiry(height, "groups", len(m.groups), m.groupsScrollOffset, func(i int) (string, azure.ActivationStatus, bool, bool, *time.Time) {
		group := m.groups[i]
		return group.DisplayName, group.Status, m.selectedGroups[i], i == m.groupsCursor && m.activeTab == TabGroups, group.ExpiresAt
	})
}

func (m Model) renderSubscriptionsList(height int) string {
	if len(m.lighthouse) == 0 {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render("📑"),
			dimStyle.Render("No subscriptions found"),
			dimStyle.Render("Check delegated access assignments"),
		)
	}

	// Filter subscriptions based on search query
	visibleIndices := make([]int, 0, len(m.lighthouse))
	for i, sub := range m.lighthouse {
		if m.searchActive && m.searchQuery != "" {
			// Search in subscription name, tenant name, and role names
			query := strings.ToLower(m.searchQuery)
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
		visibleIndices = append(visibleIndices, i)
	}

	if len(visibleIndices) == 0 && m.searchActive {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render("🔍"),
			dimStyle.Render(fmt.Sprintf("No subscriptions match \"%s\"", m.searchQuery)),
			dimStyle.Render("Try a different search term"),
		)
	}

	// Find cursor position in visible list (for position indicator)
	cursorVisibleIdx := 0
	for idx, i := range visibleIndices {
		if i == m.lightCursor {
			cursorVisibleIdx = idx
			break
		}
	}

	displayHeight := height - 1 // Reserve for scroll indicator

	// Use stored scroll offset instead of calculating from cursor
	startIdx := m.lightScrollOffset

	// Clamp scroll offset to valid range
	if len(visibleIndices) <= displayHeight {
		startIdx = 0
	} else {
		maxOffset := len(visibleIndices) - displayHeight
		if startIdx > maxOffset {
			startIdx = maxOffset
		}
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Build lines with height constraint - account for tenant headers consuming extra lines
	var lines []string
	lastTenant := ""
	actualEndIdx := startIdx

	// Determine the starting tenant context (for first visible item)
	if startIdx > 0 && startIdx < len(visibleIndices) {
		// Look at the item before startIdx to know if we're mid-tenant
		for j := startIdx - 1; j >= 0; j-- {
			prevSub := m.lighthouse[visibleIndices[j]]
			if prevSub.TenantName != "" {
				lastTenant = prevSub.TenantName
				break
			}
		}
	}

	for _, i := range visibleIndices[startIdx:] {
		sub := m.lighthouse[i]

		// Calculate how many lines this item will add
		linesNeeded := 1 // The subscription item itself
		if sub.TenantName != lastTenant && sub.TenantName != "" {
			linesNeeded++ // Tenant header
			if lastTenant != "" {
				linesNeeded++ // Spacing between tenant groups
			}
		}

		// Check if adding this item would exceed height
		if len(lines)+linesNeeded > displayHeight {
			break
		}

		// Add tenant header when tenant changes
		if sub.TenantName != lastTenant && sub.TenantName != "" {
			if lastTenant != "" {
				// Add spacing between tenant groups (skip for first)
				lines = append(lines, "")
			}
			// Tenant header with building icon
			tenantHeader := dimStyle.Bold(true).Render(fmt.Sprintf("🏢 %s", sub.TenantName))
			lines = append(lines, tenantHeader)
			lastTenant = sub.TenantName
		}
		lines = append(lines, m.renderSubscriptionItem(i, sub))
		actualEndIdx++
	}

	// Add scroll indicator if needed (content extends beyond visible area)
	hasMoreAbove := startIdx > 0
	hasMoreBelow := startIdx+actualEndIdx-startIdx < len(visibleIndices)

	if hasMoreAbove || hasMoreBelow {
		scrollInfo := dimStyle.Render(fmt.Sprintf("  ↕ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		if hasMoreAbove && hasMoreBelow {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↑↓ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		} else if hasMoreAbove {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↑ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		} else if hasMoreBelow {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↓ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		}
		lines = append(lines, scrollInfo)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderSubscriptionItem(idx int, sub azure.LighthouseSubscription) string {
	// Count selected and active roles for this subscription
	selectedCount := 0
	if m.selectedSubRoles[sub.ID] != nil {
		selectedCount = len(m.selectedSubRoles[sub.ID])
	}
	totalRoles := len(sub.EligibleRoles)

	// Count active roles
	activeCount := 0
	for _, role := range sub.EligibleRoles {
		if role.Status.IsActive() {
			activeCount++
		}
	}

	// Determine subscription status based on active roles
	subStatus := sub.Status
	if activeCount > 0 {
		subStatus = azure.StatusActive
	}

	// Build selection and active indicator
	var indicator string
	if selectedCount > 0 {
		// Show selected count in highlight color
		indicator = highlightBoldStyle.Render(fmt.Sprintf(" [%d/%d]", selectedCount, totalRoles))
	} else if activeCount > 0 {
		// Show active count in green
		indicator = activeStyle.Render(fmt.Sprintf(" ●%d", activeCount)) + dimStyle.Render(fmt.Sprintf("/%d", totalRoles))
	} else if totalRoles > 0 {
		indicator = dimStyle.Render(fmt.Sprintf(" [%d]", totalRoles))
	}

	line := fmt.Sprintf("%s %s%s", statusIcon(subStatus), truncate(sub.DisplayName, 26), indicator)

	if idx == m.lightCursor {
		// Highlighted cursor style matching the color scheme
		return cursorStyle.Padding(0, 1).Render(line)
	}
	return itemStyle.Render(line)
}

func (m Model) renderSubscriptionDetail() string {
	sub := m.getCurrentSubscription()
	if sub == nil {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			"",
			dimStyle.Render("📑"),
			"",
			dimStyle.Render("No subscription selected"),
			"",
			dimStyle.Render("Select a subscription from"),
			dimStyle.Render("the list to view details"),
		)
	}
	var lines []string

	// Title with decorative line
	lines = append(lines, detailTitleStyle.Render("━━━ 📑 Subscription Details ━━━"), "")

	// Subscription name
	lines = append(lines, detailLabelStyle.Render("Name: ")+detailValueStyle.Render(sub.DisplayName))

	// Tenant (home tenant of the subscription)
	if sub.TenantName != "" {
		lines = append(lines, detailLabelStyle.Render("Tenant: ")+detailValueStyle.Render(sub.TenantName))
	}

	// Subscription ID
	if sub.ID != "" {
		lines = append(lines, detailLabelStyle.Render("ID: ")+detailDimStyle.Render(truncate(sub.ID, 36)))
	}

	// Eligible Roles section
	lines = append(lines, "", detailDimStyle.Render("─────────────────────────────"))

	// Show focus indicator
	focusHint := ""
	if m.subRoleFocus {
		focusHint = highlightBoldStyle.Render(" [SELECTING]")
	} else if len(sub.EligibleRoles) > 0 {
		focusHint = dimStyle.Render(" (→ to select)")
	}
	lines = append(lines, detailLabelStyle.Render("Eligible Roles:")+focusHint)

	if len(sub.EligibleRoles) == 0 {
		lines = append(lines, detailDimStyle.Italic(true).Render("  (no eligible roles)"))
	} else {
		// Get selected roles for this subscription
		selectedRoles := m.selectedSubRoles[sub.ID]
		if selectedRoles == nil {
			selectedRoles = make(map[int]bool)
		}

		for i, role := range sub.EligibleRoles {
			// Checkbox for selection
			checkbox := dimStyle.Render(checkboxUnchecked)
			if selectedRoles[i] {
				checkbox = highlightBoldStyle.Render(checkboxChecked)
			}

			// Cursor indicator
			cursorPrefix := "  "
			if m.subRoleFocus && i == m.subRoleCursor {
				cursorPrefix = highlightBoldStyle.Render("> ")
			}

			// Role status icon
			roleStatus := statusIcon(role.Status)

			// Role name
			roleName := role.RoleDefinitionName
			if roleName == "" {
				roleName = "Unknown Role"
			}

			// Build the line
			line := fmt.Sprintf("%s%s %s %s", cursorPrefix, checkbox, roleStatus, roleName)

			// Apply cursor style if focused
			if m.subRoleFocus && i == m.subRoleCursor {
				lines = append(lines, cursorStyle.Render(line))
			} else {
				lines = append(lines, line)
			}

			// Show expiry for active roles
			if role.ExpiresAt != nil && role.Status.IsActive() {
				remaining := time.Until(*role.ExpiresAt)
				if remaining > 0 {
					expiryStr := formatCompactDuration(remaining)
					lines = append(lines, detailDimStyle.Render(fmt.Sprintf("       expires: %s", expiryStr)))
				}
			}
		}
	}

	// Navigation hints
	lines = append(lines, "", detailDimStyle.Render("─────────────────────────────"))
	if m.subRoleFocus {
		lines = append(lines, dimStyle.Render("↑↓ navigate │ Space select │ ← back"))
	} else if len(sub.EligibleRoles) > 0 {
		lines = append(lines, dimStyle.Render("→/Tab to select roles │ Space select all"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderLogs() string {
	logHeight := 8
	// Match the width of two side-by-side panels in header/main view
	width := m.width - 6

	var lines []string

	// Add log panel header with level indicator
	levelIndicator := dimStyle.Render("Level: ")
	switch m.logLevel {
	case LogError:
		levelIndicator += errorBoldStyle.Render("ERROR")
	case LogInfo:
		levelIndicator += detailLabelStyle.Render("INFO")
	case LogDebug:
		levelIndicator += dimStyle.Render("DEBUG")
	}
	header := dimStyle.Render("📋 Activity Log") + "  " + levelIndicator
	lines = append(lines, header)
	lines = append(lines, dimStyle.Render(strings.Repeat("─", min(width-4, 50))))

	// Get last N logs (reduced by 2 for header)
	displayHeight := logHeight - 2
	start := len(m.logs) - displayHeight
	if start < 0 {
		start = 0
	}

	if len(m.logs) == 0 {
		lines = append(lines, dimStyle.Render("  No activity yet..."))
	}

	for i := start; i < len(m.logs); i++ {
		entry := m.logs[i]

		// Skip entries below current log level
		if entry.Level > m.logLevel {
			continue
		}

		var levelIcon string
		var msgStyle lipgloss.Style
		switch entry.Level {
		case LogDebug:
			levelIcon = dimStyle.Render("○")
			msgStyle = logDebugStyle
		case LogError:
			levelIcon = errorBoldStyle.Render("●")
			msgStyle = logErrorStyle
		default:
			levelIcon = activeStyle.Render("●")
			msgStyle = logInfoStyle
		}

		timeStr := dimStyle.Render(entry.Time.Format("15:04:05"))

		// Calculate available width for message
		msgWidth := max(width-15, 20)

		// Truncate message before styling to avoid cutting ANSI codes
		msg := entry.Message
		if len(msg) > msgWidth {
			msg = msg[:msgWidth-3] + "..."
		}

		line := fmt.Sprintf("%s %s %s", levelIcon, timeStr, msgStyle.Render(msg))
		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for len(lines) < logHeight {
		lines = append(lines, "")
	}

	return logPanelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) renderStatusBar() string {
	// Duration indicator with visual selector
	var durationDisplay string
	for i, preset := range m.config.DurationPresets {
		if i < 4 {
			if i == m.durationIndex {
				durationDisplay += highlightBoldStyle.Render(fmt.Sprintf("[%dh]", preset))
			} else {
				durationDisplay += dimStyle.Render(fmt.Sprintf(" %dh ", preset))
			}
		}
	}

	// Auto-refresh status
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

	// Active items count
	activeRoles, activeGroups := m.countActiveItems()
	var activeStr string
	if activeRoles+activeGroups > 0 {
		activeStr = activeBoldStyle.Render(fmt.Sprintf("● %d active", activeRoles+activeGroups))
	} else {
		activeStr = dimStyle.Render("● 0 active")
	}

	// Selection count based on active tab
	var selected int
	switch m.activeTab {
	case TabRoles:
		selected = len(m.selectedRoles)
	case TabGroups:
		selected = len(m.selectedGroups)
	case TabSubscriptions:
		// Count total selected roles across all subscriptions
		for _, roleSelections := range m.selectedSubRoles {
			selected += len(roleSelections)
		}
	}
	var selectStr string
	if selected > 0 {
		selectStr = highlightBoldStyle.Render(fmt.Sprintf("✓ %d selected", selected))
	} else {
		selectStr = dimStyle.Render("✓ 0 selected")
	}

	// Search indicator
	var searchStr string
	if m.searchActive {
		searchStr = detailLabelStyle.Render(fmt.Sprintf(" │  🔍 \"%s\"", m.searchQuery))
	}

	// Build status line
	statusLine := fmt.Sprintf("⏱ %s  │  %s  │  %s  │  %s%s",
		durationDisplay, autoStr, activeStr, selectStr, searchStr)

	// Context-aware help hints
	helpHints := dimStyle.Render("←→ tabs │ ↑↓ navigate │ Tab switch │ Space select │ Enter activate │ / search │ ? help")

	return helpStyle.Width(m.width - 2).Render(statusLine + "\n" + helpHints)
}

func (m Model) renderHelp() string {
	// Build dynamic duration help based on config
	durationHelp := ""
	for i, preset := range m.config.DurationPresets {
		if i < 4 {
			durationHelp += fmt.Sprintf("    %d              Set duration to %d hour(s)\n", i+1, preset)
		}
	}

	// Build help with styled sections
	navSection := detailLabelStyle.Render("━━━ Navigation ━━━") + "\n" +
		dimStyle.Render("  ↑/k ↓/j") + detailValueStyle.Render("       Move cursor up/down\n") +
		dimStyle.Render("  ←/h →/l") + detailValueStyle.Render("       Switch tabs (Roles/Groups/Subs)\n") +
		dimStyle.Render("  Tab") + detailValueStyle.Render("           Cycle through tabs\n")

	selectSection := detailLabelStyle.Render("━━━ Selection & Search ━━━") + "\n" +
		dimStyle.Render("  Space") + detailValueStyle.Render("         Select/deselect item\n") +
		dimStyle.Render("  /") + detailValueStyle.Render("             Search/filter\n") +
		dimStyle.Render("  Esc") + detailValueStyle.Render("           Clear search filter\n")

	actionSection := detailLabelStyle.Render("━━━ Actions ━━━") + "\n" +
		dimStyle.Render("  Enter") + detailValueStyle.Render("         Activate selected items\n") +
		dimStyle.Render("  x/Del/BS") + detailValueStyle.Render("      Deactivate active items\n") +
		dimStyle.Render("  r/F5") + detailValueStyle.Render("          Refresh data from Azure\n")

	durationSection := detailLabelStyle.Render(fmt.Sprintf("━━━ Duration (Current: %dh) ━━━", int(m.duration.Hours()))) + "\n" + durationHelp

	settingsSection := detailLabelStyle.Render("━━━ Display & Settings ━━━") + "\n" +
		dimStyle.Render("  v") + detailValueStyle.Render("             Cycle log level\n") +
		dimStyle.Render("  c") + detailValueStyle.Render("             Copy logs to clipboard\n") +
		dimStyle.Render("  e") + detailValueStyle.Render("             Export activation history\n") +
		dimStyle.Render("  a") + detailValueStyle.Render("             Toggle auto-refresh\n") +
		dimStyle.Render("  ?") + detailValueStyle.Render("             Show/hide this help\n") +
		dimStyle.Render("  q/Ctrl+C") + detailValueStyle.Render("      Quit application\n")

	iconSection := detailLabelStyle.Render("━━━ Status Icons ━━━") + "\n" +
		activeStyle.Render("  ● Active") + "       " + lipgloss.NewStyle().Foreground(colorExpiring).Render("◐ Expiring soon\n") +
		dimStyle.Render("  ○ Inactive") + "     " + lipgloss.NewStyle().Foreground(colorPending).Render("◌ Pending approval\n")

	helpContent := "\n" + navSection + "\n" + selectSection + "\n" + actionSection + "\n" +
		durationSection + "\n" + settingsSection + "\n" + iconSection

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Help ━━━") + helpContent,
	)
}

func (m Model) renderConfirm() string {
	count := len(m.pendingActivations)
	countStr := highlightBoldStyle.Render(fmt.Sprintf("%d", count))

	// Build list of items to activate
	var itemList string
	maxShow := 5
	shown := 0
	for _, item := range m.pendingActivations {
		if shown >= maxShow {
			remaining := count - maxShow
			itemList += dimStyle.Render(fmt.Sprintf("  ... and %d more\n", remaining))
			break
		}
		switch v := item.(type) {
		case azure.Role:
			itemList += fmt.Sprintf("  %s %s\n", statusIcon(v.Status), v.DisplayName)
		case azure.Group:
			itemList += fmt.Sprintf("  %s %s\n", statusIcon(v.Status), v.DisplayName)
		case SubscriptionRoleActivation:
			itemList += fmt.Sprintf("  %s %s\n", statusIcon(v.Role.Status), v.Role.RoleDefinitionName)
			itemList += dimStyle.Render(fmt.Sprintf("     on %s\n", truncate(v.SubscriptionName, 35)))
		}
		shown++
	}

	// Duration selector visual
	var durationOptions string
	for i, preset := range m.config.DurationPresets {
		if i < 4 {
			if i == m.durationIndex {
				durationOptions += highlightBoldStyle.Render(fmt.Sprintf(" [%dh] ", preset))
			} else {
				durationOptions += dimStyle.Render(fmt.Sprintf("  %dh  ", preset))
			}
		}
	}

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Confirm Activation ━━━") + "\n\n" +
			fmt.Sprintf("Activate %s item(s):\n", countStr) +
			itemList + "\n" +
			detailLabelStyle.Render("Duration: ") + durationOptions + "\n" +
			dimStyle.Render("(Press 1-4 or Tab to change)\n\n") +
			activeStyle.Render(" [Y] Yes ") + "  " + errorBoldStyle.Render(" [N] No "),
	)
}

func (m Model) renderJustification() string {
	// Duration selector visual
	var durationOptions string
	for i, preset := range m.config.DurationPresets {
		if i < 4 {
			if i == m.durationIndex {
				durationOptions += highlightBoldStyle.Render(fmt.Sprintf(" [%dh] ", preset))
			} else {
				durationOptions += dimStyle.Render(fmt.Sprintf("  %dh  ", preset))
			}
		}
	}

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Justification Required ━━━") + "\n\n" +
			detailLabelStyle.Render("Duration: ") + durationOptions + "\n" +
			dimStyle.Render("(Press 1-4 or Tab to change)\n\n") +
			detailLabelStyle.Render("Reason for activation:") + "\n" +
			m.justificationInput.View() + "\n\n" +
			activeStyle.Render(" [Enter] Confirm ") + "  " + dimStyle.Render(" [Esc] Cancel "),
	)
}

func (m Model) renderActivating() string {
	count := len(m.pendingActivations)
	progressAnimation := spinnerDots(colorActive)

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Activating ━━━") + "\n\n" +
			fmt.Sprintf("%s Processing %d item(s)...\n\n", progressAnimation, count) +
			activeStyle.Render("  ████████████████████  ") + "\n\n" +
			dimStyle.Render("Please wait while Azure processes your request.\n") +
			dimStyle.Render("This may take a few moments."),
	)
}

func (m Model) renderConfirmDeactivate() string {
	count := len(m.pendingDeactivations)
	countStr := errorBoldStyle.Render(fmt.Sprintf("%d", count))

	// Build list of items to deactivate
	var itemList string
	maxShow := 5
	shown := 0
	for _, item := range m.pendingDeactivations {
		if shown >= maxShow {
			remaining := count - maxShow
			itemList += dimStyle.Render(fmt.Sprintf("  ... and %d more\n", remaining))
			break
		}
		switch v := item.(type) {
		case azure.Role:
			itemList += fmt.Sprintf("  %s %s\n", statusIcon(v.Status), v.DisplayName)
		case azure.Group:
			itemList += fmt.Sprintf("  %s %s\n", statusIcon(v.Status), v.DisplayName)
		case SubscriptionRoleActivation:
			itemList += fmt.Sprintf("  %s %s\n", statusIcon(v.Role.Status), v.Role.RoleDefinitionName)
			itemList += dimStyle.Render(fmt.Sprintf("     on %s\n", truncate(v.SubscriptionName, 35)))
		}
		shown++
	}

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorError).Render("━━━ Confirm Deactivation ━━━") + "\n\n" +
			fmt.Sprintf("Deactivate %s active item(s):\n", countStr) +
			itemList + "\n" +
			errorBoldStyle.Render(" [Y] Yes ") + "  " + dimStyle.Render(" [N] No "),
	)
}

func (m Model) renderDeactivating() string {
	count := len(m.pendingDeactivations)
	progressAnimation := spinnerDots(colorError)

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorError).Render("━━━ Deactivating ━━━") + "\n\n" +
			fmt.Sprintf("%s Processing %d item(s)...\n\n", progressAnimation, count) +
			errorBoldStyle.Render("  ████████████████████  ") + "\n\n" +
			dimStyle.Render("Please wait while Azure processes your request.\n") +
			dimStyle.Render("This may take a few moments."),
	)
}

func (m Model) renderSearch() string {
	// Count matches for current search input
	query := m.searchInput.Value()
	var matchInfo string
	if query != "" {
		roleMatches, groupMatches, subMatches := 0, 0, 0
		lowerQuery := strings.ToLower(query)
		for _, r := range m.roles {
			if strings.Contains(strings.ToLower(r.DisplayName), lowerQuery) {
				roleMatches++
			}
		}
		for _, g := range m.groups {
			if strings.Contains(strings.ToLower(g.DisplayName), lowerQuery) {
				groupMatches++
			}
		}
		for _, s := range m.lighthouse {
			match := strings.Contains(strings.ToLower(s.DisplayName), lowerQuery)
			if !match {
				for _, role := range s.EligibleRoles {
					if strings.Contains(strings.ToLower(role.RoleDefinitionName), lowerQuery) {
						match = true
						break
					}
				}
			}
			if match {
				subMatches++
			}
		}
		total := roleMatches + groupMatches + subMatches
		if total > 0 {
			matchInfo = activeStyle.Render(fmt.Sprintf("Found: %d roles, %d groups, %d subs", roleMatches, groupMatches, subMatches))
		} else {
			matchInfo = errorBoldStyle.Render("No matches found")
		}
	} else {
		matchInfo = dimStyle.Render("Type to search...")
	}

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Search / Filter ━━━") + "\n\n" +
			m.searchInput.View() + "\n\n" +
			matchInfo + "\n\n" +
			activeStyle.Render(" [Enter] Apply ") + "  " + dimStyle.Render(" [Esc] Cancel "),
	)
}

// Helper functions

func (m Model) renderItemList(height int, itemType string, count int, getItem func(int) (name string, status azure.ActivationStatus, selected, isCursor bool)) string {
	if count == 0 {
		// Enhanced empty state
		emptyIcon := "📭"
		if itemType == "roles" {
			emptyIcon = "🔐"
		} else if itemType == "groups" {
			emptyIcon = "👥"
		}
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render(emptyIcon),
			dimStyle.Render(fmt.Sprintf("No eligible %s found", itemType)),
			dimStyle.Render("Check your PIM assignments"),
		)
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
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render("🔍"),
			dimStyle.Render(fmt.Sprintf("No %s match \"%s\"", itemType, m.searchQuery)),
			dimStyle.Render("Try a different search term"),
		)
	}

	return strings.Join(lines, "\n")
}

// renderItemListWithExpiry renders a list with optional expiry time display
// scrollOffset is the stored scroll position (index of first visible item)
func (m Model) renderItemListWithExpiry(height int, itemType string, count int, scrollOffset int, getItem func(int) (name string, status azure.ActivationStatus, selected, isCursor bool, expiresAt *time.Time)) string {
	if count == 0 {
		emptyIcon := "📭"
		if itemType == "roles" {
			emptyIcon = "🔐"
		} else if itemType == "groups" {
			emptyIcon = "👥"
		}
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render(emptyIcon),
			dimStyle.Render(fmt.Sprintf("No eligible %s found", itemType)),
			dimStyle.Render("Check your PIM assignments"),
		)
	}

	// First pass: count visible items after filtering
	visibleIndices := make([]int, 0, count)
	for i := 0; i < count; i++ {
		name, _, _, _, _ := getItem(i)
		if m.searchActive && !strings.Contains(strings.ToLower(name), strings.ToLower(m.searchQuery)) {
			continue
		}
		visibleIndices = append(visibleIndices, i)
	}

	if len(visibleIndices) == 0 && m.searchActive {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render("🔍"),
			dimStyle.Render(fmt.Sprintf("No %s match \"%s\"", itemType, m.searchQuery)),
			dimStyle.Render("Try a different search term"),
		)
	}

	// Find cursor position in visible list (for position indicator)
	cursorVisibleIdx := 0
	for idx, i := range visibleIndices {
		_, _, _, isCursor, _ := getItem(i)
		if isCursor {
			cursorVisibleIdx = idx
			break
		}
	}

	// Use stored scroll offset instead of calculating from cursor
	displayHeight := height - 1 // Reserve 1 line for scroll indicator
	startIdx := scrollOffset

	// Clamp scroll offset to valid range
	if len(visibleIndices) <= displayHeight {
		startIdx = 0
	} else {
		maxOffset := len(visibleIndices) - displayHeight
		if startIdx > maxOffset {
			startIdx = maxOffset
		}
		if startIdx < 0 {
			startIdx = 0
		}
	}

	var lines []string
	endIdx := min(startIdx+displayHeight, len(visibleIndices))
	for _, i := range visibleIndices[startIdx:endIdx] {
		name, status, selected, isCursor, expiresAt := getItem(i)
		lines = append(lines, m.renderListItemWithExpiry(i, name, status, selected, isCursor, expiresAt))
	}

	// Add scroll indicator if there are more items
	if len(visibleIndices) > displayHeight {
		scrollInfo := dimStyle.Render(fmt.Sprintf("  ↕ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		if startIdx > 0 && endIdx < len(visibleIndices) {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↑↓ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		} else if startIdx > 0 {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↑ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		} else if endIdx < len(visibleIndices) {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↓ %d/%d", cursorVisibleIdx+1, len(visibleIndices)))
		}
		lines = append(lines, scrollInfo)
	}

	return strings.Join(lines, "\n")
}

// renderListItemWithExpiry renders a list item with optional compact expiry time
func (m Model) renderListItemWithExpiry(idx int, name string, status azure.ActivationStatus, selected, isCursor bool, expiresAt *time.Time) string {
	// Calculate available width for name (accounting for expiry suffix)
	baseWidth := m.listPanelWidth() - 6 // checkbox + status icon
	expiryWidth := 0
	var expirySuffix string

	// Add compact expiry time for active/expiring items
	if expiresAt != nil && status.IsActive() {
		remaining := time.Until(*expiresAt)
		if remaining > 0 {
			expirySuffix = formatCompactDuration(remaining)
			expiryWidth = len(expirySuffix) + 1 // +1 for space
		}
	}

	nameWidth := max(baseWidth-expiryWidth, 10)
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	// Apply search highlighting if search is active
	displayName := name
	if m.searchActive && m.searchQuery != "" {
		displayName = highlightSearchMatch(name, m.searchQuery)
	}

	// Build line with optional expiry
	var line string
	if expirySuffix != "" {
		// Color the expiry based on time remaining
		var expiryStyle lipgloss.Style
		remaining := time.Until(*expiresAt)
		switch {
		case remaining < 15*time.Minute:
			expiryStyle = lipgloss.NewStyle().Foreground(colorCritical)
		case remaining < 30*time.Minute:
			expiryStyle = lipgloss.NewStyle().Foreground(colorWarning)
		case remaining < time.Hour:
			expiryStyle = lipgloss.NewStyle().Foreground(colorExpiring)
		default:
			expiryStyle = dimStyle
		}
		line = fmt.Sprintf("%s %s %s %s", renderCheckbox(selected), statusIcon(status), displayName, expiryStyle.Render(expirySuffix))
	} else {
		line = fmt.Sprintf("%s %s %s", renderCheckbox(selected), statusIcon(status), displayName)
	}

	if isCursor {
		return cursorStyle.Render(line)
	}
	return line
}

// formatCompactDuration formats duration in a compact form like "2h" or "45m"
func formatCompactDuration(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}
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

func spinner(color lipgloss.Color) string {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := int(time.Now().UnixMilli()/100) % len(chars)
	return lipgloss.NewStyle().Foreground(color).Render(chars[idx])
}

// spinnerDots provides a dots-based spinner for activation states
func spinnerDots(color lipgloss.Color) string {
	chars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	idx := int(time.Now().UnixMilli()/80) % len(chars)
	return lipgloss.NewStyle().Foreground(color).Render(chars[idx])
}

// highlightSearchMatch highlights the search query within the text
func highlightSearchMatch(text, query string) string {
	if query == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerText, lowerQuery)
	if idx == -1 {
		return text
	}

	// Build the highlighted string
	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(colorExpiring).
		Bold(true)

	return before + highlightStyle.Render(match) + after
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

func (m Model) countExpiringItems() int {
	count := 0
	for _, r := range m.roles {
		if r.Status == StatusExpiringSoon {
			count++
		}
	}
	for _, g := range m.groups {
		if g.Status == StatusExpiringSoon {
			count++
		}
	}
	return count
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

// wrapPermission wraps a permission string at path segments if it exceeds maxWidth.
// Uses smart breaking at "/" boundaries with proper indentation for continuation lines.
func wrapPermission(perm string, maxWidth int) []string {
	if len(perm) <= maxWidth {
		return []string{perm}
	}

	// Split by "/" to find good break points
	parts := strings.Split(perm, "/")
	if len(parts) <= 2 {
		// Can't break meaningfully, just truncate with ellipsis
		return []string{perm[:maxWidth-3] + "..."}
	}

	// Build lines trying to keep under maxWidth
	var lines []string
	var current string
	indent := "    " // 4 spaces for continuation

	for i, part := range parts {
		separator := ""
		if i > 0 {
			separator = "/"
		}

		proposed := current + separator + part
		if len(proposed) > maxWidth && current != "" {
			lines = append(lines, current)
			current = indent + "/" + part
		} else {
			current = proposed
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return lines
}
