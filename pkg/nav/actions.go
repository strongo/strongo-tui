package nav

import (
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Action is an entry of the actions bar: a key binding that delivers a message.
// The bar shows the binding's help text; clicking the entry or pressing its keys
// (while the focused screen is not editing) sends Msg.
type Action struct {
	// ID identifies the action and must be unique and not empty.
	ID string
	// Binding holds the keys and the help text shown in the bar.
	Binding key.Binding
	// Msg is delivered when the action is triggered.
	Msg tea.Msg
	// Hidden keeps the action out of the bar but bound.
	Hidden bool
}

// defaultActions are the shell's own Quit and Help.
func defaultActions() []Action {
	return []Action{
		{ID: "Quit", Binding: key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "quit")), Msg: tea.QuitMsg{}},
		{ID: "Help", Binding: key.NewBinding(key.WithKeys("f1"), key.WithHelp("f1", "help")), Msg: HelpMsg{}},
	}
}

// checkActions panics for an action without an ID or with a duplicate one: they
// are programming errors.
func checkActions(existing, added []Action) {
	seen := make(map[string]bool, len(existing)+len(added))
	for _, a := range existing {
		seen[a.ID] = true
	}
	for i, a := range added {
		if a.ID == "" {
			panic("nav: actions[" + strconv.Itoa(i) + "] has no ID, help=" + a.Binding.Help().Desc)
		}
		if seen[a.ID] {
			panic("nav: an attempt to register an action with already registered ID=" + a.ID)
		}
		seen[a.ID] = true
	}
}
