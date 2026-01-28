```
 ██████╗ ██╗███╗   ███╗    ████████╗██╗   ██╗██╗
 ██╔══██╗██║████╗ ████║    ╚══██╔══╝██║   ██║██║
 ██████╔╝██║██╔████╔██║       ██║   ██║   ██║██║
 ██╔═══╝ ██║██║╚██╔╝██║       ██║   ██║   ██║██║
 ██║     ██║██║ ╚═╝ ██║       ██║   ╚██████╔╝██║
 ╚═╝     ╚═╝╚═╝     ╚═╝       ╚═╝    ╚═════╝ ╚═╝

 Azure Privileged Identity Management in your terminal
```

<p align="center">
  <a href="https://github.com/seb07-cloud/pim-tui/releases"><img src="https://img.shields.io/github/v/release/seb07-cloud/pim-tui?style=flat-square&color=00ADD8" alt="Release"></a>
  <a href="https://github.com/seb07-cloud/pim-tui/actions"><img src="https://img.shields.io/github/actions/workflow/status/seb07-cloud/pim-tui/release.yml?branch=main&style=flat-square" alt="Build Status"></a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <a href="https://goreportcard.com/report/github.com/seb07-cloud/pim-tui"><img src="https://goreportcard.com/badge/github.com/seb07-cloud/pim-tui?style=flat-square" alt="Go Report Card"></a>
  <img src="https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey?style=flat-square" alt="Platform">
</p>

---

**A fast, keyboard-driven TUI for Azure Privileged Identity Management with built-in security tier awareness.**

Activate Entra ID roles, PIM groups, and Azure RBAC roles across Lighthouse tenants—all from your terminal. PIM-TUI brings just-in-time privilege elevation to the command line with real-time status tracking, attack path visualization, and full audit trail support.

<!--
TODO: Add demo GIF showing:
1. Launching app and authenticating
2. Navigating between tabs (PIM Roles, PIM Groups, Subscriptions)
3. Selecting and activating a role with justification
4. Viewing tier information and attack paths
5. Using tree view to see role inheritance
-->

---

## Interface Preview

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Contoso Corp                               Token expires: 2h 45m  ● 3 Active│
├──────────────────────────────────────────────────────────────────────────────┤
│  [PIM Roles (5)]       PIM Groups (3)       Subscriptions (12)               │
├───────────────────────────────────┬──────────────────────────────────────────┤
│                                   │                                          │
│  ● Global Administrator      [T0] │  Role Details                            │
│  ○ Security Administrator         │  ───────────────────────────────         │
│  ○ User Administrator             │  Name: Global Administrator              │
│  ◐ Application Administrator      │  Status: Active expires 1h 23m           │
│  ○ Cloud App Security Admin       │  Tier: 0 - Control Plane                 │
│                                   │                                          │
│                                   │  WARNING: Tier 0 Role                    │
│                                   │  This role has tenant-wide admin rights. │
│                                   │  Attack Path: Direct privilege escalation│
│                                   │                                          │
├───────────────────────────────────┴──────────────────────────────────────────┤
│  Activity Log                                                                │
│  [INFO] Role activated: Global Administrator (4h)                            │
│  [INFO] Refreshed role status                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│  Duration: [4h]   Auto-refresh: 45s   Selected: 1   /search                  │
└──────────────────────────────────────────────────────────────────────────────┘

  ● Active    ◐ Expiring    ○ Inactive    ◌ Pending    [T0] Tier 0
