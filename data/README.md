# Word data

Curated word lists live in [`internal/worddata/`](../internal/worddata/).

They are embedded into the `tale` binary at build time via `go:embed`.

Regenerate with:

```bash
make wordlists    # or: go run ./cmd/prepare-wordlists
```

Statistics:

```bash
tale stats
```
