package grid

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

// numberedRows builds n rows of width columns whose values are "r<row>c<col>".
func numberedRows(n, width int) ([]Column, []Row) {
	cols := make([]Column, width)
	for c := range cols {
		cols[c] = Column{Name: fmt.Sprintf("col%d", c)}
	}
	rows := make([]Row, n)
	for i := range rows {
		values := make([]any, width)
		for c := range values {
			values[c] = fmt.Sprintf("r%dc%d", i, c)
		}
		rows[i] = Row{Key: fmt.Sprint(i), Values: values, Ref: testRef{Type: "row", ID: fmt.Sprint(i)}}
	}
	return cols, rows
}

func selectionMsgs(cmd tea.Cmd) []SelectionChangedMsg {
	var out []SelectionChangedMsg
	for _, msg := range uitest.Msgs(cmd) {
		if sel, ok := msg.(SelectionChangedMsg); ok {
			out = append(out, sel)
		}
	}
	return out
}

func TestSelectionChangedOnRowAndColumnMoves(t *testing.T) {
	cols, rows := numberedRows(3, 3)
	m := New(cols, rows, WithID("cities"))

	_, cmd := m.Update(uitest.Key("down"))
	got := selectionMsgs(cmd)
	if len(got) != 1 || got[0].Index != 1 || got[0].Column != 0 || got[0].Row.Key != "1" || got[0].ID != "cities" {
		t.Fatalf("down emitted %+v", got)
	}
	_, cmd = m.Update(uitest.Key("l"))
	got = selectionMsgs(cmd)
	if len(got) != 1 || got[0].Index != 1 || got[0].Column != 1 {
		t.Fatalf("l emitted %+v", got)
	}
	_, cmd = m.Update(uitest.Key("up"))
	if got = selectionMsgs(cmd); len(got) != 1 || got[0].Index != 0 || got[0].Column != 1 {
		t.Fatalf("up emitted %+v", got)
	}
}

func TestSelectionChangedDoesNotSpam(t *testing.T) {
	cols, rows := numberedRows(2, 2)
	m := New(cols, rows)
	for _, key := range []string{"up", "h", "x"} { // already at the first row / column
		if _, cmd := m.Update(uitest.Key(key)); cmd != nil {
			t.Fatalf("%q at the edge returned a command", key)
		}
	}
	m.Update(uitest.Key("down"))
	m.Update(uitest.Key("l"))
	for _, key := range []string{"down", "j", "l", "right"} { // already at the last row / column
		_, cmd := m.Update(uitest.Key(key))
		if len(selectionMsgs(cmd)) != 0 {
			t.Fatalf("%q at the far edge re-emitted the selection", key)
		}
	}
	if _, cmd := m.Update(featureOtherMsg{}); cmd != nil {
		t.Fatal("a foreign message emitted something")
	}
}

func TestSelectionChangedIgnoresSettersButNotKeyHandlerSelects(t *testing.T) {
	cols, rows := numberedRows(4, 3)
	m := New(cols, rows)
	m.SelectRow(2)
	m.SelectColumn(1)
	if _, cmd := m.Update(featureOtherMsg{}); cmd != nil {
		t.Fatal("setters made the next Update report a change")
	}
	// A KeyHandler that moves the selection during Update is a change made
	// because of Update: it is reported.
	m.SetKeyHandler(func(m *Model, msg tea.KeyPressMsg) (tea.Cmd, bool) {
		if msg.String() == "g" {
			m.SelectRow(0)
			return nil, true
		}
		return nil, false
	})
	_, cmd := m.Update(uitest.Key("g"))
	if got := selectionMsgs(cmd); len(got) != 1 || got[0].Index != 0 {
		t.Fatalf("key handler select emitted %+v", got)
	}
}

