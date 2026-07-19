package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ruben/tales_vault_system/internal/phrase"
	"github.com/ruben/tales_vault_system/internal/profile"
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

	opts := phrase.Options{
		AllowRepeat:          cfg.AllowRepeat,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
	}
	pass, blocks, err := phrase.GeneratePassword(p, dict, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	printGenerationOutput(p, pass, blocks, !cfg.AllowAdjacentRepeats)
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
