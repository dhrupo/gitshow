// Package ui owns the Bubble Tea Model that renders the cinematic
// replay TUI.  Layout (per blueprint §16.3):
//
//	┌──────────────────────────┐
//	│ Header                   │
//	├──────────────────────────┤
//	│ Main Playback            │
//	├──────────────────────────┤
//	│ Diff View                │
//	├──────────────────────────┤
//	│ Timeline                 │
//	└──────────────────────────┘
//
// W3 ships the static shell + keyboard navigation.  W4 plugs in the
// animation system + timeline scheduler.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

// MinSpeed and MaxSpeed bound the user-visible "speed" indicator.
const (
	MinSpeed     = 1
	MaxSpeed     = 5
	DefaultSpeed = 3
)

// CommitWithDiff bundles a commit with its already-loaded diff so the
// UI never blocks on git work during navigation.
type CommitWithDiff struct {
	Commit gitpkg.Commit
	Files  []gitpkg.ChangedFile
}

// Model is the Bubble Tea state.
type Model struct {
	// Static repo data, populated before tea.NewProgram is called.
	RepoName string
	Commits  []CommitWithDiff

	// Cursor and playback.
	idx    int
	paused bool
	speed  int

	// Viewport dimensions from tea.WindowSizeMsg.
	width  int
	height int

	// Style overrides; nil falls back to DefaultTheme.
	theme *Theme

	// quitting flips true on q / ctrl+c so View() can render a final
	// message before tea.Quit returns.
	quitting bool
}

// New builds an initial Model.  Pass an empty slice for repos with no
// history; the View renders an empty-state placeholder instead of
// panicking.
func New(repoName string, commits []CommitWithDiff) Model {
	return Model{
		RepoName: repoName,
		Commits:  commits,
		idx:      0,
		paused:   false,
		speed:    DefaultSpeed,
		theme:    DefaultTheme(),
	}
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is the message dispatcher.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case " ", "space":
		m.paused = !m.paused
		return m, nil
	case "left", "h":
		if m.idx > 0 {
			m.idx--
		}
		return m, nil
	case "right", "l":
		if m.idx < len(m.Commits)-1 {
			m.idx++
		}
		return m, nil
	case "up", "k":
		if m.speed < MaxSpeed {
			m.speed++
		}
		return m, nil
	case "down", "j":
		if m.speed > MinSpeed {
			m.speed--
		}
		return m, nil
	case "home", "g":
		m.idx = 0
		return m, nil
	case "end", "G":
		if len(m.Commits) > 0 {
			m.idx = len(m.Commits) - 1
		}
		return m, nil
	}
	return m, nil
}

// CurrentIndex returns the index of the currently focused commit.
// Useful for tests and external observers.
func (m Model) CurrentIndex() int { return m.idx }

// Paused reports whether playback is paused.
func (m Model) Paused() bool { return m.paused }

// Speed returns the current speed indicator (1..5).
func (m Model) Speed() int { return m.speed }

// Quitting reports whether the model has been told to quit.
func (m Model) Quitting() bool { return m.quitting }
