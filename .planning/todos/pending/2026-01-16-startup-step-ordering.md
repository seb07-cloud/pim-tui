---
created: 2026-01-16T07:15
title: Fix startup step ordering display
area: ui
files:
  - internal/ui/model.go
  - internal/ui/views.go
---

## Problem

During startup, loading steps sometimes show green out of order (Step 1 → Step 2 → Step 4 → app starts, with Step 3 still pending). User reported this happened once; second test worked correctly (1-3 green, then started).

This suggests a race condition where step completion messages arrive/display out of order, or step 4 can complete before step 3 finishes.

## Solution

TBD - Investigate:
1. Are steps actually completing out of order, or just displaying out of order?
2. Check if step dependencies are enforced
3. Consider forcing sequential display even if parallel completion
