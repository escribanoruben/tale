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
	BandHigh            = 0.45 // reject noun pairs closer than this (near-synonyms)
	BandLow             = 0.05 // gently pull up pairs looser than this
	BandHighWeight      = 1.0
	BandLowWeight       = 0.3
	topCombosPerSubject = 3 // verb/object combos kept per subject for pool diversity
)

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
}

// DefaultLinkWeight pairs with the offline experiment: strong enough to
// steer the subject choice, weak enough to keep scene quality.
const DefaultLinkWeight = 2.0

// ImageabilityPenalty is the score cost per low-imageability word
// (abstract noun, stative verb): from the offline experiment, cutting
// abstract top-1 phrases from 23% to 3% of blocks at negligible quality
// cost.
const ImageabilityPenalty = 0.25

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
		// verb/object table for this subject; the band couples
		// subject-object.
		d := make([]float64, len(verb)*len(obj))
		for j := range verb {
			for o := range obj {
				nn := dot(sub[i], obj[o])
				d[j*len(obj)+o] = mSVerb[i][j] + mVObj[j][o] + Band(nn)
			}
		}
		idx := make([]int, len(d))
		for t := range idx {
			idx[t] = t
		}
		sort.SliceStable(idx, func(a, b int) bool { return d[idx[a]] > d[idx[b]] })
		for _, t := range idx[:min(topCombosPerSubject, len(idx))] {
			j, o := t/len(obj), t%len(obj)
			score := (mAdj[i][bestAdj] + d[t]) / 3
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
	for _, r := range raws {
		if len(out) == ro.K {
			break
		}
		w := [4]string{words[Subject][r.sub], words[Verb][r.vb], words[Object][r.obj], words[Adjective][r.adj]}
		if !ro.AllowRepeat {
			distinct := true
			for a := 0; a < 4 && distinct; a++ {
				if picked[normalize.Key(w[a])] || pickedRoots[rootOf(w[a])] || (ro.UsedRoots != nil && ro.UsedRoots[rootOf(w[a])]) {
					distinct = false
					break
				}
				for b := a + 1; b < 4; b++ {
					if w[a] == w[b] || rootOf(w[a]) == rootOf(w[b]) {
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
			for _, x := range w {
				picked[normalize.Key(x)] = true
				if ro.UsedRoots != nil {
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

// rootOf is the cross-block root key (first four letters, lowercased):
// words sharing it ("gobierno"/"gobernar") feel like repeats in a story.
func rootOf(w string) string {
	key := normalize.Key(w)
	if len(key) > 4 {
		return key[:4]
	}
	return key
}
