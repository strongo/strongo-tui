package widgets

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tuigoff/tuigoff/pkg/theme"
)

// Tab is one entry of the tab strip.
type Tab struct {
	// ID identifies the tab for the parent.
	ID string
	// Title is the text shown in the strip.
	Title string
	// Closable adds a close mark to the tab.
	Closable bool
}

// TabsStyle selects how the tab strip is drawn; the theme supplies colours.
type TabsStyle struct {
	// Radio prefixes each tab with a filled circle when active, an empty one
	// otherwise.
	Radio bool
	// Underscore underlines the inactive tabs.
	Underscore bool
}

// Predefined tab strip styles.
var (
	// UnderlineTabsStyle underlines the inactive tabs.
	UnderlineTabsStyle = TabsStyle{Underscore: true}
	// RadioTabsStyle marks the active tab with a filled circle.
	RadioTabsStyle = TabsStyle{Radio: true}
)

// TabChangedMsg reports that the active tab changed because of a key press or a
// click. Index is the position of the new active tab.
type TabChangedMsg struct {
	ID    string
	Index int
	Tab   Tab
}

// TabCloseMsg reports a click on the close mark of a closable tab. The parent
// decides whether to call RemoveTab.
type TabCloseMsg struct {
	ID    string
	Index int
	Tab   Tab
}

// TabsKeyMap holds the key bindings of Tabs.
type TabsKeyMap struct {
	Prev key.Binding
	Next key.Binding
	// Jump activates tab n with alt+n, for n from 1 to 9.
	Jump key.Binding
}

// DefaultTabsKeyMap returns the default key bindings.
func DefaultTabsKeyMap() TabsKeyMap {
	return TabsKeyMap{
		Prev: key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "previous tab")),
		Next: key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "next tab")),
		Jump: key.NewBinding(
			key.WithKeys("alt+1", "alt+2", "alt+3", "alt+4", "alt+5", "alt+6", "alt+7", "alt+8", "alt+9"),
			key.WithHelp("alt+1..9", "go to tab")),
	}
}

// ShortHelp implements help.KeyMap.
func (k TabsKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Prev, k.Next, k.Jump} }

// FullHelp implements help.KeyMap.
func (k TabsKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

// tabsCloseMark is the close mark of a closable tab.
const tabsCloseMark = "×"

// Tabs is a one-row strip of tab titles, optionally preceded by a label. It draws
// the strip only: the parent keeps the pages and shows the content of the active
// tab itself, following TabChangedMsg. bubbles has no tab strip, so the rendering
// is our own.
//
// Left and Right switch tabs (no wrapping) and Alt+1 to Alt+9 jump to a tab,
// while focused. A click on a tab activates it, and a click on the close mark of
// a closable tab reports it. The active tab is drawn in the theme's selected
// style when the strip is focused and bold when it is not; inactive tabs are
// muted, and underlined in the Underscore style. When the strip is wider than the
// space it is scrolled so that the active tab is visible.
//
// Messages: TabChangedMsg when the active tab changes through Update (never for
// SetActive, AddTab or RemoveTab, which are setters), and TabCloseMsg on a click
// on a close mark. Mouse clicks work regardless of focus, with the x coordinate
// relative to the left edge of the strip and y 0.
//
// Tabs is a Boundary: Left is at the edge when the first tab is active, Right
// when the last one is, and Up and Down always. It is not an Editor.
type Tabs struct {
	// KeyMap holds the key bindings.
	KeyMap TabsKeyMap

	id      string
	style   TabsStyle
	label   string
	tabs    []Tab
	active  int
	focused bool
	width   int
}

// NewTabs creates a strip. id identifies it in its messages. The first tab is
// active.
func NewTabs(id string, style TabsStyle, tabs ...Tab) Tabs {
	t := Tabs{KeyMap: DefaultTabsKeyMap(), id: id, style: style, active: -1}
	for _, tab := range tabs {
		t.AddTab(tab)
	}
	return t
}

// SetLabel sets the text drawn before the strip; it may carry ANSI styling.
func (t *Tabs) SetLabel(label string) { t.label = label }

// AddTab appends a tab. The first tab added becomes active.
func (t *Tabs) AddTab(tab Tab) {
	t.tabs = append(t.tabs, tab)
	if t.active < 0 {
		t.active = 0
	}
}

// RemoveTab removes the tab at index, ignoring an index outside the strip.
// Removing the active tab activates its neighbour. It emits no message.
func (t *Tabs) RemoveTab(index int) {
	if index < 0 || index >= len(t.tabs) {
		return
	}
	t.tabs = append(t.tabs[:index:index], t.tabs[index+1:]...)
	switch {
	case len(t.tabs) == 0:
		t.active = -1
	case index < t.active:
		t.active--
	case t.active >= len(t.tabs):
		t.active = len(t.tabs) - 1
	}
}

// Tabs returns a copy of the tabs.
func (t Tabs) Tabs() []Tab { return append([]Tab(nil), t.tabs...) }

// Active returns the index of the active tab, or -1 without tabs.
func (t Tabs) Active() int { return t.active }

// SetActive activates the tab at index, ignoring an index outside the strip. It
// emits no message.
func (t *Tabs) SetActive(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.active = index
	}
}

