package main

import (
	"fmt"
	"os"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/tui"
)

func runTUI(args []string) {
	cfg := parseRunConfig(args, true)

	if err := tui.RunApp(tui.AppConfig{
		Version:              Version(),
		Language:             cfg.Language,
		Phrases:              cfg.Phrases,
		Length:               cfg.Length,
		Password:             cfg.Password,
		AllowAdjacentRepeats: cfg.AllowAdjacentRepeats,
		ForceSetup:           cfg.ForceSetup,
		SaveSettings: func(s tui.AppSettings) error {
			return persistRunConfig(runConfig{
				Language:             s.Language,
				Phrases:              s.Phrases,
				AllowAdjacentRepeats: s.AllowAdjacentRepeats,
			})
		},
		LoadProfile: func(language string) (profile.Profile, *dictionary.Dictionary, error) {
			resolved, dict, err := resolveProfileAndDictionary(language)
			if err != nil {
				return profile.Profile{}, nil, err
			}
			printUserOverrideHint(resolved)
			return resolved.Profile, dict, nil
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
