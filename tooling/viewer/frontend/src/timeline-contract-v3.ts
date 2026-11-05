import {
  timelineCandidateByID,
  validateTimelineDeliveryReference,
} from "./timeline-catalog";
import type {
  TimelineAppendEventRequest,
  TimelineCandidateRecord,
  TimelineDecisionTarget,
  TimelineDeliveryReference,
  TimelineDocumentSnapshotRequest,
  TimelineSession,
  TimelineSessionPatchRequest,
  TimelineSource,
  TimelineStage,
} from "./timeline-contract";
import {
  TimelineContractError,
  validateTimelineDecisionReference,
  validateTimelineSource,
} from "./timeline-contract";
export type TimelineActor =
  | { kind: "native_producer"; producer_id: string }
  | {
      kind: "local_viewer";
      producer_id?: never;
      origin?: TimelineLocalOrigin;
    };
export type TimelineLocalOrigin = "browser" | "cli";
export type TimelineProducer = {
  producer_id: string;
  source: TimelineSource;
  joined_at: string;
};
export type TimelineJoinRequest = {
  request_id: string;
  source: TimelineSource;
  requested_at: string;
  reason: string;
};
type TimelineAuditFields = {
  block?: TimelineBlock;
  resolution?: string;
  previous_segment?: TimelineTaskSegment;
  next_segment?: TimelineTaskSegment;
  plan_revision_id?: string;
  operation: string;
  reason: string;
  previous_state?: string;
  next_state?: string;
  previous_producer_id?: string;
  next_producer_id?: string;
  request_id?: string;
  target?: TimelineDecisionTarget;
};
type AuditCommon = "operation" | "reason" | "target";
type AuditBranch<
  O extends string,
  R extends keyof TimelineAuditFields,
  Optional extends keyof TimelineAuditFields = never,
> = {
  operation: O;
  reason: string;
  target?: TimelineDecisionTarget;
} & Required<Pick<TimelineAuditFields, R>> &
  Pick<TimelineAuditFields, Optional> &
  Partial<
    Record<
      Exclude<keyof TimelineAuditFields, AuditCommon | R | Optional>,
      never
    >
  >;
export type TimelineAudit =
  | AuditBranch<
      "stage_start" | "stage_update" | "session_patch" | "closure",
      "previous_state" | "next_state"
    >
  | AuditBranch<
      "producer_accept" | "recover_lead",
      "request_id" | "next_producer_id"
    >
  | AuditBranch<
      "foreground_transfer",
      "previous_producer_id" | "next_producer_id"
    >
  | AuditBranch<
      "recover_lead",
      "request_id" | "previous_producer_id" | "next_producer_id"
    >
  | AuditBranch<
      "task_transition",
      "previous_state" | "next_state",
      | "block"
      | "resolution"
      | "previous_segment"
      | "next_segment"
      | "plan_revision_id"
    >
  | AuditBranch<
      "session_rename" | "session_archive" | "session_pin" | "session_delete",
      never
    >;
