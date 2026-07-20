package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
)

func TestWelcomeViewShowsChoices(t *testing.T) {
	w := newWelcomeModel(AppSettings{Language: "en", Phrases: 5})
	view := w.View(nil)
	for _, want := range []string{"First run", "Use defaults", "Configure before starting", "Enter · choose"} {
		if !strings.Contains(view, want) {
			t.Fatalf("welcome view missing %q:\n%s", want, view)
		}
	}
}

func TestWelcomeEnterDefaultsStartsPicker(t *testing.T) {
	app := AppModel{
		phase:     appPhaseWelcome,
		settings:  AppSettings{Language: "en", Phrases: 5},
		languages: []string{"en"},
		welcome:   newWelcomeModel(AppSettings{Language: "en", Phrases: 5}),
		cfg: AppConfig{
			LoadProfile: func(language string) (profile.Profile, *dictionary.Dictionary, error) {
				p, err := profile.Load(profile.Files, language)
				if err != nil {
					return profile.Profile{}, nil, err
				}
				dict, err := dictionary.Default()
				if err != nil {
					return profile.Profile{}, nil, err
				}
				return p, dict, nil
			},
		},
	}
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(AppModel)
	if next.phase != appPhasePicker {
		t.Fatalf("phase = %v, want picker", next.phase)
	}
}

func TestWelcomeEnterCustomizeOpensSetup(t *testing.T) {
	app := AppModel{
		phase:     appPhaseWelcome,
		settings:  AppSettings{Language: "en", Phrases: 5},
		languages: []string{"en"},
		welcome:   newWelcomeModel(AppSettings{Language: "en", Phrases: 5}),
	}
	app.welcome.focus = welcomeChoiceCustomize
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(AppModel)
	if next.phase != appPhaseSetup {
		t.Fatalf("phase = %v, want setup", next.phase)
	}
}
