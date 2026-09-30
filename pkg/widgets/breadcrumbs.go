package widgets

import (
	"image/color"
	"net/url"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
)

// Crumb is one step of a navigation trail.
type Crumb struct {
	// Title is the text of the step.
	Title string
	// Color colours the step; the theme's hotkey colour is used when nil.
	Color color.Color
}

// CrumbSelectedMsg reports that a crumb was activated with the keyboard or the
// mouse. Index is its position in the trail, 0 being the root.
type CrumbSelectedMsg struct {
	ID    string
	Index int
	Crumb Crumb
}

// BreadcrumbsKeyMap holds the key bindings of Breadcrumbs.
type BreadcrumbsKeyMap struct {
	Prev     key.Binding
	Next     key.Binding
	Activate key.Binding
}

// DefaultBreadcrumbsKeyMap returns the default key bindings.
func DefaultBreadcrumbsKeyMap() BreadcrumbsKeyMap {
	return BreadcrumbsKeyMap{
		Prev:     key.NewBinding(key.WithKeys("left", "<"), key.WithHelp("←", "previous")),
		Next:     key.NewBinding(key.WithKeys("right", ">"), key.WithHelp("→", "next")),
		Activate: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "go")),
	}
}

// ShortHelp implements help.KeyMap.
func (k BreadcrumbsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Prev, k.Next, k.Activate}
}

// FullHelp implements help.KeyMap.
func (k BreadcrumbsKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

// Breadcrumbs is a one-line navigation trail. Left and Right move the selection
// along the trail and Enter activates the selected crumb, which is reported with
// a CrumbSelectedMsg.
//
// The selection rests on the last crumb. When the trail receives focus the
// selection jumps to the crumb before it, the parent step, so that Enter goes one
// step back. The trail is a Boundary: Left is at the edge on the first crumb and
// Right on the last.
type Breadcrumbs struct {
	// KeyMap holds the key bindings.
	KeyMap BreadcrumbsKeyMap

	id        string
	crumbs    []Crumb
	selected  int
	focused   bool
	width     int
	separator string
	sepStart  int
}

// NewBreadcrumbs creates a trail. id identifies it in CrumbSelectedMsg.
func NewBreadcrumbs(id string, crumbs ...Crumb) Breadcrumbs {
	b := Breadcrumbs{KeyMap: DefaultBreadcrumbsKeyMap(), id: id, separator: " > "}
	b.SetCrumbs(crumbs)
	return b
}

// SetCrumbs replaces the trail and rests the selection on its last crumb.
func (b *Breadcrumbs) SetCrumbs(crumbs []Crumb) {
	b.crumbs = append([]Crumb(nil), crumbs...)
	b.selected = max(len(b.crumbs)-1, 0)
	if b.focused && len(b.crumbs) > 1 {
		b.selected = len(b.crumbs) - 2
	}
}

// SetSeparator sets the text drawn between crumbs; the default is " > ".
func (b *Breadcrumbs) SetSeparator(separator string) { b.separator = separator }

// SetSeparatorStart suppresses the separators before crumb i, for trails whose
// first crumbs form one visual unit.
func (b *Breadcrumbs) SetSeparatorStart(i int) { b.sepStart = i }

// Crumbs returns a copy of the trail.
func (b Breadcrumbs) Crumbs() []Crumb { return append([]Crumb(nil), b.crumbs...) }

// Selected returns the index of the selected crumb.
func (b Breadcrumbs) Selected() int { return b.selected }

// SetSize gives the trail its width; the height is always one row.
func (b *Breadcrumbs) SetSize(width, _ int) { b.width = width }

// Focus moves the selection to the parent crumb.
func (b *Breadcrumbs) Focus() {
	if !b.focused && len(b.crumbs) > 1 {
		b.selected = len(b.crumbs) - 2
	}
	b.focused = true
}

// Blur returns the selection to the last crumb.
func (b *Breadcrumbs) Blur() {
	b.focused = false
	b.selected = max(len(b.crumbs)-1, 0)
}

// Focused reports whether the trail holds focus.
func (b Breadcrumbs) Focused() bool { return b.focused }

// AtEdge implements Boundary.
func (b Breadcrumbs) AtEdge(dir Direction) bool {
	switch dir {
	case Left:
		return b.selected == 0
	case Right:
		return b.selected == len(b.crumbs)-1
	default:
		return true
	}
}

// Update handles key presses while focused, clicks, and size messages.
func (b Breadcrumbs) Update(msg tea.Msg) (Breadcrumbs, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
	case tea.KeyPressMsg:
		if !b.focused || len(b.crumbs) == 0 {
			return b, nil
		}
		switch {
		case key.Matches(msg, b.KeyMap.Prev):
			b.selected = max(b.selected-1, 0)
		case key.Matches(msg, b.KeyMap.Next):
			b.selected = min(b.selected+1, len(b.crumbs)-1)
		case key.Matches(msg, b.KeyMap.Activate):
			return b, b.activate(b.selected)
		}
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return b, nil
		}
		for i, span := range b.spans() {
			if msg.X >= span[0] && msg.X < span[1] {
				b.selected = i
				return b, b.activate(i)
			}
		}
	}
	return b, nil
}

func (b Breadcrumbs) activate(i int) tea.Cmd {
	return Emit(CrumbSelectedMsg{ID: b.id, Index: i, Crumb: b.crumbs[i]})
}

// label is the text drawn for crumb i.
func (b Breadcrumbs) label(i int) string {
	title := strings.TrimSuffix(b.crumbs[i].Title, b.separator)
	if u, err := url.Parse(title); err == nil && u.Scheme != "" {
		title += " "
	}
	return title
}

// spans returns the column range of every crumb at the current width.
func (b Breadcrumbs) spans() [][2]int {
	spans := make([][2]int, len(b.crumbs))
	col := 0
	for i := range b.crumbs {
		w := min(ansi.StringWidth(b.label(i)), max(b.width-col, 0))
		spans[i] = [2]int{col, col + w}
		col += w
		if i >= b.sepStart && i < len(b.crumbs)-1 {
			col += min(ansi.StringWidth(b.separator), max(b.width-col, 0))
		}
	}
	return spans
}

// View renders the trail on one row of the given width.
func (b Breadcrumbs) View() string {
	sepStyle := lipgloss.NewStyle().Foreground(theme.MutedColor())
	var sb strings.Builder
	col := 0
	for i, crumb := range b.crumbs {
		if col >= b.width {
			break
		}
		fg := crumb.Color
		if fg == nil {
			fg = theme.AccentColor()
		}
		style := lipgloss.NewStyle().Foreground(fg).Faint(true)
		if i == b.selected {
			style = style.Faint(false)
			if b.focused {
				style = theme.SelectedStyle(true)
			}
		}
		text := ansi.Truncate(b.label(i), b.width-col, "")
		sb.WriteString(style.Render(text))
		col += ansi.StringWidth(text)
		if i >= b.sepStart && i < len(b.crumbs)-1 && col < b.width {
			sep := ansi.Truncate(b.separator, b.width-col, "")
			sb.WriteString(sepStyle.Render(sep))
			col += ansi.StringWidth(sep)
		}
	}
	return Fit(sb.String(), b.width, 1)
}
