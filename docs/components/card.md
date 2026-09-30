# Card

**Status:** defined · **Go:** `theme.Card` in [pkg/theme](../../pkg/theme/theme.go)

## Purpose

The one container for a unit of content: a message, a result grid, a query, a system notice. Every
message and block in a Sneat CLIs app renders through this component, so they all look related.

## Anatomy

```
▌ Header (optional, bold)
▌ Body line one
▌ Body line two, can hold a grid, markdown or any nested content
```

- **Fill:** a solid background colour chosen by the card's role. There is no border.
- **Marker column:** a narrow strip on the left. It is where focus and selection show.
- **Padding:** the same one-cell padding on the left and right for every card, so text in all cards
  starts in the same column, and the top bar aligns with it.
- **Header:** optional, bold, above the body. A default label exists per role; pass your own to
  override it.

## States

| State | Look |
|---|---|
| Default | Role fill, role text colour, marker column empty |
| Focused | Focus fill (blue family) and its contrast text; marker column filled |

## Roles

`user`, `assistant`, `system`, `error`, `block`. The role picks the fill, and nothing else about the
card changes. `block` is for rich content such as a data grid that sits in the same frame as text.

## Themes

The fill, text and focus colours each have a light and a dark value and follow the app's theme.
Contrast between fill and text is checked, not eyeballed.

## Rules

- No outer border. The fill *is* the edge; a border around a card is a box around a box.
- Never nest a card in a card. Put a grid, list or markdown *inside* one.
- One card holds one unit of content. Split long answers by block, not by inserting rules.
- Colour means state: only focus uses blue; do not use it to decorate a card.

## Go API

```go
out := theme.Card(theme.RoleAssistant, "", body, width, focused)
```

`width` is the outer width. The function returns the finished multi-line string, so it composes with
any Bubble Tea view.
