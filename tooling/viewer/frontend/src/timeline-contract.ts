import {
  isTimelineCatalogOperation,
  parseTimelineCatalogRequest,
  type TimelineCatalogRequestMap,
  timelineCatalogIdentities,
  validateTimelineCatalogs,
  validateTimelineDeliveryReference,
  validateTimelineIDs,
  validateTimelineTicketInput,
  validateTimelineTicketRecord,
  validateTimelineTicketReferences,
} from "./timeline-catalog";

export * from "./timeline-catalog";

import {
  isTimelineV3Operation,
  parseTimelineV3Request,
  resolveTimelineStageEventID,
  type TimelineActor,
  type TimelineAudit,
  type TimelineCIWait,
  type TimelineJoinRequest,
  type TimelineProducer,
  type TimelineTask,
  type TimelineV3RequestMap,
  timelineStateAllowsForeground,
  timelineStateRequiresForeground,
  validateTimelineActor,
  validateTimelineAppendV3,
  validateTimelineAuthority,
  validateTimelineSessionPatchActor,
  validateTimelineSnapshotReferencesV3,
  validateTimelineSnapshotV3,
  validateTimelineV3Graph,
} from "./timeline-contract-v3";

export * from "./timeline-contract-v3";

import type { RawSession } from "./api";

// This contract is independent of the legacy walkthrough manifest. External
// data must be parsed before it reaches typed timeline rendering.
export const TIMELINE_SCHEMA_VERSION = 3 as const;
export const TIMELINE_LIMITS = {
  body: 2 << 20,
  content: 7 << 18,
  events: 508,
  catalogRecords: 100,
  views: 50,
  images: 100,
  receipts: 1000,
  contentReceipts: 992,
  documentBytes: 256 << 10,
  documentTabs: 8,
  documents: 64,
  documentSnapshots: 50,
  aggregateDocumentBytes: 16 << 20,
  number: 2147483647,
} as const;
export type TimelineHost =
  | "codex"
  | "claude-code"
  | "grok"
  | "pi"
  | "opencode"
  | "cursor";
export type TimelineState =
  | "in_review"
  | "failed"
  | "active"
  | "awaiting_validation"
  | "paused"
  | "closed";
export type TimelineEventKind =
  | "review"
  | "state_transition"
  | "producer_joined"
  | "foreground_transferred"
  | "task_state_transition"
  | "local_action"
  | "request"
  | "work_started"
  | "stage"
  | "validation"
  | "changes_reported"
  | "iteration"
  | "approval_reported"
  | "delivery"
  | "closure";
export type TimelineSource = {
  project_root: string;
  host: TimelineHost;
  native_session_id: string;
};
export type TimelineTicketInput = {
  external_scope: string;
  external_id: string;
  key: string;
  title?: string;
  url?: string;
};
export type TimelineTicket = TimelineTicketInput & { ticket_id: string };
export type TimelineTicketSnapshot = {
  key: string;
  title?: string;
  url?: string;
};
export type TimelineTicketRef = {
  ticket_id: string;
  snapshot: TimelineTicketSnapshot;
};
export type TimelineCandidate = {
  branch: string;
  commit: string;
  base: string;
};
export type TimelineCandidateRecord = TimelineCandidate & {
  candidate_id: string;
};
export type TimelineStage =
  | "research"
  | "plan"
  | "build"
  | "verify"
  | "delivery";
export type TimelineStageState =
  | "in_review"
  | "awaiting_validation"
  | "in_progress"
  | "completed"
  | "paused"
  | "failed";
export type TimelineDecisionTarget =
  | {
      kind: "evidence";
      event_id: string;
      evidence_id: string;
      view_ids?: string[];
    }
  | { kind: "document"; event_id: string; document_revision_id: string };
export type TimelineEvidenceTarget = TimelineDecisionTarget;
export type TimelineDocument = {
  actor: TimelineActor;
  document_id: string;
  session_id: string;
  sha256: string;
  size: number;
  media_type: "text/markdown";
  received_at: string;
};
export type TimelineDocumentTab = { label: string; document_id: string };
export type TimelineDocumentSnapshot = {
  task_id?: string;
  segment_id?: string;
  document_key: string;
  predecessor_revision_id?: string;
  actor: TimelineActor;
  document_revision_id: string;
  status: "draft" | "complete";
  created_at: string;
  tabs: TimelineDocumentTab[];
};
export type TimelineDocumentResponse = {
  schema_version: 3;
  document: TimelineDocument;
  content: string;
};
export type TimelineDocumentMetadata = {
  actor: TimelineActor;
  schema_version: 3;
  operation_id: string;
  media_type: "text/markdown";
};
export type TimelineDocumentSnapshotRequest = {
  document_key: string;
  predecessor_revision_id?: string;
  expected_session_version: number;
  task_id?: string;
  segment_id?: string;
  actor: TimelineActor;
  schema_version: 3;
  operation_id: string;
  expected_version: number;
  status: "draft" | "complete";
  tabs: TimelineDocumentTab[];
};
export type TimelineSupersedes = {
  event_id: string;
  evidence_id: string;
  view_id: string;
};
export type TimelineImageRef = {
  artifact_id: string;
  caption: string;
  alt?: string;
};
export type TimelineView = {
  view_id: string;
  title: string;
  ticket_refs: TimelineTicketRef[];
  steps: string[];
  expected_result: string;
  image_refs: TimelineImageRef[];
  supersedes?: TimelineSupersedes;
};
export type TimelineViewInput = Omit<
  TimelineView,
  "view_id" | "ticket_refs"
> & { view_id?: string; ticket_ids: string[] };
export type TimelineDeliverySnapshot = {
  kind: "pull_request" | "commit";
  label: string;
  url: string;
  status: "open" | "merged" | "local";
};
export type TimelineDelivery = TimelineDeliverySnapshot & {
  delivery_id: string;
  repository_url: string;
  external_id: string;
};
export type TimelineDeliveryReference = {
  delivery_id: string;
  snapshot: TimelineDeliverySnapshot;
};
export type TimelineClosure = {
  reason: "completed" | "paused" | "abandoned";
  result: string;
};
export type TimelineEvent = {
  stage_event_id?: string;
  task_id?: string;
  segment_id?: string;
  reason?: string;
  review_status?: string;
  audit?: TimelineAudit;
  actor: TimelineActor;
  event_id: string;
  sequence: number;
  version: number;
  kind: TimelineEventKind;
  title: string;
  summary: string;
  created_at: string;
  updated_at: string;
  origin: "agent_reported" | "server";
  ticket_refs: TimelineTicketRef[];
  withdrawn?: boolean;
  candidate_id?: string;
  evidence_id?: string;
  views?: TimelineView[];
  related_event_id?: string;
  target?: TimelineEvidenceTarget;
  source_reference?: string;
  delivery_refs?: TimelineDeliveryReference[];
  closure?: TimelineClosure;
  stage?: TimelineStage;
  stage_state?: TimelineStageState;
  started_at?: string;
  finished_at?: string;
  occurred_at?: string;
  document_snapshots?: TimelineDocumentSnapshot[];
};
export type TimelineArtifact = {
  actor: TimelineActor;
  artifact_id: string;
  session_id: string;
  sha256: string;
  size: number;
  media_type: "image/png" | "image/jpeg" | "image/webp";
  received_at: string;
  captured_at: string;
  candidate_id: string;
  viewport: "desktop" | "mobile";
  theme: "light" | "dark" | "unspecified";
};
export type TimelineSession = {
  deliveries: TimelineDelivery[];
  candidates: TimelineCandidateRecord[];
  producers: TimelineProducer[];
  foreground_producer_id: string;
  pending_join_requests: TimelineJoinRequest[];
  tasks: TimelineTask[];
  waits: TimelineCIWait[];
  reason?: string;
  schema_version: 3;
  viewer_session_id: string;
  project_key: string;
  project_name: string;
  source: TimelineSource;
  version: number;
  title: string;
  summary: string;
  state: TimelineState;
  archived: boolean;
  pinned: boolean;
  created_at: string;
  updated_at: string;
  validation_count: number;
  tickets: TimelineTicket[];
  events: TimelineEvent[];
  artifacts: TimelineArtifact[];
  documents: TimelineDocument[];
  current_stage_event_id: string | null;
};
export type TimelineContentMode = "timeline_api" | "legacy";
export type TimelineAPIIndexEntry = Omit<
  TimelineSession,
  | "schema_version"
  | "events"
  | "artifacts"
  | "documents"
  | "producers"
  | "foreground_producer_id"
  | "pending_join_requests"
  | "tasks"
  | "waits"
  | "deliveries"
  | "candidates"
