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
	"sync"
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
	if c.Threads <= 0 {
		// 2 threads: more threads than the server's allowed cores makes
		// llama.cpp thrash (spin barriers) — 25× slower in practice
		c.Threads = 2
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
	Err      error    // transport failure (unreachable LLM), if any
}

// PolishStory runs the pipeline over the per-block candidate pools:
// render (with retries) -> discard broken renders -> pick the most
// natural among the valid ones. blocks holds the letter blocks, in
// order (the mnemonic invariant). Blocks with no valid render are
// returned as fallback (empty Sentence, Carriers = the pool's top
// candidate). Blocks render concurrently; ollama queues them when its
// parallel slot count is lower.
func (c *Client) PolishStory(ctx context.Context, blocks []string, pools [][][]string) []BlockPolish {
	out := make([]BlockPolish, len(pools))
	var wg sync.WaitGroup
	for i, pool := range pools {
		wg.Add(1)
		go func(i int, pool [][]string, block string) {
			defer wg.Done()
			out[i] = c.polishBlock(ctx, block, pool)
		}(i, pool, blocks[i])
	}
	wg.Wait()
	return out
}

func (c *Client) polishBlock(ctx context.Context, block string, pool [][]string) BlockPolish {
	if len(pool) == 0 {
		return BlockPolish{Fallback: true}
	}
	carriers := pool[0] // the multi-request varies these itself
	const maxAttempts = 2
	for attempt := 0; attempt < maxAttempts; attempt++ {
		extra := ""
		if attempt > 0 {
			extra = fmt.Sprintf(" ATENCIÓN: en cada frase, las iniciales de las palabras de contenido, en orden, deben deletrear %s. No omitas ninguna ni añadas otras de otra letra.", block)
		}
		prompt := fmt.Sprintf(`Estas 4 palabras son un mnemónico: sus iniciales, en orden, deletrean una clave (%s). El usuario memorizará las frases que escribas para recordar esa clave.

Tu tarea: escribe TRES frases españolas distintas, memorables y fáciles de imaginar. Cada frase debe usar las 4 palabras EN ESTE ORDEN:
%s

Reglas (para cada frase):
- Puedes cambiar cualquier palabra por otra que empiece por la MISMA letra si la frase gana naturalidad (lo que importa es la inicial).
- Puedes conjugar verbos o cambiar género y número, respetando la inicial.
- Añade solo relleno gramatical: artículos, preposiciones, pronombres, conjunciones (el, la, de, en, con, por, para, y, se, no...). Nada de contenido nuevo.
- Máximo 10 palabras por frase.
- SIN markdown (nada de ** ni *) y SIN comentarios sobre iniciales: solo las tres frases numeradas.

Ejemplo (con otras palabras):
1. El gato negro se durmió en el sofá.

Recuerda: las iniciales de las palabras principales de cada frase, en orden, deben ser exactamente %s.

Responde SOLO las tres frases, numeradas:
1. ...
2. ...
3. ...%s`, block, strings.Join(carriers, ", "), block, extra)
		resp, err := c.generate(ctx, prompt, 220)
		if err != nil {
			return BlockPolish{Carriers: carriers, Fallback: true, Err: err}
		}
		var valid []BlockPolish
		for _, raw := range parseNumbered(resp) {
			s := strings.TrimSpace(strings.Trim(extractQuoted(strings.Trim(raw, "\"")), "\"' \t"))
			s = strings.ReplaceAll(s, "\"", "")
			content, ok2 := initialsOk(s, block)
			if ok2 && len(content) == len(block) {
				valid = append(valid, BlockPolish{Sentence: s, Carriers: content})
			}
		}
		if len(valid) == 0 {
			continue // retry with the stricter reminder
		}
		if len(valid) == 1 {
			return valid[0]
		}
		// pick the most natural among the valid ones
		var lines []string
		for i, v := range valid {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, v.Sentence))
		}
		selPrompt := fmt.Sprintf("¿Qué frase suena más natural y es más fácil de recordar?\n%s\nResponde SOLO el número.", strings.Join(lines, "\n"))
		resp2, err2 := c.generate(ctx, selPrompt, 20)
		if err2 == nil {
			var n int
			if _, err := fmt.Sscanf(strings.TrimSpace(resp2), "%d", &n); err == nil && n >= 1 && n <= len(valid) {
				return valid[n-1]
			}
		}
		return valid[0]
	}
	return BlockPolish{Carriers: carriers, Fallback: true}
}

