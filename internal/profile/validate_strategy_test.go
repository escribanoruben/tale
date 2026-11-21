package profile

import "testing"

func TestValidateStrategy(t *testing.T) {
	for _, s := range []string{"", StrategyChainedStory, StrategyNounField, StrategyNounPath} {
		p := Profile{Language: "es", PasswordLength: 16, BlockSize: 4,
			Alphabet: "abc", PhraseTemplate: []string{"noun", "adj", "verb", "noun"},
			Strategy: s}
		if err := p.Validate(); err != nil {
			t.Fatalf("strategy %q should validate: %v", s, err)
		}
	}
	p := Profile{Language: "es", PasswordLength: 16, BlockSize: 4,
		Alphabet: "abc", PhraseTemplate: []string{"noun", "adj", "verb", "noun"},
		Strategy: "haiku"}
	if err := p.Validate(); err == nil {
		t.Fatal("unknown strategy should fail validation")
	}
}
