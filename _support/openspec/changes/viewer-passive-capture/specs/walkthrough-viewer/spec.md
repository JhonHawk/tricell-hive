# Walkthrough viewer delta

This delta is written against the base spec after `gh-311-minimal-research-entry` is archived, whose delta also modifies «Viewer-first walkthrough publication».

## ADDED Requirements

### Requirement: Host-observed events
The viewer SHALL accept host events reported by the local `hive hook` process through a capability-gated endpoint authenticated by the local viewer token, without a browser origin. A host event SHALL carry the host, its native session ID, the project directory, an event type, bounded data and an operation ID derived from the host event so a resend applies once; the hook SHALL resolve the session's current work and that work's active stage from the manager's current-work record and send them. Host events SHALL act as the accepted native producer of that session, SHALL be stored with the provenance `host_observed`, distinct from `agent_reported` and `server`, and SHALL be rejected for a session that is not an accepted producer of the work.

Every host event SHALL update the last activity of the works of its native session as maintenance metadata, not as a timeline event, and SHALL NOT invalidate the expected versions of agent mutations. When the current work has no agent-registered task, native task-list changes SHALL become the work's tasks, keyed `host:<native task ID>`, assigned to that session's producer and bound to the active stage, up to 50 tasks; when the work has agent-registered tasks, native task-list changes SHALL only update last activity. Without a current work or an active stage, a host event SHALL only update last activity. Host events SHALL NOT start, complete or close a stage, change producers or authority, record approvals or validations, or close, pause or fail a work; a session end SHALL update last activity only and SHALL NOT be read as process death. The UI SHALL show host-observed provenance apart from agent-published content.

#### Scenario: Native task completed in a build without a plan
- WHEN the host reports a native task created and then completed in a session whose current work has an active build stage and no agent-registered task
- THEN that work shows the task with host-observed provenance and its completion, and no stage changes state.

#### Scenario: Plan tasks already registered
- WHEN the host reports native task changes for a work whose plan tasks the agent registered
- THEN no task is added or changed and only last activity moves.

#### Scenario: Event attempts a lifecycle change
- WHEN a host event carries data that would complete a stage or close the work
- THEN the viewer rejects that effect and stores no change to stages, authority or work state.

#### Scenario: Agent write after host activity
- WHEN host events arrive between an agent's stage declaration and its next conditional mutation
- THEN the agent's mutation commits with the versions from its last receipt.

### Requirement: Resolved native identity and current work
The manager SHALL resolve the producer identity of a semantic command, when no explicit identity is given, from the first available source: a host environment variable verified against the host's official documentation and installed version, then a process-to-session map written by the host's session-start hook and found through the process ancestry. A map entry SHALL be used only when it names exactly one live session for that ancestor and is newer than the host's last session change for that process; otherwise identity is unresolved. A resolved identity is the host's actual native session ID; with no source, or an ambiguous one, the command SHALL be rejected as missing identity, and the host's conditional identity reference SHALL say how to pass it explicitly. The manager SHALL keep the current work of each (host, native session) so semantic commands may omit the work ID while exactly one work is current, and SHALL require the work ID otherwise. Opening a work SHALL register its project root when it is not registered.

#### Scenario: Claude Code without flags
- WHEN a semantic command runs in a Claude Code shell that exposes the session ID and omits identity flags
- THEN it publishes as that native session.

#### Scenario: Two sessions under one process
- WHEN the session map holds two live sessions for the nearest host ancestor
- THEN the command fails as missing identity and writes nothing.

#### Scenario: No identity source
- WHEN no flag, verified variable or session map applies
- THEN the command fails as missing identity and writes nothing.

### Requirement: Declarative stage publication without a change folder
The manager SHALL open a work without a change folder or plan revision and SHALL apply, in one command, a stage state with its reason, the registration of missing tasks with their stage segments bound to that stage, task state changes and an optional close that first completes the active stage and then reads the committed work back. Repeating the same declaration SHALL return `unchanged`. Its compact receipt SHALL include task and segment IDs. One evidence command SHALL upload captures, register the candidate when needed and append the event with its view, including a view that supersedes an earlier one. Low-level operations SHALL remain available and unchanged.

#### Scenario: Build without plan
- WHEN a producer opens a work without a change folder, declares a build stage with three tasks and their states, and closes it
- THEN the viewer shows the stage, the three tasks bound to it and the closed work, published by at most three commands, with no separate session read and no binding error.

