package tui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	rollerColumnGap = "  "
	commandName     = "tale"
)

func renderAppHeader(text uiText) string {
	return appTitleStyle.Render(commandName) + helpStyle.Render(" — "+text.commandTagline)
}

func joinPasswordRow(termWidth int, passwordPanel, entropy string) string {
	gap := rollerColumnGap
	if termWidth > 0 && termWidth < lipgloss.Width(passwordPanel)+lipgloss.Width(gap)+lipgloss.Width(entropy) {
		return passwordPanel + "\n" + entropy
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, passwordPanel, gap, entropy)
}

func fitsSideBySide(termWidth int, roller, phrases string) bool {
	if termWidth <= 0 {
		return false
	}
	gap := lipgloss.Width(rollerColumnGap)
	return termWidth >= lipgloss.Width(roller)+gap+lipgloss.Width(phrases)
}

func joinSideBySide(roller, phrases string) (combined string, rollerWidth, gapWidth, rollerYOff, phrasesYOff int) {
	gapWidth = lipgloss.Width(rollerColumnGap)
	rollerWidth = lipgloss.Width(roller)
	combined = lipgloss.JoinHorizontal(lipgloss.Center, roller, rollerColumnGap, phrases)
	combinedHeight := lipgloss.Height(combined)
	rollerYOff = (combinedHeight - lipgloss.Height(roller)) / 2
	phrasesYOff = (combinedHeight - lipgloss.Height(phrases)) / 2
	return combined, rollerWidth, gapWidth, rollerYOff, phrasesYOff
}

// viewContentWidth is the wrap width for full-screen TUI views.
func viewContentWidth(termWidth int) int {
	if termWidth <= 0 {
		return 0
	}
	return termWidth
}

func renderWrapped(style lipgloss.Style, width int, s string) string {
	if width <= 0 {
		return style.Render(s)
	}
	return style.Width(width).Render(s)
}

func offsetHitZonesColumn(zones []hitZone, baseY, baseX, columnWidth int) []hitZone {
	if len(zones) == 0 {
		return nil
	}
	out := make([]hitZone, len(zones))
	for i, z := range zones {
		nz := z
		nz.y0 += baseY
		nz.y1 += baseY
		if baseX > 0 && columnWidth > 0 {
			if z.x1 < 0 {
				nz.x0 = baseX
				nz.x1 = baseX + columnWidth - 1
			} else {
				nz.x0 += baseX
				nz.x1 += baseX
			}
		}
		out[i] = nz
	}
	return out
}
