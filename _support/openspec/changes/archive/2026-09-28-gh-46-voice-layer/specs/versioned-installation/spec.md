## ADDED Requirements

### Requirement: Optional voice layer
The manager SHALL offer an optional voice layer, off by default, as a managed span separate from the Hive guidance block in each registered user-scope instruction file. The voice SHALL NOT change release identifiers or the communication rules, and its text SHALL limit itself to the messages the user reads in the main conversation. Installing a new release SHALL keep the chosen voice and regenerate its span when the voice source changed. Removing Hive from the last consumer of a file SHALL remove that file's voice span.

#### Scenario: Voice set and removed
- **WHEN** the user sets a voice and later turns it off
- **THEN** each instruction file returns byte for byte to its content before the voice was set.

#### Scenario: Release update with a voice
- **WHEN** a new release is installed while a voice is active
- **THEN** the voice span remains, regenerated with the same choice if its source text changed.

#### Scenario: Edited voice span
- **WHEN** the user edited the voice span by hand
- **THEN** any operation on that file reports a conflict and preserves the edit.
