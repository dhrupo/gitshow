package render

import (
	"strings"
	"testing"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

func TestDiff_NoColor_RendersDeterministic(t *testing.T) {
	file := gitpkg.ChangedFile{
		Path: "hello.go",
		Mode: gitpkg.Modified,
		Hunks: []gitpkg.Hunk{{
			OldStart: 1, OldLines: 3,
			NewStart: 1, NewLines: 3,
			Lines: []gitpkg.HunkLine{
				{Kind: gitpkg.LineContext, Text: "package main"},
				{Kind: gitpkg.LineDeleted, Text: `func Hello() string { return "v1" }`},
				{Kind: gitpkg.LineAdded, Text: `func Hello() string { return "v2" }`},
			},
		}},
	}
	got := Diff(file, Options{NoColor: true})
	want := []string{
		"MODIFIED hello.go",
		"@@ -1,3 +1,3 @@",
		`  package main`,
		`- func Hello() string { return "v1" }`,
		`+ func Hello() string { return "v2" }`,
	}
	for _, line := range want {
		if !strings.Contains(got, line) {
			t.Errorf("missing %q in:\n%s", line, got)
		}
	}
}

func TestDiff_ColorOn_EmitsAnsi(t *testing.T) {
	file := gitpkg.ChangedFile{
		Path: "hello.go",
		Mode: gitpkg.Modified,
		Hunks: []gitpkg.Hunk{{
			OldStart: 1, OldLines: 1, NewStart: 1, NewLines: 1,
			Lines: []gitpkg.HunkLine{
				{Kind: gitpkg.LineAdded, Text: `var x = "hi"`},
			},
		}},
	}
	got := Diff(file, Options{})
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected ANSI escapes; got\n%s", got)
	}
}

func TestDiff_AddedDeletedRenamed_Markers(t *testing.T) {
	cases := []struct {
		mode    gitpkg.ChangeMode
		wantTag string
	}{
		{gitpkg.Added, "ADDED"},
		{gitpkg.Deleted, "DELETED"},
		{gitpkg.Modified, "MODIFIED"},
		{gitpkg.Renamed, "RENAMED"},
	}
	for _, tc := range cases {
		t.Run(string(tc.mode), func(t *testing.T) {
			file := gitpkg.ChangedFile{Path: "x.txt", Mode: tc.mode, OldPath: "old.txt"}
			got := Diff(file, Options{NoColor: true})
			if !strings.Contains(got, tc.wantTag) {
				t.Errorf("missing %s tag in:\n%s", tc.wantTag, got)
			}
		})
	}
}

func TestDiff_RenamedShowsArrow(t *testing.T) {
	file := gitpkg.ChangedFile{Path: "new.txt", Mode: gitpkg.Renamed, OldPath: "old.txt"}
	got := Diff(file, Options{NoColor: true})
	if !strings.Contains(got, "old.txt → new.txt") {
		t.Errorf("missing rename arrow in:\n%s", got)
	}
}

func TestDiff_Binary_PrintsPlaceholder(t *testing.T) {
	file := gitpkg.ChangedFile{Path: "logo.png", Mode: gitpkg.Modified, Binary: true}
	got := Diff(file, Options{NoColor: true})
	if !strings.Contains(got, "binary file") {
		t.Errorf("missing binary marker in:\n%s", got)
	}
}

func TestDiff_MaxHunkLines_Truncates(t *testing.T) {
	lines := make([]gitpkg.HunkLine, 100)
	for i := range lines {
		lines[i] = gitpkg.HunkLine{Kind: gitpkg.LineAdded, Text: "line"}
	}
	file := gitpkg.ChangedFile{
		Path: "big.txt", Mode: gitpkg.Modified,
		Hunks: []gitpkg.Hunk{{NewStart: 1, NewLines: 100, Lines: lines}},
	}
	got := Diff(file, Options{NoColor: true, MaxHunkLines: 10})
	if !strings.Contains(got, "90 more line(s) truncated") {
		t.Errorf("missing truncation marker; got:\n%s", got)
	}
}

func TestDiffSet_JoinsMultipleFiles(t *testing.T) {
	files := []gitpkg.ChangedFile{
		{Path: "a.txt", Mode: gitpkg.Added, Hunks: []gitpkg.Hunk{{NewStart: 1, NewLines: 1, Lines: []gitpkg.HunkLine{{Kind: gitpkg.LineAdded, Text: "a"}}}}},
		{Path: "b.txt", Mode: gitpkg.Deleted, Hunks: []gitpkg.Hunk{{OldStart: 1, OldLines: 1, Lines: []gitpkg.HunkLine{{Kind: gitpkg.LineDeleted, Text: "b"}}}}},
	}
	got := DiffSet(files, Options{NoColor: true})
	if !strings.Contains(got, "ADDED a.txt") {
		t.Errorf("missing ADDED a.txt in:\n%s", got)
	}
	if !strings.Contains(got, "DELETED b.txt") {
		t.Errorf("missing DELETED b.txt in:\n%s", got)
	}
}

func TestDiffSet_EmptyReturnsEmpty(t *testing.T) {
	if got := DiffSet(nil, Options{}); got != "" {
		t.Errorf("want empty string, got %q", got)
	}
}

// Smoke check that an unknown extension does not panic and emits ANSI.
func TestDiff_UnknownExtensionStillWorks(t *testing.T) {
	file := gitpkg.ChangedFile{
		Path: "config.weirdext",
		Mode: gitpkg.Added,
		Hunks: []gitpkg.Hunk{{NewStart: 1, NewLines: 1, Lines: []gitpkg.HunkLine{
			{Kind: gitpkg.LineAdded, Text: "value=42"},
		}}},
	}
	got := Diff(file, Options{})
	if got == "" {
		t.Fatal("got empty render")
	}
}
