package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// stubServer answers /api/chat with scripted responses based on the
// prompt's shape: selection prompts (¿Qué frase) get a number, render
// prompts get the carrier phrase wrapped in a fluent sentence.
func stubServer(t *testing.T, renderFor func(prompt string) string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		prompt := ""
		if len(req.Messages) > 0 {
			prompt = req.Messages[len(req.Messages)-1].Content
		}
		resp := renderFor(prompt)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"message": map[string]string{"role": "assistant", "content": resp},
		})
	}))
}

func TestPolishStoryRenderAndSelect(t *testing.T) {
	rendered := 0
	srv := stubServer(t, func(prompt string) string {
		if strings.Contains(prompt, "¿Qué frase suena") {
			return "2"
		}
		rendered++
		// echo the 4 words after "EN ESTE ORDEN:" as three numbered
		// sentences (fillers only, initials preserved)
		lines := strings.Split(prompt, "\n")
		for i, line := range lines {
			if strings.Contains(line, "EN ESTE ORDEN:") && i+1 < len(lines) {
				ws := strings.Fields(lines[i+1])
				if len(ws) < 4 {
					return "sin respuesta"
				}
				return fmt.Sprintf("1. El %s.\n2. La %s.\n3. Un %s.", strings.Join(ws, " "), strings.Join(ws, " "), strings.Join(ws, " "))
			}
		}
		return "sin respuesta"
	})
	defer srv.Close()

	c := New(Config{URL: srv.URL, Model: "stub", Threads: 2})
	pools := [][][]string{
		{{"vino", "tinto", "correr", "calle"}, {"vino", "blanco", "brindar", "copa"}},
		{{"mesa", "ancha", "cargar", "pan"}},
	}
	out := c.PolishStory(context.Background(), []string{"vtcc", "macp"}, pools)
	if len(out) != 2 {
		t.Fatalf("blocks %d, want 2", len(out))
	}
	// first block: two candidates rendered; selection stub answers "2"
	// -> the SECOND rendered candidate must be chosen
	if out[0].Fallback {
		t.Fatal("first block should not fall back")
	}
	if !strings.EqualFold(out[0].Carriers[0], "vino") {
		t.Fatalf("carriers = %v", out[0].Carriers)
	}
	if out[0].Sentence == "" {
		t.Fatal("empty sentence")
	}
	// second block: single candidate -> no selection needed
	if strings.EqualFold(out[1].Carriers[1], "copa") {
		t.Fatalf("second block carriers = %v, want the first candidate", out[1].Carriers)
	}
	if rendered < 2 {
		t.Fatalf("expected >= 2 render calls, got %d", rendered)
	}
}

func TestPolishStoryFallsBackOnGarbage(t *testing.T) {
	srv := stubServer(t, func(prompt string) string {
		if strings.Contains(prompt, "¿Qué frase suena") {
			return "1"
		}
		// garbage: carriers replaced by unrelated words
		return "La.la.la cualquier cosa sin sentido."
	})
	defer srv.Close()

	c := New(Config{URL: srv.URL, Model: "stub"})
	pools := [][][]string{{{"vino", "tinto", "correr", "calle"}}}
	out := c.PolishStory(context.Background(), []string{"vtcc", "macp"}, pools)
	if !out[0].Fallback {
		t.Fatal("garbage render must fall back")
	}
	if len(out[0].Carriers) != 4 || out[0].Carriers[0] != "vino" {
		t.Fatalf("fallback carriers = %v", out[0].Carriers)
	}
	if out[0].Sentence != "" {
		t.Fatalf("fallback sentence should be empty, got %q", out[0].Sentence)
	}
}

func TestInitialsOk(t *testing.T) {
	// block "apg": content words must have initials a, p, g in order;
	// fillers are skipped and conjugations keep the initial.
	content, ok := initialsOk("Abrió la puerta con gratitud.", "apg")
	if !ok || len(content) != 3 {
		t.Fatalf("conjugation should pass: %v ok=%v", content, ok)
	}
	if _, ok := initialsOk("Se abrió una puerta roja de par en par.", "apg"); ok {
		t.Fatal("extra content word must fail (4 content tokens for 3 letters)")
	}
	if _, ok := initialsOk("Cerró la puerta grande.", "apg"); ok {
		t.Fatal("wrong initial must fail")
	}
}

func TestGenerateUnreachable(t *testing.T) {
	c := New(Config{URL: "http://127.0.0.1:1", Model: "stub", Timeout: 200 * time.Millisecond})
	pools := [][][]string{{{"sol", "claro", "salir", "mañana"}}}
	out := c.PolishStory(context.Background(), []string{"vtcc", "macp"}, pools)
	if !out[0].Fallback {
		t.Fatal("unreachable LLM must fall back")
	}
}
