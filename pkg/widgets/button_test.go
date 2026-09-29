package widgets_test

import (
	"testing"

	"github.com/strongo/strongo-tui/pkg/theme"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func TestButtonTextAndWidth(t *testing.T) {
	plain := widgets.Button{Label: "Cancel"}
	if plain.Text() != "Cancel" || plain.Width() != 8 {
		t.Fatalf("plain: %q %d", plain.Text(), plain.Width())
	}
	hint := widgets.Button{Label: "Login", Shortcut: 'l'}
	if hint.Text() != "(l) Login" || hint.Width() != 11 {
		t.Fatalf("shortcut: %q %d", hint.Text(), hint.Width())
	}
}

func TestButtonView(t *testing.T) {
	b := widgets.Button{Label: "Login", Shortcut: 'l'}
	if got := uitest.Plain(b.View(13)); got != "  (l) Login" {
		t.Fatalf("view = %q", got)
	}
	if got := uitest.Plain(b.View(4)); got != "(l)" {
		t.Fatalf("narrow view = %q", got)
	}
	if b.View(0) != "" {
		t.Fatal("no width, no button")
	}
	b.Focused = true
	if got := uitest.Plain(b.View(13)); got != "  (l) Login" {
		t.Fatalf("focused view = %q", got)
	}
	if b.View(13) == (widgets.Button{Label: "Login", Shortcut: 'l'}).View(13) {
		t.Fatal("focus changes the styling")
	}
	plain := widgets.Button{Label: "X", Background: theme.Blue, Foreground: theme.White}
	if got := uitest.Plain(plain.View(5)); got != "  X" {
		t.Fatalf("coloured view = %q", got)
	}
	disabled := widgets.Button{Label: "X", Disabled: true}
	if disabled.View(5) == plain.View(5) {
		t.Fatal("disabled is dimmed")
	}
}
