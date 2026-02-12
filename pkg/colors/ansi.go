package colors

import "fmt"

// ANSIColor represents an ANSI color escape code
type ANSIColor string

const (
	// ANSI color codes
	Red    ANSIColor = "\x1b[91m"
	Green  ANSIColor = "\x1b[32m"
	Blue   ANSIColor = "\x1b[94m"
	Gray   ANSIColor = "\x1b[90m"
	Yellow ANSIColor = "\x1b[33m"
)

var defaultColor ANSIColor = "\x1b[39m"

// SetDefaultColor sets the default color to reset to
func SetDefaultColor(color ANSIColor) {
	defaultColor = color
}

// Colorize wraps a string with ANSI color codes
func Colorize(s string, color ANSIColor) string {
	return fmt.Sprintf("%s%s%s", color, s, defaultColor)
}

// Red returns the string colored in red
func RedText(s string) string {
	return Colorize(s, Red)
}

// Green returns the string colored in green
func GreenText(s string) string {
	return Colorize(s, Green)
}

// Blue returns the string colored in blue
func BlueText(s string) string {
	return Colorize(s, Blue)
}

// Gray returns the string colored in gray
func GrayText(s string) string {
	return Colorize(s, Gray)
}

// Yellow returns the string colored in yellow
func YellowText(s string) string {
	return Colorize(s, Yellow)
}

// Semantic color helpers
var (
	dangerColor  = Red
	successColor = Green
	warningColor = Yellow
)

// Danger returns the string in danger color (red)
func Danger(s string) string {
	return Colorize(s, dangerColor)
}

// Warning returns the string in warning color (yellow)
func Warning(s string) string {
	return Colorize(s, warningColor)
}

// Success returns the string in success color (green)
func Success(s string) string {
	return Colorize(s, successColor)
}
