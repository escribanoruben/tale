package main

import (
	"testing"

	"github.com/escribanoruben/tale/internal/profile"
)

func TestParseRunConfigMixedShortAndLong(t *testing.T) {
	cfg := parseRunConfig([]string{
		"--language", "es", // long form
		"-l", "en", // short form: last occurrence wins
		"-p", "4", // short phrases
		"-a",                       // short bool
		"--password", "abcdefghij", // long-only flag
	}, false)
	if cfg.Language != "en" {
		t.Fatalf("language %q, want en (last occurrence wins)", cfg.Language)
	}
	if cfg.Phrases != 4 {
		t.Fatalf("phrases %d, want 4", cfg.Phrases)
	}
	if !cfg.AllowAdjacentRepeats {
		t.Fatal("allow-adjacent-repeats false, want true")
	}
	if cfg.Password != "abcdefghij" {
		t.Fatalf("password %q, want abcdefghij", cfg.Password)
	}
	if cfg.ForceSetup {
		t.Fatal("force setup with explicit language")
	}
}

func TestParseRunConfigEqualsForm(t *testing.T) {
	cfg := parseRunConfig([]string{"--phrases=3", "--language=en"}, false)
	if cfg.Phrases != 3 {
		t.Fatalf("phrases %d, want 3", cfg.Phrases)
	}
	if cfg.Language != "en" {
		t.Fatalf("language %q, want en", cfg.Language)
	}
}

func TestParseRunConfigLongBool(t *testing.T) {
	cfg := parseRunConfig([]string{"--allow-adjacent-repeats", "-l", "es"}, false)
	if !cfg.AllowAdjacentRepeats {
		t.Fatal("long-form bool flag not set")
	}
}

func TestParseRunConfigTUIForceSetupWhenUnconfigured(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // no ~/.tale/settings.toml

	cfg := parseRunConfig(nil, true)
	if !cfg.ForceSetup {
		t.Fatal("unconfigured TUI without -l should force setup")
	}
	if cfg.Language != profile.DefaultLanguage() {
		t.Fatalf("language %q, want default", cfg.Language)
	}

	cfg = parseRunConfig([]string{"-l", "es"}, true)
	if cfg.ForceSetup {
		t.Fatal("explicit -l es should skip force setup")
	}
	if cfg.Language != "es" {
		t.Fatalf("language %q, want es", cfg.Language)
	}
}

func TestNoTUIShortFlag(t *testing.T) {
	args := []string{"-n", "-l", "es", "-p", "4"}
	if !hasFlag(args, "--no-tui", "-n") {
		t.Fatal("hasFlag missed -n")
	}
	rest := stripFlag(args, "--no-tui", "-n")
	if len(rest) != 4 || rest[0] != "-l" || rest[1] != "es" || rest[2] != "-p" || rest[3] != "4" {
		t.Fatalf("stripFlag = %v, want flags without -n", rest)
	}
}
