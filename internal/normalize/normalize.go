package normalize

import (
	"strings"
	"unicode/utf8"
)

// Initial returns the normalized first letter used for password matching.
func Initial(word string) string {
	word = strings.TrimSpace(strings.ToLower(word))
	if word == "" {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(word)
	if r == 'ñ' {
		return ""
	}
	r = stripAccent(r)
	if r < 'a' || r > 'z' {
		return ""
	}
	return string(r)
}

func stripAccent(r rune) rune {
	switch r {
	case 'á', 'à', 'ä', 'â', 'ã', 'å':
		return 'a'
	case 'é', 'è', 'ë', 'ê':
		return 'e'
	case 'í', 'ì', 'ï', 'î':
		return 'i'
	case 'ó', 'ò', 'ö', 'ô', 'õ':
		return 'o'
	case 'ú', 'ù', 'ü', 'û':
		return 'u'
	default:
		return r
	}
}

// Key returns a case-insensitive key for duplicate detection.
func Key(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
