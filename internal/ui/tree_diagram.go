package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/seb07-cloud/pim-tui/internal/azure"
)

// Box drawing characters for flowchart
const (
	boxTopLeft     = "╭"
	boxTopRight    = "╮"
	boxBottomLeft  = "╰"
	boxBottomRight = "╯"
	boxHorizontal  = "─"
	boxVertical    = "│"
	arrowDown      = "▼"
	arrowDownFill  = "▼"
	arrowDownEmpty = "▽"
	lineVertical   = "│"
	lineBranch     = "┼"
	lineLeft       = "┌"
	lineRight      = "┐"
	lineHoriz      = "─"
	flowDot        = "●"
	flowDotEmpty   = "○"
)

// Styles for the flowchart
var (
	userBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7B56F3")).
			Padding(0, 1).
			Bold(true)

	userBoxAnimatedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00ff00")).
				Background(lipgloss.Color("#1a1a2e")).
				Padding(0, 1).
				Bold(true)

	roleBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	tenantBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1).
			Foreground(lipgloss.Color("#888888"))

	tenantBoxAnimatedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00ff00")).
				Padding(0, 1).
				Foreground(lipgloss.Color("#00ff00"))

	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B56F3"))

	arrowAnimatedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff00")).
				Bold(true)

	connectorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	connectorAnimatedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff00"))
)

// Animation phase for diagram (maps to tree animation phases)
const (
	DiagramPhaseIdle = iota
	DiagramPhaseUser
	DiagramPhaseConnector
	DiagramPhaseRoles
	DiagramPhaseTenant
)

// getAnimationPhase maps tree animation to diagram phase
func getAnimationPhase(m TreeViewModel) int {
	if !m.animating {
		return DiagramPhaseIdle
	}
	switch m.animationPhase {
	case PhaseUserPulse:
		return DiagramPhaseUser
	case PhaseGroupsFlow:
		return DiagramPhaseConnector
	case PhaseRolesReveal:
		return DiagramPhaseRoles
	case PhaseFadeOut:
		return DiagramPhaseTenant
	default:
		return DiagramPhaseIdle
	}
}

