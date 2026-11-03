# Design

## Approaches

Selected: a high-level CLI publisher over existing v3 endpoints (D1 A). It owns bounded local reads, frozen input, durable request bookkeeping, semantic resolution and readback. Existing low-level commands remain available and unchanged.

Alternative: send a bundle to a new server-side batch endpoint. That could make a wider publication transaction server-owned, but changes the HTTP contract and requires client/server rollout. It is not selected. No persistent watcher is needed: an explicit command is the publication trigger.

## Compatibility and UI

Add `hive view work`; do not repurpose existing commands or require a protocol/storage migration. Keep the current Markdown renderer and document Sheet/tabs (D2 A). No UI source or generated bundle changes are planned. Source bytes stay original; relative links are not rewritten, HTML is skipped and images remain unsupported by that renderer. The proposal's earlier relative-link navigation idea is withdrawn by D2 A.

Client-owned directory reads produce bounded bytes for existing HTTP uploads. The server never accepts paths to scan. No automatic service start/restart, browser launch, tracker query, CI call or authority transfer is added.

## Command contract

One namespace, with strict parsing and a compact offline `hive view work contract`. Names below define the planned interface, not currently installed commands. All mutations require explicit verified `--host` and `--native-session-id`; these select the real producer, never the stored lead as a substitute. They are provenance, not proof of process liveness. `--project` is required at first binding; later the binding supplies that root. Do not invent host discovery env vars/adapters.

| Command | Semantic inputs | Result |
| --- | --- | --- |
| `work open` | `--project ROOT --change DIR --key KEY --title TEXT --summary TEXT`, or `--work-id ID --change DIR` for an existing accepted work | Resolve/create work and source binding; compact work ID/state. No documents yet. |
| `work publish` | `--work-id ID --change DIR --stage STAGE`; or creation fields of `open` when first publishing. `--document-key KEY` defaults to `plan` for plan; explicit otherwise. `--status draft\|complete` is snapshot status, not approval. | Confirm revision or `unchanged`. Stage must already exist; first-publication creation variant explicitly starts the requested stage with supplied summary/reason. |
| `work list` | `--project ROOT`, optional host/native-session pair, optional `--include-archived` | JSON items with work ID/key, title, state, observed version/revision and known source binding. |
| `work show` | `--work-id ID` | Compact observed state, bound folder and last/pending publication; no full document bodies. |
| `work stage` | `--work-id ID --stage STAGE --state STATE --reason TEXT`, optional summary; explicit prior occurrence when reentering a terminal stage | Start/update declared occurrence and retain related-event chain; never infer from filenames. |
| `work tasks register` | `--work-id ID --json FILE|-` | Validate all semantic task entries, register missing tasks, skip identical ones; changed metadata is a conflict requiring explicit update. |
| `work task update` | `--work-id ID --key T1 --stage build --state STATE --reason TEXT`, with block/resolution/review data when required | Resolve task/segment/producer, bind display occurrence if needed, apply transition; return compact state. Explicit metadata changes use `--json` with an exclusive metadata mode. |
| `work close` | `--work-id ID --outcome completed\|abandoned --reason TEXT` | Valid explicit lifecycle close; open-PR/human-review work uses stage `in_review` instead. |
| `work resume` | `--work-id ID` plus matching native source | Resume only pending technical publication, never reopen closed work or change semantic lifecycle. |
| `work reconcile` | `--work-id ID --pending ID --action archive-rejected --reason TEXT` plus matching native source | Archive a definitively rejected local intent after checking all attempted outcomes; preserve bytes/receipts and permit a new explicit publication. Never discard an unknown outcome or mutate the remote chain. |
| `work bind` | `--work-id ID --change DIR --reason TEXT` plus matching/accepted source | Explicitly replace a source-directory binding within the same project after pending requests are resolved; no implicit copy/rename of files. |

Use the v3 state/authority enums and required block/closure structures; do not create another lifecycle vocabulary. Contract examples must provide complete actionable fields (including actual outcome mappings). The client must not synthesize authorization/approval, timestamps, passing tests or CI evidence. Advanced artifacts/views/CI and membership acceptance keep their existing interfaces; the new client may use only already accepted membership.

Successful mutation output is bounded at 4096 UTF-8 bytes and includes `work_id`, `result`, confirmed session version and relevant revision/task/stage identifiers. `publish` adds document/changed counts. The list is paginated locally from the bounded existing index, default 20 items with an opaque continuation cursor, so the full server snapshot is never emitted as a model response. Missing local binding is `null`/unknown, never a guessed folder. Errors include a stable code, affected local key/path when safe, pending operation identifier and a next action. Credentials/access fragments/document content never appear in output.

## Identity and binding

