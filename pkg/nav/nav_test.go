package nav_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/strongo/strongo-tui/pkg/nav"
	"github.com/strongo/strongo-tui/pkg/nav/navtest"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

// probe is a Screen that records what it receives. It implements every
// optional interface so tests can steer the shell.
type probe struct {
	label      string
	inited     int
	keys       []string
	msgs       []tea.Msg
	mouse      [][2]int
	w, h       int
	focused    bool
	edges      map[widgets.Direction]bool
	noBoundary bool
	editing    bool
	borderless bool
	captures   map[string]bool
	hints      []key.Binding
	initCmd    tea.Cmd
}

func newProbe(label string) *probe {
	all := map[widgets.Direction]bool{widgets.Up: true, widgets.Down: true, widgets.Left: true, widgets.Right: true}
	return &probe{label: label, edges: all, captures: map[string]bool{}}
}

func (p *probe) Init() tea.Cmd {
	p.inited++
	return p.initCmd
}

func (p *probe) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	q := *p
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		q.keys = append(append([]string(nil), p.keys...), m.String())
	case tea.WindowSizeMsg:
		q.w, q.h = m.Width, m.Height
	case nav.ScreenFocusMsg:
		q.focused = m.Focused
	case tea.MouseMsg:
		mouse := m.Mouse()
		q.mouse = append(append([][2]int(nil), p.mouse...), [2]int{mouse.X, mouse.Y})
	default:
		q.msgs = append(append([]tea.Msg(nil), p.msgs...), msg)
	}
	*p = q
	return p, nil
}

func (p *probe) View() string {
	title := p.label
	if p.focused {
		title += "*"
	}
	return fmt.Sprintf("%s %dx%d", title, p.w, p.h)
}

func (p *probe) Title() string    { return p.label }
func (p *probe) Borderless() bool { return p.borderless }
func (p *probe) Editing() bool    { return p.editing }
func (p *probe) CapturesKey(msg tea.KeyPressMsg) bool {
	return p.captures[msg.String()]
}
func (p *probe) ShortHelp() []key.Binding  { return p.hints }
func (p *probe) FullHelp() [][]key.Binding { return [][]key.Binding{p.hints} }

// boundaryProbe adds widgets.Boundary to probe.
type boundaryProbe struct{ *probe }

func (b boundaryProbe) AtEdge(dir widgets.Direction) bool { return b.edges[dir] }

func (b boundaryProbe) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	_, cmd := b.probe.Update(msg)
	return b, cmd
}

// scr picks the Screen a test hands to the shell: with or without Boundary.
func scr(p *probe) nav.Screen {
	if p.noBoundary {
		return p
	}
	return boundaryProbe{p}
}

func recorded(s nav.Screen) *probe {
	if b, ok := s.(boundaryProbe); ok {
		return b.probe
	}
	return s.(*probe)
}

type fixture struct {
	h       *navtest.Harness
	menu    *probe
	content *probe
}

func newFixture(t *testing.T, options ...navtest.Option) *fixture {
	t.Helper()
	menu, content := newProbe("menu"), newProbe("content")
	menu.edges = map[widgets.Direction]bool{widgets.Up: true, widgets.Right: true}
	content.edges = map[widgets.Direction]bool{widgets.Up: true, widgets.Left: true}
	root := nav.Page{Title: "Home", Menu: scr(menu), Content: scr(content), Focus: nav.FocusToMenu}
	f := &fixture{h: navtest.New(t, root, options...), menu: menu, content: content}
	return f
}

// current returns the recorded state of the screens the shell holds now.
func (f *fixture) current() (menu, content *probe) {
	return recorded(f.h.Model().Menu()), recorded(f.h.Model().Content())
}

func TestDefaultLayout(t *testing.T) {
	h := navtest.New(t, nav.Page{Title: "Home"})
	lines := h.Lines()
	if len(lines) != 40 {
		t.Fatalf("screen has %d rows", len(lines))
	}
	if !strings.HasPrefix(lines[0], "   Home") || !strings.HasSuffix(lines[0], "(l) Login") {
		t.Fatalf("header = %q", lines[0])
	}
	if !strings.Contains(lines[39], "ctrl+q quit  f1 help") {
		t.Fatalf("actions bar = %q", lines[39])
	}
	if !strings.Contains(lines[1], "Menu") || !strings.Contains(lines[1], "Content") {
		t.Fatalf("default panels are titled:\n%s", h.View())
	}
	menuEdge := len([]rune(lines[1][:strings.Index(lines[1], "╮")]))
	if menuEdge != nav.MenuWidth-1 {
		t.Fatalf("menu is %d columns wide, want %d", menuEdge+1, nav.MenuWidth)
	}
	if w, ht := h.Size(); w != 120 || ht != 40 {
		t.Fatalf("size = %dx%d", w, ht)
	}
	if w, ht := h.Model().Size(); w != 120 || ht != 40 {
		t.Fatalf("model size = %dx%d", w, ht)
	}
}

