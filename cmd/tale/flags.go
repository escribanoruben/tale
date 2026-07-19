package main

import (
	"flag"
	"fmt"

	"github.com/ruben/tales_vault_system/internal/profile"
	"github.com/ruben/tales_vault_system/internal/settings"
)

// runConfig holds merged CLI flags and user settings.
type runConfig struct {
	Language             string
	Phrases              int
	Length               int
	Password             string
	AllowRepeat          bool
	AllowAdjacentRepeats bool
	ForceSetup           bool
}

func parseRunConfig(args []string, forTUI bool) runConfig {
	base, _ := settings.Load()
	langDefault := base.Language
	if langDefault == "" && !forTUI {
		langDefault = profile.DefaultLanguage()
	}

	fs := flag.NewFlagSet("tale", flag.ExitOnError)
	language := fs.String("language", langDefault, "language code (en, es, or custom)")
	phrases := fs.Int("phrases", base.PhrasesOr(5), "number of mnemonic phrases (blocks)")
	length := fs.Int("length", 0, "password length (overrides --phrases)")
	allowRepeat := fs.Bool("allow-repeat", base.AllowRepeat, "allow reusing words across the mnemonic")
	allowAdjacent := fs.Bool("allow-adjacent-repeats", base.AllowAdjacentRepeats, "allow consecutive identical password characters")
	password := fs.String("password", "", "existing password (TUI only, generated if empty)")
	_ = fs.Parse(args)

	languageFlag := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "language" {
			languageFlag = true
		}
	})

	cfg := runConfig{
		Language:             *language,
		Phrases:              *phrases,
		Length:               *length,
		Password:             *password,
		AllowRepeat:          *allowRepeat,
		AllowAdjacentRepeats: *allowAdjacent,
	}
	if forTUI && !languageFlag && !base.Configured() {
		cfg.ForceSetup = true
		if cfg.Language == "" {
			cfg.Language = profile.DefaultLanguage()
		}
	}
	if cfg.Language == "" && !forTUI {
		cfg.Language = profile.DefaultLanguage()
	}
	return cfg
}

func applyRunConfig(p *profile.Profile, cfg runConfig) error {
	if cfg.Length > 0 {
		return applyLength(p, cfg.Length)
	}
	if cfg.Phrases > 0 {
		if cfg.Phrases <= 0 {
			return fmt.Errorf("phrases must be positive")
		}
		p.PasswordLength = cfg.Phrases * p.BlockSize
	}
	return nil
}

func persistRunConfig(cfg runConfig) error {
	return settings.Save(settings.Settings{
		Language:             cfg.Language,
		Phrases:              cfg.Phrases,
		AllowRepeat:          cfg.AllowRepeat,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
	})
}
