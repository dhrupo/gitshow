<!-- Header / hero -->
<h1 align="center">gitshow</h1>

<p align="center">
  <strong>Apple Keynote for Pull Requests.</strong><br/>
  Turn your Git history into a cinematic terminal walkthrough.
</p>

<p align="center">
  <a href="https://github.com/dhrupo/gitshow/actions/workflows/ci.yml"><img src="https://github.com/dhrupo/gitshow/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
  <a href="https://goreportcard.com/report/github.com/dhrupo/gitshow"><img src="https://goreportcard.com/badge/github.com/dhrupo/gitshow" alt="Go Report"/></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/go-%E2%89%A51.22-00ADD8.svg" alt="Go ≥ 1.22"/></a>
  <a href="https://github.com/dhrupo/gitshow/releases/latest"><img src="https://img.shields.io/github/v/release/dhrupo/gitshow?color=blue" alt="Latest release"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/dhrupo/gitshow.svg" alt="MIT"/></a>
</p>

<p align="center">
  <img src="https://raw.githubusercontent.com/dhrupo/gitshow/main/docs/demo.gif" alt="gitshow demo" width="100%"/>
</p>

<p align="center">
  <a href="#install">Install</a> ·
  <a href="#30-second-demo">30-Second Demo</a> ·
  <a href="#use-it-for">Use Cases</a> ·
  <a href="#commands">Commands</a> ·
  <a href="#how-its-built">Architecture</a> ·
  <a href="#roadmap">Roadmap</a>
</p>

---

## Why this exists

`git log --oneline` tells you **what** changed. It doesn't tell you **why, how, the impact, the risk, or the evolution.** That's the whole part of code review that we keep losing in pull-request walls of text.

`gitshow` reads your repo's commits and turns them into a short, animated, narrated terminal walkthrough — the kind a careful human reviewer would walk a teammate through, but in 60 seconds and shareable as a gif.

It's tuned to *narrative, emotion, presentation, and storytelling* — not navigation. Lazygit nails navigation. GitHub PRs are static. VHS records, asciinema replays. `gitshow` sits in the middle: a programmable storytelling engine for Git.

---

## Install

### Homebrew (macOS · Linux)

```bash
brew install dhrupo/tap/gitshow
```

That's it. Run `gitshow replay` from inside any Git repo.

### From source

```bash
git clone https://github.com/dhrupo/gitshow
cd gitshow
go build -o gitshow ./cmd/gitshow
./gitshow --help
```

Requires Go ≥ 1.22. Builds a single static binary.

---

## 30-Second Demo

```bash
# 1. Install
brew install dhrupo/tap/gitshow

# 2. Walk into any repo
cd ~/code/your-favourite-project

# 3. Play the cinematic
gitshow replay
```

Press `space` to pause, `←/→` to jump between commits, `↑/↓` to change speed, `q` to quit.

Want to share it? Export as Markdown for a PR description, or as JSON for your dashboards:

```bash
gitshow replay --export markdown -o pr-story.md
gitshow replay --export json     -o pr-story.json
```

---

## Use It For

| Scenario | What `gitshow` does |
|---|---|
| 🎬 **Stand-ups & demos** | Stop scrolling through `git log`. Play the cinematic in your terminal as a moving picture of what changed this sprint. |
| 📝 **PR descriptions** | `gitshow replay --commits 12 --export markdown` → paste straight into the GitHub PR body. Auto-formatted, syntax-highlighted, narrated. |
| 🚢 **Release notes** | Walk the diff between two tags. Each commit reads as a beat in the release story instead of a bullet point. |
| 🎓 **Onboarding** | Hand a new hire `gitshow replay --commits 200` on a critical module. They watch how it evolved instead of reading dead code top-to-bottom. |
| 🔍 **Code archaeology** | Pause on any commit, study the diff, then resume. The TUI keeps your context the way `git log -p` never does. |
| 📺 **Conference talks** | Record the gif and drop it into your slides. The cinematic pacing reads on a projector, the way a `git log` screenshot never does. |

---

## Commands

### `gitshow replay [branch]`

The MVP command. Walks recent commits with animated reveals, syntax-highlighted diffs, and a timeline strip.

```bash
gitshow replay                            # cinematic TUI, last 20 commits
gitshow replay --commits 5                # narrow scope
gitshow replay feature/auth               # walk a branch
gitshow replay --base main --head HEAD    # diff a range
gitshow replay --tui off                  # plain stdout (pipe-friendly)
gitshow replay --export markdown > pr.md  # save as Markdown
gitshow replay --export json     > pr.json # save as JSON
```

#### Full flag reference

<details>
<summary>Click to expand</summary>