func TestScreensAreInitialisedSizedAndFocused(t *testing.T) {
	menu, content := newProbe("menu"), newProbe("content")
	menu.initCmd = func() tea.Msg { return "menu-ready" }
	content.initCmd = func() tea.Msg { return "content-ready" }
	h := navtest.New(t, nav.Page{Title: "Home", Menu: scr(menu), Content: scr(content), Focus: nav.FocusToMenu})
	m, c := recorded(h.Model().Menu()), recorded(h.Model().Content())
	if m.inited != 1 || c.inited != 1 {
		t.Fatalf("Init runs once: menu=%d content=%d", m.inited, c.inited)
	}
	if m.w != 28 || m.h != 36 || c.w != 88 || c.h != 36 {
		t.Fatalf("sizes are the areas inside the frames: menu %dx%d content %dx%d", m.w, m.h, c.w, c.h)
	}
	if !m.focused || c.focused {
		t.Fatal("the focused zone is told, the other is not")
	}
	if len(m.msgs) == 0 || len(c.msgs) == 0 {
		t.Fatal("messages of Init commands reach the screens")
	}
	h.RequireContains("menu* 28x36")
	h.RequireContains("content 88x36")
	if h.Model().AlertOpen() || h.Model().Depth() != 1 {
		t.Fatal("one page, no alert")
	}
}

func TestBorderlessScreensGetTheWholeArea(t *testing.T) {
	menu, content := newProbe("menu"), newProbe("content")
	content.borderless = true
	h := navtest.New(t, nav.Page{Title: "Home", Menu: scr(menu), Content: scr(content)})
	c := recorded(h.Model().Content())
	if c.w != 90 || c.h != 38 {
		t.Fatalf("content = %dx%d", c.w, c.h)
	}
	if !strings.Contains(h.Line(1), "content* 90x38") {
		t.Fatalf("no border, no title row:\n%s", h.View())
	}
}

func TestNarrowTerminalHidesMenu(t *testing.T) {
	f := newFixture(t)
	f.h.Resize(60, 20)
	f.h.RequireNotContains("menu")
	f.h.RequireContains("content*")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("a hidden menu gives focus to the content")
	}
	f.h.Send(nav.SetFocus(nav.FocusToMenu)())
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("a hidden menu cannot take focus")
	}
	f.h.Press("left")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("Left has no menu to go to")
	}
	f.h.Resize(120, 40)
	f.h.RequireContains("menu")
}

func TestVeryShortAndDegenerateTerminals(t *testing.T) {
	f := newFixture(t)
	f.h.Resize(120, 2)
	if lines := f.h.Lines(); len(lines) != 2 || !strings.Contains(lines[1], "quit") {
		t.Fatalf("2 rows:\n%s", f.h.View())
	}
	f.h.Resize(120, 1)
	if lines := f.h.Lines(); len(lines) != 1 || !strings.Contains(lines[0], "Home") {
		t.Fatalf("1 row:\n%s", f.h.View())
	}
	f.h.Resize(0, 0)
	if f.h.View() != "" {
		t.Fatal("nothing to draw")
	}
	f.h.Resize(4, 5)
	if lines := f.h.Lines(); len(lines) != 5 {
		t.Fatalf("a tiny terminal still renders %d rows", len(lines))
	}
}

func TestRenderedFrameRequestsAltScreenAndMouse(t *testing.T) {
	h := navtest.New(t, nav.Page{Title: "Home"})
	v := h.Model().View()
	if !v.AltScreen || v.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("view = %+v", v)
	}
	if v.Content != h.Styled() {
		t.Fatal("View renders the shell")
	}
}

func TestFocusMovesBetweenZones(t *testing.T) {
	f := newFixture(t)
	step := func(key string, want nav.FocusTo) {
		t.Helper()
		f.h.Press(key)
		if got := f.h.Model().Zone(); got != want {
			t.Fatalf("after %s: zone=%v want %v", key, got, want)
		}
	}
	step("right", nav.FocusToContent)
	step("left", nav.FocusToMenu)
	step("left", nav.FocusToMenu)
	step("up", nav.FocusToBreadcrumbs)
	step("down", nav.FocusToMenu)
	step("right", nav.FocusToContent)
	step("right", nav.FocusToContent)
	step("up", nav.FocusToBreadcrumbs)
	step("down", nav.FocusToContent)
	step("shift+tab", nav.FocusToBreadcrumbs)
	step("tab", nav.FocusToContent)
	m, c := f.current()
	if strings.Join(m.keys, ",") != "left" || strings.Join(c.keys, ",") != "right" {
		t.Fatalf("only keys that did not move focus reach the screens: menu %v content %v", m.keys, c.keys)
	}
}

