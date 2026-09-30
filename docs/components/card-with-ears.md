# Card with ears

**Status:** defined (design) · **Go:** header row of `theme.Card`; ear tab shape is the next step

## Purpose

A [card](card.md) that names who or what it belongs to. The *ear* is a small tab on the card's top
edge carrying a label and, optionally, a time. It replaces avatars, which a terminal has no room for.

## Anatomy

```
 You  10:24
╭──────────────────────╮        ear = same fill as the card,
│ Show me the top 5 …  │        rounded on its top corners only
╰──────────────────────╯
```

- **Ear:** bold label plus a dim time. Its background is the card's background, so ear and card read
  as one shape.
- **Corners:** the ear is rounded on top only. In terminals that support truecolor the rounding is
  drawn with half-block edge cells; elsewhere the ear stays square.
- **Time:** small and dim, next to the label, never inside the body.

## States

| State | Look |
|---|---|
| Default | Ear and card share the role fill |
| Focused | Both switch to the focus fill together |

## Themes

Ear colours are the card's colours, so it needs no palette of its own in either theme.

## Rules

- The ear and the card are always the same colour. A different-coloured ear is a second component.
- Use it for conversation turns. A result grid or a plain notice does not need an ear.
- One label per ear. If you need to say more, put it in the body.

## Go API

Today the label is the card's `header` argument. The tab shape (rounded top, time slot) is not yet
its own function; that is the next code change for this page.
