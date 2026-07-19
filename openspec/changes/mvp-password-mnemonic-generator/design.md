## Context

Greenfield Go project. No existing specs or code. The MVP must be a credible, blog-ready executable that demonstrates the full flow: random password → blocks → mnemonic phrases → recovery rules. Profiles are TOML files so future languages or alphabets can be added without recompiling logic.

## Goals / Non-Goals

**Goals:**

- Local-first binary with two built-in languages (`en`, `es`).
- Password-first generation with phrase derivation (never the reverse).
- 20-character default passwords (5 × 4-letter blocks).
- Fixed 4-word template per block: noun, adjective, verb, noun.
- Curated word lists and language configs **embedded** in the binary (`go:embed`, ~106 KB of text).
- Random mode (auto-pick best phrase) and TUI mode (user picks among valid candidates).
- Explicit normalization rules so phrases reliably recover the password.
- `tale stats` to inspect embedded dictionary coverage.

**Non-Goals:**

- Vault storage, browser integration, or cloud features.
- Custom user-defined phrase templates in v1.
- Runtime dictionary downloads.
- Perfect memorability modeling; v1 uses a simple deterministic heuristic.

## Decisions

### 1. Language configuration format: TOML

**Choice:** One TOML file per language under `internal/profile/{language}.toml`, embedded in the binary.

**Rationale:** Human-readable, easy to diff, no code changes for new alphabets or lengths. Public CLI selects by `--language` code (`en` / `es`).

**Example fields:**

```toml
id = "es"
language = "es"
display_name = "Spanish (default)"
password_length = 20
block_size = 4
alphabet = "abcdefghijlmnopqrstv"  # 20 letters, derived from corpus
phrase_template = ["noun", "adjective", "verb", "noun"]
data_layout = "letter"

[normalization]
strip_accents = true
lowercase = true
exclude_initial_enye = true
```

Language configs and word data ship inside the binary. Lookup path: `internal/worddata/{language}/{letter}/{nouns,adjectives,verbs}.tsv`.

### 2. Alphabet: top 20 word-initial letters per language

**Choice:** Compute from bundled word lists during data prep; store the resulting alphabet in each profile.

**Rationale:** Maximizes dictionary coverage while keeping passwords mnemonic-friendly. Exact letters are data-driven, not hand-waved.

**Methodology (maintainer script):**

1. Normalize each word's initial letter (lowercase, strip accents: `árbol` → `a`).
2. For Spanish: **exclude** words whose first character is `ñ` (they are not needed in the bundled lists).
3. Index every remaining word by its normalized initial letter (e.g. `árbol` is indexed under `a` in the Spanish noun list).
4. Count frequency of normalized initials per word.
5. Take the top 20 letters; store the alphabet in `profiles/*.toml` only. The Spanish alphabet does **not** include `ñ`.

**Alternatives considered:** Excluding rare letters manually (less reproducible); full 26-letter alphabet (worse phrase coverage).

### 3. Word lists: curated and embedded

**Choice:** Curated lists live in `internal/worddata/` and are embedded with `go:embed` at build time. No first-run download. Both languages ship in every binary (~106 KB of text; ~3 MB total binary on Linux amd64).

**Curation:**

| Language | Source | Rules |
|----------|--------|-------|
| English | Top 20k frequent words per POS | Lemma dedup (`state`/`states`), length 3–14 |
| Spanish | `doozan/spanish_data` lemmas | Frequency threshold, no diminutives |

Up to **120** words per `(letter, part of speech)` file.

**Layout:**

```text
internal/worddata/
├── en/{letter}/nouns.tsv, adjectives.tsv, verbs.tsv
└── es/{letter}/...
internal/profile/
├── en.toml
└── es.toml
```

**Distribution:** `prepare-wordlists` regenerates `internal/worddata/`. Rebuild `tale` to pick up changes. Optional `dist/data-*.zip` for mirrors only.

**Alternatives considered:** Download on first run (rejected — user preference); larger uncompressed embed (~106 KB — negligible).

### 4. Block generation with rejection

**Choice:** Generate passwords block-by-block (or regenerate individual blocks) until every 4-letter block has at least one valid phrase candidate.

**Rationale:** Avoids dead ends where `lcds` has no noun+adj+verb+noun match. Rejection cost is low with a 20-letter alphabet sized for coverage.

**Alternatives considered:** Generate full 20-char string then fail (bad UX); backtracking search (unnecessary complexity for v1).

### 5. Memorability scoring (v1 heuristic)

**Choice:** Deterministic score; highest wins in random mode.

| Factor | Effect |
|--------|--------|
| Word frequency rank in list | Prefer common words |
| Average word length 4–8 chars | Prefer readable length |
| Repeated lemma/root in block | Penalize |
| Alliteration within block | Slight penalty |

**Alternatives considered:** Random tie-break only (less memorable); ML model (out of scope).

### 6. Normalization for recovery and dictionary indexing

**Choice:** A word's effective initial letter is computed after:

1. Lowercase
2. Strip diacritics on the first character (`á→a`, `é→e`, `í→i`, `ó→o`, `ú→u`, `ü→u`)
3. Leading `h` is **not** silent; `hierro` → `h`

**Spanish-specific dictionary rules:**

- Words starting with an accented vowel are indexed under the base vowel (`árbol` → `a`, `época` → `e`).
- Words starting with `ñ` are **not** included in Spanish word lists; the password alphabet does not use `ñ`.
- Words may retain `ñ` elsewhere in the spelling (e.g. `niño`); only the normalized **first** character matters for matching.

Document these rules in CLI output.

### 7. CLI + TUI structure

**Choice:** Binary name `tale`.

```
tale stats
tale generate [--language es] [--length 20]
tale tui      [--language en] [--password <existing>]
```

- `stats`: show embedded dictionary statistics.
- `generate`: full random flow.
- `tui`: interactive word selection per block.

**Library:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI.

### 8. Repository layout

```
tales_vault_system/
├── cmd/tale/main.go
├── internal/
│   ├── worddata/      # go:embed en es
│   ├── profile/       # go:embed en.toml es.toml
│   ├── password/
│   ├── dictionary/
│   ├── phrase/
│   ├── normalize/
│   └── tui/
├── cmd/prepare-wordlists/
└── openspec/
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Some 4-letter blocks have zero valid phrases | Block-level rejection; measure coverage via `tale stats` |
| Spanish POS tagging quality varies by source | Curated lemma lists only; frequency-ranked |
| TUI adds dependency and complexity | Keep TUI isolated in `internal/tui` |
| Entropy reduced by 20-letter alphabet | Default 20 chars ≈ 86 bits; document in output |
| Dictionary updates require rebuild | prepare-wordlists + `go build`; acceptable for v1 |

## Migration Plan

Not applicable (greenfield). First release is `go build -o tale ./cmd/tale`.

## Open Questions

1. Minimum candidate count per block before accepting a password in random mode (suggest ≥ 3)?
