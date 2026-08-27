# gh-starview

[![Go 1.25](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![gh extension](https://img.shields.io/badge/gh-extension-24292F?logo=github)](https://cli.github.com/manual/gh_extension)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Termux](https://img.shields.io/badge/Termux-friendly-brightgreen?logo=android)](https://termux.dev)

I built `gh-starview` on my phone in Termux because `gh repo list` was unreadable. You get a scrollable table you can sort and filter without leaving the terminal. It caches to SQLite so it works on the subway.

```bash
gh extension install Pastalikek65/gh-starview
gh starview                      # your repos, auto-detected
gh starview Pastalikek65 --sort stars
gh starview --sort updated --limit 10
gh starview --json | jq '.[].Name'
```

![demo](https://raw.githubusercontent.com/Pastalikek65/gh-starview/main/demo.gif)

Single file cache at `~/.cache/gh-starview/gh-starview.db`. No Docker. No daemon. First fetch needs network, after that airplane mode shows your last data.

## What you get

`gh repo list` prints plain text. You cannot tell which repo needs a better README. This helps you see it:

- Press `s` for stars, `n` for name, `f` for forks, `u` for updated. `j/k` to move, `q` to quit.
- Press `/`, type a name, hit `enter` to filter. `esc` clears it.
- It detects when you are not in a TTY. In CI it prints a plain table. Set `GH_STARVIEW_PLAIN=1` to force it.
- Need JSON? `gh starview --json` pipes to `jq`.

## Install

**As a gh extension**

```bash
gh extension install Pastalikek65/gh-starview
gh starview --help
```

**From source on Termux or Debian**

```bash
git clone https://github.com/Pastalikek65/gh-starview.git
cd gh-starview
go build -ldflags="-s -w" -o gh-starview .
./gh-starview --help
gh extension install .   # local
```

**With go install**

```bash
go install github.com/Pastalikek65/gh-starview@latest
gh-starview Pastalikek65 --sort stars
```

## How to use it

```bash
# your account
gh starview
gh starview --sort stars      # default
gh starview --sort name
gh starview --sort forks
gh starview --sort updated

# someone else
gh starview torvalds --limit 5
gh starview --json --limit 20 | jq

# skip cache
gh starview --no-cache

# offline: shows a warning and your cached data
```

Keys: `↑/k` `↓/j` · `s` stars · `n` name · `f` forks · `u` updated · `/` filter · `esc` clear · `q` quit

Cache lives at `~/.cache/gh-starview/gh-starview.db` (or `XDG_CACHE_HOME`). It uses `WAL` and `0700`. If you hit a rate limit you still see your cached data and a `⚠️ rate limited` warning.

Tokens: I check `GITHUB_TOKEN`, then `GH_TOKEN`, then `gh auth token` with a 2 second timeout. I never log your token. User agent is `gh-starview/<version>`.

Completion: `gh-starview completion bash|zsh|fish`

## How it works

Go 1.25, `cobra`, `bubbletea` 1.3.4, `lipgloss` 1.1.0, `modernc.org/sqlite` 1.38.2. No CGO, so it builds on Termux. I run `CGO_ENABLED=0` and `trimpath`.

The flow is simple: `github.Client.ListRepos` handles pagination through the `Link` header and escapes the username, then `cache.Store` writes to SQLite with `url` as primary key, then `tui.Model` sorts and filters.

API call: `GET /users/:user/repos?per_page=100&sort=updated`. Rate limits: I treat `429` and `403` with `X-RateLimit-Remaining: 0` as `ErrRateLimited` and fall back to cache.

## Developing it

```bash
make test   # all packages, 75.3% total
make cover  # detailed coverage, 90% for tui, 90% github
make vet    # vet plus golangci-lint if you have it
make build  # 13M binary, 4s on my phone
./gh-starview --version # v0.2.1
```

Tests use `t.TempDir()` for SQLite and `httptest` for GitHub. No TTY needed. The `main` package has integration tests with a mock server and an offline fallback check.

## What's next

- [x] Pagination for more than 100 repos
- [x] Filter with `/`
- [x] Cross builds for android, linux, darwin with goreleaser
- [ ] Filter by fork and language
- [ ] Sparkline for stars history
- [ ] `gh starview --private` for private repos

## License

MIT — see [LICENSE](LICENSE). Built on a phone, tested on a phone.
