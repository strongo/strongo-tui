package grid

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/strongo/strongo-tui/pkg/uitest"
)

// countingSource is a synthetic RowSource that counts how often the grid
// asks it for rows, standing in for a database cursor with n records.
type countingSource struct {
	n        int
	rowCalls int
	lenCalls int
	sorted   []string // SortBy calls, "column:desc"
}

func (s *countingSource) Len() int { s.lenCalls++; return s.n }

func (s *countingSource) Row(i int) Row {
	s.rowCalls++
	return Row{
		Key:    fmt.Sprint(i),
		Values: []any{i, "name-" + fmt.Sprint(i), strings.Repeat("w", i%7)},
		Ref:    testRef{Type: "rec", ID: fmt.Sprint(i)},
	}
}

// sortingSource adds the optional Sorter interface; a descending sort reverses
// the row order.
type sortingSource struct {
	countingSource
	desc bool
}

func (s *sortingSource) SortBy(column int, desc bool) {
	s.sorted = append(s.sorted, fmt.Sprintf("%d:%v", column, desc))
	s.desc = desc
}

func (s *sortingSource) Row(i int) Row {
	if s.desc {
		i = s.n - 1 - i
	}
	return s.countingSource.Row(i)
}

func sourceColumns() []Column {
	return []Column{{Name: "id", Numeric: true}, {Name: "name"}, {Name: "w"}}
}

func newSourceGrid(src RowSource, opts ...Option) *Model {
	return New(nil, nil, append([]Option{WithRowSource(src, sourceColumns())}, opts...)...)
}

func TestRowSourceMaterialisesOnlyTheVisiblePage(t *testing.T) {
	src := &countingSource{n: 1_000_000}
	m := newSourceGrid(src)
	page := DefaultMaxVisibleRows
	if src.rowCalls > page {
		t.Fatalf("construction called Row %d times, want at most one page (%d)", src.rowCalls, page)
	}

	src.rowCalls = 0
	view := ansi.Strip(m.View(60, true))
	if src.rowCalls != 0 {
		t.Fatalf("View called Row %d times, want 0 (pages load in Update/setters, View is pure)", src.rowCalls)
	}
	if !strings.Contains(view, "Rows 1–10 of 1000000 returned") {
		t.Fatalf("footer missing the total: %q", view)
	}
	if !strings.Contains(view, "name-0") || !strings.Contains(view, "name-9") || strings.Contains(view, "name-10") {
		t.Fatalf("view is not exactly page 1: %q", view)
	}

	// Moving inside the loaded page needs no source access at all.
	m.Update(uitest.Key("down"))
	if src.rowCalls != 0 {
		t.Fatalf("a row move inside the page called Row %d times, want 0", src.rowCalls)
	}
	if m.CurrentIndex() != 1 {
		t.Fatalf("CurrentIndex() = %d, want 1", m.CurrentIndex())
	}

	// Every page-crossing or far jump loads exactly one page.
	for _, key := range []string{"pgdown", "pgdown", "end", "pgup", "home", "end", "home"} {
		src.rowCalls = 0
		m.Update(uitest.Key(key))
		if src.rowCalls > page {
			t.Fatalf("%q called Row %d times, want at most %d", key, src.rowCalls, page)
		}
		m.View(60, true)
		if src.rowCalls > page {
			t.Fatalf("View after %q called Row %d times", key, src.rowCalls)
		}
	}
}

func TestRowSourceNavigationKeys(t *testing.T) {
	src := &countingSource{n: 95}
	m := newSourceGrid(src)
	steps := []struct {
		key  string
		want int
	}{
		{"down", 1}, {"j", 2}, {"up", 1}, {"k", 0}, {"up", 0},
		{"pgdown", 10}, {"pgdown", 20}, {"pgup", 10}, {"end", 94}, {"down", 94},
		{"pgdown", 94}, {"home", 0},
	}
	for _, step := range steps {
		m.Update(uitest.Key(step.key))
		if got := m.CurrentIndex(); got != step.want {
			t.Fatalf("after %q CurrentIndex() = %d, want %d", step.key, got, step.want)
		}
		row, ok := m.CurrentRow()
		if !ok || row.Key != fmt.Sprint(step.want) {
			t.Fatalf("after %q CurrentRow() = %+v, %v; want key %d", step.key, row, ok, step.want)
		}
	}
	m.Update(uitest.Key("end"))
	if first, last := m.VisibleIndices(); first != 90 || last != 94 {
		t.Fatalf("VisibleIndices() on the last partial page = %d,%d, want 90,94", first, last)
	}
	if !strings.Contains(ansi.Strip(m.View(60, true)), "Rows 91–95 of 95 returned") {
		t.Fatal("footer of the last partial page is wrong")
	}
}

