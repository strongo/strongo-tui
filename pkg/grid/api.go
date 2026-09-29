package grid

import (
	"charm.land/lipgloss/v2"

	"github.com/strongo/strongo-tui/pkg/widgets"
)

// PinRowMsg is emitted by the "+" key for the highlighted row when that row
// has a Ref: the product decides what pinning means (a sidebar entry, a
// bookmark). Ref is the row's opaque Row.Ref; ID is the grid's WithID value.
type PinRowMsg struct {
	Ref any
	ID  string
}

// SelectionChangedMsg is emitted by Update whenever the highlighted row or the
// selected column changed because of the message just handled (cursor keys,
// filter typing, sort, a KeyHandler calling SelectRow, ...), so a screen can
// show the details of the current cell: a foreign-key target, the referrers
// of the current row. It is emitted once per change, never while the
// selection stays put, and never for the SelectRow/SelectColumn/Refresh calls
// a product makes outside Update. Row is the highlighted row (the zero Row
// and Index -1 when the grid has no rows), Index its display index (the same
// index space as SelectRow), Column the selected column, ID the WithID value.
// A sort that puts a different row (a different Row.Key) under the same
// highlighted position is also a change.
type SelectionChangedMsg struct {
	Row    Row
	Index  int
	Column int
	ID     string
}

// selection is what SelectionChangedMsg compares: the highlighted display
// index, the selected column and the highlighted row's Key.
type selection struct {
	index, column int
	key           string
}

func (m *Model) selectionNow() selection {
	row, _ := m.CurrentRow()
	return selection{index: m.CurrentIndex(), column: m.selectedColumn, key: row.Key}
}

// CellStyleFunc styles one cell: row is the row being drawn, column the
// column index and value the row's raw value for it (Absent when the row has
// none). The returned style is layered over the grid's own column style
// (alignment, selected-column emphasis): properties the function sets win.
// It is a data hook (type colours, underlined foreign-key values, error
// cells), never behaviour, and must be cheap and pure: it runs for every
// visible cell whenever the table is rebuilt. Colours must come from
// pkg/theme.
type CellStyleFunc func(row Row, column int, value any) lipgloss.Style

// WithCellStyle registers the per-cell style hook (see CellStyleFunc).
func WithCellStyle(fn CellStyleFunc) Option {
	return func(m *Model) { m.cellStyle = fn }
}

// WithID names the grid: the id is copied into every message the grid emits
// (RowActivatedMsg, PinRowMsg, SelectionChangedMsg), so a screen that holds
// several grids can tell them apart.
func WithID(id string) Option {
	return func(m *Model) { m.id = id }
}

// WithFixedColumns freezes the first n columns: they stay visible at the left
// while the remaining columns scroll horizontally under the selected-column
// navigation (h/l), e.g. an ID or name column. n is clamped to the number of
// columns. The frozen columns count against the width, so a grid too narrow
// for them shows an empty scroll region.
func WithFixedColumns(n int) Option {
	return func(m *Model) { m.fixedColumns = n }
}

// WithoutFrame makes the grid render just the table and its footer line, with
// no border, title, focus bullet or scrollbar, for a screen that already
// frames it (widgets.Frame, the navigation shell). The view switcher line is
// kept, above the table, only when extra views exist (and
// WithoutSwitcher was not given). Split layouts still draw their secondary
// view in a small card of its own.
func WithoutFrame() Option {
	return func(m *Model) { m.frameless = true }
}

// WithRowSelection makes the grid a row list: whole rows are selected and no
// column is (h/l and the arrow keys sideways do nothing, the selected column
// stays 0 and is not emphasised, SelectionChangedMsg.Column and
// RowActivatedMsg.Column are always 0). For simple name lists such as a
// table chooser. Horizontal scrolling is then not reachable.
func WithRowSelection() Option {
	return func(m *Model) { m.rowSelectOnly = true }
}

// WithFilterOnType lets printable keys start and feed the row filter directly,
// without pressing "/" first (a searchable list: type to narrow, Enter blurs
// the input, Esc clears it). Those keys are then no longer the grid's own
// h/j/k/l/s/+/digit shortcuts (arrows, home/end and Enter still are), and a
// WithKeyHandler still sees every key first. It has no effect when the filter
// is disabled or the grid is backed by a RowSource.
func WithFilterOnType() Option {
	return func(m *Model) { m.filterOnType = true }
}

var (
	_ widgets.Boundary = (*Model)(nil)
	_ widgets.Editor   = (*Model)(nil)
)

// AtEdge implements widgets.Boundary so the shell may turn an arrow key into
// a focus move: Up is true on the first row, Down on the last row (both also
// when there are no rows), Left when the first column is selected and Right
// when the last one is (both always in a WithRowSelection grid). While a split or extra view holds secondary focus the
// grid keeps Up and Down for that view (false) and Left/Right are true.
func (m *Model) AtEdge(dir widgets.Direction) bool {
	switch dir {
	case widgets.Up, widgets.Down:
		if m.secondaryFocus {
			return false
		}
		if m.src != nil {
			return m.count() == 0 || (dir == widgets.Up && m.cursor == 0) || (dir == widgets.Down && m.cursor == m.count()-1)
		}
		n := len(m.table.GetVisibleRows())
		pos := m.table.GetHighlightedRowIndex()
		return n == 0 || (dir == widgets.Up && pos == 0) || (dir == widgets.Down && pos == n-1)
	case widgets.Left:
		return m.secondaryFocus || m.rowSelectOnly || m.selectedColumn == 0
	case widgets.Right:
		return m.secondaryFocus || m.rowSelectOnly || m.selectedColumn >= len(m.columns)-1
	}
	return false
}

// Editing implements widgets.Editor: true while the grid's filter input is
// focused and owns a text cursor.
func (m *Model) Editing() bool { return m.table.GetIsFilterInputFocused() }
