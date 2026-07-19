package tui

import (
	"strings"
	"testing"
)

func TestHitZonesAlignWithViewLines(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.beginPicker()
	view := m.View()
	lines := strings.Split(view, "\n")

	for _, z := range m.hits.zones {
		if z.y0 < 0 || z.y0 >= len(lines) {
			t.Errorf("zone kind=%d y0=%d out of range (view lines=%d)", z.kind, z.y0, len(lines))
			continue
		}
		line := lines[z.y0]
		switch z.kind {
		case hitRoller:
			if z.index == 0 && !strings.Contains(line, "▸") {
				t.Errorf("roller center at y=%d expected selection marker, got: %q", z.y0, line)
			}
		case hitSlot:
			word := session.Selections[z.index][z.extra]
			if !strings.Contains(line, word) {
				t.Errorf("slot hit at y=%d expected word %q, got: %q", z.y0, word, line)
			}
		case hitBlock:
			block := session.Blocks[z.index]
			if !strings.Contains(line, block+" →") {
				t.Errorf("block hit at y=%d expected block %q, got: %q", z.y0, block, line)
			}
		}
	}
}

func TestHitZonesAlignWithRandomOverlay(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.mode = modeRandomMenu
	view := m.View()
	lines := strings.Split(view, "\n")

	for _, z := range m.hits.zones {
		if z.kind != hitOverlay {
			continue
		}
		if z.y0 < 0 || z.y0 >= len(lines) {
			t.Fatalf("overlay zone y0=%d out of range (view lines=%d)", z.y0, len(lines))
		}
		if !strings.Contains(lines[z.y0], "►") && !strings.Contains(lines[z.y0], "  ") {
			t.Fatalf("overlay zone at y=%d does not look like a menu row: %q", z.y0, lines[z.y0])
		}
	}
}
