package phrase

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/normalize"
	"github.com/escribanoruben/tale/internal/password"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/semantic"
)

const searchTopK = 20

// Options controls phrase selection behavior.
type Options struct {
	AllowRepeat          bool
	AllowAdjacentRepeats bool
	// Semantic enables embedding-based ranking (when the language ships
	// vectors); languages or templates without support fall back to the
	// frequency-based scoring.
	Semantic bool
	// Uniform selects words uniformly at random (strategy "random"),
	// instead of the frequency-ranked Best search.
	Uniform bool
}

// BlockResult holds one mnemonic phrase for a password block.
type BlockResult struct {
	Letters string
	Words   []string
}

// EntropyBits estimates password entropy in bits.
func EntropyBits(alphabetLen, passwordLen int, preventAdjacentRepeats bool) float64 {
	return password.EntropyBits(alphabetLen, passwordLen, password.Options{
		PreventAdjacentRepeats: preventAdjacentRepeats,
	})
}

// RandomForPassword selects random valid mnemonic phrases for an existing password.
func RandomForPassword(p profile.Profile, pass string, dict *dictionary.Dictionary, opts Options) ([]BlockResult, error) {
	blocks, err := password.Blocks(pass, p.BlockSize)
	if err != nil {
		return nil, err
	}
	if len(p.PhraseTemplate) != p.BlockSize {
		return nil, fmt.Errorf("phrase template has %d slots, block size is %d", len(p.PhraseTemplate), p.BlockSize)
	}

	used := make(map[string]bool)
	results := make([]BlockResult, 0, len(blocks))
	for _, block := range blocks {
		words, err := Random(p, block, dict, used, opts)
		if err != nil {
			return nil, err
		}
		if !opts.AllowRepeat {
			for _, w := range words {
				used[normalize.Key(w)] = true
			}
		}
		results = append(results, BlockResult{Letters: block, Words: words})
	}
	return results, nil
}

// ForPassword selects mnemonic phrases for an existing password.
func ForPassword(p profile.Profile, pass string, dict *dictionary.Dictionary, opts Options) ([]BlockResult, error) {
	blocks, err := password.Blocks(pass, p.BlockSize)
	if err != nil {
		return nil, err
	}
	if len(p.PhraseTemplate) != p.BlockSize {
		return nil, fmt.Errorf("phrase template has %d slots, block size is %d", len(p.PhraseTemplate), p.BlockSize)
	}

	if opts.Semantic {
		// Whole-story optimization: pools per block + exact k-best DP.
		if results, err := semanticResults(p, blocks, dict, opts); err == nil {
			return results, nil
		}
		// fall through to the per-block frequency path
	}

	used := make(map[string]bool)
	results := make([]BlockResult, 0, len(blocks))
	for _, block := range blocks {
		words, err := Best(p, block, dict, used, opts)
		if err != nil {
			return nil, err
		}
		if !opts.AllowRepeat {
			for _, w := range words {
				used[normalize.Key(w)] = true
			}
		}
		results = append(results, BlockResult{Letters: block, Words: words})
	}
	return results, nil
}

// semanticResults builds the semantic story results for fixed blocks.
func semanticResults(p profile.Profile, blocks []string, dict *dictionary.Dictionary, opts Options) ([]BlockResult, error) {
	stories, err := SemanticStories(p, blocks, dict, opts, semanticStoriesM)
	if err != nil {
		return nil, err
	}
	results := make([]BlockResult, 0, len(blocks))
	for i, block := range blocks {
		results = append(results, BlockResult{Letters: block, Words: stories[0][i]})
	}
	return results, nil
}

// linkNounsFor returns the chain targets for a block: the previous
// phrase's nouns.
func linkNounsFor(prevNouns []string) []string {
	return prevNouns
}

// HasCandidate reports whether at least one valid phrase exists for a block.
func HasCandidate(p profile.Profile, block string, dict *dictionary.Dictionary, used map[string]bool, opts Options) (bool, error) {
	lists, _, err := slotLists(p, block, dict, used, opts)
	if err != nil {
		return false, err
	}
	return searchFirst(lists, opts.AllowRepeat) != nil, nil
}

// Random returns a random valid phrase for a block.
func Random(p profile.Profile, block string, dict *dictionary.Dictionary, used map[string]bool, opts Options) ([]string, error) {
	lists, _, err := slotLists(p, block, dict, used, opts)
	if err != nil {
		return nil, err
	}
	words := searchRandom(lists, opts.AllowRepeat)
	if words == nil {
		return nil, fmt.Errorf("no phrase for block %q", block)
	}
	return words, nil
}

