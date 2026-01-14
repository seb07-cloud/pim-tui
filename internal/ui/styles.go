package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sebsebseb1982/pim-tui/internal/config"
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

	// Status icons
	iconActive   = "●"
	iconExpiring = "◐"
	iconInactive = "○"
	iconPending  = "◌"

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

	checkboxChecked   = "[x]"
	checkboxUnchecked = "[ ]"

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

	// Cursor/selection style
	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(colorHighlight).
			Bold(true)

	// Common inline styles
	dimStyle         = lipgloss.NewStyle().Foreground(colorDim)
	activeStyle      = lipgloss.NewStyle().Foreground(colorActive)
	activeBoldStyle  = lipgloss.NewStyle().Foreground(colorActive).Bold(true)
	errorBoldStyle   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	highlightBoldStyle = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)
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
		return lipgloss.NewStyle().Foreground(colorDim).Render("────")
	}

	ratio := remaining / total
	if ratio > 1 {
		ratio = 1
	}

	filled := int(float64(width) * ratio)
	empty := width - filled

	bar := progressBarFull.Render(strings.Repeat("█", filled)) +
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
}
