# tuigoff

**Meet Mr. Tuigoff: the TUI Go Friendly Framework.**
Terminal UIs in Go that feel good to use.
*The house style behind DataTug, SpecScore, inGitDB, CodeGrapher, OVDB and Sneat CLIs.*

> The name: **TUI** + **Go** + the Russian "-off" surname ending, as in Smirnoff. Mr. Tuigoff is a
> gentleman who keeps your terminal tidy.

Design rules live in [docs/design-language.md](docs/design-language.md); the original visual
reference is [docs/design-language.html](docs/design-language.html).

[![Go CI](https://github.com/tuigoff/tuigoff/actions/workflows/ci.yml/badge.svg)](https://github.com/tuigoff/tuigoff/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/tuigoff/tuigoff/badge.svg?branch=main)](https://coveralls.io/github/tuigoff/tuigoff?branch=main)

The shared terminal UI toolkit of Sneat Co., DataTug and FileTug, built on
[Bubble Tea v2](https://charm.land/bubbletea), [Bubbles](https://charm.land/bubbles)
and [Lip Gloss](https://charm.land/lipgloss), in the Elm architecture those
libraries are made for. It provides:

- a **navigation shell** (`pkg/nav`): header with breadcrumbs, menu and content
  panels, actions bar, alerts, focus zones, a stack of pages driven by messages;
- **components** (`pkg/widgets`): list, tree, form, tabs, text pane, modal,
  breadcrumbs, frame, layout helpers, all following the conventions of
  `charm.land/bubbles`;
- the **result grid** (`pkg/grid`): a sortable, filterable, virtualised table on
  top of bubble-table, with per-cell styles, fixed columns, master/detail
  messages and lazy row sources;
- the shared **theme** (`pkg/theme`), **focus ring** (`pkg/focus`), **syntax
  highlighting** (`pkg/highlight`) and the **test helpers** (`pkg/uitest`,
  `pkg/nav/navtest`).

<!-- dev-approach:v1 -->
## Our approach to development

We build with our own tooling:

- **[SpecScore](https://specscore.md)** — specify requirements as `SpecScore.md` artifacts
- **[SpecStudio](https://specscore.studio)** — author & manage specs across their lifecycle
- **[inGitDB](https://ingitdb.com)** — store structured data in Git where applicable
- **[DALgo](https://dalgo.io)** — data access layer for Go
- **[cover100.dev](https://cover100.dev)** — drive toward 100% test coverage
- **[DataTug](https://datatug.io)** — query & explore data
<!-- /dev-approach -->

## Installation

```bash
go get github.com/tuigoff/tuigoff
```

## The architecture in five rules

1. **A component is a model.** `Update(tea.Msg) (T, tea.Cmd)` and `View() string`;
   state lives in the value; setters have pointer receivers; the parent gives it
   a size with `SetSize` or a `tea.WindowSizeMsg`. No component holds a pointer to
   its parent or starts a goroutine.
2. **Behaviour is a message.** A list does not call you back; it returns a
   command that delivers `widgets.ItemSelectedMsg{ID, Index, Item}`, and your
   `Update` handles it. Every message carries the component's ID, so a screen with
   two lists can tell them apart.
3. **Asynchronous work is a command.** A load is a `tea.Cmd` that returns a
   result message. Nothing calls `Send`, nothing mutates the UI from a goroutine.
4. **Navigation is a message.** A screen returns `nav.Push(page)`,
   `nav.Pop()`, `nav.Alert(...)`; the shell owns the stack, the breadcrumbs, the
   focus and the layout.
5. **Keys are bindings.** `key.Binding` and `key.Matches`, collected in a `KeyMap`
   that implements `help.KeyMap`, so keys can be rebound and shown in the actions
   bar.

## A screen

```go
type projects struct {
	grid    *grid.Model
	loading bool
}

func newProjects() projects { return projects{loading: true} }

func (p projects) Init() tea.Cmd { return loadProjects() }        // async: a Cmd

func (p projects) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:                                       // size from the shell
		if p.grid != nil { p.grid.SetSize(msg.Width, msg.Height) }
	case nav.ScreenFocusMsg:                                      // focus from the shell
		if p.grid != nil { p.grid.SetFocused(msg.Focused) }
	case projectsLoaded:                                          // the result of the Cmd
		p.loading = false
		p.grid = newProjectGrid(msg.rows)
	case grid.RowActivatedMsg:                                    // a message from a component
		return p, nav.Push(nav.Page{Title: title(msg.Row), Content: newProject(msg.Row)})
	case tea.KeyPressMsg:
		if p.grid != nil { _, cmd := p.grid.Update(msg); return p, cmd }
	}
	return p, nil
}

func (p projects) View() string { /* p.grid.View(w, focused) or "Loading..." */ }
func (p projects) Title() string { return "Projects" }             // optional: border title
```

Wire it up:

```go
shell := nav.New(nav.Page{Title: "Home", Menu: newMenu(), Content: nav.Static("Welcome", "...")})
if err := nav.Run(shell); err != nil { log.Fatal(err) }
```

A complete runnable application is in [`examples/demo`](examples/demo): a menu
list, a grid of people loaded asynchronously, a drill-down page, a tree, a form
and highlighted text.

## The shell (`pkg/nav`)

```
+--------------------------------------------------------------+
| Home > Projects > Demo                             (l) Login |  header
+----------------+---------------------------------------------+
| menu (30 cols) | content                                     |  body
+----------------+---------------------------------------------+
| enter open  ctrl+q quit  f1 help                             |  actions bar
+--------------------------------------------------------------+
```

The shell draws the border of each panel, takes its title from a screen that
implements `Titled`, and tells the screen the size of the area inside the border
with a `tea.WindowSizeMsg`. The menu is hidden below 100 columns.

**Pages** form a stack; their titles are the breadcrumbs. `Push`, `Pop`, `PopTo`,
`Replace`, `Reset` change the stack; `SetPanels` swaps the screens of the current
page; `SetBreadcrumbs` shows a custom trail; `SetFocus` moves focus; `Alert` and
`ShowError` show a modal or an error; `SetActions` replaces the application's
actions. A page with a nil `Menu` keeps the menu of the page below.

**Optional interfaces** a screen may implement: `Titled`, `Borderless`,
`widgets.Boundary` (should this arrow key leave the screen?), `widgets.Editor`
(the screen is editing text), `KeyCapturer` (the screen claims a key the shell
would take), and `ShortHelper` (bindings listed in the actions bar).

**Focus** lives in four zones: breadcrumbs, login button, menu, content. Arrow keys
move it when the focused screen is `AtEdge` in that direction: Up at the top goes
to the breadcrumbs, Right from the menu to the content, Left from the content to
the menu, Shift+Tab to the breadcrumbs, Right past the last crumb to the login
button, Down from the header back where focus came from. The shell sends a screen
`nav.ScreenFocusMsg{Focused}` when it gains or loses focus; forward it to your
components' `Focus`/`Blur`. Screens that do not implement `Boundary` keep their
arrow keys.

**Keys**: `Ctrl+Q` always quits, `Ctrl+C` quits unless the focused screen claims
it. `nav.WithActions(nav.Action{ID, Binding, Msg})` binds application-wide keys to
messages; they are ignored while the focused screen is editing text. The shell
reports `nav.LoginMsg` (header button) and `nav.HelpMsg` (F1).

**Mouse** is on: clicks focus and activate, the wheel scrolls; a screen receives
mouse messages with coordinates relative to its own top-left cell.

## Components (`pkg/widgets`)

| Component | Built on | Messages |
|---|---|---|
| `List`, `MenuItem` | `bubbles/list` | `ItemHighlightedMsg`, `ItemSelectedMsg` |
| `Tree`, `TreeNode` | own (data-driven; `bubbles/tree` has no unselectable nodes, ID-keyed expansion or exact sizing) | `NodeHighlightedMsg`, `NodeSelectedMsg` |
| `Form`, `Field`, `FormButton` | `bubbles/textinput` | `FieldChangedMsg`, `SubmitMsg`, `CancelMsg`, `ButtonPressedMsg` |
| `TextPane` | `bubbles/viewport` | |
| `Tabs` | own strip | `TabChangedMsg`, `TabCloseMsg` |
| `Modal` | own | `ModalDoneMsg` |
| `Breadcrumbs`, `Crumb` | own | `CrumbSelectedMsg` |
| `Frame`, `Button`, `Split`, `Fit`, `Overlay`, `Center` | pure functions and values | |

Every component has a `KeyMap` of `key.Binding`s, `ShortHelp`/`FullHelp`,
`SetSize`, `Focus`/`Blur`/`Focused`, and an `AtEdge` query where navigation applies.
The conventions are documented in `pkg/widgets/doc.go`.

## The grid (`pkg/grid`)

The result grid moved here from `strongo/aichat` so that every product can use it
without depending on the chat surface. See `pkg/grid/doc.go`.

## Testing

Components are tested by driving `Update` with messages and asserting on the
view and on the messages they emit:

```go
list, cmd := list.Update(uitest.Key("enter"))
msgs := uitest.Msgs(cmd) // []tea.Msg{widgets.ItemSelectedMsg{ID: "menu", Index: 0, ...}}
got := uitest.Plain(list.View())
```

Whole screens run under `navtest`, a thin harness over the Elm loop:

```go
h := navtest.New(t, nav.Page{Title: "Home", Menu: menu, Content: welcome})
h.Press("down", "enter")             // keys, typing, clicks, wheel, resize
h.RequireContains("Projects")
h.Send(projectsLoaded{rows: rows})   // any message
h.Advance(3 * time.Second)           // fake clock for alert timers
```

`nav.Run` starts the program through a seam (`runTeaProgram`), so no test touches a
terminal.

## Development

```bash
go build ./... && go vet ./... && go test ./...
go run ./examples/demo
```

CI requires 100% statement coverage of every package.

## License

Apache License 2.0 - see [LICENSE](LICENSE).
