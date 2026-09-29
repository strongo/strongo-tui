package theme_test

import (
	"image/color"
	"testing"

	"github.com/strongo/strongo-tui/pkg/theme"
)

func rgb(c color.Color) (r, g, b uint8) {
	r16, g16, b16, _ := c.RGBA()
	return uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8)
}

func TestNamedColours(t *testing.T) {
	tests := []struct {
		name    string
		c       color.Color
		r, g, b uint8
	}{
		{"Black", theme.Black, 0, 0, 0},
		{"White", theme.White, 255, 255, 255},
		{"Gray", theme.Gray, 128, 128, 128},
		{"Grey alias", theme.Grey, 128, 128, 128},
		{"LightGrey alias", theme.LightGrey, 211, 211, 211},
		{"CornflowerBlue", theme.CornflowerBlue, 100, 149, 237},
		{"LightBlue", theme.LightBlue, 173, 216, 230},
		{"Red", theme.Red, 255, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := rgb(tt.c)
			if r != tt.r || g != tt.g || b != tt.b {
				t.Fatalf("got %d,%d,%d want %d,%d,%d", r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}

func TestSemanticColoursUseNamedHues(t *testing.T) {
	if theme.TableColumnTitle != theme.LightBlue || theme.TableTertiaryText != theme.Gray ||
		theme.TableHeaderColor != theme.WhiteSmoke || theme.TreeNodeLink != theme.Blue {
		t.Fatal("semantic colours map to their named hues")
	}
}

func TestErrorTextAndSelectedStyle(t *testing.T) {
	defer theme.SetDark(theme.Dark)
	for _, dark := range []bool{true, false} {
		theme.SetDark(dark)
		if theme.Hex(theme.ErrorColor()) == theme.Hex(theme.TextColor()) {
			t.Fatalf("dark=%v: errors must stand out from ordinary text", dark)
		}
		fbg, _ := theme.FocusSurfaceColors()
		if theme.Hex(theme.SelectedStyle(true).GetBackground()) != theme.Hex(fbg) {
			t.Fatalf("dark=%v: focused selection uses the focus surface", dark)
		}
		sbg, _ := theme.SurfaceColors()
		if theme.Hex(theme.SelectedStyle(false).GetBackground()) != theme.Hex(sbg) {
			t.Fatalf("dark=%v: blurred selection uses the neutral surface", dark)
		}
	}
	theme.SetDark(true)
	dark := theme.Hex(theme.ErrorColor())
	theme.SetDark(false)
	if theme.Hex(theme.ErrorColor()) == dark {
		t.Fatal("the error colour has a light and a dark variant")
	}
}
