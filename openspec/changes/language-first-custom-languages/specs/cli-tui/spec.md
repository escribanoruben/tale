## ADDED Requirements

### Requirement: Language flag for generate and TUI

`generate` and `tui` SHALL accept `--language` (not `--profile`) to select the language configuration, and SHALL accept `--length` consistent with that configuration’s block size.

#### Scenario: Generate with Spanish

- **WHEN** the user runs `tale generate --language es`
- **THEN** the CLI prints the password, mnemonic phrases, entropy, and recovery rules using the Spanish configuration

#### Scenario: TUI with English

- **WHEN** the user runs `tale tui --language en`
- **THEN** the system generates a password for English and opens the interactive word picker

#### Scenario: Invalid language

- **WHEN** the user specifies a language that cannot be resolved
- **THEN** the CLI prints an error and exits non-zero

### Requirement: Language management commands

The CLI SHALL provide language management subcommands to list, show, clone (`init`), and create (`new`) languages. A deprecated `profile` alias MAY forward to the same commands with a deprecation hint.

#### Scenario: List languages

- **WHEN** the user runs `tale language list`
- **THEN** the CLI lists builtin and user languages with display names and source (`builtin` or `customized`)

#### Scenario: Show language

- **WHEN** the user runs `tale language show es`
- **THEN** the CLI prints language, display name, source, alphabet, and related generation fields

### Requirement: Override visibility

When generation uses a non-builtin (user) language configuration, the CLI SHALL indicate that user configuration and/or user word data are in effect.

#### Scenario: Generate with customized language

- **WHEN** the user runs `tale generate --language es` after creating a user profile for `es`
- **THEN** the output or a clear status line indicates that user overrides are active

### Requirement: Offline operation

The CLI SHALL work without network access for `stats`, `generate`, and `tui` when using embedded or local user data.

#### Scenario: Offline generate

- **WHEN** the machine has no network connectivity
- **THEN** `tale generate --language en` succeeds using embedded data