func TestScreensAreToldWhereFocusIs(t *testing.T) {
	f := newFixture(t)
	m, c := f.current()
	if !m.focused || c.focused {
		t.Fatal("initial focus")
	}
	f.h.Press("right")
	m, c = f.current()
	if m.focused || !c.focused {
		t.Fatal("focus moved to the content")
	}
	f.h.Press("up")
	m, c = f.current()
	if m.focused || c.focused {
		t.Fatal("focus is in the header")
	}
}

func TestBoundaryDecidesWhetherArrowKeysLeave(t *testing.T) {
	f := newFixture(t)
	_, c := f.current()
	c.edges[widgets.Up] = false
	f.h.Press("right", "up")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("Up stays in a screen that is not at its top edge")
	}
	if _, c = f.current(); strings.Join(c.keys, ",") != "up" {
		t.Fatalf("the screen got the key: %v", c.keys)
	}
	c.edges[widgets.Left] = false
	f.h.Press("left")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("Left stays in a screen that is not at its left edge")
	}
	f.h.Press("shift+tab")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("Shift+Tab follows the Up edge as well")
	}
	m, _ := f.current()
	m.edges[widgets.Right] = false
	f.h.Send(nav.SetFocus(nav.FocusToMenu)())
	f.h.Press("right")
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("Right stays in a menu that is not at its right edge")
	}
}

func TestScreensWithoutBoundaryKeepArrowKeysButShiftTabLeaves(t *testing.T) {
	menu, content := newProbe("menu"), newProbe("content")
	menu.noBoundary, content.noBoundary = true, true
	h := navtest.New(t, nav.Page{Title: "Home", Menu: scr(menu), Content: scr(content)})
	h.Press("up", "left", "right")
	if got := recorded(h.Model().Content()).keys; strings.Join(got, ",") != "up,left,right" {
		t.Fatalf("arrow keys go to the screen: %v", got)
	}
	h.Press("shift+tab")
	if h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("Shift+Tab always reaches the breadcrumbs")
	}
}

func TestHeaderFocusChain(t *testing.T) {
	f := newFixture(t)
	f.h.Press("up", "right")
	if f.h.Model().Zone() != nav.FocusToLogin {
		t.Fatal("Right past the last crumb goes to Login")
	}
	f.h.Press("enter")
	if _, c := f.current(); len(c.msgs) == 0 || c.msgs[len(c.msgs)-1] != (nav.LoginMsg{}) {
		t.Fatal("Enter on Login sends LoginMsg")
	}
	f.h.Press("x")
	f.h.Press("left")
	if f.h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("Left goes back to the crumbs")
	}
	f.h.Press("right", "down")
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("Down from Login returns to where focus came from")
	}
	f.h.Press("right", "up", "right", "tab")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("Tab leaves the header too")
	}
}

func TestSetFocusCommand(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.SetFocus(nav.FocusToLogin)())
	if f.h.Model().Zone() != nav.FocusToLogin {
		t.Fatal("login")
	}
	f.h.Send(nav.SetFocus(nav.FocusToKeep)())
	if f.h.Model().Zone() != nav.FocusToLogin {
		t.Fatal("keep")
	}
	f.h.Send(nav.SetFocus(nav.FocusToBreadcrumbs)())
	f.h.Press("down")
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("the header returns to the zone it was entered from")
	}
	f.h.Send(nav.SetFocus(nav.FocusToContent)())
	f.h.Send(nav.SetFocus(nav.FocusToBreadcrumbs)())
	f.h.Press("down")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("from content")
	}
}

func TestLoginOptions(t *testing.T) {
	h := navtest.New(t, nav.Page{Title: "Home"}, navtest.WithNav(nav.WithLogin("Sign in", 's')))
	h.RequireContains("(s) Sign in")
	none := navtest.New(t, nav.Page{Title: "Home"}, navtest.WithNav(nav.WithoutLogin()))
	none.RequireNotContains("Login")
	none.Press("up", "right")
	if none.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("without a login button Right stays on the crumbs")
	}
	none.Click(110, 0)
	if none.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("no button to click")
	}
}

