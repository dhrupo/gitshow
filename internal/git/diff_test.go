package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseUnifiedDiff_Add_Modify_Delete_Rename(t *testing.T) {
	body := `diff --git a/added.txt b/added.txt
new file mode 100644
--- /dev/null
+++ b/added.txt
@@ -0,0 +1,2 @@
+hello
+world
diff --git a/mod.txt b/mod.txt
--- a/mod.txt
+++ b/mod.txt
@@ -1,3 +1,3 @@
 unchanged
-old line
+new line
 trailing
diff --git a/old.txt b/old.txt
deleted file mode 100644
--- a/old.txt
+++ /dev/null
@@ -1,1 +0,0 @@
-bye
diff --git a/oldname.txt b/newname.txt
similarity index 95%
rename from oldname.txt
rename to newname.txt
--- a/oldname.txt
+++ b/newname.txt
@@ -1,2 +1,2 @@
-hello world
+hello world!
 stays
`
	files, err := parseUnifiedDiff(body)
	if err != nil {
		t.Fatalf("parseUnifiedDiff: %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("got %d files, want 4: %#v", len(files), files)
	}

	if files[0].Mode != Added || files[0].Path != "added.txt" {
		t.Errorf("files[0] = %+v, want Added added.txt", files[0])
	}
	if len(files[0].Hunks) != 1 || len(files[0].Hunks[0].Lines) != 2 {
		t.Errorf("added file hunk shape wrong: %#v", files[0].Hunks)
	} else {
		if files[0].Hunks[0].Lines[0].Kind != LineAdded || files[0].Hunks[0].Lines[0].Text != "hello" {
			t.Errorf("added file line[0] = %+v", files[0].Hunks[0].Lines[0])
		}
	}

	if files[1].Mode != Modified || files[1].Path != "mod.txt" {
		t.Errorf("files[1] = %+v, want Modified mod.txt", files[1])
	}
	if len(files[1].Hunks) != 1 {
		t.Fatalf("modified file: want 1 hunk, got %d", len(files[1].Hunks))
	}
	gotKinds := []LineKind{}
	for _, l := range files[1].Hunks[0].Lines {
		gotKinds = append(gotKinds, l.Kind)
	}
	wantKinds := []LineKind{LineContext, LineDeleted, LineAdded, LineContext}
	for i, want := range wantKinds {
		if gotKinds[i] != want {
			t.Errorf("modified file line[%d] kind = %v, want %v", i, gotKinds[i], want)
		}
	}

	if files[2].Mode != Deleted || files[2].Path != "old.txt" {
		t.Errorf("files[2] = %+v, want Deleted old.txt", files[2])
	}

	if files[3].Mode != Renamed || files[3].Path != "newname.txt" || files[3].OldPath != "oldname.txt" {
		t.Errorf("files[3] = %+v, want Renamed oldname.txt -> newname.txt", files[3])
	}
}

func TestParseUnifiedDiff_BinaryFiles(t *testing.T) {
	body := `diff --git a/img.png b/img.png
Binary files a/img.png and b/img.png differ
`
	files, err := parseUnifiedDiff(body)
	if err != nil {
		t.Fatalf("parseUnifiedDiff: %v", err)
	}
	if len(files) != 1 || !files[0].Binary {
		t.Fatalf("expected one binary file, got %#v", files)
	}
}

func TestParseHunkHeader(t *testing.T) {
	cases := []struct {
		line     string
		oS, oL   int
		nS, nL   int
		mustFail bool
	}{
		{"@@ -1,3 +1,3 @@", 1, 3, 1, 3, false},
		{"@@ -10 +12,5 @@ context", 10, 1, 12, 5, false},
		{"@@ -0,0 +1,2 @@", 0, 0, 1, 2, false},
		{"@@ broken", 0, 0, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.line, func(t *testing.T) {
			h, err := parseHunkHeader(tc.line)
			if tc.mustFail {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHunkHeader: %v", err)
			}
			if h.OldStart != tc.oS || h.OldLines != tc.oL || h.NewStart != tc.nS || h.NewLines != tc.nL {
				t.Errorf("got -%d,%d +%d,%d; want -%d,%d +%d,%d",
					h.OldStart, h.OldLines, h.NewStart, h.NewLines,
					tc.oS, tc.oL, tc.nS, tc.nL)
			}
		})
	}
}

// Integration: a real go-git-backed repo with one file mutation.
func TestDiffFor_RealRepo_ModifiedFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitshow-diff-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	mustRun(t, dir, "git", "init", "-q", "-b", "main")
	mustRun(t, dir, "git", "config", "user.email", "test@local")
	mustRun(t, dir, "git", "config", "user.name", "Test User")
	mustRun(t, dir, "git", "config", "commit.gpgsign", "false")

	if err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n\nfunc Hello() string { return \"v1\" }\n"), 0o644); err != nil {
		t.Fatalf("write v1: %v", err)
	}
	mustRun(t, dir, "git", "add", "hello.go")
	mustRun(t, dir, "git", "commit", "-q", "-m", "v1")

	if err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n\nfunc Hello() string { return \"v2\" }\n"), 0o644); err != nil {
		t.Fatalf("write v2: %v", err)
	}
	mustRun(t, dir, "git", "add", "hello.go")
	mustRun(t, dir, "git", "commit", "-q", "-m", "v2")

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	commits, err := repo.RecentCommits("", 2)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("want 2 commits, got %d", len(commits))
	}

	files, err := repo.DiffFor(commits[0])
	if err != nil {
		t.Fatalf("DiffFor: %v", err)
	}
	if len(files) != 1 || files[0].Path != "hello.go" {
		t.Fatalf("DiffFor v2: %#v", files)
	}
	if files[0].Mode != Modified {
		t.Errorf("Mode = %v, want Modified", files[0].Mode)
	}
	if len(files[0].Hunks) == 0 {
		t.Fatal("expected at least one hunk")
	}

	gotAdded := 0
	gotDeleted := 0
	for _, h := range files[0].Hunks {
		for _, l := range h.Lines {
			if l.Kind == LineAdded && strings.Contains(l.Text, "v2") {
				gotAdded++
			}
			if l.Kind == LineDeleted && strings.Contains(l.Text, "v1") {
				gotDeleted++
			}
		}
	}
	if gotAdded == 0 || gotDeleted == 0 {
		t.Errorf("expected +v2 and -v1 lines, got %d added / %d deleted", gotAdded, gotDeleted)
	}
}

func TestDiffFor_RootCommit_TreatsEverythingAsAdded(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitshow-root-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	mustRun(t, dir, "git", "init", "-q", "-b", "main")
	mustRun(t, dir, "git", "config", "user.email", "test@local")
	mustRun(t, dir, "git", "config", "user.name", "Test User")
	mustRun(t, dir, "git", "config", "commit.gpgsign", "false")

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("line1\nline2\nline3\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	mustRun(t, dir, "git", "add", "a.txt")
	mustRun(t, dir, "git", "commit", "-q", "-m", "root")

	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	commits, err := repo.RecentCommits("", 1)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	files, err := repo.DiffFor(commits[0])
	if err != nil {
		t.Fatalf("DiffFor: %v", err)
	}
	if len(files) != 1 || files[0].Mode != Added || files[0].Path != "a.txt" {
		t.Fatalf("root diff: %#v", files)
	}
	added := 0
	for _, l := range files[0].Hunks[0].Lines {
		if l.Kind == LineAdded {
			added++
		}
	}
	if added != 3 {
		t.Errorf("added lines = %d, want 3", added)
	}
}
