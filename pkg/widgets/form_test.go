package widgets_test

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tuigoff/tuigoff/pkg/uitest"
	"github.com/tuigoff/tuigoff/pkg/widgets"
)

func formFields() []widgets.Field {
	return []widgets.Field{
		{ID: "name", Label: "Name", Placeholder: "who"},
		{ID: "pass", Label: "Password", Kind: widgets.PasswordField},
		{ID: "info", Label: "Note", Kind: widgets.StaticField, Text: "\x1b[1mfixed\x1b[0m"},
		{ID: "color", Label: "Colour", Kind: widgets.SelectField, Options: []string{"red", "green", "blue"}, Value: "green"},
	}
}

func formButtons() []widgets.FormButton {
	return []widgets.FormButton{
		{ID: "ok", Label: "OK", Role: widgets.SubmitRole},
		{ID: "no", Label: "Cancel", Role: widgets.CancelRole},
		{ID: "act", Label: "Act", Role: widgets.ActionRole},
	}
}

func newTestForm(w, h int) widgets.Form {
	f := widgets.NewForm("form", formFields(), formButtons())
	f.SetSize(w, h)
	f.Focus()
	return f
}

func formPress(f widgets.Form, keys ...string) (widgets.Form, []tea.Msg) {
	var all []tea.Msg
	for _, k := range keys {
		var cmd tea.Cmd
		if len(k) == 1 && k[0] >= 'a' && k[0] <= 'z' || k == "X" || k == "1" {
			f, cmd = f.Update(uitest.Text(k))
		} else {
			f, cmd = f.Update(uitest.Key(k))
		}
		all = append(all, uitest.Msgs(cmd)...)
	}
	return f, all
}

func formLines(f widgets.Form) []string { return uitest.Lines(f.View()) }

func TestFormRendering(t *testing.T) {
	f := newTestForm(30, 9)
	got := uitest.Plain(f.View())
	want := "    Name  \nPassword\n    Note  fixed\n  Colour  green ▾\n\n OK    Cancel    Act\n\n\n"
	// the first row shows the placeholder
	want = strings.Replace(want, "    Name  \n", "    Name  who\n", 1)
	if got != want {
		t.Fatalf("got\n%q\nwant\n%q", got, want)
	}
	for _, l := range strings.Split(f.View(), "\n") {
		if lipgloss.Width(l) != 30 {
			t.Fatalf("line width %d", lipgloss.Width(l))
		}
	}
	f.SetSize(0, 3)
	if f.View() != "" {
		t.Fatal("no size, no view")
	}
	// fixed-width control and select
	g := widgets.NewForm("g", []widgets.Field{
		{ID: "a", Label: "A", Width: 6, Value: "abcdefghij"},
		{ID: "s", Label: "S", Kind: widgets.SelectField, Width: 8, Value: "one", Options: []string{"one"}},
		{ID: "e", Label: "E", Kind: widgets.SelectField, Options: []string{"x"}},
	}, nil)
	g.SetSize(20, 3)
	if got := uitest.Plain(g.View()); got != "A  abcdef\nS  one    ▾\nE   ▾" {
		t.Fatalf("%q", got)
	}
	// no labels: controls start at column 0; no buttons: no blank row
	h := widgets.NewForm("h", []widgets.Field{{ID: "a", Value: "x"}}, nil)
	h.SetSize(5, 2)
	if got := uitest.Plain(h.View()); got != "x\n" {
		t.Fatalf("%q", got)
	}
}

