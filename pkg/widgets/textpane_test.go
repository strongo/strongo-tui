package widgets_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tuigoff/tuigoff/pkg/theme"
	"github.com/tuigoff/tuigoff/pkg/uitest"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

func numberedText(n int) string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	return strings.Join(lines, "\n")
}

func newPane(text string, w, h int) widgets.TextPane {
	p := widgets.NewTextPane("pane")
	p.SetSize(w, h)
	p.SetContent(text)
	p.Focus()
	return p
}

func TestTextPaneViewIsExactlySized(t *testing.T) {
	p := newPane("one\ntwo", 8, 4)
	lines := strings.Split(p.View(), "\n")
	if len(lines) != 4 {
		t.Fatalf("%d rows", len(lines))
	}
	for i, l := range lines {
		if ansi.StringWidth(l) != 8 {
			t.Fatalf("row %d is %d wide", i, ansi.StringWidth(l))
		}
	}
	if got := uitest.Plain(p.View()); got != "one\ntwo\n\n" {
		t.Fatalf("a short text leaves blank rows: %q", got)
	}
	p.SetSize(0, 3)
	if p.View() != "" {
		t.Fatal("zero width renders nothing")
	}
	p.SetSize(-1, -1)
	if p.View() != "" || p.ID() != "pane" {
		t.Fatal("negative size renders nothing")
	}
}

func TestTextPaneContentAndLineCount(t *testing.T) {
	p := newPane("a\r\nb\n", 10, 3)
	if p.Content() != "a\r\nb\n" || p.LineCount() != 2 {
		t.Fatalf("content %q, %d lines", p.Content(), p.LineCount())
	}
	p.AppendContent("c")
	if p.LineCount() != 3 || uitest.Plain(p.View()) != "a\nb\nc" {
		t.Fatalf("append: %q", uitest.Plain(p.View()))
	}
	p.SetContent("")
	if p.LineCount() != 0 || p.Content() != "" {
		t.Fatal("empty")
	}
}

func TestTextPaneScrollingKeys(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		want int
	}{
		{"down", []string{"down"}, 1},
		{"up at top", []string{"up"}, 0},
		{"page", []string{"pgdown"}, 3},
		{"page up", []string{"pgdown", "pgup"}, 0},
		{"end", []string{"end"}, 7},
		{"end then home", []string{"end", "home"}, 0},
		{"down past the end", []string{"end", "down"}, 7},
		{"up", []string{"end", "up"}, 6},
		{"letters are not bound", []string{"j", "k", "space", "f", "b", "g", "G"}, 0},
	}
	for _, tt := range tests {
		p := newPane(numberedText(10), 12, 3)
		for _, k := range tt.keys {
			var cmd tea.Cmd
			p, cmd = p.Update(uitest.Key(k))
			if cmd != nil {
				t.Fatalf("%s: TextPane sends no messages", tt.name)
			}
		}
		if p.YOffset() != tt.want {
			t.Errorf("%s: y offset %d, want %d", tt.name, p.YOffset(), tt.want)
		}
	}
}

func TestTextPaneUnfocusedIgnoresKeysButNotTheWheel(t *testing.T) {
	p := newPane(numberedText(20), 12, 3)
	p.Blur()
	p, _ = p.Update(uitest.Key("down"))
	if p.Focused() || p.YOffset() != 0 {
		t.Fatal("unfocused pane ignores keys")
	}
	p, _ = p.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if p.YOffset() != 3 {
		t.Fatalf("the wheel scrolls three lines: %d", p.YOffset())
	}
	p, _ = p.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if p.YOffset() != 0 {
		t.Fatal("wheel up")
	}
}

func TestTextPaneBoundary(t *testing.T) {
	p := newPane(numberedText(10), 12, 3)
	if !p.AtEdge(widgets.Up) || p.AtEdge(widgets.Down) || !p.AtEdge(widgets.Left) || !p.AtEdge(widgets.Right) {
		t.Fatal("top while wrapping")
	}
	p, _ = p.Update(uitest.Key("end"))
	if p.AtEdge(widgets.Up) || !p.AtEdge(widgets.Down) {
		t.Fatal("bottom")
	}
	var _ widgets.Boundary = p
	short := newPane("x", 12, 3)
	if !short.AtEdge(widgets.Up) || !short.AtEdge(widgets.Down) {
		t.Fatal("text that fits is at both edges")
	}
}

