## ADDED Requirements

### Requirement: Language-keyed configuration

The system SHALL resolve generation settings by language code from embedded defaults or a user TOML file at `~/.tale/profiles/{language}.toml`.

#### Scenario: Builtin English

- **WHEN** the user requests language `en` and no user profile file exists for `en`
- **THEN** the system loads the embedded English configuration (default length 20, block size 4, curated 20-letter alphabet)

#### Scenario: Builtin Spanish

- **WHEN** the user requests language `es` and no user profile file exists for `es`
- **THEN** the system loads the embedded Spanish configuration whose alphabet does not include `ñ`

#### Scenario: User override replaces builtin

- **WHEN** `~/.tale/profiles/es.toml` exists
- **THEN** the system uses that file for language `es` and does not use the embedded Spanish configuration

### Requirement: Builtin languages

The system SHALL ship embedded configurations for languages `en` and `es` only.

#### Scenario: Unknown language without user file

- **WHEN** the user requests a language with no embedded and no user configuration
- **THEN** the system returns a clear error and does not generate

### Requirement: Alphabet validation

The system SHALL require a non-empty alphabet of unique lowercase Latin letters with no fixed maximum length. Builtin `en` and `es` alphabets SHALL each contain exactly 20 letters.

#### Scenario: Custom alphabet length

- **WHEN** a user language configuration defines an alphabet of 16 unique lowercase letters
- **THEN** validation succeeds

#### Scenario: Duplicate letter rejected

- **WHEN** an alphabet contains a repeated letter
- **THEN** validation fails with a clear error

### Requirement: Letter-based data layout

Word lists SHALL be resolved from `{data_root}/{language}/{letter}/{nouns,adjectives,verbs}.tsv`.

#### Scenario: Per-letter file lookup

- **WHEN** the system needs a Spanish noun starting with `a` from embedded data
- **THEN** it reads from embedded `es/a/nouns.tsv`
