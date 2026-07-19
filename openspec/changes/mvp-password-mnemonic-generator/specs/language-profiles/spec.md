## ADDED Requirements

### Requirement: TOML language configuration loading

The system SHALL load generation settings from TOML files keyed by language code, shipped with the binary or overridden by the user.

#### Scenario: Language file loaded

- **WHEN** the user specifies `--language es`
- **THEN** the system loads the Spanish language configuration and applies its settings

### Requirement: Builtin languages

The system SHALL ship two built-in language configurations: `en` and `es`.

#### Scenario: English builtin exists

- **WHEN** the `en` language is loaded without a user override
- **THEN** it defines a 20-character password, block size 4, language `en`, and a 20-letter alphabet

#### Scenario: Spanish builtin exists

- **WHEN** the `es` language is loaded without a user override
- **THEN** it defines a 20-character password, block size 4, language `es`, and a 20-letter alphabet that does not include `ñ`

#### Scenario: Spanish alphabet excludes enye

- **WHEN** the `es` language is loaded
- **THEN** the `alphabet` field does not contain `ñ`

### Requirement: Twenty-letter alphabet for builtins

Each built-in language configuration SHALL define an alphabet of exactly 20 letters representing the most frequent word-initial letters for that language.

#### Scenario: Alphabet size

- **WHEN** either built-in language is loaded
- **THEN** the `alphabet` field contains exactly 20 unique lowercase letters

### Requirement: Letter-based data layout

Word lists SHALL be resolved from `{data_root}/{language}/{letter}/{nouns,adjectives,verbs}.tsv`.

#### Scenario: Per-letter file lookup

- **WHEN** the system needs a Spanish noun starting with `a`
- **THEN** it reads from embedded `es/a/nouns.tsv`

### Requirement: Language configuration fields

Each language configuration SHALL define at minimum: `id`, `language`, `password_length`, `block_size`, `alphabet`, `phrase_template`, `data_layout`, and `normalization` settings.

#### Scenario: Required fields present

- **WHEN** a language configuration is validated at startup
- **THEN** missing required fields cause a clear error before generation runs

### Requirement: Default language selection

The system SHALL select a default language from user settings or system locale when available among built-ins, otherwise `en`, when no `--language` flag is provided.

#### Scenario: Implicit default

- **WHEN** the user runs `tale generate` without `--language` and without a saved language preference on a non-Spanish locale
- **THEN** the `en` language configuration is used
