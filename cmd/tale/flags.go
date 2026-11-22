package main

import (
	"flag"
	"fmt"

	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/settings"
)

// runConfig holds merged CLI flags and user settings.
type runConfig struct {
	Language             string
	Phrases              int
	Length               int
	Password             string
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
	allowAdjacent := fs.Bool("allow-adjacent-repeats", base.AllowAdjacentRepeats, "allow consecutive identical password characters")
	password := fs.String("password", "", "existing password (derives its passtale; TUI edits it)")
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
	// preserve keys the TUI does not edit (LLM polish layer, ...)
	existing, err := settings.Load()
	if err != nil {
		existing = settings.Default()
	}
	return settings.Save(settings.Settings{
		Language:             cfg.Language,
		Phrases:              cfg.Phrases,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
		LLMPolish:            existing.LLMPolish,
		LLMModel:             existing.LLMModel,
		LLMURL:               existing.LLMURL,
		LLMThreads:           existing.LLMThreads,
	})
}
