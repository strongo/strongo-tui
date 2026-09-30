# Markdown

**Status:** defined · **Go:** `mdrender.Render` in [pkg/mdrender](../../pkg/mdrender/mdrender.go)

## Purpose

Turns Markdown text into styled terminal text for the body of a [card](card.md), typically an
[agent message](agent-message.md) or a document block.

## Rules

- Style follows the app theme (light or dark) automatically; pass a style name only to override it.
- Never fail visibly: if rendering errors, the raw text is shown instead.
- Width is the card's inner width, so wrapping matches the card.
- Plain text only comes out, no images; links show their text and URL.

## Go API

```go
out := mdrender.Render(text, width)
```
