## ADDED Requirements

### Requirement: Fixed phrase template

The system SHALL build each mnemonic block as exactly four words following the `phrase_template` defined in the active language profile.

#### Scenario: English template order

- **WHEN** a 4-character block `lcds` is converted to a phrase in English
- **THEN** the phrase consists of an adjective starting with `l`, a noun starting with `c`, a verb starting with `d`, and a noun starting with `s`

#### Scenario: Spanish template order

- **WHEN** a 4-character block `lcds` is converted to a phrase in Spanish
- **THEN** the phrase consists of a noun starting with `l`, an adjective starting with `c`, a verb starting with `d`, and a noun starting with `s`

### Requirement: Initial letter matching

The system SHALL match words by the normalized initial letter of each word against the corresponding character in the password block.

#### Scenario: Matching initials

- **WHEN** the block is `mrtp`
- **THEN** valid phrases use words whose normalized initials are `m`, `r`, `t`, `p` in template order

### Requirement: Normalization rules

The system SHALL normalize words for initial-letter extraction using these rules:

1. Convert to lowercase
2. Strip diacritics on the first character (`á→a`, `é→e`, `í→i`, `ó→o`, `ú→u`, `ü→u`)
3. Use the first character after normalization (no silent-letter exceptions)

#### Scenario: Accented vowel at word start

- **WHEN** the word is `árbol`
- **THEN** its normalized initial is `a`

#### Scenario: Accented vowel indexed in dictionary

- **WHEN** building the Spanish noun index for initial `a`
- **THEN** the word `árbol` is included in that index

#### Scenario: Word with enye not at start

- **WHEN** the word is `niño`
- **THEN** its normalized initial is `n` (first character is `n`, not `ñ`)

### Requirement: Spanish words starting with enye excluded

The system SHALL NOT include Spanish words whose first character is `ñ` in bundled word lists.

#### Scenario: Enye-initial word omitted

- **WHEN** the prepare-wordlists script encounters `ñu` or `ñandú`
- **THEN** those words are excluded from the Spanish bundled lists

#### Scenario: Password alphabet excludes enye

- **WHEN** the `spanish_default` profile alphabet is used for password generation
- **THEN** the alphabet does not contain the letter `ñ`

### Requirement: Dictionary scope

The system SHALL use only bundled word lists containing nouns, adjectives, and verbs for the active language.

#### Scenario: No other parts of speech

- **WHEN** building phrase candidates
- **THEN** the system does not use adverbs, prepositions, or other POS not in the bundled lists

### Requirement: Automatic phrase selection

In random mode, the system SHALL select the highest-scoring valid phrase per block using a deterministic memorability heuristic.

#### Scenario: Deterministic best phrase

- **WHEN** the same password block is processed twice in random mode
- **THEN** the same phrase is selected both times

### Requirement: Recovery instructions

The system SHALL output instructions explaining how to recover the password from the mnemonic phrases using the normalization rules.

#### Scenario: Recovery rules in output

- **WHEN** generation completes
- **THEN** the output includes a "Recovery" section describing how to extract initials from each phrase word

### Requirement: Phrase output per block

The system SHALL output one mnemonic phrase per password block, labeled with block index and letters.

#### Scenario: Five phrases for default password

- **WHEN** a 20-character password is generated
- **THEN** the output shows 5 labeled phrases, one per block
