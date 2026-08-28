package charm

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestDefaultThemeSemanticStyles(t *testing.T) {
	theme := DefaultTheme()

	assertColor(t, theme.FocusedPanel().GetBorderTopForeground(), "12")
	assertColor(t, theme.BlurredPanel().GetBorderTopForeground(), "8")
	assertColor(t, theme.Title().GetForeground(), "12")
	assertColor(t, theme.SelectedRow().GetForeground(), "0")
	assertColor(t, theme.SelectedRow().GetBackground(), "12")
	assertColor(t, theme.UnselectedRow().GetForeground(), "15")
	assertColor(t, theme.Muted().GetForeground(), "8")
	assertColor(t, theme.Help().GetForeground(), "14")
	assertColor(t, theme.Loading().GetForeground(), "14")
	assertColor(t, theme.Empty().GetForeground(), "8")
	assertColor(t, theme.Error().GetForeground(), "9")

	assertBorder(t, theme.FocusedPanel())
	if !theme.Title().GetBold() || !theme.SelectedRow().GetBold() || !theme.Error().GetBold() {
		t.Fatal("expected title, selected row, and error styles to be bold")
	}
	if !theme.Loading().GetItalic() || !theme.Empty().GetItalic() {
		t.Fatal("expected loading and empty styles to be italic")
	}
}

func TestNewThemeUsesPaletteWithoutChangingDefaultTheme(t *testing.T) {
	palette := DefaultPalette()
	palette.FocusedBorder = "1"
	palette.SelectedBackground = "2"
	palette.Error = "3"

	custom := NewTheme(palette)
	assertColor(t, custom.FocusedPanel().GetBorderTopForeground(), "1")
	assertColor(t, custom.SelectedRow().GetBackground(), "2")
	assertColor(t, custom.Error().GetForeground(), "3")

	defaultTheme := DefaultTheme()
	assertColor(t, defaultTheme.FocusedPanel().GetBorderTopForeground(), "12")
	assertColor(t, defaultTheme.SelectedRow().GetBackground(), "12")
	assertColor(t, defaultTheme.Error().GetForeground(), "9")
}

func assertBorder(t *testing.T, style lipgloss.Style) {
	t.Helper()
	_, top, right, bottom, left := style.GetBorder()
	if !top || !right || !bottom || !left {
		t.Fatalf("expected a complete panel border, got top=%t right=%t bottom=%t left=%t", top, right, bottom, left)
	}
}

func assertColor(t *testing.T, got color.Color, want string) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected colour %q, got nil", want)
	}
	gotRed, gotGreen, gotBlue, gotAlpha := got.RGBA()
	wantRed, wantGreen, wantBlue, wantAlpha := lipgloss.Color(want).RGBA()
	if gotRed != wantRed || gotGreen != wantGreen || gotBlue != wantBlue || gotAlpha != wantAlpha {
		t.Fatalf("expected colour %q, got RGBA(%d, %d, %d, %d)", want, gotRed, gotGreen, gotBlue, gotAlpha)
	}
}