func TestSelectionChangedOnFilterAndSort(t *testing.T) {
	cols, rows := numberedRows(3, 1)
	rows[0].Values, rows[1].Values, rows[2].Values = []any{"b"}, []any{"c"}, []any{"a"}
	m := New(cols, rows)
	m.SetFocused(true)
	_, cmd := m.Update(uitest.Key("s")) // same highlighted position, a different row underneath
	got := selectionMsgs(cmd)
	if len(got) != 1 || got[0].Index != 0 || got[0].Row.Key != "2" {
		t.Fatalf("sort emitted %+v, want the row now on top", got)
	}
	// Filtering "c" leaves a single visible row: the highlight moves to it.
	m.Update(uitest.Text("/"))
	var last []SelectionChangedMsg
	for _, r := range "c" {
		_, cmd = m.Update(uitest.Text(string(r)))
		last = append(last, selectionMsgs(cmd)...)
	}
	if len(last) == 0 || last[len(last)-1].Row.Key != "1" {
		t.Fatalf("filter typing emitted %+v", last)
	}
}

func TestSelectionChangedWhenTheGridBecomesEmpty(t *testing.T) {
	cols, rows := numberedRows(2, 1)
	m := New(cols, rows)
	m.SetFocused(true)
	m.Update(uitest.Text("/"))
	_, cmd := m.Update(uitest.Text("z")) // matches nothing
	got := selectionMsgs(cmd)
	if len(got) != 1 || got[0].Index != -1 || got[0].Row.Key != "" {
		t.Fatalf("emptying filter emitted %+v, want Index -1 and the zero Row", got)
	}
}

func TestRowActivatedCarriesColumnAndID(t *testing.T) {
	cols, rows := numberedRows(2, 3)
	m := New(cols, rows, WithID("g1"))
	m.SelectColumn(2)
	_, cmd := m.Update(uitest.Key("enter"))
	msgs := uitest.Msgs(cmd)
	if len(msgs) != 1 {
		t.Fatalf("enter emitted %+v", msgs)
	}
	activated, ok := msgs[0].(RowActivatedMsg)
	if !ok || activated.Column != 2 || activated.ID != "g1" || activated.Row.Key != "0" {
		t.Fatalf("activated = %+v", msgs[0])
	}
}

func TestPinRowCarriesRefAndID(t *testing.T) {
	cols, rows := numberedRows(2, 1)
	m := New(cols, rows, WithID("g1"))
	_, cmd := m.Update(uitest.Text("+"))
	msgs := uitest.Msgs(cmd)
	pin, ok := msgs[0].(PinRowMsg)
	if len(msgs) != 1 || !ok || pin.ID != "g1" || pin.Ref != (testRef{Type: "row", ID: "0"}) {
		t.Fatalf("pin = %+v", msgs)
	}
	// No Ref, and a typed nil pointer, both mean "nothing to pin".
	var typedNil *testRef
	for _, ref := range []any{nil, typedNil, map[string]string(nil), []int(nil)} {
		rows[0].Ref = ref
		m = New(cols, rows)
		if _, cmd = m.Update(uitest.Text("+")); cmd != nil {
			t.Fatalf("+ on a row with Ref %#v returned a command", ref)
		}
		if m.Current() != nil {
			t.Fatalf("Current() = %#v for Ref %#v, want plain nil", m.Current(), ref)
		}
	}
	rows[0].Ref = &testRef{Type: "ptr", ID: "1"}
	if got := New(cols, rows).Current(); got != rows[0].Ref {
		t.Fatalf("a non-nil pointer Ref was dropped: %#v", got)
	}
	rows[0].Ref = "plain-string-ref"
	if New(cols, rows).Current() != "plain-string-ref" {
		t.Fatal("a value Ref was dropped")
	}
}

func TestHomeAndEndJumpInASliceGrid(t *testing.T) {
	cols, rows := numberedRows(30, 2)
	m := New(cols, rows)
	m.Update(uitest.Key("end"))
	if m.CurrentIndex() != 29 {
		t.Fatalf("end -> %d", m.CurrentIndex())
	}
	m.Update(uitest.Key("home"))
	if m.CurrentIndex() != 0 {
		t.Fatalf("home -> %d", m.CurrentIndex())
	}
	empty := New(cols, nil)
	empty.Update(uitest.Key("end")) // nothing to move: must not panic
	if empty.CurrentIndex() != -1 {
		t.Fatal("empty grid has a current row")
	}
}

