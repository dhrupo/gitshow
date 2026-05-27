package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/dhrupo/gitshow/internal/animation"
	gitpkg "github.com/dhrupo/gitshow/internal/git"
	"github.com/dhrupo/gitshow/internal/render"
)

// View composes the four panels into the final frame.
func (m Model) View() string {
	if m.quitting {
		return "  Thanks for watching.\n"
	}
	if len(m.Commits) == 0 {
		return m.theme.EmptyState.Render("(no commits to replay)\n")
	}
	if m.width == 0 || m.height == 0 {
		// We can still render before the first WindowSizeMsg lands;
		// pick conservative defaults so a TTY without size detection
		// still gets sensible output.
		m.width = 100
		m.height = 30
	}

	header := m.renderHeader()
	main := m.renderMain()
	diff := m.renderDiff()
	timeline := m.renderTimeline()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, diff, timeline)
}

func (m Model) renderHeader() string {
	state := "▶ playing"
	if m.paused {
		state = "⏸ paused"
	}
	pos := fmt.Sprintf("[%d / %d]", m.idx+1, len(m.Commits))
	speedBar := strings.Repeat("●", m.speed) + strings.Repeat("○", MaxSpeed-m.speed)

	left := m.theme.HeaderAccent.Render("gitshow") + "  " + m.theme.HeaderMuted.Render(m.RepoName)
	right := strings.Join([]string{
		m.theme.HeaderMuted.Render(pos),
		state,
		m.theme.HeaderMuted.Render("speed " + speedBar),
	}, "  ")

	width := m.width
	if width < 40 {
		width = 40
	}
	spacing := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if spacing < 1 {
		spacing = 1
	}
	return m.theme.HeaderBar.Width(width).Render(left + strings.Repeat(" ", spacing) + right)
}

func (m Model) renderMain() string {
	c := m.Commits[m.idx].Commit
	short := c.Hash
	if len(short) > 7 {
		short = short[:7]
	}
	subject := strings.TrimSpace(c.Subject())
	body := strings.TrimSpace(strings.TrimPrefix(c.Message, c.Subject()))

	// Typewriter the subject so it reveals letter-by-letter when the
	// camera lands on this commit.  In tests (no clock) we just show
	// the whole subject.
	revealed := subject
	if m.clock != nil {
		revealed = animation.Typewriter(subject, m.commitElapsed(), subjectRevealDuration)
	}
	if revealed == "" {
		revealed = " " // keep card height stable on the first frame
	}

	lines := []string{
		m.theme.CommitSubject.Render(revealed),
		m.theme.CommitMeta.Render(fmt.Sprintf("commit %s  •  %s  •  %s",
			short, c.Author, c.Timestamp.Format("2006-01-02 15:04"))),
	}
	if body != "" {
		lines = append(lines, "", m.theme.HeaderMuted.Render(body))
	}

	card := m.theme.CommitCard.Width(m.width - 2).Render(strings.Join(lines, "\n"))
	return card
}

func (m Model) renderDiff() string {
	files := m.Commits[m.idx].Files
	if len(files) == 0 {
		body := m.theme.EmptyState.Render("(no file changes in this commit)")
		return m.theme.DiffPanel.Width(m.width - 2).Render(body)
	}

	// Reserve approximate vertical space for the other three panels
	// (header 1, main 4–6, timeline 2, borders/padding ~4).
	reserved := 12
	maxHunkLines := m.height - reserved
	if maxHunkLines < 4 {
		maxHunkLines = 4
	}

	body := render.DiffSet(files, render.Options{
		ChromaStyle:  "monokai",
		MaxHunkLines: maxHunkLines,
	})
	body = trimTrailingNewlines(body)

	// Stagger the body: reveal lines one at a time as cinematic time
	// advances.  Once the typewriter has finished spelling the subject
	// we start showing the diff line-by-line.
	if m.clock != nil {
		elapsed := m.commitElapsed() - subjectRevealDuration
		if elapsed > 0 {
			lines := strings.Split(body, "\n")
			body = animation.StaggerLines(lines, elapsed, diffLinePerInterval)
		} else {
			body = ""
		}
	}
	return m.theme.DiffPanel.Width(m.width - 2).Render(body)
}

func (m Model) renderTimeline() string {
	if len(m.Commits) == 0 {
		return ""
	}

	width := m.width - 4
	if width < 10 {
		width = 10
	}
	dots := timelineDots(len(m.Commits), m.idx, width)

	hint := m.theme.HeaderMuted.Render("  ←/→ navigate  •  ↑/↓ speed  •  space pause  •  q quit")
	return m.theme.TimelineFrame.Render(dots) + "\n" + hint
}

func timelineDots(total, current, width int) string {
	if total == 0 {
		return ""
	}
	if width < total {
		// Resample: keep position roughly proportional.
		var b strings.Builder
		for i := 0; i < width; i++ {
			mapped := i * total / width
			switch {
			case mapped == current:
				b.WriteRune('●')
			case mapped < current:
				b.WriteRune('•')
			default:
				b.WriteRune('·')
			}
		}
		return b.String()
	}
	// Plenty of room: pad each commit with a few cells.
	cellsPerCommit := width / total
	if cellsPerCommit < 1 {
		cellsPerCommit = 1
	}
	var b strings.Builder
	for i := 0; i < total; i++ {
		switch {
		case i == current:
			b.WriteRune('●')
		case i < current:
			b.WriteRune('•')
		default:
			b.WriteRune('·')
		}
		for j := 1; j < cellsPerCommit; j++ {
			b.WriteRune(' ')
		}
	}
	return b.String()
}

func trimTrailingNewlines(s string) string {
	return strings.TrimRight(s, "\n")
}

// renderQuit is exposed for completeness but unused — kept here so the
// View() switch reads naturally.
func renderQuit() string {
	return "  Thanks for watching.\n"
}

// Silence unused-import warning if gitpkg is only referenced inside
// build-tag-specific code paths.
var _ gitpkg.Commit
