package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

func fixtureCommits() []CommitWithDiff {
	t1, _ := time.Parse(time.RFC3339, "2026-05-01T12:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2026-05-02T12:00:00Z")
	return []CommitWithDiff{
		{
			Commit: gitpkg.Commit{
				Hash:      "abc1234567890",
				Author:    "Alice",
				Email:     "alice@example.com",
				Message:   "feat: add hello\n\nlonger body line.",
				Timestamp: t1,
			},
			Files: []gitpkg.ChangedFile{
				{
					Path: "hello.go", Mode: gitpkg.Added,
					Hunks: []gitpkg.Hunk{{
						OldStart: 0, OldLines: 0, NewStart: 1, NewLines: 2,
						Lines: []gitpkg.HunkLine{
							{Kind: gitpkg.LineAdded, Text: "package main"},
							{Kind: gitpkg.LineAdded, Text: `func Hello() string { return "v1" }`},
						},
					}},
				},
			},
		},
		{
			Commit: gitpkg.Commit{
				Hash:      "def4567890abc",
				Author:    "Bob",
				Email:     "bob@example.com",
				Message:   "fix: tweak hello",
				Timestamp: t2,
			},
			Files: []gitpkg.ChangedFile{
				{
					Path: "hello.go", Mode: gitpkg.Modified,
					Hunks: []gitpkg.Hunk{{
						OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
						Lines: []gitpkg.HunkLine{
							{Kind: gitpkg.LineDeleted, Text: `func Hello() string { return "v1" }`},
							{Kind: gitpkg.LineAdded, Text: `func Hello() string { return "v2" }`},
						},
					}},
				},
			},
		},
	}
}

// ----- Markdown -----

func TestMarkdown_EmptyCommits(t *testing.T) {
	var buf bytes.Buffer
	if err := Markdown(&buf, nil, Options{RepoName: "demo"}); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "demo") || !strings.Contains(s, "no commits") {
		t.Errorf("empty markdown missing header or marker:\n%s", s)
	}
}

func TestMarkdown_RendersCommitSubjectsAndDiffs(t *testing.T) {
	var buf bytes.Buffer
	if err := Markdown(&buf, fixtureCommits(), Options{RepoName: "demo"}); err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	for _, want := range []string{
		"# Replay of demo",
		"## feat: add hello",
		"## fix: tweak hello",
		"### ADDED `hello.go`",
		"### MODIFIED `hello.go`",
		"```go",
		"+ package main",
		`+ func Hello() string { return "v2" }`,
		`- func Hello() string { return "v1" }`,
		"---",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("Markdown missing %q in:\n%s", want, s)
		}
	}
}

func TestMarkdown_ExcludeBodyOmitsBody(t *testing.T) {
	var buf bytes.Buffer
	if err := Markdown(&buf, fixtureCommits(), Options{ExcludeBody: true}); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if strings.Contains(s, "longer body line") {
		t.Errorf("ExcludeBody did not strip the body:\n%s", s)
	}
}

func TestMarkdown_TruncatesLargeHunks(t *testing.T) {
	lines := make([]gitpkg.HunkLine, 50)
	for i := range lines {
		lines[i] = gitpkg.HunkLine{Kind: gitpkg.LineAdded, Text: "line"}
	}
	commits := []CommitWithDiff{{
		Commit: gitpkg.Commit{Hash: "deadbeef", Author: "A", Message: "big", Timestamp: time.Unix(0, 0)},
		Files:  []gitpkg.ChangedFile{{Path: "big.txt", Mode: gitpkg.Added, Hunks: []gitpkg.Hunk{{NewStart: 1, NewLines: 50, Lines: lines}}}},
	}}
	var buf bytes.Buffer
	if err := Markdown(&buf, commits, Options{MaxHunkLines: 5}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "45 more line(s) truncated") {
		t.Errorf("missing truncation marker; got:\n%s", buf.String())
	}
}

