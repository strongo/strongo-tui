package widgets_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func newTabStrip(style widgets.TabsStyle) widgets.Tabs {
	t := widgets.NewTabs("tabs", style,
		widgets.Tab{ID: "a", Title: "Alpha"},
		widgets.Tab{ID: "b", Title: "Beta", Closable: true},
		widgets.Tab{ID: "c", Title: "Gamma"})
	t.SetSize(40, 1)
	t.Focus()
	return t
}

func tabChanges(t *testing.T, cmd tea.Cmd) []int {
	t.Helper()
	var got []int
	for _, msg := range uitest.Msgs(cmd) {
		m, ok := msg.(widgets.TabChangedMsg)
		if !ok || m.ID != "tabs" {
			t.Fatalf("unexpected message %#v", msg)
		}
		got = append(got, m.Index)
	}
	return got
}

func TestTabsRendering(t *testing.T) {
	tests := []struct {
		name  string
		style widgets.TabsStyle
		want  string
	}{
		{"underline", widgets.UnderlineTabsStyle, " Alpha  Beta × " + " Gamma"},
		{"radio", widgets.RadioTabsStyle, " ● Alpha  ○ Beta ×  ○ Gamma"},
		{"plain", widgets.TabsStyle{}, " Alpha  Beta × " + " Gamma"},
	}
	for _, tt := range tests {
		s := newTabStrip(tt.style)
		if got := uitest.Plain(s.View()); got != tt.want {
			t.Errorf("%s: %q, want %q", tt.name, got, tt.want)
		}
		if ansi.StringWidth(s.View()) != 40 || strings.Contains(s.View(), "\n") {
			t.Errorf("%s: one row of the full width", tt.name)
		}
	}
	u := newTabStrip(widgets.UnderlineTabsStyle)
	if !strings.Contains(u.View(), "\x1b[4m") && !strings.Contains(u.View(), ";4;") && !strings.Contains(u.View(), "[4;") {
		t.Fatalf("inactive tabs are underlined: %q", u.View())
	}
	focused, blurred := u.View(), func() string { u.Blur(); return u.View() }()
	if focused == blurred || u.Focused() {
		t.Fatal("focus changes the active tab's look")
	}
}

func TestTabsLabel(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	s.SetLabel("\x1b[1mView:\x1b[m")
	if got := uitest.Plain(s.View()); !strings.HasPrefix(got, "View:  Alpha") {
		t.Fatalf("label: %q", got)
	}
	// a click on a tab takes the label's width into account
	s, cmd := s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 8, Y: 0})
	if s.Active() != 0 || len(uitest.Msgs(cmd)) != 0 {
		t.Fatal("click on the already active tab")
	}
	s, cmd = s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 17, Y: 0})
	if got := tabChanges(t, cmd); len(got) != 1 || got[0] != 1 {
		t.Fatalf("click on Beta: %v", got)
	}
}

func TestTabsSizeZero(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	s.SetSize(0, 1)
	if s.View() != "" {
		t.Fatal("zero width")
	}
	s.SetSize(-4, 1)
	if s.View() != "" {
		t.Fatal("negative width")
	}
	s, _ = s.Update(tea.WindowSizeMsg{Width: 20, Height: 5})
	if ansi.StringWidth(s.View()) != 20 {
		t.Fatal("window size message")
	}
}

func TestTabsKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want int
		msgs []int
	}{
		{"right", []string{"right"}, 1, []int{1}},
		{"right twice", []string{"right", "right"}, 2, []int{1, 2}},
		{"no wrap at the end", []string{"right", "right", "right"}, 2, []int{1, 2}},
		{"no wrap at the start", []string{"left"}, 0, nil},
		{"left", []string{"right", "left"}, 0, []int{1, 0}},
		{"jump", []string{"alt+3"}, 2, []int{2}},
		{"jump to the active tab", []string{"alt+1"}, 0, nil},
		{"jump beyond the tabs", []string{"alt+9"}, 0, nil},
		{"other keys", []string{"a", "tab", "enter", "up", "down"}, 0, nil},
	}
	for _, tt := range tests {
		s := newTabStrip(widgets.TabsStyle{})
		var got []int
		for _, k := range tt.keys {
			var cmd tea.Cmd
			s, cmd = s.Update(uitest.Key(k))
			got = append(got, tabChanges(t, cmd)...)
		}
		if s.Active() != tt.want || len(got) != len(tt.msgs) {
			t.Errorf("%s: active %d msgs %v, want %d %v", tt.name, s.Active(), got, tt.want, tt.msgs)
			continue
		}
		for i := range got {
			if got[i] != tt.msgs[i] {
				t.Errorf("%s: msgs %v, want %v", tt.name, got, tt.msgs)
			}
		}
	}
}

func TestTabsUnfocusedIgnoresKeys(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	s.Blur()
	s, cmd := s.Update(uitest.Key("right"))
	if s.Active() != 0 || cmd != nil {
		t.Fatal("unfocused strip ignores keys")
	}
}

