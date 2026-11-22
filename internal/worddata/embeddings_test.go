package worddata

import (
	"errors"
	"math"
	"testing"
)

func testVectors(dim int, words []string, fill func(w string) []float32) ([]string, [][]float32) {
	rows := make([][]float32, len(words))
	for i, w := range words {
		rows[i] = fill(w)
	}
	return words, rows
}

func TestEmbeddingsRoundtrip(t *testing.T) {
	dim := 8
	words, rows := testVectors(dim, []string{"cat", "dog", "house"}, func(w string) []float32 {
		v := make([]float32, dim)
		for i := range v {
			v[i] = float32(i+1) * float32(len(w))
		}
		return v
	})
	data, err := EncodeEmbeddings(dim, words, rows)
	if err != nil {
		t.Fatal(err)
	}
	e, err := ParseEmbeddings(data)
	if err != nil {
		t.Fatal(err)
	}
	if e.Dim != dim || len(e.Words) != 3 {
		t.Fatalf("dim=%d words=%d", e.Dim, len(e.Words))
	}
	for _, w := range words {
		if !e.Has(w) {
			t.Fatalf("missing %q", w)
		}
		got := e.Vector(w)
		norm := 0.0
		for _, x := range got {
			norm += float64(x) * float64(x)
		}
		if math.Abs(norm-1) > 1e-4 {
			t.Fatalf("%q vector norm %.6f, want 1", w, norm)
		}
	}
	if e.Has("unknown") || e.Vector("unknown") != nil {
		t.Fatal("unknown word should have no vector")
	}
}

func TestEmbeddingsCosine(t *testing.T) {
	// Two collinear vectors -> cosine 1 after quantization roundtrip.
	dim := 4
	a := []float32{1, 0, 0, 0}
	b := []float32{5, 0, 0, 0}
	data, err := EncodeEmbeddings(dim, []string{"a", "b"}, [][]float32{a, b})
	if err != nil {
		t.Fatal(err)
	}
	e, err := ParseEmbeddings(data)
	if err != nil {
		t.Fatal(err)
	}
	cos := 0.0
	for i := range e.Vector("a") {
		cos += float64(e.Vector("a")[i]) * float64(e.Vector("b")[i])
	}
	if math.Abs(cos-1) > 1e-5 {
		t.Fatalf("collinear cosine %.6f, want ~1", cos)
	}
}

func TestLoadEmbeddingsMissingLanguage(t *testing.T) {
	if _, err := LoadEmbeddings("xx"); !errors.Is(err, ErrNoEmbeddings) {
		t.Fatalf("want ErrNoEmbeddings, got %v", err)
	}
}

func TestLoadEmbeddingsBuiltinLanguages(t *testing.T) {
	common := map[string][]string{
		"en": {"house", "time", "eat"},
		"es": {"casa", "perro", "comer"},
	}
	for lang, words := range common {
		e, err := LoadEmbeddings(lang)
		if err != nil {
			t.Fatalf("%s: %v", lang, err)
		}
		if e.Dim == 0 || len(e.Words) < 1000 {
			t.Fatalf("%s: dim=%d words=%d looks wrong", lang, e.Dim, len(e.Words))
		}
		for _, w := range words {
			if !e.Has(w) {
				t.Fatalf("%s: missing common word %q", lang, w)
			}
		}
	}
}
