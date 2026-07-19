package tui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingsUIState int

const (
	settingsUIList settingsUIState = iota
	settingsUIPick
	settingsUIConfirm
)

type settingsExit int

const (
	settingsExitNone settingsExit = iota
	settingsExitUnfocus
	settingsExitApply
	settingsExitQuitApp
)

type settingsController struct {
	draft        AppSettings
	field        int
	state        settingsUIState
	pickIndex    int
	confirmIndex int // 0 = apply, 1 = discard
}

func newSettingsController(initial AppSettings) settingsController {
	initial.Phrases = snapPhraseCount(initial.Phrases)
	return settingsController{draft: initial}
}

func (c *settingsController) openPicker(languages []string) {
	c.state = settingsUIPick
	c.pickIndex = currentPickIndex(c.field, c.draft, languages)
}

func (c *settingsController) closePicker() {
	c.state = settingsUIList
}

func (c *settingsController) moveField(delta int) {
	c.field = (c.field + delta + settingFieldCount) % settingFieldCount
}

func (c *settingsController) movePick(delta int, languages []string) {
	count := pickOptionCount(c.field, languages)
	if count == 0 {
		return
	}
	c.pickIndex = (c.pickIndex + delta + count) % count
}

func (c *settingsController) confirmPick(languages []string) {
	applyPickToDraft(&c.draft, c.field, c.pickIndex, languages)
	c.closePicker()
}

func pickOptionCount(field int, languages []string) int {
	switch field {
	case settingFieldLanguage:
		return len(languages)
	case settingFieldPhrases:
		return len(commonPhraseCounts)
	case settingFieldNoAdjacent, settingFieldAllowRepeat:
		return 2
	default:
		return 0
	}
}

func currentPickIndex(field int, draft AppSettings, languages []string) int {
	switch field {
	case settingFieldLanguage:
		return languageIndex(languages, draft.Language)
	case settingFieldPhrases:
		return phrasePresetIndex(draft.Phrases)
	case settingFieldNoAdjacent:
		if !draft.AllowAdjacentRepeats {
			return 0
		}
		return 1
	case settingFieldAllowRepeat:
		if draft.AllowRepeat {
			return 0
		}
		return 1
	default:
		return 0
	}
}

func applyPickToDraft(draft *AppSettings, field, pickIndex int, languages []string) {
	switch field {
	case settingFieldLanguage:
		if pickIndex >= 0 && pickIndex < len(languages) {
			draft.Language = languages[pickIndex]
		}
	case settingFieldPhrases:
		if pickIndex >= 0 && pickIndex < len(commonPhraseCounts) {
			draft.Phrases = commonPhraseCounts[pickIndex]
		}
	case settingFieldNoAdjacent:
		draft.AllowAdjacentRepeats = pickIndex != 0
	case settingFieldAllowRepeat:
		draft.AllowRepeat = pickIndex == 0
	}
}

func settingsFieldLabel(text uiText, field int) string {
	switch field {
	case settingFieldLanguage:
		return text.settingsLanguage
	case settingFieldNoAdjacent:
		return text.settingsAllowAdjacent
	case settingFieldPhrases:
		return text.settingsPhrases
	case settingFieldAllowRepeat:
		return text.settingsAllowRepeat
	default:
		return ""
	}
}

func settingsFieldValue(text uiText, field int, draft AppSettings) string {
	switch field {
	case settingFieldLanguage:
		return languageOptionLabel(draft.Language)
	case settingFieldNoAdjacent:
		return boolWord(text, !draft.AllowAdjacentRepeats)
	case settingFieldPhrases:
		preventAdjacent := !draft.AllowAdjacentRepeats
		phrases := snapPhraseCount(draft.Phrases)
		bits := phraseEntropyBits(draft.Language, phrases, preventAdjacent)
		return fmt.Sprintf("%d (~%.0f %s)", phrases, bits, text.entropyBitsUnit)
	case settingFieldAllowRepeat:
		return boolWord(text, draft.AllowRepeat)
	default:
		return ""
	}
}

func boolWord(text uiText, yes bool) string {
	if yes {
		return text.yes
	}
	return text.no
}

