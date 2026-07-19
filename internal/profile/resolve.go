package profile

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ruben/tales_vault_system/internal/paths"
	"github.com/ruben/tales_vault_system/internal/worddata"
)

var languageRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,7}$`)

// Resolved ties a profile to its storage location and data source.
type Resolved struct {
	Profile  Profile
	Builtin  bool
	FilePath string
}

// Resolve loads the profile for a language from ~/.tale/profiles or embedded defaults.
// User profiles override built-in defaults for the same language.
func Resolve(language string) (Resolved, error) {
	if err := ValidateLanguage(language); err != nil {
		return Resolved{}, err
	}

	userPath, err := paths.ProfilePath(language)
	if err != nil {
		return Resolved{}, err
	}
	if _, err := os.Stat(userPath); err == nil {
		p, err := loadFile(userPath)
		if err != nil {
			return Resolved{}, err
		}
		if p.Language == "" {
			p.Language = language
		}
		if p.ID == "" {
			p.ID = language
		}
		if p.Language != language {
			return Resolved{}, fmt.Errorf("profile %q defines language %q", language, p.Language)
		}
		if err := p.Validate(); err != nil {
			return Resolved{}, fmt.Errorf("profile %q: %w", language, err)
		}
		return Resolved{Profile: p, Builtin: false, FilePath: userPath}, nil
	} else if !os.IsNotExist(err) {
		return Resolved{}, err
	}

	p, err := Load(Files, language)
	if err != nil {
		return Resolved{}, err
	}
	if p.Language == "" {
		p.Language = language
	}
	if p.ID == "" {
		p.ID = language
	}
	if err := p.Validate(); err != nil {
		return Resolved{}, fmt.Errorf("profile %q: %w", language, err)
	}
	return Resolved{Profile: p, Builtin: true}, nil
}

// Entry describes a known profile for listing.
type Entry struct {
	Language    string
	DisplayName string
	Builtin     bool
	Customized  bool
	Path        string
}

// List returns one entry per language (built-in plus user overrides).
func List() ([]Entry, error) {
	byLang := make(map[string]Entry)

	embedded, err := fs.ReadDir(Files, ".")
	if err != nil {
		return nil, err
	}
	for _, e := range embedded {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		lang := strings.TrimSuffix(e.Name(), ".toml")
		p, err := Load(Files, lang)
		if err != nil {
			continue
		}
		if p.Language == "" {
			p.Language = lang
		}
		byLang[p.Language] = Entry{
			Language:    p.Language,
			DisplayName: p.DisplayName,
			Builtin:     true,
		}
	}

	profilesDir, err := paths.ProfilesDir()
	if err == nil {
		entries, err := os.ReadDir(profilesDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
					continue
				}
				lang := strings.TrimSuffix(e.Name(), ".toml")
				path := filepath.Join(profilesDir, e.Name())
				p, err := loadFile(path)
				if err != nil {
					continue
				}
				if p.Language == "" {
					p.Language = lang
				}
				prev := byLang[p.Language]
				byLang[p.Language] = Entry{
					Language:    p.Language,
					DisplayName: p.DisplayName,
					Builtin:     prev.Builtin,
					Customized:  true,
					Path:        path,
				}
			}
		}
	}

	out := make([]Entry, 0, len(byLang))
	for _, e := range byLang {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Language < out[j].Language
	})
	return out, nil
}

// InitOptions configures creation of a user profile for a language.
type InitOptions struct {
	Language     string
	FromLanguage string
	DisplayName  string
	ForceData    bool
	Force        bool
}

// NewLanguageOptions configures creation of a custom language from an alphabet.
type NewLanguageOptions struct {
	Language    string
	Alphabet    string
	DisplayName string
	Force       bool
	ForceData   bool
}

// InitResult reports what was created during profile init.
type InitResult struct {
	ProfilePath string
	DataPath    string
	DataCopied  bool
	DataScaffolded bool
}

// InitUserProfile copies a profile into ~/.tale/profiles/{language}.toml
// and optionally seeds ~/.tale/data/{language}.
func InitUserProfile(opts InitOptions) (InitResult, error) {
	if err := ValidateLanguage(opts.Language); err != nil {
		return InitResult{}, fmt.Errorf("language: %w", err)
	}

	fromLang := opts.FromLanguage
	if fromLang == "" {
		fromLang = opts.Language
	}
	if err := ValidateLanguage(fromLang); err != nil {
		return InitResult{}, fmt.Errorf("from language: %w", err)
	}

	newPath, err := paths.ProfilePath(opts.Language)
	if err != nil {
		return InitResult{}, err
	}
	if _, err := os.Stat(newPath); err == nil {
		if !opts.Force {
			return InitResult{}, fmt.Errorf("profile for language %q already exists at %s (use --force to overwrite)", opts.Language, newPath)
		}
	} else if !os.IsNotExist(err) {
		return InitResult{}, err
	}

	src, err := Resolve(fromLang)
	if err != nil {
		return InitResult{}, fmt.Errorf("from profile: %w", err)
	}

	p := src.Profile
	p.ID = opts.Language
	p.Language = opts.Language
	if opts.DisplayName != "" {
		p.DisplayName = opts.DisplayName
	} else if opts.Language != fromLang {
		p.DisplayName = fmt.Sprintf("%s (custom)", strings.TrimSuffix(p.DisplayName, " (default)"))
	}
	if err := p.Validate(); err != nil {
		return InitResult{}, err
	}

	dstLang := p.Language
	dataPath, err := paths.LanguageDataDir(dstLang)
	if err != nil {
		return InitResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return InitResult{}, err
	}
	if err := writeProfileFile(newPath, p); err != nil {
		return InitResult{}, err
	}

	copied := false
	needCopy := opts.ForceData
	if !needCopy {
		empty, err := dirEmpty(dataPath)
		if err != nil {
			_ = os.Remove(newPath)
			return InitResult{}, err
		}
		needCopy = empty
	}
	if needCopy {
		if err := seedLanguageData(src, dstLang, dataPath); err != nil {
			_ = os.Remove(newPath)
			return InitResult{}, err
		}
		copied = true
	}

	return InitResult{
		ProfilePath: newPath,
		DataPath:    dataPath,
		DataCopied:  copied,
	}, nil
}

// NewUserLanguage creates a custom language with the given alphabet and scaffolds
// empty word-list stubs for each alphabet letter (no clone from another language).
func NewUserLanguage(opts NewLanguageOptions) (InitResult, error) {
	if err := ValidateLanguage(opts.Language); err != nil {
		return InitResult{}, fmt.Errorf("language: %w", err)
	}
	alphabet := strings.TrimSpace(opts.Alphabet)
	if alphabet == "" {
		return InitResult{}, fmt.Errorf("alphabet is required (pass --alphabet with unique lowercase letters)")
	}

	newPath, err := paths.ProfilePath(opts.Language)
	if err != nil {
		return InitResult{}, err
	}
	if _, err := os.Stat(newPath); err == nil {
		if !opts.Force {
			return InitResult{}, fmt.Errorf("language %q already exists at %s (use --force to overwrite)", opts.Language, newPath)
		}
	} else if !os.IsNotExist(err) {
		return InitResult{}, err
	}

	display := opts.DisplayName
	if display == "" {
		display = opts.Language
	}
	const blockSize = 4
	p := Profile{
		ID:             opts.Language,
		Language:       opts.Language,
		DisplayName:    display,
		PasswordLength: 5 * blockSize,
		BlockSize:      blockSize,
		Alphabet:       alphabet,
		PhraseTemplate: []string{"noun", "adjective", "verb", "noun"},
		DataLayout:     "letter",
	}
	if err := p.Validate(); err != nil {
		return InitResult{}, err
	}

	dataPath, err := paths.LanguageDataDir(opts.Language)
	if err != nil {
		return InitResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return InitResult{}, err
	}
	if err := writeProfileFile(newPath, p); err != nil {
		return InitResult{}, err
	}

	scaffolded := false
	needScaffold := opts.ForceData
	if !needScaffold {
		empty, err := dirEmpty(dataPath)
		if err != nil {
			_ = os.Remove(newPath)
			return InitResult{}, err
		}
		needScaffold = empty
	}
	if needScaffold {
		if err := scaffoldLanguageData(alphabet, dataPath); err != nil {
			_ = os.Remove(newPath)
			return InitResult{}, err
		}
		scaffolded = true
	}

	return InitResult{
		ProfilePath:    newPath,
		DataPath:       dataPath,
		DataScaffolded: scaffolded,
	}, nil
}

// ValidateLanguage checks language codes are safe and well-formed.
func ValidateLanguage(language string) error {
	if !languageRe.MatchString(language) {
		return fmt.Errorf("invalid language %q (use lowercase letters, digits, and -)", language)
	}
	return nil
}

func loadFile(path string) (Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	if err := toml.Unmarshal(data, &p); err != nil {
		return Profile{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return p, nil
}

func writeProfileFile(path string, p Profile) error {
	if p.DataLayout == "" {
		p.DataLayout = "letter"
	}
	content := fmt.Sprintf(`id = %q
