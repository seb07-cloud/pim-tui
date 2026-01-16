# PIM-TUI

```
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║   ██████╗ ██╗███╗   ███╗    ████████╗██╗   ██╗██╗                 ║
║   ██╔══██╗██║████╗ ████║    ╚══██╔══╝██║   ██║██║                 ║
║   ██████╔╝██║██╔████╔██║       ██║   ██║   ██║██║                 ║
║   ██╔═══╝ ██║██║╚██╔╝██║       ██║   ██║   ██║██║                 ║
║   ██║     ██║██║ ╚═╝ ██║       ██║   ╚██████╔╝██║                 ║
║   ╚═╝     ╚═╝╚═╝     ╚═╝       ╚═╝    ╚═════╝ ╚═╝                 ║
║                                                                   ║
║   Azure Privileged Identity Management - Terminal UI              ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
```

A terminal UI for Azure Privileged Identity Management built with Go and Bubble Tea.

**Version:** 0.1.0

---

## Features

- **Azure Authentication** - Uses Azure SDK DefaultAzureCredential (Azure CLI, managed identity, environment)
- **PIM Roles** - List and activate eligible Azure AD directory roles
- **PIM Groups** - List and activate eligible PIM-enabled group memberships
- **Lighthouse** - View delegated subscriptions linked to PIM groups
- **Multi-select** - Select multiple roles/groups for batch activation
- **Duration presets** - Quick selection (1h, 2h, 4h, 8h)
- **Justification** - Prompted when required by policy
- **Auto-refresh** - Configurable automatic status refresh
- **Colorful UI** - Status indicators, progress bars, colored logs
- **Clipboard** - Copy log entries to clipboard

---

## Screenshots

### Main View (Tabbed Layout)
```
┌──────────────────────────────────┬──────────────────────────────────────┐
│  ██████╗ ██╗███╗   ███╗          │  Tenant:  Contoso                    │
│  ██╔══██╗██║████╗ ████║          │  ID:      xxxxxxxx-xxxx-xxxx-xxxx    │
│  ██████╔╝██║██╔████╔██║          │  User:    admin@contoso.com          │
│  ██╔═══╝ ██║██║╚██╔╝██║          │  Active:  2 roles, 1 groups          │
│  ██║     ██║██║ ╚═╝ ██║          │  Refresh: Auto (45s)                 │
│  ╚═╝     ╚═╝╚═╝     ╚═╝          │  Version: v0.1.0                     │
└──────────────────────────────────┴──────────────────────────────────────┘
  [Roles]  Groups                              (Tab/←→ to switch)
┌──────────────────────────────────┬──────────────────────────────────────┐
│  ● PIM Roles                     │  Role Details                        │
│  ─────────────────────────────── │  ──────────────────────────────────  │
│  [x] ● Global Admin    ████ 2h   │  Name: Global Administrator          │
│  [ ] ○ User Admin      ────      │  Status: ● Active                    │
│  [ ] ○ Billing Reader  ────      │  Expires: 1h 45m                     │
│                                  │                                      │
│                                  │  Permissions:                        │
│                                  │  • microsoft.directory/*/allTasks    │
│                                  │  • microsoft.azure.*/allEntities     │
├──────────────────────────────────┴──────────────────────────────────────┤
│  [INFO]  14:32:01 Authentication successful                             │
│  [INFO]  14:32:02 Connected to tenant: Contoso                          │
└─────────────────────────────────────────────────────────────────────────┘
  ↑↓ navigate  Tab/←→ switch tab  Space select  Enter activate  ? help  q quit
```