func TestPushPopAndBreadcrumbs(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Push(nav.Page{Title: "Projects", Content: scr(newProbe("projects"))})())
	if f.h.Model().Depth() != 2 {
		t.Fatal("depth after push")
	}
	f.h.RequireContains("Home > Projects")
	f.h.RequireContains("projects*")
	f.h.RequireContains("menu")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("a pushed page focuses its content by default")
	}
	if got := f.h.Model().Breadcrumbs(); len(got) != 2 || got[1].Title != "Projects" {
		t.Fatalf("crumbs = %v", got)
	}
	f.h.Send(nav.Push(nav.Page{Title: "Demo", Content: scr(newProbe("demo"))})())
	f.h.RequireContains("Home > Projects > Demo")
	f.h.Send(nav.Pop()())
	f.h.RequireContains("projects*")
	f.h.RequireNotContains("Demo")
	f.h.Send(nav.Pop()())
	f.h.Send(nav.Pop()())
	if f.h.Model().Depth() != 1 {
		t.Fatal("the root page stays")
	}
	f.h.RequireContains("menu*")
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("popping back focuses where the revealed page wants it")
	}
	if m, _ := f.current(); m != f.menu {
		t.Fatal("popping back reveals the same menu")
	}
}

func TestPushWithOwnMenuAndPop(t *testing.T) {
	f := newFixture(t)
	other := newProbe("other-menu")
	f.h.Send(nav.Push(nav.Page{Title: "Deep", Menu: scr(other), Content: scr(newProbe("deep")), Focus: nav.FocusToMenu})())
	f.h.RequireContains("other-menu*")
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("page focus")
	}
	if f.menu.focused {
		t.Fatal("the menu that went under is blurred")
	}
	f.h.Send(nav.Pop()())
	f.h.RequireContains("menu")
	f.h.RequireNotContains("other-menu")
	f.h.Send(nav.Push(nav.Page{Title: "Keep", Content: scr(newProbe("keep")), Focus: nav.FocusToKeep})())
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("FocusToKeep leaves focus in the menu")
	}
}

func TestPopToReplaceAndReset(t *testing.T) {
	f := newFixture(t)
	push := func(title string) {
		f.h.Send(nav.Push(nav.Page{Title: title, Content: scr(newProbe(title))})())
	}
	push("a")
	push("b")
	push("c")
	f.h.Send(nav.PopTo(2)())
	f.h.RequireContains("Home > a")
	f.h.RequireNotContains("Home > a > b")
	f.h.Send(nav.PopTo(5)())
	if f.h.Model().Depth() != 2 {
		t.Fatal("PopTo deeper than the stack does nothing")
	}
	f.h.Send(nav.PopTo(0)())
	if f.h.Model().Depth() != 1 {
		t.Fatal("PopTo clamps to the root")
	}
	push("x")
	f.h.Send(nav.Replace(nav.Page{Title: "y", Content: scr(newProbe("y"))})())
	f.h.RequireContains("Home > y")
	f.h.RequireNotContains("Home > x")
	f.h.Send(nav.Pop()())
	f.h.Send(nav.Replace(nav.Page{Title: "root2", Content: scr(newProbe("root2"))})())
	f.h.RequireContains("root2")
	if f.h.Model().Depth() != 1 {
		t.Fatal("replacing the root keeps depth 1")
	}
	push("z")
	f.h.Send(nav.Reset(nav.Page{Title: "fresh", Content: scr(newProbe("fresh")), Focus: nav.FocusToMenu})())
	if f.h.Model().Depth() != 1 || f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("Reset shows a single page")
	}
	f.h.RequireContains("fresh")
	f.h.RequireContains("menu")
	f.h.Send(nav.Reset(nav.Page{Title: "own", Menu: scr(newProbe("m2")), Content: scr(newProbe("c2"))})())
	f.h.RequireContains("m2")
}

func TestSetPanelsKeepsCrumbsAndNilScreens(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.SetPanels(nil, scr(newProbe("next")), nav.FocusToContent)())
	f.h.RequireContains("next*")
	f.h.RequireContains("menu")
	f.h.Send(nav.SetPanels(scr(newProbe("menu2")), nil, nav.FocusToKeep)())
	f.h.RequireContains("menu2 ")
	f.h.RequireContains("next*")
	if f.h.Model().Depth() != 1 || len(f.h.Model().Breadcrumbs()) != 1 {
		t.Fatal("SetPanels does not touch the stack")
	}
	f.h.Send(nav.SetPanels(nil, nil, nav.FocusToContent)())
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("focus target")
	}
}