language = %q
display_name = %q
password_length = %d
block_size = %d
alphabet = %q
phrase_template = [%s]
data_layout = %q

[normalization]
strip_accents = true
lowercase = true
exclude_initial_enye = true
`, p.ID, p.Language, p.DisplayName, p.PasswordLength, p.BlockSize, p.Alphabet,
		quoteList(p.PhraseTemplate), p.DataLayout)
	return os.WriteFile(path, []byte(content), 0o644)
}

func quoteList(items []string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = fmt.Sprintf("%q", item)
	}
	return strings.Join(parts, ", ")
}

func dirEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func seedLanguageData(src Resolved, dstLang, dstPath string) error {
	srcLang := src.Profile.Language
	if src.Builtin {
		return copyEmbeddedLanguage(srcLang, dstLang, dstPath)
	}
	srcPath, err := paths.LanguageDataDir(srcLang)
	if err != nil {
		return err
	}
	return copyDirTree(srcPath, dstPath)
}

var stubPOSFiles = []string{"nouns.tsv", "adjectives.tsv", "verbs.tsv"}

func scaffoldLanguageData(alphabet, dstPath string) error {
	if err := os.MkdirAll(dstPath, 0o755); err != nil {
		return err
	}
	for _, r := range alphabet {
		letterDir := filepath.Join(dstPath, string(r))
		if err := os.MkdirAll(letterDir, 0o755); err != nil {
			return err
		}
		for _, name := range stubPOSFiles {
			path := filepath.Join(letterDir, name)
			if err := os.WriteFile(path, nil, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyEmbeddedLanguage(srcLang, dstLang, dstPath string) error {
	if err := os.MkdirAll(dstPath, 0o755); err != nil {
		return err
	}
	langFS, err := fs.Sub(worddata.Words, srcLang)
	if err != nil {
		return fmt.Errorf("embedded language %q: %w", srcLang, err)
	}
	return fs.WalkDir(langFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dstPath, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(langFS, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func copyDirTree(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source %s is not a directory", src)
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