```

---

## Features

### 🔐 Core Functionality

| Feature               | Description                                            |
|-----------------------|--------------------------------------------------------|
| **Entra ID PIM Roles**| Activate/deactivate Azure AD privileged roles          |
| **PIM Groups**        | Manage privileged access groups with linked roles      |
| **Azure RBAC**        | Manage roles across Lighthouse (delegated) subscriptions|
| **Batch Activation**  | Activate multiple roles simultaneously                 |
| **Justification**     | Mandatory reason input for audit compliance            |

### 🛡️ Security Tier Awareness

PIM-TUI classifies every role by security tier, so you always know the blast radius:

| Tier       | Risk Level | Description                              | Visual    |
|------------|------------|------------------------------------------|-----------|
| **Tier 0** | Critical   | Control Plane - tenant-wide admin rights | 🔴 Red    |
| **Tier 1** | High       | Data plane with significant access       | 🟠 Orange |
| **Tier 2** | Medium     | Service support and management           | 🟡 Yellow |
| **Tier 3** | Low        | Standard user-level access               | 🟢 Green  |

- Attack path visualization for Tier 0 roles
- Escalation risk warnings
- Tier data sourced from [azure-tiering](https://github.com/emiliensocchi/azure-tiering)

### 📊 Visualization & Monitoring

- **Tree View**: ASCII diagram showing role inheritance flow to tenant
- **Real-time Status**: Active role count, expiry countdown, token status
- **Activity Logs**: Filterable logs (ERROR/INFO/DEBUG)
- **Activation History**: Track all activations with timestamps

### ⚡ Productivity

- **Vim-style Navigation**: `j/k`, `h/l`, `g/G` bindings
- **Quick Duration**: Press `1-4` for preset durations
- **Auto-refresh**: Configurable interval with countdown display
- **Search/Filter**: Real-time filtering with `/`
- **Clipboard Export**: Copy logs and history with `c`/`e`

---

## Quick Start

### Prerequisites

- Azure CLI installed and authenticated (`az login`)
- Eligible PIM roles in your tenant

### Installation

**Using Go:**
```bash
go install github.com/seb07-cloud/pim-tui/cmd/pim-tui@latest
```

**Pre-built Binaries:**

Download from [Releases](https://github.com/seb07-cloud/pim-tui/releases):

| Platform | Architecture          | Download                    |
|----------|-----------------------|-----------------------------|
| Windows  | x86_64                | `pim-tui-windows-amd64.exe` |
| Linux    | x86_64                | `pim-tui-linux-amd64`       |
| Linux    | ARM64                 | `pim-tui-linux-arm64`       |
| macOS    | x86_64                | `pim-tui-darwin-amd64`      |
| macOS    | ARM64 (Apple Silicon) | `pim-tui-darwin-arm64`      |

**Build from Source:**
```bash
git clone https://github.com/seb07-cloud/pim-tui.git
cd pim-tui
go build -o pim-tui ./cmd/pim-tui
```

### Run

```bash
# Authenticate with Azure CLI first
az login

# Launch PIM-TUI
pim-tui
```

---

## Usage

### Keybindings

<details>
<summary><strong>Navigation</strong></summary>

| Key          | Action         |
|--------------|----------------|
| `↑` / `k`    | Move up        |
| `↓` / `j`    | Move down      |
| `←` / `h`    | Previous tab   |
| `→` / `l`    | Next tab       |
| `Tab`        | Cycle tabs     |
| `g` / `Home` | Jump to top    |
| `G` / `End`  | Jump to bottom |

</details>

<details>
<summary><strong>Actions</strong></summary>

| Key            | Action                  |
|----------------|-------------------------|
| `Space`        | Select/deselect item    |
| `Enter`        | Activate selected roles |
| `x` / `Delete` | Deactivate active roles |
| `y` / `Enter`  | Confirm dialog          |
| `n` / `Esc`    | Cancel dialog           |

</details>

<details>
<summary><strong>Duration & Settings</strong></summary>

| Key | Action                  |
|-----|-------------------------|
| `1` | Set duration to 1 hour  |
| `2` | Set duration to 2 hours |
| `3` | Set duration to 4 hours |
| `4` | Set duration to 8 hours |
| `d` | Cycle through presets   |
| `a` | Toggle auto-refresh     |
| `v` | Cycle log level         |
| `r` | Manual refresh          |

</details>

<details>
<summary><strong>Display & Export</strong></summary>

| Key | Action                    |
|-----|---------------------------|
| `?` | Show help                 |
| `/` | Search/filter             |
| `t` | Open tree view            |
| `c` | Copy logs to clipboard    |
| `e` | Export activation history |
| `q` | Quit                      |

</details>

### Common Workflows

**Activate a Role:**
```
1. Navigate to role with j/k
2. Press Space to select
3. Press 1-4 to set duration (optional)
4. Press Enter to activate
5. Confirm with y
6. Enter justification and press Enter
```

**View Role Inheritance:**
```
1. Press t to open tree view
2. Navigate with j/k
3. Press a to animate flow
4. Press Esc to close
```

**Filter Subscriptions:**
```
1. Go to Subscriptions tab with l
2. Start typing to filter
3. Press Esc to clear filter
```

---

## Configuration

PIM-TUI reads configuration from `~/.config/pim-tui/config.yaml`:

```yaml
# Default activation duration in hours
default_duration: 4

