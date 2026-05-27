package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// Run launches the TUI program with the supplied model.  Returns an
// error if Bubble Tea fails to initialize (e.g. stdout is not a TTY).
func Run(m Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// IsStdoutTTY reports whether stdout is attached to a terminal.  The
// CLI uses this to decide whether to launch the cinematic TUI or fall
// back to the plain stdout dump.
func IsStdoutTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
