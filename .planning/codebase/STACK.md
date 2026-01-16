# Technology Stack

**Analysis Date:** 2025-01-16

## Languages

**Primary:**
- Go 1.25.5 - All application code

**Secondary:**
- None

## Runtime

**Environment:**
- Go 1.25.5
- Cross-platform (Windows, Linux, macOS)

**Package Manager:**
- Go Modules
- Lockfile: `go.sum` (present, 96 lines)

## Frameworks

**Core:**
- Bubbletea v1.3.10 - Terminal UI framework (Elm architecture)
- Bubbles v0.21.0 - UI components (help, textinput)
- Lipgloss v1.1.0 - Terminal styling and layout

**Testing:**
- None configured (no test files detected)

**Build/Dev:**
- Standard `go build`
- No Makefile or build scripts detected

## Key Dependencies

**Critical:**
- `github.com/Azure/azure-sdk-for-go/sdk/azcore` v1.21.0 - Azure core SDK for API requests
- `github.com/Azure/azure-sdk-for-go/sdk/azidentity` v1.13.1 - Azure authentication (DefaultAzureCredential, AzureCLICredential)
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI framework

**Infrastructure:**
- `gopkg.in/yaml.v3` v3.0.1 - YAML config parsing
- `github.com/atotto/clipboard` v0.1.4 - Clipboard operations for log export

**Azure Authentication (Indirect):**
- `github.com/AzureAD/microsoft-authentication-library-for-go` v1.6.0 - MSAL for OAuth flows
- `github.com/golang-jwt/jwt/v5` v5.3.0 - JWT token handling
- `github.com/pkg/browser` - Browser-based auth flows

## Configuration

**Environment:**
- No `.env` file used
- Authentication via Azure CLI (`az login`) or DefaultAzureCredential chain
- Configuration file: `$XDG_CONFIG_HOME/pim-tui/config.yaml` or `~/.config/pim-tui/config.yaml`

**Config Schema:**
```yaml
default_duration: 4           # Default activation duration in hours
duration_presets: [1, 2, 4, 8] # Quick-select duration options
log_level: "info"             # debug, info, error
auto_refresh_interval: 60     # Seconds between auto-refresh
auto_refresh_enabled: true    # Enable auto-refresh
theme:
  color_active: "#00ff00"     # Green
  color_expiring: "#ffff00"   # Yellow
  color_inactive: "#808080"   # Gray
  color_pending: "#00bfff"    # Blue
  color_error: "#ff0000"      # Red
  color_highlight: "#7d56f4"  # Purple
  color_border: "#444444"
```

**Build:**
- No build configuration files
- Standard `go build ./cmd/pim-tui`
- Produces `pim-tui` (Linux/Mac) or `pim-tui.exe` (Windows)

## Platform Requirements

**Development:**
- Go 1.25.5+
- Azure CLI installed (`az` command available)
- Azure subscription with PIM access

**Production:**
- Compiled binary (no runtime dependencies)
- Azure CLI installed for authentication
- Terminal with Unicode support (for status icons)
- Windows/Linux/macOS supported

## Build Commands

```bash
# Build for current platform
go build -o pim-tui ./cmd/pim-tui

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o pim-tui.exe ./cmd/pim-tui

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o pim-tui ./cmd/pim-tui

# Run directly
go run ./cmd/pim-tui
```

## Version

- Application version: `0.1.0` (hardcoded in `cmd/pim-tui/main.go`)

---

*Stack analysis: 2025-01-16*
