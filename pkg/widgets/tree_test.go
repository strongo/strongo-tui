package widgets_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/strongo/strongo-tui/pkg/theme"
	"github.com/strongo/strongo-tui/pkg/uitest"
	"github.com/strongo/strongo-tui/pkg/widgets"
)

func treeSample() []widgets.TreeNode {
	return []widgets.TreeNode{
		{ID: "a", Text: "Alpha", Children: []widgets.TreeNode{
			{ID: "a1", Text: "A one"},
			{ID: "a2", Text: "A two", Children: []widgets.TreeNode{{ID: "a2x", Text: "deep"}}},
		}},
		{ID: "b", Text: "Beta"},
	}
}

func newTestTree(h int) widgets.Tree {
	t := widgets.NewTree("tree", treeSample()...)
	t.SetSize(20, h)
	t.Focus()
	return t
}

func treeCursor(t widgets.Tree) string {
	n, _ := t.Current()
	return n.ID
}

func treePress(t widgets.Tree, keys ...string) (widgets.Tree, []tea.Msg) {
	var all []tea.Msg
	for _, k := range keys {
		var cmd tea.Cmd
		t, cmd = t.Update(uitest.Key(k))
		all = append(all, uitest.Msgs(cmd)...)
	}
	return t, all
}

func TestTreeRenderingGuidesAndMarkers(t *testing.T) {
	tr := newTestTree(6)
	if got, want := uitest.Plain(tr.View()), "▸ Alpha\n  Beta\n\n\n\n"; got != want {
		t.Fatalf("collapsed:\n%q\nwant\n%q", got, want)
	}
	tr.ExpandAll()
	want := "▾ Alpha\n├─   A one\n└─ ▾ A two\n   └─   deep\n  Beta\n"
	if got := uitest.Plain(tr.View()); got != want {
		t.Fatalf("expanded:\n%q\nwant\n%q", got, want)
	}
	for _, l := range strings.Split(tr.View(), "\n") {
		if w := lipgloss.Width(l); w != 20 {
			t.Fatalf("width %d", w)
		}
	}
	if len(strings.Split(tr.View(), "\n")) != 6 {
		t.Fatal("height")
	}
	tr.SetSize(0, 5)
	if tr.View() != "" {
		t.Fatal("zero size renders nothing")
	}
}

func TestTreeGuidesNestedNotLast(t *testing.T) {
	tr := widgets.NewTree("t", widgets.TreeNode{ID: "r", Text: "r", Children: []widgets.TreeNode{
		{ID: "c1", Text: "c1", Children: []widgets.TreeNode{{ID: "g", Text: "g"}}},
		{ID: "c2", Text: "c2"},
	}})
	tr.SetSize(20, 5)
	tr.ExpandAll()
	want := "▾ r\n├─ ▾ c1\n│  └─   g\n└─   c2\n"
	if got := uitest.Plain(tr.View()); got != want {
		t.Fatalf("%q want %q", got, want)
	}
}

func TestTreeColorsAndSelectedRow(t *testing.T) {
	tr := widgets.NewTree("t",
		widgets.TreeNode{ID: "a", Text: "\x1b[31mred\x1b[0m"},
		widgets.TreeNode{ID: "b", Text: "blue", Color: theme.LightBlue})
	tr.SetSize(12, 2)
	lines := strings.Split(tr.View(), "\n")
	if strings.Contains(lines[0], "\x1b[31m") {
		t.Fatal("the selected row loses its own styling")
	}
	if !strings.Contains(lines[1], "38;") {
		t.Fatalf("node colour missing: %q", lines[1])
	}
	blurred := lines[0]
	tr.Focus()
	if tr.View() == "" || strings.Split(tr.View(), "\n")[0] == blurred {
		t.Fatal("focused selection looks different")
	}
}

