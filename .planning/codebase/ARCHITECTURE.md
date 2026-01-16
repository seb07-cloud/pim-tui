# Architecture

**Analysis Date:** 2026-01-16

## Pattern Overview

**Overall:** Model-View-Update (MVU) / Elm Architecture via Bubble Tea

**Key Characteristics:**
- Single immutable model containing all application state
- Unidirectional data flow: User Input -> Messages -> Update -> View
- Commands (tea.Cmd) for asynchronous side effects (API calls)
- Separation between UI layer (`internal/ui`) and API layer (`internal/azure`)

## Layers

**Presentation Layer (internal/ui):**
- Purpose: Terminal UI rendering and user input handling
- Location: `internal/ui/`
- Contains: Model state, view functions, key bindings, styles
- Depends on: `internal/azure` (types), `internal/config`, Bubble Tea/Lipgloss
- Used by: `cmd/pim-tui/main.go`

**API Layer (internal/azure):**
- Purpose: Azure PIM and Graph API communication
- Location: `internal/azure/`
- Contains: Client, HTTP requests, API response parsing, domain types
- Depends on: Azure SDK, standard library HTTP
- Used by: `internal/ui` (via messages and commands)

**Configuration Layer (internal/config):**
- Purpose: User configuration loading and defaults
- Location: `internal/config/`
- Contains: Config struct, YAML parsing, default values
- Depends on: `gopkg.in/yaml.v3`
- Used by: `internal/ui`, `cmd/pim-tui/main.go`

**Entry Point (cmd/pim-tui):**
- Purpose: Application bootstrap
- Location: `cmd/pim-tui/main.go`
- Contains: Main function, program initialization
- Depends on: `internal/ui`, `internal/config`
- Used by: End user (CLI execution)

## Data Flow

**Initialization Flow:**

1. `main.go` loads config via `config.Load()`
2. Creates UI model with `ui.NewModel(cfg, version)`
3. Starts Bubble Tea program with `tea.NewProgram(m, tea.WithAltScreen())`
4. Model.Init() returns `initClientCmd()` and `tickCmd()`
5. `initClientCmd` creates Azure client asynchronously
6. On success, sends `clientReadyMsg` triggering tenant/data loading

**Data Loading Flow:**

1. `clientReadyMsg` received -> store client, start `loadTenantCmd()`
2. `tenantLoadedMsg` received -> store tenant, start parallel loading:
   - `loadRolesCmd()` -> `rolesLoadedMsg`
   - `loadGroupsCmd()` -> `groupsLoadedMsg`
   - `loadLighthouseCmd()` -> `lighthouseLoadedMsg`
3. `checkLoadingComplete()` transitions to `StateNormal` when all loaded

**Activation Flow:**

1. User selects items with Space key
2. User presses Enter -> `initiateActivation()` -> `StateConfirm`
3. User confirms with Y -> `StateJustification`
4. User enters justification, presses Enter -> `startActivation()`
5. Async command calls `client.ActivateRole/Group/AzureRole()`
6. Returns `activationDoneMsg` -> refresh data, clear selections

**State Management:**
- All state in single `Model` struct
- State transitions via `State` enum (StateLoading, StateNormal, StateConfirm, etc.)
- Selection state via maps: `selectedRoles`, `selectedGroups`, `selectedSubRoles`
- Cursor position via integers: `rolesCursor`, `groupsCursor`, `lightCursor`

## Key Abstractions

**Model (internal/ui/model.go):**
- Purpose: Central state container implementing tea.Model interface
- Examples: `internal/ui/model.go` lines 93-163
- Pattern: Implements Init(), Update(), View() for Bubble Tea

**Client (internal/azure/client.go):**
- Purpose: Azure API communication with multiple auth strategies
- Examples: `internal/azure/client.go`
- Pattern: Lazy credential initialization, az CLI fallback, request retry with backoff

**Message Types (internal/ui/model.go):**
- Purpose: Typed messages for state transitions
- Examples: `clientReadyMsg`, `rolesLoadedMsg`, `activationDoneMsg`
- Pattern: One message type per async operation result

**Domain Types (internal/azure/types.go):**
- Purpose: Business domain entities
- Examples: `Role`, `Group`, `LighthouseSubscription`, `EligibleAzureRole`
- Pattern: Flat structs with activation status tracking

## Entry Points

**CLI Entry (cmd/pim-tui/main.go):**
- Location: `cmd/pim-tui/main.go`
- Triggers: User runs `pim-tui` binary
- Responsibilities: Load config, create model, run tea.Program

**Model.Init (internal/ui/model.go):**
- Location: `internal/ui/model.go` line 234
- Triggers: Bubble Tea program starts
- Responsibilities: Return initial commands for client creation and tick timer

**Model.Update (internal/ui/model.go):**
- Location: `internal/ui/model.go` line 353
- Triggers: Any message received (key press, async result, tick)
- Responsibilities: State transitions, return follow-up commands

## Error Handling

**Strategy:** Error messages stored in model, displayed in UI, logged

**Patterns:**
- `errMsg` struct with error and source field for categorization
- Non-fatal errors (roles/groups loading) mark as loaded and continue
- Fatal errors (auth, tenant) transition to `StateError` state
- Error state allows retry (R key) or quit (Q key)
- All errors logged via `m.log(LogError, ...)` for activity panel

**Recovery:**
- Auth failures: User can press R to retry authentication
- API failures: Individual data sources fail independently
- Rate limiting: `azRestRequestWithRetry` implements exponential backoff

## Cross-Cutting Concerns

**Logging:**
- Internal logging via `Model.log()` method
- Log entries stored in `Model.logs` slice (max 100)
- Three levels: LogError, LogInfo, LogDebug
- Displayed in activity log panel at bottom of UI
- Copyable to clipboard with C key

**Validation:**
- Justification required before activation (checked in `startActivation`)
- Selection validation before showing confirm dialog

**Authentication:**
- Dual strategy: `az rest` CLI command preferred, Azure SDK fallback
- Credentials cached in Client struct
- Separate credentials for Graph API and PIM API
- Windows cross-platform support via `cmd /c az` pattern

**API Endpoints:**
- Graph API: `https://graph.microsoft.com/v1.0`
- PIM Governance API: `https://api.azrbac.mspim.azure.com/api/v2/privilegedAccess`
- ARM API: `https://management.azure.com`

---

*Architecture analysis: 2026-01-16*
