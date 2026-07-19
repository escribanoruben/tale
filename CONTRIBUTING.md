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
