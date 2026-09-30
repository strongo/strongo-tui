package grid

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tuigoff/tuigoff/pkg/widgets"
)

// Update handles a message and returns the receiver (the grid is a mutable
// component, not a value) plus a command. Key handling order: the filter
// input (while focused) always wins; then, ONLY while secondary focus holds
// the active non-table view (see SecondaryFocus/SetSecondaryFocus — always
// true for a non-split view, since the table isn't reachable there), that
// ExtraView's own Update; then the product KeyHandler (see WithKeyHandler),
// which can claim any key including Enter, digits or Tab; then the grid's
// own defaults: h/l select a column (scrolling it into view), up/k and
// down/j move the table row WHILE THE TABLE HAS FOCUS (a split layout's
// primary pane, or the plain table view), home/end jump to the first/last
// row, digit keys switch views (1 is always the table), Tab toggles focus
// between the table and a split secondary view, s sorts by the selected
// column, Enter emits RowActivatedMsg, + emits PinRowMsg for the highlighted
// row's Ref, / opens the built-in filter.
//
// Whenever the highlighted row or the selected column changed because of the
// message, Update also emits one SelectionChangedMsg (a change made with the
// SelectRow/SelectColumn setters is not reported). It emits nothing when the
// selection is unchanged, so holding a key at the last row does not spam.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	before := m.selectionNow()
	cmd := m.update(msg)
	now := m.selectionNow()
	if now == before {
		return m, cmd
	}
	row, _ := m.CurrentRow()
	changed := widgets.Emit(SelectionChangedMsg{Row: row, Index: now.index, Column: now.column, ID: m.id})
	return m, tea.Batch(cmd, changed)
}

func (m *Model) update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	if m.table.GetIsFilterInputFocused() {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return cmd
	}
	if m.secondaryFocus {
		if i := int(m.view) - firstExtraView; i >= 0 && i < len(m.extraViews) && m.extraViews[i].Update != nil {
			if cmd, handled := m.extraViews[i].Update(m, keyMsg); handled {
				return cmd
			}
		}
	}
	if m.keyHandler != nil {
		if cmd, handled := m.keyHandler(m, keyMsg); handled {
			return cmd
		}
	}
	if m.filterOnType && m.filterEnabled() && !m.secondaryFocus && keyMsg.Text != "" && keyMsg.Mod == 0 {
		var cmd tea.Cmd
		m.table, cmd = m.table.StartFilterTyping().Update(msg)
		return cmd
	}
	switch keyMsg.String() {
	case "left", "h":
		if m.rowSelectOnly {
			return nil
		}
		if m.selectedColumn > 0 {
			m.SelectColumn(m.selectedColumn - 1)
		}
		return nil
	case "right", "l":
		if m.rowSelectOnly {
			return nil
		}
		if m.selectedColumn+1 < len(m.columns) {
			m.SelectColumn(m.selectedColumn + 1)
		}
		return nil
	case "up", "k":
		// bubble-table's own RowUp always wraps to the last row; the grid
		// itself does not, matching h/l's clamped column navigation above.
		// Gated on !secondaryFocus rather than m.view == ViewTable: a split
		// layout's primary (table) pane keeps its own row cursor reachable
		// even while a non-table view occupies the secondary pane, as long
		// as secondary focus hasn't been Tab'd onto that view.
		//
		// This moves by POSITION within the table's own visible/filtered
		// row set (see moveRow), not by CurrentIndex()'s source-row
		// arithmetic: under an active filter, adjacent visible rows are not
		// adjacent source rows.
		if !m.secondaryFocus {
			m.moveRow(-1)
			return nil
		}
	case "down", "j":
		if !m.secondaryFocus {
			m.moveRow(1)
			return nil
		}
	case "home":
		if !m.secondaryFocus {
			m.moveRow(-farRows)
			return nil
		}
	case "end":
		if !m.secondaryFocus {
			m.moveRow(farRows)
			return nil
		}
	case "pgup":
		if m.src != nil && !m.secondaryFocus {
			m.moveRow(-m.pageSize())
			return nil
		}
	case "pgdown":
		if m.src != nil && !m.secondaryFocus {
			m.moveRow(m.pageSize())
			return nil
		}
	case "1":
		m.ShowView(ViewTable)
		return nil
	case "tab":
		m.ToggleSecondaryFocusIfSplit()
		return nil
	case "enter":
		if row, ok := m.CurrentRow(); ok {
			return widgets.Emit(RowActivatedMsg{Row: row, Column: m.selectedColumn, ID: m.id})
		}
		return nil
	case "+":
		if ref := m.Current(); ref != nil {
			return widgets.Emit(PinRowMsg{Ref: ref, ID: m.id})
		}
		return nil
	case "s":
		m.Sort(m.selectedColumn)
		return nil
	}
	if n := extraViewKeyIndex(keyMsg.String()); n >= 0 && n < len(m.extraViews) {
		m.ShowView(View(firstExtraView + n))
		return nil
	}
	if m.secondaryFocus {
		// A registered ExtraView with no Update of its own does not consume
		// navigation keys while it holds secondary focus.
		return nil
	}
	// The table has focus — the plain table view, or a split layout's
	// primary pane while secondary focus hasn't been Tab'd onto the active
	// non-table view — so any bubble-table key not already handled above
	// (e.g. pgup/pgdown of a slice-backed grid) still reaches it.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return cmd
}

// farRows is a row delta larger than any real row count, for home/end.
const farRows = 1 << 30

// moveRow moves the highlighted row by delta positions, clamped to the first
// and last row. A slice-backed grid moves within the table's visible
// (filtered) rows; a RowSource grid moves its absolute cursor and loads the
// new page when the cursor leaves the current one.
func (m *Model) moveRow(delta int) {
	if m.src != nil {
		if m.count() > 0 {
			m.moveCursor(m.cursor + delta)
		}
		return
	}
	n := len(m.table.GetVisibleRows())
	if n == 0 {
		return
	}
	m.table = m.table.WithHighlightedRow(max(0, min(n-1, m.table.GetHighlightedRowIndex()+delta)))
}
