package widgets

import (
	"image/color"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TextPaneKeyMap holds the key bindings of TextPane. Letters are not bound.
type TextPaneKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Left     key.Binding
	Right    key.Binding
}

// DefaultTextPaneKeyMap returns the default key bindings.
func DefaultTextPaneKeyMap() TextPaneKeyMap {
	return TextPaneKeyMap{
		Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		Top:      key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "top")),
		Bottom:   key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "bottom")),
		Left:     key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "scroll left")),
		Right:    key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "scroll right")),
	}
}

// ShortHelp implements help.KeyMap.
func (k TextPaneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.PageUp, k.PageDown}
}

// FullHelp implements help.KeyMap.
func (k TextPaneKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.PageUp, k.PageDown}, {k.Top, k.Bottom, k.Left, k.Right}}
}

// TextPane shows scrollable, optionally wrapped text and is built on bubbles'
// viewport, which does the scrolling, the soft wrapping and the horizontal
// cropping. The wrapper adds the key map (arrows, page keys, home and
// end; letters are not bound), a default text colour, the Boundary queries, and
// the exact-size View.
//
// The content may carry ANSI styling (lipgloss output, theme.RedText, highlight
// output). Styling survives soft wrapping because the viewport cuts every wrapped
// row out of the styled line with ansi.Cut, which re-opens the styles that are
// active at the start of the row and closes them at its end (the tests verify
// this). A style that stays open across a line break, which the viewport would
// otherwise lose on the next line, is closed at the end of the line and re-opened
// at the start of the next by carryStyles when the content is set.
//
// Wrapping is on by default and breaks lines at exactly the width (the viewport
// does not break at word boundaries). With SetWrap(false), Left and Right scroll
// horizontally. The wheel scrolls three lines. The vertical position is kept when
// the pane is resized and clamped to the text; SetContent scrolls to the top and
// AppendContent keeps the position.
//
// TextPane sends no messages. It is a Boundary: Up is at the edge when scrolled
// to the top, Down at the bottom, Left when scrolled fully left (always while
// wrapping) and Right when scrolled fully right (always while wrapping). It is
// not an Editor.
type TextPane struct {
	// KeyMap holds the key bindings.
	KeyMap TextPaneKeyMap

	id        string
	vp        viewport.Model
	content   string
	textColor color.Color
	focused   bool
}

// NewTextPane creates an empty pane that wraps long lines.
func NewTextPane(id string) TextPane {
	vp := viewport.New()
	vp.SoftWrap = true
	vp.MouseWheelDelta = 3
	return TextPane{KeyMap: DefaultTextPaneKeyMap(), id: id, vp: vp}
}

// ID returns the id the pane was created with.
func (t TextPane) ID() string { return t.id }

// SetContent replaces the text and scrolls to the top.
func (t *TextPane) SetContent(s string) {
	t.content = s
	t.refresh()
	t.vp.SetYOffset(0)
	t.vp.SetXOffset(0)
}

// AppendContent adds s to the end of the text without moving the scroll
// position; call GotoBottom afterwards to follow a stream.
func (t *TextPane) AppendContent(s string) {
	t.content += s
	t.refresh()
}

// Content returns the text as it was given.
func (t TextPane) Content() string { return t.content }

// SetTextColor sets the foreground colour of text that has no colour of its own.
func (t *TextPane) SetTextColor(c color.Color) {
	t.textColor = c
	t.refresh()
}

// SetWrap turns soft wrapping on (the default) or off, which enables horizontal
// scrolling.
func (t *TextPane) SetWrap(wrap bool) {
	if wrap {
		t.vp.SetXOffset(0)
	}
	t.vp.SoftWrap = wrap
	t.clamp()
}

// SetSize sets the size of the pane; the scroll position is kept and clamped.
func (t *TextPane) SetSize(width, height int) {
	t.vp.SetWidth(max(width, 0))
	t.vp.SetHeight(max(height, 0))
	t.clamp()
}

// clamp brings the scroll offsets back inside the text after a size change.
func (t *TextPane) clamp() {
	t.vp.SetYOffset(t.vp.YOffset())
	t.vp.SetXOffset(t.vp.XOffset())
}

// YOffset returns the first visible line, counting wrapped lines.
func (t TextPane) YOffset() int { return t.vp.YOffset() }

// XOffset returns the first visible column; it is 0 while wrapping.
func (t TextPane) XOffset() int { return t.vp.XOffset() }

// ScrollTo scrolls so that line y and column x are the first visible ones,
// clamped to the text.
func (t *TextPane) ScrollTo(y, x int) {
	t.vp.SetYOffset(y)
	t.vp.SetXOffset(x)
}

// GotoTop scrolls to the first line.
func (t *TextPane) GotoTop() { t.vp.SetYOffset(0) }

