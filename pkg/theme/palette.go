package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Named colours use the X11 names and values. They are fixed hues for text that
// must not follow the light or dark variant, such as a cell of a given data type
// or a node of a given kind.
var (
	Black                = lipgloss.Color("#000000")
	White                = lipgloss.Color("#FFFFFF")
	WhiteSmoke           = lipgloss.Color("#F5F5F5")
	Gray                 = lipgloss.Color("#808080")
	Grey                 = Gray
	DarkGray             = lipgloss.Color("#A9A9A9")
	DarkSlateGray        = lipgloss.Color("#2F4F4F")
	LightGray            = lipgloss.Color("#D3D3D3")
	LightGrey            = LightGray
	Red                  = lipgloss.Color("#FF0000")
	DarkRed              = lipgloss.Color("#8B0000")
	PaleVioletRed        = lipgloss.Color("#DB7093")
	LightCoral           = lipgloss.Color("#F08080")
	LightSalmon          = lipgloss.Color("#FFA07A")
	Orange               = lipgloss.Color("#FFA500")
	Gold                 = lipgloss.Color("#FFD700")
	Yellow               = lipgloss.Color("#FFFF00")
	LightYellow          = lipgloss.Color("#FFFFE0")
	LightGoldenrodYellow = lipgloss.Color("#FAFAD2")
	Green                = lipgloss.Color("#008000")
	Lime                 = lipgloss.Color("#00FF00")
	Cyan                 = lipgloss.Color("#00FFFF")
	Blue                 = lipgloss.Color("#0000FF")
	DarkBlue             = lipgloss.Color("#00008B")
	CornflowerBlue       = lipgloss.Color("#6495ED")
	LightBlue            = lipgloss.Color("#ADD8E6")
	LightSteelBlue       = lipgloss.Color("#B0C4DE")
	Magenta              = lipgloss.Color("#FF00FF")
)

// Semantic colours of data screens. They are fixed hues; use FocusColor,
// MutedColor and AccentColor for colours that must follow the terminal
// background.
var (
	// TableColumnTitle colours the title of a table column.
	TableColumnTitle color.Color = LightBlue
	// TableTertiaryText colours de-emphasised table text.
	TableTertiaryText color.Color = Gray
	// TableHeaderColor colours a table header row.
	TableHeaderColor color.Color = WhiteSmoke
	// TreeNodeLink colours a tree node that navigates somewhere.
	TreeNodeLink color.Color = Blue
)

// ErrorColor is the colour of error text: a red that reads on both variants.
func ErrorColor() color.Color { return pick(lipgloss.Color("#B3261E"), lipgloss.Color("#FF6B6B")) }

// TextColor is the colour of ordinary text on the neutral surface.
func TextColor() color.Color {
	_, fg := SurfaceColors()
	return fg
}

// SelectedStyle styles the selected row, cell or node: the shared focus accent
// while the component holds focus, a quieter surface tint while it does not.
func SelectedStyle(focused bool) lipgloss.Style {
	if focused {
		bg, fg := FocusSurfaceColors()
		return lipgloss.NewStyle().Bold(true).Background(bg).Foreground(fg)
	}
	bg, fg := SurfaceColors()
	return lipgloss.NewStyle().Bold(true).Background(bg).Foreground(fg)
}
