package paths

import (
	"os"
	"path/filepath"
)

const configDirName = ".tale"

// ConfigDir returns ~/.tale (created on demand by callers if needed).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName), nil
}

// ProfilesDir returns ~/.tale/profiles.
func ProfilesDir() (string, error) {
	root, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profiles"), nil
}

// DataDir returns ~/.tale/data.
func DataDir() (string, error) {
	root, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "data"), nil
}

// ProfilePath returns ~/.tale/profiles/{id}.toml.
func ProfilePath(id string) (string, error) {
	dir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".toml"), nil
}

// LanguageDataDir returns ~/.tale/data/{language}.
func LanguageDataDir(language string) (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, language), nil
}
