# Grid block

**Status:** defined · **Go:** [pkg/gridblock](../../pkg/gridblock/gridblock.go)

## Purpose

Lets a data grid (`pkg/grid`) be an entry in a [transcript](transcript.md). It is a thin adapter:
the grid keeps drawing and handling keys, the adapter only says how the transcript should treat it.

## Anatomy

The grid's own bordered frame, with its title inline, exactly as on its own. The transcript draws
no card around it.

## States

Those of the grid: focused (blue border, bullet), unfocused, filtering, empty. The block is always
focusable, so the transcript cursor stops on it.

## Themes

Nothing is added; the grid follows the app theme.

## Rules

- Self-framed: no card fill around a grid, and no second frame inside it.
- The grid's messages pass through untouched, including `grid.PinRowMsg` from the `+` key. What
  pinning means (a sidebar, a bookmark) is decided by the composing screen.
- The row's `Ref` may hold an `entity.Ref` or a pointer to one; `Current()` reports it, or nil.

## Go API

```go
blk := gridblock.Wrap(grid.New(columns, rows, grid.WithTitle("results")))
tr.Append(transcript.Entry{ID: "g1", Block: blk})
ref := gridblock.EntityRef(gridModel) // *entity.Ref of the highlighted row, or nil
```
