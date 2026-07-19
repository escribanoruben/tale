package tui

import "github.com/charmbracelet/lipgloss"

type blockPalette struct {
	active   string // phrase / password letter (foreground on dark bg)
	inactive string
	chipBG   string // selection pill & cursor background
	chipFG   string // text on chipBG
}

// Six hues spaced around the wheel (no green/cyan cluster) for dark terminals.
// chipBG/chipFG are tuned per hue so pills stay readable (e.g. dark text on yellow).
var blockPalettes = []blockPalette{
	{active: "208", inactive: "172", chipBG: "166", chipFG: "233"}, // orange
	{active: "218", inactive: "169", chipBG: "168", chipFG: "233"}, // pink
	{active: "220", inactive: "179", chipBG: "136", chipFG: "233"}, // yellow
	{active: "203", inactive: "167", chipBG: "167", chipFG: "233"}, // coral
	{active: "75", inactive: "32", chipBG: "25", chipFG: "255"},    // blue
	{active: "141", inactive: "97", chipBG: "55", chipFG: "255"},    // purple
}

func blockPaletteFor(index int) blockPalette {
	return blockPalettes[index%len(blockPalettes)]
}

func blockTextStyle(p blockPalette, phraseActive bool) lipgloss.Style {
	if phraseActive {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.active))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(p.inactive))
}

// blockCursorStyle highlights the active letter in the password (no padding — keeps columns aligned).
func blockCursorStyle(p blockPalette) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(p.chipFG)).
		Background(lipgloss.Color(p.chipBG))
}

// blockSlotStyle highlights the active word as a pill tinted with the block hue.
func blockSlotStyle(p blockPalette) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(p.chipFG)).
		Background(lipgloss.Color(p.chipBG)).
		Padding(0, 1)
}

func renderBlockWord(word string, blockIdx int) string {
	return blockSlotStyle(blockPaletteFor(blockIdx)).Render(word)
}
