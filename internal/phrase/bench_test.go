package phrase

import (
	"testing"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
)

func BenchmarkGenerateSemantic(b *testing.B) {
	p, err := profile.Load(profile.Files, "es")
	if err != nil {
		b.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		b.Fatal(err)
	}
	opts := Options{Semantic: true}
	for b.Loop() {
		if _, _, err := GeneratePassword(p, dict, opts); err != nil {
			b.Fatal(err)
		}
	}
}
