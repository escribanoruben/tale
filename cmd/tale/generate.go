package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/llm"
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

	base, _ := settings.Load()
	opts := phrase.Options{
		AllowRepeat:          false,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
	}
	storyish := p.Strategy == "" || p.Strategy == profile.StrategyStory || p.Strategy == profile.StrategyAliasStory
	if p.Strategy == profile.StrategyRandom {
		// uniform words: the classic mode, no embeddings involved
		opts.Uniform = true
	} else if base.SemanticEnabled() && storyish {
		opts.Semantic = phrase.SemanticAvailable(p, dict)
	}

	if storyish && opts.Semantic && base.LLMPolish {
		runGeneratePasstale(p, dict, cfg, base)
		return
	}

	pass, blocks, err := phrase.GeneratePassword(p, dict, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	printGenerationOutput(p, pass, blocks, !cfg.AllowAdjacentRepeats)
}

// runGeneratePasstale is the top mode: semantic story + local LLM polish.
// Output pairs each fluent sentence with its carrier recap, so recovery
// never depends on the render.
func runGeneratePasstale(p profile.Profile, dict *dictionary.Dictionary, cfg runConfig, base settings.Settings) {
	opts := phrase.Options{
		AllowRepeat:          false,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
		Semantic:             true,
	}
	pass, blocks, pools, err := phrase.GenerateStoryBlocks(p, dict, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}

	model := base.LLMModel
	if model == "" {
		model = llm.DefaultModel
	}
	client := llm.New(llm.Config{
		URL:     base.LLMURL,
		Model:   model,
		Threads: base.LLMThreads,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	polish := client.PolishStory(ctx, pools)

	entropy := phrase.EntropyBits(len(p.Alphabet), len(pass), !cfg.AllowAdjacentRepeats)
	strategy := p.Strategy
	if strategy == "" || strategy == profile.StrategyAliasStory {
		strategy = profile.StrategyStory
	}
	fmt.Printf("Password: %s\n\n", pass)
	fmt.Printf("Passtale (%s + %s):\n", strategy, model)
	for i, block := range blocks {
		pb := polish[i]
		if pb.Fallback || pb.Sentence == "" {
			fmt.Printf("  [%s] %s\n", block, strings.Join(pb.Carriers, " "))
		} else {
			fmt.Printf("  [%s] %s\n", block, pb.Sentence)
		}
		fmt.Printf("      portadoras: %s\n", strings.Join(pb.Carriers, " "))
	}
	if !cfg.AllowAdjacentRepeats && len(pass) > 1 && len(p.Alphabet) >= 2 {
		fmt.Printf("\nEntropy: ~%.1f bits (no adjacent repeats)\n", entropy)
	} else {
		fmt.Printf("\nEntropy: ~%.1f bits (%d chars × log₂(%d))\n", entropy, len(pass), len(p.Alphabet))
	}
	fmt.Println("\nRecovery:")
	fmt.Println("  Take the first letter of each carrier word in the \"portadoras\" lines, in order")
	fmt.Println("  (lowercase, strip accents on the first character, ñ→n). Concatenate to recover the password.")
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
