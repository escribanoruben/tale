package worddata

import (
	"bufio"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

var posFiles = map[string]string{
	"noun":      "nouns.tsv",
	"adjective": "adjectives.tsv",
	"verb":      "verbs.tsv",
}

var languageLabels = map[string]string{
	"en": "English",
	"es": "Spanish",
}

// LanguageStats holds aggregated counts for one language.
type LanguageStats struct {
	Code        string
	Label       string
	Alphabet    string
	Letters     []string
	TotalsByPOS map[string]int
	Buckets     map[string]map[string]int
	TotalWords  int
	MinBucket   int
	MaxBucket   int
	AvgBucket   float64
}

// Report is the full statistics report for embedded word data.
type Report struct {
	Languages map[string]LanguageStats
}

// Stats scans embedded word lists and returns usage statistics.
func Stats() (Report, error) {
	report := Report{Languages: make(map[string]LanguageStats)}

	for code, label := range languageLabels {
		langFS, err := fs.Sub(Words, code)
		if err != nil {
			return Report{}, fmt.Errorf("language %s: %w", code, err)
		}

		letters, err := listLetters(langFS)
		if err != nil {
			return Report{}, fmt.Errorf("language %s letters: %w", code, err)
		}

		buckets := make(map[string]map[string]int, len(letters))
		totalsByPOS := map[string]int{"noun": 0, "adjective": 0, "verb": 0}
		allCounts := make([]int, 0, len(letters)*len(posFiles))

		for _, letter := range letters {
			buckets[letter] = make(map[string]int, len(posFiles))
			for pos, filename := range posFiles {
				count, err := countWords(langFS, letter+"/"+filename)
				if err != nil {
					return Report{}, fmt.Errorf("%s/%s/%s: %w", code, letter, filename, err)
				}
				buckets[letter][pos] = count
				totalsByPOS[pos] += count
				allCounts = append(allCounts, count)
			}
		}

		minBucket, maxBucket := 0, 0
		avgBucket := 0.0
		if len(allCounts) > 0 {
			minBucket = allCounts[0]
			maxBucket = allCounts[0]
			sum := 0
			for _, count := range allCounts {
				if count < minBucket {
					minBucket = count
				}
				if count > maxBucket {
					maxBucket = count
				}
				sum += count
			}
			avgBucket = float64(sum) / float64(len(allCounts))
		}

		report.Languages[code] = LanguageStats{
			Code:        code,
			Label:       label,
			Alphabet:    strings.Join(letters, ""),
			Letters:     letters,
			TotalsByPOS: totalsByPOS,
			Buckets:     buckets,
			TotalWords:  totalsByPOS["noun"] + totalsByPOS["adjective"] + totalsByPOS["verb"],
			MinBucket:   minBucket,
			MaxBucket:   maxBucket,
			AvgBucket:   avgBucket,
		}
	}

	return report, nil
}

func listLetters(langFS fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(langFS, ".")
	if err != nil {
		return nil, err
	}
	letters := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 1 {
			letters = append(letters, entry.Name())
		}
	}
	sort.Strings(letters)
	return letters, nil
}

func countWords(langFS fs.FS, path string) (int, error) {
	file, err := langFS.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count, scanner.Err()
}

// FormatReport renders a human-readable statistics report.
func FormatReport(report Report, minThreshold int) string {
	var b strings.Builder

	codes := make([]string, 0, len(report.Languages))
	for code := range report.Languages {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	grandTotal := 0
	for _, code := range codes {
		info := report.Languages[code]
		grandTotal += info.TotalWords

		fmt.Fprintf(&b, "=== %s (%s) ===\n", info.Label, info.Code)
		fmt.Fprintf(&b, "Alphabet (%d letters): %s\n", len(info.Letters), info.Alphabet)
		fmt.Fprintf(&b, "Total words: %s  (nouns %s, adjectives %s, verbs %s)\n",
			formatInt(info.TotalWords),
			formatInt(info.TotalsByPOS["noun"]),
			formatInt(info.TotalsByPOS["adjective"]),
			formatInt(info.TotalsByPOS["verb"]),
		)
		fmt.Fprintf(&b, "Per bucket: min %d, max %d, avg %.1f\n\n",
			info.MinBucket, info.MaxBucket, info.AvgBucket)

		fmt.Fprintf(&b, "%-8s%6s%6s%6s%6s\n", "letter", "nou", "adj", "ver", "sum")
		fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 32))
		for _, letter := range info.Letters {
			row := info.Buckets[letter]
			sum := row["noun"] + row["adjective"] + row["verb"]
			fmt.Fprintf(&b, "%-8s%6d%6d%6d%6d\n",
				letter, row["noun"], row["adjective"], row["verb"], sum)
		}
		colNouns, colAdj, colVer := 0, 0, 0
		for _, letter := range info.Letters {
			colNouns += info.Buckets[letter]["noun"]
			colAdj += info.Buckets[letter]["adjective"]
			colVer += info.Buckets[letter]["verb"]
		}
		fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 32))
		fmt.Fprintf(&b, "%-8s%6d%6d%6d%6d\n\n", "total", colNouns, colAdj, colVer, info.TotalWords)

		var low []string
		for _, letter := range info.Letters {
			for pos := range posFiles {
				count := info.Buckets[letter][pos]
				if count < minThreshold {
					low = append(low, fmt.Sprintf("%s/%s: %d", letter, pos, count))
				}
			}
		}
		if len(low) > 0 {
			fmt.Fprintf(&b, "Buckets below %d words:\n", minThreshold)
			for _, item := range low {
				fmt.Fprintf(&b, "  - %s\n", item)
			}
			b.WriteByte('\n')
		}
	}

	if len(codes) > 1 {
		fmt.Fprintf(&b, "Grand total: %s words across %d languages\n", formatInt(grandTotal), len(codes))
	}

	return b.String()
}

func formatInt(value int) string {
	s := fmt.Sprintf("%d", value)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append([]string{s}, parts...)
	}
	return strings.Join(parts, ",")
}
