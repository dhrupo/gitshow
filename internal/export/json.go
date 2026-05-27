package export

import (
	"encoding/json"
	"io"
	"strings"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

// JSONReport is the top-level schema written by JSON exports.  Fields
// use snake_case for consumer friendliness; struct tags ensure that
// regardless of Go field naming.
type JSONReport struct {
	SchemaVersion int          `json:"schema_version"`
	Repo          string       `json:"repo,omitempty"`
	GeneratedAt   string       `json:"generated_at,omitempty"`
	Commits       []JSONCommit `json:"commits"`
}

// JSONCommit is the per-commit shape.
type JSONCommit struct {
	Hash      string     `json:"hash"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body,omitempty"`
	Author    string     `json:"author"`
	Email     string     `json:"email,omitempty"`
	Timestamp string     `json:"timestamp"`
	Files     []JSONFile `json:"files"`
}

// JSONFile mirrors gitpkg.ChangedFile with explicit JSON tags.
type JSONFile struct {
	Path    string     `json:"path"`
	OldPath string     `json:"old_path,omitempty"`
	Mode    string     `json:"mode"`
	Binary  bool       `json:"binary,omitempty"`
	Hunks   []JSONHunk `json:"hunks,omitempty"`
}

// JSONHunk mirrors gitpkg.Hunk.
type JSONHunk struct {
	OldStart int        `json:"old_start"`
	OldLines int        `json:"old_lines"`
	NewStart int        `json:"new_start"`
	NewLines int        `json:"new_lines"`
	Lines    []JSONLine `json:"lines"`
}

// JSONLine mirrors gitpkg.HunkLine with a string Kind.
type JSONLine struct {
	Kind string `json:"kind"` // "+", "-", or " "
	Text string `json:"text"`
}

// JSON writes a JSONReport encoded as indented JSON to w.
func JSON(w io.Writer, commits []CommitWithDiff, opts Options) error {
	report := BuildJSONReport(commits, opts)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// BuildJSONReport constructs the JSONReport without serializing it.
// Useful for tests that want to assert against the shape directly.
func BuildJSONReport(commits []CommitWithDiff, opts Options) JSONReport {
	out := JSONReport{
		SchemaVersion: SchemaVersion,
		Repo:          opts.RepoName,
		Commits:       []JSONCommit{},
	}
	if !opts.GeneratedAt.IsZero() {
		out.GeneratedAt = opts.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	for _, c := range commits {
		jc := JSONCommit{
			Hash:      c.Commit.Hash,
			Subject:   c.Commit.Subject(),
			Author:    c.Commit.Author,
			Email:     c.Commit.Email,
			Timestamp: c.Commit.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
			Files:     []JSONFile{},
		}
		if !opts.ExcludeBody {
			body := strings.TrimSpace(strings.TrimPrefix(c.Commit.Message, c.Commit.Subject()))
			jc.Body = body
		}
		for _, f := range c.Files {
			jc.Files = append(jc.Files, buildJSONFile(f, opts))
		}
		out.Commits = append(out.Commits, jc)
	}
	return out
}

func buildJSONFile(f gitpkg.ChangedFile, opts Options) JSONFile {
	jf := JSONFile{
		Path:    f.Path,
		OldPath: f.OldPath,
		Mode:    string(f.Mode),
		Binary:  f.Binary,
		Hunks:   []JSONHunk{},
	}
	for _, h := range f.Hunks {
		jf.Hunks = append(jf.Hunks, buildJSONHunk(h, opts))
	}
	return jf
}

func buildJSONHunk(h gitpkg.Hunk, opts Options) JSONHunk {
	jh := JSONHunk{
		OldStart: h.OldStart,
		OldLines: h.OldLines,
		NewStart: h.NewStart,
		NewLines: h.NewLines,
		Lines:    []JSONLine{},
	}
	limit := len(h.Lines)
	truncated := 0
	if opts.MaxHunkLines > 0 && limit > opts.MaxHunkLines {
		truncated = limit - opts.MaxHunkLines
		limit = opts.MaxHunkLines
	}
	for i := 0; i < limit; i++ {
		jh.Lines = append(jh.Lines, JSONLine{
			Kind: lineKindToString(h.Lines[i].Kind),
			Text: h.Lines[i].Text,
		})
	}
	if truncated > 0 {
		jh.Lines = append(jh.Lines, JSONLine{
			Kind: " ",
			Text: truncatedPlaceholder(truncated),
		})
	}
	return jh
}

func lineKindToString(kind gitpkg.LineKind) string {
	switch kind {
	case gitpkg.LineAdded:
		return "+"
	case gitpkg.LineDeleted:
		return "-"
	default:
		return " "
	}
}

func truncatedPlaceholder(n int) string {
	if n == 1 {
		return "... 1 more line truncated"
	}
	// Simple sprintf-equivalent without importing fmt twice.
	return formatTruncated(n)
}

func formatTruncated(n int) string {
	// Avoid fmt.Sprintf to keep this file fmt-free; tiny manual itoa.
	return "... " + intoa(n) + " more lines truncated"
}

func intoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
