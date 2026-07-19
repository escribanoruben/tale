package wordprepare

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ruben/tales_vault_system/internal/normalize"
)

type wordEntry struct {
	word   string
	pos    string
	rank   int
	weight int
}

var (
	wordREEn = regexp.MustCompile(`^[a-z]+$`)
	wordREES = regexp.MustCompile(`^[a-záéíóúüñ]+$`)
	esSuffix = regexp.MustCompile(`(ito|ita|illo|illa|ote|ota)$`)
)

var englishSources = map[string]string{
	"nouns":      "https://raw.githubusercontent.com/david47k/top-english-wordlists/master/top_english_nouns_lower_20000.txt",
	"adjectives": "https://raw.githubusercontent.com/david47k/top-english-wordlists/master/top_english_adjs_lower_20000.txt",
	"verbs":      "https://raw.githubusercontent.com/david47k/top-english-wordlists/master/top_english_verbs_lower_20000.txt",
}

const spanishFrequencyCSV = "https://raw.githubusercontent.com/doozan/spanish_data/master/frequency.csv"

var spanishPOSMap = map[string]string{
	"n":   "noun",
	"adj": "adjective",
	"v":   "verb",
}

func download(url string) (string, error) {
	fmt.Printf("  downloading %s\n", url)
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func normalizedInitial(word string) string {
	return normalize.Initial(word)
}

func isValidWord(word, language string) bool {
	word = strings.TrimSpace(strings.ToLower(word))
	if len(word) < 3 || len(word) > 14 {
		return false
	}
	if language == "en" && len(word) == 3 && strings.HasSuffix(word, ".") {
		return false
	}
	if strings.ContainsAny(word, " -/.") {
		return false
	}
	if language == "en" {
		return wordREEn.MatchString(word)
	}
	return wordREES.MatchString(word)
}

func englishLemmaKey(word, pos string) string {
	w := strings.ToLower(word)
	switch pos {
	case "noun":
		if len(w) > 4 && strings.HasSuffix(w, "ies") {
			return w[:len(w)-3] + "y"
		}
		if len(w) > 3 && strings.HasSuffix(w, "s") &&
			!strings.HasSuffix(w, "ss") && !strings.HasSuffix(w, "us") &&
			!strings.HasSuffix(w, "is") && !strings.HasSuffix(w, "as") &&
			!strings.HasSuffix(w, "os") {
			return w[:len(w)-1]
		}
	case "verb":
		if len(w) > 5 && strings.HasSuffix(w, "ing") {
			base := w[:len(w)-3]
			if len(base) > 2 && base[len(base)-1] == base[len(base)-2] {
				base = base[:len(base)-1]
			}
			return base
		}
		if len(w) > 4 && strings.HasSuffix(w, "ied") {
			return w[:len(w)-3] + "y"
		}
		if len(w) > 3 && strings.HasSuffix(w, "ed") {
			base := w[:len(w)-2]
			if len(base) > 2 && base[len(base)-1] == base[len(base)-2] {
				base = base[:len(base)-1]
			}
			return base
		}
		if len(w) > 4 && strings.HasSuffix(w, "es") {
			return w[:len(w)-2]
		}
		if len(w) > 3 && strings.HasSuffix(w, "s") {
			return w[:len(w)-1]
		}
	case "adjective":
		if len(w) > 5 && strings.HasSuffix(w, "iest") {
			return w[:len(w)-4] + "y"
		}
		if len(w) > 4 && strings.HasSuffix(w, "ier") {
			return w[:len(w)-3] + "y"
		}
		if len(w) > 4 && strings.HasSuffix(w, "est") {
			return w[:len(w)-3]
		}
		if len(w) > 3 && strings.HasSuffix(w, "er") {
			return w[:len(w)-2]
		}
	}
	return w
}

func spanishIsCommonEnough(word string, weight, rank int) bool {
	if esSuffix.MatchString(word) {
		return false
	}
	if weight >= minSpanishFrequency {
		return true
	}
	if rank <= 400 {
		return true
	}
	return weight >= 500
}

func loadEnglish() (map[string][]wordEntry, error) {
	buckets := map[string][]wordEntry{
		"noun":       {},
		"adjective":  {},
		"verb":       {},
	}
	posNames := map[string]string{
		"nouns":      "noun",
		"adjectives": "adjective",
		"verbs":      "verb",
	}

	for posKey, url := range englishSources {
		posName := posNames[posKey]
		text, err := download(url)
		if err != nil {
			return nil, err
		}

		rank := 0
		seenWords := make(map[string]struct{})
		seenLemmas := make(map[string]struct{})

		scanner := bufio.NewScanner(strings.NewReader(text))
		for scanner.Scan() {
			word := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if word == "" {
				continue
			}
			if _, ok := seenWords[word]; ok {
				continue
			}
			if !isValidWord(word, "en") {
				continue
			}
			if isBlocked(word, "en") {
				continue
			}
			if normalizedInitial(word) == "" {
				continue
			}
			lemma := englishLemmaKey(word, posName)
			if _, ok := seenLemmas[lemma]; ok {
				continue
			}
			rank++
			seenWords[word] = struct{}{}
			seenLemmas[lemma] = struct{}{}
			buckets[posName] = append(buckets[posName], wordEntry{
				word:   word,
				pos:    posName,
				rank:   rank,
				weight: 20000 - rank,
			})
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return buckets, nil
}

func loadSpanish() (map[string][]wordEntry, error) {
	buckets := map[string][]wordEntry{
		"noun":       {},
		"adjective":  {},
		"verb":       {},
	}

	text, err := download(spanishFrequencyCSV)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return buckets, nil
	}

	header := make(map[string]int)
	for i, name := range records[0] {
		header[name] = i
	}

	rankByPOS := map[string]int{}
	seenByPOS := map[string]map[string]struct{}{
		"noun":      {},
		"adjective": {},
		"verb":      {},
	}

	for _, row := range records[1:] {
		posIdx, ok := header["pos"]
		if !ok || posIdx >= len(row) {
			continue
		}
		posName, ok := spanishPOSMap[strings.TrimSpace(row[posIdx])]
		if !ok {
			continue
		}

		spanishIdx, ok := header["spanish"]
		if !ok || spanishIdx >= len(row) {
			continue
		}
		word := strings.TrimSpace(strings.ToLower(row[spanishIdx]))
		if word == "" {
			continue
		}
		if _, ok := seenByPOS[posName][word]; ok {
			continue
		}
		if !isValidWord(word, "es") {
			continue
		}
		if isBlocked(word, "es") {
			continue
		}
		if normalizedInitial(word) == "" {
			continue
		}

		weight := 1
		if countIdx, ok := header["count"]; ok && countIdx < len(row) {
			if _, err := fmt.Sscanf(strings.TrimSpace(row[countIdx]), "%d", &weight); err != nil {
				weight = 1
			}
		}

		rankByPOS[posName]++
		rank := rankByPOS[posName]
		if !spanishIsCommonEnough(word, weight, rank) {
			continue
		}
		seenByPOS[posName][word] = struct{}{}
		buckets[posName] = append(buckets[posName], wordEntry{
			word:   word,
			pos:    posName,
			rank:   rank,
			weight: weight,
		})
	}

	return buckets, nil
}
