package password

import (
	"strings"
	"testing"
)

func TestGeneratePreventsAdjacentRepeatsByDefault(t *testing.T) {
	alphabet := "abcd"
	for i := 0; i < 50; i++ {
		pass, err := Generate(alphabet, 20, DefaultOptions())
		if err != nil {
			t.Fatal(err)
		}
		if HasAdjacentRepeats(pass) {
			t.Fatalf("password %q has adjacent repeats", pass)
		}
	}
}

func TestGenerateAfterRespectsPreviousRune(t *testing.T) {
	alphabet := "ab"
	opts := DefaultOptions()
	for i := 0; i < 50; i++ {
		seg, err := GenerateAfter(alphabet, 4, 'a', opts)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(seg, "a") {
			t.Fatalf("segment %q repeats leading a", seg)
		}
		if HasAdjacentRepeats(seg) {
			t.Fatalf("segment %q has internal adjacent repeats", seg)
		}
	}
}

func TestGenerateAllowsAdjacentRepeatsWhenDisabled(t *testing.T) {
	alphabet := "a"
	opts := Options{PreventAdjacentRepeats: false}
	pass, err := Generate(alphabet, 5, opts)
	if err != nil {
		t.Fatal(err)
	}
	if pass != "aaaaa" {
		t.Fatalf("password = %q, want aaaaa", pass)
	}
}

func TestGenerateAdjacentRepeatsRequiresAlphabetSize(t *testing.T) {
	_, err := Generate("a", 2, DefaultOptions())
	if err == nil {
		t.Fatal("expected error for single-letter alphabet")
	}
}