export type TimelineBlock = {
  reason: string;
  steps: Array<string>;
  owner: string;
  resume_when: string;
};
export type TimelineTaskSegment = {
  segment_id: string;
  planned_stage: TimelineStage;
  order: number;
  state: TimelineTaskState;
  assigned_producer_id: string;
  display_stage_event_id?: string;
  block?: TimelineBlock;
  review_kind?: string;
  withdrawn?: boolean;
};
export type TimelineTask = {
  plan_order: number;
  task_id: string;
  task_key: string;
  title: string;
  summary: string;
  task_version: number;
  plan_revision_id?: string;
  segments: Array<TimelineTaskSegment>;
  state: TimelineTaskState;
  withdrawn?: boolean;
};
export type TimelineTaskSegmentInput = {
  planned_stage: TimelineStage;
  order: number;
  assigned_producer_id?: string;
};
export type TimelineNativeOperation = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
};
export type TimelineJoinRequestInput = {
  schema_version: 3;
  operation_id: string;
  source: TimelineSource;
  expected_session_version: number;
  reason: string;
};
export type TimelineProducerAcceptRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  request_id: string;
};
export type TimelineForegroundTransferRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  target_producer_id: string;
};
export type TimelineRecoverLeadRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  request_id: string;
};
export type TimelineTaskCreateRequest = {
  plan_order: number;
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  task_key: string;
  title: string;
  summary: string;
  plan_revision_id?: string;
  segments: Array<TimelineTaskSegmentInput>;
};
export type TimelineTaskPatchRequest = {
  plan_order?: number;
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  expected_task_version: number;
  title?: string;
  summary?: string;
  plan_revision_id?: string;
  decision_target?: TimelineDecisionTarget;
  segment_id?: string;
  order?: number;
  assigned_producer_id?: string;
  display_stage_event_id?: string;
  withdrawn?: boolean;
};
export type TimelineTaskTransitionRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  expected_task_version: number;
  segment_id: string;
  state: TimelineTaskState;
  block?: TimelineBlock;
  review_kind?: string;
  resolution?: string;
};
export type TimelineRequiredCheck = {
  check_key: string;
  label: string;
};
export type TimelineCheckPolicy = {
  policy_id: string;
  confirmed: boolean;
  authority: string;
  source_url: string;
  observed_at: string;
  required_checks: Array<TimelineRequiredCheck>;
};
export type TimelineCIExternalRun = {
  commit: string;
  run_id: string;
  url: string;
};
export type TimelineCIAttempt = {
  attempt_id: string;
  generation: number;
  check_keys: Array<string>;
  external_run: TimelineCIExternalRun;
  source_url: string;
  observed_at: string;
  actor: TimelineActor;
};
export type TimelineCIObservation = {
  observation_id: string;
  attempt_id: string;
  check_key: string;
  external_run: TimelineCIExternalRun;
  state: TimelineWaitState;
  summary: string;
  observed_at: string;
  received_at: string;
  actor: TimelineActor;
};
export type TimelineCIWait = {
  wait_id: string;
  wait_version: number;
  stage_event_id: string;
  task_id?: string;
  segment_id?: string;
  candidate_id: string;
  pull_request_ref: TimelineDeliveryReference;
  policy: TimelineCheckPolicy;
  state: TimelineWaitState;
  next_action?: string;
  previous_wait_id?: string;
  applicable_attempt_id: Record<string, string>;
  attempts: Array<TimelineCIAttempt>;
  observations: Array<TimelineCIObservation>;
};
export type TimelineWaitCreateRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  target_stage_event_id: string;
  task_id?: string;
  segment_id?: string;
  candidate_id: string;
  pull_request_delivery_id: string;
  policy: TimelineCheckPolicy;
  previous_wait_id?: string;
  next_action?: string;
};
export type TimelineWaitAttemptStartRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  expected_wait_version: number;
  check_keys: Array<string>;
  external_run: TimelineCIExternalRun;
  source_url: string;
  observed_at: string;
};
export type TimelineWaitObserveRequest = {
  schema_version: 3;
  operation_id: string;
  actor: TimelineActor;
  expected_session_version: number;
  reason: string;
  expected_wait_version: number;
  attempt_id: string;
  check_key: string;
  external_run: TimelineCIExternalRun;
  state: TimelineWaitState;
  summary: string;
  observed_at: string;
};
const shapes: Record<string, Array<[string, string, boolean]>> = {
  TimelineActor: [
    ["kind", "string", false],
    ["producer_id", "string", true],
    ["origin", "string", true],
  ],
  TimelineProducer: [
    ["producer_id", "string", false],
    ["source", "TimelineSource", false],
    ["joined_at", "string", false],
  ],
  TimelineJoinRequest: [
    ["request_id", "string", false],
    ["source", "TimelineSource", false],
    ["requested_at", "string", false],
    ["reason", "string", false],
  ],
  TimelineAudit: [
    ["block", "*TimelineBlock", true],
    ["resolution", "string", true],
    ["previous_segment", "*TimelineTaskSegment", true],
    ["next_segment", "*TimelineTaskSegment", true],
    ["plan_revision_id", "string", true],
    ["operation", "string", false],
    ["reason", "string", false],
    ["previous_state", "string", true],
    ["next_state", "string", true],
    ["previous_producer_id", "string", true],
    ["next_producer_id", "string", true],
    ["request_id", "string", true],
    ["target", "*TimelineDecisionTarget", true],
  ],
  TimelineBlock: [
    ["reason", "string", false],
    ["steps", "[]string", false],
    ["owner", "string", false],
    ["resume_when", "string", false],
  ],
  TimelineTaskSegment: [
    ["segment_id", "string", false],
    ["planned_stage", "TimelineStage", false],
    ["order", "int64", false],
    ["state", "string", false],
    ["assigned_producer_id", "string", false],
    ["display_stage_event_id", "string", true],
    ["block", "*TimelineBlock", true],
    ["review_kind", "string", true],
    ["withdrawn", "bool", true],
  ],
  TimelineTask: [
    ["task_id", "string", false],
    ["task_key", "string", false],
    ["plan_order", "int64", false],
    ["title", "string", false],
    ["summary", "string", false],
    ["task_version", "int64", false],
    ["plan_revision_id", "string", true],
    ["segments", "[]TimelineTaskSegment", false],
    ["state", "string", false],
    ["withdrawn", "bool", true],
  ],
  TimelineTaskSegmentInput: [
    ["planned_stage", "TimelineStage", false],
    ["order", "int64", false],
    ["assigned_producer_id", "string", true],
  ],
  TimelineNativeOperation: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
  ],
  TimelineJoinRequestInput: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["source", "TimelineSource", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
  ],
  TimelineProducerAcceptRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["request_id", "string", false],
  ],
  TimelineForegroundTransferRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["target_producer_id", "string", false],
  ],
  TimelineRecoverLeadRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["request_id", "string", false],
  ],
  TimelineTaskCreateRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["task_key", "string", false],
    ["plan_order", "int64", false],
    ["title", "string", false],
    ["summary", "string", false],
    ["plan_revision_id", "string", true],
    ["segments", "[]TimelineTaskSegmentInput", false],
  ],
  TimelineTaskPatchRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["expected_task_version", "int64", false],
    ["title", "*string", true],
    ["summary", "*string", true],
    ["plan_revision_id", "string", true],
    ["decision_target", "*TimelineDecisionTarget", true],
    ["segment_id", "string", true],
    ["order", "*int64", true],
    ["assigned_producer_id", "string", true],
    ["display_stage_event_id", "string", true],
    ["withdrawn", "*bool", true],
    ["plan_order", "int64", true],
  ],
  TimelineTaskTransitionRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["expected_task_version", "int64", false],
    ["segment_id", "string", false],
    ["state", "string", false],
    ["block", "*TimelineBlock", true],
    ["review_kind", "string", true],
    ["resolution", "string", true],
  ],
  TimelineRequiredCheck: [
    ["check_key", "string", false],
    ["label", "string", false],
  ],
  TimelineCheckPolicy: [
    ["policy_id", "string", false],
    ["confirmed", "bool", false],
    ["authority", "string", false],
    ["source_url", "string", false],
    ["observed_at", "string", false],
    ["required_checks", "[]TimelineRequiredCheck", false],
  ],
  TimelineCIExternalRun: [
    ["commit", "string", false],
    ["run_id", "string", false],
    ["url", "string", false],
  ],
  TimelineCIAttempt: [
    ["attempt_id", "string", false],
    ["generation", "int64", false],
    ["check_keys", "[]string", false],
    ["external_run", "TimelineCIExternalRun", false],
    ["source_url", "string", false],
    ["observed_at", "string", false],
    ["actor", "TimelineActor", false],
  ],
  TimelineCIObservation: [
    ["observation_id", "string", false],
    ["attempt_id", "string", false],
    ["check_key", "string", false],
    ["external_run", "TimelineCIExternalRun", false],
    ["state", "string", false],
    ["summary", "string", false],
    ["observed_at", "string", false],
    ["received_at", "string", false],
    ["actor", "TimelineActor", false],
  ],
  TimelineCIWait: [
    ["wait_id", "string", false],
    ["wait_version", "int64", false],
    ["stage_event_id", "string", false],
    ["task_id", "string", true],
    ["segment_id", "string", true],
    ["candidate_id", "string", false],
    ["pull_request_ref", "TimelineDeliveryReference", false],
    ["policy", "TimelineCheckPolicy", false],
    ["state", "string", false],
    ["next_action", "string", true],
    ["previous_wait_id", "string", true],
    ["applicable_attempt_id", "map[string]string", false],
    ["attempts", "[]TimelineCIAttempt", false],
    ["observations", "[]TimelineCIObservation", false],
  ],
  TimelineWaitCreateRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["target_stage_event_id", "string", false],
    ["task_id", "string", true],
    ["segment_id", "string", true],
    ["candidate_id", "string", false],
    ["pull_request_delivery_id", "string", false],
    ["policy", "TimelineCheckPolicy", false],
    ["previous_wait_id", "string", true],
    ["next_action", "string", true],
  ],
  TimelineWaitAttemptStartRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["expected_wait_version", "int64", false],
    ["check_keys", "[]string", false],
    ["external_run", "TimelineCIExternalRun", false],
    ["source_url", "string", false],
    ["observed_at", "string", false],
  ],
  TimelineWaitObserveRequest: [
    ["schema_version", "int", false],
    ["operation_id", "string", false],
    ["actor", "TimelineActor", false],
    ["expected_session_version", "int64", false],
    ["reason", "string", false],
    ["expected_wait_version", "int64", false],
    ["attempt_id", "string", false],
    ["check_key", "string", false],
    ["external_run", "TimelineCIExternalRun", false],
    ["state", "string", false],
    ["summary", "string", false],
    ["observed_at", "string", false],
  ],
};

export const TIMELINE_V3_LIMITS = {
  producers: 16,
  joinRequests: 16,
  tasks: 100,
  segments: 5,
  blockSteps: 10,
  requiredChecks: 50,
  attempts: 20,
  observations: 500,
  waits: 50,
  waitText: 500,
  contentEvents: 500,
  events: 508,
} as const;
export type TimelineTaskState =
  | "not_started"
  | "in_progress"
  | "in_review"
  | "blocked"
  | "completed";
export type TimelineWaitState =
  | "waiting"
  | "succeeded"
  | "failed"
  | "cancelled"
  | "unknown";