#### Scenario: Captures for validation
- WHEN a producer publishes two captures with a validation view that supersedes an earlier view in one evidence command
- THEN the viewer shows the new view with both images, linked to the candidate, and the earlier view as superseded.

### Requirement: Queued semantic publication while the viewer is unavailable
When the viewer is not running or does not answer, a semantic open, stage declaration or evidence command SHALL record its semantic input and operation identity durably, report it as queued, and SHALL NOT start the viewer. A queued open SHALL hold its work key, from which the viewer ID is derived when it is sent. The next semantic command or hook run with a running viewer SHALL send queued declarations in order before its own, building their requests against the current remote state; a declaration that no longer applies, such as a stage already completed, SHALL be reported as rejected with its reason and SHALL NOT overwrite remote state. Locks SHALL be held per work; commands on different works SHALL NOT block each other.

#### Scenario: Viewer stopped at session start
- WHEN a producer opens a work and declares a build stage while the viewer is stopped, and the viewer later runs again
- THEN both are reported queued, and the next command creates the work and the stage once, in order.

#### Scenario: Remote state moved on
- WHEN a queued declaration would complete a stage that another producer already completed
- THEN that part is reported rejected and the stage keeps its recorded state.

### Requirement: Hook registration per host
The manager SHALL install, update and remove only its own hook entries: entries whose command runs `hive hook` in Claude Code's user settings and Codex's and Cursor's hooks files, its own native hook file in Grok, its own event plugin in OpenCode's plugin directory and its own extension in Pi, leaving every other hook and setting byte-identical and keeping a backup of each edited JSON file. It SHALL NOT write Codex hook trust; its report SHALL tell the user to approve the hook in Codex. Hook processes SHALL print nothing, finish within the host's time limit and exit successfully on any error. Host support not verified against current documentation and the installed version SHALL be reported as unverified.

#### Scenario: Existing third-party hooks
- WHEN Claude Code's settings already hold other tools' hooks and Hive installs and later removes its hooks
- THEN the other hooks and settings stay byte-identical after both operations.

## MODIFIED Requirements

### Requirement: Required native session registration
New timeline sessions SHALL require the resolved registered project, canonical host and actual native session ID, and MAY carry a work key. An identity resolved by the manager from a verified host source, as «Resolved native identity and current work» sets, is the actual native session ID, not a synthesized one. The server SHALL derive a distinct viewer ID from the project, host, native session ID and the effective work key: the given work key, or, when absent, the creating operation ID. It SHALL enforce exact uniqueness across concurrent requests, retries and restarts. Sessions stored before work keys existed SHALL keep their IDs. Missing identity SHALL be rejected rather than synthesized. Native IDs SHALL remain opaque and separate from viewer storage IDs. Existing registration SHALL not overwrite metadata or reopen closed/deleted work. API registration SHALL recover persisted root/project bindings before hashing identity; temporary metadata failure SHALL not replace a known project key with a path fallback, and unresolved/conflicting identity SHALL fail explicitly.

#### Scenario: Resume or concurrent creation
- WHEN two clients from the same native session register the same project/host/native-session tuple with the same work key
- THEN both receive the same viewer ID and existing session content remains intact.

#### Scenario: Two works in one native session
- WHEN one native session registers two different work keys, or two creations without a work key with different operation IDs
- THEN the server creates two viewer sessions.

#### Scenario: Separation
- WHEN the native ID is equal but project or host differs
- THEN the server creates separate viewer sessions.

### Requirement: Proportional identity-aware publication procedure
Shared walkthrough instructions SHALL name host capabilities rather than concrete host tool identifiers. Identity acquisition SHALL belong to the manager, which resolves it from verified host sources; on the semantic path the procedure SHALL NOT ask the agent to discover, guess or pass its own native ID, except as each host's one conditional identity reference, verified against current official documentation and installed versions, sets when the manager reports identity unresolved; unsupported discovery SHALL not be represented as a native ID. The canonical walkthrough procedure SHALL open or resume the work, keep its work key in the work checkpoint (the manager keeps the returned IDs), publish each stage change with one declaration, upload immutable new captures with one evidence command and publish typed meaningful milestones; tasks the host already reports SHALL NOT be registered again by hand. A mutation's returned receipt (IDs and versions) SHALL confirm that write; the procedure SHALL read the committed session back before claiming to the user that a publication is visible, at close, and when resuming after an interruption or a version conflict, and SHALL NOT require a session read after every mutation; a close declaration that reads the work back satisfies the read at close. The procedure SHALL be read once per session and again after a compaction, not at each stage change. Initial affected-layout verification SHALL cover relevant desktop/mobile widths and supported themes; correction rounds SHALL check affected regions/neighbors, relevant states/widths and theme-sensitive behavior only. Navigation changes SHALL exercise their affected sequence. Publication consumers SHALL follow flow-walkthrough; review/capture consumers SHALL follow the single verify-owned conditional UI policy without a conflicting full-matrix iteration instruction. Human responses SHALL continue through native host questions in this delivery.

