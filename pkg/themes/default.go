package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/strongo/strongo-tui/pkg/colors"
)

func defaultTheme() Theme {
	return theme{
		focusedBorderColor:   colors.DefaultFocusedBorderColor,
		focusedGraphicsColor: tcell.ColorWhite,
		focusedSelectedTextStyle: tcell.StyleDefault.
			Background(tcell.ColorWhite).
			Foreground(tcell.ColorBlack),

		blurredBorderColor:   colors.DefaultBlurBorderColor,
		blurredGraphicsColor: tcell.ColorGray,
		blurredSelectedTextStyle: tcell.StyleDefault.
			Background(tcell.ColorGray).
			Foreground(tcell.ColorWhite),

		labelColor:       colors.LabelColor,
		tableHeaderColor: colors.TableHeaderColor,
		hotkeyColor:      colors.HotkeyColor,
	}
}
