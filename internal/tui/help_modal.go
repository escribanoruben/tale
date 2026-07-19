package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderHelpModal(text uiText) string {
	border := panelStyle.BorderForeground(lipgloss.Color("117"))
	var lines []string
	lines = append(lines, appTitleStyle.Render(text.helpTitle))
	lines = append(lines, "")
	for _, p := range text.helpBody {
		lines = append(lines, labelStyle.Render(p))
		lines = append(lines, "")
	}
	lines = append(lines, helpStyle.Render(text.helpDismiss))
	return border.Render(strings.Join(lines, "\n"))
}
