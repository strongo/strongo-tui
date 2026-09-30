package gridblock

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tuigoff/tuigoff/pkg/entity"
	"github.com/tuigoff/tuigoff/pkg/grid"
	"github.com/tuigoff/tuigoff/pkg/transcript"
)

func newBlock(ref any) *Block {
	cols := []grid.Column{{Name: "name"}}
	rows := []grid.Row{{Values: []any{"alpha"}, Ref: ref}}
	return Wrap(grid.New(cols, rows, grid.WithTitle("results")))
}

func key(s string) tea.KeyPressMsg {
	r := []rune(s)
	return tea.KeyPressMsg{Code: r[0], Text: s}
}

func TestBlockContract(t *testing.T) {
	b := newBlock(nil)
	if !b.Focusable() || !b.SelfFramed() {
		t.Fatal("Focusable and SelfFramed must both be true")
	}
	if b.Title() != "results" {
		t.Errorf("Title = %q", b.Title())
	}
	if !strings.Contains(b.View(40, true), "alpha") {
		t.Error("View lacks row content")
	}
	got, _ := b.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got != transcript.Block(b) {
		t.Error("Update must return the same Block")
	}
}

func TestCurrentPointerValueAndNone(t *testing.T) {
	ref := entity.Ref{Type: "k", Title: "one"}
	if got := newBlock(&ref).Current(); got == nil || !got.Same(ref) || got.Title != ref.Title {
		t.Errorf("pointer ref: got %v", got)
	}
	if got := newBlock(ref).Current(); got == nil || !got.Same(ref) || got.Title != ref.Title {
		t.Errorf("value ref: got %v", got)
	}
	if got := newBlock(nil).Current(); got != nil {
		t.Errorf("no ref: got %v", got)
	}
	if got := newBlock("other").Current(); got != nil {
		t.Errorf("foreign ref type: got %v", got)
	}
	var nilPtr *entity.Ref
	if got := EntityRef(newBlock(nilPtr).Model); got != nil {
		t.Errorf("typed nil: got %v", got)
	}
}

func TestPinPassesThroughUntranslated(t *testing.T) {
	ref := entity.Ref{Type: "k", Title: "one"}
	_, cmd := newBlock(&ref).Update(key("+"))
	if cmd == nil {
		t.Fatal("expected a command")
	}
	msg, ok := cmd().(grid.PinRowMsg)
	if !ok {
		t.Fatalf("want grid.PinRowMsg, got %#v", msg)
	}
	if got := RefOf(msg.Ref); got == nil || !got.Same(ref) {
		t.Fatalf("PinRowMsg.Ref = %#v, want %v", msg.Ref, ref)
	}
}