func TestFixedColumnsStayVisibleWhileTheRestScrolls(t *testing.T) {
	cols, rows := numberedRows(3, 6)
	for c := range cols {
		cols[c].Name = fmt.Sprintf("column%d", c)
	}
	m := New(cols, rows, WithFixedColumns(1))
	m.SetWidth(40)
	before := ansi.Strip(m.TableView())
	if !strings.Contains(before, "column0") || strings.Contains(before, "column5") {
		t.Fatalf("initial table: %q", before)
	}
	m.SelectColumn(5)
	after := ansi.Strip(m.TableView())
	if !strings.Contains(after, "column0") || !strings.Contains(after, "r0c0") {
		t.Fatalf("the frozen column scrolled away: %q", after)
	}
	if !strings.Contains(after, "column5") || strings.Contains(after, "column1") {
		t.Fatalf("the rest did not scroll: %q", after)
	}
	if m.ColumnOffset() == 0 {
		t.Fatal("ColumnOffset() did not advance")
	}
	if first, last := m.VisibleColumnRange(); first != 1 || last != 6 {
		t.Fatalf("VisibleColumnRange() = %d,%d, want 1,6", first, last)
	}
	// Selecting a frozen column never scrolls; going back leaves it visible.
	m.SelectColumn(0)
	if !strings.Contains(ansi.Strip(m.TableView()), "column0") {
		t.Fatal("frozen column vanished after selecting it")
	}
	m.SelectColumn(1)
	if got := ansi.Strip(m.TableView()); !strings.Contains(got, "column1") || !strings.Contains(got, "column0") {
		t.Fatalf("scrolling back left lost a column: %q", got)
	}
}

func TestFixedColumnsClampedToTheColumnCount(t *testing.T) {
	cols, rows := numberedRows(2, 2)
	m := New(cols, rows, WithFixedColumns(9))
	if m.fixed() != 2 {
		t.Fatalf("fixed() = %d, want it clamped to 2", m.fixed())
	}
	m = New(cols, rows, WithFixedColumns(-3))
	if m.fixed() != 0 {
		t.Fatalf("fixed() = %d, want 0", m.fixed())
	}
	m.SelectColumn(1)
	if !strings.Contains(ansi.Strip(m.View(30, true)), "r0c1") {
		t.Fatal("grid without fixed columns broke")
	}
}

func TestFixedColumnsDoNotFitDegradesGracefully(t *testing.T) {
	cols, rows := numberedRows(2, 4)
	m := New(cols, rows, WithFixedColumns(3))
	m.SetWidth(8) // narrower than the frozen columns alone
	if first, last := m.VisibleColumnRange(); first != 0 || last != 0 {
		t.Fatalf("VisibleColumnRange() = %d,%d, want 0,0", first, last)
	}
	_ = m.View(8, true) // must not panic
}

func TestStyleHookStylesCells(t *testing.T) {
	cols := []Column{{Name: "n", Numeric: true}, {Name: "fk"}, {Name: "note"}}
	rows := []Row{
		{Key: "0", Values: []any{1, "abc", Absent}},
		{Key: "1", Values: []any{2, "def", "ok"}},
	}
	var calls []string
	m := New(cols, rows, WithCellStyle(func(row Row, column int, value any) lipgloss.Style {
		calls = append(calls, fmt.Sprintf("%s/%d", row.Key, column))
		if column == 1 {
			return lipgloss.NewStyle().Underline(true)
		}
		return lipgloss.NewStyle()
	}))
	if len(calls) < 6 {
		t.Fatalf("hook called %d times, want once per cell: %v", len(calls), calls)
	}
	m.SetWidth(40)
	view := m.TableView()
	if !underlineOn(view, "d") {
		t.Fatalf("foreign-key cell is not underlined: %q", view)
	}
	if underlineOn(view, "o") {
		t.Fatalf("plain cell got underlined: %q", view)
	}
	plain := ansi.Strip(view)
	if !strings.Contains(plain, "abc") || !strings.Contains(plain, "ok") {
		t.Fatalf("cell text lost: %q", plain)
	}
}

