# Requirements: pim-tui

**Defined:** 2026-01-16
**Core Value:** Fast, reliable role activation without leaving the terminal

## v1 Requirements

Requirements for v1.1 Refactor & Reliability milestone. Each maps to roadmap phases.

### Architecture

- [x] **ARCH-01**: All Azure API calls use native REST with `azidentity` (no `az rest` CLI shelling)
- [x] **ARCH-02**: Codebase follows consistent patterns (simplified, cleaned up)

### Performance

- [x] **PERF-01**: Subscription fetching uses tenant name cache (one lookup per tenant, not per subscription)
- [x] **PERF-02**: API responses support pagination for large result sets

### UI

- [x] **UI-01**: Scrolling within a panel keeps all panels fixed in position (only content scrolls)

### Testing

- [x] **TEST-01**: Azure client methods have unit test coverage
- [x] **TEST-02**: UI state transitions have unit test coverage
- [x] **TEST-03**: Config loading has unit test coverage

### Reliability

- [x] **REL-01**: Parallel goroutines use proper synchronization (no race conditions)
- [x] **REL-02**: Group activation uses actual roleDefinitionId from assignment (not hardcoded "member")
- [x] **REL-03**: All errors are logged (no silent swallowing)
- [x] **REL-04**: Dead code removed (spinnerPulse function, etc.)

### Robustness

- [x] **ROB-01**: Application handles SIGINT/SIGTERM gracefully
- [x] **ROB-02**: Credentials refresh automatically during long sessions
- [x] **ROB-03**: Justification input is validated (character limits, control characters filtered)

## v2 Requirements

Deferred to future releases.

(None — this milestone is refactor-focused)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| New features | This is refactor only |
| Offline mode / data caching | Deferred to future milestone |
| Persistent file logging | In-memory logging sufficient |
| GUI version | Terminal-only product |

## Traceability

Which phases cover which requirements. Updated by create-roadmap.

| Requirement | Phase | Status |
|-------------|-------|--------|
| ARCH-01 | Phase 1 | Complete |
| ARCH-02 | Phase 2 | Complete |
| PERF-01 | Phase 3 | Complete |
| PERF-02 | Phase 3 | Complete |
| UI-01 | Phase 4 | Complete |
| TEST-01 | Phase 7 | Complete |
| TEST-02 | Phase 7 | Complete |
| TEST-03 | Phase 7 | Complete |
| REL-01 | Phase 5 | Complete |
| REL-02 | Phase 5 | Complete |
| REL-03 | Phase 5 | Complete |
| REL-04 | Phase 2 | Complete |
| ROB-01 | Phase 6 | Complete |
| ROB-02 | Phase 6 | Complete |
| ROB-03 | Phase 6 | Complete |

**Coverage:**
- v1 requirements: 15 total
- Mapped to phases: 15 ✓
- Unmapped: 0

---
*Requirements defined: 2026-01-16*
*Last updated: 2026-01-16 — Phase mappings added by create-roadmap*
