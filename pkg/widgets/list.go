package widgets

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/tuigoff/tuigoff/pkg/theme"
)

// MenuItem is a ready-made list item: a label, an optional second line of
// detail, an optional shortcut key and an opaque reference the parent can use to
// find out what the item stands for.
type MenuItem struct {
	// ID identifies the item for the parent.
	ID string
	// Label is the first line of the item.
	Label string
	// Detail, when not empty, is shown muted on a second line.
	Detail string
	// Shortcut, when not zero, is drawn as "(x) " before the label and selects
	// the item when the key is typed.
	Shortcut rune
	// Ref is any value the parent wants to get back with the item.
	Ref any
}

// Title implements list.DefaultItem.
func (i MenuItem) Title() string { return i.Label }

// Description implements list.DefaultItem.
func (i MenuItem) Description() string { return i.Detail }

// FilterValue implements list.Item.
func (i MenuItem) FilterValue() string { return i.Label }

// ItemHighlightedMsg reports that the highlighted item of a List changed
// because of a message passed to Update. Index is the position among the visible
// items (those matching the filter, when there is one).
type ItemHighlightedMsg struct {
	ID    string
	Index int
	Item  list.Item
}

// ItemSelectedMsg reports that the highlighted item of a List was chosen with
// Enter, with its shortcut key, or by a click on the already highlighted row.
type ItemSelectedMsg struct {
	ID    string
	Index int
	Item  list.Item
}

// listConfirm is the binding that chooses the highlighted item.
var listConfirm = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))

// DefaultListKeyMap returns the key bindings of a List: bubbles' own map with
// j, k, h, l, g, G, u, d, b, f and the quit and help keys taken away, because
// letters belong to item shortcuts and Left and Right belong to the shell.
func DefaultListKeyMap() list.KeyMap {
	km := list.DefaultKeyMap()
	km.CursorUp = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up"))
	km.CursorDown = key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down"))
	km.PrevPage = key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "prev page"))
	km.NextPage = key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "next page"))
	km.GoToStart = key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "first"))
	km.GoToEnd = key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "last"))
	km.ShowFullHelp = key.NewBinding()
	km.CloseFullHelp = key.NewBinding()
	km.Quit = key.NewBinding()
	km.ForceQuit = key.NewBinding()
	return km
}

// listDelegate draws the rows of a List: one line per item, or two when some
// item has a detail. It is a value so that the model can be copied freely.
type listDelegate struct {
	focused     bool
	twoLines    bool
	hasShortcut bool
}

func (d listDelegate) Height() int {
	if d.twoLines {
		return 2
	}
	return 1
}

func (listDelegate) Spacing() int { return 0 }

