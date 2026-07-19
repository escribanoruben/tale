package profile

import (
	"os"
	"sort"
	"strings"

	"golang.org/x/text/language"
)

// SystemLanguage returns the primary language tag from the environment (e.g. "es").
// It returns an empty string when no usable locale is set.
func SystemLanguage() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		v := os.Getenv(key)
		if v == "" || v == "C" || v == "POSIX" {
			continue
		}
		if lang := parseLocaleTag(v); lang != "" {
			return lang
		}
	}
	return ""
}

// DefaultLanguage picks the system locale when a profile exists for it,
// otherwise falls back to English.
func DefaultLanguage() string {
	available, err := AvailableLanguages()
	if err != nil || len(available) == 0 {
		return "en"
	}
	set := make(map[string]struct{}, len(available))
	for _, lang := range available {
		set[lang] = struct{}{}
	}
	if sys := SystemLanguage(); sys != "" {
		if _, ok := set[sys]; ok {
			return sys
		}
	}
	if _, ok := set["en"]; ok {
		return "en"
	}
	return available[0]
}

// AvailableLanguages returns sorted language codes with a built-in or user profile.
func AvailableLanguages() ([]string, error) {
	entries, err := List()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		seen[e.Language] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for lang := range seen {
		out = append(out, lang)
	}
	sort.Strings(out)
	return out, nil
}

func parseLocaleTag(raw string) string {
	tag, err := language.Parse(raw)
	if err != nil {
		return parseLocaleFallback(raw)
	}
	base, conf := tag.Base()
	if conf == language.No {
		return parseLocaleFallback(raw)
	}
	return base.String()
}

func parseLocaleFallback(raw string) string {
	lang := strings.Split(raw, ".")[0]
	lang = strings.Split(lang, "@")[0]
	if i := strings.IndexByte(lang, '_'); i >= 0 {
		lang = lang[:i]
	}
	if i := strings.IndexByte(lang, '-'); i >= 0 {
		lang = lang[:i]
	}
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" || lang == "c" || lang == "posix" {
		return ""
	}
	return lang
}
