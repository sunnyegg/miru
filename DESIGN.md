---
name: Miru
description: Matte MPV-bezel client — posters are the field, orange is the hit.
colors:
  background: "#111111"
  foreground: "#f2f2f2"
  accent: "#ff5f1f"
  on-accent: "#111111"
  bezel: "#0a0a0a"
  card: "#181818"
  muted: "#1c1c1c"
  secondary: "#1c1c1c"
  on-secondary: "#f2f2f2"
  muted-foreground: "#c2c2c2"
  border: "#2a2a2a"
  destructive: "#ef4444"
  on-destructive: "#111111"
  ring: "#ff5f1f"
typography:
  headline:
    fontFamily: "Source Sans 3 Variable, Source Sans 3, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 2rem
  title:
    fontFamily: "Source Sans 3 Variable, Source Sans 3, sans-serif"
    fontSize: "1rem"
    fontWeight: 500
    lineHeight: 1.5rem
  body:
    fontFamily: "Source Sans 3 Variable, Source Sans 3, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.25rem
  label:
    fontFamily: "Source Sans 3 Variable, Source Sans 3, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1rem
rounded:
  none: "0px"
spacing:
  xs: "8px"
  sm: "12px"
  md: "16px"
  lg: "20px"
  xl: "24px"
components:
  button-play:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "0 20px"
    height: "44px"
    width: "96px"
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.on-secondary}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "0 16px"
    height: "44px"
  button-muted:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "0 12px"
    height: "44px"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.muted-foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "0 12px"
    height: "44px"
  button-destructive:
    backgroundColor: "{colors.destructive}"
    textColor: "{colors.on-destructive}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "0 16px"
    height: "44px"
  input:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    padding: "0 12px"
    height: "44px"
  nav-item:
    backgroundColor: "transparent"
    textColor: "{colors.muted-foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    height: "44px"
  nav-item-current:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.none}"
    height: "44px"
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.none}"
    padding: "16px"
---

# Design System: Miru

## Overview

**Creative North Star: "MPV Bezel"**

Miru is a player bezel, not a storefront. The field is matte black; chrome (the left rail and the playback status strip) sits one step darker. Type is pale phosphor. The only loud color is accent orange, used as the hit: primary actions, progress, the selected mark. Every control is a rectangle. One workhorse sans carries titles, rows, and labels.

Library is the signature surface: square posters fill the field; selecting a show opens an episode list in the field. Click a row to play. Import and AniList match stay on the sheet above the field. Other destinations inherit the same tokens, bezel rail, zero-radius chrome, and type — not the Library two-step navigation.

Confirmed visual rejection: a streaming-app card grid with rounded tiles and a Play on every row.

**Key Characteristics:**

- Matte field with darker bezel chrome
- Single orange hit (Play, progress, selected marks)
- Zero-radius rectangles; square posters
- One sans: Source Sans 3 Variable, self-hosted
- Hairline rules and tonal stacking; no shadows
- Named left rail; Settings docked at the bottom
- Lucide icons at 16px, stroke 1.75; Play is the only filled glyph

## Colors

A near-black bezel with phosphor type and one orange hit. Neutrals do the chrome; orange does the action.

### Primary

- **Accent Orange** (`accent`): Primary actions (search, save, start download), playback and download progress fills, the playing episode row mark, the highlighted poster outline, the current rail hairline, caret, text selection, and the scrollbar thumb. Its job is the hit, not the fill.

### Neutral

- **Matte Field** (`background`): The page and Library poster field.
- **Deep Bezel** (`bezel`): Left rail and the playback status strip. Always darker than the field.
- **Sheet** (`card`): Inset panels and list blocks on inherited views; episode rows and the AniList match sheet on Library.
- **Panel** (`muted`, `secondary`): Same value, two jobs — muted fills (skeletons, input wells, poster fallbacks, nav hover) and secondary buttons (Import, Search, Download, Connect). Secondary hover steps to muted only when the control is already on secondary; muted controls hover to secondary.
- **Phosphor** (`foreground`, `on-secondary`): Body type and type on secondary fills. `on-accent` and `on-destructive` use the field value so type on orange and on fault red stays dark.
- **Dim Type** (`muted-foreground`): Captions, metadata, idle rail labels, empty field copy.
- **Hairline** (`border`): Full-weight on Library match sheet. Inherited views often stroke at 40% of this value.
- **Fault Red** (`destructive`): Init banner, error toasts, error copy, cancel-download fill. Not a second brand accent.

**The Play Hit Rule.** Orange is the hit: primary actions, progress, the playing row mark, the highlighted poster outline, the current rail mark. Do not fill chrome, sheets, or poster tiles with it.