func TestCrumbActivationPopsOrRunsOnCrumb(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Push(nav.Page{Title: "a", Content: scr(newProbe("a"))})())
	f.h.Send(nav.Push(nav.Page{Title: "b", Content: scr(newProbe("b"))})())
	f.h.Press("up", "enter") // the parent crumb, "a"
	if f.h.Model().Depth() != 2 || f.h.Model().Zone() != nav.FocusToContent {
		t.Fatalf("activating the parent crumb pops to it: depth=%d zone=%v", f.h.Model().Depth(), f.h.Model().Zone())
	}
	f.h.Press("up", "enter")
	f.h.Press("up", "left", "enter")
	if f.h.Model().Depth() != 1 {
		t.Fatal("activating the root crumb pops to the root")
	}

	custom := nav.Page{Title: "custom", Content: scr(newProbe("custom")), OnCrumb: func() tea.Msg { return "custom-crumb" }}
	f.h.Send(nav.Push(custom)())
	f.h.Send(nav.Push(nav.Page{Title: "below", Content: scr(newProbe("below"))})())
	f.h.Press("up", "enter")
	if _, c := f.current(); c.msgs[len(c.msgs)-1] != "custom-crumb" {
		t.Fatal("OnCrumb replaces the pop")
	}
	if f.h.Model().Depth() != 3 {
		t.Fatal("no pop happened")
	}
	f.h.Send(widgets.CrumbSelectedMsg{ID: "someone.else", Index: 0})
	f.h.Send(widgets.CrumbSelectedMsg{ID: "nav.crumbs", Index: 99})
	f.h.Send(widgets.CrumbSelectedMsg{ID: "nav.crumbs", Index: 2})
	if f.h.Model().Depth() != 3 {
		t.Fatal("foreign, out of range, or last-crumb selections do nothing")
	}
}

func TestSetBreadcrumbsOverridesUntilNextNavigation(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.SetBreadcrumbs(widgets.Crumb{Title: "Alpha"}, widgets.Crumb{Title: "Beta"})())
	f.h.RequireContains("Alpha > Beta")
	f.h.Press("up", "enter")
	if f.h.Model().Depth() != 1 {
		t.Fatal("the override does not change the stack")
	}
	f.h.Send(nav.Push(nav.Page{Title: "P", Content: scr(newProbe("p"))})())
	f.h.RequireContains("Home > P")
	f.h.Send(nav.SetBreadcrumbs(widgets.Crumb{Title: "X"})())
	f.h.Send(nav.Pop()())
	f.h.RequireNotContains("X")
	f.h.Send(nav.SetBreadcrumbs(widgets.Crumb{Title: "Y"})())
	f.h.Send(nav.Reset(nav.Page{Title: "R", Content: scr(newProbe("r"))})())
	f.h.RequireNotContains("Y")
}

func TestKeysReachTheFocusedScreen(t *testing.T) {
	f := newFixture(t)
	f.h.Press("x", "down")
	m, c := f.current()
	if strings.Join(m.keys, ",") != "x,down" || len(c.keys) != 0 {
		t.Fatalf("menu keys = %v content keys = %v", m.keys, c.keys)
	}
}

func TestKeyCapturersClaimShellKeys(t *testing.T) {
	f := newFixture(t)
	m, _ := f.current()
	m.captures["ctrl+c"] = true
	f.h.Press("ctrl+c")
	m, _ = f.current()
	if f.h.Quit() || strings.Join(m.keys, ",") != "ctrl+c" {
		t.Fatal("a screen that claims Ctrl+C keeps it")
	}
	m.captures["ctrl+c"] = false
	f.h.Press("ctrl+c")
	if !f.h.Quit() {
		t.Fatal("unclaimed Ctrl+C quits")
	}
}

func TestCtrlQAlwaysQuits(t *testing.T) {
	f := newFixture(t)
	m, _ := f.current()
	m.editing = true
	m.captures["ctrl+q"] = true
	f.h.Press("ctrl+q")
	if !f.h.Quit() || len(m.keys) != 0 {
		t.Fatal("Ctrl+Q quits before any screen sees it")
	}
}