#### Scenario: Publication without per-write reads
- WHEN a producer starts a stage and later updates it in the same session
- THEN it uses the receipt returned by the first declaration for the update, without reading the session in between
- AND it reads the session back before telling the user the stage is visible.

### Requirement: Versioned event and immutable artifact publication
The local server SHALL accept typed, bounded protocol v3 events, images and Markdown documents under work-scoped authentication, assign durable IDs/versions/sequences and preserve immutable bytes. Native publication mutations SHALL carry an accepted producer and operation ID; host-observed events SHALL carry the accepted producer of their reporting session and an operation ID derived from the host event, as «Host-observed events» sets. Local viewer management mutations SHALL use the separate local_viewer actor and SHALL NOT impersonate a native producer. Exact replay SHALL return its durable receipt after restart; changed operation payloads, stale mutations and cross-work references SHALL fail without overwrite.

Server-side durable receipts SHALL preserve successful DELETE replay after restart, with replay preceding stale-version or lifecycle rejection for v3 operations. Image bytes and origin SHALL remain immutable; substantive evidence changes SHALL create an iteration linked to the prior view. Each validation payload SHALL retain its immutable evidence_id independently of copy edits, preserving instructions/images after restart.

Stage-associated events SHALL receive server-owned stage_event_id from a validated target_stage_event_id or the lead's foreground/latest preceding non-withdrawn stage. Non-lead publishers SHALL provide explicit stage targets. Clients SHALL NOT set the output field. Stages, closure and global membership/handoff events SHALL carry none. Protocol v1 events SHALL NOT be imported or reconstructed.

#### Scenario: Concurrent producers
- WHEN two publishers use the same expected version
- THEN one commits and one conflicts without overwriting; exact retry preserves receipt identity.

#### Scenario: Foreign stage
- WHEN a publisher targets a foreign or withdrawn stage
- THEN publication fails without content or state mutation.

### Requirement: Independent task progress and task detail
Tasks SHALL belong to the work with stable task_id, unique task_key, short title, plan reference when available and explicit required stage segments. The lead SHALL declare the full plan including future tasks without needing fake stage occurrences. Each segment SHALL have planned_stage, local state, accepted assignee, order and an explicit binding to a real matching stage occurrence when available. Stage panes SHALL display only their bound segments; future-stage tasks SHALL remain accessible in the plan and global count, not inside Build.

A task may span verify and delivery without becoming multiple tasks. Each stage projection SHALL show only its segment state; the task Sheet SHALL show aggregate state and the complete stage-labeled history. Global count SHALL deduplicate task_id and count completed only when all its required active segments are complete. Stage counts SHALL use only their assigned segments. No foreground change SHALL automatically move, clone or complete a task/segment. Assignment changes SHALL be explicit, versioned and audited.

Segment states SHALL be not_started, in_progress, in_review, blocked and completed. Aggregate task state SHALL be blocked if any active required segment is blocked; otherwise in_review if any is in review; completed if all are complete; not_started if all are not started; in_progress otherwise. State SHALL be displayed only in the row indicator with empty circle, loader, pause, alert or check and accessible text/tooltip. Indicators SHALL use semantic context colors alongside their distinct shapes and accessible labels; the icon stroke SHALL NOT receive a second decorative container border. Neither index, commits, checklist text nor process observation SHALL infer completion; a task state declared by the host's native task list and reported under «Host-observed events» is that session producer's declaration, not an inference. No heuristic percentage SHALL be displayed.

