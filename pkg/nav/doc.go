// Package nav is the navigation shell of strongo-tui: one Bubble Tea model that
// lays out a header with breadcrumbs, a menu panel, a content panel and an
// actions bar, routes keys, mouse events and messages to the screens it hosts,
// keeps a stack of pages, and shows alerts.
//
// The shell follows the Elm architecture. Screens are models, navigation is a
// message, and nothing calls back:
//
//	tui := nav.New(nav.Page{Title: "Home", Menu: mainMenu, Content: welcome})
//	tea.NewProgram(tui).Run()            // or nav.Run(tui)
//
//	// inside a screen's Update:
//	return s, nav.Push(nav.Page{Title: "Projects", Content: newProjects()})
//
// # Layout
//
//	+--------------------------------------------------------------+
//	| Home > Projects > Demo                             (l) Login |  header
//	+----------------+---------------------------------------------+
//	| menu (30 cols) | content                                     |  body
//	|                |                                             |
//	+----------------+---------------------------------------------+
//	| Enter select  Ctrl+Q Quit  F1 Help                          |  actions bar
//	+--------------------------------------------------------------+
//
// The menu is hidden on terminals narrower than MinWidthForMenu columns. The
// shell draws the border of each panel (widgets.Frame), takes its title from a
// screen that implements Titled, and hands the screen the size of the area
// inside the border as a tea.WindowSizeMsg.
//
// # Screens
//
// A Screen is a tea.Model-shaped value: Init, Update returning the updated
// Screen, and View returning a string. Optional interfaces let a screen tell the
// shell more:
//
//   - Titled: the title in the panel's border;
//   - Borderless: draw the panel without a border;
//   - widgets.Boundary: whether an arrow key should leave the screen (see Focus);
//   - widgets.Editor: the screen is editing text, so application-wide key
//     bindings step aside;
//   - KeyCapturer: the screen claims a key the shell would otherwise handle;
//   - ShortHelper: the bindings listed in the actions bar while it has focus.
//
// The shell sends a screen a tea.WindowSizeMsg with its own size whenever it is
// mounted or resized, and a ScreenFocusMsg when it gains or loses focus. Every
// other message that is not a key or mouse event, such as the result of an
// asynchronous load, is delivered to the menu and the content of the current
// page.
//
// # Pages and messages
//
// Pages form a stack. Their titles are the breadcrumbs. A screen navigates by
// returning one of the commands below; the shell handles the message they carry:
//
//	Push, Pop, PopTo, Replace, Reset   the stack
//	SetPanels                          the screens of the current page
//	SetBreadcrumbs                     the trail, for flows outside the stack
//	SetFocus                           the focus zone
//	Alert, ShowError                   modals and error content
//	SetActions                         the application's actions
//
// The shell in turn reports LoginMsg (the header button), HelpMsg (F1) and
// whatever messages the application's Actions carry.
//
// # Focus
//
// Focus lives in one of four zones: the breadcrumbs, the login button, the menu
// and the content. Arrow keys move it between zones when the focused screen is
// at the edge in that direction: Up at the top goes to the breadcrumbs, Right
// from the menu to the content, Left from the content to the menu, Shift+Tab
// from a panel to the breadcrumbs, Right past the last breadcrumb to the login
// button, Down from the header back to where focus came from. A screen that does
// not implement widgets.Boundary keeps its arrow keys.
//
// # Keys
//
// Bindings are key.Binding values in KeyMap. Ctrl+Q always quits. Actions bind
// application-wide keys to messages and are ignored while the focused screen is
// editing text. Ctrl+C quits unless the focused screen claims it.
//
// # Testing
//
// Package navtest drives a Model without a terminal.
package nav