func TestActionsBindKeysAndSuspendWhileEditing(t *testing.T) {
	type savedMsg struct{}
	action := nav.Action{ID: "Save", Binding: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")), Msg: savedMsg{}}
	f := newFixture(t, navtest.WithNav(nav.WithActions(action, nav.Action{
		ID: "Hidden", Binding: key.NewBinding(key.WithKeys("ctrl+h")), Msg: "hidden-fired", Hidden: true,
	}, nav.Action{
		ID: "Off", Binding: key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "off"), key.WithDisabled()), Msg: "off-fired",
	})))
	f.h.RequireContains("ctrl+s save")
	f.h.RequireNotContains("ctrl+h")
	f.h.RequireNotContains("ctrl+o")
	f.h.Press("ctrl+s", "ctrl+h")
	m, _ := f.current()
	if len(m.msgs) < 2 || m.msgs[len(m.msgs)-2] != any(savedMsg{}) || m.msgs[len(m.msgs)-1] != "hidden-fired" || len(m.keys) != 0 {
		t.Fatalf("actions send their messages: %v %v", m.msgs, m.keys)
	}
	m.editing = true
	f.h.Press("ctrl+s")
	m, _ = f.current()
	if strings.Join(m.keys, ",") != "ctrl+s" {
		t.Fatal("actions step aside while the screen edits text")
	}
	f.h.Press("f1")
	m.editing = false
	f.h.Press("f1")
	if m, _ = f.current(); m.msgs[len(m.msgs)-1] != any(nav.HelpMsg{}) {
		t.Fatal("F1 sends HelpMsg")
	}
}

func TestActionsBarShowsScreenHelp(t *testing.T) {
	menu := newProbe("menu")
	menu.hints = []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "gone"), key.WithDisabled()),
	}
	h := navtest.New(t, nav.Page{Title: "Home", Menu: scr(menu), Focus: nav.FocusToMenu})
	h.RequireContains("enter open")
	h.RequireNotContains("gone")
	h.Press("right")
	h.RequireNotContains("enter open")
}

func TestActionRegistrationRules(t *testing.T) {
	mustPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatalf("%s must panic", name)
			}
		}()
		fn()
	}
	mustPanic("empty ID", func() {
		nav.New(nav.Page{}, nav.WithActions(nav.Action{Binding: key.NewBinding(key.WithHelp("x", "no id"))}))
	})
	mustPanic("duplicate of a default", func() { nav.New(nav.Page{}, nav.WithActions(nav.Action{ID: "Quit"})) })
	mustPanic("duplicate among new", func() {
		nav.New(nav.Page{}, nav.WithActions(nav.Action{ID: "A"}, nav.Action{ID: "A"}))
	})
	mustPanic("bad SetActions", func() {
		h := navtest.New(t, nav.Page{})
		h.Send(nav.SetActions(nav.Action{ID: "Help"})())
	})
}

func TestSetActionsReplacesApplicationActions(t *testing.T) {
	a := nav.Action{ID: "A", Binding: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "alpha")), Msg: "alpha-fired"}
	b := nav.Action{ID: "B", Binding: key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "beta")), Msg: "beta-fired"}
	f := newFixture(t, navtest.WithNav(nav.WithActions(a)))
	f.h.RequireContains("alpha")
	f.h.Send(nav.SetActions(b)())
	f.h.RequireNotContains("alpha")
	f.h.RequireContains("beta")
	f.h.RequireContains("quit")
	f.h.Press("ctrl+b")
	if m, _ := f.current(); m.msgs[len(m.msgs)-1] != "beta-fired" {
		t.Fatal("the new action works")
	}
}

func TestClickingActions(t *testing.T) {
	f := newFixture(t)
	f.h.Click(3, 39)
	if !f.h.Quit() {
		t.Fatal("clicking the Quit action quits")
	}
	f.h.Click(16, 39)
	if m, _ := f.current(); m.msgs[len(m.msgs)-1] != any(nav.HelpMsg{}) {
		t.Fatal("clicking Help sends HelpMsg")
	}
	before := len(recorded(f.h.Model().Menu()).msgs)
	f.h.Click(60, 39)
	if len(recorded(f.h.Model().Menu()).msgs) != before {
		t.Fatal("clicking the empty part of the bar does nothing")
	}
}

func TestAsyncMessagesReachBothScreens(t *testing.T) {
	f := newFixture(t)
	f.h.Send("hello")
	m, c := f.current()
	if m.msgs[len(m.msgs)-1] != "hello" || c.msgs[len(c.msgs)-1] != "hello" {
		t.Fatal("custom messages reach both screens")
	}
	f.h.Send(tea.BackgroundColorMsg{})
}

func TestShowErrorReplacesContent(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.ShowError(errors.New("disk on fire"))())
	f.h.RequireContains("disk on fire")
	f.h.RequireContains("Error")
	f.h.RequireContains("menu")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("focus goes to the content")
	}
	if _, ok := f.h.Model().Content().(interface{ Title() string }); !ok {
		t.Fatal("the error screen has a title")
	}
	f.h.Press("up")
	if f.h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("a static screen never keeps an arrow key")
	}
}

