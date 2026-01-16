# Roadmap: pim-tui

## Overview

This roadmap transforms pim-tui from a working v0.1.0 prototype into a production-quality v1.1 release. The journey starts with architectural foundation (removing CLI shelling), then systematically addresses performance, UI, reliability, robustness, and testing. Each phase builds on the previous, with the native REST migration enabling all subsequent improvements.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Native REST Migration** - Replace az CLI shelling with native azidentity + REST
- [x] **Phase 2: Codebase Cleanup** - Remove dead code, establish consistent patterns
- [ ] **Phase 3: Performance Optimization** - Cache tenant names, add pagination
- [ ] **Phase 4: UI Scrolling Fix** - Panels fixed, content scrolls independently
- [ ] **Phase 5: Reliability Fixes** - Fix race conditions, role lookups, error logging
- [ ] **Phase 6: Robustness** - Graceful shutdown, credential refresh, input validation
- [ ] **Phase 7: Test Coverage** - Unit tests for Azure client, UI, and config

## Phase Details

### Phase 1: Native REST Migration
**Goal**: All Azure API calls use native REST with azidentity (no az CLI shelling)
**Depends on**: Nothing (first phase)
**Requirements**: ARCH-01
**Success Criteria** (what must be TRUE):
  1. Application authenticates using azidentity.AzureCLICredential
  2. All Graph API calls use direct HTTP with Authorization header
  3. All ARM API calls use direct HTTP with Authorization header
  4. No `az rest` or `exec.Command("az")` calls exist in codebase
  5. All existing functionality works identically
**Research**: Unlikely (internal refactoring, SDK fallback path already exists)
**Plans**: 3 plans in 2 waves

Plans:
- [x] 01-01: Simplify NewClient and remove az rest from client.go
- [x] 01-02: Remove az rest from lighthouse.go ARM requests
- [x] 01-03: Clean up imports and verify migration

### Phase 2: Codebase Cleanup
**Goal**: Clean, consistent code patterns with dead code removed
**Depends on**: Phase 1
**Requirements**: ARCH-02, REL-04
**Success Criteria** (what must be TRUE):
  1. Unused functions removed (spinnerPulse, etc.)
  2. Consistent error handling pattern across all files
  3. Consistent naming conventions
  4. No commented-out code blocks
**Research**: Unlikely (internal patterns)
**Plans**: 1 plan in 1 wave

Plans:
- [x] 02-01: Remove dead code and improve error handling consistency
- [x] 02-02: Remove renderExpiryLine dead code (gap closure)

### Phase 3: Performance Optimization
**Goal**: Fast subscription loading with tenant caching and pagination support
**Depends on**: Phase 2
**Requirements**: PERF-01, PERF-02
**Success Criteria** (what must be TRUE):
  1. Tenant names cached (one lookup per tenant ID)
  2. Subscription list loads in <3 seconds for typical user
  3. Pagination handles users with 100+ items gracefully
  4. Progress indicator shows during long operations
**Research**: Unlikely (internal patterns)
**Plans**: 1 plan in 1 wave

Plans:
- [x] 03-01: Cache tenant names and add pagination support

### Phase 4: UI Scrolling Fix
**Goal**: Panels stay fixed in position, only content scrolls within each panel
**Depends on**: Phase 3
**Requirements**: UI-01
**Success Criteria** (what must be TRUE):
  1. Scrolling in Roles panel doesn't move Groups/Lighthouse panels
  2. Each panel maintains independent scroll position
  3. Panel headers remain visible during scroll
  4. Scroll position preserved when switching tabs
**Research**: Unlikely (Bubble Tea patterns exist in codebase)
**Plans**: 1 plan in 1 wave (includes human verification checkpoint)

Plans:
- [ ] 04-01: Add independent scroll offsets per panel with fixed headers

### Phase 5: Reliability Fixes
**Goal**: No race conditions, correct role lookups, proper error logging
**Depends on**: Phase 4
**Requirements**: REL-01, REL-02, REL-03
**Success Criteria** (what must be TRUE):
  1. No data races detected under `-race` flag
  2. Group activation uses actual roleDefinitionId from eligibility response
  3. All API errors logged with context (endpoint, status, body)
  4. User sees error feedback for failed operations
**Research**: Unlikely (internal patterns)
**Plans**: 1 plan in 1 wave

Plans:
- [ ] 05-01: Fix race conditions and roleDefinitionId handling

### Phase 6: Robustness
**Goal**: Graceful handling of signals, credentials, and user input
**Depends on**: Phase 5
**Requirements**: ROB-01, ROB-02, ROB-03
**Success Criteria** (what must be TRUE):
  1. Ctrl+C exits cleanly with cleanup
  2. Token refresh happens before expiry during long sessions
  3. Justification rejects invalid input with user feedback
  4. Application state preserved on graceful exit
**Research**: Unlikely (Go standard patterns)
**Plans**: 1 plan in 1 wave

Plans:
- [ ] 06-01: Add graceful shutdown and input validation

### Phase 7: Test Coverage
**Goal**: Unit tests for critical paths with meaningful coverage
**Depends on**: Phase 6
**Requirements**: TEST-01, TEST-02, TEST-03
**Success Criteria** (what must be TRUE):
  1. Azure client methods have tests with mocked HTTP responses
  2. UI state machine has tests for all valid transitions
  3. Config loading tests cover valid, invalid, and missing files
  4. `go test ./...` passes with >70% coverage on target packages
**Research**: Unlikely (Go testing patterns)
**Plans**: 1 plan in 1 wave

Plans:
- [ ] 07-01: Add unit tests for Azure client and types

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Native REST Migration | 3/3 | Complete | 2026-01-16 |
| 2. Codebase Cleanup | 2/2 | Complete | 2026-01-16 |
| 3. Performance Optimization | 1/1 | Complete | 2026-01-16 |
| 4. UI Scrolling Fix | 0/1 | Not started | - |
| 5. Reliability Fixes | 0/1 | Not started | - |
| 6. Robustness | 0/1 | Not started | - |
| 7. Test Coverage | 0/1 | Not started | - |

---
*Roadmap created: 2026-01-16*
*Milestone: v1.1 Refactor & Reliability*
