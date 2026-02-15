package button

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newScreen(w, h int) tcell.SimulationScreen {
	screen := tcell.NewSimulationScreen("")
	_ = screen.Init()
	screen.SetSize(w, h)
	return screen
}

func TestNewWithShortcut(t *testing.T) {
	b := NewWithShortcut("OK", 'o')
	if b == nil {
		t.Fatal("expected non-nil button")
	}
	if b.shortcut != 'o' {
		t.Errorf("expected shortcut 'o', got %c", b.shortcut)
	}
	if b.GetLabel() != "OK" {
		t.Errorf("expected label 'OK', got %q", b.GetLabel())
	}
	if b.Button == nil {
		t.Fatal("expected embedded Button to be non-nil")
	}
}

func TestSetShortcutStyle(t *testing.T) {
	b := NewWithShortcut("OK", 'o')
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	result := b.SetShortcutStyle(style)
	if result != b {
		t.Error("expected SetShortcutStyle to return same button for chaining")
	}
	if b.shortcutStyle != style {
		t.Error("expected shortcutStyle to be updated")
	}
}

func TestDraw_Normal(t *testing.T) {
	screen := newScreen(80, 25)
	b := NewWithShortcut("OK", 'o')
	b.SetRect(0, 0, 20, 3)

	b.Draw(screen)

	// Verify shortcut and label are drawn at the vertical center (y=1 for height=3)
	// Find "(o) OK" somewhere in row 1
	found := false
	for x := 0; x < 20; x++ {
		mainc, style, _ := screen.Get(x, 1)
		_, _ = mainc, style
		if mainc == "(" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected shortcut text to be drawn")
	}
}

func TestDraw_Focused(t *testing.T) {
	screen := newScreen(80, 25)
	b := NewWithShortcut("OK", 'o')
	b.SetRect(0, 0, 20, 3)
	b.Focus(nil)

	b.Draw(screen)

	// Verify focused styling is applied (yellow background)
	// Find the shortcut text and check its style
	for x := 0; x < 20; x++ {
		mainc, style, _ := screen.Get(x, 1)
		if mainc == "(" {
			_, bg, _ := style.Decompose()
			if bg != tcell.ColorYellow {
				t.Errorf("expected yellow background for focused shortcut, got %v", bg)
			}
			break
		}
	}
}

func TestDraw_Disabled(t *testing.T) {
	screen := newScreen(80, 25)
	b := NewWithShortcut("OK", 'o')
	b.SetRect(0, 0, 20, 3)
	b.SetDisabled(true)

	b.Draw(screen)
	// Coverage: exercises the disabled branch
}

func TestDraw_ZeroDimensions(t *testing.T) {
	screen := newScreen(80, 25)
	b := NewWithShortcut("OK", 'o')
	b.SetRect(0, 0, 0, 0)

	// Should not panic with zero dimensions
	b.Draw(screen)
}

func TestDraw_NarrowWidth(t *testing.T) {
	screen := newScreen(80, 25)
	b := NewWithShortcut("OK", 'o')
	b.SetRect(0, 0, 4, 1)

	// Narrow enough to clip both shortcut and label text
	// "(o)" = 3 chars, " OK" = 3 chars, total = 6 > width 4
	// This covers startX < x branch and clipping branches
	b.Draw(screen)
}

func TestDraw_ZeroShortcutStyle(t *testing.T) {
	screen := newScreen(80, 25)
	b := NewWithShortcut("OK", 'o')
	b.shortcutStyle = tcell.Style{}
	b.SetRect(0, 0, 20, 3)

	// Covers the shortcutStyle == (tcell.Style{}) fallback branch
	b.Draw(screen)
}
