package widgets_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/strongo/strongo-tui/pkg/theme"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func newTrail() widgets.Breadcrumbs {
	b := widgets.NewBreadcrumbs("trail",
		widgets.Crumb{Title: "Home"}, widgets.Crumb{Title: "Projects"}, widgets.Crumb{Title: "Demo"})
	b.SetSize(40, 1)
	return b
}

func selectedIndexes(t *testing.T, cmd tea.Cmd) []int {
	t.Helper()
	var got []int
	for _, msg := range uitest.Msgs(cmd) {
		m, ok := msg.(widgets.CrumbSelectedMsg)
		if !ok || m.ID != "trail" {
			t.Fatalf("unexpected message %#v", msg)
		}
		got = append(got, m.Index)
	}
	return got
}

func TestBreadcrumbsSelectionLifecycle(t *testing.T) {
	b := newTrail()
	if b.Selected() != 2 || b.AtEdge(widgets.Left) || !b.AtEdge(widgets.Right) {
		t.Fatal("a fresh trail rests on the last crumb")
	}
	b.Focus()
	b.Focus()
	if !b.Focused() || b.Selected() != 1 {
		t.Fatal("focus lands on the parent crumb, once")
	}
	b, _ = b.Update(uitest.Key("left"))
	b.Focus()
	if b.Selected() != 0 || !b.AtEdge(widgets.Left) || b.AtEdge(widgets.Right) {
		t.Fatal("focusing again keeps the user's selection")
	}
	if !b.AtEdge(widgets.Up) || !b.AtEdge(widgets.Down) {
		t.Fatal("vertical moves always leave a one-row trail")
	}
	b.Blur()
	if b.Focused() || b.Selected() != 2 {
		t.Fatal("blur returns to the last crumb")
	}
}

func TestBreadcrumbsKeysEmitSelection(t *testing.T) {
	b := newTrail()
	b.Focus() // selection: Projects
	var cmd tea.Cmd
	b, cmd = b.Update(uitest.Key("enter"))
	if got := selectedIndexes(t, cmd); len(got) != 1 || got[0] != 1 {
		t.Fatalf("enter on the parent: %v", got)
	}
	b, _ = b.Update(uitest.Key("right"))
	b, _ = b.Update(uitest.Key("right"))
	if b.Selected() != 2 {
		t.Fatal("right stops at the last crumb")
	}
	b, _ = b.Update(uitest.Key("<"))
	b, _ = b.Update(uitest.Key("<"))
	b, _ = b.Update(uitest.Key("<"))
	if b.Selected() != 0 {
		t.Fatal("left stops at the first crumb")
	}
	b, _ = b.Update(uitest.Key(">"))
	b, cmd = b.Update(uitest.Key("enter"))
	if got := selectedIndexes(t, cmd); len(got) != 1 || got[0] != 1 {
		t.Fatalf("enter after >: %v", got)
	}
	msg := uitest.Msgs(cmd)[0].(widgets.CrumbSelectedMsg)
	if msg.Crumb.Title != "Projects" {
		t.Fatalf("the message carries the crumb: %+v", msg)
	}
	if _, cmd = b.Update(uitest.Key("x")); cmd != nil {
		t.Fatal("other keys do nothing")
	}
}

func TestBreadcrumbsIgnoreKeysUntilFocused(t *testing.T) {
	b := newTrail()
	if b2, cmd := b.Update(uitest.Key("enter")); cmd != nil || b2.Selected() != b.Selected() {
		t.Fatal("an unfocused trail ignores keys")
	}
	empty := widgets.NewBreadcrumbs("empty")
	empty.Focus()
	if _, cmd := empty.Update(uitest.Key("enter")); cmd != nil {
		t.Fatal("an empty trail has nothing to activate")
	}
	if empty.View() != "" && empty.Selected() != 0 {
		t.Fatal("an empty trail renders nothing")
	}
}

