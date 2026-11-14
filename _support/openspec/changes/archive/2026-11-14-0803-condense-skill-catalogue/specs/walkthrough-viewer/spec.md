## MODIFIED Requirements

### Requirement: Proportional identity-aware publication procedure
Shared walkthrough instructions SHALL name host capabilities rather than concrete host tool identifiers. Identity acquisition SHALL belong to the manager, which resolves it from verified host sources; on the semantic path the procedure SHALL NOT ask the agent to discover, guess or pass its own native ID, except as each host's one conditional identity reference, verified against current official documentation and installed versions, sets when the manager reports identity unresolved; unsupported discovery SHALL not be represented as a native ID. The canonical walkthrough procedure SHALL publish to the viewer only UI captures: a flow that produces no capture SHALL create no viewer work, and a flow that does SHALL publish each batch of captures with one capture command. At the work's completion phase it SHALL close the work, or declare it in_review, paused or awaiting validation, as «Viewer finalization before deferred hygiene» sets. The procedure SHALL NOT ask the agent to declare stages or stage states during the flow, register or update tasks, publish documents, read the publication contract or read the session back after a capture; the capture command's receipt SHALL confirm that write, and the completion declaration's read-back SHALL confirm the work at close. The procedure SHALL be read once per session and again after a compaction. Initial affected-layout verification SHALL cover relevant desktop/mobile widths and supported themes; correction rounds SHALL check affected regions/neighbors, relevant states/widths and theme-sensitive behavior only. Navigation changes SHALL exercise their affected sequence. Publication consumers SHALL follow the walkthrough reference of flow-verify; review/capture consumers SHALL follow the single verify-owned conditional UI policy without a conflicting full-matrix iteration instruction. Human responses SHALL continue through native host questions in this delivery.

#### Scenario: Flow without captures
- WHEN a producer plans, builds and closes a change with no UI capture
- THEN no viewer command runs during the flow and the viewer shows no work for it.

#### Scenario: Publication without per-write reads
- WHEN a producer publishes two batches of captures in the same work
- THEN each batch is one capture command confirmed by its receipt, with no session read in between
- AND the work is read back once, by the completion declaration.
