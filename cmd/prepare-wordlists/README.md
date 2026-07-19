# prepare-wordlists

Regenerates curated English/Spanish word lists from upstream sources.

## Usage

```bash
go run ./cmd/prepare-wordlists
```

Run from the repository root (or pass `--repo /path/to/tale`).

## Output

1. `internal/worddata/{en,es}/{letter}/{nouns,adjectives,verbs}.tsv`
2. `internal/profile/*.toml`
3. `dist/data-{en,es}.zip` (optional release archives)

After regenerating, rebuild the binary:

```bash
go build -o tale ./cmd/tale
./tale stats
```

Statistics are available via `tale stats` (embedded data) — no separate stats tool needed.

## Curation

| Language | Source | Processing |
|----------|--------|------------|
| English | Top 20k nouns/adjectives/verbs by frequency | Lemma dedup, length 3–14 |
| Spanish | `doozan/spanish_data` lemmas | Frequency threshold, no diminutives |

Up to **120** words per `(language, letter, part of speech)` file.

Builtin alphabet sizes (top initials by corpus weight): **22** for English (`j`, `k` added), **20** for Spanish.