func TestStyleHookReceivesTheRawValue(t *testing.T) {
	var got []any
	cols := []Column{{Name: "a"}}
	New(cols, []Row{{Values: []any{42}}, {Values: nil}}, WithCellStyle(func(_ Row, _ int, value any) lipgloss.Style {
		got = append(got, value)
		return lipgloss.NewStyle()
	}))
	if len(got) < 2 || got[0] != 42 || got[1] != Absent {
		t.Fatalf("hook values = %#v, want the raw 42 then Absent", got)
	}
}

// underlineOn reports whether text is rendered with the underline attribute
// in styled output (an SGR sequence containing 4 right before the text).
func underlineOn(styled, text string) bool {
	i := strings.Index(styled, text)
	if i < 0 {
		return false
	}
	j := strings.LastIndex(styled[:i], "\x1b[")
	if j < 0 {
		return false
	}
	seq := styled[j:i]
	return strings.Contains(seq, ";4;") || strings.Contains(seq, "[4;") || strings.Contains(seq, ";4m") || strings.HasSuffix(seq, ";4")
}

func TestWithoutFrameRendersTableAndFooterOnly(t *testing.T) {
	cols, rows := numberedRows(3, 2)
	m := New(cols, rows, WithTitle("Secret title"), WithoutFrame())
	lines := strings.Split(ansi.Strip(m.View(40, true)), "\n")
	if len(lines) != 1+3+1 {
		t.Fatalf("frameless view has %d lines, want header + 3 rows + footer: %q", len(lines), lines)
	}
	joined := strings.Join(lines, "\n")
	for _, border := range []string{"╭", "╰", "●", "○", "▐", "Secret title"} {
		if strings.Contains(joined, border) {
			t.Fatalf("frameless view contains %q: %q", border, joined)
		}
	}
	if !strings.HasPrefix(lines[0], "col0") && !strings.Contains(lines[0], "col0") {
		t.Fatalf("first line is not the column header: %q", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], "Rows 1–3 of 3 returned") {
		t.Fatalf("last line is not the footer: %q", lines[len(lines)-1])
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w != 40 {
			t.Fatalf("line width %d != 40: %q", w, line)
		}
	}
	if m.NaturalWidth() >= New(cols, rows).NaturalWidth() {
		t.Fatal("frameless NaturalWidth() should be smaller: no border columns")
	}
}

func TestWithoutFrameKeepsSwitcherOnlyWithExtraViews(t *testing.T) {
	cols, rows := numberedRows(2, 2)
	with := New(cols, rows, WithoutFrame(), WithExtraViews(CardView("")))
	lines := strings.Split(ansi.Strip(with.View(70, true)), "\n")
	if len(lines) != 1+1+2+1 || !strings.Contains(lines[0], "1 Table") || !strings.Contains(lines[0], "2 Current row") {
		t.Fatalf("switcher line missing: %q", lines)
	}
	hidden := New(cols, rows, WithoutFrame(), WithExtraViews(CardView("")), WithoutSwitcher())
	if got := strings.Split(ansi.Strip(hidden.View(60, true)), "\n"); len(got) != 4 {
		t.Fatalf("WithoutSwitcher frameless view has %d lines: %q", len(got), got)
	}
	// An extra view shows its own body between the switcher and the footer.
	with.ShowView(View(firstExtraView))
	view := ansi.Strip(with.View(70, true))
	if !strings.Contains(view, "r0c0") || !strings.Contains(view, "1 Table") {
		t.Fatalf("card view frameless: %q", view)
	}
}

func TestWithoutFrameToleratesTinyWidth(t *testing.T) {
	cols, rows := numberedRows(2, 2)
	m := New(cols, rows, WithoutFrame())
	if got := m.View(1, false); got == "" {
		t.Fatal("empty view")
	}
}

