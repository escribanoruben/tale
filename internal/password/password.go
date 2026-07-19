package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Options controls password generation behavior.
type Options struct {
	// PreventAdjacentRepeats avoids consecutive identical characters (default on).
	PreventAdjacentRepeats bool
}

// DefaultOptions returns the standard generation settings.
func DefaultOptions() Options {
	return Options{PreventAdjacentRepeats: true}
}

// Generate creates a random lowercase password using letters from alphabet.
func Generate(alphabet string, length int, opts Options) (string, error) {
	return GenerateAfter(alphabet, length, 0, opts)
}

// GenerateAfter creates a password segment, optionally continuing after prev.
// prev == 0 means no leading constraint.
func GenerateAfter(alphabet string, length int, prev rune, opts Options) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	if alphabet == "" {
		return "", fmt.Errorf("alphabet must not be empty")
	}
	letters := []rune(alphabet)
	if opts.PreventAdjacentRepeats && length > 1 && len(letters) < 2 {
		return "", fmt.Errorf("alphabet needs at least 2 letters when preventing adjacent repeats")
	}
	if opts.PreventAdjacentRepeats && length > 0 && prev != 0 && len(letters) < 2 {
		return "", fmt.Errorf("alphabet needs at least 2 letters when preventing adjacent repeats")
	}

	max := big.NewInt(int64(len(letters)))
	out := make([]rune, length)
	for i := range out {
		forbidden := rune(0)
		if opts.PreventAdjacentRepeats {
			switch {
			case i > 0:
				forbidden = out[i-1]
			case prev != 0:
				forbidden = prev
			}
		}

		picked, err := pickLetter(letters, max, forbidden)
		if err != nil {
			return "", err
		}
		out[i] = picked
	}
	return string(out), nil
}

func pickLetter(letters []rune, max *big.Int, forbidden rune) (rune, error) {
	if forbidden == 0 {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return 0, err
		}
		return letters[n.Int64()], nil
	}

	// Rejection sampling; alphabet is small so this is cheap.
	for attempt := 0; attempt < 1000; attempt++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return 0, err
		}
		letter := letters[n.Int64()]
		if letter != forbidden {
			return letter, nil
		}
	}
	return 0, fmt.Errorf("could not pick a non-repeating letter from alphabet")
}

// Blocks splits password into fixed-size blocks.
func Blocks(password string, blockSize int) ([]string, error) {
	if blockSize <= 0 {
		return nil, fmt.Errorf("block size must be positive")
	}
	if len(password)%blockSize != 0 {
		return nil, fmt.Errorf("password length %d is not divisible by block size %d", len(password), blockSize)
	}
	blocks := make([]string, 0, len(password)/blockSize)
	for i := 0; i < len(password); i += blockSize {
		blocks = append(blocks, password[i:i+blockSize])
	}
	return blocks, nil
}

// LettersAt returns the letter required for each slot in a block.
func LettersAt(block string, slot int) (byte, error) {
	if slot < 0 || slot >= len(block) {
		return 0, fmt.Errorf("slot %d out of range for block %q", slot, block)
	}
	return block[slot], nil
}

// HasAdjacentRepeats reports whether s contains identical consecutive runes.
func HasAdjacentRepeats(s string) bool {
	runes := []rune(s)
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			return true
		}
	}
	return false
}
