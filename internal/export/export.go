// Package export turns a slice of CommitWithDiff into a serialized
// artifact: Markdown for human consumption, JSON for machines.
//
// Both formats are deterministic given the same input: no timestamps
// of "now", no map iteration, nothing that would jitter snapshot
// tests.  GeneratedAt is taken from the explicit options parameter.
package export

import (
	"time"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

// SchemaVersion is the JSON export schema version.  Bump when the
// JSON shape changes in a breaking way.
const SchemaVersion = 1

// CommitWithDiff mirrors ui.CommitWithDiff so callers from either
// surface can reuse a single shape.  Keeping the type local to the
// export package avoids a UI ↔ export circular dependency.
type CommitWithDiff struct {
	Commit gitpkg.Commit
	Files  []gitpkg.ChangedFile
}

// Options control formatting decisions shared between Markdown and
// JSON exporters.  The zero value is a sensible default.
type Options struct {
	// RepoName is the project label shown in the generated header.
	RepoName string
	// GeneratedAt is embedded in the output.  Zero means "omit" so
	// snapshot tests can be byte-exact.
	GeneratedAt time.Time
	// ExcludeBody hides the commit body text (multi-paragraph
	// description) from the export, keeping only the subject line.
	// Useful when you want a compact summary.
	ExcludeBody bool
	// MaxHunkLines truncates large hunks before serialization.  Zero
	// means no truncation.  Truncation adds a synthetic placeholder
	// line so the output never silently loses information.
	MaxHunkLines int
}
