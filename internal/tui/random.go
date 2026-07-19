package tui

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/escribanoruben/tales_vault_system/internal/normalize"
	"github.com/escribanoruben/tales_vault_system/internal/phrase"
)

type uiMode int

const (
	modeNormal uiMode = iota
	modeRandomMenu
	modeChangeLetter
	modeHelp
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
