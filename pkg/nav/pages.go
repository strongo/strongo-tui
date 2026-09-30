package nav

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

// windowSize is the message that tells a screen its size.
func windowSize(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }

// tellMenu and tellContent tell a screen that it gained or lost focus, unless it
// already knows.
func (m *Model) tellMenu(focused bool) tea.Cmd {
	if m.menuFocused == focused {
		return nil
	}
	m.menuFocused = focused
	s, cmd := m.menu().Update(ScreenFocusMsg{Focused: focused})
	m.setMenu(s)
	return cmd
}

func (m *Model) tellContent(focused bool) tea.Cmd {
	if m.contentFocused == focused {
		return nil
	}
	m.contentFocused = focused
	s, cmd := m.content().Update(ScreenFocusMsg{Focused: focused})
	m.setContent(s)
	return cmd
}

// syncFocus tells the screens and the breadcrumbs where focus is.
func (m *Model) syncFocus() tea.Cmd {
	if m.zone == FocusToBreadcrumbs {
		m.crumbs.Focus()
	} else {
		m.crumbs.Blur()
	}
	return tea.Batch(m.tellMenu(m.zone == FocusToMenu), m.tellContent(m.zone == FocusToContent))
}

// resizeMenu and resizeContent send a screen its size.
func (m *Model) resizeMenu() tea.Cmd {
	w, h := m.menuSize()
	s, cmd := m.menu().Update(windowSize(w, h))
	m.setMenu(s)
	return cmd
}

func (m *Model) resizeContent() tea.Cmd {
	w, h := m.contentSize()
	s, cmd := m.content().Update(windowSize(w, h))
	m.setContent(s)
	return cmd
}

// relayout applies a new terminal size or menu visibility.
func (m *Model) relayout() tea.Cmd {
	m.zone = m.resolveZone(m.zone)
	m.crumbs.SetSize(m.geometry().crumbs.w, 1)
	return tea.Batch(m.resizeMenu(), m.resizeContent(), m.syncFocus())
}

// mountContent installs a content screen: it is initialised, sized and told
// whether it holds focus.
func (m *Model) mountContent(s Screen) tea.Cmd {
	m.setContent(s)
	m.contentFocused = false
	init := s.Init()
	return tea.Batch(init, m.resizeContent())
}

func (m *Model) mountMenu(s Screen) tea.Cmd {
	m.setMenu(s)
	m.menuFocused = false
	init := s.Init()
	return tea.Batch(init, m.resizeMenu())
}

// arrive moves focus to where a page wants it and tells everyone.
func (m *Model) arrive(focus FocusTo) tea.Cmd {
	m.zone = m.resolveZone(focus)
	return m.syncFocus()
}

// refreshCrumbs rebuilds the breadcrumb trail from the pages, or the override.
func (m *Model) refreshCrumbs() {
	if m.crumbsOverride != nil {
		m.crumbs.SetCrumbs(m.crumbsOverride)
		return
	}
	crumbs := make([]widgets.Crumb, len(m.pages))
	for i, p := range m.pages {
		crumbs[i] = widgets.Crumb{Title: p.Title, Color: p.Color}
	}
	m.crumbs.SetCrumbs(crumbs)
}

// push shows page above the current one.
func (m *Model) push(page Page) tea.Cmd {
	cmds := []tea.Cmd{m.tellContent(false)}
	if page.Menu != nil {
		cmds = append(cmds, m.tellMenu(false))
	}
	m.crumbsOverride = nil
	m.pages = append(m.pages, page)
	if page.Menu != nil {
		cmds = append(cmds, m.mountMenu(page.Menu))
	}
	cmds = append(cmds, m.mountContent(m.pages[m.top()].Content))
	m.refreshCrumbs()
	return tea.Batch(append(cmds, m.arrive(page.Focus))...)
}

// popTo closes pages until depth remain.
func (m *Model) popTo(depth int) tea.Cmd {
	depth = max(depth, 1)
	if depth >= len(m.pages) {
		return nil
	}
	menuBefore := m.menuPage()
	m.pages = m.pages[:depth]
	m.crumbsOverride = nil
	// The revealed screens are different from the ones that were showing.
	m.contentFocused = false
	cmds := []tea.Cmd{m.resizeContent()}
	if m.menuPage() != menuBefore {
		m.menuFocused = false
		cmds = append(cmds, m.resizeMenu())
	}
	m.refreshCrumbs()
	return tea.Batch(append(cmds, m.arrive(m.pages[m.top()].Focus))...)
}

// replace swaps the current page.
func (m *Model) replace(page Page) tea.Cmd {
	if len(m.pages) == 1 {
		return m.reset(page)
	}
	menuBefore := m.menuPage()
	m.pages = m.pages[:m.top()]
	// The screens that were focused are gone.
	m.contentFocused = false
	if m.menuPage() != menuBefore {
		m.menuFocused = false
	}
	return m.push(page)
}

// reset shows page as the only page. A page without a menu keeps the menu that
// is showing, which is not initialised again.
func (m *Model) reset(page Page) tea.Cmd {
	keptMenu := page.Menu == nil
	if keptMenu {
		page.Menu = m.pages[m.menuPage()].Menu
	}
	m.pages = []Page{page}
	m.crumbsOverride = nil
	m.contentFocused = false
	var cmds []tea.Cmd
	if keptMenu {
		cmds = append(cmds, m.resizeMenu())
	} else {
		m.menuFocused = false
		cmds = append(cmds, m.mountMenu(page.Menu))
	}
	cmds = append(cmds, m.mountContent(page.Content))
	m.refreshCrumbs()
	return tea.Batch(append(cmds, m.arrive(page.Focus))...)
}

// setPanels replaces the screens of the current page.
func (m *Model) setPanels(menu, content Screen, focus FocusTo) tea.Cmd {
	var cmds []tea.Cmd
	if menu != nil {
		cmds = append(cmds, m.mountMenu(menu))
	}
	if content != nil {
		cmds = append(cmds, m.mountContent(content))
	}
	if focus == FocusToKeep {
		focus = m.zone
	}
	return tea.Batch(append(cmds, m.arrive(focus))...)
}

// setFocus moves focus to a zone.
func (m *Model) setFocus(to FocusTo) tea.Cmd {
	if to == FocusToKeep {
		return nil
	}
	if to == FocusToBreadcrumbs || to == FocusToLogin {
		return m.focusHeader(to)
	}
	m.zone = m.resolveZone(to)
	return m.syncFocus()
}

// focusHeader moves focus into the header, remembering the panel it came from.
func (m *Model) focusHeader(to FocusTo) tea.Cmd {
	if m.zone == FocusToMenu || m.zone == FocusToContent {
		m.from = m.zone
	}
	m.zone = to
	return m.syncFocus()
}

// leaveHeader returns focus to the panel it came from.
func (m *Model) leaveHeader() tea.Cmd {
	m.zone = m.resolveZone(m.from)
	m.from = FocusToMenu
	return m.syncFocus()
}

// crumbSelected handles the activation of a breadcrumb.
func (m *Model) crumbSelected(msg widgets.CrumbSelectedMsg) tea.Cmd {
	if msg.ID != crumbsID || msg.Index >= len(m.pages) {
		return nil
	}
	page := m.pages[msg.Index]
	if page.OnCrumb != nil {
		return page.OnCrumb
	}
	if msg.Index == m.top() {
		return nil
	}
	return m.popTo(msg.Index + 1)
}