// GotoBottom scrolls so that the last line is visible.
func (t *TextPane) GotoBottom() { t.vp.GotoBottom() }

// LineCount returns the number of lines at the current width, wrapped lines
// counted separately.
func (t TextPane) LineCount() int { return t.vp.TotalLineCount() }

// Focus makes the pane handle keys.
func (t *TextPane) Focus() { t.focused = true }

// Blur stops the pane handling keys.
func (t *TextPane) Blur() { t.focused = false }

// Focused reports whether the pane holds focus.
func (t TextPane) Focused() bool { return t.focused }

// AtEdge implements Boundary.
func (t TextPane) AtEdge(dir Direction) bool {
	switch dir {
	case Up:
		return t.vp.AtTop()
	case Down:
		return t.vp.AtBottom()
	case Left:
		return t.vp.SoftWrap || t.vp.XOffset() == 0
	default:
		return t.vp.SoftWrap || t.vp.HorizontalScrollPercent() >= 1
	}
}

// ShortHelp implements help.KeyMap.
func (t TextPane) ShortHelp() []key.Binding { return t.KeyMap.ShortHelp() }

// FullHelp implements help.KeyMap.
func (t TextPane) FullHelp() [][]key.Binding { return t.KeyMap.FullHelp() }

// Update handles key presses while focused, the mouse wheel, and size messages.
func (t TextPane) Update(msg tea.Msg) (TextPane, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		if t.focused {
			t.handleKey(msg)
		}
	case tea.MouseWheelMsg:
		t.vp, _ = t.vp.Update(msg)
	}
	return t, nil
}

func (t *TextPane) handleKey(msg tea.KeyPressMsg) {
	switch {
	case key.Matches(msg, t.KeyMap.Up):
		t.vp.ScrollUp(1)
	case key.Matches(msg, t.KeyMap.Down):
		t.vp.ScrollDown(1)
	case key.Matches(msg, t.KeyMap.PageUp):
		t.vp.PageUp()
	case key.Matches(msg, t.KeyMap.PageDown):
		t.vp.PageDown()
	case key.Matches(msg, t.KeyMap.Top):
		t.GotoTop()
	case key.Matches(msg, t.KeyMap.Bottom):
		t.GotoBottom()
	case key.Matches(msg, t.KeyMap.Left):
		t.vp.ScrollLeft(textPaneStep)
	case key.Matches(msg, t.KeyMap.Right):
		t.vp.ScrollRight(textPaneStep)
	}
}

// textPaneStep is how many columns Left and Right scroll.
const textPaneStep = 6

// View renders the visible text in exactly the size given to SetSize.
func (t TextPane) View() string {
	w, h := t.vp.Width(), t.vp.Height()
	if w <= 0 || h <= 0 {
		return ""
	}
	return Fit(t.vp.View(), w, h)
}

// refresh prepares the content for the viewport: carriage returns and a final
// newline dropped, styles carried across line breaks, and the default text colour
// applied.
func (t *TextPane) refresh() {
	s := strings.ReplaceAll(t.content, "\r", "")
	s = strings.TrimSuffix(s, "\n")
	lines := strings.Split(s, "\n")
	carryStyles(lines)
	if prefix := t.colorPrefix(); prefix != "" {
		reapply := strings.NewReplacer("\x1b[m", "\x1b[m"+prefix, "\x1b[0m", "\x1b[0m"+prefix)
		for i, line := range lines {
			lines[i] = prefix + reapply.Replace(line) + "\x1b[m"
		}
	}
	t.vp.SetContentLines(lines)
	t.clamp()
}

// colorPrefix is the escape sequence that selects the default text colour, or "".
func (t TextPane) colorPrefix() string {
	if t.textColor == nil {
		return ""
	}
	sample := lipgloss.NewStyle().Foreground(t.textColor).Render("x")
	prefix, _, _ := strings.Cut(sample, "x")
	return prefix
}

// sgrSequence matches a Select Graphic Rendition escape sequence.
var sgrSequence = regexp.MustCompile("\x1b\\[([0-9;:]*)m")

// carryStyles makes every line self-contained: a style that is still open at
// the end of a line is closed there and re-opened at the start of the next one,
// so each line can be cut, aligned or drawn on its own.
func carryStyles(lines []string) {
	var open []string
	for i, line := range lines {
		prefix := strings.Join(open, "")
		for _, m := range sgrSequence.FindAllStringSubmatch(line, -1) {
			if m[1] == "" || m[1] == "0" {
				open = open[:0]
			} else {
				open = append(open, m[0])
			}
		}
		if len(open) > 0 {
			line += "\x1b[m"
		}
		lines[i] = prefix + line
	}
}
