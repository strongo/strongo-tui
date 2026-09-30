# Tuigoff TUI design language

Shared by every Sneat-ecosystem terminal app: DataTug, SpecScore, inGitDB, CodeGrapher, OVDB and the
Sneat CLI. Dense and data-first, but calm enough for long sessions. The terminal is a canvas:
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

## Palette (roles)

| Role | Hex |
|---|---|
| Main canvas | `#080D12` |
| Side surface | `#101820` |
| Raised / card | `#16212B` |
| Faint border | `#293A48` |
| Focus | `#20A4FF` |
| Attention | `#E9C34B` |
| Ready | `#3BD17F` |
| Text / dim text | `#D9E2EA` / `#8FA0AE` |

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