func renderSettingsSummary(s AppSettings) string {
	text := uiForLang(s.Language)
	parts := []string{
		fmt.Sprintf("%s %s", text.settingsLanguage, languageOptionLabel(s.Language)),
		fmt.Sprintf("%s %s", text.settingsAllowAdjacent, boolWord(text, !s.AllowAdjacentRepeats)),
		fmt.Sprintf("%s %s", text.settingsPhrases, settingsFieldValue(text, settingFieldPhrases, s)),
		fmt.Sprintf("%s %s", text.settingsAllowRepeat, boolWord(text, s.AllowRepeat)),
	}
	return labelStyle.Render(text.settingsBarPrefix) + " " + joinUIParts(parts, " · ")
}

// renderSettingsSummaryBar draws the compact settings strip (picker, unfocused).
func renderSettingsSummaryBar(s AppSettings) (string, []hitZone) {
	text := uiForLang(s.Language)
	body := renderSettingsSummary(s)
	body += "\n\n" + helpStyle.Render(text.settingsFocusHint)
	rendered := panelStyle.Render(body)
	h := visualHeight(rendered)
	return rendered, []hitZone{fullRow(0, h-1, hitSettingsPanel, 0)}
}

// renderSettingsModal draws the full settings editor as a modal panel.
func renderSettingsModal(s AppSettings, languages []string, ctrl settingsController) (string, []hitZone) {
	text := uiForLang(s.Language)
	style := panelStyle.BorderForeground(lipgloss.Color("205"))
	var lines []string
	var contentHits []hitZone
	lines = append(lines, titleStyle.Render(text.settingsTitle))
	lines = append(lines, "")

	switch ctrl.state {
	case settingsUIPick:
		pickLines, pickHits := settingsPickerLines(text, ctrl, languages, len(lines))
		lines = append(lines, pickLines...)
		contentHits = append(contentHits, pickHits...)
		lines = append(lines, "", helpStyle.Render(text.settingsPickHelp))
	case settingsUIConfirm:
		listLines, _ := settingsListLines(text, ctrl, len(lines))
		lines = append(lines, listLines...)
		lines = append(lines, "")
		kind := settingsChangeKind(s, ctrl.draft)
		confirmLines, confirmHits := settingsConfirmLines(text, ctrl, kind, len(lines))
		lines = append(lines, confirmLines...)
		contentHits = append(contentHits, confirmHits...)
		lines = append(lines, "", helpStyle.Render(text.settingsConfirmHelp))
	default:
		listLines, listHits := settingsListLines(text, ctrl, len(lines))
		lines = append(lines, listLines...)
		contentHits = append(contentHits, listHits...)
		lines = append(lines, "")
		if kind := settingsChangeKind(s, ctrl.draft); kind != settingsChangeNone {
			lines = append(lines, renderSettingsPendingHint(ctrl.draft.Language, kind), "")
		}
		lines = append(lines, helpStyle.Render(text.settingsListHelp))
	}

	rendered := style.Render(strings.Join(lines, "\n"))
	h := visualHeight(rendered)
	out := []hitZone{fullRow(0, h-1, hitSettingsPanel, 0)}
	for _, z := range contentHits {
		out = append(out, fullRow(z.y0+1, z.y1+1, z.kind, z.index))
	}
	return rendered, out
}

// renderSettingsBar draws the compact strip shown above the modal while editing.
func renderSettingsBar(s AppSettings) string {
	text := uiForLang(s.Language)
	body := renderSettingsSummary(s)
	body += "\n\n" + helpStyle.Render(text.settingsModalDismiss)
	return panelStyle.Render(body)
}

// renderSettingsWidget draws the settings panel and returns hit zones relative to the
// rendered panel (row 0 = top border). Used by the first-run setup screen.
func renderSettingsWidget(s AppSettings, languages []string, ctrl settingsController, focused bool) (string, []hitZone) {
	if !focused {
		return renderSettingsSummaryBar(s)
	}
	return renderSettingsModal(s, languages, ctrl)
}

func settingsListLines(text uiText, ctrl settingsController, lineBase int) ([]string, []hitZone) {
	var lines []string
	var hits []hitZone
	for field := 0; field < settingFieldCount; field++ {
		label := settingsFieldLabel(text, field)
		value := settingsFieldValue(text, field, ctrl.draft)
		row := fmt.Sprintf("%s · %s", label, value)
		if field == ctrl.field {
			lines = append(lines, "▸ "+activeStyle.Render(row))
		} else {
			lines = append(lines, "  "+labelStyle.Render(row))
		}
		y := lineBase + field
		hits = append(hits, fullRow(y, y, hitSettingsField, field))
	}
	return lines, hits
}

