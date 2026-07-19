package phrase

import (
	"github.com/ruben/tales_vault_system/internal/dictionary"
	"github.com/ruben/tales_vault_system/internal/normalize"
)

func scorePhrase(words []string, buckets []*dictionary.Bucket) int64 {
	var score int64
	for i, w := range words {
		rank := buckets[i].Rank(w)
		score += int64(buckets[i].Size()+1-rank) * 100

		n := len(w)
		switch {
		case n >= 4 && n <= 8:
			score += 50
		case n < 4:
			score -= 20
		default:
			score -= 10
		}
	}

	roots := make(map[string]int)
	initials := make(map[string]int)
	for _, w := range words {
		root := rootKey(w)
		roots[root]++
		initials[normalize.Initial(w)]++
	}
	for _, count := range roots {
		if count > 1 {
			score -= int64(count-1) * 80
		}
	}
	for _, count := range initials {
		if count > 1 {
			score -= int64(count-1) * 30
		}
	}
	return score
}