> & { mode: "timeline_api" };
export type TimelineLegacyIndexEntry = { mode: "legacy"; session: RawSession };
export type TimelineIndexEntry =
  | TimelineAPIIndexEntry
  | TimelineLegacyIndexEntry;
export type TimelineOperationKind =
  | "ticket_create"
  | "ticket_patch"
  | "delivery_create"
  | "delivery_patch"
  | "candidate_create"
  | "join_request"
  | "producer_accept"
  | "foreground_transfer"
  | "recover_lead"
  | "task_create"
  | "task_patch"
  | "task_transition"
  | "wait_create"
  | "wait_attempt_start"
  | "wait_observe"
  | "create_session"
  | "session_patch"
  | "session_delete"
  | "append_event"
  | "event_patch"
  | "event_withdraw"
  | "artifact_add"
  | "document_add"
  | "document_snapshot";
type TimelineOperationBase = {
  schema_version: 3;
  operation_id: string;
  viewer_session_id: string;
  session_version: number;
  current_stage_event_id: string | null;
};
type TimelineEventOperationFields = {
  event_id: string;
  event_version: number;
  sequence: number;
  evidence_id?: string;
  view_ids?: string[];
};
export type TimelineOperationResult = TimelineOperationBase &
  (
    | { kind: "ticket_create"; ticket_id: string; created: boolean }
    | { kind: "ticket_patch"; ticket_id: string }
    | { kind: "delivery_create"; delivery_id: string; created: boolean }
    | { kind: "delivery_patch"; delivery_id: string }
    | { kind: "candidate_create"; candidate_id: string; created: boolean }
    | { kind: "create_session"; created: boolean }
    | { kind: "session_patch" }
    | { kind: "join_request"; request_id: string }
    | {
        kind: "producer_accept" | "recover_lead";
        request_id: string;
        producer_id: string;
      }
    | { kind: "foreground_transfer"; producer_id: string }
    | {
        kind: "task_create" | "task_patch" | "task_transition";
        task_id: string;
        task_version: number;
      }
    | { kind: "wait_create"; wait_id: string; wait_version: number }
    | {
        kind: "wait_attempt_start";
        wait_id: string;
        wait_version: number;
        attempt_id: string;
        generation: number;
      }
    | {
        kind: "wait_observe";
        wait_id: string;
        wait_version: number;
        observation_id: string;
      }
    | { kind: "session_delete"; deleted: true }
    | ({ kind: "append_event" | "event_patch" } & TimelineEventOperationFields)
    | ({
        kind: "event_withdraw";
        withdrawn: true;
      } & TimelineEventOperationFields)
    | { kind: "artifact_add"; artifact_id: string; created: boolean }
    | {
        kind: "document_add";
        document_id: string;
        sha256: string;
        size: number;
        created: boolean;
      }
    | ({
        kind: "document_snapshot";
        document_revision_id: string;
      } & TimelineEventOperationFields)
  );
// Internal receipt contracts are exported for persistence/tests, never members
// of the public TimelineSession or TimelineSessionResponse data.
export type TimelineOperationReceipt = {
  operation_id: string;
  request_sha256: string;
  completed_at: string;
  result: TimelineOperationResult;
};
export type TimelineSessionResponse = {
  schema_version: 3;
  mode: "timeline_api";
  session: TimelineSession;
  artifact_urls?: Record<string, string>;
  document_urls?: Record<string, string>;
  created?: boolean;
};
export type TimelineCreateSessionRequest = {
  schema_version: 3;
  operation_id: string;
  source: TimelineSource;
  title: string;
  summary: string;
  tickets: TimelineTicketInput[];
};
export type TimelineMutationRequest = {
  actor: TimelineActor;
  schema_version: 3;
  operation_id: string;
  expected_version: number;
};
export type TimelineSessionPatchRequest = TimelineMutationRequest & {
  reason?: string;
} & Partial<
    Pick<TimelineSession, "title" | "summary" | "state" | "archived" | "pinned">
  >;
export type TimelineEventPatchRequest = TimelineMutationRequest &
  Partial<Pick<TimelineEvent, "title" | "summary" | "stage_state">> & {
    expected_session_version?: number;
    reason?: string;
  };
export type TimelineAppendEventRequest = {
  target_stage_event_id?: string;
  task_id?: string;
  segment_id?: string;
  reason?: string;
  review_status?: string;
  actor: TimelineActor;
  schema_version: 3;
  operation_id: string;
  kind: TimelineEventKind;
  title: string;
  summary: string;
  ticket_ids: string[];
  candidate_id?: string;
  views?: TimelineViewInput[];
  related_event_id?: string;
  target?: TimelineEvidenceTarget;
  source_reference?: string;
  delivery_ids?: string[];
  closure?: TimelineClosure;
  stage?: TimelineStage;
  stage_state?: "in_progress";
  expected_session_version?: number;
  occurred_at?: string;
};
export type TimelineArtifactMetadata = {
  actor: TimelineActor;
  schema_version: 3;
  operation_id: string;
  captured_at: string;
  candidate_id: string;
  viewport: "desktop" | "mobile";
  theme: "light" | "dark" | "unspecified";
  media_type: "image/png" | "image/jpeg" | "image/webp";
};
export type TimelineContractKind =
  | keyof TimelineV3RequestMap
  | keyof TimelineCatalogRequestMap
  | "create_session"
  | "session_patch"
  | "event_patch"
  | "mutation"
  | "append_event"
  | "artifact_metadata"
  | "session"
  | "index_entry"
  | "operation_result"
  | "operation_receipt"
  | "document_metadata"
  | "document_snapshot";