### Loading Screen
```
 ██████╗ ██╗███╗   ███╗    ████████╗██╗   ██╗██╗
 ██╔══██╗██║████╗ ████║    ╚══██╔══╝██║   ██║██║
 ██████╔╝██║██╔████╔██║       ██║   ██║   ██║██║
 ██╔═══╝ ██║██║╚██╔╝██║       ██║   ██║   ██║██║
 ██║     ██║██║ ╚═╝ ██║       ██║   ╚██████╔╝██║
 ╚═╝     ╚═╝╚═╝     ╚═╝       ╚═╝    ╚═════╝ ╚═╝

                    v0.1.0
            Connected to: Contoso

        ⠋ Loading PIM roles and groups...

  ✓ Initializing Azure SDK
  ✓ Loading tenant info
  ⠋ Loading PIM roles
  ⠋ Loading PIM groups
```

---

## Project Structure

```
pim-tui/
├── cmd/
│   └── pim-tui/
│       └── main.go              # Entry point, version
├── internal/
│   ├── azure/
│   │   ├── types.go             # Data types (Role, Group, Tenant, etc.)
│   │   ├── client.go            # Azure SDK client, auth, HTTP requests
│   │   ├── pim.go               # PIM roles API
│   │   ├── groups.go            # PIM groups API
│   │   └── lighthouse.go        # Lighthouse subscriptions API
│   ├── ui/
│   │   ├── model.go             # Bubble Tea model, state, update logic
│   │   ├── views.go             # All view rendering (loading, main, help, etc.)
│   │   ├── styles.go            # Lipgloss styles, colors, icons
│   │   └── keys.go              # Key bindings
│   └── config/
│       └── config.go            # YAML config loading
├── go.mod
├── go.sum
├── .gitignore
└── PLAN.md
```

---

## Key Bindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Navigate up |
| `↓` / `j` | Navigate down |
| `←` / `h` | Switch to Roles tab |
| `→` / `l` | Switch to Groups tab |
| `Tab` | Switch tabs |
| `Space` | Toggle selection |
| `Enter` | Activate selected items |
| `x` / `Del` / `BS` | Deactivate selected active items |
| `/` | Search/filter roles and groups |
| `Esc` | Clear search / cancel action |
| `L` | Toggle Lighthouse mode |
| `r` / `R` | Refresh data |
| `a` | Toggle auto-refresh |
| `1-4` | Set duration (1h, 2h, 4h, 8h) |
| `v` | Cycle log level (error/info/debug) |
| `c` | Copy logs to clipboard |
| `e` | Export activation history |
| `?` | Show/hide help |
| `q` / `Ctrl+C` | Quit |

---

## Configuration

**Location**: `~/.config/pim-tui/config.yaml`

```yaml
# Default activation duration in hours
default_duration: 4

# Duration presets shown in UI (hours)
duration_presets:
  - 1
  - 2
  - 4
  - 8

# Log level: error, info, debug
log_level: info

# Auto-refresh interval in seconds
auto_refresh_interval: 60

# Auto-refresh enabled on startup
auto_refresh_enabled: true

# Custom theme colors (hex format)
theme:
  color_active: "#00ff00"    # Green - active status
  color_expiring: "#ffff00"  # Yellow - expiring soon
  color_inactive: "#808080"  # Gray - inactive status
  color_pending: "#00bfff"   # Blue - pending approval
  color_error: "#ff0000"     # Red - errors
  color_highlight: "#7d56f4" # Purple - highlights/accents
  color_border: "#444444"    # Dark gray - borders
```

---

## Installation & Usage

### Build from source
```bash
git clone https://github.com/seb07-cloud/pim-tui
cd pim-tui
go build -o pim-tui ./cmd/pim-tui
./pim-tui
```

### Prerequisites
- Go 1.21+
- Azure CLI authenticated (`az login`)
- Appropriate Azure AD permissions (see below)

---

## Azure Permissions Required

### Microsoft Graph API
- `RoleManagement.Read.Directory` - List eligible roles
- `RoleManagement.ReadWrite.Directory` - Activate roles
- `PrivilegedAccess.Read.AzureADGroup` - List eligible PIM groups
- `PrivilegedAccess.ReadWrite.AzureADGroup` - Activate group membership
- `Directory.Read.All` - Read tenant/organization info
- `User.Read` - Get current user ID

