---
created: 2026-01-16T07:15
title: Smart wrap long permission strings at path segments
area: ui
files:
  - internal/ui/views.go
---

## Problem

In the Role Details panel, long permission strings wrap mid-word when they exceed the panel width. This breaks readability:

Current (bad):
```
• microsoft.directory/users/authenticationMethods/allProperties/allTask
  s
```

Desired (good):
```
• microsoft.directory/users
     /authenticationMethods/allProperties/allTasks
```

The user wants intelligent wrapping at path segment boundaries (/) with proper indentation on continuation lines.

## Solution

TBD - Approaches:
1. Split permission string at `/` boundaries when wrapping
2. Calculate available width, find best break point at path segment
3. Indent continuation lines to align with content after bullet
4. Consider truncation with `...` as fallback for very long paths
