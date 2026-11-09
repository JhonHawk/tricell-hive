## MODIFIED Requirements

### Requirement: Viewer-first walkthrough publication
Substantive publication SHALL use only the compatible v3 API and confirmed readback. A missing executable or an incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon or producing any other format; a stopped viewer SHALL do the same, except that semantic open, stage, evidence and capture declarations SHALL be queued as «Queued semantic publication while the viewer is unavailable» sets. Previously generated HTML pages and user-selected artifact units SHALL not be deleted or rewritten by the removal of the renderer. Research and plan SHALL NOT publish to the viewer; a work's viewer session SHALL start at its first UI capture in a build or verify flow that was explicitly invoked (its command typed, its route accepted, research continuing into build at the user's request, or build continuing into verify), created with a summary of about one paragraph on what the work is about. The agent SHALL NOT open the viewer on its own: at each stop where the viewer holds captures for the user to review (a UI walkthrough, including the build's first rendered view) it SHALL offer to open it inside what that stop already asks, until the user accepts once in the conversation. When the stop asks through the host's question tool, the offer SHALL be one more option of the stop's own approval, acceptance or validation question, never a separate question, a question with its own decision ID or its own question-tool call; choosing it SHALL open the work once and ask that same question again, unchanged, with the same ID and without the open option, keeping the answers to the round's other questions. The open option SHALL NOT be the first option and SHALL NOT be starred. Otherwise, or when the question already holds the most options the tool accepts, the offer SHALL be a reply line of the stop's waiting line or next-step block, and using it SHALL open the work while the stop keeps waiting for its answer. After acceptance the agent SHALL NOT offer or open it again unless the user asks.

#### Scenario: Binary present, server incompatible
- **WHEN** publication finds a v1 server
- **THEN** it requests the coordinated v3 update and preserves inputs without HTML substitution

#### Scenario: User opens the viewer before deciding
- **WHEN** a UI validation stop asks its acceptance question through the host's question tool and the user picks the open-viewer option
- **THEN** the work opens in the viewer once
- **AND** the same acceptance question is asked again with the same ID and without the open-viewer option, and the round's other answers stand

#### Scenario: Plan approval without the viewer
- **WHEN** a plan reaches its approval stop
- **THEN** no viewer work exists for it and the stop offers no open-viewer option

### Requirement: Proportional identity-aware publication procedure
Shared walkthrough instructions SHALL name host capabilities rather than concrete host tool identifiers. Identity acquisition SHALL belong to the manager, which resolves it from verified host sources; on the semantic path the procedure SHALL NOT ask the agent to discover, guess or pass its own native ID, except as each host's one conditional identity reference, verified against current official documentation and installed versions, sets when the manager reports identity unresolved; unsupported discovery SHALL not be represented as a native ID. The canonical walkthrough procedure SHALL publish to the viewer only UI captures: a flow that produces no capture SHALL create no viewer work, and a flow that does SHALL publish each batch of captures with one capture command. At the work's completion phase it SHALL close the work, or declare it in_review, paused or awaiting validation, as «Viewer finalization before deferred hygiene» sets. The procedure SHALL NOT ask the agent to declare stages or stage states during the flow, register or update tasks, publish documents, read the publication contract or read the session back after a capture; the capture command's receipt SHALL confirm that write, and the completion declaration's read-back SHALL confirm the work at close. The procedure SHALL be read once per session and again after a compaction. Initial affected-layout verification SHALL cover relevant desktop/mobile widths and supported themes; correction rounds SHALL check affected regions/neighbors, relevant states/widths and theme-sensitive behavior only. Navigation changes SHALL exercise their affected sequence. Publication consumers SHALL follow flow-walkthrough; review/capture consumers SHALL follow the single verify-owned conditional UI policy without a conflicting full-matrix iteration instruction. Human responses SHALL continue through native host questions in this delivery.

#### Scenario: Flow without captures
- WHEN a producer plans, builds and closes a change with no UI capture
- THEN no viewer command runs during the flow and the viewer shows no work for it.

#### Scenario: Publication without per-write reads
- WHEN a producer publishes two batches of captures in the same work
- THEN each batch is one capture command confirmed by its receipt, with no session read in between
- AND the work is read back once, by the completion declaration.

### Requirement: Declared flow stages and immutable documents
Stages SHALL include research, plan, build, verify and delivery. Stage states SHALL include in_progress, in_review, awaiting_validation, paused, completed and failed. Work states SHALL distinguish active, in_review, awaiting_validation, paused, closed and failed. Labels SHALL be short, reasons separate. One accepted lead SHALL control at most one foreground; global/foreground review or validation changes SHALL be atomic. Parallel review SHALL NOT change that foreground. completed/failed occurrences SHALL be terminal; re-entry SHALL create a linked occurrence.

Distinct transitions SHALL append immutable server-timestamped audit events with declared producer, reason and previous/next states. Membership/handoff and all other new mutations SHALL retain actor provenance and audit; local browser renaming, archiving, pinning and confirmed deletion SHALL use local_viewer actor provenance with existing scope/origin/version checks; audit events SHALL NOT be editable or withdrawable. A producer that declares a user question SHALL publish awaiting_validation first; the canonical procedure declares none during a flow. Responses SHALL reference concrete targets when known, never be inferred from silence, and resumption SHALL be explicit. Closure SHALL require no foreground.

Documents from all five stage types SHALL use one read-only Markdown Sheet, full width mobile/two-thirds desktop, fixed scrollable tabs and one close action/publication indicator. Only structured plan deltas SHALL group under Specs. Work-scoped document_key SHALL separate independent revision chains. Successors SHALL identify their predecessor and cross-occurrence continuation SHALL require linked same-type stages. Open content SHALL not change until user action. Only a later revision in that chain SHALL trigger an update notice; switching SHALL preserve the selected label or announce its removal. Comparisons SHALL identify added/changed/removed tabs by label/document identity; the first revision SHALL be labeled initial.

#### Scenario: Declared wait for a choice
- WHEN a producer declares a user question on a work
- THEN work/foreground show Validation with a separate reason and immutable transition before asking.
- AND a complete published document does not imply approval or completed work.

#### Scenario: Parallel review
- WHEN a collaborator posts review progress or findings during build
- THEN its declared producer and stage target are visible while build stays In progress.

#### Scenario: Independent documents
- WHEN separate document keys use identical tab labels
- THEN comparisons and update notices stay separate, including across linked same-type occurrences.

## ADDED Requirements

### Requirement: Capture publication in one command
The manager SHALL provide one capture command that publishes a batch of captures for one screen, named by its base view title. When the session has no open work, it SHALL open one with the given key, or the current branch when no key is given, and the given summary or a default one. At send time it SHALL start the target stage when it has no occurrence, resume it when paused, and open a linked occurrence when its last one is terminal. When no candidate is given, it SHALL use the project's current commit, resolved when the command runs, and SHALL refuse without publishing when that commit's branch or base cannot be resolved. At send time it SHALL read the work's views, take those whose title without a trailing ` · #N` equals the base title, supersede the one with the highest N and publish the view titled with the base title followed by ` · #N+1`, or ` · #1` when none exists. Repeating the latest iteration's images, captions and candidate SHALL return `unchanged` and SHALL NOT create an iteration. While the viewer is unavailable, the command SHALL be queued as its own operation, with its images frozen, and resolved at send time in queue order. Its compact receipt SHALL include the view ID, the iteration number and the base title. The command SHALL belong to the build and verify flows of the offline contract. The existing open, stage, evidence and close operations SHALL remain available and unchanged.

#### Scenario: First capture of a session
- WHEN a producer with no open work runs the capture command for the Login screen with one image
- THEN the viewer shows a new work with a started stage and the view "Login · #1" linked to the current commit, published by one command.

#### Scenario: Later iteration of the same screen
- WHEN the producer captures Login again after a correction, and then captures Profile
- THEN "Login · #2" supersedes "Login · #1", and "Profile · #1" supersedes nothing.

#### Scenario: Two captures while the viewer is stopped
- WHEN the producer captures Login twice while the viewer is stopped, and the viewer later runs again
- THEN both commands report queued without starting the viewer, and the next command or hook run creates the work, the stage, "Login · #1" and "Login · #2" superseding it, once and in order.

#### Scenario: Unchanged repeat
- WHEN the producer repeats the latest Login capture with the same images, captions and commit
- THEN the command returns unchanged and no "Login · #3" exists.