func TestTextPaneHorizontalScrolling(t *testing.T) {
	p := newPane("0123456789abcdefghij", 8, 2)
	p, _ = p.Update(uitest.Key("right"))
	if p.XOffset() != 0 {
		t.Fatal("no horizontal scroll while wrapping")
	}
	p.SetWrap(false)
	if got := uitest.Plain(p.View()); got != "01234567\n" {
		t.Fatalf("unwrapped: %q", got)
	}
	if !p.AtEdge(widgets.Left) || p.AtEdge(widgets.Right) {
		t.Fatal("left end of a long line")
	}
	p, _ = p.Update(uitest.Key("right"))
	if p.XOffset() != 6 || p.AtEdge(widgets.Left) {
		t.Fatalf("right scrolls: %d", p.XOffset())
	}
	p, _ = p.Update(uitest.Key("right"))
	p, _ = p.Update(uitest.Key("right"))
	if p.XOffset() != 12 || !p.AtEdge(widgets.Right) {
		t.Fatalf("clamped at the end: %d", p.XOffset())
	}
	if got := uitest.Plain(p.View()); got != "cdefghij\n" {
		t.Fatalf("scrolled: %q", got)
	}
	p, _ = p.Update(uitest.Key("left"))
	if p.XOffset() != 6 {
		t.Fatal("left")
	}
	p.SetWrap(true)
	if p.XOffset() != 0 || !p.AtEdge(widgets.Right) {
		t.Fatal("wrapping again resets the column")
	}
}

func TestTextPaneWrapsLongLines(t *testing.T) {
	p := newPane("0123456789abcdef", 6, 4)
	if got := uitest.Plain(p.View()); got != "012345\n6789ab\ncdef\n" {
		t.Fatalf("wrapped: %q", got)
	}
	if p.LineCount() != 3 {
		t.Fatal("wrapped lines count separately")
	}
}

func TestTextPaneANSISurvivesWrapping(t *testing.T) {
	red := theme.RedText("0123456789abcdef")
	p := newPane(red, 6, 3)
	rows := strings.Split(p.View(), "\n")
	for i, row := range rows[:3] {
		if !strings.Contains(row, "\x1b[") {
			t.Fatalf("row %d lost its style: %q", i, row)
		}
	}
	if uitest.Plain(p.View()) != "012345\n6789ab\ncdef" {
		t.Fatalf("plain: %q", uitest.Plain(p.View()))
	}
	unwrapped := newPane(red, 6, 1)
	unwrapped.SetWrap(false)
	unwrapped.ScrollTo(0, 4)
	if !strings.Contains(unwrapped.View(), "\x1b[") || uitest.Plain(unwrapped.View()) != "456789" {
		t.Fatalf("cropped row keeps its style: %q", unwrapped.View())
	}
}

func TestTextPaneStyleOpenAcrossLinesIsCarried(t *testing.T) {
	p := newPane("\x1b[31mred\nstill red\x1b[m plain\nnormal", 20, 3)
	rows := strings.Split(p.View(), "\n")
	if !strings.Contains(rows[1], "\x1b[31m") {
		t.Fatalf("second line re-opens the style: %q", rows[1])
	}
	if strings.Contains(rows[2], "\x1b[31m") {
		t.Fatalf("the style ended: %q", rows[2])
	}
}

func TestTextPaneDefaultTextColor(t *testing.T) {
	p := newPane("plain \x1b[1mbold\x1b[m after\nsecond", 20, 2)
	plain := p.View()
	p.SetTextColor(theme.LightBlue)
	coloured := p.View()
	if coloured == plain || !strings.Contains(coloured, "\x1b[") {
		t.Fatal("the default colour is applied")
	}
	if uitest.Plain(coloured) != uitest.Plain(plain) {
		t.Fatal("colour does not change the text")
	}
	prefix, _, _ := strings.Cut(lipgloss.NewStyle().Foreground(theme.LightBlue).Render("x"), "x")
	if strings.Count(strings.Split(coloured, "\n")[0], prefix) != 2 {
		t.Fatalf("the colour is re-applied after the reset: %q", coloured)
	}
}

func TestTextPaneResizeKeepsScrollPosition(t *testing.T) {
	p := newPane(numberedText(30), 10, 5)
	p.ScrollTo(12, 0)
	p.SetSize(20, 5)
	p.SetSize(20, 8)
	if p.YOffset() != 12 {
		t.Fatalf("resizing keeps the position: %d", p.YOffset())
	}
	p, _ = p.Update(tea.WindowSizeMsg{Width: 10, Height: 4})
	if p.YOffset() != 12 || len(strings.Split(p.View(), "\n")) != 4 {
		t.Fatal("window size message")
	}
	p.SetSize(10, 40)
	if p.YOffset() != 0 {
		t.Fatalf("a taller pane clamps the position: %d", p.YOffset())
	}
}

func TestTextPaneScrollToAndGoto(t *testing.T) {
	p := newPane(numberedText(20), 10, 4)
	p.GotoBottom()
	if p.YOffset() != 16 {
		t.Fatalf("bottom: %d", p.YOffset())
	}
	p.GotoTop()
	p.ScrollTo(99, 99)
	if p.YOffset() != 16 || p.XOffset() != 0 {
		t.Fatal("clamped")
	}
	p.SetContent("new")
	if p.YOffset() != 0 {
		t.Fatal("SetContent scrolls to the top")
	}
	p.SetContent(numberedText(20))
	p.GotoBottom()
	p.AppendContent("\nmore")
	if p.YOffset() != 16 {
		t.Fatal("AppendContent keeps the position")
	}
}

func TestTextPaneHelp(t *testing.T) {
	p := widgets.NewTextPane("x")
	if len(p.ShortHelp()) != 4 || len(p.FullHelp()) != 2 {
		t.Fatal("help")
	}
}
