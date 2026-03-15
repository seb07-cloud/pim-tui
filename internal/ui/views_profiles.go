package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/seb07-cloud/pim-tui/internal/azure"
)

func (m Model) renderProfilesList(height int) string {
	if len(m.profiles) == 0 {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			dimStyle.Render("📋"),
			dimStyle.Render("No profiles configured"),
			dimStyle.Render("Add profiles to ~/.config/pim-tui/profiles.yaml"),
		)
	}

	displayHeight := height - 1 // Reserve 1 line for scroll indicator
	startIdx := m.profilesScrollOffset

	// Clamp scroll offset
	if len(m.profiles) <= displayHeight {
		startIdx = 0
	} else {
		maxOffset := len(m.profiles) - displayHeight
		if startIdx > maxOffset {
			startIdx = maxOffset
		}
		if startIdx < 0 {
			startIdx = 0
		}
	}

	var lines []string
	endIdx := min(startIdx+displayHeight, len(m.profiles))
	for i := startIdx; i < endIdx; i++ {
		p := m.profiles[i]
		isCursor := i == m.profilesCursor && m.activeTab == TabProfiles

		// Build display: "ProfileName (N items) [TierBadge]"
		entryCount := len(p.Entries)
		name := fmt.Sprintf("%s (%d items)", p.Name, entryCount)

		// Resolve to get tier badge (lightweight — just for display)
		rp := resolveProfile(p, m.roles, m.groups, m.lighthouse)
		tierStr := profileTierBadgeStr(&rp)
		if tierStr != "" {
			name += " " + TierBadge(tierStr)
		}

		// Status indicator based on resolution
		statusStr := dimStyle.Render("○")
		if rp.AllValid {
			statusStr = activeStyle.Render("●")
		} else {
			statusStr = lipgloss.NewStyle().Foreground(colorError).Render("●")
		}

		// Truncate name to fit
		nameWidth := max(m.listPanelContentWidth()-6, 10)
		if lipgloss.Width(name) > nameWidth {
			name = ansi.Truncate(name, nameWidth, "...")
		}

		line := fmt.Sprintf("  %s %s", statusStr, name)

		if isCursor {
			line = cursorStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// Scroll indicator
	if len(m.profiles) > displayHeight {
		scrollInfo := dimStyle.Render(fmt.Sprintf("  ↕ %d/%d", m.profilesCursor+1, len(m.profiles)))
		if startIdx > 0 && endIdx < len(m.profiles) {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↑↓ %d/%d", m.profilesCursor+1, len(m.profiles)))
		} else if startIdx > 0 {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↑ %d/%d", m.profilesCursor+1, len(m.profiles)))
		} else if endIdx < len(m.profiles) {
			scrollInfo = dimStyle.Render(fmt.Sprintf("  ↓ %d/%d", m.profilesCursor+1, len(m.profiles)))
		}
		lines = append(lines, scrollInfo)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderProfileDetail() string {
	if len(m.profiles) == 0 || m.profilesCursor >= len(m.profiles) {
		return lipgloss.JoinVertical(lipgloss.Center,
			"",
			"",
			dimStyle.Render("📋"),
			"",
			dimStyle.Render("No profile selected"),
			"",
			dimStyle.Render("Select a profile from the list"),
			dimStyle.Render("to view its details"),
		)
	}

	p := m.profiles[m.profilesCursor]
	rp := resolveProfile(p, m.roles, m.groups, m.lighthouse)

	var lines []string
	lines = append(lines, detailTitleStyle.Render("━━━ 📋 Profile Details ━━━"), "")
	lines = append(lines, detailLabelStyle.Render("Name: ")+detailValueStyle.Render(p.Name))

	if p.Description != "" {
		lines = append(lines, detailLabelStyle.Render("Desc: ")+detailDimStyle.Render(p.Description))
	}

	if p.Duration > 0 {
		lines = append(lines, detailLabelStyle.Render("Duration: ")+detailValueStyle.Render(fmt.Sprintf("%dh", p.Duration)))
	}

	// Tier badge
	tierStr := profileTierBadgeStr(&rp)
	if tierStr != "" {
		tierLine := detailLabelStyle.Render("Max Tier: ") + TierBadge(tierStr)
		switch tierStr {
		case "0":
			tierLine += " " + TierStyle("0").Render("(Control Plane)")
		case "1":
			tierLine += " " + TierStyle("1").Render("(High Privilege)")
		case "2":
			tierLine += " " + TierStyle("2").Render("(Medium Privilege)")
		case "3":
			tierLine += " " + TierStyle("3").Render("(Low Privilege)")
		}
		lines = append(lines, tierLine)
	}

	// Justification template
	if p.Justification != "" {
		justLine := p.Justification
		// Highlight template variables
		justLine = templateVarPattern.ReplaceAllStringFunc(justLine, func(match string) string {
			return highlightBoldStyle.Render(match)
		})
		lines = append(lines, detailLabelStyle.Render("Justification: ")+justLine)
	}

	// Validation status
	lines = append(lines, "")
	if rp.AllValid {
		lines = append(lines, activeBoldStyle.Render("✓ All entries match eligible roles"))
	} else {
		lines = append(lines, errorBoldStyle.Render("✗ Some entries do not match"))
	}

	// Entry list
	lines = append(lines, "", detailDimStyle.Render("─── Entries ───"))
	for _, entry := range rp.Entries {
		icon := activeStyle.Render("✓")
		if !entry.Matched {
			icon = errorBoldStyle.Render("✗")
		}

		typeIcon := "🔐"
		switch entry.Entry.Type {
		case "group":
			typeIcon = "👥"
		case "subscription-role":
			typeIcon = "📑"
		}

		entryLine := fmt.Sprintf("  %s %s %s", icon, typeIcon, entry.Entry.Name)
		if entry.Entry.Type == "subscription-role" && entry.Entry.Subscription != "" {
			entryLine += dimStyle.Render(" on " + entry.Entry.Subscription)
		}
		lines = append(lines, entryLine)

		if !entry.Matched && entry.Warning != "" {
			lines = append(lines, "    "+lipgloss.NewStyle().Foreground(colorError).Render(entry.Warning))
		}
	}

	lines = append(lines, "", dimStyle.Render("Press Enter to activate"))

	return strings.Join(lines, "\n")
}

func (m Model) renderProfileVars() string {
	if m.selectedProfile == nil || len(m.selectedProfile.Variables) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, detailLabelStyle.Render("Profile: ")+detailValueStyle.Render(m.selectedProfile.Profile.Name))
	lines = append(lines, "")

	// Show justification template with filled/unfilled variables
	template := m.selectedProfile.Profile.Justification
	preview := template
	for i, varName := range m.selectedProfile.Variables {
		if i < m.profileVarIndex {
			// Already filled
			val := m.profileVarValues[varName]
			preview = strings.ReplaceAll(preview, "{{"+varName+"}}", activeBoldStyle.Render(val))
		} else if i == m.profileVarIndex {
			// Currently editing
			preview = strings.ReplaceAll(preview, "{{"+varName+"}}", highlightBoldStyle.Render("▏"+varName))
		}
		// Unfilled variables keep their {{variable}} display
	}
	lines = append(lines, detailLabelStyle.Render("Justification: ")+preview)
	lines = append(lines, "")

	// Current variable input
	varName := m.selectedProfile.Variables[m.profileVarIndex]
	lines = append(lines, detailLabelStyle.Render(fmt.Sprintf("Enter %s:", varName)))
	lines = append(lines, m.profileVarInputs[m.profileVarIndex].View())
	lines = append(lines, "")

	// Progress indicator
	progress := fmt.Sprintf("Variable %d of %d", m.profileVarIndex+1, len(m.selectedProfile.Variables))
	lines = append(lines, dimStyle.Render(progress))
	lines = append(lines, "")
	lines = append(lines, activeStyle.Render(" [Enter] Next ")+
		"  "+dimStyle.Render(" [Esc] Cancel "))

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Profile Variables ━━━") + "\n\n" +
			strings.Join(lines, "\n"),
	)
}

func (m Model) renderProfileConfirm() string {
	if m.selectedProfile == nil {
		return ""
	}

	rp := m.selectedProfile
	count := len(rp.Entries)
	countStr := highlightBoldStyle.Render(fmt.Sprintf("%d", count))

	// Tier 0 warning
	tierWarning := ""
	if rp.MaxTier == 0 {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(colorTier0).
			Bold(true).
			Padding(0, 1)
		tierWarning = warningStyle.Render(" ⚠ CONTROL PLANE ACCESS ") + "\n"
		tierWarning += TierStyle("0").Render("Tier 0 role grants tenant-wide admin rights") + "\n\n"
	}

	// Build entry list with match status
	var itemList string
	maxShow := 8
	shown := 0
	for _, entry := range rp.Entries {
		if shown >= maxShow {
			remaining := count - maxShow
			itemList += dimStyle.Render(fmt.Sprintf("  ... and %d more\n", remaining))
			break
		}

		matchIcon := activeStyle.Render("✓")
		if !entry.Matched {
			matchIcon = errorBoldStyle.Render("✗")
		}

		switch v := entry.MatchedAs.(type) {
		case azure.Role:
			badge := getRoleTierBadge(v.RoleDefinitionID)
			itemList += fmt.Sprintf("  %s %s %s%s\n", matchIcon, statusIcon(v.Status), v.DisplayName, badge)
		case azure.Group:
			itemList += fmt.Sprintf("  %s %s %s\n", matchIcon, statusIcon(v.Status), v.DisplayName)
		case SubscriptionRoleActivation:
			badge := ""
			if tier, found := azure.GetAzureTier(v.Role.RoleDefinitionID); found {
				badge = " " + TierBadge(tier.Tier)
			}
			itemList += fmt.Sprintf("  %s %s %s%s\n", matchIcon, statusIcon(v.Role.Status), v.Role.RoleDefinitionName, badge)
			itemList += dimStyle.Render(fmt.Sprintf("     on %s\n", v.SubscriptionName))
		default:
			// Unmatched entry
			itemList += fmt.Sprintf("  %s   %s", matchIcon, entry.Entry.Name)
			if entry.Warning != "" {
				itemList += " " + lipgloss.NewStyle().Foreground(colorError).Render("("+entry.Warning+")")
			}
			itemList += "\n"
		}
		shown++
	}

	// Duration selector
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

	// Justification preview
	justification := m.justificationInput.Value()

	// Action buttons
	var actionLine string
	if rp.AllValid {
		actionLine = activeStyle.Render(" [Y] Activate ") + "  " + dimStyle.Render(" [N] Cancel ")
	} else {
		actionLine = errorBoldStyle.Render("Cannot activate — unmatched entries") + "\n" +
			dimStyle.Render(" [Esc] Back ")
	}

	return confirmStyle.Width(m.dialogWidth()).Render(
		titleStyle.Foreground(colorHighlight).Render("━━━ Activate Profile ━━━") + "\n\n" +
			tierWarning +
			detailLabelStyle.Render("Profile: ") + detailValueStyle.Render(rp.Profile.Name) + "\n" +
			fmt.Sprintf("Activate %s item(s):\n", countStr) +
			itemList + "\n" +
			detailLabelStyle.Render("Justification: ") + detailValueStyle.Render(justification) + "\n\n" +
			detailLabelStyle.Render("Duration: ") + durationOptions + "\n" +
			dimStyle.Render("(Press 1-4 or Tab to change)\n\n") +
			actionLine,
	)
}
