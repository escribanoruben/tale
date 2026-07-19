package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ruben/tales_vault_system/internal/dictionary"
	"github.com/ruben/tales_vault_system/internal/profile"
)

func testSession(t *testing.T) *Session {
	t.Helper()
	p, err := profile.Load(profile.Files, "en")
	if err != nil {
		t.Fatal(err)
	}
	dict, err := dictionary.Default()
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSession(p, "abcd", dict, true, true)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestHitTargetsTopmostWins(t *testing.T) {
	h := &hitTargets{}
	h.add(0, 5, hitSettingsPanel, 0)
	h.add(2, 2, hitSettingsField, 1)

	z, ok := h.at(3, 2)
	if !ok {
		t.Fatal("expected hit")
	}
	if z.kind != hitSettingsField || z.index != 1 {
		t.Fatalf("got kind=%v index=%d", z.kind, z.index)
	}
}

func TestHitTargetsRespectsX(t *testing.T) {
	h := &hitTargets{}
	h.addXY(1, 10, 14, hitSlot, 0, 2)
	if _, ok := h.at(5, 1); ok {
		t.Fatal("expected miss left of word")
	}
	z, ok := h.at(12, 1)
	if !ok || z.kind != hitSlot || z.extra != 2 {
		t.Fatalf("expected slot hit, got ok=%v zone=%+v", ok, z)
	}
}

func TestWheelDelta(t *testing.T) {
	if wheelDelta(tea.MouseMsg{Button: tea.MouseButtonWheelUp}) != -1 {
		t.Fatal("wheel up should be -1")
	}
	if wheelDelta(tea.MouseMsg{Button: tea.MouseButtonWheelDown}) != 1 {
		t.Fatal("wheel down should be 1")
	}
	if wheelDelta(tea.MouseMsg{Button: tea.MouseButtonLeft}) != 0 {
		t.Fatal("left click is not a wheel delta")
	}
}

func TestHandleMouseWheelOpensPicker(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	_ = m.View()

	var target hitZone
	for _, z := range m.hits.zones {
		if z.kind == hitSlot && z.index == 0 && z.extra == 1 {
			target = z
			break
		}
	}
	if target.kind != hitSlot {
		t.Fatal("expected slot hit zone")
	}

	updated, _ := m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      target.x0,
		Y:      target.y0,
	})
	m = updated.(Model)
	if m.mode != modePicker {
		t.Fatalf("mode=%v want picker", m.mode)
	}
	if m.session.BlockIndex != 0 || m.session.SlotIndex != 1 {
		t.Fatalf("got block=%d slot=%d, want 0/1", m.session.BlockIndex, m.session.SlotIndex)
	}
}

func TestHandleMouseDoubleClickOpensPicker(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	_ = m.View()

	var target hitZone
	for _, z := range m.hits.zones {
		if z.kind == hitSlot && z.index == 0 && z.extra == 2 {
			target = z
			break
		}
	}
	if target.kind != hitSlot {
		t.Fatal("expected slot hit zone")
	}

	click := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: target.x0, Y: target.y0}
	updated, _ := m.handleMouse(click)
	m = updated.(Model)
	if m.mode == modePicker {
		t.Fatal("single click should not open picker")
	}

	m.click.at = time.Now().Add(-100 * time.Millisecond)
	updated, _ = m.handleMouse(click)
	m = updated.(Model)
	if m.mode != modePicker {
		t.Fatalf("mode=%v want picker after double click", m.mode)
	}
	if m.session.SlotIndex != 2 {
		t.Fatalf("slot=%d want 2", m.session.SlotIndex)
	}
}

func TestHandleMouseDoubleClickOnRollerConfirms(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.beginPicker()
	before := m.session.Selections[0][0]
	_ = m.View()

	var target hitZone
	for _, z := range m.hits.zones {
		if z.kind == hitRoller && z.index == 1 {
			target = z
			break
		}
	}
	if target.kind != hitRoller {
		t.Fatal("expected roller hit zone")
	}

	click := tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: target.x0, Y: target.y0}
	updated, _ := m.handleMouse(click)
	m = updated.(Model)

	m.click.at = time.Now().Add(-100 * time.Millisecond)
	updated, _ = m.handleMouse(click)
	m = updated.(Model)
	if m.mode != modeNormal {
		t.Fatalf("mode=%v want normal after confirm", m.mode)
	}
	if m.session.Selections[0][0] == before && len(m.session.visibleCandidates()) > 1 {
		t.Fatal("expected roller double click to confirm a new word")
	}
}

func TestHandleMouseWheelMovesRoller(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.beginPicker()
	beforeIdx := m.session.RollerIndex
	beforeWord := m.session.Selections[m.session.BlockIndex][m.session.SlotIndex]
	_ = m.View()

	var target hitZone
	for _, z := range m.hits.zones {
		if z.kind == hitRoller && z.index == 0 {
			target = z
			break
		}
	}
	if target.kind != hitRoller {
		t.Fatal("expected roller hit zone")
	}

	updated, _ := m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      target.x0,
		Y:      target.y0,
	})
	m = updated.(Model)
	if len(m.session.visibleCandidates()) > 1 && m.session.RollerIndex == beforeIdx {
		t.Fatal("expected wheel down to move roller preview")
	}
	if m.session.Selections[m.session.BlockIndex][m.session.SlotIndex] != beforeWord {
		t.Fatal("expected wheel in picker to preview only, not apply selection")
	}
}

func TestHandleMouseClickRollerOffset(t *testing.T) {
	session := testSession(t)
	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	m.beginPicker()
	view := m.View()
	if view == "" {
		t.Fatal("empty view")
	}

	var rollerY int
	found := false
	for _, z := range m.hits.zones {
		if z.kind == hitRoller && z.index == 1 {
			rollerY = z.y0
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected roller hit zone for offset +1")
	}

	beforeIdx := m.session.RollerIndex
	updated, _ := m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      rollerY, // X unused for full-row roller hits
		Y:      rollerY,
	})
	m = updated.(Model)
	visible := m.session.visibleCandidates()
	if len(visible) == 0 {
		t.Fatal("no candidates")
	}
	want := (beforeIdx + 1 + len(visible)) % len(visible)
	if m.session.RollerIndex != want {
		t.Fatalf("RollerIndex=%d want %d", m.session.RollerIndex, want)
	}
}

func TestHandleMouseClickPhraseWordSelectsSlot(t *testing.T) {
	session := testSession(t)
	session.BlockIndex = 0
	session.SlotIndex = 0
	if err := session.refreshCandidates(); err != nil {
		t.Fatal(err)
	}
	session.syncRollerToSelection()

	m := NewModel(session, AppSettings{Language: "en", Phrases: 1}, []string{"en"})
	_ = m.View()

	var target hitZone
	found := false
	for _, z := range m.hits.zones {
		if z.kind == hitSlot && z.index == 0 && z.extra == 2 {
			target = z
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected hit zone for phrase word at slot 2")
	}

	updated, _ := m.handleMouse(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      target.x0,
		Y:      target.y0,
	})
	m = updated.(Model)
	if m.session.BlockIndex != 0 || m.session.SlotIndex != 2 {
		t.Fatalf("got block=%d slot=%d, want 0/2", m.session.BlockIndex, m.session.SlotIndex)
	}
}
