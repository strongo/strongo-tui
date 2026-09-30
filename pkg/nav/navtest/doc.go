// Package navtest drives a nav.Model without a terminal, so screens can be
// tested by pressing keys and reading what would be on screen.
//
//	h := navtest.New(t, nav.Page{Title: "Home", Menu: menu, Content: content})
//	h.Press("down", "enter")
//	h.RequireContains("Project created")
//
// The harness sends a window size first (120x40 unless WithSize says otherwise),
// feeds every message to the shell, runs the commands that come back and feeds
// their messages in turn, and reports what the shell rendered as plain text:
// styling stripped, trailing spaces trimmed. It is a thin helper over the Elm
// loop; tests that only concern a screen can drive its Update directly with the
// helpers of package uitest.
//
// # Time
//
// Alert timers do not use the wall clock. Advance moves a fake clock forward and
// delivers the timers that fall due.
//
// # Commands
//
// Commands are executed synchronously, in order. A command that blocks for
// longer than the harness timeout fails the test.
package navtest
