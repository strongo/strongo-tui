package themes

import (
	"testing"

	catppuccin "github.com/catppuccin/go"
	"github.com/gdamore/tcell/v2"
)

func TestCatppuccinColor(t *testing.T) {
	c := catppuccin.Mocha.Lavender()
	got := CatppuccinColor(c)

	// Verify it's an RGB color
	if got&tcell.ColorIsRGB == 0 {
		t.Error("expected RGB color flag to be set")
	}

	// Verify the RGB values match
	r, g, b := got.RGB()
	if int32(r) != int32(c.RGB[0]) || int32(g) != int32(c.RGB[1]) || int32(b) != int32(c.RGB[2]) {
		t.Errorf("RGB mismatch: got (%d,%d,%d), want (%d,%d,%d)",
			r, g, b, c.RGB[0], c.RGB[1], c.RGB[2])
	}
}

func TestCatppuccinTheme(t *testing.T) {
	theme := CatppuccinTheme(catppuccin.Mocha)

	if _, ok := theme.(catppuccinTheme); !ok {
		t.Fatal("CatppuccinTheme should return a catppuccinTheme")
	}
}

func TestCatppuccinThemeMethods(t *testing.T) {
	flavor := catppuccin.Mocha
	th := CatppuccinTheme(flavor)

	colorTests := []struct {
		name string
		got  tcell.Color
		want tcell.Color
	}{
		{"FocusedBorderColor", th.FocusedBorderColor(), CatppuccinColor(flavor.Lavender())},
		{"FocusedGraphicsColor", th.FocusedGraphicsColor(), CatppuccinColor(flavor.Text())},
		{"BlurredBorderColor", th.BlurredBorderColor(), CatppuccinColor(flavor.Surface1())},
		{"BlurredGraphicsColor", th.BlurredGraphicsColor(), CatppuccinColor(flavor.Overlay0())},
		{"LabelColor", th.LabelColor(), CatppuccinColor(flavor.Blue())},
		{"TableHeaderColor", th.TableHeaderColor(), CatppuccinColor(flavor.Mauve())},
		{"HotkeyColor", th.HotkeyColor(), CatppuccinColor(flavor.Yellow())},
	}
	for _, tt := range colorTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	t.Run("FocusedSelectedTextStyle", func(t *testing.T) {
		want := tcell.StyleDefault.
			Background(CatppuccinColor(flavor.Surface1())).
			Foreground(CatppuccinColor(flavor.Text()))
		if got := th.FocusedSelectedTextStyle(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("BlurredSelectedTextStyle", func(t *testing.T) {
		want := tcell.StyleDefault.
			Background(CatppuccinColor(flavor.Surface0())).
			Foreground(CatppuccinColor(flavor.Subtext0()))
		if got := th.BlurredSelectedTextStyle(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestCatppuccinThemeAllFlavors(t *testing.T) {
	flavors := []struct {
		name   string
		flavor catppuccin.Flavor
	}{
		{"Latte", catppuccin.Latte},
		{"Frappe", catppuccin.Frappe},
		{"Macchiato", catppuccin.Macchiato},
		{"Mocha", catppuccin.Mocha},
	}
	for _, f := range flavors {
		t.Run(f.name, func(t *testing.T) {
			th := CatppuccinTheme(f.flavor)
			// Verify all methods return non-zero colors
			if th.FocusedBorderColor() == 0 {
				t.Error("FocusedBorderColor returned zero")
			}
			if th.BlurredBorderColor() == 0 {
				t.Error("BlurredBorderColor returned zero")
			}
			if th.LabelColor() == 0 {
				t.Error("LabelColor returned zero")
			}
		})
	}
}
