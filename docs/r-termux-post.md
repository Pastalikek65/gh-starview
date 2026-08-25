# r/termux — Showcase Post Draft for gh-starview

**Status: DRAFT — copy-paste to Reddit. Verify flair/rules live before posting (Reddit blocked our stealth browser, inferred from Termux community docs + explorer).**

## Title (pick one, keep under 100 chars)

Option A (recommended, explorer pattern):
```
[Showcase] Built gh-starview on my phone in Termux — sortable TUI for gh repo list with offline cache (13MB, no CGO)
```

Option B (shorter):
```
[Showcase] gh-starview — TUI for gh repo list built entirely in Termux (offline SQLite, 13MB)
```

## Flair

- Use **Showcase** (if not found, **Showoff**). Check https://www.reddit.com/r/termux/ submit page — do NOT use Question/Help. If flair is required and missing, AutoMod will remove.

## Timing

- **Tue-Thu 14:00-18:00 UTC** (best). Avoid Mon 00-06 UTC and weekends. Post once, stay 90 min to reply to first comments.

## Body (copy-paste, keep code blocks)

I built `gh-starview` on my phone in Termux because `gh repo list` was unreadable on a small screen. You get a scrollable table you can sort and filter without leaving the terminal. It caches to SQLite so it works on the subway.

![demo](https://raw.githubusercontent.com/Pastalikek65/gh-starview/main/demo.gif)

**Why this is Termux-related (for Rule 4):**
- Built in Termux proot Debian on aarch64, `go build` 4s on my phone
- `CGO_ENABLED=0` + `modernc.org/sqlite` 1.38.2 pure Go — `mattn/go-sqlite3` fails in Termux, this one builds
- `android-arm64` goreleaser asset — `gh` on Termux reports `android-arm64` not `linux-arm64`, so I ship both
- 13M binary, <40M RAM, single file `~/.cache/gh-starview/gh-starview.db` WAL `0700`, works offline after first fetch
- TTY-aware: `bubbletea` TUI when interactive, plain table when piped (`GH_STARVIEW_PLAIN=1`) and `--json` for scripts
- Handles `>100` repos via `Link` header pagination, 10s timeout, falls back to cache on `429`/`403`

**Install — 3 ways (Termux):**

```bash
# 1. as gh extension (needs gh CLI)
gh extension install Pastalikek65/gh-starview
gh starview --help

# 2. go install
go install github.com/Pastalikek65/gh-starview@latest

# 3. from source in Termux
git clone https://github.com/Pastalikek65/gh-starview.git
cd gh-starview
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o gh-starview .
./gh-starview --help
```

**Usage:**

```bash
gh starview                      # your repos, auto-detected via gh auth
gh starview Pastalikek65 --sort stars
gh starview --sort updated --limit 10
gh starview --json | jq '.[].Name'

# keys: s stars, n name, f forks, u updated, / filter (type + enter), esc clear, j/k nav, q quit
```

Tech: Go 1.25, `cobra`, `bubbletea` 1.3.4, `lipgloss` 1.1.0, `modernc.org/sqlite`, `Link` pagination. Repo: https://github.com/Pastalikek65/gh-starview — MIT. Tested on Termux aarch64, feedback welcome, especially install reports on `android-arm64`.

---

## Checklist Before Posting

- [ ] Open https://www.reddit.com/r/termux/ → check sidebar Rules + flair list (Showcase vs Showoff) — our browser was blocked (network security), inferred. Verify live.
- [ ] `gh release view v0.2.1` shows `gh-starview-android-arm64` asset exists (it does, 13M)
- [ ] Demo gif loads: https://raw.githubusercontent.com/Pastalikek65/gh-starview/main/demo.gif (47K, 4-frame animated)
- [ ] No "please star" wording (Rule 7). Disclosure: this is your own project (expected for Showcase).
- [ ] Keep post under 300 words + code blocks + gif. Reply to first 3 comments within 60 min to boost.

## Risks (from explorer)

- Link-drop without Termux proof → removal. Mitigated by leading with "Built on my phone in Termux" + CGO/4s/android-arm64 details.
- Wrong flair → filtered. Verify live.
- Low-traffic hour → buried. Use Tue-Thu 14-18 UTC.
