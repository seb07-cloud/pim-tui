---
phase: 01-native-rest-migration
plan: 01
subsystem: api
tags: [azidentity, azure-cli, authentication, http-client]

# Dependency graph
requires: []
provides:
  - SDK-only authentication via AzureCLICredential
  - Simplified Client struct without useAzRest flag
  - graphRequest with retry logic for 429 responses
  - pimRequest with retry logic for 429 responses
affects: [01-02, 01-03, testing, future-mock-http]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - AzureCLICredential for all Azure authentication
    - Exponential backoff retry (1s, 2s, 4s) for 429 rate limiting
    - Single credential shared between Graph and PIM APIs

key-files:
  created: []
  modified:
    - internal/azure/client.go

key-decisions:
  - "Use AzureCLICredential instead of DefaultAzureCredential for explicit az login dependency"
  - "Share single credential instance between cred and pimCred (same token works for different scopes)"
  - "Implement retry with exponential backoff inline in each request method"

patterns-established:
  - "SDK token acquisition: cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{scope}})"
  - "Retry pattern: max 3 retries with 1<<attempt second delays for 429 responses"
  - "Body reader reset: Re-marshal JSON body before each retry to reset io.Reader position"

# Metrics
duration: 4min
completed: 2026-01-16
---

# Phase 1 Plan 1: Remove az CLI Shelling Summary

**Native azidentity SDK authentication replacing all az CLI subprocess execution in client.go**

## Performance

- **Duration:** 4 min
- **Started:** 2026-01-16T05:51:55Z
- **Completed:** 2026-01-16T05:56:05Z
- **Tasks:** 3
- **Files modified:** 1

## Accomplishments

- Removed all az CLI subprocess execution (exec.Command, azCommand function)
- Replaced DefaultAzureCredential with AzureCLICredential for explicit az login dependency
- Removed useAzRest field and conditional branching from Client struct
- Added retry logic with exponential backoff for 429 rate limiting to graphRequest and pimRequest
- Simplified NewClient to initialize both credentials upfront (no lazy initialization)

## Task Commits

All three tasks were completed in a single atomic commit since they all modified the same file with interdependent changes:

1. **Task 1: Simplify NewClient to SDK-only authentication** - `59d61ec` (refactor)
2. **Task 2: Remove az rest request functions and simplify graphRequest** - `59d61ec` (refactor)
3. **Task 3: Simplify pimRequest to SDK-only path** - `59d61ec` (refactor)

_Note: Tasks were combined into single commit because all changes were to client.go and removing az CLI code required simultaneous updates to NewClient, graphRequest, and pimRequest._

## Files Created/Modified

- `internal/azure/client.go` - Simplified Azure client with SDK-only authentication

### Code Removed (Lines of Code)

- `azCommand()` function (~27 lines) - Cross-platform az CLI path detection
- `azRestRequest()` function (~3 lines) - az rest wrapper
- `azRestRequestWithRetry()` function (~40 lines) - az rest with retry logic
- `azRestRequestWithResource()` function (~3 lines) - az rest with custom resource
- `useAzRest` field and all conditional branches (~15 lines)
- Lazy initialization of pimCred (~7 lines)
- Unused imports: `os/exec`, `runtime`, `strings`

### Code Added

- Retry logic in graphRequest (~20 lines)
- Retry logic in pimRequest (~20 lines)

**Net reduction:** ~55 lines removed

## Decisions Made

1. **AzureCLICredential over DefaultAzureCredential** - More explicit about requiring `az login`, matches user expectation that Azure CLI must be authenticated
2. **Single credential for both cred and pimCred** - Same AzureCLICredential works for Graph, PIM, and ARM scopes; no need for separate instances
3. **Inline retry logic** - Each request method has its own retry loop rather than shared helper, keeps code locality and clarity

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all changes compiled and verified on first attempt.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 01-02 (lighthouse.go ARM requests) already completed in parallel (commit c4258b8)
- Plan 01-03 (pim.go Entra ID role requests) ready to execute
- All az CLI code removed from client.go; lighthouse.go and pim.go can follow same pattern
- Build passes with all changes

---
*Phase: 01-native-rest-migration*
*Completed: 2026-01-16*