func (listDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

// listHintWidth is the width of the shortcut column.
const listHintWidth = 4

func (d listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	width := m.Width()
	title, detail := item.FilterValue(), ""
	if di, ok := item.(list.DefaultItem); ok {
		title, detail = di.Title(), di.Description()
	}
	var shortcut rune
	if mi, ok := item.(MenuItem); ok {
		shortcut = mi.Shortcut
	}
	hint := ""
	if d.hasShortcut {
		hint = strings.Repeat(" ", listHintWidth)
		if shortcut != 0 {
			hint = "(" + string(shortcut) + ") "
		}
	}
	first := hint + title
	second := strings.Repeat(" ", ansi.StringWidth(hint)) + detail

	if index == m.Index() && !m.SettingFilter() {
		style := theme.SelectedStyle(d.focused)
		out := style.Render(PadRight(ansi.Strip(first), width))
		if d.twoLines {
			out += "\n" + style.Render(PadRight(ansi.Strip(second), width))
		}
		fmt.Fprint(w, out) //nolint:errcheck
		return
	}
	out := PadRight(first, width)
	if hint != "" && shortcut != 0 {
		accent := lipgloss.NewStyle().Foreground(theme.AccentColor()).Render(strings.TrimRight(hint, " "))
		out = accent + " " + PadRight(title, width-listHintWidth)
	}
	if d.twoLines {
		muted := lipgloss.NewStyle().Foreground(theme.MutedColor())
		out += "\n" + muted.Render(PadRight(second, width))
	}
	fmt.Fprint(w, out) //nolint:errcheck
}

// List is a compact vertical menu built on bubbles' list.Model, which supplies
// the cursor movement, paging and optional filtering. The wrapper turns off the
// title, status bar, help and pagination dots, draws its own theme-coloured
// rows, and reports what happens with messages.
//
// Rows are one line high, or two when some item has a detail. The highlighted
// row is drawn across the full width in the theme's selected style. bubbles
// pages the list instead of scrolling it: moving past the last row of a page
// shows the next page.
//
// Messages, returned as commands by Update:
//   - ItemHighlightedMsg when a key, a click or the wheel changes the
//     highlighted item; SetItems and Select are setters and never emit
//   - ItemSelectedMsg on Enter, on the shortcut key of a MenuItem (after an
//     ItemHighlightedMsg when that moved the highlight), and on a left click on
//     the row that is already highlighted; a click on another row only
//     highlights it
//
// Keys, while focused: up, down, home, end, pgup and pgdown, all without
// wrapping around, and Enter. Letters are free for shortcuts. Unfocused, keys
// are ignored. Mouse clicks and the wheel work regardless of focus, with
// coordinates relative to the top-left cell; the wheel moves the highlight one
// row per notch.
//
// Filtering is off until SetFilteringEnabled(true); then "/" starts it, and while
// the filter input is active the list is an Editor: Editing reports true and
// every key goes to bubbles unmodified. The parent must pass all messages, not
// only keys, to Update while filtering, because bubbles delivers the matches
// asynchronously. Enabling filtering costs one row for the filter input.
//
// List is a Boundary: Up is at the edge when the first item is highlighted (or
// there are no items), Down when the last one is, and Left and Right always.
type List struct {
	id      string
	model   list.Model
	focused bool
}

// NewList creates a list. id identifies it in its messages.
func NewList(id string, items ...list.Item) List {
	l := List{id: id}
	l.model = list.New(nil, listDelegate{}, 0, 0)
	l.model.SetShowTitle(false)
	l.model.SetShowStatusBar(false)
	l.model.SetShowHelp(false)
	l.model.SetShowPagination(false)
	l.model.SetFilteringEnabled(false)
	l.model.Styles.TitleBar = lipgloss.NewStyle()
	l.model.FilterInput.Prompt = "/ "
	l.SetKeyMap(DefaultListKeyMap())
	l.SetItems(items...)
	return l
}

// KeyMap returns the key bindings of the underlying list model.
func (l List) KeyMap() list.KeyMap { return l.model.KeyMap }

// SetKeyMap replaces the key bindings. The quit bindings stay disabled.
func (l *List) SetKeyMap(km list.KeyMap) {
	l.model.KeyMap = km
	l.model.DisableQuitKeybindings()
	l.model.SetFilteringEnabled(l.model.FilteringEnabled())
}

// SetFilteringEnabled turns the "/" filter on or off; it is off by default.
func (l *List) SetFilteringEnabled(enabled bool) {
	l.model.SetFilteringEnabled(enabled)
	l.model.SetSize(l.model.Width(), l.model.Height())
}

// FilteringEnabled reports whether the filter is available.
func (l List) FilteringEnabled() bool { return l.model.FilteringEnabled() }

// SetItems replaces the items, ending any filter and keeping the highlight at
// the same position, or on the last item when there are fewer. It emits no
// message.
func (l *List) SetItems(items ...list.Item) {
	l.model.ResetFilter()
	index := l.model.Index()
	l.model.SetDelegate(l.delegateFor(items))
	l.model.SetItems(items)
	l.model.Select(max(min(index, len(items)-1), 0))
}

// Items returns the items, ignoring any filter.
func (l List) Items() []list.Item { return l.model.Items() }

func (l List) delegateFor(items []list.Item) listDelegate {
	d := listDelegate{focused: l.focused}
	for _, item := range items {
		if di, ok := item.(list.DefaultItem); ok && di.Description() != "" {
			d.twoLines = true
		}
		if mi, ok := item.(MenuItem); ok && mi.Shortcut != 0 {
			d.hasShortcut = true
		}
	}
	return d
}

// SetSize sets the size of the list in cells.
func (l *List) SetSize(width, height int) { l.model.SetSize(max(width, 0), max(height, 0)) }

// Focus makes the list handle keys and draws the highlight in the focused style.
func (l *List) Focus() {
	l.focused = true
	l.model.SetDelegate(l.delegateFor(l.model.Items()))
}

// Blur stops the list handling keys and dims the highlight.
func (l *List) Blur() {
	l.focused = false
	l.model.SetDelegate(l.delegateFor(l.model.Items()))
}

// Focused reports whether the list holds focus.
func (l List) Focused() bool { return l.focused }

// Editing implements Editor: true while the filter input is being typed into.
func (l List) Editing() bool { return l.focused && l.model.SettingFilter() }

// Index returns the position of the highlighted item among the visible items.
func (l List) Index() int { return l.model.Index() }

// SelectedItem returns the highlighted item, or nil when there is none.
func (l List) SelectedItem() list.Item { return l.model.SelectedItem() }

// Select highlights the item at index among the visible items, clamped to the
// list. It emits no message.
func (l *List) Select(index int) {
	l.model.Select(max(min(index, len(l.model.VisibleItems())-1), 0))
}

// AtEdge implements Boundary.
func (l List) AtEdge(dir Direction) bool {
	n := len(l.model.VisibleItems())
	switch dir {
	case Up:
		return n == 0 || l.model.Index() == 0
	case Down:
		return n == 0 || l.model.Index() == n-1
	default:
		return true
	}
}

// ShortHelp implements help.KeyMap.
func (l List) ShortHelp() []key.Binding {
	km := l.model.KeyMap
	return []key.Binding{km.CursorUp, km.CursorDown, listConfirm, km.Filter}
}

// FullHelp implements help.KeyMap.
func (l List) FullHelp() [][]key.Binding {
	km := l.model.KeyMap
	return [][]key.Binding{
		{km.CursorUp, km.CursorDown, km.PrevPage, km.NextPage, km.GoToStart, km.GoToEnd},
		{listConfirm, km.Filter, km.ClearFilter},
	}
}

// listPos identifies what is highlighted.
type listPos struct{ index, global int }

func (l List) pos() listPos { return listPos{l.model.Index(), l.model.GlobalIndex()} }

// forward passes msg to the bubbles model and reports a changed highlight.
func (l List) forward(msg tea.Msg) (List, tea.Cmd) {
	before := l.pos()
	var cmd tea.Cmd
	l.model, cmd = l.model.Update(msg)
	if l.pos() == before || l.model.SelectedItem() == nil {
		return l, cmd
	}
	return l, tea.Batch(cmd, l.highlighted())
}

func (l List) highlighted() tea.Cmd {
	return Emit(ItemHighlightedMsg{ID: l.id, Index: l.model.Index(), Item: l.model.SelectedItem()})
}

func (l List) selected() tea.Cmd {
	return Emit(ItemSelectedMsg{ID: l.id, Index: l.model.Index(), Item: l.model.SelectedItem()})
}

// shortcutIndex returns the position of the visible item with the shortcut r.
func (l List) shortcutIndex(r rune) int {
	for i, item := range l.model.VisibleItems() {
		if mi, ok := item.(MenuItem); ok && mi.Shortcut == r {
			return i
		}
	}
	return -1
}

// headerRows is the number of rows above the items: the filter row.
func (l List) headerRows() int {
	if l.model.FilteringEnabled() {
		return 1
	}
	return 0
}

// Update handles keys while focused, mouse clicks and wheel notches, size
// messages, and passes everything else (bubbles' asynchronous filter results and
// cursor blinks) to the underlying model.
func (l List) Update(msg tea.Msg) (List, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.SetSize(msg.Width, msg.Height)
		return l, nil
	case tea.KeyPressMsg:
		return l.updateKey(msg)
	case tea.MouseClickMsg:
		return l.updateClick(msg)
	case tea.MouseWheelMsg:
		return l.updateWheel(msg)
	}
	return l.forward(msg)
}

