## ADDED Requirements

### Requirement: Embedded word lists

The `tale` binary SHALL embed curated English and Spanish word lists at build time using `go:embed`.

#### Scenario: Words available without network

- **WHEN** the user runs `tale` on a machine with no network connectivity
- **THEN** word lists for `en` and `es` are available immediately

#### Scenario: Embedded layout

- **WHEN** the binary is built
- **THEN** it includes `internal/worddata/{en,es}/{letter}/{nouns,adjectives,verbs}.tsv`

### Requirement: Embedded language configurations

The `tale` binary SHALL embed `en` and `es` TOML language configurations at build time.

#### Scenario: Language configs ship with binary

- **WHEN** the binary is built
- **THEN** it includes `internal/profile/{en,es}.toml` with alphabet and generation settings

### Requirement: No first-run download

The application SHALL NOT require downloading word lists on first run for built-in languages.

#### Scenario: No setup required for words

- **WHEN** a user installs `tale` and runs `tale stats`
- **THEN** statistics are shown without any prior setup step

### Requirement: Stats command

The application SHALL provide `tale stats` to inspect embedded word-list counts by language, letter, and part of speech.

#### Scenario: Embedded stats

- **WHEN** the user runs `tale stats`
- **THEN** the CLI prints per-language totals and a letter × POS table from embedded data

#### Scenario: JSON stats

- **WHEN** the user runs `tale stats --json`
- **THEN** the CLI prints machine-readable statistics
