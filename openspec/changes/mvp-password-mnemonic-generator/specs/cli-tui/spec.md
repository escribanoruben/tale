## ADDED Requirements

### Requirement: Binary name

The CLI binary SHALL be named `tale`.

#### Scenario: Invoking the application

- **WHEN** the user runs `tale --help`
- **THEN** the Tales Vault System help text is displayed

### Requirement: Generate command

The system SHALL provide a `generate` subcommand that performs full random password and mnemonic generation.

#### Scenario: Successful generate run

- **WHEN** the user runs `tale generate --language es`
- **THEN** the CLI prints the password, mnemonic phrases, entropy, and recovery rules to stdout

### Requirement: TUI command

The system SHALL provide a `tui` subcommand for interactive mnemonic construction.

#### Scenario: TUI with generated password

- **WHEN** the user runs `tale tui --language en`
- **THEN** the system generates a password and opens an interactive UI to select words per block

#### Scenario: TUI with existing password

- **WHEN** the user runs `tale tui --password abcdefghijklmnopqrst`
- **THEN** the system uses the provided password and opens the UI without regenerating it

### Requirement: TUI word selection

In TUI mode, the system SHALL present filtered word candidates per grammatical slot and allow the user to select one word per slot for each block.

#### Scenario: English candidate filtering

- **WHEN** the user navigates block 1 with letters `lcds` in TUI mode with language `en`
- **THEN** adjective choices come from `data/en/l/adjectives.tsv`, nouns from `c/nouns.tsv`, verbs from `d/verbs.tsv`, nouns from `s/nouns.tsv`

#### Scenario: Spanish candidate filtering

- **WHEN** the user navigates block 1 with letters `lcds` in TUI mode with language `es`
- **THEN** noun choices come from `data/es/l/nouns.tsv`, adjectives from `c/adjectives.tsv`, verbs from `d/verbs.tsv`, nouns from `s/nouns.tsv`

### Requirement: Language and length flags

`generate` and `tui` SHALL accept `--language` and `--length` flags consistent with language configuration defaults.

#### Scenario: Length override in TUI

- **WHEN** the user runs `tale tui --length 16`
- **THEN** a 16-character password is used (4 blocks of 4)

### Requirement: Exit codes

The CLI SHALL exit with code 0 on success and non-zero on validation or generation failure.

#### Scenario: Invalid language

- **WHEN** the user specifies a language that cannot be resolved
- **THEN** the CLI prints an error message and exits with a non-zero code

### Requirement: Offline operation

The CLI SHALL work without network access.

#### Scenario: Offline stats and generation

- **WHEN** the machine has no network connectivity
- **THEN** `tale stats` and `generate`/`tui` commands use embedded or local user word data