func TestStaticScreen(t *testing.T) {
	h := navtest.New(t, nav.Page{Title: "Home", Content: nav.Static("Notes", "one two three four five six seven eight nine ten")}, navtest.WithSize(120, 12))
	h.RequireContains("Notes")
	h.RequireContains("one two three")
	if nav.Static("t", "x").Init() != nil {
		t.Fatal("no Init command")
	}
}

func TestQuitCommand(t *testing.T) {
	h := navtest.New(t, nav.Page{Title: "Home"})
	h.Run(nav.Quit())
	if !h.Quit() {
		t.Fatal("Quit")
	}
}

func TestMouseFocusAndCoordinates(t *testing.T) {
	f := newFixture(t)
	f.h.Click(60, 10)
	_, c := f.current()
	if f.h.Model().Zone() != nav.FocusToContent || len(c.mouse) != 1 || c.mouse[0] != [2]int{29, 8} {
		t.Fatalf("content click: zone=%v mouse=%v", f.h.Model().Zone(), c.mouse)
	}
	f.h.Click(5, 3)
	m, _ := f.current()
	if f.h.Model().Zone() != nav.FocusToMenu || len(m.mouse) != 1 || m.mouse[0] != [2]int{4, 1} {
		t.Fatalf("menu click: zone=%v mouse=%v", f.h.Model().Zone(), m.mouse)
	}
	f.h.Wheel(60, 10, false)
	f.h.Wheel(5, 10, true)
	m, c = f.current()
	if f.h.Model().Zone() != nav.FocusToMenu || len(c.mouse) != 2 || len(m.mouse) != 2 {
		t.Fatal("the wheel scrolls the screen under the pointer without moving focus")
	}
	f.h.Send(tea.MouseMotionMsg(tea.Mouse{X: 60, Y: 10}))
	f.h.Send(tea.MouseReleaseMsg(tea.Mouse{X: 60, Y: 10, Button: tea.MouseLeft}))
	f.h.Send(tea.MouseClickMsg(tea.Mouse{X: 60, Y: 10, Button: tea.MouseRight}))
	if _, c = f.current(); len(c.mouse) != 2 || f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("motion, release and right clicks are ignored")
	}
}

func TestMouseOnHeader(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Push(nav.Page{Title: "Deep", Content: scr(newProbe("deep"))})())
	f.h.Click(4, 0)
	if f.h.Model().Depth() != 1 {
		t.Fatal("clicking a crumb navigates to it")
	}
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("focus goes back to the panels after navigating")
	}
	f.h.Click(110, 0)
	if f.h.Model().Zone() != nav.FocusToLogin {
		t.Fatal("clicking Login focuses it")
	}
	if _, c := f.current(); len(c.msgs) == 0 || c.msgs[len(c.msgs)-1] != any(nav.LoginMsg{}) {
		t.Fatal("clicking Login sends LoginMsg")
	}
	f.h.Wheel(110, 0, false)
	f.h.Wheel(4, 0, false)
	f.h.Click(1, 0)
	if f.h.Model().Zone() != nav.FocusToLogin {
		t.Fatal("clicks and wheel elsewhere in the header change nothing")
	}
}

func TestAlertLifecycle(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Alert("Oops", "Something broke", 0, nav.FocusToContent)())
	if !f.h.Model().AlertOpen() {
		t.Fatal("alert is open")
	}
	f.h.RequireContains("Oops")
	f.h.RequireContains("Something broke")
	f.h.RequireContains("OK")
	f.h.Press("down", "x")
	m, c := f.current()
	if len(m.keys) != 0 || len(c.keys) != 0 {
		t.Fatal("an alert swallows keys")
	}
	f.h.Press("enter")
	if f.h.Model().AlertOpen() {
		t.Fatal("Enter dismisses")
	}
	f.h.RequireNotContains("Something broke")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("focus goes back to the requested zone")
	}
}

func TestAlertEscKeepsFocus(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Alert("T", "M", 0, nav.FocusToKeep)())
	f.h.Press("esc")
	if f.h.Model().AlertOpen() || f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("Esc dismisses and FocusToKeep leaves focus alone")
	}
}

func TestAlertAutoCloses(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Alert("Saved", "All good", 2*time.Second, nav.FocusToContent)())
	f.h.RequireContains("All good")
	f.h.Advance(time.Second)
	f.h.RequireContains("All good")
	f.h.Advance(time.Second)
	f.h.RequireNotContains("All good")
	if f.h.Model().Zone() != nav.FocusToContent {
		t.Fatal("focus returns after auto-close")
	}
}

