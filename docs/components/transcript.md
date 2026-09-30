# Transcript

**Status:** defined · **Go:** [pkg/transcript](../../pkg/transcript/transcript.go)

## Purpose

The scrolling history of a conversation or log: an ordered list of entries, each drawn as a
[card](card.md), with a focus cursor that walks the entries that can take focus. It is not a chat
screen: it has no composer, no network and no idea what produces its text.

## Anatomy

```
▌ You                      user entry (plain text)
▌ Show me the top 5 …

▌ Assistant                assistant entry (text or Markdown)
▌ Here are the five …

  ╭ results ─────────╮     a self-framed block (a data grid) draws its own border
  ╰──────────────────╯

▌ error: …                 system / error entry
```

- **Entries:** `Role` (`user`, `assistant`, `system`), `Text`, optional `Markdown`, or a `Block`.
- **Blocks:** rich entries that render and handle keys themselves. Optional capabilities decide how
  they are framed: `Titled` (card header), `Roled` (card colour), `SelfFramed` (no card, the block
  draws its own border), `EscCapturer`, `WheelConsumer`, `EntityBlock` (reports the row under the
  cursor as an [entity reference](#entity-reference)).
- **Spacing:** one blank line between entries.

## States

| State | Look |
|---|---|
| Unfocused | Every card in its role fill |
| Focused stop | That card in the focus fill with the marker column filled; the viewport scrolls it into view |
| Streaming | The last entry grows as deltas are appended; only it is re-rendered |
| Scrolled up | New content does not pull the view back down; it follows only when already at the bottom |

## Themes

All colour comes from `pkg/theme`; the transcript adds none. Both themes are covered by the card
and role pages.

## Rules

- A Block that is prose-like sits in a card; a grid-like Block draws its own border and gets no
  card. Never both.
- The transcript never knows what a Block contains. It routes keys to the focused Block only and
  broadcasts every other message to all Blocks (or to one entry when the message implements
  `Targeted`).
- Rendering is cached per entry and invalidated on content, width or focus change.

## Entity reference

`entity.Ref` (`pkg/entity`) is the plain-data reference a Block reports through `Current()`:
`Type`, `Keys` and a display `Title`. `Same` compares `Type` and `Keys` and ignores `Title`. Layers
above tuigoff alias or convert their own reference type to it.

## Go API

```go
tr := transcript.New(transcript.WithMarkdownRenderer(mdrender.Render))
tr.SetSize(width, height)
tr.Append(transcript.Entry{Role: transcript.RoleUser, Text: "hello"})
tr.AppendDelta("a1", "streamed ")   // creates or extends an assistant entry
tr.Focus(0)                          // first focusable entry
out := tr.View()
```