func TestFormFocusAndKeys(t *testing.T) {
	f := newTestForm(30, 9)
	if !f.Focused() || !f.Editing() || !f.AtEdge(widgets.Up) {
		t.Fatal("starts on the first field")
	}
	f, msgs := formPress(f, "a", "b")
	if f.Value("name") != "ab" || len(msgs) != 2 {
		t.Fatalf("typing: %q %v", f.Value("name"), msgs)
	}
	if m := msgs[1].(widgets.FieldChangedMsg); m.ID != "form" || m.FieldID != "name" || m.Value != "ab" {
		t.Fatalf("%#v", m)
	}
	if f.AtEdge(widgets.Up) && f.AtEdge(widgets.Down) {
		t.Fatal("edges")
	}
	if !f.AtEdge(widgets.Right) || f.AtEdge(widgets.Left) {
		t.Fatal("cursor at the end")
	}
	f, _ = formPress(f, "left", "left")
	if !f.AtEdge(widgets.Left) || f.AtEdge(widgets.Right) {
		t.Fatal("cursor at the start")
	}
	f, msgs = formPress(f, "right", "backspace", "delete")
	if f.Value("name") != "" && f.Value("name") != "a" {
		t.Fatalf("editing keys: %q", f.Value("name"))
	}
	if len(msgs) == 0 {
		t.Fatal("deleting reports")
	}
	f, msgs = formPress(f, "enter") // next field
	if f.Editing() != true || len(msgs) != 0 {
		t.Fatal("enter moves on")
	}
	f, _ = formPress(f, "s", "e", "c")
	if f.Value("pass") != "sec" {
		t.Fatal("password typed")
	}
	if strings.Contains(uitest.Plain(f.View()), "sec") {
		t.Fatal("password must be hidden")
	}
	f, _ = formPress(f, "tab") // skips the static field, lands on the select
	if f.Editing() || !f.AtEdge(widgets.Left) == false {
		t.Fatal("select focus")
	}
	f, _ = formPress(f, "shift+tab")
	if f.Value("pass") != "sec" || !f.Editing() {
		t.Fatal("back to the password")
	}
	f, _ = formPress(f, "up", "up")
	if !f.Editing() || !f.AtEdge(widgets.Up) {
		t.Fatal("up stops at the first field")
	}
	f, _ = formPress(f, "tab", "tab", "tab")
	if f.Editing() || !f.AtEdge(widgets.Down) {
		t.Fatal("first button")
	}
	if !f.AtEdge(widgets.Left) || f.AtEdge(widgets.Right) || f.AtEdge(widgets.Up) {
		t.Fatal("button edges")
	}
	f, _ = formPress(f, "right", "right", "right")
	if !f.AtEdge(widgets.Right) || f.AtEdge(widgets.Left) {
		t.Fatal("last button")
	}
	f, _ = formPress(f, "left", "left", "left")
	if !f.AtEdge(widgets.Left) {
		t.Fatal("first button again")
	}
	f, _ = formPress(f, "down", "tab")
	f, _ = formPress(f, "up")
	if f.Editing() == true {
		t.Fatal("up from the second button goes to the select, not a text field")
	}
	f, _ = formPress(f, "tab", "tab", "tab", "tab", "tab", "tab")
	if !f.AtEdge(widgets.Right) {
		t.Fatal("tab clamps at the last button")
	}
}

func TestFormButtonsMessages(t *testing.T) {
	f := newTestForm(30, 9)
	f.SetValue("name", "Ann")
	f.FocusField("color")
	f, _ = formPress(f, "tab")
	f, msgs := formPress(f, "enter")
	want := map[string]string{"name": "Ann", "pass": "", "color": "green"}
	if len(msgs) != 1 || !reflect.DeepEqual(msgs[0], widgets.SubmitMsg{ID: "form", ButtonID: "ok", Values: want}) {
		t.Fatalf("submit: %#v", msgs)
	}
	f, msgs = formPress(f, "right", "space")
	if !reflect.DeepEqual(msgs, []tea.Msg{widgets.CancelMsg{ID: "form", ButtonID: "no"}}) {
		t.Fatalf("cancel button: %#v", msgs)
	}
	f, msgs = formPress(f, "right", "enter")
	if !reflect.DeepEqual(msgs, []tea.Msg{widgets.ButtonPressedMsg{ID: "form", ButtonID: "act", Values: want}}) {
		t.Fatalf("action: %#v", msgs)
	}
	_, msgs = formPress(f, "esc")
	if !reflect.DeepEqual(msgs, []tea.Msg{widgets.CancelMsg{ID: "form"}}) {
		t.Fatalf("esc: %#v", msgs)
	}
}

