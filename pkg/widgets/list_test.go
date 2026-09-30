package widgets_test

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tuigoff/tuigoff/pkg/uitest"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

func newMenu() widgets.List {
	l := widgets.NewList("menu",
		widgets.MenuItem{ID: "a", Label: "Alpha", Shortcut: 'a'},
		widgets.MenuItem{ID: "b", Label: "Beta"},
		widgets.MenuItem{ID: "c", Label: "Gamma", Shortcut: 'g'},
		widgets.MenuItem{ID: "d", Label: "Delta"},
	)
	l.SetSize(20, 4)
	l.Focus()
	return l
}

// listMsgs describes the emitted messages as "H<index>" and "S<index>".
func listMsgs(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	var parts []string
	for _, msg := range uitest.Msgs(cmd) {
		switch m := msg.(type) {
		case widgets.ItemHighlightedMsg:
			if m.ID != "menu" || m.Item == nil {
				t.Fatalf("bad highlight %#v", m)
			}
			parts = append(parts, "H"+string(rune('0'+m.Index)))
		case widgets.ItemSelectedMsg:
			if m.ID != "menu" || m.Item == nil {
				t.Fatalf("bad selection %#v", m)
			}
			parts = append(parts, "S"+string(rune('0'+m.Index)))
		default:
			t.Fatalf("unexpected message %#v", msg)
		}
	}
	return strings.Join(parts, " ")
}

func TestListMenuItemImplementsListItem(t *testing.T) {
	var item list.DefaultItem = widgets.MenuItem{Label: "L", Detail: "D"}
	if item.Title() != "L" || item.Description() != "D" || item.FilterValue() != "L" {
		t.Fatal("MenuItem accessors")
	}
}

func TestListRendering(t *testing.T) {
	l := newMenu()
	want := "(a) Alpha\n    Beta\n(g) Gamma\n    Delta"
	if got := uitest.Plain(l.View()); got != want {
		t.Fatalf("view:\n%s", got)
	}
	for i, line := range strings.Split(l.View(), "\n") {
		if w := ansi.StringWidth(line); w != 20 {
			t.Fatalf("line %d width %d", i, w)
		}
	}
	if !strings.Contains(l.View(), "48;2;") {
		t.Fatal("the highlighted row carries the selected background")
	}
	l.Blur()
	if l.Focused() || l.View() == "" || strings.Contains(l.View(), "111;177;255") {
		t.Fatal("blurred list still renders, with the quiet highlight")
	}
}

func TestListDetailRowsAndPlainItems(t *testing.T) {
	l := widgets.NewList("menu", widgets.MenuItem{Label: "One", Detail: "first"}, plainItem("Two"), widgets.MenuItem{Label: "Three"})
	l.SetSize(12, 6)
	want := "One\nfirst\nTwo\n\nThree"
	if got := uitest.Plain(l.View()); got != want+"\n" {
		t.Fatalf("view: %q", got)
	}
	l.SetSize(0, 6)
	if l.View() != "" {
		t.Fatal("zero width renders nothing")
	}
	l.SetSize(-3, -3)
	if l.View() != "" {
		t.Fatal("negative size renders nothing")
	}
}

type plainItem string

func (p plainItem) FilterValue() string { return string(p) }

func TestListEmpty(t *testing.T) {
	l := widgets.NewList("menu")
	l.SetSize(6, 2)
	if l.View() != "  \n  "[:0]+"      \n      " {
		t.Fatalf("empty view %q", l.View())
	}
	if !l.AtEdge(widgets.Up) || !l.AtEdge(widgets.Down) || l.SelectedItem() != nil {
		t.Fatal("an empty list is at every edge")
	}
	l.Focus()
	var cmd tea.Cmd
	l, cmd = l.Update(uitest.Key("enter"))
	l, cmd2 := l.Update(uitest.Key("down"))
	if cmd != nil || len(uitest.Msgs(cmd2)) != 0 {
		t.Fatal("nothing to select")
	}
	l, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 1, Y: 0})
	if cmd != nil {
		t.Fatal("click on nothing")
	}
}

