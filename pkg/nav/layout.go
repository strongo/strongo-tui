package nav

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tuigoff/tuigoff/pkg/theme"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

// rect is a rectangle of terminal cells.
type rect struct{ x, y, w, h int }

func (r rect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

// geometry is where each part of the shell sits at the current size.
type geometry struct {
	crumbs  rect
	login   rect
	menu    rect
	content rect
	actions rect
}

func (m Model) geometry() geometry {
	w, h := m.width, m.height
	left, right := theme.HintsInset()
	inner := max(w-left-right, 0)
	loginW := 0
	if !m.noLogin {
		loginW = min(m.login.Width(), inner)
	}
	bodyH := max(h-2, 0)
	menuW := 0
	if m.menuVisible() {
		menuW = min(MenuWidth, w)
	}
	return geometry{
		crumbs:  rect{left, 0, inner - loginW, 1},
		login:   rect{left + inner - loginW, 0, loginW, 1},
		menu:    rect{0, 1, menuW, bodyH},
		content: rect{menuW, 1, w - menuW, bodyH},
		actions: rect{0, max(h-1, 0), w, 1},
	}
}

// frameFor is the frame drawn around a screen.
func frameFor(s Screen, focused bool) widgets.Frame {
	f := widgets.NewFrame().WithFocus(focused)
	if t, ok := s.(Titled); ok {
		f = f.WithTitle(t.Title())
	}
	if b, ok := s.(Borderless); ok && b.Borderless() {
		f = f.WithoutBorders()
	}
	return f
}

// menuSize and contentSize are the sizes screens are told: the area inside the
// frame.
func (m Model) menuSize() (int, int) {
	g := m.geometry()
	return frameFor(m.menu(), false).Inner(g.menu.w, g.menu.h)
}

func (m Model) contentSize() (int, int) {
	g := m.geometry()
	return frameFor(m.content(), false).Inner(g.content.w, g.content.h)
}

// actionBar renders the actions bar text and returns the column range of every
// visible action, for clicks.
func (m Model) actionBar() (text string, spans [][2]int, actions []Action) {
	left, _ := theme.StatusInset()
	keyStyle := lipgloss.NewStyle().Foreground(theme.AccentColor())
	descStyle := lipgloss.NewStyle().Foreground(theme.MutedColor())
	col := left
	var parts []string
	if s := m.focusedScreen(); s != nil {
		for _, b := range shortHelp(s) {
			if !b.Enabled() {
				continue
			}
			h := b.Help()
			parts = append(parts, keyStyle.Render(h.Key)+" "+descStyle.Render(h.Desc))
			col += ansi.StringWidth(h.Key+" "+h.Desc) + 2
		}
	}
	for _, a := range m.actions {
		if a.Hidden || !a.Binding.Enabled() {
			continue
		}
		h := a.Binding.Help()
		w := ansi.StringWidth(h.Key + " " + h.Desc)
		parts = append(parts, keyStyle.Render(h.Key)+" "+descStyle.Render(h.Desc))
		spans = append(spans, [2]int{col, col + w})
		actions = append(actions, a)
		col += w + 2
	}
	return strings.Join(parts, "  "), spans, actions
}

// View implements tea.Model.
func (m Model) View() tea.View {
	v := tea.NewView(m.Render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// Render returns the whole screen as an ANSI string, the frame View displays.
func (m Model) Render() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	g := m.geometry()
	head := m.crumbs.View()
	if !m.noLogin {
		head += m.login.View(g.login.w)
	}
	lines := []string{theme.Bar(m.width, head)}
	if g.menu.h > 0 {
		body := m.panel(m.content(), g.content, m.zone == FocusToContent)
		if g.menu.w > 0 {
			body = joinColumns(m.panel(m.menu(), g.menu, m.zone == FocusToMenu), body)
		}
		lines = append(lines, body)
	}
	if m.height > 1 {
		bar, _, _ := m.actionBar()
		lines = append(lines, theme.StatusLine(m.width, bar))
	}
	screen := strings.Join(lines, "\n")
	if m.alert != nil {
		screen = widgets.Center(screen, m.alert.box(m.width, m.height), m.width, m.height)
	}
	return screen
}

// panel draws a screen in its frame.
func (m Model) panel(s Screen, r rect, focused bool) string {
	f := frameFor(s, focused)
	w, h := f.Inner(r.w, r.h)
	body := ""
	if w > 0 && h > 0 {
		body = s.View()
	}
	return f.Render(body, r.w, r.h)
}

// joinColumns places two equally tall blocks side by side.
func joinColumns(left, right string) string {
	l, r := strings.Split(left, "\n"), strings.Split(right, "\n")
	for i := range l {
		l[i] += r[i]
	}
	return strings.Join(l, "\n")
}
