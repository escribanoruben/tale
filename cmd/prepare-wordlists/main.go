package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/escribanoruben/tale/internal/wordprepare"
)

func main() {
	repo := flag.String("repo", "", "repository root (default: auto-detect from go.mod)")
	flag.Parse()

	fmt.Println("Tales Vault word list preparation (curated)")

	_, issues, err := wordprepare.Prepare(wordprepare.Options{RepoRoot: *repo})
	if err != nil {
		fmt.Fprintf(os.Stderr, "prepare-wordlists: %v\n", err)
		os.Exit(1)
	}
	if len(issues) > 0 {
		os.Exit(1)
	}
}
