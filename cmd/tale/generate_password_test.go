package main

import (
	"strings"
	"testing"

	"github.com/escribanoruben/tale/internal/profile"
)

func TestValidateProvidedPassword(t *testing.T) {
	p, err := profile.Load(profile.Files, "es")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateProvidedPassword(p, "ochfevftqmbjfndj"); err != nil {
		t.Fatalf("valid password rejected: %v", err)
	}
	err = validateProvidedPassword(p, "ochfevftqmbjfnd") // 15 letters
	if err == nil || !strings.Contains(err.Error(), "multiple of the block size") {
		t.Fatalf("want block-size error, got %v", err)
	}
	err = validateProvidedPassword(p, "ochfevftqm1jfndj") // digit
	if err == nil || !strings.Contains(err.Error(), "outside the language alphabet") {
		t.Fatalf("want alphabet error, got %v", err)
	}
	err = validateProvidedPassword(p, "ochfevftqmñjfndj") // ñ, 16 letters
	if err == nil || !strings.Contains(err.Error(), "outside the language alphabet") {
		t.Fatalf("want alphabet error for ñ, got %v", err)
	}
}
