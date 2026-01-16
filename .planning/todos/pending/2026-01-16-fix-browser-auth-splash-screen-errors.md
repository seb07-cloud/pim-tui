---
created: 2026-01-16T10:30
title: Fix browser auth splash screen errors causing duplicates
area: ui
files:
  - internal/azure/client.go:59-100
  - internal/ui/views.go:196-230
  - internal/ui/model.go:646-681
---

## Problem

When logging in via browser authentication, error messages are shown on the splash screen which causes text duplication/movement issues:

1. **WSL interoperability errors** appear at bottom of terminal:
   - `grep: /proc/sys/fs/binfmt_misc/WSLInterop: No such file or directory`
   - `WSL Interopability is disabled. Please enable it before using WSL.`
   - `[error] WSL Interoperability is disabled. Please enable it before using WSL.`

2. **Duplicate text lines** appear in the UI:
   - "Opening browser for authentication..." shows twice with different spinner states
   - Other UI elements duplicate during auth flow

Root causes investigated but not fully resolved:
- Browser launcher (wslview/xdg-open) prints errors to stderr outside Bubble Tea's control
- Terminal not clearing properly between renders despite ANSI clear codes
- syscall.Dup2 stderr redirect not affecting subprocess output
- Azure SDK logging suppression not sufficient

Attempts made:
- Added ANSI clear screen codes (`\033[2J\033[H`)
- Suppressed Azure SDK logging via `log.SetListener(nil)`
- Redirected stderr at fd level with syscall.Dup2
- Set BROWSER env var to use Windows cmd.exe directly
- Simplified view to reduce render complexity

## Solution

TBD - Requires deeper investigation into:
1. VS Code terminal + WSL rendering behavior
2. Alternative browser opening approach that doesn't use shell scripts
3. Possibly running browser open in separate process with fully isolated stdio
4. Consider whether interactive browser auth is viable in WSL or if alternative auth flow needed
