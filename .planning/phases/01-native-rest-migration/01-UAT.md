---
status: diagnosed
phase: 01-native-rest-migration
source: [01-01-SUMMARY.md, 01-02-SUMMARY.md, 01-03-SUMMARY.md]
started: 2026-01-16T06:15:00Z
updated: 2026-01-16T06:22:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Application launches and authenticates
expected: Run the app. It should start and show the main TUI interface without subprocess errors.
result: issue
reported: "yes but not all tenants did get resolved against their tenant names, some work tho"
severity: major

### 2. Subscription list loads
expected: After launch, subscriptions should load and display. No difference from before the migration.
result: pass

### 3. Role assignments display correctly
expected: Navigate to a subscription with PIM roles. Role eligibilities should load and display.
result: pass

## Summary

total: 3
passed: 2
issues: 1
pending: 0
skipped: 0

## Gaps

- truth: "All tenant names resolve correctly"
  status: failed
  reason: "User reported: not all tenants did get resolved against their tenant names, some work tho"
  severity: major
  test: 1
  root_cause: "Pre-existing issue: getTenantNameByID uses Graph API tenantRelationships/findTenantInformationByTenantId which fails for tenants where user lacks cross-tenant query permissions. Not introduced by Phase 1."
  artifacts:
    - path: "internal/azure/lighthouse.go"
      issue: "getTenantNameByID fails silently for inaccessible tenants, falls back to showing tenant ID"
  missing:
    - "Phase 3 (PERF-01) will add tenant name caching to address this"
  debug_session: ""
  note: "This is a pre-existing issue scheduled for Phase 3, not a Phase 1 regression"
