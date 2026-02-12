package colors

import "github.com/gdamore/tcell/v2"

// TCell color constants for terminal UI
var (
	// Table colors
	TableColumnTitle  = tcell.ColorLightBlue
	TableTertiaryText = tcell.ColorGray
	TableHeaderColor  = tcell.ColorWhiteSmoke

	// Tree/Navigation colors
	TreeNodeLink = tcell.ColorBlue

	// Border colors
	DefaultFocusedBorderColor = tcell.ColorCornflowerBlue
	DefaultBlurBorderColor    = tcell.ColorGray

	// General UI colors
	LabelColor  = tcell.ColorDarkGray
	HotkeyColor = tcell.ColorWhite
)
