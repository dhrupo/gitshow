# gitshow

> Apple Keynote for Pull Requests.

A cinematic terminal storytelling engine for Git history.

`gitshow` turns commits, PRs, and releases into animated walkthroughs
you can play live in your terminal or export as a shareable gif.

```bash
gitshow replay              # cinematic replay of recent commits
gitshow pr 142              # walk a pull request
gitshow release v2.0.0      # release story
gitshow export gif          # share it
```

## Status

Early development — Phase 1 (`gitshow replay`).

## Vision

`git log --oneline` tells you *what* changed. `gitshow` tells you
*why, how, impact, risk, and evolution*. Existing alternatives miss
the gap:

| Tool | Weakness gitshow exploits |
|---|---|
| `git log` | Ugly, no narrative |
| Lazygit | Navigation-focused, not presentation |
| GitHub PRs | Static, web-only |
| VHS | Recording only, no Git intelligence |
| asciinema | Low-level playback, not story |

## Install

```bash
brew install gitshow        # (not yet published)
```

Or build from source:

```bash
git clone https://github.com/dhrupo/gitshow
cd gitshow
go build -o gitshow ./cmd/gitshow
./gitshow --help
```

## Phase 1 — MVP

- [x] Repo scaffold (Cobra + go-git)
- [ ] Commit + diff parsing (Chroma syntax highlighting)
- [ ] Bubble Tea TUI shell, 4-panel layout, keyboard nav
- [ ] Animation primitives (typewriter, stagger, fade)
- [ ] Markdown + JSON export
- [ ] Cinematic demo gif in this README

Excluded from MVP: AI summaries, GitHub integration, HTML / GIF / SVG
exports, theme variety, collaboration. Those land in Phases 2–6.

## Keyboard

| Key | Action |
|---|---|
| `space` | pause / resume |
| `←` / `→` | previous / next commit |
| `↑` / `↓` | faster / slower |
| `q` | quit |

## Development

```bash
go build -o gitshow ./cmd/gitshow      # build the binary
go test ./...                           # run unit + snapshot tests
./gitshow replay --commits 5            # try it on the current repo
```

## License

MIT — see [LICENSE](./LICENSE).
