// Package semantic ranks mnemonic phrases by embedding proximity.
//
// Phrases are scored with an anchored structure derived from the profile
// template: the subject noun links to the adjective and to the verb, the
// verb links to the object noun, and the two nouns must be related but
// distinct ("prudentially far"): a distance band rejects near-synonym
// pairs while gently discouraging unrelated ones.
//
//	score = cos(noun1, adj) + cos(noun1, verb) + cos(verb, noun2) + band(cos(noun1, noun2))
//
// Ranking is exact and deterministic: same vectors and same inputs yield
// the same candidate pool. Password entropy is unaffected — block letters
// are drawn before any word is chosen.
package semantic

import (
	"sort"

	"github.com/escribanoruben/tale/internal/normalize"
	"github.com/escribanoruben/tale/internal/worddata"
)

// Slot roles within the template (indexes into the slot lists).
const (
	Subject = iota // noun1: adjective attaches to it
	Verb
	Object // noun2
	Adjective
)

// Band tuning for cos(noun1, noun2), from the offline experiment
// (openspec/changes/semantic-phrases/design.md): mush rejection at
// negligible score cost.
const (
	BandHigh       = 0.45 // reject noun pairs closer than this (near-synonyms)
	BandLow        = 0.05 // gently pull up pairs looser than this
	BandHighWeight = 1.0
	BandLowWeight  = 0.3
	// adjective band: same synonym-trap guard as the noun band, with a
	// softer threshold ("luna plateada" survives, "silencio callado" falls)
	AdjBandHigh   = 0.60
	AdjBandWeight = 0.5
	// topCombosPerSubject: verb/object combos kept per subject for pool
	// diversity (experiment sweet spot: 10).
	topCombosPerSubject = 10
)

// AdjBand scores the noun1-adjective edge: penalty only for
// near-synonym pairs above AdjBandHigh.
func AdjBand(cos float64) float64 {
	if cos > AdjBandHigh {
		return -AdjBandWeight * (cos - AdjBandHigh)
	}
	return 0
}

// Candidate is one ranked phrase.
type Candidate struct {
	Words []string // ordered by template slot position
	Score float64  // mean anchored cosine (band and link included)
}

// RankOptions tunes one ranking pass over a block.
type RankOptions struct {
	// Used holds excluded full-word keys (normalize.Key); ignored with AllowRepeat.
	Used map[string]bool
	// UsedRoots holds excluded word-root keys across blocks (nil to skip).
	UsedRoots map[string]bool
	// AllowRepeat disables exclusion checks.
	AllowRepeat bool
	// K is the number of diverse candidates to return.
	K int
	// LowImageability marks words that are hard to picture (abstract
	// nouns, stative verbs); each hit costs ImageabilityPenalty.
	LowImageability map[string]struct{}
	// MaxSharedWords relaxes the within-pool diversity: a candidate is
	// skipped only when it shares more than this many words with an
	// already-picked one (0 = current disjoint behavior; 1 lets the
	// same noun appear with different verb/object scenes).
	MaxSharedWords int
	// CombosPerSubject overrides topCombosPerSubject when > 0.
	CombosPerSubject int
	// AdjWeight / VerbWeight scale the anchored edges (nouns -> verbs ->
	// adjectives priority). 0 keeps the defaults.
	AdjWeight  float64
	VerbWeight float64
}

// DefaultLinkWeight pairs with the offline experiment: strong enough to
// steer the subject choice, weak enough to keep scene quality.
const DefaultLinkWeight = 2.0

// ImageabilityPenalty is the score cost per low-imageability word
// (abstract noun, stative verb): from the offline experiment, cutting
// abstract top-1 phrases from 23% to 3% of blocks at negligible quality
// cost.
const ImageabilityPenalty = 0.25

// Role hierarchy (experiment "nouns -> verbs -> adjectives"): the noun
// spine already dominates at story level (chain/closure edges are
// noun-only, weight 2.0); within a block the verb edges (the action)
// weigh more than the adjective edge (the decoration).
const (
	DefaultAdjWeight  = 0.5
	DefaultVerbWeight = 1.0
)

