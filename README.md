# gh-starview

[![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev)
[![gh extension](https://img.shields.io/badge/gh-extension-24292F?logo=github)](https://cli.github.com/manual/gh_extension)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Termux](https://img.shields.io/badge/Termux-friendly-brightgreen?logo=android)](https://termux.dev)

**TUI for GitHub repo metrics — `gh` CLI extension with offline SQLite cache.**

`gh starview` lists your repos in a sortable terminal table (stars / forks / updated) — built for **Termux Debian** on phone: single binary <15MB, <40MB RAM, `CGO=0`, `modernc.org/sqlite` WAL cache, `go build` 4s.

```bash
gh extension install Pastalikek65/gh-starview
gh starview                      # your repos (auto-detected via gh auth)
gh starview Pastalikek65 --sort stars
gh starview --sort updated --limit 10
gh starview --json | jq '.[].Name'
```

![demo](https://raw.githubusercontent.com/Pastalikek65/gh-starview/main/demo.gif)

> **Phone-friendly:** no Docker, no daemon, single `~/.cache/gh-starview/gh-starview.db` file. Works offline after first fetch.

## Why

`gh repo list` is plain text — you can't see at a glance which README is weak or which repo needs archiving. `gh-starview` gives you:

- **Sortable TUI** (`bubbletea` + `lipgloss`): `s` stars, `n` name, `f` forks, `u` updated, `j/k` navigate, `q` quit
- **Offline cache:** SQLite WAL (`modernc.org/sqlite`, pure Go, no CGO) — airplane mode still shows last fetch
- **Plain fallback:** auto-detects non-TTY (CI, `GH_STARVIEW_PLAIN=1`) → prints lipgloss table without TUI
- **JSON mode:** `gh starview --json` for scripting

## Install

### As `gh` extension (recommended)

```bash
gh extension install Pastalikek65/gh-starview
gh starview --help
```

### From source (Termux / Debian)

```bash
git clone https://github.com/Pastalikek65/gh-starview.git
cd gh-starview
go build -ldflags="-s -w" -o gh-starview .
./gh-starview --help
# optional: install as extension locally
gh extension install .
```

### `go install`

```bash
go install github.com/Pastalikek65/gh-starview@latest
gh-starview Pastalikek65 --sort stars
```

## Usage

```bash
# your account (detected via `gh api user`)
gh starview
gh starview --sort stars      # default
gh starview --sort name
gh starview --sort forks
gh starview --sort updated

# any user
gh starview torvalds --limit 5
gh starview --json --limit 20 | jq

# bypass cache (force fetch)
gh starview --no-cache

# offline — uses cache, shows [offline] banner on stderr
# (after first successful fetch)
```

**Keys in TUI:** `↑/k` `↓/j` navigate · `s` stars · `n` name · `f` forks · `u` updated · `q` quit

**Cache:** `~/.cache/gh-starview/gh-starview.db` (or `$XDG_CACHE_HOME/gh-starview/...`). `PRAGMA journal_mode=WAL`.

## Tech

- **Go 1.24**, `cobra`, `bubbletea` `v1.3.4`, `lipgloss` `v1.1.0`, `modernc.org/sqlite` `v1.38.2` (pure Go, no CGO)
- Data flow: `github.Client.ListRepos` → `cache.Store.Upsert` → `tui.Model`
- API: `GET /users/:user/repos?per_page=100&sort=updated` (REST, paginated v1 = 100 repos)
- Rate-limit handling: shows cached `[rate-limited]` banner instead of failing

## Development

```bash
make test   # go test ./... -count=1 -timeout 30s
make vet    # go vet ./...
make build  # go build -ldflags="-s -w" -o gh-starview .
```

**Tests:** `internal/cache` (WAL + sort) · `internal/github` (httptest mock + rate-limit) · `internal/tui` (sort/filter) — all <100ms.

## Roadmap

- [ ] Pagination (>100 repos) via `Link` header
- [ ] `--fork` filter toggle, language filter `/`
- [ ] Stars sparkline (history via GraphQL)
- [ ] `gh starview --private` (needs `repo` scope, uses `/user/repos`)

## License

MIT — see [LICENSE](LICENSE)
