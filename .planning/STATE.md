# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-01-16)

**Core value:** Fast, reliable role activation without leaving the terminal
**Current focus:** v1.2 UI Polish & Auth UX - Phase 8 ready for planning

## Current Position

Milestone: v1.2 UI Polish & Auth UX
Phase: 8 of 9 (UI Polish)
Plan: 0 of ? in current phase - Not planned
Status: Ready to plan Phase 8
Last activity: 2026-01-16 - Added phases 8-9 from todos

Progress: ░░░░░░░░░░ 0%

## Milestones

### v1.1 Refactor & Reliability (COMPLETE)
- Phases: 1-7
- Plans completed: 13
- Duration: ~49 minutes
- Audit: PASSED (15/15 requirements)

### v1.2 UI Polish & Auth UX (CURRENT)
- Phases: 8-9
- Plans completed: 0
- Status: Planning

## Performance Metrics

**Velocity (v1.1):**
- Total plans completed: 13
- Average duration: 4 min
- Total execution time: 0.82 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-native-rest-migration | 3/3 | 14 min | 5 min |
| 02-codebase-cleanup | 2/2 | 2 min | 1 min |
| 03-performance-optimization | 1/1 | 3 min | 3 min |
| 04-ui-scrolling-fix | 1/1 | 8 min | 8 min |
| 05-reliability-fixes | 1/1 | 3 min | 3 min |
| 06-robustness | 1/1 | 4 min | 4 min |
| 07-test-coverage | 3/3 | 9 min | 3 min |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Native REST over msgraph-sdk-go (avoid new dependency, more control)
- Fix all CONCERNS.md items (clean slate for future development)
- SDK-only path mandatory - removed useAzRest fallback from all files
- ARM API retry pattern: max 3 retries, exponential backoff 1s/2s/4s, only on 429
- AzureCLICredential preferred over DefaultAzureCredential (explicit az login requirement)
- Single credential instance shared for Graph, PIM, and ARM scopes
- Package documentation pattern: Include auth requirements and architecture notes
- Intentionally ignored errors: Document with comment, use named err with nil check
- renderExpiryLine removed as dead code - inline expiry logic is sufficient
- Tenant caching: Separate fetch phases for parallel efficiency, then cache unique tenant names
- Pagination: In-function @odata.nextLink loops rather than centralized helper
- Independent scroll offsets per panel with stored position (04-01)
- Fixed height constraint for subscriptions panel to prevent visual overflow (04-01)
- RoleDefinitionID stored on Group struct from eligibility response (05-01)
- Error logging pattern: log.Printf("[component] context: %v", err) (05-01)
- Context cancellation for graceful shutdown via tea.WithContext (06-01)
- Justification validation rejects ASCII 0-31 (except tab/newline/CR) and DEL (06-01)
- Same-package testing for internal function access (07-01)
- Table-driven tests with descriptive names pattern (07-01)
- testModel helper for creating Model in specific initial state (07-03)
- Update testing via message injection and state assertion (07-03)
- httptest.Server + custom RoundTripper for Azure API mocking (07-02)
- mockCredential for static token in tests (07-02)

### Roadmap Evolution

- Phases 8-9 added: Todos converted to phases for v1.2 milestone (2026-01-16)

### Pending Todos

0 todos - all converted to roadmap phases:
- Fix startup step ordering display → Phase 8 (UI Polish)
- Improve role selection cursor visibility → Phase 8 (UI Polish)
- Smart wrap long permission strings → Phase 8 (UI Polish)
- Add in-app az login → Phase 9 (In-App Authentication)

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-01-16T09:50:00Z
Stopped at: Created phases 8-9 from todos
Resume file: None
