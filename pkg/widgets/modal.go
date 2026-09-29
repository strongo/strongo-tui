package widgets

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
)

// modalMaxTextWidth is the widest a modal's wrapped text gets.
const modalMaxTextWidth = 60

// ModalColors overrides the colours of a Modal. A nil colour uses the theme's:
// the focus colour for the border, the text colour for the text, no background,
// and the default look for unfocused buttons.
type ModalColors struct {
	Background       color.Color
	Text             color.Color
	Border           color.Color
	ButtonBackground color.Color
	ButtonText       color.Color
}

// ModalDoneMsg reports that the modal was answered: a button was pressed with
// Enter or a click, or Esc was pressed, in which case Index is -1 and Label is
// empty. The parent removes the modal when it receives it.
type ModalDoneMsg struct {
	ID    string
	Index int
	Label string
}

// ModalKeyMap holds the key bindings of Modal.
type ModalKeyMap struct {
	Prev    key.Binding
	Next    key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

// DefaultModalKeyMap returns the default key bindings.
func DefaultModalKeyMap() ModalKeyMap {
	return ModalKeyMap{
		Prev:    key.NewBinding(key.WithKeys("left", "shift+tab"), key.WithHelp("←", "previous button")),
		Next:    key.NewBinding(key.WithKeys("right", "tab"), key.WithHelp("→", "next button")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
		Cancel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

// ShortHelp implements help.KeyMap.
func (k ModalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Prev, k.Next, k.Confirm, k.Cancel}
}

// FullHelp implements help.KeyMap.
func (k ModalKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

// modalQuit are the keys a modal leaves alone so that the shell can quit.
var modalQuit = key.NewBinding(key.WithKeys("ctrl+c", "ctrl+q"))

// Modal is a dialog: a bordered box with a title in its border, wrapped text and
// a row of buttons drawn with Button. View returns the box only, at its natural
// size; the shell overlays it on the screen with Center, so the modal need not
// know what is behind it. bubbles has no dialog, and Frame draws neither a
// background nor a custom border colour, so the box is drawn here.
//
// A modal owns the input, so it is always focused and has no Focus or Blur:
// Left, Right, Tab and Shift+Tab move between the buttons (wrapping around),
// Enter presses the focused button, Esc answers with index -1, and every other
// key press is swallowed. Only Ctrl+C and Ctrl+Q are left alone, Update
// returning the model and no command, so that the shell can quit. A left click
// on a button presses it and every other click is swallowed. Mouse coordinates
// are relative to the top-left corner of the box, which is where the shell puts
// the box's top-left cell.
//
// Messages: ModalDoneMsg. A tea.WindowSizeMsg sets the space the box must fit in,
// like SetMaxSize. Modal is neither a Boundary nor an Editor.
type Modal struct {
	// KeyMap holds the key bindings.
	KeyMap ModalKeyMap

	id      string
	title   string
	text    string
	buttons []string
	colors  ModalColors
	maxW    int
	maxH    int
	active  int
}

// NewModal creates a modal without text or buttons. id identifies it in
// ModalDoneMsg.
func NewModal(id string) Modal { return Modal{KeyMap: DefaultModalKeyMap(), id: id} }

// SetTitle sets the title drawn in the top border.
func (m *Modal) SetTitle(title string) { m.title = title }

// SetText sets the message; Size and View wrap it.
func (m *Modal) SetText(text string) { m.text = text }

// SetButtons replaces the buttons and focuses the first one.
func (m *Modal) SetButtons(labels ...string) {
	m.buttons = append([]string(nil), labels...)
	m.active = 0
}

// SetColors overrides the theme's colours; nil fields keep the theme's.
func (m *Modal) SetColors(colors ModalColors) { m.colors = colors }

// SetMaxSize sets the space the box must fit in. The text is wrapped at the
// smaller of maxWidth-4 and 60 columns, and text rows that do not fit the height
// are dropped. Zero or less means unlimited, which still wraps at 60 columns.
func (m *Modal) SetMaxSize(maxWidth, maxHeight int) { m.maxW, m.maxH = maxWidth, maxHeight }

// Active returns the index of the focused button.
func (m Modal) Active() int { return m.active }

// Size returns the natural size of the box.
func (m Modal) Size() (w, h int) {
	l := m.layout()
	return l.inner + 4, l.height
}

// ShortHelp implements help.KeyMap.
func (m Modal) ShortHelp() []key.Binding { return m.KeyMap.ShortHelp() }

// FullHelp implements help.KeyMap.
func (m Modal) FullHelp() [][]key.Binding { return m.KeyMap.FullHelp() }

// Update handles key presses, left clicks and size messages. See Modal for
// which keys are swallowed.
func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetMaxSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return m, nil
		}
		l := m.layout()
		if msg.Y != l.buttonsY {
			return m, nil
		}
		for i, r := range l.rects {
			if msg.X >= r[0] && msg.X < r[1] {
				m.active = i
				return m, m.done(i)
			}
		}
	}
	return m, nil
}

func (m Modal) updateKey(msg tea.KeyPressMsg) (Modal, tea.Cmd) {
	n := len(m.buttons)
	switch {
	case key.Matches(msg, modalQuit):
		return m, nil
	case n > 0 && key.Matches(msg, m.KeyMap.Prev):
		m.active = (m.active + n - 1) % n
	case n > 0 && key.Matches(msg, m.KeyMap.Next):
		m.active = (m.active + 1) % n
	case n > 0 && key.Matches(msg, m.KeyMap.Confirm):
		return m, m.done(m.active)
	case key.Matches(msg, m.KeyMap.Cancel):
		return m, Emit(ModalDoneMsg{ID: m.id, Index: -1})
	}
	return m, nil
}

func (m Modal) done(i int) tea.Cmd {
	return Emit(ModalDoneMsg{ID: m.id, Index: i, Label: m.buttons[i]})
}

// modalLayout is where everything sits in the box, computed from the model.
type modalLayout struct {
	lines    []string // wrapped text
	inner    int      // width of the content area
	height   int
	buttonsY int
	rects    [][2]int // column ranges of the buttons, relative to the box
	left     int      // blank columns before the buttons
}

func (m Modal) button(i int) Button {
	return Button{
		Label:      m.buttons[i],
		Focused:    i == m.active,
		Background: m.colors.ButtonBackground,
		Foreground: m.colors.ButtonText,
	}
}

// layout wraps the text, sizes the box and locates the buttons.
func (m Modal) layout() modalLayout {
	var l modalLayout
	limit := modalMaxTextWidth
	if m.maxW > 0 {
		limit = max(min(limit, m.maxW-4), 1)
	}
	if m.text != "" {
		l.lines = strings.Split(ansi.Wrap(m.text, limit, ""), "\n")
	}
	if m.maxH > 0 {
		room := m.maxH - 2
		if len(m.buttons) > 0 {
			room -= 2
		}
		l.lines = l.lines[:max(min(len(l.lines), room), 0)]
	}
	buttonsWidth := 2 * max(len(m.buttons)-1, 0)
	for i := range m.buttons {
		buttonsWidth += m.button(i).Width()
	}
	if m.title != "" {
		l.inner = ansi.StringWidth(m.title) + 2
	}
	for _, line := range l.lines {
		l.inner = max(l.inner, ansi.StringWidth(line))
	}
	l.inner = max(l.inner, buttonsWidth, 1)
	if m.maxW > 0 {
		l.inner = max(min(l.inner, m.maxW-4), 1)
	}

	l.height = 2 + len(l.lines)
	if len(m.buttons) > 0 {
		if len(l.lines) > 0 {
			l.height++
		}
		l.buttonsY = l.height - 1
		l.height++
		l.left = max((l.inner-buttonsWidth)/2, 0)
		x := 2 + l.left
		for i := range m.buttons {
			if i > 0 {
				x += 2
			}
			end := x + m.button(i).Width()
			l.rects = append(l.rects, [2]int{min(x, 2+l.inner), min(end, 2+l.inner)})
			x = end
		}
	}
	return l
}

func modalStyle(fg, bg color.Color) lipgloss.Style {
	s := lipgloss.NewStyle()
	if fg != nil {
		s = s.Foreground(fg)
	}
	if bg != nil {
		s = s.Background(bg)
	}
	return s
}

// View renders the box at its natural size: every line has the same width.
func (m Modal) View() string {
	l := m.layout()
	borderColor, textColor := m.colors.Border, m.colors.Text
	if borderColor == nil {
		borderColor = theme.FocusColor()
	}
	if textColor == nil {
		textColor = theme.TextColor()
	}
	bg := m.colors.Background
	borderSty, textSty, bgSty := modalStyle(borderColor, bg), modalStyle(textColor, bg), modalStyle(nil, bg)

	wall := borderSty.Render("│") + bgSty.Render(" ")
	wallR := bgSty.Render(" ") + borderSty.Render("│")
	row := func(content string) string { return wall + content + wallR }

	out := []string{m.topBorder(borderSty, l.inner)}
	for _, line := range l.lines {
		out = append(out, row(textSty.Render(AlignIn(line, l.inner, AlignCenter))))
	}
	if len(m.buttons) > 0 {
		if len(l.lines) > 0 {
			out = append(out, row(bgSty.Render(strings.Repeat(" ", l.inner))))
		}
		out = append(out, row(m.buttonRow(bgSty, l)))
	}
	out = append(out, borderSty.Render("╰"+strings.Repeat("─", l.inner+2)+"╯"))
	return strings.Join(out, "\n")
}

// topBorder draws the top edge with the title in it.
func (m Modal) topBorder(sty lipgloss.Style, inner int) string {
	if m.title == "" {
		return sty.Render("╭" + strings.Repeat("─", inner+2) + "╮")
	}
	segment := ansi.Truncate(" "+ansi.Strip(m.title)+" ", inner+1, "")
	return sty.Render("╭─" + segment + strings.Repeat("─", inner+1-ansi.StringWidth(segment)) + "╮")
}

// buttonRow draws the buttons centred in the content area.
func (m Modal) buttonRow(bgSty lipgloss.Style, l modalLayout) string {
	var sb strings.Builder
	sb.WriteString(bgSty.Render(strings.Repeat(" ", l.left)))
	for i := range m.buttons {
		if i > 0 {
			sb.WriteString(bgSty.Render("  "))
		}
		b := m.button(i)
		sb.WriteString(b.View(b.Width()))
	}
	row := ansi.Truncate(sb.String(), l.inner, "")
	return row + bgSty.Render(strings.Repeat(" ", max(l.inner-ansi.StringWidth(row), 0)))
}
