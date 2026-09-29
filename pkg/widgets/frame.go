package widgets

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
)

// Frame is the bordered, titled box a panel is drawn in. The border takes the
// theme's focused colour while Focused is true and its blurred colour otherwise.
// Frame is a plain value: With... methods return a modified copy, and Render is a
// pure function of the frame and its arguments.
//
//	frame := widgets.NewFrame().WithTitle("Projects").WithFocus(true)
//	out := frame.Render(list.View(), 40, 12)
type Frame struct {
	// Title is drawn in the top border.
	Title string
	// Focused selects the focused border colour.
	Focused bool
	// Borders draws the border; without it the content fills the whole rectangle.
	Borders bool
	// Padding is the space between the border and the content.
	Padding Padding
	// TitleAlign positions the title along the top border.
	TitleAlign Align
}

// Padding is a space in rows and columns.
type Padding struct{ Top, Right, Bottom, Left int }

// NewFrame returns a bordered frame without padding and with a centred title.
func NewFrame() Frame { return Frame{Borders: true, TitleAlign: AlignCenter} }

// WithTitle sets the title.
func (f Frame) WithTitle(title string) Frame { f.Title = title; return f }

// WithFocus sets whether the frame is drawn as focused.
func (f Frame) WithFocus(focused bool) Frame { f.Focused = focused; return f }

// WithoutBorders removes the border.
func (f Frame) WithoutBorders() Frame { f.Borders = false; return f }

// WithPadding sets the padding between the border and the content.
func (f Frame) WithPadding(top, right, bottom, left int) Frame {
	f.Padding = Padding{top, right, bottom, left}
	return f
}

// WithTitleAlign positions the title along the top border.
func (f Frame) WithTitleAlign(align Align) Frame { f.TitleAlign = align; return f }

func (f Frame) edge() int {
	if f.Borders {
		return 1
	}
	return 0
}

// Origin is the position of the content's top-left cell inside the frame.
func (f Frame) Origin() (x, y int) {
	return f.edge() + f.Padding.Left, f.edge() + f.Padding.Top
}

// Inner returns the size of the content area of a frame drawn into
// width x height; a dimension is never negative.
func (f Frame) Inner(width, height int) (w, h int) {
	e := f.edge()
	return max(width-2*e-f.Padding.Left-f.Padding.Right, 0),
		max(height-2*e-f.Padding.Top-f.Padding.Bottom, 0)
}

// Render draws content, which should be Inner(width, height) cells, inside the
// frame and returns exactly height lines of width columns.
func (f Frame) Render(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	iw, ih := f.Inner(width, height)
	if iw <= 0 || ih <= 0 {
		if f.Borders && width >= 2 && height >= 2 {
			middle := make([]string, height-2)
			for i := range middle {
				middle[i] = strings.Repeat(" ", width-2)
			}
			return f.border(strings.Join(middle, "\n"), width, height)
		}
		return Blank(width, height)
	}
	body := Fit(content, iw, ih)
	pad := func(line string) string {
		return strings.Repeat(" ", f.Padding.Left) + line + strings.Repeat(" ", f.Padding.Right)
	}
	lines := make([]string, 0, height)
	for range f.Padding.Top {
		lines = append(lines, pad(strings.Repeat(" ", iw)))
	}
	for _, line := range strings.Split(body, "\n") {
		lines = append(lines, pad(line))
	}
	for range f.Padding.Bottom {
		lines = append(lines, pad(strings.Repeat(" ", iw)))
	}
	inner := strings.Join(lines, "\n")
	if !f.Borders {
		return inner
	}
	return f.border(inner, width, height)
}

func (f Frame) borderColor() color.Color {
	if f.Focused {
		return theme.FocusColor()
	}
	return theme.MutedColor()
}

// border surrounds content (already width-2 columns wide) with the border.
func (f Frame) border(content string, width, height int) string {
	style := lipgloss.NewStyle().Foreground(f.borderColor())
	edge := lipgloss.RoundedBorder()
	inner := width - 2
	lines := make([]string, 0, height)
	lines = append(lines, style.Render(edge.TopLeft)+f.titleRow(style, edge.Top, inner)+style.Render(edge.TopRight))
	left, right := style.Render(edge.Left), style.Render(edge.Right)
	if height > 2 {
		for _, line := range strings.Split(content, "\n") {
			lines = append(lines, left+line+right)
		}
	}
	lines = append(lines, style.Render(edge.BottomLeft+strings.Repeat(edge.Bottom, inner)+edge.BottomRight))
	return strings.Join(lines, "\n")
}

// titleRow renders the top border between the corners, with the title in it.
func (f Frame) titleRow(style lipgloss.Style, dash string, inner int) string {
	if f.Title == "" || inner < 5 {
		return style.Render(strings.Repeat(dash, inner))
	}
	label := " " + ansi.Truncate(f.Title, inner-2, "…") + " "
	room := inner - ansi.StringWidth(label)
	var before int
	switch f.TitleAlign {
	case AlignLeft:
		before = min(1, room)
	case AlignRight:
		before = max(room-1, 0)
	default:
		before = room / 2
	}
	text := lipgloss.NewStyle().Bold(true).Foreground(theme.TextColor()).Render(label)
	return style.Render(strings.Repeat(dash, before)) + text + style.Render(strings.Repeat(dash, room-before))
}
