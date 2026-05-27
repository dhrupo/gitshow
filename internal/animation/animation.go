// Package animation owns the timing math for gitshow's cinematic
// effects.  Every function here is pure: it takes a duration (the
// time elapsed since the animation started) and a total duration,
// and returns the current state of the effect.  This keeps the
// Bubble Tea Update loop trivial and the tests fast.
package animation

import (
	"strings"
	"time"
	"unicode/utf8"
)

// Progress returns a normalized 0..1 progress value, clamped at both
// ends so callers don't have to think about overshoot or negative
// elapsed durations.
func Progress(elapsed, total time.Duration) float64 {
	if total <= 0 {
		return 1.0
	}
	if elapsed <= 0 {
		return 0.0
	}
	if elapsed >= total {
		return 1.0
	}
	return float64(elapsed) / float64(total)
}

// Typewriter returns the rune-prefix of `text` that should be visible
// after `elapsed` of a `total` reveal.  At elapsed=0 the result is
// empty; at elapsed>=total the full text is returned.  The math
// counts runes, so multi-byte characters (emoji, CJK) reveal as one
// unit.
func Typewriter(text string, elapsed, total time.Duration) string {
	if text == "" {
		return ""
	}
	if elapsed <= 0 {
		return ""
	}
	totalRunes := utf8.RuneCountInString(text)
	if elapsed >= total || total <= 0 {
		return text
	}
	revealed := int(float64(totalRunes) * Progress(elapsed, total))
	if revealed <= 0 {
		return ""
	}
	if revealed >= totalRunes {
		return text
	}
	count := 0
	for i := range text {
		if count == revealed {
			return text[:i]
		}
		count++
	}
	return text
}

// Stagger reports how many of `items` are currently visible given
// `elapsed` time, with each item taking `perItem` to reveal.  Useful
// for revealing diff lines one at a time.  Returned value is clamped
// to [0, len(items)].
func Stagger(elapsed time.Duration, perItem time.Duration, items int) int {
	if items <= 0 || perItem <= 0 {
		return 0
	}
	if elapsed <= 0 {
		return 0
	}
	n := int(elapsed / perItem)
	if n < 0 {
		return 0
	}
	if n > items {
		return items
	}
	return n
}

// StaggerLines is a tiny convenience wrapper that joins the first
// Stagger(...) lines of `lines` with newlines.  Useful when the
// caller doesn't need the integer count itself.
func StaggerLines(lines []string, elapsed, perItem time.Duration) string {
	n := Stagger(elapsed, perItem, len(lines))
	if n == 0 {
		return ""
	}
	return strings.Join(lines[:n], "\n")
}

// EaseInOutCubic is a smooth easing curve in [0,1].  Phase 4 keeps it
// as a primitive for future fade / slide transitions even though the
// MVP doesn't use it heavily yet.
func EaseInOutCubic(t float64) float64 {
	switch {
	case t <= 0:
		return 0
	case t >= 1:
		return 1
	case t < 0.5:
		return 4 * t * t * t
	default:
		return 1 - pow3(-2*t+2)/2
	}
}

func pow3(x float64) float64 { return x * x * x }
