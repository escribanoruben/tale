package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/escribanoruben/tales_vault_system/internal/dictionary"
	"github.com/escribanoruben/tales_vault_system/internal/profile"
)

func TestAppModelNormalNavigation(t *testing.T) {
	session := testSession(t)
	app := AppModel{
		phase:     appPhasePicker,
		settings:  AppSettings{Language: "en", Phrases: 1},
		languages: []string{"en"},
		picker:    NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"}),
	}
	before := app.picker.session.SlotIndex

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	app = updated.(AppModel)
	if app.picker.session.SlotIndex == before {
		t.Fatalf("expected normal-mode slot navigation, slot still %d", before)
	}
	if app.picker.mode != modeNormal {
		t.Fatalf("expected normal mode, mode=%v", app.picker.mode)
	}
}

func TestAppModelPickerNavigation(t *testing.T) {
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
	app := AppModel{
		phase:     appPhasePicker,
		settings:  AppSettings{Language: "en", Phrases: 5},
		languages: []string{"en"},
		picker:    NewModel(session, AppSettings{Language: "en", Phrases: 5}, []string{"en"}),
	}
	before := app.picker.session.SlotIndex

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	app = updated.(AppModel)
	if app.picker.session.SlotIndex == before {
		t.Fatalf("expected slot navigation through AppModel, slot still %d", before)
	}
}
