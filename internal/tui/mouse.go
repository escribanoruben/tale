package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const doubleClickInterval = 450 * time.Millisecond

type hitKind int

const (
	hitSettingsPanel hitKind = iota
	hitSettingsField
	hitSettingsPick
	hitSettingsConfirm
	hitRoller
	hitBlock
	hitSlot
	hitCurrentWord
	hitRollerPanel
	hitOverlay
)

type hitZone struct {
	y0, y1 int
	x0, x1 int // x1 < 0 means full row width
	kind   hitKind
	index  int // block / field / offset / option
	extra  int // slot index for hitSlot
}

type hitTargets struct {
	zones []hitZone
}

func fullRow(y0, y1 int, kind hitKind, index int) hitZone {
	return hitZone{y0: y0, y1: y1, x0: 0, x1: -1, kind: kind, index: index}
}

func (h *hitTargets) reset() {
	if h == nil {
		return
	}
	h.zones = h.zones[:0]
}

func (h *hitTargets) add(y0, y1 int, kind hitKind, index int) {
	if h == nil || y1 < y0 {
		return
	}
	h.zones = append(h.zones, fullRow(y0, y1, kind, index))
}

func (h *hitTargets) addXY(y, x0, x1 int, kind hitKind, index, extra int) {
	if h == nil || x1 < x0 {
		return
	}
	h.zones = append(h.zones, hitZone{y0: y, y1: y, x0: x0, x1: x1, kind: kind, index: index, extra: extra})
}

func (h *hitTargets) addOffset(baseY int, zones []hitZone) {
	if h == nil {
		return
	}
	for _, z := range zones {
		h.zones = append(h.zones, hitZone{
			y0:    baseY + z.y0,
			y1:    baseY + z.y1,
			x0:    z.x0,
			x1:    z.x1,
			kind:  z.kind,
			index: z.index,
			extra: z.extra,
		})
	}
}

// at returns the topmost hit zone containing (x, y). Later zones win.
// Zones with x1 < 0 match any column.
func (h *hitTargets) at(x, y int) (hitZone, bool) {
	if h == nil {
		return hitZone{}, false
	}
	for i := len(h.zones) - 1; i >= 0; i-- {
		z := h.zones[i]
		if y < z.y0 || y > z.y1 {
			continue
		}
		if z.x1 >= 0 && (x < z.x0 || x > z.x1) {
			continue
		}
		return z, true
	}
	return hitZone{}, false
}

func visualHeight(s string) int {
	if s == "" {
		return 0
	}
	return lipgloss.Height(s)
}

// viewLineY returns the 0-based row index where the next WriteString would begin.
// It matches strings.Split(view, "\n") line numbering used by the terminal.
func viewLineY(b *strings.Builder) int {
	return strings.Count(b.String(), "\n")
}

func wheelDelta(msg tea.MouseMsg) int {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		return -1
	case tea.MouseButtonWheelDown:
		return 1
	default:
		return 0
	}
}

func isLeftClick(msg tea.MouseMsg) bool {
	return msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
}

type clickTracker struct {
	zone hitZone
	ok   bool
	at   time.Time
}

func (z hitZone) sameTarget(other hitZone) bool {
	return z.kind == other.kind && z.index == other.index && z.extra == other.extra
}

func (c *clickTracker) note(zone hitZone, ok bool, now time.Time) bool {
	if !ok {
		c.ok = false
		return false
	}
	double := c.ok && c.zone.sameTarget(zone) && now.Sub(c.at) <= doubleClickInterval
	c.zone = zone
	c.ok = true
	c.at = now
	return double
}

func mouseProgramOpts() []tea.ProgramOption {
	return []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
}
