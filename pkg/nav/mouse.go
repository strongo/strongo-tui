package nav

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

// relative returns a click or wheel message with its coordinates moved by
// (-dx, -dy), so a screen sees them relative to its own top-left cell.
func relative(msg tea.MouseMsg, dx, dy int) tea.MouseMsg {
	m := msg.Mouse()
	m.X, m.Y = m.X-dx, m.Y-dy
	if _, isClick := msg.(tea.MouseClickMsg); isClick {
		return tea.MouseClickMsg(m)
	}
	return tea.MouseWheelMsg(m)
}

// handleMouse routes a mouse event to the part of the screen under the pointer.
// A left click also moves focus there; the wheel scrolls without moving focus.
func (m *Model) handleMouse(msg tea.MouseMsg) tea.Cmd {
	mouse := msg.Mouse()
	if m.alert != nil {
		return m.alertMouse(msg)
	}
	var click bool
	switch msg.(type) {
	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft {
			return nil
		}
		click = true
	case tea.MouseWheelMsg:
	default:
		return nil
	}
	g := m.geometry()
	switch {
	case g.crumbs.contains(mouse.X, mouse.Y):
		var cmds []tea.Cmd
		if click {
			cmds = append(cmds, m.focusHeader(FocusToBreadcrumbs))
		}
		crumbs, cmd := m.crumbs.Update(relative(msg, g.crumbs.x, g.crumbs.y))
		m.crumbs = crumbs
		return tea.Batch(append(cmds, cmd)...)
	case click && g.login.contains(mouse.X, mouse.Y):
		return tea.Batch(m.focusHeader(FocusToLogin), widgets.Emit(LoginMsg{}))
	case g.menu.contains(mouse.X, mouse.Y):
		return m.mouseToPanel(msg, g.menu, FocusToMenu, click)
	case g.content.contains(mouse.X, mouse.Y):
		return m.mouseToPanel(msg, g.content, FocusToContent, click)
	case click && g.actions.contains(mouse.X, mouse.Y):
		return m.clickAction(mouse.X)
	}
	return nil
}

// mouseToPanel focuses a panel on a click and hands the event to its screen.
func (m *Model) mouseToPanel(msg tea.MouseMsg, r rect, zone FocusTo, click bool) tea.Cmd {
	var cmds []tea.Cmd
	if click {
		cmds = append(cmds, m.setFocus(zone))
	}
	s := m.menu()
	if zone == FocusToContent {
		s = m.content()
	}
	frame := frameFor(s, false)
	ox, oy := frame.Origin()
	iw, ih := frame.Inner(r.w, r.h)
	mouse := msg.Mouse()
	x, y := mouse.X-r.x-ox, mouse.Y-r.y-oy
	if x < 0 || y < 0 || x >= iw || y >= ih {
		return tea.Batch(cmds...)
	}
	s, cmd := s.Update(relative(msg, r.x+ox, r.y+oy))
	if zone == FocusToContent {
		m.setContent(s)
	} else {
		m.setMenu(s)
	}
	return tea.Batch(append(cmds, cmd)...)
}

// clickAction triggers the action bar item at column x.
func (m *Model) clickAction(x int) tea.Cmd {
	_, spans, actions := m.actionBar()
	for i, span := range spans {
		if x >= span[0] && x < span[1] {
			return widgets.Emit(actions[i].Msg)
		}
	}
	return nil
}

// alertMouse lets the open alert receive clicks on its buttons; clicks elsewhere
// are swallowed.
func (m *Model) alertMouse(msg tea.MouseMsg) tea.Cmd {
	lines := strings.Split(m.alert.box(m.width, m.height), "\n")
	boxW := 0
	for _, l := range lines {
		boxW = max(boxW, ansi.StringWidth(l))
	}
	x0, y0 := max((m.width-boxW)/2, 0), max((m.height-len(lines))/2, 0)
	mouse := msg.Mouse()
	if mouse.X < x0 || mouse.Y < y0 || mouse.X >= x0+boxW || mouse.Y >= y0+len(lines) {
		return nil
	}
	modal, cmd := m.alert.modal.Update(relative(msg, x0, y0))
	m.alert.modal = modal
	return cmd
}
