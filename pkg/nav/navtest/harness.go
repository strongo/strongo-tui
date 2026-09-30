package navtest

import (
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tuigoff/tuigoff/pkg/nav"
	"github.com/tuigoff/tuigoff/pkg/uitest"
)

// Default terminal size and command timeout.
const (
	DefaultWidth   = 120
	DefaultHeight  = 40
	DefaultTimeout = 5 * time.Second
)

// Option customises a Harness.
type Option func(*Harness)

// WithSize sets the terminal size the harness reports.
func WithSize(width, height int) Option {
	return func(h *Harness) { h.width, h.height = width, height }
}

// WithTimeout sets how long a single command may run before the test fails.
func WithTimeout(d time.Duration) Option {
	return func(h *Harness) { h.timeout = d }
}

// WithNav passes options to nav.New.
func WithNav(options ...nav.Option) Option {
	return func(h *Harness) { h.navOptions = append(h.navOptions, options...) }
}

type timer struct {
	due time.Duration
	fn  func(time.Time) tea.Msg
}

// Harness drives a nav.Model in a test.
type Harness struct {
	tb         testing.TB
	model      nav.Model
	width      int
	height     int
	timeout    time.Duration
	navOptions []nav.Option
	quit       bool
	now        time.Duration
	timers     []timer
}

// New creates a shell whose root page is root and starts it.
func New(tb testing.TB, root nav.Page, options ...Option) *Harness {
	tb.Helper()
	h := &Harness{tb: tb, width: DefaultWidth, height: DefaultHeight, timeout: DefaultTimeout}
	for _, option := range options {
		option(h)
	}
	h.model = nav.New(root, append(h.navOptions, nav.WithTick(h.tick))...)
	h.Run(h.model.Init())
	h.Resize(h.width, h.height)
	return h
}

// tick is the fake clock behind alert timers: the command it returns registers
// the timer when the harness executes it.
func (h *Harness) tick(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		h.timers = append(h.timers, timer{due: h.now + d, fn: fn})
		return nil
	}
}

// Model returns the shell in its current state.
func (h *Harness) Model() nav.Model { return h.model }

// Send delivers msg to the shell and runs the commands that come back.
func (h *Harness) Send(msg tea.Msg) *Harness {
	h.tb.Helper()
	m, cmd := h.model.Update(msg)
	h.model = m.(nav.Model)
	return h.Run(cmd)
}

// Run executes cmd and feeds the resulting messages to the shell. Batches and
// sequences are unpacked; tea.Quit is recorded and reported by Quit.
func (h *Harness) Run(cmd tea.Cmd) *Harness {
	h.tb.Helper()
	if cmd == nil {
		return h
	}
	msg := h.execute(cmd)
	switch m := msg.(type) {
	case nil:
	case tea.QuitMsg:
		h.quit = true
	default:
		if cmds, ok := asCmds(m); ok {
			for _, c := range cmds {
				h.Run(c)
			}
			return h
		}
		h.Send(m)
	}
	return h
}

// execute runs cmd, failing the test if it blocks longer than the timeout.
func (h *Harness) execute(cmd tea.Cmd) tea.Msg {
	h.tb.Helper()
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		return msg
	case <-time.After(h.timeout):
		h.tb.Fatalf("navtest: command did not finish within %v", h.timeout)
		return nil
	}
}

// asCmds recognises tea.BatchMsg and the unexported sequence message: any slice
// of commands.
func asCmds(msg tea.Msg) ([]tea.Cmd, bool) {
	v := reflect.ValueOf(msg)
	if v.Kind() != reflect.Slice || v.Type().Elem() != reflect.TypeFor[tea.Cmd]() {
		return nil, false
	}
	cmds := make([]tea.Cmd, v.Len())
	for i := range cmds {
		cmds[i], _ = v.Index(i).Interface().(tea.Cmd)
	}
	return cmds, true
}

// Press sends key presses named as in uitest.Key: "down", "enter", "shift+tab",
// "ctrl+q", "a".
func (h *Harness) Press(keys ...string) *Harness {
	h.tb.Helper()
	for _, key := range keys {
		h.Send(uitest.Key(key))
	}
	return h
}

// Type sends text one character at a time, like typing into a field.
func (h *Harness) Type(text string) *Harness {
	h.tb.Helper()
	for _, r := range text {
		h.Send(uitest.Text(string(r)))
	}
	return h
}

// Click sends a left click at column x, row y.
func (h *Harness) Click(x, y int) *Harness {
	h.tb.Helper()
	return h.Send(tea.MouseClickMsg(tea.Mouse{X: x, Y: y, Button: tea.MouseLeft}))
}

// Wheel sends a wheel scroll at column x, row y; up scrolls up.
func (h *Harness) Wheel(x, y int, up bool) *Harness {
	h.tb.Helper()
	button := tea.MouseWheelDown
	if up {
		button = tea.MouseWheelUp
	}
	return h.Send(tea.MouseWheelMsg(tea.Mouse{X: x, Y: y, Button: button}))
}

// Resize sends a window size message.
func (h *Harness) Resize(width, height int) *Harness {
	h.tb.Helper()
	h.width, h.height = width, height
	return h.Send(tea.WindowSizeMsg{Width: width, Height: height})
}

// Advance moves the fake clock forward by d and delivers the timers that fall
// due, in order.
func (h *Harness) Advance(d time.Duration) *Harness {
	h.tb.Helper()
	h.now += d
	for {
		next := -1
		for i, t := range h.timers {
			if t.due <= h.now && (next < 0 || t.due < h.timers[next].due) {
				next = i
			}
		}
		if next < 0 {
			return h
		}
		t := h.timers[next]
		h.timers = append(h.timers[:next], h.timers[next+1:]...)
		h.Send(t.fn(time.Time{}))
	}
}

// Quit reports whether the shell asked the program to exit.
func (h *Harness) Quit() bool { return h.quit }

// Size returns the terminal size the harness reports.
func (h *Harness) Size() (width, height int) { return h.width, h.height }

// Styled returns the screen with its ANSI styling.
func (h *Harness) Styled() string { return h.model.Render() }

// View returns the screen as plain text.
func (h *Harness) View() string { return uitest.Plain(h.model.Render()) }

// Lines returns the screen as plain text lines.
func (h *Harness) Lines() []string { return uitest.Lines(h.model.Render()) }

// Line returns row y of the screen as plain text, or "" outside the screen.
func (h *Harness) Line(y int) string {
	lines := h.Lines()
	if y < 0 || y >= len(lines) {
		return ""
	}
	return lines[y]
}

// Contains reports whether the plain-text screen contains s.
func (h *Harness) Contains(s string) bool { return strings.Contains(h.View(), s) }

// RequireContains fails the test unless the screen contains s.
func (h *Harness) RequireContains(s string) *Harness {
	h.tb.Helper()
	if !h.Contains(s) {
		h.tb.Fatalf("navtest: screen does not contain %q:\n%s", s, h.View())
	}
	return h
}

// RequireNotContains fails the test if the screen contains s.
func (h *Harness) RequireNotContains(s string) *Harness {
	h.tb.Helper()
	if h.Contains(s) {
		h.tb.Fatalf("navtest: screen unexpectedly contains %q:\n%s", s, h.View())
	}
	return h
}
