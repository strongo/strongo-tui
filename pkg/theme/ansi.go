package theme

import "fmt"

// ANSIColor is an ANSI SGR escape sequence that switches the foreground colour.
type ANSIColor string

// ANSI foreground colour sequences.
const (
	ANSIRed    ANSIColor = "\x1b[91m"
	ANSIGreen  ANSIColor = "\x1b[32m"
	ANSIBlue   ANSIColor = "\x1b[94m"
	ANSIGray   ANSIColor = "\x1b[90m"
	ANSIYellow ANSIColor = "\x1b[33m"
)

var defaultColor ANSIColor = "\x1b[39m"

// SetDefaultColor sets the sequence appended after coloured text to restore the
// foreground colour.
func SetDefaultColor(color ANSIColor) { defaultColor = color }

// Colorize wraps s in color and the reset sequence set by SetDefaultColor. The
// result can be placed inside any text that is rendered through a component.
func Colorize(s string, color ANSIColor) string {
	return fmt.Sprintf("%s%s%s", color, s, defaultColor)
}

// RedText returns s coloured red.
func RedText(s string) string { return Colorize(s, ANSIRed) }

// GreenText returns s coloured green.
func GreenText(s string) string { return Colorize(s, ANSIGreen) }

// BlueText returns s coloured blue.
func BlueText(s string) string { return Colorize(s, ANSIBlue) }

// GrayText returns s coloured gray.
func GrayText(s string) string { return Colorize(s, ANSIGray) }

// YellowText returns s coloured yellow.
func YellowText(s string) string { return Colorize(s, ANSIYellow) }

// Danger returns s in the danger colour (red).
func Danger(s string) string { return Colorize(s, ANSIRed) }

// Warning returns s in the warning colour (yellow).
func Warning(s string) string { return Colorize(s, ANSIYellow) }

// Success returns s in the success colour (green).
func Success(s string) string { return Colorize(s, ANSIGreen) }
