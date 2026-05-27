package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

type replayOpts struct {
	commits int
	branch  string
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

	// Week 1 deliverable: dump raw commit messages.  Week 3 swaps
	// this for a Bubble Tea cinematic player.
	for _, c := range commits {
		fmt.Printf("%s  %s  %s\n", c.Hash[:7], c.Author, c.Subject())
	}
	return nil
}
