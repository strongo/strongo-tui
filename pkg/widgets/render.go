package widgets

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Align positions text inside a cell or a line.
type Align int

// Alignments.
const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// PadRight pads s with spaces on the right to width columns, or truncates it
// (with an ellipsis) when it is wider.
func PadRight(s string, width int) string {
	return AlignIn(s, width, AlignLeft)
}

// AlignIn fits s into exactly width columns: wider text is truncated with an
// ellipsis, narrower text is padded according to align.
func AlignIn(s string, width int, align Align) string {
	if width <= 0 {
		return ""
	}
	w := ansi.StringWidth(s)
	if w > width {
		return ansi.Truncate(s, width, "…")
	}
	gap := width - w
	switch align {
	case AlignCenter:
		left := gap / 2
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", gap-left)
	case AlignRight:
		return strings.Repeat(" ", gap) + s
	default:
		return s + strings.Repeat(" ", gap)
	}
}

// Fit reshapes s into exactly height lines of exactly width columns, padding
// short lines and blank rows and truncating what does not fit.
func Fit(s string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	out := make([]string, height)
	for i := range out {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		out[i] = PadRight(line, width)
	}
	return strings.Join(out, "\n")
}

// Blank returns a width x height block of spaces.
func Blank(width, height int) string {
	return Fit("", width, height)
}

// Overlay draws top over base with its top-left corner at column x, row y. Both
// may carry ANSI styling. Parts of top that fall outside base are clipped.
func Overlay(base, top string, x, y int) string {
	x = max(x, 0)
	baseLines := strings.Split(base, "\n")
	baseWidth := 0
	for _, line := range baseLines {
		baseWidth = max(baseWidth, ansi.StringWidth(line))
	}
	for i, topLine := range strings.Split(top, "\n") {
		row := y + i
		if row < 0 || row >= len(baseLines) {
			continue
		}
		line := baseLines[row]
		topLine = ansi.Truncate(topLine, max(baseWidth-x, 0), "")
		lineWidth := ansi.StringWidth(line)
		topWidth := ansi.StringWidth(topLine)
		left := ansi.Truncate(line, x, "")
		if pad := x - ansi.StringWidth(left); pad > 0 {
			left += strings.Repeat(" ", pad)
		}
		right := ""
		if end := x + topWidth; end < lineWidth {
			right = ansi.Cut(line, end, lineWidth)
		}
		baseLines[row] = left + "\x1b[m" + topLine + "\x1b[m" + right
	}
	return strings.Join(baseLines, "\n")
}

// Center overlays top in the middle of a width x height base.
func Center(base, top string, width, height int) string {
	lines := strings.Split(top, "\n")
	topWidth := 0
	for _, l := range lines {
		topWidth = max(topWidth, ansi.StringWidth(l))
	}
	return Overlay(base, top, max((width-topWidth)/2, 0), max((height-len(lines))/2, 0))
}
