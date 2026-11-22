package tui

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/escribanoruben/tale/internal/normalize"
	"github.com/escribanoruben/tale/internal/phrase"
)

type uiMode int

const (
	modeNormal uiMode = iota
	modeRandomMenu
	modeChangeLetter
	modeHelp
	modeQuitConfirm
)

func (s *Session) phraseOpts() phrase.Options {
	return phrase.Options{AllowRepeat: false}
}

func (s *Session) rebuildUsed() {
	used := make(map[string]bool)
	for bi := range s.Selections {
		for _, w := range s.Selections[bi] {
			if w != "" {
				used[normalize.Key(w)] = true
			}
		}
	}
	s.Used = used
}

func (s *Session) randomizeScope(scope string) error {
	s.Err = nil
	if scope == "l" {
		return s.llmReroll()
	}
	switch scope {
	case "w":
		return s.randomizeWord()
	case "p":
		return s.randomizePhrase()
	case "a":
		return s.randomizeAll()
	default:
		return nil
	}
}

// llmReroll re-renders the current block's phrase with the LLM polish
// layer (story mode + llm_polish enabled). The render may substitute
// words; the initials pipeline validates it, so the password always
// recovers.
func (s *Session) llmReroll() error {
	if s.LLM == nil {
		return fmt.Errorf("el re-render con IA requiere llm_polish = true en settings")
	}
	if !s.Storyish || len(s.Pools) == 0 {
		return fmt.Errorf("el re-render con IA está disponible en el modo historia")
	}
	// the current selection leads the pool so the LLM re-renders what is
	// on screen; the rest of the pool offers variety on retries
	poolView := [][]string{s.Selections[s.BlockIndex]}
	seen := map[string]bool{normalize.Key(s.Selections[s.BlockIndex][0]): true}
	for _, c := range s.Pools[s.BlockIndex] {
		if len(c) != len(s.Selections[s.BlockIndex]) || seen[normalize.Key(c[0])] {
			continue
		}
		seen[normalize.Key(c[0])] = true
		poolView = append(poolView, c)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pb := s.LLM.PolishBlock(ctx, s.Blocks[s.BlockIndex], poolView)
	if pb.Fallback || len(pb.Carriers) != len(s.Selections[s.BlockIndex]) {
		return fmt.Errorf("la IA no generó un render válido para este bloque; prueba otra vez")
	}
	for si, w := range pb.Carriers {
		s.Selections[s.BlockIndex][si] = w
	}
	s.rebuildUsed()
	return s.refreshCandidates()
}

// semanticRerollPhrase picks a random candidate from the block's
// semantic pool, skipping used words; falls back to the uniform phrase
// when the pool is exhausted.
func (s *Session) semanticRerollPhrase() error {
	used := s.usedExceptBlock(s.BlockIndex)
	var options [][]string
	for _, c := range s.Pools[s.BlockIndex] {
		ok := true
		for _, w := range c {
			if used[normalize.Key(w)] {
				ok = false
				break
			}
		}
		if ok {
			options = append(options, c)
		}
	}
	if len(options) == 0 {
		return s.randomizePhrase()
	}
	idx, err := randIndex(len(options))
	if err != nil {
		return err
	}
	for si, w := range options[idx] {
		s.Selections[s.BlockIndex][si] = w
	}
	s.rebuildUsed()
	return s.refreshCandidates()
}

// semanticRerollAll applies the semantic reroll to every block.
func (s *Session) semanticRerollAll() error {
	for range s.Selections {
		if err := s.semanticRerollPhrase(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) randomizeWord() error {
	if err := s.refreshCandidates(); err != nil {
		return err
	}
	if len(s.Candidates) == 0 {
		return fmt.Errorf("no words available for this slot")
	}
	idx, err := randIndex(len(s.Candidates))
	if err != nil {
		return err
	}
	s.RollerIndex = idx
	s.applySelection()
	return nil
}

func (s *Session) randomizePhrase() error {
	if s.Storyish && len(s.Pools) > 0 {
		return s.semanticRerollPhrase()
	}
	used := s.usedExceptBlock(s.BlockIndex)
	words, err := phrase.Random(s.Profile, s.Blocks[s.BlockIndex], s.dict, used, s.phraseOpts())
	if err != nil {
		return err
	}
	for si, w := range words {
		s.Selections[s.BlockIndex][si] = w
	}
	s.rebuildUsed()
	if err := s.refreshCandidates(); err != nil {
		return err
	}
	s.syncRollerToSelection()
	return nil
}

func (s *Session) randomizeAll() error {
	if s.Storyish && len(s.Pools) > 0 {
		return s.semanticRerollAll()
	}
	results, err := phrase.RandomForPassword(s.Profile, s.Password, s.dict, s.phraseOpts())
	if err != nil {
		return err
	}
	for i, r := range results {
		copy(s.Selections[i], r.Words)
	}
	s.rebuildUsed()
	if err := s.refreshCandidates(); err != nil {
		return err
	}
	s.syncRollerToSelection()
	return nil
}

func (s *Session) usedExceptBlock(blockIndex int) map[string]bool {
	used := make(map[string]bool)
	for bi, sel := range s.Selections {
		if bi == blockIndex {
			continue
		}
		for _, w := range sel {
			if w != "" {
				used[normalize.Key(w)] = true
			}
		}
	}
	return used
}

func (s *Session) changeLetter(letter rune) error {
	if !strings.ContainsRune(s.Profile.Alphabet, letter) {
		return fmt.Errorf("letter %q is not in the profile alphabet", letter)
	}
	flat := s.BlockIndex*len(s.Profile.PhraseTemplate) + s.SlotIndex
	if flat < 0 || flat >= len(s.Password) {
		return fmt.Errorf("slot out of range")
	}

	pass := []byte(s.Password)
	pass[flat] = byte(letter)
	s.Password = string(pass)

	blocks, err := splitBlocks(s.Password, s.Profile.BlockSize)
	if err != nil {
		return err
	}
	s.Blocks = blocks

	if err := s.refreshCandidates(); err != nil {
		return err
	}
	if len(s.Candidates) == 0 {
		return fmt.Errorf("no words available for letter %q", string(letter))
	}
	return s.randomizeWord()
}

func randIndex(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("empty choice set")
	}
	max := big.NewInt(int64(n))
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}