func TestBreadcrumbsSetCrumbs(t *testing.T) {
	b := newTrail()
	b.Focus()
	b.SetCrumbs([]widgets.Crumb{{Title: "A"}, {Title: "B"}, {Title: "C"}, {Title: "D"}})
	if b.Selected() != 2 || len(b.Crumbs()) != 4 {
		t.Fatalf("a focused trail keeps the parent selected: %d", b.Selected())
	}
	b.Blur()
	b.SetCrumbs([]widgets.Crumb{{Title: "Only"}})
	if b.Selected() != 0 {
		t.Fatal("a single crumb is selected")
	}
	crumbs := b.Crumbs()
	crumbs[0].Title = "changed"
	if b.Crumbs()[0].Title != "Only" {
		t.Fatal("Crumbs returns a copy")
	}
}

func TestBreadcrumbsView(t *testing.T) {
	b := newTrail()
	if got := uitest.Plain(b.View()); got != "Home > Projects > Demo" {
		t.Fatalf("view = %q", got)
	}
	b.SetSize(10, 1)
	if got := uitest.Plain(b.View()); got != "Home > Pro" {
		t.Fatalf("clipped = %q", got)
	}
	b.SetSize(4, 1)
	if got := uitest.Plain(b.View()); got != "Home" {
		t.Fatalf("clipped to the first crumb = %q", got)
	}
	b.SetSize(0, 1)
	if b.View() != "" {
		t.Fatal("no width, no trail")
	}
	b.SetSize(40, 1)
	blurred := b.View()
	b.Focus()
	if b.View() == blurred {
		t.Fatal("focus highlights the selected crumb")
	}
	b, _ = b.Update(tea.WindowSizeMsg{Width: 6, Height: 1})
	if got := uitest.Plain(b.View()); got != "Home >" {
		t.Fatalf("size messages resize the trail: %q", got)
	}
}

func TestBreadcrumbsSeparatorsColoursAndURLs(t *testing.T) {
	b := widgets.NewBreadcrumbs("t",
		widgets.Crumb{Title: "https://example.com"},
		widgets.Crumb{Title: "api", Color: theme.Green},
		widgets.Crumb{Title: "v1"})
	b.SetSeparator(" / ")
	b.SetSeparatorStart(1)
	b.SetSize(60, 1)
	if got := uitest.Plain(b.View()); got != "https://example.com api / v1" {
		t.Fatalf("view = %q", got)
	}
	c := widgets.NewBreadcrumbs("t", widgets.Crumb{Title: "Home > "}, widgets.Crumb{Title: "x"})
	c.SetSize(20, 1)
	if got := uitest.Plain(c.View()); got != "Home > x" {
		t.Fatalf("a title that ends with the separator is not doubled: %q", got)
	}
}

func TestBreadcrumbsMouse(t *testing.T) {
	b := newTrail()
	click := func(x int, button tea.MouseButton) (widgets.Breadcrumbs, tea.Cmd) {
		return b.Update(tea.MouseClickMsg(tea.Mouse{X: x, Button: button}))
	}
	var cmd tea.Cmd
	b, cmd = click(2, tea.MouseLeft)
	if got := selectedIndexes(t, cmd); len(got) != 1 || got[0] != 0 || b.Selected() != 0 {
		t.Fatalf("click on Home: %v", got)
	}
	b, cmd = click(9, tea.MouseLeft)
	if got := selectedIndexes(t, cmd); len(got) != 1 || got[0] != 1 {
		t.Fatalf("click on Projects: %v", got)
	}
	for _, x := range []int{6, 39} {
		if _, cmd = click(x, tea.MouseLeft); cmd != nil {
			t.Fatalf("click at %d is not on a crumb", x)
		}
	}
	if _, cmd = click(2, tea.MouseRight); cmd != nil {
		t.Fatal("right click ignored")
	}
	if _, cmd = b.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelUp})); cmd != nil {
		t.Fatal("wheel ignored")
	}
	narrow := newTrail()
	narrow.SetSize(3, 1)
	if _, cmd = narrow.Update(tea.MouseClickMsg(tea.Mouse{X: 2, Button: tea.MouseLeft})); cmd == nil {
		t.Fatal("a clipped crumb is still clickable where it shows")
	}
}

func TestBreadcrumbsKeyMapHelp(t *testing.T) {
	km := widgets.DefaultBreadcrumbsKeyMap()
	if len(km.ShortHelp()) != 3 || len(km.FullHelp()) != 1 {
		t.Fatal("help lists the bindings")
	}
}
