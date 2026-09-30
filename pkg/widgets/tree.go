package widgets

import (
	"image/color"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
)

// TreeNode is one node of a Tree. Nodes are plain data: the tree keeps no
// pointers into them, and expansion and the cursor are tracked by ID.
type TreeNode struct {
	// ID identifies the node and must be unique in the tree.
	ID string
	// Text is the caption; it may carry ANSI styling and emoji.
	Text string
	// Ref is an arbitrary value the application attaches to the node.
	Ref any
	// Color colours the caption when the node is not the selected row; nil keeps
	// the terminal default.
	Color color.Color
	// Children are the nodes below this one.
	Children []TreeNode
	// Unselectable marks a group header: it is drawn (and can be expanded with
	// SetExpanded, ExpandAll or a click on its marker) but the cursor never rests
	// on it.
	Unselectable bool
}

// NodeHighlightedMsg reports that the cursor moved to a node because of a key or
// a click. It is not sent for changes made with Select, SetRoots or SetChildren.
type NodeHighlightedMsg struct {
	ID   string
	Node TreeNode
}

// NodeSelectedMsg reports that a node was activated: Enter on the cursor node,
// or a click on the node that already holds the cursor. The node is never
// changed by the message itself; Enter afterwards toggles a node that has
// children.
type NodeSelectedMsg struct {
	ID   string
	Node TreeNode
}

// TreeKeyMap holds the key bindings of Tree.
type TreeKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Expand   key.Binding
	Collapse key.Binding
	Activate key.Binding
	Toggle   key.Binding
}

// DefaultTreeKeyMap returns the default key bindings.
func DefaultTreeKeyMap() TreeKeyMap {
	return TreeKeyMap{
		Up:       key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:     key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		Home:     key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "first")),
		End:      key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "last")),
		Expand:   key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "expand")),
		Collapse: key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "collapse")),
		Activate: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Toggle:   key.NewBinding(key.WithKeys("space", " "), key.WithHelp("space", "toggle")),
	}
}

// ShortHelp implements help.KeyMap.
func (k TreeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Expand, k.Collapse, k.Activate}
}

// FullHelp implements help.KeyMap.
func (k TreeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Home, k.End},
		{k.Expand, k.Collapse, k.Toggle, k.Activate},
	}
}

// Tree is a scrollable, data-driven hierarchy of TreeNode values. Several roots
// give the "hidden root" layout: they are drawn at the left edge.
//
// Keys: Up, Down, PageUp, PageDown, Home and End move the cursor between
// visible selectable nodes. Right expands a collapsed node, or moves to the
// first selectable child of an expanded one. Left collapses an expanded node
// below the top level, or moves to its selectable parent; on a top-level node it
// does nothing. Enter reports NodeSelectedMsg and then toggles a node that has
// children, Space only toggles. A click selects the row (NodeSelectedMsg when it
// already held the cursor) and a click on the expand marker toggles it; the
// wheel scrolls without moving the cursor. Cursor moves made by Update are
// reported with NodeHighlightedMsg.
//
// Boundary: AtEdge(Up) is true on the first selectable visible node, Down on the
// last, Left when Left would do nothing (a top-level node, or a node whose
// parent is a group header and that cannot collapse) and Right when Right would
// do nothing (a leaf, or an expanded node without a selectable child). Tree is
// not an Editor.
//
// Why it does not wrap charm.land/bubbles/v2/tree: that component is built on
// a mutable *Node graph whose Open and Close mutate shared nodes, has no
// unselectable nodes, keys expansion by pointer rather than by ID, has no
// per-node colour, and renders the whole tree into a viewport on every move
// instead of exactly width x height rows. Tree needs all of those, so it draws
// the rows itself (with the guides and markers below) and reuses only key and
// help from bubbles.
type Tree struct {
	// KeyMap holds the key bindings.
	KeyMap TreeKeyMap

	id       string
	roots    []TreeNode
	expanded map[string]bool
	cursor   string
	offset   int
	width    int
	height   int
	focused  bool
}

// NewTree creates a tree. id identifies it in messages. The first selectable
// node holds the cursor; nothing is expanded.
func NewTree(id string, roots ...TreeNode) Tree {
	t := Tree{KeyMap: DefaultTreeKeyMap(), id: id, expanded: map[string]bool{}}
	t.SetRoots(roots...)
	return t
}

// treeRow is one visible row.
type treeRow struct {
	node   TreeNode
	parent string
	root   bool
	prefix string
}

const (
	guideTee  = "├─ "
	guideEnd  = "└─ "
	guidePipe = "│  "
	guideGap  = "   "
)