# Quick duration presets (mapped to keys 1-4)
duration_presets: [1, 2, 4, 8]

# Log level: error, info, debug
log_level: info

# Auto-refresh interval in seconds
auto_refresh_interval: 60
auto_refresh_enabled: true

# Theme customization
theme:
  color_active: "#00ff00"      # Active roles
  color_expiring: "#ffff00"    # Expiring soon (<30m)
  color_inactive: "#808080"    # Inactive roles
  color_pending: "#00bfff"     # Pending approval
  color_error: "#ff0000"       # Errors and Tier 0
  color_highlight: "#7d56f4"   # Selected items
  color_border: "#444444"      # Panel borders
```

If the config file doesn't exist, defaults are used.

---

## Authentication

PIM-TUI supports two authentication methods:

### 1. Azure CLI (Recommended)

```bash
az login
pim-tui
```

Credentials are read from the Azure CLI cache. No additional setup required.

### 2. Browser OAuth (Fallback)

If Azure CLI credentials aren't available:
1. Press `L` when prompted
2. Complete authentication in your browser
3. Credentials are cached in `~/.config/pim-tui/auth_record.json`

**Token Expiry:** The status bar shows remaining token validity. Re-authenticate when expired.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                           PIM-TUI                                │
├──────────────────────────────────────────────────────────────────┤
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    │
│  │   TUI    │    │  State   │    │  Azure   │    │  Config  │    │
│  │  Views   │◄──►│ Machine  │◄──►│  Client  │    │  Loader  │    │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘    │
│       ▲                               │                          │
│       │                               ▼                          │
│  ┌──────────┐              ┌─────────────────────┐               │
│  │  Bubble  │              │    Azure APIs       │               │
│  │   Tea    │              ├─────────────────────┤               │
│  └──────────┘              │ • Graph API         │               │ 
│                            │ • PIM Governance    │               │
│                            │ • ARM (Lighthouse)  │               │
│                            └─────────────────────┘               │ 
└──────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
pim-tui/
├── cmd/pim-tui/
│   └── main.go              # Entry point
├── internal/
│   ├── azure/               # Azure API client
│   │   ├── client.go        # HTTP client, auth
│   │   ├── pim.go           # Entra ID PIM API
│   │   ├── groups.go        # PIM groups API
│   │   ├── lighthouse.go    # Azure RBAC API
│   │   ├── tiers.go         # Security tier lookup
│   │   └── data/            # Embedded tier data
│   ├── ui/                  # TUI layer
│   │   ├── model.go         # State machine
│   │   ├── views.go         # Rendering
│   │   ├── styles.go        # Theming
│   │   └── tree_*.go        # Tree visualization
│   └── config/              # Configuration
└── .github/workflows/       # CI/CD
```

---

## Roadmap

### Current (v1.2.x)
- [x] Entra ID PIM role management
- [x] PIM group activation with linked roles
- [x] Azure RBAC via Lighthouse
- [x] Security tier classification
- [x] Tree view visualization
- [x] Activity logging and history export
- [x] Configurable themes

### Planned
- [ ] Multi-tenant switcher
- [ ] Custom tier overrides

---

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
git clone https://github.com/seb07-cloud/pim-tui.git
cd pim-tui
go mod download
go build -o pim-tui ./cmd/pim-tui
./pim-tui
```

### Running Tests

```bash
go test ./...
```

---

## Acknowledgments

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) - Azure integration

### Security Tier Data

- [azure-tiering](https://github.com/emiliensocchi/azure-tiering) by Emilien Socchi

---

## License

This project is available under the MIT License. See [LICENSE](LICENSE) for details.

---

<p align="center">
  <sub>Built with ❤️ for Azure security engineers who live in the terminal</sub>
</p>