// buildFlowDiagram creates an ASCII flowchart visualization
// Layout: You (top) -> Roles (middle) -> Single Tenant (bottom)
func buildFlowDiagram(m TreeViewModel) string {
	// Get active roles
	var activeRoles []azure.Role
	for _, node := range m.nodes {
		if role, ok := node.Value.(azure.Role); ok {
			activeRoles = append(activeRoles, role)
		}
	}

	if len(activeRoles) == 0 {
		return buildEmptyDiagram(m.width)
	}

	// Get animation state
	animPhase := getAnimationPhase(m)
	animFrame := m.animationFrame

	// Build role boxes first to get actual widths
	roleBoxes := buildRoleBoxes(activeRoles, animPhase, animFrame, m.flowProgress)

	// Calculate box centers and total width for proper line alignment
	boxWidths := make([]int, len(roleBoxes))
	boxCenters := make([]int, len(roleBoxes))
	spacing := 4 // Space between boxes
	pos := 0
	for i, box := range roleBoxes {
		boxWidths[i] = lipgloss.Width(box)
		boxCenters[i] = pos + boxWidths[i]/2
		pos += boxWidths[i] + spacing
	}
	// Total width is sum of all box widths plus spacing between (not after last)
	rolesWidth := pos - spacing // Remove trailing spacing

	// Build roles row manually to ensure exact positioning
	// Split each box into lines and combine them horizontally
	boxLines := make([][]string, len(roleBoxes))
	maxHeight := 0
	for i, box := range roleBoxes {
		boxLines[i] = strings.Split(box, "\n")
		if len(boxLines[i]) > maxHeight {
			maxHeight = len(boxLines[i])
		}
	}

	// Build rolesRow line by line
	var rolesRowLines []string
	for lineIdx := 0; lineIdx < maxHeight; lineIdx++ {
		var lineBuilder strings.Builder
		for boxIdx, lines := range boxLines {
			if lineIdx < len(lines) {
				lineBuilder.WriteString(lines[lineIdx])
				// Pad if this line is shorter than box width
				lineWidth := lipgloss.Width(lines[lineIdx])
				if lineWidth < boxWidths[boxIdx] {
					lineBuilder.WriteString(strings.Repeat(" ", boxWidths[boxIdx]-lineWidth))
				}
			} else {
				// Pad with spaces if this box has fewer lines
				lineBuilder.WriteString(strings.Repeat(" ", boxWidths[boxIdx]))
			}
			if boxIdx < len(boxLines)-1 {
				lineBuilder.WriteString(strings.Repeat(" ", spacing))
			}
		}
		rolesRowLines = append(rolesRowLines, lineBuilder.String())
	}
	rolesRow := strings.Join(rolesRowLines, "\n")

	// Build tenant box
	tenantBox := buildTenantBox(m.tenant, animPhase, animFrame)

	// Animation styles
	lineStyle := connectorStyle
	arrowLineStyle := arrowStyle
	if animPhase >= DiagramPhaseConnector {
		lineStyle = connectorAnimatedStyle
		arrowLineStyle = arrowAnimatedStyle
	}

	// Build user box
	boxStyle := userBoxStyle
	if animPhase == DiagramPhaseUser && (animFrame/10)%2 == 0 {
		boxStyle = userBoxAnimatedStyle
	}
	userBox := boxStyle.Render("👤 You")

	// Use midpoint between first and last box centers for vertical alignment
	// This matches the visual center of the spread/converge lines
	diagramCenter := rolesWidth / 2
	if len(boxCenters) >= 2 {
		diagramCenter = (boxCenters[0] + boxCenters[len(boxCenters)-1]) / 2
	} else if len(boxCenters) == 1 {
		diagramCenter = boxCenters[0]
	}

	var lines []string

	// Helper to build a line with a single character at a specific position
	// Uses character-by-character building to match spread/converge line positioning
	buildSingleCharLine := func(pos int, totalWidth int, char string, style lipgloss.Style) string {
		var sb strings.Builder
		for i := 0; i < totalWidth; i++ {
			if i == pos {
				sb.WriteString(style.Render(char))
			} else {
				sb.WriteString(" ")
			}
		}
		return sb.String()
	}

	// Helper to build lines with a box centered at a specific position
	// Handles multi-line boxes by padding each line individually
	buildCenteredBoxLine := func(centerPos int, totalWidth int, box string) string {
		boxWidth := lipgloss.Width(box)
		startPos := centerPos - boxWidth/2
		if startPos < 0 {
			startPos = 0
		}

		leftPad := strings.Repeat(" ", startPos)
		boxLines := strings.Split(box, "\n")
		var resultLines []string

		for _, line := range boxLines {
			lineWidth := lipgloss.Width(line)
			rightPad := ""
			remaining := totalWidth - startPos - lineWidth
			if remaining > 0 {
				rightPad = strings.Repeat(" ", remaining)
			}
			resultLines = append(resultLines, leftPad+line+rightPad)
		}

		return strings.Join(resultLines, "\n")
	}

	// Row 1: User box (centered at diagramCenter)
	lines = append(lines, buildCenteredBoxLine(diagramCenter, rolesWidth, userBox))

	// Row 2: Vertical connector (at diagram center)
	vertChar := lineVertical
	if animPhase >= DiagramPhaseConnector && (animFrame/5)%2 == 0 {
		vertChar = flowDot
	}
	lines = append(lines, buildSingleCharLine(diagramCenter, rolesWidth, vertChar, arrowLineStyle))

	// Row 3: Spread line (if multiple roles) - connects box centers
	if len(activeRoles) > 1 {
		spreadLine := buildSpreadLineAtCenters(boxCenters, rolesWidth, diagramCenter, lineStyle, animPhase, animFrame)
		lines = append(lines, spreadLine)
	}

	// Row 4: Down arrows to roles (at box centers)
	arrowsLine := buildDownArrowsAtCenters(boxCenters, boxWidths, rolesWidth, arrowLineStyle, animPhase, animFrame)
	lines = append(lines, arrowsLine)

	// Row 5: Role boxes
	lines = append(lines, rolesRow)

	// Row 6: Vertical lines from roles (at box centers)
	vertLines := buildVertLinesAtCenters(boxCenters, rolesWidth, lineStyle, animPhase, animFrame)
	lines = append(lines, vertLines)

	// Row 7: Converge line (if multiple roles) - connects box centers
	if len(activeRoles) > 1 {
		convergeLine := buildConvergeLineAtCenters(boxCenters, rolesWidth, diagramCenter, lineStyle, animPhase, animFrame)
		lines = append(lines, convergeLine)
	}

	// Row 8: Down arrow to tenant (at diagram center)
	tenantArrowStyle := arrowStyle
	if animPhase >= DiagramPhaseTenant {
		tenantArrowStyle = arrowAnimatedStyle
	}
	lines = append(lines, buildSingleCharLine(diagramCenter, rolesWidth, arrowDown, tenantArrowStyle))

	// Row 9: Tenant box (centered at diagramCenter)
	lines = append(lines, buildCenteredBoxLine(diagramCenter, rolesWidth, tenantBox))

	return strings.Join(lines, "\n")
}