func TestFormNeverConsumesQuitKeys(t *testing.T) {
	f := newTestForm(30, 9)
	for _, k := range []string{"ctrl+q", "ctrl+c"} {
		g, cmd := f.Update(uitest.Key(k))
		if cmd != nil || !reflect.DeepEqual(g, f) {
			t.Fatalf("%s consumed", k)
		}
	}
	f.Blur()
	g, cmd := f.Update(uitest.Text("a"))
	if cmd != nil || f.Editing() || g.Value("name") != "" {
		t.Fatal("blurred forms ignore keys")
	}
	f.Blur()
	f.Focus()
	if !f.Editing() {
		t.Fatal("refocus")
	}
}

func TestFormAcceptFilter(t *testing.T) {
	f := widgets.NewForm("f", []widgets.Field{{ID: "n", Label: "N", Accept: func(text string, last rune) bool {
		return last >= '0' && last <= '9'
	}}}, nil)
	f.SetSize(10, 1)
	f.Focus()
	f, msgs := formPress(f, "1", "a", "x")
	if f.Value("n") != "1" || len(msgs) != 1 {
		t.Fatalf("accept: %q %v", f.Value("n"), msgs)
	}
	f, msgs = formPress(f, "backspace")
	if f.Value("n") != "" || len(msgs) != 1 {
		t.Fatal("deletion is not filtered")
	}
}

func TestFormPasteAndOtherMessages(t *testing.T) {
	f := newTestForm(30, 9)
	f, cmd := f.Update(tea.PasteMsg{Content: "pasted"})
	if f.Value("name") != "pasted" {
		t.Fatal("paste")
	}
	if m := uitest.Msgs(cmd); len(m) != 1 || m[0].(widgets.FieldChangedMsg).Value != "pasted" {
		t.Fatalf("%v", m)
	}
	_, cmd = f.Update(tea.PasteMsg{Content: ""})
	if cmd != nil {
		t.Fatal("empty paste changes nothing")
	}
	f, cmd = f.Update(uitest.Key("ctrl+v"))
	if cmd == nil {
		t.Fatal("ctrl+v asks for the clipboard")
	}
	f.FocusField("color")
	g, cmd := f.Update("unrelated")
	if cmd != nil || g.Value("name") != "pasted" {
		t.Fatal("messages for a select are ignored")
	}
	f, _ = f.Update(tea.WindowSizeMsg{Width: 12, Height: 4})
	if len(strings.Split(f.View(), "\n")) != 4 {
		t.Fatal("window size")
	}
}

func TestFormSelectKeys(t *testing.T) {
	f := newTestForm(30, 12)
	f.FocusField("color")
	if f.Editing() || f.AtEdge(widgets.Left) || f.AtEdge(widgets.Right) {
		t.Fatal("middle option: neither edge")
	}
	f, msgs := formPress(f, "right")
	if f.Value("color") != "blue" || len(msgs) != 1 {
		t.Fatalf("cycle %q", f.Value("color"))
	}
	if !f.AtEdge(widgets.Right) {
		t.Fatal("last option")
	}
	f, msgs = formPress(f, "right")
	if len(msgs) != 0 {
		t.Fatal("no wrap")
	}
	f, _ = formPress(f, "left", "left", "left")
	if f.Value("color") != "red" || !f.AtEdge(widgets.Left) {
		t.Fatal("first option")
	}
	f, _ = formPress(f, "enter")
	lines := formLines(f)
	if !strings.Contains(strings.Join(lines, "\n"), "red\n") || lines[4] != "" && !strings.HasPrefix(strings.TrimSpace(lines[4]), "red") {
		t.Fatalf("open list:\n%s", strings.Join(lines, "\n"))
	}
	if f.AtEdge(widgets.Up) || f.AtEdge(widgets.Down) || f.AtEdge(widgets.Left) || f.AtEdge(widgets.Right) {
		t.Fatal("an open list has no edges")
	}
	f, msgs = formPress(f, "down", "down", "down", "up", "enter")
	if f.Value("color") != "green" || len(msgs) != 1 {
		t.Fatalf("pick: %q %v", f.Value("color"), msgs)
	}
	f, msgs = formPress(f, "space", "enter")
	if len(msgs) != 0 {
		t.Fatal("picking the same option is silent")
	}
	f, msgs = formPress(f, "down", "esc")
	if f.Value("color") != "green" || len(msgs) != 0 {
		t.Fatal("esc closes without picking")
	}
	f, _ = formPress(f, "down", "x", "space")
	if f.Value("color") != "green" {
		t.Fatal("space picks the cursor option")
	}
	// a select without options
	e := widgets.NewForm("e", []widgets.Field{{ID: "s", Label: "S", Kind: widgets.SelectField}}, nil)
	e.SetSize(10, 4)
	e.Focus()
	e, msgs = formPress(e, "right", "left", "enter", "enter")
	if len(msgs) != 0 || e.Value("s") != "" {
		t.Fatal("empty select")
	}
}

