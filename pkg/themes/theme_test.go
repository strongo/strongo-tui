package themes

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestThemeStructMethods(t *testing.T) {
	th := theme{
		focusedBorderColor:       tcell.ColorRed,
		focusedGraphicsColor:     tcell.ColorGreen,
		focusedSelectedTextStyle: tcell.StyleDefault.Foreground(tcell.ColorBlue),
		blurredBorderColor:       tcell.ColorYellow,
		blurredGraphicsColor:     tcell.ColorPurple,
		blurredSelectedTextStyle: tcell.StyleDefault.Foreground(tcell.ColorOrange),
		labelColor:               tcell.ColorPink,
		tableHeaderColor:         tcell.ColorTeal,
		hotkeyColor:              tcell.ColorWhite,
	}

	if got := th.FocusedBorderColor(); got != tcell.ColorRed {
		t.Errorf("FocusedBorderColor: got %v, want %v", got, tcell.ColorRed)
	}
	if got := th.FocusedGraphicsColor(); got != tcell.ColorGreen {
		t.Errorf("FocusedGraphicsColor: got %v, want %v", got, tcell.ColorGreen)
	}
	if got := th.FocusedSelectedTextStyle(); got != tcell.StyleDefault.Foreground(tcell.ColorBlue) {
		t.Errorf("FocusedSelectedTextStyle: got %v, want %v", got, tcell.StyleDefault.Foreground(tcell.ColorBlue))
	}
	if got := th.BlurredBorderColor(); got != tcell.ColorYellow {
		t.Errorf("BlurredBorderColor: got %v, want %v", got, tcell.ColorYellow)
	}
	if got := th.BlurredGraphicsColor(); got != tcell.ColorPurple {
		t.Errorf("BlurredGraphicsColor: got %v, want %v", got, tcell.ColorPurple)
	}
	if got := th.BlurredSelectedTextStyle(); got != tcell.StyleDefault.Foreground(tcell.ColorOrange) {
		t.Errorf("BlurredSelectedTextStyle: got %v, want %v", got, tcell.StyleDefault.Foreground(tcell.ColorOrange))
	}
	if got := th.LabelColor(); got != tcell.ColorPink {
		t.Errorf("LabelColor: got %v, want %v", got, tcell.ColorPink)
	}
	if got := th.TableHeaderColor(); got != tcell.ColorTeal {
		t.Errorf("TableHeaderColor: got %v, want %v", got, tcell.ColorTeal)
	}
	if got := th.HotkeyColor(); got != tcell.ColorWhite {
		t.Errorf("HotkeyColor: got %v, want %v", got, tcell.ColorWhite)
	}
}

func TestCurrentThemeIsNotNil(t *testing.T) {
	if CurrentTheme == nil {
		t.Fatal("CurrentTheme should not be nil")
	}
}

func TestDefaultBorderWithoutPadding(t *testing.T) {
	box := tview.NewBox()
	DefaultBorderWithoutPadding(box)

	if got := box.GetBorderAttributes(); got&tcell.AttrDim == 0 {
		t.Errorf("expected border attributes to include AttrDim, got %v", got)
	}

	// Trigger focus and blur callbacks to cover the closures
	box.Focus(nil)
	box.Blur()
}

func TestDefaultBorderWithPadding(t *testing.T) {
	box := tview.NewBox()
	DefaultBorderWithPadding(box)

	// Verify border is set (padding has no getter, but we verify no panic)
	if got := box.GetBorderColor(); got != CurrentTheme.BlurredBorderColor() {
		t.Errorf("border color: got %v, want %v", got, CurrentTheme.BlurredBorderColor())
	}
}

func TestSetPanelTitle(t *testing.T) {
	box := tview.NewBox()
	SetPanelTitle(box, "Test Title")

	if got := box.GetTitle(); got != "Test Title" {
		t.Errorf("title: got %q, want %q", got, "Test Title")
	}
}