### Azure Resource Manager (for Lighthouse)
- Reader access to delegated subscriptions

### Authentication Setup

The default Azure CLI token doesn't include Graph API permissions for PIM. You have two options:

**Option 1: Use Azure CLI with specific scopes**
```bash
az login --scope https://graph.microsoft.com/.default
```

**Option 2: Create an App Registration**
1. Register an app in Azure AD
2. Add the required Microsoft Graph API permissions (see above)
3. Grant admin consent
4. Use one of these auth methods:
   - Client credentials: Set `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
   - Certificate: Set `AZURE_CLIENT_ID`, `AZURE_CLIENT_CERTIFICATE_PATH`, `AZURE_TENANT_ID`

**Common Error:**
```
API error 403: PermissionScopeNotGranted
```
This means the authenticated identity lacks the required Graph API permissions.

---

## Dependencies

```
github.com/charmbracelet/bubbletea    - TUI framework
github.com/charmbracelet/bubbles      - TUI components
github.com/charmbracelet/lipgloss     - Styling
github.com/Azure/azure-sdk-for-go     - Azure authentication
github.com/atotto/clipboard           - Clipboard support
gopkg.in/yaml.v3                      - Config parsing
```

---

## Status Icons

| Icon | Color | Meaning |
|------|-------|---------|
| `●` | Green | Active |
| `◐` | Yellow | Expiring soon (< 30 min) |
| `○` | Gray | Inactive |
| `◌` | Blue | Pending approval |

---

## Implementation Status

### ✅ Completed
- [x] Azure SDK authentication with DefaultAzureCredential
- [x] Tenant info display (name + ID)
- [x] PIM roles listing (eligible + active status)
- [x] PIM groups listing (name, description, status)
- [x] Multi-select across panels
- [x] Duration presets (1h, 2h, 4h, 8h)
- [x] Batch activation with confirmation
- [x] Justification prompt
- [x] Progress bars for active items
- [x] Log panel with timestamps
- [x] Log level cycling (error/info/debug)
- [x] Auto-refresh with toggle
- [x] Lighthouse mode for delegated subscriptions
- [x] Help overlay
- [x] Loading screen with progress steps
- [x] Error handling with retry
- [x] Copy logs to clipboard
- [x] Version display in header
- [x] YAML config file support

### ✅ Recently Added
- [x] Deactivate active roles/groups (x/Del/Backspace)
- [x] Filter/search roles and groups (/ key)
- [x] Export activation history (e key)
- [x] Custom themes/colors in config
- [x] Tabbed UI (Roles / Groups tabs)
- [x] Role details panel with permissions
- [x] Group details panel with linked roles

### 🔄 Future Enhancements
- [ ] Fetch group linked Azure RBAC roles dynamically
- [ ] Multiple tenant support

---

## Architecture Notes

### Bubble Tea Model
The application uses a single Bubble Tea model with states:
- `StateLoading` - Initial authentication and data loading
- `StateNormal` - Main interactive view
- `StateConfirm` - Activation confirmation dialog
- `StateConfirmDeactivate` - Deactivation confirmation dialog
- `StateJustification` - Justification input
- `StateActivating` - Processing activations
- `StateDeactivating` - Processing deactivations
- `StateSearch` - Search/filter input
- `StateHelp` - Help overlay
- `StateError` - Authentication/critical error

### Azure API Flow
1. `NewClient()` - Create credential with DefaultAzureCredential
2. `GetCurrentUser()` - Validate authentication
3. `GetTenant()` - Load organization info
4. `GetRoles()` / `GetGroups()` - Load eligible and active assignments
5. `ActivateRole()` / `ActivateGroup()` - Self-activate with justification

### Tick-based Updates
- 100ms tick for smooth spinner animations
- Auto-refresh check on each tick (when enabled)
- Progress bars update in real-time

---

## License

MIT
