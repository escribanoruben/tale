package tui

import (
	"testing"

	"github.com/ruben/tales_vault_system/internal/dictionary"
	"github.com/ruben/tales_vault_system/internal/normalize"
	"github.com/ruben/tales_vault_system/internal/profile"
)

func TestSessionRandomFill(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	pass := "abcd" + "efgh" + "ilmn" + "oprs" + "tuvw"
	s, err := NewSession(p, pass, dict, false, true)
	if err != nil {
		t.Fatal(err)
	}

	used := make(map[string]bool)
	for bi, block := range s.Blocks {
		for si, w := range s.Selections[bi] {
			if w == "" {
				t.Fatalf("empty selection at block %d slot %d", bi, si)
			}
			key := normalize.Key(w)
			if used[key] {
				t.Fatalf("repeated word %q", w)
			}
			used[key] = true
			if normalize.Initial(w) != string(block[si]) {
				t.Fatalf("word %q does not match letter %q", w, block[si])
			}
		}
	}
}

func TestSessionNavigatePreservesWord(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	first := s.Selections[0][0]

	if err := s.moveSlotInBlock(1); err != nil {
		t.Fatal(err)
	}
	if err := s.moveSlotInBlock(-1); err != nil {
		t.Fatal(err)
	}
	if s.Selections[0][0] != first {
		t.Fatalf("expected word %q preserved after navigation, got %q", first, s.Selections[0][0])
	}
}

func TestMoveSlotCrossesPhrasesAndWraps(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcdefgh", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Blocks) != 2 {
		t.Fatalf("blocks = %d, want 2", len(s.Blocks))
	}

	s.BlockIndex = 0
	s.SlotIndex = 3
	if err := s.moveSlotInBlock(1); err != nil {
		t.Fatal(err)
	}
	if s.BlockIndex != 1 || s.SlotIndex != 0 {
		t.Fatalf("after next at end: block=%d slot=%d, want 1/0", s.BlockIndex, s.SlotIndex)
	}

	s.BlockIndex = 1
	s.SlotIndex = 3
	if err := s.moveSlotInBlock(1); err != nil {
		t.Fatal(err)
	}
	if s.BlockIndex != 0 || s.SlotIndex != 0 {
		t.Fatalf("after wrap forward: block=%d slot=%d, want 0/0", s.BlockIndex, s.SlotIndex)
	}

	s.BlockIndex = 0
	s.SlotIndex = 0
	if err := s.moveSlotInBlock(-1); err != nil {
		t.Fatal(err)
	}
	if s.BlockIndex != 1 || s.SlotIndex != 3 {
		t.Fatalf("after wrap backward: block=%d slot=%d, want 1/3", s.BlockIndex, s.SlotIndex)
	}

	s.BlockIndex = 1
	s.SlotIndex = 0
	if err := s.moveSlotInBlock(-1); err != nil {
		t.Fatal(err)
	}
	if s.BlockIndex != 0 || s.SlotIndex != 3 {
		t.Fatalf("after prev at start: block=%d slot=%d, want 0/3", s.BlockIndex, s.SlotIndex)
	}
}

func TestSessionRandomizeWord(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	original := s.Selections[0][0]

	if err := s.randomizeWord(); err != nil {
		t.Fatal(err)
	}
	if s.Selections[0][0] == original {
		t.Fatal("expected word to change after randomize")
	}
	if s.Password != "abcd" {
		t.Fatalf("password changed to %q, want abcd", s.Password)
	}
}

func TestSessionRandomizePhrase(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]string(nil), s.Selections[0]...)

	if err := s.randomizePhrase(); err != nil {
		t.Fatal(err)
	}
	if joinWords(s.Selections[0]) == joinWords(before) {
		t.Fatal("expected phrase to change")
	}
	if s.Password != "abcd" {
		t.Fatalf("password changed to %q, want abcd", s.Password)
	}
}

func TestSessionChangeLetter(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.changeLetter('m'); err != nil {
		t.Fatal(err)
	}
	if s.Password != "mbcd" {
		t.Fatalf("password = %q, want mbcd", s.Password)
	}
	if normalize.Initial(s.Selections[0][0]) != "m" {
		t.Fatalf("word %q should start with m", s.Selections[0][0])
	}
}

func joinWords(words []string) string {
	out := ""
	for i, w := range words {
		if i > 0 {
			out += " "
		}
		out += w
	}
	return out
}

func TestSessionMoveBlock(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd"+"efgh", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	firstBlockWord := s.Selections[0][0]

	if err := s.moveBlock(1); err != nil {
		t.Fatal(err)
	}
	if s.BlockIndex != 1 {
		t.Fatalf("block index = %d, want 1", s.BlockIndex)
	}
	if err := s.moveBlock(-1); err != nil {
		t.Fatal(err)
	}
	if s.BlockIndex != 0 || s.Selections[0][0] != firstBlockWord {
		t.Fatal("expected to return to first block preserving selections")
	}
}

func TestSessionChangeWord(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd", dict, false, true)
	if err != nil {
		t.Fatal(err)
	}
	original := s.Selections[0][0]
	if len(s.Candidates) < 2 {
		t.Skip("not enough candidates to test word change")
	}

	s.previewRoller(1)
	s.applySelection()
	if s.Selections[0][0] == original {
		t.Fatal("expected word to change after move")
	}
	if s.Used[normalize.Key(original)] {
		t.Fatal("expected previous word to be freed after change")
	}
	if !s.Used[normalize.Key(s.Selections[0][0])] {
		t.Fatal("expected new word to be marked used")
	}
}
