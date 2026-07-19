package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ruben/tales_vault_system/internal/paths"
	"github.com/ruben/tales_vault_system/internal/profile"
)

func runLanguage(args []string) {
	if len(args) == 0 {
		languageUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		runLanguageInit(args[1:])
	case "new":
		runLanguageNew(args[1:])
	case "list":
		runLanguageList(args[1:])
	case "show":
		runLanguageShow(args[1:])
	case "help", "-h", "--help":
		languageUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown language command %q\n\n", args[0])
		languageUsage()
		os.Exit(1)
	}
}

func runProfile(args []string) {
	fmt.Fprintln(os.Stderr, "note: 'tale profile' is deprecated; use 'tale language' instead.")
	runLanguage(args)
}

func runLanguageInit(args []string) {
	fs := flag.NewFlagSet("language init", flag.ExitOnError)
	from := fs.String("from", "", "source language (default: same as target)")
	displayName := fs.String("display-name", "", "human-readable language name")
	force := fs.Bool("force", false, "overwrite existing user language config")
	forceData := fs.Bool("force-data", false, "overwrite existing ~/.tale/data/{language}")
	lang, flagArgs, err := splitPositional(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "usage: tale language init <language> [--from LANG]")
		os.Exit(1)
	}
	_ = fs.Parse(flagArgs)

	result, err := profile.InitUserProfile(profile.InitOptions{
		Language:     lang,
		FromLanguage: *from,
		DisplayName:  *displayName,
		Force:        *force,
		ForceData:    *forceData,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "language init: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created language config: %s\n", result.ProfilePath)
	fmt.Printf("Word data dir:          %s\n", result.DataPath)
	if result.DataCopied {
		fmt.Println("Word lists copied from source.")
	} else {
		fmt.Println("Using existing word lists (not overwritten).")
	}
	fmt.Println("\nThis user config replaces any built-in defaults for this language.")
	fmt.Println("Edit the TOML to customize alphabet, length, or phrase template.")
	fmt.Println("Edit TSV files under the word data dir to customize words.")
}

func runLanguageNew(args []string) {
	fs := flag.NewFlagSet("language new", flag.ExitOnError)
	alphabet := fs.String("alphabet", "", "required unique lowercase letters for this language")
	displayName := fs.String("display-name", "", "human-readable language name")
	force := fs.Bool("force", false, "overwrite existing user language config")
	forceData := fs.Bool("force-data", false, "overwrite existing ~/.tale/data/{language}")
	lang, flagArgs, err := splitPositional(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "usage: tale language new <language> --alphabet LETTERS")
		os.Exit(1)
	}
	_ = fs.Parse(flagArgs)

	if *alphabet == "" {
		fmt.Fprintln(os.Stderr, "language new: --alphabet is required (unique lowercase letters)")
		os.Exit(1)
	}

	result, err := profile.NewUserLanguage(profile.NewLanguageOptions{
		Language:    lang,
		Alphabet:    *alphabet,
		DisplayName: *displayName,
		Force:       *force,
		ForceData:   *forceData,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "language new: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created language config: %s\n", result.ProfilePath)
	fmt.Printf("Word data dir:          %s\n", result.DataPath)
	if result.DataScaffolded {
		fmt.Println("Empty word-list stubs created for each alphabet letter.")
	} else {
		fmt.Println("Using existing word lists (not overwritten).")
	}
	if len(*alphabet) < 10 {
		fmt.Fprintf(os.Stderr, "warning: alphabet length %d is short; entropy will be low.\n", len(*alphabet))
	}
	fmt.Println("\nNext steps:")
	fmt.Println("  Fill nouns.tsv, adjectives.tsv, and verbs.tsv under each letter folder.")
	fmt.Println("  Until those lists have words, generate/tui will fail for uncovered letters.")
	fmt.Printf("  Then: tale generate --language %s\n", lang)
}

func runLanguageList(args []string) {
	_ = flag.NewFlagSet("language list", flag.ExitOnError).Parse(args)

	entries, err := profile.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "language list: %v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LANGUAGE\tSOURCE\tDISPLAY NAME")
	for _, e := range entries {
		source := "builtin"
		if e.Customized {
			source = "customized"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.Language, source, e.DisplayName)
	}
	_ = w.Flush()

	configDir, _ := paths.ConfigDir()
	if configDir != "" {
		fmt.Printf("\nUser languages: %s/profiles/{language}.toml\n", configDir)
		fmt.Printf("User data:      %s/data/{language}\n", configDir)
	}
}

func runLanguageShow(args []string) {
	fs := flag.NewFlagSet("language show", flag.ExitOnError)
	lang, flagArgs, err := splitPositional(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "usage: tale language show <language>")
		os.Exit(1)
	}
	_ = fs.Parse(flagArgs)

	resolved, err := profile.Resolve(lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "language show: %v\n", err)
		os.Exit(1)
	}

	p := resolved.Profile
	source := "builtin"
	if !resolved.Builtin {
		source = "user (overrides built-in for this language)"
	}
	fmt.Printf("Language:        %s\n", p.Language)
	fmt.Printf("Display name:    %s\n", p.DisplayName)
	fmt.Printf("Source:          %s\n", source)
	fmt.Printf("Password length: %d\n", p.PasswordLength)
	fmt.Printf("Block size:      %d\n", p.BlockSize)
	fmt.Printf("Alphabet (%d):   %s\n", len(p.Alphabet), p.Alphabet)
	fmt.Printf("Phrase template: %v\n", p.PhraseTemplate)
	if resolved.FilePath != "" {
		fmt.Printf("Config file:     %s\n", resolved.FilePath)
	}
	dataDir, err := paths.LanguageDataDir(p.Language)
	if err == nil {
		fmt.Printf("Word data dir:   %s\n", dataDir)
		if !resolved.Builtin {
			fmt.Println("Word data:       user directory only (no embedded fallback)")
		}
	}
}

// splitPositional takes the first non-flag argument as the language code so that
// flags may appear before or after it (Go's flag package stops at the first positional).
func splitPositional(args []string) (positional string, flagArgs []string, err error) {
	valueFlags := map[string]bool{
		"-from": true, "--from": true,
		"-display-name": true, "--display-name": true,
		"-alphabet": true, "--alphabet": true,
	}
	flagArgs = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			rest := args[i+1:]
			if positional == "" && len(rest) > 0 {
				positional = rest[0]
				flagArgs = append(flagArgs, rest[1:]...)
			} else {
				flagArgs = append(flagArgs, rest...)
			}
			break
		}
		if strings.HasPrefix(a, "-") {
			flagArgs = append(flagArgs, a)
			name, _, hasEq := strings.Cut(a, "=")
			if !hasEq && valueFlags[name] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flagArgs = append(flagArgs, args[i])
			}
			continue
		}
		if positional != "" {
			return "", nil, fmt.Errorf("unexpected argument %q", a)
		}
		positional = a
	}
	if positional == "" {
		return "", nil, fmt.Errorf("missing language")
	}
	return positional, flagArgs, nil
}

func languageUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  tale language list
  tale language show <language>
  tale language init <language> [--from LANG] [--display-name NAME] [--force] [--force-data]
  tale language new <language> --alphabet LETTERS [--display-name NAME] [--force] [--force-data]

One configuration per language. User files in ~/.tale/profiles/{language}.toml replace built-ins.
language new scaffolds empty word lists for the given alphabet; init --from clones an existing language.

Examples:
  tale language init es
  tale language init de --from en --display-name "German"
  tale language new pt --alphabet=abcdefghijlmnopqrstuvxz --display-name Portuguese
  tale generate --language es
`)
}
