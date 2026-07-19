package worddata

import (
	"bufio"
	"io/fs"
	"strings"
	"testing"
)

func TestWordListsAlphabetical(t *testing.T) {
	err := fs.WalkDir(Words, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tsv") {
			return nil
		}

		data, err := Words.ReadFile(path)
		if err != nil {
			return err
		}

		var prev string
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			word := line
			if tab := strings.IndexByte(line, '\t'); tab >= 0 {
				word = strings.TrimSpace(line[:tab])
			}
			if prev != "" && word < prev {
				t.Errorf("%s: %q out of order after %q", path, word, prev)
			}
			prev = word
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
}
