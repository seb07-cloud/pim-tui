---
created: 2026-01-16T07:15
title: Improve role selection cursor visibility
area: ui
files:
  - internal/ui/views.go
  - internal/ui/styles.go
---

## Problem

When navigating the roles list, the current selection is hard to see. The checkbox/indicator moves slightly but doesn't stand out enough. User suggests making the selected row white or more prominently highlighted.

Current: Subtle indicator movement
Desired: Clear visual highlighting (e.g., white background, inverted colors, or bold text for selected row)

## Solution

TBD - Options:
1. Add background color to selected row (lipgloss style)
2. Invert foreground/background for selected item
3. Add a more visible cursor indicator (e.g., `>` or `│`)
4. Combine approaches for maximum visibility