// Band scores a noun-pair cosine: 0 inside the sweet spot, penalties outside.
func Band(cos float64) float64 {
	if cos > BandHigh {
		return -BandHighWeight * (cos - BandHigh)
	}
	if cos < BandLow {
		return -BandLowWeight * (BandLow - cos)
	}
	return 0
}

// Rank returns the top-k word-distinct candidates for one block.
//
// lists holds one candidate word list per template slot; roles maps
// semantic roles to slot indexes. Slot words without vectors are dropped.
// Candidates in the pool share no words or roots (diversity), so k
// options are k distinct scenes.
func Rank(e *worddata.Embeddings, lists [][]string, roles [4]int, ro RankOptions) []Candidate {
	if ro.K <= 0 {
		ro.K = 1
	}
	words := make([][]string, 4)
	vecs := make([][]([]float32), 4)
	for r := 0; r < 4; r++ {
		for _, w := range lists[roles[r]] {
			key := normalize.Key(w)
			if !ro.AllowRepeat {
				if ro.Used[key] || (ro.UsedRoots != nil && ro.UsedRoots[key]) {
					continue
				}
			}
			v := e.Vector(w)
			if v == nil {
				continue
			}
			words[r] = append(words[r], w)
			vecs[r] = append(vecs[r], v)
		}
		if len(words[r]) == 0 {
			return nil
		}
	}
	sub, verb, obj, adj := vecs[Subject], vecs[Verb], vecs[Object], vecs[Adjective]

	dim := e.Dim
	dot := func(a, b []float32) float64 {
		var s float64
		for i := 0; i < dim; i++ {
			s += float64(a[i]) * float64(b[i])
		}
		return s
	}

	// Cosine matrices: subject x adjective, subject x verb, verb x object.
	mAdj := make([][]float64, len(sub))
	mSVerb := make([][]float64, len(sub))
	for i := range sub {
		mAdj[i] = make([]float64, len(adj))
		for j := range adj {
			mAdj[i][j] = dot(sub[i], adj[j])
		}
		mSVerb[i] = make([]float64, len(verb))
		for j := range verb {
			mSVerb[i][j] = dot(sub[i], verb[j])
		}
	}
	mVObj := make([][]float64, len(verb))
	for j := range verb {
		mVObj[j] = make([]float64, len(obj))
		for o := range obj {
			mVObj[j][o] = dot(verb[j], obj[o])
		}
	}

	type raw struct {
		score             float64
		sub, adj, vb, obj int
	}
	var raws []raw
	for i := range sub {
		bestAdj := 0
		for j := range adj {
			if mAdj[i][j] > mAdj[i][bestAdj] {
				bestAdj = j
			}
		}
		// subject-object cosines, computed once per subject
		nn := make([]float64, len(obj))
		for o := range obj {
			nn[o] = dot(sub[i], obj[o])
		}
		// verb/object table for this subject; the band couples
		// subject-object.
		d := make([]float64, len(verb)*len(obj))
		for j := range verb {
			for o := range obj {
				d[j*len(obj)+o] = mSVerb[i][j] + mVObj[j][o] + Band(nn[o])
			}
		}
		// keep the top verb/object combos per subject by linear scan
		// (deterministic: earlier index wins ties) — sorting the whole
		// table would dominate runtime
		combos := ro.CombosPerSubject
		if combos <= 0 {
			combos = topCombosPerSubject
		}
		for _, t := range topN(d, combos) {
			j, o := t/len(obj), t%len(obj)
			wa, wv := ro.AdjWeight, ro.VerbWeight
			if wa <= 0 {
				wa = DefaultAdjWeight
			}
			if wv <= 0 {
				wv = DefaultVerbWeight
			}
			score := (wa*mAdj[i][bestAdj] + wv*(mSVerb[i][j]+mVObj[j][o])) / (wa + 2*wv)
			// synonym guard on the adjective edge
			score += AdjBand(mAdj[i][bestAdj])
			// imageability: concrete scenes beat abstract soup
			if len(ro.LowImageability) > 0 {
				for _, w := range []string{words[Subject][i], words[Verb][j], words[Object][o], words[Adjective][bestAdj]} {
					if _, low := ro.LowImageability[normalize.Key(w)]; low {
						score -= ImageabilityPenalty
					}
				}
			}
			raws = append(raws, raw{score, i, bestAdj, j, o})
		}
	}

	sort.SliceStable(raws, func(a, b int) bool { return raws[a].score > raws[b].score })

	var out []Candidate
	picked := make(map[string]bool)
	pickedRoots := make(map[string]bool)
	var pickedWords [][]string // for soft-diversity overlap checks
	for _, r := range raws {
		if len(out) == ro.K {
			break
		}
		w := [4]string{words[Subject][r.sub], words[Verb][r.vb], words[Object][r.obj], words[Adjective][r.adj]}
		if !ro.AllowRepeat {
			distinct := true
			for a := 0; a < 4 && distinct; a++ {
				if picked[normalize.Key(w[a])] || (ro.MaxSharedWords == 0 && pickedRoots[rootOf(w[a])]) || (ro.UsedRoots != nil && ro.UsedRoots[rootOf(w[a])]) {
					distinct = false
					break
				}
				// mush guard: a candidate never repeats a root within
				// itself ("entrevista...entrevistar"), whatever the
				// pool's diversity mode
				for b := a + 1; b < 4; b++ {
					if w[a] == w[b] || rootOf(w[a]) == rootOf(w[b]) {
						distinct = false
						break
					}
				}
			}
			if distinct && ro.MaxSharedWords > 0 {
				// soft diversity: skip only when a picked candidate
				// shares more than MaxSharedWords with this one
				for _, p := range pickedWords {
					shared := 0
					for _, x := range p {
						if normalize.Key(x) == normalize.Key(w[0]) || normalize.Key(x) == normalize.Key(w[1]) ||
							normalize.Key(x) == normalize.Key(w[2]) || normalize.Key(x) == normalize.Key(w[3]) {
							shared++
						}
					}
					if shared > ro.MaxSharedWords {
						distinct = false
						break
					}
				}
			}
			if !distinct {
				continue
			}
		}
		ordered := make([]string, len(lists))
		for role, slot := range roles {
			ordered[slot] = w[role]
		}
		out = append(out, Candidate{Words: ordered, Score: r.score})
		if !ro.AllowRepeat {
			pw := make([]string, 0, 4)
			for _, x := range w {
				picked[normalize.Key(x)] = true
				pw = append(pw, x)
			}
			pickedWords = append(pickedWords, pw)
			if ro.MaxSharedWords == 0 {
				for _, x := range w {
					pickedRoots[rootOf(x)] = true
				}
			}
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// topN returns the indexes of the n largest values of d, largest first.
// Linear scan with running insertion (deterministic: earlier index wins
// ties); n is small, so a full sort would only dominate runtime.
func topN(d []float64, n int) []int {
	if n > len(d) {
		n = len(d)
	}
	idx := make([]int, 0, n)
	for t, v := range d {
		if len(idx) < n {
			idx = append(idx, t)
			// insertion into the (short) running list
			for i := len(idx) - 1; i > 0 && v > d[idx[i-1]]; i-- {
				idx[i], idx[i-1] = idx[i-1], idx[i]
			}
			continue
		}
		if v <= d[idx[n-1]] {
			continue
		}
		idx[n-1] = t
		for i := n - 1; i > 0 && d[idx[i]] > d[idx[i-1]]; i-- {
			idx[i], idx[i-1] = idx[i-1], idx[i]
		}
	}
	return idx
}

// rootOf is the cross-block root key (first four letters, lowercased):
// words sharing it ("gobierno"/"gobernar") feel like repeats in a story.
func rootOf(w string) string {
	key := normalize.Key(w)
	if len(key) > 4 {
		return key[:4]
	}
	return key
}