func TestListKeysMoveWithoutWrapping(t *testing.T) {
	tests := []struct {
		keys []string
		want int
		msgs string
	}{
		{[]string{"down"}, 1, "H1"},
		{[]string{"down", "down", "down"}, 3, "H1 H2 H3"},
		{[]string{"up"}, 0, ""},
		{[]string{"end"}, 3, "H3"},
		{[]string{"end", "down"}, 3, "H3"},
		{[]string{"end", "home"}, 0, "H3 H0"},
		{[]string{"pgdown"}, 0, ""},
		{[]string{"j", "k", "left", "right", "z", "?"}, 0, ""},
	}
	for _, tt := range tests {
		l := newMenu()
		var msgs []string
		for _, k := range tt.keys {
			var cmd tea.Cmd
			l, cmd = l.Update(uitest.Key(k))
			if s := listMsgs(t, cmd); s != "" {
				msgs = append(msgs, s)
			}
		}
		if l.Index() != tt.want || strings.Join(msgs, " ") != tt.msgs {
			t.Errorf("%v: index %d msgs %q, want %d %q", tt.keys, l.Index(), strings.Join(msgs, " "), tt.want, tt.msgs)
		}
	}
}

func TestListPaging(t *testing.T) {
	var items []list.Item
	for i := range 10 {
		items = append(items, widgets.MenuItem{Label: string(rune('A' + i))})
	}
	l := widgets.NewList("menu", items...)
	l.SetSize(8, 4)
	l.Focus()
	l, cmd := l.Update(uitest.Key("pgdown"))
	if l.Index() != 4 || listMsgs(t, cmd) != "H4" {
		t.Fatalf("pgdown lands on the next page: %d", l.Index())
	}
	if got := uitest.Plain(l.View()); got != "E\nF\nG\nH" {
		t.Fatalf("second page: %q", got)
	}
	l, _ = l.Update(uitest.Key("pgup"))
	if l.Index() != 0 {
		t.Fatalf("pgup back to the first page: %d", l.Index())
	}
	l, _ = l.Update(uitest.Key("end"))
	if l.Index() != 9 {
		t.Fatal("end")
	}
}

func TestListUnfocusedIgnoresKeys(t *testing.T) {
	l := newMenu()
	l.Blur()
	l, cmd := l.Update(uitest.Key("down"))
	if l.Index() != 0 || cmd != nil {
		t.Fatal("unfocused list ignores keys")
	}
	l, cmd = l.Update(uitest.Key("enter"))
	if cmd != nil {
		t.Fatal("unfocused list ignores enter")
	}
}

func TestListEnterAndShortcutsSelect(t *testing.T) {
	l := newMenu()
	l, cmd := l.Update(uitest.Key("enter"))
	if got := listMsgs(t, cmd); got != "S0" {
		t.Fatalf("enter: %q", got)
	}
	l, cmd = l.Update(uitest.Key("g"))
	if got := listMsgs(t, cmd); got != "H2 S2" || l.Index() != 2 {
		t.Fatalf("shortcut: %q index %d", got, l.Index())
	}
	l, cmd = l.Update(uitest.Key("g"))
	if got := listMsgs(t, cmd); got != "S2" {
		t.Fatalf("shortcut on the highlighted item: %q", got)
	}
	l, cmd = l.Update(uitest.Key("x"))
	if cmd != nil && len(uitest.Msgs(cmd)) != 0 {
		t.Fatal("unknown letters do nothing")
	}
	l, _ = l.Update(uitest.Key("ctrl+g"))
	l, _ = l.Update(uitest.Key("up"))
	l, _ = l.Update(tea.KeyPressMsg(tea.Key{Code: 'a', Text: "ab"}))
	if l.Index() != 1 {
		t.Fatal("multi-rune text is not a shortcut")
	}
	if item, ok := l.SelectedItem().(widgets.MenuItem); !ok || item.ID != "b" {
		t.Fatal("selected item")
	}
}

