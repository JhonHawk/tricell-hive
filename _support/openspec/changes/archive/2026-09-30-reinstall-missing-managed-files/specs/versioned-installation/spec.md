## ADDED Requirements

### Requirement: Deleted managed content is reinstalled
The system SHALL treat managed content that the user deleted by hand as absent: `hive install` and `hive update` SHALL recreate it as a fresh installation would, and removal SHALL complete without writing when what it would delete is already gone. The system SHALL keep refusing content that was edited rather than deleted.

#### Scenario: Deleted managed file
- **WHEN** a skill, agent or link that Hive installed was deleted and the user runs `hive install` or `hive update`
- **THEN** the file is written again with its managed bytes and mode, and `hive status` no longer reports drift for it.

#### Scenario: Deleted managed block
- **WHEN** the Hive or voice block was removed from a user file, or the user file that held it was deleted
- **THEN** `hive install` adds the block again, creating the file when needed, and preserves the user's text outside the block.

#### Scenario: Removal of deleted content
- **WHEN** the user removes a host whose managed content was already deleted
- **THEN** removal completes without writing and drops the record; a shared resource still used by another host stays absent until the next install or update.

#### Scenario: Edited content still refused
- **WHEN** managed content was edited, its permissions changed, or its markers were damaged
- **THEN** install, update and remove refuse with a message naming the path, and preserve the edit.
