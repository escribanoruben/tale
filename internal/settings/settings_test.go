package settings_test

import (
	"path/filepath"
	"testing"

	"github.com/ruben/tales_vault_system/internal/settings"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	in := settings.Settings{
		Language:             "es",
		Phrases:              4,
		AllowAdjacentRepeats: true,
	}
	if err := settings.Save(in); err != nil {
		t.Fatal(err)
	}

	out, err := settings.Load()
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("got %+v, want %+v", out, in)
	}

	path, err := settings.Path()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "settings.toml" {
		t.Fatalf("unexpected path %s", path)
	}
}

func TestLoadMissingUsesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	s, err := settings.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.Configured() {
		t.Fatal("expected unconfigured defaults")
	}
	if s.PhrasesOr(0) != 5 {
		t.Fatalf("phrases = %d, want 5", s.PhrasesOr(0))
	}
}

func TestResolveForBatch(t *testing.T) {
	s := settings.Default().Resolve(true)
	if s.Language == "" {
		t.Fatal("expected language for batch")
	}
	if s.Phrases != 5 {
		t.Fatalf("phrases = %d, want 5", s.Phrases)
	}
}
