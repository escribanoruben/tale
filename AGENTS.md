# Agent guidance — Tales Vault System

IDE-agnostic instructions for coding agents working in this repository.

## Project

Go CLI that generates CSPRNG passwords and derives mnemonic phrases from a grammatical template per language (e.g. EN: `adjective → noun → verb → noun`; ES: `noun → adjective → verb → noun` per 4-character block). Local-first: no network at runtime; word lists are embedded.

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

Word-list regeneration (maintainers, needs network): `make wordlists`. Semantic
embedding assets (maintainers, needs network): `make embeddings`.

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

## Current status (Nov 2026)

- **Forgejo is the primary remote** (`origin` = ssh://forgejo-kilvin.gila-spectrum.ts.net/ruben/tale.git); GitHub is a push mirror — never push to GitHub manually.
- **Mnemonic architecture**: three user-selectable strategies in the profile TOML (`strategy`): `story` (default; whole-story DP over per-block candidate pools, chain edges + circular closure), `random` (uniform), and noun strategies `noun_field`/`noun_path` (adaptive centroid / association path; EXPERIMENTAL). The mnemonic artifact is called the **passtale**.
- **LLM polish is BETA and opt-in** (`llm_polish = true` in `~/.tale/settings.toml`): local ollama only (`llm_model`, default `qwen2.5:1.5b`; `qwen3.5:2b` tested), chat endpoint with `think: false`, `num_thread: 2` (more threads thrash this box), validation by initials always guarantees a valid passtale (fallback = telegraphic carriers). Render hit-rate varies by block; do not rely on the LLM for correctness.
- **Key invariants**: password letters are CSPRNG-drawn before any word selection (word choice never affects entropy); recovery = initials of the carrier words printed as `portadoras` (fillers are decoration); `phrase_template` supports the trío shape `[noun verb noun adj]` as an alternative.
- **Open changes** (`openspec list`): `phrase-strategies` (15/17; pending CI tasks only), `semantic-phrases` (open; most tasks done, remaining are minor doc/CI items). `mvp-password-mnemonic-generator` still has open tasks from the original MVP.
- **Machine quirks (kilvin)**: ollama pinned fixes — governor Performance, no `AllowedCPUs` pin, `OLLAMA_NUM_PARALLEL=4`; 4 cores, ~7 GB free RAM → CPU models up to ~3B; `systemctl edit ollama` opens with `SYSTEMD_EDITOR=hx`.
