package main

import (
	"fmt"
	"runtime/debug"
	"strings"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version returns the release version, falling back to module build info for go install builds.
func Version() string {
	if v := strings.TrimSpace(version); v != "" && v != "dev" {
		return v
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := strings.TrimSpace(info.Main.Version); v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	if version != "" {
		return version
	}
	return "dev"
}

func printVersion() {
	fmt.Printf("tale %s (commit %s, built %s)\n", Version(), commit, date)
}
