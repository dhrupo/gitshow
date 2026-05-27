package git

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// DiffFor returns the files changed by `commit` vs its first parent.
// For root commits (no parents), every file in the tree is reported as
// Added.
func (r *Repo) DiffFor(commit Commit) ([]ChangedFile, error) {
	hash := plumbing.NewHash(commit.Hash)
	c, err := r.r.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("commit %s: %w", commit.Hash, err)
	}

	if c.NumParents() == 0 {
		return rootDiff(c)
	}

	parent, err := c.Parent(0)
	if err != nil {
		return nil, fmt.Errorf("parent of %s: %w", commit.Hash, err)
	}
	patch, err := parent.Patch(c)
	if err != nil {
		return nil, fmt.Errorf("patch %s..%s: %w", parent.Hash, c.Hash, err)
	}

	return parseGoGitPatch(patch)
}

func rootDiff(c *object.Commit) ([]ChangedFile, error) {
	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}
	var out []ChangedFile
	err = tree.Files().ForEach(func(f *object.File) error {
		body, err := f.Contents()
		if err != nil {
			return err
		}
		out = append(out, ChangedFile{
			Path: f.Name,
			Mode: Added,
			Hunks: []Hunk{{
				NewStart: 1,
				NewLines: countLines(body),
				Lines:    bodyAsAddedLines(body),
			}},
		})
		return nil
	})
	return out, err
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	n := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

func bodyAsAddedLines(s string) []HunkLine {
	if s == "" {
		return nil
	}
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	var lines []HunkLine
	for scanner.Scan() {
		lines = append(lines, HunkLine{Kind: LineAdded, Text: scanner.Text()})
	}
	return lines
}

// parseGoGitPatch converts a go-git Patch into our ChangedFile slice.
// go-git's String() output is a standard unified diff; we walk it
// rather than the object.Patch FilePatches() API because the latter
// strips line-level +/- markers.
func parseGoGitPatch(patch interface{ String() string }) ([]ChangedFile, error) {
	body := patch.String()
	if body == "" {
		return nil, nil
	}
	return parseUnifiedDiff(body)
}

// parseUnifiedDiff is a minimal unified-diff parser.  It handles:
//   - "diff --git a/<path> b/<path>" boundaries
//   - "new file" / "deleted file" / "rename from" / "rename to" markers
//   - "Binary files differ" lines
//   - "--- a/x" / "+++ b/x" pre-hunk headers
//   - "@@ -a,b +c,d @@" hunk headers
//
// It is intentionally lenient: unknown lines are skipped, never panic.
func parseUnifiedDiff(body string) ([]ChangedFile, error) {
	var out []ChangedFile
	var cur *ChangedFile
	var hunk *Hunk

	flushHunk := func() {
		if cur != nil && hunk != nil {
			cur.Hunks = append(cur.Hunks, *hunk)
			hunk = nil
		}
	}
	flushFile := func() {
		flushHunk()
		if cur != nil {
			out = append(out, *cur)
			cur = nil
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flushFile()
			path := pathFromDiffHeader(line)
			cur = &ChangedFile{Path: path, Mode: Modified}
		case cur == nil:
			// Skip prefatory noise.
			continue
		case strings.HasPrefix(line, "new file mode"):
			cur.Mode = Added
		case strings.HasPrefix(line, "deleted file mode"):
			cur.Mode = Deleted
		case strings.HasPrefix(line, "rename from "):
			cur.Mode = Renamed
			cur.OldPath = strings.TrimPrefix(line, "rename from ")
		case strings.HasPrefix(line, "rename to "):
			cur.Path = strings.TrimPrefix(line, "rename to ")
		case strings.HasPrefix(line, "Binary files "):
			cur.Binary = true
		case strings.HasPrefix(line, "--- "):
			// Track old path if we couldn't pull it from the diff header.
			if strings.HasPrefix(line, "--- a/") && cur.OldPath == "" {
				cur.OldPath = strings.TrimPrefix(line, "--- a/")
			}
		case strings.HasPrefix(line, "+++ "):
			if strings.HasPrefix(line, "+++ b/") && cur.Path == "" {
				cur.Path = strings.TrimPrefix(line, "+++ b/")
			}
		case strings.HasPrefix(line, "@@"):
			flushHunk()
			h, err := parseHunkHeader(line)
			if err == nil {
				hunk = h
			}
		case hunk != nil && len(line) > 0:
			switch line[0] {
			case '+':
				hunk.Lines = append(hunk.Lines, HunkLine{Kind: LineAdded, Text: line[1:]})
			case '-':
				hunk.Lines = append(hunk.Lines, HunkLine{Kind: LineDeleted, Text: line[1:]})
			case ' ':
				hunk.Lines = append(hunk.Lines, HunkLine{Kind: LineContext, Text: line[1:]})
			}
		case hunk != nil && len(line) == 0:
			hunk.Lines = append(hunk.Lines, HunkLine{Kind: LineContext, Text: ""})
		}
	}
	flushFile()
	return out, nil
}

// pathFromDiffHeader pulls the right-hand-side path out of
// "diff --git a/<path> b/<path>".  Quoted paths (with spaces) are
// handled by walking from " b/" backward.
func pathFromDiffHeader(line string) string {
	if i := strings.Index(line, " b/"); i >= 0 {
		return strings.TrimSpace(line[i+len(" b/"):])
	}
	return ""
}

// parseHunkHeader parses "@@ -a,b +c,d @@ ..."  where the ",b" and ",d"
// are optional (default 1).
func parseHunkHeader(line string) (*Hunk, error) {
	rest := strings.TrimPrefix(line, "@@")
	end := strings.Index(rest, "@@")
	if end < 0 {
		return nil, fmt.Errorf("malformed hunk header: %q", line)
	}
	parts := strings.Fields(rest[:end])
	if len(parts) < 2 {
		return nil, fmt.Errorf("hunk header missing ranges: %q", line)
	}
	oldStart, oldLines, err := parseHunkRange(parts[0])
	if err != nil {
		return nil, err
	}
	newStart, newLines, err := parseHunkRange(parts[1])
	if err != nil {
		return nil, err
	}
	return &Hunk{
		OldStart: oldStart,
		OldLines: oldLines,
		NewStart: newStart,
		NewLines: newLines,
	}, nil
}

func parseHunkRange(s string) (start, lines int, err error) {
	// "-a,b" or "+a,b" or "-a" or "+a"
	if len(s) < 2 {
		return 0, 0, fmt.Errorf("hunk range too short: %q", s)
	}
	body := s[1:] // strip the leading - or +
	if i := strings.IndexByte(body, ','); i >= 0 {
		start, err = strconv.Atoi(body[:i])
		if err != nil {
			return 0, 0, err
		}
		lines, err = strconv.Atoi(body[i+1:])
		return start, lines, err
	}
	start, err = strconv.Atoi(body)
	return start, 1, err
}
