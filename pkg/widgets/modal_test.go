package widgets_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func newConfirmModal() widgets.Modal {
	m := widgets.NewModal("confirm")
	m.SetTitle("Confirm")
	m.SetText("Delete the file?")
	m.SetButtons("Yes", "No")
	m.SetMaxSize(80, 24)
	return m
}

func modalDone(t *testing.T, cmd tea.Cmd) []widgets.ModalDoneMsg {
	t.Helper()
	var got []widgets.ModalDoneMsg
	for _, msg := range uitest.Msgs(cmd) {
		m, ok := msg.(widgets.ModalDoneMsg)
		if !ok || m.ID != "confirm" {
			t.Fatalf("unexpected message %#v", msg)
		}
		got = append(got, m)
	}
	return got
}

func assertModalUniformWidth(t *testing.T, box string) {
	t.Helper()
	lines := strings.Split(box, "\n")
	w := ansi.StringWidth(lines[0])
	for i, l := range lines {
		if ansi.StringWidth(l) != w {
			t.Fatalf("line %d is %d wide, want %d:\n%s", i, ansi.StringWidth(l), w, uitest.Plain(box))
		}
	}
}

func TestModalBoxLayout(t *testing.T) {
	m := newConfirmModal()
	want := strings.Join([]string{
		"╭─ Confirm ────────╮",
		"│ Delete the file? │",
		"│                  │",
		"│    Yes    No     │",
		"╰──────────────────╯",
	}, "\n")
	got := uitest.Plain(m.View())
	if got != want {
		t.Fatalf("box:\n%s\nwant:\n%s", got, want)
	}
	assertModalUniformWidth(t, m.View())
	if w, h := m.Size(); w != 20 || h != 5 {
		t.Fatalf("size %dx%d", w, h)
	}
}

func TestModalBoxShapes(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(m *widgets.Modal)
		wantW       int
		wantH       int
		firstLine   string
		widthAtMost int
	}{
		{"empty", func(m *widgets.Modal) {}, 5, 2, "╭───╮", 0},
		{"title only", func(m *widgets.Modal) { m.SetTitle("Hi") }, 8, 2, "╭─ Hi ─╮", 0},
		{"text only", func(m *widgets.Modal) { m.SetText("Hello") }, 9, 3, "╭───────╮", 0},
		{"buttons only", func(m *widgets.Modal) { m.SetButtons("OK") }, 8, 3, "╭──────╮", 0},
		{"long text wraps at 60", func(m *widgets.Modal) {
			m.SetText(strings.Repeat("word ", 30))
			m.SetButtons("OK")
			m.SetMaxSize(0, 0)
		}, 0, 0, "", 64},
		{"width limit", func(m *widgets.Modal) {
			m.SetText(strings.Repeat("word ", 30))
			m.SetMaxSize(30, 0)
		}, 0, 0, "", 30},
		{"tiny width", func(m *widgets.Modal) {
			m.SetText("abcdef")
			m.SetMaxSize(3, 0)
		}, 5, 0, "", 5},
		{"height limit drops text rows", func(m *widgets.Modal) {
			m.SetText("1\n2\n3\n4\n5\n6")
			m.SetButtons("OK")
			m.SetMaxSize(0, 7)
		}, 0, 7, "", 0},
		{"height limit without buttons", func(m *widgets.Modal) {
			m.SetText("1\n2\n3\n4\n5\n6")
			m.SetMaxSize(0, 4)
		}, 0, 4, "", 0},
		{"height too small for any text", func(m *widgets.Modal) {
			m.SetText("1\n2")
			m.SetButtons("OK")
			m.SetMaxSize(0, 1)
		}, 0, 3, "", 0},
		{"long title is truncated", func(m *widgets.Modal) {
			m.SetTitle("A very long title indeed")
			m.SetText("short")
			m.SetMaxSize(14, 0)
		}, 14, 3, "", 14},
	}
	for _, tt := range tests {
		m := widgets.NewModal("m")
		tt.setup(&m)
		box := m.View()
		assertModalUniformWidth(t, box)
		lines := strings.Split(uitest.Plain(box), "\n")
		w, h := m.Size()
		if ansi.StringWidth(lines[0]) != w {
			t.Errorf("%s: Size width %d, box %d", tt.name, w, ansi.StringWidth(lines[0]))
		}
		if h != len(lines) {
			t.Errorf("%s: Size height %d, box %d", tt.name, h, len(lines))
		}
		if tt.wantW != 0 && w != tt.wantW {
			t.Errorf("%s: width %d, want %d", tt.name, w, tt.wantW)
		}
		if tt.wantH != 0 && h != tt.wantH {
			t.Errorf("%s: height %d, want %d\n%s", tt.name, h, tt.wantH, uitest.Plain(box))
		}
		if tt.firstLine != "" && lines[0] != tt.firstLine {
			t.Errorf("%s: first line %q, want %q", tt.name, lines[0], tt.firstLine)
		}
		if tt.widthAtMost != 0 && w > tt.widthAtMost {
			t.Errorf("%s: width %d exceeds %d", tt.name, w, tt.widthAtMost)
		}
	}
}

func TestModalColors(t *testing.T) {
	m := newConfirmModal()
	def := m.View()
	m.SetColors(widgets.ModalColors{
		Background: theme.LightBlue, Text: theme.Red, Border: theme.Red,
		ButtonBackground: theme.LightBlue, ButtonText: theme.Red,
	})
	coloured := m.View()
	if coloured == def || uitest.Plain(coloured) != uitest.Plain(def) {
		t.Fatal("colours change the look, not the text")
	}
	assertModalUniformWidth(t, coloured)
	bg := lipgloss.NewStyle().Background(theme.LightBlue).Render(" ")
	bgSeq, _, _ := strings.Cut(bg, " ")
	bgSeq = strings.TrimSuffix(strings.TrimPrefix(bgSeq, "\x1b["), "m")
	for i, line := range strings.Split(coloured, "\n") {
		if !strings.Contains(line, bgSeq) {
			t.Fatalf("line %d has no background", i)
		}
	}
	if strings.Contains(def, bgSeq) {
		t.Fatal("no background by default")
	}
}

