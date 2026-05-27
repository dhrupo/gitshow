package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// initFixtureRepo creates a tiny git repo on disk with `n` commits.
// Returns the repo path; caller is responsible for cleanup.
func initFixtureRepo(t *testing.T, n int) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gitshow-fixture-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	mustRun(t, dir, "git", "init", "-q", "-b", "main")
	mustRun(t, dir, "git", "config", "user.email", "test@local")
	mustRun(t, dir, "git", "config", "user.name", "Test User")
	mustRun(t, dir, "git", "config", "commit.gpgsign", "false")

	for i := 1; i <= n; i++ {
		fname := filepath.Join(dir, "file.txt")
		body := []byte("revision " + itoa(i) + "\n")
		if err := os.WriteFile(fname, body, 0o644); err != nil {
			t.Fatalf("write fixture file: %v", err)
		}
		mustRun(t, dir, "git", "add", "file.txt")
		mustRun(t, dir, "git", "commit", "-q", "-m", "commit "+itoa(i)+"\n\nlonger body line.")
	}
	return dir
}

func itoa(n int) string {
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

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestCommit_Subject(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want string
	}{
		{"empty", "", ""},
		{"single line", "fix the thing", "fix the thing"},
		{"multi line", "fix the thing\n\nlonger explanation", "fix the thing"},
		{"trailing newline only", "fix the thing\n", "fix the thing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Commit{Message: tc.msg}
			if got := c.Subject(); got != tc.want {
				t.Errorf("Subject() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestOpen_FailsOnNonRepo(t *testing.T) {
	dir, err := os.MkdirTemp("", "gitshow-nonrepo-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	if _, err := Open(dir); err == nil {
		t.Fatalf("Open(%q) succeeded; expected error on non-repo", dir)
	}
}

func TestRecentCommits_ReturnsNewestFirstUpToLimit(t *testing.T) {
	repoPath := initFixtureRepo(t, 5)
	r, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got, err := r.RecentCommits("", 3)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d commits, want 3", len(got))
	}
	// Newest-first: subjects should be commit 5, 4, 3.
	wantSubjects := []string{"commit 5", "commit 4", "commit 3"}
	for i, c := range got {
		if c.Subject() != wantSubjects[i] {
			t.Errorf("commit[%d].Subject() = %q, want %q", i, c.Subject(), wantSubjects[i])
		}
		if c.Author != "Test User" {
			t.Errorf("commit[%d].Author = %q, want Test User", i, c.Author)
		}
		if c.Hash == "" {
			t.Errorf("commit[%d].Hash is empty", i)
		}
	}
}

func TestRecentCommits_LimitGreaterThanHistoryReturnsAll(t *testing.T) {
	repoPath := initFixtureRepo(t, 2)
	r, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	got, err := r.RecentCommits("", 50)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d commits, want 2", len(got))
	}
}

func TestRecentCommits_ZeroLimitReturnsEmpty(t *testing.T) {
	repoPath := initFixtureRepo(t, 3)
	r, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	got, err := r.RecentCommits("", 0)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d commits, want 0", len(got))
	}
}

func TestRecentCommits_UnknownBranch(t *testing.T) {
	repoPath := initFixtureRepo(t, 1)
	r, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := r.RecentCommits("does-not-exist", 5); err == nil {
		t.Fatal("expected error for unknown branch, got nil")
	}
}

func TestRecentCommits_TimestampPreserved(t *testing.T) {
	repoPath := initFixtureRepo(t, 1)
	r, err := Open(repoPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	got, err := r.RecentCommits("", 1)
	if err != nil {
		t.Fatalf("RecentCommits: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected 1 commit")
	}
	if got[0].Timestamp.IsZero() {
		t.Error("Timestamp is zero; expected commit time")
	}
	if time.Since(got[0].Timestamp) > 5*time.Minute {
		t.Errorf("Timestamp %v looks stale", got[0].Timestamp)
	}
}
