package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) rollerOpen() bool {
	return m.mode == modePicker || m.mode == modeSearch
}

func renderRollerModal(s *Session, mode uiMode, text uiText) (string, []hitZone) {
	p := blockPaletteFor(s.BlockIndex)
	border := panelStyle.BorderForeground(lipgloss.Color(p.active))
	visible := s.visibleCandidates()

	var lines []string
	var hits []hitZone
	line := 0

	lines = append(lines, titleStyle.Render(fmt.Sprintf("%s · %s «%s» · %s",
		text.currentWord,
		text.letter,
		s.requiredLetter(),
		posLabel(s.Profile.Language, s.currentPOSLabel()),
	)))
	line++

	if len(visible) == 0 {
		lines = append(lines, warnStyle.Render(text.noWords))
		line++
	} else {
		if mode == modeSearch {
			lines = append(lines, labelStyle.Render("/"+s.SearchQuery+"_")+"  "+
				helpStyle.Render(fmt.Sprintf(text.searchCount, len(visible))))
			line++
		}
		rollerStart := line
		for i, rl := range strings.Split(renderRoller(visible, s.RollerIndex, s.BlockIndex), "\n") {
			lines = append(lines, rl)
			hits = append(hits, fullRow(rollerStart+i, rollerStart+i, hitRoller, i-visibleRadius))
			line++
		}
	}

	help := text.pickerHelp
	if mode == modeSearch {
		help = text.searchHelp
	}
	lines = append(lines, "", helpStyle.Render(help))

	rendered := border.Render(strings.Join(lines, "\n"))
	h := visualHeight(rendered)
	out := []hitZone{fullRow(0, h-1, hitRollerPanel, 0)}
	for _, z := range hits {
		out = append(out, fullRow(z.y0+1, z.y1+1, z.kind, z.index))
	}
	return rendered, out
}