// Best returns the highest-scoring valid phrase for a block.
func Best(p profile.Profile, block string, dict *dictionary.Dictionary, used map[string]bool, opts Options) ([]string, error) {
	lists, buckets, err := slotLists(p, block, dict, used, opts)
	if err != nil {
		return nil, err
	}
	words := searchBest(lists, buckets, opts.AllowRepeat)
	if words == nil {
		return nil, fmt.Errorf("no phrase for block %q", block)
	}
	return words, nil
}

// GeneratePassword creates a password and matching mnemonic phrases with block-level rejection.
func GeneratePassword(p profile.Profile, dict *dictionary.Dictionary, opts Options) (string, []BlockResult, error) {
	numBlocks := p.PasswordLength / p.BlockSize
	if p.PasswordLength%p.BlockSize != 0 {
		return "", nil, fmt.Errorf("password length %d not divisible by block size %d", p.PasswordLength, p.BlockSize)
	}

	if opts.Semantic {
		if pass, results, err := generatePasswordSemantic(p, dict, opts, numBlocks); err == nil {
			return pass, results, nil
		}
		// fall through to the per-block frequency path
	}

	used := make(map[string]bool)
	blockLetters := make([]string, 0, numBlocks)
	results := make([]BlockResult, 0, numBlocks)
	pwOpts := password.Options{PreventAdjacentRepeats: !opts.AllowAdjacentRepeats}
	var prev rune

	for range numBlocks {
		var block string
		var words []string
		found := false
		for attempt := 0; attempt < 10_000; attempt++ {
			b, err := password.GenerateAfter(p.Alphabet, p.BlockSize, prev, pwOpts)
			if err != nil {
				return "", nil, err
			}
			ok, err := HasCandidate(p, b, dict, used, opts)
			if err != nil || !ok {
				continue
			}
			words, err = Best(p, b, dict, used, opts)
			if err != nil {
				continue
			}
			block = b
			found = true
			break
		}
		if !found {
			return "", nil, fmt.Errorf("could not generate valid block after many attempts")
		}
		prev = rune(block[len(block)-1])
		blockLetters = append(blockLetters, block)
		if !opts.AllowRepeat {
			for _, w := range words {
				used[normalize.Key(w)] = true
			}
		}
		results = append(results, BlockResult{Letters: block, Words: words})
	}

	return strings.Join(blockLetters, ""), results, nil
}

// generatePasswordSemantic draws block letters and then optimizes the
// whole narrative over their pools (exact k-best DP). Falls back to the
// per-block frequency path through the caller when it fails.
// StoryPoolsForPassword builds the per-block candidate pools for an
// existing password (story mode: the TUI picker and the polish layer).
func StoryPoolsForPassword(p profile.Profile, pass string, dict *dictionary.Dictionary, opts Options) ([][][]string, error) {
	blocks, err := password.Blocks(pass, p.BlockSize)
	if err != nil {
		return nil, err
	}
	pools := make([][][]string, len(blocks))
	for i, block := range blocks {
		cands, err := semanticPool(p, block, dict, opts, semanticPoolK)
		if err != nil {
			return nil, err
		}
		pool := make([][]string, len(cands))
		for j, c := range cands {
			pool[j] = c.Words
		}
		pools[i] = pool
	}
	return pools, nil
}

// GenerateStoryBlocks draws the block letters (with adjacency rules) and
// builds the per-block candidate pools, for the whole-story DP or the
// LLM polish layer.
func GenerateStoryBlocks(p profile.Profile, dict *dictionary.Dictionary, opts Options) (string, []string, [][][]string, error) {
	numBlocks := p.PasswordLength / p.BlockSize
	if p.PasswordLength%p.BlockSize != 0 {
		return "", nil, nil, fmt.Errorf("password length %d not divisible by block size %d", p.PasswordLength, p.BlockSize)
	}
	pwOpts := password.Options{PreventAdjacentRepeats: !opts.AllowAdjacentRepeats}
	blockLetters := make([]string, 0, numBlocks)
	var prev rune
	for range numBlocks {
		var block string
		found := false
		for attempt := 0; attempt < 10_000; attempt++ {
			b, err := password.GenerateAfter(p.Alphabet, p.BlockSize, prev, pwOpts)
			if err != nil {
				return "", nil, nil, err
			}
			ok, err := HasCandidate(p, b, dict, nil, opts)
			if err != nil || !ok {
				continue
			}
			block = b
			found = true
			break
		}
		if !found {
			return "", nil, nil, fmt.Errorf("could not generate valid block after many attempts")
		}
		prev = rune(block[len(block)-1])
		blockLetters = append(blockLetters, block)
	}
	pools := make([][][]string, numBlocks)
	for i, block := range blockLetters {
		cands, err := semanticPool(p, block, dict, opts, semanticPoolK)
		if err != nil {
			return "", nil, nil, err
		}
		pool := make([][]string, len(cands))
		for j, c := range cands {
			pool[j] = c.Words
		}
		pools[i] = pool
	}
	return strings.Join(blockLetters, ""), blockLetters, pools, nil
}

