package semantic

import (
	"sort"

	"github.com/escribanoruben/tale/internal/worddata"
)

// Story is one complete narrative: one phrase per block, ranked.
type Story struct {
	Phrases [][]string // story.Phrases[block] = slot-ordered words
	Score   float64    // sum of phrase scores + chain/closure bonuses
}

// Stories optimizes the whole narrative at once.
//
// pools holds the per-block candidate pools (from Rank, one per block in
// block order). It builds a layered DAG — node = candidate phrase, edge =
// chain bonus between consecutive blocks (subject of the next phrase vs
// the previous phrase's nouns) — and runs an exact k-best dynamic
// program over the layers, then adds the closure bonus (final block's
// object noun back to the first phrase's subject) to the reconstructed
// paths and re-ranks. With the per-first-node structure folded into a
// single pass this scales O(K^2) in the pool size, which lets the pools
// be wide (hundreds of nodes) without runtime blow-ups. Deterministic.
func Stories(e *worddata.Embeddings, pools [][]Candidate, roles [4]int, m int) []Story {
	if m <= 0 {
		m = 1
	}
	n := len(pools)
	if n == 0 {
		return nil
	}
	for i := range pools {
		if len(pools[i]) == 0 {
			return nil
		}
	}
	if n == 1 {
		out := make([]Story, 0, min(m, len(pools[0])))
		for i, c := range pools[0] {
			if i == m {
				break
			}
			out = append(out, Story{Phrases: [][]string{c.Words}, Score: c.Score})
		}
		return out
	}

	dim := e.Dim
	dot := func(a, b []float32) float64 {
		var s float64
		for i := 0; i < dim; i++ {
			s += float64(a[i]) * float64(b[i])
		}
		return s
	}
	// Per-node anchor vectors.
	n1 := make([][][]float32, n)
	n2 := make([][][]float32, n)
	for i := range pools {
		for _, c := range pools[i] {
			if roles[Subject] >= len(c.Words) || roles[Object] >= len(c.Words) {
				continue
			}
			v1 := e.Vector(c.Words[roles[Subject]])
			v2 := e.Vector(c.Words[roles[Object]])
			if v1 == nil || v2 == nil {
				continue
			}
			n1[i] = append(n1[i], v1)
			n2[i] = append(n2[i], v2)
		}
		if len(n1[i]) != len(pools[i]) {
			return nil // a node lacks vectors; pools must be pre-filtered
		}
	}

	chain := func(i, a, b int) float64 {
		d1 := dot(n1[i+1][b], n1[i][a])
		d2 := dot(n1[i+1][b], n2[i][a])
		return DefaultLinkWeight * max(d1, d2)
	}

	type step struct {
		score    float64
		prev     int
		prevRank int
	}
	const keepPerNode = 64 // k-best paths kept per node
	layers := make([][][]step, n)
	layers[0] = make([][]step, len(pools[0]))
	for j := range pools[0] {
		layers[0][j] = []step{{score: pools[0][j].Score}}
	}
	for i := 1; i < n; i++ {
		layers[i] = make([][]step, len(pools[i]))
		for j := range pools[i] {
			var cands []step
			for a := range pools[i-1] {
				lw := chain(i-1, a, j)
				for r, st := range layers[i-1][a] {
					cands = append(cands, step{st.score + lw + pools[i][j].Score, a, r})
				}
			}
			sort.SliceStable(cands, func(x, y int) bool { return cands[x].score > cands[y].score })
			if len(cands) > keepPerNode {
				cands = cands[:keepPerNode]
			}
			layers[i][j] = cands
		}
	}

	// Final nodes x ranks, closed back to their first phrase.
	type fin struct {
		score float64
		node  int
		rank  int
	}
	var finals []fin
	for j := range pools[n-1] {
		for r := range layers[n-1][j] {
			finals = append(finals, fin{layers[n-1][j][r].score, j, r})
		}
	}
	sort.SliceStable(finals, func(x, y int) bool { return finals[x].score > finals[y].score })
	const backtrackBudget = 2000
	if len(finals) > backtrackBudget {
		finals = finals[:backtrackBudget]
	}

	var all []Story
	for _, f := range finals {
		phrases := make([][]string, n)
		node, rank := f.node, f.rank
		for i := n - 1; i >= 1; i-- {
			phrases[i] = pools[i][node].Words
			st := layers[i][node][rank]
			node, rank = st.prev, st.prevRank
		}
		phrases[0] = pools[0][node].Words
		// closure: the story's last object noun back to its opening subject
		score := f.score + DefaultLinkWeight*dot(n2[n-1][f.node], n1[0][node])
		all = append(all, Story{Phrases: phrases, Score: score})
	}

	sort.SliceStable(all, func(x, y int) bool { return all[x].Score > all[y].Score })
	if len(all) > m {
		all = all[:m]
	}
	return all
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
