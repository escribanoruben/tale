package phrase

import (
	"fmt"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/password"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/semantic"
	"github.com/escribanoruben/tale/internal/worddata"
)

// Noun-strategy wiring: the password is spelled by the initials of X
// nouns (one per letter), sharing a semantic field (noun_field) or
// chained by consecutive adjacency (noun_path).

// nounCandidates builds the per-letter candidate noun lists for a
// password: embedded-vector present, not low-imageability.
func nounCandidates(p profile.Profile, letters []byte, dict *dictionary.Dictionary, emb *worddata.Embeddings) ([][]string, error) {
	lowImg, err := worddata.LoadImageability(p.Language)
	if err != nil {
		return nil, err
	}
	cands := make([][]string, len(letters))
	for i, letter := range letters {
		bucket, err := dict.BucketFor(p.Language, string(letter), dictionary.Noun)
		if err != nil {
			return nil, err
		}
		for _, w := range bucket.Words {
			if _, low := lowImg[w]; low {
				continue
			}
			if emb.Has(w) {
				cands[i] = append(cands[i], w)
			}
		}
		if len(cands[i]) == 0 {
			return nil, fmt.Errorf("no noun candidates for letter %q (strategy %q)", letter, p.Strategy)
		}
	}
	return cands, nil
}

// GenerateNouns creates a password and its noun mnemonic under a noun
// strategy (noun_field or noun_path).
func GenerateNouns(p profile.Profile, dict *dictionary.Dictionary, opts Options) (string, []string, error) {
	if _, err := worddata.LoadEmbeddings(p.Language); err != nil {
		return "", nil, fmt.Errorf("strategy %q requires embedded semantic vectors: %w", p.Strategy, err)
	}
	pwOpts := password.Options{PreventAdjacentRepeats: !opts.AllowAdjacentRepeats}
	letters := make([]byte, 0, p.PasswordLength)
	var prev rune
	for range p.PasswordLength {
		b, err := password.GenerateAfter(p.Alphabet, 1, prev, pwOpts)
		if err != nil {
			return "", nil, err
		}
		letters = append(letters, b[0])
		prev = rune(b[0])
	}
	words, err := NounsForPassword(p, string(letters), dict)
	if err != nil {
		return "", nil, err
	}
	return string(letters), words, nil
}

// NounsForPassword selects the mnemonic nouns for an existing password.
func NounsForPassword(p profile.Profile, pass string, dict *dictionary.Dictionary) ([]string, error) {
	emb, err := worddata.LoadEmbeddings(p.Language)
	if err != nil {
		return nil, fmt.Errorf("strategy %q requires embedded semantic vectors: %w", p.Strategy, err)
	}
	letters := []byte(pass)
	cands, err := nounCandidates(p, letters, dict, emb)
	if err != nil {
		return nil, err
	}
	switch p.Strategy {
	case profile.StrategyNounPath:
		return semantic.NounPath(emb, cands)
	case profile.StrategyNounField, "":
		return semantic.NounField(emb, cands)
	default:
		return nil, fmt.Errorf("strategy %q does not produce noun mnemonics", p.Strategy)
	}
}
