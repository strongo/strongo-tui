# Agent message

**Status:** defined · **Go:** `theme.RoleAssistant` with `theme.Card`; transcript entries in `strongo/aichat` (transcript code is moving into tuigoff; see the scope note in the [index](README.md))

## Purpose

The assistant's turn: an answer, possibly streamed, possibly containing a grid, a query or a code
block.

## Anatomy

A [card with an ear](card-with-ears.md), role `assistant`, at the left, using the full transcript
width.

```
 DataTug  10:24
╭──────────────────────────────────────╮
│ Top 5 customers by total invoice:    │
│  # Customer          Total           │
│  1 Helena Holý       49.62           │
╰──────────────────────────────────────╯
```

- **Ear label:** the *product's* name (DataTug, SpecScore …), not "Assistant". The default label
  "Assistant" is only a fallback.
- **Body:** Markdown rendered for the terminal. Rich results (a grid, an HTTP response) are their own
  `block` entries placed in the same frame.
- **Streaming:** text is appended as it arrives; the card grows downward and the view follows unless
  the person has scrolled up.

## States

| State | Look |
|---|---|
| Streaming | `assistant` fill, body still growing, no time in the ear until it finishes |
| Done | Ear shows the finish time |
| Focused | Focus fill |
| Error | Use the `error` role card instead; never colour a normal answer red |

## Themes

The `assistant` fill sits between the canvas and the focus colour in both themes and keeps its text
contrast above the accessibility floor.

## Rules

- One card per answer. Tool output and grids get their own `block` cards below it.
- Do not put the model name in the ear; it belongs in the composer bar.
- Never show partial Markdown that would reflow when it completes; render on block boundaries.

## Go API

```go
out := theme.Card(theme.RoleAssistant, "DataTug", body, width, focused)
```
