# Contributing to gh-starview

Thanks for your interest! This is a Termux-friendly, single-binary Go project.

## Quick Start

```bash
git clone https://github.com/Pastalikek65/gh-starview.git
cd gh-starview
go vet ./...
go test ./... -count=1 -timeout 30s -cover
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=dev" -o gh-starview .
./gh-starview --help
GH_STARVIEW_PLAIN=1 ./gh-starview Pastalikek65 --limit 5 --json | jq
```

## Development Rules

- **TDD:** add failing test first (`internal/*_test.go`), then minimal fix, then `make test`
- **Phone-friendly:** keep `go build` <10s, binary <15M, `CGO_ENABLED=0`, `modernc.org/sqlite` only
- **No speculative features:** YAGNI — only what `CHANGELOG.md` asks for
- **Surgical edits:** don't refactor unrelated files; match existing style

## Pull Requests

1. Fork, create branch `feat/<short-name>`
2. `make vet && make test` must pass (CI checks `60%` coverage)
3. Update `README.md` and `CHANGELOG.md` if user-facing
4. Open PR with clear description + `Fixes #<issue>` if any

## Release

- Tag `v*` pushes trigger `goreleaser` (`.goreleaser.yaml`) → 5 binaries + `checksums.txt`
- Check `gh release view <tag>` and `gh extension install Pastalikek65/gh-starview`

## Questions?

Open an issue or discussion — we respond within 2 days.
