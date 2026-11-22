package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSetupViewShowsStartAction(t *testing.T) {
	s := newSetupModel([]string{"en", "es"}, AppSettings{Language: "en", Phrases: 5})
	view := s.View(nil)
	if strings.Count(view, "j/k · move") != 1 {
		t.Fatalf("expected single setup help line:\n%s", view)
	}
	for _, want := range []string{"Start", "Generation settings"} {
		if !strings.Contains(view, want) {
			t.Fatalf("setup view missing %q:\n%s", want, view)
		}
	}
}

func TestSetupEnterOnStartBeginsPicker(t *testing.T) {
	setup := newSetupModel([]string{"en"}, AppSettings{Language: "en", Phrases: 5})
	setup.ctrl.focusOnAction = true
	exit := setup.update(tea.KeyMsg{Type: tea.KeyEnter})
	if exit != settingsExitStart {
		t.Fatalf("exit = %v, want start", exit)
	}
}
