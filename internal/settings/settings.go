package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/escribanoruben/tales_vault_system/internal/paths"
	"github.com/escribanoruben/tales_vault_system/internal/profile"
)

const defaultPhrases = 5

// Settings holds user defaults stored in ~/.tale/settings.toml.
type Settings struct {
	Language             string `toml:"language"`
	Phrases              int    `toml:"phrases"`
	AllowAdjacentRepeats bool   `toml:"allow_adjacent_repeats"`
}

// Default returns built-in defaults (language unset until the user chooses).
func Default() Settings {
	return Settings{Phrases: defaultPhrases}
}

// Configured reports whether the user has chosen a language.
func (s Settings) Configured() bool {
	return s.Language != ""
}

// PhrasesOr returns phrases, falling back to defaultPhrases.
func (s Settings) PhrasesOr(fallback int) int {
	if s.Phrases > 0 {
		return s.Phrases
	}
	if fallback > 0 {
		return fallback
	}
	return defaultPhrases
}

// Resolve fills unset fields for batch or interactive use.
func (s Settings) Resolve(forBatch bool) Settings {
	out := s
	if out.Phrases <= 0 {
		out.Phrases = defaultPhrases
	}
	if out.Language == "" && forBatch {
		out.Language = profile.DefaultLanguage()
	}
	return out
}

// Path returns ~/.tale/settings.toml.
func Path() (string, error) {
	dir, err := paths.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.toml"), nil
}

// Load reads settings from disk, or returns defaults if missing.
func Load() (Settings, error) {
	path, err := Path()
	if err != nil {
		return Default(), err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Default(), err
	}
	var s Settings
	if err := toml.Unmarshal(data, &s); err != nil {
		return Default(), fmt.Errorf("parse %s: %w", path, err)
	}
	if s.Phrases <= 0 {
		s.Phrases = defaultPhrases
	}
	return s, nil
}

// Save writes settings to ~/.tale/settings.toml.
func Save(s Settings) error {
	if s.Phrases <= 0 {
		s.Phrases = defaultPhrases
	}
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
