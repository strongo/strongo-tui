package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/strongo/strongo-tui/pkg/colors"
)

// Theme defines the color scheme for terminal UI components
type Theme struct {
	FocusedBorderColor       tcell.Color
	FocusedGraphicsColor     tcell.Color
	FocusedSelectedTextStyle tcell.Style

	BlurredBorderColor       tcell.Color
	BlurredGraphicsColor     tcell.Color
	BlurredSelectedTextStyle tcell.Style

	LabelColor tcell.Color

	TableHeaderColor tcell.Color

	HotkeyColor tcell.Color
}

// CurrentTheme is the currently active theme
var CurrentTheme = Theme{
	FocusedBorderColor:   colors.DefaultFocusedBorderColor,
	FocusedGraphicsColor: tcell.ColorWhite,
	FocusedSelectedTextStyle: tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack),

	BlurredBorderColor:   colors.DefaultBlurBorderColor,
	BlurredGraphicsColor: tcell.ColorGray,
	BlurredSelectedTextStyle: tcell.StyleDefault.
		Background(tcell.ColorGray).
		Foreground(tcell.ColorWhite),

	LabelColor:       colors.LabelColor,
	TableHeaderColor: colors.TableHeaderColor,
	HotkeyColor:      colors.HotkeyColor,
}

// DefaultBorderWithPadding applies default border styling with padding
func DefaultBorderWithPadding(box *tview.Box) {
	DefaultBorderWithoutPadding(box)
	box.SetBorderPadding(1, 1, 2, 2)
}

// DefaultBorderWithoutPadding applies default border styling without padding
func DefaultBorderWithoutPadding(box *tview.Box) {
	box.SetBorder(true)
	box.SetBorderColor(CurrentTheme.BlurredBorderColor)
	box.SetBorderAttributes(tcell.AttrDim)
	box.SetFocusFunc(func() {
		box.SetBorderColor(CurrentTheme.FocusedBorderColor)
	})
	box.SetBlurFunc(func() {
		box.SetBorderColor(CurrentTheme.BlurredBorderColor)
	})
}

// SetPanelTitle sets up a box with a centered title and default border
func SetPanelTitle(box *tview.Box, title string) {
	DefaultBorderWithPadding(box)
	box.SetTitle(title)
	box.SetTitleAlign(tview.AlignCenter)
	box.SetTitleColor(tview.Styles.TitleColor)
}
