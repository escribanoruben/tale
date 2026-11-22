package tui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fmt"
	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/llm"
	"github.com/escribanoruben/tale/internal/profile"
)

func storyTestSession(t *testing.T) *Session {
	t.Helper()
	p, err := profile.Load(profile.Files, "es")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}
	pass := "sldpbtmrhctfgqbf" // 16 letters, 4 blocks
	s, err := NewSession(p, pass, dict, true)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestStorySessionSeedsSemanticStory(t *testing.T) {
	s := storyTestSession(t)
	if !s.Storyish {
		t.Fatal("es should be storyish by default")
	}
	if len(s.Pools) != len(s.Blocks) {
		t.Fatalf("pools %d, want %d", len(s.Pools), len(s.Blocks))
	}
	for bi, block := range s.Blocks {
		for si, w := range s.Selections[bi] {
			if w == "" {
				t.Fatalf("empty selection in story fill: block %d slot %d", bi, si)
			}
			if normalizeInitial(w) != string(block[si]) {
				t.Fatalf("word %q does not match block letter (block %s)", w, block)
			}
		}
	}
}

func normalizeInitial(w string) string {
	r := []rune(strings.ToLower(w))
	return string(r[0])
}

func TestSemanticRerollPhraseKeepsInitials(t *testing.T) {
	s := storyTestSession(t)
	before := strings.Join(s.Selections[0], "")
	if err := s.randomizeScope("p"); err != nil {
		t.Fatal(err)
	}
	after := s.Selections[0]
	if strings.Join(after, "") == before {
		t.Skip("reroll picked the same phrase; acceptable")
	}
	for si, w := range after {
		if normalizeInitial(w) != string(s.Blocks[0][si]) {
			t.Fatalf("rerolled word %q breaks the block initials", w)
		}
	}
}

func TestLLMRerollAppliesCarriers(t *testing.T) {
	s := storyTestSession(t)
	// stub LLM: render echo with a same-initial substitution on the
	// first word ("silla" -> "soga"), keeping initials s-l-d-p per block
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		prompt := req.Messages[len(req.Messages)-1].Content
		ws := extractWords(prompt)
		resp := fmt.Sprintf("1. El %s.\n2. La %s.\n3. Un %s.", strings.Join(ws, " "), strings.Join(ws, " "), strings.Join(ws, " "))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]string{"role": "assistant", "content": resp},
		})
	}))
	defer srv.Close()
	s.LLM = llm.New(llm.Config{URL: srv.URL, Model: "stub", Threads: 1})

	if err := s.randomizeScope("l"); err != nil {
		t.Fatal(err)
	}
	if s.Err != nil {
		t.Fatalf("llm reroll errored: %v", s.Err)
	}
	for si, w := range s.Selections[0] {
		if normalizeInitial(w) != string(s.Blocks[0][si]) {
			t.Fatalf("post-reroll word %q breaks initials", w)
		}
	}
}

// extractWords pulls the 4 carrier words from the render prompt (the
// line after "EN ESTE ORDEN:").
func extractWords(prompt string) []string {
	lines := strings.Split(prompt, "\n")
	for i, line := range lines {
		if strings.Contains(line, "EN ESTE ORDEN") && i+1 < len(lines) {
			return strings.Fields(lines[i+1])
		}
	}
	return nil
}
