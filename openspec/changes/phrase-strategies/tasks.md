## 1. Profile configuration

- [x] 1.1 `Strategy` field in `profile.Profile` (`toml:"strategy"`) + validation: empty or one of chained_story | noun_field | noun_path
- [x] 1.2 Validation test for the strategy key

## 2. Noun engine (`internal/semantic/nouns.go`)

- [x] 2.1 `NounField`: adaptive centroid (CSPRNG seed, 3 rounds), word+root dedupe
- [x] 2.2 `NounPath`: k-best Viterbi with stutter guard (0.75 / 2.0) + mask-and-retry word dedupe (greedy fallback on original candidates)
- [x] 2.3 Engine tests with crafted vectors: field coherence (one axis), path dedupe, stutter guard, exhausted-candidates error

## 3. Wiring

- [x] 3.1 `phrase.GenerateNouns` / `phrase.NounsForPassword`: CSPRNG letters + per-letter candidates (vector-present, non-low-imageability) + strategy dispatch; clear error without vectors
- [x] 3.2 `cmd/tale generate`: branch on strategy, noun-specific output and recovery text; `LoadEmbeddings` also reads user-data `embeddings.bin` (custom languages can ship vectors)
- [x] 3.3 Settings/TUI: noun strategies are CLI-generate for now (documented; TUI cycling slice will cover them)

## 4. Docs and verification

- [x] 4.1 README/profile docs for `strategy`
- [x] 4.2 `go test ./...`, manual smoke (es noun_field + noun_path, 16 letters)
- [ ] 4.3 CI green on GitHub after merge

## 5. Strategy names, passtale and LLM polish

- [x] 5.1 Strategy values renamed: `story` (alias `chained_story`), `random` (uniform words, no embeddings), noun strategies unchanged
- [x] 5.2 Trío template shape supported: [noun verb noun adj] maps the trailing adjective to the first noun (roles fix)
- [x] 5.3 `internal/llm`: polish pipeline (render candidates with retries, discard broken, pick most natural), carrier-stem validators, graceful fallback; local ollama only
- [x] 5.4 Settings: `llm_polish`, `llm_model`, `llm_url`, `llm_threads`; `LoadEmbeddings` also reads user-data `embeddings.bin`
- [x] 5.5 CLI passtale output: fluent sentence + carrier recap per block; `phrase.GenerateStoryBlocks` exported
- [ ] 5.6 CI green on GitHub after merge