func TestSetSizeFitsThePageToTheHeight(t *testing.T) {
	cols, rows := numberedRows(50, 2)
	m := New(cols, rows)
	m.SetSize(40, 8) // top border, header, 5 rows, bottom border
	if m.Height() != 8 || m.Width() != 40 {
		t.Fatalf("size = %dx%d", m.Width(), m.Height())
	}
	view := ansi.Strip(m.View(40, true))
	if n := len(strings.Split(view, "\n")); n != 8 {
		t.Fatalf("framed view has %d lines, want 8:\n%s", n, view)
	}
	if !strings.Contains(view, "Rows 1–5 of 50") {
		t.Fatalf("page is not 5 rows: %q", view)
	}
	m.SetSize(40, 13)
	if first, last := m.VisibleIndices(); last-first+1 != 10 {
		t.Fatalf("page after growing = %d..%d", first, last)
	}
	m.SetSize(40, 1) // never below one row
	if first, last := m.VisibleIndices(); last-first+1 != 1 {
		t.Fatalf("page at tiny height = %d..%d", first, last)
	}
	m.SetSize(40, 0) // back to the WithMaxVisibleRows page
	if first, last := m.VisibleIndices(); last-first+1 != DefaultMaxVisibleRows {
		t.Fatalf("page after clearing the height = %d..%d", first, last)
	}
}

func TestSetSizeIsCappedByExplicitMaxVisibleRows(t *testing.T) {
	cols, rows := numberedRows(50, 2)
	m := New(cols, rows, WithMaxVisibleRows(3))
	m.SetSize(40, 30)
	if first, last := m.VisibleIndices(); last-first+1 != 3 {
		t.Fatalf("explicit cap ignored: %d..%d", first, last)
	}
	m.SetSize(40, 5) // fits only 2 rows: the height is the tighter bound
	if first, last := m.VisibleIndices(); last-first+1 != 2 {
		t.Fatalf("fit ignored: %d..%d", first, last)
	}
	// An explicit 0 ("no paging") lets the height decide.
	unpaged := New(cols, rows, WithMaxVisibleRows(0))
	unpaged.SetSize(40, 9)
	if first, last := unpaged.VisibleIndices(); last-first+1 != 6 {
		t.Fatalf("unpaged fit = %d..%d, want 6 rows", first, last)
	}
}

func TestSetSizeFramelessChrome(t *testing.T) {
	cols, rows := numberedRows(50, 2)
	plain := New(cols, rows, WithoutFrame())
	plain.SetSize(40, 8) // header + footer + 6 rows
	view := ansi.Strip(plain.View(40, true))
	if n := len(strings.Split(view, "\n")); n != 8 {
		t.Fatalf("frameless view has %d lines, want 8:\n%s", n, view)
	}
	switcher := New(cols, rows, WithoutFrame(), WithExtraViews(CardView("")))
	switcher.SetSize(40, 8) // switcher + header + footer + 5 rows
	view = ansi.Strip(switcher.View(40, true))
	if n := len(strings.Split(view, "\n")); n != 8 {
		t.Fatalf("frameless view with switcher has %d lines, want 8:\n%s", n, view)
	}
}

func TestSetSizeKeepsTheHighlightedRow(t *testing.T) {
	cols, rows := numberedRows(50, 2)
	m := New(cols, rows)
	m.SelectRow(23)
	m.SetSize(40, 8)
	if m.CurrentIndex() != 23 {
		t.Fatalf("CurrentIndex() after SetSize = %d", m.CurrentIndex())
	}
	if first, last := m.VisibleIndices(); first > 23 || last < 23 {
		t.Fatalf("the highlighted row is off the page: %d..%d", first, last)
	}
}

func TestSetSizeReloadsARowSourcePage(t *testing.T) {
	src := &countingSource{n: 1000}
	m := newSourceGrid(src)
	m.SelectRow(47)
	src.rowCalls = 0
	m.SetSize(50, 9) // 6 rows per page
	if src.rowCalls != 6 {
		t.Fatalf("SetSize loaded %d rows, want one page of 6", src.rowCalls)
	}
	if first, last := m.VisibleIndices(); first != 42 || last != 47 {
		t.Fatalf("page = %d..%d, want 42..47", first, last)
	}
	if n := len(strings.Split(ansi.Strip(m.View(50, true)), "\n")); n != 9 {
		t.Fatalf("view has %d lines, want 9", n)
	}
}