func TestTreeNavigationAndMessages(t *testing.T) {
	tr := newTestTree(10)
	tr.ExpandAll()
	var msgs []tea.Msg
	tr, msgs = treePress(tr, "down")
	if treeCursor(tr) != "a1" || len(msgs) != 1 {
		t.Fatalf("down: %s %v", treeCursor(tr), msgs)
	}
	if m := msgs[0].(widgets.NodeHighlightedMsg); m.ID != "tree" || m.Node.ID != "a1" {
		t.Fatalf("%#v", m)
	}
	tr, _ = treePress(tr, "down", "down", "down")
	if treeCursor(tr) != "b" {
		t.Fatalf("got %s", treeCursor(tr))
	}
	tr, msgs = treePress(tr, "down")
	if len(msgs) != 0 || !tr.AtEdge(widgets.Down) {
		t.Fatal("down at the end does nothing")
	}
	tr, _ = treePress(tr, "home")
	if treeCursor(tr) != "a" || !tr.AtEdge(widgets.Up) {
		t.Fatal("home")
	}
	tr, msgs = treePress(tr, "up", "home")
	if len(msgs) != 0 {
		t.Fatal("no move, no message")
	}
	tr, _ = treePress(tr, "end")
	if treeCursor(tr) != "b" {
		t.Fatal("end")
	}
	tr, _ = treePress(tr, "pgup")
	if treeCursor(tr) != "a" {
		t.Fatalf("pgup %s", treeCursor(tr))
	}
	tr, _ = treePress(tr, "pgdown")
	if treeCursor(tr) != "b" {
		t.Fatalf("pgdown %s", treeCursor(tr))
	}
	small := newTestTree(2)
	small.ExpandAll()
	small, _ = treePress(small, "pgdown")
	if treeCursor(small) != "a2" {
		t.Fatalf("pgdown by a page: %s", treeCursor(small))
	}
	small, _ = treePress(small, "pgup")
	if treeCursor(small) != "a" {
		t.Fatalf("pgup by a page: %s", treeCursor(small))
	}
	// vertical keys that hit no bindings and unfocused trees do nothing
	small.Blur()
	if small.Focused() {
		t.Fatal("blur")
	}
	small2, msgs := treePress(small, "down")
	if treeCursor(small2) != "a" || len(msgs) != 0 {
		t.Fatal("unfocused ignores keys")
	}
	tr, msgs = treePress(tr, "x")
	if len(msgs) != 0 {
		t.Fatal("unbound key")
	}
}

func TestTreeExpandCollapseKeys(t *testing.T) {
	tr := newTestTree(10)
	tr, msgs := treePress(tr, "right")
	if !tr.Expanded("a") || treeCursor(tr) != "a" || len(msgs) != 0 {
		t.Fatal("right expands")
	}
	tr, _ = treePress(tr, "right")
	if treeCursor(tr) != "a1" {
		t.Fatal("right on an expanded node enters it")
	}
	if !tr.AtEdge(widgets.Right) {
		t.Fatal("leaf: right is at the edge")
	}
	tr, msgs = treePress(tr, "right")
	if len(msgs) != 0 {
		t.Fatal("right on a leaf does nothing")
	}
	if tr.AtEdge(widgets.Left) {
		t.Fatal("a child can go left to its parent")
	}
	tr, _ = treePress(tr, "left")
	if treeCursor(tr) != "a" {
		t.Fatal("left goes to the parent")
	}
	if !tr.AtEdge(widgets.Left) {
		t.Fatal("top level: left is at the edge")
	}
	tr, _ = treePress(tr, "left")
	if !tr.Expanded("a") {
		t.Fatal("left on a top-level node does not collapse")
	}
	tr, _ = treePress(tr, "down", "down", "right")
	if treeCursor(tr) != "a2" || !tr.Expanded("a2") {
		t.Fatal("expand a2")
	}
	if tr.AtEdge(widgets.Left) {
		t.Fatal("expanded: left collapses")
	}
	tr, _ = treePress(tr, "left")
	if tr.Expanded("a2") || treeCursor(tr) != "a2" {
		t.Fatal("left collapses")
	}
	if tr.AtEdge(widgets.Right) {
		t.Fatal("collapsed with children: right expands")
	}
}

func TestTreeEnterAndSpace(t *testing.T) {
	tr := newTestTree(10)
	tr, msgs := treePress(tr, "enter")
	if !tr.Expanded("a") || len(msgs) != 1 {
		t.Fatal("enter selects then toggles")
	}
	if m := msgs[0].(widgets.NodeSelectedMsg); m.ID != "tree" || m.Node.ID != "a" || len(m.Node.Children) != 2 {
		t.Fatalf("%#v", m)
	}
	tr, msgs = treePress(tr, "space")
	if tr.Expanded("a") || len(msgs) != 0 {
		t.Fatal("space only toggles")
	}
	tr, _ = treePress(tr, "end")
	tr, msgs = treePress(tr, "enter", "space")
	if len(msgs) != 1 {
		t.Fatal("leaf: enter selects, space is silent")
	}
}

