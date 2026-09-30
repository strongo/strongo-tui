package highlight

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// DefaultStyle is the Chroma style used by YAML and Language.
const DefaultStyle = "dracula"

// Colorize tokenises text with lexer and returns it with each token wrapped in
// the SGR sequences of the named Chroma style. A nil lexer selects Chroma's
// plain-text fallback. An unknown style falls back to Chroma's default style.
func Colorize(text, styleName string, lexer chroma.Lexer) (string, error) {
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return "", err
	}
	style := styles.Get(styleName)

	var sb strings.Builder
	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)
		if entry.IsZero() {
			sb.WriteString(token.Value)
			continue
		}
		sb.WriteString(render(entry, token.Value))
	}
	return sb.String(), nil
}

// render styles one token. Line breaks stay outside the styled span so that
// every line can be cut or wrapped independently.
func render(entry chroma.StyleEntry, value string) string {
	style := lipgloss.NewStyle().
		Bold(entry.Bold == chroma.Yes).
		Italic(entry.Italic == chroma.Yes).
		Underline(entry.Underline == chroma.Yes)
	if entry.Colour.IsSet() {
		style = style.Foreground(lipgloss.Color(entry.Colour.String()))
	}
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = style.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

// Language highlights text as the named language (a Chroma lexer name or
// alias such as "yaml", "json" or "sql") using DefaultStyle. An unknown
// language returns the text as plain text.
func Language(text, language string) (string, error) {
	return Colorize(text, DefaultStyle, lexers.Get(language))
}

// YAML highlights YAML text using DefaultStyle.
func YAML(text string) (string, error) { return Language(text, "yaml") }