func TestTabsClicks(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	s.Blur() // clicks work regardless of focus
	// " Alpha  Beta ×  Gamma": Alpha 0-6, Beta 7-14 (mark at 13), Gamma 15-21
	s, cmd := s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 9, Y: 0})
	if got := tabChanges(t, cmd); len(got) != 1 || got[0] != 1 || s.Active() != 1 {
		t.Fatalf("click on Beta: %v", got)
	}
	_, cmd = s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 9, Y: 0})
	if cmd != nil {
		t.Fatal("no message when the tab does not change")
	}
	_, cmd = s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 13, Y: 0})
	msgs := uitest.Msgs(cmd)
	closeMsg, ok := msgs[0].(widgets.TabCloseMsg)
	if len(msgs) != 1 || !ok || closeMsg.ID != "tabs" || closeMsg.Index != 1 || closeMsg.Tab.ID != "b" {
		t.Fatalf("close click: %#v", msgs)
	}
	s, cmd = s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 17, Y: 0})
	if got := tabChanges(t, cmd); len(got) != 1 || got[0] != 2 {
		t.Fatal("click on Gamma")
	}
	// the mark of a tab that is not closable is just the tab
	for _, c := range []tea.MouseClickMsg{
		{Button: tea.MouseRight, X: 9, Y: 0},
		{Button: tea.MouseLeft, X: 9, Y: 1},
		{Button: tea.MouseLeft, X: 30, Y: 0},
	} {
		if _, cmd := s.Update(c); cmd != nil {
			t.Fatalf("click %+v is ignored", c)
		}
	}
	if _, cmd := s.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown}); cmd != nil {
		t.Fatal("the wheel is ignored")
	}
}

func TestTabsScrollToActiveTab(t *testing.T) {
	s := widgets.NewTabs("tabs", widgets.TabsStyle{},
		widgets.Tab{Title: "Alpha"}, widgets.Tab{Title: "Beta"}, widgets.Tab{Title: "Gamma"}, widgets.Tab{Title: "Delta"})
	s.SetSize(14, 1)
	s.Focus()
	if got := uitest.Plain(s.View()); got != " Alpha  Beta" {
		t.Fatalf("first: %q", got)
	}
	s.SetActive(3)
	if got := uitest.Plain(s.View()); !strings.HasSuffix(got, "Delta") || ansi.StringWidth(s.View()) != 14 {
		t.Fatalf("active tab is visible: %q", got)
	}
	// clicks use the scrolled positions: the last column is on Delta
	s, cmd := s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 13, Y: 0})
	if cmd != nil || s.Active() != 3 {
		t.Fatal("click on the visible active tab")
	}
	s, cmd = s.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 0, Y: 0})
	if got := tabChanges(t, cmd); len(got) != 1 || got[0] != 2 {
		t.Fatalf("first visible column belongs to Gamma: %v", got)
	}
}

func TestTabsAddRemoveAndActive(t *testing.T) {
	s := widgets.NewTabs("tabs", widgets.TabsStyle{})
	if s.Active() != -1 || s.AtEdge(widgets.Left) != true || !s.AtEdge(widgets.Right) {
		t.Fatal("empty strip")
	}
	s.SetSize(20, 1)
	if s.View() != strings.Repeat(" ", 20) {
		t.Fatal("empty strip renders blank")
	}
	s, cmd := s.Update(uitest.Key("right"))
	if cmd != nil {
		t.Fatal("nothing to switch to")
	}
	s.SetActive(0)
	if s.Active() != -1 {
		t.Fatal("set active outside the strip is ignored")
	}
	s.AddTab(widgets.Tab{ID: "a", Title: "A"})
	s.AddTab(widgets.Tab{ID: "b", Title: "B"})
	s.AddTab(widgets.Tab{ID: "c", Title: "C"})
	s.AddTab(widgets.Tab{ID: "d", Title: "D"})
	if s.Active() != 0 || len(s.Tabs()) != 4 {
		t.Fatal("first added tab is active")
	}
	s.SetActive(2)
	s.SetActive(9)
	s.SetActive(-1)
	if s.Active() != 2 {
		t.Fatal("set active")
	}
	s.RemoveTab(0)
	if s.Active() != 1 || s.Tabs()[s.Active()].ID != "c" {
		t.Fatal("removing an earlier tab keeps the active tab")
	}
	s.RemoveTab(2)
	if s.Active() != 1 {
		t.Fatal("removing a later tab keeps the active tab")
	}
	s.RemoveTab(1)
	if s.Active() != 0 || s.Tabs()[0].ID != "b" {
		t.Fatalf("removing the active tab activates a neighbour: %d", s.Active())
	}
	s.RemoveTab(7)
	s.RemoveTab(-1)
	s.RemoveTab(0)
	if s.Active() != -1 || len(s.Tabs()) != 0 {
		t.Fatal("removing the last tab")
	}
	tabs := s.Tabs()
	if len(tabs) != 0 {
		t.Fatal("copy")
	}
}

func TestTabsRemoveActiveLast(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	s.SetActive(2)
	s.RemoveTab(2)
	if s.Active() != 1 {
		t.Fatalf("active moves to the new last tab: %d", s.Active())
	}
}

func TestTabsBoundary(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	if !s.AtEdge(widgets.Left) || s.AtEdge(widgets.Right) || !s.AtEdge(widgets.Up) || !s.AtEdge(widgets.Down) {
		t.Fatal("first tab")
	}
	s.SetActive(1)
	if s.AtEdge(widgets.Left) || s.AtEdge(widgets.Right) {
		t.Fatal("middle tab")
	}
	s.SetActive(2)
	if s.AtEdge(widgets.Left) || !s.AtEdge(widgets.Right) {
		t.Fatal("last tab")
	}
	var _ widgets.Boundary = s
}

func TestTabsHelp(t *testing.T) {
	s := newTabStrip(widgets.TabsStyle{})
	if len(s.ShortHelp()) != 3 || len(s.FullHelp()) != 1 {
		t.Fatal("help")
	}
}
