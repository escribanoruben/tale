package tui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingsExit int

const (
	settingsExitNone settingsExit = iota
	settingsExitUnfocus
	settingsExitApply
	settingsExitQuitApp
	settingsExitConfirmDiscard
)

const (
	settingsActionApply = iota
	settingsActionDiscard
	settingsActionCount
)

type settingsController struct {
	draft         AppSettings
	focusField    int
	focusOption   int
	focusOnAction bool
	focusAction   int
}

func newSettingsController(initial AppSettings) settingsController {
	initial.Phrases = snapPhraseCount(initial.Phrases)
	return settingsController{draft: initial}
}

func (c *settingsController) syncFocusFromDraft(languages []string) {
	c.focusOnAction = false
	c.focusField = settingFieldLanguage
	c.focusOption = currentPickIndex(c.focusField, c.draft, languages)
}

func (c *settingsController) jumpToSection(field int, languages []string) {
	if field < 0 || field >= settingFieldCount {
		return
	}
	c.focusOnAction = false
	c.focusField = field
	c.focusOption = currentPickIndex(field, c.draft, languages)
}

func (c *settingsController) hasPendingChanges(applied AppSettings) bool {
	return settingsChangeKind(applied, c.draft) != settingsChangeNone
}

func (c *settingsController) totalOptionRows(languages []string) int {
	n := 0
	for field := 0; field < settingFieldCount; field++ {
		n += fieldOptionCount(field, languages)
	}
	return n
}

func (c *settingsController) totalFocusRows(languages []string, applied AppSettings) int {
	n := c.totalOptionRows(languages)
	if c.hasPendingChanges(applied) {
		n += settingsActionCount
	}
	return n
}

func (c *settingsController) flatFocusIndex(languages []string) int {
	idx := 0
	for field := 0; field < settingFieldCount; field++ {
		count := fieldOptionCount(field, languages)
		if field == c.focusField {
			return idx + c.focusOption
		}
		idx += count
	}
	return 0
}

func (c *settingsController) setFlatFocusIndex(i int, languages []string) {
	remaining := i
	for field := 0; field < settingFieldCount; field++ {
		count := fieldOptionCount(field, languages)
		if remaining < count {
			c.focusField = field
			c.focusOption = remaining
			return
		}
		remaining -= count
	}
}

func (c *settingsController) currentFocusIndex(languages []string, applied AppSettings) int {
	if c.focusOnAction {
		return c.totalOptionRows(languages) + c.focusAction
	}
	return c.flatFocusIndex(languages)
}

func (c *settingsController) moveFocus(delta int, languages []string, applied AppSettings) {
	total := c.totalFocusRows(languages, applied)
	if total == 0 {
		return
	}
	next := (c.currentFocusIndex(languages, applied) + delta + total) % total
	optRows := c.totalOptionRows(languages)
	if next < optRows {
		c.focusOnAction = false
		c.setFlatFocusIndex(next, languages)
		return
	}
	c.focusOnAction = true
	c.focusAction = next - optRows
}

func (c *settingsController) selectFocused(languages []string) {
	applyPickToDraft(&c.draft, c.focusField, c.focusOption, languages)
}

func (c *settingsController) selectOption(field, option int, languages []string) {
	c.focusOnAction = false
	c.focusField = field
	c.focusOption = option
	applyPickToDraft(&c.draft, field, option, languages)
}

func (c *settingsController) focusActionRow(action int, languages []string, applied AppSettings) {
	if !c.hasPendingChanges(applied) {
		return
	}
	if action < 0 || action >= settingsActionCount {
		return
	}
	c.focusOnAction = true
	c.focusAction = action
}

func fieldOptionCount(field int, languages []string) int {
	switch field {
	case settingFieldLanguage:
		return len(languages)
	case settingFieldPhrases:
		return len(commonPhraseCounts)
	case settingFieldNoAdjacent:
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
	}
}

