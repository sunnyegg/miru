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

## Screenshots are evidence (mandatory)

No finding is reported without a screenshot backing it. A snapshot tree proves
what is in the DOM, never what it looks like, so every claim — pass or fail,
including "blocked" — ships with the image that shows it.

```bash
mkdir -p /tmp/miru-ui-verify
agent-browser screenshot /tmp/miru-ui-verify/<view>-<state>.png
```

The path is positional; `--path` and other flags are not part of the command.
Use absolute paths, one file per state, with the view and the state in the name
(`search-torrent-results.png`, `search-batch-bar-3-selected.png`,
`downloads-empty.png`). Re-screenshot after each interaction that changes the
screen, so a flow yields a before/after pair rather than a single final shot.

Attach each file in the report with markdown image syntax and the absolute path
(`![search batch bar](/tmp/miru-ui-verify/search-batch-bar-3-selected.png)`), not
a bare path or a description of what it contains — a path nobody opens is not
evidence. Keep the files until the report is delivered; never clean them up
first.

If the model cannot read images, say so in the report, still capture and attach
the screenshots for the operator, and base pass/fail on DOM evidence
(`agent-browser snapshot -i`, computed styles, measured heights) instead of
claiming to have looked at the picture. Never omit the screenshot because the
model cannot see it.

When a check cannot run at all (indexer unreachable, empty data set, no dev
server), the blocked finding also gets a screenshot of what the app actually
showed, e.g. the inline error alert.

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

1. `agent-browser screenshot /tmp/miru-ui-verify/<view>.png` for the current
   page, and again after any state change you judge.
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
   repeat. Use `agent-browser snapshot -i` to find navigation elements. Every
   view visited gets its own screenshot.
5. Report findings: pass or fail per check, each with its screenshot attached.

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
3. Screenshot the focus ring and any element whose label or contrast you are
   judging.
4. Report findings per check, each with its screenshot attached.

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
3. After each flow, `agent-browser snapshot -i` to confirm the expected state,
   and screenshot the end state (plus any intermediate state the flow is meant
   to show, such as a dialog, a bulk bar or a toast).
4. Report pass/fail per flow, each with its screenshot attached.

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
5. Screenshot each error and empty state while it is on screen — notices and
   alerts disappear, so capture before moving on.
6. Report findings per check, each with its screenshot attached.

#### `full`

Run `visual` → `a11y` → `interaction` → `errors` in order. Collect all
findings into a single report at the end.

### 4. Report

After all checks complete, produce a summary that carries the screenshots:

```
## UI Verification Report

**Mode:** <mode>
**URL:** http://127.0.0.1:34115

### Visual
- [PASS/FAIL] Border radius: zero
  ![library](/tmp/miru-ui-verify/library.png)
- [PASS/FAIL] Control height: 44px
  ![search torrent](/tmp/miru-ui-verify/search-torrent-results.png)
- [BLOCKED] <check that could not run, and why>
  ![](/tmp/miru-ui-verify/search-torrent-blocked.png)

### Accessibility
- [PASS/FAIL] Focus ring visible
  ![focus ring](/tmp/miru-ui-verify/settings-focus-ring.png)

(only include sections relevant to the mode)

### Screenshots
| File | Shows |
| --- | --- |
| `/tmp/miru-ui-verify/library.png` | Library grid, zero radius, 44px controls |
```

Rules for the report:

- Every line item has at least one image directly under it. No image, no claim.
- A failure also names the specific element or component that violated the rule,
  next to its screenshot.
- `BLOCKED` is its own verdict: it is not a pass, and it still carries a
  screenshot of whatever the app showed.

### 5. Close

```bash
agent-browser close
```

Leave the screenshots on disk after closing the browser: the report references
them by path, and the operator may want to look at them again.
