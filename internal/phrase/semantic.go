package phrase

import (
	"fmt"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/semantic"
	"github.com/escribanoruben/tale/internal/worddata"
)

const (
	// semanticPoolK is the number of ranked candidates per block (the
	// TUI cycles through them); more nodes per layer widen the DP
	// search without changing its guarantees. semanticStoriesM is how
	// many complete stories the DP keeps for browsing. MaxSharedWords=1
	// lets the same noun appear in several scene variants.
	semanticPoolK    = 100
	semanticStoriesM = 8
)

// semanticRoles maps semantic roles to template slot indexes.
// Requires a template with exactly two nouns, one verb and one adjective;
// the subject is the noun adjacent to the adjective in reading order.
// Returns ok=false for templates the anchored score cannot express
// (those languages keep the frequency-based scoring).
func semanticRoles(template []string) ([4]int, bool) {
	var roles [4]int
	if len(template) != 4 {
		return roles, false
	}
	var nouns []int
	verbIdx, adjIdx := -1, -1
	for i, pos := range template {
		switch pos {
		case "noun":
			nouns = append(nouns, i)
		case "verb":
			if verbIdx >= 0 {
				return roles, false
			}
			verbIdx = i
		case "adjective":
			if adjIdx >= 0 {
				return roles, false
			}
			adjIdx = i
		default:
			return roles, false
		}
	}
	if len(nouns) != 2 || verbIdx < 0 || adjIdx < 0 {
		return roles, false
	}
	subject := -1
	for _, n := range nouns {
		if abs(n-adjIdx) == 1 {
			subject = n
			break
		}
	}
	if subject < 0 {
		return roles, false
	}
	object := nouns[0]
	if object == subject {
		object = nouns[1]
	}
	roles[semantic.Subject] = subject
	roles[semantic.Verb] = verbIdx
	roles[semantic.Object] = object
	roles[semantic.Adjective] = adjIdx
	return roles, true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// SemanticAvailable reports whether semantic ranking can run for this
// profile: supported template, embedded vectors, and at least one
// vector-backed word per role and alphabet letter.
func SemanticAvailable(p profile.Profile, dict *dictionary.Dictionary) bool {
	if _, ok := semanticRoles(p.PhraseTemplate); !ok {
		return false
	}
	emb, err := worddata.LoadEmbeddings(p.Language)
	if err != nil {
		return false
	}
	posByRole := [4]dictionary.POS{dictionary.Noun, dictionary.Verb, dictionary.Noun, dictionary.Adjective}
	for _, letter := range p.Alphabet {
		for _, posName := range posByRole {
			bucket, err := dict.BucketFor(p.Language, string(letter), posName)
			if err != nil {
				return false
			}
			if !hasVectorizedWord(emb, bucket.Words) {
				return false
			}
		}
	}
	return true
}

func hasVectorizedWord(emb *worddata.Embeddings, words []string) bool {
	for _, w := range words {
		if emb.Has(w) {
			return true
		}
	}
	return false
}

// semanticPool is the context-free pool generator returning ranked
// candidates (used by SemanticPool and SemanticStories).
func semanticPool(p profile.Profile, block string, dict *dictionary.Dictionary, opts Options, k int) ([]semantic.Candidate, error) {
	roles, ok := semanticRoles(p.PhraseTemplate)
	if !ok {
		return nil, fmt.Errorf("semantic ranking unsupported for template %v", p.PhraseTemplate)
	}
	emb, err := worddata.LoadEmbeddings(p.Language)
	if err != nil {
		return nil, err
	}
	lowImg, err := worddata.LoadImageability(p.Language)
	if err != nil {
		return nil, err
	}
	lists, buckets, err := slotLists(p, block, dict, nil, opts)
	if err != nil {
		return nil, err
	}
	// Semantic ranking scores the full bucket (up to 120 words): the
	// frequency cap would let function words dominate the candidate set.
	for i := range lists {
		lists[i] = buckets[i].Words
	}
	return semantic.Rank(emb, lists, roles, semantic.RankOptions{
		K: k, LowImageability: lowImg, MaxSharedWords: 1,
	}), nil
}

// SemanticPool ranks the top-k candidate phrases for one block under the
// anchored semantic score. Context-free: cross-block constraints (word
// and root reuse, chaining) are applied later, over the whole story.
// Returns worddata.ErrNoEmbeddings for languages without embedded
// vectors (callers fall back to frequency scoring).
func SemanticPool(p profile.Profile, block string, dict *dictionary.Dictionary, opts Options, k int) ([][]string, error) {
	cands, err := semanticPool(p, block, dict, opts, k)
	if err != nil {
		return nil, err
	}
	out := make([][]string, len(cands))
	for i, c := range cands {
		out[i] = c.Words
	}
	return out, nil
}

// SemanticStories optimizes the whole narrative at once: per-block
// candidate pools + exact k-best dynamic programming over chain bonuses
// (subject vs previous phrase's nouns) and the closure bonus (final
// object noun vs first subject noun). Stories are returned best first,
// limited to ones without repeated words or roots across blocks. When
// every story repeats something, the least constrained ones are still
// returned so callers always get output.
func SemanticStories(p profile.Profile, blocks []string, dict *dictionary.Dictionary, opts Options, m int) ([][][]string, error) {
	roles, ok := semanticRoles(p.PhraseTemplate)
	if !ok {
		return nil, fmt.Errorf("semantic ranking unsupported for template %v", p.PhraseTemplate)
	}
	emb, err := worddata.LoadEmbeddings(p.Language)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks to narrate")
	}
	pools := make([][]semantic.Candidate, len(blocks))
	for i, block := range blocks {
		cands, err := semanticPool(p, block, dict, opts, semanticPoolK)
		if err != nil {
			return nil, err
		}
		pools[i] = cands
	}
	stories := semantic.Stories(emb, pools, roles, m*8)
	if len(stories) == 0 {
		return nil, fmt.Errorf("no semantic story for the given blocks")
	}

	// Cross-block word repetition is allowed: a repeated word reads as a
	// refrain and can aid recall. Stories come ranked by the DP score.
	out := make([][][]string, 0, len(stories))
	for _, st := range stories {
		out = append(out, st.Phrases)
	}
	return out, nil
}
