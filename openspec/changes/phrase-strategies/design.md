## Context

The semantic engine ships one mnemonic shape (chained story phrases). The user wants the shape to be a profile choice: the proven story at different lengths, plus noun-only modes where the password is spelled by X nouns sharing a semantic field or forming an association path. All strategies share the same security model: password letters are drawn by CSPRNG before any word selection.

## Decisions

- **Config surface**: one new profile key, `strategy` (`chained_story` default). No new length keys: `password_length` remains the letter count; noun strategies consume it directly (one noun per letter), chained_story derives phrases as today. 16-letter/4-phrase and 20-letter/5-phrase are the same strategy with different lengths.
- **Noun strategies use only nouns** (no adjectives/verbs): per-letter candidate lists come from the noun buckets, filtered by embedded-vector presence and excluding the low-imageability set (abstract nouns would wreck field coherence).
- **`noun_field` = adaptive centroid**: seed noun picked by CSPRNG among the first letter's candidates; 3 rounds of per-letter argmax + centroid re-normalization. Deterministic given the seed; the seed is the only CSPRNG input (entropy unaffected). Word/root dedupe during selection. Chosen metric from the experiment: mean pairwise cosine 0.353 vs 0.155 random, stutters 0.10/story.
- **`noun_path` = k-best Viterbi over consecutive positions** with edge weight `cos − 2.0·max(0, cos − 0.75)` (stutter guard: the weak −0.5 guard still produced 3.6 synonym hops/story; the 2.0 guard brings it to 0.93). Word repetition is not Markovian, so it is enforced by a mask-and-retry loop: re-run the DP with nodes containing repeated words masked (max 4 retries), falling back to a greedy deduplicated pass if the mask loop saturates. Measured: consecutive cosine 0.562 vs 0.154 random, ~90 ms.
- **Field is the recommended default for `noun_*`**: uniform coherence ("one room of the memory palace", e.g. the grocery list: tomate, naranja, queso, carne…) with zero stutters and no DP cost; `noun_path` is the narrative alternative.
- **Failure mode**: noun strategies REQUIRE embedded vectors — a custom language without them gets an explicit error (not a silent fallback, which would produce unoptimized words with no warning).
- **TUI**: noun strategies land CLI-only this slice; the TUI keeps its current story/random flows until the cycling slice.

## Risks / Trade-offs

- Field tightness depends on the letters: random 16-letter sequences still reach ~0.35 mean pairwise cosine with the adaptive centroid, but very "wide" letter mixes may drift (mitigation: the centroid re-centering is the mechanism; accept the residual).
- Greedy fallback in `noun_path` sacrifices global optimality when the mask loop saturates; rare in practice (repeats need degenerate letter mixes).
- `phrase_template` is ignored by noun strategies but still validated (block-size consistency) — documented, not a trap: noun strategies keep the standard block grouping only for output formatting.
