# Changelog

All notable changes to `gh-starview` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2026-08-25
### Fixed
- `goreleaser` ldflags: move `-trimpath` from `ldflags` to `flags` (fix `flag provided but not defined` on darwin)
- Go 1.25 upgrade (fix 5 stdlib vulns: GO-2026-6218, GO-2026-6090, etc.)

### Added
- Release assets for `darwin-amd64`/`darwin-arm64` + `checksums.txt`
- Topics `bubbletea cli gh-extension github go sqlite termux tui`

## [0.2.0] - 2026-08-25
### Added
- Interactive filter `/` with `textinput` (enter/esc/backspace/ctrl+u), `90.4%` tui coverage
- Pagination via `Link` header (`rel="next"`), `PathEscape` + user validation (`^[a-zA-Z0-9-]{1,39}$`)
- Timeout `10s` (http.Client) + `15s` context, `429` handling, `X-RateLimit-Remaining`
- Cache `url TEXT PRIMARY KEY` + migration, `0700` perms + `TempDir` fallback
- `main` integration tests (mock server + offline fallback), total `69.6%`
- `demo.gif` 900x700, `goreleaser` + `release.yml`, CI coverage gate `60%`
- `SECURITY.md` `CODE_OF_CONDUCT.md` `CONTRIBUTING.md` `.editorconfig` `dependabot`

### Changed
- `README`: offline banner `⚠️`, keys `/`+`esc`, pagination docs, `make cover` 69.6%
- `Makefile`: `CGO_ENABLED=0` `trimpath` `-X main.version` `cover/lint`
- Binary removed from git (`git rm --cached`), use release assets only

## [0.1.0] - 2026-08-25
### Added
- Initial `gh` extension: `bubbletea` + `lipgloss` TUI, `cobra` CLI, `modernc.org/sqlite` WAL cache
- `gh starview [user] [--sort stars|name|forks|updated] [--limit 1-100] [--json] [--no-cache]`
- Plain fallback for non-TTY (`GH_STARVIEW_PLAIN=1`), JSON mode, offline cache
- `internal/cache|config|github|tui` with TDD, `ci.yml` + `go test` <0.2s
