package padding

import (
	"github.com/rivo/tview"
	"github.com/strongo/strongo-tui/pkg/components/boxed"
)

// Box wraps a primitive with a border and padding.
// Padding is real layout padding (via Flex), not Box border padding.
func Box[T tview.Primitive](
	content T,
	title string,
	paddingTop, paddingBottom, paddingLeft, paddingRight int,
) boxed.WithBoxType[*tview.Flex] {

	// Horizontal padding
	hPad := tview.NewFlex().
		AddItem(nil, paddingLeft, 0, false).
		AddItem(content, 0, 1, true).
		AddItem(nil, paddingRight, 0, false)

	// Vertical padding
	padded := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, paddingTop, 0, false).
		AddItem(hPad, 0, 1, true).
		AddItem(nil, paddingBottom, 0, false)

	// Border box
	box := tview.NewBox().
		SetBorder(true).
		SetTitle(title)

	flex := tview.NewFlex().
		AddItem(box, 0, 1, false).
		AddItem(padded, 0, 1, true)
	return boxed.WithBoxType[*tview.Flex]{
		Box:       box,
		Primitive: flex,
	}
}
