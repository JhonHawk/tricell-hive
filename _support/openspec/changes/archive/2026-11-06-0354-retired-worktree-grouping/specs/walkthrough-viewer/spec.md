## MODIFIED Requirements

### Requirement: Explicit presentation grouping independent of project authority

The viewer SHALL derive a read-only navigation group from an exact registered original source root's valid `.engram/config.json` `project_name`. Equal trimmed case-sensitive explicit names SHALL share a presentation group across different canonical project keys. Missing configuration SHALL retain canonical-project grouping. Invalid or unavailable configuration SHALL retain canonical-project grouping and expose a bounded recoverable diagnostic without file content. An absent registered source root MAY use the unambiguous current explicit name of another available registered source represented in the index with the exact same persisted canonical project key. This exception SHALL preserve original origin metadata and SHALL NOT infer identity from equal display names; missing witnesses or conflicting explicit names SHALL keep the fallback and diagnostic. Configuration lookup SHALL NOT ascend, scan child repositories, invoke Engram or follow configuration symlinks. Reading and grouping SHALL NOT rewrite stored works or change identities, provenance, collaboration eligibility, memberships or receipts.

#### Scenario: Existing works in separate projects share a declared name
- WHEN two stored works have distinct project keys and their original registered roots declare the same valid name
- THEN navigation shows one named group with both works and the correct visible count
- AND their IDs, project keys, sources and join boundaries remain unchanged.

#### Scenario: Configuration changes after publication
- WHEN the explicit name changes, is removed or becomes invalid
- THEN the next completed index refresh changes only the presentation, using canonical fallback on absence or invalidity
- AND invalidity is reported without hiding the works or changing their stored bytes.