Reuse the existing opaque work ID: project identity + host + native session ID + explicit stable work key. Key does not depend on title, content hash or current working directory. The first CLI call requires a key; no random-key fallback for this high-level path. Preserve exact creation request and operation ID before sending; an uncertain first result is recovered by exact replay, not another creation operation. Ordinary updates GET the recorded work rather than call create with fresh operation IDs.

Canonicalize an explicitly provided root/change path once; `pwd` only resolves relative input. Bind work ID, canonical project root, canonical change directory and producer source in local state. Within one native source, a directory cannot silently select another key/work; `open` with a conflicting key or `publish` with a conflicting directory fails. A different change directory plus key produces a different work in the same native session. Native session end and work completion remain independent.

An existing `--work-id` can be used by another session only after existing v3 join/accept has established membership; resolve that session's producer rather than impersonating the original. Per-work source bindings may exist for accepted producers in other worktrees. Rebinding requires `work bind`, matching project identity, explicit reason, no unresolved pending transaction and a verified target directory. Preserve prior paths as local history, not an automatic alias fallback. Closed/deleted works remain closed/deleted; neither binding, list, GET nor publication reopens them.

## Source selection and snapshots

Default plan selection: required `proposal.md` and `tasks.md`, optional `design.md`, then existing `specs/**/spec.md` sorted by normalized relative path. Order: proposal, design if present, tasks, specs. Unique labels are `Proposal`, `Design`, `Tasks`, and each spec's relative path. `--document RELATIVE_PATH` may add explicitly selected Markdown extras; reject duplicate resolved paths or labels. Other stages use an explicit document selection and key; do not require a change-folder schema for an already retained research/verification document. Plans kept in conversation can use that explicit-file mode with a bounded source root; do not force a new change folder merely for display.

Enumerate only the selected patterns; enforce directory traversal/entry budgets and the existing document/tab/aggregate limits before any mutation. No entire-repository traversal, dependency directories, implicit linked-file reads or silent truncation. More than eight selected documents fails with the count/limit and selection guidance. An explicit document selection mode replaces default discovery only when the user/agent deliberately requests it; do not silently drop specs to fit.

Reject path escapes, symlink components below the canonical source root, nonregular files, shared-inode hardlinks and changed file identity. Use descriptor-based validation and bounded reads. Reject growth/shrink or metadata changes during collection, then validate the collected set again before freezing; recommend an explicit retry if the source changed. A file changed after freeze does not mutate the pending request. The snapshot represents the frozen set, not a claim of a filesystem-wide transaction. Freeze every selected file and its relative path, label, SHA-256, size and request context before network mutation. An omitted optional file is removed only from the next snapshot, never from historical documents. Missing required/unreadable files fail preflight.

No-op comparison uses the current remote chain tip, ordered labels/byte hashes, document key, status, stage occurrence and any task/segment context. A local fingerprint alone is insufficient because another actor can write. Same bytes in a new occurrence/status/context may require a new snapshot. A semantic no-op performs reads but no remote mutation (not even a new create/upload receipt). Upload only missing content; already stored immutable documents may be referenced when the server confirms their bytes/provenance are suitable.

## Client journal and recovery

Use `<state>/publisher/v1/`, a sibling of `viewer/`, directory mode 0700 and journal/frozen files mode 0600. No token is persisted here; use the existing authenticated transport. Keep source plans in the repository. Journal schema is versioned; malformed/unknown versions fail without deletion. Use an independent nonblocking publisher lock, released by process exit; do not hold the manager/viewer server locks over HTTP. A single lock per publisher state directory is sufficient for v1 and reports `busy` rather than queueing model-visible polling.

State machine:

1. `prepared`: all selected bytes, source identity and semantic intent durably stored.
2. `resolving`: exact create request persisted if needed; otherwise GET existing work/producer.
3. `uploading`: each exact upload request/op ID persisted before send, each receipt saved after confirmation.
4. `snapshot_pending`: current chain/version read, exact snapshot request persisted before send.
5. `verifying`: confirm remote work, revision/tabs and downloaded document sizes/hashes against frozen inputs.
6. `confirmed`: persist receipt/fingerprint and compact result; task-owned frozen scratch may be reclaimed only after readback.
7. `needs_reconciliation`: preserve inputs and request bytes on uncertain errors; resume technical pending work before accepting a new operation for that work.
8. `rejected_conflict`: the exact attempted request received a definitive precondition rejection from the server; preserve its request, bytes and rejection evidence. It blocks replacement until explicitly reconciled.
9. `archived_rejected`: explicit reconciliation preserves the entire rejected intent outside the active pending slot and frees the work for a new explicit operation. This is local bookkeeping, not a successful publication or remote rollback.

Atomic journal writes use temp + file sync + rename + directory sync with safe file checks; do not introduce a runtime/dependency for persistence. Keep receipt metadata sufficient for exact replay until confirmed. Retain unconfirmed bytes without age-based cleanup. Never put new files under the server-owned v3 inventory.

