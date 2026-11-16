// Command prepare-embeddings regenerates the embedded semantic vector
// assets (internal/worddata/{lang}/embeddings.bin) from fastText Common
// Crawl vectors restricted to tale's embedded vocabulary.
//
// Maintainer-only and networked (like cmd/prepare-wordlists); runtime
// generation never calls this. Vectors: cc.{lang}.300.vec.gz from
// https://dl.fbaipublicfiles.com/fasttext/vectors-crawl/ (CC-BY-SA 3.0 —
// see NOTICE). Words without vectors are skipped and reported; semantic
// ranking simply ignores them.
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/escribanoruben/tale/internal/worddata"
)

const vecURL = "https://dl.fbaipublicfiles.com/fasttext/vectors-crawl/%s"

func main() {
	langs := flag.String("lang", "en,es", "comma-separated language codes")
	vecDir := flag.String("vec-dir", "", "directory with pre-downloaded cc.{lang}.300.vec.gz (downloads to a temp dir otherwise)")
	flag.Parse()

	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}
	for _, lang := range strings.Split(*langs, ",") {
		lang = strings.TrimSpace(lang)
		if lang == "" {
			continue
		}
		if err := prepare(root, lang, *vecDir); err != nil {
			fatal(fmt.Errorf("%s: %w", lang, err))
		}
	}
}

func prepare(root, lang, vecDir string) error {
	vocab := collectVocab(filepath.Join(root, "internal", "worddata", lang))
	if len(vocab) == 0 {
		return fmt.Errorf("no embedded words found under internal/worddata/%s", lang)
	}
	fmt.Printf("%s: %d unique words in vocabulary\n", lang, len(vocab))

	src, cleanup, err := openVectors(lang, vecDir)
	if err != nil {
		return err
	}
	defer cleanup()

	gz, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	vectors, dim, err := extract(gz, vocab)
	if err != nil {
		return err
	}
	if len(vectors) == 0 {
		return fmt.Errorf("no vocabulary words found in vectors")
	}
	words := make([]string, 0, len(vectors))
	for w := range vectors {
		words = append(words, w)
	}
	sort.Strings(words)
	rows := make([][]float32, len(words))
	for i, w := range words {
		rows[i] = vectors[w]
	}
	fmt.Printf("%s: extracted %d/%d vectors (%.1f%%), dim %d\n",
		lang, len(words), len(vocab), 100*float64(len(words))/float64(len(vocab)), dim)

	data, err := worddata.EncodeEmbeddings(dim, words, rows)
	if err != nil {
		return err
	}
	out := filepath.Join(root, "internal", "worddata", lang, "embeddings.bin")
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("%s: wrote %s (%.1f MB)\n", lang, out, float64(len(data))/1e6)
	return nil
}

// openVectors returns a reader over cc.{lang}.300.vec.gz, downloading it
// when no pre-downloaded copy exists.
func openVectors(lang, vecDir string) (io.ReadCloser, func(), error) {
	name := fmt.Sprintf("cc.%s.300.vec.gz", lang)
	local := filepath.Join(vecDir, name)
	if vecDir != "" {
		f, err := os.Open(local)
		if err == nil {
			return f, func() { f.Close() }, nil
		}
		if !os.IsNotExist(err) {
			return nil, nil, err
		}
	}
	tmp := filepath.Join(os.TempDir(), name)
	if f, err := os.Open(tmp); err == nil {
		fmt.Printf("%s: using cached %s\n", lang, tmp)
		return f, func() { f.Close() }, nil
	}
	url := fmt.Sprintf(vecURL, name)
	fmt.Printf("%s: downloading %s (about 1 GB, be patient)\n", lang, url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("download %s: %s", url, resp.Status)
	}
	out, err := os.Create(tmp)
	if err != nil {
		return nil, nil, err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return nil, nil, err
	}
	if err := out.Close(); err != nil {
		return nil, nil, err
	}
	f, err := os.Open(tmp)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}

// extract streams the .vec file keeping vectors for vocabulary words.
func extract(r io.Reader, vocab map[string]bool) (map[string][]float32, int, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	if !sc.Scan() {
		return nil, 0, fmt.Errorf("empty vector file")
	}
	var dim int
	if _, err := fmt.Sscanf(sc.Text(), "%d %d", new(int), &dim); err != nil {
		return nil, 0, fmt.Errorf("bad header %q: %w", sc.Text(), err)
	}
	vectors := make(map[string][]float32, len(vocab))
	for sc.Scan() {
		line := sc.Text()
		sp := strings.IndexByte(line, ' ')
		if sp <= 0 {
			continue
		}
		word := strings.ToLower(line[:sp])
		if !vocab[word] {
			continue
		}
		if _, seen := vectors[word]; seen {
			continue
		}
		fields := strings.Fields(line[sp+1:])
		if len(fields) != dim {
			continue
		}
		vec := make([]float32, dim)
		ok := true
		for i, f := range fields {
			x, err := strconv.ParseFloat(f, 32)
			if err != nil {
				ok = false
				break
			}
			vec[i] = float32(x)
		}
		if ok {
			vectors[word] = vec
			if len(vectors) == len(vocab) {
				break
			}
		}
	}
	return vectors, dim, sc.Err()
}

// collectVocab gathers lowercase first-column words from all TSVs.
func collectVocab(dir string) map[string]bool {
	vocab := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return vocab
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		for _, pos := range []string{"nouns.tsv", "adjectives.tsv", "verbs.tsv"} {
			data, err := os.ReadFile(filepath.Join(dir, e.Name(), pos))
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(data), "\n") {
				w := strings.ToLower(strings.TrimSpace(strings.SplitN(line, "\t", 2)[0]))
				if w != "" {
					vocab[w] = true
				}
			}
		}
	}
	return vocab
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s upward", dir)
		}
		dir = parent
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "prepare-embeddings: %v\n", err)
	os.Exit(1)
}
