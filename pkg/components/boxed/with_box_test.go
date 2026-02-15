package boxed

import (
	"testing"

	"github.com/rivo/tview"
)

func TestWithBoxType_GetBox(t *testing.T) {
	box := tview.NewBox()
	tv := tview.NewTextView()
	wbt := WithBoxType[*tview.TextView]{
		Primitive: tv,
		Box:       box,
	}
	if got := wbt.GetBox(); got != box {
		t.Errorf("GetBox() returned wrong box")
	}
}

func TestWithBoxType_GetPrimitive(t *testing.T) {
	tv := tview.NewTextView()
	wbt := WithBoxType[*tview.TextView]{
		Primitive: tv,
		Box:       tview.NewBox(),
	}
	if got := wbt.GetPrimitive(); got != tv {
		t.Errorf("GetPrimitive() returned wrong primitive")
	}
}

func TestWithDefaultBorders(t *testing.T) {
	tv := tview.NewTextView()
	box := tview.NewBox()
	result := WithDefaultBorders(tv, box)

	if result.Primitive != tv {
		t.Errorf("expected Primitive to be the text view")
	}
	if result.Box != box {
		t.Errorf("expected Box to be the provided box")
	}
}

func TestWithBordersWithoutPadding(t *testing.T) {
	tv := tview.NewTextView()
	box := tview.NewBox()
	result := WithBordersWithoutPadding(tv, box)

	if result.Primitive != tv {
		t.Errorf("expected Primitive to be the text view")
	}
	if result.Box != box {
		t.Errorf("expected Box to be the provided box")
	}
}

func TestWithBoxWithoutBorder(t *testing.T) {
	tv := tview.NewTextView()
	box := tview.NewBox()
	result := WithBoxWithoutBorder(tv, box)

	if result.Primitive != tv {
		t.Errorf("expected Primitive to be the text view")
	}
	if result.Box != box {
		t.Errorf("expected Box to be the provided box")
	}
}

func TestPrimitiveWithBoxInterface(t *testing.T) {
	tv := tview.NewTextView()
	box := tview.NewBox()
	wbt := WithDefaultBorders(tv, box)

	var _ PrimitiveWithBox = wbt
}
