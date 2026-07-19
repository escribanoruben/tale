## 1. Language new API

- [x] 1.1 Add `InitOptions`/API for `NewUserLanguage` with required alphabet, display name, and force flags
- [x] 1.2 Implement scaffold of `~/.tale/data/{lang}/{letter}/{nouns,adjectives,verbs}.tsv` stubs for alphabet letters only
- [x] 1.3 Add unit tests for custom language creation, alphabet validation, and scaffold layout

## 2. CLI language commands

- [x] 2.1 Add `tale language` subcommand group (`list`, `show`, `init`, `new`) wiring existing + new APIs
- [x] 2.2 Implement `language new <code> --alphabet=...` with clear error when alphabet is missing
- [x] 2.3 Keep `tale profile` as deprecated alias that forwards to `language` with a one-line hint
- [x] 2.4 Print next-steps after `new` (edit TSV files) and indicate user override on `generate`/`show`

## 3. Docs and Makefile

- [x] 3.1 Update README quick start to `--language en|es` and document `language new` / override semantics
- [x] 3.2 Fix Makefile `ARGS` default to `tui` (language from settings/locale, not Makefile)
- [x] 3.3 Update CLI help text that still mentions profiles as user-facing selectors

## 4. OpenSpec coherence

- [x] 4.1 Update `mvp-password-mnemonic-generator` specs (`language-profiles`, `cli-tui`, `embedded-word-data`) to language-first wording so both open changes agree
- [x] 4.2 Update MVP `design.md` / `proposal.md` references to `--profile` where they contradict runtime

## 5. Verification

- [x] 5.1 Add/adjust CLI or package tests for `language new` and override messaging
- [x] 5.2 Run `go test ./...` and smoke `go run ./cmd/tale language list`
