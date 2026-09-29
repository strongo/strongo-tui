package uitest

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

var namedKeys = map[string]rune{
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"pgup":      tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	"enter":     tea.KeyEnter,
	"tab":       tea.KeyTab,
	"esc":       tea.KeyEscape,
	"backspace": tea.KeyBackspace,
	"delete":    tea.KeyDelete,
	"insert":    tea.KeyInsert,
	"space":     tea.KeySpace,
	"f1":        tea.KeyF1,
	"f2":        tea.KeyF2,
	"f3":        tea.KeyF3,
	"f4":        tea.KeyF4,
	"f5":        tea.KeyF5,
	"f6":        tea.KeyF6,
	"f7":        tea.KeyF7,
	"f8":        tea.KeyF8,
	"f9":        tea.KeyF9,
	"f10":       tea.KeyF10,
	"f11":       tea.KeyF11,
	"f12":       tea.KeyF12,
}

var modifiers = map[string]tea.KeyMod{
	"ctrl":  tea.ModCtrl,
	"alt":   tea.ModAlt,
	"shift": tea.ModShift,
	"meta":  tea.ModMeta,
	"super": tea.ModSuper,
}

// Key builds a key press from its textual name, the same text the press's
// String method returns: "a", "enter", "up", "shift+tab", "ctrl+q", "alt+down",
// "f1". "backtab" is accepted as an alias of "shift+tab", "escape" of "esc".
// It panics on an unknown name, which is a bug in the test.
func Key(name string) tea.KeyPressMsg {
	switch name {
	case "backtab":
		name = "shift+tab"
	case "escape":
		name = "esc"
	}
	var mod tea.KeyMod
	rest := name
	for {
		head, tail, found := strings.Cut(rest, "+")
		m, isMod := modifiers[head]
		if !found || !isMod || tail == "" {
			break
		}
		mod |= m
		rest = tail
	}
	if code, ok := namedKeys[rest]; ok {
		key := tea.Key{Code: code, Mod: mod}
		if code == tea.KeySpace && mod == 0 {
			key.Text = " "
		}
		return tea.KeyPressMsg(key)
	}
	if utf8.RuneCountInString(rest) != 1 {
		panic(fmt.Sprintf("uitest: unknown key %q", name))
	}
	r, _ := utf8.DecodeRuneInString(rest)
	key := tea.Key{Code: r, Mod: mod}
	if mod&^tea.ModShift == 0 {
		key.Text = rest
	}
	return tea.KeyPressMsg(key)
}

// Text builds the key press of a rune typed into a text field.
func Text(s string) tea.KeyPressMsg {
	r, _ := utf8.DecodeRuneInString(s)
	return tea.KeyPressMsg(tea.Key{Code: r, Text: s})
}

// Plain strips styling from a rendered view and trims trailing spaces from every
// line, so assertions read like the screen.
func Plain(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

// Lines is Plain split into lines.
func Lines(view string) []string {
	return strings.Split(Plain(view), "\n")
}

// Msgs runs cmd and returns the messages it produces, in order. Batches and
// sequences are unpacked, and commands that produce no message are skipped, so
// a test can assert on what a component reported:
//
//	list, cmd := list.Update(uitest.Key("enter"))
//	msgs := uitest.Msgs(cmd) // []tea.Msg{widgets.ItemSelectedMsg{...}}
//
// Commands run synchronously; a command that blocks blocks the test.
func Msgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	switch m := msg.(type) {
	case nil:
		return nil
	case tea.BatchMsg:
		return flatten(m)
	}
	if v := reflect.ValueOf(msg); v.Kind() == reflect.Slice && v.Type().Elem() == reflect.TypeFor[tea.Cmd]() {
		cmds := make([]tea.Cmd, v.Len())
		for i := range cmds {
			cmds[i], _ = v.Index(i).Interface().(tea.Cmd)
		}
		return flatten(cmds)
	}
	return []tea.Msg{msg}
}

func flatten(cmds []tea.Cmd) []tea.Msg {
	var msgs []tea.Msg
	for _, c := range cmds {
		msgs = append(msgs, Msgs(c)...)
	}
	return msgs
}
