package ui

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

func fixtureCommits(n int) []CommitWithDiff {
	out := make([]CommitWithDiff, n)
	for i := 0; i < n; i++ {
		out[i] = CommitWithDiff{
			Commit: gitpkg.Commit{
				Hash:      "abcdef1234567890",
				Author:    "Test User",
				Message:   "commit " + itoa(i+1) + "\n\nbody",
				Timestamp: time.Now(),
			},
			Files: []gitpkg.ChangedFile{
				{Path: "hello.go", Mode: gitpkg.Modified, Hunks: []gitpkg.Hunk{{NewStart: 1, NewLines: 1, Lines: []gitpkg.HunkLine{{Kind: gitpkg.LineAdded, Text: "line"}}}}},
			},
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func send(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, _ := m.Update(msg)
	mm, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned non-Model: %v (%v)", next, reflect.TypeOf(next))
	}
	return mm
}

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestModel_InitialState(t *testing.T) {
	m := New("repo", fixtureCommits(5))
	if m.CurrentIndex() != 0 {
		t.Errorf("CurrentIndex = %d, want 0", m.CurrentIndex())
	}
	if m.Speed() != DefaultSpeed {
		t.Errorf("Speed = %d, want %d", m.Speed(), DefaultSpeed)
	}
	if m.Paused() {
		t.Error("model started paused; want playing")
	}
}

func TestModel_SpaceTogglesPause(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	m = send(t, m, key(" "))
	if !m.Paused() {
		t.Error("space did not pause")
	}
	m = send(t, m, key(" "))
	if m.Paused() {
		t.Error("second space did not resume")
	}
}

func TestModel_ArrowKeys_Navigate(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRight})
	if m.CurrentIndex() != 1 {
		t.Errorf("after right: idx = %d, want 1", m.CurrentIndex())
	}
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRight})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRight})
	if m.CurrentIndex() != 2 {
		t.Errorf("right past end: idx = %d, want 2 (clamped)", m.CurrentIndex())
	}
	m = send(t, m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.CurrentIndex() != 1 {
		t.Errorf("after left: idx = %d, want 1", m.CurrentIndex())
	}
}

func TestModel_VimKeys_Navigate(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	m = send(t, m, key("l"))
	if m.CurrentIndex() != 1 {
		t.Errorf("after l: idx = %d, want 1", m.CurrentIndex())
	}
	m = send(t, m, key("h"))
	if m.CurrentIndex() != 0 {
		t.Errorf("after h: idx = %d, want 0", m.CurrentIndex())
	}
}

func TestModel_UpDown_AdjustsSpeed(t *testing.T) {
	m := New("repo", fixtureCommits(2))
	m = send(t, m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Speed() != DefaultSpeed+1 {
		t.Errorf("up: speed = %d, want %d", m.Speed(), DefaultSpeed+1)
	}
	for i := 0; i < 10; i++ {
		m = send(t, m, tea.KeyMsg{Type: tea.KeyUp})
	}
	if m.Speed() != MaxSpeed {
		t.Errorf("clamped speed = %d, want %d", m.Speed(), MaxSpeed)
	}
	for i := 0; i < 10; i++ {
		m = send(t, m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.Speed() != MinSpeed {
		t.Errorf("clamped speed = %d, want %d", m.Speed(), MinSpeed)
	}
}

func TestModel_HomeEndJumpToBounds(t *testing.T) {
	m := New("repo", fixtureCommits(10))
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnd})
	if m.CurrentIndex() != 9 {
		t.Errorf("after end: idx = %d, want 9", m.CurrentIndex())
	}
	m = send(t, m, tea.KeyMsg{Type: tea.KeyHome})
	if m.CurrentIndex() != 0 {
		t.Errorf("after home: idx = %d, want 0", m.CurrentIndex())
	}
}

func TestModel_QTriggersQuit(t *testing.T) {
	m := New("repo", fixtureCommits(2))
	next, cmd := m.Update(key("q"))
	mm := next.(Model)
	if !mm.Quitting() {
		t.Error("q did not set quitting")
	}
	if cmd == nil {
		t.Error("q did not return a Cmd (expected tea.Quit)")
	}
}

func TestModel_WindowSizeMsg_StoresDimensions(t *testing.T) {
	m := New("repo", fixtureCommits(1))
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	mm := next.(Model)
	if mm.width != 120 || mm.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", mm.width, mm.height)
	}
}

func TestModel_View_EmptyCommits(t *testing.T) {
	m := New("repo", nil)
	v := m.View()
	if !strings.Contains(v, "no commits") {
		t.Errorf("empty view should mention 'no commits'; got %q", v)
	}
}