func TestModalKeys(t *testing.T) {
	tests := []struct {
		name   string
		keys   []string
		active int
	}{
		{"right", []string{"right"}, 1},
		{"tab wraps", []string{"tab", "tab"}, 0},
		{"left wraps", []string{"left"}, 1},
		{"shift+tab wraps", []string{"shift+tab"}, 1},
		{"other keys are swallowed", []string{"a", "up", "down", "space", "f1", "alt+1"}, 0},
	}
	for _, tt := range tests {
		m := newConfirmModal()
		for _, k := range tt.keys {
			var cmd tea.Cmd
			m, cmd = m.Update(uitest.Key(k))
			if cmd != nil {
				t.Fatalf("%s: %s emitted", tt.name, k)
			}
		}
		if m.Active() != tt.active {
			t.Errorf("%s: active %d, want %d", tt.name, m.Active(), tt.active)
		}
	}
}

func TestModalDoneMessages(t *testing.T) {
	m := newConfirmModal()
	_, cmd := m.Update(uitest.Key("enter"))
	if got := modalDone(t, cmd); len(got) != 1 || got[0].Index != 0 || got[0].Label != "Yes" {
		t.Fatalf("enter on the first button: %v", got)
	}
	m, _ = m.Update(uitest.Key("right"))
	_, cmd = m.Update(uitest.Key("enter"))
	if got := modalDone(t, cmd); len(got) != 1 || got[0].Index != 1 || got[0].Label != "No" {
		t.Fatalf("enter on the second button: %v", got)
	}
	_, cmd = m.Update(uitest.Key("esc"))
	if got := modalDone(t, cmd); len(got) != 1 || got[0].Index != -1 || got[0].Label != "" {
		t.Fatalf("esc: %v", got)
	}
}

func TestModalWithoutButtons(t *testing.T) {
	m := widgets.NewModal("confirm")
	m.SetText("Working...")
	for _, k := range []string{"left", "right", "tab", "shift+tab", "enter", "x"} {
		var cmd tea.Cmd
		m, cmd = m.Update(uitest.Key(k))
		if cmd != nil || m.Active() != 0 {
			t.Fatalf("%s does nothing", k)
		}
	}
	_, cmd := m.Update(uitest.Key("esc"))
	if got := modalDone(t, cmd); len(got) != 1 || got[0].Index != -1 {
		t.Fatal("esc still cancels")
	}
	m, cmd = m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 3, Y: 1})
	if cmd != nil {
		t.Fatal("click on nothing")
	}
}

func TestModalLeavesQuitKeysAlone(t *testing.T) {
	m := newConfirmModal()
	for _, k := range []string{"ctrl+c", "ctrl+q"} {
		m2, cmd := m.Update(uitest.Key(k))
		if cmd != nil || m2.Active() != m.Active() {
			t.Fatalf("%s must reach the shell untouched", k)
		}
	}
}

func TestModalMouse(t *testing.T) {
	m := newConfirmModal()
	// buttons row is at y 3; Yes spans columns 4-8, No 11-14 (relative to the box)
	m2, cmd := m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 6, Y: 3})
	if got := modalDone(t, cmd); len(got) != 1 || got[0].Index != 0 || m2.Active() != 0 {
		t.Fatalf("click on Yes: %v", got)
	}
	m2, cmd = m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 13, Y: 3})
	if got := modalDone(t, cmd); len(got) != 1 || got[0].Index != 1 || got[0].Label != "No" || m2.Active() != 1 {
		t.Fatalf("click on No: %v", got)
	}
	for _, c := range []tea.MouseClickMsg{
		{Button: tea.MouseLeft, X: 6, Y: 1},
		{Button: tea.MouseLeft, X: 10, Y: 3},
		{Button: tea.MouseLeft, X: 0, Y: 3},
		{Button: tea.MouseRight, X: 6, Y: 3},
	} {
		m3, cmd := m.Update(c)
		if cmd != nil || m3.Active() != 0 {
			t.Fatalf("click %+v is swallowed", c)
		}
	}
	if _, cmd := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown}); cmd != nil {
		t.Fatal("the wheel is swallowed")
	}
}

func TestModalButtonRowClippedByMaxSize(t *testing.T) {
	m := widgets.NewModal("confirm")
	m.SetButtons("First button", "Second button")
	m.SetMaxSize(20, 0)
	assertModalUniformWidth(t, m.View())
	w, _ := m.Size()
	if w != 20 {
		t.Fatalf("width %d", w)
	}
	// the second button is mostly clipped; a click beyond the box does nothing
	_, cmd := m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 19, Y: 1})
	if cmd != nil {
		t.Fatal("click on the border")
	}
}

func TestModalWindowSizeAndSetButtons(t *testing.T) {
	m := newConfirmModal()
	m.SetText(strings.Repeat("word ", 30))
	m, _ = m.Update(tea.WindowSizeMsg{Width: 24, Height: 10})
	if w, h := m.Size(); w > 24 || h > 10 {
		t.Fatalf("fits the window: %dx%d", w, h)
	}
	m, _ = m.Update(uitest.Key("right"))
	m.SetButtons("A", "B", "C")
	if m.Active() != 0 {
		t.Fatal("SetButtons focuses the first button")
	}
	if len(m.ShortHelp()) != 4 || len(m.FullHelp()) != 1 {
		t.Fatal("help")
	}
}
