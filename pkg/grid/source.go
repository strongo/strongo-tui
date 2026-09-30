package grid

import (
	"github.com/charmbracelet/x/ansi"
)

// RowSource supplies rows lazily, for a result too large to hold as a []Row
// (a table with a million records, a paged database cursor). Pass one with
// WithRowSource instead of a row slice.
//
// The grid asks the source for the rows of ONE page only (at most the page
// size, see WithMaxVisibleRows and SetSize), when the page is loaded — on
// construction, when the highlighted row moves onto another page, and on
// SetSize/Refresh — never on View and never for all Len() rows. Rendering and
// scrolling therefore cost O(page size), independent of Len().
//
// Sort and filter need all the data, so for a RowSource grid:
//   - the "/" filter is disabled (a product that wants search filters its own
//     source and calls Refresh);
//   - sort (the "s" key, Model.Sort) is a no-op unless the source also
//     implements Sorter, in which case the grid delegates to it;
//   - IndexForKey returns -1 and Rows returns only the loaded page.
//
// Column widths cannot be measured over all rows: they are sized from the
// header and every page loaded so far and only ever grow, so scrolling never
// makes a column shrink.
type RowSource interface {
	// Len is the total number of rows. It is called once per key press and
	// once per page load, so it must be cheap.
	Len() int
	// Row returns row i (0 <= i < Len()) in display order.
	Row(i int) Row
}

// Sorter is an optional interface of a RowSource: it can reorder itself.
// SortBy is called by Model.Sort (the "s" key) with the column index and
// direction; the grid then reloads its page. Ordering is the source's
// business (typically an ORDER BY re-query).
type Sorter interface {
	SortBy(column int, desc bool)
}

// WithRowSource backs the grid with a lazy RowSource and its columns instead
// of a []Row (pass nil rows to New). See RowSource for what sort, filter and
// the row accessors do in this mode.
func WithRowSource(src RowSource, columns []Column) Option {
	return func(m *Model) {
		m.src = src
		m.columns = append([]Column(nil), columns...)
	}
}

// Refresh reloads the current page from a RowSource after the source's data
// changed (the cursor is clamped to the new Len()). It does not emit
// SelectionChangedMsg and is a no-op for a slice-backed grid.
func (m *Model) Refresh() {
	if m.src == nil {
		return
	}
	m.loadWindow()
	m.rebuildTable()
}

// count is the total number of rows: Len() of the source or len(rows).
func (m *Model) count() int {
	if m.src != nil {
		return m.src.Len()
	}
	return len(m.rows)
}

// tableRows are the rows (and their formatted cells) the inner table is built
// from: all rows of a slice-backed grid, the loaded page of a RowSource grid.
func (m *Model) tableRows() ([]Row, [][]string) {
	if m.src != nil {
		return m.win, m.winCells
	}
	return m.rows, m.cells
}

// highlightedPosition is the absolute position of the highlighted row within
// the rows the scrollbar spans.
func (m *Model) highlightedPosition() int {
	if m.src != nil {
		return m.cursor
	}
	return m.table.GetHighlightedRowIndex()
}

// loadWindow clamps the cursor and materialises exactly the page that holds
// it: at most pageSize() calls of src.Row.
func (m *Model) loadWindow() {
	n := m.src.Len()
	m.cursor = max(0, min(m.cursor, n-1))
	size := m.pageSize()
	m.top = m.cursor / size * size
	end := min(m.top+size, n)
	m.win = make([]Row, 0, max(0, end-m.top))
	m.winCells = make([][]string, 0, cap(m.win))
	if len(m.seenWidth) < len(m.columns) {
		m.seenWidth = append(m.seenWidth, make([]int, len(m.columns)-len(m.seenWidth))...)
	}
	for i := m.top; i < end; i++ {
		row := m.src.Row(i)
		cells := make([]string, len(m.columns))
		for c := range m.columns {
			cells[c] = cellText(row, c)
			m.seenWidth[c] = max(m.seenWidth[c], ansi.StringWidth(cells[c]))
		}
		m.win = append(m.win, row)
		m.winCells = append(m.winCells, cells)
	}
}

// moveCursor moves a RowSource grid's highlighted row to the absolute index
// (clamped). Within the loaded page only the highlight moves; onto another
// page that page is loaded and the table rebuilt.
func (m *Model) moveCursor(index int) {
	index = max(0, min(index, m.count()-1))
	size := m.pageSize()
	if index/size*size == m.top && index-m.top < len(m.win) {
		m.cursor = index
		m.table = m.table.WithHighlightedRow(index - m.top)
		return
	}
	m.cursor = index
	m.loadWindow()
	m.rebuildTable()
}

// selectSourceRow is SelectRow for a RowSource grid.
func (m *Model) selectSourceRow(index int) {
	if m.count() == 0 {
		return
	}
	m.moveCursor(index)
}

// sortSource is Sort for a RowSource grid: delegated to a Sorter source, a
// no-op otherwise. The highlighted position is kept, like a slice sort.
func (m *Model) sortSource(column int) {
	sorter, ok := m.src.(Sorter)
	if !ok {
		return
	}
	if m.sortColumn == column {
		m.sortDesc = !m.sortDesc
	} else {
		m.sortColumn, m.sortDesc = column, false
	}
	sorter.SortBy(column, m.sortDesc)
	m.loadWindow()
	m.rebuildTable()
}
