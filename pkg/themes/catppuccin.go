package themes

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/gdamore/tcell/v2"
)

func CatppuccinColor(c catppuccin.Color) tcell.Color {
	return tcell.ColorIsRGB | tcell.NewRGBColor(
		int32(c.RGB[0]),
		int32(c.RGB[1]),
		int32(c.RGB[2]),
	)
}

func CatppuccinTheme(flavor catppuccin.Flavor) Theme {
	return catppuccinTheme{
		flavor: flavor,
	}
}

type catppuccinTheme struct {
	flavor catppuccin.Flavor
}

func (t catppuccinTheme) FocusedBorderColor() tcell.Color {
	return CatppuccinColor(t.flavor.Lavender())
}

func (t catppuccinTheme) FocusedGraphicsColor() tcell.Color {
	return CatppuccinColor(t.flavor.Text())
}

func (t catppuccinTheme) FocusedSelectedTextStyle() tcell.Style {
	return tcell.StyleDefault.
		Background(CatppuccinColor(t.flavor.Surface1())).
		Foreground(CatppuccinColor(t.flavor.Text()))
}

func (t catppuccinTheme) BlurredBorderColor() tcell.Color {
	return CatppuccinColor(t.flavor.Surface1())
}

func (t catppuccinTheme) BlurredGraphicsColor() tcell.Color {
	return CatppuccinColor(t.flavor.Overlay0())
}

func (t catppuccinTheme) BlurredSelectedTextStyle() tcell.Style {
	return tcell.StyleDefault.
		Background(CatppuccinColor(t.flavor.Surface0())).
		Foreground(CatppuccinColor(t.flavor.Subtext0()))
}

func (t catppuccinTheme) LabelColor() tcell.Color {
	return CatppuccinColor(t.flavor.Blue())
}

func (t catppuccinTheme) TableHeaderColor() tcell.Color {
	return CatppuccinColor(t.flavor.Mauve())
}

func (t catppuccinTheme) HotkeyColor() tcell.Color {
	return CatppuccinColor(t.flavor.Yellow())
}
