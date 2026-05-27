// Package git is a thin wrapper around go-git that exposes only the
// data shapes gitshow cares about.  Other packages should depend on
// this, never on go-git directly, so we can swap implementations
// (libgit2, shelled-out git) later if profiling demands it.
package git

import (
	"fmt"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Commit is the gitshow-internal commit shape.  Mirrors blueprint §14.1.
type Commit struct {
	Hash      string
	Author    string
	Email     string
	Message   string
	Timestamp time.Time
}

// ChangeMode tags how a file moved between commit and its parent.
type ChangeMode string

const (
	Added    ChangeMode = "added"
	Modified ChangeMode = "modified"
	Deleted  ChangeMode = "deleted"
	Renamed  ChangeMode = "renamed"
)

// ChangedFile is one file that moved between two commits.  Hunks may
// be empty for binary or pure-rename changes.
type ChangedFile struct {
	Path    string     // new path (or old path when deleted)
	OldPath string     // previous path when renamed; empty otherwise
	Mode    ChangeMode
	Binary  bool
	Hunks   []Hunk
}

// LineKind tags a single line of a diff hunk.
type LineKind int

const (
	LineContext LineKind = iota
	LineAdded
	LineDeleted
)

// HunkLine is one line within a hunk.  Text never includes the
// leading +/-/space marker.
type HunkLine struct {
	Kind LineKind
	Text string
}

// Hunk is one @@ block from a unified diff.
type Hunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []HunkLine
}

// Subject returns the first line of the commit message.
func (c Commit) Subject() string {
	if c.Message == "" {
		return ""
	}
	if i := strings.IndexByte(c.Message, '\n'); i >= 0 {
		return c.Message[:i]
	}
	return c.Message
}

// Repo is gitshow's view of a git repository.
type Repo struct {
	r *gogit.Repository
}

// Open returns a Repo rooted at path.  Returns an error if path is
// not inside a git work tree.
func Open(path string) (*Repo, error) {
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}
	return &Repo{r: r}, nil
}

// RecentCommits returns up to `limit` commits ending at the named
// branch (or HEAD when branch is empty).  Newest first.  When limit
// is <= 0, returns an empty slice.
func (r *Repo) RecentCommits(branch string, limit int) ([]Commit, error) {
	if limit <= 0 {
		return []Commit{}, nil
	}

	start, err := r.startHash(branch)
	if err != nil {
		return nil, err
	}

	iter, err := r.r.Log(&gogit.LogOptions{From: start})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	defer iter.Close()

	out := make([]Commit, 0, limit)
	err = iter.ForEach(func(c *object.Commit) error {
		out = append(out, fromGogitCommit(c))
		if len(out) >= limit {
			return errStop
		}
		return nil
	})
	if err != nil && err != errStop {
		return nil, err
	}
	return out, nil
}

func (r *Repo) startHash(branch string) (plumbing.Hash, error) {
	if branch == "" {
		head, err := r.r.Head()
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("HEAD: %w", err)
		}
		return head.Hash(), nil
	}

	ref, err := r.r.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err == nil {
		return ref.Hash(), nil
	}
	// Fall back to remote-tracking refs (e.g. origin/main).
	ref, err = r.r.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
	if err == nil {
		return ref.Hash(), nil
	}
	return plumbing.ZeroHash, fmt.Errorf("branch %q not found", branch)
}

func fromGogitCommit(c *object.Commit) Commit {
	return Commit{
		Hash:      c.Hash.String(),
		Author:    c.Author.Name,
		Email:     c.Author.Email,
		Message:   c.Message,
		Timestamp: c.Author.When,
	}
}

// errStop is a sentinel used to bail out of go-git iterators early.
var errStop = fmt.Errorf("stop")
