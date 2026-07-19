package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type setupModel struct {
	languages []string
	ctrl      settingsController
}

func newSetupModel(languages []string, initial AppSettings) setupModel {
	ctrl := newSettingsController(initial)
	ctrl.syncFocusFromDraft(languages)
	return setupModel{
		languages: languages,
		ctrl:      ctrl,
	}
}

func (s *setupModel) toSettings() AppSettings {
	s.ctrl.draft.Phrases = snapPhraseCount(s.ctrl.draft.Phrases)
	return s.ctrl.draft
}

func (s *setupModel) update(msg tea.KeyMsg) {
	handleSettingsKey(&s.ctrl, msg, s.languages, s.ctrl.draft)
}

func (s setupModel) View(hits *hitTargets) string {
	if hits != nil {
		hits.reset()
	}
	text := uiForLang(s.ctrl.draft.Language)
	var b strings.Builder

	b.WriteString(renderAppHeader(text))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(text.setupSubtitle))
	b.WriteString("\n\n")

	widget, zones := renderSettingsWidget(s.ctrl.draft, s.languages, s.ctrl, true)
	if hits != nil {
		hits.addOffset(viewLineY(&b), zones)
	}
	b.WriteString(widget)

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render(text.setupHelp))
	return b.String()
}
