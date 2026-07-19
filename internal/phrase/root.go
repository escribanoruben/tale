package phrase

import "strings"

func rootKey(word string) string {
	w := strings.ToLower(strings.TrimSpace(word))
	if len(w) > 4 {
		return w[:4]
	}
	return w
}