func TestAtEdgeInASliceGrid(t *testing.T) {
	cols, rows := numberedRows(3, 3)
	m := New(cols, rows)
	edges := func() [4]bool {
		return [4]bool{m.AtEdge(widgets.Up), m.AtEdge(widgets.Down), m.AtEdge(widgets.Left), m.AtEdge(widgets.Right)}
	}
	if got := edges(); got != [4]bool{true, false, true, false} {
		t.Fatalf("top-left edges = %v", got)
	}
	m.Update(uitest.Key("down"))
	m.Update(uitest.Key("l"))
	if got := edges(); got != [4]bool{false, false, false, false} {
		t.Fatalf("middle edges = %v", got)
	}
	m.Update(uitest.Key("end"))
	m.Update(uitest.Key("right"))
	if got := edges(); got != [4]bool{false, true, false, true} {
		t.Fatalf("bottom-right edges = %v", got)
	}
	if m.AtEdge(widgets.Direction(99)) {
		t.Fatal("an unknown direction is an edge")
	}
	empty := New(cols, nil)
	if !empty.AtEdge(widgets.Up) || !empty.AtEdge(widgets.Down) {
		t.Fatal("an empty grid is at every vertical edge")
	}
	none := New(nil, nil)
	if !none.AtEdge(widgets.Left) || !none.AtEdge(widgets.Right) {
		t.Fatal("a grid without columns is at every horizontal edge")
	}
}

func TestAtEdgeInARowSourceGrid(t *testing.T) {
	m := newSourceGrid(&countingSource{n: 25})
	if !m.AtEdge(widgets.Up) || m.AtEdge(widgets.Down) {
		t.Fatal("first row of a source grid")
	}
	m.Update(uitest.Key("down"))
	if m.AtEdge(widgets.Up) || m.AtEdge(widgets.Down) {
		t.Fatal("middle of a source grid")
	}
	m.Update(uitest.Key("end"))
	if m.AtEdge(widgets.Up) || !m.AtEdge(widgets.Down) {
		t.Fatal("last row of a source grid")
	}
	if !newSourceGrid(&countingSource{n: 0}).AtEdge(widgets.Down) {
		t.Fatal("empty source is at the bottom edge")
	}
}

func TestAtEdgeWhileASecondaryViewHasFocus(t *testing.T) {
	cols, rows := numberedRows(3, 3)
	m := New(cols, rows, WithExtraViews(CardView("")))
	m.Update(uitest.Key("2"))
	if !m.SecondaryFocus() {
		t.Fatal("no secondary focus")
	}
	if m.AtEdge(widgets.Up) || m.AtEdge(widgets.Down) || !m.AtEdge(widgets.Left) || !m.AtEdge(widgets.Right) {
		t.Fatal("secondary focus keeps up/down for the view and releases left/right")
	}
}

func TestEditingWhileTheFilterInputIsFocused(t *testing.T) {
	cols, rows := numberedRows(3, 1)
	m := New(cols, rows)
	m.SetFocused(true)
	if m.Editing() {
		t.Fatal("editing before the filter opened")
	}
	m.Update(uitest.Text("/"))
	if !m.Editing() || !m.CapturesEsc() {
		t.Fatal("not editing while the filter input is focused")
	}
	m.Update(uitest.Key("esc"))
	if m.Editing() {
		t.Fatal("still editing after esc")
	}
}

// featureOtherMsg is a message the grid does not understand.
type featureOtherMsg struct{}

func TestFooterShowsTheActiveFilter(t *testing.T) {
	cols, rows := numberedRows(5, 2)
	m := New(cols, rows)
	m.SetFocused(true)
	if strings.Contains(m.Footer(), "•  /") || strings.Contains(m.Footer(), "/") {
		t.Fatalf("filter shown before it was used: %q", m.Footer())
	}
	m.Update(uitest.Text("/"))
	m.Update(uitest.Text("r"))
	m.Update(uitest.Text("2"))
	if got := m.Footer(); !strings.HasSuffix(got, "• /r2▏") {
		t.Fatalf("Footer() while typing = %q", got)
	}
	m.Update(uitest.Key("enter")) // blur the input, keep the filter
	if got := m.Footer(); !strings.HasSuffix(got, "• /r2") {
		t.Fatalf("Footer() with a blurred filter = %q", got)
	}
}

