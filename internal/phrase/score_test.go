package phrase

import (
	"testing"

	"github.com/escribanoruben/tale/internal/dictionary"
)

func TestScorePhraseUsesFrequencyRank(t *testing.T) {
	buckets := []*dictionary.Bucket{
		mustBucket(t, "alpha\t1\nzebra\t2\n"),
		mustBucket(t, "beta\t1\nzulu\t2\n"),
		mustBucket(t, "gamma\t1\nzeta\t2\n"),
		mustBucket(t, "delta\t1\nzinc\t2\n"),
	}

	frequent := scorePhrase([]string{"alpha", "beta", "gamma", "delta"}, buckets)
	rare := scorePhrase([]string{"zebra", "zulu", "zeta", "zinc"}, buckets)
	if frequent <= rare {
		t.Fatalf("frequent score %d should beat rare score %d", frequent, rare)
	}
}

func mustBucket(t *testing.T, data string) *dictionary.Bucket {
	t.Helper()
	bucket, err := dictionary.ParseBucketData(data)
	if err != nil {
		t.Fatal(err)
	}
	return bucket
}
