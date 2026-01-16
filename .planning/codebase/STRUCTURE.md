# Codebase Structure

**Analysis Date:** 2026-01-16

## Directory Layout

```
pim-tui/
├── cmd/
│   └── pim-tui/
│       └── main.go         # Application entry point
├── internal/
│   ├── azure/              # Azure API client and types
│   │   ├── client.go       # HTTP client, auth, request helpers
│   │   ├── groups.go       # PIM group operations
│   │   ├── lighthouse.go   # Azure subscription/RBAC operations
│   │   ├── pim.go          # PIM Entra role operations
│   │   └── types.go        # Domain type definitions
│   ├── config/
│   │   └── config.go       # Configuration loading and defaults
│   └── ui/
│       ├── keys.go         # Reserved for key binding migration
│       ├── model.go        # Bubble Tea model (state, update, init)
│       ├── roles_builtin.go # Built-in role permission mappings
│       ├── styles.go       # Lipgloss styles and colors
│       └── views.go        # View rendering functions
├── .planning/
│   └── codebase/           # Codebase analysis documents
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── pim-tui                 # Linux binary (built)
└── pim-tui.exe             # Windows binary (built)
```

## Directory Purposes

**cmd/pim-tui/:**
- Purpose: Application entry point
- Contains: Single `main.go` file
- Key files: `main.go` (28 lines)

**internal/azure/:**
- Purpose: Azure API communication layer
- Contains: HTTP client, authentication, API operations for PIM
- Key files:
  - `client.go`: Authentication, request helpers (333 lines)
  - `pim.go`: Entra ID role activation/deactivation (204 lines)
  - `groups.go`: PIM group operations (241 lines)
  - `lighthouse.go`: Azure RBAC/subscription operations (470 lines)
  - `types.go`: Domain type definitions (106 lines)

**internal/config/:**
- Purpose: Application configuration
- Contains: Config struct, YAML loading, defaults
- Key files: `config.go` (75 lines)

**internal/ui/:**
- Purpose: Terminal user interface
- Contains: Bubble Tea model, views, styles, state management
- Key files:
  - `model.go`: State management, Update loop, commands (1151 lines)
  - `views.go`: All View rendering functions (1520 lines)
  - `styles.go`: Lipgloss color and style definitions (241 lines)
  - `roles_builtin.go`: Built-in role permission data (218 lines)
  - `keys.go`: Placeholder for future key binding migration (5 lines)

## Key File Locations

**Entry Points:**
- `cmd/pim-tui/main.go`: Application bootstrap, Bubble Tea program start

**Configuration:**
- `internal/config/config.go`: Config struct, Load() function, Default()
- User config file: `$XDG_CONFIG_HOME/pim-tui/config.yaml` or `~/.config/pim-tui/config.yaml`

**Core Logic:**
- `internal/ui/model.go`: All state, Update() message handling, Init()
- `internal/azure/client.go`: Azure authentication, HTTP request handling
- `internal/azure/pim.go`: Entra role API operations
- `internal/azure/groups.go`: PIM group API operations
- `internal/azure/lighthouse.go`: Azure RBAC API operations

**Testing:**
- No test files present (test coverage gap)

**Styling:**
- `internal/ui/styles.go`: All color definitions, Lipgloss styles
- `internal/ui/views.go`: Rendering functions using styles

## Naming Conventions

**Files:**
- lowercase_snake.go: All Go source files
- Functionality-based: `client.go`, `pim.go`, `groups.go`, `model.go`, `views.go`
- Type-focused: `types.go`, `config.go`, `styles.go`

**Directories:**
- lowercase: `cmd`, `internal`, `azure`, `config`, `ui`
- Standard Go layout: `cmd/` for binaries, `internal/` for private packages

**Functions:**
- Exported: PascalCase (`NewClient`, `GetRoles`, `NewModel`)
- Unexported: camelCase (`graphRequest`, `renderHeader`, `initClientCmd`)
- Command functions: Suffix with `Cmd` (`loadRolesCmd`, `tickCmd`)
- Message types: Suffix with `Msg` (`clientReadyMsg`, `rolesLoadedMsg`)

**Types:**
- Structs: PascalCase (`Model`, `Client`, `Role`, `Group`)
- Enums: PascalCase type with const block (`State`, `Tab`, `LogLevel`)
- Interfaces: Not used (concrete types only)

**Variables:**
- Package-level styles: camelCase (`panelStyle`, `colorActive`)
- Constants: PascalCase or camelCase depending on export

## Where to Add New Code

**New API Operation:**
- Add to existing file in `internal/azure/` based on domain:
  - Entra roles: `internal/azure/pim.go`
  - PIM groups: `internal/azure/groups.go`
  - Azure RBAC: `internal/azure/lighthouse.go`
- Add new types to `internal/azure/types.go`

**New UI State:**
- Add State constant to `internal/ui/model.go` (StateXxx)
- Add fields to Model struct in `internal/ui/model.go`
- Add message type in `internal/ui/model.go`
- Handle in Update() switch in `internal/ui/model.go`

**New View/Dialog:**
- Add render function to `internal/ui/views.go`
- Pattern: `func (m Model) renderXxx() string`
- Call from main View() function in `internal/ui/views.go`

**New Configuration Option:**
- Add field to `Config` struct in `internal/config/config.go`
- Set default in `Default()` function
- Theme options go in `ThemeConfig` struct

**New Styles:**
- Add to `internal/ui/styles.go`
- Follow existing pattern: package-level var with lipgloss.NewStyle()

**New Key Binding:**
- Add case to `handleKeyPress()` in `internal/ui/model.go`
- Document in renderHelp() in `internal/ui/views.go`

## Special Directories

**.planning/:**
- Purpose: Project planning and analysis documents
- Generated: By analysis tools
- Committed: Yes

**.planning/codebase/:**
- Purpose: Codebase mapping documents for Claude agents
- Generated: By `/gsd:map-codebase` command
- Committed: Yes

**.claude/:**
- Purpose: Claude Code configuration and commands
- Generated: Manual/tool setup
- Committed: Yes (provides context for AI assistance)

**Built Binaries:**
- `pim-tui`: Linux binary
- `pim-tui.exe`: Windows binary
- Generated: `go build` in project root
- Committed: Yes (for distribution)

---

*Structure analysis: 2026-01-16*
