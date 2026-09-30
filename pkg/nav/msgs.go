package nav

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

// Navigation is done with messages. A screen returns one of the commands below
// from its Update, and the shell handles the message the command delivers.

// PushMsg opens a page on top of the stack.
type PushMsg struct{ Page Page }

// Push returns the command that opens page above the current one; its title
// becomes the last breadcrumb.
func Push(page Page) tea.Cmd { return widgets.Emit(PushMsg{Page: page}) }

// PopMsg closes the current page.
type PopMsg struct{}

// Pop returns the command that closes the current page. The root page stays.
func Pop() tea.Cmd { return widgets.Emit(PopMsg{}) }

// PopToMsg closes pages until Depth remain.
type PopToMsg struct{ Depth int }

// PopTo returns the command that closes pages until depth remain: 1 goes to the
// root page, 2 to the page above it, and so on.
func PopTo(depth int) tea.Cmd { return widgets.Emit(PopToMsg{Depth: depth}) }

// ReplaceMsg swaps the current page.
type ReplaceMsg struct{ Page Page }

// Replace returns the command that replaces the current page, keeping the depth
// of the stack.
func Replace(page Page) tea.Cmd { return widgets.Emit(ReplaceMsg{Page: page}) }

// ResetMsg drops the whole stack and shows a single page.
type ResetMsg struct{ Page Page }

// Reset returns the command that replaces the whole stack with page.
func Reset(page Page) tea.Cmd { return widgets.Emit(ResetMsg{Page: page}) }

// SetPanelsMsg replaces the screens of the current page.
type SetPanelsMsg struct {
	// Menu replaces the menu; nil keeps it.
	Menu Screen
	// Content replaces the content; nil keeps it.
	Content Screen
	// Focus is where focus goes afterwards.
	Focus FocusTo
}

// SetPanels returns the command that replaces the menu and, or, the content of
// the current page without touching the breadcrumbs. A nil screen is kept.
func SetPanels(menu, content Screen, focus FocusTo) tea.Cmd {
	return widgets.Emit(SetPanelsMsg{Menu: menu, Content: content, Focus: focus})
}

// SetBreadcrumbsMsg replaces the trail.
type SetBreadcrumbsMsg struct{ Crumbs []widgets.Crumb }

// SetBreadcrumbs returns the command that shows crumbs instead of the titles of
// the stack until the next navigation. Activating one of them navigates to the
// page at the same position when there is one.
func SetBreadcrumbs(crumbs ...widgets.Crumb) tea.Cmd {
	return widgets.Emit(SetBreadcrumbsMsg{Crumbs: crumbs})
}

// SetFocusMsg moves focus.
type SetFocusMsg struct{ To FocusTo }

// SetFocus returns the command that moves focus to a zone.
func SetFocus(to FocusTo) tea.Cmd { return widgets.Emit(SetFocusMsg{To: to}) }

// AlertMsg shows a modal.
type AlertMsg struct {
	Title   string
	Message string
	// Timeout closes the alert by itself after that long; zero keeps it open
	// until it is dismissed.
	Timeout time.Duration
	// FocusBack is where focus goes when the alert closes.
	FocusBack FocusTo
}

// Alert returns the command that shows a red modal with an OK button above the
// layout. Enter or Esc dismisses it; a positive timeout closes it after that
// time. Showing a new alert replaces the current one.
func Alert(title, message string, timeout time.Duration, focusBack FocusTo) tea.Cmd {
	return widgets.Emit(AlertMsg{Title: title, Message: message, Timeout: timeout, FocusBack: focusBack})
}

// ErrorMsg shows an error in the content panel.
type ErrorMsg struct{ Err error }

// ShowError returns the command that replaces the content of the current page
// with err in red, keeping the menu.
func ShowError(err error) tea.Cmd { return widgets.Emit(ErrorMsg{Err: err}) }

// SetActionsMsg replaces the application's actions.
type SetActionsMsg struct{ Actions []Action }

// SetActions returns the command that replaces the actions the application
// registered, keeping the shell's own Quit and Help.
func SetActions(actions ...Action) tea.Cmd { return widgets.Emit(SetActionsMsg{Actions: actions}) }

// LoginMsg is sent when the button at the right end of the header is activated.
type LoginMsg struct{}

// HelpMsg is sent when the Help action is triggered. The shell does nothing
// else with it; an application can show its own help.
type HelpMsg struct{}

// Quit returns the command that ends the program.
func Quit() tea.Cmd { return tea.Quit }
