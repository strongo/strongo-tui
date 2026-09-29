// Command demo is a complete strongo-tui application in the Elm architecture: a
// menu list, a grid loaded asynchronously, a drill-down page, a tree, a form and
// highlighted text, all inside the navigation shell.
//
// Read it top to bottom as the pattern for a screen:
//
//   - a screen is a value with Init, Update and View;
//   - Update forwards messages to its components, gives them the size and the
//     focus the shell announces, and reacts to the messages the components emit;
//   - navigation, alerts and asynchronous loads are commands.
package main

import (
	"fmt"
	"os"
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/strongo/strongo-tui/pkg/grid"
	"github.com/strongo/strongo-tui/pkg/highlight"
	"github.com/strongo/strongo-tui/pkg/nav"
	"github.com/strongo/strongo-tui/pkg/theme"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

// run and exit are seams: tests replace them so no terminal is ever started.
var (
	run  = nav.Run
	exit = os.Exit
)

const configYAML = `name: demo
visibility: public
tables:
  - users
  - orders
`

// --- menu ---------------------------------------------------------------

// menu is the left panel: a list of screens. Selecting an item pushes a page.
type menu struct{ list widgets.List }

func newMenu() menu {
	return menu{list: widgets.NewList("menu",
		widgets.MenuItem{ID: "people", Label: "People", Detail: "loaded asynchronously", Shortcut: 'p'},
		widgets.MenuItem{ID: "tree", Label: "Tree", Shortcut: 't'},
		widgets.MenuItem{ID: "form", Label: "Form", Shortcut: 'f'},
		widgets.MenuItem{ID: "text", Label: "Text", Shortcut: 'x'},
	)}
}

func (m menu) Init() tea.Cmd { return nil }

func (m menu) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case nav.ScreenFocusMsg:
		if msg.Focused {
			m.list.Focus()
		} else {
			m.list.Blur()
		}
		return m, nil
	case widgets.ItemSelectedMsg:
		return m, open(msg.Item.(widgets.MenuItem).ID)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m menu) View() string                      { return m.list.View() }
func (m menu) Title() string                     { return "Screens" }
func (m menu) AtEdge(dir widgets.Direction) bool { return m.list.AtEdge(dir) }
func (m menu) ShortHelp() []key.Binding          { return m.list.ShortHelp() }

// open returns the navigation command for a menu item.
func open(id string) tea.Cmd {
	switch id {
	case "people":
		return nav.Push(nav.Page{Title: "People", Content: newPeople()})
	case "tree":
		return nav.Push(nav.Page{Title: "Tree", Content: newTree()})
	case "form":
		return nav.Push(nav.Page{Title: "Form", Content: newForm()})
	default:
		return nav.Push(nav.Page{Title: "Text", Content: newText()})
	}
}

// --- people: a grid loaded asynchronously, then a drill-down -------------

type person struct {
	ID    int
	Name  string
	Score int
}

// peopleLoaded is the result of the load command.
type peopleLoaded struct{ people []person }

// loadPeople stands for a database or network call: it runs off the event loop
// and reports back with a message.
func loadPeople() tea.Cmd {
	return func() tea.Msg {
		people := make([]person, 0, 200)
		for i := 1; i <= 200; i++ {
			people = append(people, person{ID: i, Name: fmt.Sprintf("Person %d", i), Score: i * 7 % 100})
		}
		return peopleLoaded{people: people}
	}
}

type people struct {
	grid    *grid.Model
	loading bool
	w, h    int
	focused bool
}

func newPeople() people { return people{loading: true} }

func (p people) Init() tea.Cmd { return loadPeople() }

func (p people) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.w, p.h = msg.Width, msg.Height
		if p.grid != nil {
			p.grid.SetSize(p.w, p.h)
		}
		return p, nil
	case nav.ScreenFocusMsg:
		p.focused = msg.Focused
		if p.grid != nil {
			p.grid.SetFocused(p.focused)
		}
		return p, nil
	case peopleLoaded:
		p.loading = false
		p.grid = newPeopleGrid(msg.people)
		p.grid.SetSize(p.w, p.h)
		p.grid.SetFocused(p.focused)
		return p, nil
	case grid.RowActivatedMsg:
		return p, nav.Push(nav.Page{Title: fmt.Sprint(msg.Row.Values[1]), Content: newDetail(msg.Row)})
	}
	if p.grid == nil {
		return p, nil
	}
	_, cmd := p.grid.Update(msg)
	return p, cmd
}

func newPeopleGrid(list []person) *grid.Model {
	columns := []grid.Column{{Name: "ID", Numeric: true}, {Name: "Name"}, {Name: "Score", Numeric: true}}
	rows := make([]grid.Row, len(list))
	for i, p := range list {
		rows[i] = grid.Row{Key: strconv.Itoa(p.ID), Values: []any{p.ID, p.Name, p.Score}}
	}
	return grid.New(columns, rows,
		grid.WithID("people"),
		grid.WithoutFrame(),
		grid.WithoutSwitcher(),
		grid.WithFixedColumns(1),
		grid.WithCellStyle(func(_ grid.Row, column int, value any) lipgloss.Style {
			if score, ok := value.(int); ok && column == 2 && score >= 90 {
				return lipgloss.NewStyle().Foreground(theme.Green)
			}
			return lipgloss.NewStyle()
		}),
	)
}

