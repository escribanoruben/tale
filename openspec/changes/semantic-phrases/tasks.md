## 1. Word-list curation (prerequisite)

- [ ] 1.1 EN `nouns.tsv`: remove proper nouns and demonyms flagged by the experiment (norway, ben, andrew, canadian, september, june, vienna, ukrainian, brazil, japan, kenya, thomas, …); script-assisted scan + manual pass
- [ ] 1.2 EN `verbs.tsv`: remove auxiliaries and defective forms (will, should, may, might, could, don, …); verify every letter keeps ≥ 1 verb after removal
- [ ] 1.3 ES spot-check: remove filler verbs (ser, haber) and particles (porqué) where they were placed in the wrong POS list; confirm no coverage hole
- [x] 1.4 Suitability pass: extend `internal/wordprepare/blocklist/{en,es}.txt` (sexual aggression, gore, morbidity, hate, weapons, disease dread, disgust) and purge the committed TSVs of the same words (151 occurrences); every letter × POS bucket keeps ≥ 8 words
- [ ] 1.5 Regenerate embedded word data and confirm tests still pass

## 2. Embeddings asset and pipeline

- [ ] 2.1 `cmd/prepare-embeddings`: maintainer-only, networked; downloads `cc.{lang}.300.vec.gz`, extracts vectors for embedded vocabulary, writes `data/{lang}/embeddings.bin` (magic + version + dim + word list + int8 quantization with per-vector scale)
- [ ] 2.2 Loader in `internal/worddata`: parse embeddings.bin, expose word → vector and dot/cosine helpers; fail soft (empty index) when asset absent
- [ ] 2.3 `make embeddings` target; document regeneration and CC-BY-SA 3.0 attribution in NOTICE and `data/README.md`
- [ ] 2.4 Generate committed assets for `en` and `es`

## 3. Semantic scoring and selection

- [ ] 3.1 `internal/semantic`: cosine over int8 vectors, anchored score `cos(n1,v) + cos(v,n2) + cos(adj,n1)`, distance band (τ_hi 0.45, τ_lo 0.05, weights 1.0/0.3) on `cos(n1,n2)`
- [ ] 3.2 Exact top-K per block (K = 10) with diversity demotion; deterministic; respects `used` word set
- [ ] 3.3 Selection mode wiring: enabled by default for builtin languages with vectors; settings/profile key `semantic = false` disables; silent fallback to current scoring for languages without vectors
- [ ] 3.4 Template constants: slot roles (n1, verb, n2, adj) derived from `PhraseTemplate` so scoring follows the profile

## 4. Template change (builtin languages)

- [ ] 4.1 Switch builtin `es` and `en` `phrase_template` to `noun, verb, noun, adjective` in profiles
- [ ] 4.2 Coverage test: every alphabet letter × every template POS has ≥ 1 word after the swap and after curation
- [ ] 4.3 Update `builtin_templates_test`, README examples, and `language-profiles` spec wording

## 5. TUI candidate cycling

- [ ] 5.1 Picker shows rank 1 by default with a position indicator ("3/10"); Tab / Shift+Tab (or R) cycles the pool per block
- [ ] 5.2 Chosen candidate marks its words used for cross-block constraints; cycling never triggers regeneration of the password
- [ ] 5.3 CLI `generate` prints the top-1 phrase (no cycling) and hints that the TUI offers alternatives

## 6. Docs and verification

- [ ] 6.1 README: semantic mode, template change, regeneration command; CONTRIBUTING if release steps affected (no)
- [ ] 6.2 Spec deltas reviewed against implementation (`semantic-mnemonics`, `cli-tui`, `language-profiles`, `embedded-word-data`)
- [ ] 6.3 Determinism test: same password → same pool twice; fallback test: language without vectors → current scoring
- [ ] 6.4 `go test ./...`, `make build`, manual TUI smoke test (es + en)
- [ ] 6.5 CI green on GitHub after merge (push mirror)
