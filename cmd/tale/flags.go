package main

import (
	"fmt"
	"os"

	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/settings"
	"github.com/spf13/pflag"
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

// Short-flag map (kept collision-free):
//
//	-l LANG   --language                -p N   --phrases
//	-a        --allow-adjacent-repeats  -n     --no-tui (top level)
//
// --length and --password stay long-only: -p is taken by phrases and a
// short for the rarer, longer-valued flags would invite mistypes.
func parseRunConfig(args []string, forTUI bool) runConfig {
	base, _ := settings.Load()
	langDefault := base.Language
	if langDefault == "" && !forTUI {
		langDefault = profile.DefaultLanguage()
	}

	fs := pflag.NewFlagSet("tale", pflag.ExitOnError)
	fs.Usage = runFlagUsage
	language := fs.StringP("language", "l", langDefault, "language code (en, es, or custom)")
	phrases := fs.IntP("phrases", "p", base.PhrasesOr(5), "number of mnemonic phrases (blocks)")
	length := fs.Int("length", 0, "password length (overrides --phrases)")
	allowAdjacent := fs.BoolP("allow-adjacent-repeats", "a", base.AllowAdjacentRepeats, "allow consecutive identical password characters")
	password := fs.String("password", "", "existing password (derives its passtale; TUI edits it)")
	_ = fs.Parse(args)

	languageFlag := false
	fs.Visit(func(f *pflag.Flag) {
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

// runFlagUsage lists the shared TUI/generate flags with both spellings.
func runFlagUsage() {
	fmt.Fprint(os.Stderr, `Options (TUI and generate):
  -l, --language LANG           language code (en, es, or a custom language; saved to ~/.tale/settings.toml)
  -p, --phrases N               number of mnemonic phrases (default from settings or 5)
      --length N                password length (overrides --phrases)
      --password PASS           derive the passtale from an existing password (the TUI lets you edit it)
  -a, --allow-adjacent-repeats  allow consecutive identical password letters
  -h, --help                    show this help
`)
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
		CheckUpdates:         existing.CheckUpdates,
	})
}
