package tui

import "strings"

func renderHelpView(text uiText, termWidth int) string {
	innerW := viewContentWidth(termWidth)
	var lines []string
	lines = append(lines, renderAppHeader(text))
	lines = append(lines, "")
	for _, p := range text.helpBody {
		lines = append(lines, renderWrapped(labelStyle, innerW, p))
		lines = append(lines, "")
	}
	lines = append(lines, titleStyle.Render(text.helpShortcutsTitle))
	lines = append(lines, "")
	for _, line := range text.helpShortcuts {
		lines = append(lines, renderWrapped(helpStyle, innerW, line))
	}
	lines = append(lines, "")
	lines = append(lines, renderWrapped(helpStyle, innerW, text.helpDismiss))
	return strings.Join(lines, "\n")
}
