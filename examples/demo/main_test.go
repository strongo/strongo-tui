package main

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/strongo/strongo-tui/pkg/nav"
	"github.com/strongo/strongo-tui/pkg/nav/navtest"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func newHarness(t *testing.T) *navtest.Harness {
	t.Helper()
	return navtest.New(t, nav.Page{
		Title:   "Demo",
		Menu:    newMenu(),
		Content: nav.Static("Welcome", "Pick a screen from the menu."),
		Focus:   nav.FocusToMenu,
	})
}

func TestMenuOpensPeopleAndLoadsThemAsynchronously(t *testing.T) {
	h := newHarness(t)
	h.RequireContains("Screens")
	h.RequireContains("Pick a screen")

	h.Press("enter") // People is highlighted first
	h.RequireContains("Demo > People")
	h.RequireContains("Person 1")
	h.RequireContains("Person 2")
	if h.Model().Depth() != 2 {
		t.Fatal("People is a page on the stack")
	}
}

func TestPeopleGridDrillsDownIntoARow(t *testing.T) {
	h := newHarness(t)
	h.Press("enter")
	h.Press("down", "down", "enter")
	h.RequireContains("Demo > People > Person 3")
	h.RequireContains("name: Person 3")
	h.Press("up")
	if h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("the pane is at its top: Up goes to the breadcrumbs")
	}
	h.Press("enter") // the parent crumb, People
	h.RequireNotContains("Demo > People > Person 3")
	h.RequireContains("Demo > People")
}

func TestShortcutsOpenScreens(t *testing.T) {
	h := newHarness(t)
	h.Press("t")
	h.RequireContains("alpha")
	h.RequireNotContains("Demo > People")

	h.Press("left", "left") // the first Left collapses alpha, the second leaves the tree
	h.Press("x")
	h.RequireContains("visibility: public")
	h.RequireContains("Demo > Tree > Text")
}

func TestTreeSelectionShowsAnAlert(t *testing.T) {
	h := newHarness(t)
	h.Press("t")
	h.Press("enter")
	h.RequireContains("Opening alpha")
	h.Press("enter")
	h.RequireNotContains("Opening alpha")
	h.Press("down", "down", "enter") // beta, then an unselectable group cannot be reached
	h.RequireContains("Opening beta")
}

func TestFormValidatesAndCancels(t *testing.T) {
	h := newHarness(t)
	h.Press("f")
	h.RequireContains("Title")
	h.Press("tab", "tab", "tab", "enter") // Create with no title
	h.RequireContains("Please type a title first")
	h.Press("enter")
	h.Press("shift+tab", "shift+tab", "shift+tab")
	h.Type("Acme")
	h.Press("tab", "tab", "tab", "enter")
	h.RequireContains("Project Acme created")
	h.Press("esc")
	h.Press("tab", "enter") // Cancel
	if h.Model().Depth() != 1 {
		t.Fatal("Cancel pops the form page")
	}
}

func TestQuitFromAnywhere(t *testing.T) {
	h := newHarness(t)
	h.Press("ctrl+q")
	if !h.Quit() {
		t.Fatal("Ctrl+Q quits")
	}
}

func TestMainRunsTheApp(t *testing.T) {
	previousRun, previousExit := run, exit
	t.Cleanup(func() { run, exit = previousRun, previousExit })

	var ran bool
	run = func(nav.Model, ...tea.ProgramOption) error { ran = true; return nil }
	exit = func(int) { t.Fatal("must not exit on success") }
	main()
	if !ran {
		t.Fatal("main must run the app")
	}

	code := -1
	run = func(nav.Model, ...tea.ProgramOption) error { return errors.New("boom") }
	exit = func(c int) { code = c }
	main()
	if code != 1 {
		t.Fatalf("exit code = %d", code)
	}
}

func TestPeopleScreenBeforeTheLoadCompletes(t *testing.T) {
	var s nav.Screen = newPeople()
	s, _ = s.Update(tea.WindowSizeMsg{Width: 40, Height: 5})
	s, _ = s.Update(nav.ScreenFocusMsg{Focused: true})
	if got := s.View(); !strings.Contains(got, "Loading people") {
		t.Fatalf("view = %q", got)
	}
	if _, cmd := s.Update(uitest.Key("down")); cmd != nil {
		t.Fatal("keys are ignored until the grid exists")
	}
	if !s.(people).AtEdge(widgets.Up) || s.(people).Editing() {
		t.Fatal("no grid: at every edge, not editing")
	}
	s, _ = s.Update(loadPeople()())
	if got := s.View(); strings.Contains(got, "Loading") || !strings.Contains(got, "Person 1") {
		t.Fatalf("after the load: %q", got)
	}
}

func TestTreeLeafSelectionDoesNothing(t *testing.T) {
	h := newHarness(t)
	h.Press("t")
	h.Press("down", "enter") // alpha/readme.md has no Ref
	h.RequireNotContains("Opening")
}

func TestLeavingTheFormBlursIt(t *testing.T) {
	h := newHarness(t)
	h.Press("f")
	h.Press("f1") // application keys step aside while the form edits text: F1 is text-safe here
	h.Press("up")
	if h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("Up from the first field leaves the form")
	}
	h.Press("down")
	h.RequireContains("Title")
}

func TestPeopleEditingWhileFilterIsOpen(t *testing.T) {
	h := newHarness(t)
	h.Press("enter")
	h.Press("f1")
}

func TestUpFromTheMenuReachesTheBreadcrumbs(t *testing.T) {
	h := newHarness(t)
	h.Press("up")
	if h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("the first item is at the top edge")
	}
}
