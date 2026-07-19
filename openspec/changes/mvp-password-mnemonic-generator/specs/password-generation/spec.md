## ADDED Requirements

### Requirement: Cryptographically secure password generation

The system SHALL generate passwords using a cryptographically secure pseudo-random number generator (`crypto/rand`).

#### Scenario: Password uses CSPRNG

- **WHEN** the user runs password generation
- **THEN** each character is selected using `crypto/rand` (not `math/rand`)

### Requirement: Configurable password length

The system SHALL support a configurable password length via profile or CLI flag, with a default of 20 characters.

#### Scenario: Default length

- **WHEN** the user runs `generate` without `--length`
- **THEN** the generated password is exactly 20 characters long

#### Scenario: Custom length

- **WHEN** the user passes `--length 24`
- **THEN** the generated password is exactly 24 characters long

### Requirement: Language-specific alphabet

The system SHALL generate passwords using only characters from the alphabet defined in the active profile.

#### Scenario: Alphabet constraint

- **WHEN** a password is generated with profile `spanish_default`
- **THEN** every character in the password belongs to that profile's `alphabet` field

### Requirement: Lowercase output

The system SHALL output passwords in lowercase only.

#### Scenario: No uppercase in password

- **WHEN** a password is generated
- **THEN** the password contains only lowercase letters from the profile alphabet

### Requirement: Block splitting

The system SHALL split the password into consecutive blocks of size defined by the profile (`block_size`, default 4).

#### Scenario: Default block layout

- **WHEN** a 20-character password is generated with `block_size = 4`
- **THEN** the password is divided into 5 blocks of 4 characters each

### Requirement: Rejection on uncovered blocks

The system SHALL regenerate a block (or the full password) when any block has zero valid mnemonic phrase candidates.

#### Scenario: Block without candidates is rejected

- **WHEN** a generated block has no valid noun-adjective-verb-noun phrase matching its letters
- **THEN** the system regenerates until all blocks have at least one valid phrase candidate

### Requirement: Entropy reporting

The system SHALL display approximate entropy in bits for the generated password.

#### Scenario: Entropy shown in output

- **WHEN** generation completes successfully
- **THEN** the output includes entropy calculated as `length × log2(alphabet_size)`
