# comment-audit

## ADDED Requirements

### Requirement: Vendored code stays out of the audit and inside the batch check

The comment audit SHALL leave vendored third-party code out of its inventory and its lanes, listed by its path from the repository root in the ratchet's `EXCLUDED`, and the ratchet's `--same-code` check SHALL fail when a batch changes, adds, deletes or moves out any tracked file of any type under an excluded path, so leaving code out of the audit never leaves it out of the proof that a batch removed only comments.

#### Scenario: A batch damages vendored code
- **GIVEN** `vendor/lib.js` listed in `EXCLUDED`
- **WHEN** a removal batch changes `vendor/lib.js` and `--same-code` runs against the record commit
- **THEN** it exits 1 and names `vendor/lib.js`

#### Scenario: A lane child does not read vendored comments
- **GIVEN** a lane whose package holds a vendored folder listed in `EXCLUDED`
- **WHEN** the skill briefs the lane's child
- **THEN** the child marks only the comments of the lane's inventory file, which leaves out that folder
