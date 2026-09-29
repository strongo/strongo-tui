package nav

import tea "charm.land/bubbletea/v2"

// runTeaProgram runs a program to completion. It is the shell's only path to a
// terminal, so tests replace it.
var runTeaProgram = func(p *tea.Program) (tea.Model, error) { return p.Run() }

// Run runs the shell as a full-screen program with mouse support and blocks
// until it exits. options are passed to tea.NewProgram, for example to redirect
// input and output in tests.
func Run(m Model, options ...tea.ProgramOption) error {
	_, err := runTeaProgram(tea.NewProgram(m, options...))
	return err
}
