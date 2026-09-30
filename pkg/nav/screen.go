package nav

import (
	"image/color"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Screen is what the menu and content panels host. It is shaped like a
// tea.Model whose View returns a string, so the components of
// charm.land/bubbles and strongo-tui compose into screens the usual way.
type Screen interface {
	// Init returns the command to run when the screen is mounted.
	Init() tea.Cmd
	// Update handles a message and returns the updated screen.
	Update(msg tea.Msg) (Screen, tea.Cmd)
	// View renders the screen into the size the shell last announced with a
	// tea.WindowSizeMsg.
	View() string
}

// Titled is implemented by screens that have a title for the border of their
// panel.
type Titled interface{ Title() string }

// Borderless is implemented by screens that draw their panel without a border.
type Borderless interface{ Borderless() bool }

// KeyCapturer is implemented by screens that claim keys the shell would
// otherwise handle itself, such as Ctrl+C for copy or Esc while a filter is
// open. A claimed key goes to the screen untouched.
type KeyCapturer interface {
	CapturesKey(msg tea.KeyPressMsg) bool
}

// ScreenFocusMsg tells a screen that it gained or lost focus. Screens forward it
// to their components' Focus and Blur.
type ScreenFocusMsg struct{ Focused bool }

// FocusTo names a place focus can go.
type FocusTo int

// Focus targets. The zero value is the content, which is where focus goes after
// navigation unless a page says otherwise.
const (
	// FocusToContent focuses the content panel.
	FocusToContent FocusTo = iota
	// FocusToMenu focuses the menu panel.
	FocusToMenu
	// FocusToKeep leaves focus where it is.
	FocusToKeep
	// FocusToBreadcrumbs focuses the breadcrumb trail.
	FocusToBreadcrumbs
	// FocusToLogin focuses the button at the right end of the header.
	FocusToLogin
)

// Page is one step of the navigation stack.
type Page struct {
	// Title is the breadcrumb of the page.
	Title string
	// Color colours the breadcrumb; the theme's accent colour when nil.
	Color color.Color
	// Menu is the screen in the menu panel. A nil Menu keeps the menu of the
	// page below.
	Menu Screen
	// Content is the screen in the content panel.
	Content Screen
	// Focus is where focus goes when the page is shown. The zero value is the
	// content.
	Focus FocusTo
	// OnCrumb, when not nil, runs instead of PopTo when the page's breadcrumb is
	// activated; use it for flows that rebuild their screens rather than pop.
	OnCrumb tea.Cmd
}

// ShortHelper is implemented by screens that list key bindings in the actions
// bar while they have focus. Every help.KeyMap is one.
type ShortHelper interface{ ShortHelp() []key.Binding }

// shortHelp returns the bindings s lists in the actions bar, if it does.
func shortHelp(s Screen) []key.Binding {
	if k, ok := s.(ShortHelper); ok {
		return k.ShortHelp()
	}
	return nil
}
