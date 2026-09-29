package nav

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

// KeyMap holds the key bindings of the shell.
type KeyMap struct {
	// Quit ends the program at any time, even while a screen is editing text.
	Quit key.Binding
	// Interrupt ends the program unless the focused screen claims the key.
	Interrupt key.Binding
	// Up, Down, Left and Right move focus between zones when the focused screen
	// is at the edge in that direction.
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	// ToHeader moves focus from a panel to the breadcrumbs.
	ToHeader key.Binding
	// LeaveHeader returns focus from the header to the panels.
	LeaveHeader key.Binding
	// Activate presses the focused header button.
	Activate key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:        key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "quit")),
		Interrupt:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Up:          key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:        key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		Left:        key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "left")),
		Right:       key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "right")),
		ToHeader:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "breadcrumbs")),
		LeaveHeader: key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("↓", "back")),
		Activate:    key.NewBinding(key.WithKeys("enter", "space"), key.WithHelp("enter", "select")),
	}
}

// ShortHelp implements help.KeyMap.
func (k KeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Quit} }

// FullHelp implements help.KeyMap.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Left, k.Right}, {k.ToHeader, k.Quit}}
}

// editing reports whether the screen is editing text.
func editing(s Screen) bool {
	e, ok := s.(widgets.Editor)
	return ok && e.Editing()
}

// handleKey routes a key press: quit first, then an open alert, then keys the
// focused screen claims, then the application's actions, then the zone.
func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	if key.Matches(msg, m.keys.Quit) {
		return tea.Quit
	}
	if m.alert != nil {
		return m.alertKey(msg)
	}
	s := m.focusedScreen()
	if c, ok := s.(KeyCapturer); ok && c.CapturesKey(msg) {
		return m.toScreen(msg)
	}
	if key.Matches(msg, m.keys.Interrupt) {
		return tea.Quit
	}
	if a, ok := m.bound(msg); ok && !editing(s) {
		return widgets.Emit(a.Msg)
	}
	switch m.zone {
	case FocusToBreadcrumbs:
		return m.crumbsKey(msg)
	case FocusToLogin:
		return m.loginKey(msg)
	default:
		return m.bodyKey(msg, s)
	}
}

// toScreen delivers a message to the screen that holds focus.
func (m *Model) toScreen(msg tea.Msg) tea.Cmd {
	if m.zone == FocusToMenu {
		s, cmd := m.menu().Update(msg)
		m.setMenu(s)
		return cmd
	}
	s, cmd := m.content().Update(msg)
	m.setContent(s)
	return cmd
}

// bodyKey handles a key while a panel holds focus: arrow keys at the edge of the
// screen move focus to the neighbouring zone, everything else is the screen's.
func (m *Model) bodyKey(msg tea.KeyPressMsg, s Screen) tea.Cmd {
	boundary, isBoundary := s.(widgets.Boundary)
	atEdge := func(dir widgets.Direction) bool { return isBoundary && boundary.AtEdge(dir) }
	switch {
	case key.Matches(msg, m.keys.ToHeader) && (!isBoundary || atEdge(widgets.Up)):
		return m.focusHeader(FocusToBreadcrumbs)
	case key.Matches(msg, m.keys.Up) && atEdge(widgets.Up):
		return m.focusHeader(FocusToBreadcrumbs)
	case key.Matches(msg, m.keys.Right) && m.zone == FocusToMenu && atEdge(widgets.Right):
		return m.setFocus(FocusToContent)
	case key.Matches(msg, m.keys.Left) && m.zone == FocusToContent && m.menuVisible() && atEdge(widgets.Left):
		return m.setFocus(FocusToMenu)
	}
	return m.toScreen(msg)
}

// crumbsKey handles a key while the breadcrumbs hold focus.
func (m *Model) crumbsKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.LeaveHeader):
		return m.leaveHeader()
	case key.Matches(msg, m.keys.Right) && m.crumbs.AtEdge(widgets.Right) && !m.noLogin:
		m.zone = FocusToLogin
		return m.syncFocus()
	}
	crumbs, cmd := m.crumbs.Update(msg)
	m.crumbs = crumbs
	return cmd
}

// loginKey handles a key while the header button holds focus.
func (m *Model) loginKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.LeaveHeader):
		return m.leaveHeader()
	case key.Matches(msg, m.keys.Left):
		m.zone = FocusToBreadcrumbs
		return m.syncFocus()
	case key.Matches(msg, m.keys.Activate):
		return widgets.Emit(LoginMsg{})
	}
	return nil
}