func TestListMouse(t *testing.T) {
	l := newMenu()
	l.Blur() // the mouse works regardless of focus
	l, cmd := l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 3, Y: 2})
	if got := listMsgs(t, cmd); got != "H2" {
		t.Fatalf("click on another row: %q", got)
	}
	l, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 3, Y: 2})
	if got := listMsgs(t, cmd); got != "S2" {
		t.Fatalf("click on the highlighted row: %q", got)
	}
	_, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseRight, X: 3, Y: 0})
	if cmd != nil {
		t.Fatal("only the left button clicks")
	}
	_, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 3, Y: -1})
	if cmd != nil {
		t.Fatal("above the list")
	}
	_, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 3, Y: 9})
	if cmd != nil {
		t.Fatal("below the items")
	}
	l, cmd = l.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if got := listMsgs(t, cmd); got != "H1" {
		t.Fatalf("wheel up: %q", got)
	}
	l, _ = l.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	l, cmd = l.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if l.Index() != 0 || cmd != nil {
		t.Fatal("the wheel stops at the top")
	}
	l, cmd = l.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if got := listMsgs(t, cmd); got != "H1" {
		t.Fatalf("wheel down: %q", got)
	}
	_, cmd = l.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if cmd != nil {
		t.Fatal("horizontal wheel is ignored")
	}
}

func TestListMouseOnTwoLineRowsAndPages(t *testing.T) {
	l := widgets.NewList("menu",
		widgets.MenuItem{Label: "One", Detail: "1"}, widgets.MenuItem{Label: "Two", Detail: "2"},
		widgets.MenuItem{Label: "Three", Detail: "3"}, widgets.MenuItem{Label: "Four", Detail: "4"})
	l.SetSize(10, 4)
	l, cmd := l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 0, Y: 3})
	if got := listMsgs(t, cmd); got != "H1" {
		t.Fatalf("second row spans y 2 and 3: %q", got)
	}
	l, _ = l.Update(uitest.Key("down")) // ignored: unfocused
	l.Focus()
	l, _ = l.Update(uitest.Key("down"))
	if l.Index() != 2 || !strings.Contains(uitest.Plain(l.View()), "Three") {
		t.Fatalf("the next page shows item 3: %q", uitest.Plain(l.View()))
	}
	l, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 0, Y: 2})
	if got := listMsgs(t, cmd); got != "H3" {
		t.Fatalf("click on a later page: %q", got)
	}
}

func TestListSetItemsAndSelect(t *testing.T) {
	l := newMenu()
	l.Select(3)
	if l.Index() != 3 || len(l.Items()) != 4 {
		t.Fatal("select")
	}
	l.SetItems(widgets.MenuItem{Label: "Only"}, widgets.MenuItem{Label: "Two"})
	if l.Index() != 1 {
		t.Fatalf("highlight is clamped to the last item, got %d", l.Index())
	}
	l.Select(-5)
	if l.Index() != 0 {
		t.Fatal("select clamps low")
	}
	l.Select(9)
	if l.Index() != 1 {
		t.Fatal("select clamps high")
	}
	l.SetItems()
	if l.Index() != 0 || l.SelectedItem() != nil {
		t.Fatal("no items")
	}
}

func TestListBoundary(t *testing.T) {
	l := newMenu()
	if !l.AtEdge(widgets.Up) || l.AtEdge(widgets.Down) || !l.AtEdge(widgets.Left) || !l.AtEdge(widgets.Right) {
		t.Fatal("first item")
	}
	l, _ = l.Update(uitest.Key("down"))
	if l.AtEdge(widgets.Up) || l.AtEdge(widgets.Down) {
		t.Fatal("middle")
	}
	l, _ = l.Update(uitest.Key("end"))
	if l.AtEdge(widgets.Up) || !l.AtEdge(widgets.Down) {
		t.Fatal("last item")
	}
	var _ widgets.Boundary = l
	var _ widgets.Editor = l
}

