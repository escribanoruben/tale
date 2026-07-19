package dictionary

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/escribanoruben/tales_vault_system/internal/normalize"
	"github.com/escribanoruben/tales_vault_system/internal/worddata"
)

// Bucket is a letter/POS word list: alphabetical for display, ranked for selection.
type Bucket struct {
	Words  []string
	rankOf map[string]int // normalized key -> 1-based frequency rank (lower is better)
	size   int
}

// Size returns the number of words in the bucket.
func (b *Bucket) Size() int {
	return b.size
}

// Rank returns the 1-based frequency rank for a word (lower is better).
func (b *Bucket) Rank(word string) int {
	if rank, ok := b.rankOf[normalize.Key(word)]; ok {
		return rank
	}
	return b.size + 1
}

// TopByFrequency returns up to k most frequent words, optionally excluding used keys.
func (b *Bucket) TopByFrequency(k int, used map[string]bool) []string {
	type item struct {
		word string
		rank int
	}
	items := make([]item, 0, len(b.Words))
	for _, word := range b.Words {
		key := normalize.Key(word)
		if used != nil && used[key] {
			continue
		}
		items = append(items, item{word: word, rank: b.rankOf[key]})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].rank != items[j].rank {
			return items[i].rank < items[j].rank
		}
		return items[i].word < items[j].word
	})
	if len(items) > k {
		items = items[:k]
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.word
	}
	return out
}

func (d *Dictionary) bucketCacheKey(language, letter string, pos POS) string {
	return language + "/" + letter + "/" + string(pos)
}

// BucketFor loads a letter/POS bucket with alphabetical words and frequency ranks.
func (d *Dictionary) BucketFor(language, letter string, pos POS) (*Bucket, error) {
	if d.cache == nil {
		d.cache = make(map[string]*Bucket)
	}
	key := d.bucketCacheKey(language, letter, pos)
	if bucket, ok := d.cache[key]; ok {
		return bucket, nil
	}

	filename, ok := posFiles[pos]
	if !ok {
		return nil, fmt.Errorf("unknown pos %q", pos)
	}

	var data []byte
	var err error
	if d.dataRoot != "" {
		path := filepath.Join(d.dataRoot, language, letter, filename)
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
	} else {
		path := language + "/" + letter + "/" + filename
		data, err = fs.ReadFile(worddata.Words, path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
	}

	bucket, err := parseBucket(string(data))
	if err != nil {
		return nil, err
	}
	d.cache[key] = bucket
	return bucket, nil
}

func parseBucket(data string) (*Bucket, error) {
	type record struct {
		word string
		rank int
	}

	var records []record
	lineRank := 0
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lineRank++
		word, rank, err := parseWordLine(line, lineRank)
		if err != nil {
			return nil, err
		}
		if word == "" {
			continue
		}
		records = append(records, record{word: word, rank: rank})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].word < records[j].word
	})

	bucket := &Bucket{
		Words:  make([]string, len(records)),
		rankOf: make(map[string]int, len(records)),
		size:   len(records),
	}
	for i, rec := range records {
		bucket.Words[i] = rec.word
		bucket.rankOf[normalize.Key(rec.word)] = rec.rank
	}
	return bucket, nil
}

// ParseBucketData parses a TSV word bucket (tests and tooling).
func ParseBucketData(data string) (*Bucket, error) {
	return parseBucket(data)
}

func parseWordLine(line string, defaultRank int) (word string, rank int, err error) {
	if tab := strings.IndexByte(line, '\t'); tab >= 0 {
		word = strings.TrimSpace(line[:tab])
		rankField := strings.TrimSpace(line[tab+1:])
		if rankField == "" {
			return word, defaultRank, nil
		}
		rank, err = strconv.Atoi(rankField)
		if err != nil || rank < 1 {
			return "", 0, fmt.Errorf("invalid frequency rank %q for word %q", rankField, word)
		}
		return word, rank, nil
	}
	return line, defaultRank, nil
}