export class TimelineContractError extends Error {
  constructor(
    public field: string,
    public code: string,
  ) {
    super(`${field}: ${code}`);
  }
}
type ObjectValue = Record<string, unknown>;
const hosts = ["codex", "claude-code", "grok", "pi", "opencode", "cursor"];
const states = [
  "active",
  "in_review",
  "awaiting_validation",
  "paused",
  "closed",
  "failed",
];
const kinds = [
  "review",
  "state_transition",
  "producer_joined",
  "foreground_transferred",
  "task_state_transition",
  "local_action",
  "request",
  "work_started",
  "stage",
  "validation",
  "changes_reported",
  "iteration",
  "approval_reported",
  "delivery",
  "closure",
];
const opaqueID = /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/;
const operationID = /^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$/;
const sha256 = /^[0-9a-f]{64}$/;
function fail(field: string, code = "invalid_value"): never {
  throw new TimelineContractError(field, code);
}
function object(
  value: unknown,
  required: string[],
  optional: string[] = [],
  field = "body",
): ObjectValue {
  if (!value || typeof value !== "object" || Array.isArray(value))
    fail(field, "object_required");
  const obj = value as ObjectValue;
  for (const key of required)
    if (!Object.hasOwn(obj, key)) fail(key, "required_field");
  for (const key of Object.keys(obj))
    if (![...required, ...optional].includes(key)) fail(field, "unknown_field");
  for (const [key, item] of Object.entries(obj))
    if (item === null && key !== "current_stage_event_id")
      fail(field, "null_not_allowed");
  return obj;
}
function text(
  value: unknown,
  field: string,
  max: number,
  required = true,
): asserts value is string {
  if (
    typeof value !== "string" ||
    [...value].length > max ||
    (required && !value.trim()) ||
    [...value].some(
      (char) => /\p{Cc}/u.test(char) && char !== "\n" && char !== "\t",
    )
  )
    fail(field, "invalid_text");
}
function integer(
  value: unknown,
  field: string,
  min = 1,
  max: number = TIMELINE_LIMITS.number,
): asserts value is number {
  if (
    typeof value !== "number" ||
    !Number.isInteger(value) ||
    value < min ||
    value > max
  )
    fail(field, "invalid_positive_integer");
}
function id(value: unknown, field: string): asserts value is string {
  if (typeof value !== "string" || !opaqueID.test(value))
    fail(field, "invalid_id");
}
function array(value: unknown, field: string, max: number, min = 0): unknown[] {
  if (!Array.isArray(value) || value.length < min || value.length > max)
    fail(field, "invalid_array");
  return value;
}
function timestamp(value: unknown, field: string) {
  text(value, field, 64);
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d{1,9})?(Z|[+-]\d{2}:\d{2})$/.exec(
      value,
    );
  if (!match || !Number.isFinite(Date.parse(value)))
    fail(field, "invalid_timestamp");
  const [, y, m, d, h, minute, second, zone] = match;
  const date = new Date(0);
  date.setUTCFullYear(Number(y), Number(m) - 1, Number(d));
  if (
    date.getUTCFullYear() !== Number(y) ||
    date.getUTCMonth() !== Number(m) - 1 ||
    date.getUTCDate() !== Number(d) ||
    Number(h) > 23 ||
    Number(minute) > 59 ||
    Number(second) > 59 ||
    (zone !== "Z" &&
      (Number(zone.slice(1, 3)) > 23 || Number(zone.slice(4)) > 59))
  )
    fail(field, "invalid_timestamp");
}
export function isSafeTimelineURL(value: unknown): value is string {
  if (
    typeof value !== "string" ||
    [...value].length > 4096 ||
    /[\p{White_Space}\p{Cc}\\]/u.test(value) ||
    /%(?![0-9a-f]{2})/i.test(value)
  )
    return false;
  const authority =
    /^https:\/\/([A-Za-z0-9._-]+|\[[0-9A-Fa-f:.]+\])(:[0-9]+)?([/?#]|$)/i.exec(
      value,
    );
  if (!authority || (authority[2] && Number(authority[2].slice(1)) > 65535))
    return false;
  const rawHost = authority[1].toLowerCase();
  if (!rawHost.startsWith("[")) {
    const last = rawHost.replace(/\.$/, "").split(".").at(-1) ?? "";
    if (/^[0-9]+$/.test(last) || /^0x[0-9a-f]*$/.test(last)) {
      const octets = rawHost.split(".");
      if (
        octets.length !== 4 ||
        octets.some(
          (part) => !/^(0|[1-9][0-9]*)$/.test(part) || Number(part) > 255,
        )
      )
        return false;
    }
  }
  try {
    const parsed = new URL(value);
    return (
      parsed.protocol === "https:" &&
      !!parsed.hostname &&
      !parsed.username &&
      !parsed.password
    );
  } catch {
    return false;
  }
}
export function isSafeTimelineArtifactURL(
  value: unknown,
  sessionID: string,
): value is string {
  if (typeof value !== "string" || !sha256.test(sessionID)) return false;
  const prefix = `/files/${sessionID}/`;
  if (!value.startsWith(prefix)) return false;
  const [capability, ...segments] = value.slice(prefix.length).split("/");
  return (
    sha256.test(capability) &&
    segments.length > 0 &&
    segments.every(
      (part) => /^[A-Za-z0-9._-]+$/.test(part) && part !== "." && part !== "..",
    )
  );
}
export function validateTimelineSource(value: unknown) {
  const obj = object(value, ["project_root", "host", "native_session_id"]);
  text(obj.project_root, "source.project_root", 4096);
  if (
    !(obj.project_root as string).startsWith("/") ||
    (obj.project_root as string).includes("\0")
  )
    fail("source.project_root", "absolute_root_required");
  if (!hosts.includes(obj.host as string))
    fail("source.host", "unsupported_host");
  text(obj.native_session_id, "source.native_session_id", 256);
  if (/[\p{White_Space}\p{Cc}]/u.test(obj.native_session_id as string))
    fail("source.native_session_id", "native_id_required");
}
const source = validateTimelineSource;
export function timelineTicketIdentity(ref: { ticket_id: string }): string {
  return ref.ticket_id;
}
/** Checks the header fields a session and its index entry share. */
function summaryHeader(obj: ObjectValue) {
  if (
    typeof obj.viewer_session_id !== "string" ||
    !sha256.test(obj.viewer_session_id)
  )
    fail("viewer_session_id", "invalid_id");
  source(obj.source);
  text(obj.project_key, "project_key", 2048);
  text(obj.project_name, "project_name", 200);
  integer(obj.version, "version");
  text(obj.title, "title", 200);
  text(obj.summary, "summary", 4000);
  if (!states.includes(obj.state as string)) fail("state", "unsupported_state");
  for (const field of ["archived", "pinned"])
    if (typeof obj[field] !== "boolean") fail(field);
  timestamp(obj.created_at, "created_at");
  timestamp(obj.updated_at, "updated_at");
}
function tickets(value: unknown, full = false, bootstrap = false) {
  // A full ticket list over the limit gets the Go validator's code; event
  // ticket references keep the generic array error.
  if (
    full &&
    Array.isArray(value) &&
    value.length > TIMELINE_LIMITS.catalogRecords
  )
    fail("tickets", "array_required_or_limit");
  const list = array(value, "tickets", TIMELINE_LIMITS.catalogRecords);
  if (!full) {
    validateTimelineTicketReferences(list);
    return;
  }
  for (const item of list) {
    if (bootstrap) validateTimelineTicketInput(item);
    else validateTimelineTicketRecord(item);
  }
}
function target(value: unknown) {
  if (!value || typeof value !== "object" || Array.isArray(value))
    fail("target", "object_required");
  const kind = (value as ObjectValue).kind;
  if (kind === "document") {
    const obj = object(value, ["kind", "event_id", "document_revision_id"]);
    id(obj.event_id, "target.event_id");
    id(obj.document_revision_id, "target.document_revision_id");
    return;
  }
  const obj = object(value, ["kind", "event_id", "evidence_id"], ["view_ids"]);
  if (kind !== "evidence") fail("target.kind", "unsupported_target");
  id(obj.event_id, "target.event_id");
  id(obj.evidence_id, "target.evidence_id");
  if (obj.view_ids !== undefined) {
    const seen = new Set<string>();
    for (const item of array(obj.view_ids, "target.view_ids", 50)) {
      id(item, "target.view_ids");
      if (seen.has(item)) fail("target.view_ids", "duplicate_id");
      seen.add(item);
    }
  }
}
function view(value: unknown, output = false) {
  const obj = object(
    value,
    [
      "title",
      output ? "ticket_refs" : "ticket_ids",
      "steps",
      "expected_result",
      "image_refs",
      ...(output ? ["view_id"] : []),
    ],
    ["supersedes", ...(!output ? ["view_id"] : [])],
  );
  if (obj.view_id !== undefined) id(obj.view_id, "views.view_id");
  text(obj.title, "views.title", 200);
  if (output) tickets(obj.ticket_refs);
  else validateTimelineIDs(obj.ticket_ids, "views.ticket_ids");
  for (const step of array(obj.steps, "views.steps", 100, 1))
    text(step, "views.steps", 2000);
  text(obj.expected_result, "views.expected_result", 4000);
  const seen = new Set<string>();
  for (const item of array(obj.image_refs, "views.image_refs", 100, 1)) {
    const image = object(item, ["artifact_id", "caption"], ["alt"]);
    id(image.artifact_id, "views.image_refs.artifact_id");
    text(image.caption, "views.image_refs.caption", 1000);
    if (image.alt !== undefined) text(image.alt, "views.image_refs.alt", 2000);
    if (seen.has(image.artifact_id))
      fail("views.image_refs", "duplicate_artifact");
    seen.add(image.artifact_id);
  }
  if (obj.supersedes !== undefined) {
    const prev = object(obj.supersedes, ["event_id", "evidence_id", "view_id"]);
    id(prev.event_id, "supersedes.event_id");
    id(prev.evidence_id, "supersedes.evidence_id");
    id(prev.view_id, "supersedes.view_id");
    if (obj.view_id !== undefined && obj.view_id !== prev.view_id)
      fail("views.supersedes", "view_mismatch");
  }
}
function operation(obj: ObjectValue, version = false) {
  if (obj.schema_version !== 3) fail("schema_version", "unsupported_protocol");
  if (
    typeof obj.operation_id !== "string" ||
    !operationID.test(obj.operation_id)
  )
    fail("operation_id", "invalid_id");
  if (version) integer(obj.expected_version, "expected_version");
}
function eventPayload(obj: ObjectValue, output = false) {
  if (obj.kind === "review") {
    if (
      !["in_progress", "completed", "failed"].includes(
        obj.review_status as string,
      )
    )
      fail("review_status", "invalid_state");
    if (output) id(obj.stage_event_id, "stage_event_id");
  } else if (obj.review_status !== undefined)
    fail("review_status", "forbidden_for_kind");
  if (!kinds.includes(obj.kind as string)) fail("kind", "unsupported_kind");
  text(obj.title, "title", 200, obj.kind !== "stage");
  text(obj.summary, "summary", 4000);
  if (output) tickets(obj.ticket_refs);
  else validateTimelineIDs(obj.ticket_ids, "ticket_ids");
  if (obj.candidate_id !== undefined) id(obj.candidate_id, "candidate_id");
  if (obj.related_event_id !== undefined)
    id(obj.related_event_id, "related_event_id");
  if (obj.occurred_at !== undefined) timestamp(obj.occurred_at, "occurred_at");
  if (obj.kind === "stage") {
    if (!stages.includes(obj.stage as string)) fail("stage");
    if (!stageStates.includes(obj.stage_state as string)) fail("stage_state");
    if (!output) {
      if (obj.stage_state !== "in_progress") fail("stage_state");
      integer(obj.expected_session_version, "expected_session_version");
    }
  } else if (
    obj.stage !== undefined ||
    obj.stage_state !== undefined ||
    obj.expected_session_version !== undefined
  )
    fail("stage", "forbidden_for_kind");
  const evidence = obj.kind === "validation" || obj.kind === "iteration";
  const decision =
    obj.kind === "changes_reported" || obj.kind === "approval_reported";
  if (evidence) {
    const views = array(obj.views, "views", 50, 1);
    const seen = new Set<string>();
    let hasPrior = false;
    for (const item of views) {
      view(item, output);
      const v = item as ObjectValue;
      if (v.view_id !== undefined) {
        if (seen.has(v.view_id as string))
          fail("views.view_id", "duplicate_id");
        seen.add(v.view_id as string);
      }
      hasPrior ||= v.supersedes !== undefined;
    }
    if (obj.kind === "iteration" && !hasPrior)
      fail("views.supersedes", "iteration_reference_required");
  } else if (obj.views !== undefined) fail("views", "forbidden_for_kind");
  if (decision) {
    target(obj.target);
    text(obj.source_reference, "source_reference", 1000);
  } else if (obj.target !== undefined || obj.source_reference !== undefined)
    fail("target", "forbidden_for_kind");
  if (obj.kind === "delivery") {
    if (output) {
      const refs = array(obj.delivery_refs, "delivery_refs", 100, 1);
      const seen = new Set<string>();
      for (const ref of refs) {
        validateTimelineDeliveryReference(ref);
        const id = (ref as TimelineDeliveryReference).delivery_id;
        if (seen.has(id)) fail("delivery_refs", "duplicate_id");
        seen.add(id);
      }
    } else validateTimelineIDs(obj.delivery_ids, "delivery_ids", true);
  } else if (obj.delivery_ids !== undefined || obj.delivery_refs !== undefined)
    fail("delivery_ids", "forbidden_for_kind");
  if (obj.kind === "closure") {
    const c = object(obj.closure, ["reason", "result"]);
    if (!["completed", "paused", "abandoned"].includes(c.reason as string))
      fail("closure", "invalid_reason");
    text(c.result, "closure.result", 4000);
  } else if (obj.closure !== undefined) fail("closure", "forbidden_for_kind");
}
const eventRequired = ["kind", "title", "summary", "ticket_ids"];
const eventOptional = [
  "target_stage_event_id",
  "task_id",
  "segment_id",
  "reason",
  "review_status",
  "candidate_id",
  "views",
  "related_event_id",
  "target",
  "source_reference",
  "delivery_ids",
  "closure",
  "stage",
  "stage_state",
  "expected_session_version",
  "occurred_at",
];
export type TimelineRequestMap = TimelineV3RequestMap &
  TimelineCatalogRequestMap & {
    create_session: TimelineCreateSessionRequest;
    session_patch: TimelineSessionPatchRequest;
    event_patch: TimelineEventPatchRequest;
    mutation: TimelineMutationRequest;
    append_event: TimelineAppendEventRequest;
    artifact_metadata: TimelineArtifactMetadata;
    document_metadata: TimelineDocumentMetadata;
    document_snapshot: TimelineDocumentSnapshotRequest;
  };
export function parseTimelineRequest<K extends keyof TimelineRequestMap>(
  kind: K,
  value: unknown,
): TimelineRequestMap[K] {
  if (
    value &&
    typeof value === "object" &&
    "schema_version" in value &&
    value.schema_version !== 3
  )
    fail("schema_version", "unsupported_protocol");
  if (isTimelineCatalogOperation(kind))
    return parseTimelineCatalogRequest(kind, value) as TimelineRequestMap[K];
  if (isTimelineV3Operation(kind))
    return parseTimelineV3Request(
      kind,
      value,
    ) as unknown as TimelineRequestMap[K];
  let obj: ObjectValue = {};
  switch (kind) {
    case "create_session":
      obj = object(value, [
        "schema_version",
        "operation_id",
        "source",
        "title",
        "summary",
        "tickets",
      ]);
      operation(obj);
      source(obj.source);
      text(obj.title, "title", 200);
      text(obj.summary, "summary", 4000);
      tickets(obj.tickets, true, true);
      break;
    case "session_patch": {
      const fields = [
        "title",
        "summary",
        "state",
        "archived",
        "pinned",
        "reason",
      ];
      obj = object(
        value,
        ["schema_version", "operation_id", "expected_version", "actor"],
        fields,
      );
      operation(obj, true);
      if (!fields.some((field) => Object.hasOwn(obj, field)))
        fail("body", "empty_patch");
      if (obj.title !== undefined) text(obj.title, "title", 200);
      if (obj.summary !== undefined) text(obj.summary, "summary", 4000);
      if (obj.state !== undefined && !states.includes(obj.state as string))
        fail("state", "unsupported_state");
      for (const field of ["archived", "pinned"])
        if (obj[field] !== undefined && typeof obj[field] !== "boolean")
          fail(field);
      break;
    }
    case "event_patch":
      obj = object(
        value,
        ["schema_version", "operation_id", "expected_version", "actor"],
        [
          "title",
          "summary",
          "stage_state",
          "expected_session_version",
          "reason",
        ],
      );
      operation(obj, true);
      if (
        obj.title === undefined &&
        obj.summary === undefined &&
        obj.stage_state === undefined
      )
        fail("body", "empty_patch");
      if (obj.title !== undefined) text(obj.title, "title", 200, false);
      if (obj.summary !== undefined) text(obj.summary, "summary", 4000);
      if (obj.stage_state !== undefined) {
        text(obj.reason, "reason", 500);
        if (!stageStates.includes(obj.stage_state as string))
          fail("stage_state");
        integer(obj.expected_session_version, "expected_session_version");
      } else if (obj.expected_session_version !== undefined)
        fail("expected_session_version", "forbidden_for_kind");
      break;
    case "document_metadata":
      obj = object(value, [
        "schema_version",
        "operation_id",
        "media_type",
        "actor",
      ]);
      operation(obj);
      if (obj.media_type !== "text/markdown") fail("media_type");
      break;
    case "document_snapshot":
      obj = object(
        value,
        [
          "schema_version",
          "operation_id",
          "expected_version",
          "actor",
          "document_key",
          "expected_session_version",
          "status",
          "tabs",
        ],
        ["predecessor_revision_id", "task_id", "segment_id"],
      );
      operation(obj, true);
      snapshotFields(obj);
      break;
    case "mutation":
      obj = object(value, [
        "schema_version",
        "operation_id",
        "expected_version",
        "actor",
      ]);
      operation(obj, true);
      break;
    case "append_event":
      obj = object(
        value,
        ["schema_version", "operation_id", "actor", ...eventRequired],
        eventOptional,
      );
      operation(obj);
      eventPayload(obj);
      break;
    case "artifact_metadata":
      obj = object(value, [
        "schema_version",
        "operation_id",
        "actor",
        "captured_at",
        "candidate_id",
        "viewport",
        "theme",
        "media_type",
      ]);
      operation(obj);
      timestamp(obj.captured_at, "captured_at");
      id(obj.candidate_id, "candidate_id");
      if (!["desktop", "mobile"].includes(obj.viewport as string))
        fail("viewport");
      if (!["light", "dark", "unspecified"].includes(obj.theme as string))
        fail("theme");
      if (
        !["image/png", "image/jpeg", "image/webp"].includes(
          obj.media_type as string,
        )
      )
        fail("media_type");
      break;
  }
  if (kind !== "create_session")
    validateTimelineActor(
      obj.actor,
      kind !== "session_patch" && kind !== "mutation",
    );
  if (kind === "session_patch")
    validateTimelineSessionPatchActor(obj as TimelineSessionPatchRequest);
  if (kind === "append_event")
    validateTimelineAppendV3(obj as unknown as TimelineAppendEventRequest);
  if (kind === "document_snapshot")
    validateTimelineSnapshotV3(obj as TimelineDocumentSnapshotRequest);
  return obj as unknown as TimelineRequestMap[K];
}

/** Checks references against the current session, under the caller's mutation boundary. */
export function validateTimelineEventReferences(
  session: TimelineSession,
  request: TimelineAppendEventRequest,
): void {
  resolveTimelineStageEventID(session, request);
  parseTimelineRequest("append_event", request);
  if (session.state === "closed") fail("state", "session_closed");
  if (request.kind === "closure" && session.current_stage_event_id !== null)
    fail("current_stage_event_id", "stage_active");
  if (request.kind === "stage") {
    if (request.expected_session_version !== session.version)
      fail("expected_session_version", "version_conflict");
    const latest = session.events
      .filter(
        (e) => e.kind === "stage" && e.stage === request.stage && !e.withdrawn,
      )
      .at(-1);
    if (latest) {
      if (
        latest.stage_state === "in_progress" ||
        latest.stage_state === "paused"
      )
        fail("stage", "stage_resume_required");
      if (request.related_event_id !== latest.event_id)
        fail("related_event_id", "stage_reentry_reference_required");
    }
  }
  checkReferences(session, request, false);
}
export function validateTimelineDecisionReference(
  s: TimelineSession,
  t: TimelineDecisionTarget,
): void {
  target(t);
  checkReferences(
    s,
    { ticket_ids: [], target: t } as unknown as TimelineAppendEventRequest,
    false,
  );
}
function checkReferences(
  session: TimelineSession,
  request: TimelineAppendEventRequest,
  restored: boolean,
) {
  const knownTickets = new Set(session.tickets.map(timelineTicketIdentity));
  const checkTickets = (refs: string[]) => {
    for (const ref of refs)
      if (!knownTickets.has(ref)) fail("ticket_refs", "unknown_ticket");
  };
  checkTickets(request.ticket_ids);
  if (request.candidate_id) {
    if (
      !session.candidates.some((c) => c.candidate_id === request.candidate_id)
    )
      fail("candidate_id", "unknown_candidate");
  }
  for (const id of request.delivery_ids ?? []) {
    if (!session.deliveries.some((d) => d.delivery_id === id))
      fail("delivery_ids", "unknown_delivery");
  }
  const events = new Map(
    session.events.map((event) => [event.event_id, event]),
  );
  const knownViews = new Set(
    session.events.flatMap((event) =>
      (event.views ?? []).map((v) => v.view_id),
    ),
  );
  const knownArtifacts = new Set(
    session.artifacts
      .filter((a) => a.session_id === session.viewer_session_id)
      .map((a) => a.artifact_id),
  );
  if (restored)
    for (const v of request.views ?? [])
      if (v.view_id) knownViews.add(v.view_id);
  const checkTarget = (t: TimelineEvidenceTarget) => {
    const previous = events.get(t.event_id);
    if (t.kind === "document") {
      if (
        previous?.kind !== "stage" ||
        (!restored && previous.withdrawn) ||
        !previous.document_snapshots?.some(
          (snapshot) =>
            snapshot.document_revision_id === t.document_revision_id,
        )
      )
        fail("target", "unknown_document_revision");
      return;
    }
    if (
      !previous?.evidence_id ||
      previous.evidence_id !== t.evidence_id ||
      (!restored && previous.withdrawn)
    )
      fail("target", "unknown_evidence");
    for (const viewID of t.view_ids ?? [])
      if (!previous.views?.some((v) => v.view_id === viewID))
        fail("target.view_ids", "unknown_view");
  };
  if (request.related_event_id && !events.has(request.related_event_id))
    fail("related_event_id", "unknown_event");
  if (request.target) checkTarget(request.target);
  for (const v of request.views ?? []) {
    checkTickets(v.ticket_ids);
    if (v.view_id && !knownViews.has(v.view_id))
      fail("views.view_id", "unknown_view");
    for (const image of v.image_refs)
      if (!knownArtifacts.has(image.artifact_id))
        fail("views.image_refs", "unknown_artifact");
    if (v.supersedes)
      checkTarget({
        kind: "evidence",
        event_id: v.supersedes.event_id,
        evidence_id: v.supersedes.evidence_id,
        view_ids: [v.supersedes.view_id],
      });
  }
}
export function parseTimelineSession(value: unknown): TimelineSession {
  if (
    value &&
    typeof value === "object" &&
    "schema_version" in value &&
    value.schema_version !== 3
  )
    fail("schema_version", "unsupported_protocol");
  const obj = object(
    value,
    [
      "schema_version",
      "producers",
      "foreground_producer_id",
      "pending_join_requests",
      "tasks",
      "waits",
      "deliveries",
      "candidates",
      "viewer_session_id",
      "project_key",
      "project_name",
      "source",
      "version",
      "title",
      "summary",
      "state",
      "archived",
      "pinned",
      "created_at",
      "updated_at",
      "validation_count",
      "tickets",
      "events",
      "artifacts",
      "documents",
      "current_stage_event_id",
    ],
    ["reason"],
  );
  if (obj.schema_version !== 3) fail("schema_version", "unsupported_protocol");
  summaryHeader(obj);
  integer(obj.validation_count, "validation_count", 0, TIMELINE_LIMITS.events);
  tickets(obj.tickets, true);
  validateTimelineCatalogs(obj as unknown as TimelineSession);
  validateTimelineV3Graph(obj as unknown as TimelineSession);
  const artifacts = array(obj.artifacts, "artifacts", TIMELINE_LIMITS.body);
  const seenArtifacts = new Set<string>();
  for (const value of artifacts) {
    const a = object(value, [
      "artifact_id",
      "actor",
      "session_id",
      "sha256",
      "size",
      "media_type",
      "received_at",
      "captured_at",
      "candidate_id",
      "viewport",
      "theme",
    ]);
    id(a.artifact_id, "artifacts.artifact_id");
    if (
      a.session_id !== obj.viewer_session_id ||
      seenArtifacts.has(a.artifact_id)
    )
      fail("artifacts", "wrong_owner_or_duplicate");
    seenArtifacts.add(a.artifact_id);
    if (typeof a.sha256 !== "string" || !sha256.test(a.sha256))
      fail("artifacts.sha256");
    integer(a.size, "artifacts.size", 1, 20 << 20);
    timestamp(a.received_at, "artifacts.received_at");
    parseTimelineRequest("artifact_metadata", {
      schema_version: 3,
      operation_id: "restore",
      actor: a.actor,
      captured_at: a.captured_at,
      candidate_id: a.candidate_id,
      viewport: a.viewport,
      theme: a.theme,
      media_type: a.media_type,
    });
  }
  const events = array(obj.events, "events", TIMELINE_LIMITS.events);
  const ids = new Set<string>();
  const evidenceIDs = new Set<string>();
  let sequence = 0;
  let count = 0;
  const session = obj as unknown as TimelineSession;
  const prefix: TimelineSession = { ...session, events: [] };
  for (const value of events) {
    const e = object(
      value,
      [
        "event_id",
        "sequence",
        "version",
        "created_at",
        "updated_at",
        "origin",
        "actor",
        ...eventRequired.map((f) => (f === "ticket_ids" ? "ticket_refs" : f)),
      ],
      [
        "withdrawn",
        "stage_event_id",
        "audit",
        "evidence_id",
        "started_at",
        "finished_at",
        "document_snapshots",
        ...eventOptional
          .filter((f) => f !== "expected_session_version")
          .map((f) => (f === "delivery_ids" ? "delivery_refs" : f)),
      ],
    );
    id(e.event_id, "events.event_id");
    if (ids.has(e.event_id)) fail("events.event_id", "duplicate_id");
    ids.add(e.event_id);
    integer(e.sequence, "events.sequence");
    if (e.sequence <= sequence) fail("events.sequence", "not_increasing");
    sequence = e.sequence;
    integer(e.version, "events.version");
    timestamp(e.created_at, "events.created_at");
    timestamp(e.updated_at, "events.updated_at");
    if (e.origin !== "agent_reported" && e.origin !== "server")
      fail("events.origin", "invalid_origin");
    if (e.withdrawn !== undefined && typeof e.withdrawn !== "boolean")
      fail("events.withdrawn");
    eventPayload(e, true);
    validateStageOutput(e);
    if (e.kind === "validation" || e.kind === "iteration") {
      id(e.evidence_id, "events.evidence_id");
      if (evidenceIDs.has(e.evidence_id))
        fail("events.evidence_id", "duplicate_id");
      evidenceIDs.add(e.evidence_id);
      if (!e.withdrawn) count++;
    } else if (e.evidence_id !== undefined)
      fail("events.evidence_id", "forbidden_for_kind");
    const topTickets = new Set(
      (e.ticket_refs as TimelineTicketRef[]).map(timelineTicketIdentity),
    );
    for (const v of (e.views ?? []) as TimelineView[])
      for (const ref of v.ticket_refs)
        if (!topTickets.has(timelineTicketIdentity(ref)))
          fail("events.ticket_refs", "missing_view_ticket");
    checkReferences(
      prefix,
      {
        ...e,
        ticket_ids: (e.ticket_refs as TimelineTicketRef[]).map(
          (r) => r.ticket_id,
        ),
        delivery_ids: (
          e.delivery_refs as TimelineDeliveryReference[] | undefined
        )?.map((r) => r.delivery_id),
        views: ((e.views ?? []) as TimelineView[]).map((v) => ({
          ...v,
          ticket_ids: v.ticket_refs.map((r) => r.ticket_id),
        })),
      } as unknown as TimelineAppendEventRequest,
      true,
    );
    prefix.events.push(e as unknown as TimelineEvent);
  }
  if (obj.validation_count !== count)
    fail("validation_count", "count_mismatch");
  validateR3Session(session);
  return session;
}
export function parseTimelineSessionResponse(
  value: unknown,
): TimelineSessionResponse {
  const obj = object(
    value,
    ["schema_version", "mode", "session"],
    ["artifact_urls", "document_urls", "created"],
  );
  if (obj.schema_version !== 3 || obj.mode !== "timeline_api")
    fail("mode", "unsupported_version_or_mode");
  const session = parseTimelineSession(obj.session);
  if (obj.created !== undefined && typeof obj.created !== "boolean")
    fail("created");
  if (obj.artifact_urls !== undefined) {
    if (
      !obj.artifact_urls ||
      typeof obj.artifact_urls !== "object" ||
      Array.isArray(obj.artifact_urls)
    )
      fail("artifact_urls");
    const ids = new Set(session.artifacts.map((a) => a.artifact_id));
    for (const [key, value] of Object.entries(obj.artifact_urls))
      if (
        !ids.has(key) ||
        !isSafeTimelineArtifactURL(value, session.viewer_session_id)
      )
        fail("artifact_urls", "invalid_artifact_url");
  }
  if (obj.document_urls !== undefined) {
    const urls = object(
      obj.document_urls,
      [],
      session.documents.map((d) => d.document_id),
    );
    for (const [key, url] of Object.entries(urls))
      if (!isSafeTimelineDocumentURL(url, session.viewer_session_id, key))
        fail("document_urls", "invalid_document_url");
  }
  return obj as unknown as TimelineSessionResponse;
}

// Parse strings before JSON.parse loses duplicate keys. The scanner only tracks
// grammar/key identity; the native parser still handles JSON lexical validity.
export function parseTimelineJSON(raw: string): unknown {
  if (!raw || new TextEncoder().encode(raw).length > TIMELINE_LIMITS.body)
    fail("body", "invalid_size_or_encoding");
  let index = 0;
  const space = () => {
    while (/[\t\r\n ]/.test(raw[index] ?? "") && index < raw.length) index++;
  };
  const string = () => {
    const start = index++;
    while (index < raw.length) {
      const char = raw[index++];
      if (char === "\\") index++;
      else if (char === '"')
        return JSON.parse(raw.slice(start, index)) as string;
    }
    return fail("body", "invalid_json");
  };
  const visit = (depth: number, nullable = false): void => {
    if (depth > 32) fail("body", "too_deep");
    space();
    if (raw[index] === "{") {
      index++;
      space();
      const seen = new Set<string>();
      if (raw[index] === "}") {
        index++;
        return;
      }
      while (index < raw.length) {
        space();
        if (raw[index] !== '"') fail("body", "invalid_json");
        const key = string();
        if (seen.has(key)) fail("body", "duplicate_key");
        seen.add(key);
        space();
        if (raw[index++] !== ":") fail("body", "invalid_json");
        visit(depth + 1, key === "current_stage_event_id");
        space();
        const end = raw[index++];
        if (end === "}") return;
        if (end !== ",") fail("body", "invalid_json");
      }
      fail("body", "invalid_json");
    } else if (raw[index] === "[") {
      index++;
      space();
      if (raw[index] === "]") {
        index++;
        return;
      }
      while (index < raw.length) {
        visit(depth + 1);
        space();
        const end = raw[index++];
        if (end === "]") return;
        if (end !== ",") fail("body", "invalid_json");
      }
      fail("body", "invalid_json");
    } else if (raw[index] === '"') string();
    else {
      const start = index;
      while (index < raw.length && !/[,\]}\t\r\n ]/.test(raw[index])) index++;
      if (index === start) fail("body", "invalid_json");
      if (
        /^[0-9-]/.test(raw.slice(start, index)) &&
        !/^-?(0|[1-9][0-9]*)$/.test(raw.slice(start, index))
      )
        fail("body", "invalid_integer_literal");
      if (raw.slice(start, index) === "null" && !nullable)
        fail("body", "null_not_allowed");
    }
  };
  try {
    visit(0);
    space();
    if (index !== raw.length) fail("body", "trailing_data");
    return JSON.parse(raw);
  } catch (error) {
    if (error instanceof TimelineContractError) throw error;
    return fail("body", "invalid_json");
  }
}
export function decodeTimelineContract(
  kind: TimelineContractKind,
  raw: string,
): unknown {
  const value = parseTimelineJSON(raw);
  switch (kind) {
    case "session":
      return parseTimelineSession(value);
    case "index_entry":
      return parseTimelineIndexEntry(value);
    case "operation_result":
      return parseTimelineOperationResult(value);
    case "operation_receipt":
      return parseTimelineOperationReceipt(value);
    default:
      return parseTimelineRequest(kind, value);
  }
}