func TestTreeUnselectableAndEmpty(t *testing.T) {
	tr := widgets.NewTree("t",
		widgets.TreeNode{ID: "g", Text: "Group", Unselectable: true, Children: []widgets.TreeNode{
			{ID: "x", Text: "x"}, {ID: "y", Text: "y", Unselectable: true},
		}},
		widgets.TreeNode{ID: "z", Text: "z"})
	tr.SetSize(20, 6)
	tr.Focus()
	if treeCursor(tr) != "z" {
		t.Fatalf("first selectable, got %q", treeCursor(tr))
	}
	if !tr.AtEdge(widgets.Up) {
		t.Fatal("z is the first selectable")
	}
	if tr.Select("g") || tr.Select("nope") || tr.Select("y") {
		t.Fatal("cannot select unselectable or missing nodes")
	}
	if !tr.SetExpanded("g", true) {
		t.Fatal("headers expand")
	}
	if !tr.Select("x") || treeCursor(tr) != "x" {
		t.Fatal("select")
	}
	if !tr.AtEdge(widgets.Left) {
		t.Fatal("parent is a header: nothing to collapse into")
	}
	tr, _ = treePress(tr, "left")
	if treeCursor(tr) != "x" {
		t.Fatal("left is inert under a header")
	}
	tr, _ = treePress(tr, "down")
	if treeCursor(tr) != "z" {
		t.Fatalf("skips the unselectable: %s", treeCursor(tr))
	}
	tr, _ = treePress(tr, "up")
	if treeCursor(tr) != "x" {
		t.Fatal("up skips too")
	}
	// header with only unselectable children: right is at the edge once open
	h := widgets.NewTree("h", widgets.TreeNode{ID: "p", Text: "p", Children: []widgets.TreeNode{{ID: "q", Text: "q", Unselectable: true}}})
	h.Expand("p")
	if !h.AtEdge(widgets.Right) {
		t.Fatal("no selectable child")
	}
	h.Focus()
	h, msgs := treePress(h, "right")
	if treeCursor(h) != "p" || len(msgs) != 0 {
		t.Fatal("right finds no child")
	}

	empty := widgets.NewTree("e")
	empty.SetSize(5, 2)
	empty.Focus()
	if _, ok := empty.Current(); ok {
		t.Fatal("empty has no current")
	}
	for _, d := range []widgets.Direction{widgets.Up, widgets.Down, widgets.Left, widgets.Right} {
		if !empty.AtEdge(d) {
			t.Fatal("empty is at every edge")
		}
	}
	empty, msgs = treePress(empty, "down", "enter")
	if len(msgs) != 0 || uitest.Plain(empty.View()) != "\n" {
		t.Fatal("empty tree")
	}
}

func TestTreeDataUpdates(t *testing.T) {
	tr := newTestTree(10)
	tr.Expand("a")
	tr.Select("a2")
	tr.SetRoots(widgets.TreeNode{ID: "a", Text: "A2", Children: []widgets.TreeNode{
		{ID: "a2", Text: "again"}, {ID: "n", Text: "new"}}})
	if treeCursor(tr) != "a2" || !tr.Expanded("a") {
		t.Fatal("cursor and expansion survive by id")
	}
	tr.SetRoots(widgets.TreeNode{ID: "q", Text: "q"})
	if treeCursor(tr) != "q" || len(tr.Roots()) != 1 {
		t.Fatal("cursor falls back to the first selectable node")
	}
	tr.SetRoots(widgets.TreeNode{ID: "r", Text: "r", Children: []widgets.TreeNode{{ID: "r1", Text: "r1"}}})
	tr.Expand("r")
	tr.Select("r1")
	if !tr.SetChildren("r", []widgets.TreeNode{{ID: "r1", Text: "r1"}, {ID: "r2", Text: "r2", Children: []widgets.TreeNode{{ID: "deepest", Text: "d"}}}}) {
		t.Fatal("set children")
	}
	if treeCursor(tr) != "r1" {
		t.Fatal("lazy load keeps the cursor")
	}
	if !tr.SetChildren("deepest", []widgets.TreeNode{{ID: "kid", Text: "k"}}) {
		t.Fatal("nested set children")
	}
	if tr.SetChildren("missing", nil) {
		t.Fatal("missing node")
	}
	if !tr.Select("kid") || !tr.Expanded("deepest") && !tr.Expanded("r2") {
		t.Fatal("select opens ancestors")
	}
	if !tr.Expanded("r2") || !tr.Expanded("deepest") {
		t.Fatal("ancestors open")
	}
	// collapsing an ancestor moves the cursor to it
	tr.Collapse("r2")
	if treeCursor(tr) != "r2" {
		t.Fatalf("cursor %s", treeCursor(tr))
	}
	tr.Select("r1")
	tr.SetExpanded("r", false)
	if treeCursor(tr) != "r" {
		t.Fatal("cursor follows the collapse")
	}
	if tr.SetExpanded("nope", true) {
		t.Fatal("unknown node")
	}
	tr.Expand("r")
	tr.CollapseAll()
	if tr.Expanded("r") {
		t.Fatal("collapse all")
	}
	tr.SetExpanded("r", false) // already closed: no-op
	// data passed in is not modified
	src := treeSample()
	shared := widgets.NewTree("s", src...)
	shared.SetChildren("a", nil)
	if len(src[0].Children) != 2 {
		t.Fatal("caller data was modified")
	}
	// Elm value semantics: an updated copy does not disturb the original
	orig := newTestTree(10)
	next, _ := orig.Update(uitest.Key("right"))
	if orig.Expanded("a") || !next.Expanded("a") {
		t.Fatal("expansion map is copy-on-write")
	}
}

