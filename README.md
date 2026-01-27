# PIM-TUI

```
    ____  ________  ___   __________  ______
   / __ \/  _/  |/  /  /_  __/ / / //  _/
  / /_/ // // /|_/ /____/ / / / / / / /
 / ____// // /  / /_____/ / / /_/ /_/ /
/_/   /___/_/  /_/     /_/  \____//___/

Azure Privileged Identity Management - Terminal User Interface
```

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build](https://img.shields.io/github/actions/workflow/status/seb07-cloud/pim-tui/release.yml?branch=main)
![Release](https://img.shields.io/github/v/release/seb07-cloud/pim-tui)
![Go Report Card](https://goreportcard.com/badge/github.com/seb07-cloud/pim-tui)

---

**Manage your Azure PIM role activations without leaving the terminal.**

PIM-TUI is a powerful Terminal User Interface for Azure Privileged Identity Management that allows security administrators and DevOps engineers to view, activate, and manage privileged role assignments directly from the command line. No more switching to the Azure Portal for quick role activations.

<!-- TODO: Add animated GIF demo showing:
     1. Launching the app and authenticating
     2. Navigating between tabs (Roles, Groups, Subscriptions)
     3. Selecting and activating a role with justification
     4. Viewing the role inheritance tree
-->

---

## Screenshots

### Main Interface

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  PIM-TUI v1.2.2                                    contoso.onmicrosoft.com      │
├──────────────────────────────┬──────────────────────────────────────────────────┤
│  🔐 Roles (3) │ 👥 Groups │ 📑 Subscriptions                                     │
├──────────────────────────────┼──────────────────────────────────────────────────┤
│                              │                                                   │
│  ○ Global Administrator     │  Role Details                                     │
│  ● Security Administrator   │  ─────────────────────────────────────────────── │
│  ◐ User Administrator       │  Name: Security Administrator                    │
│  ○ Application Administrator│  Status: ● Active                                │
│  ○ Cloud App Security Admin │  Tier: T0 - Control Plane                        │
│                              │  Expires: 2h 45m remaining                       │
│                              │  ▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░ 68%                       │
│                              │                                                   │
│                              │  Permissions:                                    │
│                              │  • Manage security policies                       │
│                              │  • Configure identity protection                 │
│                              │  • Review security reports                        │
│                              │                                                   │
├──────────────────────────────┴──────────────────────────────────────────────────┤
│  [12:34:05] ✓ Activated Security Administrator for 4h                           │
│  ↑/↓ navigate │ Space select │ Enter activate │ t tree │ ? help │ q quit        │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Role Inheritance Tree View

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  Role Inheritance Tree                                                   [ESC]  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  👤 user@contoso.com                                                            │
│   │                                                                              │
│   ├── 👥 Security Operations Group                                              │
│   │    ├── 🔐 Security Administrator (T0)                                       │
│   │    └── 🔐 Security Reader                                                   │
│   │                                                                              │
│   └── 👥 IT Admins                                                              │
│        ├── 🔐 User Administrator (T1)                                           │
│        └── 📑 Subscription: Production                                          │
│             └── 🔐 Contributor                                                  │
│                                                                                  │
│  ──────────────────────────────────────────────────────────────────────────────│
│  Press 'a' to toggle animation │ ←/→ navigate │ ESC to close                    │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Features

### 🔐 Core PIM Management

- **Entra ID Roles** — View and activate eligible Privileged Identity Management roles
- **Group Memberships** — Manage PIM-enabled group member and owner assignments
- **Azure Subscriptions** — Handle Lighthouse-delegated subscription role activations
- **Multi-Select Activation** — Activate multiple roles simultaneously with a single justification

### 🛡️ Security-First Design

- **Security Tier Classification** — Visual indicators (T0-T3) based on [azure-tiering](https://github.com/emiliensocchi/azure-tiering)
  - 🔴 **T0 - Control Plane**: Tenant-wide administrative rights
  - 🟠 **T1 - High Privilege**: Significant access capabilities
  - 🟡 **T2 - Medium Privilege**: Limited administrative scope
  - 🟢 **T3 - Low Privilege**: Read-only or minimal permissions
- **Escalation Path Warnings** — Alerts for roles that can lead to privilege escalation
- **T0 Activation Confirmation** — Extra confirmation step for high-risk role activations

### 📊 Real-Time Monitoring

- **Status Tracking** — Live updates showing Active, Expiring, Inactive, and Pending states
- **Duration Progress Bars** — Visual countdown of remaining activation time
- **Expiry Warnings** — Color-coded alerts when roles are expiring soon (<30 min)
- **Auto-Refresh** — Configurable automatic data refresh intervals

### 🌳 Visualization

- **Role Inheritance Tree** — ASCII diagram showing how roles flow through group memberships
- **Animated Tree View** — Optional animation showing activation propagation paths
- **Built-in Role Permissions** — Embedded documentation for 50+ Entra ID roles

### ⚡ Productivity Features

- **Keyboard-Driven** — Full Vim-style navigation (hjkl) plus intuitive shortcuts
- **Quick Search** — Filter roles, groups, and subscriptions instantly
- **Activation History** — Track all activations with export to clipboard
- **Activity Logging** — Detailed logs with adjustable verbosity levels
- **Clipboard Integration** — Copy logs and history for documentation

---

## Quick Start

```bash
# Install with Go
go install github.com/seb07-cloud/pim-tui/cmd/pim-tui@latest

# Run
pim-tui
```

On first launch, PIM-TUI will open your browser for Azure authentication. Credentials are cached for seamless subsequent launches.

---

## Installation

### Pre-built Binaries

Download the latest release for your platform from [Releases](https://github.com/seb07-cloud/pim-tui/releases):

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 | `pim-tui-linux-amd64` |
| Linux | arm64 | `pim-tui-linux-arm64` |
| macOS | Intel | `pim-tui-darwin-amd64` |
| macOS | Apple Silicon | `pim-tui-darwin-arm64` |
| Windows | amd64 | `pim-tui-windows-amd64.exe` |

```bash
# Linux/macOS example
curl -L https://github.com/seb07-cloud/pim-tui/releases/latest/download/pim-tui-linux-amd64 -o pim-tui
chmod +x pim-tui
sudo mv pim-tui /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/seb07-cloud/pim-tui.git
cd pim-tui
go build -o pim-tui ./cmd/pim-tui
```

### Package Managers

<details>
<summary>Homebrew (coming soon)</summary>

```bash
brew tap seb07-cloud/tap
brew install pim-tui
```

</details>

<details>
<summary>Scoop (Windows - coming soon)</summary>

```powershell
scoop bucket add seb07-cloud https://github.com/seb07-cloud/scoop-bucket
scoop install pim-tui
```

</details>

---

## Usage

### Keybindings

| Key | Action |
|-----|--------|
| **Navigation** | |
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `h` | Previous tab |
| `→` / `l` | Next tab |
| `Tab` | Cycle tabs |
| **Selection & Actions** | |
| `Space` | Toggle selection |
| `Enter` | Activate selected roles |
| `x` / `Delete` | Deactivate active roles |
| `1-4` | Quick-set duration (1h, 2h, 4h, 8h) |
| `d` | Cycle duration presets |
| **Views** | |
| `t` | Open role inheritance tree |
| `/` | Open search |
| `Esc` | Close dialog / Clear filter |
| `?` | Show help |
| **Data & Logs** | |
| `r` / `R` / `F5` | Refresh data from Azure |
| `a` | Toggle auto-refresh |
| `v` | Cycle log level (ERROR/INFO/DEBUG) |
| `c` | Copy logs to clipboard |
| `e` | Export activation history |
| **Exit** | |
| `q` / `Ctrl+C` | Quit |

### Status Icons

| Icon | Meaning |
|------|---------|
| `●` | Active |
| `◐` | Expiring soon (<30 min) |
| `○` | Inactive / Eligible |
| `◌` | Pending activation |

### Tab Overview

- **🔐 Roles** — Entra ID PIM role assignments
- **👥 Groups** — PIM-enabled group memberships (Member/Owner)
- **📑 Subscriptions** — Lighthouse-delegated Azure subscriptions

---

## Configuration

PIM-TUI stores its configuration at `~/.config/pim-tui/config.yaml`:

```yaml
# Duration settings
default_duration: 4              # Default activation duration in hours
duration_presets: [1, 2, 4, 8]   # Available duration options

# Behavior
log_level: "info"                # Log verbosity: debug, info, error
auto_refresh_interval: 60        # Auto-refresh interval in seconds
auto_refresh_enabled: true       # Enable auto-refresh on startup

# Theme customization
theme:
  color_active: "#00ff00"        # Active status color
  color_expiring: "#ffff00"      # Expiring soon color
  color_inactive: "#808080"      # Inactive status color
  color_pending: "#00bfff"       # Pending status color
  color_error: "#ff0000"         # Error message color
  color_highlight: "#7d56f4"     # Selection highlight
  color_border: "#444444"        # Border color
```

---

## Authentication

PIM-TUI supports multiple authentication methods (in order of priority):

1. **Cached Credentials** — Silent authentication using stored tokens
2. **Azure CLI** — Uses existing `az login` session if available
3. **Interactive Browser** — Opens system browser for OAuth flow

Tokens are securely cached in your OS config directory for seamless restarts.

### Required Permissions

Your Azure account needs the following to use PIM-TUI:

- Eligible role assignments in Azure AD PIM
- For Subscriptions tab: Azure Lighthouse delegations or subscription-level PIM assignments

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          PIM-TUI                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   UI Layer  │  │   Config    │  │      Azure Client       │  │
│  │  (Bubble Tea)│  │   (YAML)    │  │   (azidentity + SDK)    │  │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬────────────┘  │
│         │                │                      │                │
│         └────────────────┴──────────────────────┘                │
│                          │                                       │
├──────────────────────────┼───────────────────────────────────────┤
│                          ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    Azure APIs                                ││
│  │  • Microsoft Graph API (roles, groups)                       ││
│  │  • PIM Governance API (activations)                         ││
│  │  • Azure Resource Manager (subscriptions)                   ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
pim-tui/
├── cmd/pim-tui/
│   └── main.go              # Entry point
├── internal/
│   ├── azure/               # Azure API client & authentication
│   │   ├── client.go        # Auth & token management
│   │   ├── pim.go           # Entra ID role operations
│   │   ├── groups.go        # Group membership handling
│   │   ├── lighthouse.go    # Subscription & delegated access
│   │   └── tiers.go         # Security tier classification
│   ├── config/
│   │   └── config.go        # YAML configuration
│   └── ui/
│       ├── model.go         # TUI state machine
│       ├── views.go         # Rendering & components
│       ├── styles.go        # Colors & styling
│       ├── tree_*.go        # Role inheritance visualization
│       └── roles_builtin.go # Embedded role permissions
└── go.mod
```

---

## Security Tier System

PIM-TUI includes built-in security tier classifications based on the [azure-tiering](https://github.com/emiliensocchi/azure-tiering) project:

| Tier | Risk Level | Description | Examples |
|------|------------|-------------|----------|
| **T0** | 🔴 Critical | Control Plane - Full tenant control | Global Admin, Privileged Role Admin |
| **T1** | 🟠 High | Significant privileges with escalation potential | User Admin, Exchange Admin |
| **T2** | 🟡 Medium | Limited administrative scope | Helpdesk Admin, Groups Admin |
| **T3** | 🟢 Low | Read-only or minimal impact | Security Reader, Reports Reader |

When activating T0 roles, PIM-TUI displays additional warnings about potential attack scenarios and requires explicit confirmation.

---

## Roadmap

- [x] Entra ID PIM role management
- [x] PIM group membership activation
- [x] Azure subscription role activation (Lighthouse)
- [x] Role inheritance tree visualization
- [x] Security tier classification
- [x] Multi-platform binary releases
- [x] Configuration file support
- [x] Activation history & export
- [ ] Homebrew formula
- [ ] Scoop manifest (Windows)
- [ ] AUR package (Arch Linux)
- [ ] Role activation scheduling
- [ ] Favorites/bookmarks for frequently used roles
- [ ] Custom tier overrides
- [ ] Notification integration (desktop alerts)

---

## Contributing

Contributions are welcome! Here's how to get started:

### Development Setup

```bash
# Clone the repository
git clone https://github.com/seb07-cloud/pim-tui.git
cd pim-tui

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o pim-tui ./cmd/pim-tui

# Run in development
go run ./cmd/pim-tui
```

### Guidelines

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes with clear messages
4. **Test** your changes thoroughly
5. **Push** to your fork
6. **Open** a Pull Request

### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Add tests for new functionality
- Update documentation as needed

---

## Acknowledgments

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — The fun, functional TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions for terminal apps
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components for Bubble Tea
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) — Official Azure client libraries

### Data Sources

- [azure-tiering](https://github.com/emiliensocchi/azure-tiering) by Emilien Socchi — Security tier classifications for Azure roles

### Inspiration

- [lazygit](https://github.com/jesseduffield/lazygit) — Simple terminal UI for git commands
- [k9s](https://github.com/derailed/k9s) — Kubernetes CLI to manage clusters

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <sub>Built with ❤️ for the terminal-loving Azure community</sub>
</p>
