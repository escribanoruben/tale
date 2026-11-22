## Context

`tale` derives mnemonic phrases (whose initials spell the password blocks) from curated word lists: ~120 words per letter per POS, ~5,700-6,900 unique words per language. Selection today is exhaustive search under a frequency/length heuristic, which favors word soup.

Password entropy is irrelevant to word choice: block letters are drawn by CSPRNG *before* words are selected, and the initials of all valid phrases cover exactly the letter space — so any selection strategy leaves the password search space untouched. The only quality axis is memorability.

## Experiment results (offline, fastText cc.{en,es}.300 vectors restricted to tale's vocabulary)

Scripts ran against the real embedded lists (profiles `es.toml` / `en.toml`, templates read from config). Coverage of the vocabulary in fastText: **99.7 %** both languages (5,731/5,747 ES; 6,848/6,866 EN). Missing words are rare terms; selection skips them.

Anchored score = mean of `cos(n1,adj) + cos(n1,v) + cos(v,n2)` (subject noun, verb, object noun; adjective attaches to n1 only).

| Strategy | ES mean | EN mean |
|---|---|---|
| Uniform random (current `Random`) | 0.137 | 0.070 |
| Greedy A (fix n1+adj pair → verb → n2) | 0.441 | 0.491 |
| Greedy B (fix n1+verb → adj → n2) | 0.428 | 0.459 |
| **Exact global optimum** | **0.468** | **0.516** |

Key findings:

- **Exact beats greedy and is cheaper than it looks**: because the adjective depends only on n1 and the object noun only on the verb, the global argmax decomposes into 3 small (120×120) cosine matrices plus an argmax — milliseconds even in pure Go. Greedy search is unnecessary.
- **Distance band on `cos(n1,n2)`** (penalize pairs above τ_hi = 0.45, gently push up from below τ_lo = 0.05; weights 1.0 / 0.3): cuts near-synonym mush in ES top-1s from 4/25 blocks to 1/25 (max pair cosine 0.684 → 0.457) at a score cost of −0.004. Top-1 pairs already sit near the sweet spot (~0.3); the band is a guardrail.
- **Template `noun verb noun adjective`** (trío as the core, adjective as a trailing tag of n1): produces clearly more vivid scenes in ES ("asesino cazar guarida homicida", "paciente medicar fármaco médico") at identical pairwise score (0.461 vs 0.465). The template swap changes which letter feeds which POS pool, which is why it was measured, not assumed.
- **EN bottleneck is data, not model**: top candidates include proper nouns ("norway ben andrew canadian"), auxiliaries as verbs ("will should consideration take"), ordinal noise ("third including addition fourth"). Curation of the EN TSVs is a prerequisite for the EN template switch to read well.

## Decisions

