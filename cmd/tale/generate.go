package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/phrase"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/settings"
)

func runGenerate(args []string) {
	cfg := parseRunConfig(args, false)
	cfg = mergeBatchDefaults(cfg)

	resolved, dict := loadProfileFatal("generate", cfg.Language)
	p := resolved.Profile
	printUserOverrideHint(resolved)
	if err := applyRunConfig(&p, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	if p.Strategy == profile.StrategyNounField || p.Strategy == profile.StrategyNounPath {
		runGenerateNouns(p, dict, cfg)
		return
	}

	opts := phrase.Options{
		AllowRepeat:          false,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
	}
	if base, _ := settings.Load(); base.SemanticEnabled() {
		opts.Semantic = phrase.SemanticAvailable(p, dict)
	}
	pass, blocks, err := phrase.GeneratePassword(p, dict, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	printGenerationOutput(p, pass, blocks, !cfg.AllowAdjacentRepeats)
}

// runGenerateNouns handles the noun-only strategies: the password is
// spelled by the initials of X nouns.
func runGenerateNouns(p profile.Profile, dict *dictionary.Dictionary, cfg runConfig) {
	opts := phrase.Options{AllowAdjacentRepeats: cfg.AllowAdjacentRepeats}
	pass, words, err := phrase.GenerateNouns(p, dict, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	entropy := phrase.EntropyBits(len(p.Alphabet), len(pass), !cfg.AllowAdjacentRepeats)
	fmt.Printf("Password: %s\n\n", pass)
	fmt.Printf("Mnemonic nouns (%s):\n  %s\n", p.Strategy, strings.Join(words, " "))
	if !cfg.AllowAdjacentRepeats && len(pass) > 1 && len(p.Alphabet) >= 2 {
		fmt.Printf("\nEntropy: ~%.1f bits (no adjacent repeats)\n", entropy)
	} else {
		fmt.Printf("\nEntropy: ~%.1f bits (%d chars × log₂(%d))\n", entropy, len(pass), len(p.Alphabet))
	}
	fmt.Println("\nRecovery:")
	fmt.Println("  Take the first letter of each noun, in order (lowercase, strip accents on the first character, ñ→n).")
	fmt.Println("  Concatenate the initials to recover the password.")
}

func mergeBatchDefaults(cfg runConfig) runConfig {
	if cfg.Phrases <= 0 {
		cfg.Phrases = 5
	}
	return cfg
}

func applyLength(p *profile.Profile, length int) error {
	if length == 0 {
		return nil
	}
	if length <= 0 {
		return fmt.Errorf("length must be positive")
	}
	if length%p.BlockSize != 0 {
		return fmt.Errorf("length %d is not divisible by block size %d", length, p.BlockSize)
	}
	p.PasswordLength = length
	return nil
}

func printGenerationOutput(p profile.Profile, pass string, blocks []phrase.BlockResult, preventAdjacentRepeats bool) {
	entropy := phrase.EntropyBits(len(p.Alphabet), len(pass), preventAdjacentRepeats)
	fmt.Printf("Password: %s\n\n", pass)
	fmt.Println("Mnemonic phrases:")
	for _, b := range blocks {
		fmt.Printf("  [%s] %s\n", b.Letters, strings.Join(b.Words, " "))
	}
	if preventAdjacentRepeats && len(pass) > 1 && len(p.Alphabet) >= 2 {
		fmt.Printf("\nEntropy: ~%.1f bits (no adjacent repeats)\n", entropy)
	} else {
		fmt.Printf("\nEntropy: ~%.1f bits (%d chars × log₂(%d))\n", entropy, len(pass), len(p.Alphabet))
	}
	fmt.Println("\nRecovery:")
	fmt.Println("  Take the first letter of each word in order (lowercase, strip accents on the first character, ñ→n).")
	fmt.Println("  Concatenate block initials to recover the password.")
}
