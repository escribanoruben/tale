package normalize

import "testing"

func TestInitial(t *testing.T) {
	tests := []struct {
		word string
		want string
	}{
		{"árbol", "a"},
		{"época", "e"},
		{"niño", "n"},
		{"Hierro", "h"},
		{"  mango ", "m"},
		{"", ""},
		{"ñu", ""},
	}
	for _, tc := range tests {
		got := Initial(tc.word)
		if got != tc.want {
			t.Errorf("Initial(%q) = %q, want %q", tc.word, got, tc.want)
		}
	}
}

func TestKey(t *testing.T) {
	if Key("  Mango ") != "mango" {
		t.Fatalf("Key normalization failed: %q", Key("  Mango "))
	}
}
