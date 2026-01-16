# pim-tui

## What This Is

A terminal UI for Azure Privileged Identity Management (PIM). Allows users to view and activate eligible Entra ID roles, PIM groups, and Azure RBAC roles (including Lighthouse/cross-tenant subscriptions) from a single keyboard-driven interface.

## Core Value

Fast, reliable role activation without leaving the terminal. If activation doesn't work, nothing else matters.

## Requirements

### Validated

<!-- Shipped and confirmed working — inferred from existing codebase -->

- ✓ View eligible and active Entra ID roles — existing
- ✓ View eligible and active PIM groups — existing
- ✓ View eligible and active Azure RBAC roles (Lighthouse subscriptions) — existing
- ✓ Activate/deactivate roles with justification — existing
- ✓ Multi-select activation (batch multiple items) — existing
- ✓ Tab navigation between Roles, Groups, Lighthouse panels — existing
- ✓ Status indicators (active, expiring soon, inactive) — existing
- ✓ Activity log panel with clipboard copy — existing
- ✓ Configurable theme colors — existing
- ✓ Configurable duration presets — existing
- ✓ Cross-platform (Windows, Linux, macOS) — existing

### Active

<!-- Current scope — v1.1 Refactor & Reliability milestone -->

**Architecture:**
- [ ] Remove `az rest` CLI shelling — switch to native REST with `azidentity`
- [ ] Simplify and clean up codebase using consistent patterns

**Performance:**
- [ ] Fix slow subscription fetching — cache tenant names, reduce API calls
- [ ] Add pagination support for users with many roles/groups/subscriptions

**UI:**
- [ ] Fix scrolling behavior — panels stay fixed, only content scrolls

**Reliability:**
- [ ] Add test coverage for Azure client methods
- [ ] Add test coverage for UI state transitions
- [ ] Add test coverage for config loading
- [ ] Fix race condition in parallel goroutines
- [ ] Fix hardcoded "member" roleDefinitionId for groups
- [ ] Add proper error logging (no silent swallowing)
- [ ] Remove dead code (unused spinnerPulse function)

**Robustness:**
- [ ] Add graceful shutdown handling
- [ ] Add credential refresh for long sessions
- [ ] Implement proper input validation on justification

### Out of Scope

<!-- Explicit boundaries for this milestone -->

- New features beyond current functionality — this is refactor only
- Offline mode / caching of previous data — deferred to future
- Persistent logging to file — current in-memory is sufficient
- GUI version — terminal-only

## Context

**Current state:**
- Working v0.1.0 with all core features functional
- Architecture: Bubble Tea (Elm architecture) with `internal/ui` and `internal/azure` layers
- Auth: Currently shells out to `az rest` CLI command, falls back to Azure SDK
- Performance: Subscription loading is slow due to per-subscription tenant name lookups
- UI bug: Scrolling in one panel causes all panels to shift position
- Technical debt: Zero test coverage, large monolithic files, some race conditions

**Codebase analysis:**
- See `.planning/codebase/` for detailed architecture, structure, and concerns documentation
- Key files: `internal/ui/model.go` (1150 lines), `internal/ui/views.go` (1519 lines)
- 15+ items identified in CONCERNS.md requiring attention

## Constraints

- **Tech stack**: Go with Bubble Tea/Lipgloss — no change
- **Auth**: Must use Azure CLI credentials (`azidentity.AzureCLICredential`) — users already have `az login`
- **Compatibility**: Must maintain same keyboard shortcuts and workflow
- **Dependencies**: Prefer native REST over adding msgraph-sdk-go dependency

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Native REST over msgraph-sdk-go | Already have azidentity, avoid new dependency, more control | — Pending |
| Fix all CONCERNS.md items | Technical debt compounds; clean slate for future development | — Pending |
| Panels fixed, content scrolls | Consistent with standard TUI patterns, less visual noise | — Pending |

---
*Last updated: 2026-01-16 after initialization*
