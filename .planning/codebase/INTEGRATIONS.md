# External Integrations

**Analysis Date:** 2025-01-16

## APIs & External Services

### Microsoft Graph API

**Purpose:** User and tenant information retrieval

**Endpoints Used:**
- `https://graph.microsoft.com/v1.0/me` - Current user ID and profile
- `https://graph.microsoft.com/v1.0/organization` - Tenant details
- `https://graph.microsoft.com/v1.0/tenantRelationships/findTenantInformationByTenantId` - Lighthouse tenant names

**Implementation:**
- Client: `internal/azure/client.go`
- Auth: Azure CLI or DefaultAzureCredential
- Scope: `https://graph.microsoft.com/.default`

### Azure PIM Governance API

**Purpose:** Privileged Identity Management for Entra ID roles and groups

**Base URL:** `https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess`

**Endpoints Used:**
- `GET /aadroles/roleAssignments` - List eligible/active Entra ID roles
- `POST /aadroles/roleAssignmentRequests` - Activate/deactivate Entra roles
- `GET /aadGroups/roleAssignments` - List eligible/active PIM groups
- `POST /aadGroups/roleAssignmentRequests` - Activate/deactivate PIM groups
- `GET /aadGroups/resources/{id}` - Get group display name

**Implementation:**
- Client: `internal/azure/pim.go`, `internal/azure/groups.go`
- Auth: Azure CLI or AzureCLICredential
- Scope: `https://api.azrbac.mspim.azure.com/.default`

### Azure Resource Manager (ARM) API

**Purpose:** Azure RBAC PIM for subscriptions (Lighthouse)

**Base URL:** `https://management.azure.com`

**Endpoints Used:**
- `GET /subscriptions/{id}` - Subscription details
- `GET /providers/Microsoft.Authorization/roleEligibilityScheduleInstances` - Eligible Azure RBAC roles
- `GET /providers/Microsoft.Authorization/roleAssignmentScheduleInstances` - Active Azure RBAC roles
- `PUT {scope}/providers/Microsoft.Authorization/roleAssignmentScheduleRequests/{id}` - Activate/deactivate Azure roles

**Implementation:**
- Client: `internal/azure/lighthouse.go`
- Auth: Azure CLI or DefaultAzureCredential
- Scope: `https://management.azure.com/.default`

## Authentication

### Primary: Azure CLI

**Method:** Shell out to `az rest` command

**How it works:**
1. Application first tests if `az rest` works for Graph API
2. If successful, all API calls use `az rest --method <METHOD> --url <URL>`
3. Handles Windows (`cmd /c az`) and Unix paths automatically

**Benefits:**
- Leverages existing Azure CLI authentication
- No token management in application
- Works with MFA, device code, service principals

**Implementation:**
```go
// internal/azure/client.go
func azCommand(args ...string) *exec.Cmd {
    if runtime.GOOS == "windows" {
        cmdArgs := append([]string{"/c", "az"}, args...)
        return exec.Command("cmd", cmdArgs...)
    }
    return exec.Command("az", args...)
}
```

### Fallback: Azure SDK DefaultAzureCredential

**When used:** If `az rest` test fails

**Credential chain:**
1. Environment variables
2. Managed Identity
3. Azure CLI
4. Visual Studio Code
5. Azure PowerShell

**Implementation:**
```go
// internal/azure/client.go
cred, err := azidentity.NewDefaultAzureCredential(nil)
```

### PIM-Specific Credential

**For PIM API calls:** Uses `AzureCLICredential` directly when SDK fallback is active

**Implementation:**
```go
// internal/azure/client.go
cred, err := azidentity.NewAzureCLICredential(nil)
```

## Data Storage

**Databases:**
- None (stateless application)

**File Storage:**
- Config file only: `~/.config/pim-tui/config.yaml`
- No persistent data storage

**Caching:**
- In-memory only (not persisted):
  - User ID cached in `Client.userID`
  - Tenant info cached in `Client.tenant`
  - PIM credential cached in `Client.pimCred`

## Monitoring & Observability

**Error Tracking:**
- None (errors displayed in TUI log panel)

**Logs:**
- In-memory log buffer (last 100 entries)
- Displayed in TUI
- Can be copied to clipboard (`c` key)
- Levels: DEBUG, INFO, ERROR

## Rate Limiting

**Azure API Rate Limiting:**
- Implemented exponential backoff retry (1s, 2s, 4s)
- Handles HTTP 429 "Too Many Requests" responses
- Max 3 retries

**Implementation:**
```go
// internal/azure/client.go
maxRetries := 3
for attempt := 0; attempt <= maxRetries; attempt++ {
    if attempt > 0 {
        time.Sleep(time.Duration(1<<attempt) * time.Second)
    }
    // ... make request ...
}
```

## Parallel Request Handling

**Data fetching uses goroutines with sync.WaitGroup:**
- Roles and groups fetched in parallel
- Eligible and active assignments fetched in parallel
- Group names fetched in parallel
- Subscription tenant details fetched in parallel

**Implementation pattern:**
```go
var wg sync.WaitGroup
wg.Add(2)
go func() {
    defer wg.Done()
    eligible, eligibleErr = c.GetEligibleRoles(ctx)
}()
go func() {
    defer wg.Done()
    active, activeErr = c.GetActiveRoles(ctx)
}()
wg.Wait()
```

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None

## Required Permissions

### Azure AD / Entra ID

For PIM role and group activation:
- User must have eligible PIM role assignments
- No specific application permissions needed (uses delegated auth)

### Azure RBAC (Lighthouse)

For subscription role activation:
- User must have eligible role assignments on subscriptions
- Works with Azure Lighthouse delegated access

## Environment Variables

**Optional (for SDK fallback auth):**
- `AZURE_TENANT_ID` - Azure tenant ID
- `AZURE_CLIENT_ID` - Service principal client ID
- `AZURE_CLIENT_SECRET` - Service principal secret

**Note:** Most users will authenticate via `az login` and no env vars are needed.

## Timeouts

| Operation | Timeout |
|-----------|---------|
| HTTP client | 30 seconds |
| Auth test | 30 seconds |
| Role/Group load | 30 seconds |
| Lighthouse load | 60 seconds |
| User info load | 10 seconds |

## Error Handling

**API errors include:**
- HTTP status code
- Response body for debugging
- Source context (roles, groups, lighthouse, auth)

**User-facing:**
- Troubleshooting tips based on error type
- Retry option on auth failure

---

*Integration audit: 2025-01-16*
