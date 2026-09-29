package uitest_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/strongo/strongo-tui/pkg/uitest"
)

func TestKeyRoundTripsThroughString(t *testing.T) {
	for _, name := range []string{
		"a", "Z", "1", "up", "down", "left", "right", "home", "end", "pgup", "pgdown",
		"enter", "tab", "esc", "backspace", "delete", "space", "f1", "f12",
		"shift+tab", "ctrl+q", "ctrl+c", "alt+down", "ctrl+alt+x", "alt+c",
	} {
		if got := uitest.Key(name).String(); got != name {
			t.Errorf("Key(%q).String() = %q", name, got)
		}
	}
}

func TestKeyAliasesAndText(t *testing.T) {
	if uitest.Key("backtab").String() != "shift+tab" || uitest.Key("escape").String() != "esc" {
		t.Fatal("aliases must resolve to canonical names")
	}
	if got := uitest.Key("a").Text; got != "a" {
		t.Fatalf("text of a = %q", got)
	}
	if got := uitest.Key("space").Text; got != " " {
		t.Fatalf("text of space = %q", got)
	}
	if got := uitest.Key("ctrl+q").Text; got != "" {
		t.Fatalf("modified keys carry no text, got %q", got)
	}
	if got := uitest.Key("shift+a"); got.Mod != tea.ModShift || got.Text != "a" {
		t.Fatalf("shift+a = %+v", got)
	}
	if got := uitest.Text("é"); got.Text != "é" || got.String() != "é" {
		t.Fatalf("Text = %+v", got)
	}
}

func TestKeyUnknownPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("unknown key must panic")
		}
	}()
	uitest.Key("nonsense")
}

func TestKeyTrailingPlusIsARune(t *testing.T) {
	if got := uitest.Key("+").String(); got != "+" {
		t.Fatalf("got %q", got)
	}
	if got := uitest.Key("ctrl++").String(); got != "ctrl++" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainAndLines(t *testing.T) {
	view := "\x1b[31mred\x1b[0m   \nplain  "
	if got := uitest.Plain(view); got != "red\nplain" {
		t.Fatalf("Plain = %q", got)
	}
	if got := uitest.Lines(view); len(got) != 2 || got[1] != "plain" {
		t.Fatalf("Lines = %q", got)
	}
}

func TestMsgs(t *testing.T) {
	msg := func(v string) tea.Cmd { return func() tea.Msg { return v } }
	if got := uitest.Msgs(nil); got != nil {
		t.Fatalf("nil command: %v", got)
	}
	if got := uitest.Msgs(func() tea.Msg { return nil }); got != nil {
		t.Fatalf("no message: %v", got)
	}
	if got := uitest.Msgs(msg("a")); len(got) != 1 || got[0] != "a" {
		t.Fatalf("single: %v", got)
	}
	got := uitest.Msgs(tea.Batch(msg("a"), tea.Sequence(msg("b"), msg("c")), nil))
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("batch and sequence: %v", got)
	}
}
