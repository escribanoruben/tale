package wordprepare

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	minWordsPerBucket   = 8
	maxWordsPerBucket   = 120
	minSpanishFrequency = 2000
)

// Builtin alphabet sizes: curated from corpus frequency per language.
var topLetterCountByLanguage = map[string]int{
	"en": 22,
	"es": 20,
}

var posFiles = map[string]string{
	"noun":       "nouns.tsv",
	"adjective":  "adjectives.tsv",
	"verb":       "verbs.tsv",
}

var posOrder = []string{"noun", "adjective", "verb"}

// Options configures word list preparation.
type Options struct {
	RepoRoot string
}

// LanguageResult summarizes one prepared language.
type LanguageResult struct {
	Language  string
	ProfileID string
	Letters   []string
	Counts    map[bucketKey]int
}

type bucketKey struct {
	pos    string
	letter string
}

// Prepare downloads sources, curates word lists, writes profiles, and packages archives.
func Prepare(opts Options) ([]LanguageResult, []string, error) {
	root, err := resolveRepoRoot(opts.RepoRoot)
	if err != nil {
		return nil, nil, err
	}

	dataDir := filepath.Join(root, "internal", "worddata")
	profilesDir := filepath.Join(root, "internal", "profile")
	distDir := filepath.Join(root, "dist")

	var results []LanguageResult
	var allIssues []string

	loaders := []struct {
		language string
		load     func() (map[string][]wordEntry, error)
	}{
		{"en", loadEnglish},
		{"es", loadSpanish},
	}

	for _, item := range loaders {
		fmt.Printf("\nProcessing %s...\n", item.language)
		buckets, err := item.load()
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", item.language, err)
		}

		topN := topLetterCountByLanguage[item.language]
		if topN <= 0 {
			topN = 20
		}
		letters := computeTopLetters(buckets, topN)
		fmt.Printf("  top %d initials by corpus weight: %s\n", topN, strings.Join(letters, ""))

		trimmed := trimToBucketLimits(buckets, letters)
		counts := bucketCounts(trimmed, letters)

		if err := writeLanguageLayout(dataDir, item.language, letters, trimmed); err != nil {
			return nil, nil, err
		}

		profileID := item.language
		if err := writeProfile(profilesDir, item.language, profileID, letters); err != nil {
			return nil, nil, err
		}

		issues := printCoverage(item.language, letters, counts)
		allIssues = append(allIssues, issues...)
		results = append(results, LanguageResult{
			Language:  item.language,
			ProfileID: profileID,
			Letters:   letters,
			Counts:    counts,
		})
	}

	fmt.Println("\nPackaging language archives...")
	if err := packageLanguageArchives(dataDir, distDir, results); err != nil {
		return nil, nil, err
	}

	fmt.Println("\nDone.")
	fmt.Printf("Data written to %s\n", dataDir)
	fmt.Printf("Profiles written to %s\n", profilesDir)
	fmt.Printf("Archives written to %s\n", distDir)

	return results, allIssues, nil
}

func resolveRepoRoot(arg string) (string, error) {
	if arg != "" {
		return filepath.Abs(arg)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root (no go.mod); use --repo")
		}
		dir = parent
	}
}

func computeTopLetters(buckets map[string][]wordEntry, n int) []string {
	counts := make(map[string]int)
	for _, entries := range buckets {
		for _, entry := range entries {
			initial := normalizedInitial(entry.word)
			if initial != "" {
				counts[initial] += entry.weight
			}
		}
	}

	type pair struct {
		letter string
		weight int
	}
	pairs := make([]pair, 0, len(counts))
	for letter, weight := range counts {
		pairs = append(pairs, pair{letter, weight})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].weight != pairs[j].weight {
			return pairs[i].weight > pairs[j].weight
		}
		return pairs[i].letter < pairs[j].letter
	})

	letters := make([]string, 0, n)
	for i := 0; i < len(pairs) && i < n; i++ {
		letters = append(letters, pairs[i].letter)
	}
	sort.Strings(letters)
	return letters
}

func trimToBucketLimits(buckets map[string][]wordEntry, letters []string) map[bucketKey][]wordEntry {
	letterSet := make(map[string]struct{}, len(letters))
	for _, letter := range letters {
		letterSet[letter] = struct{}{}
	}

	byBucket := make(map[bucketKey][]wordEntry)
	for pos, entries := range buckets {
		for _, entry := range entries {
			initial := normalizedInitial(entry.word)
			if _, ok := letterSet[initial]; ok {
				key := bucketKey{pos: pos, letter: initial}
				byBucket[key] = append(byBucket[key], entry)
			}
		}
	}

	trimmed := make(map[bucketKey][]wordEntry)
	for _, pos := range posOrder {
		for _, letter := range letters {
			key := bucketKey{pos: pos, letter: letter}
			items := append([]wordEntry(nil), byBucket[key]...)
			sort.Slice(items, func(i, j int) bool {
				return items[i].rank < items[j].rank
			})
			if len(items) > maxWordsPerBucket {
				items = items[:maxWordsPerBucket]
			}
			trimmed[key] = items
		}
	}
	return trimmed
}