func TestFormLongListScrollsAndForm(t *testing.T) {
	var opts []string
	for i := range 12 {
		opts = append(opts, strings.Repeat(string(rune('a'+i)), 2))
	}
	f := widgets.NewForm("l", []widgets.Field{
		{ID: "t", Label: "T"},
		{ID: "s", Label: "S", Kind: widgets.SelectField, Options: opts},
	}, []widgets.FormButton{{ID: "ok", Label: "OK"}})
	f.SetSize(20, 5)
	f.Focus()
	f.FocusField("s")
	f, _ = formPress(f, "enter")
	if got := uitest.Plain(f.View()); !strings.Contains(got, "aa") || len(strings.Split(got, "\n")) != 5 {
		t.Fatalf("%q", got)
	}
	for range 11 {
		f, _ = formPress(f, "down")
	}
	got := uitest.Plain(f.View())
	if !strings.Contains(got, "ll") || strings.Contains(got, "aa") {
		t.Fatalf("list scrolls with the cursor:\n%s", got)
	}
	f, msgs := formPress(f, "enter")
	if f.Value("s") != "ll" || len(msgs) != 1 {
		t.Fatal("picked far option")
	}
	f, _ = formPress(f, "enter")
	f, _ = formPress(f, "esc")
	// wheel scrolls the form
	f.SetSize(20, 2)
	f, _ = f.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	f, _ = f.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if !strings.HasPrefix(uitest.Plain(f.View()), "T") {
		t.Fatal("scrolled to the top")
	}
	for range 10 {
		f, _ = f.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	}
	f, _ = f.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if got := uitest.Plain(f.View()); !strings.HasSuffix(got, "OK") {
		t.Fatalf("wheel clamps at the bottom: %q", got)
	}
	// the focused row stays visible
	f.FocusField("t")
	if !strings.HasPrefix(uitest.Plain(f.View()), "T") {
		t.Fatal("focus scrolls up")
	}
	f, _ = formPress(f, "tab", "tab")
	if got := uitest.Plain(f.View()); !strings.HasSuffix(got, "OK") {
		t.Fatalf("focus scrolls down: %q", got)
	}
	// blur closes an open list
	f.FocusField("s")
	f, _ = formPress(f, "enter")
	f.Blur()
	if strings.Contains(uitest.Plain(f.View()), "aa") {
		t.Fatal("blur closes the list")
	}
}

