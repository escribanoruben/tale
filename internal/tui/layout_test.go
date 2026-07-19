package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderAppHeader(t *testing.T) {
	text := uiForLang("es")
	header := renderAppHeader(text)
	if !strings.Contains(header, "tale") {
		t.Fatalf("header missing command name:\n%s", header)
	}
	if !strings.Contains(header, "Tales Vault System") {
		t.Fatalf("header missing tagline:\n%s", header)
	}
}

func TestSettingsFocusShowsFullView(t *testing.T) {
	m := NewModel(testSession(t), AppSettings{Language: "en", Phrases: 5}, []string{"en", "es"})
	m.settingsFocus = true
	m.width = 120
	view := m.View()
	if !strings.Contains(view, "Generation settings") {
		t.Fatalf("settings view missing title:\n%s", view)
	}
	if strings.Contains(view, "Password:") {
		t.Fatal("settings view should not show picker chrome")
	}
}

func TestJoinPasswordRowPlacesEntropyBesidePassword(t *testing.T) {
	panel := passwordPanelStyle.Render("Password: abcdefgh")
	entropy := helpStyle.Render("Entropy: ~88.0 bits")

	row := joinPasswordRow(160, panel, entropy)
	if !strings.Contains(row, "~88.0 bits") {
		t.Fatalf("expected entropy in password row:\n%s", row)
	}
}

func TestJoinPasswordRowStacksWhenNarrow(t *testing.T) {
	panel := passwordPanelStyle.Render("Password: abcdefgh")
	entropy := helpStyle.Render("Entropy: ~88.0 bits")

	row := joinPasswordRow(40, panel, entropy)
	if lipgloss.Height(row) <= lipgloss.Height(panel) {
		t.Fatalf("expected stacked password row in narrow layout, height=%d\n%s", lipgloss.Height(row), row)
	}
}

func TestJoinSideBySideCentersShorterColumn(t *testing.T) {
	roller := panelStyle.Render("one\ntwo\nthree\nfour\nfive")
	phrases := "alpha\nbeta"

	combined, _, _, rollerYOff, phrasesYOff := joinSideBySide(roller, phrases)
	if phrasesYOff <= 0 {
		t.Fatalf("expected phrases to be offset downward, phrasesYOff=%d", phrasesYOff)
	}
	if rollerYOff != 0 {
		t.Fatalf("expected roller to stay top-aligned when taller, rollerYOff=%d", rollerYOff)
	}
	if lipgloss.Height(combined) <= lipgloss.Height(phrases) {
		t.Fatal("expected combined height to match taller roller column")
	}
}

func TestSideBySideLayoutPlacesRollerLeft(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.width = 160
	_ = m.View()

	var rollerZone hitZone
	found := false
	for _, z := range m.hits.zones {
		if z.kind == hitRoller && z.index == 0 {
			rollerZone = z
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected roller hit zone")
	}
	if rollerZone.x0 > 5 {
		t.Fatalf("expected roller on the left in wide layout, x0=%d", rollerZone.x0)
	}
}

func TestStackedLayoutKeepsFullWidthRoller(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.width = 60
	_ = m.View()

	for _, z := range m.hits.zones {
		if z.kind == hitRoller && z.index == 0 && z.x1 >= 0 {
			t.Fatalf("expected full-width roller zone in narrow layout, got x0=%d x1=%d", z.x0, z.x1)
		}
	}
}
