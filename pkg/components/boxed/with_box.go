package boxed

import (
	"github.com/rivo/tview"
	"github.com/strongo/strongo-tui/pkg/themes"
)

// PrimitiveWithBox is an interface for primitives that have a Box
type PrimitiveWithBox interface {
	tview.Primitive
	GetBox() *tview.Box
}

var _ PrimitiveWithBox = (*WithBoxType[tview.Primitive])(nil)

// WithBoxType wraps a primitive with a Box for border management
type WithBoxType[T tview.Primitive] struct {
	tview.Primitive
	Box *tview.Box
}

// GetBox returns the Box associated with this primitive
func (p WithBoxType[T]) GetBox() *tview.Box {
	return p.Box
}

// GetPrimitive returns the wrapped primitive
func (p WithBoxType[T]) GetPrimitive() T {
	return p.Primitive.(T)
}

// WithDefaultBorders wraps a primitive with a box and applies default border styling with padding
func WithDefaultBorders[T tview.Primitive](p T, box *tview.Box) WithBoxType[T] {
	themes.DefaultBorderWithPadding(box)
	return WithBoxType[T]{
		Primitive: p,
		Box:       box,
	}
}

// WithBordersWithoutPadding wraps a primitive with a box and applies default border styling without padding
func WithBordersWithoutPadding[T tview.Primitive](p T, box *tview.Box) WithBoxType[T] {
	themes.DefaultBorderWithoutPadding(box)
	return WithBoxType[T]{
		Primitive: p,
		Box:       box,
	}
}

// WithBoxWithoutBorder wraps a primitive with a box but without applying border styling
func WithBoxWithoutBorder[T tview.Primitive](p T, box *tview.Box) WithBoxType[T] {
	return WithBoxType[T]{
		Primitive: p,
		Box:       box,
	}
}
