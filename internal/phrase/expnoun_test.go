package phrase

// Temporary experiment: sample outputs for noun strategies. Run:
//   go test ./internal/phrase -run TestExperimentNounExamples -v
import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/worddata"
)

func expCandidates2(t *testing.T, p profile.Profile, dict *dictionary.Dictionary, emb *worddata.Embeddings) map[byte][]string {
	lowImg, _ := worddata.LoadImageability("es")
	out := map[byte][]string{}
	for _, r := range p.Alphabet {
		letter := byte(r)
		bucket, err := dict.BucketFor("es", string(letter), dictionary.Noun)
		if err != nil {
			t.Fatal(err)
		}
		var ws []string
		for _, w := range bucket.Words {
			if _, low := lowImg[w]; low {
				continue
			}
			if emb.Has(w) {
				ws = append(ws, w)
			}
		}
		if len(ws) > 0 {
			out[letter] = ws
		}
	}
	return out
}

func expDot2(e *worddata.Embeddings, a, b string) float64 {
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

func expPath2(letters []byte, cands map[byte][]string, e *worddata.Embeddings) []string {
	n := len(letters)
	best := make([][]float64, n)
	prev := make([][]int, n)
	for range cands[letters[0]] {
		best[0] = append(best[0], 0)
		prev[0] = append(prev[0], -1)
	}
	edge := func(a, b string) float64 {
		c := expDot2(e, a, b)
		if c > 0.75 {
			c -= 2.0 * (c - 0.75)
		}
		return c
	}
	for i := 1; i < n; i++ {
		for _, w := range cands[letters[i]] {
			bi, bv := -1, math.Inf(-1)
			for a, wp := range cands[letters[i-1]] {
				v := best[i-1][a] + edge(wp, w)
				if v > bv {
					bv, bi = v, a
				}
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
	idx := make([]int, n)
	idx[n-1] = bi
	for i := n - 1; i > 0; i-- {
		idx[i-1] = prev[i][idx[i]]
	}
	ws := make([]string, n)
	for i := range ws {
		ws[i] = cands[letters[i]][idx[i]]
	}
	return ws
}

func expField2(letters []byte, cands map[byte][]string, e *worddata.Embeddings, rng *rand.Rand) []string {
	seed := cands[letters[0]][rng.Intn(len(cands[letters[0]]))]
	centroid := append([]float32(nil), e.Vector(seed)...)
	var chosen []string
	for iter := 0; iter < 3; iter++ {
		chosen = nil
		roots := map[string]bool{}
		sum := make([]float64, len(centroid))
		for _, l := range letters {
			bw, bv := "", math.Inf(-1)
			for _, w := range cands[l] {
				r := w
				if len(r) > 4 {
					r = r[:4]
				}
				if roots[r] {
					continue
				}
				v := e.Vector(w)
				var d float64
				for k := range centroid {
					d += float64(v[k] * centroid[k])
				}
				if d > bv {
					bv, bw = d, w
				}
			}
			if bw == "" {
				bw = cands[l][0]
			}
			chosen = append(chosen, bw)
			r := bw
			if len(r) > 4 {
				r = r[:4]
			}
			roots[r] = true
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
	return chosen
}

func TestExperimentNounExamples(t *testing.T) {
	p := testProfile(t, "es")
	dict := testDict(t)
	emb, err := worddata.LoadEmbeddings("es")
	if err != nil {
		t.Fatal(err)
	}
	cands := expCandidates2(t, p, dict, emb)
	rng := rand.New(rand.NewSource(7))
	const L = 16
	for i := 0; i < 5; i++ {
		letters := make([]byte, L)
		for j := range letters {
			letters[j] = p.Alphabet[rng.Intn(len(p.Alphabet))]
		}
		password := string(letters)
		fmt.Printf("\n=== password %s ===\n", password)
		fmt.Printf("  campo: %v\n", expField2(letters, cands, emb, rng))
		fmt.Printf("  camino: %v\n", expPath2(letters, cands, emb))
	}
}
