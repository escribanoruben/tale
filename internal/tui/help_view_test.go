package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/escribanoruben/tale/internal/profile"
)

func TestHelpViewIncludesShortcuts(t *testing.T) {
	text := uiFor(profile.Profile{Language: "en"})
	view := renderHelpView(text, 0)
	for _, want := range []string{"Shortcuts", "h/l", "Enter", "Esc"} {
		if !strings.Contains(view, want) {
			t.Fatalf("help view missing %q:\n%s", want, view)
		}
	}
}

func TestHelpViewFitsNarrowWidth(t *testing.T) {
	text := uiFor(profile.Profile{Language: "es"})
	view := renderHelpView(text, 58)
	if lipgloss.Width(view) > 58 {
		t.Fatalf("help view wider than terminal: got %d want <= 58\n%s", lipgloss.Width(view), view)
	}
	if !strings.Contains(view, "Enter") {
		t.Fatalf("wrapped help view missing shortcuts:\n%s", view)
	}
}

func TestViewHidesFooterHelpLine(t *testing.T) {
	m := NewModel(testSession(t), AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	view := m.View()
	if strings.Contains(view, "F1") {
		t.Fatal("view should not show legacy footer help")
	}
	if strings.Contains(view, "h/l · slot") && strings.Contains(view, "j/k · phrase") {
		t.Fatal("shortcut bar should not appear in main view")
	}
}

func TestHelpModeShowsFullView(t *testing.T) {
	m := NewModel(testSession(t), AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.mode = modeHelp
	view := m.View()
	if !strings.Contains(view, "Shortcuts") {
		t.Fatalf("help mode should show help view:\n%s", view)
	}
	if strings.Contains(view, "Password:") {
		t.Fatal("help view should not show picker chrome")
	}
}