func TestColumnMaxWidthCapsAColumn(t *testing.T) {
	long := strings.Repeat("x", 50)
	cols := []Column{{Name: "a", MaxWidth: 12}, {Name: "b"}}
	m := New(cols, []Row{{Values: []any{long, long}}})
	m.SetWidth(100)
	if got := m.columnWidthFor(100, 0); got != 12 {
		t.Fatalf("capped column width = %d, want 12", got)
	}
	if got := m.columnWidthFor(100, 1); got != DefaultMaxColumnWidth {
		t.Fatalf("default column width = %d, want %d", got, DefaultMaxColumnWidth)
	}
	if want := m.chromeWidth() + 12 + 1 + DefaultMaxColumnWidth; m.NaturalWidth() != want {
		t.Fatalf("NaturalWidth() = %d, want %d", m.NaturalWidth(), want)
	}
}

func TestSetDataReplacesRowsAndResetsState(t *testing.T) {
	cols, rows := numberedRows(5, 3)
	m := New(cols, rows, WithFixedColumns(1))
	m.SetFocused(true)
	m.Sort(1)
	m.SelectRow(3)
	m.SelectColumn(2)

	loaded := []Row{{Key: "a", Values: []any{"x0", "x1"}}, {Key: "b", Values: []any{"y0", "y1"}}}
	m.SetData([]Column{{Name: "p"}, {Name: "q"}}, loaded)
	if col, _ := m.SortState(); col != -1 {
		t.Fatalf("sort state kept: %d", col)
	}
	if m.CurrentIndex() != 0 || len(m.Rows()) != 2 || len(m.Columns()) != 2 {
		t.Fatalf("state after SetData: index %d rows %d columns %d", m.CurrentIndex(), len(m.Rows()), len(m.Columns()))
	}
	if m.SelectedColumn() != 1 {
		t.Fatalf("selected column = %d, want it clamped to 1", m.SelectedColumn())
	}
	if view := ansi.Strip(m.View(40, true)); !strings.Contains(view, "x0") || !strings.Contains(view, "Rows 1–2 of 2") {
		t.Fatalf("view after SetData: %q", view)
	}
	if m.fixed() != 1 {
		t.Fatalf("fixed() = %d", m.fixed())
	}
	// A shrinking column set re-clamps the frozen columns, an empty one works.
	m.SetData([]Column{{Name: "only"}}, nil)
	if m.fixed() != 1 || m.CurrentIndex() != -1 || m.SelectedColumn() != 0 {
		t.Fatalf("after emptying: fixed %d index %d column %d", m.fixed(), m.CurrentIndex(), m.SelectedColumn())
	}
	m.SetData(nil, nil)
	if m.SelectedColumn() != 0 || m.fixed() != 0 {
		t.Fatal("no columns")
	}
	if _, cmd := m.Update(featureOtherMsg{}); cmd != nil {
		t.Fatal("SetData made the next Update report a change")
	}
}

func TestSetDataDropsAnActiveFilterAndSwitchesOutOfSourceMode(t *testing.T) {
	m := newSourceGrid(&countingSource{n: 500})
	cols, rows := numberedRows(3, 2)
	m.SetData(cols, rows)
	if m.count() != 3 || m.src != nil {
		t.Fatalf("SetData did not switch to a slice: count %d", m.count())
	}
	m.SetFocused(true)
	m.Update(uitest.Text("/")) // the filter exists again
	if !m.Editing() {
		t.Fatal("filter unavailable after leaving source mode")
	}
	m.SetData(cols, rows)
	if m.Editing() || m.table.GetCurrentFilter() != "" {
		t.Fatal("SetData kept the filter")
	}
}

func TestSetRowSourceLoadsTheFirstPageOfTheNewSource(t *testing.T) {
	cols, rows := numberedRows(3, 3)
	m := New(cols, rows)
	src := &countingSource{n: 1_000_000}
	m.SetRowSource(src, sourceColumns())
	if src.rowCalls != DefaultMaxVisibleRows {
		t.Fatalf("Row calls = %d, want one page", src.rowCalls)
	}
	if m.CurrentIndex() != 0 || m.count() != 1_000_000 {
		t.Fatalf("index %d count %d", m.CurrentIndex(), m.count())
	}
	next := &countingSource{n: 20}
	m.SelectRow(15)
	m.SetRowSource(next, sourceColumns())
	if m.CurrentIndex() != 0 {
		t.Fatalf("highlight kept across sources: %d", m.CurrentIndex())
	}
}

