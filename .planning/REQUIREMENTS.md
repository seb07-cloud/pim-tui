# Requirements: pim-tui

**Defined:** 2026-01-16
**Core Value:** Fast, reliable role activation without leaving the terminal

## v1 Requirements

Requirements for v1.1 Refactor & Reliability milestone. Each maps to roadmap phases.

### Architecture

- [x] **ARCH-01**: All Azure API calls use native REST with `azidentity` (no `az rest` CLI shelling)
- [ ] **ARCH-02**: Codebase follows consistent patterns (simplified, cleaned up)

### Performance

- [ ] **PERF-01**: Subscription fetching uses tenant name cache (one lookup per tenant, not per subscription)
- [ ] **PERF-02**: API responses support pagination for large result sets

### UI

- [ ] **UI-01**: Scrolling within a panel keeps all panels fixed in position (only content scrolls)

### Testing

- [ ] **TEST-01**: Azure client methods have unit test coverage
- [ ] **TEST-02**: UI state transitions have unit test coverage
- [ ] **TEST-03**: Config loading has unit test coverage

### Reliability

- [ ] **REL-01**: Parallel goroutines use proper synchronization (no race conditions)
- [ ] **REL-02**: Group activation uses actual roleDefinitionId from assignment (not hardcoded "member")
- [ ] **REL-03**: All errors are logged (no silent swallowing)
- [ ] **REL-04**: Dead code removed (spinnerPulse function, etc.)

### Robustness

- [ ] **ROB-01**: Application handles SIGINT/SIGTERM gracefully
- [ ] **ROB-02**: Credentials refresh automatically during long sessions
- [ ] **ROB-03**: Justification input is validated (character limits, control characters filtered)

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
| ARCH-02 | Phase 2 | Pending |
| PERF-01 | Phase 3 | Pending |
| PERF-02 | Phase 3 | Pending |
| UI-01 | Phase 4 | Pending |
| TEST-01 | Phase 7 | Pending |
| TEST-02 | Phase 7 | Pending |
| TEST-03 | Phase 7 | Pending |
| REL-01 | Phase 5 | Pending |
| REL-02 | Phase 5 | Pending |
| REL-03 | Phase 5 | Pending |
| REL-04 | Phase 2 | Pending |
| ROB-01 | Phase 6 | Pending |
| ROB-02 | Phase 6 | Pending |
| ROB-03 | Phase 6 | Pending |

**Coverage:**
- v1 requirements: 15 total
- Mapped to phases: 15 ✓
- Unmapped: 0

---
*Requirements defined: 2026-01-16*
*Last updated: 2026-01-16 — Phase mappings added by create-roadmap*
