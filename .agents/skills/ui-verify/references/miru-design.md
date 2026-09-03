# Miru Design System Reference

Extracted from AGENTS.md and `frontend/src/style.css`. Use this as the
authoritative checklist during visual verification.

## Layout & Shape

- **Border radius:** zero on all elements. Every `--radius-*` token is `0px`.
  If any element has rounded corners, it is a violation.
- **Control height:** interactive controls (buttons, inputs, selects) are 44px
  tall.
- **Library poster grid and episode list** are custom layouts. Do not expect
  them to follow generic shadcn patterns.

## Color

Dark-only theme. All colors come from CSS custom properties and Tailwind
tokens. No raw hex values in components.

| Token | Hex | Usage |
|-------|-----|-------|
| `--background` | `#111111` | Page background |
| `--foreground` | `#f2f2f2` | Primary text |
| `--card` | `#181818` | Card/surface background |
| `--secondary` / `--muted` | `#1c1c1c` | Subtle surfaces |
| `--muted-foreground` | `#c2c2c2` | Secondary text |
| `--primary` / `--accent` / `--ring` | `#a78bfa` | Hit color only |
| `--destructive` | `#ef4444` | Errors |
| `--border` / `--input` | `#2a2a2a` | Borders |
| `--bezel` / `--sidebar` | `#0a0a0a` | Deepest surfaces |

**Purple rule:** `#a78bfa` (accent/primary) is reserved for active/hit states
— Play button, progress bar, selected mark, focus ring, caret, selection
highlight. It must not appear as a general decoration, background, or
non-interactive element color.

**Tailwind tokens:** use `bg-background`, `text-muted-foreground`,
`border-border`, etc. — never raw hex in JSX or CSS modules.

## Typography

- **Font family:** Source Sans 3 Variable (`--font-sans`). No other font
  families anywhere in the app.
- **Font smoothing:** `font-smoothing: antialiased` on body text.
- **Optical sizing:** enabled via the variable font.
- **Headings:** use `text-balance` for balanced line breaks.
- **Selection color:** accent background (`#a78bfa`) with accent-foreground
  text.

## Focus & Accessibility

- **Focus ring:** `outline: 2px solid var(--ring)` (`#a78bfa`) with
  `outline-offset: 2px`. Visible on all focusable elements via `:focus-visible`.
- **Scrollbar:** thin, accent-colored thumb on bezel track.
- **Live regions:** dynamic content updates (toasts, loading states) must use
  `aria-live` or equivalent so screen readers announce changes.
- **Interactive elements:** must have accessible names (visible text,
  `aria-label`, or `aria-labelledby`).

## Icons

Import from `frontend/src/components/Icons.tsx`. Do not import Lucide
directly — the Icons file re-exports and may customize them.

## CSS Rules

- No CSS modules. All styling via Tailwind utility classes.
- No new hex color values. Use existing tokens only.
- No new font families.
