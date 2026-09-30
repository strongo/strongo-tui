# Components

The building blocks every Sneat CLIs terminal app draws from. Each component has one page that says
what it is, how it looks in both themes, when to use it, and where the Go code lives. The rules
they all follow are in the [design language](../design-language.md).

**Status:** `defined` = the design is written down and the code exists; `planned` = named, page not
written yet.

## Scope: tuigoff renders, products compose

tuigoff owns **representation**: how things look, lay out, focus and scroll. A component takes plain
data (text, labels, IDs, styles, callbacks) and knows nothing about LLMs, sessions, databases or any
product. Higher-level components that need that knowledge, such as a full chat screen, live in the
product or library that owns it (for example `strongo/aichat`). They wrap or compose tuigoff
components and never re-implement their rendering. Dependencies point one way: products → aichat →
tuigoff.

A quick test for where code belongs: if it would still make sense in a terminal app that has never
heard of an LLM, it is tuigoff.

## Surfaces

| Component | What it is | Status |
|---|---|---|
| [Card](card.md) | A filled, role-coloured block that holds one unit of content | defined |
| [Card with ears](card-with-ears.md) | A card with a small header tab (the "ear") on its top edge | defined |

## Conversation

| Component | What it is | Status |
|---|---|---|
| [User message](user-message.md) | The person's turn: a card with a "You" ear | defined |
| [Agent message](agent-message.md) | The assistant's turn: a card with the product's name in the ear | defined |

## Next up (planned)

Composer, status bar, top bar with project selector, explorer tree with filter and counts, data
grid, tabs, modal, form, breadcrumbs, alert. Each gets a page here before its code is called stable.

## Writing a component page

Use the same headings as the existing pages: **Purpose**, **Anatomy**, **States**, **Themes**,
**Rules**, **Go API**. Show every state in light and dark. Say what the component is *not* for.
