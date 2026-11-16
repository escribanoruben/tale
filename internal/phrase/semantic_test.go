package phrase

import (
	"strings"
	"testing"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
)

func testProfile(t *testing.T, id string) profile.Profile {
	t.Helper()
	p, err := profile.Load(profile.Files, id)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func testDict(t *testing.T) *dictionary.Dictionary {
	t.Helper()
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}
	return dict
}

func TestSemanticAvailableBuiltin(t *testing.T) {
	dict := testDict(t)
	for _, id := range []string{"en", "es"} {
		p := testProfile(t, id)
		if !SemanticAvailable(p, dict) {
			t.Fatalf("%s: semantic ranking should be available", id)
		}
	}
}

func TestSemanticPoolInitialsSpellBlock(t *testing.T) {
	p := testProfile(t, "es")
	dict := testDict(t)
	opts := Options{AllowRepeat: false}
	block := "casa"
	pool, err := SemanticPool(p, block, dict, opts, semanticPoolK)
	if err != nil {
		t.Fatal(err)
	}
	if len(pool) == 0 {
		t.Fatal("empty pool")
	}
	for _, words := range pool {
		if len(words) != len(p.PhraseTemplate) {
			t.Fatalf("phrase %v does not match template", words)
		}
		for i, w := range words {
			initial := strings.ToLower(string([]rune(w)[0]))
			if initial != string(block[i]) {
				t.Fatalf("word %q initial %q != block letter %q", w, initial, block[i])
			}
		}
	}
	// Diversity: no word shared across pool candidates.
	seen := map[string]bool{}
	for _, words := range pool {
		for _, w := range words {
			if seen[w] {
				t.Fatalf("word %q repeated across pool", w)
			}
			seen[w] = true
		}
	}
}

func TestSemanticPoolMissingVectors(t *testing.T) {
	p := testProfile(t, "es")
	p.Language = "zz" // no embeddings shipped for it
	dict := testDict(t)
	if SemanticAvailable(p, dict) {
		t.Fatal("semantic must be unavailable without vectors")
	}
	if _, err := SemanticPool(p, "casa", dict, Options{}, 5); err == nil {
		t.Fatal("expected error for language without vectors")
	}
}

func TestSemanticDeterministic(t *testing.T) {
	p := testProfile(t, "es")
	dict := testDict(t)
	opts := Options{AllowRepeat: false}
	a, err := SemanticPool(p, "casa", dict, opts, semanticPoolK)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SemanticPool(p, "casa", dict, opts, semanticPoolK)
	if err != nil {
		t.Fatal(err)
	}
	for i := range a {
		if strings.Join(a[i], " ") != strings.Join(b[i], " ") {
			t.Fatalf("pool not deterministic:\n%v\n%v", a[i], b[i])
		}
	}
}

func TestPhraseSubject(t *testing.T) {
	p := testProfile(t, "es") // noun adj verb noun -> subject at slot 0
	words := []string{"perro", "listo", "ladra", "carta"}
	if got := phraseSubject(p, words); got != "perro" {
		t.Fatalf("subject = %q, want perro", got)
	}
}

func TestSemanticStoriesShapes(t *testing.T) {
	p := testProfile(t, "es")
	dict := testDict(t)
	blocks := []string{"casa", "mesa"}
	stories, err := SemanticStories(p, blocks, dict, Options{}, semanticStoriesM)
	if err != nil {
		t.Fatal(err)
	}
	if len(stories) == 0 {
		t.Fatal("no stories")
	}
	for _, story := range stories {
		if len(story) != len(blocks) {
			t.Fatalf("story blocks %d, want %d", len(story), len(blocks))
		}
		seen := map[string]bool{}
		for bi, words := range story {
			for i, w := range words {
				initial := strings.ToLower(string([]rune(w)[0]))
				if initial != string(blocks[bi][i]) {
					t.Fatalf("word %q initial != block letter (story %v)", w, story)
				}
				if seen[w] {
					t.Fatalf("word %q repeated across story", w)
				}
				seen[w] = true
			}
		}
	}
	// determinism
	again, err := SemanticStories(p, blocks, dict, Options{}, semanticStoriesM)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(stories[0][0], " ") != strings.Join(again[0][0], " ") {
		t.Fatal("stories not deterministic")
	}
}
