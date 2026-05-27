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
	"time"

	tea "github.com/charmbracelet/bubbletea"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
	"github.com/dhrupo/gitshow/internal/timeline"
)

// FrameInterval is the wall-clock tick interval (~30 fps).  Animations
// read the timeline clock, not the wall clock, so changing speed
// adjusts how much cinematic time passes per tick.
const FrameInterval = 33 * time.Millisecond

// commitHoldDuration is how long the cinematic camera stays on a
// single commit before auto-advancing.  Tuned for "watchable" pacing
// from blueprint §5.2.
const commitHoldDuration = 6 * time.Second

// subjectRevealDuration is how long the typewriter takes to spell
// out a commit subject from blank to fully typed.
const subjectRevealDuration = 700 * time.Millisecond

// diffLinePerInterval is the gap between successive diff-line
// stagger reveals.
const diffLinePerInterval = 18 * time.Millisecond

// tickMsg is the periodic frame heartbeat.
type tickMsg time.Time

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

	// clock drives the per-commit animation timing.  It also gates
	// auto-advance (we move to the next commit when commitHoldDuration
	// of cinematic time has elapsed since the camera landed here).
	clock *timeline.Timeline

	// quitting flips true on q / ctrl+c so View() can render a final
	// message before tea.Quit returns.
	quitting bool
}

// speedToMultiplier maps the user-facing 1..5 indicator to the
// timeline speed multiplier.  3 is "normal", 1 is slow, 5 is fast.
func speedToMultiplier(speed int) float64 {
	switch speed {
	case 1:
		return 0.5
	case 2:
		return 0.75
	case 3:
		return 1.0
	case 4:
		return 1.5
	case 5:
		return 2.0
	default:
		return 1.0
	}
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
		clock:    timeline.New(speedToMultiplier(DefaultSpeed)),
	}
}

// Init satisfies tea.Model.  Starts the frame-tick ping.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(FrameInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// commitElapsed reports cinematic time since the camera landed on the
// current commit.
func (m Model) commitElapsed() time.Duration {
	if m.clock == nil {
		return 0
	}
	return m.clock.Elapsed()
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
	case tickMsg:
		return m.handleTick()
	}
	return m, nil
}

func (m Model) handleTick() (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, nil
	}
	// Auto-advance to the next commit when we've held this one long
	// enough.  Paused freezes the clock, so the comparison naturally
	// stops moving while the user thinks.
	if !m.paused && m.idx < len(m.Commits)-1 && m.commitElapsed() >= commitHoldDuration {
		m.idx++
		m.resetClock()
	}
	return m, tickCmd()
}

func (m *Model) resetClock() {
	if m.clock == nil {
		m.clock = timeline.New(speedToMultiplier(m.speed))
		return
	}
	m.clock.Reset()
	if m.paused {
		m.clock.Pause()
	}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit
	case " ", "space":
		m.paused = !m.paused
		if m.clock != nil {
			if m.paused {
				m.clock.Pause()
			} else {
				m.clock.Resume()
			}
		}
		return m, nil
	case "left", "h":
		if m.idx > 0 {
			m.idx--
			m.resetClock()
		}
		return m, nil
	case "right", "l":
		if m.idx < len(m.Commits)-1 {
			m.idx++
			m.resetClock()
		}
		return m, nil
	case "up", "k":
		if m.speed < MaxSpeed {
			m.speed++
			if m.clock != nil {
				m.clock.SetSpeed(speedToMultiplier(m.speed))
			}
		}
		return m, nil
	case "down", "j":
		if m.speed > MinSpeed {
			m.speed--
			if m.clock != nil {
				m.clock.SetSpeed(speedToMultiplier(m.speed))
			}
		}
		return m, nil
	case "home", "g":
		if m.idx != 0 {
			m.idx = 0
			m.resetClock()
		}
		return m, nil
	case "end", "G":
		if len(m.Commits) > 0 && m.idx != len(m.Commits)-1 {
			m.idx = len(m.Commits) - 1
			m.resetClock()
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
