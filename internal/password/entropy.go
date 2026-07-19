package password

import (
	"fmt"
	"math"
)

// EntropyBits estimates password entropy in bits for the generation model.
func EntropyBits(alphabetLen, passwordLen int, opts Options) float64 {
	if alphabetLen <= 1 || passwordLen <= 0 {
		return 0
	}
	n := float64(alphabetLen)
	if opts.PreventAdjacentRepeats && passwordLen > 1 {
		if alphabetLen < 2 {
			return 0
		}
		return math.Log2(n) + float64(passwordLen-1)*math.Log2(n-1)
	}
	return float64(passwordLen) * math.Log2(n)
}

// FormatEntropy renders a human-readable entropy summary.
func FormatEntropy(alphabetLen, passwordLen int, opts Options) string {
	bits := EntropyBits(alphabetLen, passwordLen, opts)
	if opts.PreventAdjacentRepeats && passwordLen > 1 && alphabetLen >= 2 {
		return fmt.Sprintf("~%.1f bits (no adjacent repeats)", bits)
	}
	return fmt.Sprintf("~%.1f bits (%d chars × log₂(%d))", bits, passwordLen, alphabetLen)
}
