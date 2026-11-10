## MODIFIED Requirements

### Requirement: Fixed-role context delivery without authority expansion
The manager SHALL generate the two fixed reference snapshots from its validated release payload into the installed plugin, serializing content as data and validating derived identity before writes. When explicitly active, the loader SHALL resolve the native agent's identifier and supply that generated snapshot for `hive-review-plan` or `hive-review-harness` before model dispatch without opening reference or metadata files in its request loop. It SHALL NOT infer prompt intent, alter tools, permissions or generation settings, or persist message changes. Source resources missing or invalid at generation SHALL reject delivery before writes; runtime missing-snapshot diagnostics SHALL be bounded and SHALL NOT claim native request abortion. The global core's obligation to report dependencies remains. Unknown roles SHALL remain unchanged.

The fixed mappings SHALL be `hive-review-plan` to
`flow-plan/SKILL.md` and `hive-review-harness` to
`harness-audit/SKILL.md`. The identifier SHALL be resolved from the native
agent lookup's `data.id`, then its `id`, then the hook event's agent value (also
when the lookup fails), and SHALL NOT be taken from the display name.
Deduplication SHALL affect only parts added by this plugin to the current
request, not foreign content or persistent session flags.

#### Scenario: Covered native role
- **WHEN** a context request belongs to a covered role and its generated snapshot is present
- **THEN** one copy is supplied by the loader to the outgoing request, and a later request receives that snapshot without depending on persistent session flags or reopening mutable reference paths.

#### Scenario: Reference files change after generation
- **WHEN** a local reference is edited or its directory replaced after the plugin snapshot was generated
- **THEN** the request loop supplies the previously generated snapshot, performs no live read of that reference, and changes only after regeneration and plugin reload.

