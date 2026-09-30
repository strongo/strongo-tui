package gridblock

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tuigoff/tuigoff/pkg/entity"
	"github.com/tuigoff/tuigoff/pkg/grid"
	"github.com/tuigoff/tuigoff/pkg/transcript"
)

// Block adapts *grid.Model to transcript.Block, transcript.EntityBlock and
// transcript.SelfFramed. Model is embedded, so every other grid method
// (Title, CapturesEsc, SetSize, SetFocused, SelectRow, ...) is promoted
// unchanged, and View(width, focused) satisfies transcript.Block as is.
//
// The grid's own messages (grid.PinRowMsg, grid.SelectionChangedMsg, ...) are
// returned untouched: what "pin" means is the composing layer's decision.
type Block struct{ *grid.Model }

var (
	_ transcript.Block       = (*Block)(nil)
	_ transcript.EntityBlock = (*Block)(nil)
	_ transcript.SelfFramed  = (*Block)(nil)
	_ transcript.Titled      = (*Block)(nil)
	_ transcript.EscCapturer = (*Block)(nil)
)

// Wrap returns m as a transcript block.
func Wrap(m *grid.Model) *Block { return &Block{Model: m} }

// Focusable reports that a grid block always takes a focus stop.
func (b *Block) Focusable() bool { return true }

// SelfFramed reports that the grid draws its own complete border, so
// transcript must not wrap it in a card fill.
func (b *Block) SelfFramed() bool { return true }

// Update forwards msg to the grid and returns the same Block (identity is
// kept).
func (b *Block) Update(msg tea.Msg) (transcript.Block, tea.Cmd) {
	_, cmd := b.Model.Update(msg)
	return b, cmd
}

// Current returns the entity under the cursor, or nil.
func (b *Block) Current() *entity.Ref { return EntityRef(b.Model) }

// EntityRef returns the entity.Ref stored (as pointer or value) in the Ref of
// m's highlighted row, or nil when there is none.
func EntityRef(m *grid.Model) *entity.Ref { return RefOf(m.Current()) }

// RefOf returns the entity.Ref held by v (a *entity.Ref or an entity.Ref, as
// found in grid.Row.Ref and grid.PinRowMsg.Ref), or nil for anything else,
// including a nil pointer.
func RefOf(v any) *entity.Ref {
	switch r := v.(type) {
	case *entity.Ref:
		return r
	case entity.Ref:
		return &r
	}
	return nil
}
