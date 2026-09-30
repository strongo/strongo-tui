package navtest_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tuigoff/tuigoff/pkg/nav"
	"github.com/tuigoff/tuigoff/pkg/nav/navtest"
)

// fakeTB records failures instead of ending the test. Fatalf unwinds with a
// panic the test recovers from, like the real one stops the goroutine.
type fakeTB struct {
	testing.TB
	failure string
}

type stopped struct{}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.failure = strings.TrimSpace(format)
	panic(stopped{})
}

// expectFailure runs fn and reports what the fake test recorded, or "" when fn
// returned normally.
func expectFailure(fake *fakeTB, fn func()) (failure string) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(stopped); !ok {
				panic(r)
			}
			failure = fake.failure
		}
	}()
	fn()
	return ""
}

// note is a Screen that records the messages it receives.
type note struct {
	title string
	got   *[]tea.Msg
	w, h  int
	typed string
}

func (n note) Init() tea.Cmd { return nil }

func (n note) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		n.typed += m.String()
	case tea.WindowSizeMsg:
		n.w, n.h = m.Width, m.Height
	}
	*n.got = append(*n.got, msg)
	return n, nil
}

func (n note) View() string  { return n.title + " " + n.typed }
func (n note) Title() string { return n.title }

func newNote(title string) (note, *[]tea.Msg) {
	got := &[]tea.Msg{}
	return note{title: title, got: got}, got
}

func newHarness(t *testing.T, options ...navtest.Option) *navtest.Harness {
	t.Helper()
	return navtest.New(t, nav.Page{Title: "Home"}, options...)
}

func TestNewSendsWindowSize(t *testing.T) {
	h := newHarness(t)
	if w, ht := h.Size(); w != navtest.DefaultWidth || ht != navtest.DefaultHeight {
		t.Fatalf("size = %dx%d", w, ht)
	}
	small := newHarness(t, navtest.WithSize(80, 24))
	if w, ht := small.Model().Size(); w != 80 || ht != 24 {
		t.Fatalf("WithSize: %dx%d", w, ht)
	}
	if len(small.Lines()) != 24 {
		t.Fatal("the screen has as many rows as the terminal")
	}
}

func TestScreenAccessors(t *testing.T) {
	h := newHarness(t)
	if h.Line(0) == "" || h.Line(-1) != "" || h.Line(1000) != "" {
		t.Fatal("Line")
	}
	if strings.Contains(h.View(), "\x1b") || !strings.Contains(h.Styled(), "\x1b") {
		t.Fatal("View is plain, Styled keeps ANSI")
	}
	if !h.Contains("Home") || h.Contains("nowhere") {
		t.Fatal("Contains")
	}
	if h.RequireContains("Home") != h || h.RequireNotContains("nowhere") != h {
		t.Fatal("assertions chain")
	}
}

func TestRequireFailures(t *testing.T) {
	fake := &fakeTB{}
	h := navtest.New(fake, nav.Page{Title: "Home"})
	if got := expectFailure(fake, func() { h.RequireContains("missing") }); !strings.Contains(got, "does not contain") {
		t.Fatalf("failure = %q", got)
	}
	if got := expectFailure(fake, func() { h.RequireNotContains("Home") }); !strings.Contains(got, "unexpectedly contains") {
		t.Fatalf("failure = %q", got)
	}
}

func TestPressTypeClickWheelResize(t *testing.T) {
	content, got := newNote("notes")
	h := navtest.New(t, nav.Page{Title: "Home", Content: content})
	h.Press("a", "enter")
	h.Type("bc")
	h.RequireContains("notes aenterb")
	h.Click(60, 5).Wheel(60, 5, false).Wheel(60, 5, true)
	if h.Resize(90, 30) != h {
		t.Fatal("Resize chains")
	}
	if w, ht := h.Size(); w != 90 || ht != 30 {
		t.Fatalf("size = %dx%d", w, ht)
	}
	var mouse int
	for _, msg := range *got {
		if _, ok := msg.(tea.MouseMsg); ok {
			mouse++
		}
	}
	if mouse != 3 {
		t.Fatalf("mouse messages reached the screen: %d", mouse)
	}
}

func TestQuitIsRecorded(t *testing.T) {
	h := newHarness(t)
	if h.Quit() {
		t.Fatal("not yet")
	}
	h.Press("ctrl+q")
	if !h.Quit() {
		t.Fatal("Ctrl+Q recorded")
	}
}

func TestRunUnpacksBatchesAndSequences(t *testing.T) {
	h := newHarness(t)
	var got []string
	record := func(name string) tea.Cmd {
		return func() tea.Msg { got = append(got, name); return nil }
	}
	h.Run(nil)
	h.Run(tea.Batch(record("a"), record("b")))
	h.Run(tea.Sequence(record("c"), record("d")))
	if strings.Join(got, "") != "abcd" {
		t.Fatalf("commands ran in order: %v", got)
	}
	h.Run(tea.Batch(tea.Quit, record("e")))
	if !h.Quit() {
		t.Fatal("a quit inside a batch is recorded")
	}
	h.Run(func() tea.Msg { return "hello" })
}

func TestBlockedCommandFailsTheTest(t *testing.T) {
	fake := &fakeTB{}
	h := navtest.New(fake, nav.Page{Title: "Home"}, navtest.WithTimeout(20*time.Millisecond))
	release := make(chan struct{})
	defer close(release)
	got := expectFailure(fake, func() { h.Run(func() tea.Msg { <-release; return nil }) })
	if !strings.Contains(got, "did not finish") {
		t.Fatalf("failure = %q", got)
	}
}

func TestAdvanceDeliversTimersInOrder(t *testing.T) {
	h := newHarness(t)
	h.Send(nav.Alert("Later", "second", 3*time.Second, nav.FocusToKeep)())
	h.Send(nav.Alert("Sooner", "first", time.Second, nav.FocusToKeep)())
	h.RequireContains("first")
	h.Advance(500 * time.Millisecond)
	h.RequireContains("first")
	h.Advance(time.Hour)
	h.RequireNotContains("first")
	if h.Advance(time.Second) != h {
		t.Fatal("Advance chains")
	}
	if h.Model().AlertOpen() {
		t.Fatal("both alerts are over")
	}
}

func TestWithNavPassesOptions(t *testing.T) {
	h := newHarness(t, navtest.WithNav(nav.WithoutLogin()))
	h.RequireNotContains("Login")
}