func TestFormSetFieldsKeepsValuesAndFocus(t *testing.T) {
	f := newTestForm(30, 9)
	f, _ = formPress(f, "h", "i")
	f.FocusField("color")
	f, _ = formPress(f, "right")
	f.SetFields([]widgets.Field{
		{ID: "extra", Label: "Extra", Value: "x"},
		{ID: "name", Label: "Name"},
		{ID: "color", Label: "Colour", Kind: widgets.SelectField, Options: []string{"blue", "pink"}},
		{ID: "pass", Label: "P", Kind: widgets.SelectField, Options: []string{"1"}},
	})
	if f.Value("name") != "hi" || f.Value("color") != "blue" || f.Value("extra") != "x" || f.Value("pass") != "" {
		t.Fatalf("values: %v", f.Values())
	}
	f, _ = formPress(f, "left")
	if f.Value("color") != "blue" {
		t.Fatal("focus stayed on the select (first option, nothing to the left)")
	}
	f.SetFields([]widgets.Field{{ID: "color", Label: "C", Kind: widgets.SelectField, Options: []string{"pink"}}})
	if f.Value("color") != "" {
		t.Fatal("chosen option gone: cleared")
	}
	f.SetFields(nil)
	if f.Editing() || !f.Focused() {
		t.Fatal("focus falls to a button")
	}
	if !f.AtEdge(widgets.Up) {
		t.Fatal("first button is the first control")
	}
	f.FocusField("nope")
	f.SetButtons(nil)
	if !f.AtEdge(widgets.Down) || f.Editing() {
		t.Fatal("nothing left to focus")
	}
	f, msgs := formPress(f, "tab", "enter", "left", "right")
	if len(msgs) != 0 {
		t.Fatal("empty form")
	}
	f.Focus()
	f.SetFields([]widgets.Field{{ID: "a", Label: "A"}})
	f.SetButtons([]widgets.FormButton{{ID: "b", Label: "B"}})
	f.FocusField("a")
	f.SetButtons([]widgets.FormButton{{ID: "b", Label: "B"}, {ID: "c", Label: "C"}})
	if !f.Editing() {
		t.Fatal("focus kept when buttons change")
	}
	f.SetValue("nope", "x")
	f.SetValue("a", "long")
	if f.Value("a") != "long" || f.Value("nope") != "" {
		t.Fatal("set value")
	}
	if f.FocusField("b") {
		t.Fatal("buttons are not fields")
	}
}

