## ADDED Requirements

### Requirement: Cycle semantic candidates in the TUI

The interactive picker SHALL display the top-ranked semantic candidate by default with a pool position indicator, and SHALL let the user cycle forward/backward through the block's candidate pool without regenerating the password.

#### Scenario: Cycle forward and back

- **WHEN** the user presses the cycle keys in the picker
- **THEN** the displayed phrase moves to the next/previous pool candidate and the indicator shows the new position (e.g. "3/10")

#### Scenario: Default candidate

- **WHEN** the picker opens for a block
- **THEN** it shows the rank-1 candidate with the indicator "1/10"
