## ADDED Requirements

### Requirement: Strategy configuration

Language profiles SHALL accept a `strategy` key with values `chained_story` (default), `noun_field`, or `noun_path`. For `chained_story` the phrase count derives from `password_length / block_size`; for noun strategies the password length equals the noun count (one noun per password letter).

#### Scenario: Default strategy

- **WHEN** a profile omits `strategy`
- **THEN** generation uses the chained story strategy

#### Scenario: Invalid strategy

- **WHEN** a profile declares an unknown strategy
- **THEN** profile validation fails with a clear error

### Requirement: Chained story strategy

`chained_story` SHALL produce one phrase per block whose initials spell the password, optimized by the whole-story dynamic program (anchored scoring, chain edges between consecutive phrases, closure edge, refrains allowed) as specified in `semantic-mnemonics`.

#### Scenario: Chained generation

- **WHEN** a 20-letter password is generated with `chained_story`
- **THEN** five phrases are produced whose initials spell the password

### Requirement: Noun field strategy

`noun_field` SHALL select one noun per password letter such that all nouns orbit a common semantic centroid (seeded by CSPRNG among the first letter's candidates, refined over re-centering rounds), with no repeated words or roots and low-imageability words excluded. Noun initials, in order, spell the password.

#### Scenario: Field coherence

- **WHEN** nouns are selected for a password
- **THEN** the mean pairwise cosine between chosen nouns substantially exceeds the random baseline

#### Scenario: No repeated words

- **WHEN** the field is selected
- **THEN** no word or root appears twice

### Requirement: Noun path strategy

`noun_path` SHALL select one noun per password letter maximizing consecutive-adjacency cosine (exact dynamic program over consecutive positions), with near-synonym hops (cosine above 0.75) penalized at weight 2.0 and no repeated words in the path.

#### Scenario: Path fluency

- **WHEN** nouns are selected for a password
- **THEN** each chosen noun is substantially closer to the next chosen noun than the random baseline

#### Scenario: No repeated words

- **WHEN** the optimal path repeats a word
- **THEN** candidates containing that word are masked and the dynamic program is re-run until the path is repetition-free

### Requirement: Noun strategies require vectors

`noun_field` and `noun_path` SHALL fail with a clear error when the language has no embedded vectors (custom languages); `chained_story` keeps its silent fallback to frequency scoring.

#### Scenario: Custom language with noun strategy

- **WHEN** a language without embedded vectors declares `noun_field`
- **THEN** generation reports that the strategy requires semantic vectors instead of silently degrading
