## Context

`tale` is built and released through GitHub Actions on the GitHub mirror; Forgejo is the primary remote and syncs via its push mirror. The sibling project `odoo_mcp` ships a proven CI/release setup on the same model; this change ports the parts that fit a TUI binary project.

## Decisions

- **svu pinned to v3.2.4** (as a `go` tool in `go.mod`): v3.4.x requires Go ≥ 1.27 while the local toolchain is 1.26 (`GOTOOLCHAIN=local`). Bump the pin when the toolchain moves to 1.27+.
- **svu as a `go` tool, not a binary dependency**: tools live in the module graph only; the `tale` binary and its runtime dependency set are unaffected.
- **No gitleaks as a local `go` tool**: its dependency tree is large for a lint convenience. CI uses `gitleaks/gitleaks-action@v2` (installs its own binary). Local scans can run `go tool gitleaks` later if ever needed.
- **`GITLEAKS_ENABLE_COMMENTS: false` from day one**: Dependabot PRs run with a read-only token; comment attempts fail the job ("Resource not accessible by integration").
- **No Dockerfile / ghcr publishing**: `tale` is an interactive TUI; GoReleaser already distributes cross-platform binaries with checksums. Containers add no value for this shape of project.
- **No GoReleaser changes**: `release.sh` only cuts the tag and syncs the mirror; version injection stays in GoReleaser ldflags (`main.version` from `{{.Version}}`), so there is no version constant to edit in source.
- **Release script triggers the Forgejo push-mirror sync via API** after pushing the tag: a tag-only push may not trigger the automatic sync, and the GitHub Release workflow only starts once the tag reaches GitHub.

## Risks / Trade-offs

- Conventional Commits are a convention, not enforced: svu fails politely ("no commits warrant a release") and `BUMP=` / `VERSION=` overrides cover irregular histories.
- The push-mirror sync depends on stored Forgejo credentials (`~/.git-credentials` or the configured credential helper); on failure the script warns and the tag lands with the next commit push.
