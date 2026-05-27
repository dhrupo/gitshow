// Package timeline implements the playback clock that gitshow's UI
// animates against.
//
// The clock advances in scaled time: 1 real second equals
// `speed * 1 second` of clock time.  Pause freezes the clock.  Seek
// jumps it to an absolute position.  Code that wants to read the
// current animation state asks for Elapsed(), which is computed from
// the wall clock plus accumulated pause time and the current speed.
//
// The package is pure - it never starts goroutines or fires events
// on its own.  The UI is responsible for sending tick messages.
package timeline

import (
	"sync"
	"time"
)

// MinSpeedMul and MaxSpeedMul bound the speed multiplier.  These mirror
// ui.MinSpeed / ui.MaxSpeed but are duplicated here so the timeline
// package can be imported without pulling in Bubble Tea.
const (
	MinSpeedMul = 0.25
	MaxSpeedMul = 4.0
)

// Now is the wall-clock source.  Override in tests to inject a fake
// clock without touching the package's public API.
var Now = time.Now

// Timeline tracks elapsed cinematic time.  Safe for concurrent use,
// although gitshow today only touches it from the UI goroutine.
type Timeline struct {
	mu sync.Mutex

	startedAt    time.Time
	speed        float64
	paused       bool
	pausedAt     time.Time

	// pausedAccum is the total wall-clock duration spent paused so we
	// can subtract it from the elapsed math.
	pausedAccum time.Duration

	// baseOffset is added to Elapsed(), used by Seek() to jump.
	baseOffset time.Duration
}

// New returns a Timeline that starts running immediately at the given
// speed.  If speed is 0 it defaults to 1.0.
func New(speed float64) *Timeline {
	if speed == 0 {
		speed = 1.0
	}
	return &Timeline{
		startedAt: Now(),
		speed:     clampSpeed(speed),
	}
}

// Elapsed returns the cinematic time that has passed since the
// timeline started, accounting for pause and speed multiplier.
func (t *Timeline) Elapsed() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.elapsedLocked()
}

func (t *Timeline) elapsedLocked() time.Duration {
	now := Now()
	wall := now.Sub(t.startedAt) - t.pausedAccum
	if t.paused {
		wall -= now.Sub(t.pausedAt)
	}
	if wall < 0 {
		wall = 0
	}
	scaled := time.Duration(float64(wall) * t.speed)
	return t.baseOffset + scaled
}

// Pause freezes the clock.  Idempotent.
func (t *Timeline) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.paused {
		return
	}
	t.paused = true
	t.pausedAt = Now()
}

// Resume restarts the clock.  Idempotent.
func (t *Timeline) Resume() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.paused {
		return
	}
	t.pausedAccum += Now().Sub(t.pausedAt)
	t.paused = false
}

// Paused reports the pause state.
func (t *Timeline) Paused() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.paused
}

// Speed returns the current speed multiplier.
func (t *Timeline) Speed() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.speed
}

// SetSpeed updates the speed multiplier.  The current Elapsed() value
// is preserved at the moment of the change so animations don't jump.
func (t *Timeline) SetSpeed(s float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	current := t.elapsedLocked()
	t.speed = clampSpeed(s)
	t.startedAt = Now()
	t.pausedAccum = 0
	if t.paused {
		t.pausedAt = Now()
	}
	t.baseOffset = current
}

// Seek jumps the timeline to an absolute cinematic time.  Negative
// values clamp to zero.
func (t *Timeline) Seek(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if d < 0 {
		d = 0
	}
	t.startedAt = Now()
	t.pausedAccum = 0
	if t.paused {
		t.pausedAt = Now()
	}
	t.baseOffset = d
}

// Reset returns the timeline to its initial state at t=0, playing,
// with the same speed.
func (t *Timeline) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.startedAt = Now()
	t.pausedAccum = 0
	t.paused = false
	t.baseOffset = 0
}

func clampSpeed(s float64) float64 {
	if s < MinSpeedMul {
		return MinSpeedMul
	}
	if s > MaxSpeedMul {
		return MaxSpeedMul
	}
	return s
}
