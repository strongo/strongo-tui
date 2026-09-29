// Package uitest holds the TTY-free helpers used to test widgets and screens:
// building key presses from their textual names and reading rendered views as
// plain text.
//
// The helpers depend only on Bubble Tea, so widget packages can use them without
// importing the navigation shell. Shell-level tests use package nav/navtest,
// which builds on these helpers.
package uitest
