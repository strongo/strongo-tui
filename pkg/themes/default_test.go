package themes

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/strongo/strongo-tui/pkg/colors"
)

func TestDefaultTheme(t *testing.T) {
	dt := defaultTheme()

	colorTests := []struct {
		name string
		got  tcell.Color
		want tcell.Color
	}{
		{"FocusedBorderColor", dt.FocusedBorderColor(), colors.DefaultFocusedBorderColor},
		{"FocusedGraphicsColor", dt.FocusedGraphicsColor(), tcell.ColorWhite},
		{"BlurredBorderColor", dt.BlurredBorderColor(), colors.DefaultBlurBorderColor},
		{"BlurredGraphicsColor", dt.BlurredGraphicsColor(), tcell.ColorGray},
		{"LabelColor", dt.LabelColor(), colors.LabelColor},
		{"TableHeaderColor", dt.TableHeaderColor(), colors.TableHeaderColor},
		{"HotkeyColor", dt.HotkeyColor(), colors.HotkeyColor},
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
			Background(tcell.ColorWhite).
			Foreground(tcell.ColorBlack)
		if got := dt.FocusedSelectedTextStyle(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("BlurredSelectedTextStyle", func(t *testing.T) {
		want := tcell.StyleDefault.
			Background(tcell.ColorGray).
			Foreground(tcell.ColorWhite)
		if got := dt.BlurredSelectedTextStyle(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
