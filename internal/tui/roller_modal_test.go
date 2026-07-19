package tui

import (
	"strings"
	"testing"

	"github.com/escribanoruben/tale/internal/dictionary"
	"github.com/escribanoruben/tale/internal/profile"
)

func TestSessionCycleWord(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSession(p, "abcd", dict, true)
	if err != nil {
		t.Fatal(err)
	}
	original := s.Selections[0][0]
	if len(s.Candidates) < 2 {
		t.Skip("not enough candidates to test word change")
	}

	s.cycleWord(1)
	if s.Selections[0][0] == original {
		t.Fatal("expected word to change after cycleWord")
	}
}

func TestRollerPanelShowsRoller(t *testing.T) {
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}
	pass := "abcd" + "efgh" + "ilmn" + "oprs" + "tuvw"
	s, err := NewSession(p, pass, dict, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.refreshCandidates(); err != nil {
		t.Fatal(err)
	}

	text := uiFor(p)
	panel, hits := renderRollerPanel(s, text)
	if !strings.Contains(panel, "▸") {
		t.Fatalf("panel missing roller selection marker:\n%s", panel)
	}
	foundRoller := false
	for _, z := range hits {
		if z.kind == hitRoller {
			foundRoller = true
			break
		}
	}
	if !foundRoller {
		t.Fatal("expected roller hit zones")
	}
}

func TestViewAlwaysShowsRollerPanel(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	view := m.View()
	if !strings.Contains(view, "▸") {
		t.Fatalf("view missing always-visible roller:\n%s", view)
	}
}
