package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seb07-cloud/pim-tui/internal/azure"
)

// TreeNode represents a single node in the flattened tree structure.
// Nodes are organized as: User (level 0) -> Groups and Roles (level 1).
// Both groups and roles are direct children of the user node.
type TreeNode struct {
	ID          int         // Unique identifier for this node
	Level       int         // Depth: 0=user, 1=group or role
	Label       string      // Display text
	Value       interface{} // Underlying data: azure.Group or azure.Role
	HasChildren bool        // Whether node can be expanded
}

// TreeViewModel manages tree view state and navigation.
// It implements the Bubbletea tea.Model interface for keyboard-driven interaction.
type TreeViewModel struct {
	cursor        int            // Current selection index in visible nodes
	nodes         []TreeNode     // Flattened list of currently visible nodes
	expanded      map[int]bool   // Tracks which nodes are expanded by ID
	width         int            // Rendering width
	height        int            // Rendering height
	groups        []azure.Group  // Source data: groups
	roles         []azure.Role   // Source data: roles
	tenant        *azure.Tenant  // Current tenant info for display
	showingDetail bool           // True when detail panel is visible
	detailRole    *azure.Role    // Role being displayed in detail panel

	// Animation state
	animating      bool // True when animation is running
	animationID    int  // Unique ID for current animation (prevents stale ticks)
	animationPhase int  // Current phase (PhaseIdle, PhaseUserPulse, etc.)
	animationFrame int  // Frame counter within current phase
	flowProgress   int  // Index of current node in flow sequence (group or role index)
}

// NewTreeView creates a new TreeViewModel with the specified dimensions.
func NewTreeView(width, height int) TreeViewModel {
	return TreeViewModel{
		expanded: make(map[int]bool),
		nodes:    make([]TreeNode, 0),
		width:    width,
		height:   height,
	}
}

// SetData populates the tree with groups, roles, and tenant data.
// Only activated roles are shown in the tree.
func (m *TreeViewModel) SetData(groups []azure.Group, roles []azure.Role, tenant *azure.Tenant) {
	m.groups = groups
	m.roles = roles
	m.tenant = tenant
	m.rebuildVisibleNodes()
}

