package dictionary

import (
	"testing"
)

func TestParseBucketAlphabeticalWithRanks(t *testing.T) {
	data := "zebra\t2\nalpha\t1\nmiddle\t3\n"
	bucket, err := parseBucket(data)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"alpha", "middle", "zebra"}
	for i, word := range want {
		if bucket.Words[i] != word {
			t.Fatalf("Words[%d] = %q, want %q", i, bucket.Words[i], word)
		}
	}
	if bucket.Rank("alpha") != 1 {
		t.Fatalf("alpha rank = %d, want 1", bucket.Rank("alpha"))
	}
	if bucket.Rank("zebra") != 2 {
		t.Fatalf("zebra rank = %d, want 2", bucket.Rank("zebra"))
	}
}

func TestTopByFrequency(t *testing.T) {
	data := "zebra\t4\nalpha\t1\nbravo\t2\ncharlie\t3\n"
	bucket, err := parseBucket(data)
	if err != nil {
		t.Fatal(err)
	}
	top := bucket.TopByFrequency(2, nil)
	if len(top) != 2 {
		t.Fatalf("got %d words, want 2", len(top))
	}
	if top[0] != "alpha" || top[1] != "bravo" {
		t.Fatalf("top by frequency = %v, want [alpha bravo]", top)
	}
}

func TestParseBucketLegacySingleColumn(t *testing.T) {
	data := "rare\ncommon\n"
	bucket, err := parseBucket(data)
	if err != nil {
		t.Fatal(err)
	}
	if bucket.Rank("rare") != 1 || bucket.Rank("common") != 2 {
		t.Fatalf("line order ranks: rare=%d common=%d", bucket.Rank("rare"), bucket.Rank("common"))
	}
	if bucket.Words[0] != "common" || bucket.Words[1] != "rare" {
		t.Fatalf("alphabetical words = %v", bucket.Words)
	}
}