// rows flattens the visible part of the tree.
func (t Tree) rows() []treeRow {
	var out []treeRow
	for _, r := range t.roots {
		t.collect(r, "", true, true, "", &out)
	}
	return out
}

func (t Tree) collect(n TreeNode, parent string, root bool, last bool, lasts string, out *[]treeRow) {
	prefix := ""
	if !root {
		prefix = lasts + guideTee
		if last {
			prefix = lasts + guideEnd
		}
	}
	*out = append(*out, treeRow{node: n, parent: parent, root: root, prefix: prefix})
	if !t.expanded[n.ID] {
		return
	}
	childLasts := lasts
	if !root {
		childLasts += guidePipe
		if last {
			childLasts = lasts + guideGap
		}
	}
	for i, c := range n.Children {
		t.collect(c, n.ID, false, i == len(n.Children)-1, childLasts, out)
	}
}

// find returns the ids from the roots down to the node with the id, or nil.
func find(nodes []TreeNode, id string) []TreeNode {
	for _, n := range nodes {
		if n.ID == id {
			return []TreeNode{n}
		}
		if p := find(n.Children, id); p != nil {
			return append([]TreeNode{n}, p...)
		}
	}
	return nil
}

func indexOf(rows []treeRow, id string) int {
	return slices.IndexFunc(rows, func(r treeRow) bool { return r.node.ID == id })
}

// fixCursor keeps the cursor on a visible selectable node: it stays, else moves
// to the nearest visible selectable ancestor, else to the first selectable row.
func (t *Tree) fixCursor() {
	rows := t.rows()
	if i := indexOf(rows, t.cursor); i >= 0 && !rows[i].node.Unselectable {
		t.reveal(rows)
		return
	}
	path := find(t.roots, t.cursor)
	for k := len(path) - 1; k >= 0; k-- {
		if i := indexOf(rows, path[k].ID); i >= 0 && !path[k].Unselectable {
			t.cursor = path[k].ID
			t.reveal(rows)
			return
		}
	}
	t.cursor = ""
	for _, r := range rows {
		if !r.node.Unselectable {
			t.cursor = r.node.ID
			break
		}
	}
	t.reveal(rows)
}

// reveal scrolls so the cursor row is visible and the offset is in range.
func (t *Tree) reveal(rows []treeRow) {
	if i := indexOf(rows, t.cursor); i >= 0 && t.height > 0 {
		t.offset = min(max(t.offset, i-t.height+1), i)
	}
	t.offset = t.clampOffset(len(rows))
}

func (t Tree) clampOffset(n int) int {
	return max(min(t.offset, n-t.height), 0)
}

// SetRoots replaces the data. Expansion is kept by ID, and so is the cursor when
// its node still exists; otherwise the first selectable node takes it.
func (t *Tree) SetRoots(roots ...TreeNode) {
	t.roots = slices.Clone(roots)
	t.fixCursor()
}

// SetChildren replaces the children of the node id, for lazy loading, and keeps
// the cursor. It reports whether the node exists.
func (t *Tree) SetChildren(id string, children []TreeNode) bool {
	roots, ok := replaceChildren(t.roots, id, slices.Clone(children))
	if ok {
		t.roots = roots
		t.fixCursor()
	}
	return ok
}

// replaceChildren copies the spine down to the node so shared data is never
// modified in place.
func replaceChildren(nodes []TreeNode, id string, children []TreeNode) ([]TreeNode, bool) {
	for i, n := range nodes {
		if n.ID == id {
			out := slices.Clone(nodes)
			out[i].Children = children
			return out, true
		}
		if sub, ok := replaceChildren(n.Children, id, children); ok {
			out := slices.Clone(nodes)
			out[i].Children = sub
			return out, true
		}
	}
	return nodes, false
}

// Roots returns a copy of the top-level nodes.
func (t Tree) Roots() []TreeNode { return slices.Clone(t.roots) }

// SetExpanded opens or closes the node id. It reports whether the node exists.
func (t *Tree) SetExpanded(id string, open bool) bool {
	if find(t.roots, id) == nil {
		return false
	}
	if t.expanded[id] != open {
		t.expanded = setFlag(t.expanded, id, open)
		t.fixCursor()
	}
	return true
}

// setFlag copies the map, so models sharing it are not affected.
func setFlag(m map[string]bool, id string, open bool) map[string]bool {
	out := make(map[string]bool, len(m)+1)
	for k, v := range m {
		out[k] = v
	}
	if open {
		out[id] = true
	} else {
		delete(out, id)
	}
	return out
}

// Expand opens the node id.
func (t *Tree) Expand(id string) { t.SetExpanded(id, true) }

