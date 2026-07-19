package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSettingsViewShowsAllOptions(t *testing.T) {
	ctrl := newSettingsController(AppSettings{Language: "en", Phrases: 5})
	ctrl.syncFocusFromDraft([]string{"en", "es"})
	view, _ := renderSettingsView(AppSettings{Language: "en", Phrases: 5}, []string{"en", "es"}, ctrl, 100)
	for _, want := range []string{"Language", "English", "Español", "Phrases", "No adjacent", "typing fast"} {
		if !strings.Contains(view, want) {
			t.Fatalf("settings view missing %q:\n%s", want, view)
		}
	}
	if !strings.Contains(view, "[x]") || !strings.Contains(view, "[ ]") {
		t.Fatalf("expected checkbox markers in settings view:\n%s", view)
	}
}

func TestSettingsMoveFocusAndSelect(t *testing.T) {
	ctrl := newSettingsController(AppSettings{Language: "en", Phrases: 5})
	ctrl.syncFocusFromDraft([]string{"en", "es"})
	languages := []string{"en", "es"}

	applied := AppSettings{Language: "en", Phrases: 5}
	for ctrl.draft.Language != "es" {
		ctrl.moveFocus(1, languages, applied)
		ctrl.selectFocused(languages)
	}
	if ctrl.draft.Language != "es" {
		t.Fatalf("language = %q, want es", ctrl.draft.Language)
	}
}

func TestSettingsViewShowsActionRowsWhenPending(t *testing.T) {
	applied := AppSettings{Language: "en", Phrases: 5}
	ctrl := newSettingsController(applied)
	ctrl.syncFocusFromDraft([]string{"en", "es"})
	ctrl.draft.Phrases = 6
	view, _ := renderSettingsView(applied, []string{"en", "es"}, ctrl, 100)
	for _, want := range []string{"Apply changes", "Discard and go back"} {
		if !strings.Contains(view, want) {
			t.Fatalf("settings view missing action %q:\n%s", want, view)
		}
	}
}

func TestSettingsViewHidesActionRowsWithoutChanges(t *testing.T) {
	applied := AppSettings{Language: "en", Phrases: 5}
	ctrl := newSettingsController(applied)
	ctrl.syncFocusFromDraft([]string{"en", "es"})
	view, _ := renderSettingsView(applied, []string{"en", "es"}, ctrl, 100)
	if strings.Contains(view, "Apply changes") {
		t.Fatal("settings view should not show actions without pending changes")
	}
}

func TestSettingsEscDiscardsWithoutApply(t *testing.T) {
	applied := AppSettings{Language: "en", Phrases: 5}
	ctrl := newSettingsController(applied)
	ctrl.draft.Phrases = 6
	exit := handleSettingsKey(&ctrl, tea.KeyMsg{Type: tea.KeyEsc}, []string{"en", "es"}, applied)
	if exit != settingsExitUnfocus {
		t.Fatalf("esc exit = %v, want unfocus", exit)
	}
}

func TestSettingsActionApplyCommits(t *testing.T) {
	applied := AppSettings{Language: "en", Phrases: 5}
	ctrl := newSettingsController(applied)
	ctrl.draft.Phrases = 6
	ctrl.focusActionRow(settingsActionApply, []string{"en", "es"}, applied)
	exit := handleSettingsKey(&ctrl, tea.KeyMsg{Type: tea.KeySpace}, []string{"en", "es"}, applied)
	if exit != settingsExitApply {
		t.Fatalf("apply exit = %v, want apply", exit)
	}
}
