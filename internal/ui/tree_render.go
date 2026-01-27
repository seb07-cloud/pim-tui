package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/seb07-cloud/pim-tui/internal/azure"
)

// Animation highlight styles - designed for maximum visibility
var (
	// Bright pulse: white text on purple background (very prominent)
	pulseStyleBright = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(colorHighlight)

	// Dim pulse: yellow/gold text on dark background (distinct from normal cursor)
	pulseStyleDim = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffd700")).
		Background(lipgloss.Color("#333333"))

	flowArrowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ff00")).
		Bold(true)
)

// getNodeAnimationStyle returns the appropriate style for a node during animation.
// Handles phase-specific highlighting for the flow animation effect.
func (m TreeViewModel) getNodeAnimationStyle(nodeIndex int, node TreeNode) lipgloss.Style {
	if !m.animating {
		// Not animating, return normal selection style
		if nodeIndex == m.cursor {
			return lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
		}
		return lipgloss.NewStyle()
	}

	switch m.animationPhase {
	case PhaseUserPulse:
		if node.Level == 0 {
			// Pulse effect: alternate every 10 frames
			if (m.animationFrame/10)%2 == 0 {
				return pulseStyleBright
			}
			return pulseStyleDim
		}

	case PhaseGroupsFlow:
		if _, ok := node.Value.(azure.Group); ok {
			groupIndex := m.getGroupIndex(nodeIndex)
			if groupIndex == m.flowProgress {
				// Currently flowing - bright highlight
				return pulseStyleBright
			} else if groupIndex < m.flowProgress {
				// Already flowed through - subtle highlight
				return lipgloss.NewStyle().Foreground(colorHighlight)
			}
		}

	case PhaseRolesReveal:
		if _, ok := node.Value.(azure.Role); ok {
			roleIndex := m.getRoleIndex(nodeIndex)
			if roleIndex <= m.flowProgress {
				// Revealed - use tier color
				if role, ok := node.Value.(azure.Role); ok {
					if tier, found := azure.GetEntraTier(role.RoleDefinitionID); found {
						return TierStyle(tier.Tier).Bold(true)
					}
				}
				return activeStyle
			}
			// Not yet revealed - dim
			return dimStyle
		}

	case PhaseFadeOut:
		// Gradually return to normal - just show normal styling
		if nodeIndex == m.cursor {
			return lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
		}
	}

	// Default during animation: dim non-animated nodes
	if nodeIndex == m.cursor {
		return lipgloss.NewStyle().Bold(true).Foreground(colorHighlight)
	}
	return lipgloss.NewStyle()
}

// buildLipglossTree creates a lipgloss tree structure from TreeViewModel state.
// Shows flow visualization: You -> Activated Roles -> Tenant
func buildLipglossTree(m TreeViewModel) *tree.Tree {
	// Determine root node label with cursor/animation highlighting
	rootLabel := "👤 You"
	if m.animating || m.cursor == 0 {
		if len(m.nodes) > 0 {
			rootStyle := m.getNodeAnimationStyle(0, m.nodes[0])
			// Add pulsing indicator during user pulse phase
			if m.animating && m.animationPhase == PhaseUserPulse {
				if (m.animationFrame/10)%2 == 0 {
					rootLabel = ">> " + rootStyle.Render("👤 You") + " <<"
				} else {
					rootLabel = "> " + rootStyle.Render("👤 You") + " <"
				}
			} else {
				rootLabel = rootStyle.Render("👤 You")
			}
		}
	}

	// Create root tree with rounded enumerator for modern look
	t := tree.Root(rootLabel).Enumerator(tree.RoundedEnumerator)

	// Build tenant label for reuse
	tenantLabel := "🏢 Unknown Tenant"
	if m.tenant != nil {
		tenantLabel = fmt.Sprintf("🏢 %s (%s)", m.tenant.DisplayName, m.tenant.ID)
	}

	// Build tree structure: You -> Active Roles -> Tenant
	for i, node := range m.nodes {
		if node.Level == 0 {
			// Root node "You" - already handled above, skip
			continue
		}

		// Check if this is a role (only roles shown now - only activated ones)
		if role, ok := node.Value.(azure.Role); ok {
			// Apply animation or cursor highlighting
			style := m.getNodeAnimationStyle(i, node)

			// During animation roles reveal, handle dimmed unrevealed roles
			if m.animating && m.animationPhase == PhaseRolesReveal {
				roleIndex := m.getRoleIndex(i)
				if roleIndex > m.flowProgress {
					// Not yet revealed - show dimmed
					roleTree := tree.Root(dimStyle.Render("🔑 " + node.Label))
					roleTree.Child(dimStyle.Render(tenantLabel))
					t.Child(roleTree)
					continue
				}
			}

			// Build role label with tier badge
			roleLabel := node.Label
			var tierBadge string
			if tier, found := azure.GetEntraTier(role.RoleDefinitionID); found {
				tierBadge = " " + TierBadge(tier.Tier)
				if m.animating {
					roleLabel = style.Render("🔑 "+roleLabel) + tierBadge
				} else if i == m.cursor {
					roleLabel = style.Render("🔑 "+roleLabel) + tierBadge
				} else {
					tierStyle := TierStyle(tier.Tier)
					roleLabel = tierStyle.Render("🔑 "+roleLabel) + tierBadge
				}
			} else {
				roleLabel = style.Render("🔑 " + roleLabel)
			}

			// Create role subtree with tenant as child
			roleTree := tree.Root(roleLabel)

			// Add flow arrow during animation
			if m.animating && m.animationPhase == PhaseRolesReveal {
				roleTree.Child(flowArrowStyle.Render("→ ") + activeStyle.Render(tenantLabel))
			} else {
				roleTree.Child(tenantLabel)
			}

			t.Child(roleTree)
		}
	}

	return t
}
