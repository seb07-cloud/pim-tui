---
phase: 01-native-rest-migration
plan: 02
subsystem: api
tags: [azure, arm, rest, authentication, sdk, retry]

# Dependency graph
requires:
  - phase: none
    provides: none
provides:
  - SDK-only ARM request functions in lighthouse.go
  - Retry logic with exponential backoff for 429 responses
affects: [01-03-cleanup, future ARM API calls]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "SDK token acquisition for ARM API (management.azure.com)"
    - "Retry with exponential backoff (1s, 2s, 4s) for 429 rate limiting"

key-files:
  created: []
  modified:
    - internal/azure/lighthouse.go

key-decisions:
  - "Removed useAzRest fallback - SDK-only path is now mandatory"
  - "Added retry logic to armRequest and armRequestWithBody (max 3 retries)"

patterns-established:
  - "ARM API retry pattern: max 3 retries, exponential backoff 1s/2s/4s, only on 429"

# Metrics
duration: 8min
completed: 2026-01-16
---

# Phase 01 Plan 02: ARM Request SDK Migration Summary

**SDK-only ARM request functions with 429 retry logic in lighthouse.go**

## Performance

- **Duration:** 8 min
- **Started:** 2026-01-16T05:46:00Z
- **Completed:** 2026-01-16T05:54:42Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments

- Removed `useAzRest` branch from `armRequest()` function
- Removed `useAzRest` branch from `armRequestWithBody()` function
- Added retry logic with exponential backoff for 429 Too Many Requests responses
- Both functions now use `cred.GetToken()` with ARM scope exclusively

## Task Commits

Each task was committed atomically:

1. **Task 1 + Task 2: Simplify ARM request functions** - `c4258b8` (feat)
   - Combined commit since both tasks modify the same file for the same purpose

**Plan metadata:** (to be added after summary creation)

## Files Created/Modified

- `internal/azure/lighthouse.go` - ARM request functions simplified to SDK-only path with retry logic

## Decisions Made

- **SDK-only path mandatory**: Removed the `useAzRest` fallback completely. The application now requires Azure CLI credentials to be available via the SDK.
- **Retry pattern for ARM API**: Added the same retry pattern used in the az rest functions - max 3 retries with exponential backoff (1s, 2s, 4s) on 429 status codes only.

## Deviations from Plan

None - plan executed as written.

**Note on build status**: The `go build ./...` verification step will fail until plan 01-01 is also completed. This is expected because:
- Plan 01-01 removes the `useAzRest` field from Client struct
- Plan 01-02 removes references to `useAzRest` from lighthouse.go
- Both plans are in wave 1 and can run in parallel
- The build passes only after all wave 1 plans complete

## Issues Encountered

- **Pre-existing uncommitted changes**: The lighthouse.go file already had significant uncommitted changes in the working tree from a previous session. These changes included a more complete implementation with Azure role eligibility and activation support. The SDK migration was applied on top of this existing work.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- lighthouse.go ARM functions ready for use
- Awaiting plan 01-01 completion for full build verification
- Plan 01-03 (wave 2) can proceed once both wave 1 plans complete

---
*Phase: 01-native-rest-migration*
*Completed: 2026-01-16*
