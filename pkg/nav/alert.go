package nav

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuigoff/tuigoff/pkg/theme"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

// alertExpiredMsg closes the alert with the given id once its time is up.
type alertExpiredMsg struct{ id int }

type alertState struct {
	id        int
	modal     widgets.Modal
	focusBack FocusTo
}

// box renders the modal for a terminal of the given size.
func (a *alertState) box(width, height int) string {
	a.modal.SetMaxSize(width, height)
	return a.modal.View()
}

// showAlert opens an alert, replacing the current one.
func (m *Model) showAlert(msg AlertMsg) tea.Cmd {
	id := 1
	if m.alert != nil {
		id = m.alert.id + 1
	}
	modal := widgets.NewModal(alertID)
	modal.SetTitle(msg.Title)
	modal.SetText(msg.Message)
	modal.SetButtons("OK")
	modal.SetColors(widgets.ModalColors{
		Background:       theme.DarkRed,
		Text:             theme.White,
		Border:           theme.WhiteSmoke,
		ButtonBackground: theme.Black,
		ButtonText:       theme.White,
	})
	m.alert = &alertState{id: id, modal: modal, focusBack: msg.FocusBack}
	if msg.Timeout <= 0 {
		return nil
	}
	return m.tick(msg.Timeout, func(time.Time) tea.Msg { return alertExpiredMsg{id: id} })
}

// closeAlert closes the alert with the given id, if it is still the open one.
func (m *Model) closeAlert(id int) tea.Cmd {
	if m.alert == nil || m.alert.id != id {
		return nil
	}
	back := m.alert.focusBack
	m.alert = nil
	return m.setFocus(back)
}

// modalDone handles the dismissal of the alert modal.
func (m *Model) modalDone(msg widgets.ModalDoneMsg) tea.Cmd {
	if msg.ID != alertID || m.alert == nil {
		return nil
	}
	return m.closeAlert(m.alert.id)
}

// alertKey routes a key press to the open alert. Ctrl+C still quits.
func (m *Model) alertKey(msg tea.KeyPressMsg) tea.Cmd {
	if key.Matches(msg, m.keys.Interrupt) {
		return tea.Quit
	}
	modal, cmd := m.alert.modal.Update(msg)
	m.alert.modal = modal
	return cmd
}

// staticScreen shows a text.
type staticScreen struct {
	title string
	text  string
	err   bool
	w, h  int
}

// Static returns a screen that shows text under a title. It is what an empty
// panel shows, and a handy placeholder.
func Static(title, text string) Screen { return staticScreen{title: title, text: text} }

func (s staticScreen) Init() tea.Cmd { return nil }

func (s staticScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		s.w, s.h = size.Width, size.Height
	}
	return s, nil
}

func (s staticScreen) View() string {
	style := lipgloss.NewStyle().Width(max(s.w, 1))
	if s.err {
		style = style.Foreground(theme.ErrorColor())
	}
	return widgets.Fit(style.Render(s.text), s.w, s.h)
}

// Title implements Titled.
func (s staticScreen) Title() string { return s.title }

// AtEdge implements widgets.Boundary: a static text never keeps an arrow key.
func (s staticScreen) AtEdge(widgets.Direction) bool { return true }

// errorScreen is the content ShowError installs.
func errorScreen(err error) Screen {
	return staticScreen{title: "Error", text: err.Error(), err: true}
}
