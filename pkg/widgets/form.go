package widgets

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/strongo/strongo-tui/pkg/theme"
)

// FieldKind is the kind of a form field.
type FieldKind int

// Field kinds.
const (
	// TextField is a single-line text input.
	TextField FieldKind = iota
	// PasswordField is a text input that hides what is typed.
	PasswordField
	// SelectField is a single choice among Options.
	SelectField
	// StaticField is a read-only line of text that focus skips.
	StaticField
)

// Field describes one labelled row of a Form.
type Field struct {
	// ID identifies the field in Value, SetValue and the messages; it must be
	// unique in the form.
	ID string
	// Label is the caption in front of the control.
	Label string
	// Kind selects the control.
	Kind FieldKind
	// Value is the initial text, or for a SelectField the initially chosen
	// option (empty or unknown: nothing chosen).
	Value string
	// Placeholder is shown in an empty text field.
	Placeholder string
	// Width is the width of the control in columns; zero fills the row (text) or
	// fits the value (select).
	Width int
	// Options are the choices of a SelectField.
	Options []string
	// Text is the content of a StaticField; it may carry ANSI styling.
	Text string
	// Accept filters typed and pasted text. It receives the text the field would
	// hold after an insertion and the inserted rune; returning false drops the
	// insertion. It is the one function a Field holds: a pure input filter, that
	// is validation data and not behaviour, so it must not have side effects.
	Accept func(text string, last rune) bool
}

// ButtonRole says what pressing a FormButton reports.
type ButtonRole int

// Button roles.
const (
	// SubmitRole buttons report SubmitMsg.
	SubmitRole ButtonRole = iota
	// CancelRole buttons report CancelMsg.
	CancelRole
	// ActionRole buttons report ButtonPressedMsg.
	ActionRole
)

// FormButton is a push button below the fields of a Form.
type FormButton struct {
	// ID identifies the button in the messages.
	ID string
	// Label is the caption.
	Label string
	// Role selects the message a press reports.
	Role ButtonRole
}

// FieldChangedMsg reports that the user changed the value of a field by typing,
// deleting, pasting or choosing an option. SetValue does not send it.
type FieldChangedMsg struct {
	ID      string
	FieldID string
	Value   string
}

// SubmitMsg reports that a SubmitRole button was pressed.
type SubmitMsg struct {
	ID       string
	ButtonID string
	Values   map[string]string
}

// CancelMsg reports that a CancelRole button was pressed, or Esc was, in which
// case ButtonID is empty.
type CancelMsg struct {
	ID       string
	ButtonID string
}

// ButtonPressedMsg reports that an ActionRole button was pressed.
type ButtonPressedMsg struct {
	ID       string
	ButtonID string
	Values   map[string]string
}

// FormKeyMap holds the key bindings of Form.
type FormKeyMap struct {
	Next   key.Binding
	Prev   key.Binding
	Submit key.Binding
	Cancel key.Binding
	Left   key.Binding
	Right  key.Binding
	// Open opens a closed select.
	Open key.Binding
	// Press presses the focused button.
	Press key.Binding
}

// DefaultFormKeyMap returns the default key bindings.
func DefaultFormKeyMap() FormKeyMap {
	return FormKeyMap{
		Next:   key.NewBinding(key.WithKeys("tab", "down"), key.WithHelp("tab", "next")),
		Prev:   key.NewBinding(key.WithKeys("shift+tab", "up"), key.WithHelp("shift+tab", "previous")),
		Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next / press")),
		Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Left:   key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "previous button / option")),
		Right:  key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "next button / option")),
		Open:   key.NewBinding(key.WithKeys("enter", "space", " ", "down"), key.WithHelp("enter", "open")),
		Press:  key.NewBinding(key.WithKeys("space", " "), key.WithHelp("space", "press")),
	}
}

// ShortHelp implements help.KeyMap.
func (k FormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Submit, k.Cancel}
}

// FullHelp implements help.KeyMap.
func (k FormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Next, k.Prev, k.Submit, k.Cancel}, {k.Left, k.Right, k.Open, k.Press}}
}