func TestRowSelectionGridHasNoColumnSelection(t *testing.T) {
	cols, rows := numberedRows(3, 3)
	m := New(cols, rows, WithRowSelection())
	for _, key := range []string{"l", "right", "h", "left"} {
		if _, cmd := m.Update(uitest.Key(key)); cmd != nil {
			t.Fatalf("%q returned a command", key)
		}
	}
	m.SelectColumn(2)
	if m.SelectedColumn() != 0 {
		t.Fatalf("SelectedColumn() = %d, want 0", m.SelectedColumn())
	}
	_, cmd := m.Update(uitest.Key("down"))
	if got := selectionMsgs(cmd); len(got) != 1 || got[0].Column != 0 || got[0].Index != 1 {
		t.Fatalf("down emitted %+v", got)
	}
	if !m.AtEdge(widgets.Left) || !m.AtEdge(widgets.Right) {
		t.Fatal("a row list is at both horizontal edges")
	}
	_, cmd = m.Update(uitest.Key("enter"))
	if activated := uitest.Msgs(cmd)[0].(RowActivatedMsg); activated.Column != 0 || activated.Row.Key != "1" {
		t.Fatalf("activated = %+v", activated)
	}
	// No selected-column emphasis: the accent the default grid puts on the
	// selected column is absent from the row list.
	const accent = "38;2;232;194;90"
	if !strings.Contains(New(cols, rows).TableView(), accent) {
		t.Fatal("test premise: the default grid should show the selected-column accent")
	}
	if strings.Contains(m.TableView(), accent) {
		t.Fatal("a row list shows the selected-column accent")
	}
}

func TestFilterOnTypeNarrowsTheListWhileTyping(t *testing.T) {
	cols := []Column{{Name: "name"}}
	rows := []Row{{Key: "0", Values: []any{"users"}}, {Key: "1", Values: []any{"orders"}}, {Key: "2", Values: []any{"order_items"}}}
	m := New(cols, rows, WithRowSelection(), WithFilterOnType())
	m.SetFocused(true)
	for _, r := range "ord" {
		m.Update(uitest.Text(string(r)))
	}
	if !m.Editing() {
		t.Fatal("typing did not open the filter")
	}
	if first, last := m.VisibleIndices(); last-first+1 != 2 {
		t.Fatalf("filter left %d..%d visible, want the 2 order tables", first, last)
	}
	if got := m.CurrentIndex(); got != 1 {
		t.Fatalf("CurrentIndex() = %d, want 1 (the first match)", got)
	}
	// "s" and "l" are typed text now, not sort/column keys.
	m.Update(uitest.Text("e"))
	if col, _ := m.SortState(); col != -1 {
		t.Fatal("typing sorted the grid")
	}
	m.Update(uitest.Key("enter")) // blur, keep filter
	if m.Editing() {
		t.Fatal("still editing after enter")
	}
	// A modified key press, and arrows, stay the grid's own.
	m.Update(tea.KeyPressMsg{Text: "x", Code: 'x', Mod: tea.ModAlt})
	if m.Editing() {
		t.Fatal("alt+x opened the filter")
	}
}

func TestFilterOnTypeIsInertWithoutAFilter(t *testing.T) {
	cols, rows := numberedRows(3, 1)
	for name, m := range map[string]*Model{
		"disabled": New(cols, rows, WithFilterOnType(), WithFilterDisabled()),
		"source":   newSourceGrid(&countingSource{n: 30}, WithFilterOnType()),
	} {
		m.SetFocused(true)
		m.Update(uitest.Text("z"))
		if m.Editing() {
			t.Fatalf("%s: typing opened a filter that should not exist", name)
		}
	}
	// Secondary focus keeps its keys.
	m := New(cols, rows, WithFilterOnType(), WithExtraViews(ExtraView{Label: "X", Render: func(*Model, int, int) string { return "" }}))
	m.SetFocused(true)
	m.ShowView(View(firstExtraView))
	m.Update(uitest.Text("z"))
	if m.Editing() {
		t.Fatal("typing opened the filter behind a secondary view")
	}
}
