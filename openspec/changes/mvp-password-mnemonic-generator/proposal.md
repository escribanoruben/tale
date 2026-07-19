## Why

Secure passwords are hard to remember, while user-chosen memorable phrases are often predictable. Tales Vault System inverts the flow: generate the password first with a CSPRNG, then derive mnemonic phrases from fixed grammatical templates so users never pick letters or words upfront. A working executable in English and Spanish is required before publishing the companion blog post.

## What Changes

- Introduce a Go binary named `tale` with curated word lists and language configurations **embedded** at build time (`go:embed`).
- Split passwords into 4-character blocks and build mnemonic phrases using **noun → adjective → verb → noun**.
- Ship built-in languages `en` and `es` in the binary (selected with `--language`).
- Support **random mode** (automatic) and **TUI mode** (user picks words from candidates).
- Default password length: 20 characters (5 phrases of 4 words each).
- Provide `tale stats` to inspect embedded dictionary coverage.

## Capabilities

### New Capabilities

- `password-generation`: CSPRNG password creation, alphabet constraints, block splitting, rejection when blocks lack phrase coverage.
- `mnemonic-phrases`: Dictionary lookup, phrase assembly, memorability scoring, normalization for recovery.
- `language-profiles`: Embedded TOML language configs keyed by language code; letter-based data layout.
- `embedded-word-data`: `go:embed` word lists and language configs; `tale stats` command.
- `cli-tui`: `tale` binary with `generate`, `tui`, and `stats` subcommands.

### Modified Capabilities

<!-- None — greenfield project -->

## Impact

- Go module with embedded `internal/worddata/` (~106 KB text, ~3 MB binary total).
- No first-run download or `~/.tale/data/` requirement.
- Maintainer script regenerates `internal/worddata/` before release builds.

## Non-goals

- User-chosen initials or free-form phrase composition.
- Cloud sync, account management, or password vault storage.
- Additional languages beyond English and Spanish in v1.
- Runtime dictionary downloads.