// selectRows caps the option rows an open select adds to the form.
const selectRows = 8

// formGap is the number of columns between a label and its control, and between
// buttons.
const formGap = 2

// fieldState is the runtime state of one field, kept beside the Field data.
type fieldState struct {
	in   textinput.Model
	opt  int
	open bool
	cur  int
	top  int
}

// Form is a vertical form of labelled fields above a row of buttons. Fields sit
// one per row with labels right-aligned to the widest one, then a blank row and
// the buttons, aligned left. When the height is too small the form scrolls so
// the focused row stays visible.
//
// Text and password fields wrap charm.land/bubbles/v2/textinput (virtual cursor
// on, blinking off, so no timer messages). Select fields and layout are own
// code: bubbles has no select, and a form needs the option list to take rows
// from the layout while it is open (capped at eight, scrolling inside).
//
// Messages: FieldChangedMsg when Update changes a value, SubmitMsg,
// CancelMsg and ButtonPressedMsg for buttons, and CancelMsg with an empty
// ButtonID for Esc.
//
// Keys (FormKeyMap): Tab and Down move to the next control, Shift+Tab and Up to
// the previous one; Enter on a field moves on and on a button presses it; Left
// and Right move between buttons; Esc cancels. On a closed select Enter, Space
// and Down open the option list, Up and Down move in it, Enter picks and Esc
// closes it, and Left and Right step through the options. Ctrl+Q and Ctrl+C are
// never consumed. Clicks focus a field or press a button.
//
// Boundary: AtEdge(Up) is true on the first control, Down on the buttons (or the
// last field when there are none), Left when the text cursor is at the start, on
// the first button, or on a select at its first option, and Right when the
// cursor is at the end, on the last button, or on a select at its last option.
// While a select list is open no edge is reached. Form is an Editor while a text
// or password field is focused.
type Form struct {
	// KeyMap holds the key bindings.
	KeyMap FormKeyMap

	id      string
	fields  []Field
	buttons []FormButton
	st      []fieldState
	focus   int
	focused bool
	width   int
	height  int
	offset  int
}

// NewForm creates a form. id identifies it in messages. It starts blurred, with
// the first field (or button) as the control that receives focus.
func NewForm(id string, fields []Field, buttons []FormButton) Form {
	f := Form{KeyMap: DefaultFormKeyMap(), id: id, focus: -1}
	f.SetFields(fields)
	f.SetButtons(buttons)
	return f
}

func newFieldState(fd Field) fieldState {
	s := fieldState{opt: -1}
	switch fd.Kind {
	case TextField, PasswordField:
		s.in = textinput.New()
		s.in.Prompt = ""
		s.in.Placeholder = fd.Placeholder
		if fd.Kind == PasswordField {
			s.in.EchoMode = textinput.EchoPassword
		}
		s.in.SetStyles(formInputStyles())
		s.in.SetValue(fd.Value)
		s.in.CursorEnd()
	case SelectField:
		s.opt = slices.Index(fd.Options, fd.Value)
	}
	return s
}

// formInputStyles are the text input styles, read from the theme at call time.
// The cursor does not blink.
func formInputStyles() textinput.Styles {
	text := lipgloss.NewStyle().Foreground(theme.TextColor()).Underline(true)
	state := textinput.StyleState{
		Text:        text,
		Placeholder: lipgloss.NewStyle().Foreground(theme.MutedColor()).Underline(true),
		Suggestion:  lipgloss.NewStyle().Foreground(theme.MutedColor()),
		Prompt:      lipgloss.NewStyle(),
	}
	var s textinput.Styles
	s.Focused, s.Blurred = state, state
	s.Cursor = textinput.CursorStyle{Color: theme.TextColor(), Shape: tea.CursorBlock}
	return s
}

func (f Form) isText(i int) bool {
	return i >= 0 && i < len(f.fields) && (f.fields[i].Kind == TextField || f.fields[i].Kind == PasswordField)
}