func (l List) updateKey(msg tea.KeyPressMsg) (List, tea.Cmd) {
	if !l.focused {
		return l, nil
	}
	if !l.model.SettingFilter() && l.model.SelectedItem() != nil {
		if key.Matches(msg, listConfirm) {
			return l, l.selected()
		}
		if r, n := utf8.DecodeRuneInString(msg.Text); n > 0 && n == len(msg.Text) {
			if i := l.shortcutIndex(r); i >= 0 {
				var cmds []tea.Cmd
				if i != l.model.Index() {
					l.model.Select(i)
					cmds = append(cmds, l.highlighted())
				}
				return l, tea.Sequence(append(cmds, l.selected())...)
			}
		}
	}
	return l.forward(msg)
}

func (l List) updateClick(msg tea.MouseClickMsg) (List, tea.Cmd) {
	if msg.Button != tea.MouseLeft || l.model.SettingFilter() {
		return l, nil
	}
	height := listDelegate{twoLines: l.twoLines()}.Height()
	y := msg.Y - l.headerRows()
	if y < 0 {
		return l, nil
	}
	i := l.model.Paginator.Page*l.model.Paginator.PerPage + y/height
	if i >= len(l.model.VisibleItems()) {
		return l, nil
	}
	if i == l.model.Index() {
		return l, l.selected()
	}
	l.model.Select(i)
	return l, l.highlighted()
}

func (l List) twoLines() bool { return l.delegateFor(l.model.Items()).twoLines }

func (l List) updateWheel(msg tea.MouseWheelMsg) (List, tea.Cmd) {
	before := l.pos()
	switch msg.Button {
	case tea.MouseWheelUp:
		l.model.CursorUp()
	case tea.MouseWheelDown:
		l.model.CursorDown()
	}
	if l.pos() == before || l.model.SelectedItem() == nil {
		return l, nil
	}
	return l, l.highlighted()
}

// View renders the list in exactly the size given to SetSize.
func (l List) View() string {
	w, h := l.model.Width(), l.model.Height()
	if w <= 0 || h <= 0 {
		return ""
	}
	m := l.model
	m.Styles.NoItems = lipgloss.NewStyle().Foreground(theme.MutedColor())
	s := m.FilterInput.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(theme.AccentColor())
	s.Focused.Text = lipgloss.NewStyle().Foreground(theme.TextColor())
	m.FilterInput.SetStyles(s)
	if len(m.Items()) == 0 && !m.FilteringEnabled() {
		return Blank(w, h)
	}
	return Fit(m.View(), w, h)
}
