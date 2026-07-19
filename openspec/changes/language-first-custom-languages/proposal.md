## Why

The product model is language-first (`en` / `es`), but OpenSpec still describes `--profile english_default` / `spanish_default`, and creating a truly custom language still means cloning another language then hand-editing TOML and data. Users need a clear language selector, documented full overrides, and a dedicated path to add a new language with its own alphabet (not locked to 20 letters).

## What Changes

- Treat **language code** as the public selector (`--language en|es`, plus user-defined codes); retire profile IDs as the user-facing concept (**BREAKING** for any docs/scripts still using `--profile`).
- Document and keep **full override** semantics: a user language config replaces the builtin for that code; user word data replaces embedded lists for that language (no hybrid merge).
- Add a dedicated **custom language** command that requires an alphabet (variable length), writes `~/.tale/profiles/{lang}.toml`, and scaffolds `~/.tale/data/{lang}/` for those letters only.
- Keep `init --from` as a clone shortcut for near-forks of an existing language.
- Align README, Makefile defaults, and OpenSpec with the language-first model; update the stale MVP change specs for coherence.
- Surface clear messages when running against user overrides (and how builtins no longer apply).

## Non-goals

- Interactive wizard beyond required flags (alphabet via flag or simple prompt is enough).
- Hybrid merge of embedded + user word lists.
- Multiple presets/flavors per language.
- Non-ASCII alphabet letters / `ñ` in alphabets (normalization and folder layout stay as today).

## Capabilities

### New Capabilities

- `custom-languages`: Create and manage user-defined languages with a required alphabet and empty or stubbed letter data layout.

### Modified Capabilities

- `language-profiles`: Language codes replace profile IDs; builtins are `en`/`es`; user TOML overrides builtin for the same language; alphabet length is free for custom languages (builtins remain 20-letter curated sets).
- `cli-tui`: `generate` / `tui` use `--language`; language management CLI (`language` / rename from `profile` where appropriate); override messaging.
- `embedded-word-data`: User data root takes over entirely when the language is not builtin-backed.

## Impact

- CLI: `cmd/tale` flags, help, `profile` subcommand (rename or alias to `language`), README, Makefile `ARGS`.
- Config/data: `internal/profile` init/new paths, `internal/dictionary` override messaging, paths under `~/.tale/`.
- OpenSpec: this change’s delta specs; coherent updates to `mvp-password-mnemonic-generator` specs still describing `--profile`.
- Blog/executable requirement unchanged: English and Spanish remain the shipped builtins.
