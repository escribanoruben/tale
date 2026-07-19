## 1. Project Setup

- [x] 1.1 Initialize Go module and repository layout per design.md
- [x] 1.2 Add `cmd/tvs/main.go` with `stats` subcommand
- [ ] 1.3 Add `generate` and `tui` subcommand skeletons

## 2. Word List Data

- [x] 2.1 Create word data README with sources and licenses
- [x] 2.2 Implement `scripts/prepare-wordlists` for EN noun/adjective/verb lists
- [x] 2.3 Extend prepare script for ES lists (exclude `ñ`-initial; `árbol` → `a`)
- [x] 2.4 Output letter-based layout under `internal/worddata/`
- [x] 2.5 Curate lists (lemma dedup EN, frequency lemmas ES)

## 3. Embedded Data

- [x] 3.1 `go:embed` word lists in `internal/worddata/`
- [x] 3.2 `go:embed` profiles in `internal/profile/`
- [x] 3.3 Implement `tvs stats` with table and `--json` output
- [ ] 3.4 Implement `internal/dictionary` loader from embedded FS

## 4. Core Libraries

- [ ] 4.1 Implement `internal/normalize` (lowercase, strip accents on first character, initial extraction)
- [ ] 4.2 Implement `internal/profile` TOML parser from embedded FS
- [ ] 4.3 Wire dictionary lookup for `{lang}/{letter}/{pos}.tsv`
- [ ] 4.4 Implement `internal/phrase` candidate search for noun-adjective-verb-noun blocks
- [ ] 4.5 Implement `internal/phrase` memorability scorer (deterministic)
- [ ] 4.6 Implement `internal/password` CSPRNG generation with block-level rejection

## 5. Generate Command

- [ ] 5.1 Wire `tvs generate`: load profile → generate password → build phrases → print output
- [ ] 5.2 Add entropy calculation and recovery instructions to output
- [ ] 5.3 Add `--profile` and `--length` flags with sensible defaults

## 6. TUI Mode

- [ ] 6.1 Add Bubble Tea dependency and `internal/tui` package
- [ ] 6.2 Implement block-by-block navigation with per-slot word lists
- [ ] 6.3 Wire `tvs tui` with optional `--password` flag
- [ ] 6.4 Print final password + selected phrases on TUI completion

## 7. Testing and Release

- [ ] 7.1 Unit tests: normalization, block splitting, initial-letter matching
- [ ] 7.2 Unit tests: phrase recovery round-trip (phrase initials → password block)
- [ ] 7.3 Integration test: `generate` with both default profiles
- [x] 7.4 Document build/run instructions in project README
- [ ] 7.5 Release binary builds for Linux/macOS/Windows
