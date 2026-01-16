# Codebase Concerns

**Analysis Date:** 2026-01-16

## Tech Debt

**Zero Test Coverage:**
- Issue: No test files exist in the entire codebase (confirmed via glob search for `*_test.go`)
- Files: All files in `internal/azure/`, `internal/ui/`, `internal/config/`
- Impact: Any refactoring or bug fix risks introducing regressions. Critical security code (authentication, role activation) is untested.
- Fix approach: Add unit tests starting with `internal/azure/` package. Mock HTTP responses for API tests. Use table-driven tests for status calculations in `internal/azure/types.go`.

**Large Monolithic View File:**
- Issue: `internal/ui/views.go` is 1519 lines - contains all rendering logic in one file
- Files: `internal/ui/views.go`
- Impact: Hard to navigate, maintain, and test. Adding new views increases file size further.
- Fix approach: Split into separate files by view type: `views_loading.go`, `views_main.go`, `views_dialogs.go`, `views_help.go`. Each render method can be in its own file.

**Large Model File:**
- Issue: `internal/ui/model.go` is 1150 lines - contains all state management and update logic
- Files: `internal/ui/model.go`
- Impact: Difficult to understand state transitions. Event handling is one giant switch statement.
- Fix approach: Extract key handlers into separate files: `handlers_navigation.go`, `handlers_activation.go`, `handlers_search.go`. Consider command pattern for complex operations.

**Hardcoded Group Role Definition:**
- Issue: Group activation hardcodes `"member"` as roleDefinitionId
- Files: `internal/azure/groups.go:201`, `internal/azure/groups.go:227`
- Impact: May fail for groups with "Owner" role type or other custom roles
- Fix approach: Fetch the actual roleDefinitionId from the group's eligible assignment data and store it in the Group struct.

**Silent Error Swallowing:**
- Issue: Some errors are silently ignored without logging
- Files: `internal/azure/groups.go:74` (getGroupName error ignored), `internal/azure/lighthouse.go:317-348` (active assignments query errors silently ignored)
- Impact: Debugging production issues becomes difficult. Users may not know why data is incomplete.
- Fix approach: At minimum, log errors at debug level. Consider propagating partial success with warnings.

## Known Bugs

**Potential Race Condition in Goroutines:**
- Symptoms: Concurrent goroutines access shared error variables without synchronization
- Files: `internal/azure/pim.go:110-122`, `internal/azure/groups.go:156-170`
- Trigger: Parallel fetch of eligible and active roles/groups assigns to `eligibleErr` and `activeErr` without mutex
- Workaround: Currently works because each goroutine writes to a different variable, but this is fragile

**Unused spinnerPulse Function:**
- Symptoms: Dead code that is never called
- Files: `internal/ui/views.go:1409-1414`
- Trigger: N/A - dead code
- Workaround: None needed, but should be removed or used

## Security Considerations

**Token Handling via Azure CLI:**
- Risk: Application shells out to `az rest` command, which may expose tokens in process arguments
- Files: `internal/azure/client.go:152-198`
- Current mitigation: Uses `az rest` which handles tokens internally
- Recommendations: Consider using SDK credential flow exclusively where possible. Audit process listing exposure on shared systems.

**Justification Stored in Memory:**
- Risk: Justification text persists in activation history in memory
- Files: `internal/ui/model.go:62-69`, `internal/ui/model.go:1017-1037`
- Current mitigation: Memory is cleared on application exit
- Recommendations: Consider not storing justification text in history, or truncating it

**No Input Sanitization on Justification:**
- Risk: Justification text is sent directly to Azure API without sanitization
- Files: `internal/ui/model.go:579`, `internal/azure/pim.go:161`
- Current mitigation: Azure API likely sanitizes, but no client-side validation
- Recommendations: Add character limit validation (already limited to 500), filter control characters

**External Command Execution:**
- Risk: Executes Azure CLI commands which could be compromised if PATH is manipulated
- Files: `internal/azure/client.go:38-64`
- Current mitigation: Tries to find `az` in well-known paths on non-Windows systems
- Recommendations: Consider validating `az` binary checksum or using SDK exclusively

## Performance Bottlenecks

**Sequential Activation of Multiple Items:**
- Problem: Activating multiple roles/groups happens sequentially in a loop
- Files: `internal/ui/model.go:1039-1058`
- Cause: Each activation waits for previous to complete before starting next
- Improvement path: Use errgroup to parallelize activations with controlled concurrency

