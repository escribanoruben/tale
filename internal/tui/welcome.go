package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const welcomeChoiceCount = 2

const (
	welcomeChoiceDefaults = iota
	welcomeChoiceCustomize
)

type welcomeModel struct {
	defaults AppSettings
	focus    int
}

func newWelcomeModel(defaults AppSettings) welcomeModel {
	return welcomeModel{defaults: defaults}
}

func (w *welcomeModel) moveFocus(delta int) {
	w.focus = (w.focus + delta + welcomeChoiceCount) % welcomeChoiceCount
}

func (w welcomeModel) choice() int {
	return w.focus
}

func (w welcomeModel) View(hits *hitTargets) string {
	if hits != nil {
		hits.reset()
	}
	text := uiForLang(w.defaults.Language)
	var b strings.Builder

	b.WriteString(renderAppHeader(text))
	b.WriteString("\n\n")

	innerW := viewContentWidth(0)
	var lines []string
	var contentHits []hitZone
	lines = append(lines, appTitleStyle.Render(text.welcomeTitle))
	lines = append(lines, "")
	for _, p := range text.welcomeBody {
		lines = append(lines, renderWrapped(labelStyle, innerW, p))
	}
	lines = append(lines, "")

	y := len(lines)
	options := []string{
		w.defaultsOptionLabel(text),
		text.welcomeCustomizeOption,
	}
	for i, opt := range options {
		focused := i == w.focus
		lines = append(lines, welcomeOptionLine(focused, innerW, opt))
		contentHits = append(contentHits, hitZone{
			y0: y, y1: y, x0: 0, x1: -1,
			kind: hitWelcomeChoice, index: i,
		})
		y++
	}
	lines = append(lines, "", renderWrapped(helpStyle, innerW, text.welcomeHelp))

	body := strings.Join(lines, "\n")
	panel := panelStyle.BorderForeground(lipgloss.Color("205")).Render(body)
	if hits != nil {
		hits.addOffset(viewLineY(&b), offsetHitZonesColumn(contentHits, 1, 1, 0))
	}
	b.WriteString(panel)
	return b.String()
}

func (w welcomeModel) defaultsOptionLabel(text uiText) string {
	return fmt.Sprintf(text.welcomeDefaultsOption,
		languageOptionLabel(w.defaults.Language),
		w.defaults.Phrases,
	)
}

func welcomeOptionLine(focused bool, innerW int, label string) string {
	prefix := "  "
	if focused {
		prefix = "> "
	}
	line := prefix + label
	if focused {
		return renderWrapped(activeStyle, innerW, line)
	}
	return renderWrapped(labelStyle, innerW, line)
}

func (w *welcomeModel) update(msg tea.KeyMsg) {
	switch msg.String() {
	case "j", "down":
		w.moveFocus(1)
	case "k", "up":
		w.moveFocus(-1)
	}
}
