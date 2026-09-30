package nav

import (
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/tuigoff/tuigoff/pkg/theme"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

// Layout constants.
const (
	// MenuWidth is the width of the menu panel in columns.
	MenuWidth = 30
	// MinWidthForMenu is the narrowest terminal that still shows the menu.
	MinWidthForMenu = 100

	defaultWidth  = 80
	defaultHeight = 24

	crumbsID = "nav.crumbs"
	alertID  = "nav.alert"
)

// TickFunc schedules a message, like tea.Tick. Alerts use it to close by
// themselves; tests replace it to control time.
type TickFunc func(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd

// Model is the root model of the shell. Create it with New and run it as a
// tea.Model, with tea.NewProgram or Run.
type Model struct {
	keys    KeyMap
	tick    TickFunc
	actions []Action

	pages  []Page
	zone   FocusTo
	from   FocusTo
	crumbs widgets.Breadcrumbs
	// crumbsOverride, when not nil, replaces the trail derived from the pages.
	crumbsOverride []widgets.Crumb
	login          widgets.Button
	noLogin        bool

	// menuFocused and contentFocused are what the screens were last told.
	menuFocused    bool
	contentFocused bool

	alert *alertState

	width, height int
	sized         bool
}

var _ tea.Model = Model{}

// Option customises New.
type Option func(*Model)

// WithKeyMap replaces the shell's key bindings.
func WithKeyMap(keys KeyMap) Option { return func(m *Model) { m.keys = keys } }

// WithActions registers application actions in the actions bar, after the
// shell's own Quit and Help. It panics for an action without an ID or with an ID
// that is already registered.
func WithActions(actions ...Action) Option {
	return func(m *Model) {
		checkActions(m.actions, actions)
		m.actions = append(m.actions, actions...)
	}
}

// WithTick replaces the timer used for alert expiry; the default is tea.Tick.
func WithTick(tick TickFunc) Option { return func(m *Model) { m.tick = tick } }

// WithLogin sets the label and shortcut hint of the button at the right end of
// the header; the default is "Login" with the hint (l).
func WithLogin(label string, shortcut rune) Option {
	return func(m *Model) { m.login = widgets.Button{Label: label, Shortcut: shortcut} }
}

// WithoutLogin removes the button at the right end of the header.
func WithoutLogin() Option { return func(m *Model) { m.noLogin = true } }

// New creates a shell showing root as its only page. The root page's Menu and
// Content are mounted when the program starts.
func New(root Page, options ...Option) Model {
	m := Model{
		keys:    DefaultKeyMap(),
		tick:    tea.Tick,
		actions: defaultActions(),
		width:   defaultWidth,
		height:  defaultHeight,
		crumbs:  widgets.NewBreadcrumbs(crumbsID),
		login:   widgets.Button{Label: "Login", Shortcut: 'l'},
		zone:    focusOrDefault(root.Focus),
		from:    FocusToMenu,
	}
	for _, option := range options {
		option(&m)
	}
	if root.Menu == nil {
		root.Menu = Static("Menu", "")
	}
	if root.Content == nil {
		root.Content = Static("Content", "")
	}
	m.pages = []Page{root}
	m.zone = m.resolveZone(m.zone)
	m.crumbs.SetSize(m.geometry().crumbs.w, 1)
	m.refreshCrumbs()
	return m
}

// focusOrDefault turns a page's Focus into a zone: focus that stays put is the
// content when there is nothing to keep.
func focusOrDefault(f FocusTo) FocusTo {
	if f == FocusToKeep {
		return FocusToContent
	}
	return f
}

// Init starts the screens of the root page. Their sizes arrive with the first
// tea.WindowSizeMsg.
func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.menu().Init(), m.content().Init())
}

// Zone returns the zone that holds focus.
func (m Model) Zone() FocusTo { return m.zone }

// Depth returns the number of pages on the stack.
func (m Model) Depth() int { return len(m.pages) }

// Breadcrumbs returns the trail as shown in the header.
func (m Model) Breadcrumbs() []widgets.Crumb { return m.crumbs.Crumbs() }

// Menu returns the screen in the menu panel.
func (m Model) Menu() Screen { return m.menu() }

