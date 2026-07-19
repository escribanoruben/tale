package tui

import "testing"

func TestBlockPaletteCycles(t *testing.T) {
	n := len(blockPalettes)
	if blockPaletteFor(0).active != blockPaletteFor(n).active {
		t.Fatal("expected palette to cycle after one full round")
	}
	if blockPaletteFor(1).active == blockPaletteFor(0).active {
		t.Fatal("expected distinct palettes for adjacent blocks")
	}
}
