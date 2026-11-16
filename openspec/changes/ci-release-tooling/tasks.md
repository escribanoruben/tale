## 1. CI

- [x] 1.1 Add `go-mod-tidy` job (tidy + `git diff --exit-code go.mod go.sum`)
- [x] 1.2 Add gitleaks job (full history, `GITLEAKS_ENABLE_COMMENTS: false`, PR-safe permissions)
- [x] 1.3 Bump `actions/setup-go` v5 → v7 in CI and Release workflows

## 2. Dependency automation

- [x] 2.1 Add `.github/dependabot.yml` (weekly gomod + github-actions)

## 3. Release tooling

- [x] 3.1 Add svu as `go` tool in `go.mod`
- [x] 3.2 Add `scripts/next-version.sh` (svu-based bump, `BUMP=` override)
- [x] 3.3 Add `scripts/release.sh` (tests → tag → push Forgejo → push-mirror sync via API)
- [x] 3.4 Add Makefile targets `next-version`, `release-plan`, `release`

## 4. Docs

- [x] 4.1 Update `CONTRIBUTING.md` release section (svu flow, Conventional Commits, mirror)

## 5. Verification

- [x] 5.1 Run `go test ./...`, `make release-plan`, and `openspec validate ci-release-tooling`
- [ ] 5.2 CI green on GitHub after merge (via push mirror)