func TestModel_View_AfterQuit_PrintsThanks(t *testing.T) {
	m := New("repo", fixtureCommits(1))
	next, _ := m.Update(key("q"))
	mm := next.(Model)
	v := mm.View()
	if !strings.Contains(v, "Thanks") {
		t.Errorf("quit view should thank user; got %q", v)
	}
}

func TestModel_View_RendersAllFourPanels(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 30})

	// Seek past the subject typewriter + the diff stagger so we render
	// the fully-revealed frame.
	if m.clock != nil {
		m.clock.Seek(commitHoldDuration)
	}

	view := m.View()
	if !strings.Contains(view, "gitshow") {
		t.Error("header missing 'gitshow'")
	}
	if !strings.Contains(view, "commit ") {
		t.Error("main panel missing commit meta")
	}
	if !strings.Contains(view, "@@") {
		t.Error("diff panel missing hunk header at fully-revealed frame")
	}
	if !strings.Contains(view, "navigate") {
		t.Error("timeline hints missing")
	}
}

// At the very first frame, the diff stagger hasn't started yet, so the
// hunk should NOT be visible.  This protects the animation timing.
func TestModel_View_DiffIsAnimatedNotImmediatelyVisible(t *testing.T) {
	m := New("repo", fixtureCommits(1))
	m, _ = updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.clock != nil {
		m.clock.Seek(0)
	}
	view := m.View()
	if strings.Contains(view, "@@") {
		t.Errorf("at t=0, diff should not yet be revealed; got:\n%s", view)
	}
}

// updateModel is a tiny helper used to thread tea messages through
// the Model interface without losing the concrete type.
func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

func TestTimelineDots_WidthGreaterThanCommits(t *testing.T) {
	rs := []rune(timelineDots(3, 1, 12))
	// 12 / 3 = 4 rune slots per commit; positions 0, 4, 8 carry markers.
	if rs[0] != '•' { // past
		t.Errorf("cell 0 = %q, want past dot", string(rs[0]))
	}
	if rs[4] != '●' { // current
		t.Errorf("cell 4 = %q, want current dot", string(rs[4]))
	}
	if rs[8] != '·' { // future
		t.Errorf("cell 8 = %q, want future dot", string(rs[8]))
	}
}

func TestTimelineDots_WidthSmallerThanCommits_Resamples(t *testing.T) {
	got := timelineDots(100, 50, 20)
	if rcount := len([]rune(got)); rcount != 20 {
		t.Errorf("rune len = %d, want 20", rcount)
	}
	if !strings.ContainsRune(got, '●') {
		t.Errorf("missing current marker: %q", got)
	}
}

// quick guard that we're not accidentally pulling in OS-specific code
// that breaks cross-platform builds.
func TestModel_BuildsOnAllPlatforms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("informational")
	}
}

// When enough cinematic time has elapsed, tick should advance to the
// next commit and reset the per-commit clock.
func TestModel_Tick_AutoAdvances(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	if m.clock == nil {
		t.Fatal("expected clock to be initialised")
	}
	// Jump past the commit hold duration so the next tick should
	// auto-advance us off commit 0 onto commit 1.
	m.clock.Seek(commitHoldDuration + time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	mm := next.(Model)
	if mm.CurrentIndex() != 1 {
		t.Errorf("after long tick, idx = %d, want 1", mm.CurrentIndex())
	}
}

func TestModel_Tick_DoesNotAdvanceWhenPaused(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	m = send(t, m, key(" ")) // pause
	if m.clock != nil {
		m.clock.Seek(commitHoldDuration + time.Second)
	}
	next, _ := m.Update(tickMsg(time.Now()))
	mm := next.(Model)
	if mm.CurrentIndex() != 0 {
		t.Errorf("paused tick advanced idx to %d, want 0", mm.CurrentIndex())
	}
}

func TestModel_Tick_StopsAtLastCommit(t *testing.T) {
	m := New("repo", fixtureCommits(2))
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnd}) // jump to last
	if m.clock != nil {
		m.clock.Seek(commitHoldDuration + time.Second)
	}
	next, _ := m.Update(tickMsg(time.Now()))
	mm := next.(Model)
	if mm.CurrentIndex() != 1 {
		t.Errorf("at last commit, tick should stay put; got idx=%d", mm.CurrentIndex())
	}
}

func TestModel_NavigationResetsClock(t *testing.T) {
	m := New("repo", fixtureCommits(3))
	if m.clock == nil {
		t.Fatal("expected clock")
	}
	m.clock.Seek(2 * time.Second)
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRight})
	if got := m.clock.Elapsed(); got > 100*time.Millisecond {
		t.Errorf("clock should reset on right; got elapsed=%v", got)
	}
}
