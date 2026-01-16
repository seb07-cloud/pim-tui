# pim-tui

## What This Is

A terminal UI for Azure Privileged Identity Management (PIM). Allows users to view and activate eligible Entra ID roles, PIM groups, and Azure RBAC roles (including Lighthouse/cross-tenant subscriptions) from a single keyboard-driven interface. Now with in-app browser authentication.

## Core Value

Fast, reliable role activation without leaving the terminal. If activation doesn't work, nothing else matters.

## Requirements

### Validated

<!-- Shipped and confirmed working -->

**Original (v0.1):**
- ✓ View eligible and active Entra ID roles — v0.1
- ✓ View eligible and active PIM groups — v0.1
- ✓ View eligible and active Azure RBAC roles (Lighthouse subscriptions) — v0.1
- ✓ Activate/deactivate roles with justification — v0.1
- ✓ Multi-select activation (batch multiple items) — v0.1
- ✓ Tab navigation between Roles, Groups, Lighthouse panels — v0.1
- ✓ Status indicators (active, expiring soon, inactive) — v0.1
- ✓ Activity log panel with clipboard copy — v0.1
- ✓ Configurable theme colors — v0.1
- ✓ Configurable duration presets — v0.1
- ✓ Cross-platform (Windows, Linux, macOS) — v0.1

**v1.1 Refactor & Reliability:**
- ✓ Native REST with azidentity (no `az rest` CLI) — v1.1
- ✓ Clean, consistent code patterns — v1.1
- ✓ Tenant name caching for performance — v1.1
- ✓ Pagination for large result sets — v1.1
- ✓ Panels fixed, content scrolls independently — v1.1
- ✓ Unit tests for Azure client, UI, config — v1.1
- ✓ Race condition fixes — v1.1
- ✓ RoleDefinitionId from eligibility response — v1.1
- ✓ Proper error logging — v1.1
- ✓ Dead code removed — v1.1
- ✓ Graceful shutdown handling — v1.1
- ✓ Credential refresh for long sessions — v1.1
- ✓ Justification input validation — v1.1

**v1.2 UI Polish & Auth UX:**
- ✓ Deterministic startup step ordering — v1.2
- ✓ High-contrast cursor visibility — v1.2
- ✓ Permission string wrapping at path segments — v1.2
- ✓ In-app browser authentication — v1.2

### Active

<!-- Next milestone scope -->

(None defined — v1.2 milestone complete, planning next milestone)

### Out of Scope

<!-- Explicit boundaries -->

- Offline mode / caching of previous data — deferred
- Persistent logging to file — in-memory sufficient
- GUI version — terminal-only product
- Device code authentication — blocked by security policies, browser auth used instead

## Context

**Current state:**
- Shipped v1.2 with 7,313 LOC Go
- Architecture: Bubble Tea (Elm architecture) with `internal/ui` and `internal/azure` layers
- Auth: Native azidentity with AzureCLICredential or InteractiveBrowserCredential
- Performance: Tenant names cached, pagination supported
- UI: Panels fixed, high-contrast cursor, permission wrapping
- Test coverage: Azure client HTTP mocking, UI state transitions, config loading

**Tech stack:**
- Go 1.25+
- Bubble Tea / Lipgloss for TUI
- Azure SDK for Go (azcore, azidentity)

## Constraints

- **Tech stack**: Go with Bubble Tea/Lipgloss — no change
- **Auth**: AzureCLICredential (primary) or InteractiveBrowserCredential (in-app)
- **Compatibility**: Maintain same keyboard shortcuts and workflow
- **Dependencies**: Native REST preferred over msgraph-sdk-go

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Native REST over msgraph-sdk-go | Already have azidentity, avoid new dependency, more control | ✓ Good |
| Fix all CONCERNS.md items | Technical debt compounds; clean slate for future development | ✓ Good |
| Panels fixed, content scrolls | Consistent with standard TUI patterns, less visual noise | ✓ Good |
| Browser auth over device code | Device code often blocked by security policies | ✓ Good |
| High-contrast cursor (white/black) | Maximum visibility on any terminal theme | ✓ Good |
| Path-segment wrapping for permissions | Readable long strings in detail panel | ✓ Good |
| ANSI clear screen for auth states | Clean full-screen rendering during auth | ✓ Good |

---
*Last updated: 2026-01-16 after v1.2 milestone*