func (f Form) isSelect(i int) bool {
	return i >= 0 && i < len(f.fields) && f.fields[i].Kind == SelectField
}

// focusable reports whether the control i (fields first, then buttons) can hold
// focus.
func (f Form) focusable(i int) bool {
	return i >= 0 && i < len(f.fields)+len(f.buttons) && (i >= len(f.fields) || f.fields[i].Kind != StaticField)
}

// firstStop returns the first focusable control, or -1.
func (f Form) firstStop() int {
	for i := range len(f.fields) + len(f.buttons) {
		if f.focusable(i) {
			return i
		}
	}
	return -1
}

// focusKey names the focused control so it can be found after a rebuild.
func (f Form) focusKey() string {
	switch {
	case f.focus < 0:
		return ""
	case f.focus < len(f.fields):
		return "f:" + f.fields[f.focus].ID
	}
	return "b:" + f.buttons[f.focus-len(f.fields)].ID
}

// SetFields replaces the fields. Typed values of fields whose ID and kind
// survive are kept, and so is the focus when its control survives.
func (f *Form) SetFields(fields []Field) {
	key := f.focusKey()
	old, oldFields := f.st, f.fields
	f.fields = slices.Clone(fields)
	f.st = make([]fieldState, len(fields))
	for i, fd := range f.fields {
		f.st[i] = newFieldState(fd)
		j := slices.IndexFunc(oldFields, func(o Field) bool { return o.ID == fd.ID && o.Kind == fd.Kind })
		if j < 0 {
			continue
		}
		switch fd.Kind {
		case TextField, PasswordField:
			f.st[i].in.SetValue(old[j].in.Value())
			f.st[i].in.SetCursor(old[j].in.Position())
		case SelectField:
			if old[j].opt >= 0 {
				f.st[i].opt = slices.Index(fd.Options, oldFields[j].Options[old[j].opt])
			}
		}
	}
	f.refocus(key)
}

// SetButtons replaces the buttons, keeping the focus when its control survives.
func (f *Form) SetButtons(buttons []FormButton) {
	key := f.focusKey()
	f.buttons = slices.Clone(buttons)
	f.refocus(key)
}

// refocus restores the focus after the controls changed.
func (f *Form) refocus(key string) {
	total := len(f.fields) + len(f.buttons)
	f.focus = -1
	for i := range total {
		if f.focusable(i) && f.focusKeyOf(i) == key {
			f.focus = i
		}
	}
	if f.focus < 0 {
		f.focus = f.firstStop()
	}
	f.syncInputs()
	f.syncFocus()
	f.reveal()
}

func (f Form) focusKeyOf(i int) string {
	if i < len(f.fields) {
		return "f:" + f.fields[i].ID
	}
	return "b:" + f.buttons[i-len(f.fields)].ID
}

// Value returns the value of the field id: the text, or the chosen option of a
// select. Unknown and static fields have none.
func (f Form) Value(id string) string {
	i := slices.IndexFunc(f.fields, func(fd Field) bool { return fd.ID == id })
	switch {
	case f.isText(i):
		return f.st[i].in.Value()
	case f.isSelect(i) && f.st[i].opt >= 0:
		return f.fields[i].Options[f.st[i].opt]
	}
	return ""
}

// Values returns the value of every text, password and select field by ID.
func (f Form) Values() map[string]string {
	out := map[string]string{}
	for _, fd := range f.fields {
		if fd.Kind != StaticField {
			out[fd.ID] = f.Value(fd.ID)
		}
	}
	return out
}

// SetValue sets the value of the field id without sending a message. A value
// that is not one of a select's options clears the choice.
func (f *Form) SetValue(id, value string) {
	i := slices.IndexFunc(f.fields, func(fd Field) bool { return fd.ID == id })
	switch {
	case f.isText(i):
		f.st[i].in.SetValue(value)
		f.st[i].in.CursorEnd()
	case f.isSelect(i):
		f.st[i].opt = slices.Index(f.fields[i].Options, value)
	}
}

