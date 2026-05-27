package animation

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestProgress(t *testing.T) {
	cases := []struct {
		elapsed, total time.Duration
		want           float64
	}{
		{0, time.Second, 0.0},
		{500 * time.Millisecond, time.Second, 0.5},
		{time.Second, time.Second, 1.0},
		{2 * time.Second, time.Second, 1.0}, // overshoot clamp
		{-time.Second, time.Second, 0.0},   // negative clamp
		{500 * time.Millisecond, 0, 1.0},   // zero total ⇒ done
	}
	for _, tc := range cases {
		got := Progress(tc.elapsed, tc.total)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("Progress(%v,%v) = %v, want %v", tc.elapsed, tc.total, got, tc.want)
		}
	}
}

func TestTypewriter_Boundaries(t *testing.T) {
	text := "hello"
	cases := []struct {
		elapsed, total time.Duration
		want           string
	}{
		{0, time.Second, ""},
		{time.Second, time.Second, "hello"},
		{2 * time.Second, time.Second, "hello"},
		{-time.Second, time.Second, ""},
	}
	for _, tc := range cases {
		got := Typewriter(text, tc.elapsed, tc.total)
		if got != tc.want {
			t.Errorf("Typewriter(elapsed=%v) = %q, want %q", tc.elapsed, got, tc.want)
		}
	}
}

func TestTypewriter_MidwayReveal(t *testing.T) {
	text := "hello world" // 11 runes
	got := Typewriter(text, 500*time.Millisecond, time.Second)
	// At 0.5 progress, reveal floor(11*0.5) = 5 runes.
	if got != "hello" {
		t.Errorf("midway typewriter = %q, want %q", got, "hello")
	}
}

func TestTypewriter_EmojiCountsAsOneRune(t *testing.T) {
	text := "ab🚀cd"             // 5 runes, but multi-byte
	got := Typewriter(text, 60*time.Millisecond, 100*time.Millisecond)
	// 0.6 * 5 = 3.0 → reveal 3 runes "ab🚀".
	want := "ab🚀"
	if got != want {
		t.Errorf("emoji typewriter = %q, want %q", got, want)
	}
}

func TestTypewriter_EmptyText(t *testing.T) {
	if got := Typewriter("", time.Second, time.Second); got != "" {
		t.Errorf("Typewriter('') = %q", got)
	}
}

func TestStagger(t *testing.T) {
	per := 100 * time.Millisecond
	cases := []struct {
		elapsed time.Duration
		items   int
		want    int
	}{
		{0, 5, 0},
		{50 * time.Millisecond, 5, 0},
		{100 * time.Millisecond, 5, 1},
		{350 * time.Millisecond, 5, 3},
		{time.Second, 5, 5},          // clamp at items
		{2 * time.Second, 5, 5},      // overshoot
		{-time.Second, 5, 0},         // negative
		{time.Second, 0, 0},          // zero items
		{time.Second, -3, 0},         // negative items
	}
	for _, tc := range cases {
		got := Stagger(tc.elapsed, per, tc.items)
		if got != tc.want {
			t.Errorf("Stagger(elapsed=%v, items=%d) = %d, want %d", tc.elapsed, tc.items, got, tc.want)
		}
	}
}

func TestStagger_ZeroPerItemReturnsZero(t *testing.T) {
	if got := Stagger(time.Second, 0, 5); got != 0 {
		t.Errorf("Stagger(perItem=0) = %d, want 0", got)
	}
}

func TestStaggerLines_JoinsRevealedLines(t *testing.T) {
	lines := []string{"a", "b", "c", "d"}
	got := StaggerLines(lines, 250*time.Millisecond, 100*time.Millisecond)
	want := strings.Join(lines[:2], "\n")
	if got != want {
		t.Errorf("StaggerLines = %q, want %q", got, want)
	}
}

func TestStaggerLines_AllRevealedAfterEnoughTime(t *testing.T) {
	lines := []string{"a", "b", "c"}
	got := StaggerLines(lines, 10*time.Second, 100*time.Millisecond)
	if got != "a\nb\nc" {
		t.Errorf("StaggerLines (all) = %q", got)
	}
}

func TestEaseInOutCubic(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{-0.1, 0},
		{0, 0},
		{0.5, 0.5},
		{1, 1},
		{1.5, 1},
	}
	for _, tc := range cases {
		got := EaseInOutCubic(tc.in)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("EaseInOutCubic(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
	// Symmetric around 0.5.
	a := EaseInOutCubic(0.25)
	b := 1 - EaseInOutCubic(0.75)
	if math.Abs(a-b) > 1e-9 {
		t.Errorf("EaseInOutCubic not symmetric: e(0.25)=%v vs 1-e(0.75)=%v", a, b)
	}
}
