package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderRollerPanel(s *Session, text uiText) (string, []hitZone) {
	p := blockPaletteFor(s.BlockIndex)
	border := panelStyle.BorderForeground(lipgloss.Color(p.active))

	var lines []string
	var hits []hitZone

	if len(s.Candidates) == 0 {
		lines = append(lines, warnStyle.Render(text.noWords))
	} else {
		for i, rl := range strings.Split(renderRoller(s.Candidates, s.RollerIndex, s.BlockIndex), "\n") {
			lines = append(lines, rl)
			hits = append(hits, fullRow(i, i, hitRoller, i-visibleRadius))
		}
	}

	rendered := border.Render(strings.Join(lines, "\n"))
	h := visualHeight(rendered)
	out := []hitZone{fullRow(0, h-1, hitRollerPanel, 0)}
	for _, z := range hits {
		out = append(out, fullRow(z.y0+1, z.y1+1, z.kind, z.index))
	}
	return rendered, out
}
