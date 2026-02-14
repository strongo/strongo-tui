package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Theme interface {
	FocusedBorderColor() tcell.Color
	FocusedGraphicsColor() tcell.Color
	FocusedSelectedTextStyle() tcell.Style

	BlurredBorderColor() tcell.Color
	BlurredGraphicsColor() tcell.Color
	BlurredSelectedTextStyle() tcell.Style

	LabelColor() tcell.Color

	TableHeaderColor() tcell.Color

	HotkeyColor() tcell.Color
}

var _ Theme = theme{}

// theme defines the color scheme for terminal UI components
type theme struct {
	focusedBorderColor       tcell.Color
	focusedGraphicsColor     tcell.Color
	focusedSelectedTextStyle tcell.Style

	blurredBorderColor       tcell.Color
	blurredGraphicsColor     tcell.Color
	blurredSelectedTextStyle tcell.Style

	labelColor tcell.Color

	tableHeaderColor tcell.Color

	hotkeyColor tcell.Color
}

func (t theme) FocusedBorderColor() tcell.Color       { return t.focusedBorderColor }
func (t theme) FocusedGraphicsColor() tcell.Color     { return t.focusedGraphicsColor }
func (t theme) FocusedSelectedTextStyle() tcell.Style { return t.focusedSelectedTextStyle }
func (t theme) BlurredBorderColor() tcell.Color       { return t.blurredBorderColor }
func (t theme) BlurredGraphicsColor() tcell.Color     { return t.blurredGraphicsColor }
func (t theme) BlurredSelectedTextStyle() tcell.Style { return t.blurredSelectedTextStyle }
func (t theme) LabelColor() tcell.Color               { return t.labelColor }
func (t theme) TableHeaderColor() tcell.Color         { return t.tableHeaderColor }
func (t theme) HotkeyColor() tcell.Color              { return t.hotkeyColor }

// CurrentTheme is the currently active theme
var CurrentTheme = defaultTheme()

// DefaultBorderWithPadding applies default border styling with padding
func DefaultBorderWithPadding(box *tview.Box) {
	DefaultBorderWithoutPadding(box)
	box.SetBorderPadding(1, 1, 2, 2)
}

// DefaultBorderWithoutPadding applies default border styling without padding
func DefaultBorderWithoutPadding(box *tview.Box) {
	box.SetBorder(true)
	box.SetBorderColor(CurrentTheme.BlurredBorderColor())
	box.SetBorderAttributes(tcell.AttrDim)
	box.SetFocusFunc(func() {
		box.SetBorderColor(CurrentTheme.FocusedBorderColor())
	})
	box.SetBlurFunc(func() {
		box.SetBorderColor(CurrentTheme.BlurredBorderColor())
	})
}

// SetPanelTitle sets up a box with a centered title and default border
func SetPanelTitle(box *tview.Box, title string) {
	DefaultBorderWithPadding(box)
	box.SetTitle(title)
	box.SetTitleAlign(tview.AlignCenter)
	box.SetTitleColor(tview.Styles.TitleColor)
}
