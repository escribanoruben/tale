package main

import "testing"

func TestVersionUsesInjectedValue(t *testing.T) {
	prev := version
	t.Cleanup(func() { version = prev })
	version = "1.2.3"
	if got := Version(); got != "1.2.3" {
		t.Fatalf("Version() = %q, want 1.2.3", got)
	}
}

func TestVersionFallsBackToDev(t *testing.T) {
	prev := version
	t.Cleanup(func() { version = prev })
	version = "dev"
	if got := Version(); got != "dev" {
		t.Fatalf("Version() = %q, want dev", got)
	}
}
