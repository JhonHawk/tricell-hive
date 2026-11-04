package main

import (
	"encoding/json"
	"fmt"
	"io"
)

const viewWorkContract = `hive view work — semantic publisher over protocol v3

All mutations require --host HOST --native-session-id ID, verified from the native session. Accepted hosts: claude-code, codex, grok, pi, opencode, cursor. These identify the actual producer; accepted membership is required for an existing work. No command starts/restarts the viewer or opens a browser. --home and --state-dir select existing manager state.

open --project ROOT --change DIR --key KEY --title TEXT --summary TEXT
open --work-id ID --change DIR
publish --work-id ID --change DIR --stage STAGE [--document-key KEY] [--status draft|complete]
publish --project ROOT --change DIR --key KEY --title TEXT --summary TEXT --stage STAGE --reason TEXT [publication options]
list --project ROOT [--host HOST --native-session-id ID] [--include-archived] [--limit 1..100] [--cursor OPAQUE]
show --work-id ID
stage --work-id ID --stage STAGE --state STATE --reason TEXT [--summary TEXT] [--previous-stage OCCURRENCE]
tasks register --work-id ID --plan-revision CONFIRMED_REVISION --json FILE|- [--reason TEXT]
task update --work-id ID --key KEY --stage STAGE --state STATE --reason TEXT [transition options]
task update --work-id ID --key KEY --reason TEXT --json FILE|-
close --work-id ID --outcome completed|abandoned --reason TEXT
resume --work-id ID
reconcile --work-id ID --pending ID --action archive-rejected --reason TEXT
bind --work-id ID --change DIR --reason TEXT
contract

Creation fields and --work-id are exclusive. --project is mandatory on first binding; later roots come from the binding. Canonical change paths are bound to explicit keys, never cwd/title/content-derived identity. bind explicitly replaces a directory within the same project after pending work resolves. show/list are reads and never create/reopen work. list defaults to 20 items, reports unknown bindings as null, and returns an opaque next cursor.

STAGE: research|plan|build|verify|delivery. Stage STATE: in_progress|in_review|awaiting_validation|completed|paused|failed. A terminal stage needs explicit --previous-stage for reentry. Existing publication never starts a stage. The creation variant explicitly starts the requested stage with --summary/--reason. Snapshot status is draft|complete, not approval.

Plan defaults: proposal.md and tasks.md required, design.md optional, sorted specs/**/spec.md. --document RELATIVE.md adds extras. --explicit-documents replaces defaults and requires at least one --document. Non-plan stages require explicit --document and --document-key. The plan key defaults to plan. Maximum 8 tabs, 256 KiB per UTF-8 Markdown; no truncation. Paths/links/types/stability and observable document limits are checked before mutation. The server enforces remaining v3 quotas; partial uploads never count as a published snapshot. No linked-file traversal. Bytes freeze before network mutation.

Registration JSON is a nonempty array (max 100):
[{"key":"T1","title":"Implement export","summary":"Add the requested export","order":1,"segments":[{"stage":"build","order":1},{"stage":"verify","order":2}]}]
Optional segment assignee: {"host":"codex","native_session_id":"verified-native-session"}; it must resolve to an accepted producer. --plan-revision references a confirmed published plan, never an unconfirmed edit. Entire batches validate before writes; identical registration is skipped without resetting progress. Unknown/duplicate JSON fields, duplicate task keys/stages and technical IDs/versions are rejected.

Task keys use T followed by a positive integer from 1 (for example T1; T0 is rejected). A segment order is unique per stage across all tasks of the batch, not within each task, and a task holds at most one segment per stage. A rejected --json names the field and the rule, never a value: tasks[<i>].<field> for the list of tasks register and metadata.<field> for task update --json, with segments[<j>] inside a task (for example tasks[1].key: duplicate task key). Task STATE: not_started|in_progress|in_review|blocked|completed, with existing v3 transition rules. blocked requires --block-reason TEXT --block-owner TEXT --block-step TEXT (repeatable, 1..10) --resume-when TEXT. Leaving blocked requires --resolution TEXT. Optional --review-kind human|agent is valid only for in_review. Metadata --json is exclusive with transition fields; accepts title, summary, order, stage, segment_order, assignee, withdrawn, plan_revision, decision_reference. Segment edits require stage; metadata needs a confirmed plan_revision or decision_reference when v3 requires it. No Markdown status inference, approval, membership creation or CI synthesis.

close --outcome completed maps to v3 closure reason completed, and abandoned maps to abandoned; --reason becomes the honest closure result/reason. Active-stage/authority checks still apply. Review/CI waits are distinct; use in_review rather than claiming completion. resume recovers technical pending requests and never reopens work. reconcile archives only a definitively rejected exact pending intent after all unknown outcomes resolve; it retains bytes/receipts/rejection evidence locally and never overwrites remote content.

Success mutation JSON <=4096 UTF-8 bytes: work_id, result, session_version, relevant revision_id/task_id/stage_id; publish adds document_count/changed_count. Success requires confirmed readback. unchanged means no remote mutation. Errors carry a stable code, safe field/path, pending ID when present and next action. No credentials or document bodies in stdout/errors. Journal: <state>/publisher/v1/, outside viewer inventory, 0700 directories/0600 files. Unknown outcomes retain exact bytes and operation identity; resume unchanged, never replace pending intent. Unavailable runtime/client: upgrade Hive; never substitute an agent-generated helper. Low-level hive view contract and its advanced operations remain available.
`

func writeViewWorkBoundedJSON(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("invalid_result: cannot encode result")
	}
	if len(data)+1 > 4096 {
		return fmt.Errorf("invalid_result: mutation output exceeds 4096 bytes")
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}
