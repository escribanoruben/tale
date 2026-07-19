package profile

import "testing"

func TestBuiltinPhraseTemplates(t *testing.T) {
	tests := []struct {
		language string
		want     []string
	}{
		{
			language: "en",
			want:     []string{"adjective", "noun", "verb", "noun"},
		},
		{
			language: "es",
			want:     []string{"noun", "adjective", "verb", "noun"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.language, func(t *testing.T) {
			p, err := Load(Files, tc.language)
			if err != nil {
				t.Fatal(err)
			}
			if len(p.PhraseTemplate) != len(tc.want) {
				t.Fatalf("template len %d, want %d: %v", len(p.PhraseTemplate), len(tc.want), p.PhraseTemplate)
			}
			for i := range tc.want {
				if p.PhraseTemplate[i] != tc.want[i] {
					t.Fatalf("slot %d = %q, want %q (full: %v)", i, p.PhraseTemplate[i], tc.want[i], p.PhraseTemplate)
				}
			}
		})
	}
}
