# Tales Vault System

Generate cryptographically random passwords and derive memorable mnemonic phrases from them — without choosing letters or words yourself.

The password comes first (CSPRNG). Mnemonic phrases are built afterward from a fixed grammatical template so recovery is predictable.

## How it works

```
[CSPRNG] → password (20 chars, language alphabet)
              ↓
         blocks of 4: lcds | mrtp | ...
              ↓
    noun + adjective + verb + noun (per block)
              ↓
    password + mnemonic phrases + recovery rules
```

**Example** (English block `lcds`):

- *Lion cowardly drifts silver*
- *Lamb clumsy dances slowly*

Each word's normalized initial must match the block letter (`l`, `c`, `d`, `s`).

## Design principles

| Principle | Detail |
|-----------|--------|
| Randomness first | Password is generated; the phrase is derived, never the reverse |
| Restricted alphabet | Top 20 most frequent word-initial letters per language |
| Fixed template | `noun → adjective → verb → noun` per 4-character block |
| No letter picking | User does not choose initials; TUI mode only picks among valid words |
| Local-first | Works fully offline — word lists are embedded in the binary |
| Single binary | ~3 MB with curated EN + ES dictionaries (~13k words, ~106 KB of text) |

## Status

| Component | Status |
|-----------|--------|
| Word lists (EN, ES) | Ready — `internal/worddata/` |
| Profiles | Ready — embedded in binary |
| `tale stats` | Ready |
| `tale generate` | Ready — automatic password + mnemonic phrases |
| `tale tui` | Ready — interactive word picker (rodillo) |

## Quick start

```bash
make build          # or: go build -o tale ./cmd/tale
./tale stats         # or: make stats

# Generate password + mnemonic phrases (automatic)
./tale generate --profile english_default
./tale generate --profile english_default --length 16
./tale generate --profile spanish_default

# Interactive mnemonic picker
./tale tui --profile english_default
./tale tui --profile english_default --password abcdefghijklmnopqrst
./tale tui --profile spanish_default

# Install (optional)
make install        # or: go install ./cmd/tale

# Custom profile from a built-in template
tale profile init my_english --from english_default
tale generate --profile my_english
```

Run tests: `make test` (or `go test ./...`).

See `make help` for all targets.

## Languages

| Language | Profile | Alphabet |
|----------|---------|----------|
| English | `english_default` | 20 letters: `abcdefghilmnoprstuvw` |
| Spanish | `spanish_default` | 20 letters: `abcdefghijlmnopqrstv` |

Custom profiles can be created under `~/.tale/profiles` (see `tale profile init`).

Spanish normalization rules:

- Accented vowels at word start map to the base vowel (`árbol` → `a`)
- Words starting with `ñ` are excluded; the alphabet does not include `ñ`

## Repository layout

```text
├── cmd/
│   ├── tale/                   # CLI entrypoint
│   └── prepare-wordlists/      # regenerate word data (maintainers)
├── internal/
│   ├── worddata/               # curated lists (go:embed)
│   ├── profile/                # TOML profiles (go:embed)
│   ├── password/               # CSPRNG generation
│   ├── dictionary/             # embedded word lookup
│   ├── phrase/                 # mnemonic selection + scoring
│   ├── normalize/              # initial-letter normalization
│   ├── tui/                    # Bubble Tea interactive picker
│   └── wordprepare/            # word list curation pipeline
└── openspec/
```

## Maintainers

Regenerate curated word lists:

```bash
make wordlists    # regenerate word data (requires network)
make build
make stats
```

See [internal/worddata/README.md](internal/worddata/README.md) for curation rules and sources.

## Development

OpenSpec change: [`openspec/changes/mvp-password-mnemonic-generator/`](openspec/changes/mvp-password-mnemonic-generator/)

Contributing notes (including **GitHub issues in English**): [CONTRIBUTING.md](CONTRIBUTING.md)

```bash
openspec status --change mvp-password-mnemonic-generator
```

## License

TBD. Word list sources have their own licenses — see [internal/worddata/README.md](internal/worddata/README.md).
