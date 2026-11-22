## Why

The sibling project `odoo_mcp` (Forgejo primary, GitHub mirror) has a proven CI and release setup: Dependabot for Go modules and Actions, a secret-scanning job, a `go mod tidy` guard, and an svu-based release flow that pushes the tag through the Forgejo push mirror. `tale` still relies on manual Dependabot-less updates, has no secret scanning, and releases require hand-pushing tags without a documented bump calculation.

## What Changes

- Add **Dependabot** config: weekly `gomod` and `github-actions` updates with `chore(deps)` / `chore(ci)` commit prefixes.
- Harden **CI**: add a `go mod tidy` alignment check and a **gitleaks** job that scans full history (with `GITLEAKS_ENABLE_COMMENTS: false` so Dependabot's read-only token does not fail the job); bump `actions/setup-go` v5 → v7.
- Add **svu** as a `go` tool (`go.mod` tool directive) and release targets: `make next-version`, `make release-plan`, `make release` (SemVer bump derived from Conventional Commits, or forced with `BUMP=` / `VERSION=`).
- Add `scripts/release.sh`: runs tests, tags, pushes `main` + tag to Forgejo (`origin`), triggers the Forgejo push-mirror sync via API so the tag reaches GitHub and starts the GoReleaser workflow.
- Update `CONTRIBUTING.md` release section to the new flow; note Conventional Commits as the convention that enables automatic bumps.

## Non-goals

- Docker image / ghcr publishing: `tale` is an interactive TUI distributed via GoReleaser binaries; containerization adds no value here.
- Gitleaks as a local `go` tool: its dependency tree would bloat `go.mod` for a lint convenience; CI installs it via the official action.
- Rewriting existing commit history into Conventional Commits.

## Capabilities

No product capability changes: repository infrastructure only (CI, dependency automation, release tooling).

## Impact

- `.github/workflows/ci.yml`, `.github/dependabot.yml` (new).
- `go.mod`/`go.sum` (tool directive for svu; no runtime dependencies).
- `Makefile`, `scripts/next-version.sh` (new), `scripts/release.sh` (new), `CONTRIBUTING.md`.
- Releases keep working exactly as before on GitHub (GoReleaser builds Linux/macOS/Windows binaries and checksums).
