// Package commands wires the gitshow CLI surface via Cobra.
package commands

import (
	"github.com/spf13/cobra"
)

// NewRootCmd returns the gitshow root command tree.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "gitshow",
		Short:         "Apple Keynote for Pull Requests — cinematic Git storytelling in your terminal.",
		Long:          longRootDescription,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newReplayCmd())
	return root
}

const longRootDescription = `gitshow turns Git history into a cinematic terminal walkthrough.

Phase 1 ships 'gitshow replay' — an animated, syntax-highlighted replay
of recent commits.  Future phases add PR walkthroughs, AI summaries,
HTML/GIF exports, and a hosted sharing layer.

Examples:
  gitshow replay
  gitshow replay --commits 50
  gitshow replay --branch feature/auth
`