func bucketCounts(trimmed map[bucketKey][]wordEntry, letters []string) map[bucketKey]int {
	counts := make(map[bucketKey]int)
	for _, pos := range posOrder {
		for _, letter := range letters {
			key := bucketKey{pos: pos, letter: letter}
			counts[key] = len(trimmed[key])
		}
	}
	return counts
}

func cleanupLanguageDir(langDir string, letters []string) error {
	letterSet := make(map[string]struct{}, len(letters))
	for _, letter := range letters {
		letterSet[letter] = struct{}{}
	}

	entries, err := os.ReadDir(langDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(langDir, entry.Name())
		if entry.IsDir() {
			if _, ok := letterSet[entry.Name()]; !ok {
				if err := os.RemoveAll(path); err != nil {
					return err
				}
			}
			continue
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func writeWordFile(path string, entries []wordEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	type ranked struct {
		word string
		rank int
	}
	rows := make([]ranked, len(entries))
	for i, entry := range entries {
		rows[i] = ranked{word: entry.word, rank: i + 1}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].word < rows[j].word
	})
	var b strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&b, "%s\t%d\n", row.word, row.rank)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeLanguageLayout(dataDir, language string, letters []string, trimmed map[bucketKey][]wordEntry) error {
	langDir := filepath.Join(dataDir, language)
	if err := cleanupLanguageDir(langDir, letters); err != nil {
		return err
	}
	for _, letter := range letters {
		letterDir := filepath.Join(langDir, letter)
		for pos, filename := range posFiles {
			key := bucketKey{pos: pos, letter: letter}
			if err := writeWordFile(filepath.Join(letterDir, filename), trimmed[key]); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeProfile(profilesDir, language, profileID string, letters []string) error {
	display := map[string]string{
		"en": "English (default)",
		"es": "Spanish (default)",
	}[language]
	content := fmt.Sprintf(`id = %q
language = %q
display_name = %q
password_length = 20
block_size = 4
alphabet = %q
phrase_template = ["noun", "adjective", "verb", "noun"]
data_layout = "letter"

[normalization]
strip_accents = true
lowercase = true
exclude_initial_enye = true
`, profileID, language, display, strings.Join(letters, ""))

	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(profilesDir, profileID+".toml"), []byte(content), 0o644)
}

func printCoverage(language string, letters []string, counts map[bucketKey]int) []string {
	var issues []string
	fmt.Printf("\n%s coverage (%d letters: %s)\n", strings.ToUpper(language), len(letters), strings.Join(letters, ""))
	fmt.Printf("%-12s ", "POS")
	for _, letter := range letters {
		fmt.Printf("%4s", letter)
	}
	fmt.Println()

	for _, pos := range posOrder {
		fmt.Printf("%-12s ", pos)
		for _, letter := range letters {
			count := counts[bucketKey{pos: pos, letter: letter}]
			fmt.Printf("%4d", count)
			if count < minWordsPerBucket {
				issues = append(issues, fmt.Sprintf("%s %s '%s': %d words (< %d)", language, pos, letter, count, minWordsPerBucket))
			}
		}
		fmt.Println()
	}

	total := 0
	for _, pos := range posOrder {
		for _, letter := range letters {
			total += counts[bucketKey{pos: pos, letter: letter}]
		}
	}
	fmt.Printf("Total words: %d\n", total)
	if len(issues) > 0 {
		fmt.Println("Warnings:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	} else {
		fmt.Println("All buckets meet minimum word count.")
	}
	return issues
}

func packageLanguageArchives(dataDir, distDir string, results []LanguageResult) error {
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return err
	}
	for _, result := range results {
		source := filepath.Join(dataDir, result.Language)
		archive := filepath.Join(distDir, "data-"+result.Language+".zip")
		if err := os.Remove(archive); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := zipDirectory(archive, source, result.Language); err != nil {
			return err
		}
		fmt.Printf("  packaged %s\n", archive)
	}
	return nil
}

func zipDirectory(archivePath, sourceDir, baseName string) error {
	out, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(filepath.Join(baseName, rel))

		if d.IsDir() {
			_, err := zw.Create(rel + "/")
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = rel
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
