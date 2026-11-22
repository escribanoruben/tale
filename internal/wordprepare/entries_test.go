package wordprepare

import "testing"

func TestEnglishLemmaKey(t *testing.T) {
	cases := []struct {
		word, pos, want string
	}{
		{"states", "noun", "state"},
		{"studies", "noun", "study"},
		{"running", "verb", "run"},
		{"happier", "adjective", "happy"},
	}
	for _, tc := range cases {
		if got := englishLemmaKey(tc.word, tc.pos); got != tc.want {
			t.Errorf("englishLemmaKey(%q, %q) = %q, want %q", tc.word, tc.pos, got, tc.want)
		}
	}
}

func TestIsValidWord(t *testing.T) {
	if !isValidWord("hello", "en") {
		t.Fatal("expected hello to be valid")
	}
	if isValidWord("ab", "en") {
		t.Fatal("expected short word to be invalid")
	}
	if isValidWord("hello-world", "en") {
		t.Fatal("expected hyphenated word to be invalid")
	}
	if !isValidWord("árbol", "es") {
		t.Fatal("expected accented spanish word to be valid")
	}
}

func TestIsBlocked(t *testing.T) {
	if !isBlocked("fuck", "en") {
		t.Fatal("expected fuck to be blocked")
	}
	if isBlocked("hello", "en") {
		t.Fatal("expected hello not to be blocked")
	}
	if !isBlocked("puta", "es") {
		t.Fatal("expected puta to be blocked")
	}
	if isBlocked("casa", "es") {
		t.Fatal("expected casa not to be blocked")
	}
	// suitability blocklist (sexual aggression, gore, morbidity)
	for _, w := range []string{"rapist", "murder", "corpse", "torture", "nazi", "hiv"} {
		if !isBlocked(w, "en") {
			t.Errorf("expected %q to be blocked", w)
		}
	}
	for _, w := range []string{"violador", "fornicar", "asesinato", "torturar", "cadáver", "sida"} {
		if !isBlocked(w, "es") {
			t.Errorf("expected %q to be blocked", w)
		}
	}
}

func TestSpanishIsCommonEnough(t *testing.T) {
	if spanishIsCommonEnough("perrito", 5000, 10) {
		t.Fatal("expected diminutive to be excluded")
	}
	if !spanishIsCommonEnough("casa", 5000, 10) {
		t.Fatal("expected frequent word to be included")
	}
}
