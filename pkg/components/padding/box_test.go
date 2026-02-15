package padding

import (
	"testing"

	"github.com/rivo/tview"
)

func TestBox(t *testing.T) {
	tv := tview.NewTextView().SetText("hello")
	result := Box(tv, "Test Title", 1, 2, 3, 4)

	if result.Box == nil {
		t.Fatal("expected Box to be non-nil")
	}
	if result.Primitive == nil {
		t.Fatal("expected Primitive to be non-nil")
	}
	if result.Box.GetTitle() != "Test Title" {
		t.Errorf("expected title %q, got %q", "Test Title", result.Box.GetTitle())
	}
}

func TestBox_ZeroPadding(t *testing.T) {
	tv := tview.NewTextView()
	result := Box(tv, "", 0, 0, 0, 0)

	if result.Box == nil {
		t.Fatal("expected Box to be non-nil")
	}
	if result.Primitive == nil {
		t.Fatal("expected Primitive to be non-nil")
	}
}
