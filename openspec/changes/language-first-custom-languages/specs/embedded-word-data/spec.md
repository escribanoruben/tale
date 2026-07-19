## ADDED Requirements

### Requirement: Embedded word data for builtins

For builtin language configurations, the system SHALL load word lists from data embedded in the binary.

#### Scenario: Builtin uses embed

- **WHEN** language `en` resolves as builtin
- **THEN** word lookups use embedded `en/{letter}/{pos}.tsv` files

### Requirement: User word data replaces embed

When a language resolves as a user configuration (non-builtin), the system SHALL load word lists only from `~/.tale/data/{language}/` and SHALL NOT fall back to embedded lists for missing files.

#### Scenario: User data only

- **WHEN** language `de` resolves as a user configuration
- **THEN** word lookups read exclusively under the user data directory for `de`

#### Scenario: Missing user word file

- **WHEN** a required TSV is missing under the user data directory for a customized language
- **THEN** the system returns a clear error (no silent fallback to embed)

### Requirement: Stats inspects embedded data

`tale stats` SHALL report counts from embedded word data.

#### Scenario: Stats offline

- **WHEN** the user runs `tale stats`
- **THEN** the CLI prints per-language totals from embedded data without network access
