package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestQuitConfirmEscDoesNotQuitImmediately(t *testing.T) {
	m := NewModel(testSession(t), AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("expected no quit command on first Esc")
	}
	if m.mode != modeQuitConfirm {
		t.Fatalf("mode = %v, want quit confirm", m.mode)
	}
	if m.session.Quitting {
		t.Fatal("session should not be quitting yet")
	}
}

func TestQuitConfirmYesQuits(t *testing.T) {
	m := NewModel(testSession(t), AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.mode = modeQuitConfirm
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !m.session.Quitting {
		t.Fatal("session should be quitting")
	}
}

func TestQuitConfirmNoReturnsToPicker(t *testing.T) {
	m := NewModel(testSession(t), AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.mode = modeQuitConfirm
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.mode != modeNormal {
		t.Fatalf("mode = %v, want normal", m.mode)
	}
	if strings.Contains(m.View(), "Quit and lose") {
		t.Fatal("quit prompt should be hidden after cancel")
	}
}
