# Testing Patterns

**Analysis Date:** 2026-01-16

## Test Framework

**Runner:**
- Not configured - no test files exist in codebase
- No `*_test.go` files detected

**Assertion Library:**
- Not configured

**Run Commands:**
```bash
go test ./...              # Would run all tests
go test -v ./...           # Verbose output
go test -cover ./...       # Coverage report
```

## Test File Organization

**Location:**
- No tests currently exist
- Go convention: co-located with source files (e.g., `client_test.go` alongside `client.go`)

**Naming:**
- Go convention: `*_test.go` suffix

**Recommended Structure:**
```
internal/
├── azure/
│   ├── client.go
│   ├── client_test.go     # (missing)
│   ├── pim.go
│   ├── pim_test.go        # (missing)
│   └── ...
├── config/
│   ├── config.go
│   └── config_test.go     # (missing)
└── ui/
    ├── model.go
    └── model_test.go      # (missing)
```

## Test Structure

**Recommended Pattern for This Codebase:**
```go
func TestGetRoles_Success(t *testing.T) {
    // Arrange
    client := setupMockClient(t)

    // Act
    roles, err := client.GetRoles(context.Background())

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(roles) == 0 {
        t.Error("expected roles, got none")
    }
}
```

**Table-Driven Tests (Recommended):**
```go
func TestStatusFromExpiry(t *testing.T) {
    tests := []struct {
        name     string
        expiry   *time.Time
        expected ActivationStatus
    }{
        {"nil expiry", nil, StatusInactive},
        {"far future", timePtr(time.Now().Add(2*time.Hour)), StatusActive},
        {"expiring soon", timePtr(time.Now().Add(20*time.Minute)), StatusExpiringSoon},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := StatusFromExpiry(tc.expiry)
            if got != tc.expected {
                t.Errorf("got %v, want %v", got, tc.expected)
            }
        })
    }
}
```

## Mocking

**Framework:**
- Not established - no mocking in use

**Recommended Approach:**
- Interface-based mocking for Azure API calls
- HTTP test server for API response mocking

**Mock Pattern for Azure Client:**
```go
// Define interface for testability
type AzureAPI interface {
    GetRoles(ctx context.Context) ([]Role, error)
    ActivateRole(ctx context.Context, roleDefID, scopeID, justification string, duration time.Duration) error
}

// Mock implementation
type MockAzureClient struct {
    GetRolesFunc     func(ctx context.Context) ([]Role, error)
    ActivateRoleFunc func(ctx context.Context, roleDefID, scopeID, justification string, duration time.Duration) error
}

func (m *MockAzureClient) GetRoles(ctx context.Context) ([]Role, error) {
    return m.GetRolesFunc(ctx)
}
```

**What to Mock:**
- Azure API responses (`internal/azure/`)
- HTTP client calls
- Time-dependent functions

**What NOT to Mock:**
- Pure functions (e.g., `StatusFromExpiry`, `formatDuration`)
- Configuration loading (use temp files)
- Lipgloss styling (test output strings instead)

## Fixtures and Factories

**Test Data:**
- Not established - should create `testdata/` directories

**Recommended Location:**
- `internal/azure/testdata/` for API response fixtures
- `internal/config/testdata/` for config file samples

**Factory Pattern Example:**
```go
func testRole(overrides ...func(*Role)) Role {
    r := Role{
        ID:               "test-id",
        DisplayName:      "Test Role",
        RoleDefinitionID: "def-123",
        Status:           StatusInactive,
    }
    for _, override := range overrides {
        override(&r)
    }
    return r
}

// Usage
role := testRole(func(r *Role) {
    r.Status = StatusActive
})
```

## Coverage

**Requirements:** Not enforced

**View Coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Types

**Unit Tests:**
- Target: Pure functions in `internal/azure/types.go`
- Target: Configuration parsing in `internal/config/config.go`
- Target: Helper functions in `internal/ui/views.go`

**Integration Tests:**
- Target: Azure API client with mock HTTP server
- Target: Full Model update cycles in Bubbletea

**E2E Tests:**
- Not applicable for TUI application
- Manual testing via running the application

## Testable Components

**Easy to Test (pure functions):**
- `StatusFromExpiry()` in `internal/azure/types.go`
- `formatDuration()` in `internal/ui/views.go`
- `formatCompactDuration()` in `internal/ui/views.go`
- `truncate()` in `internal/ui/views.go`
- `clampCursor()` in `internal/ui/model.go`
- `indexOf()` in `internal/ui/model.go`
- `parseLogLevel()` in `internal/ui/model.go`
- `GetRolePermissions()` in `internal/ui/roles_builtin.go`

**Requires Mocking:**
- `Client.GetRoles()` - needs mock HTTP responses
- `Client.ActivateRole()` - needs mock HTTP responses
- `Model.Update()` - needs message simulation

**Difficult to Test:**
- `View()` methods - output contains ANSI codes
- Interactive keyboard handling

## Common Patterns

**Context with Timeout:**
```go
func TestGetRoles_Timeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    // Slow mock
    client := &MockClient{
        GetRolesFunc: func(ctx context.Context) ([]Role, error) {
            time.Sleep(100 * time.Millisecond)
            return nil, nil
        },
    }

    _, err := client.GetRoles(ctx)
    if err == nil {
        t.Error("expected timeout error")
    }
}
```

**Error Testing:**
```go
func TestNewClient_AuthFailure(t *testing.T) {
    // Setup environment to fail auth
    t.Setenv("AZURE_TENANT_ID", "invalid")

    _, err := NewClient()

    if err == nil {
        t.Error("expected error, got nil")
    }
    if !strings.Contains(err.Error(), "authentication") {
        t.Errorf("unexpected error: %v", err)
    }
}
```

## Priority Test Targets

**High Priority (critical functionality):**
1. `internal/azure/types.go` - Status calculation logic
2. `internal/config/config.go` - Configuration loading and defaults
3. `internal/ui/model.go` - Cursor movement and selection logic

**Medium Priority:**
4. `internal/azure/client.go` - API request building
5. `internal/ui/views.go` - Duration formatting functions

**Lower Priority:**
6. `internal/ui/styles.go` - Style application
7. `internal/ui/roles_builtin.go` - Static permission mapping

---

*Testing analysis: 2026-01-16*
