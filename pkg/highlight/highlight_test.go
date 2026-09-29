package highlight_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/highlight"
)

const sample = "name: demo\nitems:\n  - one\n  - 2\n# comment\n"

func TestYAMLKeepsTextAndAddsStyling(t *testing.T) {
	got, err := highlight.YAML(sample)
	if err != nil {
		t.Fatal(err)
	}
	if ansi.Strip(got) != sample {
		t.Fatalf("stripped output differs:\n%q", ansi.Strip(got))
	}
	if !strings.Contains(got, "\x1b[") {
		t.Fatal("expected SGR sequences")
	}
}

func TestLanguage(t *testing.T) {
	got, err := highlight.Language(`{"a": [1, true]}`, "json")
	if err != nil || ansi.Strip(got) != `{"a": [1, true]}` || !strings.Contains(got, "\x1b[") {
		t.Fatalf("json: %q %v", got, err)
	}
	plain, err := highlight.Language("just words", "no-such-language")
	if err != nil || ansi.Strip(plain) != "just words" {
		t.Fatalf("unknown language: %q %v", plain, err)
	}
}

func TestColorizeNilLexerAndStyles(t *testing.T) {
	got, err := highlight.Colorize("x: 1", "no-such-style", nil)
	if err != nil || ansi.Strip(got) != "x: 1" {
		t.Fatalf("got %q %v", got, err)
	}
	for _, style := range []string{"github", "monokai", "dracula"} {
		out, err := highlight.Colorize(sample, style, lexers.Get("yaml"))
		if err != nil || ansi.Strip(out) != sample {
			t.Fatalf("%s: %v", style, err)
		}
	}
}

func TestColorizeStyleAttributes(t *testing.T) {
	// The colorful style makes keywords bold.
	out, err := highlight.Colorize("SELECT 1\n", "colorful", lexers.Get("sql"))
	if err != nil || !strings.Contains(out, "\x1b[1;") {
		t.Fatalf("expected bold keyword: %q %v", out, err)
	}
	// The pygments style leaves some tokens unstyled and italicises comments.
	out, err = highlight.Colorize("a: 1\n# c\n", "pygments", lexers.Get("yaml"))
	if err != nil || ansi.Strip(out) != "a: 1\n# c\n" || !strings.Contains(out, "\x1b[3;") {
		t.Fatalf("expected italic comment: %q %v", out, err)
	}
	// Underlined tokens are styled too.
	out, err = highlight.Colorize("http://example.com", "monokai", lexers.Get("markdown"))
	if err != nil || ansi.Strip(out) != "http://example.com" {
		t.Fatalf("markdown: %q %v", out, err)
	}
}

type failingLexer struct{ chroma.Lexer }

func (failingLexer) Tokenise(*chroma.TokeniseOptions, string) (chroma.Iterator, error) {
	return nil, errors.New("lexer failed")
}

func TestColorizeReportsLexerError(t *testing.T) {
	if _, err := highlight.Colorize("x", highlight.DefaultStyle, failingLexer{}); err == nil {
		t.Fatal("lexer errors are returned")
	}
}
