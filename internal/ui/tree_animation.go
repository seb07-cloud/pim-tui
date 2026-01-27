package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/seb07-cloud/pim-tui/internal/azure"
)

// Animation phase constants for the flow animation state machine.
// The animation progresses through these phases sequentially to visualize
// permission inheritance flow: User -> Groups -> Roles.
const (
	PhaseIdle       int = iota // No animation running
	PhaseUserPulse             // Highlight user node (root)
	PhaseGroupsFlow            // Cascade through groups
	PhaseRolesReveal           // Pulse roles with tier colors
	PhaseFadeOut               // Return to normal
)

// Animation timing constants.
// Based on 30fps for smooth but slower animation (~33ms per frame).
const (
	AnimationFPS      = 30
	AnimationInterval = time.Second / 30 // ~33ms per frame
	UserPulseFrames   = 45               // 1.5s at 30fps
	GroupFlowFrames   = 30               // 1s per group
	RoleRevealFrames  = 24               // 800ms per role
	FadeOutFrames     = 30               // 1s fade
)

// animationTickMsg is the tick message type for animation frames.
// It's separate from the regular tickMsg to avoid conflicts with auto-refresh timing.
// The id field allows cancellation by invalidating pending ticks when animation stops.
type animationTickMsg struct {
	time time.Time
	id   int
}

// animationTick schedules the next animation frame with the given animation ID.
// Returns a tea.Cmd that will send an animationTickMsg after AnimationInterval.
func animationTick(id int) tea.Cmd {
	return tea.Tick(AnimationInterval, func(t time.Time) tea.Msg {
		return animationTickMsg{time: t, id: id}
	})
}

// StartAnimation initiates the flow animation on the TreeViewModel.
// Returns a tea.Cmd that schedules the first animation tick.
func (m *TreeViewModel) StartAnimation() tea.Cmd {
	m.animating = true
	m.animationID++            // New ID invalidates pending ticks from any previous animation
	m.animationPhase = PhaseUserPulse
	m.animationFrame = 0
	m.flowProgress = 0
	return animationTick(m.animationID)
}

// StopAnimation cancels any running animation immediately.
// Increments animationID to invalidate any pending tick messages.
func (m *TreeViewModel) StopAnimation() {
	m.animating = false
	m.animationPhase = PhaseIdle
	m.animationID++ // Invalidate pending ticks
}

// UpdateAnimation handles animation tick messages and advances the state machine.
// Returns a tea.Cmd for the next tick, or nil when animation completes.
func (m *TreeViewModel) UpdateAnimation(msg animationTickMsg) tea.Cmd {
	// Reject stale ticks from cancelled animations
	if msg.id != m.animationID || !m.animating {
		return nil
	}

	m.animationFrame++

	switch m.animationPhase {
	case PhaseUserPulse:
		if m.animationFrame >= UserPulseFrames {
			m.animationPhase = PhaseGroupsFlow
			m.animationFrame = 0
			m.flowProgress = 0
		}

	case PhaseGroupsFlow:
		if m.animationFrame >= GroupFlowFrames {
			m.animationFrame = 0
			m.flowProgress++

			// Count groups by type
			groupCount := len(m.groups)

			if m.flowProgress >= groupCount {
				m.animationPhase = PhaseRolesReveal
				m.flowProgress = 0
			}
		}

	case PhaseRolesReveal:
		if m.animationFrame >= RoleRevealFrames {
			m.animationFrame = 0
			m.flowProgress++

			// Count roles by type
			roleCount := len(m.roles)

			if m.flowProgress >= roleCount {
				m.animationPhase = PhaseFadeOut
				m.animationFrame = 0
			}
		}

	case PhaseFadeOut:
		if m.animationFrame >= FadeOutFrames {
			// Loop back to start instead of stopping
			m.animationPhase = PhaseUserPulse
			m.animationFrame = 0
			m.flowProgress = 0
		}
	}

	return animationTick(m.animationID)
}

// IsAnimating returns whether the animation is currently running.
func (m TreeViewModel) IsAnimating() bool {
	return m.animating
}

// getGroupIndex returns the ordinal position of a group node in the visible nodes.
// Returns -1 if the nodeIndex doesn't correspond to a group node.
func (m TreeViewModel) getGroupIndex(nodeIndex int) int {
	count := 0
	for i, node := range m.nodes {
		if _, ok := node.Value.(azure.Group); ok {
			if i == nodeIndex {
				return count
			}
			count++
		}
	}
	return -1
}

// getRoleIndex returns the ordinal position of a role node in the visible nodes.
// Returns -1 if the nodeIndex doesn't correspond to a role node.
func (m TreeViewModel) getRoleIndex(nodeIndex int) int {
	count := 0
	for i, node := range m.nodes {
		if _, ok := node.Value.(azure.Role); ok {
			if i == nodeIndex {
				return count
			}
			count++
		}
	}
	return -1
}
