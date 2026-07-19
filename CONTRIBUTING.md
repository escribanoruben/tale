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

Releases use [SemVer](https://semver.org/) git tags (`v0.1.0`, `v1.2.3`) and [GoReleaser](https://goreleaser.com/).

1. Merge changes to `main`.
2. Tag and push:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. The **Release** workflow builds binaries for Linux, macOS, and Windows (amd64 + arm64) and publishes a GitHub Release with checksums.

CI runs on every push/PR to `main` (`go test ./...` + build smoke test).
