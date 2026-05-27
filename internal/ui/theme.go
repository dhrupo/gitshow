package ui

import "github.com/charmbracelet/lipgloss"

// Theme bundles the lipgloss Styles used by the UI.  Phase 1 ships a
// single default theme; Phase 5 ships a registry (nord / dracula /
// catppuccin / gruvbox / tokyo-night).
type Theme struct {
	HeaderBar      lipgloss.Style
	HeaderAccent   lipgloss.Style
	HeaderMuted    lipgloss.Style
	CommitCard     lipgloss.Style
	CommitSubject  lipgloss.Style
	CommitAuthor   lipgloss.Style
	CommitMeta     lipgloss.Style
	DiffPanel      lipgloss.Style
	TimelineFrame  lipgloss.Style
	TimelinePast   lipgloss.Style
	TimelineCurr   lipgloss.Style
	TimelineFuture lipgloss.Style
	StatusBar      lipgloss.Style
	EmptyState     lipgloss.Style
}

// DefaultTheme returns the built-in gitshow default theme.  The colors
// are deliberately conservative so they look acceptable on both dark
// and light terminals.
func DefaultTheme() *Theme {
	const (
		accent  = lipgloss.Color("#7AA2F7") // soft blue
		warm    = lipgloss.Color("#E0AF68") // amber
		past    = lipgloss.Color("#565F89") // muted
		curr    = lipgloss.Color("#9ECE6A") // green
		future  = lipgloss.Color("#3B4261") // dim
		text    = lipgloss.Color("#C0CAF5") // foreground
		muted   = lipgloss.Color("#7C82A0") // muted text
		divider = lipgloss.Color("#414868")
	)

	return &Theme{
		HeaderBar:      lipgloss.NewStyle().Foreground(text).Bold(true).Padding(0, 1),
		HeaderAccent:   lipgloss.NewStyle().Foreground(accent).Bold(true),
		HeaderMuted:    lipgloss.NewStyle().Foreground(muted),
		CommitCard:     lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(divider).Padding(0, 1),
		CommitSubject:  lipgloss.NewStyle().Foreground(text).Bold(true),
		CommitAuthor:   lipgloss.NewStyle().Foreground(accent),
		CommitMeta:     lipgloss.NewStyle().Foreground(muted).Italic(true),
		DiffPanel:      lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(divider).Padding(0, 1),
		TimelineFrame:  lipgloss.NewStyle().Foreground(divider),
		TimelinePast:   lipgloss.NewStyle().Foreground(past),
		TimelineCurr:   lipgloss.NewStyle().Foreground(curr).Bold(true),
		TimelineFuture: lipgloss.NewStyle().Foreground(future),
		StatusBar:      lipgloss.NewStyle().Foreground(warm),
		EmptyState:     lipgloss.NewStyle().Foreground(muted).Italic(true),
	}
}