// Collapse closes the node id.
func (t *Tree) Collapse(id string) { t.SetExpanded(id, false) }

// Expanded reports whether the node id is open.
func (t Tree) Expanded(id string) bool { return t.expanded[id] }

// ExpandAll opens every node that has children.
func (t *Tree) ExpandAll() {
	all := map[string]bool{}
	var walk func(nodes []TreeNode)
	walk = func(nodes []TreeNode) {
		for _, n := range nodes {
			if len(n.Children) > 0 {
				all[n.ID] = true
				walk(n.Children)
			}
		}
	}
	walk(t.roots)
	t.expanded = all
	t.fixCursor()
}

// CollapseAll closes every node.
func (t *Tree) CollapseAll() {
	t.expanded = map[string]bool{}
	t.fixCursor()
}

// Select moves the cursor to the node id, opening its ancestors so it is
// visible, without sending a message. It reports false when the node does not
// exist or is unselectable.
func (t *Tree) Select(id string) bool {
	path := find(t.roots, id)
	if path == nil || path[len(path)-1].Unselectable {
		return false
	}
	for _, n := range path[:len(path)-1] {
		if !t.expanded[n.ID] {
			t.expanded = setFlag(t.expanded, n.ID, true)
		}
	}
	t.cursor = id
	t.reveal(t.rows())
	return true
}

// Current returns the node under the cursor, with its children.
func (t Tree) Current() (TreeNode, bool) {
	if p := find(t.roots, t.cursor); p != nil {
		return p[len(p)-1], true
	}
	return TreeNode{}, false
}

// SetSize sets the size of the view.
func (t *Tree) SetSize(width, height int) {
	t.width, t.height = width, height
	t.reveal(t.rows())
}

// Focus gives the tree the keyboard.
func (t *Tree) Focus() { t.focused = true }

// Blur takes the keyboard away.
func (t *Tree) Blur() { t.focused = false }

// Focused reports whether the tree holds focus.
func (t Tree) Focused() bool { return t.focused }

// selectable returns the indexes of the selectable rows.
func selectable(rows []treeRow) []int {
	var out []int
	for i, r := range rows {
		if !r.node.Unselectable {
			out = append(out, i)
		}
	}
	return out
}

// AtEdge implements Boundary.
func (t Tree) AtEdge(dir Direction) bool {
	rows := t.rows()
	cur := indexOf(rows, t.cursor)
	if cur < 0 {
		return true
	}
	sel := selectable(rows)
	switch dir {
	case Up:
		return sel[0] == cur
	case Down:
		return sel[len(sel)-1] == cur
	case Left:
		r := rows[cur]
		if r.root {
			return true
		}
		if t.expanded[r.node.ID] && len(r.node.Children) > 0 {
			return false
		}
		p := indexOf(rows, r.parent)
		return rows[p].node.Unselectable
	}
	n := rows[cur].node
	if len(n.Children) == 0 {
		return true
	}
	return t.expanded[n.ID] && !slices.ContainsFunc(n.Children, func(c TreeNode) bool { return !c.Unselectable })
}

// Update handles key presses while focused, clicks, the wheel, and size
// messages.
func (t Tree) Update(msg tea.Msg) (Tree, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		if !t.focused {
			return t, nil
		}
		return t.handleKey(msg)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return t.click(msg.X, msg.Y)
		}
	case tea.MouseWheelMsg:
		rows := t.rows()
		switch msg.Button {
		case tea.MouseWheelUp:
			t.offset = max(t.offset-1, 0)
		case tea.MouseWheelDown:
			t.offset = t.clampOffset(len(rows))
			t.offset = min(t.offset+1, max(len(rows)-t.height, 0))
		}
	}
	return t, nil
}

// moveTo puts the cursor on rows[i] and reports the change.
func (t Tree) moveTo(rows []treeRow, i int) (Tree, tea.Cmd) {
	if rows[i].node.ID == t.cursor {
		return t, nil
	}
	t.cursor = rows[i].node.ID
	t.reveal(rows)
	return t, Emit(NodeHighlightedMsg{ID: t.id, Node: rows[i].node})
}

