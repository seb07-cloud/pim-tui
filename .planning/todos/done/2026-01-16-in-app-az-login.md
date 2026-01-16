---
created: 2026-01-16T09:34
title: Add in-app az login authentication option
area: auth
files:
  - internal/azure/client.go
  - internal/ui/model.go
---

## Problem

Currently, when the app starts up it checks if the user is authenticated via Azure CLI credentials. If not authenticated, the app shows an error and must be closed. The user then has to:
1. Close the app
2. Run `az login` in terminal
3. Restart the app

This is disruptive to the workflow. Users want to authenticate directly from within the app without leaving and restarting.

## Solution

TBD - Approaches:
1. Detect auth failure at startup, show "Not authenticated" state instead of error
2. Add UI option/key binding to trigger `az login` flow (could spawn subprocess)
3. After successful login, refresh credentials and continue to normal loading
4. Consider: `az login --use-device-code` for non-interactive terminal environments
5. Consider: Show clear instructions if interactive login isn't possible
