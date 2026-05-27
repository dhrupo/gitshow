// Package render turns parsed gitshow data structures into ANSI
// terminal output.  The diff renderer uses Chroma to syntax-highlight
// each added/deleted line by detecting the file's language from its
// path, then wraps the result with green/red background bars that look
// like a typical PR-review diff.
package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	gitpkg "github.com/dhrupo/gitshow/internal/git"
)

const (
	ansiReset       = "\x1b[0m"
	ansiBoldRed     = "\x1b[1;31m"
	ansiBoldGreen   = "\x1b[1;32m"
	ansiBoldCyan    = "\x1b[1;36m"
	ansiDimGray     = "\x1b[2;37m"
	ansiPlusFg      = "\x1b[32m"
	ansiMinusFg     = "\x1b[31m"
	ansiHunkHeader  = "\x1b[36m"
	ansiFilePathFg  = "\x1b[1;37m"
	ansiModifiedTag = "\x1b[33m"
	ansiAddedTag    = "\x1b[32m"
	ansiDeletedTag  = "\x1b[31m"
	ansiRenamedTag  = "\x1b[35m"
)

// Options control how diffs are rendered.
type Options struct {
	// ChromaStyle is a Chroma style name (e.g. "monokai", "dracula").
	// Empty falls back to "monokai".
	ChromaStyle string
	// NoColor disables every ANSI escape, useful for snapshot tests
	// or non-tty output.
	NoColor bool
	// MaxHunkLines caps how many lines per hunk we render before
	// emitting a "... N more lines truncated" marker.  Zero means no
	// cap.
	MaxHunkLines int
}

// Diff renders a single ChangedFile to a string.  The output is safe
// to write directly to an ANSI-capable terminal.
func Diff(file gitpkg.ChangedFile, opts Options) string {
	style := styleFor(opts.ChromaStyle)

	var buf bytes.Buffer
	writeFileHeader(&buf, file, opts)

	if file.Binary {
		writeBinaryNote(&buf, opts)
		return buf.String()
	}

	lexer := lexerForPath(file.Path)
	for _, hunk := range file.Hunks {
		writeHunkHeader(&buf, hunk, opts)
		writeHunkLines(&buf, hunk, lexer, style, opts)
	}
	return buf.String()
}

// DiffSet renders a list of files into one string with blank lines
// between them.
func DiffSet(files []gitpkg.ChangedFile, opts Options) string {
	if len(files) == 0 {
		return ""
	}
	parts := make([]string, 0, len(files))
	for _, f := range files {
		parts = append(parts, Diff(f, opts))
	}
	return strings.Join(parts, "\n")
}

func styleFor(name string) *chroma.Style {
	if name == "" {
		name = "monokai"
	}
	s := styles.Get(name)
	if s == nil {
		s = styles.Fallback
	}
	return s
}

func lexerForPath(path string) chroma.Lexer {
	l := lexers.Match(path)
	if l == nil {
		l = lexers.Fallback
	}
	return chroma.Coalesce(l)
}

func writeFileHeader(buf *bytes.Buffer, file gitpkg.ChangedFile, opts Options) {
	tag := modeTag(file.Mode)
	path := file.Path
	if file.Mode == gitpkg.Renamed && file.OldPath != "" {
		path = fmt.Sprintf("%s → %s", file.OldPath, file.Path)
	}
	if opts.NoColor {
		fmt.Fprintf(buf, "%s %s\n", strings.ToUpper(string(file.Mode)), path)
		return
	}
	fmt.Fprintf(buf, "%s%s%s %s%s%s\n",
		tag, strings.ToUpper(string(file.Mode)), ansiReset,
		ansiFilePathFg, path, ansiReset)
}

func modeTag(mode gitpkg.ChangeMode) string {
	switch mode {
	case gitpkg.Added:
		return ansiAddedTag
	case gitpkg.Deleted:
		return ansiDeletedTag
	case gitpkg.Renamed:
		return ansiRenamedTag
	default:
		return ansiModifiedTag
	}
}

func writeBinaryNote(buf *bytes.Buffer, opts Options) {
	if opts.NoColor {
		buf.WriteString("  (binary file - no preview)\n")
		return
	}
	fmt.Fprintf(buf, "%s  (binary file - no preview)%s\n", ansiDimGray, ansiReset)
}

func writeHunkHeader(buf *bytes.Buffer, h gitpkg.Hunk, opts Options) {
	header := fmt.Sprintf("@@ -%d,%d +%d,%d @@", h.OldStart, h.OldLines, h.NewStart, h.NewLines)
	if opts.NoColor {
		fmt.Fprintf(buf, "  %s\n", header)
		return
	}
	fmt.Fprintf(buf, "  %s%s%s\n", ansiHunkHeader, header, ansiReset)
}

func writeHunkLines(buf *bytes.Buffer, h gitpkg.Hunk, lexer chroma.Lexer, style *chroma.Style, opts Options) {
	limit := len(h.Lines)
	truncated := 0
	if opts.MaxHunkLines > 0 && limit > opts.MaxHunkLines {
		truncated = limit - opts.MaxHunkLines
		limit = opts.MaxHunkLines
	}
	for i := 0; i < limit; i++ {
		writeHunkLine(buf, h.Lines[i], lexer, style, opts)
	}
	if truncated > 0 {
		if opts.NoColor {
			fmt.Fprintf(buf, "  ... %d more line(s) truncated\n", truncated)
		} else {
			fmt.Fprintf(buf, "  %s... %d more line(s) truncated%s\n", ansiDimGray, truncated, ansiReset)
		}
	}
}

func writeHunkLine(buf *bytes.Buffer, line gitpkg.HunkLine, lexer chroma.Lexer, style *chroma.Style, opts Options) {
	marker, fg := lineMarker(line.Kind)
	body := line.Text
	if !opts.NoColor && lexer != nil {
		highlighted, err := highlight(body, lexer, style)
		if err == nil && highlighted != "" {
			body = highlighted
		}
	}
	if opts.NoColor {
		fmt.Fprintf(buf, "%s %s\n", marker, body)
		return
	}
	fmt.Fprintf(buf, "%s%s%s %s%s\n", fg, marker, ansiReset, body, ansiReset)
}

func lineMarker(kind gitpkg.LineKind) (marker, fg string) {
	switch kind {
	case gitpkg.LineAdded:
		return "+", ansiPlusFg
	case gitpkg.LineDeleted:
		return "-", ansiMinusFg
	default:
		return " ", ansiDimGray
	}
}

// highlight runs Chroma against a single line.  Chroma was designed
// for whole-file lexing; running it per line gives "good enough" colors
// for keywords, strings, and numbers but loses some context.  Cheap to
// improve later if needed.
func highlight(line string, lexer chroma.Lexer, style *chroma.Style) (string, error) {
	if line == "" {
		return "", nil
	}
	it, err := lexer.Tokenise(nil, line)
	if err != nil {
		return "", err
	}
	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, it); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}