Each row SHALL expose available review content before expansion through compact evidence/document/wait/history affordances. Contentful rows SHALL provide independent disclosure and title actions: disclosure expands concise summary/optional relevant evidence; title opens the common-width task Sheet with Summary/Evidence/History and the selected segment in context. A task containing only a description and assignment/stage metadata, without linked evidence, documents, waits, blocking details or meaningful transition history, SHALL use inline disclosure rather than open a mostly empty Sheet. Refresh, state changes and return from Sheet SHALL preserve expansion/focus/scroll keyed by work/task/segment. Open task snapshots SHALL update only on explicit action and only a newer task_version SHALL trigger the notice. Any mutation changing visible task snapshot content, including linked evidence, document revisions and CI observations/attempt selection, SHALL atomically advance each affected task_version with the durable mutation/receipt. Unlinked uploads, unrelated tasks, no-ops and exact retries SHALL NOT advance it. Nested documents SHALL have a return path without stacked Sheets.

Blocked transitions SHALL require reason, concrete steps, owner and resume condition; incomplete payload SHALL fail unchanged. Blocking a segment SHALL NOT automatically block others or pause work. The lead or segment-assigned accepted producer SHALL perform conditional transitions and retain immutable actor audit and receipts. completed segments may reopen explicitly with a reason; terminal-stage re-entry SHALL use a linked new occurrence and explicit segment binding. Old evidence/history SHALL remain. Local viewer SHALL not approve/complete tasks. Withdrawal SHALL preserve history and exclude the withdrawn requirement from current counts.

Task content SHALL identify task_id/segment_id and its actual stage occurrence; references SHALL belong to that work, segment and origin stage. Activity from multiple stages SHALL stay attributable without duplicating task entities or blobs. Per-stage expansions SHALL show only relevant content, while the Sheet aggregates it. Decisions SHALL retain exact evidence/revision targets. Images SHALL be optional. Bounds SHALL fail atomically without preventing terminal cleanup.

#### Scenario: CAM task crossing stages
- WHEN Build tasks T1–T6 are complete and T7 has a completed verify segment and pending delivery segment
- THEN Build shows 6/6, Verify 1/1 and Delivery 0/1 when that stage exists, while the work remains 6/7 unique tasks.
- AND T7 does not appear within Build or create a fake future Delivery occurrence.

#### Scenario: Independent indicator states
- WHEN T1 and T3 segments are completed, T2 blocked, T4 in review and T5 not started
- THEN each indicator uses its explicit local state and the stage count is 2/5 regardless of index.

#### Scenario: Block and correct
- WHEN incomplete unblock instructions are submitted
- THEN transition fails unchanged.
- WHEN a completed segment reopens explicitly or a blocked one resolves
- THEN it retains IDs, evidence and immutable history across retry/restart.

### Requirement: On-demand high-level publication

The manager SHALL provide the existing additive hive view work semantic interface, reading explicitly bounded local sources, freezing selected Markdown and publishing through authenticated protocol v3 operations at matching timeline contract revision 3. It SHALL return compact confirmed results without callers authoring transport IDs, versions, hashes or request bodies. Existing low-level operation meanings, native payload identity and Markdown rendering SHALL be preserved after the coordinated update; operation against unmodified older server revisions SHALL not be promised. Publication SHALL NOT start/restart the viewer or open a browser.

#### Scenario: Publish and revise a plan
- WHEN a matching-revision caller publishes a valid plan folder and later its edited contents to the same work
- THEN two immutable revisions retain original bytes and the CLI confirms their revision IDs.

#### Scenario: Equal content and context
- WHEN current remote content and semantic publication context match the requested set
- THEN the command returns unchanged without remote mutation.

### Requirement: Durable client publication recovery
The manager SHALL persist versioned, permission-restricted publisher bookkeeping outside the server-owned viewer inventory. Every attempted mutating request SHALL have its exact bytes and operation identity durably recorded before transmission. A semantic declaration queued while the viewer is unavailable SHALL have its semantic input and operation identity durably recorded before the command returns, and the requests built when it is sent SHALL follow the same rule. Uncertain outcomes SHALL replay unchanged requests before accepting replacement work. Successful publication SHALL require exact revision/document readback; partial uploads SHALL not be described as a published revision. Unconfirmed inputs, queued declarations included, SHALL survive process exit and SHALL not expire by age. The local management commands (`hive view work rename`, `archive`, `restore`, `pin`, `unpin`, `recover` and `delete`) are exempt from this bookkeeping: after an uncertain outcome they SHALL resend the identical request once, which the viewer answers from its stored receipt, and only when that is uncertain too re-read the work and report whether the action holds; their expected version or a deleted work SHALL keep a retry from applying twice.

#### Scenario: Response lost after server commit
- WHEN the CLI exits after a committed snapshot but before recording its acknowledgment
- THEN resume replays the original operation and confirms the same revision without a duplicate.

