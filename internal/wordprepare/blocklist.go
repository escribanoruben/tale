package wordprepare

import (
	"embed"
	"strings"
)

//go:embed blocklist/*.txt
var blocklistFiles embed.FS

var blocklists = mustLoadBlocklists()

func mustLoadBlocklists() map[string]map[string]struct{} {
	out := make(map[string]map[string]struct{})
	for _, lang := range []string{"en", "es"} {
		data, err := blocklistFiles.ReadFile("blocklist/" + lang + ".txt")
		if err != nil {
			panic("wordprepare: load blocklist " + lang + ": " + err.Error())
		}
		set := make(map[string]struct{})
		for line := range strings.Lines(string(data)) {
			word := strings.TrimSpace(line)
			if word == "" || strings.HasPrefix(word, "#") {
				continue
			}
			set[word] = struct{}{}
		}
		out[lang] = set
	}
	return out
}

func isBlocked(word, language string) bool {
	set, ok := blocklists[language]
	if !ok {
		return false
	}
	_, ok = set[strings.ToLower(strings.TrimSpace(word))]
	return ok
}
