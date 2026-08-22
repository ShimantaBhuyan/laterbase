---
name: laterbase
description: A warm-editorial link-archiving dashboard with dark toggle
colors:
  canvas: "#faf7f2"
  surface: "#f4f0ea"
  surface-hover: "#ede8e0"
  surface-elevated: "#e6dfd5"
  border: "#dbd3c9"
  border-hover: "#cdc3b6"
  text-primary: "#1c1814"
  text-secondary: "#6b6257"
  text-muted: "#9c9388"
  accent: "#5e6ad2"
  accent-hover: "#6b77e0"
  accent-soft: "#eef0fa"
  green: "#22c55e"
  amber: "#f59e0b"
  red: "#ef4444"
  canvas-dark: "#08090a"
  surface-dark: "rgba(255, 255, 255, 0.02)"
  surface-hover-dark: "rgba(255, 255, 255, 0.04)"
  surface-elevated-dark: "rgba(255, 255, 255, 0.05)"
  border-dark: "rgba(255, 255, 255, 0.08)"
  border-hover-dark: "rgba(255, 255, 255, 0.12)"
  text-primary-dark: "#f7f8f8"
  text-secondary-dark: "#8a8f98"
  text-muted-dark: "#62666d"
typography:
  display:
    fontFamily: "Inter, system-ui, -apple-system, sans-serif"
    fontWeight: 590
    lineHeight: 1.3
    letterSpacing: "-0.02em"
  title:
    fontFamily: "Inter, system-ui, -apple-system, sans-serif"
    fontWeight: 510
    lineHeight: 1.4
    letterSpacing: "normal"
  body:
    fontFamily: "Inter, system-ui, -apple-system, sans-serif"
    fontWeight: 400
    fontSize: "14px"
    lineHeight: 1.5
    letterSpacing: "normal"
  label:
    fontFamily: "Inter, system-ui, -apple-system, sans-serif"
    fontWeight: 510
    fontSize: "12px"
    lineHeight: 1.4
    letterSpacing: "normal"
rounded:
  sm: "4px"
  md: "6px"
  lg: "8px"
  xl: "12px"
  pill: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "20px"
  xl: "24px"
components:
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text-primary}"
    rounded: "{rounded.lg}"
    padding: "{spacing.md}"
    border: "1px solid {colors.border}"
  card-hover:
    backgroundColor: "{colors.surface-hover}"
    border: "1px solid {colors.border-hover}"
  pill-badge:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text-secondary}"
    rounded: "{rounded.pill}"
    padding: "2px 10px"
    typography: "{typography.label}"
  ghost-input:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.md}"
    border: "1px solid transparent"
    padding: "8px 12px"
  ghost-input-focus:
    border: "1px solid {colors.border}"
  nav-link:
    rounded: "{rounded.md}"
    padding: "4px 12px"
    typography: "{typography.label}"
  nav-link-active:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text-primary}"
  status-dot:
    size: "8px"
    rounded: "50%"
  theme-toggle:
    backgroundColor: "{colors.surface}"
    border: "1px solid {colors.border}"
    rounded: "{rounded.md}"
    padding: "4px 8px"
---
# Design System: laterbase

## 1. Overview

**Creative North Star: "The Archive"**

Cold storage for links. The interface behaves like a well-organized archival system: each link is a filed item with a clear label, a status, and a place. Surfaces are warm paper-toned by default — quality card stock in a reading room — with the option to pull the curtains for a dark scanning mode. The indigo accent is the only color that signals action, like the librarian's stamp marking something as processed.

The system rejects bloated bookmark managers and decorative dashboards. No charts, no recommendations, no content algorithms. Every surface exists to frame a link card or a detail view. Borders are whisper-thin, backgrounds are near-transparent, and the typography carries the hierarchy.

**Key Characteristics:**
- Light-warm default, dark toggle for scanning sessions
- Translucent surfaces with whisper-thin borders
- Single indigo accent, used sparingly for interactive elements
- Typographic hierarchy over color coding
- Flat depth with one hover shadow for elevated states

## 2. Colors

A warm editorial palette by default — not beige, not sand, but the tone of uncoated paper stock under warm light. The indigo accent sits cool against the warm neutrals, creating a quiet tension that makes interactive elements unmistakable.

