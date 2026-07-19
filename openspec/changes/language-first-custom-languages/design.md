## Context

Runtime already resolves generation by language code (`en` / `es`): embedded TOML under `internal/profile/{lang}.toml`, optional user override at `~/.tale/profiles/{lang}.toml`, and word lists from embed or `~/.tale/data/{lang}/` with no merge. OpenSpec and some docs still describe `--profile english_default`. Adding a custom language today clones another language’s alphabet and data via `tale profile init --from`, which is a poor fit when the alphabet differs.

Word-list provenance: curated lists are prepared offline (`prepare-wordlists`) and embedded; user lists are local files the user owns. Normalization for recovery remains `internal/normalize` (strip accents, `ñ`→`n`); builtin alphabets exclude letters that would conflict with that mapping in Spanish.

## Goals / Non-Goals

**Goals:**

- Make language the only public selector for generation and TUI.
- Keep full override (config + data) when a user language file exists.
- Provide `language new` (or equivalent) that requires an alphabet and scaffolds letter folders for that alphabet only.
- Align docs and OpenSpec (this change + stale MVP specs) with the language-first model.
- Clear stderr/stdout hints when user overrides are active.

**Non-Goals:**

- Full interactive wizard; hybrid embed+user merge; per-language presets; non-ASCII alphabets.

## Decisions

### 1. Public CLI: `--language`, not `--profile`

**Choice:** Keep `--language`; do not reintroduce `--profile`. Default language from settings or locale/`en`.

**Alternatives:** Soft-alias `--profile` → language (rejected for this change to avoid two names).

### 2. Rename management subcommand to `language`

**Choice:** Prefer `tale language {list,show,init,new}`; keep `tale profile …` as a thin deprecated alias printing a one-line hint, or remove if unused externally.

**Alternatives:** Only document “profile means language config” (worse UX).

### 3. `language new` vs `language init --from`

| Command | Role |
|---------|------|
| `language new <code> --alphabet=…` | Custom language: required alphabet, default phrase template/block size, scaffold empty TSV stubs per letter × POS |
| `language init <code> [--from LANG]` | Clone existing language config + seed data (current behavior) |

Alphabet validation: unique lowercase ASCII letters (same rules as today). No fixed length of 20; optional soft warning if length &lt; 10. Builtin `en`/`es` alphabets remain curated 20-letter sets.

### 4. Full override, no merge

When `~/.tale/profiles/{lang}.toml` exists → user config. When that profile is non-builtin → dictionary reads only `~/.tale/data` for that language. Deleting the user profile file restores the builtin (if any).

**Uncovered letter blocks:** generation/TUI already surface missing word errors; custom languages with empty stubs MUST fail clearly until the user fills TSV files (same rejection path as empty candidate lists—no silent regeneration into another letter).

### 5. Docs and OpenSpec coherence

Update README/Makefile; rewrite MVP change specs that still say `--profile` so both open changes agree. This change’s specs are the normative language-first contract going forward.

## Risks / Trade-offs

- [Empty stubs break generate until filled] → Scaffold with header comments or README in data dir; CLI prints next steps after `language new`.
- [Users expect merge with embed] → Document override explicitly; show “user data” in `language show` / generate preamble.
- [Deprecated `profile` alias confuses] → Single hint line; remove later.
- [Alphabet too short → weak passwords] → Entropy display already reflects alphabet size; optional warning on `new`.

## Migration Plan

1. Ship CLI/docs; no data migration required.
2. Scripts using `--profile` must switch to `--language en|es`.
3. Existing `~/.tale/profiles/*.toml` keep working (already language-keyed).

## Open Questions

- None blocking: alias retention duration can be decided at apply time (default: keep alias one release).
