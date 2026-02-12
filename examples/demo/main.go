package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/strongo/strongo-tui/pkg/colors"
	"github.com/strongo/strongo-tui/pkg/components/button"
	"github.com/strongo/strongo-tui/pkg/themes"
)

func main() {
	app := tview.NewApplication()

	// Create a text view with colored content
	welcomeText := colors.Success("Welcome to strongo-tui!") + "\n\n" +
		"This demo showcases the components:\n\n" +
		colors.BlueText("• Colors") + " - ANSI and tcell color support\n" +
		colors.BlueText("• Themes") + " - Consistent styling\n" +
		colors.BlueText("• Buttons") + " - Buttons with keyboard shortcuts\n" +
		colors.BlueText("• Boxes") + " - Flexible container components\n\n" +
		colors.Warning("Use Tab/Shift+Tab to navigate between buttons") + "\n" +
		colors.GrayText("Press Ctrl+C to quit")

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(welcomeText)

	// Apply themed border
	themes.SetPanelTitle(textView.Box, " strongo-tui Demo ")

	// Create status text
	statusText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(colors.GrayText("Ready"))

	// Create buttons with shortcuts
	okButton := button.NewWithShortcut("OK", 'o')
	cancelButton := button.NewWithShortcut("Cancel", 'c')
	helpButton := button.NewWithShortcut("Help", 'h')

	// Button handlers
	okButton.SetSelectedFunc(func() {
		statusText.SetText(colors.Success("✓ OK pressed!"))
	})
	cancelButton.SetSelectedFunc(func() {
		statusText.SetText(colors.Danger("✗ Cancelled"))
	})
	helpButton.SetSelectedFunc(func() {
		statusText.SetText(colors.BlueText("ℹ Press buttons or use keyboard shortcuts (o/c/h)"))
	})

	// Create button bar
	buttons := tview.NewFlex().
		AddItem(okButton, 10, 0, true).
		AddItem(nil, 2, 0, false).
		AddItem(cancelButton, 14, 0, false).
		AddItem(nil, 2, 0, false).
		AddItem(helpButton, 12, 0, false)

	// Main layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(buttons, 3, 0, true).
		AddItem(nil, 1, 0, false).
		AddItem(statusText, 1, 0, false)

	// Keyboard shortcuts
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'o', 'O':
			statusText.SetText(colors.Success("✓ OK pressed (keyboard)!"))
			return nil
		case 'c', 'C':
			statusText.SetText(colors.Danger("✗ Cancelled (keyboard)"))
			return nil
		case 'h', 'H':
			statusText.SetText(colors.BlueText("ℹ Shortcuts: (o) OK, (c) Cancel, (h) Help, Ctrl+C to quit"))
			return nil
		case 'q', 'Q':
			app.Stop()
			return nil
		}
		return event
	})

	// Run application
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
	}
}