func TestNewAlertReplacesOldOneAndStaleTimersAreIgnored(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Alert("First", "one", time.Second, nav.FocusToKeep)())
	f.h.Send(nav.Alert("Second", "two", time.Hour, nav.FocusToKeep)())
	f.h.Advance(2 * time.Second)
	f.h.RequireContains("two")
	f.h.RequireNotContains("one")
	f.h.Press("enter")
	f.h.Advance(2 * time.Hour)
	if f.h.Model().AlertOpen() {
		t.Fatal("a timer of a closed alert does nothing")
	}
	f.h.Send(widgets.ModalDoneMsg{ID: "someone.else"})
	f.h.Send(widgets.ModalDoneMsg{ID: "nav.alert"})
}

func TestAlertQuitKeys(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Alert("T", "M", 0, nav.FocusToKeep)())
	f.h.Press("ctrl+c")
	if !f.h.Quit() {
		t.Fatal("Ctrl+C quits with an alert open")
	}
	g := newFixture(t)
	g.h.Send(nav.Alert("T", "M", 0, nav.FocusToKeep)())
	g.h.Press("ctrl+q")
	if !g.h.Quit() {
		t.Fatal("Ctrl+Q quits with an alert open")
	}
}

func TestAlertMouse(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Alert("T", "M", 0, nav.FocusToKeep)())
	f.h.Click(0, 0)
	if !f.h.Model().AlertOpen() {
		t.Fatal("clicks outside the alert are swallowed")
	}
	f.h.Wheel(60, 20, false)
	var row, col int
	for y, line := range f.h.Lines() {
		if x := strings.Index(line, "OK"); x >= 0 {
			row, col = y, len([]rune(line[:x]))
		}
	}
	f.h.Click(col, row)
	if f.h.Model().AlertOpen() {
		t.Fatalf("clicking OK dismisses (clicked %d,%d)\n%s", col, row, f.h.View())
	}
}

func TestWithKeyMapRebindsShellKeys(t *testing.T) {
	keys := nav.DefaultKeyMap()
	keys.Up = key.NewBinding(key.WithKeys("k"))
	f := newFixture(t, navtest.WithNav(nav.WithKeyMap(keys)))
	f.h.Press("up")
	if f.h.Model().Zone() != nav.FocusToMenu {
		t.Fatal("Up is no longer bound")
	}
	f.h.Press("k")
	if f.h.Model().Zone() != nav.FocusToBreadcrumbs {
		t.Fatal("k is Up now")
	}
	if len(keys.ShortHelp()) != 1 || len(keys.FullHelp()) != 2 {
		t.Fatal("help")
	}
}

var _ = uitest.Key

func TestRootFocusToKeepMeansContent(t *testing.T) {
	h := navtest.New(t, nav.Page{Title: "Home", Focus: nav.FocusToKeep})
	if h.Model().Zone() != nav.FocusToContent {
		t.Fatal("there is nothing to keep at the start")
	}
}

func TestClickOnPanelBorderFocusesWithoutReachingTheScreen(t *testing.T) {
	f := newFixture(t)
	f.h.Click(60, 1) // the content's top border
	_, c := f.current()
	if f.h.Model().Zone() != nav.FocusToContent || len(c.mouse) != 0 {
		t.Fatalf("border click: zone=%v mouse=%v", f.h.Model().Zone(), c.mouse)
	}
}

func TestReplaceDropsTheFocusedScreensOfTheOldPage(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Push(nav.Page{Title: "a", Menu: scr(newProbe("menu-a")), Content: scr(newProbe("a")), Focus: nav.FocusToMenu})())
	f.h.Send(nav.Replace(nav.Page{Title: "b", Content: scr(newProbe("b"))})())
	f.h.RequireContains("Home > b")
	f.h.RequireContains("menu")
	f.h.RequireNotContains("menu-a")
	if m, _ := f.current(); m != f.menu {
		t.Fatal("the menu of the page below is back")
	}
	if f.menu.focused {
		t.Fatal("it has not been told it is focused: focus is in the content")
	}
}

func TestResetKeepsAnInheritedMenuWithoutInitialisingItAgain(t *testing.T) {
	f := newFixture(t)
	f.h.Send(nav.Reset(nav.Page{Title: "fresh", Content: scr(newProbe("fresh"))})())
	if f.menu.inited != 1 {
		t.Fatalf("the kept menu was initialised %d times", f.menu.inited)
	}
	if m, _ := f.current(); m != f.menu {
		t.Fatal("the menu is kept")
	}
}