// FocusField moves the focus to the field id and reports whether it exists and
// can take focus.
func (f *Form) FocusField(id string) bool {
	i := slices.IndexFunc(f.fields, func(fd Field) bool { return fd.ID == id })
	if !f.focusable(i) {
		return false
	}
	f.setFocus(i)
	return true
}

// SetSize sets the size of the view.
func (f *Form) SetSize(width, height int) {
	f.width, f.height = width, height
	f.syncInputs()
	f.reveal()
}

// Focus gives the form the keyboard.
func (f *Form) Focus() {
	f.focused = true
	if f.focus < 0 {
		f.focus = f.firstStop()
	}
	f.syncFocus()
}

// Blur takes the keyboard away and closes an open select list.
func (f *Form) Blur() {
	f.focused = false
	f.closeLists()
	f.syncFocus()
}

// Focused reports whether the form holds focus.
func (f Form) Focused() bool { return f.focused }

func (f *Form) closeLists() {
	for i := range f.st {
		f.st[i].open = false
	}
}

func (f *Form) setFocus(i int) {
	f.closeLists()
	f.focus = i
	f.syncFocus()
	f.reveal()
}

// syncFocus gives the text cursor to the focused text field only.
func (f *Form) syncFocus() {
	for i := range f.st {
		if !f.isText(i) {
			continue
		}
		if f.focused && i == f.focus {
			f.st[i].in.Focus()
		} else {
			f.st[i].in.Blur()
		}
	}
}

// syncInputs sets the width of every text input to its control width.
func (f *Form) syncInputs() {
	lay := f.layout()
	for i := range f.st {
		if f.isText(i) {
			f.st[i].in.SetWidth(max(lay.ctrlW[i]-1, 1))
		}
	}
}

// formLayout is where everything sits, a pure function of the model.
type formLayout struct {
	colX    int
	ctrlW   []int
	rowY    []int
	rowH    []int
	buttonY int
	buttonX []int
	total   int
}

func (f Form) layout() formLayout {
	labelW := 0
	for _, fd := range f.fields {
		labelW = max(labelW, ansi.StringWidth(fd.Label))
	}
	lay := formLayout{buttonY: -1}
	if labelW > 0 {
		lay.colX = labelW + formGap
	}
	avail := max(f.width-lay.colX, 1)
	y := 0
	for i, fd := range f.fields {
		w := avail
		if fd.Width > 0 {
			w = min(fd.Width, avail)
		}
		lay.ctrlW = append(lay.ctrlW, w)
		h := 1
		if f.st[i].open {
			h += min(len(fd.Options), selectRows)
		}
		lay.rowY = append(lay.rowY, y)
		lay.rowH = append(lay.rowH, h)
		y += h
	}
	if len(f.buttons) > 0 {
		if len(f.fields) > 0 {
			y++
		}
		lay.buttonY = y
		x := 0
		for _, b := range f.buttons {
			lay.buttonX = append(lay.buttonX, x)
			x += f.button(b, false).Width() + formGap
		}
		y++
	}
	lay.total = y
	return lay
}

func (f Form) button(b FormButton, focused bool) Button {
	return Button{Label: b.Label, Focused: focused}
}

// reveal scrolls so the focused row, with its open list, is visible.
func (f *Form) reveal() {
	lay := f.layout()
	if f.focus >= 0 && f.height > 0 {
		start, end := lay.buttonY, lay.buttonY+1
		if f.focus < len(f.fields) {
			start, end = lay.rowY[f.focus], lay.rowY[f.focus]+1
			if s := f.st[f.focus]; s.open {
				end += s.cur - s.top + 1 // down to the cursor row of the open list
			}
		}
		start = max(start, end-f.height)
		f.offset = min(max(f.offset, end-f.height), start)
	}
	f.offset = max(min(f.offset, lay.total-f.height), 0)
}

// Editing implements Editor: a text or password field holds the cursor.
func (f Form) Editing() bool { return f.focused && f.isText(f.focus) }

