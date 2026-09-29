package widgets_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func TestAlignIn(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		width int
		align widgets.Align
		want  string
	}{
		{"left pad", "ab", 5, widgets.AlignLeft, "ab   "},
		{"right pad", "ab", 5, widgets.AlignRight, "   ab"},
		{"center pad odd gap", "ab", 5, widgets.AlignCenter, " ab  "},
		{"truncate", "abcdef", 4, widgets.AlignLeft, "abc…"},
		{"zero width", "abc", 0, widgets.AlignLeft, ""},
		{"ansi width ignored", "\x1b[31mab\x1b[0m", 4, widgets.AlignLeft, "\x1b[31mab\x1b[0m  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := widgets.AlignIn(tt.in, tt.width, tt.align); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
	if widgets.PadRight("x", 3) != "x  " {
		t.Fatal("PadRight must pad on the right")
	}
}

func TestFitAndBlank(t *testing.T) {
	got := widgets.Fit("ab\ncdefg\nx\ny", 3, 3)
	want := "ab \ncd…\nx  "
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if widgets.Fit("x", 0, 3) != "" || widgets.Fit("x", 3, 0) != "" {
		t.Fatal("degenerate sizes render nothing")
	}
	if widgets.Blank(2, 2) != "  \n  " {
		t.Fatal("Blank must be spaces")
	}
}

func TestOverlay(t *testing.T) {
	base := "aaaaaa\nbbbbbb\ncccccc"
	got := ansi.Strip(widgets.Overlay(base, "XY\nZ", 2, 1))
	if got != "aaaaaa\nbbXYbb\nccZccc" {
		t.Fatalf("got %q", got)
	}
	// Off-screen rows are dropped and a negative x is clamped.
	got = ansi.Strip(widgets.Overlay(base, "Q\nR", -3, 2))
	if got != "aaaaaa\nbbbbbb\nQccccc" {
		t.Fatalf("got %q", got)
	}
	// Above the base is dropped too.
	got = ansi.Strip(widgets.Overlay("ab", "Q\nR", 0, -1))
	if got != "Rb" {
		t.Fatalf("got %q", got)
	}
	// A short base line is padded so the overlay lands at x.
	got = ansi.Strip(widgets.Overlay("a\nbbbbb", "Z", 3, 0))
	if got != "a  Z\nbbbbb" {
		t.Fatalf("got %q", got)
	}
}

func TestCenter(t *testing.T) {
	base := widgets.Blank(7, 5)
	got := ansi.Strip(widgets.Center(base, "ab\ncd", 7, 5))
	lines := strings.Split(got, "\n")
	if lines[1] != "  ab   " || lines[2] != "  cd   " {
		t.Fatalf("got %q", lines)
	}
	// A box larger than the screen is anchored at the top-left corner.
	got = ansi.Strip(widgets.Center("xx", "abcd", 2, 1))
	if got != "ab" {
		t.Fatalf("got %q", got)
	}
}