func (p people) View() string {
	if p.grid == nil {
		return widgets.Fit("Loading people…", p.w, p.h)
	}
	return widgets.Fit(p.grid.View(p.w, p.focused), p.w, p.h)
}

func (p people) Title() string { return "People" }

func (p people) AtEdge(dir widgets.Direction) bool {
	return p.grid == nil || p.grid.AtEdge(dir)
}

func (p people) Editing() bool { return p.grid != nil && p.grid.Editing() }

// detail shows the row that was activated as highlighted YAML.
type detail struct{ pane widgets.TextPane }

func newDetail(row grid.Row) detail {
	yaml := fmt.Sprintf("id: %v\nname: %v\nscore: %v\n", row.Values[0], row.Values[1], row.Values[2])
	pane := widgets.NewTextPane("detail")
	text, _ := highlight.YAML(yaml) // the built-in YAML lexer does not fail
	pane.SetContent(text)
	return detail{pane: pane}
}

func (d detail) Init() tea.Cmd { return nil }

func (d detail) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	if focus, ok := msg.(nav.ScreenFocusMsg); ok {
		if focus.Focused {
			d.pane.Focus()
		} else {
			d.pane.Blur()
		}
		return d, nil
	}
	var cmd tea.Cmd
	d.pane, cmd = d.pane.Update(msg)
	return d, cmd
}

func (d detail) View() string                      { return d.pane.View() }
func (d detail) Title() string                     { return "Person" }
func (d detail) AtEdge(dir widgets.Direction) bool { return d.pane.AtEdge(dir) }

// --- tree ----------------------------------------------------------------

type treeScreen struct{ tree widgets.Tree }

func newTree() treeScreen {
	roots := []widgets.TreeNode{{
		ID: "projects", Text: "Projects", Unselectable: true, Children: []widgets.TreeNode{
			{ID: "alpha", Text: "alpha", Ref: "alpha", Children: []widgets.TreeNode{{ID: "alpha/readme", Text: "readme.md"}}},
			{ID: "beta", Text: "beta", Ref: "beta"},
		},
	}}
	tree := widgets.NewTree("tree", roots...)
	tree.ExpandAll()
	return treeScreen{tree: tree}
}

func (s treeScreen) Init() tea.Cmd { return nil }

func (s treeScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case nav.ScreenFocusMsg:
		if msg.Focused {
			s.tree.Focus()
		} else {
			s.tree.Blur()
		}
		return s, nil
	case widgets.NodeSelectedMsg:
		if name, ok := msg.Node.Ref.(string); ok {
			return s, nav.Alert("Project", "Opening "+name, 0, nav.FocusToContent)
		}
		return s, nil
	}
	var cmd tea.Cmd
	s.tree, cmd = s.tree.Update(msg)
	return s, cmd
}

func (s treeScreen) View() string                      { return s.tree.View() }
func (s treeScreen) Title() string                     { return "Tree" }
func (s treeScreen) AtEdge(dir widgets.Direction) bool { return s.tree.AtEdge(dir) }

// --- form ----------------------------------------------------------------

type formScreen struct{ form widgets.Form }

func newForm() formScreen {
	return formScreen{form: widgets.NewForm("new-project",
		[]widgets.Field{
			{ID: "title", Label: "Title", Kind: widgets.TextField, Width: 30},
			{ID: "location", Label: "Location", Kind: widgets.TextField, Value: "~/demo", Width: 30},
			{ID: "visibility", Label: "Visibility", Kind: widgets.SelectField, Options: []string{"Public", "Private"}},
		},
		[]widgets.FormButton{
			{ID: "create", Label: "Create", Role: widgets.SubmitRole},
			{ID: "cancel", Label: "Cancel", Role: widgets.CancelRole},
		})}
}

func (s formScreen) Init() tea.Cmd { return nil }

func (s formScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case nav.ScreenFocusMsg:
		if msg.Focused {
			s.form.Focus()
		} else {
			s.form.Blur()
		}
		return s, nil
	case widgets.SubmitMsg:
		title := msg.Values["title"]
		if title == "" {
			return s, nav.Alert("Missing title", "Please type a title first", 0, nav.FocusToContent)
		}
		return s, nav.Alert("Created", "Project "+title+" created", 0, nav.FocusToContent)
	case widgets.CancelMsg:
		return s, nav.Pop()
	}
	var cmd tea.Cmd
	s.form, cmd = s.form.Update(msg)
	return s, cmd
}

func (s formScreen) View() string                      { return s.form.View() }
func (s formScreen) Title() string                     { return "New project" }
func (s formScreen) AtEdge(dir widgets.Direction) bool { return s.form.AtEdge(dir) }
func (s formScreen) Editing() bool                     { return s.form.Editing() }

// --- text ----------------------------------------------------------------

func newText() nav.Screen {
	pane := widgets.NewTextPane("config")
	text, _ := highlight.YAML(configYAML)
	pane.SetContent(text)
	return detail{pane: pane}
}

// --- application ---------------------------------------------------------

// newApp builds the shell: the menu on the left, a welcome text on the right.
func newApp() nav.Model {
	return nav.New(nav.Page{
		Title:   "Demo",
		Menu:    newMenu(),
		Content: nav.Static("Welcome", "Pick a screen from the menu."),
		Focus:   nav.FocusToMenu,
	})
}

func main() {
	if err := run(newApp()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		exit(1)
	}
}
