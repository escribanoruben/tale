package semantic

import (
	"strings"
	"testing"

	"github.com/escribanoruben/tale/internal/worddata"
)

func fieldTestEmbeddings(t *testing.T) *worddata.Embeddings {
	t.Helper()
	// Two semantic axes: sun/light and sea/water. Candidates per
	// position mix both so the strategies can pick.
	words := []string{"sol", "luz", "sal", "mar", "oro", "ola", "astro", "marea"}
	rows := [][]float32{
		{1, 0, 0},       // sol (sun axis)
		{0.9, 0.1, 0},   // luz (sun axis)
		{0, 1, 0},       // sal (sea axis)
		{0.1, 0.9, 0},   // mar (sea axis)
		{0.8, 0.2, 0},   // oro (sun-ish)
		{0.2, 0.8, 0},   // ola (sea-ish)
		{0.85, 0.15, 0}, // astro (sun axis)
		{0.15, 0.85, 0}, // marea (sea axis)
	}
	data, err := worddata.EncodeEmbeddings(3, words, rows)
	if err != nil {
		t.Fatal(err)
	}
	e, err := worddata.ParseEmbeddings(data)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func fieldCands() [][]string {
	return [][]string{
		{"sol", "sal"},     // letter a
		{"luz", "mar"},     // letter l
		{"oro", "ola"},     // letter o
		{"astro", "marea"}, // letter s
	}
}

func TestNounFieldPicksOneAxis(t *testing.T) {
	e := fieldTestEmbeddings(t)
	ws, err := NounField(e, fieldCands())
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 4 {
		t.Fatalf("want 4 nouns, got %d", len(ws))
	}
	seen := map[string]bool{}
	var axes []string
	for _, w := range ws {
		if seen[w] {
			t.Fatalf("word repeated: %v", ws)
		}
		seen[w] = true
		switch w {
		case "sol", "luz", "oro", "astro":
			axes = append(axes, "sun")
		case "sal", "mar", "ola", "marea":
			axes = append(axes, "sea")
		default:
			t.Fatalf("unknown word %q", w)
		}
	}
	// one coherent axis, no mixing halfway
	if axes[0] == "sun" {
		for _, a := range axes {
			if a != "sun" {
				t.Fatalf("field mixed axes: %v (words %v)", axes, ws)
			}
		}
	} else {
		for _, a := range axes {
			if a != "sea" {
				t.Fatalf("field mixed axes: %v (words %v)", axes, ws)
			}
		}
	}
}

func TestNounPathNoRepeatedWords(t *testing.T) {
	e := fieldTestEmbeddings(t)
	// Degenerate-ish case: identical candidate lists everywhere, so the
	// plain DP would happily repeat words; the mask loop must fix it
	// (greedy fallback runs on the unmasked candidates).
	cands := [][]string{
		{"sol", "sal", "mar", "oro"},
		{"sol", "sal", "mar", "oro"},
		{"sol", "sal", "mar", "oro"},
	}
	ws, err := NounPath(e, cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 3 {
		t.Fatalf("want 3 nouns, got %d", len(ws))
	}
	seen := map[string]bool{}
	for _, w := range ws {
		if seen[w] {
			t.Fatalf("word repeated in path: %v", ws)
		}
		seen[w] = true
	}
}

func TestNounFieldExhaustedCandidates(t *testing.T) {
	e := fieldTestEmbeddings(t)
	// more positions than distinct roots -> clean error, not a crash
	cands := [][]string{
		{"sol", "sal"}, {"sol", "sal"}, {"sol", "sal"}, {"sol", "sal"},
	}
	if _, err := NounField(e, cands); err == nil {
		t.Fatal("expected error when candidates exhaust")
	}
}

func TestNounPathStutterGuard(t *testing.T) {
	e := fieldTestEmbeddings(t)
	// "luz luz luz" is a synonym hop chain (cos 1.0); the guard plus
	// dedupe must prevent it when alternatives exist
	cands := [][]string{
		{"luz", "sal"},
		{"luz", "mar"},
		{"luz", "ola"},
	}
	ws, err := NounPath(e, cands)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(ws, "|") == "luz|luz|luz" {
		t.Fatalf("stutter chain survived the guard: %v", ws)
	}
}
