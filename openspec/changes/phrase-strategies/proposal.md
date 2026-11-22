## Why

The semantic engine currently hard-codes one strategy (chained story phrases). Users want different mnemonic shapes for different keys: the proven chained story at 16 letters (4 phrases) or 20 letters (5 phrases) — same strategy, different count — and a new noun-only mode where the password is spelled by X nouns that share a semantic field (one mental scene) or form an association path where each word is close to the next. Making the strategy a profile key turns all of these into TOML choices.

## What Changes

- Add a `strategy` profile key: `chained_story` (default, current behavior), `noun_field`, `noun_path`. For noun strategies the password length equals the noun count (one noun per password letter); `block_size` grouping stays as-is.
- **`noun_field`**: the X nouns orbit an adaptive centroid (seeded by CSPRNG, 3 re-centering rounds) — one semantic field, one mental scene. No repeated words or roots; low-imageability words excluded.
- **`noun_path`**: consecutive-adjacency optimization (k-best Viterbi) with a stutter guard (pairs closer than 0.75 are penalized at weight 2.0) and a mask-and-retry loop so no word repeats.
- Experiment numbers (ES, 30 passwords, 16 nouns): field pairwise cosine 0.353 vs 0.155 random (stutters 0.10/story, 2.7 ms); path consecutive 0.562 vs 0.154 (stutters 0.93/story with guard, ~90 ms).
- Strategy-specific recovery text (nouns list vs phrases) in the CLI.

## Non-goals

- TUI support for noun strategies (candidate cycling comes with the TUI slice).
- Mixing strategies per key; per-word roles (adjectives/verbs) in noun strategies.
- Changing the built-in default password length (es/en stay at 20 letters).

## Capabilities

### New capability: `phrase-strategies`

Strategy selection, the three strategies, and their requirements on vectors and output.

## Impact

- `internal/profile` (strategy field + validation), `internal/semantic` (noun strategies), `internal/phrase` (wiring), `cmd/tale` (generate output), openspec `semantic-phrases` unaffected (its scope remains the story engine).
- Security posture unchanged: password letters are CSPRNG-drawn before any word selection.