An unknown outcome or lost response replays the unchanged operation ID/body. GET does not expose receipts, so matching final visible state alone cannot establish that a specific uncertain write committed. A definite version/predecessor rejection can prepare a new request/op ID only after fresh reads establish that authority, target stage and chain are still compatible. If another actor published a different chain tip or closed the work, stop with an actionable conflict; do not silently supersede/reopen. No unlimited retries: bounded transient retries inside one command, then pending error with explicit resume.

To escape a terminal conflict, `work reconcile` requires that exact pending ID, explicit reason and the original producer or already accepted authority. Under the publisher lock it validates the journal, resolves every unknown attempted outcome by exact replay and confirms that the failed step has a definitive server rejection (for example `version_conflict` or `current_predecessor_required`). A timeout, malformed response, `outcome_unknown`, unexplained `operation_conflict` or unresolved earlier step is not proof of rejection and refuses archival. The command atomically moves the known-rejected intent to retained local history, with its frozen bytes, all prior successful receipts and rejection evidence; it performs no replacement snapshot or blob deletion. A crash during archival must recover to exactly one active-or-archived classification. A subsequent `publish` is a new explicit intent based on the current remote chain and source bytes. If the work is closed/deleted, that new publication still fails under lifecycle rules. Test a predecessor conflict followed by reconciliation and a valid new publication in the same still-active work, plus refusal to archive uncertain requests.

Server operations are individually atomic; the full multi-request publication is not. Partial uploads can remain stored if the final snapshot fails. Never roll back by deleting server blobs, and never label the group published before the final snapshot and readback. Restarting the server while the client waits preserves this contract. Work deletion/tombstones override pending publication; report unavailable and preserve local evidence rather than resurrecting.

## Tasks and lifecycle

The task registration input is a compact JSON array with `key`, `title`, `summary`, `order` and `segments` (`stage`, `order`, optional accepted assignee). It references the selected confirmed plan revision explicitly; do not link to an unconfirmed or merely latest local edit. Runtime operation IDs, server IDs and expected versions are never agent-authored fields. Validate the entire batch and semantic duplicates before creating any task; preserve per-task receipts to resume a partially registered batch. Do not derive task status from Markdown checkboxes or parse prose for ownership.

Resolve task key and stage segment from current work state; preserve lead/assigned-producer rules. Register once, skip identical registration, require explicit metadata updates/withdrawal for changed scope. Batch registration must not reset progress. Metadata updates accept only existing v3 mutable fields expressed through semantic keys and a plan/decision reference where required. Starting/completing a segment must not approve a plan, complete another segment or close a work. `blocked` retains required reason, responsible party, unblock steps and resolution criterion; review/completion preserve the original v3 validation requirements.

Stage reentry uses an explicit new occurrence linked to its predecessor. Publication uses the selected occurrence and never manufactures stage transitions except the explicit first-publish creation path. User validation and CI waits stay distinct. `work close` follows the existing honest outcome and active-stage preconditions; `resume` is recovery, not lifecycle reopening.

## Impact and instruction ownership

Implementation homes: `tooling/cli/view_work*.go` for parsing/adapters; a focused CLI-owned publisher module or package for journal/scan/sequence logic, chosen to avoid exporting unrelated viewer internals. Reuse the existing transport, index reader and typed v3 validators. Small shared helper extraction is allowed only when these consumers need it; no separate daemon, universal adapter or new dependency.

Update canonical publication procedure and compact contract together. Preserve identity, authorization, snapshot timing, human approval, evidence provenance and cleanup intent while moving technical steps into the binary. Review every caller across global guidance, flow skills/references, agent roles, site pages, architecture docs and pilot detectors. Global guidance changes are not expected; measure its before/after byte count and report the delta in delivery. Do not change the separate research-trigger decision or adopt unrelated edits from the current branch. Do not add fixture regression cases or run model pilots. This is an explicitly requested capability; it is not justified solely as an unapproved new behavioral rule.

README, llms.txt, site viewer/affected flow pages and deployment-manager docs must describe the new commands and preserved limits in the same delivery. Maintain old low-level contract for advanced consumers; missing new-client capability is reported with an upgrade action, not an agent-generated substitute helper.

## Recovery and rollout

Build from the then-current development base after reconciling concurrent work, keeping this change isolated. Delivery is automatic (D3 A), with native Codex `/review` before push (D4 A), once the user starts build. Do not publish the unrelated current branch commits as part of this work. No server migration or viewer frontend rebuild is required by this design. A maintainer refresh, if later authorized by the repository marker/delivery rules, rebuilds binary before deploying content; preserve the existing viewer service unless the integrated change meets the repository's restart rule. Site documentation changes still require their own check/lint/build and local preview.

Rollback restores the prior binary/content while preserving `<state>/publisher/v1/` for forward recovery. Existing server content remains valid v3; no automatic journal or document deletion. A downgraded client cannot resume high-level pending operations, so report that limitation rather than silently discarding them.
