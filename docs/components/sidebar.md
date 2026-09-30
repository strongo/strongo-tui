# Sidebar

**Status:** defined · **Go:** [pkg/sidebar](../../pkg/sidebar/sidebar.go)

## Purpose

A side panel holding an ordered list of pinned entities (the working context): rows the person
wants to keep in view while the main pane changes. It is not a file tree or a navigator; it lists
`entity.Ref` values and lets the person open or drop them.

## Anatomy

```
 Pinned
▌ Team standup
  Dentist, Fri 10:00
```

- **Header:** a panel header, "Pinned" by default, replaceable with `WithTitle`.
- **Rows:** one line per entity, drawn by a renderer the caller supplies (default: the ref's
  `Title`). Empty list shows "(empty)".
- **Width:** the panel stores a chat/sidebar split percentage, clamped to 40 to 75, for the
  composing screen to use.

## States

| State | Look |
|---|---|
| Unfocused | Rows in plain text, cursor row not highlighted |
| Focused | Cursor row in the selected-row fill |
| Empty | Header and "(empty)" |
| Hidden | `Visible()` is false; the composing screen draws nothing |

## Themes

Colours come from `pkg/theme` (`PanelHeader`, `SelectedRow`); nothing is defined here.

## Rules

- Adding an entity that is already pinned (same `Type` and `Keys`) changes nothing.
- The panel emits `RemoveMsg` (x, Delete, Backspace) and `OpenMsg` (Enter); it never acts on the
  entity itself.
- Keys reach the panel only while it holds focus. Show/hide and resizing shortcuts belong to the
  composing screen.

## Go API

```go
sb := sidebar.New(func(ref entity.Ref, width int) string { return ref.Title }).WithTitle("Pinned")
sb.Add(entity.Ref{Type: "task", Keys: map[string]string{"id": "7"}, Title: "Write docs"})
out := sb.View(width, focused)
```
