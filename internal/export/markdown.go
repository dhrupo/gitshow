package export

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

// Markdown writes the human-readable Markdown export to w.  Format:
//
//	# Replay of <RepoName>            (header line, optional)
//	_<N> commits, generated <timestamp>_
//
//	## <subject>
//	_<author> · <hash> · <timestamp>_
//
//	<body, when present and not excluded>
//
//	### <Mode> <path>
//
//	```<lang>
//	  context line
//	- removed line
//	+ added line
//	```
//
// Output is deterministic: no map iteration, no "now"-flavoured
// fields, no rune-jitter.
func Markdown(w io.Writer, commits []CommitWithDiff, opts Options) error {
	if len(commits) == 0 {
		_, err := io.WriteString(w, emptyMarkdown(opts))
		return err
	}

	var b strings.Builder
	writeMarkdownHeader(&b, len(commits), opts)
	for i, c := range commits {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		writeMarkdownCommit(&b, c, opts)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func emptyMarkdown(opts Options) string {
	var b strings.Builder
	if opts.RepoName != "" {
		fmt.Fprintf(&b, "# Replay of %s\n\n", opts.RepoName)
	}
	b.WriteString("_(no commits to replay)_\n")
	return b.String()
}

func writeMarkdownHeader(b *strings.Builder, n int, opts Options) {
	if opts.RepoName != "" {
		fmt.Fprintf(b, "# Replay of %s\n", opts.RepoName)
	} else {
		b.WriteString("# gitshow replay\n")
	}
	switch {
	case n == 1 && !opts.GeneratedAt.IsZero():
		fmt.Fprintf(b, "_1 commit, generated %s_\n\n", opts.GeneratedAt.UTC().Format("2006-01-02 15:04:05 MST"))
	case n == 1:
		b.WriteString("_1 commit_\n\n")
	case !opts.GeneratedAt.IsZero():
		fmt.Fprintf(b, "_%d commits, generated %s_\n\n", n, opts.GeneratedAt.UTC().Format("2006-01-02 15:04:05 MST"))
	default:
		fmt.Fprintf(b, "_%d commits_\n\n", n)
	}
}

func writeMarkdownCommit(b *strings.Builder, c CommitWithDiff, opts Options) {
	subject := strings.TrimSpace(c.Commit.Subject())
	if subject == "" {
		subject = "(no subject)"
	}
	fmt.Fprintf(b, "## %s\n", subject)

	short := c.Commit.Hash
	if len(short) > 7 {
		short = short[:7]
	}
	fmt.Fprintf(b, "_%s · `%s` · %s_\n\n",
		mdEscape(c.Commit.Author),
		short,
		c.Commit.Timestamp.UTC().Format("2006-01-02 15:04 MST"))

	if !opts.ExcludeBody {
		body := strings.TrimSpace(strings.TrimPrefix(c.Commit.Message, c.Commit.Subject()))
		if body != "" {
			b.WriteString(body)
			b.WriteString("\n\n")
		}
	}

	if len(c.Files) == 0 {
		b.WriteString("_(no file changes)_\n")
		return
	}

	for _, f := range c.Files {
		writeMarkdownFile(b, f, opts)
	}
}

func writeMarkdownFile(b *strings.Builder, f gitpkg.ChangedFile, opts Options) {
	switch f.Mode {
	case gitpkg.Renamed:
		fmt.Fprintf(b, "### RENAMED `%s` → `%s`\n\n", f.OldPath, f.Path)
	default:
		fmt.Fprintf(b, "### %s `%s`\n\n",
			strings.ToUpper(string(f.Mode)), f.Path)
	}

	if f.Binary {
		b.WriteString("_(binary file - no preview)_\n\n")
		return
	}

	lang := markdownLangHint(f.Path)
	if lang == "" {
		b.WriteString("```diff\n")
	} else {
		fmt.Fprintf(b, "```%s\n", lang)
	}
	for _, h := range f.Hunks {
		fmt.Fprintf(b, "@@ -%d,%d +%d,%d @@\n", h.OldStart, h.OldLines, h.NewStart, h.NewLines)
		limit := len(h.Lines)
		truncated := 0
		if opts.MaxHunkLines > 0 && limit > opts.MaxHunkLines {
			truncated = limit - opts.MaxHunkLines
			limit = opts.MaxHunkLines
		}
		for i := 0; i < limit; i++ {
			line := h.Lines[i]
			switch line.Kind {
			case gitpkg.LineAdded:
				fmt.Fprintf(b, "+ %s\n", line.Text)
			case gitpkg.LineDeleted:
				fmt.Fprintf(b, "- %s\n", line.Text)
			default:
				fmt.Fprintf(b, "  %s\n", line.Text)
			}
		}
		if truncated > 0 {
			fmt.Fprintf(b, "... %d more line(s) truncated\n", truncated)
		}
	}
	b.WriteString("```\n\n")
}

// markdownLangHint maps a file extension to the language identifier
// Markdown code fences expect.  Always lowercase, never panics.
func markdownLangHint(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs":
		return "js"
	case ".ts", ".tsx":
		return "ts"
	case ".vue":
		return "vue"
	case ".py":
		return "python"
	case ".php":
		return "php"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".java":
		return "java"
	case ".kt", ".kts":
		return "kotlin"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".swift":
		return "swift"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	case ".md", ".markdown":
		return "markdown"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	default:
		return ""
	}
}

// mdEscape escapes the minimal set of Markdown characters that can
// mangle author names.  We only worry about underscores and asterisks
// because those are the realistic chars an author might pick.
func mdEscape(s string) string {
	r := strings.NewReplacer("_", "\\_", "*", "\\*")
	return r.Replace(s)
}
