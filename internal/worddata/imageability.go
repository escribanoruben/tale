package worddata

import (
	"errors"
	"io/fs"
	"strings"
)

// LoadImageability returns the set of low-imageability words for a
// language (imageability.txt next to the word lists): abstract nouns and
// stative verbs that are hard to picture and easy to forget. A missing
// file yields an empty set — languages without the asset simply get no
// penalty.
func LoadImageability(lang string) (map[string]struct{}, error) {
	data, err := fs.ReadFile(Words, lang+"/imageability.txt")
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	set := make(map[string]struct{})
	for _, line := range strings.Split(string(data), "\n") {
		w := strings.ToLower(strings.TrimSpace(line))
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		set[w] = struct{}{}
	}
	return set, nil
}
