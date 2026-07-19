# Word list data

Curated noun, adjective, and verb lists embedded in the `tale` binary via `go:embed`.

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

Use `tale stats` to inspect word counts per language, letter, and part of speech.

## Sources and licenses

| Language | Source | License |
|----------|--------|---------|
| English | david47k/top-english-wordlists | Check upstream repo |
| Spanish | doozan/spanish_data `frequency.csv` | CC-BY-SA 3.0 |
