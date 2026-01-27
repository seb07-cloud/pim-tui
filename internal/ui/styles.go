package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/seb07-cloud/pim-tui/internal/config"
)

var (
	// Colors (defaults, can be overridden by config)
	colorActive       = lipgloss.Color("#00ff00") // Green
	colorExpiring     = lipgloss.Color("#ffff00") // Yellow
	colorInactive     = lipgloss.Color("#808080") // Gray
	colorPending      = lipgloss.Color("#00bfff") // Blue
	colorError        = lipgloss.Color("#ff0000") // Red
	colorHighlight    = lipgloss.Color("#7d56f4") // Purple
	colorBorder       = lipgloss.Color("#444444")
	colorBorderActive = lipgloss.Color("#7d56f4")
	colorDim          = lipgloss.Color("#666666")
	colorWarning      = lipgloss.Color("#ff8c00") // Orange - for low time remaining
	colorCritical     = lipgloss.Color("#ff4444") // Light red - for very low time

	// Tier colors - security tier visual indicators
	// Based on Azure/Entra tiering: lower tier = higher privilege = more dangerous
	colorTier0 = lipgloss.Color("#ff0000") // Red - Control Plane (highest privilege)
	colorTier1 = lipgloss.Color("#ff8c00") // Orange - High privilege
	colorTier2 = lipgloss.Color("#ffff00") // Yellow - Medium privilege
	colorTier3 = lipgloss.Color("#00ff00") // Green - Low privilege (safest)

	// Status icons - enhanced with more expressive symbols
	iconActive   = "●"
	iconExpiring = "◐"
	iconInactive = "○"
	iconPending  = "◌"
	iconWarning  = "⚠"

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff"))

	tenantStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activePanelStyle = panelStyle.
				BorderForeground(colorBorderActive)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHighlight).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(colorHighlight).
				Bold(true)

	logPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	logDebugStyle = lipgloss.NewStyle().Foreground(colorDim)
	logInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	logErrorStyle = lipgloss.NewStyle().Foreground(colorError)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	progressBarFull  = lipgloss.NewStyle().Foreground(colorActive)
	progressBarEmpty = lipgloss.NewStyle().Foreground(colorDim)

	// Enhanced checkbox styling with Unicode characters
	checkboxChecked   = "▣"
	checkboxUnchecked = "▢"

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorHighlight).
			Padding(1, 2).
			Align(lipgloss.Center)

	// Detail panel styles
	detailTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
	detailLabelStyle = lipgloss.NewStyle().Foreground(colorPending).Bold(true)
	detailValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	detailDimStyle   = lipgloss.NewStyle().Foreground(colorDim)

	// Cursor/selection style - high contrast inverted for visibility
	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ffffff")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)

	// Common inline styles
	dimStyle           = lipgloss.NewStyle().Foreground(colorDim)
	activeStyle        = lipgloss.NewStyle().Foreground(colorActive)
	activeBoldStyle    = lipgloss.NewStyle().Foreground(colorActive).Bold(true)
	errorBoldStyle     = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	highlightBoldStyle = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)

	// Tab styles - active tab is prominent with filled background
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(colorHighlight).
			Padding(0, 3).
			Border(lipgloss.RoundedBorder(), true, true, false, true).
			BorderForeground(colorHighlight)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				BorderForeground(colorBorder)
)

func statusIcon(status ActivationStatus) string {
	switch status {
	case StatusActive:
		return lipgloss.NewStyle().Foreground(colorActive).Render(iconActive)
	case StatusExpiringSoon:
		return lipgloss.NewStyle().Foreground(colorExpiring).Render(iconExpiring)
	case StatusPending:
		return lipgloss.NewStyle().Foreground(colorPending).Render(iconPending)
	default:
		return lipgloss.NewStyle().Foreground(colorInactive).Render(iconInactive)
	}
}

