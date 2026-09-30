# User message

**Status:** defined · **Go:** `theme.RoleUser` with `theme.Card`

## Purpose

The person's turn in a conversation: what they typed, shown back in the transcript.

## Anatomy

A [card with an ear](card-with-ears.md), role `user`, ear label **You** and the time sent.

```
                       You  10:24
                ╭─────────────────────────╮
                │ Show me the top 5 …     │
                ╰─────────────────────────╯
```

- **Alignment:** pushed to the right, at most 70% of the transcript width, so it is distinct from the
  agent's turn at a glance.
- **Text:** plain, not rendered as Markdown; what was typed is what is shown.
- **Ear label:** "You". Do not use the person's name or initials.

## States

| State | Look |
|---|---|
| Default | `user` role fill |
| Focused | Focus fill, when the transcript cursor is on the message |
| Editing | Not defined yet |

## Themes

The `user` fill is the terminal background tinted toward a soft blue (`#4A7FB5`), in light and dark
alike, so it stays quieter than the focus colour and reads clearly against the assistant's warm grey.

## Rules

- Never truncate a user message silently; wrap it, and offer collapse for very long ones.
- The ear time is the send time in the person's locale.
- Do not add reactions, buttons or menus on the card. Actions live in the actions bar.

## Go API

```go
out := theme.Card(theme.RoleUser, "You", text, width, focused)
```