// AtEdge implements Boundary.
func (f Form) AtEdge(dir Direction) bool {
	if f.focus < 0 {
		return true
	}
	nf := len(f.fields)
	if f.isSelect(f.focus) && f.st[f.focus].open {
		return false
	}
	switch dir {
	case Up:
		return f.focus == f.firstStop()
	case Down:
		return f.focus >= nf || len(f.buttons) == 0 && f.focus == nf-1
	case Left:
		switch {
		case f.isText(f.focus):
			return f.st[f.focus].in.Position() == 0
		case f.isSelect(f.focus):
			return f.st[f.focus].opt <= 0
		}
		return f.focus <= nf
	}
	switch {
	case f.isText(f.focus):
		return f.st[f.focus].in.Position() == len([]rune(f.st[f.focus].in.Value()))
	case f.isSelect(f.focus):
		return f.st[f.focus].opt >= len(f.fields[f.focus].Options)-1
	}
	return f.focus == nf+len(f.buttons)-1
}

// neverConsumed are the keys the application owns; the form leaves them alone.
var neverConsumed = key.NewBinding(key.WithKeys("ctrl+q", "ctrl+c"))

// Update handles key presses while focused, clicks, the wheel, pastes and size
// messages.
func (f Form) Update(msg tea.Msg) (Form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.SetSize(msg.Width, msg.Height)
		return f, nil
	case tea.KeyPressMsg:
		if !f.focused || f.focus < 0 || key.Matches(msg, neverConsumed) {
			return f, nil
		}
		f.st = slices.Clone(f.st)
		return f.handleKey(msg)
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return f, nil
		}
		f.st = slices.Clone(f.st)
		return f.click(msg.X, msg.Y)
	case tea.MouseWheelMsg:
		lay := f.layout()
		switch msg.Button {
		case tea.MouseWheelUp:
			f.offset = max(f.offset-1, 0)
		case tea.MouseWheelDown:
			f.offset = max(min(f.offset+1, lay.total-f.height), 0)
		}
		return f, nil
	}
	// Anything else, such as a clipboard paste, belongs to the text cursor.
	if f.focused && f.isText(f.focus) {
		f.st = slices.Clone(f.st)
		return f.edit(msg)
	}
	return f, nil
}

func (f Form) handleKey(msg tea.KeyPressMsg) (Form, tea.Cmd) {
	km := f.KeyMap
	nf := len(f.fields)
	if f.isSelect(f.focus) && f.st[f.focus].open {
		return f.listKey(msg)
	}
	switch {
	case key.Matches(msg, km.Cancel):
		return f, Emit(CancelMsg{ID: f.id})
	case f.isSelect(f.focus) && key.Matches(msg, km.Open):
		f.openList(f.focus)
		return f, nil
	case f.isSelect(f.focus) && key.Matches(msg, km.Left, km.Right):
		return f.cycle(key.Matches(msg, km.Right))
	case key.Matches(msg, km.Next):
		f.move(1)
		return f, nil
	case key.Matches(msg, km.Prev):
		f.move(-1)
		return f, nil
	case f.focus < nf && key.Matches(msg, km.Submit):
		f.move(1)
		return f, nil
	case f.focus >= nf && key.Matches(msg, km.Submit, km.Press):
		return f, f.press(f.focus - nf)
	case f.focus >= nf && key.Matches(msg, km.Right):
		if f.focus+1 < nf+len(f.buttons) {
			f.setFocus(f.focus + 1)
		}
		return f, nil
	case f.focus >= nf && key.Matches(msg, km.Left):
		if f.focus > nf {
			f.setFocus(f.focus - 1)
		}
		return f, nil
	case f.isText(f.focus):
		return f.edit(msg)
	}
	return f, nil
}

// move focuses the next focusable control in the direction, if any.
func (f *Form) move(dir int) {
	for j := f.focus + dir; j >= 0 && j < len(f.fields)+len(f.buttons); j += dir {
		if f.focusable(j) {
			f.setFocus(j)
			return
		}
	}
}

