package nav

import (
	"errors"
	"io"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func stubProgram(t *testing.T, fn func(p *tea.Program) (tea.Model, error)) {
	t.Helper()
	previous := runTeaProgram
	runTeaProgram = fn
	t.Cleanup(func() { runTeaProgram = previous })
}

func TestRunUsesTheProgramSeam(t *testing.T) {
	want := errors.New("boom")
	stubProgram(t, func(p *tea.Program) (tea.Model, error) { return nil, want })
	if err := Run(New(Page{Title: "Home"})); !errors.Is(err, want) {
		t.Fatalf("Run = %v", err)
	}
}

func TestRunHeadlessProgramEndToEnd(t *testing.T) {
	input, release := io.Pipe()
	t.Cleanup(func() { _ = release.Close() })
	quit := New(Page{Title: "Home"})
	done := make(chan error, 1)
	go func() {
		done <- Run(quit, tea.WithInput(input), tea.WithOutput(io.Discard), tea.WithWindowSize(100, 30))
	}()
	// Ctrl+Q typed into the program's input quits it.
	go func() {
		time.Sleep(100 * time.Millisecond)
		_, _ = release.Write([]byte{0x11})
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("program did not exit")
	}
}