export function parseTimelineIndexEntry(value: unknown): TimelineIndexEntry {
  if (!value || typeof value !== "object" || Array.isArray(value))
    fail("body", "object_required");
  if ((value as ObjectValue).mode === "legacy")
    fail("mode", "unsupported_protocol");
  if ((value as ObjectValue).mode !== "timeline_api")
    fail("mode", "unsupported_mode");
  const entry = object(value, [
    "mode",
    "viewer_session_id",
    "project_key",
    "project_name",
    "source",
    "version",
    "title",
    "summary",
    "state",
    "archived",
    "pinned",
    "created_at",
    "updated_at",
    "validation_count",
    "tickets",
    "current_stage_event_id",
  ]);
  // The order matches the Go validator, so an entry with several faults
  // reports the same first code in both languages.
  if (
    timelineStateRequiresForeground(entry.state) &&
    entry.current_stage_event_id === null
  )
    fail("current_stage_event_id", "foreground_required");
  summaryHeader(entry);
  tickets(entry.tickets, true);
  timelineCatalogIdentities().addTickets(
    entry.tickets as TimelineSession["tickets"],
  );
  // An integer out of range gets the Go validator's code; other values keep
  // the generic integer error.
  const count = entry.validation_count;
  if (typeof count === "number" && Number.isInteger(count)) {
    if (count < 0 || count > TIMELINE_LIMITS.events)
      fail("validation_count", "invalid_count");
  } else integer(count, "validation_count", 0, TIMELINE_LIMITS.events);
  // The summary carries the stage pointer but not the event it points to.
  if (entry.current_stage_event_id !== null) {
    id(entry.current_stage_event_id, "current_stage_event_id");
    if (!timelineStateAllowsForeground(entry.state))
      fail("state", "foreground_requires_active_session");
  }
  return entry as TimelineAPIIndexEntry;
}
const v3ResultFields: Record<string, string[]> = {
  ticket_create: ["ticket_id", "created"],
  ticket_patch: ["ticket_id"],
  delivery_create: ["delivery_id", "created"],
  delivery_patch: ["delivery_id"],
  candidate_create: ["candidate_id", "created"],
  join_request: ["request_id"],
  producer_accept: ["request_id", "producer_id"],
  foreground_transfer: ["producer_id"],
  recover_lead: ["request_id", "producer_id"],
  task_create: ["task_id", "task_version"],
  task_patch: ["task_id", "task_version"],
  task_transition: ["task_id", "task_version"],
  wait_create: ["wait_id", "wait_version"],
  wait_attempt_start: ["wait_id", "wait_version", "attempt_id", "generation"],
  wait_observe: ["wait_id", "wait_version", "observation_id"],
};
export function parseTimelineOperationResult(
  value: unknown,
): TimelineOperationResult {
  const base = [
    "schema_version",
    "operation_id",
    "kind",
    "viewer_session_id",
    "session_version",
    "current_stage_event_id",
  ];
  if (!value || typeof value !== "object" || Array.isArray(value))
    fail("body", "object_required");
  const kind = (value as ObjectValue).kind;
  const eventFields = ["event_id", "event_version", "sequence"];
  let required = [...base];
  let optional: string[] = [];
  switch (kind) {
    case "create_session":
      required.push("created");
      break;
    case "session_patch":
      break;
    case "session_delete":
      required.push("deleted");
      break;
    case "append_event":
    case "event_patch":
      required = [...required, ...eventFields];
      optional = ["evidence_id", "view_ids"];
      break;
    case "event_withdraw":
      required = [...required, ...eventFields, "withdrawn"];
      optional = ["evidence_id", "view_ids"];
      break;
    case "document_add":
      required.push("document_id", "sha256", "size", "created");
      break;
    case "document_snapshot":
      required.push(...eventFields, "document_revision_id");
      break;
    case "artifact_add":
      required.push("artifact_id", "created");
      break;
    default:
      if (typeof kind !== "string" || !Object.hasOwn(v3ResultFields, kind))
        fail("kind", "unsupported_operation");
      required.push(...v3ResultFields[kind]);
  }
  const result = object(value, required, optional);
  operation(result);
  integer(result.session_version, "session_version");
  if (result.current_stage_event_id !== null)
    id(result.current_stage_event_id, "current_stage_event_id");
  if (
    typeof result.viewer_session_id !== "string" ||
    !sha256.test(result.viewer_session_id)
  )
    fail("viewer_session_id", "invalid_id");
  if (result.created !== undefined && typeof result.created !== "boolean")
    fail("created");
  if (kind === "session_delete" && result.deleted !== true)
    fail("deleted", "completed_delete_required");
  if (eventFields.every((field) => Object.hasOwn(result, field))) {
    id(result.event_id, "event_id");
    integer(result.event_version, "event_version");
    integer(result.sequence, "sequence");
  }
  if (kind === "event_withdraw" && result.withdrawn !== true)
    fail("withdrawn", "completed_withdrawal_required");
  if (result.evidence_id !== undefined) id(result.evidence_id, "evidence_id");
  if (result.view_ids !== undefined) {
    if (result.evidence_id === undefined)
      fail("view_ids", "invalid_evidence_scope");
    const seen = new Set<string>();
    for (const viewID of array(result.view_ids, "view_ids", 50)) {
      id(viewID, "view_ids");
      if (seen.has(viewID)) fail("view_ids", "duplicate_id");
      seen.add(viewID);
    }
  }
  for (const field of ["ticket_id", "delivery_id", "candidate_id"]) {
    if (result[field] !== undefined) id(result[field], field);
  }
  if (kind === "artifact_add") id(result.artifact_id, "artifact_id");
  if (kind === "document_add") {
    id(result.document_id, "document_id");
    if (typeof result.sha256 !== "string" || !sha256.test(result.sha256))
      fail("sha256");
    integer(result.size, "size", 0, TIMELINE_LIMITS.documentBytes);
  }
  if (kind === "document_snapshot")
    id(result.document_revision_id, "document_revision_id");
  if (typeof kind === "string" && Object.hasOwn(v3ResultFields, kind))
    for (const field of v3ResultFields[kind]) {
      if (field.endsWith("_version") || field === "generation")
        integer(result[field], field);
      else if (field !== "created") id(result[field], field);
    }
  return result as TimelineOperationResult;
}
export function parseTimelineOperationReceipt(
  value: unknown,
): TimelineOperationReceipt {
  const receipt = object(value, [
    "operation_id",
    "request_sha256",
    "completed_at",
    "result",
  ]);
  if (
    typeof receipt.operation_id !== "string" ||
    !operationID.test(receipt.operation_id)
  )
    fail("operation_id", "invalid_id");
  if (
    typeof receipt.request_sha256 !== "string" ||
    !sha256.test(receipt.request_sha256)
  )
    fail("request_sha256", "invalid_digest");
  timestamp(receipt.completed_at, "completed_at");
  const result = parseTimelineOperationResult(receipt.result);
  if (result.operation_id !== receipt.operation_id)
    fail("result.operation_id", "operation_mismatch");
  return receipt as TimelineOperationReceipt;
}
export function validateTimelineOperationRetry(
  receipt: TimelineOperationReceipt,
  operation: string,
  digest: string,
): TimelineOperationResult {
  parseTimelineOperationReceipt(receipt);
  if (receipt.operation_id !== operation || receipt.request_sha256 !== digest)
    fail("operation_id", "operation_conflict");
  return receipt.result;
}

