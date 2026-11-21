package profile

import (
	"fmt"
	"strings"
	"unicode"
)

// Validate checks profile fields required for generation.
func (p Profile) Validate() error {
	if strings.TrimSpace(p.Language) == "" {
		return fmt.Errorf("language must not be empty")
	}
	if p.BlockSize <= 0 {
		return fmt.Errorf("block_size must be positive")
	}
	if p.PasswordLength <= 0 {
		return fmt.Errorf("password_length must be positive")
	}
	if p.PasswordLength%p.BlockSize != 0 {
		return fmt.Errorf("password_length %d is not divisible by block_size %d", p.PasswordLength, p.BlockSize)
	}
	if len(p.PhraseTemplate) != p.BlockSize {
		return fmt.Errorf("phrase_template has %d slots, block_size is %d", len(p.PhraseTemplate), p.BlockSize)
	}
	if strings.TrimSpace(p.Alphabet) == "" {
		return fmt.Errorf("alphabet must not be empty")
	}
	seen := make(map[rune]struct{})
	for _, r := range p.Alphabet {
		if !unicode.IsLower(r) || !unicode.IsLetter(r) {
			return fmt.Errorf("alphabet must contain only lowercase letters, got %q", string(r))
		}
		if _, ok := seen[r]; ok {
			return fmt.Errorf("alphabet contains duplicate letter %q", string(r))
		}
		seen[r] = struct{}{}
	}
	if p.DataLayout != "" && p.DataLayout != "letter" {
		return fmt.Errorf("unsupported data_layout %q", p.DataLayout)
	}
	switch p.Strategy {
	case "", StrategyRandom, StrategyStory, StrategyAliasStory, StrategyNounField, StrategyNounPath:
	default:
		return fmt.Errorf("unsupported strategy %q (want random, story, noun_field or noun_path)", p.Strategy)
	}
	return nil
}