func settingsPickerLines(text uiText, ctrl settingsController, languages []string, lineBase int) ([]string, []hitZone) {
	var lines []string
	var hits []hitZone
	lines = append(lines, labelStyle.Render(settingsFieldLabel(text, ctrl.field)))

	var options []string
	switch ctrl.field {
	case settingFieldLanguage:
		for _, lang := range languages {
			options = append(options, languageOptionLabel(lang))
		}
	case settingFieldPhrases:
		_, blockSize := profileMetrics(ctrl.draft.Language)
		preventAdjacent := !ctrl.draft.AllowAdjacentRepeats
		for _, count := range commonPhraseCounts {
			passLen := phrasePasswordLength(count, blockSize)
			bits := phraseEntropyBits(ctrl.draft.Language, count, preventAdjacent)
			summary := fmt.Sprintf(text.phraseOptionSummary, count, passLen, math.Round(bits))
			options = append(options, summary+" — "+phraseUseCase(text, count))
		}
	case settingFieldNoAdjacent, settingFieldAllowRepeat:
		options = []string{boolWord(text, true), boolWord(text, false)}
	}
	for i, line := range options {
		lines = append(lines, pickerLine(i == ctrl.pickIndex, line))
		y := lineBase + 1 + i
		hits = append(hits, fullRow(y, y, hitSettingsPick, i))
	}
	return lines, hits
}

func pickerLine(active bool, line string) string {
	if active {
		return "  ► " + activeStyle.Render(line)
	}
	return "    " + labelStyle.Render(line)
}

func settingsConfirmLines(text uiText, ctrl settingsController, kind, lineBase int) ([]string, []hitZone) {
	var lines []string
	var hits []hitZone
	lines = append(lines, titleStyle.Render(text.settingsConfirmTitle))
	optBase := lineBase + 1
	if msg := renderSettingsPendingWarning(ctrl.draft.Language, kind); msg != "" {
		lines = append(lines, msg)
		optBase++
	}
	options := []string{text.settingsConfirmApply, text.settingsConfirmDiscard}
	for i, line := range options {
		lines = append(lines, pickerLine(i == ctrl.confirmIndex, line))
		y := optBase + i
		hits = append(hits, fullRow(y, y, hitSettingsConfirm, i))
	}
	return lines, hits
}

func (c *settingsController) openConfirm() {
	c.state = settingsUIConfirm
	c.confirmIndex = 0
}

func handleSettingsKey(ctrl *settingsController, msg tea.KeyMsg, languages []string, applied AppSettings) settingsExit {
	switch msg.String() {
	case "esc":
		switch ctrl.state {
		case settingsUIPick:
			ctrl.closePicker()
		case settingsUIConfirm:
			ctrl.state = settingsUIList
		case settingsUIList:
			if settingsChangeKind(applied, ctrl.draft) != settingsChangeNone {
				ctrl.openConfirm()
			} else {
				return settingsExitUnfocus
			}
		}
	case "q", "ctrl+c":
		if ctrl.state == settingsUIList || ctrl.state == settingsUIConfirm {
			return settingsExitUnfocus
		}
		return settingsExitQuitApp
	case "j", "down":
		switch ctrl.state {
		case settingsUIPick:
			ctrl.movePick(1, languages)
		case settingsUIConfirm:
			ctrl.confirmIndex = 1
		default:
			ctrl.moveField(1)
		}
	case "k", "up":
		switch ctrl.state {
		case settingsUIPick:
			ctrl.movePick(-1, languages)
		case settingsUIConfirm:
			ctrl.confirmIndex = 0
		default:
			ctrl.moveField(-1)
		}
	case "enter":
		switch ctrl.state {
		case settingsUIPick:
			ctrl.confirmPick(languages)
		case settingsUIConfirm:
			if ctrl.confirmIndex == 0 {
				return settingsExitApply
			}
			return settingsExitUnfocus
		default:
			ctrl.openPicker(languages)
		}
	case "y":
		if ctrl.state == settingsUIConfirm {
			return settingsExitApply
		}
	case "n":
		if ctrl.state == settingsUIConfirm {
			return settingsExitUnfocus
		}
	default:
		if ctrl.state == settingsUIList {
			if field, ok := settingsFieldForKey(msg); ok {
				ctrl.field = field
				ctrl.openPicker(languages)
			}
		}
	}
	return settingsExitNone
}
