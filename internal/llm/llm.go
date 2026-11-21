// Package llm adds the polish layer on top of the semantic engine: the
// embeddings propose candidate phrases per block, a local LLM renders
// each candidate as a fluent sentence (with retries) and picks the most
// natural. Recovery never depends on the render: the carrier words are
// validated in order and returned alongside the sentence, and the CLI
// prints them as the recovery recap.
//
// The LLM is expected on a local ollama (localhost); nothing here talks
// to remote services. Runtime failures degrade gracefully: blocks whose
// render fails keep the telegraphic phrase.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DefaultModel is the ollama model the polish layer uses when the
// settings leave it empty: small enough for CPU, good Spanish.
const DefaultModel = "qwen2.5:1.5b"

type Config struct {
	URL         string        // default http://localhost:11434
	Model       string        // e.g. qwen2.5:1.5b
	Threads     int           // compute threads for the LLM (0 = server default)
	Temperature float64       // 0 keeps the server default
	Timeout     time.Duration // per request, default 90s
}

func (c Config) withDefaults() Config {
	if c.URL == "" {
		c.URL = "http://localhost:11434"
	}
	if c.Timeout <= 0 {
		c.Timeout = 90 * time.Second
	}
	if c.Temperature <= 0 {
		c.Temperature = 0.4 // deterministic sampling degenerates small models
	}
	return c
}

// Client talks to a local ollama server.
type Client struct {
	cfg  Config
	http *http.Client
}

// New builds a client from the config.
func New(cfg Config) *Client {
	c := cfg.withDefaults()
	return &Client{cfg: c, http: &http.Client{Timeout: c.Timeout}}
}

// BlockPolish is the polished output for one block: the fluent sentence
// (if the LLM produced a valid render) and the carrier words that spell
// the block (always present).
type BlockPolish struct {
	Sentence string   // fluent render; empty when the LLM failed -> use telegraphic
	Carriers []string // validated carrier words, in order (initials = block)
	Fallback bool     // true when the render failed and the telegraphic form is kept
}

// PolishStory runs the pipeline over the per-block candidate pools:
// render (with retries) -> discard broken renders -> pick the most
// natural among the valid ones. Blocks with no valid render are returned
// as fallback (empty Sentence, Carriers = the pool's top candidate).
func (c *Client) PolishStory(ctx context.Context, pools [][][]string) []BlockPolish {
	out := make([]BlockPolish, 0, len(pools))
	for _, pool := range pools {
		out = append(out, c.polishBlock(ctx, pool))
	}
	return out
}

func (c *Client) polishBlock(ctx context.Context, pool [][]string) BlockPolish {
	if len(pool) == 0 {
		return BlockPolish{Fallback: true}
	}
	const maxCandidates = 3
	const maxAttempts = 3
	var valid []BlockPolish
	for i, carriers := range pool {
		if i >= maxCandidates {
			break
		}
		for attempt := 0; attempt < maxAttempts; attempt++ {
			extra := ""
			if attempt > 0 {
				extra = fmt.Sprintf(" Las palabras %s deben aparecer EN ESE ORDEN exacto, sin cambiarlas ni reordenarlas.", strings.Join(carriers, ", "))
			}
			prompt := fmt.Sprintf("Escribe en español una frase memorable usando estas 4 palabras EN ESTE ORDEN y sin cambiarlas: %s.\nSolo puedes DECORAR añadiendo artículos, preposiciones, pronombres y conjunciones (el, la, un, de, en, con, por, para, y, me, se, no...). Está PROHIBIDO añadir otras palabras de contenido (nada de sustantivos, verbos o adverbios nuevos).%s Responde solo la frase, sin comillas ni explicaciones.", strings.Join(carriers, " "), extra)
			resp, err := c.generate(ctx, prompt, 90)
			if err != nil {
				break // LLM unreachable: no point retrying
			}
			s := strings.TrimSpace(extractQuoted(strings.Trim(resp, "\"")))
			if carriersOk(s, carriers) {
				valid = append(valid, BlockPolish{Sentence: trimToFirstCarrier(s, carriers), Carriers: carriers})
				break
			}
		}
	}
	if len(valid) == 0 {
		return BlockPolish{Carriers: pool[0], Fallback: true}
	}
	if len(valid) == 1 {
		return valid[0]
	}
	sel := 0
	var lines []string
	for i, v := range valid {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, v.Sentence))
	}
	prompt := fmt.Sprintf("¿Qué frase suena más natural y es más fácil de recordar?\n%s\nResponde SOLO el número.", strings.Join(lines, "\n"))
	resp, err := c.generate(ctx, prompt, 20)
	if err == nil {
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(resp), "%d", &n); err == nil && n >= 1 && n <= len(valid) {
			sel = n - 1
		}
	}
	return valid[sel]
}

// generate posts one prompt to ollama and returns the response text.
func (c *Client) generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if c.cfg.Model == "" {
		return "", fmt.Errorf("llm: no model configured")
	}
	payload := map[string]any{
		"model":  c.cfg.Model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]any{
			"temperature": c.cfg.Temperature,
			"num_predict": maxTokens,
		},
	}
	if c.cfg.Threads > 0 {
		payload["options"].(map[string]any)["num_thread"] = c.cfg.Threads
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.URL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm: status %s", resp.Status)
	}
	var out struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Response), nil
}

// trimToFirstCarrier strips chatty preambles ("En este orden, ...:")
// by cutting everything before the first carrier occurrence.
func trimToFirstCarrier(sentence string, carriers []string) string {
	hay := strings.Fields(strings.ToLower(sentence))
	if len(hay) == 0 {
		return sentence
	}
	target := strings.ToLower(carriers[0])
	for i, tok := range hay {
		tok = strings.Trim(tok, ".,;:!?¡¿\"'«»()[]")
		if sharedPrefixLen(tok, target) >= 4 {
			// rebuild keeping original casing/spacing from the first hit
			lower := strings.ToLower(sentence)
			pos := 0
			count := 0
			for pos < len(lower) && count < i {
				// skip one word
				for pos < len(lower) && lower[pos] == ' ' {
					pos++
				}
				for pos < len(lower) && lower[pos] != ' ' {
					pos++
				}
				count++
			}
			for pos < len(lower) && lower[pos] == ' ' {
				pos++
			}
			if pos > 0 && pos < len(sentence) {
				return strings.TrimSpace(sentence[pos:])
			}
			return sentence
		}
	}
	return sentence
}
func carriersOk(sentence string, carriers []string) bool {
	hay := strings.Fields(strings.ToLower(sentence))
	idx := 0
	for _, carrier := range carriers {
		target := strings.ToLower(carrier)
		found := false
		for ; idx < len(hay); idx++ {
			tok := strings.Trim(hay[idx], ".,;:!?¡¿\"'«»()[]")
			if sharedPrefixLen(tok, target) >= 4 {
				found = true
				idx++
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// sharedPrefixLen counts the common leading characters of two words.
func sharedPrefixLen(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}

// extractQuoted pulls the quoted segment out of a chatty response, if any.
func extractQuoted(resp string) string {
	if a := strings.Index(resp, "'"); a >= 0 {
		if b := strings.LastIndex(resp, "'"); b > a+1 {
			return strings.TrimSpace(resp[a+1 : b])
		}
	}
	if a := strings.Index(resp, "\""); a >= 0 {
		if b := strings.LastIndex(resp, "\""); b > a+1 {
			return strings.TrimSpace(resp[a+1 : b])
		}
	}
	return resp
}