#### Scenario: Competing publisher or external change
- WHEN a second publisher holds the client lock, or another actor changes the selected stage/chain incompatibly
- THEN the operation reports busy/conflict without silently overwriting, closing or reopening work.

#### Scenario: Recover from a definitive rejection
- WHEN the server has definitively rejected the exact pending request, every earlier attempted outcome is resolved, and an authorized caller explicitly reconciles that pending intent with a reason
- THEN the client atomically archives its bytes/receipts/rejection evidence locally, leaves remote content untouched and permits a new explicit publication against the current chain.
- WHEN any attempted outcome remains unknown
- THEN archival is refused and exact replay remains the recovery path.
- WHEN the work was closed or deleted
- THEN reconciliation never reopens or recreates it.

### Requirement: Bounded publication interface and compatibility

Successful mutation responses SHALL fit 4096 UTF-8 bytes without full source documents or credentials. The work list SHALL retain bounded pages and honest missing-binding state. High- and low-level clients SHALL share protocol v3 operation meanings at matching timeline contract revision 3; a coordinated update SHALL replace the prior requirement to operate against an unmodified older server. Publisher files SHALL remain outside viewer storage. The compact contract and canonical procedure SHALL preserve publication timing, producer authority, human validation and evidence checks while the binary owns bookkeeping. Native pending-operation identities/bytes and retained publication receipts SHALL survive the update for explicit resume.

#### Scenario: Coordinated server and clients
- WHEN high- and low-level clients use the matching upgraded server
- THEN they operate on preserved work histories, keep publisher journals outside viewer storage and retain explicit interfaces for unsupported advanced operations.

### Requirement: Viewer-first walkthrough publication
Substantive publication SHALL use only the compatible v3 API and confirmed readback. A missing executable or an incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon or producing any other format; a stopped viewer SHALL do the same, except that semantic open, stage and evidence declarations SHALL be queued as «Queued semantic publication while the viewer is unavailable» sets. Previously generated HTML pages and user-selected artifact units SHALL not be deleted or rewritten by the removal of the renderer. Research SHALL NOT publish to the viewer; a work's viewer session SHALL start at its first plan or build flow that was explicitly invoked (its command typed, its route accepted, or research continuing into build at the user's request), created with a summary of about one paragraph on what the work is about. The agent SHALL NOT open the viewer on its own: at each stop where the viewer holds something for the user to review (a plan at its approval stop, a UI walkthrough, including the build's first rendered view) it SHALL offer to open it inside what that stop already asks (one more question of the stop's round, or a reply line of its waiting line or next-step block), never as a question with its own decision ID or its own question-tool call, until the user accepts once in the conversation; after acceptance it SHALL open the work once and not ask again unless the user asks.

#### Scenario: Binary present, server incompatible
- WHEN publication finds a v1 server
- THEN it requests the coordinated v3 update and preserves inputs without HTML substitution.

#### Scenario: User declines
- WHEN the user declines to open the viewer at a review stop
- THEN the viewer stays closed and the offer is repeated at the next review stop.

#### Scenario: Research report
- WHEN a research report ends, with or without a route to plan or build
- THEN nothing is published to the viewer and no open-viewer offer is made
- AND the first plan or build flow of that work creates its viewer session with a one-paragraph summary.

### Requirement: Coordinated timeline contract revision

The upgraded server, CLI and UI SHALL require timeline contract revision 3 while preserving schema/storage version 3 identities and all existing evidence. Matching revision SHALL be acknowledged before upgraded clients write; old or missing revisions SHALL receive a safe actionable update/reload error before timeline mutations. No simultaneous old-client projection SHALL be required. Bootstrap/authentication and existing file-capability reads SHALL retain their access protections.

The shared session/index projection SHALL expose required last_activity_at, nullable closed_at and autoarchive_at, and nullable archive_origin limited to manual/automatic. Stored-read normalization SHALL accept valid pre-policy records without weakening strict public validation. Contract mismatch SHALL preserve publisher journal inputs, operation identity and stored receipts; it SHALL NOT trigger automatic replay, request replacement or deletion.

#### Scenario: Old tab or CLI
- WHEN an outdated client attempts a timeline mutation
- THEN the server requests update/reload before writing, and the pending operation/input remains available for explicit recovery with the upgraded client.

#### Scenario: Upgraded client against old server
- WHEN the server does not acknowledge the new contract revision
- THEN the upgraded client refuses mutation before sending a write.