func buildFieldOptions(text uiText, field int, draft AppSettings, languages []string) []string {
	switch field {
	case settingFieldLanguage:
		opts := make([]string, len(languages))
		for i, lang := range languages {
			opts[i] = languageOptionLabel(lang)
		}
		return opts
	case settingFieldPhrases:
		_, blockSize := profileMetrics(draft.Language)
		preventAdjacent := !draft.AllowAdjacentRepeats
		opts := make([]string, len(commonPhraseCounts))
		for i, count := range commonPhraseCounts {
			passLen := phrasePasswordLength(count, blockSize)
			bits := phraseEntropyBits(draft.Language, count, preventAdjacent)
			summary := fmt.Sprintf(text.phraseOptionSummary, count, passLen, math.Round(bits))
			opts[i] = summary + " — " + phraseUseCase(text, count)
		}
		return opts
	case settingFieldNoAdjacent:
		return []string{text.settingsAdjacentPrevent, text.settingsAdjacentAllow}
	default:
		return nil
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

func selectionBox(selected bool) string {
	if selected {
		return "[x]"
	}
	return "[ ]"
}

func renderSettingsSummary(s AppSettings) string {
	text := uiForLang(s.Language)
	parts := []string{
		fmt.Sprintf("%s %s", text.settingsLanguage, languageOptionLabel(s.Language)),
		fmt.Sprintf("%s %s", text.settingsAllowAdjacent, boolWord(text, !s.AllowAdjacentRepeats)),
		fmt.Sprintf("%s %s", text.settingsPhrases, settingsFieldValue(text, settingFieldPhrases, s)),
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

// renderSettingsView draws the full-screen settings editor.
func renderSettingsView(s AppSettings, languages []string, ctrl settingsController, termWidth int) (string, []hitZone) {
	text := uiForLang(s.Language)
	innerW := viewContentWidth(termWidth)
	var lines []string
	var contentHits []hitZone
	lines = append(lines, appTitleStyle.Render(text.settingsTitle))
	lines = append(lines, "")
	lines = append(lines, renderWrapped(labelStyle, innerW, text.settingsSubtitle))
	lines = append(lines, "")

	groupLines, groupHits := settingsOptionGroups(text, ctrl, languages, len(lines), innerW)
	lines = append(lines, groupLines...)
	contentHits = append(contentHits, groupHits...)

	if kind := settingsChangeKind(s, ctrl.draft); kind != settingsChangeNone {
		lines = append(lines, "")
		if kind == settingsChangePassword {
			lines = append(lines, renderWrapped(warnStyle, innerW, text.settingsRegenPassword))
		} else {
			lines = append(lines, renderSettingsPendingHint(ctrl.draft.Language, kind))
		}
		actionLines, actionHits := settingsActionLines(text, ctrl, s, len(lines), innerW)
		lines = append(lines, actionLines...)
		contentHits = append(contentHits, actionHits...)
	}

	lines = append(lines, "", renderWrapped(helpStyle, innerW, settingsListHelp(text, s, ctrl.draft)))
	return strings.Join(lines, "\n"), contentHits
}

func renderSettingsDiscardConfirm(text uiText, innerW int) string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, renderWrapped(warnStyle, innerW, text.settingsDiscardPrompt))
	lines = append(lines, renderWrapped(helpStyle, innerW, text.settingsDiscardHelp))
	return strings.Join(lines, "\n")
}

func settingsListHelp(text uiText, applied, draft AppSettings) string {
	if settingsChangeKind(applied, draft) != settingsChangeNone {
		return text.settingsListHelpPending
	}
	return text.settingsListHelp
}

// renderSettingsWidget draws the settings panel and returns hit zones relative to the
// rendered panel (row 0 = top border). Used by the first-run setup screen.
func renderSettingsWidget(s AppSettings, languages []string, ctrl settingsController, focused bool) (string, []hitZone) {
	if !focused {
		return renderSettingsSummaryBar(s)
	}
	body, hits := renderSettingsView(s, languages, ctrl, 0)
	return panelStyle.BorderForeground(lipgloss.Color("205")).Render(body), offsetHitZonesColumn(hits, 1, 1, 0)
}

func settingsOptionGroups(text uiText, ctrl settingsController, languages []string, lineBase, innerW int) ([]string, []hitZone) {
	var lines []string
	var hits []hitZone
	y := lineBase

	for field := 0; field < settingFieldCount; field++ {
		lines = append(lines, renderWrapped(labelStyle, innerW, settingsFieldLabel(text, field)))
		y++

		options := buildFieldOptions(text, field, ctrl.draft, languages)
		selected := currentPickIndex(field, ctrl.draft, languages)
		for i, opt := range options {
			focused := !ctrl.focusOnAction && field == ctrl.focusField && i == ctrl.focusOption
			lines = append(lines, settingsOptionLine(i == selected, focused, innerW, opt))
			hits = append(hits, hitZone{
				y0: y, y1: y, x0: 0, x1: -1,
				kind: hitSettingsPick, index: i, extra: field,
			})
			y++
		}
		lines = append(lines, "")
		y++
	}
	return lines, hits
}

func settingsActionLines(text uiText, ctrl settingsController, applied AppSettings, lineBase, innerW int) ([]string, []hitZone) {
	if !ctrl.hasPendingChanges(applied) {
		return nil, nil
	}
	var lines []string
	var hits []hitZone
	y := lineBase
	lines = append(lines, "")
	y++

	actions := []string{text.settingsConfirmApply, text.settingsConfirmDiscard}
	for i, label := range actions {
		focused := ctrl.focusOnAction && ctrl.focusAction == i
		lines = append(lines, settingsActionLine(focused, innerW, label))
		hits = append(hits, fullRow(y, y, hitSettingsConfirm, i))
		y++
	}
	return lines, hits
}

func settingsOptionLine(selected, focused bool, innerW int, label string) string {
	prefix := "  "
	if focused {
		prefix = "> "
	}
	line := prefix + selectionBox(selected) + " " + label
	if focused {
		return renderWrapped(activeStyle, innerW, line)
	}
	return renderWrapped(labelStyle, innerW, line)
}

func settingsActionLine(focused bool, innerW int, label string) string {
	prefix := "  "
	if focused {
		prefix = "> "
	}
	line := prefix + "[ ] " + label
	if focused {
		return renderWrapped(activeStyle, innerW, line)
	}
	return renderWrapped(labelStyle, innerW, line)
}

func handleSettingsKey(ctrl *settingsController, msg tea.KeyMsg, languages []string, applied AppSettings) settingsExit {
	switch msg.String() {
	case "esc":
		if ctrl.hasPendingChanges(applied) {
			return settingsExitConfirmDiscard
		}
		return settingsExitUnfocus
	case "j", "down":
		ctrl.moveFocus(1, languages, applied)
	case "k", "up":
		ctrl.moveFocus(-1, languages, applied)
	case " ", "enter":
		if ctrl.focusOnAction && ctrl.hasPendingChanges(applied) {
			if ctrl.focusAction == settingsActionApply {
				return settingsExitApply
			}
			return settingsExitUnfocus
		}
		ctrl.selectFocused(languages)
	default:
		if field, ok := settingsFieldForKey(msg); ok {
			ctrl.jumpToSection(field, languages)
		}
	}
	return settingsExitNone
}
