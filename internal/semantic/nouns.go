package semantic

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/escribanoruben/tale/internal/normalize"
	"github.com/escribanoruben/tale/internal/worddata"
)

// Noun-only strategies: the password is spelled by the initials of X
// nouns (one per password letter). Candidates arrive pre-filtered per
// position (embedded-vector present, not low-imageability).

// Noun strategy tuning, from the noun-strategy experiment
// (openspec/changes/phrase-strategies/design.md).
const (
	// PathStutterHigh: consecutive pairs closer than this are
	// near-synonyms and get penalized.
	PathStutterHigh = 0.75
	// PathStutterPenalty: strong enough to bring synonym hops from 3.6
	// to ~1 per story in the experiment.
	PathStutterPenalty = 2.0
	// fieldRounds: centroid re-centering rounds.
	fieldRounds = 3
)

// ErrNoCandidates reports a position without usable noun candidates.
var ErrNoCandidates = errors.New("semantic: no noun candidates for a password letter")

// NounField selects one noun per position orbiting an adaptive centroid:
// a seed noun (CSPRNG among the first position's candidates) is refined
// over re-centering rounds, so the chosen words share a semantic field —
// one mental scene. No repeated words or roots.
func NounField(e *worddata.Embeddings, cands [][]string) ([]string, error) {
	for i, c := range cands {
		if len(c) == 0 {
			return nil, fmt.Errorf("%w: position %d", ErrNoCandidates, i)
		}
	}
	n := big.NewInt(int64(len(cands[0])))
	idx, err := rand.Int(rand.Reader, n)
	if err != nil {
		return nil, err
	}
	seed := cands[0][idx.Int64()]
	centroid := append([]float32(nil), e.Vector(seed)...)
	var chosen []string
	for range fieldRounds {
		chosen = nil
		roots := map[string]bool{}
		sum := make([]float64, len(centroid))
		for _, c := range cands {
			bw, bv := "", math.Inf(-1)
			for _, w := range c {
				r := rootOf(w)
				if roots[r] {
					continue
				}
				v := e.Vector(w)
				if v == nil {
					continue
				}
				var d float64
				for k := range centroid {
					d += float64(v[k] * centroid[k])
				}
				if d > bv {
					bv, bw = d, w
				}
			}
			if bw == "" {
				return nil, fmt.Errorf("%w: all remaining candidates share roots", ErrNoCandidates)
			}
			chosen = append(chosen, bw)
			roots[rootOf(bw)] = true
			v := e.Vector(bw)
			for k := range v {
				sum[k] += float64(v[k])
			}
		}
		var norm float64
		for k := range sum {
			norm += sum[k] * sum[k]
		}
		norm = math.Sqrt(norm)
		if norm == 0 {
			norm = 1
		}
		for k := range centroid {
			centroid[k] = float32(sum[k] / norm)
		}
	}
	return chosen, nil
}

// NounPath selects one noun per position maximizing consecutive-adjacency
// cosine (exact k-best Viterbi with a stutter guard on near-synonym
// hops), then enforces word uniqueness by masking candidates containing
// repeated words and re-running the DP (greedy deduplicated pass as a
// last resort).
func NounPath(e *worddata.Embeddings, cands [][]string) ([]string, error) {
	viterbi := func() ([]string, bool) {
		n := len(cands)
		best := make([][]float64, n)
		prev := make([][]int, n)
		for range cands[0] {
			best[0] = append(best[0], 0)
			prev[0] = append(prev[0], -1)
		}
		edge := func(a, b string) float64 {
			c := cosine(e, a, b)
			if c > PathStutterHigh {
				c -= PathStutterPenalty * (c - PathStutterHigh)
			}
			return c
		}
		for i := 1; i < n; i++ {
			for _, w := range cands[i] {
				bi, bv := -1, math.Inf(-1)
				for a, wp := range cands[i-1] {
					v := best[i-1][a] + edge(wp, w)
					if v > bv {
						bv, bi = v, a
					}
				}
				if bi < 0 {
					return nil, false
				}
				best[i] = append(best[i], bv)
				prev[i] = append(prev[i], bi)
			}
		}
		bi, bv := -1, math.Inf(-1)
		for a, v := range best[n-1] {
			if v > bv {
				bv, bi = v, a
			}
		}
		if bi < 0 {
			return nil, false
		}
		idx := make([]int, n)
		idx[n-1] = bi
		for i := n - 1; i > 0; i-- {
			idx[i-1] = prev[i][idx[i]]
		}
		ws := make([]string, n)
		for i := range ws {
			ws[i] = cands[i][idx[i]]
		}
		return ws, true
	}

	masked := map[string]bool{}
	original := cands
	maskedCands := append([][]string(nil), cands...)
	for attempt := 0; attempt < 4; attempt++ {
		ws, ok := viterbi()
		if !ok {
			break
		}
		repeated := repeatedWords(ws)
		if len(repeated) == 0 {
			return ws, nil
		}
		for w := range repeated {
			masked[w] = true
			for i := range maskedCands {
				filtered := maskedCands[i][:0:0]
				for _, c := range maskedCands[i] {
					if !masked[normalize.Key(c)] {
						filtered = append(filtered, c)
					}
				}
				maskedCands[i] = filtered
			}
		}
	}
	// greedy deduplicated fallback (original candidates)
	return greedyPath(e, original)
}

// repeatedWords returns words (normalized keys) appearing more than once.
func repeatedWords(ws []string) map[string]bool {
	count := map[string]int{}
	for _, w := range ws {
		count[normalize.Key(w)]++
	}
	out := map[string]bool{}
	for w, n := range count {
		if n > 1 {
			out[w] = true
		}
	}
	return out
}

// greedyPath picks per position the candidate closest to the previous
// chosen word, skipping used words. Last-resort path builder.
func greedyPath(e *worddata.Embeddings, cands [][]string) ([]string, error) {
	var out []string
	used := map[string]bool{}
	var prev string
	for i, c := range cands {
		if len(c) == 0 {
			return nil, fmt.Errorf("%w: position %d exhausted by dedupe", ErrNoCandidates, i)
		}
		bw, bv := "", math.Inf(-1)
		for _, w := range c {
			if used[normalize.Key(w)] {
				continue
			}
			var d float64
			if i > 0 {
				d = cosine(e, prev, w)
			} else {
				d = 1 // first position: any candidate is fine, pick first
			}
			if d > bv || bw == "" {
				bv, bw = d, w
			}
		}
		if bw == "" {
			return nil, fmt.Errorf("%w: position %d exhausted by dedupe", ErrNoCandidates, i)
		}
		out = append(out, bw)
		used[normalize.Key(bw)] = true
		prev = bw
	}
	return out, nil
}

func cosine(e *worddata.Embeddings, a, b string) float64 {
	va, vb := e.Vector(a), e.Vector(b)
	if va == nil || vb == nil {
		return 0
	}
	var s float64
	for i := range va {
		s += float64(va[i]) * float64(vb[i])
	}
	return s
}
