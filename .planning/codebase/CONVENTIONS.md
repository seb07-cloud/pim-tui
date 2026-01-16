# Coding Conventions

**Analysis Date:** 2026-01-16

## Naming Patterns

**Files:**
- Snake case for Go files: `roles_builtin.go`, not `rolesBuiltin.go`
- Package-level organization: one primary concept per file
- Example: `internal/azure/pim.go`, `internal/azure/groups.go`, `internal/azure/lighthouse.go`

**Functions:**
- CamelCase with exported functions PascalCase
- Verb-noun pattern: `GetRoles`, `ActivateRole`, `DeactivateGroup`
- Private functions lowercase: `loadTenantCmd`, `getGroupName`
- Command functions suffix: `loadRolesCmd`, `tickCmd`, `initClientCmd`

**Variables:**
- camelCase for local variables: `userID`, `roleDefinitionID`, `justification`
- Short names acceptable in tight loops: `i`, `r`, `g`
- Descriptive names for struct fields: `DisplayName`, `RoleDefinitionID`

**Types:**
- PascalCase for exported types: `Client`, `Role`, `Group`, `Model`
- Struct types describe entities: `LighthouseSubscription`, `EligibleAzureRole`
- Interface types not used (no explicit interfaces defined)
- Message types for Bubbletea: suffix with `Msg` - `clientReadyMsg`, `rolesLoadedMsg`

**Constants:**
- PascalCase for exported: `StatusActive`, `TabRoles`
- camelCase for package-private: `graphBaseURL`, `pimBaseURL`
- Enum pattern using iota: `StatusInactive ActivationStatus = iota`

## Code Style

**Formatting:**
- Standard `gofmt` formatting (no explicit config)
- Tabs for indentation
- No trailing whitespace

**Linting:**
- No explicit linter configuration (`.golangci.yml` not present)
- Standard Go conventions followed

## Import Organization

**Order:**
1. Standard library imports
2. External dependencies (blank line separator)
3. Internal project imports (blank line separator)

**Path Aliases:**
- Tea framework alias: `tea "github.com/charmbracelet/bubbletea"`
- No other path aliases used

**Example from `internal/ui/model.go`:**
```go
import (
    "context"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/atotto/clipboard"
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"

    "github.com/seb07-cloud/pim-tui/internal/azure"
    "github.com/seb07-cloud/pim-tui/internal/config"
)
```

## Error Handling

**Patterns:**
- Return errors as last return value: `func (c *Client) GetRoles(ctx context.Context) ([]Role, error)`
- Wrap errors with context using `fmt.Errorf`: `fmt.Errorf("failed to get token: %w", err)`
- Check errors immediately after function call
- Non-critical errors logged, not returned (e.g., user info fetch in `loadUserInfoCmd`)

**Error wrapping example from `internal/azure/client.go`:**
```go
if err != nil {
    return nil, fmt.Errorf("failed to get token: %w", err)
}
```

**Error messages in UI:**
- Logged via `m.log(LogError, "message: %v", err)`
- User-facing errors shown in error state view

## Logging

**Framework:** Custom log system in `internal/ui/model.go`

**Levels:**
- `LogError` (0) - Errors and failures
- `LogInfo` (1) - Normal operations
- `LogDebug` (2) - Verbose debugging

**Patterns:**
- Log via Model method: `m.log(LogInfo, "format string", args...)`
- Log entries kept in circular buffer (max 100 entries)
- Logs displayed in UI panel, copyable to clipboard

**Example:**
```go
m.log(LogInfo, "Loaded %d eligible roles", len(m.roles))
m.log(LogError, "Activation failed: %v", msg.err)
m.log(LogDebug, "Post-activation refresh...")
```

## Comments

**When to Comment:**
- Package-level documentation for types
- Function documentation for exported functions
- Inline comments for non-obvious logic
- URL references for external documentation

**Style:**
- Full sentences with periods
- No redundant comments on obvious code

**Example from `internal/azure/client.go`:**
```go
// NewClient creates a new Azure client
// It tries 'az rest' first (which handles Graph API auth properly), then falls back to SDK
func NewClient() (*Client, error) {
```

## Function Design

**Size:**
- Most functions under 50 lines
- Larger view functions in `views.go` are acceptable (render functions)
- Complex logic split into helper functions

**Parameters:**
- Context as first parameter for API calls: `func (c *Client) GetRoles(ctx context.Context) ([]Role, error)`
- Pointer receivers for methods that modify state: `func (m *Model) log(...)`
- Value receivers for read-only methods: `func (m Model) View() string`

**Return Values:**
- Single return for simple operations
- (result, error) tuple for fallible operations
- Named returns not used

## Module Design

**Exports:**
- Only export what's needed by other packages
- Internal helper functions kept private

**Barrel Files:**
- Not used - direct imports to specific files

**Package Organization:**
- `internal/azure/` - Azure API client and types
- `internal/config/` - Configuration loading
- `internal/ui/` - Bubbletea TUI components
- `cmd/pim-tui/` - Main entry point

## Type Patterns

**Struct Tags:**
- JSON tags for API responses: `json:"id"`
- YAML tags for config: `yaml:"default_duration"`

**Enums with iota:**
```go
type ActivationStatus int

const (
    StatusInactive ActivationStatus = iota
    StatusActive
    StatusExpiringSoon
    StatusPending
)
```

**Method on enum types:**
```go
func (s ActivationStatus) String() string {
    switch s {
    case StatusActive:
        return "Active"
    // ...
    }
}
```

## Concurrency Patterns

**Parallel API calls:**
- Use `sync.WaitGroup` for coordinating parallel requests
- Mutex for protecting shared state in goroutines

**Example from `internal/azure/pim.go`:**
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

## Bubbletea Patterns

**Model Structure:**
- Single `Model` struct holds all UI state
- Fields grouped by concern with comments
- Maps for selection state: `selectedRoles map[int]bool`

**Message Types:**
- Private message structs: `type rolesLoadedMsg struct{ roles []azure.Role }`
- Error messages include source: `type errMsg struct { err error; source string }`

**Command Pattern:**
- Commands return `tea.Cmd`
- Closures capture dependencies
- Timeouts via `context.WithTimeout`

**Example:**
```go
func loadRolesCmd(client *azure.Client) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        roles, err := client.GetRoles(ctx)
        if err != nil {
            return errMsg{fmt.Errorf("failed to load roles: %w", err), "roles"}
        }
        return rolesLoadedMsg{roles}
    }
}
```

## Style Variables Pattern

**Lipgloss styles defined as package variables in `internal/ui/styles.go`:**
```go
var (
    colorActive    = lipgloss.Color("#00ff00")
    activeStyle    = lipgloss.NewStyle().Foreground(colorActive)
    activeBoldStyle = lipgloss.NewStyle().Foreground(colorActive).Bold(true)
)
```

**Theme application:**
- `ApplyTheme()` function updates colors
- `rebuildStyles()` reconstructs style objects with new colors

---

*Convention analysis: 2026-01-16*