// press reports the press of button j.
func (f Form) press(j int) tea.Cmd {
	b := f.buttons[j]
	switch b.Role {
	case SubmitRole:
		return Emit(SubmitMsg{ID: f.id, ButtonID: b.ID, Values: f.Values()})
	case CancelRole:
		return Emit(CancelMsg{ID: f.id, ButtonID: b.ID})
	}
	return Emit(ButtonPressedMsg{ID: f.id, ButtonID: b.ID, Values: f.Values()})
}

// editBindings are the text input bindings that edit text or move the cursor.
func editBindings(in textinput.Model) []key.Binding {
	k := in.KeyMap
	return []key.Binding{
		k.CharacterForward, k.CharacterBackward, k.WordForward, k.WordBackward,
		k.DeleteWordBackward, k.DeleteWordForward, k.DeleteAfterCursor, k.DeleteBeforeCursor,
		k.DeleteCharacterBackward, k.DeleteCharacterForward, k.LineStart, k.LineEnd, k.Paste,
	}
}

// edit feeds msg to the focused text input and reports a change of its value.
func (f Form) edit(msg tea.Msg) (Form, tea.Cmd) {
	i := f.focus
	s := &f.st[i]
	if k, ok := msg.(tea.KeyPressMsg); ok && k.Text == "" && !key.Matches(k, editBindings(s.in)...) {
		return f, nil
	}
	old, oldPos := s.in.Value(), s.in.Position()
	var cmd tea.Cmd
	s.in, cmd = s.in.Update(msg)
	text := s.in.Value()
	if text == old {
		return f, cmd
	}
	if accept := f.fields[i].Accept; accept != nil && len([]rune(text)) > len([]rune(old)) {
		if !accept(text, []rune(text)[s.in.Position()-1]) {
			s.in.SetValue(old)
			s.in.SetCursor(oldPos)
			return f, cmd
		}
	}
	return f, tea.Batch(cmd, Emit(FieldChangedMsg{ID: f.id, FieldID: f.fields[i].ID, Value: text}))
}

func (f *Form) openList(i int) {
	s := &f.st[i]
	s.open = true
	s.cur = max(s.opt, 0)
	s.top = min(max(s.top, s.cur-selectRows+1), s.cur)
	f.reveal()
}

// cycle steps the focused select to the next or previous option.
func (f Form) cycle(forward bool) (Form, tea.Cmd) {
	s := &f.st[f.focus]
	n := len(f.fields[f.focus].Options)
	if n == 0 {
		return f, nil
	}
	next := max(s.opt-1, 0)
	if forward {
		next = min(s.opt+1, n-1)
	}
	return f.choose(f.focus, next)
}

// choose makes option opt of select i the value and reports a change.
func (f Form) choose(i, opt int) (Form, tea.Cmd) {
	s := &f.st[i]
	if s.opt == opt {
		return f, nil
	}
	s.opt = opt
	return f, Emit(FieldChangedMsg{ID: f.id, FieldID: f.fields[i].ID, Value: f.fields[i].Options[opt]})
}

// listKey handles keys while a select list is open; every key is consumed.
func (f Form) listKey(msg tea.KeyPressMsg) (Form, tea.Cmd) {
	km := f.KeyMap
	s := &f.st[f.focus]
	n := len(f.fields[f.focus].Options)
	switch {
	case key.Matches(msg, km.Cancel):
		s.open = false
		f.reveal()
	case key.Matches(msg, km.Next):
		s.cur = min(s.cur+1, n-1)
	case key.Matches(msg, km.Prev):
		s.cur = max(s.cur-1, 0)
	case key.Matches(msg, km.Submit, km.Press):
		s.open = false
		f.reveal()
		if n > 0 {
			return f.choose(f.focus, s.cur)
		}
	}
	s.top = min(max(s.top, s.cur-selectRows+1), s.cur)
	f.reveal()
	return f, nil
}

