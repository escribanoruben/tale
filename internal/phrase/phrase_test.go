package phrase

import (
	"strings"
	"testing"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/normalize"
	"github.com/escribanoruben/tale/internal/profile"
)

func TestBestNoRepeatWithinBlock(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	words, err := Best(p, "aaaa", dict, nil, Options{AllowRepeat: false})
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, w := range words {
		key := normalize.Key(w)
		if seen[key] {
			t.Fatalf("repeated word %q in phrase %v", w, words)
		}
		seen[key] = true
	}
}

func TestForPasswordNoRepeatAcrossBlocks(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	p.PasswordLength = 8 // 2 blocks — faster test, same uniqueness rules
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	pass, blocks, err := GeneratePassword(p, dict, Options{AllowRepeat: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(pass) != p.PasswordLength {
		t.Fatalf("password length %d, want %d", len(pass), p.PasswordLength)
	}

	used := make(map[string]bool)
	for _, b := range blocks {
		for i, w := range b.Words {
			if normalize.Initial(w) != string(b.Letters[i]) {
				t.Fatalf("word %q initial mismatch for block %q slot %d", w, b.Letters, i)
			}
			key := normalize.Key(w)
			if used[key] {
				t.Fatalf("repeated word %q across blocks", w)
			}
			used[key] = true
		}
	}
}

func TestRandomForPasswordNoRepeatAcrossBlocks(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	p.PasswordLength = 8
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	pass := "abcd" + "efgh"
	blocks, err := RandomForPassword(p, pass, dict, Options{AllowRepeat: false})
	if err != nil {
		t.Fatal(err)
	}

	used := make(map[string]bool)
	for _, b := range blocks {
		for i, w := range b.Words {
			if normalize.Initial(w) != string(b.Letters[i]) {
				t.Fatalf("word %q initial mismatch for block %q slot %d", w, b.Letters, i)
			}
			key := normalize.Key(w)
			if used[key] {
				t.Fatalf("repeated word %q across blocks", w)
			}
			used[key] = true
		}
	}
}

func TestBestDeterministic(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	block := "mrtp"
	a, err := Best(p, block, dict, nil, Options{AllowRepeat: true})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Best(p, block, dict, nil, Options{AllowRepeat: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(a, " ") != strings.Join(b, " ") {
		t.Fatalf("non-deterministic best phrase: %v vs %v", a, b)
	}
}

func TestEntropyBits(t *testing.T) {
	got := EntropyBits(20, 20, false)
	if got < 85 || got > 87 {
		t.Fatalf("entropy %.2f outside expected ~86 bits", got)
	}
	gotAdj := EntropyBits(20, 20, true)
	if gotAdj >= got {
		t.Fatalf("adjacent-restricted entropy %.2f should be below %.2f", gotAdj, got)
	}
}
