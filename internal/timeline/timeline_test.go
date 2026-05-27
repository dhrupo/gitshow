package timeline

import (
	"sync"
	"testing"
	"time"
)

// fakeClock advances only when Tick() is called.  Lets us drive the
// timeline deterministically without time.Sleep.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Tick(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func withFakeClock(t *testing.T) *fakeClock {
	t.Helper()
	clock := &fakeClock{now: time.Unix(1_700_000_000, 0)}
	prev := Now
	Now = clock.Now
	t.Cleanup(func() { Now = prev })
	return clock
}

func TestNew_StartsAtZero(t *testing.T) {
	withFakeClock(t)
	tl := New(1.0)
	if tl.Elapsed() != 0 {
		t.Errorf("fresh Elapsed = %v, want 0", tl.Elapsed())
	}
}

func TestElapsed_AdvancesAtSpeedOne(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(2 * time.Second)
	got := tl.Elapsed()
	if got != 2*time.Second {
		t.Errorf("Elapsed = %v, want 2s", got)
	}
}

func TestElapsed_ScalesWithSpeed(t *testing.T) {
	c := withFakeClock(t)
	tl := New(2.0)
	c.Tick(time.Second)
	if tl.Elapsed() != 2*time.Second {
		t.Errorf("at speed=2, 1s wall = %v, want 2s", tl.Elapsed())
	}
}

func TestPause_FreezesClock(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(time.Second)
	tl.Pause()
	c.Tick(5 * time.Second)
	if tl.Elapsed() != time.Second {
		t.Errorf("paused Elapsed = %v, want 1s", tl.Elapsed())
	}
}

func TestResume_ContinuesFromPauseTime(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(time.Second)
	tl.Pause()
	c.Tick(5 * time.Second)
	tl.Resume()
	c.Tick(2 * time.Second)
	// 1s before pause + 2s after = 3s total
	if got := tl.Elapsed(); got != 3*time.Second {
		t.Errorf("post-resume Elapsed = %v, want 3s", got)
	}
}

func TestPauseIsIdempotent(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(time.Second)
	tl.Pause()
	c.Tick(time.Second)
	tl.Pause() // second pause must not double-count
	c.Tick(time.Second)
	tl.Resume()
	if got := tl.Elapsed(); got != time.Second {
		t.Errorf("double-pause Elapsed = %v, want 1s", got)
	}
}

func TestSetSpeed_PreservesCurrentElapsed(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(time.Second)
	before := tl.Elapsed()
	tl.SetSpeed(2.0)
	after := tl.Elapsed()
	if before != after {
		t.Errorf("SetSpeed jumped elapsed: before=%v after=%v", before, after)
	}
	c.Tick(time.Second)
	if got := tl.Elapsed(); got != 3*time.Second {
		// 1s (at 1x) + 2s (at 2x) = 3s
		t.Errorf("post-SetSpeed Elapsed = %v, want 3s", got)
	}
}

func TestSetSpeed_Clamps(t *testing.T) {
	withFakeClock(t)
	tl := New(1.0)
	tl.SetSpeed(99)
	if tl.Speed() != MaxSpeedMul {
		t.Errorf("speed = %v, want %v", tl.Speed(), MaxSpeedMul)
	}
	tl.SetSpeed(0.001)
	if tl.Speed() != MinSpeedMul {
		t.Errorf("speed = %v, want %v", tl.Speed(), MinSpeedMul)
	}
}

func TestSeek_JumpsToAbsolute(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(time.Second)
	tl.Seek(10 * time.Second)
	if got := tl.Elapsed(); got != 10*time.Second {
		t.Errorf("immediately after Seek: %v, want 10s", got)
	}
	c.Tick(2 * time.Second)
	if got := tl.Elapsed(); got != 12*time.Second {
		t.Errorf("after Seek+tick: %v, want 12s", got)
	}
}

func TestSeek_NegativeClampsToZero(t *testing.T) {
	c := withFakeClock(t)
	tl := New(1.0)
	c.Tick(5 * time.Second)
	tl.Seek(-3 * time.Second)
	if got := tl.Elapsed(); got != 0 {
		t.Errorf("seek(-3s) = %v, want 0", got)
	}
}

func TestReset_RestartsFromZero(t *testing.T) {
	c := withFakeClock(t)
	tl := New(2.0)
	c.Tick(5 * time.Second)
	tl.Reset()
	if got := tl.Elapsed(); got != 0 {
		t.Errorf("post-Reset Elapsed = %v, want 0", got)
	}
	if tl.Speed() != 2.0 {
		t.Errorf("Reset clobbered speed = %v, want 2.0", tl.Speed())
	}
	if tl.Paused() {
		t.Error("Reset should leave timeline playing")
	}
}

func TestNew_ZeroSpeedDefaultsToOne(t *testing.T) {
	withFakeClock(t)
	tl := New(0)
	if tl.Speed() != 1.0 {
		t.Errorf("speed = %v, want 1.0", tl.Speed())
	}
}