**The Bezel Stack Rule.** Rail and status strip sit on deep bezel; the field is one step lighter; sheets sit one step lighter than the field. Depth is this stack, not a shadow.

## Typography

**Display Font:** none (no display face)
**Body Font:** Source Sans 3 Variable, with Source Sans 3, then sans-serif
**Label/Mono Font:** none (same family; no mono)

**Character:** One workhorse sans at UI sizes. Semibold for page titles and the wordmark; medium for Play, poster names, and sheet titles; regular for everything else. Tight, player-OSD, never editorial.

### Hierarchy

- **Headline** (600, 1.5rem / 2rem): Page titles (Library, Watching, Search, Downloads, Airing, Settings). Library tightens tracking; other pages leave it default — do not require tracking-tight.
- **Title** (500, 1rem / 1.5rem): Sheet headings and card headings (Match AniList, day blocks, settings groups). The expanded wordmark is 1.125rem / 600 on the rail from the `sm` breakpoint; collapsed it is 11px / 600, centered.
- **Body** (400, 0.875rem / 1.25rem): Default UI — nav labels, buttons, empty states, notices, form copy. Medium (500) at this size is the Play label and poster titles.
- **Label** (400, 0.75rem / 1rem): Episode counts, torrent metadata, helper lines under fields.

**The One Voice Rule.** Source Sans 3 Variable is the only family. No second sans, no display serif, no monospace as a voice.

## Layout

Full-height row: a thin left rail, then the field. The rail is 3rem (48px) wide, 11rem (176px) from `sm` (640px). Labels are screen-reader-only when collapsed and visible from `sm`. Settings is the last item, pushed to the bottom. The wordmark sits in the rail header.

Library owns the field: main has no page padding and does not scroll as a whole. Header is 20px horizontal, 20px top, 12px bottom. The poster grid is `auto-fill` at min 8.5rem, 12px gaps, and scrolls. Selecting a show replaces the field with a scrollable episode list; Back returns to the grid.

Inherited views (Watching, Search, Downloads, Calendar, Settings) use 24px main padding and 24px section gaps. Calendar day blocks go two columns from `lg` (1024px). Watching list goes two columns from `lg`.

Watching is an AniList list manager, not a poster field. Header carries a **Watching / Completed / Planning / Paused / Dropped / Repeating** filter row (selected = muted fill, idle = ghost — not orange). Search AniList sits in a sheet card above the list: 44px input, secondary Search, result rows with cover + title + list-status caption, and an orange **Watch** action only when the title is not already CURRENT. List cards keep the inherited row layout; **Edit** (ghost) opens a centered sheet overlay on bezel scrim where status, progress, score, dates, repeat, private, and notes can be changed — not a direct status toggle on the row. Status changes remove the row from the active filter after a successful save.

Rhythm is 8 / 12 / 16 / 20 / 24px. Interactive chrome is at least 44px tall.

**The Field-and-Rail Rule.** Named destinations live in the rail. The field is the work. Do not add a top app bar.

## Elevation & Depth

Flat. No `box-shadow`. Depth is the bezel stack (bezel → field → sheet → panel) plus 1px hairlines. Playback progress is an 8px-tall accent fill on a muted track at the bottom of the playing episode row, and a bezel status strip above main. Download progress is an 8px-tall accent fill on a muted track. Highlighted posters use a 2px accent outline, offset 2px — the same language as `:focus-visible`.

Focus is a 2px `ring` outline, offset 2px, on the same orange as primary actions. Text selection is orange on field-dark. The caret is orange. Scrollbars are thin: orange thumb on bezel track. Color changes use 200ms. `prefers-reduced-motion: reduce` collapses transition duration.

**The Flat Bezel Rule.** Surfaces are flat at rest and in motion. No drop shadow, no glow, no blur.

## Shapes

Every radius token is 0px. Posters are square (`aspect-square`). Primary buttons are rectangles (minimum 96px wide where locked, 44px tall), not pills and not circles. Episode rows are sheet-filled rectangles, 44px min height, with a left accent hairline when playing. Inputs, sheets, notices, and nav items are square-cornered rectangles.

Hairlines are 1px. Library uses full `border`. Inherited views often use that hairline at 40% opacity. Do not round the 40% stroke into a different radius.

**The Zero Radius Rule.** Controls, posters, sheets, inputs, and notices are square. Do not introduce a radius scale.

## Components

### Buttons

