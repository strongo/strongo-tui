// Package widgets is the component kit of strongo-tui. Its components follow
// the Elm architecture of Bubble Tea: a component is a model with
// Update(tea.Msg) (T, tea.Cmd) and View() string, its state lives in the model,
// and what happens in it is reported to the parent as messages.
//
// # Conventions
//
// Every component follows the same shape as the models in charm.land/bubbles:
//
//   - a constructor returns a value: NewList(id, items);
//   - setters have pointer receivers (SetSize, SetItems, Focus, Blur), and
//     Update has a value receiver and returns the updated model;
//   - the parent gives the component its size with SetSize, or by forwarding a
//     tea.WindowSizeMsg to Update; View returns exactly that many lines of that
//     many columns;
//   - a component ignores key presses until Focus is called, and mouse events
//     are given in coordinates relative to its top-left cell;
//   - keys are key.Binding values collected in a KeyMap field, so an
//     application can rebind them and a help bar can list them (each KeyMap
//     implements help.KeyMap);
//   - colours come from the active theme (package theme) when View runs.
//
// # Messages instead of callbacks
//
// A component never calls back into its parent. It returns a command that
// delivers a message, and the parent handles the message in its own Update:
//
//	list, cmd = list.Update(msg)
//	...
//	case widgets.ItemSelectedMsg:
//		if msg.ID == "projects" { return m, openProject(msg.Item) }
//
// Each message carries the ID the component was created with, so a screen with
// two lists can tell them apart. Asynchronous work is a tea.Cmd that returns a
// result message; components never start goroutines.
//
// The catalogue of messages:
//
//	List         ItemHighlightedMsg, ItemSelectedMsg
//	Tree         NodeHighlightedMsg, NodeSelectedMsg
//	Form         FieldChangedMsg, SubmitMsg, CancelMsg, ButtonPressedMsg
//	Tabs         TabChangedMsg, TabCloseMsg
//	Modal        ModalDoneMsg
//	Breadcrumbs  CrumbSelectedMsg
//
// # Focus movement
//
// A screen that wants the shell (package nav) to move focus with the arrow keys
// implements Boundary: the shell asks AtEdge before it turns Up at the top of a
// list into "go to the breadcrumbs". Components that edit text implement Editor,
// which suspends application-wide key bindings while a cursor is active.
//
// # Framing and layout
//
// Frame draws the bordered, titled box every panel sits in, with the focus
// colour on its border; Split divides a length into fixed and proportional
// parts; Fit, Overlay and Center are the string helpers behind them. All are
// pure functions of their arguments.
package widgets
