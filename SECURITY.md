# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :x:                |

## Reporting a Vulnerability

Please **do not** open a public issue. Instead:

1. Use GitHub's private vulnerability reporting: https://github.com/Pastalikek65/gh-starview/security/advisories/new
2. Or email the maintainer via GitHub profile.

We aim to respond within 48 hours and to release a fix within 14 days for critical issues.

## Token Handling

- `GITHUB_TOKEN` > `GH_TOKEN` > `gh auth token` (2s timeout) — never logged
- No token is written to cache (`~/.cache/gh-starview/gh-starview.db` stores only public repo metadata)
- Use `GITHUB_TOKEN` with minimal `public_repo` scope for `gh-starview`; do not use `repo` unless `--private` is implemented.

## Dependencies

- `go 1.25` (fixes GO-2026-6218 etc.), `modernc.org/sqlite` pure Go (no CGO)
- Run `govulncheck ./...` before release; `dependabot` auto-updates weekly.