const stages = ["research", "plan", "build", "verify", "delivery"];
const stageStates = [
  "in_progress",
  "in_review",
  "awaiting_validation",
  "completed",
  "paused",
  "failed",
];
function snapshotFields(s: ObjectValue) {
  if (!["draft", "complete"].includes(s.status as string)) fail("status");
  const labels = new Set<string>();
  for (const value of array(s.tabs, "tabs", TIMELINE_LIMITS.documentTabs, 1)) {
    const t = object(value, ["label", "document_id"]);
    text(t.label, "tabs.label", 100);
    id(t.document_id, "tabs.document_id");
    if (labels.has(t.label as string)) fail("tabs.label", "duplicate_label");
    labels.add(t.label as string);
  }
}
function validateStageOutput(e: ObjectValue) {
  if (e.kind !== "stage") {
    if (
      e.started_at !== undefined ||
      e.finished_at !== undefined ||
      e.document_snapshots !== undefined
    )
      fail("stage", "forbidden_for_kind");
    return;
  }
  timestamp(e.started_at, "started_at");
  const start = Date.parse(e.started_at as string),
    updated = Date.parse(e.updated_at as string);
  if (start !== Date.parse(e.created_at as string) || updated < start)
    fail("started_at", "server_time_mismatch");
  if (e.stage_state === "completed" || e.stage_state === "failed") {
    timestamp(e.finished_at, "finished_at");
    if (
      Date.parse(e.finished_at as string) < start ||
      Date.parse(e.finished_at as string) > updated
    )
      fail("finished_at", "server_time_mismatch");
  } else if (e.finished_at !== undefined)
    fail("finished_at", "forbidden_for_stage_state");
  if (e.document_snapshots !== undefined)
    for (const value of array(
      e.document_snapshots,
      "document_snapshots",
      TIMELINE_LIMITS.documentSnapshots,
    )) {
      const s = object(
        value,
        [
          "document_revision_id",
          "actor",
          "document_key",
          "status",
          "created_at",
          "tabs",
        ],
        ["predecessor_revision_id", "task_id", "segment_id"],
      );
      id(s.document_revision_id, "document_revision_id");
      timestamp(s.created_at, "document_snapshots.created_at");
      snapshotFields(s);
      if (
        Date.parse(s.created_at as string) < start ||
        Date.parse(s.created_at as string) > updated
      )
        fail("document_snapshots.created_at", "server_time_mismatch");
    }
}
function validateR3Session(s: TimelineSession) {
  const docs = new Set<string>(),
    hashes = new Set<string>();
  let total = 0;
  for (const value of array(
    s.documents,
    "documents",
    TIMELINE_LIMITS.documents,
  )) {
    const d = object(value, [
      "document_id",
      "actor",
      "session_id",
      "sha256",
      "size",
      "media_type",
      "received_at",
    ]);
    id(d.document_id, "document_id");
    if (
      d.session_id !== s.viewer_session_id ||
      docs.has(d.document_id as string)
    )
      fail("documents", "wrong_owner_or_duplicate");
    if (
      typeof d.sha256 !== "string" ||
      !sha256.test(d.sha256) ||
      hashes.has(d.sha256)
    )
      fail("documents.sha256");
    integer(d.size, "documents.size", 0, TIMELINE_LIMITS.documentBytes);
    if (d.media_type !== "text/markdown") fail("documents.media_type");
    timestamp(d.received_at, "documents.received_at");
    docs.add(d.document_id as string);
    hashes.add(d.sha256);
    total += d.size as number;
  }
  if (total > TIMELINE_LIMITS.aggregateDocumentBytes)
    fail("documents", "aggregate_limit");
  let active: string | null = null;
  const latestStages = new Map<TimelineStage, TimelineEvent>();
  const revisions = new Set<string>();
  for (const e of s.events) {
    if (e.kind !== "stage") continue;
    if (!e.withdrawn && e.stage) {
      const previous = latestStages.get(e.stage);
      if (
        previous &&
        ((previous.stage_state !== "completed" &&
          previous.stage_state !== "failed") ||
          e.related_event_id !== previous.event_id)
      )
        fail("related_event_id", "stage_reentry_reference_required");
      latestStages.set(e.stage, e);
    }
    if (
      ["in_progress", "in_review", "awaiting_validation"].includes(
        e.stage_state ?? "",
      )
    ) {
      if (active || e.withdrawn)
        fail("current_stage_event_id", "multiple_or_withdrawn_foreground");
      active = e.event_id;
    }
    for (const r of e.document_snapshots ?? []) {
      if (revisions.has(r.document_revision_id))
        fail("document_revision_id", "duplicate_id");
      revisions.add(r.document_revision_id);
      for (const t of r.tabs)
        if (!docs.has(t.document_id))
          fail("tabs.document_id", "unknown_document");
    }
  }
  if (revisions.size > TIMELINE_LIMITS.documentSnapshots)
    fail("document_snapshots", "session_limit");
  if (s.current_stage_event_id !== null)
    id(s.current_stage_event_id, "current_stage_event_id");
  if (active !== s.current_stage_event_id)
    fail("current_stage_event_id", "foreground_mismatch");
  if (active && !timelineStateAllowsForeground(s.state))
    fail("state", "foreground_requires_active_session");
}
export function isSafeTimelineDocumentURL(
  value: unknown,
  sessionID: string,
  documentID: string,
): value is string {
  return (
    typeof value === "string" &&
    sha256.test(sessionID) &&
    /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/.test(documentID) &&
    value === `/api/timeline/sessions/${sessionID}/documents/${documentID}`
  );
}
export function validateTimelineDocumentBytes(value: Uint8Array): void {
  if (value.length > TIMELINE_LIMITS.documentBytes)
    fail("document", "invalid_utf8_or_size");
  try {
    new TextDecoder("utf-8", { fatal: true }).decode(value);
  } catch {
    fail("document", "invalid_utf8_or_size");
  }
}
export function validateTimelineEventPatch(
  s: TimelineSession,
  eventID: string,
  r: TimelineEventPatchRequest,
): void {
  parseTimelineRequest("event_patch", r);
  const e = s.events.find((e) => e.event_id === eventID);
  if (!e) fail("event_id", "unknown_event");
  if (r.title !== undefined) text(r.title, "title", 200, e.kind !== "stage");
  if (e.audit) fail("event_id", "immutable_audit");
  validateTimelineAuthority(
    s,
    e.kind === "stage" ? "stage_update" : "event_patch",
    r.actor,
    e.actor.producer_id,
  );
  if (e.version !== r.expected_version)
    fail("expected_version", "version_conflict");
  if (r.stage_state === undefined) return;
  if (e.kind !== "stage") fail("stage_state", "forbidden_for_kind");
  if (r.expected_session_version !== s.version)
    fail("expected_session_version", "version_conflict");
  if (s.state === "closed" || e.withdrawn)
    fail("stage_state", "stage_unavailable");
  if (e.stage_state === r.stage_state) return;
  if (e.stage_state === "completed" || e.stage_state === "failed")
    fail("stage_state", "stage_terminal");
}
export function validateTimelineSessionPatch(
  s: TimelineSession,
  r: TimelineSessionPatchRequest,
): void {
  parseTimelineRequest("session_patch", r);
  validateTimelineAuthority(
    s,
    r.actor.kind === "local_viewer"
      ? r.archived !== undefined
        ? "session_archive"
        : r.pinned !== undefined
          ? "session_pin"
          : "session_rename"
      : "session_patch",
    r.actor,
  );
  if (s.version !== r.expected_version)
    fail("expected_version", "version_conflict");
  if (r.state === "closed" && s.current_stage_event_id !== null)
    fail("current_stage_event_id", "stage_active");
}
export function validateTimelineEventWithdrawal(
  s: TimelineSession,
  eventID: string,
  r: TimelineMutationRequest,
): void {
  parseTimelineRequest("mutation", r);
  const e = s.events.find((e) => e.event_id === eventID);
  if (!e) fail("event_id", "unknown_event");
  if (e.audit) fail("event_id", "immutable_audit");
  validateTimelineAuthority(
    s,
    e.kind === "stage" ? "stage_update" : "event_withdraw",
    r.actor,
    e.actor.producer_id,
  );
  if (e.version !== r.expected_version)
    fail("expected_version", "version_conflict");
  if (e.kind === "stage" && e.stage_state === "in_progress")
    fail("current_stage_event_id", "stage_active");
}
export function validateTimelineDocumentSnapshot(
  s: TimelineSession,
  eventID: string,
  r: TimelineDocumentSnapshotRequest,
): void {
  validateTimelineSnapshotReferencesV3(s, eventID, r);
  parseTimelineRequest("document_snapshot", r);
  if (s.state === "closed") fail("state", "session_closed");
  const e = s.events.find((e) => e.event_id === eventID);
  if (e?.kind !== "stage" || e.withdrawn) fail("event_id", "stage_required");
  if (e.version !== r.expected_version)
    fail("expected_version", "version_conflict");
  if (
    s.events.reduce((n, e) => n + (e.document_snapshots?.length ?? 0), 0) >=
    TIMELINE_LIMITS.documentSnapshots
  )
    fail("document_snapshots", "session_limit");
  for (const t of r.tabs)
    if (
      !s.documents.some(
        (d) =>
          d.document_id === t.document_id &&
          d.session_id === s.viewer_session_id,
      )
    )
      fail("tabs.document_id", "unknown_document");
}

