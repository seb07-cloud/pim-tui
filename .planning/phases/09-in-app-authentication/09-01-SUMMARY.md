---
phase: 09-in-app-authentication
plan: 01
status: complete
duration: ~15 min
commits:
  - 9ec7fb7 (device code - reverted)
  - 22b0237 (device code - reverted)
  - 0b3e9e4 (device code - reverted)
  - 7d6f3c7 (browser auth)
---

## Summary

Implemented in-app browser authentication allowing users to log in without restarting the app.

## What Changed

### internal/azure/client.go
- Added `AuthenticateWithBrowser()` function using `azidentity.NewInteractiveBrowserCredential`
- Opens system browser for Azure AD authentication
- Returns configured Client on successful authentication

### internal/ui/model.go
- Added `StateUnauthenticated` and `StateAuthenticating` states
- Added `authRequiredMsg`, `authCompleteMsg` message types
- Added `authCancelFunc` field for cancellation support
- Modified `initClientCmd()` to detect auth failures and return `authRequiredMsg`
- Added `startAuthCmd()` that calls `azure.AuthenticateWithBrowser`
- Handle 'L' key in StateUnauthenticated to start browser auth
- Handle Esc to cancel authentication in progress

### internal/ui/views.go
- Added `renderUnauthenticated()` view with ASCII logo
- Shows "Authentication Required" with login/quit options
- Shows "Waiting for browser sign-in..." during authentication
- Added ANSI clear screen for full-screen auth states
- Uses `lipgloss.Place` for centered content

### internal/ui/keys.go
- Added 'L' keybinding for login action

## Approach Change

Original plan specified device code flow. User requested browser authentication instead due to device code often being blocked by security policies. Browser auth provides better UX with automatic browser popup.

## Verification

- [x] `go build ./...` compiles
- [x] `go test ./...` passes
- [x] App shows friendly "Authentication Required" screen when not logged in
- [x] 'L' key starts browser authentication
- [x] Browser opens to Azure login page
- [x] App transitions to loading after successful auth
- [x] App works normally after in-app authentication
- [x] Esc cancels authentication flow

## Decisions

- Browser auth over device code (user security requirement)
- Native azidentity.InteractiveBrowserCredential (simpler than custom MSAL wrapper)
- ANSI clear screen codes for clean full-screen rendering
