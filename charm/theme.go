// Package charm provides semantic Lip Gloss styles for terminal applications.
//
// It deliberately owns presentation tokens only. Layout, focus management, and
// application state remain the responsibility of the consuming application.
package charm

import "charm.land/lipgloss/v2"

// Palette contains the terminal colours used to build a Theme. Colour values
// use any format accepted by lipgloss.Color, such as "12", "#7D56F4", or
// "magenta".
//
// Use DefaultPalette as a starting point when customizing a theme.
type Palette struct {
	FocusedBorder      string
	BlurredBorder      string
	Title              string
	Text               string
	Muted              string
	Help               string
	SelectedText       string
	SelectedBackground string
	Error              string
}

// Theme is an immutable collection of semantic Lip Gloss styles. Its methods
// return style values, so applications can derive a one-off variation without
// changing the theme shared by other views.
type Theme struct {
	focusedPanel  lipgloss.Style
	blurredPanel  lipgloss.Style
	title         lipgloss.Style
	selectedRow   lipgloss.Style
	unselectedRow lipgloss.Style
	muted         lipgloss.Style
	help          lipgloss.Style
	loading       lipgloss.Style
	empty         lipgloss.Style
	err           lipgloss.Style
}

// DefaultPalette returns the colours used by DefaultTheme.
func DefaultPalette() Palette {
	return Palette{
		FocusedBorder:      "12",
		BlurredBorder:      "8",
		Title:              "12",
		Text:               "15",
		Muted:              "8",
		Help:               "14",
		SelectedText:       "0",
		SelectedBackground: "12",
		Error:              "9",
	}
}

// DefaultTheme returns the standard Filetug-ready theme. Each call returns a
// value, not a process-wide mutable theme.
func DefaultTheme() Theme {
	return NewTheme(DefaultPalette())
}

// NewTheme creates a Theme from a colour palette.
func NewTheme(palette Palette) Theme {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	return Theme{
		focusedPanel: panel.BorderForeground(lipgloss.Color(palette.FocusedBorder)),
		blurredPanel: panel.BorderForeground(lipgloss.Color(palette.BlurredBorder)),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Title)).
			Bold(true),
		selectedRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.SelectedText)).
			Background(lipgloss.Color(palette.SelectedBackground)).
			Bold(true),
		unselectedRow: lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Text)),
		muted:         lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Muted)),
		help:          lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Help)),
		loading: lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Help)).
			Italic(true),
		empty: lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Muted)).
			Italic(true),
		err: lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Error)).
			Bold(true),
	}
}

// FocusedPanel styles the border of a panel that currently receives input.
func (theme Theme) FocusedPanel() lipgloss.Style { return theme.focusedPanel }

// BlurredPanel styles the border of a panel that does not receive input.
func (theme Theme) BlurredPanel() lipgloss.Style { return theme.blurredPanel }

// Title styles a panel or section title.
func (theme Theme) Title() lipgloss.Style { return theme.title }

// SelectedRow styles the active row in a list or table.
func (theme Theme) SelectedRow() lipgloss.Style { return theme.selectedRow }

// UnselectedRow styles an inactive row in a list or table.
func (theme Theme) UnselectedRow() lipgloss.Style { return theme.unselectedRow }

// Muted styles secondary, low-emphasis text.
func (theme Theme) Muted() lipgloss.Style { return theme.muted }

// Help styles keyboard hints and other guidance.
func (theme Theme) Help() lipgloss.Style { return theme.help }

// Loading styles an in-progress status.
func (theme Theme) Loading() lipgloss.Style { return theme.loading }

// Empty styles an empty-state message.
func (theme Theme) Empty() lipgloss.Style { return theme.empty }

// Error styles an error message.
func (theme Theme) Error() lipgloss.Style { return theme.err }