// buildRoleBoxes creates individual role boxes
func buildRoleBoxes(roles []azure.Role, animPhase int, animFrame int, flowProgress int) []string {
	var boxes []string
	for i, role := range roles {
		tierBadge := ""
		style := roleBoxStyle
		if tier, found := azure.GetEntraTier(role.RoleDefinitionID); found {
			tierBadge = " " + TierBadge(tier.Tier)
			style = style.BorderForeground(getTierColor(tier.Tier))
		}

		// Animation effect
		if animPhase == DiagramPhaseRoles && i == flowProgress {
			style = style.BorderForeground(lipgloss.Color("#00ff00")).Bold(true)
			if (animFrame/8)%2 == 0 {
				style = style.Background(lipgloss.Color("#1a2f1a"))
			}
		} else if animPhase >= DiagramPhaseTenant {
			style = style.BorderForeground(lipgloss.Color("#00ff00"))
		}

		content := fmt.Sprintf("🔑 %s%s", role.DisplayName, tierBadge)
		boxes = append(boxes, style.Render(content))
	}
	return boxes
}

// buildTenantBox creates the tenant box
func buildTenantBox(tenant *azure.Tenant, animPhase int, animFrame int) string {
	tenantName := "Unknown Tenant"
	tenantID := "unknown"
	if tenant != nil {
		tenantName = tenant.DisplayName
		tenantID = tenant.ID
	}

	style := tenantBoxStyle
	idStyle := dimStyle
	if animPhase == DiagramPhaseTenant {
		style = tenantBoxAnimatedStyle
		idStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
		if (animFrame/8)%2 == 0 {
			style = style.Background(lipgloss.Color("#1a2f1a"))
		}
	}

	content := fmt.Sprintf("🏢 %s\n%s", tenantName, idStyle.Render(tenantID))
	return style.Render(content)
}

// buildSpreadLineAtCenters creates ┌───┬───┐ line connecting at box centers with outward flow
func buildSpreadLineAtCenters(boxCenters []int, totalWidth int, mid int, style lipgloss.Style, animPhase int, animFrame int) string {
	if len(boxCenters) < 2 {
		return style.Render(lineBranch)
	}

	firstCenter := boxCenters[0]
	lastCenter := boxCenters[len(boxCenters)-1]

	var sb strings.Builder
	for i := 0; i < totalWidth; i++ {
		// Check if at a box center
		isBoxCenter := false
		for _, c := range boxCenters {
			if i == c {
				isBoxCenter = true
				break
			}
		}

		if i < firstCenter || i > lastCenter {
			sb.WriteString(" ")
		} else if i == firstCenter {
			sb.WriteString(style.Render(lineLeft))
		} else if i == lastCenter {
			sb.WriteString(style.Render(lineRight))
		} else if isBoxCenter {
			sb.WriteString(style.Render(lineBranch))
		} else if i == mid {
			sb.WriteString(style.Render(lineBranch))
		} else if animPhase >= DiagramPhaseConnector {
			distFromCenter := i - mid
			if distFromCenter < 0 {
				distFromCenter = -distFromCenter
			}
			if (animFrame/3-distFromCenter)%5 == 0 {
				sb.WriteString(style.Render(flowDot))
			} else {
				sb.WriteString(style.Render(lineHoriz))
			}
		} else {
			sb.WriteString(style.Render(lineHoriz))
		}
	}
	return sb.String()
}