/** Reads bytes as data only; renderers must still disable raw HTML and remote images. */
export async function parseTimelineDocumentResponse(
  value: unknown,
): Promise<TimelineDocumentResponse> {
  const r = object(value, ["schema_version", "document", "content"]);
  if (r.schema_version !== 3) fail("schema_version", "unsupported_protocol");
  const d = object(r.document, [
    "document_id",
    "actor",
    "session_id",
    "sha256",
    "size",
    "media_type",
    "received_at",
  ]);
  id(d.document_id, "document.document_id");
  if (
    typeof d.session_id !== "string" ||
    !sha256.test(d.session_id) ||
    typeof d.sha256 !== "string" ||
    !sha256.test(d.sha256)
  )
    fail("document", "invalid_metadata");
  integer(d.size, "document.size", 0, TIMELINE_LIMITS.documentBytes);
  if (d.media_type !== "text/markdown") fail("document.media_type");
  timestamp(d.received_at, "document.received_at");
  if (typeof r.content !== "string") fail("content");
  const bytes = new TextEncoder().encode(r.content);
  validateTimelineDocumentBytes(bytes);
  // Reject unpaired UTF-16 surrogates instead of hashing replacement characters.
  if (new TextDecoder("utf-8", { ignoreBOM: true }).decode(bytes) !== r.content)
    fail("content", "invalid_utf8_or_size");
  const digest = await crypto.subtle.digest("SHA-256", bytes);
  const hash = Array.from(new Uint8Array(digest), (b) =>
    b.toString(16).padStart(2, "0"),
  ).join("");
  if (bytes.length !== d.size || hash !== d.sha256)
    fail("document", "content_mismatch");
  return r as unknown as TimelineDocumentResponse;
}