- **Source vectors**: fastText Common Crawl `cc.{en,es}.300.vec.gz` (license CC-BY-SA 3.0 → attribution in NOTICE and `data/README.md`). Multilingual family keeps EN/ES vector spaces in one recipe; regeneration is a maintainer-only, networked step like `make wordlists`.
- **Storage format**: custom compact binary per language (`data/{lang}/embeddings.bin`): magic + version + dim + word list + int8-quantized matrix (per-vector max-abs scale). ~1.7 MB/language raw, ~1 MB compressed; both languages ≈ +2-3 MB embedded in the binary. No ANN index needed at ~7k vectors — brute-force dot products.
- **Scoring**: anchored score + distance band as above; all constants in one place (`internal/semantic`), surfaced as documented constants, not config, for v1.
- **Imageability penalty**: a curated per-language list (`imageability.txt`: 118 abstract nouns + 22 stative verbs for ES, 115 + 15 for EN) costs −0.25 per hit on the anchored mean, so concrete scenes outrank abstract soup without banning the words (they remain as last resort). Measured offline: ES top-1 phrases containing abstract/stative words drop from 23 % to 3 % of blocks. Future refinement: per-word concreteness norms (Brysbaert CC-BY for EN; license check needed for ES) embedded next to the vectors.
- **Role hierarchy (nouns → verbs → adjectives)**: the noun spine dominates at story level (chain/closure edges are noun-only, weight 2.0); within a block the anchored edges are weighted — verb edges (the action) at 1.0, adjective edge (the decoration) at 0.5, normalized so the score stays cosine-scaled. The adjective argmax per subject is unchanged (a global rescale cannot flip it), so embedding-driven gender agreement is unaffected.
- **Gender agreement by embedding (no grammar code)**: Spanish adjective lists carry derived feminine forms as first-class entries (mechanical derivation in `prepare-wordlists` and in the committed data: `-o→-a` plus `-dor/-ón/-án/-ete/-ote/-ín→+a`; invariables skipped). fastText encodes the gender signal, so the anchored `cos(noun1, adjective)` edge picks the agreeing form automatically — verified offline on 15/15 noun/adjective pairs (including the `problema`/`día` masculine-in-`-a` and `foto` feminine-in-`-o` exceptions, which need no exception lists). The swap preserves the first letter, so mnemonic initials and entropy are unaffected.
- **Selection**: exact top-K per block (K = 10 default). Deterministic: same password → same pool (no RNG in ranking). Diversity: candidates sharing words **or roots** with higher-ranked ones are demoted, so the pool offers 10 distinct scenes. Cross-block word reuse (`used` set) is respected exactly as today, plus cross-block root reuse (`usedRoots`: "gobierno"/"gobernar" feel like repeats in a story).
- **Whole-story optimization (layered DAG + k-best DP)**: instead of freezing each phrase greedily, per-block candidate pools (K = 100, soft diversity: candidates may share ≤ 1 word, so the same noun appears in several scene variants; 10 verb/object combos per subject) form a layered graph; chain edges score the subject of block i+1 against the previous phrase's nouns (weight 2.0) and a closure edge ties the final object noun back to the first subject. A single-pass k-best DP (O(K²), per-node path budget 64) returns the best complete story plus ranked alternatives; the closure bonus is applied on reconstruction and re-ranked with the top-2000 backtracked paths. Measured (15 passwords ES, wide vs narrow graphs): per-scene quality flat, chain +, closure +55 %, abstract words 2 → 0, runtime ~0.3 s. A global topic anchor was rejected (topic lottery: a bad anchor degrades every phrase).
- **Refrains allowed**: cross-block word repetition is not penalized — a repeated word reads as a refrain and can aid recall (user decision); only within-phrase root repeats remain rejected.
- **TUI**: default shows rank 1; Tab / Shift+Tab (or R) cycles the pool per block with a position indicator ("3/10"). The user is the final judge — covers residual garbage and uncomfortable picks without an ML answer.
- **Template change**: builtin `es` and `en` switch to `["noun","verb","noun","adjective"]`, applied **after** list curation for EN. BREAKING for builtin mnemonic format; validated by a coverage test (every letter × POS non-empty).
- **Fallback**: if a language has no embedded vectors (custom/user languages), semantic mode is silently disabled and current scoring applies. No error, no behavioral cliff.
- **Mode default**: semantic ranking is enabled by default for builtin languages with vectors; a settings/profile key (`semantic = false`) disables it.
- **Runtime stays offline**: embeddings ship in the binary; no network in generate/TUI paths (AGENTS.md constraint preserved).

## Risks / Trade-offs

- **Binary size** +2-3 MB for both languages — accepted; alternative (dropping dims to 128) loses little ranking quality but complicates the pipeline for marginal savings; revisit if size matters.
- **Uncomfortable words** surface more often under semantic ranking (frequency-blind): "violador fornicar sexo homicida". Handled in scope: the suitability curation purges sexual aggression, gore, morbidity, hate, weapons and disease-dread vocabulary from the blocklists and committed lists (phrases are meant to be memorized — no unpleasant imagery). The TUI cycling remains the final guard.
- **Curation scope creep**: the EN lists carried more damage than the experiment first showed — proper nouns (164 in a first sweep, then a second full pass: persons, places, institutions, demonyms), auxiliaries and verb inflections (`-ed`/`-ing`), function words in the wrong POS (`would`, `kinda`, `that`, `something`, …), acronym fragments (`nasa`, `gdp`, `vii`), and slurs (`negro`). Systematic category sweeps plus blocklist entries fixed them; every letter × POS bucket keeps ≥ 8 words (EN min 35, ES min 8).
- **fastText drift**: regeneration with newer vectors may shift rankings; determinism is per-binary (vectors are embedded), so regeneration is an explicit release decision.
