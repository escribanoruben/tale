package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// stubServer answers /api/generate with scripted responses based on the
// prompt's shape: selection prompts (¿Qué frase) get a number, render
// prompts get the carrier phrase wrapped in a fluent sentence.
func stubServer(t *testing.T, renderFor func(prompt string) string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := renderFor(req.Prompt)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"response": resp})
	}))
}

func TestPolishStoryRenderAndSelect(t *testing.T) {
	rendered := 0
	srv := stubServer(t, func(prompt string) string {
		if strings.Contains(prompt, "¿Qué frase suena") {
			return "2"
		}
		rendered++
		// echo a fluent sentence using the carriers in order
		var carriers string
		for _, line := range strings.Split(prompt, "\n") {
			if strings.HasPrefix(line, "Escribe") {
				carriers = line
			}
		}
		return "Frase fluida con " + carriers + " dentro."
	})
	defer srv.Close()

	c := New(Config{URL: srv.URL, Model: "stub", Threads: 2})
	pools := [][][]string{
		{{"vino", "tinto", "correr", "calle"}, {"vino", "blanco", "brindar", "copa"}},
		{{"mesa", "ancha", "cargar", "pan"}},
	}
	out := c.PolishStory(context.Background(), pools)
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
	if rendered < 3 {
		t.Fatalf("expected >= 3 render calls, got %d", rendered)
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
	out := c.PolishStory(context.Background(), pools)
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

func TestCarriersOkStemMatching(t *testing.T) {
	carriers := []string{"abrir", "puerta", "grande"}
	if !carriersOk("Ayer se abrió la puerta grande.", carriers) {
		t.Fatal("conjugation must pass the stem match")
	}
	if carriersOk("Se abrió una puerta roja de par en par.", carriers) {
		t.Fatal("reordered carriers must fail")
	}
	if carriersOk("Cerró la puerta grande.", carriers) {
		t.Fatal("missing carrier must fail")
	}
}

func TestGenerateUnreachable(t *testing.T) {
	c := New(Config{URL: "http://127.0.0.1:1", Model: "stub", Timeout: 200 * time.Millisecond})
	pools := [][][]string{{{"sol", "claro", "salir", "mañana"}}}
	out := c.PolishStory(context.Background(), pools)
	if !out[0].Fallback {
		t.Fatal("unreachable LLM must fall back")
	}
}
