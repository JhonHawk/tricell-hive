## MODIFIED Requirements

### Requirement: Current source classification
The server SHALL derive `tracked`, `local` or `unknown` source status for every stored copy before publishing its initial or reconstructed view and on each complete scan. It SHALL place tracked copies outside the normal project/session list, preserve their files, keys, ingestion time, pins, archive state and existing retention behavior, and show an explanatory entry without rendering the document's instructions. Authenticated normal reads of a tracked copy's manifest or files SHALL return 409 without document bytes; authenticated archive/restore, pin and delete operations SHALL remain available. Missing sources without positive tracking evidence and failed classification SHALL be shown as unknown with a warning. Local status SHALL NOT claim verified agent authorship. The list and open document SHALL update source warnings when source status or reason changes, even when content hashes remain unchanged. Classification SHALL NOT require changes to stored per-copy metadata.

#### Scenario: Tracking changes without content changes
- **WHEN** an existing source becomes tracked or stops being tracked while its bytes remain unchanged and its file remains present
- **THEN** the next complete scan changes its placement without resetting its ingestion time or pin.

#### Scenario: Source missing or tracking unavailable
- **WHEN** the source is missing without positive tracking evidence or its tracking cannot be determined
- **THEN** its retained copy remains accessible with an explicit unknown-origin warning rather than a verified-authorship claim.