// parseNumbered extracts the numbered lines ("1. ...") of a response,
// tolerating markdown bold markers and leading asterisks.
func parseNumbered(resp string) []string {
	var out []string
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "**")
		if len(line) > 2 && (line[0] == '1' || line[0] == '2' || line[0] == '3') &&
			(line[1] == '.' || line[1] == ')') {
			out = append(out, strings.TrimSpace(strings.ReplaceAll(line[2:], "**", "")))
		}
	}
	return out
}

// generate posts one prompt to ollama (chat endpoint, thinking
// disabled — qwen3.5-style models otherwise burn the token budget
// reasoning) and returns the response text.
func (c *Client) generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if c.cfg.Model == "" {
		return "", fmt.Errorf("llm: no model configured")
	}
	payload := map[string]any{
		"model":    c.cfg.Model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"stream":   false,
		"think":    false,
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.URL+"/api/chat", bytes.NewReader(body))
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
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Message.Content), nil
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

// initialsOk walks the sentence tokens: whitelisted fillers are skipped,
// every other token is a content word whose initial must match the next
// letter of the block, in order. Returns the content words as they
// appear in the sentence (for the carrier recap).
func initialsOk(sentence, block string) (content []string, ok bool) {
	toks := strings.Fields(strings.ToLower(sentence))
	li := 0 // next block letter index
	for _, tok := range toks {
		tok = strings.Trim(tok, ".,;:!?¡¿\"'«»()[]")
		if tok == "" || fillerWhitelist[strings.ToLower(tok)] {
			continue
		}
		if li >= len(block) {
			return nil, false // too many content words
		}
		r := []rune(tok)
		if len(r) == 0 || r[0] != rune(block[li]) {
			return nil, false
		}
		content = append(content, tok)
		li++
	}
	return content, li == len(block)
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

// fillerWhitelist: the only words the LLM may add around the content
// words (articles, prepositions, conjunctions, clitic/possessive
// pronouns and a few grammatical adverbs). Anything else in the render
// counts as a content word and must match the block's initials.
var fillerWhitelist = map[string]bool{}

func init() {
	fillers := []string{
		"el", "la", "los", "las", "un", "una", "unos", "unas",
		"a", "ante", "bajo", "cabe", "con", "contra", "de", "del", "desde",
		"en", "entre", "hacia", "hasta", "para", "por", "según", "sin",
		"sobre", "tras", "mediante", "durante",
		"y", "e", "o", "u", "que", "pero", "aunque", "si", "como", "mientras",
		"me", "te", "se", "lo", "le", "nos", "os", "les", "mi", "tu", "su",
		"mis", "tus", "sus", "nuestro", "nuestra", "vuestro", "vuestra",
		"este", "esta", "estos", "estas", "ese", "esa", "esos", "esas",
		"aquel", "aquella", "mí", "ti", "él", "ella", "ellos", "ellas",
		"nosotros", "nosotras", "vosotros", "vosotras", "quien", "quienes",
		"cual", "cuales",
		"no", "más", "muy", "tan", "ya", "también", "tampoco", "solo",
		"sólo", "ambos", "ambas", "al", "dél",
	}
	for _, w := range fillers {
		fillerWhitelist[w] = true
	}
}

// PolishBlock renders one block's phrase (public entry for the TUI).
func (c *Client) PolishBlock(ctx context.Context, block string, pool [][]string) BlockPolish {
	return c.polishBlock(ctx, block, pool)
}