// SetSize updates the tree view dimensions.
func (m *TreeViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init implements tea.Model.
func (m TreeViewModel) Init() tea.Cmd {
	return nil
}

// View implements tea.Model and renders the tree with a styled container.
func (m TreeViewModel) View() string {
	// If showing role detail, render detail panel instead of tree
	if m.showingDetail && m.detailRole != nil {
		title := panelTitleStyle.Render("Role Inheritance Map")
		detail := renderRoleDetail(m.detailRole, m.width)
		content := title + "\n\n" + detail

		container := panelStyle.
			Width(m.width).
			Height(m.height)

		return container.Render(content)
	}

	// Build flowchart diagram with width for proper centering
	diagram := buildFlowDiagram(m)

	// Title
	title := panelTitleStyle.Render("Active Roles Flow")

	// Help text with tier legend - dynamic based on animation state
	legend := RenderTierLegend()
	var help string
	if m.animating {
		help = helpStyle.Render("a: stop animation | Esc: back")
	} else {
		help = helpStyle.Render("j/k: navigate | Enter: details | a: animate flow | Esc: back")
	}

	// Wrap diagram in a fixed-width block to preserve internal alignment
	// This prevents lipgloss.Place from centering each line independently
	diagramWidth := lipgloss.Width(diagram)
	fixedDiagram := lipgloss.NewStyle().
		Width(diagramWidth).
		Render(diagram)

	// Main content (title + diagram) - will be centered as a block
	mainContent := title + "\n\n" + fixedDiagram

	// Footer content (legend + help) - will be at bottom
	footer := legend + "\n" + help

	// Calculate dimensions
	availableWidth := m.width - 2   // Account for border
	availableHeight := m.height - 2 // Account for border
	footerHeight := lipgloss.Height(footer)

	// Center the main content in the space above the footer
	mainAreaHeight := availableHeight - footerHeight - 1 // -1 for spacing
	centeredMain := lipgloss.Place(
		availableWidth,
		mainAreaHeight,
		lipgloss.Center,
		lipgloss.Center,
		mainContent,
	)

	// Footer centered horizontally at bottom
	centeredFooter := lipgloss.NewStyle().
		Width(availableWidth).
		Align(lipgloss.Center).
		Render(footer)

	// Combine: centered main + footer at bottom
	fullContent := centeredMain + "\n" + centeredFooter

	// Container with border
	container := panelStyle.
		Width(m.width).
		Height(m.height)

	return container.Render(fullContent)
}

// Update implements tea.Model and handles keyboard navigation.
// Follows W3C TreeView keyboard standards with vim-style alternatives.
func (m TreeViewModel) Update(msg tea.Msg) (TreeViewModel, tea.Cmd) {
	// If showing detail, Enter/Esc/Space dismisses it
	if m.showingDetail {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter", "esc", " ":
				m.showingDetail = false
				m.detailRole = nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Navigate up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Navigate down
		case "down", "j":
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}

		// Show role detail on Enter
		case "enter", " ":
			if len(m.nodes) > 0 && m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				// Show detail for role nodes (check by type)
				if role, ok := node.Value.(azure.Role); ok {
					m.showingDetail = true
					m.detailRole = &role
				}
			}

		// Expand or move to first child
		case "right", "l":
			if len(m.nodes) > 0 && m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				if node.HasChildren && !m.expanded[node.ID] {
					// Expand closed node
					m.expanded[node.ID] = true
					m.rebuildVisibleNodes()
				} else if m.expanded[node.ID] && m.cursor < len(m.nodes)-1 {
					// Already expanded, move to first child
					m.cursor++
				}
			}

		// Collapse or move to parent
		case "left", "h":
			if len(m.nodes) > 0 && m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				if m.expanded[node.ID] {
					// Collapse expanded node
					m.expanded[node.ID] = false
					m.rebuildVisibleNodes()
					// Clamp cursor after rebuild
					if m.cursor >= len(m.nodes) {
						m.cursor = len(m.nodes) - 1
					}
				} else if node.Level > 0 {
					// Move to parent node (search backwards for node.Level-1)
					for i := m.cursor - 1; i >= 0; i-- {
						if m.nodes[i].Level == node.Level-1 {
							m.cursor = i
							break
						}
					}
				}
			}

		// Jump to beginning
		case "home", "g":
			m.cursor = 0

		// Jump to end
		case "end", "G":
			if len(m.nodes) > 0 {
				m.cursor = len(m.nodes) - 1
			}
		}
	}

	return m, nil
}

// SelectedNode returns the currently selected node, or nil if no nodes exist.
func (m TreeViewModel) SelectedNode() *TreeNode {
	if len(m.nodes) == 0 || m.cursor < 0 || m.cursor >= len(m.nodes) {
		return nil
	}
	return &m.nodes[m.cursor]
}

// Nodes returns the current list of visible nodes.
func (m TreeViewModel) Nodes() []TreeNode {
	return m.nodes
}

// Cursor returns the current cursor position.
func (m TreeViewModel) Cursor() int {
	return m.cursor
}

// IsExpanded checks if a node is expanded.
func (m TreeViewModel) IsExpanded(nodeID int) bool {
	return m.expanded[nodeID]
}

// rebuildVisibleNodes creates a flattened list of visible nodes based on expand state.
// Shows: You -> Activated Roles -> Tenant (flow visualization)
func (m *TreeViewModel) rebuildVisibleNodes() {
	m.nodes = make([]TreeNode, 0)
	nodeID := 0

	// Filter to only activated roles
	var activeRoles []azure.Role
	for _, role := range m.roles {
		if role.Status.IsActive() {
			activeRoles = append(activeRoles, role)
		}
	}

	// Level 0: Root node "You" (always visible)
	m.nodes = append(m.nodes, TreeNode{
		ID:          nodeID,
		Level:       0,
		Label:       "You",
		Value:       nil,
		HasChildren: len(activeRoles) > 0,
	})
	nodeID++

	// Level 1: Only ACTIVATED roles (showing flow to tenant)
	for _, role := range activeRoles {
		m.nodes = append(m.nodes, TreeNode{
			ID:          nodeID,
			Level:       1,
			Label:       role.DisplayName,
			Value:       role,
			HasChildren: m.tenant != nil, // Has tenant as child
		})
		nodeID++
	}

	// Store active roles count for animation
	m.roles = activeRoles
}

