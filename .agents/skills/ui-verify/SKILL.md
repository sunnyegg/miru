---
name: ui-verify
description: >-
  Verify Miru UI work visually using a browser. Opens the Wails dev server in
  agent-browser, takes screenshots, checks layout, colors, typography,
  accessibility, interaction flows, and error states against the Miru design
  system. Use after frontend changes, PR review, or when the user asks to
  verify, check, review, or test the UI visually.
argument-hint: "[visual|a11y|interaction|errors|full]"
---

# UI Verify

Browser-based verification of Miru's frontend. Uses `agent-browser` to open the
Wails dev server, inspect pages, and report pass/fail findings.

## Prerequisites

`agent-browser` must be installed. If missing:

```bash
npm i -g agent-browser && agent-browser install
```

Confirm with `agent-browser --version`. If the command is not available, stop
and ask the user to install it.

Load the core workflow before running any commands:

```bash
agent-browser skills get core
```

## Mode

Parse `$ARGUMENTS` for the mode keyword. Default to `visual` if none is given.

| Mode | What it checks |
|------|----------------|
| `visual` | Layout, colors, typography, spacing against Miru design system |
| `a11y` | Contrast, focus visibility, keyboard navigation, ARIA labels |
| `interaction` | Click through key flows, verify state transitions |
| `errors` | Trigger error/empty states, verify alerts and `notice()` |
| `full` | Run all four modes sequentially |

## Workflow

### 1. Ensure dev server is running

```bash
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:34115
```

If the response is not `200`, tell the user to run `make dev` and wait for it
to be ready. Do not start the server yourself — it is a Wails process that
needs a display.

### 2. Open the app

```bash
agent-browser open http://127.0.0.1:34115
```

### 3. Run checks per mode

Load the design reference before visual checks:

```
Read .agents/skills/ui-verify/references/miru-design.md
```

#### `visual`

1. `agent-browser screenshot` the current page.
2. `agent-browser snapshot -i` to get the element tree.
3. Check against the design reference:
   - Border radius is zero on all elements.
   - Interactive controls are 44px tall.
   - Purple (`#a78bfa`) is used only for active/hit states (Play, progress,
     selected mark), not for general decoration.
   - Colors come from Tailwind tokens (`bg-background`, `text-muted-foreground`,
     `border-border`), not raw hex.
   - No CSS modules or inline color values.
   - Font is Source Sans 3 Variable. No other font families.
   - Headings use `text-balance`. Body uses `font-smoothing: antialiased`.
4. Navigate to each major view (Library, Search, Downloads, Settings) and
   repeat. Use `agent-browser snapshot -i` to find navigation elements.
5. Report findings: pass or fail per check with the screenshot as evidence.

#### `a11y`

1. `agent-browser snapshot -i` to get the full element tree.
2. Check:
   - All interactive elements have accessible names (label, aria-label, or
     visible text).
   - Focus ring is visible: `outline: 2px solid var(--ring)` with offset.
   - Tab order is logical — tab through the page with `agent-browser press Tab`
     and verify focus moves in reading order.
   - Color contrast: text on backgrounds meets WCAG AA (4.5:1 for body, 3:1
     for large text). The key pairs to check:
     - `#f2f2f2` on `#111111` (foreground on background) — 17.4:1
     - `#c2c2c2` on `#111111` (muted-foreground on background) — 10.4:1
     - `#f2f2f2` on `#181818` (foreground on card) — 15.3:1
     - `#a78bfa` on `#111111` (accent on background) — 7.0:1
   - Live regions exist for dynamic content updates (toasts, loading states).
3. Report findings per check.

#### `interaction`

1. Identify key flows from the current page using `agent-browser snapshot -i`.
2. For each flow, execute the steps and verify the outcome:
   - **Library → Play**: click a poster, verify episode list appears, click
     Play, verify playback UI state.
   - **Search**: type in search input, verify results appear, select a result.
   - **Settings**: navigate to Settings, toggle a setting, verify the change
     persists after re-navigation.
   - **Downloads**: navigate to Downloads, verify tab switching works, verify
     torrent list renders.
3. After each flow, `agent-browser snapshot -i` to confirm the expected state.
4. Report pass/fail per flow with the final snapshot.

#### `errors`

1. Check load failure handling: navigate to a view that fetches data, observe
   if an inline Alert appears on failure (not a blank screen or console error).
2. Check empty states: navigate to a view with no data (e.g. empty library),
   verify a meaningful empty state is shown.
3. Check form validation: submit a form with invalid input, verify error
   messages appear inline.
4. Verify error patterns match conventions:
   - Load failures use inline Alert.
   - Action failures use `notice(..., true)`.
   - No swallowed errors (no empty catch blocks visible in snapshot).
5. Report findings per check.

#### `full`

Run `visual` → `a11y` → `interaction` → `errors` in order. Collect all
findings into a single report at the end.

### 4. Report

After all checks complete, produce a summary:

```
## UI Verification Report

**Mode:** <mode>
**URL:** http://127.0.0.1:34115

### Visual
- [PASS/FAIL] Border radius: zero
- [PASS/FAIL] Control height: 44px
- ...

### Accessibility
- [PASS/FAIL] Focus ring visible
- ...

(only include sections relevant to the mode)
```

If any check fails, include the screenshot and the specific element or
component that violated the rule.

### 5. Close

```bash
agent-browser close
```
