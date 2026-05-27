package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
	"github.com/dhrupo/gitshow/internal/render"
	"github.com/dhrupo/gitshow/internal/ui"
)

type replayOpts struct {
	commits      int
	branch       string
	noDiff       bool
	noColor      bool
	maxHunkLines int
	chromaStyle  string
	tui          string // "auto" (default), "on", or "off"
}

func newReplayCmd() *cobra.Command {
	opts := &replayOpts{}

	cmd := &cobra.Command{
		Use:   "replay [branch]",
		Short: "Cinematic replay of recent commits in the current repository.",
		Long: `Replay walks the most recent N commits of the current repository
(or the named branch) and renders them as a cinematic terminal
walkthrough.  In Phase 1 the output is a static dump; the animated
Bubble Tea TUI lands in week 3.

Examples:
  gitshow replay
  gitshow replay --commits 5
  gitshow replay feature/auth
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.branch = args[0]
			}
			return runReplay(opts)
		},
	}

	cmd.Flags().IntVarP(&opts.commits, "commits", "n", 20, "number of recent commits to replay (default 20)")
	cmd.Flags().StringVarP(&opts.branch, "branch", "b", "", "branch to replay (default: current HEAD)")
	cmd.Flags().BoolVar(&opts.noDiff, "no-diff", false, "skip per-file diffs; show only commit headers")
	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "disable ANSI colors (useful for piping to a file)")
	cmd.Flags().IntVar(&opts.maxHunkLines, "max-hunk-lines", 80, "max lines per hunk before truncation; 0 = unlimited")
	cmd.Flags().StringVar(&opts.chromaStyle, "chroma-style", "monokai", "Chroma syntax theme (monokai / dracula / nord / ...)")
	cmd.Flags().StringVar(&opts.tui, "tui", "auto", "interactive TUI mode: auto / on / off")

	return cmd
}

func runReplay(opts *replayOpts) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine current directory: %w", err)
	}

	repo, err := gitpkg.Open(cwd)
	if err != nil {
		return fmt.Errorf("opening repository at %s: %w", cwd, err)
	}

	commits, err := repo.RecentCommits(opts.branch, opts.commits)
	if err != nil {
		return fmt.Errorf("reading commits: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("(no commits found)")
		return nil
	}

	if shouldUseTUI(opts) {
		return runReplayTUI(repo, commits, cwd)
	}
	return runReplayDump(repo, commits, opts)
}

func shouldUseTUI(opts *replayOpts) bool {
	switch strings.ToLower(opts.tui) {
	case "on", "true", "yes":
		return true
	case "off", "false", "no":
		return false
	default:
		return ui.IsStdoutTTY()
	}
}

func runReplayTUI(repo *gitpkg.Repo, commits []gitpkg.Commit, cwd string) error {
	loaded := make([]ui.CommitWithDiff, 0, len(commits))
	for _, c := range commits {
		files, err := repo.DiffFor(c)
		if err != nil {
			// Don't fail the whole replay because one commit's diff
			// can't be loaded; record an empty diff and keep going.
			fmt.Fprintf(os.Stderr, "warning: could not load diff for %s: %v\n", c.Hash[:7], err)
			files = nil
		}
		loaded = append(loaded, ui.CommitWithDiff{Commit: c, Files: files})
	}
	repoName := filepath.Base(cwd)
	return ui.Run(ui.New(repoName, loaded))
}

func runReplayDump(repo *gitpkg.Repo, commits []gitpkg.Commit, opts *replayOpts) error {
	renderOpts := render.Options{
		ChromaStyle:  opts.chromaStyle,
		NoColor:      opts.noColor,
		MaxHunkLines: opts.maxHunkLines,
	}

	for i, c := range commits {
		if i > 0 {
			fmt.Println()
		}
		printCommitHeader(c, opts.noColor)
		if opts.noDiff {
			continue
		}
		files, err := repo.DiffFor(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  (could not load diff: %v)\n", err)
			continue
		}
		if len(files) == 0 {
			continue
		}
		fmt.Print(render.DiffSet(files, renderOpts))
	}
	return nil
}

func printCommitHeader(c gitpkg.Commit, noColor bool) {
	short := c.Hash
	if len(short) > 7 {
		short = short[:7]
	}
	subject := strings.TrimSpace(c.Subject())
	when := c.Timestamp.Format("2006-01-02 15:04")
	if noColor {
		fmt.Printf("commit %s  %s  %s\n  %s\n", short, c.Author, when, subject)
		return
	}
	fmt.Printf("\x1b[1;33mcommit %s\x1b[0m  \x1b[36m%s\x1b[0m  \x1b[2m%s\x1b[0m\n  \x1b[1m%s\x1b[0m\n",
		short, c.Author, when, subject)
}