const stages = ["research", "plan", "build", "verify", "delivery"];
const taskStates = [
  "not_started",
  "in_progress",
  "in_review",
  "blocked",
  "completed",
];
const waitStates = ["waiting", "succeeded", "failed", "cancelled", "unknown"];
const opaque = /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/;
const documentKey = /^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$/;
type Obj = Record<string, unknown>;
function fail(field: string, code = "invalid_value"): never {
  throw new TimelineContractError(field, code);
}
function record(value: unknown): Obj {
  if (!value || typeof value !== "object" || Array.isArray(value))
    fail("body", "object_required");
  return value as Obj;
}
function text(
  value: unknown,
  field: string,
  max: number,
  required = true,
): void {
  if (
    typeof value !== "string" ||
    [...value].length > max ||
    (required && !value.trim()) ||
    [...value].some((c) => {
      const n = c.codePointAt(0) ?? 0;
      return (n < 32 && n !== 9 && n !== 10) || (n >= 127 && n <= 159);
    })
  )
    fail(field);
}
function id(value: unknown, field: string): void {
  if (typeof value !== "string" || !opaque.test(value))
    fail(field, "invalid_id");
}
function number(value: unknown, field: string): void {
  if (
    !Number.isInteger(value) ||
    (value as number) < 1 ||
    (value as number) > 2147483647
  )
    fail(field);
}
function time(value: unknown, field: string): void {
  if (
    typeof value !== "string" ||
    !/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?(?:Z|[+-]\d\d:\d\d)$/.test(
      value,
    ) ||
    !Number.isFinite(Date.parse(value))
  )
    fail(field);
}
function url(value: unknown, field: string): void {
  if (
    typeof value !== "string" ||
    !/^https:\/\/[A-Za-z0-9.-]+(?::\d+)?(?:[/?#]|$)/.test(value) ||
    /\s/.test(value) ||
    [...value].some((c) => {
      const n = c.codePointAt(0) ?? 0;
      return n < 32 || (n >= 127 && n <= 159);
    })
  )
    fail(field);
  try {
    const u = new URL(value);
    if (u.username || u.password || !u.hostname) fail(field);
  } catch {
    fail(field);
  }
}
function array(value: unknown, field: string, max: number, min = 0): unknown[] {
  if (!Array.isArray(value) || value.length < min || value.length > max)
    fail(field, "invalid_count");
  return value;
}
function source(value: unknown): void {
  validateTimelineSource(value);
}
export function validateTimelineActor(
  value: unknown,
  native = false,
): asserts value is TimelineActor {
  const v = record(value);
  if (
    Object.keys(v).some((k) => !["kind", "producer_id", "origin"].includes(k))
  )
    fail("actor", "unknown_field");
  if (v.kind === "native_producer") {
    id(v.producer_id, "actor.producer_id");
    if (Object.hasOwn(v, "origin")) fail("actor", "invalid_actor");
  } else if (v.kind !== "local_viewer" || Object.hasOwn(v, "producer_id"))
    fail("actor", "invalid_actor");
  else if (
    Object.hasOwn(v, "origin") &&
    v.origin !== "browser" &&
    v.origin !== "cli"
  )
    fail("actor.origin", "invalid_origin");
  if (native && v.kind !== "native_producer")
    fail("actor", "native_producer_required");
}
function shape(name: string, value: unknown): void {
  const v = record(value),
    fields = shapes[name];
  if (!fields) fail(name, "unsupported_shape");
  for (const k of Object.keys(v))
    if (!fields.some((f) => f[0] === k)) fail(k, "unknown_field");
  for (const [k, type, optional] of fields) {
    if (v[k] === undefined) {
      if (!optional) fail(k, "required_field");
      continue;
    }
    if (v[k] === null) fail(k, "null_not_allowed");
    typed(type, v[k], k);
  }
}
function typed(type: string, value: unknown, field: string): void {
  if (type.startsWith("*")) {
    typed(type.slice(1), value, field);
    return;
  }
  if (type.startsWith("[]")) {
    for (const v of array(value, field, 1000)) typed(type.slice(2), v, field);
    return;
  }
  if (type.startsWith("map[string]")) {
    for (const v of Object.values(record(value)))
      typed(type.slice(11), v, field);
    return;
  }
  if (type === "string" || type === "TimelineStage") {
    if (typeof value !== "string") fail(field);
    return;
  }
  if (type === "int64" || type === "int") {
    number(value, field);
    return;
  }
  if (type === "bool") {
    if (typeof value !== "boolean") fail(field);
    return;
  }
  if (type === "TimelineSource") {
    source(value);
    return;
  }
  if (type === "TimelineDeliveryReference") {
    validateTimelineDeliveryReference(value);
    return;
  }
  if (type === "TimelineCandidate") {
    const v = record(value);
    if (Object.keys(v).some((k) => !["branch", "commit", "base"].includes(k)))
      fail(field, "unknown_field");
    for (const k of ["branch", "commit", "base"]) text(v[k], k, 200);
    return;
  }
  if (type === "TimelineDecisionTarget") {
    const v = record(value),
      keys =
        v.kind === "document"
          ? ["kind", "event_id", "document_revision_id"]
          : ["kind", "event_id", "evidence_id", "view_ids"];
    if (
      !["document", "evidence"].includes(v.kind as string) ||
      Object.keys(v).some((k) => !keys.includes(k))
    )
      fail(field);
    id(v.event_id, "event_id");
    id(
      v.kind === "document" ? v.document_revision_id : v.evidence_id,
      "target",
    );
    if (v.view_ids !== undefined)
      for (const idValue of array(v.view_ids, "view_ids", 50))
        id(idValue, "view_ids");
    return;
  }
  shape(type, value);
}
function reason(v: unknown): void {
  text(v, "reason", 500);
}
function block(value: unknown): void {
  shape("TimelineBlock", value);
  const v = value as TimelineBlock;
  reason(v.reason);
  for (const s of array(v.steps, "steps", 10, 1)) text(s, "steps", 500);
  text(v.owner, "owner", 200);
  text(v.resume_when, "resume_when", 500);
}
function run(v: TimelineCIExternalRun): void {
  text(v.commit, "commit", 200);
  text(v.run_id, "run_id", 200);
  url(v.url, "url");
}
function policy(v: TimelineCheckPolicy): void {
  shape("TimelineCheckPolicy", v);
  id(v.policy_id, "policy_id");
  text(v.authority, "authority", 500);
  url(v.source_url, "source_url");
  time(v.observed_at, "observed_at");
  const seen = new Set<string>();
  for (const c of array(
    v.required_checks,
    "required_checks",
    TIMELINE_V3_LIMITS.requiredChecks,
  ) as TimelineRequiredCheck[]) {
    id(c.check_key, "check_key");
    text(c.label, "label", 200);
    if (seen.has(c.check_key)) fail("check_key", "duplicate_key");
    seen.add(c.check_key);
  }
}
export type TimelineV3RequestMap = {
  join_request: TimelineJoinRequestInput;
  producer_accept: TimelineProducerAcceptRequest;
  foreground_transfer: TimelineForegroundTransferRequest;
  recover_lead: TimelineRecoverLeadRequest;
  task_create: TimelineTaskCreateRequest;
  task_patch: TimelineTaskPatchRequest;
  task_transition: TimelineTaskTransitionRequest;
  wait_create: TimelineWaitCreateRequest;
  wait_attempt_start: TimelineWaitAttemptStartRequest;
  wait_observe: TimelineWaitObserveRequest;
};
const operationShapes: Record<keyof TimelineV3RequestMap, string> = {
  join_request: "TimelineJoinRequestInput",
  producer_accept: "TimelineProducerAcceptRequest",
  foreground_transfer: "TimelineForegroundTransferRequest",
  recover_lead: "TimelineRecoverLeadRequest",
  task_create: "TimelineTaskCreateRequest",
  task_patch: "TimelineTaskPatchRequest",
  task_transition: "TimelineTaskTransitionRequest",
  wait_create: "TimelineWaitCreateRequest",
  wait_attempt_start: "TimelineWaitAttemptStartRequest",
  wait_observe: "TimelineWaitObserveRequest",
};
export function isTimelineV3Operation(
  kind: string,
): kind is keyof TimelineV3RequestMap {
  return Object.hasOwn(operationShapes, kind);
}
export function parseTimelineV3Request<K extends keyof TimelineV3RequestMap>(
  kind: K,
  value: unknown,
): TimelineV3RequestMap[K] {
  const v = record(value);
  if (v.schema_version !== 3) fail("schema_version", "unsupported_protocol");
  shape(operationShapes[kind], v);
  if (
    typeof v.operation_id !== "string" ||
    !/^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$/.test(v.operation_id)
  )
    fail("operation_id");
  number(v.expected_session_version, "expected_session_version");
  reason(v.reason);
  if (kind === "join_request") {
    source(v.source);
    return v as TimelineV3RequestMap[K];
  }
  validateTimelineActor(v.actor, kind !== "recover_lead");
  if (
    kind === "recover_lead" &&
    (v.actor as TimelineActor).kind !== "local_viewer"
  )
    fail("actor", "forbidden_actor");
  switch (kind) {
    case "producer_accept":
    case "recover_lead":
      id(v.request_id, "request_id");
      break;
    case "foreground_transfer":
      id(v.target_producer_id, "target_producer_id");
      break;
    case "task_create": {
      if (
        typeof v.task_key !== "string" ||
        !/^T[1-9][0-9]{0,126}$/.test(v.task_key)
      )
        fail("task_key");
      text(v.title, "title", 120);
      text(v.summary, "summary", 4000, false);
      const seen = new Set<string>();
      for (const s of array(
        v.segments,
        "segments",
        5,
        1,
      ) as TimelineTaskSegmentInput[]) {
        if (!stages.includes(s.planned_stage) || seen.has(s.planned_stage))
          fail("segments", "invalid_stage");
        seen.add(s.planned_stage);
        number(s.order, "order");
        if (s.assigned_producer_id !== undefined)
          id(s.assigned_producer_id, "assigned_producer_id");
      }
      break;
    }
    case "task_patch": {
      if (v.plan_order !== undefined) number(v.plan_order, "plan_order");
      if (v.summary !== undefined) text(v.summary, "summary", 4000, false);
      number(v.expected_task_version, "expected_task_version");
      if (
        ![
          "title",
          "summary",
          "plan_revision_id",
          "segment_id",
          "withdrawn",
          "plan_order",
        ].some((k) => Object.hasOwn(v, k))
      )
        fail("body", "empty_patch");
      if (v.title !== undefined) text(v.title, "title", 120);
      if (v.order !== undefined) number(v.order, "order");
      if (
        (v.order !== undefined ||
          v.assigned_producer_id !== undefined ||
          v.display_stage_event_id !== undefined) &&
        v.segment_id === undefined
      )
        fail("segment_id", "required_for_segment_patch");
      break;
    }
    case "task_transition":
      number(v.expected_task_version, "expected_task_version");
      id(v.segment_id, "segment_id");
      if (!taskStates.includes(v.state as string)) fail("state");
      if (v.state === "blocked") {
        if (v.block === undefined) fail("block", "required_for_blocked");
        block(v.block);
      } else if (v.block !== undefined) fail("block", "forbidden_for_state");
      if (
        v.review_kind !== undefined &&
        (v.state !== "in_review" ||
          !["human", "agent"].includes(v.review_kind as string))
      )
        fail("review_kind");
      break;
    case "wait_create":
      if (v.next_action !== undefined)
        text(v.next_action, "next_action", TIMELINE_V3_LIMITS.waitText);
      id(v.target_stage_event_id, "target_stage_event_id");
      id(v.candidate_id, "candidate_id");
      id(v.pull_request_delivery_id, "pull_request_delivery_id");
      policy(v.policy as TimelineCheckPolicy);
      if ((v.task_id === undefined) !== (v.segment_id === undefined))
        fail("task_id", "task_segment_pair_required");
      break;
    case "wait_attempt_start": {
      number(v.expected_wait_version, "expected_wait_version");
      const seen = new Set<string>();
      for (const k of array(
        v.check_keys,
        "check_keys",
        TIMELINE_V3_LIMITS.requiredChecks,
        1,
      )) {
        id(k, "check_keys");
        if (seen.has(k as string)) fail("check_keys", "duplicate_key");
        seen.add(k as string);
      }
      run(v.external_run as TimelineCIExternalRun);
      url(v.source_url, "source_url");
      time(v.observed_at, "observed_at");
      break;
    }
    case "wait_observe":
      number(v.expected_wait_version, "expected_wait_version");
      id(v.attempt_id, "attempt_id");
      id(v.check_key, "check_key");
      if (!waitStates.includes(v.state as string)) fail("state");
      run(v.external_run as TimelineCIExternalRun);
      time(v.observed_at, "observed_at");
      text(v.summary, "summary", TIMELINE_V3_LIMITS.waitText);
      break;
  }
  return v as TimelineV3RequestMap[K];
}
export function validateTimelineSessionPatchActor(
  r: TimelineSessionPatchRequest,
): void {
  validateTimelineActor(r.actor);
  if (r.actor.kind === "local_viewer") {
    if (r.summary !== undefined || r.state !== undefined)
      fail("actor", "forbidden_actor");
  } else if (
    r.title !== undefined ||
    r.archived !== undefined ||
    r.pinned !== undefined
  )
    fail("actor", "local_viewer_required");
  if (r.state !== undefined || r.reason !== undefined) reason(r.reason);
}
export function validateTimelineAppendV3(r: TimelineAppendEventRequest): void {
  validateTimelineActor(r.actor, true);
  if (
    [
      "state_transition",
      "producer_joined",
      "foreground_transferred",
      "task_state_transition",
      "local_action",
    ].includes(r.kind)
  )
    fail("kind", "server_only");
  if (r.kind === "stage" || r.kind === "closure") {
    if (
      r.target_stage_event_id !== undefined ||
      r.task_id !== undefined ||
      r.segment_id !== undefined
    )
      fail("target_stage_event_id", "forbidden_for_kind");
    reason(r.reason);
  }
  if (r.target_stage_event_id !== undefined)
    id(r.target_stage_event_id, "target_stage_event_id");
  if ((r.task_id === undefined) !== (r.segment_id === undefined))
    fail("task_id", "task_segment_pair_required");
  if (r.task_id !== undefined) {
    id(r.task_id, "task_id");
    id(r.segment_id, "segment_id");
    id(r.target_stage_event_id, "target_stage_event_id");
  }
  if (r.kind === "review") {
    if (!["in_progress", "completed", "failed"].includes(r.review_status ?? ""))
      fail("review_status");
    id(r.target_stage_event_id, "target_stage_event_id");
  } else if (r.review_status !== undefined)
    fail("review_status", "forbidden_for_kind");
}
export function validateTimelineSnapshotV3(
  r: TimelineDocumentSnapshotRequest,
): void {
  validateTimelineActor(r.actor, true);
  if (!documentKey.test(r.document_key)) fail("document_key", "invalid_key");
  if (r.predecessor_revision_id !== undefined)
    id(r.predecessor_revision_id, "predecessor_revision_id");
  number(r.expected_session_version, "expected_session_version");
  if ((r.task_id === undefined) !== (r.segment_id === undefined))
    fail("task_id", "task_segment_pair_required");
}
export function aggregateTimelineTaskState(
  segments: TimelineTaskSegment[],
): TimelineTaskState {
  const active = segments.filter((s) => !s.withdrawn);
  if (active.some((s) => s.state === "blocked")) return "blocked";
  if (active.some((s) => s.state === "in_review")) return "in_review";
  if (active.length && active.every((s) => s.state === "completed"))
    return "completed";
  if (active.every((s) => s.state === "not_started")) return "not_started";
  return "in_progress";
}
export function validateTimelineTaskTransition(
  previous: string,
  r: TimelineTaskTransitionRequest,
): void {
  parseTimelineV3Request("task_transition", r);
  if (previous === r.state) return;
  const allowed: Record<string, string[]> = {
    not_started: ["in_progress", "blocked"],
    in_progress: ["in_review", "blocked", "completed"],
    in_review: ["in_progress", "blocked", "completed"],
    blocked: ["in_progress"],
    completed: ["in_progress"],
  };
  if (!allowed[previous]?.includes(r.state))
    fail("state", "invalid_transition");
  if (previous === "blocked") text(r.resolution, "resolution", 500);
}
export function validateTimelineAuthority(
  s: TimelineSession,
  operation: string,
  actor: TimelineActor,
  owner = "",
): void {
  validateTimelineActor(actor);
  const local = [
    "session_rename",
    "session_archive",
    "session_pin",
    "session_delete",
    "recover_lead",
  ].includes(operation);
  if (actor.kind === "local_viewer") {
    if (local) return;
    fail("actor", "forbidden_actor");
  }
  if (local) fail("actor", "local_viewer_required");
  const p = s.producers.find((p) => p.producer_id === actor.producer_id);
  if (!p) fail("actor", "producer_not_accepted");
  const lead = actor.producer_id === s.foreground_producer_id;
  if (
    [
      "producer_accept",
      "foreground_transfer",
      "session_patch",
      "stage_start",
      "stage_update",
      "closure",
      "task_create",
      "task_patch",
      "ticket_create",
      "ticket_patch",
      "delivery_create",
      "delivery_patch",
    ].includes(operation)
  ) {
    if (!lead) fail("actor", "lead_required");
  } else if (
    ["wait_create", "wait_attempt_start", "wait_observe"].includes(operation)
  ) {
    if (!lead && (!owner || owner !== actor.producer_id))
      fail("actor", "segment_assignee_required");
  } else if (operation === "event_patch") {
    if (owner && owner !== actor.producer_id)
      fail("actor", "content_owner_required");
  } else if (operation === "event_withdraw") {
    if (owner && owner !== actor.producer_id && !lead)
      fail("actor", "content_owner_required");
  } else if (
    ![
      "append_event",
      "candidate_create",
      "artifact_add",
      "document_add",
      "document_snapshot",
      "task_transition",
      "wait_create",
      "wait_attempt_start",
      "wait_observe",
    ].includes(operation)
  )
    fail("operation", "unsupported_operation");
}
export function validateTimelineTaskReference(
  s: TimelineSession,
  taskID: string | undefined,
  segmentID: string | undefined,
  stageID: string | undefined,
): void {
  if (!taskID && !segmentID) return;
  const t = s.tasks.find((t) => t.task_id === taskID && !t.withdrawn),
    seg = t?.segments.find((s) => s.segment_id === segmentID && !s.withdrawn);
  if (!seg) fail("task_id", "unknown_task_segment");
  if (!stageID || seg.display_stage_event_id !== stageID)
    fail("segment_id", "stage_binding_mismatch");
  if (
    !s.events.some(
      (e) =>
        e.event_id === stageID &&
        e.kind === "stage" &&
        !e.withdrawn &&
        e.stage === seg.planned_stage,
    )
  )
    fail("task_id", "unknown_task_segment");
}
// Stored evidence keeps its original occurrence after rebinding or retirement.
// New mutation inputs still require the current binding above.
function validateTimelineHistoricalTaskReference(
  s: TimelineSession,
  taskID: string | undefined,
  segmentID: string | undefined,
  stageID: string | undefined,
): void {
  if (!taskID && !segmentID) return;
  const task = s.tasks.find((t) => t.task_id === taskID);
  const segment = task?.segments.find((seg) => seg.segment_id === segmentID);
  if (
    segment &&
    s.events.some(
      (e) =>
        e.event_id === stageID &&
        e.kind === "stage" &&
        e.stage === segment.planned_stage,
    )
  )
    return;
  fail("task_id", "unknown_task_segment");
}

export function aggregateTimelineCIWait(
  policy: TimelineCheckPolicy,
  applicable: Record<string, string>,
  observed: Record<string, TimelineCIObservation>,
): TimelineWaitState {
  if (!policy.confirmed || !policy.required_checks.length) return "unknown";
  let waiting = false,
    cancelled = false,
    unknown = false;
  for (const c of policy.required_checks) {
    const o = observed[c.check_key];
    if (
      !o ||
      !applicable[c.check_key] ||
      o.attempt_id !== applicable[c.check_key]
    ) {
      unknown = true;
      continue;
    }
    if (o.state === "failed") return "failed";
    if (o.state === "cancelled") cancelled = true;
    if (o.state === "unknown") unknown = true;
    if (o.state === "waiting") waiting = true;
  }
  return cancelled
    ? "cancelled"
    : unknown
      ? "unknown"
      : waiting
        ? "waiting"
        : "succeeded";
}
export function sameTimelineExternalRunIdentity(
  a: TimelineCIExternalRun,
  b: TimelineCIExternalRun,
): boolean {
  return a.commit === b.commit && a.run_id === b.run_id;
}

export function validateTimelineCIWait(
  w: TimelineCIWait,
  candidate: TimelineCandidateRecord,
): void {
  if (!candidate) fail("candidate_id", "candidate_context_required");
  if (w.next_action !== undefined)
    text(w.next_action, "next_action", TIMELINE_V3_LIMITS.waitText);
  shape("TimelineCIWait", w);
  policy(w.policy);
  id(w.candidate_id, "candidate_id");
  for (const field of ["branch", "commit", "base"] as const)
    text(candidate[field], field, 256);
  if (candidate.candidate_id !== w.candidate_id)
    fail("candidate_id", "candidate_mismatch");
  validateTimelineDeliveryReference(w.pull_request_ref);
  if (!waitStates.includes(w.state)) fail("state");
  array(w.attempts, "attempts", TIMELINE_V3_LIMITS.attempts);
  array(w.observations, "observations", 500);
  const required = new Set(w.policy.required_checks.map((c) => c.check_key)),
    attempts = new Map<string, TimelineCIAttempt>();
  let generation = 0;
  for (const a of w.attempts) {
    if (a.external_run.commit !== candidate.commit)
      fail("commit", "candidate_mismatch");
    id(a.attempt_id, "attempt_id");
    if (attempts.has(a.attempt_id) || a.generation <= generation)
      fail("generation", "non_monotonic_attempt");
    generation = a.generation;
    const seen = new Set<string>();
    for (const k of array(
      a.check_keys,
      "check_keys",
      TIMELINE_V3_LIMITS.requiredChecks,
      1,
    ) as string[]) {
      if (!required.has(k) || seen.has(k))
        fail("check_key", "unknown_or_duplicate");
      seen.add(k);
    }
    if (
      [...attempts.values()].some((old) =>
        sameTimelineExternalRunIdentity(old.external_run, a.external_run),
      )
    )
      fail("external_run", "attempt_already_known");
    run(a.external_run);
    url(a.source_url, "source_url");
    time(a.observed_at, "observed_at");
    validateTimelineActor(a.actor, true);
    attempts.set(a.attempt_id, a);
  }
  for (const [k, idValue] of Object.entries(w.applicable_attempt_id)) {
    const a = attempts.get(idValue);
    if (!required.has(k) || !a?.check_keys.includes(k))
      fail("applicable_attempt_id", "invalid_attempt");
    if (
      w.attempts.some(
        (other) =>
          other.check_keys.includes(k) && other.generation > a.generation,
      )
    )
      fail("applicable_attempt_id", "superseded_attempt");
  }
  const observed: Record<string, TimelineCIObservation> = {},
    ids = new Set<string>();
  for (const o of w.observations) {
    text(o.summary, "summary", TIMELINE_V3_LIMITS.waitText);
    const a = attempts.get(o.attempt_id);
    if (
      !a?.check_keys.includes(o.check_key) ||
      a.external_run.commit !== o.external_run.commit ||
      a.external_run.run_id !== o.external_run.run_id ||
      a.external_run.url !== o.external_run.url ||
      ids.has(o.observation_id)
    )
      fail("observations", "unknown_or_mixed_attempt");
    id(o.observation_id, "observation_id");
    ids.add(o.observation_id);
    if (!waitStates.includes(o.state)) fail("state");
    time(o.observed_at, "observed_at");
    time(o.received_at, "received_at");
    validateTimelineActor(o.actor, true);
    if (
      w.applicable_attempt_id[o.check_key] === o.attempt_id &&
      (!observed[o.check_key] ||
        Date.parse(observed[o.check_key].observed_at) <
          Date.parse(o.observed_at))
    )
      observed[o.check_key] = o;
  }
  if (
    w.state !==
    aggregateTimelineCIWait(w.policy, w.applicable_attempt_id, observed)
  )
    fail("state", "wait_aggregate_mismatch");
}
// States that must point at a foreground stage event; the session and index
// validators share them.
export function timelineStateRequiresForeground(state: unknown): boolean {
  return state === "in_review" || state === "awaiting_validation";
}
// States that may hold a foreground stage event.
export function timelineStateAllowsForeground(state: unknown): boolean {
  return state === "active" || timelineStateRequiresForeground(state);
}
export function validateTimelineV3Graph(s: TimelineSession): void {
  if (s.schema_version !== 3) fail("schema_version", "unsupported_protocol");
  array(s.producers, "producers", 16, 1);
  array(s.pending_join_requests, "pending_join_requests", 16);
  array(s.tasks, "tasks", 100);
  array(s.waits, "waits", TIMELINE_V3_LIMITS.waits);
  if (
    timelineStateRequiresForeground(s.state) &&
    s.current_stage_event_id === null
  )
    fail("current_stage_event_id", "foreground_required");
  if (s.current_stage_event_id) {
    const e = s.events.find((e) => e.event_id === s.current_stage_event_id);
    if (
      e &&
      s.state !== (e.stage_state === "in_progress" ? "active" : e.stage_state)
    )
      fail("state", "foreground_state_mismatch");
  }
  const members = new Set<string>(),
    sources = new Set<string>();
  for (const p of s.producers) {
    shape("TimelineProducer", p);
    id(p.producer_id, "producer_id");
    source(p.source);
    const key = nativeSourceIdentity(p.source);
    if (members.has(p.producer_id) || sources.has(key))
      fail("producers", "duplicate_identity");
    time(p.joined_at, "joined_at");
    members.add(p.producer_id);
    sources.add(key);
  }
  if (
    !members.has(s.foreground_producer_id) ||
    !s.producers.some(
      (p) =>
        p.source.project_root === s.source.project_root &&
        p.source.host === s.source.host &&
        p.source.native_session_id === s.source.native_session_id,
    )
  )
    fail("foreground_producer_id", "accepted_original_and_lead_required");
  const requests = new Set<string>();
  for (const r of s.pending_join_requests) {
    shape("TimelineJoinRequest", r);
    id(r.request_id, "request_id");
    source(r.source);
    const key = nativeSourceIdentity(r.source);
    if (sources.has(key) || requests.has(r.request_id))
      fail("pending_join_requests", "duplicate_identity");
    reason(r.reason);
    time(r.requested_at, "requested_at");
    sources.add(key);
    requests.add(r.request_id);
  }
  const acceptedActor = (a: TimelineActor) => {
    validateTimelineActor(a, true);
    if (!members.has(a.producer_id ?? ""))
      fail("actor", "producer_not_accepted");
  };
  for (const d of s.documents) acceptedActor(d.actor);
  for (const a of s.artifacts) {
    acceptedActor(a.actor);
    timelineCandidateByID(s, a.candidate_id);
  }
  const stagesMap = new Map(
      s.events.filter((e) => e.kind === "stage").map((e) => [e.event_id, e]),
    ),
    revisions = new Map<string, TimelineSession["events"][number]>(),
    chains = new Map<string, string>();
  for (const e of s.events) {
    validateTimelineActor(e.actor);
    if (e.actor.kind === "native_producer" && !members.has(e.actor.producer_id))
      fail("actor", "producer_not_accepted");
    for (const r of e.document_snapshots ?? []) {
      acceptedActor(r.actor);
      if (
        !documentKey.test(r.document_key) ||
        (r.predecessor_revision_id ?? "") !== (chains.get(r.document_key) ?? "")
      )
        fail("document_key", "invalid_chain");
      const prev = revisions.get(r.predecessor_revision_id ?? "");
      if (
        prev &&
        prev.event_id !== e.event_id &&
        (prev.stage !== e.stage || e.related_event_id !== prev.event_id)
      )
        fail("predecessor_revision_id", "stage_chain_mismatch");
      chains.set(r.document_key, r.document_revision_id);
      revisions.set(r.document_revision_id, e);
    }
  }
  const ids = new Set<string>(),
    keys = new Set<string>(),
    segments = new Set<string>(),
    orders = new Set<string>();
  const planOrders = new Set<number>();
  for (const t of s.tasks) {
    number(t.plan_order, "plan_order");
    if (!t.withdrawn) {
      if (planOrders.has(t.plan_order))
        fail("plan_order", "duplicate_plan_order");
      planOrders.add(t.plan_order);
    }
    shape("TimelineTask", t);
    id(t.task_id, "task_id");
    if (
      ids.has(t.task_id) ||
      keys.has(t.task_key) ||
      !/^T[1-9][0-9]{0,126}$/.test(t.task_key)
    )
      fail("tasks", "invalid_identity_or_segments");
    ids.add(t.task_id);
    keys.add(t.task_key);
    text(t.title, "title", 120);
    number(t.task_version, "task_version");
    if (t.plan_revision_id && !revisions.has(t.plan_revision_id))
      fail("plan_revision_id", "unknown_revision");
    const kinds = new Set<string>();
    for (const seg of array(
      t.segments,
      "segments",
      5,
      1,
    ) as TimelineTaskSegment[]) {
      id(seg.segment_id, "segment_id");
      if (
        segments.has(seg.segment_id) ||
        kinds.has(seg.planned_stage) ||
        !stages.includes(seg.planned_stage) ||
        !taskStates.includes(seg.state) ||
        !members.has(seg.assigned_producer_id)
      )
        fail("segments", "invalid_segment");
      segments.add(seg.segment_id);
      kinds.add(seg.planned_stage);
      number(seg.order, "order");
      const key = `${seg.planned_stage}:${seg.order}`;
      if (!seg.withdrawn && !t.withdrawn) {
        if (orders.has(key)) fail("order", "duplicate_stage_order");
        orders.add(key);
      }
      if (
        seg.display_stage_event_id &&
        stagesMap.get(seg.display_stage_event_id)?.stage !== seg.planned_stage
      )
        fail("display_stage_event_id", "stage_binding_mismatch");
      if (seg.state === "blocked") {
        if (!seg.block) fail("block", "required_for_blocked");
        block(seg.block);
      } else if (seg.block) fail("block", "forbidden_for_state");
    }
    if (t.state !== aggregateTimelineTaskState(t.segments))
      fail("state", "aggregate_mismatch");
  }
  for (const e of s.events) {
    if (
      e.audit ||
      [
        "state_transition",
        "producer_joined",
        "foreground_transferred",
        "task_state_transition",
        "local_action",
      ].includes(e.kind)
    )
      validateTimelineAuditReferences(s, e);
    if (e.stage_event_id) {
      const stage = stagesMap.get(e.stage_event_id);
      if (
        !stage ||
        e.kind === "stage" ||
        e.kind === "closure" ||
        stage.sequence > e.sequence
      )
        fail("stage_event_id", "invalid_stage_reference");
    }
    validateTimelineHistoricalTaskReference(
      s,
      e.task_id,
      e.segment_id,
      e.stage_event_id,
    );
  }
  for (const e of s.events)
    for (const r of e.document_snapshots ?? [])
      validateTimelineHistoricalTaskReference(
        s,
        r.task_id,
        r.segment_id,
        e.event_id,
      );
  const waits = new Set<string>();
  for (const w of s.waits) {
    const d = s.deliveries.find(
      (d) => d.delivery_id === w.pull_request_ref.delivery_id,
    );
    if (!d) fail("delivery_id", "unknown_delivery");
    if (d.kind !== "pull_request")
      fail("pull_request_ref", "pull_request_required");
    id(w.wait_id, "wait_id");
    if (waits.has(w.wait_id)) fail("wait_id", "duplicate_id");
    if (w.previous_wait_id && !waits.has(w.previous_wait_id))
      fail("previous_wait_id", "unknown_wait");
    waits.add(w.wait_id);
    number(w.wait_version, "wait_version");
    if (stagesMap.get(w.stage_event_id)?.stage !== "delivery")
      fail("stage_event_id", "delivery_required");
    validateTimelineHistoricalTaskReference(
      s,
      w.task_id,
      w.segment_id,
      w.stage_event_id,
    );
    for (const a of w.attempts) acceptedActor(a.actor);
    for (const o of w.observations) acceptedActor(o.actor);
    validateTimelineCIWait(w, timelineCandidateByID(s, w.candidate_id));
  }
}
export function validateTimelineV3OperationReferences(
  s: TimelineSession,
  kind: keyof TimelineV3RequestMap,
  objectID: string,
  value: unknown,
): void {
  const r = parseTimelineV3Request(
    kind,
    value,
  ) as TimelineV3RequestMap[keyof TimelineV3RequestMap];
  if (r.expected_session_version !== s.version)
    fail("expected_session_version", "version_conflict");
  if ("source" in r) {
    const identity = nativeSourceIdentity(r.source);
    if (
      s.producers.some((p) => nativeSourceIdentity(p.source) === identity) ||
      s.pending_join_requests.some(
        (q) => nativeSourceIdentity(q.source) === identity,
      )
    )
      fail("source", "duplicate_identity");
    if (s.pending_join_requests.length >= 16)
      fail("pending_join_requests", "limit");
    return;
  }
  if (s.state === "closed" || s.state === "failed")
    fail("state", "session_terminal");
  let assignee = "";
  if (kind === "wait_create") {
    const req = r as TimelineWaitCreateRequest;
    assignee = waitAssignee(s, req.task_id, req.segment_id);
  }
  if (kind === "wait_attempt_start" || kind === "wait_observe") {
    const w = s.waits.find((w) => w.wait_id === objectID);
    if (w) assignee = waitAssignee(s, w.task_id, w.segment_id);
  }
  validateTimelineAuthority(s, kind, r.actor, assignee);
  const member = (idValue: string) =>
      s.producers.some((p) => p.producer_id === idValue),
    revision = (idValue: string) =>
      s.events.some((e) =>
        e.document_snapshots?.some((r) => r.document_revision_id === idValue),
      );
  if (kind === "producer_accept" || kind === "recover_lead") {
    const req = r as TimelineProducerAcceptRequest;
    if (!s.pending_join_requests.some((q) => q.request_id === req.request_id))
      fail("request_id", "pending_request_required");
    if (s.producers.length >= 16) fail("producers", "limit");
    return;
  }
  if (kind === "foreground_transfer") {
    if (!member((r as TimelineForegroundTransferRequest).target_producer_id))
      fail("target_producer_id", "producer_not_accepted");
    return;
  }
  if (kind === "task_create") {
    const req = r as TimelineTaskCreateRequest;
    if (s.tasks.some((t) => !t.withdrawn && t.plan_order === req.plan_order))
      fail("plan_order", "duplicate_plan_order");
    if (s.tasks.length >= 100) fail("tasks", "limit");
    if (s.tasks.some((t) => t.task_key === req.task_key))
      fail("task_key", "duplicate_key");
    if (req.plan_revision_id && !revision(req.plan_revision_id))
      fail("plan_revision_id", "unknown_revision");
    for (const seg of req.segments) {
      if (seg.assigned_producer_id && !member(seg.assigned_producer_id))
        fail("assigned_producer_id", "producer_not_accepted");
      if (
        s.tasks.some(
          (t) =>
            !t.withdrawn &&
            t.segments.some(
              (old) =>
                !old.withdrawn &&
                old.planned_stage === seg.planned_stage &&
                old.order === seg.order,
            ),
        )
      )
        fail("order", "duplicate_stage_order");
    }
    return;
  }
  if (kind === "task_patch" || kind === "task_transition") {
    const t = s.tasks.find((t) => t.task_id === objectID && !t.withdrawn),
      req = r as TimelineTaskPatchRequest | TimelineTaskTransitionRequest;
    if (!t) fail("task_id", "unknown_task");
    if (t.task_version !== req.expected_task_version)
      fail("expected_task_version", "version_conflict");
    if (kind === "task_patch") {
      const req = r as TimelineTaskPatchRequest;
      if (
        req.plan_order !== undefined &&
        req.plan_order !== t.plan_order &&
        !req.plan_revision_id &&
        !req.decision_target
      )
        fail("plan_order", "scope_reference_required");
      if (
        req.plan_order !== undefined &&
        s.tasks.some(
          (t) =>
            !t.withdrawn &&
            t.task_id !== objectID &&
            t.plan_order === req.plan_order,
        )
      )
        fail("plan_order", "duplicate_plan_order");
      if (req.decision_target)
        validateTimelineDecisionReference(s, req.decision_target);
      if (req.plan_revision_id && !revision(req.plan_revision_id))
        fail("plan_revision_id", "unknown_revision");
      if (req.assigned_producer_id && !member(req.assigned_producer_id))
        fail("assigned_producer_id", "producer_not_accepted");
      if (req.segment_id) {
        const seg = t.segments.find((seg) => seg.segment_id === req.segment_id);
        if (!seg) fail("segment_id", "unknown_task_segment");
        if (
          req.display_stage_event_id &&
          !s.events.some(
            (e) =>
              e.event_id === req.display_stage_event_id &&
              e.kind === "stage" &&
              !e.withdrawn &&
              e.stage === seg.planned_stage,
          )
        )
          fail("display_stage_event_id", "stage_binding_mismatch");
      }
    } else {
      const req = r as TimelineTaskTransitionRequest,
        seg = t.segments.find(
          (seg) => seg.segment_id === req.segment_id && !seg.withdrawn,
        );
      if (!seg) fail("segment_id", "unknown_task_segment");
      if (
        r.actor.producer_id !== s.foreground_producer_id &&
        r.actor.producer_id !== seg.assigned_producer_id
      )
        fail("actor", "segment_assignee_required");
      validateTimelineTaskReference(
        s,
        t.task_id,
        seg.segment_id,
        seg.display_stage_event_id,
      );
      if (
        seg.state === "completed" &&
        req.state !== "completed" &&
        s.events.some(
          (e) =>
            e.event_id === seg.display_stage_event_id &&
            (e.stage_state === "completed" || e.stage_state === "failed"),
        )
      )
        fail("display_stage_event_id", "reentry_binding_required");
      validateTimelineTaskTransition(seg.state, req);
    }
    return;
  }
  if (kind === "wait_create") {
    const req = r as TimelineWaitCreateRequest;
    timelineCandidateByID(s, req.candidate_id);
    const d = s.deliveries.find(
      (d) => d.delivery_id === req.pull_request_delivery_id,
    );
    if (!d) fail("delivery_id", "unknown_delivery");
    if (d.kind !== "pull_request")
      fail("pull_request_delivery_id", "pull_request_required");
    if (s.waits.length >= TIMELINE_V3_LIMITS.waits) fail("waits", "limit");
    if (
      !s.events.some(
        (e) =>
          e.event_id === req.target_stage_event_id &&
          e.kind === "stage" &&
          !e.withdrawn &&
          e.stage === "delivery",
      )
    )
      fail("target_stage_event_id", "delivery_required");
    validateTimelineTaskReference(
      s,
      req.task_id,
      req.segment_id,
      req.target_stage_event_id,
    );
    if (req.previous_wait_id) {
      const old = s.waits.find((w) => w.wait_id === req.previous_wait_id);
      if (!old) fail("previous_wait_id", "unknown_wait");
      if (
        old.candidate_id === req.candidate_id &&
        old.policy.policy_id === req.policy.policy_id
      )
        fail("previous_wait_id", "same_candidate_policy");
    }
    return;
  }
  const w = s.waits.find((w) => w.wait_id === objectID),
    req = r as TimelineWaitAttemptStartRequest | TimelineWaitObserveRequest;
  if (!w) fail("wait_id", "unknown_wait");
  if (w.wait_version !== req.expected_wait_version)
    fail("expected_wait_version", "version_conflict");
  if (kind === "wait_attempt_start") {
    const req = r as TimelineWaitAttemptStartRequest;
    if (
      req.external_run.commit !==
      timelineCandidateByID(s, w.candidate_id).commit
    )
      fail("commit", "candidate_mismatch");
    if (w.attempts.length >= TIMELINE_V3_LIMITS.attempts)
      fail("attempts", "limit");
    for (const k of req.check_keys)
      if (!w.policy.required_checks.some((c) => c.check_key === k))
        fail("check_key", "unknown_check");
    if (
      w.attempts.some((a) =>
        sameTimelineExternalRunIdentity(a.external_run, req.external_run),
      )
    )
      fail("external_run", "attempt_already_known");
  } else {
    const req = r as TimelineWaitObserveRequest;
    if (w.observations.length >= 500) fail("observations", "limit");
    if (
      !w.attempts.some(
        (a) =>
          a.attempt_id === req.attempt_id &&
          a.check_keys.includes(req.check_key) &&
          a.external_run.commit === req.external_run.commit &&
          a.external_run.run_id === req.external_run.run_id &&
          a.external_run.url === req.external_run.url,
      )
    )
      fail("attempt_id", "unknown_or_mixed_attempt");
  }
}
export function validateTimelineEventCapacity(
  eventCount: number,
  newEvents: number,
  lifecycle: boolean,
): void {
  if (
    eventCount < 0 ||
    newEvents < 0 ||
    eventCount + newEvents > (lifecycle ? 508 : 500)
  )
    fail("events", "limit");
}
// The final candidate must retain room for closure and its immutable audit.
export function validateTimelineClosureReserve(s: TimelineSession): void {
  if (s.state === "closed") return;
  const required =
    2 + (s.current_stage_event_id !== null || s.state === "failed" ? 1 : 0);
  if (s.events.length + required > 508) fail("events", "limit");
}

export function resolveTimelineStageEventID(
  s: TimelineSession,
  r: TimelineAppendEventRequest,
): string | undefined {
  validateTimelineAuthority(
    s,
    r.kind === "stage"
      ? "stage_start"
      : r.kind === "closure"
        ? "closure"
        : "append_event",
    r.actor,
  );
  if (["stage", "closure", "request", "work_started"].includes(r.kind)) {
    if (r.target_stage_event_id)
      fail("target_stage_event_id", "forbidden_for_kind");
    return;
  }
  let target = r.target_stage_event_id;
  if (!target) {
    if (r.actor.producer_id !== s.foreground_producer_id)
      fail("target_stage_event_id", "collaborator_target_required");
    target =
      s.current_stage_event_id ??
      s.events.filter((e) => e.kind === "stage" && !e.withdrawn).at(-1)
        ?.event_id;
  }
  if (
    target &&
    !s.events.some(
      (e) => e.event_id === target && e.kind === "stage" && !e.withdrawn,
    )
  )
    fail("target_stage_event_id", "stage_unavailable");
  validateTimelineTaskReference(s, r.task_id, r.segment_id, target);
  return target;
}
export function validateTimelineSnapshotReferencesV3(
  s: TimelineSession,
  eventID: string,
  r: TimelineDocumentSnapshotRequest,
): void {
  validateTimelineAuthority(s, "document_snapshot", r.actor);
  if (r.expected_session_version !== s.version)
    fail("expected_session_version", "version_conflict");
  validateTimelineTaskReference(s, r.task_id, r.segment_id, eventID);
  const current = s.events.find((e) => e.event_id === eventID);
  let latest: string | undefined,
    previous: TimelineSession["events"][number] | undefined;
  for (const e of s.events)
    for (const snap of e.document_snapshots ?? [])
      if (snap.document_key === r.document_key) {
        latest = snap.document_revision_id;
        previous = e;
      }
  if (r.predecessor_revision_id !== latest)
    fail("predecessor_revision_id", "current_predecessor_required");
  if (
    previous &&
    previous.event_id !== eventID &&
    (previous.stage !== current?.stage ||
      current?.related_event_id !== previous.event_id)
  )
    fail("document_key", "stage_chain_mismatch");
}

function waitAssignee(
  s: TimelineSession,
  taskID: string | undefined,
  segmentID: string | undefined,
): string {
  return (
    s.tasks
      .find((t) => t.task_id === taskID && !t.withdrawn)
      ?.segments.find(
        (seg) =>
          seg.segment_id === segmentID &&
          !seg.withdrawn &&
          seg.planned_stage === "delivery",
      )?.assigned_producer_id ?? ""
  );
}

/** A server origin does not replace the closed per-operation audit contract. */
export function validateTimelineAuditEvent(
  e: TimelineSession["events"][number],
): void {
  const a = e.audit;
  if (!a || e.origin !== "server" || e.withdrawn)
    fail("audit", "server_event_required");
  shape("TimelineAudit", a);
  validateTimelineActor(e.actor);
  reason(a.reason);
  const allowed = new Set(["operation", "reason", "target"]),
    allow = (...keys: string[]) => {
      for (const k of keys) allowed.add(k);
    };
  switch (e.kind) {
    case "state_transition": {
      if (
        !["stage_start", "stage_update", "session_patch", "closure"].includes(
          a.operation,
        )
      )
        fail("audit.operation", "invalid_audit_operation");
      validateTimelineActor(e.actor, true);
      let states = [
        "active",
        "in_review",
        "awaiting_validation",
        "paused",
        "closed",
        "failed",
      ];
      if (a.operation === "stage_start" || a.operation === "stage_update") {
        states = [
          "in_progress",
          "in_review",
          "awaiting_validation",
          "paused",
          "completed",
          "failed",
        ];
        id(e.stage_event_id, "stage_event_id");
      }
      if (
        !states.includes(a.previous_state ?? "") ||
        !states.includes(a.next_state ?? "") ||
        a.previous_state === a.next_state
      )
        fail("audit", "invalid_state_transition");
      allow("previous_state", "next_state");
      break;
    }
    case "producer_joined":
      if (!["producer_accept", "recover_lead"].includes(a.operation))
        fail("audit.operation", "invalid_audit_operation");
      id(a.request_id, "audit.request_id");
      id(a.next_producer_id, "audit.next_producer_id");
      allow("request_id", "next_producer_id");
      break;
    case "foreground_transferred":
      if (!["foreground_transfer", "recover_lead"].includes(a.operation))
        fail("audit.operation", "invalid_audit_operation");
      id(a.previous_producer_id, "audit.previous_producer_id");
      id(a.next_producer_id, "audit.next_producer_id");
      if (a.previous_producer_id === a.next_producer_id)
        fail("audit", "invalid_producer_transition");
      allow("previous_producer_id", "next_producer_id");
      if (a.operation === "recover_lead") {
        id(a.request_id, "audit.request_id");
        allow("request_id");
      }
      break;
    case "task_state_transition": {
      if (a.operation !== "task_transition")
        fail("audit.operation", "invalid_audit_operation");
      validateTimelineActor(e.actor, true);
      id(e.task_id, "task_id");
      id(e.segment_id, "segment_id");
      id(e.stage_event_id, "stage_event_id");
      const transitions: Record<string, string[]> = {
        not_started: ["in_progress", "blocked"],
        in_progress: ["in_review", "blocked", "completed"],
        in_review: ["in_progress", "blocked", "completed"],
        blocked: ["in_progress"],
        completed: ["in_progress"],
      };
      if (!transitions[a.previous_state ?? ""]?.includes(a.next_state ?? ""))
        fail("audit", "invalid_state_transition");
      allow(
        "previous_state",
        "next_state",
        "block",
        "resolution",
        "previous_segment",
        "next_segment",
        "plan_revision_id",
      );
      if (a.next_state === "blocked" && !a.block)
        fail("block", "required_for_blocked");
      if (a.block) {
        if (a.next_state !== "blocked" && a.previous_state !== "blocked")
          fail("block", "forbidden_for_state");
        block(a.block);
      }
      if (a.previous_state === "blocked") text(a.resolution, "resolution", 500);
      else if (a.resolution !== undefined)
        fail("resolution", "forbidden_for_state");
      for (const seg of [a.previous_segment, a.next_segment])
        if (seg && seg.segment_id !== e.segment_id)
          fail("audit.segment_id", "unknown_task_segment");
      break;
    }
    case "local_action":
      if (
        ![
          "session_rename",
          "session_archive",
          "session_pin",
          "session_delete",
        ].includes(a.operation) ||
        e.actor.kind !== "local_viewer"
      )
        fail("audit.operation", "invalid_audit_operation");
      break;
    default:
      fail("audit", "server_event_required");
  }
  if (a.operation === "recover_lead") {
    if (e.actor.kind !== "local_viewer") fail("actor", "local_viewer_required");
  } else if (
    e.kind === "producer_joined" ||
    e.kind === "foreground_transferred"
  )
    validateTimelineActor(e.actor, true);
  for (const key of Object.keys(a))
    if (!allowed.has(key)) fail(`audit.${key}`, "forbidden_for_operation");
}
function validateTimelineAuditReferences(
  s: TimelineSession,
  e: TimelineSession["events"][number],
): void {
  validateTimelineAuditEvent(e);
  const a = e.audit as TimelineAudit;
  for (const idValue of [a.previous_producer_id, a.next_producer_id])
    if (idValue && !s.producers.some((p) => p.producer_id === idValue))
      fail("audit.producer_id", "producer_not_accepted");
  if (
    a.plan_revision_id &&
    !s.events.some((e) =>
      e.document_snapshots?.some(
        (r) => r.document_revision_id === a.plan_revision_id,
      ),
    )
  )
    fail("plan_revision_id", "unknown_revision");
  if (a.target)
    validateTimelineDecisionReference(
      { ...s, events: s.events.filter((prev) => prev.sequence < e.sequence) },
      a.target,
    );
}

/** Pure mirror for a server-resolved, non-JSON canonical project key.
 * The T4/T5 server must resolve registered roots at create, join, accept/recover
 * and load (original source, accepted producers, pending requests). The browser
 * cannot establish this proof. Keep every actual Source tuple unchanged.
 */
export function validateTimelineResolvedProject(
  s: TimelineSession,
  value: TimelineSource,
  resolvedProjectKey: string,
): void {
  validateTimelineSource(value);
  if (!resolvedProjectKey) fail("source", "project_identity_unavailable");
  if (resolvedProjectKey !== s.project_key) fail("source", "foreign_project");
}

// Project scope is the work's server-resolved key; preserve first actual root.
function nativeSourceIdentity(value: TimelineSource): string {
  return JSON.stringify([value.host, value.native_session_id]);
}