func renderProgressBar(remaining, total float64, width int) string {
	if total <= 0 || remaining <= 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render(strings.Repeat("░", width))
	}

	ratio := remaining / total
	if ratio > 1 {
		ratio = 1
	}

	filled := int(float64(width) * ratio)
	empty := width - filled

	// Choose color based on remaining time ratio
	var barColor lipgloss.Color
	switch {
	case ratio > 0.5:
		barColor = colorActive // Green - plenty of time
	case ratio > 0.25:
		barColor = colorExpiring // Yellow - getting low
	case ratio > 0.1:
		barColor = colorWarning // Orange - low
	default:
		barColor = colorCritical // Red - critical
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	bar := barStyle.Render(strings.Repeat("█", filled)) +
		progressBarEmpty.Render(strings.Repeat("░", empty))

	return bar
}

// ApplyTheme applies custom theme colors from config
func ApplyTheme(theme config.ThemeConfig) {
	if theme.ColorActive != "" {
		colorActive = lipgloss.Color(theme.ColorActive)
	}
	if theme.ColorExpiring != "" {
		colorExpiring = lipgloss.Color(theme.ColorExpiring)
	}
	if theme.ColorInactive != "" {
		colorInactive = lipgloss.Color(theme.ColorInactive)
	}
	if theme.ColorPending != "" {
		colorPending = lipgloss.Color(theme.ColorPending)
	}
	if theme.ColorError != "" {
		colorError = lipgloss.Color(theme.ColorError)
	}
	if theme.ColorHighlight != "" {
		colorHighlight = lipgloss.Color(theme.ColorHighlight)
		colorBorderActive = lipgloss.Color(theme.ColorHighlight)
	}
	if theme.ColorBorder != "" {
		colorBorder = lipgloss.Color(theme.ColorBorder)
	}

	// Rebuild styles with new colors
	rebuildStyles()
}

func rebuildStyles() {
	tenantStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder)

	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)

	activePanelStyle = panelStyle.
		BorderForeground(colorBorderActive)

	panelTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorHighlight).
		Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(colorHighlight).
		Bold(true)

	logPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)

	logErrorStyle = lipgloss.NewStyle().Foreground(colorError)

	progressBarFull = lipgloss.NewStyle().Foreground(colorActive)

	confirmStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(colorHighlight).
		Padding(1, 2).
		Align(lipgloss.Center)

	detailTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
	detailLabelStyle = lipgloss.NewStyle().Foreground(colorPending).Bold(true)
	detailDimStyle = lipgloss.NewStyle().Foreground(colorDim)

	// Rebuild cursor style for theme customization
	cursorStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#ffffff")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true)
}

// TierColor returns the appropriate color for a given tier level.
// Tier values: "0" (highest privilege/red) through "3" (lowest privilege/green).
// Unknown tier values return a dim gray color.
func TierColor(tier string) lipgloss.Color {
	switch tier {
	case "0":
		return colorTier0
	case "1":
		return colorTier1
	case "2":
		return colorTier2
	case "3":
		return colorTier3
	default:
		return colorDim // Unknown tier
	}
}

// TierStyle returns a styled string renderer for the given tier.
func TierStyle(tier string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(TierColor(tier))
}

// TierBadge returns a compact tier indicator like "T0" with appropriate color.
func TierBadge(tier string) string {
	if tier == "" {
		return dimStyle.Render("T?")
	}
	return TierStyle(tier).Bold(true).Render("T" + tier)
}

// RenderTierLegend returns a formatted tier legend for display.
// Format: "Tiers: T0 Control | T1 High | T2 Medium | T3 Low"
func RenderTierLegend() string {
	return dimStyle.Render("Tiers: ") +
		TierStyle("0").Bold(true).Render("T0") + dimStyle.Render(" Control") +
		dimStyle.Render(" | ") +
		TierStyle("1").Bold(true).Render("T1") + dimStyle.Render(" High") +
		dimStyle.Render(" | ") +
		TierStyle("2").Bold(true).Render("T2") + dimStyle.Render(" Medium") +
		dimStyle.Render(" | ") +
		TierStyle("3").Bold(true).Render("T3") + dimStyle.Render(" Low")
}
