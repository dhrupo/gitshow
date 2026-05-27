# gitshow

[![CI](https://github.com/dhrupo/gitshow/actions/workflows/ci.yml/badge.svg)](https://github.com/dhrupo/gitshow/actions/workflows/ci.yml)
[![go report](https://goreportcard.com/badge/github.com/dhrupo/gitshow)](https://goreportcard.com/report/github.com/dhrupo/gitshow)
[![go version](https://img.shields.io/badge/go-%E2%89%A51.22-00ADD8.svg)](go.mod)
[![license](https://img.shields.io/github/license/dhrupo/gitshow.svg)](LICENSE)

> **Apple Keynote for Pull Requests.**
> A cinematic terminal storytelling engine for Git history.

![demo](https://raw.githubusercontent.com/dhrupo/gitshow/main/docs/demo.gif)

`gitshow` turns the last *N* commits of your repo into an animated
walkthrough you can play live in your terminal or hand off as a
Markdown / JSON artifact for code-review, release notes, or onboarding.
`git log --oneline` tells you *what* changed; `gitshow` tells you
*why, how, impact, risk, and evolution.*

```bash
gitshow replay                  # cinematic replay in your terminal
gitshow replay --export md > pr-story.md
gitshow replay --export json > pr-story.json
```

## Why this exists

| Tool | Weakness gitshow exploits |
|---|---|
| `git log` | Ugly, no narrative |
| Lazygit | Navigation-focused, not presentation |
| GitHub PRs | Static, web-only |
| VHS | Recording only, no Git intelligence |
| asciinema | Low-level playback, not story |

`gitshow` sits in the middle: a short, animated narrative your team
can actually pay attention to.

## Install

### From source

```bash
git clone https://github.com/dhrupo/gitshow
cd gitshow
go build -o gitshow ./cmd/gitshow
./gitshow --help
```

Requires Go ≥ 1.22.

### Coming soon

```bash
brew install gitshow       # not yet published
```

## Quick start

From any Git repository:

```bash
gitshow replay              # cinematic TUI replay (auto-detects TTY)
gitshow replay --commits 5  # only the last 5 commits
gitshow replay feature/auth # specific branch
```

In the live TUI:

| Key | Action |
|---|---|
| `space` | pause / resume |
| `←` / `h` | previous commit |
| `→` / `l` | next commit |
| `↑` / `k` | faster (1.0× → 2.0×) |
| `↓` / `j` | slower (1.0× → 0.5×) |
| `home` / `g` | jump to first commit |
| `end` / `G` | jump to last commit |
| `q` / `esc` / `ctrl+c` | quit |

When stdout is not a terminal (`gitshow replay > out.txt`), it falls
back to a deterministic stdout dump suitable for piping. Force the
behaviour with `--tui on` or `--tui off`.

## Commands

### `gitshow replay [branch]`

The MVP command. Walks recent commits with animated reveals,
syntax-highlighted diffs, and a timeline strip.

```text
Common flags:
  -n, --commits <N>          number of commits to replay (default 20)
  -b, --branch <ref>         branch to replay (default: HEAD)
      --no-diff              show commit headers only
      --no-color             disable ANSI colors (for piping)
      --max-hunk-lines <N>   truncate large hunks (default 80; 0 = unlimited)
      --chroma-style <name>  Chroma syntax theme (monokai / dracula / nord ...)
      --tui <auto|on|off>    interactive TUI mode

Export flags:
      --export <format>      markdown / json
  -o, --output <path>        write export to a file (default: stdout)
      --exclude-body         in exports, keep only subject lines

Help:
  -h, --help                 show help
  -v, --version              show version
```

### Exports

```bash
gitshow replay --export markdown -o pr-story.md
gitshow replay --export json     -o pr-story.json
```

The Markdown export reads like a narrated PR (subject, meta, body,
fenced syntax-highlighted diffs with language hints). The JSON export
has `schema_version: 1` and a stable shape — wire it into dashboards,
GitHub issue creation, or team backlogs.

## Output anatomy

The live TUI is a 4-panel layout:

```
┌──────────────────────────────────────────────────┐
│ gitshow   <repo>          [3 / 8]  ▶ playing  ●●●○○│  Header
├──────────────────────────────────────────────────┤
│ ╭──────────────────────────────────────────────╮ │
│ │ feat: add OAuth login                        │ │  Main playback
│ │ commit abc1234 · Alice · 2026-05-27 12:00    │ │
│ │                                              │ │
│ │ Adds Google + GitHub OAuth providers, ...    │ │
│ ╰──────────────────────────────────────────────╯ │
├──────────────────────────────────────────────────┤
│  MODIFIED app/auth.go                            │  Diff view
│  @@ -10,3 +10,7 @@                                │
│  + provider := oauth.NewGoogle(cfg.GoogleClient) │
│  ...                                              │
├──────────────────────────────────────────────────┤
│ • • • ● · · · ·                                   │  Timeline
│   ←/→ navigate · ↑/↓ speed · space pause · q quit │
└──────────────────────────────────────────────────┘
```

- **Header** — repo, position, play/pause indicator, speed bar.
- **Main playback** — typewriter-revealed commit subject and metadata.
- **Diff view** — Chroma-syntax-highlighted unified diff, line-by-line
  stagger reveal.
- **Timeline** — past (`•`) / current (`●`) / future (`·`) dots,
  proportional to commit count.

## How it's built

| Module | Responsibility |
|---|---|
| `cmd/gitshow` | Cobra entrypoint |
| `internal/git` | go-git wrapper: commits, diffs, branches |
| `internal/render` | Chroma syntax highlighting → ANSI |
| `internal/animation` | Typewriter / Stagger / EaseInOutCubic pure helpers |
| `internal/timeline` | Cinematic clock: pause / resume / speed / seek |
| `internal/ui` | Bubble Tea Model + View + 4-panel layout |
| `internal/export` | Markdown + JSON writers |

The animation system is pure: every effect is a function of `elapsed`
and `total` durations. The timeline package provides the clock; the
Bubble Tea Model reads from it on each frame tick (~30 fps). This
keeps `Update()` trivial and the unit tests blazing fast (75 tests run
in well under a second).

## Roadmap

This is Phase 1 of a 6-phase plan ([blueprint][plan]).

- [x] **Phase 1** — Core replay engine (this repo today)
- [ ] **Phase 2** — `gitshow pr <num>` PR walkthrough mode
- [ ] **Phase 3** — Optional AI summaries (`--ai` flag)
- [ ] **Phase 4** — HTML / GIF / SVG / asciinema exports
- [ ] **Phase 5** — GitHub integration
- [ ] **Phase 6** — Hosted sharing platform

[plan]: docs/

## Development

```bash
git clone https://github.com/dhrupo/gitshow
cd gitshow
go build -o gitshow ./cmd/gitshow
go test ./...              # 75 unit + integration tests, ~5s total
go vet ./...
```

CI runs on Ubuntu + macOS across Go 1.22 / 1.23 / 1.24.

### Re-record the demo GIF

The demo at the top of this README is driven by
[vhs](https://github.com/charmbracelet/vhs):

```bash
go install ./cmd/gitshow   # or symlink the built binary onto your PATH
vhs docs/demo.tape         # writes docs/demo.gif
```

### Project structure

```
cmd/gitshow/                  Cobra entrypoint
internal/
  git/                        go-git wrapper + diff parser
  render/                     Chroma highlighting → ANSI
  animation/                  Pure timing helpers
  timeline/                   Cinematic clock
  ui/                         Bubble Tea Model + View
  export/                     Markdown + JSON writers
testdata/                     Synthetic git fixtures (reserved)
docs/                         Demo GIF + tape sources
```

## Contributing

Bug reports, PRs, and "this looks ugly on my terminal, here's a
screenshot" issues all welcome. The pre-1.0 surface is intentionally
small; please grill scope before opening a big PR.

## License

MIT — see [LICENSE](./LICENSE).
