package widgets

import tea "charm.land/bubbletea/v2"

// Direction is a direction of navigation.
type Direction int

// Directions.
const (
	Up Direction = iota
	Down
	Left
	Right
)

// Boundary is implemented by components that can tell whether the selection
// sits at an edge. The shell asks the focused screen before it turns an arrow
// key into a focus move: Up at the top edge goes to the breadcrumbs, Left at the
// left edge goes to the menu, and so on. A component that is not a Boundary
// keeps all arrow keys for itself.
//
// AtEdge is a pure query of the model state, so it is as testable as View.
type Boundary interface {
	AtEdge(dir Direction) bool
}

// Editor is implemented by components that may currently own a text cursor.
// While Editing reports true the shell keeps application-wide key bindings out
// of the way so that text-editing keys keep their meaning.
type Editor interface {
	Editing() bool
}

// Emit returns a command that delivers msg to Update. Components report what
// happened with messages, and Emit is how they send them.
func Emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}