func generatePasswordSemantic(p profile.Profile, dict *dictionary.Dictionary, opts Options, numBlocks int) (string, []BlockResult, error) {
	pass, blockLetters, _, err := GenerateStoryBlocks(p, dict, opts)
	if err != nil {
		return "", nil, err
	}
	stories, err := SemanticStories(p, blockLetters, dict, opts, semanticStoriesM)
	if err != nil {
		return "", nil, err
	}
	results := make([]BlockResult, 0, numBlocks)
	for i, block := range blockLetters {
		results = append(results, BlockResult{Letters: block, Words: stories[0][i]})
	}
	return pass, results, nil
}

// phraseNouns returns the words sitting in the template's noun slots.
func phraseNouns(p profile.Profile, words []string) []string {
	var nouns []string
	for i, pos := range p.PhraseTemplate {
		if pos == "noun" && i < len(words) {
			nouns = append(nouns, words[i])
		}
	}
	return nouns
}

// phraseSubject returns the subject noun of a picked phrase (the noun
// adjacent to the adjective, per the anchored roles), or "" when the
// template cannot express it.
func phraseSubject(p profile.Profile, words []string) string {
	roles, ok := semanticRoles(p.PhraseTemplate)
	if !ok || roles[semantic.Subject] >= len(words) {
		return ""
	}
	return words[roles[semantic.Subject]]
}

func slotLists(p profile.Profile, block string, dict *dictionary.Dictionary, used map[string]bool, opts Options) ([][]string, []*dictionary.Bucket, error) {
	if len(block) != len(p.PhraseTemplate) {
		return nil, nil, fmt.Errorf("block %q length %d does not match template", block, len(p.PhraseTemplate))
	}
	lists := make([][]string, len(p.PhraseTemplate))
	buckets := make([]*dictionary.Bucket, len(p.PhraseTemplate))
	for i, posName := range p.PhraseTemplate {
		pos, err := dictionary.ParsePOS(posName)
		if err != nil {
			return nil, nil, err
		}
		letter := string(block[i])
		bucket, err := dict.BucketFor(p.Language, letter, pos)
		if err != nil {
			return nil, nil, err
		}
		buckets[i] = bucket
		var usedFilter map[string]bool
		if !opts.AllowRepeat {
			usedFilter = used
		}
		words := bucket.TopByFrequency(searchTopK, usedFilter)
		if len(words) == 0 {
			return nil, nil, fmt.Errorf("no words for %s starting with %q", posName, letter)
		}
		lists[i] = words
	}
	return lists, buckets, nil
}

func searchRandom(lists [][]string, allowRepeat bool) []string {
	shuffled := make([][]string, len(lists))
	for i, list := range lists {
		shuffled[i] = shuffleCopy(list)
	}
	return searchFirst(shuffled, allowRepeat)
}

func shuffleCopy(words []string) []string {
	out := append([]string(nil), words...)
	for i := len(out) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			break
		}
		j := int(n.Int64())
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func searchFirst(lists [][]string, allowRepeat bool) []string {
	var out []string
	var used map[string]bool
	if !allowRepeat {
		used = make(map[string]bool)
	}
	path := make([]string, 0, len(lists))
	var walk func(depth int) bool
	walk = func(depth int) bool {
		if depth == len(lists) {
			out = append([]string(nil), path...)
			return true
		}
		for _, w := range lists[depth] {
			if !allowRepeat {
				key := normalize.Key(w)
				if used[key] {
					continue
				}
				used[key] = true
			}
			path = append(path, w)
			if walk(depth + 1) {
				return true
			}
			path = path[:len(path)-1]
			if !allowRepeat {
				delete(used, normalize.Key(w))
			}
		}
		return false
	}
	if walk(0) {
		return out
	}
	return nil
}

func searchBest(lists [][]string, buckets []*dictionary.Bucket, allowRepeat bool) []string {
	var best []string
	var bestScore int64 = math.MinInt64
	var used map[string]bool
	if !allowRepeat {
		used = make(map[string]bool)
	}
	combo := make([]string, 0, len(lists))
	var walk func(depth int)
	walk = func(depth int) {
		if depth == len(lists) {
			words := append([]string(nil), combo...)
			score := scorePhrase(words, buckets)
			if score > bestScore {
				bestScore = score
				best = words
			}
			return
		}
		for _, w := range lists[depth] {
			if !allowRepeat {
				key := normalize.Key(w)
				if used[key] {
					continue
				}
				used[key] = true
			}
			combo = append(combo, w)
			walk(depth + 1)
			combo = combo[:len(combo)-1]
			if !allowRepeat {
				delete(used, normalize.Key(w))
			}
		}
	}
	walk(0)
	return best
}
