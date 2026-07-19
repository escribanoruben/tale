package profile

import "testing"

func TestSystemLanguage(t *testing.T) {
	tests := []struct {
		env  string
		want string
	}{
		{"es_ES.UTF-8", "es"},
		{"en_US", "en"},
		{"C", ""},
		{"POSIX", ""},
	}
	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv("LC_ALL", "")
			t.Setenv("LC_MESSAGES", "")
			t.Setenv("LANG", tc.env)
			if got := SystemLanguage(); got != tc.want {
				t.Fatalf("SystemLanguage() = %q, want %q", got, tc.want)
			}
		})
	}
}