```text
Replay flags:
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

</details>

---

## Keyboard Cheat Sheet

| Key | Action |
|:---:|---|
| `space` | Pause / resume |
| `←` `→` *or* `h` `l` | Previous / next commit |
| `↑` `↓` *or* `k` `j` | Faster (up to 2.0×) / slower (down to 0.5×) |
| `g` / `home` | Jump to oldest commit |
| `G` / `end` | Jump to newest commit |
| `q` *or* `esc` *or* `ctrl+c` | Quit |

> When stdout is **not** a terminal (e.g. `gitshow replay | cat`), the cinematic is silently replaced with a deterministic stdout dump — perfect for CI, scripts, and pipes. Force it with `--tui on` / `--tui off`.

---

## Output Anatomy

The live TUI is a 4-panel cinematic frame:

```
┌──────────────────────────────────────────────────────────────────┐
│ gitshow   my-repo            [3 / 8]  ▶ playing  ●●●○○            │  ← Header
├──────────────────────────────────────────────────────────────────┤
│ ╭──────────────────────────────────────────────────────────────╮ │
│ │ feat: add OAuth login support                                │ │
│ │ commit abc1234 · Alice · 2026-05-27 12:00                    │ │  ← Main playback
│ │                                                              │ │     (typewriter
│ │ Adds Google + GitHub OAuth providers, session                │ │      reveal)
│ │ persistence in Redis, and a /me endpoint.                    │ │
│ ╰──────────────────────────────────────────────────────────────╯ │
├──────────────────────────────────────────────────────────────────┤
│  MODIFIED app/auth.go                                            │
│  @@ -10,3 +10,7 @@                                                │  ← Diff view
│  + provider := oauth.NewGoogle(cfg.GoogleClient)                  │     (Chroma
│  + session, err := store.New(provider, redis)                     │      highlight,
│  - // TODO: pick a provider                                       │      staggered)
├──────────────────────────────────────────────────────────────────┤
│ • • • ● · · · ·                                                   │  ← Timeline
│   ←/→ navigate · ↑/↓ speed · space pause · q quit                 │
└──────────────────────────────────────────────────────────────────┘
```

### Markdown export

```markdown
# Replay of my-repo
_8 commits, generated 2026-05-27 14:32:11 UTC_

## feat: add OAuth login support
_Alice · `abc1234` · 2026-05-27 12:00 UTC_

Adds Google + GitHub OAuth providers, session persistence in Redis,
and a /me endpoint.

### MODIFIED `app/auth.go`

​```go
@@ -10,3 +10,7 @@
+ provider := oauth.NewGoogle(cfg.GoogleClient)
+ session, err := store.New(provider, redis)
- // TODO: pick a provider
​```
```

### JSON export (schema version 1)

```json
{
  "schema_version": 1,
  "repo": "my-repo",
  "generated_at": "2026-05-27T14:32:11Z",
  "commits": [
    {
      "hash": "abc1234567890",
      "subject": "feat: add OAuth login support",
      "author": "Alice",
      "timestamp": "2026-05-27T12:00:00Z",
      "files": [
        {
          "path": "app/auth.go",
          "mode": "modified",
          "hunks": [{ "old_start": 10, "new_start": 10, "lines": [...] }]
        }
      ]
    }
  ]
}
```

Wire the JSON straight into a dashboard, generate GitHub issues, or feed it to another tool — the schema is documented and versioned.

---

## How It's Built

`gitshow` is a small, focused Go binary. No CGO, no native deps, no Python sidecar.

| Layer | Package | Responsibility |
|---|---|---|
| Entry | `cmd/gitshow` | Cobra command tree |
| Git | `internal/git` | `go-git` wrapper · commits · unified-diff parser |
| Render | `internal/render` | Chroma syntax highlighting → ANSI |
| Animation | `internal/animation` | `Typewriter` / `Stagger` / `EaseInOutCubic` — pure timing helpers |
| Timeline | `internal/timeline` | Cinematic clock with pause / resume / speed / seek |
| TUI | `internal/ui` | Bubble Tea `Model` + 4-panel `View` + keyboard handling |
| Export | `internal/export` | Deterministic Markdown + JSON writers |

The animation system is a chain of **pure functions of time**. Every effect (typewriter reveal, diff line stagger, fade) is `f(elapsed, total) → state`. The timeline package owns the clock; the Bubble Tea Model reads from it on each frame tick (~30 fps). That keeps `Update()` trivial, makes the unit tests blazing fast (75 tests in well under a second), and means recordings are reproducible from a tape.

### Render pipeline

```
git log ──▶ internal/git ──▶ Chroma (Lexer/Style) ──▶ ANSI buffer
                │
                ▼
        internal/animation (Typewriter, Stagger)
                │
                ▼
        internal/timeline (clock, pause, seek, speed)
                │
                ▼
        ┌───────┴───────┐
        ▼               ▼
   Bubble Tea TUI     Markdown / JSON exporter
   (live mode)        (headless mode)
```

---

## Compared To

| Tool | Strength | Where `gitshow` differs |
|---|---|---|
| **`git log`** | Universal | Ugly, no narrative, no animation |
| **Lazygit** | Best-in-class for navigation | `gitshow` is for *presentation*, not navigation |
| **GitHub PRs** | Async review | Static HTML, web-only, no pacing |
| **VHS** | Records terminals to gif | `gitshow` *understands* commits — VHS just records keypresses |
| **asciinema** | Web playback of recorded sessions | `gitshow` generates the playback programmatically from real history |
| **Lookatme** / **Slides** | Terminal slide decks | Generic — no Git intelligence |

The competitive edge is the *Git intelligence layer*: `gitshow` knows what a commit, a hunk, an added line means. Everything else either watches your terminal or just shows raw text.

---

## Roadmap

This is **Phase 1** of a six-phase plan.

- [x] **Phase 1 — Core replay engine** (this repo today, v1.0.0)
- [ ] **Phase 2 — PR walkthrough mode** · `gitshow pr <num>` via the `gh` CLI
- [ ] **Phase 3 — Optional AI summaries** · `--ai` flag, deterministic prompts, multiple providers
- [ ] **Phase 4 — HTML / GIF / SVG / asciinema exports** · shareable browser playback
- [ ] **Phase 5 — GitHub integration** · auto-detect PR from current branch, cross-reference issues
- [ ] **Phase 6 — Hosted sharing platform** · upload + share `.gitshow` artifacts

Each phase ships only when the prior one is genuinely *good*, not when it has the right number of features.

---

## Development

```bash
git clone https://github.com/dhrupo/gitshow
cd gitshow

go build -o gitshow ./cmd/gitshow      # build the binary
go test ./...                           # 75 unit + integration tests, ~5s
go vet ./...                            # static checks

./gitshow replay --commits 5            # try it on this repo itself
```

CI runs the test suite on **Ubuntu + macOS** across **Go 1.22 / 1.23 / 1.24**.

### Re-record the demo gif

The header gif is driven by [vhs](https://github.com/charmbracelet/vhs). To regenerate:

```bash
go install ./cmd/gitshow   # or symlink the built binary onto your PATH
vhs docs/demo.tape         # writes docs/demo.gif
```

The tape source lives at `docs/demo.tape` so the recording is reproducible.

### Releases

Cutting a release is one command:

```bash
git tag -a vX.Y.Z -m "release notes here"
git push origin main --tags
```

GoReleaser then handles the rest: cross-platform binaries, a GitHub release, checksums, and an auto-updated `Formula/gitshow.rb` in [dhrupo/homebrew-tap](https://github.com/dhrupo/homebrew-tap).

### Project structure

```
cmd/gitshow/        Cobra entrypoint
internal/
  git/              go-git wrapper + diff parser
  render/           Chroma highlighting → ANSI
  animation/        Pure timing helpers
  timeline/         Cinematic clock
  ui/               Bubble Tea Model + View
  export/           Markdown + JSON writers
docs/               Demo gif + tape source
testdata/           Synthetic git fixtures
```

---

## Contributing

Bug reports, PRs, and "this looks ugly on my terminal, here's a screenshot" issues all welcome. The pre-1.0 surface is intentionally small — please grill scope before opening a big PR.

If you want to help:

- **Try `gitshow replay` on your repo** and open an issue when something looks weird.
- **Add a fixture** to `testdata/` if your repo trips a bug — the smaller, the better.
- **Theme it** — `internal/theme` ships one default; Phase 5 will accept community themes.

---

## Acknowledgements

`gitshow` stands on the shoulders of giants:

- 🌿 [**charm.sh**](https://charm.sh) — Bubble Tea, Lip Gloss, VHS. The TUI ecosystem this project rides on.
- 📜 [**alecthomas/chroma**](https://github.com/alecthomas/chroma) — universal syntax highlighting that just works.
- 🔧 [**go-git/go-git**](https://github.com/go-git/go-git) — a pure-Go git implementation, so the binary stays single-file.
- 🍻 [**Homebrew**](https://brew.sh) — making `brew install` the easiest distribution path on Earth.

---

## License

MIT — see [LICENSE](./LICENSE). Use it, fork it, ship it.

<p align="center">
  Built by <a href="https://github.com/dhrupo">@dhrupo</a> · star the repo if it made you smile.
</p>