func TestMarkdown_BinaryFile(t *testing.T) {
	commits := []CommitWithDiff{{
		Commit: gitpkg.Commit{Hash: "x", Author: "A", Message: "b", Timestamp: time.Unix(0, 0)},
		Files:  []gitpkg.ChangedFile{{Path: "logo.png", Mode: gitpkg.Modified, Binary: true}},
	}}
	var buf bytes.Buffer
	if err := Markdown(&buf, commits, Options{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "binary file") {
		t.Errorf("missing binary placeholder; got:\n%s", buf.String())
	}
}

func TestMarkdown_RenamedShowsArrow(t *testing.T) {
	commits := []CommitWithDiff{{
		Commit: gitpkg.Commit{Hash: "x", Author: "A", Message: "r", Timestamp: time.Unix(0, 0)},
		Files:  []gitpkg.ChangedFile{{Path: "new.txt", OldPath: "old.txt", Mode: gitpkg.Renamed}},
	}}
	var buf bytes.Buffer
	if err := Markdown(&buf, commits, Options{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "old.txt` → `new.txt") {
		t.Errorf("missing rename arrow; got:\n%s", buf.String())
	}
}

func TestMarkdownLangHint(t *testing.T) {
	cases := map[string]string{
		"foo.go":    "go",
		"foo.ts":    "ts",
		"foo.tsx":   "ts",
		"foo.py":    "python",
		"foo.php":   "php",
		"foo.weird": "",
		"NOEXT":     "",
	}
	for path, want := range cases {
		if got := markdownLangHint(path); got != want {
			t.Errorf("markdownLangHint(%q) = %q, want %q", path, got, want)
		}
	}
}

// ----- JSON -----

func TestJSON_EmptyCommits_HasSchemaAndEmptyArray(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, nil, Options{RepoName: "demo"}); err != nil {
		t.Fatal(err)
	}
	var r JSONReport
	if err := json.Unmarshal(buf.Bytes(), &r); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if r.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", r.SchemaVersion, SchemaVersion)
	}
	if r.Repo != "demo" {
		t.Errorf("Repo = %q, want demo", r.Repo)
	}
	if len(r.Commits) != 0 {
		t.Errorf("Commits len = %d, want 0", len(r.Commits))
	}
}

func TestJSON_RealCommits_SerializeShape(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, fixtureCommits(), Options{RepoName: "demo"}); err != nil {
		t.Fatal(err)
	}

	var r JSONReport
	if err := json.Unmarshal(buf.Bytes(), &r); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(r.Commits) != 2 {
		t.Fatalf("Commits len = %d, want 2", len(r.Commits))
	}
	if r.Commits[0].Subject != "feat: add hello" {
		t.Errorf("commit[0].Subject = %q", r.Commits[0].Subject)
	}
	if r.Commits[0].Body != "longer body line." {
		t.Errorf("commit[0].Body = %q", r.Commits[0].Body)
	}
	if len(r.Commits[1].Files) != 1 {
		t.Fatalf("commit[1].Files len = %d, want 1", len(r.Commits[1].Files))
	}
	if r.Commits[1].Files[0].Path != "hello.go" || r.Commits[1].Files[0].Mode != "modified" {
		t.Errorf("commit[1].Files[0] = %+v", r.Commits[1].Files[0])
	}
	if len(r.Commits[1].Files[0].Hunks) == 0 {
		t.Fatal("expected hunks on modified file")
	}
	kinds := []string{}
	for _, l := range r.Commits[1].Files[0].Hunks[0].Lines {
		kinds = append(kinds, l.Kind)
	}
	if want := []string{"-", "+"}; !equalStringSlices(kinds, want) {
		t.Errorf("line kinds = %v, want %v", kinds, want)
	}
}

func TestJSON_GeneratedAt_IncludedWhenSet(t *testing.T) {
	var buf bytes.Buffer
	gen := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	if err := JSON(&buf, fixtureCommits(), Options{GeneratedAt: gen}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"generated_at": "2026-05-27T12:00:00Z"`) {
		t.Errorf("expected generated_at in output:\n%s", buf.String())
	}
}

func TestJSON_TruncationAppendsSyntheticLine(t *testing.T) {
	lines := make([]gitpkg.HunkLine, 50)
	for i := range lines {
		lines[i] = gitpkg.HunkLine{Kind: gitpkg.LineAdded, Text: "line"}
	}
	commits := []CommitWithDiff{{
		Commit: gitpkg.Commit{Hash: "x", Author: "A", Message: "big", Timestamp: time.Unix(0, 0)},
		Files:  []gitpkg.ChangedFile{{Path: "big.txt", Mode: gitpkg.Added, Hunks: []gitpkg.Hunk{{NewStart: 1, NewLines: 50, Lines: lines}}}},
	}}

	r := BuildJSONReport(commits, Options{MaxHunkLines: 3})
	if len(r.Commits[0].Files[0].Hunks[0].Lines) != 4 { // 3 kept + 1 placeholder
		t.Fatalf("line count = %d, want 4", len(r.Commits[0].Files[0].Hunks[0].Lines))
	}
	last := r.Commits[0].Files[0].Hunks[0].Lines[3]
	if !strings.Contains(last.Text, "47 more lines truncated") {
		t.Errorf("expected truncation placeholder, got %q", last.Text)
	}
}

func TestJSON_ExcludeBody(t *testing.T) {
	r := BuildJSONReport(fixtureCommits(), Options{ExcludeBody: true})
	if r.Commits[0].Body != "" {
		t.Errorf("expected empty body, got %q", r.Commits[0].Body)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
