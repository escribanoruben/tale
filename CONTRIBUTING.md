# Contributing

## GitHub issues

**Write issues in English** (title, description, comments).

This keeps the issue tracker consistent and easy to search for anyone following the project.

## Development

See [README.md](README.md) for build and test instructions.

OpenSpec changes live under [`openspec/changes/`](openspec/changes/).

```bash
make build
make test
make wordlists    # regenerate word data (requires network)
tale profile list
openspec validate <change-id>
```

## Pull requests

- Keep PR descriptions in English when possible.
- Focused diffs; match existing code style.
- Run `make test` before submitting.

## Releases

Releases use [SemVer](https://semver.org/) tags (`v0.1.0`, `v1.2.3`) and [GoReleaser](https://goreleaser.com/).
The version bump is computed with [svu](https://svu.js.org) from **[Conventional Commits](https://www.conventionalcommits.org/)**
since the last `v*` tag (`fix:`/`perf:`/`revert:` → patch, `feat:` → minor,
`feat!:` or `BREAKING CHANGE:` → major; `chore:`, `docs:`, `ci:`, … don't count).

1. Merge changes to `main` (write commits with `feat:` / `fix:` prefixes so bumps are automatic).
2. Review the bump and cut the release:
   ```bash
   make release-plan            # show current tag, bump type and next version
   make release                 # auto bump from Conventional Commits
   make release BUMP=patch      # force a bump (e.g. only chores)
   make release VERSION=0.2.0   # explicit version
   ```
   This runs the tests, tags, and pushes `main` + tag to **Forgejo** (`origin`); the Forgejo push mirror replicates the tag to GitHub.
3. The tag on **GitHub** starts the **Release** workflow: GoReleaser builds binaries for Linux, macOS, and Windows (amd64 + arm64) and publishes a GitHub Release with checksums.

> First release on a fresh clone without prior tags: use `make release VERSION=x.y.z`.

CI runs on every push/PR to `main` (`go test ./...` + build smoke test, `go mod tidy` alignment, and gitleaks secret scanning of the full history).
Dependency updates come from Dependabot (weekly `gomod` + `github-actions`).
