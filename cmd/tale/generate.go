package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/llm"
	"github.com/escribanoruben/tale/internal/password"
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

	if cfg.Password != "" {
		if err := validateProvidedPassword(p, cfg.Password); err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
		// an explicit password fixes the length
		p.PasswordLength = len(cfg.Password)
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
	if storyish && opts.Semantic && cfg.Password != "" {
		// passtale for an existing password without the polish layer
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

// validateProvidedPassword checks an externally generated password:
// length must be a multiple of the block size and every character must
// belong to the language alphabet.
func validateProvidedPassword(p profile.Profile, pass string) error {
	var invalid []rune
	for _, r := range pass {
		if !strings.ContainsRune(p.Alphabet, r) {
			invalid = append(invalid, r)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("cannot derive a passtale: characters outside the language alphabet %q: %c", p.Alphabet, invalid)
	}
	if n := utf8.RuneCountInString(pass); n%p.BlockSize != 0 {
		return fmt.Errorf("cannot derive a passtale: password length %d is not a multiple of the block size %d", n, p.BlockSize)
	}
	return nil
}

// runGeneratePasstale is the top mode: semantic story + local LLM polish.
// Output pairs each fluent sentence with its carrier recap, so recovery
// never depends on the render. When cfg.Password is set, the passtale is
// derived for that password instead of generating a new one.
func runGeneratePasstale(p profile.Profile, dict *dictionary.Dictionary, cfg runConfig, base settings.Settings) {
	opts := phrase.Options{
		AllowRepeat:          false,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
		Semantic:             true,
	}
	var pass string
	var blocks []string
	var pools [][][]string
	var err error
	if cfg.Password != "" {
		pass = cfg.Password
		pools, err = phrase.StoryPoolsForPassword(p, pass, dict, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
		blocks, err = password.Blocks(pass, p.BlockSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
	} else {
		pass, blocks, pools, err = phrase.GenerateStoryBlocks(p, dict, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
	}

	model := base.LLMModel
	if model == "" {
		model = llm.DefaultModel
	}
	var polish []llm.BlockPolish
	if base.LLMPolish {
		client := llm.New(llm.Config{
			URL:     base.LLMURL,
			Model:   model,
			Threads: base.LLMThreads,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		polish = client.PolishStory(ctx, blocks, pools)
	} else {
		// no polish: the whole-story DP result is the passtale
		stories, err := phrase.SemanticStories(p, blocks, dict, opts, 1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
		for _, words := range stories[0] {
			polish = append(polish, llm.BlockPolish{Carriers: words})
		}
	}

	entropy := phrase.EntropyBits(len(p.Alphabet), len(pass), !cfg.AllowAdjacentRepeats)
	strategy := p.Strategy
	if strategy == "" || strategy == profile.StrategyAliasStory {
		strategy = profile.StrategyStory
	}
	fmt.Printf("Password: %s\n\n", pass)
	if base.LLMPolish {
		fmt.Printf("Passtale (%s + %s):\n", strategy, model)
	} else {
		fmt.Printf("Passtale (%s):\n", strategy)
	}
	debugPools := os.Getenv("TALE_DEBUG_POOLS") != ""
	for i, block := range blocks {
		pb := polish[i]
		if debugPools {
			var shown [][]string
			for j, c := range pools[i] {
				if j >= 3 {
					break
				}
				shown = append(shown, c)
			}
			fmt.Printf("  [%s] candidatas embeddings:\n", block)
			for _, c := range shown {
				fmt.Printf("      · %s\n", strings.Join(c, " "))
			}
		}
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
	var pass string
	var words []string
	var err error
	if cfg.Password != "" {
		pass = cfg.Password
		words, err = phrase.NounsForPassword(p, pass, dict)
	} else {
		pass, words, err = phrase.GenerateNouns(p, dict, opts)
	}
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