func TestListSizeMessageAndKeyMap(t *testing.T) {
	l := newMenu()
	l, cmd := l.Update(tea.WindowSizeMsg{Width: 10, Height: 2})
	if cmd != nil || len(strings.Split(l.View(), "\n")) != 2 {
		t.Fatal("window size message resizes")
	}
	km := l.KeyMap()
	km.CursorDown = km.CursorUp
	l.SetKeyMap(km)
	if l.KeyMap().ForceQuit.Enabled() || l.KeyMap().Quit.Enabled() {
		t.Fatal("quit bindings stay disabled")
	}
	if len(l.ShortHelp()) == 0 || len(l.FullHelp()) == 0 {
		t.Fatal("help")
	}
	if _, cmd := l.Update(uitest.Key("ctrl+c")); cmd != nil {
		t.Fatal("ctrl+c is left alone")
	}
}

func TestListFiltering(t *testing.T) {
	l := newMenu()
	if l.FilteringEnabled() {
		t.Fatal("filtering is opt-in")
	}
	l.SetFilteringEnabled(true)
	l.SetSize(20, 5)
	if !l.FilteringEnabled() || l.Editing() {
		t.Fatal("enabled but not editing")
	}
	l, cmd := l.Update(uitest.Key("/"))
	if !l.Editing() {
		t.Fatal("slash starts the filter")
	}
	_ = cmd
	// letters (even shortcut letters) go to the filter input
	for _, r := range "ta" {
		var c tea.Cmd
		l, c = l.Update(uitest.Text(string(r)))
		l = settleFilter(t, l, c)
	}
	if !strings.Contains(uitest.Plain(l.View()), "/ ta") {
		t.Fatalf("filter input shows: %q", uitest.Plain(l.View()))
	}
	if l.SelectedItem().(widgets.MenuItem).ID != "b" {
		t.Fatalf("filtered to Beta, got %v", l.SelectedItem())
	}
	// enter accepts the filter
	l, cmd = l.Update(uitest.Key("enter"))
	if l.Editing() || len(uitest.Msgs(cmd)) != 0 {
		t.Fatalf("enter accepts the filter without selecting")
	}
	// click while applied selects
	l, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: 1})
	if got := listMsgs(t, cmd); got != "S0" {
		t.Fatalf("click on the filtered row: %q", got)
	}
	// esc clears the filter
	l, _ = l.Update(uitest.Key("esc"))
	if len(l.Items()) != 4 || l.SelectedItem() == nil {
		t.Fatal("cleared")
	}
	// clicks are ignored while typing
	l, _ = l.Update(uitest.Key("/"))
	if _, cmd = l.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 2, Y: 1}); cmd != nil {
		t.Fatal("clicks are ignored while typing")
	}
	// SetItems ends the filter
	l.SetItems(widgets.MenuItem{Label: "New"})
	if l.Editing() {
		t.Fatal("set items ends the filter")
	}
	l.SetFilteringEnabled(false)
	if l.Editing() {
		t.Fatal("disabled")
	}
	// blurred lists are never editing
	l.SetFilteringEnabled(true)
	l, _ = l.Update(uitest.Key("/"))
	l.Blur()
	if l.Editing() {
		t.Fatal("blurred")
	}
}

// settleFilter runs the commands returned while typing a filter and feeds the
// resulting messages (the filter matches) back into the list.
func settleFilter(t *testing.T, l widgets.List, cmd tea.Cmd) widgets.List {
	t.Helper()
	for _, msg := range uitest.Msgs(cmd) {
		var c tea.Cmd
		l, c = l.Update(msg)
		_ = c
	}
	return l
}

func TestListFilterWithoutMatches(t *testing.T) {
	l := newMenu()
	l.SetFilteringEnabled(true)
	l, _ = l.Update(uitest.Key("/"))
	var c tea.Cmd
	l, c = l.Update(uitest.Text("z"))
	l = settleFilter(t, l, c)
	if l.SelectedItem() != nil || len(strings.Split(l.View(), "\n")) != 4 {
		t.Fatal("nothing highlighted, view keeps its size")
	}
	empty := widgets.NewList("menu")
	empty.SetFilteringEnabled(true)
	empty.SetSize(20, 3)
	if !strings.Contains(uitest.Plain(empty.View()), "No items") {
		t.Fatalf("notice: %q", uitest.Plain(empty.View()))
	}
}
