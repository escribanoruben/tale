# Agent guidance — Tales Vault System

IDE-agnostic instructions for coding agents working in this repository.

## Project

Go CLI that generates CSPRNG passwords and derives mnemonic phrases from a fixed template (`noun → adjective → verb → noun` per 4-character block). Local-first: no network at runtime; word lists are embedded.

- Entrypoint: `cmd/tale`
- Languages: built-in `en` / `es`; custom under `~/.tale`
- Specs and changes: `openspec/`

## Commands

```bash
make build
make test
make help
openspec validate <change-id>
openspec status --change <change-id>
```

Word-list regeneration (maintainers, needs network): `make wordlists`.

## Conventions

- Prefer English for GitHub issues, PR descriptions, OpenSpec artifacts, and project docs (`CONTRIBUTING.md`).
- Match existing Go style; keep diffs focused.
- Language-first model: resolve by language code (`en` / `es`), not legacy profile names like `english_default`.
- User overrides at `~/.tale/profiles/{lang}.toml` replace embedded defaults for that code; customized languages load words only from `~/.tale/data/{lang}/` (no merge with embed).

## OpenSpec

Plan and track work under `openspec/changes/`. Use the OpenSpec CLI (`openspec …`); do not depend on IDE-specific tooling or directories.

When implementing a change, follow that change’s `tasks.md` and keep specs/design coherent with the language-first model.

## Security

- Password generation must stay CSPRNG-based; never invent “randomness” from non-crypto sources.
- Do not add network calls to runtime password/mnemonic paths.
- Never commit secrets, credentials, or personal `~/.tale` data.
