package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/escribanoruben/tale/internal/paths"
)

func TestResolveBuiltin(t *testing.T) {
	resolved, err := Resolve("en")
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.Builtin {
		t.Fatal("expected builtin profile")
	}
	if len(resolved.Profile.Alphabet) != 22 {
		t.Fatalf("alphabet length %d, want 22", len(resolved.Profile.Alphabet))
	}
}

func TestInitUserProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	result, err := InitUserProfile(InitOptions{
		Language: "es",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProfilePath == "" {
		t.Fatal("expected profile path")
	}
	if !result.DataCopied {
		t.Fatal("expected word data to be copied")
	}

	dataPath, err := paths.LanguageDataDir("es")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dataPath, "a", "nouns.tsv")); err != nil {
		t.Fatalf("expected copied nouns file: %v", err)
	}

	resolved, err := Resolve("es")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Builtin {
		t.Fatal("expected user profile")
	}
	if resolved.Profile.Language != "es" {
		t.Fatalf("language = %q, want es", resolved.Profile.Language)
	}
}

func TestInitUserProfileNewLanguage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	result, err := InitUserProfile(InitOptions{
		Language:     "de",
		FromLanguage: "en",
		DisplayName:  "German",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DataCopied {
		t.Fatal("expected german data copy")
	}
	if _, err := os.Stat(filepath.Join(result.DataPath, "a", "nouns.tsv")); err != nil {
		t.Fatalf("expected german seed data: %v", err)
	}
}

func TestNewUserLanguage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	result, err := NewUserLanguage(NewLanguageOptions{
		Language:    "pt",
		Alphabet:    "ab",
		DisplayName: "Portuguese",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DataScaffolded {
		t.Fatal("expected data scaffold")
	}
	if _, err := os.Stat(filepath.Join(result.DataPath, "a", "nouns.tsv")); err != nil {
		t.Fatalf("expected stub nouns: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.DataPath, "b", "verbs.tsv")); err != nil {
		t.Fatalf("expected stub verbs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.DataPath, "c")); !os.IsNotExist(err) {
		t.Fatalf("unexpected letter folder outside alphabet: %v", err)
	}

	resolved, err := Resolve("pt")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Builtin {
		t.Fatal("expected user language")
	}
	if resolved.Profile.Alphabet != "ab" {
		t.Fatalf("alphabet = %q, want ab", resolved.Profile.Alphabet)
	}
}

func TestNewUserLanguageRequiresAlphabet(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := NewUserLanguage(NewLanguageOptions{Language: "xx"})
	if err == nil {
		t.Fatal("expected error for missing alphabet")
	}
}

func TestNewUserLanguageRejectsDuplicateAlphabet(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := NewUserLanguage(NewLanguageOptions{
		Language: "xx",
		Alphabet: "aba",
	})
	if err == nil {
		t.Fatal("expected validation error for duplicate letter")
	}
}

func TestUserProfileOverridesBuiltin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := InitUserProfile(InitOptions{
		Language: "en",
	})
	if err != nil {
		t.Fatal(err)
	}

	profilesDir, err := paths.ProfilesDir()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(profilesDir, "en.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	content = stringsReplaceOnce(content, `password_length = 20`, `password_length = 16`)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	resolved, err := Resolve("en")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Builtin {
		t.Fatal("expected user override")
	}
	if resolved.Profile.PasswordLength != 16 {
		t.Fatalf("password_length = %d, want 16", resolved.Profile.PasswordLength)
	}
}

func TestDefaultLanguageUsesSystemLocale(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LANG", "es_ES.UTF-8")

	if got := DefaultLanguage(); got != "es" {
		t.Fatalf("DefaultLanguage() = %q, want es", got)
	}
}

func TestDefaultLanguageFallsBackToEnglish(t *testing.T) {
	t.Setenv("LANG", "fr_FR.UTF-8")

	if got := DefaultLanguage(); got != "en" {
		t.Fatalf("DefaultLanguage() = %q, want en", got)
	}
}

func stringsReplaceOnce(s, old, new string) string {
	i := indexOf(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
