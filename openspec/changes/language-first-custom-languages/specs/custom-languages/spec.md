## ADDED Requirements

### Requirement: Create custom language with alphabet

The system SHALL provide a command to create a new user language that requires an alphabet of unique lowercase Latin letters (any positive length meeting validation rules).

#### Scenario: New language with alphabet flag

- **WHEN** the user runs `tale language new de --alphabet=abcdefghijklmnopqr` (or equivalent)
- **THEN** the system writes `~/.tale/profiles/de.toml` with that alphabet and language code `de`

#### Scenario: Alphabet required

- **WHEN** the user runs `tale language new de` without providing an alphabet
- **THEN** the system fails with a clear error asking for `--alphabet` (or prompts once in interactive mode)

### Requirement: Scaffold letter data for custom alphabet

After creating a custom language, the system SHALL create `~/.tale/data/{language}/{letter}/` directories for each alphabet letter with empty or stub `nouns.tsv`, `adjectives.tsv`, and `verbs.tsv` files, and SHALL NOT copy another language’s word lists unless the user explicitly uses clone/`init --from`.

#### Scenario: Scaffold matches alphabet

- **WHEN** a new language is created with alphabet `ab`
- **THEN** the data directory contains `a/` and `b/` letter folders with the three POS TSV stubs and does not contain folders for letters outside that alphabet

### Requirement: Clone remains available

The system SHALL retain a clone/init path that copies configuration and seeds word data from an existing language for near-forks.

#### Scenario: Init from English

- **WHEN** the user runs `tale language init pt --from en`
- **THEN** a user profile for `pt` is created and word data is seeded from the English source unless existing data is preserved per force flags

### Requirement: Uncovered letters fail clearly

Until the user fills word lists, generation or TUI for a custom language SHALL fail with a clear error when a password letter has no candidates, without changing the password letter silently.

#### Scenario: Empty stub blocks generation

- **WHEN** the user generates a mnemonic for a new custom language whose TSV stubs are still empty
- **THEN** the command errors describing the missing words for the required letter and POS
