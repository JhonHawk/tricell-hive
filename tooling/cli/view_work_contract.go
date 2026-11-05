package main

import (
	"encoding/json"
	"fmt"
	"io"
)

const viewWorkContract = `hive view work — semantic publisher over protocol v3

Identity: --host HOST --native-session-id ID are optional. Without them the command uses the host's session variable, or the process-to-session map its session-start hook wrote; no source, or an ambiguous one, is rejected (pass both flags then). Without --work-id, a command acts on the work this session opened and has not closed; with several open, --work-id is required. Hosts: claude-code, codex, grok, pi, opencode, cursor; an existing work needs accepted membership. No command starts/restarts the viewer or opens a browser. --home and --state-dir select existing manager state.

A build without a plan takes three commands, each one reply of at most 4096 bytes:
open --key KEY --title TEXT --summary TEXT [--project ROOT]
sync --stage STAGE --state STATE --reason TEXT [--tasks FILE|-] [--summary TEXT] [--previous-stage OCCURRENCE] [--close completed|abandoned]
evidence --stage STAGE --view-title TEXT --image FILE[:caption] ... [--candidate COMMIT [--branch B --base B]] [--supersedes VIEW] [--step TEXT ...] [--expected TEXT] [--viewport desktop|mobile] [--theme light|dark|unspecified] [--reason TEXT]
close --outcome completed|abandoned --reason TEXT

open without --change makes a work with no change folder and no plan revision; --project defaults to the repository you run in. sync is one declaration per stage: it starts or updates the stage (started first when declared in another state than in_progress), registers the missing tasks bound to that stage (only without a change folder or published plan; else use tasks register) and moves each task to its state through the allowed transitions. --tasks is a JSON array [{"key":"T1","title":"Parse input","state":"in_progress"}] with optional summary, block {reason,steps,owner,resume_when} and resolution (leaving blocked); tasks it omits are left alone. Repeating a declaration answers unchanged. --close completes the active stage, closes the work and reads it back; close does the same. With --state completed or failed the task states are applied first. With the viewer stopped or silent, open, sync and evidence reply queued (work_id is a local handle) and never start it; the next command sends the queue in order first and lists what no longer applies under queue.rejected. The reply carries stage_id, state and per task task_id, segment_id and state (tasks_omitted counts those that do not fit). evidence uploads the captures, registers the candidate when its commit is new (else the latest) and appends a validation event with one view; --supersedes VIEW makes it an iteration replacing that view; repeating it changes nothing. Its reply carries candidate_id, event_id, evidence_id, view_ids and artifact_ids.

Advanced commands (tasks register and close behave as above):
open --project ROOT --change DIR --key KEY --title TEXT --summary TEXT | open --work-id ID --change DIR
publish --work-id ID --change DIR --stage STAGE [--document-key KEY] [--status draft|complete]
publish --project ROOT --change DIR --key KEY --title TEXT --summary TEXT --stage STAGE --reason TEXT [publication options]
list --project ROOT [--host HOST --native-session-id ID] [--include-archived] [--limit 1..100] [--cursor OPAQUE]
show --work-id ID
stage --work-id ID --stage STAGE --state STATE --reason TEXT [--summary TEXT] [--previous-stage OCCURRENCE]
tasks register --work-id ID [--plan-revision CONFIRMED_REVISION] --json FILE|- [--reason TEXT] (revision optional without a change folder or plan)
task update --work-id ID --key KEY --stage STAGE --state STATE --reason TEXT [transition options]
task update --work-id ID --key KEY --reason TEXT --json FILE|-
close --work-id ID --outcome completed|abandoned --reason TEXT (completes the active stage first)
resume --work-id ID
reconcile --work-id ID --pending ID --action archive-rejected --reason TEXT
bind --work-id ID --change DIR --reason TEXT
rename --work-id ID --title TEXT
archive --work-id ID
restore --work-id ID
pin --work-id ID
unpin --work-id ID
recover --work-id ID --request-id ID --reason TEXT
delete --work-id ID --confirm ID
contract

Creation fields and --work-id are exclusive. Change paths are bound to explicit keys, never cwd/title/content-derived identity; bind replaces a directory within the same project after pending work resolves. show/list are reads and never create/reopen work; list defaults to 20 items and returns an opaque next cursor.

STAGE: research|plan|build|verify|delivery. Stage STATE: in_progress|in_review|awaiting_validation|completed|paused|failed. A terminal stage needs explicit --previous-stage for reentry. Existing publication never starts a stage. The creation variant explicitly starts the requested stage with --summary/--reason. Snapshot status is draft|complete, not approval.

Plan defaults: proposal.md and tasks.md required, design.md optional, sorted specs/**/spec.md. --document RELATIVE.md adds extras; --explicit-documents replaces defaults and needs at least one. Non-plan stages require explicit --document and --document-key. The plan key defaults to plan. At most 8 tabs, 256 KiB per Markdown, no truncation, no linked-file traversal; limits are checked before mutation and bytes freeze before it.

Registration JSON, a nonempty array (max 100):
[{"key":"T1","title":"Implement export","summary":"Add the requested export","order":1,"segments":[{"stage":"build","order":1},{"stage":"verify","order":2}]}]
Optional segment assignee: {"host":"codex","native_session_id":"verified-native-session"}; it must resolve to an accepted producer. --plan-revision (a confirmed published plan) is required only for a work with a change folder or a published plan. Batches validate before writes; identical registration is skipped without resetting progress. Unknown/duplicate fields, duplicate keys/stages and technical IDs/versions are rejected.

Task keys use T followed by a positive integer from 1 (for example T1; T0 is rejected). A segment order is unique per stage across all tasks of the batch, not within each task, and a task holds at most one segment per stage. A rejected --json names the field and the rule, never a value: tasks[<i>].<field> for the list of tasks register and metadata.<field> for task update --json, with segments[<j>] inside a task (for example tasks[1].key: duplicate task key). Task STATE: not_started|in_progress|in_review|blocked|completed, with existing v3 transition rules. blocked requires --block-reason TEXT --block-owner TEXT --block-step TEXT (repeatable, 1..10) --resume-when TEXT. Leaving blocked requires --resolution TEXT. --review-kind human|agent is valid only for in_review. Metadata --json is exclusive with transition fields; it accepts title, summary, order, stage, segment_order, assignee, withdrawn, plan_revision, decision_reference, and needs a confirmed plan_revision or decision_reference when v3 requires it. No status inference, approval, membership creation or CI synthesis.

close --outcome completed maps to v3 closure reason completed, and abandoned maps to abandoned; --reason becomes the honest closure result/reason; authority checks still apply. Use in_review rather than claiming completion. resume recovers technical pending requests and never reopens work. reconcile archives only a definitively rejected pending intent after all unknown outcomes resolve, keeping its evidence locally.

Management commands (rename, archive, restore, pin, unpin, recover, delete) take neither flag and act as the local_viewer actor with origin cli (declared provenance, not authentication); who may run them: the viewer publication guidance «Shared work and authority». They need a viewer that announces local_actions_cli in /api/health, else viewer_update_required. recover accepts the pending same-project request named by --request-id (the new session asks first with hive view producer request) and makes it lead; it never infers that the previous lead ended nor reopens a closed work. delete sends nothing unless --confirm equals --work-id; work_gone and delete_pending are errors with a non-zero exit. They keep no journal: after an uncertain response they resend once with the same operation ID, then re-read the work and report whether the action holds, or not_applied / outcome_unknown.

Success mutation JSON <=4096 UTF-8 bytes: work_id, result, session_version, relevant revision_id/task_id/stage_id; publish adds document_count/changed_count; close adds state. Success requires confirmed readback; unchanged means no remote mutation. Errors carry a stable code, safe field/path, pending ID and next action, never credentials or document bodies. Journal: <state>/publisher/v1/, 0700 directories/0600 files. Unknown outcomes keep exact bytes and operation identity; resume them unchanged. Unavailable runtime: upgrade Hive; never substitute an agent-generated helper. Low-level hive view contract remains available.
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