- **Shape:** Square (0px). Minimum height 44px. No border.
- **Play / primary:** Orange fill, field-dark type, 0 20px padding, medium 0.875rem. Search, Downloads start, and Settings save use this fill; min-width 96px only where the design locks it.
- **Secondary:** Panel fill, phosphor type, 0 16px padding (0 12px on compact chrome). Import, Search (match sheet), Download, Connect, Today, Try again. Hover steps to muted.
- **Muted:** Muted fill, phosphor type. Detect / browse path, Calendar Previous / Next. Hover steps to secondary.
- **Ghost:** Transparent, dim type, hover to phosphor. Back, Skip, Disconnect (destructive type instead of dim), Open folder.
- **Destructive fill:** Fault red, field-dark type. Cancel download.
- **Disabled:** `not-allowed` cursor, 50% opacity.
- **Hover / Focus:** Play has no fill hover; secondary and muted swap panel steps. Focus is the global 2px orange outline, offset 2px.

### Cards / Containers

- **Corner Style:** Square (0px).
- **Background:** Sheet (`card`) for match, forms, list rows, episode rows, day blocks. Bezel for rail and playback status strip.
- **Shadow Strategy:** None.
- **Border:** Full hairline on the Library match sheet. Inherited empty/error states use hairline at 40%, dashed on some empty states.
- **Internal Padding:** 16px typical; 32px on empty/error blocks; 12px on Watching rows.

### Inputs / Fields

- **Style:** Square well, muted fill (Library match, Search, Downloads) or sheet fill (Settings). Hairline stroke: full on Library, 40% on inherited views. 12px horizontal padding, 44px min height. Body type.
- **Focus:** 2px orange outline, offset 2px (`ring`).
- **Error / Disabled:** Error is copy and banners in fault red, not a glowing ring. Disabled follows the 50% opacity rule.

### Navigation

- **Style:** Bezel column, 1px hairline on the right. Wordmark in the header. Literal destination names: Library, Watching, Search, Downloads, Airing, Settings. Settings is last, `margin-top: auto`.
- **Default:** Transparent, dim type, 44px tall, left hairline transparent.
- **Hover:** Muted fill, phosphor type.
- **Current:** Muted fill, phosphor type, left hairline in orange (`border-l` + `accent`).
- **Collapsed (`< sm`):** 48px wide, labels not visible, wordmark 11px centered. Expanded: 176px, labels visible, wordmark 1.125rem left-aligned.

### Icons

Lucide, 16px (`h-4 w-4`), stroke 1.75, `currentColor`. Rail destinations and ghost actions stay outline: Library is the poster grid, Watching the eye, Search, Downloads, Airing the calendar, Settings the gear, Folder for Open folder. Play is the one filled triangle. Semantic wrappers live in `frontend/src/components/Icons.tsx` — do not import Lucide names from views.

### Poster Field

Square tiles, min 8.5rem, 12px gaps. Cover is `object-cover` on a muted square; missing cover is muted with dim title anchored at the bottom. Caption is medium body + label: AniList watch progress for linked shows (`5 / 12 · 7 left`), or local file count when not linked. Highlighted tile (currently playing show): 2px orange outline, offset 2px. No Play on the tile.

### Show Page

Opens when a poster is selected. Header: ghost Back (44px) + show title; Import stays on the right. Field is a scrollable episode list — sheet rows, 44px min height, 4px vertical gap.

For bound shows, rows follow the AniList catalog (`1..N`): slots with a local file are playable; aired gaps show dim copy `No file`; not-yet-aired slots show `Akan datang`. Unbound shows list local files only. Click a playable row to play; no separate Play button per row. The playing row gets a left accent hairline, a filled Play icon, and an 8px progress track on the row bottom. Busy row shows “Starting…” at 50% opacity. Missing and upcoming rows are not buttons.

Match AniList is a sheet above the field, never inside the episode list.

### Notices

Fixed, 16px from the bottom-right. Secondary fill for info; destructive fill for errors. Body type, 16px 12px padding, max 24rem. Init failure is a full-width destructive bar above main, not a toast.

## Do's and Don'ts

### Do:

- **Do** browse by square posters; open the episode list when a show is selected.
- **Do** play by clicking an episode row; no Play button on every row or poster.
- **Do** park Import and AniList match in the header / sheet, never in the episode list.
- **Do** mark the current rail destination with a left orange hairline on a muted fill.
- **Do** paint playback and download progress as orange fills on muted tracks.
- **Do** keep every corner at 0px and every primary control at least 44px tall.

### Don't:

- **Don't** put a Play control on every poster or every row.
- **Don't** round posters, Play, sheets, or inputs into pills or cards with radius.
- **Don't** fill chrome, nav, or poster tiles with orange.
- **Don't** introduce a second type family or a display face.
- **Don't** add drop shadows, blurs, or lifted cards.
- **Don't** add a top app bar; destinations stay in the named left rail.
