---
name: microsoft-graph-go-backend
description: >-
  Develops backend services using Microsoft Graph API with Go standard library, native REST calls and msgraph-sdk-go. Enforces Google Go style, production-grade performance,
  and complete implementations. Use when building M365 integrations, Entra ID
  automation, tenant management, Lighthouse/cross-tenant subscriptions, Azure PIM,
  or when user mentions Graph API, Azure AD, ARM API, delegated access,
  users, groups, Teams, SharePoint, OneDrive, Exchange programmatic access.
---

# Graph API Backend Development (Go)

## Enforced Standards

These are non-negotiable for all code produced:

**Quality:** Google Go Style Guide (https://google.github.io/styleguide/go/). Clarity over cleverness. Every error handled. All exports documented. Table-driven tests.

**Performance:** Single reusable `http.Client`. Context with timeout on every call. Batch operations. Stream responses. Delta queries over full syncs.

**Completeness:** No shortcuts. Wrap errors with context. Retry with backoff. Validate inputs. Structured logging. Never log secrets.

## Graph API Reference

**Base URL:** `https://graph.microsoft.com/v1.0`

### Token Request (Client Credentials)

```
POST https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token
Content-Type: application/x-www-form-urlencoded

client_id={clientID}&client_secret={secret}&scope=https://graph.microsoft.com/.default&grant_type=client_credentials
```

Note: `.default` suffix on scope is required — `https://graph.microsoft.com/` alone fails silently.

### Request Headers

```
Authorization: Bearer {token}
Content-Type: application/json
ConsistencyLevel: eventual    # Required for $count, $search, advanced $filter
```

### Batch Request Structure

```json
POST /$batch

{
  "requests": [
    {"id": "1", "method": "GET", "url": "/users/user-id-1"},
    {"id": "2", "method": "GET", "url": "/users/user-id-2"}
  ]
}
```

Max 20 requests. Responses keyed by `id`.

### Delta Query Pattern

Initial: `GET /users/delta` → returns `@odata.deltaLink`

Subsequent: `GET {deltaLink}` → returns changes since last call

Store `deltaLink` persistently between syncs.

### Error Response

```json
{
  "error": {
    "code": "Request_ResourceNotFound",
    "message": "...",
    "innerError": {
      "request-id": "abc-123",  // Log this for support escalation
      "date": "2025-01-16T..."
    }
  }
}
```

**Rate limits:** 429 returns `Retry-After` header in seconds.

## Permissions

Request minimum necessary. Common application permissions:
- Users: `User.Read.All` / `User.ReadWrite.All`
- Groups: `Group.Read.All` / `Group.ReadWrite.All`
- Mail: `Mail.Send`
- Files: `Files.ReadWrite.All`
- Teams: `ChannelMessage.Send`

## Azure API Landscape

Three distinct APIs serve different purposes:

| API | Base URL | Scope | Purpose |
|-----|----------|-------|---------|
| **Graph API** | `graph.microsoft.com/v1.0` | `graph.microsoft.com/.default` | Directory: users, groups, tenants, M365 services |
| **ARM API** | `management.azure.com` | `management.azure.com/.default` | Azure Resources: subscriptions, RBAC, PIM for Azure roles |
| **PIM Governance API** | `api.azrbac.mspim.azure.com` | `api.azrbac.mspim.azure.com/.default` | Entra ID PIM: directory roles, privileged group membership |

### When to Use Which

- **Entra ID roles** (Global Admin, User Admin, etc.) → PIM Governance API
- **Azure RBAC roles** (Contributor, Owner on subscriptions) → ARM API
- **User/group info, tenant names, M365 data** → Graph API

### PIM Governance API (Entra ID Roles & Groups)

**Base URL:** `https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess`

```
GET /aadRoles/resources/{tenantId}/roleAssignments?$filter=subjectId eq '{userId}'
GET /aadGroups/resources/{groupId}/roleAssignments?$filter=subjectId eq '{userId}'
```

Use for activating Entra ID directory roles and privileged group memberships.

## Cross-Tenant Tenant Resolution

To get tenant display names for external/Lighthouse tenants (not your own), use:

```
GET https://graph.microsoft.com/v1.0/tenantRelationships/findTenantInformationByTenantId(tenantId='{tenantId}')
```

Response: `{"tenantId": "...", "displayName": "Customer Name", "defaultDomainName": "customer.onmicrosoft.com"}`

**Note:** `/organization` only returns YOUR tenant. Use `findTenantInformationByTenantId` for any external tenant ID.

## ARM API Reference (Azure RBAC PIM)

**Base URL:** `https://management.azure.com`

**Scope:** `https://management.azure.com/.default`

Use ARM API for Azure resource roles (Contributor, Owner, Reader on subscriptions/resource groups). NOT for Entra ID directory roles.

### Subscription Details

```
GET https://management.azure.com/subscriptions/{subscriptionId}?api-version=2022-12-01
```

Returns `homeTenantId` (customer tenant for Lighthouse) and `tenantId`. Use `homeTenantId` when available.

### PIM Role Eligibility (Azure RBAC)

```
GET https://management.azure.com/providers/Microsoft.Authorization/roleEligibilityScheduleInstances?api-version=2020-10-01&$filter=asTarget()
```

Returns all eligible PIM role assignments for the current user across all subscriptions.

### PIM Role Activation

```
PUT https://management.azure.com{scope}/providers/Microsoft.Authorization/roleAssignmentScheduleRequests/{requestId}?api-version=2020-10-01

{
  "properties": {
    "principalId": "{userId}",
    "roleDefinitionId": "{roleDefId}",
    "requestType": "SelfActivate",
    "justification": "...",
    "scheduleInfo": {
      "startDateTime": "2025-01-16T...",
      "expiration": {"type": "AfterDuration", "duration": "PT8H"}
    }
  }
}
```

### Active Role Assignments

```
GET https://management.azure.com/providers/Microsoft.Authorization/roleAssignmentScheduleInstances?api-version=2020-10-01&$filter=asTarget()
```

Filter by `assignmentType == "Activated"` to find PIM-activated roles (vs permanent).

## External References

- Graph API reference: https://learn.microsoft.com/en-us/graph/api/overview
- ARM API reference: https://learn.microsoft.com/en-us/rest/api/azure/
- PIM for Azure Resources: https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-resource-roles-overview
- Graph Explorer: https://developer.microsoft.com/graph/graph-explorer
- Go style: https://google.github.io/styleguide/go/
- Go Graph SDK: https://github.com/microsoftgraph/msgraph-sdk-go
- MS Graph SDKs: https://learn.microsoft.com/en-us/graph/sdks/sdks-overview