// Content returns the screen in the content panel.
func (m Model) Content() Screen { return m.content() }

// AlertOpen reports whether an alert is showing.
func (m Model) AlertOpen() bool { return m.alert != nil }

// Size returns the terminal size the shell last saw.
func (m Model) Size() (width, height int) { return m.width, m.height }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.pages = slices.Clone(m.pages)
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height, m.sized = msg.Width, msg.Height, true
		cmd = m.relayout()
	case tea.BackgroundColorMsg:
		theme.SetTerminalBackground(msg.Color)
	case tea.KeyPressMsg:
		cmd = m.handleKey(msg)
	case tea.MouseMsg:
		cmd = m.handleMouse(msg)
	case widgets.CrumbSelectedMsg:
		cmd = m.crumbSelected(msg)
	case widgets.ModalDoneMsg:
		cmd = m.modalDone(msg)
	case alertExpiredMsg:
		cmd = m.closeAlert(msg.id)
	case AlertMsg:
		cmd = m.showAlert(msg)
	case ErrorMsg:
		cmd = m.setPanels(nil, errorScreen(msg.Err), FocusToContent)
	case PushMsg:
		cmd = m.push(msg.Page)
	case PopMsg:
		cmd = m.popTo(len(m.pages) - 1)
	case PopToMsg:
		cmd = m.popTo(msg.Depth)
	case ReplaceMsg:
		cmd = m.replace(msg.Page)
	case ResetMsg:
		cmd = m.reset(msg.Page)
	case SetPanelsMsg:
		cmd = m.setPanels(msg.Menu, msg.Content, msg.Focus)
	case SetBreadcrumbsMsg:
		m.crumbsOverride = slices.Clone(msg.Crumbs)
		m.refreshCrumbs()
	case SetFocusMsg:
		cmd = m.setFocus(msg.To)
	case SetActionsMsg:
		defaults := defaultActions()
		checkActions(defaults, msg.Actions)
		m.actions = append(defaults, msg.Actions...)
	default:
		cmd = m.forward(msg)
	}
	return m, cmd
}

// bind reports whether msg matches one of the shell's own action bindings, and
// the action it triggers.
func (m Model) bound(msg tea.KeyPressMsg) (Action, bool) {
	for _, a := range m.actions {
		if key.Matches(msg, a.Binding) {
			return a, true
		}
	}
	return Action{}, false
}

// top is the index of the current page.
func (m Model) top() int { return len(m.pages) - 1 }

// menuPage is the index of the page whose menu is shown.
func (m Model) menuPage() int {
	for i := m.top(); i > 0; i-- {
		if m.pages[i].Menu != nil {
			return i
		}
	}
	return 0
}

func (m Model) menu() Screen    { return m.pages[m.menuPage()].Menu }
func (m Model) content() Screen { return m.pages[m.top()].Content }

func (m *Model) setMenu(s Screen)    { m.pages[m.menuPage()].Menu = s }
func (m *Model) setContent(s Screen) { m.pages[m.top()].Content = s }

// menuVisible reports whether the terminal is wide enough for the menu. Until
// the terminal has reported its size the menu is assumed to fit.
func (m Model) menuVisible() bool { return !m.sized || m.width >= MinWidthForMenu }

// resolveZone maps a requested zone to one that can hold focus.
func (m Model) resolveZone(z FocusTo) FocusTo {
	if z == FocusToMenu && !m.menuVisible() {
		return FocusToContent
	}
	if z == FocusToKeep {
		return m.zone
	}
	return z
}

// focusedScreen returns the screen that holds focus, or nil in the header.
func (m Model) focusedScreen() Screen {
	switch m.zone {
	case FocusToMenu:
		return m.menu()
	case FocusToContent:
		return m.content()
	default:
		return nil
	}
}

// forward delivers a message to the menu and the content of the current page.
func (m *Model) forward(msg tea.Msg) tea.Cmd {
	menu, c1 := m.menu().Update(msg)
	m.setMenu(menu)
	content, c2 := m.content().Update(msg)
	m.setContent(content)
	return tea.Batch(c1, c2)
}