**Unbounded Tenant Name Lookups:**
- Problem: For each subscription, makes a separate API call to fetch tenant display name
- Files: `internal/azure/lighthouse.go:271-307`
- Cause: No caching of tenant names across subscriptions from same tenant
- Improvement path: Build tenant name cache before subscription loop, reuse for all subscriptions with same tenant ID

**No Pagination Handling:**
- Problem: API responses assume all data fits in single response
- Files: `internal/azure/pim.go:47`, `internal/azure/groups.go:48`, `internal/azure/lighthouse.go:202`
- Cause: No @odata.nextLink handling in response parsing
- Improvement path: Add pagination support for users with many roles/groups/subscriptions

## Fragile Areas

**UI Rendering Width Calculations:**
- Files: `internal/ui/views.go:178-194`, `internal/ui/views.go:311-313`, `internal/ui/views.go:536-538`
- Why fragile: Multiple hardcoded width ratios (9/20, 45%) that must stay in sync. Magic numbers for minimum widths.
- Safe modification: Add named constants for layout ratios. Consider a layout configuration struct.
- Test coverage: None - UI rendering is completely untested

**State Machine Transitions:**
- Files: `internal/ui/model.go:353-486`, `internal/ui/model.go:512-762`
- Why fragile: State transitions scattered across large switch statements. Easy to miss a transition or create invalid state.
- Safe modification: Document valid state transitions. Consider using a state machine library or explicit transition table.
- Test coverage: None - state transitions are untested

**Azure API Response Parsing:**
- Files: `internal/azure/pim.go:13-32`, `internal/azure/groups.go:13-37`, `internal/azure/lighthouse.go:29-95`
- Why fragile: Deeply nested JSON structures. API changes could silently break parsing.
- Safe modification: Add response validation. Use generated types from API spec if available.
- Test coverage: None - API responses should be tested with fixtures

## Scaling Limits

**In-Memory Log Buffer:**
- Current capacity: Last 100 log entries
- Limit: `internal/ui/model.go:348-350` - hard limit at 100 entries
- Scaling path: Consider log rotation or persistent logging for long-running sessions

**Activation History:**
- Current capacity: Unbounded - grows with every activation
- Limit: `internal/ui/model.go:154` - no limit, could consume memory over time
- Scaling path: Add max history size with rotation, or persist to file

## Dependencies at Risk

**Go Version:**
- Risk: Uses `go 1.25.5` which is a future version (current stable is ~1.22)
- Impact: May have been a typo; could cause build issues
- Migration plan: Verify and downgrade to latest stable Go version

**Azure SDK:**
- Risk: Using azidentity v1.13.1 and azcore v1.21.0 - need to track security updates
- Impact: Authentication vulnerabilities could affect token handling
- Migration plan: Implement dependabot or similar for dependency updates

## Missing Critical Features

**No Graceful Shutdown:**
- Problem: No handling of SIGINT/SIGTERM for cleanup
- Blocks: Users may lose in-progress operations if terminal is closed

**No Credential Refresh:**
- Problem: If Azure CLI credentials expire during long session, operations will fail
- Blocks: Long-running sessions require manual restart

**No Offline Mode:**
- Problem: Application requires network access; no caching of previously loaded data
- Blocks: Cannot view previously activated roles when offline

## Test Coverage Gaps

**All Azure Client Methods:**
- What's not tested: Every method in `internal/azure/` package
- Files: `internal/azure/client.go`, `internal/azure/pim.go`, `internal/azure/groups.go`, `internal/azure/lighthouse.go`
- Risk: Authentication failures, API changes, edge cases in role activation all undetected
- Priority: High - these are the core security-sensitive operations

**UI State Transitions:**
- What's not tested: State machine transitions (Loading -> Normal -> Confirm -> Activating, etc.)
- Files: `internal/ui/model.go`
- Risk: Invalid state combinations, stuck states, missed transitions
- Priority: High - affects user experience and correctness

**Configuration Loading:**
- What's not tested: Config parsing, defaults, theme loading
- Files: `internal/config/config.go`
- Risk: Invalid config could crash application
- Priority: Medium - affects startup behavior

**Duration/Time Formatting:**
- What's not tested: `formatDuration`, `formatCompactDuration`, `StatusFromExpiry`
- Files: `internal/ui/views.go:1381-1394`, `internal/ui/views.go:1509-1519`, `internal/azure/types.go:33-41`
- Risk: Edge cases (negative durations, nil times, timezone issues)
- Priority: Medium - affects display correctness

---

*Concerns audit: 2026-01-16*
