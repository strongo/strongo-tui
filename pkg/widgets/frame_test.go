package widgets_test

import (
	"strings"
	"testing"

	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func TestFrameDrawsBorderTitleAndPadding(t *testing.T) {
	f := widgets.NewFrame().WithTitle("Files").WithPadding(1, 2, 1, 2)
	got := uitest.Plain(f.Render("a\nb", 14, 7))
	want := strings.Join([]string{
		"╭── Files ───╮",
		"│            │",
		"│  a         │",
		"│  b         │",
		"│            │",
		"│            │",
		"╰────────────╯",
	}, "\n")
	if got != want {
		t.Fatalf("view:\n%s\nwant:\n%s", got, want)
	}
	if w, h := f.Inner(14, 7); w != 8 || h != 3 {
		t.Fatalf("inner = %dx%d", w, h)
	}
	if x, y := f.Origin(); x != 3 || y != 2 {
		t.Fatalf("origin = %d,%d", x, y)
	}
}

func TestFrameTitleAlignmentAndBorderless(t *testing.T) {
	left := widgets.NewFrame().WithTitle("T").WithTitleAlign(widgets.AlignLeft)
	if got := uitest.Plain(left.Render("x", 8, 3)); got != "╭─ T ──╮\n│x     │\n╰──────╯" {
		t.Fatalf("left:\n%s", got)
	}
	right := left.WithTitleAlign(widgets.AlignRight)
	if got := strings.Split(uitest.Plain(right.Render("x", 8, 3)), "\n")[0]; got != "╭── T ─╮" {
		t.Fatalf("right: %q", got)
	}
	bare := widgets.NewFrame().WithoutBorders().WithPadding(1, 0, 0, 2)
	if got := uitest.Plain(bare.Render("x\ny", 6, 3)); got != "\n  x\n  y" {
		t.Fatalf("borderless: %q", got)
	}
	if x, y := bare.Origin(); x != 2 || y != 1 {
		t.Fatalf("borderless origin = %d,%d", x, y)
	}
}

func TestFrameFocusColoursBorderOnly(t *testing.T) {
	f := widgets.NewFrame().WithTitle("x")
	blurred := f.Render("body", 12, 3)
	focused := f.WithFocus(true).Render("body", 12, 3)
	if blurred == focused {
		t.Fatal("focus must change the border colour")
	}
	if uitest.Plain(blurred) != uitest.Plain(focused) {
		t.Fatal("focus must not change the text")
	}
}

func TestFrameDegenerateSizes(t *testing.T) {
	f := widgets.NewFrame()
	if f.Render("x", 0, 3) != "" || f.Render("x", 3, 0) != "" {
		t.Fatal("degenerate")
	}
	if got := uitest.Plain(f.Render("x", 2, 2)); got != "╭╮\n╰╯" {
		t.Fatalf("border only:\n%s", got)
	}
	if got := uitest.Plain(f.Render("x", 1, 4)); got != "\n\n\n" {
		t.Fatalf("too narrow for a border: %q", got)
	}
	padded := f.WithPadding(1, 1, 1, 1)
	if got := uitest.Plain(padded.Render("x", 4, 4)); got != "╭──╮\n│  │\n│  │\n╰──╯" {
		t.Fatalf("no room for content:\n%s", got)
	}
	if w, h := padded.Inner(2, 2); w != 0 || h != 0 {
		t.Fatal("inner never goes negative")
	}
	titled := f.WithTitle("Long title")
	if got := strings.Split(uitest.Plain(titled.Render("x", 4, 3)), "\n")[0]; got != "╭──╮" {
		t.Fatalf("narrow title: %q", got)
	}
	if got := strings.Split(uitest.Plain(titled.Render("x", 9, 3)), "\n")[0]; got != "╭ Long… ╮" {
		t.Fatalf("truncated title: %q", got)
	}
}