func TestTreeScrolling(t *testing.T) {
	var kids []widgets.TreeNode
	for i := range 10 {
		kids = append(kids, widgets.TreeNode{ID: string(rune('a' + i)), Text: strings.Repeat(string(rune('a'+i)), 3)})
	}
	tr := widgets.NewTree("s", kids...)
	tr.Focus()
	tr, _ = tr.Update(tea.WindowSizeMsg{Width: 10, Height: 3})
	tr, _ = treePress(tr, "end")
	if got := uitest.Plain(tr.View()); got != "  hhh\n  iii\n  jjj" {
		t.Fatalf("cursor stays visible: %q", got)
	}
	tr, _ = tr.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if got := uitest.Plain(tr.View()); got != "  ggg\n  hhh\n  iii" || treeCursor(tr) != "j" {
		t.Fatalf("wheel up scrolls only: %q", got)
	}
	for range 20 {
		tr, _ = tr.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	}
	if got := uitest.Plain(tr.View()); got != "  hhh\n  iii\n  jjj" {
		t.Fatalf("wheel down clamps: %q", got)
	}
	for range 20 {
		tr, _ = tr.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	}
	tr, _ = tr.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if !strings.HasPrefix(uitest.Plain(tr.View()), "  aaa") {
		t.Fatal("wheel up clamps at the top")
	}
	tr, _ = treePress(tr, "home", "down")
	tr, _ = treePress(tr, "end", "home")
	if !strings.HasPrefix(uitest.Plain(tr.View()), "  aaa") {
		t.Fatal("home scrolls back")
	}
}

func TestTreeMouse(t *testing.T) {
	tr := newTestTree(10)
	tr.Blur()
	click := func(tr widgets.Tree, x, y int) (widgets.Tree, []tea.Msg) {
		tr, cmd := tr.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: x, Y: y})
		return tr, uitest.Msgs(cmd)
	}
	tr, msgs := click(tr, 5, 1)
	if treeCursor(tr) != "b" || len(msgs) != 1 {
		t.Fatalf("click selects (even unfocused): %s", treeCursor(tr))
	}
	if _, ok := msgs[0].(widgets.NodeHighlightedMsg); !ok {
		t.Fatalf("%T", msgs[0])
	}
	tr, msgs = click(tr, 5, 1)
	if _, ok := msgs[0].(widgets.NodeSelectedMsg); !ok || len(msgs) != 1 {
		t.Fatal("click on the current node selects it")
	}
	tr, msgs = click(tr, 0, 0)
	if !tr.Expanded("a") || len(msgs) != 0 || treeCursor(tr) != "b" {
		t.Fatal("marker click toggles without moving")
	}
	tr, _ = click(tr, 1, 0)
	if tr.Expanded("a") {
		t.Fatal("marker click toggles back")
	}
	tr.ExpandAll()
	tr, _ = click(tr, 4, 1)
	if treeCursor(tr) != "a1" {
		t.Fatalf("click on a child: %s", treeCursor(tr))
	}
	for _, c := range [][2]int{{0, -1}, {0, 50}, {0, 10}} {
		if _, msgs = click(tr, c[0], c[1]); len(msgs) != 0 {
			t.Fatal("click outside the rows")
		}
	}
	tr, cmd := tr.Update(tea.MouseClickMsg{Button: tea.MouseRight, X: 5, Y: 0})
	if cmd != nil || treeCursor(tr) != "a1" {
		t.Fatal("only the left button")
	}
	hdr := widgets.NewTree("h", widgets.TreeNode{ID: "g", Text: "G", Unselectable: true, Children: []widgets.TreeNode{{ID: "k", Text: "k"}}},
		widgets.TreeNode{ID: "z", Text: "z"})
	hdr.SetSize(10, 4)
	hdr, msgs = click(hdr, 5, 0)
	if len(msgs) != 0 || treeCursor(hdr) != "z" {
		t.Fatal("a click on a header caption does nothing")
	}
	hdr, _ = click(hdr, 0, 0)
	if !hdr.Expanded("g") {
		t.Fatal("headers expand from the marker")
	}
}

func TestTreeKeyMapHelp(t *testing.T) {
	km := widgets.DefaultTreeKeyMap()
	if len(km.ShortHelp()) != 5 || len(km.FullHelp()) != 3 {
		t.Fatal("help")
	}
	tr := newTestTree(5)
	tr.KeyMap.Activate.SetEnabled(false)
	_, msgs := treePress(tr, "enter")
	if len(msgs) != 0 {
		t.Fatal("bindings come from the key map")
	}
}
