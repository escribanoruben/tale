# tale

**A CLI implementation of Tales Vault System.**

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
| Language-adaptive alphabet | Each language defines its own restricted letter set from corpus frequency — not a global fixed size |
| Mnemonics over raw entropy | Fewer initials trade a small entropy cost for better phrase quality; default passwords still carry ~86–89 bits |
| Fixed template | `noun → adjective → verb → noun` per 4-character block |
| No letter picking | User does not choose initials; TUI mode only picks among valid words |
| Local-first | Works fully offline — word lists are embedded in the binary |
| Single binary | ~3 MB with curated EN + ES dictionaries (~13k words, ~106 KB of text) |

Built-in alphabets are **asymmetric on purpose**: English uses 22 letters (`j`, `k` added after the original top 20); Spanish stays at 20. Initial-letter distributions differ by language, so each builtin picks its own balance between dictionary coverage and memorable words. Custom languages set any alphabet in `~/.tale/profiles/{lang}.toml` — length is not locked to the builtins.

Restricting the alphabet does not collapse security: dropping rare initials costs only a few bits compared with the full Latin set, while block-level rejection keeps every 4-letter chunk phraseable. Need more entropy without widening the alphabet? Increase `--length` instead.

## Status

| Component | Status |
|-----------|--------|
| Word lists (EN, ES) | Ready — `internal/worddata/` |
| Languages | Ready — embedded `en` / `es` (+ custom under `~/.tale`) |
| `tale stats` | Ready |
| `tale generate` | Ready — automatic password + mnemonic phrases |
| `tale tui` | Ready — interactive word picker (rodillo) |

## Quick start

### Install

#### With Go (recommended)

If you have [Go 1.22+](https://go.dev/dl/) installed:

```bash
go install github.com/escribanoruben/tale/cmd/tale@latest
```

Pin a specific release:

```bash
go install github.com/escribanoruben/tale/cmd/tale@v0.1.1
```

> **Note:** Release `v0.1.0` used the previous module path `github.com/escribanoruben/tales_vault_system`. From `v0.1.1` onward, use `github.com/escribanoruben/tale`.

Ensure `$(go env GOPATH)/bin` is on your `PATH`, then run `tale`.

#### Pre-built binary (no Go required)

Download from [GitHub Releases](https://github.com/escribanoruben/tale/releases/latest):

| OS | CPU | Archive |
|----|-----|---------|
| Linux | x86_64 (Intel/AMD) | `*_linux_amd64.tar.gz` |
| Linux | arm64 (Raspberry Pi, etc.) | `*_linux_arm64.tar.gz` |
| macOS | Apple Silicon (M1/M2/M3) | `*_darwin_arm64.tar.gz` |
| macOS | Intel | `*_darwin_amd64.tar.gz` |
| Windows | x86_64 | `*_windows_amd64.zip` |
| Windows | arm64 | `*_windows_arm64.zip` |

Extract the `tale` binary (on Windows, `tale.exe`) and place it on your `PATH`.

Verify with `tale version`. Checksums are in `checksums.txt` on the same release page.

#### From source

```bash
git clone https://github.com/escribanoruben/tale.git
cd tale
make build          # or: go build -o tale ./cmd/tale
make install        # optional: install to $GOPATH/bin
```

### Run

```bash
make run            # TUI without installing (language from ~/.tale/settings or system locale)
./tale stats        # or: make stats
./tale version      # show version, commit, and build date

# Generate password + mnemonic phrases (automatic)
./tale generate --language en
./tale generate --language en --length 16
./tale generate --language es

# Interactive mnemonic picker
./tale tui --language en
./tale tui --language en --password abcdefghijklmnopqrst
./tale tui --language es

# Customize a built-in language (replaces embedded defaults for that code)
tale language init es
tale generate --language es

# Add a custom language with its own alphabet (scaffolds empty word lists)
tale language new pt --alphabet=abcdefghijlmnopqrstuvxz --display-name Portuguese
# Fill ~/.tale/data/pt/{letter}/*.tsv, then:
tale generate --language pt
```

Run tests: `make test` (or `go test ./...`).

See `make help` for all targets.

## Languages

| Language | Code | Alphabet |
|----------|------|----------|
| English | `en` | 22 letters: `abcdefghijklmnoprstuvw` |
| Spanish | `es` | 20 letters: `abcdefghijlmnopqrstv` |

User language files live under `~/.tale/profiles/{language}.toml` and **replace** any built-in for that code. Customized languages load words only from `~/.tale/data/{language}/` (no embedded fallback). See `tale language list|show|init|new`.

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

OpenSpec changes:

- [`openspec/changes/mvp-password-mnemonic-generator/`](openspec/changes/mvp-password-mnemonic-generator/)
- [`openspec/changes/language-first-custom-languages/`](openspec/changes/language-first-custom-languages/)

Contributing notes (including **GitHub issues in English**): [CONTRIBUTING.md](CONTRIBUTING.md)

```bash
openspec status --change language-first-custom-languages
```

## License

TBD. Word list sources have their own licenses — see [internal/worddata/README.md](internal/worddata/README.md).