func TestFormOnlyStaticFocusLandsNowhere(t *testing.T) {
	f := widgets.NewForm("s", []widgets.Field{{ID: "i", Label: "I", Kind: widgets.StaticField, Text: "t"}}, nil)
	f.SetSize(10, 2)
	f.Focus()
	if f.Editing() || !f.AtEdge(widgets.Up) || f.FocusField("i") {
		t.Fatal("static only")
	}
	f, msgs := formPress(f, "x", "enter")
	if len(msgs) != 0 {
		t.Fatal("nothing to do")
	}
	if f.Value("i") != "" || len(f.Values()) != 0 {
		t.Fatal("static has no value")
	}
	// a static field last, with buttons above it in focus order: down at the last field
	g := widgets.NewForm("g", []widgets.Field{{ID: "a", Label: "A"}}, nil)
	g.SetSize(10, 2)
	g.Focus()
	if !g.AtEdge(widgets.Down) || !g.AtEdge(widgets.Up) {
		t.Fatal("single field, no buttons")
	}
	two := widgets.NewForm("g", []widgets.Field{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, nil)
	two.Focus()
	if two.AtEdge(widgets.Down) || !two.AtEdge(widgets.Up) {
		t.Fatal("first of two fields")
	}
	two, _ = formPress(two, "down")
	if !two.AtEdge(widgets.Down) || two.AtEdge(widgets.Up) {
		t.Fatal("last field of two")
	}
	two, _ = formPress(two, "down", "enter")
	if !two.Editing() {
		t.Fatal("enter on the last field stays")
	}
}

func TestFormButtonsOnly(t *testing.T) {
	f := widgets.NewForm("b", nil, formButtons())
	f.SetSize(30, 3)
	f.Focus()
	if got := uitest.Plain(f.View()); got != " OK    Cancel    Act\n\n" {
		t.Fatalf("%q", got)
	}
	if f.Editing() || !f.AtEdge(widgets.Up) || !f.AtEdge(widgets.Down) {
		t.Fatal("buttons only")
	}
	f, _ = formPress(f, "up", "shift+tab")
	if !f.AtEdge(widgets.Left) {
		t.Fatal("stays on the first button")
	}
}

func TestFormMouse(t *testing.T) {
	f := newTestForm(30, 12)
	f.Blur()
	click := func(f widgets.Form, x, y int) (widgets.Form, []tea.Msg) {
		f, cmd := f.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: x, Y: y})
		return f, uitest.Msgs(cmd)
	}
	f, msgs := click(f, 12, 1)
	if len(msgs) != 0 {
		t.Fatal("focus click is silent")
	}
	f.Focus()
	if !f.Editing() || f.Value("pass") != "" {
		t.Fatal("click focused the password field")
	}
	f, _ = click(f, 12, 2) // static: nothing
	if f.AtEdge(widgets.Up) {
		t.Fatal("static click keeps focus")
	}
	f, _ = click(f, 12, 3) // select opens
	if !strings.Contains(uitest.Plain(f.View()), "  blue") {
		t.Fatalf("select opened:\n%s", uitest.Plain(f.View()))
	}
	f, _ = click(f, 12, 3) // closes
	if strings.Contains(uitest.Plain(f.View()), "  blue") {
		t.Fatal("second click closes")
	}
	f, _ = click(f, 12, 3) // open again (already focused)
	f, msgs = click(f, 12, 6)
	if f.Value("color") != "blue" || len(msgs) != 1 {
		t.Fatalf("pick with the mouse: %q %v", f.Value("color"), msgs)
	}
	f, _ = click(f, 12, 3)
	f, msgs = click(f, 0, 5) // outside the list column
	if len(msgs) != 0 {
		t.Fatal("left of the list closes it")
	}
	f, msgs = click(f, 3, 4) // blank row
	if len(msgs) != 0 {
		t.Fatal("blank row")
	}
	f, msgs = click(f, 1, 5)
	if len(msgs) != 1 {
		t.Fatalf("OK button: %v", msgs)
	}
	if _, ok := msgs[0].(widgets.SubmitMsg); !ok {
		t.Fatalf("%T", msgs[0])
	}
	_, msgs = click(f, 8, 5)
	if _, ok := msgs[0].(widgets.CancelMsg); !ok {
		t.Fatal("second button")
	}
	for _, c := range [][2]int{{1, -1}, {200, 5}, {1, 40}} {
		if _, msgs = click(f, c[0], c[1]); len(msgs) != 0 {
			t.Fatalf("outside %v", c)
		}
	}
	g, cmd := f.Update(tea.MouseClickMsg{Button: tea.MouseRight, X: 1, Y: 5})
	if cmd != nil || !reflect.DeepEqual(g, f) {
		t.Fatal("left button only")
	}
	// the text field click after scroll uses form coordinates
	small := newTestForm(30, 3)
	small.FocusField("color")
	small, _ = small.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 12, Y: 0})
	if !small.Editing() {
		t.Fatal("scrolled form: row 0 is the password field, not the name field")
	}
}

func TestFormValueSemantics(t *testing.T) {
	orig := newTestForm(30, 9)
	next, _ := orig.Update(uitest.Text("z"))
	if orig.Value("name") != "" || next.Value("name") != "z" {
		t.Fatal("Update returns a copy")
	}
	km := widgets.DefaultFormKeyMap()
	if len(km.ShortHelp()) != 4 || len(km.FullHelp()) != 2 {
		t.Fatal("help")
	}
}

func TestFormUnboundKeysAndSelectSetValue(t *testing.T) {
	f := newTestForm(30, 9)
	f, msgs := formPress(f, "f1", "home", "end")
	if len(msgs) != 0 || f.Value("name") != "" {
		t.Fatal("non-text keys do not edit")
	}
	f.SetValue("color", "blue")
	f.SetValue("color", "purple")
	if f.Value("color") != "" {
		t.Fatal("unknown option clears the select")
	}
	f.SetValue("color", "blue")
	f.FocusField("color")
	f, _ = formPress(f, "tab")
	f, msgs = formPress(f, "x", "f1")
	if len(msgs) != 0 {
		t.Fatal("unbound key on a button")
	}
}