// click focuses the clicked field or presses the clicked button.
func (f Form) click(x, y int) (Form, tea.Cmd) {
	lay := f.layout()
	y += f.offset
	if y < 0 {
		return f, nil
	}
	for i, fd := range f.fields {
		top := lay.rowY[i]
		if y < top || y >= top+lay.rowH[i] {
			continue
		}
		switch {
		case y > top: // a row of the open list
			s := &f.st[i]
			s.open = false
			if opt := s.top + y - top - 1; x >= lay.colX && opt < len(fd.Options) {
				return f.choose(i, opt)
			}
		case fd.Kind == StaticField:
		case f.isSelect(i) && f.focus == i && f.st[i].open:
			f.st[i].open = false
		case f.isSelect(i) && f.focus == i:
			f.openList(i)
		default:
			f.setFocus(i)
			if f.isSelect(i) {
				f.openList(i)
			}
		}
		return f, nil
	}
	if y != lay.buttonY {
		return f, nil
	}
	for j, b := range f.buttons {
		if x >= lay.buttonX[j] && x < lay.buttonX[j]+f.button(b, false).Width() {
			f.setFocus(len(f.fields) + j)
			return f, f.press(j)
		}
	}
	return f, nil
}

// View renders exactly height rows of width columns, scrolled so the focused
// row is visible.
func (f Form) View() string {
	if f.width <= 0 || f.height <= 0 {
		return ""
	}
	lay := f.layout()
	label := lipgloss.NewStyle().Foreground(theme.MutedColor())
	var lines []string
	for i, fd := range f.fields {
		focused := f.focused && f.focus == i
		ls := label
		if focused {
			ls = lipgloss.NewStyle().Foreground(theme.FocusColor())
		}
		row := ""
		if lay.colX > 0 {
			row = ls.Render(AlignIn(fd.Label, lay.colX-formGap, AlignRight)) + strings.Repeat(" ", formGap)
		}
		lines = append(lines, row+f.control(i, lay.ctrlW[i], focused))
		if f.st[i].open {
			lines = append(lines, f.list(i, lay)...)
		}
	}
	if len(f.buttons) > 0 {
		if len(f.fields) > 0 {
			lines = append(lines, "")
		}
		var row []string
		for j, b := range f.buttons {
			btn := f.button(b, f.focused && f.focus == len(f.fields)+j)
			row = append(row, btn.View(btn.Width()))
		}
		lines = append(lines, strings.Join(row, strings.Repeat(" ", formGap)))
	}
	off := f.offset
	off = max(min(off, len(lines)-f.height), 0)
	return Fit(strings.Join(lines[off:], "\n"), f.width, f.height)
}

// control renders the control of field i in w columns.
func (f Form) control(i, w int, focused bool) string {
	fd := f.fields[i]
	switch fd.Kind {
	case StaticField:
		return PadRight(fd.Text, w)
	case SelectField:
		opt := ""
		if f.st[i].opt >= 0 {
			opt = fd.Options[f.st[i].opt]
		}
		text := opt + " ▾"
		if fd.Width > 0 {
			text = PadRight(opt, max(w-2, 0)) + " ▾"
		}
		style := lipgloss.NewStyle().Foreground(theme.TextColor()).Underline(true)
		if focused {
			style = theme.SelectedStyle(true)
		}
		return style.Render(ansi.Truncate(text, w, ""))
	}
	in := f.st[i].in
	in.SetStyles(formInputStyles())
	in.SetWidth(max(w-1, 1))
	return ansi.Truncate(in.View(), w, "")
}

// list renders the open option list of select i below its row.
func (f Form) list(i int, lay formLayout) []string {
	s := f.st[i]
	opts := f.fields[i].Options
	width := 0
	for _, o := range opts {
		width = max(width, ansi.StringWidth(o))
	}
	width = min(width+2, max(f.width-lay.colX, 1))
	bg, fg := theme.SurfaceColors()
	plain := lipgloss.NewStyle().Background(bg).Foreground(fg)
	var out []string
	for r := s.top; r < len(opts) && r < s.top+selectRows; r++ {
		style := plain
		if r == s.cur {
			style = theme.SelectedStyle(true)
		}
		out = append(out, strings.Repeat(" ", lay.colX)+style.Render(AlignIn(" "+ansi.Strip(opts[r]), width, AlignLeft)))
	}
	return out
}
