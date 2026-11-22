package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/escribanoruben/tale/internal/worddata"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runTUI(nil)
		return
	}

	switch args[0] {
	case "version", "-v", "--version":
		printVersion()
		return
	case "update":
		runUpdate(args[1:])
	case "help", "-h", "--help":
		usage()
		return
	case "stats":
		runStats(args[1:])
	case "generate":
		runGenerate(args[1:])
	case "tui":
		runTUI(args[1:])
	case "language":
		runLanguage(args[1:])
	case "profile":
		runProfile(args[1:])
	default:
		if hasFlag(args, "--no-tui", "-n") {
			runGenerate(stripFlag(args, "--no-tui", "-n"))
			return
		}
		if strings.HasPrefix(args[0], "-") {
			runTUI(args)
			return
		}
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage()
		os.Exit(1)
	}
}

func hasFlag(args []string, names ...string) bool {
	for _, a := range args {
		for _, name := range names {
			if a == name {
				return true
			}
		}
	}
	return false
}

func stripFlag(args []string, names ...string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		drop := false
		for _, name := range names {
			if a == name {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, a)
		}
	}
	return out
}

func runStats(args []string) {
	fs := pflag.NewFlagSet("stats", pflag.ExitOnError)
	jsonOut := fs.BoolP("json", "j", false, "output JSON")
	min := fs.IntP("min", "m", 8, "highlight buckets with fewer than N words")
	_ = fs.Parse(args)

	report, err := worddata.Stats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		os.Exit(1)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "stats: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Print("Embedded word data\n\n")
	fmt.Print(worddata.FormatReport(report, *min))
}

func usage() {
	fmt.Fprintf(os.Stderr, `tale — Tales Vault System

Usage:
  tale [options]                         interactive TUI (default)
  tale -n [options]                      generate password + passtale (no TUI)
  tale generate [options]                same as -n / --no-tui
  tale tui [options]                     interactive TUI
  tale language list|show <language>     manage languages
  tale language init <language> [--from LANG] [--display-name NAME] [--force] [--force-data]
  tale language new <language> --alphabet LETTERS [--display-name NAME] [--force] [--force-data]
  tale profile …                         deprecated alias for language …
  tale stats [--json/-j] [--min/-m N]    show embedded word-list statistics
  tale version                           show version, commit, and build date
  tale update [--check/-c] [--yes/-y]    update the binary from the latest release
  tale help                              show this help

Options (TUI and generate):
  -l, --language LANG           language code (en, es, or a custom language; saved to ~/.tale/settings.toml)
  -p, --phrases N               number of mnemonic phrases (default from settings or 5)
      --length N                password length (overrides --phrases)
      --password PASS           derive the passtale from an existing password (the TUI lets you edit it)
  -a, --allow-adjacent-repeats  allow consecutive identical password letters
  -n, --no-tui                  generate instead of starting the TUI

Settings file: ~/.tale/settings.toml
Languages:     ~/.tale/profiles/{language}.toml (overrides built-in en/es)
Word data:     ~/.tale/data/{language}/ (required for customized languages; no embed fallback)

TUI: Space/⌫ word  ←→ slot  ↑↓ phrase  ·  n/p h/l j/k vim  s settings  ? help  Enter show result  Esc quit
`)
}
