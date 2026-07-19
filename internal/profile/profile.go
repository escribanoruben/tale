package profile

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/BurntSushi/toml"
)

// Profile holds generation settings for a language.
type Profile struct {
	ID             string   `toml:"id"`
	Language       string   `toml:"language"`
	DisplayName    string   `toml:"display_name"`
	PasswordLength int      `toml:"password_length"`
	BlockSize      int      `toml:"block_size"`
	Alphabet       string   `toml:"alphabet"`
	PhraseTemplate []string `toml:"phrase_template"`
	DataLayout     string   `toml:"data_layout"`
}

// Load reads a profile by id from the embedded filesystem.
func Load(files embed.FS, id string) (Profile, error) {
	path := id + ".toml"
	data, err := fs.ReadFile(files, path)
	if err != nil {
		return Profile{}, fmt.Errorf("profile %q: %w", id, err)
	}
	var p Profile
	if err := toml.Unmarshal(data, &p); err != nil {
		return Profile{}, fmt.Errorf("profile %q: %w", id, err)
	}
	if p.ID == "" {
		p.ID = id
	}
	return p, nil
}