func TestRowSourceAccessors(t *testing.T) {
	src := &countingSource{n: 40}
	m := newSourceGrid(src)
	if got := len(m.Rows()); got != DefaultMaxVisibleRows {
		t.Fatalf("Rows() returned %d rows, want only the loaded page (%d)", got, DefaultMaxVisibleRows)
	}
	if got := m.Cell(33, 1); got != "name-33" {
		t.Fatalf("Cell(33,1) = %q", got)
	}
	for _, bad := range [][2]int{{-1, 0}, {40, 0}, {0, -1}, {0, 3}} {
		if got := m.Cell(bad[0], bad[1]); got != "" {
			t.Fatalf("Cell(%d,%d) = %q, want empty", bad[0], bad[1], got)
		}
	}
	if got := m.IndexForKey("3"); got != -1 {
		t.Fatalf("IndexForKey in source mode = %d, want -1", got)
	}
	if ref := m.Current(); ref != (testRef{Type: "rec", ID: "0"}) {
		t.Fatalf("Current() = %v", ref)
	}
	m.SelectRow(25)
	if m.CurrentIndex() != 25 || m.Rows()[0].Key != "20" {
		t.Fatalf("SelectRow(25): index %d, first loaded row %q", m.CurrentIndex(), m.Rows()[0].Key)
	}
	m.SelectRow(1000)
	if m.CurrentIndex() != 39 {
		t.Fatalf("SelectRow past the end = %d, want 39", m.CurrentIndex())
	}
	m.SelectRow(-5)
	if m.CurrentIndex() != 0 {
		t.Fatalf("SelectRow before the start = %d, want 0", m.CurrentIndex())
	}
	if !strings.Contains(m.TableView(), "name-0") {
		t.Fatal("TableView misses the first page")
	}
}

func TestRowSourceEmpty(t *testing.T) {
	src := &countingSource{n: 0}
	m := newSourceGrid(src)
	if m.CurrentIndex() != -1 {
		t.Fatalf("CurrentIndex() = %d, want -1", m.CurrentIndex())
	}
	if _, ok := m.CurrentRow(); ok {
		t.Fatal("empty source has a current row")
	}
	if m.Current() != nil {
		t.Fatal("empty source has a current Ref")
	}
	m.SelectRow(3)
	for _, key := range []string{"down", "end", "pgdown", "enter", "+"} {
		if _, cmd := m.Update(uitest.Key(key)); cmd != nil {
			t.Fatalf("%q on an empty source returned a command", key)
		}
	}
	if got := m.Footer(); got != "No rows returned." {
		t.Fatalf("Footer() = %q", got)
	}
	if src.rowCalls != 0 {
		t.Fatalf("empty source got %d Row calls", src.rowCalls)
	}
}

func TestRowSourceRefreshClampsAfterShrink(t *testing.T) {
	src := &countingSource{n: 100}
	m := newSourceGrid(src)
	m.SelectRow(95)
	src.n = 12
	m.Refresh()
	if m.CurrentIndex() != 11 {
		t.Fatalf("CurrentIndex() after the source shrank = %d, want 11", m.CurrentIndex())
	}
	if first, last := m.VisibleIndices(); first != 10 || last != 11 {
		t.Fatalf("VisibleIndices() = %d,%d, want 10,11", first, last)
	}
	slice := New(sourceColumns(), []Row{{Values: []any{1, "a", "b"}}})
	slice.Refresh() // no-op for a slice-backed grid
	if slice.CurrentIndex() != 0 {
		t.Fatal("Refresh disturbed a slice-backed grid")
	}
}

