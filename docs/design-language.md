# Tuigoff TUI design language

Shared by every Sneat Co. terminal app: DataTug, SpecScore, inGitDB, CodeGrapher, OVDB and the
Sneat Co. CLIs. Dense and data-first, but calm enough for long sessions. The terminal is a canvas:
background changes define regions, borders define structure, and bright colour is reserved for state
and focus.

The visual reference, with a full sample screen, is [design-language.html](design-language.html). It
was drawn for DataTug, so treat its product names as examples; the rules below are the generic part.

## Principles

1. **Surfaces before boxes.** Side panels are borderless and slightly lighter than the main canvas.
2. **Borders mean structure.** Regions are separated by space and background; internal separators
   are used sparingly.
3. **Colour means state.** Blue is focus/selection, yellow is warning/attention, green is
   ready/success.
4. **Dense, not cramped.** One-cell gaps and padding. Tables may be dense; controls need room.

## Themes: light and dark

Light and dark are both first-class. The theme is detected from the terminal's real background
(the terminal is asked for it once at start-up), falling back to dark when it does not answer, and
an app may override it with `theme.SetDark`. Every colour below is read at draw time, so a change
takes effect on the next frame.

Surfaces are not fixed hex values. A card's fill is the terminal's actual background blended 20% of
the way toward the role's hue, so it reads as "that role, tinted onto this terminal" on a plain dark
grey, a bright white or a hued theme such as Solarized. Text on it is whichever of a near-white or a
near-black has the higher contrast.

| Token (`pkg/theme`) | Dark | Light |
|---|---|---|
| Focus / selection, `FocusColor` | `#6FB1FF` | `#1A5FC7` |
| Text on a focused surface | `#0B1220` | `#FFFFFF` |
| Muted text and unfocused borders, `MutedColor` | `#A3A8B1` | `#54544C` |
| Accent for keys and shortcuts, `AccentColor` | `#E8C25A` | `#734B00` |
| Error text, `ErrorColor` | `#FF6B6B` | `#B3261E` |
| Top / status bar background | `#2E3440` | `#D8DCE6` |
| Top / status bar text | `#ECEFF4` | `#101820` |

Role hues, the same in both themes (they only tint the card fill):

| Role | Hue |
|---|---|
| user | `#4A7FB5` |
| assistant | `#9A9488` |
| system | `#C7A934` |
| error | `#B5544A` |
| block | `#8B93A8` |

The [reference screen](design-language.html) was drawn in dark only, with its own sample colours
(canvas `#080D12`, focus `#20A4FF`, ready `#3BD17F`); treat those as illustration and the tables
above as the tokens.

## Layout and borders

- One terminal cell of gap between regions and as the outer margin; the canvas stays visible between
  panels. Side panels never touch the terminal edge or each other.
- No outer boxes around side panels or the status bar. The central workspace has no side borders.
- Tabs and the composer/input area use horizontal separators only.
- Strong (blue) borders are reserved for focused controls or structured content, never permanent
  panel decoration. Avoid boxes around boxes.

## Components

- **Filtered counts:** `Tables (2/11)`: the numerator (visible matches) is bright/bold, the
  denominator (total) is dim.
- **Search field:** replaces the static heading it filters; focus shows as a blue underline; fuzzy
  match across all node kinds.
- **Selectors:** a top-bar selector keeps an explicit `▾` affordance; the closed state is compact
  and the menu opens downward over the workspace.
- **Message ears:** the ear background matches the message card, rounded top corners only, no
  avatars.
- **Persistent input:** the composer belongs to the main-area shell and stays pinned to the bottom
  on every screen, separated only by a top rule.
- **Trees:** sections are root nodes (no redundant project-root row); `▸`/`▾` for collapsed/expanded.
- **Status bar:** borderless; `● Ready` in green, context on the left, key hints on the right.