### Primary
- **Accent Indigo** (#5e6ad2 → oklch(0.62 0.14 270)): Buttons, interactive pills, active filters. The only saturated color on screen. Its rarity is the point — when it appears, the user knows they can act.
- **Accent Hover** (#6b77e0 → oklch(0.67 0.14 272)): Button hover state, focus rings.
- **Accent Soft** (#eef0fa → oklch(0.94 0.025 270)): Selected filter chips, hover highlights on interactive badges.

### Neutral
- **Warm Canvas** (#faf7f2 → oklch(0.97 0.006 70)): Page background. A true warm off-white, not a tinted gray.
- **Warm Surface** (#f4f0ea → oklch(0.94 0.006 70)): Card and container backgrounds. One step above canvas.
- **Warm Surface Hover** (#ede8e0 → oklch(0.91 0.006 70)): Elevated card state, interactive hover.
- **Warm Surface Elevated** (#e6dfd5 → oklch(0.88 0.006 70)): Dialogs, dropdowns, modals.
- **Warm Border** (#dbd3c9 → oklch(0.85 0.006 65)): Default card and divider borders.
- **Warm Border Hover** (#cdc3b6 → oklch(0.80 0.007 65)): Hovered card and input borders.
- **Ink** (#1c1814 → oklch(0.15 0.008 70)): Primary body text. Warm near-black, not pure #000.
- **Warm Gray** (#6b6257 → oklch(0.45 0.01 65)): Secondary text, metadata, non-interactive labels.
- **Warm Muted** (#9c9388 → oklch(0.6 0.008 65)): Placeholder text, timestamps, disabled content.

### Dark Theme
The dark palette flips the warm neutrals to cold near-blacks:
- **Dark Canvas** (#08090a): True near-black background.
- **Dark Surface** (rgba white at 2%): Translucent cards emerging from darkness.
- **Dark Text Primary** (#f7f8f8): Cool white for headings.
- **Dark Text Secondary** (#8a8f98): Cool gray for metadata.
- **Dark Accent**: Same indigo (#5e6ad2), unchanged across themes.

### System Colors
- **Green** (#22c55e): Ready / success status.
- **Amber** (#f59e0b): Fetching / processing status.
- **Red** (#ef4444): Failed / paywalled / inaccessible status.

## 3. Typography

**Display & Body Font:** Inter Variable (with system-ui and -apple-system fallback)

Inter was chosen for its generous x-height and restrained letterforms — editorial in body sizes, precise in labels. The variable axis (opsz 14–32) means display sizes tighten subtly as they scale up, while body sizes stay open and readable.

**Character:** Engineering meets editorial. Inter's mechanical precision at label sizes contrasts with its warmth at body sizes. The pairing is a single family; the tension comes from weight and tracking, not from a font switch.

### Hierarchy
- **Title** (590 weight, 18px, 1.3 line-height, -0.02em letter-spacing): Link titles and page headings. The signature weight across the app.
- **Label** (510 weight, 12px, 1.4 line-height): Pill badges, nav links, metadata. 510 is the secret weapon — visually between 500 and 600, it reads as "semi-bold" without the weight of full bold.
- **Body** (400 weight, 14px, 1.5 line-height): Summaries, descriptions, detail text. Contained to 65ch max-width in detail views.
- **Caption** (400 weight, 12px, 1.4 line-height, var(--text-muted)): Timestamps, secondary metadata, time-ago labels.

### Named Rules
**The 510 Rule.** Weight 510 is used for nav links, pill badges, and interactive labels. It is not used for headings (590) or body copy (400). Its consistency across interactive elements builds a visual grammar: "if it's 510, you can click or tap it."

## 4. Elevation

The system uses tonal layering as its primary depth mechanism: each nested surface is one step darker on the warm-neutral ramp. This keeps the interface flat by default and avoids the visual noise of multiple box-shadows.

One subtle shadow is reserved for elevated states:
- **Hover Shadow** (`box-shadow: 0 4px 12px rgba(0,0,0,0.06)`): Applied to card hover in light mode. Low blur, low opacity — enough to lift without floating.
- **Dialog Shadow** (`box-shadow: 0 8px 24px rgba(0,0,0,0.10)`): Modal dialogs and dropdown menus.
- **Dark Mode**: No shadows. Depth is conveyed entirely through surface opacity (the translucent white-on-black system).

### Named Rules
**The Flat-by-Default Rule.** Surfaces are flat at rest. The hover shadow is the only depth cue, and it appears only as a response to cursor position. Dialogs are the sole exception — they need enough shadow to separate from the page.

## 5. Components

### Cards
The system's primary container. Each card represents one link.

- **Shape:** Gently rounded corners (8px / var(--radius-lg)).
- **Background:** Warm Surface (#f4f0ea) in light mode; Dark Surface (rgba white 2%) in dark mode.
- **Border:** 1px solid Warm Border (#dbd3c9) in light; 1px solid rgba white 8% in dark.
- **Hover:** Background shifts one step darker; border shifts one step more opaque; 0 4px 12px rgba(0,0,0,0.06) shadow appears.
- **Internal Padding:** 16px all sides.
- **Child spacing:** 6px between metadata row elements, 10px between summary and badges.

### Pill Badges
Group name and source type labels.

- **Shape:** Full pill shape (9999px radius).
- **Typography:** 12px, weight 510.
- **Group badge:** Background uses group color at 15% opacity with full-opacity text in the same hue.
- **Source badge:** Default surface background with secondary text color.
- **Padding:** 2px vertical, 10px horizontal.

### Ghost Inputs
Search field and inline inputs.

- **Shape:** 6px radius.
- **Background:** Surface color at rest.
- **Border:** Transparent at rest; shifts to border color on focus.
- **Typography:** Inherit body with text-primary color.
- **Placeholder:** Text-muted color (contrast: 4.5:1 minimum against canvas).
- **Padding:** 8px 12px.

### Navigation
Top bar links.

- **Typography:** 13px, weight 510.
- **Default:** Text-secondary color, no background.
- **Active:** Surface background with text-primary color.
- **Hover:** Subtle background shift.
- **Transition:** 0.15s ease background and color.

### Status Dots
Inline circular indicators for link processing status.

- **Size:** 8px diameter.
- **Colors:** Green (ready), Amber (fetching/processing), Red (failed/inaccessible/paywalled), Muted (new).
- **Placement:** Alongside status label in card and detail views.

### Add Link Dialog
Modal overlay for pasting a URL.

- **Overlay:** Canvas color at 60% opacity.
- **Container:** Surface Elevated background, Dialog shadow, 12px radius.
- **Padding:** 24px.
- **Actions:** Accent button (Submit), ghost button (Cancel).

## 6. Do's and Don'ts

### Do:
- **Do** use the whisper-border system (1px at 8% opacity in dark, var(--border) in light). Borders are barely-there guides, not container outlines.
- **Do** keep the indigo accent on ≤10% of any given screen. Its rarity signals interactivity.
- **Do** use weight 590 for titles and 510 for interactive labels — the two-weight system is the hierarchy.
- **Do** use tonal layering for depth. One step darker per nesting level.
- **Do** cap body text at 65ch in detail views for comfortable reading.
- **Do** use `text-wrap: balance` on link titles and page headings.
- **Do** show a hover shadow on cards — 0 4px 12px rgba(0,0,0,0.06) — but only on hover, never at rest.

### Don't:
- **Don't** use box-shadows as default container decoration. Cards are flat at rest. Depth comes from tonal layering.
- **Don't** use gradient text, glassmorphism, or decorative blurs. The system is flat and honest.
- **Don't** add charts, widgets, or data visualizations. This is an archive, not a dashboard.
- **Don't** use border-left or border-right accent stripes on cards. Use the full border system or nothing.
- **Don't** let card radii exceed 12px. The 8px default keeps cards looking like paper, not bubbles.
- **Don't** pair box-shadows with borders on the same element. Pick one: the tonal border OR the hover shadow, never both simultaneously as decoration.
- **Don't** use warm beige/sand tones. The warm neutral is paper-toned (oklch 0.97 0.006 70), not sand (oklch 0.90 0.04 80).
- **Don't** add numbered section markers (01 / 02 / 03) or small-uppercase section eyebrows. The typography carries the hierarchy without scaffolding.
