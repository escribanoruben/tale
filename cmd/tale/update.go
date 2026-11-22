package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/escribanoruben/tale/internal/update"
)

func runUpdate(args []string) {
	fs := pflag.NewFlagSet("update", pflag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage:
  tale update [--check/-c] [--yes/-y]

Checks the latest GitHub release and updates the tale binary in place.
Artifacts are verified against the published checksums.txt before the
swap; on any failure the current binary is kept.

  -c, --check  only report the latest release and up-to-date status
  -y, --yes    skip the confirmation prompt
`)
	}
	checkOnly := fs.BoolP("check", "c", false, "check for a new release without updating")
	yes := fs.BoolP("yes", "y", false, "update without asking for confirmation")
	_ = fs.Parse(args)

	current := Version()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	latest, err := update.Check(ctx, nil, current)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		os.Exit(1)
	}
	if !latest.Found {
		fmt.Println("No published release found.")
		return
	}

	if *checkOnly {
		if latest.Newer {
			fmt.Printf("tale %s -> %s available: run `tale update`\n", current, latest.Latest)
		} else {
			fmt.Printf("tale %s is up to date (latest release: %s).\n", current, latest.Latest)
		}
		return
	}

	if !latest.Newer && current != "dev" {
		fmt.Printf("tale %s is up to date (latest release: %s).\n", current, latest.Latest)
		return
	}

	fmt.Printf("tale %s -> %s\n", current, latest.Latest)
	if !*yes {
		fmt.Print("Update now? [y/N] ")
		answer := ""
		if scanner := bufio.NewScanner(os.Stdin); scanner.Scan() {
			answer = strings.ToLower(strings.TrimSpace(scanner.Text()))
		}
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return
		}
	}

	from, to, err := update.Apply(ctx, nil, current)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated tale %s -> %s.\n", from, to)
}
