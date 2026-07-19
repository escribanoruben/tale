package main

import (
	"fmt"
	"os"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
)

func resolveProfileAndDictionary(id string) (profile.Resolved, *dictionary.Dictionary, error) {
	resolved, err := profile.Resolve(id)
	if err != nil {
		return profile.Resolved{}, nil, err
	}
	dict, err := dictionary.ForProfile(resolved)
	if err != nil {
		return profile.Resolved{}, nil, err
	}
	return resolved, dict, nil
}

func loadProfileFatal(cmd, id string) (profile.Resolved, *dictionary.Dictionary) {
	resolved, dict, err := resolveProfileAndDictionary(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
		os.Exit(1)
	}
	return resolved, dict
}

func printUserOverrideHint(resolved profile.Resolved) {
	if resolved.Builtin {
		return
	}
	fmt.Fprintf(os.Stderr, "Using user language override for %s (config + word data under ~/.tale; embedded defaults do not apply).\n", resolved.Profile.Language)
}
