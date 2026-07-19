package dictionary

import (
	"fmt"
	"strings"

	"github.com/escribanoruben/tales_vault_system/internal/normalize"
	"github.com/escribanoruben/tales_vault_system/internal/paths"
	"github.com/escribanoruben/tales_vault_system/internal/profile"
)

// POS identifies a grammatical category.
type POS string

const (
	Noun      POS = "noun"
	Adjective POS = "adjective"
	Verb      POS = "verb"
)

var posFiles = map[POS]string{
	Noun:      "nouns.tsv",
	Adjective: "adjectives.tsv",
	Verb:      "verbs.tsv",
}

// Dictionary provides access to word lists from embed or user data dirs.
type Dictionary struct {
	dataRoot string
	cache    map[string]*Bucket
}

// Default returns a dictionary backed by embedded word data.
func Default() (*Dictionary, error) {
	return &Dictionary{}, nil
}

// ForProfile returns a dictionary for the given resolved profile.
func ForProfile(resolved profile.Resolved) (*Dictionary, error) {
	if resolved.Builtin {
		return Default()
	}
	root, err := paths.DataDir()
	if err != nil {
		return nil, err
	}
	return &Dictionary{dataRoot: root}, nil
}

// WordsFor returns words in alphabetical order for display (e.g. TUI roller).
func (d *Dictionary) WordsFor(language, letter string, pos POS) ([]string, error) {
	bucket, err := d.BucketFor(language, letter, pos)
	if err != nil {
		return nil, err
	}
	return bucket.Words, nil
}

// ParsePOS maps template names to POS values.
func ParsePOS(name string) (POS, error) {
	switch strings.ToLower(name) {
	case "noun", "nouns":
		return Noun, nil
	case "adjective", "adjectives", "adj":
		return Adjective, nil
	case "verb", "verbs":
		return Verb, nil
	default:
		return "", fmt.Errorf("unknown pos %q", name)
	}
}

// FilterAvailable removes already-used words when unique selection is required.
func FilterAvailable(words []string, used map[string]bool) []string {
	if len(used) == 0 {
		return words
	}
	out := make([]string, 0, len(words))
	for _, w := range words {
		key := normalize.Key(w)
		if !used[key] {
			out = append(out, w)
		}
	}
	return out
}
