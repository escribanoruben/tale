package tui

import (
	"fmt"
	"math"

	"github.com/escribanoruben/tales_vault_system/internal/password"
	"github.com/escribanoruben/tales_vault_system/internal/profile"
)

const (
	settingFieldLanguage = iota
	settingFieldNoAdjacent
	settingFieldPhrases
	settingFieldCount
)

const (
	settingsChangeNone = iota
	settingsChangePassword
)

var commonPhraseCounts = []int{3, 4, 5, 6, 8}

func languageIndex(languages []string, lang string) int {
	for i, l := range languages {
		if l == lang {
			return i
		}
	}
	return 0
}

func languageOptionLabel(lang string) string {
	switch lang {
	case "en":
		return "English"
	case "es":
		return "Español"
	default:
		return lang
	}
}

func snapPhraseCount(phrases int) int {
	if phrases <= 0 {
		return 5
	}
	best := commonPhraseCounts[0]
	for _, count := range commonPhraseCounts {
		if abs(count-phrases) < abs(best-phrases) {
			best = count
		}
	}
	return best
}

func phrasePresetIndex(phrases int) int {
	target := snapPhraseCount(phrases)
	for i, count := range commonPhraseCounts {
		if count == target {
			return i
		}
	}
	return 2
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func profileMetrics(lang string) (alphabetLen, blockSize int) {
	alphabetLen = 20
	blockSize = 4
	resolved, err := profile.Resolve(lang)
	if err != nil {
		return alphabetLen, blockSize
	}
	if len(resolved.Profile.Alphabet) > 0 {
		alphabetLen = len(resolved.Profile.Alphabet)
	}
	if resolved.Profile.BlockSize > 0 {
		blockSize = resolved.Profile.BlockSize
	}
	return alphabetLen, blockSize
}

func phrasePasswordLength(phrases, blockSize int) int {
	return phrases * blockSize
}

func phraseEntropyBits(lang string, phrases int, preventAdjacent bool) float64 {
	alphabetLen, blockSize := profileMetrics(lang)
	passLen := phrasePasswordLength(phrases, blockSize)
	return password.EntropyBits(alphabetLen, passLen, password.Options{
		PreventAdjacentRepeats: preventAdjacent,
	})
}

func phraseUseCase(text uiText, phrases int) string {
	switch phrases {
	case 3:
		return text.phraseUseCase3
	case 4:
		return text.phraseUseCase4
	case 5:
		return text.phraseUseCase5
	case 6:
		return text.phraseUseCase6
	case 8:
		return text.phraseUseCase8
	default:
		return text.phraseUseCase5
	}
}

func settingsChangeKind(applied, draft AppSettings) int {
	if applied.Language != draft.Language ||
		applied.Phrases != draft.Phrases ||
		applied.AllowAdjacentRepeats != draft.AllowAdjacentRepeats {
		return settingsChangePassword
	}
	return settingsChangeNone
}

func renderSettingsPendingWarning(lang string, kind int) string {
	text := uiForLang(lang)
	switch kind {
	case settingsChangePassword:
		return warnStyle.Render(text.settingsRegenPassword)
	default:
		return ""
	}
}

func renderSettingsPendingHint(lang string, kind int) string {
	text := uiForLang(lang)
	switch kind {
	case settingsChangePassword:
		return helpStyle.Render(text.settingsPendingHint)
	default:
		return ""
	}
}

func compactPhraseSummary(s AppSettings) string {
	text := uiForLang(s.Language)
	preventAdjacent := !s.AllowAdjacentRepeats
	phrases := snapPhraseCount(s.Phrases)
	bits := phraseEntropyBits(s.Language, phrases, preventAdjacent)
	return fmt.Sprintf("%s: %d (~%.0f %s)", text.settingsPhrases, phrases, bits, text.entropyBitsUnit)
}

func phraseOptionSummaryLine(text uiText, lang string, count int, preventAdjacent bool) string {
	_, blockSize := profileMetrics(lang)
	passLen := phrasePasswordLength(count, blockSize)
	bits := phraseEntropyBits(lang, count, preventAdjacent)
	return fmt.Sprintf(text.phraseOptionSummary, count, passLen, math.Round(bits))
}

func joinUIParts(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}