func TestRowSourceSortNeedsASorter(t *testing.T) {
	plain := newSourceGrid(&countingSource{n: 30})
	plain.Update(uitest.Key("s"))
	if column, _ := plain.SortState(); column != -1 {
		t.Fatalf("sort state changed for a source without Sorter: %d", column)
	}

	src := &sortingSource{countingSource: countingSource{n: 30}}
	m := newSourceGrid(src)
	m.Update(uitest.Key("s"))
	m.Update(uitest.Key("s"))
	m.Sort(2)
	if got := strings.Join(src.sorted, " "); got != "0:false 0:true 2:false" {
		t.Fatalf("SortBy calls = %q", got)
	}
	m.Sort(0)
	m.Sort(0)
	if row, _ := m.CurrentRow(); row.Key != "29" {
		t.Fatalf("after a descending sort the first row is %q, want 29", row.Key)
	}
	if !strings.Contains(m.Footer(), "sort id") {
		t.Fatalf("footer lacks the sort indicator: %q", m.Footer())
	}
}

func TestRowSourceFilterIsDisabled(t *testing.T) {
	m := newSourceGrid(&countingSource{n: 30})
	m.SetFocused(true)
	m.Update(uitest.Text("/"))
	if m.Editing() || m.CapturesEsc() {
		t.Fatal("the filter opened on a RowSource grid")
	}
}

func TestRowSourceZeroMaxVisibleRowsStillPages(t *testing.T) {
	src := &countingSource{n: 500}
	m := newSourceGrid(src, WithMaxVisibleRows(0))
	if src.rowCalls != DefaultMaxVisibleRows {
		t.Fatalf("Row calls = %d, want the default page (%d)", src.rowCalls, DefaultMaxVisibleRows)
	}
	if m.paneHeight() != DefaultMaxVisibleRows {
		t.Fatalf("paneHeight() = %d", m.paneHeight())
	}
}

func TestRowSourceColumnWidthsNeverShrink(t *testing.T) {
	src := &countingSource{n: 40}
	m := newSourceGrid(src)
	wide := m.columnWidthFor(80, 2) // the "w" column grows with i%7 within a page
	m.SelectRow(21)
	if got := m.columnWidthFor(80, 2); got < wide {
		t.Fatalf("column width shrank from %d to %d while scrolling", wide, got)
	}
	if m.NaturalWidth() < 2 {
		t.Fatal("NaturalWidth() is empty")
	}
}

func TestRowSourceCardViewShowsCurrentRow(t *testing.T) {
	m := newSourceGrid(&countingSource{n: 300}, WithExtraViews(CardView("")))
	m.SelectRow(205)
	m.ShowView(View(firstExtraView))
	if got := m.ActiveViewContent(40, 10); !strings.Contains(got, "name-205") {
		t.Fatalf("card view = %q", got)
	}
}

func TestRowSourceSecondaryFocusKeepsNavigationKeys(t *testing.T) {
	m := newSourceGrid(&countingSource{n: 300}, WithExtraViews(ExtraView{Label: "X", Render: func(*Model, int, int) string { return "" }}))
	m.ShowView(View(firstExtraView))
	for _, key := range []string{"pgup", "pgdown", "home", "end"} {
		m.Update(uitest.Key(key))
		if m.CurrentIndex() != 0 {
			t.Fatalf("%q moved the table while a secondary view held focus", key)
		}
	}
}

func TestRowSourceSplitLayoutRendersPrimaryPane(t *testing.T) {
	m := newSourceGrid(&countingSource{n: 300}, WithExtraViews(CardView("")),
		WithSplitLayout(func(total, natural int, v View) SplitLayout {
			return SplitLayout{Split: true, PrimaryWidth: total / 2, SecondaryWidth: total - total/2}
		}))
	m.SelectRow(42)
	m.ShowView(View(firstExtraView))
	view := ansi.Strip(m.View(80, true))
	if !strings.Contains(view, "name-42") {
		t.Fatalf("split view lost the highlighted row: %q", view)
	}
}

func BenchmarkRowSourceScroll(b *testing.B) {
	src := &countingSource{n: 1_000_000}
	m := newSourceGrid(src)
	down, pgdown := uitest.Key("down"), uitest.Key("pgdown")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%20 == 0 {
			m.Update(pgdown)
		} else {
			m.Update(down)
		}
		_ = m.View(100, true)
	}
	b.ReportMetric(float64(src.rowCalls)/float64(b.N), "Row-calls/op")
}

func BenchmarkRowSourceConstruct(b *testing.B) {
	src := &countingSource{n: 1_000_000}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = newSourceGrid(src)
	}
}