// SetSize gives the strip its width; the height is always one row.
func (t *Tabs) SetSize(width, _ int) { t.width = max(width, 0) }

// Focus makes the strip handle keys.
func (t *Tabs) Focus() { t.focused = true }

// Blur stops the strip handling keys.
func (t *Tabs) Blur() { t.focused = false }

// Focused reports whether the strip holds focus.
func (t Tabs) Focused() bool { return t.focused }

// AtEdge implements Boundary.
func (t Tabs) AtEdge(dir Direction) bool {
	switch dir {
	case Left:
		return t.active <= 0
	case Right:
		return t.active == len(t.tabs)-1
	default:
		return true
	}
}

// ShortHelp implements help.KeyMap.
func (t Tabs) ShortHelp() []key.Binding { return t.KeyMap.ShortHelp() }

// FullHelp implements help.KeyMap.
func (t Tabs) FullHelp() [][]key.Binding { return t.KeyMap.FullHelp() }

// Update handles key presses while focused, left clicks, and size messages.
func (t Tabs) Update(msg tea.Msg) (Tabs, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		if !t.focused {
			return t, nil
		}
		switch {
		case key.Matches(msg, t.KeyMap.Prev):
			return t.activate(t.active - 1)
		case key.Matches(msg, t.KeyMap.Next):
			return t.activate(t.active + 1)
		case key.Matches(msg, t.KeyMap.Jump):
			s := msg.String()
			return t.activate(int(s[len(s)-1]-'0') - 1)
		}
	case tea.MouseClickMsg:
		return t.click(msg)
	}
	return t, nil
}

// activate makes tab i active and reports the change; an index outside the strip
// or the active tab itself changes nothing.
func (t Tabs) activate(i int) (Tabs, tea.Cmd) {
	if i < 0 || i >= len(t.tabs) || i == t.active {
		return t, nil
	}
	t.active = i
	return t, Emit(TabChangedMsg{ID: t.id, Index: i, Tab: t.tabs[i]})
}

func (t Tabs) click(msg tea.MouseClickMsg) (Tabs, tea.Cmd) {
	if msg.Button != tea.MouseLeft || msg.Y != 0 {
		return t, nil
	}
	spans, scroll := t.layout()
	x := msg.X + scroll
	for i, span := range spans {
		if x < span.start || x >= span.end {
			continue
		}
		if t.tabs[i].Closable && x == span.end-2 {
			return t, Emit(TabCloseMsg{ID: t.id, Index: i, Tab: t.tabs[i]})
		}
		return t.activate(i)
	}
	return t, nil
}

// tabSpan is the column range of a tab in the whole strip, before scrolling.
type tabSpan struct{ start, end int }

// text is the text drawn for tab i, padded with a space on each side.
func (t Tabs) text(i int) string {
	tab := t.tabs[i]
	text := " " + tab.Title + " "
	if t.style.Radio {
		glyph := "○"
		if i == t.active {
			glyph = "●"
		}
		text = " " + glyph + text
	}
	if tab.Closable {
		text += tabsCloseMark + " "
	}
	return text
}

// labelWidth is the width of the label and the space after it.
func (t Tabs) labelWidth() int {
	if t.label == "" {
		return 0
	}
	return ansi.StringWidth(t.label) + 1
}

// layout computes where every tab sits and how far the strip is scrolled so that
// the active tab is visible. It is a pure function of the model, used by both
// View and the mouse handling.
func (t Tabs) layout() (spans []tabSpan, scroll int) {
	spans = make([]tabSpan, len(t.tabs))
	col := t.labelWidth()
	for i := range t.tabs {
		w := ansi.StringWidth(t.text(i))
		spans[i] = tabSpan{col, col + w}
		col += w
	}
	if t.active >= 0 {
		a := spans[t.active]
		scroll = min(max(a.end-t.width, 0), a.start)
	}
	return spans, scroll
}

// View renders the strip on one row of the width given to SetSize.
func (t Tabs) View() string {
	if t.width <= 0 {
		return ""
	}
	var sb strings.Builder
	if t.label != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(theme.MutedColor()).Render(t.label + " "))
	}
	for i := range t.tabs {
		sb.WriteString(t.tabStyle(i).Render(t.text(i)))
	}
	_, scroll := t.layout()
	return Fit(ansi.Cut(sb.String(), scroll, scroll+t.width), t.width, 1)
}

// tabStyle is the style of tab i.
func (t Tabs) tabStyle(i int) lipgloss.Style {
	if i != t.active {
		style := lipgloss.NewStyle().Foreground(theme.MutedColor())
		return style.Underline(t.style.Underscore)
	}
	if t.focused {
		return theme.SelectedStyle(true)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(theme.TextColor())
}
