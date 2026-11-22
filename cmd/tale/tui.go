package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
	"github.com/escribanoruben/tale/internal/settings"
	"github.com/escribanoruben/tale/internal/tui"
	"github.com/escribanoruben/tale/internal/update"
)

func runTUI(args []string) {
	cfg := parseRunConfig(args, true)

	// Opt-in update notice: one throttled check (max once per 24h) so
	// the welcome screen can point to `tale update`. Default off, and
	// never part of the generation path.
	var updateNotice string
	if base, _ := settings.Load(); base.CheckUpdates {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		updateNotice = update.Notice(ctx, Version(), 24*time.Hour)
	}

	if err := tui.RunApp(tui.AppConfig{
		Version:              Version(),
		UpdateNotice:         updateNotice,
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
