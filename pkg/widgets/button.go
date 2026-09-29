package widgets

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
)

// Button is the look of a push button: a label with an optional shortcut hint in
// front of it, "(l) Login". It is a plain value; components that contain
// buttons keep the state (which one is focused) and render them with View.
type Button struct {
	// Label is the caption.
	Label string
	// Shortcut, when not zero, is shown before the label.
	Shortcut rune
	// Focused draws the button highlighted.
	Focused bool
	// Disabled draws the button dimmed.
	Disabled bool
	// Background and Foreground override the colours of an unfocused button.
	Background color.Color
	Foreground color.Color
}

// Text is the plain caption: the shortcut hint, when there is one, and the label.
func (b Button) Text() string {
	if b.Shortcut == 0 {
		return b.Label
	}
	return "(" + string(b.Shortcut) + ") " + b.Label
}

// Width is the width that fits the caption with one column of padding each side.
func (b Button) Width() int { return ansi.StringWidth(b.Text()) + 2 }

func (b Button) style() lipgloss.Style {
	switch {
	case b.Disabled:
		return lipgloss.NewStyle().Foreground(theme.MutedColor())
	case b.Focused:
		return theme.SelectedStyle(true)
	}
	style := lipgloss.NewStyle().Foreground(theme.TextColor())
	if b.Background != nil {
		style = style.Background(b.Background)
	}
	if b.Foreground != nil {
		style = style.Foreground(b.Foreground)
	}
	return style
}

// View renders the caption centred in exactly width columns.
func (b Button) View(width int) string {
	if width <= 0 {
		return ""
	}
	style := b.style()
	caption := style.Render(b.Label)
	if b.Shortcut != 0 {
		hint := lipgloss.NewStyle().Foreground(theme.AccentColor())
		if b.Focused {
			hint = style
		}
		caption = hint.Render("("+string(b.Shortcut)+")") + style.Render(" "+b.Label)
	}
	pad := max(width-ansi.StringWidth(caption), 0)
	left := pad / 2
	row := style.Render(strings.Repeat(" ", left)) + caption + style.Render(strings.Repeat(" ", pad-left))
	return ansi.Truncate(row, width, "")
}