func (t Tree) handleKey(msg tea.KeyPressMsg) (Tree, tea.Cmd) {
	rows := t.rows()
	cur := indexOf(rows, t.cursor)
	if cur < 0 {
		return t, nil
	}
	node := rows[cur].node
	sel := selectable(rows)
	page := max(t.height, 1)
	// first and last selectable rows within [lo, hi].
	firstIn := func(lo, hi int) int {
		for _, i := range sel {
			if i >= lo && i <= hi {
				return i
			}
		}
		return -1
	}
	lastIn := func(lo, hi int) int {
		for k := len(sel) - 1; k >= 0; k-- {
			if sel[k] >= lo && sel[k] <= hi {
				return sel[k]
			}
		}
		return -1
	}
	target := -1
	km := t.KeyMap
	switch {
	case key.Matches(msg, km.Up):
		target = lastIn(0, cur-1)
	case key.Matches(msg, km.Down):
		target = firstIn(cur+1, len(rows)-1)
	case key.Matches(msg, km.Home):
		target = sel[0]
	case key.Matches(msg, km.End):
		target = sel[len(sel)-1]
	case key.Matches(msg, km.PageUp):
		target = firstIn(max(cur-page, 0), cur-1)
	case key.Matches(msg, km.PageDown):
		target = lastIn(cur+1, min(cur+page, len(rows)-1))
	case key.Matches(msg, km.Expand):
		switch {
		case len(node.Children) == 0:
		case !t.expanded[node.ID]:
			t.expanded = setFlag(t.expanded, node.ID, true)
			t.reveal(t.rows())
		default:
			target = t.firstChildRow(rows, cur)
		}
	case key.Matches(msg, km.Collapse):
		return t.collapseKey(rows, cur)
	case key.Matches(msg, km.Toggle):
		t.toggle(node)
	case key.Matches(msg, km.Activate):
		cmd := Emit(NodeSelectedMsg{ID: t.id, Node: node})
		t.toggle(node)
		return t, cmd
	}
	if target >= 0 {
		return t.moveTo(rows, target)
	}
	return t, nil
}

// firstChildRow is the row of the first selectable direct child of rows[cur], or -1.
func (t Tree) firstChildRow(rows []treeRow, cur int) int {
	for i := cur + 1; i < len(rows); i++ {
		if rows[i].parent == rows[cur].node.ID && !rows[i].node.Unselectable {
			return i
		}
	}
	return -1
}

func (t Tree) collapseKey(rows []treeRow, cur int) (Tree, tea.Cmd) {
	r := rows[cur]
	if r.root {
		return t, nil
	}
	if t.expanded[r.node.ID] && len(r.node.Children) > 0 {
		t.expanded = setFlag(t.expanded, r.node.ID, false)
		t.reveal(t.rows())
		return t, nil
	}
	p := indexOf(rows, r.parent)
	if rows[p].node.Unselectable {
		return t, nil
	}
	return t.moveTo(rows, p)
}

func (t *Tree) toggle(n TreeNode) {
	if len(n.Children) > 0 {
		t.expanded = setFlag(t.expanded, n.ID, !t.expanded[n.ID])
		t.fixCursor()
	}
}

func (t Tree) click(x, y int) (Tree, tea.Cmd) {
	rows := t.rows()
	i := t.offset + y
	if y < 0 || y >= t.height || i >= len(rows) {
		return t, nil
	}
	r := rows[i]
	markerAt := ansi.StringWidth(r.prefix)
	if len(r.node.Children) > 0 && x >= markerAt && x < markerAt+2 {
		t.toggle(r.node)
		return t, nil
	}
	switch {
	case r.node.Unselectable:
		return t, nil
	case r.node.ID == t.cursor:
		return t, Emit(NodeSelectedMsg{ID: t.id, Node: r.node})
	}
	return t.moveTo(rows, i)
}

func (t Tree) marker(n TreeNode) string {
	switch {
	case len(n.Children) == 0:
		return "  "
	case t.expanded[n.ID]:
		return "▾ "
	}
	return "▸ "
}

// View renders exactly height rows of width columns, scrolled to the offset.
func (t Tree) View() string {
	if t.width <= 0 || t.height <= 0 {
		return ""
	}
	rows := t.rows()
	off := t.clampOffset(len(rows))
	guide := lipgloss.NewStyle().Foreground(theme.MutedColor())
	lines := make([]string, 0, t.height)
	for i := off; i < len(rows) && len(lines) < t.height; i++ {
		r := rows[i]
		if r.node.ID == t.cursor {
			plain := r.prefix + t.marker(r.node) + ansi.Strip(r.node.Text)
			lines = append(lines, theme.SelectedStyle(t.focused).Render(AlignIn(plain, t.width, AlignLeft)))
			continue
		}
		text := r.node.Text
		if r.node.Color != nil {
			text = lipgloss.NewStyle().Foreground(r.node.Color).Render(text)
		}
		lines = append(lines, PadRight(guide.Render(r.prefix)+t.marker(r.node)+text, t.width))
	}
	return Fit(strings.Join(lines, "\n"), t.width, t.height)
}