// renderRoleDetail creates a styled detail panel for a role.
func renderRoleDetail(role *azure.Role, width int) string {
	if role == nil {
		return ""
	}

	// Build detail lines using existing detail styles
	lines := []string{
		detailTitleStyle.Render("Role Details"),
		"",
		detailLabelStyle.Render("Name: ") + detailValueStyle.Render(role.DisplayName),
	}

	// Add tier information if available
	if tier, found := azure.GetEntraTier(role.RoleDefinitionID); found {
		tierLine := detailLabelStyle.Render("Security Tier: ") + TierBadge(tier.Tier)
		switch tier.Tier {
		case "0":
			tierLine += " " + TierStyle("0").Render("(Control Plane - Highest Risk)")
		case "1":
			tierLine += " " + TierStyle("1").Render("(High Privilege)")
		case "2":
			tierLine += " " + TierStyle("2").Render("(Medium Privilege)")
		case "3":
			tierLine += " " + TierStyle("3").Render("(Low Privilege)")
		}
		lines = append(lines, tierLine)

		// Add path type if available
		if tier.PathType != "" {
			pathStyle := detailValueStyle
			if tier.PathType == "Direct" {
				pathStyle = TierStyle("0") // Red for direct escalation
			}
			lines = append(lines, detailLabelStyle.Render("Path Type: ")+pathStyle.Render(tier.PathType))
		}

		// Add attack path details for Tier 0 roles
		if tier.Tier == "0" {
			lines = append(lines, "", detailDimStyle.Render("─── Attack Path ───"))
			if tier.ShortestPath != "" {
				// Truncate for panel width
				pathText := tier.ShortestPath
				maxLen := width - 20
				if maxLen < 20 {
					maxLen = 20
				}
				if len(pathText) > maxLen {
					pathText = pathText[:maxLen-3] + "..."
				}
				lines = append(lines, detailLabelStyle.Render("Via: ")+TierStyle("0").Render(pathText))
			}
			if tier.Example != "" {
				// Truncate for panel width
				example := tier.Example
				maxLen := width - 20
				if maxLen < 20 {
					maxLen = 20
				}
				if len(example) > maxLen {
					example = example[:maxLen-3] + "..."
				}
				lines = append(lines, detailLabelStyle.Render("Risk: ")+detailDimStyle.Render(example))
			}
		}
	}

	if role.Description != "" {
		// Wrap description if too long
		desc := role.Description
		if len(desc) > width-10 {
			desc = desc[:width-13] + "..."
		}
		lines = append(lines, detailLabelStyle.Render("Description: ")+detailValueStyle.Render(desc))
	}

	lines = append(lines, detailLabelStyle.Render("Status: ")+detailValueStyle.Render(role.Status.String()))

	if role.MaxDuration > 0 {
		lines = append(lines, detailLabelStyle.Render("Max Duration: ")+detailValueStyle.Render(role.MaxDuration.String()))
	}

	if role.DirectoryScopeID != "" {
		lines = append(lines, detailLabelStyle.Render("Scope: ")+detailValueStyle.Render(role.DirectoryScopeID))
	}

	if len(role.Permissions) > 0 {
		permCount := len(role.Permissions)
		if permCount > 3 {
			lines = append(lines, detailLabelStyle.Render("Permissions: ")+detailValueStyle.Render(fmt.Sprintf("%d total", permCount)))
		} else {
			lines = append(lines, detailLabelStyle.Render("Permissions: ")+detailValueStyle.Render(strings.Join(role.Permissions, ", ")))
		}
	}

	lines = append(lines, "", detailDimStyle.Render("Press Enter or Esc to close"))

	content := strings.Join(lines, "\n")

	// Use existing panel style with border
	return panelStyle.
		Width(width - 4).
		Render(content)
}