// buildDownArrowsAtCenters creates arrows at exact box center positions
func buildDownArrowsAtCenters(boxCenters []int, boxWidths []int, totalWidth int, style lipgloss.Style, animPhase int, animFrame int) string {
	var sb strings.Builder
	for i := 0; i < totalWidth; i++ {
		isCenter := false
		for _, c := range boxCenters {
			if i == c {
				isCenter = true
				break
			}
		}
		if isCenter {
			arrow := arrowDown
			if animPhase >= DiagramPhaseConnector && (animFrame/5)%2 == 0 {
				arrow = arrowDownFill
			}
			sb.WriteString(style.Render(arrow))
		} else {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

// buildVertLinesAtCenters creates vertical lines at exact box center positions
func buildVertLinesAtCenters(boxCenters []int, totalWidth int, style lipgloss.Style, animPhase int, animFrame int) string {
	var sb strings.Builder
	for i := 0; i < totalWidth; i++ {
		isCenter := false
		for _, c := range boxCenters {
			if i == c {
				isCenter = true
				break
			}
		}
		if isCenter {
			char := lineVertical
			if animPhase >= DiagramPhaseRoles && (animFrame/5)%2 == 0 {
				char = flowDot
			}
			sb.WriteString(style.Render(char))
		} else {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

// buildConvergeLineAtCenters creates ╰───┴───╯ line connecting at box centers with inward flow
func buildConvergeLineAtCenters(boxCenters []int, totalWidth int, mid int, style lipgloss.Style, animPhase int, animFrame int) string {
	if len(boxCenters) < 2 {
		return style.Render(lineBranch)
	}

	firstCenter := boxCenters[0]
	lastCenter := boxCenters[len(boxCenters)-1]

	var sb strings.Builder
	for i := 0; i < totalWidth; i++ {
		// Check if at a box center
		isBoxCenter := false
		for _, c := range boxCenters {
			if i == c {
				isBoxCenter = true
				break
			}
		}

		if i < firstCenter || i > lastCenter {
			sb.WriteString(" ")
		} else if i == firstCenter {
			sb.WriteString(style.Render(boxBottomLeft))
		} else if i == lastCenter {
			sb.WriteString(style.Render(boxBottomRight))
		} else if isBoxCenter {
			sb.WriteString(style.Render(lineBranch))
		} else if i == mid {
			sb.WriteString(style.Render(lineBranch))
		} else if animPhase >= DiagramPhaseTenant {
			distFromEdge := i - firstCenter
			if lastCenter-i < distFromEdge {
				distFromEdge = lastCenter - i
			}
			if (animFrame/3-distFromEdge)%5 == 0 {
				sb.WriteString(style.Render(flowDot))
			} else {
				sb.WriteString(style.Render(lineHoriz))
			}
		} else {
			sb.WriteString(style.Render(lineHoriz))
		}
	}
	return sb.String()
}

func buildEmptyDiagram(panelWidth int) string {
	userBox := userBoxStyle.Render("👤 You")
	boxWidth := lipgloss.Width(userBox)

	// Center the content
	availWidth := panelWidth - 4
	leftPad := (availWidth - boxWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	pad := strings.Repeat(" ", leftPad)

	msg := dimStyle.Render("No active roles to display.\nActivate a role to see the flow.")
	msgWidth := lipgloss.Width(msg)
	msgPad := (availWidth - msgWidth) / 2
	if msgPad < 0 {
		msgPad = 0
	}

	return pad + userBox + "\n\n" + strings.Repeat(" ", msgPad) + msg
}

func getTierColor(tier string) lipgloss.Color {
	switch tier {
	case "0":
		return lipgloss.Color("#ff6b6b") // Red - Critical
	case "1":
		return lipgloss.Color("#ffa500") // Orange - High
	case "2":
		return lipgloss.Color("#90EE90") // Green - Medium
	case "3":
		return lipgloss.Color("#87CEEB") // Light Blue - Low
	default:
		return lipgloss.Color("#888888") // Gray
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func intersperse(items []string, sep string) []string {
	if len(items) == 0 {
		return items
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		result = append(result, item)
		if i < len(items)-1 {
			result = append(result, sep)
		}
	}
	return result
}
