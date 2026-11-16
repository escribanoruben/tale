# Word list data

Curated noun, adjective, and verb lists embedded in the `tale` binary via `go:embed`,
plus int8-quantized semantic vectors per language (`embeddings.bin`).

## Layout

```text
data/
├── en/
│   ├── a/
│   │   ├── nouns.tsv
│   │   ├── adjectives.tsv
│   │   └── verbs.tsv
│   └── ...
└── es/
    └── ...
```

Each file contains one word per line in **alphabetical** order. A second column holds the within-bucket **frequency rank** (1 = most common, tab-separated):

```text
ability	15
absence	3
```

Up to **120** curated words per file.

The alphabet for each language is defined in `profiles/*.toml` (22 letters for English, 20 for Spanish by default). User profiles live in `~/.tale/profiles/`.

## User install layout

Not used — word lists ship inside the `tale` binary. Rebuild after regenerating data:

```bash
go run ./cmd/prepare-wordlists
go build -o tale ./cmd/tale
```

## Curation rules

### English

- Source: top 20k frequent words per part of speech ([david47k/top-english-wordlists](https://github.com/david47k/top-english-wordlists))
- Deduplicate inflected forms (`states` → keep `state`, `studies` → keep `study`)
- Drop abbreviations, words shorter than 3 or longer than 14 characters

### Spanish

- Source: frequency-ranked **lemmas** only ([doozan/spanish_data](https://github.com/doozan/spanish_data) `frequency.csv`)
- No inflected forms from Wiktionary dumps
- Exclude diminutives (`-ito`, `-ita`, `-illo`, …)
- Minimum corpus frequency threshold for inclusion
- Accented vowels at word start map to base vowel (`árbol` → `a`); no `ñ`-initial words
- **Adjectives carry derived feminine forms** (`quemado` → `quemada`): fastText vectors encode gender agreement, so the semantic scorer pairs adjectives with the noun's gender without any grammar code. Feminines inherit the masculine entry's rank and do not count against the per-bucket cap.

## Alphabets (in profiles)

| Language | Letters |
|----------|---------|
| English | `abcdefghijklmnoprstuvw` |
| Spanish | `abcdefghijlmnopqrstv` |

## Regenerating

```bash
go run ./cmd/prepare-wordlists   # or: make wordlists
```

Produces `data/{en,es}/…` and release zips in `dist/data-{en,es}.zip`.

## Semantic embeddings

`{en,es}/embeddings.bin` hold one int8-quantized 300-dim vector per embedded
word, used by `internal/semantic` to rank mnemonic phrases (see
`openspec/changes/semantic-phrases/`). Format: `TEMB` magic + version +
dim + word table + int8 rows + per-vector float32 scales; vectors are
L2-normalized at load time.

Regenerate (maintainers, needs network; `VECS=dir` reuses downloaded
`cc.{lang}.300.vec.gz` files):

```bash
make embeddings            # downloads fastText cc.{en,es}.300.vec.gz
make embeddings VECS=dir   # reuses pre-downloaded copies
```

Words without vectors (rare terms) are skipped by semantic ranking and
reported by the tool. Languages without an embeddings file silently fall
back to frequency-based scoring.

Use `tale stats` to inspect word counts per language, letter, and part of speech.

## Sources and licenses

Full attributions: [NOTICE](../../NOTICE) at the repository root.

| Language | Path | Upstream | License |
|----------|------|----------|---------|
| English | `internal/worddata/en/` | [david47k/top-english-wordlists](https://github.com/david47k/top-english-wordlists) | [CC-BY 3.0](https://creativecommons.org/licenses/by/3.0/) |
| Spanish | `internal/worddata/es/` | [doozan/spanish_data](https://github.com/doozan/spanish_data) `frequency.csv` (from [hermitdave/FrequencyWords](https://github.com/hermitdave/FrequencyWords)) | [CC-BY-SA 3.0](https://creativecommons.org/licenses/by-sa/3.0/) |

The English lists require attribution when redistributed. The Spanish lists are
derivative works and must stay under CC-BY-SA 3.0 (the curated TSV files in this
directory satisfy that obligation). The tale Go code remains under MIT; see
[LICENSE](../../LICENSE).
