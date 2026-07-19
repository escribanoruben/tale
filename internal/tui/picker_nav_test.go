package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
)

func TestHandlePickerNavigatesSlot(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 5}, []string{"en"})
	before := m.session.SlotIndex

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(Model)
	if m.session.SlotIndex == before {
		t.Fatalf("expected slot to advance from %d", before)
	}
	if m.mode != modeNormal {
		t.Fatalf("expected normal mode, mode=%v", m.mode)
	}
}

func TestEnterFinishesWithCurrentPassword(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.session.Done {
		t.Fatal("expected Enter to mark session done")
	}
	if cmd == nil {
		t.Fatal("expected Enter to quit program")
	}
}

func TestHandlePickerNavigatesBlock(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}
	session, err := NewSession(p, "abcd"+"efgh"+"ilmn"+"oprs"+"tuvw", dict, true)
	if err != nil {
		t.Fatal(err)
	}
	m := NewModel(session, AppSettings{Language: "en", Phrases: 5}, []string{"en"})
	before := m.session.BlockIndex

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.session.BlockIndex == before {
		t.Fatalf("expected block to advance from %d", before)
	}
}
