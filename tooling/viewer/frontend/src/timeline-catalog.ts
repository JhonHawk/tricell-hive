import {
  isSafeTimelineURL,
  type TimelineCandidate,
  type TimelineCandidateRecord,
  TimelineContractError,
  type TimelineDeliverySnapshot,
  type TimelineNativeOperation,
  type TimelineSession,
  type TimelineTicketInput,
  type TimelineTicketRef,
} from "./timeline-contract";
import { validateTimelineActor } from "./timeline-contract-v3";
export type TimelineTicketCreateRequest = TimelineNativeOperation &
  TimelineTicketInput;
export type TimelineTicketPatchRequest = TimelineNativeOperation & {
  key?: string;
  title?: string;
  url?: string;
};
export type TimelineDeliveryCreateRequest = TimelineNativeOperation &
  TimelineDeliverySnapshot & { repository_url: string; external_id: string };
export type TimelineDeliveryPatchRequest = TimelineNativeOperation & {
  label?: string;
  url?: string;
  status?: "open" | "merged" | "local";
};
export type TimelineCandidateCreateRequest = TimelineNativeOperation &
  TimelineCandidate;
export type TimelineCatalogRequestMap = {
  ticket_create: TimelineTicketCreateRequest;
  ticket_patch: TimelineTicketPatchRequest;
  delivery_create: TimelineDeliveryCreateRequest;
  delivery_patch: TimelineDeliveryPatchRequest;
  candidate_create: TimelineCandidateCreateRequest;
};
function fail(field: string, code = "invalid_value"): never {
  throw new TimelineContractError(field, code);
}
type Obj = Record<string, unknown>;
function object(
  value: unknown,
  required: string[],
  optional: string[] = [],
): Obj {
  if (!value || typeof value !== "object" || Array.isArray(value))
    fail("body", "object_required");
  const o = value as Obj;
  for (const k of Object.keys(o)) {
    if (![...required, ...optional].includes(k)) fail("body", "unknown_field");
    if (o[k] === null) fail(k, "null_not_allowed");
  }
  for (const k of required) if (!Object.hasOwn(o, k)) fail(k, "required_field");
  return o;
}
function text(v: unknown, field: string, max: number, required = true) {
  if (
    typeof v !== "string" ||
    [...v].length > max ||
    (required && /^\p{White_Space}*$/u.test(v)) ||
    [...v].some((r) => /\p{Cc}/u.test(r) && r !== "\n" && r !== "\t")
  )
    fail(field, "invalid_text");
}
function id(v: unknown, field: string) {
  if (typeof v !== "string" || !/^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/.test(v))
    fail(field, "invalid_id");
}
function external(v: unknown) {
  text(v, "external_id", 256);
  if (/[\p{White_Space}\p{Cc}]/u.test(v as string))
    fail("external_id", "invalid_external_id");
}
function https(v: unknown, field: string) {
  if (typeof v !== "string" || [...v].length > 4096 || !isSafeTimelineURL(v))
    fail(field, "invalid_https_url");
}
export function canonicalTimelineIdentityURL(raw: string): string {
  https(raw, "identity_url");
  if (raw.includes("?") || raw.includes("#"))
    fail("identity_url", "invalid_identity_url");
  const match = /^https:\/\/([^/?#]+)(.*)$/i.exec(raw);
  if (!match) fail("identity_url", "invalid_identity_url");
  for (const segment of match[2].split("/")) {
    const decoded = segment.replace(/%2e/gi, ".");
    if (decoded === "." || decoded === "..")
      fail("identity_url", "invalid_identity_url");
  }
  return locator(raw).replace(/\/+$/, "");
}
function locator(raw: string | undefined): string {
  if (raw === undefined) fail("url", "invalid_https_url");
  const match = /^https:\/\/([^/?#]+)(.*)$/i.exec(raw);
  if (!match) fail("url", "invalid_https_url");
  const host = match[1].toLowerCase().replace(/:0*443$/, "");
  return `https://${host}${match[2]}`;
}
export function validateTimelineTicketInput(value: unknown) {
  const o = object(
    value,
    ["external_scope", "external_id", "key"],
    ["title", "url"],
  );
  ticketValues(o);
}
function ticketValues(o: Obj) {
  canonicalTimelineIdentityURL(o.external_scope as string);
  external(o.external_id);
  text(o.key, "key", 100);
  if (o.title !== undefined) text(o.title, "title", 300, false);
  if (o.url !== undefined) https(o.url, "url");
}
export function validateTimelineTicketRecord(value: unknown) {
  const o = object(
    value,
    ["ticket_id", "external_scope", "external_id", "key"],
    ["title", "url"],
  );
  id(o.ticket_id, "ticket_id");
  ticketValues(o);
}
export function validateTimelineIDs(
  value: unknown,
  field: string,
  required = false,
) {
  if (
    !Array.isArray(value) ||
    value.length > 100 ||
    (required && !value.length)
  )
    fail(field, "array_required_or_limit");
  const seen = new Set<string>();
  for (const v of value as unknown[]) {
    id(v, field);
    if (seen.has(v as string)) fail(field, "duplicate_id");
    seen.add(v as string);
  }
}
export function validateTimelineTicketReferences(value: unknown) {
  if (!Array.isArray(value) || value.length > 100)
    fail("ticket_refs", "array_required_or_limit");
  const seen = new Set<string>();
  for (const v of value as unknown[]) {
    const o = object(v, ["ticket_id", "snapshot"]);
    id(o.ticket_id, "ticket_id");
    if (seen.has(o.ticket_id as string)) fail("ticket_refs", "duplicate_id");
    seen.add(o.ticket_id as string);
    const snap = object(o.snapshot, ["key"], ["title", "url"]);
    text(snap.key, "snapshot.key", 100);
    if (snap.title !== undefined)
      text(snap.title, "snapshot.title", 300, false);
    if (snap.url !== undefined) https(snap.url, "snapshot.url");
  }
}
function delivery(o: Obj) {
  if (
    !["pull_request", "commit"].includes(o.kind as string) ||
    !["open", "merged", "local"].includes(o.status as string)
  )
    fail("delivery", "invalid_kind_or_status");
  text(o.label, "label", 200);
  https(o.url, "url");
}
export function validateTimelineDeliveryReference(value: unknown) {
  const o = object(value, ["delivery_id", "snapshot"]);
  id(o.delivery_id, "delivery_id");
  delivery(object(o.snapshot, ["kind", "label", "url", "status"]));
}
export function timelineCandidateByID(
  s: TimelineSession,
  id: string,
): TimelineCandidateRecord {
  const c = s.candidates.find((c) => c.candidate_id === id);
  if (!c) fail("candidate_id", "unknown_candidate");
  return c;
}
/** Tracks the identities and locators that tickets, deliveries and candidates
 * of one catalog must not repeat. */
export function timelineCatalogIdentities() {
  const ids = new Set<string>(),
    natural = new Set<string>(),
    urls = new Set<string>();
  function add(idValue: string, key: string, url?: string) {
    id(idValue, "catalog.id");
    if (ids.has(idValue) || natural.has(key))
      fail("catalog", "duplicate_identity");
    ids.add(idValue);
    natural.add(key);
    if (url) {
      if (urls.has(url)) fail("catalog", "entity_conflict");
      urls.add(url);
    }
  }
  /** Validates each ticket record and registers its identity and URL. */
  function addTickets(tickets: TimelineSession["tickets"]) {
    for (const t of tickets) {
      validateTimelineTicketRecord(t);
      add(
        t.ticket_id,
        `ticket\0${canonicalTimelineIdentityURL(t.external_scope)}\0${t.external_id}`,
        t.url ? `ticket\0${locator(t.url)}` : undefined,
      );
    }
  }
  return { add, addTickets };
}
export function validateTimelineCatalogs(s: TimelineSession) {
  for (const list of [s.tickets, s.deliveries, s.candidates])
    if (!Array.isArray(list) || list.length > 100)
      fail("catalog", "array_required_or_limit");
  const identities = timelineCatalogIdentities();
  const add = identities.add;
  identities.addTickets(s.tickets);
  for (const d of s.deliveries) {
    const o = object(d, [
      "delivery_id",
      "repository_url",
      "external_id",
      "kind",
      "label",
      "url",
      "status",
    ]);
    delivery(o);
    external(d.external_id);
    add(
      d.delivery_id,
      `delivery\0${canonicalTimelineIdentityURL(d.repository_url)}\0${d.kind}\0${d.external_id}`,
      `delivery\0${locator(d.url)}`,
    );
  }
  for (const c of s.candidates) {
    object(c, ["candidate_id", "branch", "commit", "base"]);
    for (const k of ["branch", "commit", "base"] as const) text(c[k], k, 256);
    add(c.candidate_id, `candidate\0${c.branch}\0${c.commit}\0${c.base}`);
  }
}
const base = [
  "schema_version",
  "operation_id",
  "actor",
  "expected_session_version",
  "reason",
];
const shapes = {
  ticket_create: [
    ["external_scope", "external_id", "key"],
    ["title", "url"],
  ],
  ticket_patch: [[], ["key", "title", "url"]],
  delivery_create: [
    ["repository_url", "external_id", "kind", "label", "url", "status"],
    [],
  ],
  delivery_patch: [[], ["label", "url", "status"]],
  candidate_create: [["branch", "commit", "base"], []],
} as const;
export function isTimelineCatalogOperation(
  kind: string,
): kind is keyof TimelineCatalogRequestMap {
  return Object.hasOwn(shapes, kind);
}
export function parseTimelineCatalogRequest<
  K extends keyof TimelineCatalogRequestMap,
>(kind: K, value: unknown): TimelineCatalogRequestMap[K] {
  const [required, optional] = shapes[kind];
  const o = object(value, [...base, ...required], [...optional]);
  if (o.schema_version !== 3) fail("schema_version", "unsupported_protocol");
  if (
    typeof o.operation_id !== "string" ||
    !/^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$/.test(o.operation_id)
  )
    fail("operation_id", "invalid_id");
  validateTimelineActor(o.actor, true);
  if (
    !Number.isSafeInteger(o.expected_session_version) ||
    Number(o.expected_session_version) < 1 ||
    Number(o.expected_session_version) > 2147483647
  )
    fail("expected_session_version", "invalid_number");
  text(o.reason, "reason", 500);
  if (kind.endsWith("_patch") && !optional.some((k) => Object.hasOwn(o, k)))
    fail("body", "empty_patch");
  if (kind === "ticket_create") ticketValues(o);
  if (kind === "ticket_patch") {
    if (o.key !== undefined) text(o.key, "key", 100);
    if (o.title !== undefined) text(o.title, "title", 300, false);
    if (o.url !== undefined && o.url !== "") https(o.url, "url");
  }
  if (kind === "delivery_create") {
    canonicalTimelineIdentityURL(o.repository_url as string);
    external(o.external_id);
    delivery(o);
  }
  if (kind === "delivery_patch") {
    if (o.label !== undefined) text(o.label, "label", 200);
    if (o.url !== undefined) https(o.url, "url");
    if (
      o.status !== undefined &&
      !["open", "merged", "local"].includes(o.status as string)
    )
      fail("status", "invalid_kind_or_status");
  }
  if (kind === "candidate_create")
    for (const k of ["branch", "commit", "base"]) text(o[k], k, 256);
  return o as unknown as TimelineCatalogRequestMap[K];
}
export function captureTimelineTicketReferences(
  s: TimelineSession,
  ids: string[],
): TimelineTicketRef[] {
  validateTimelineIDs(ids, "ticket_ids");
  return ids.map((id) => {
    const t = s.tickets.find((t) => t.ticket_id === id);
    if (!t) fail("ticket_ids", "unknown_ticket");
    const { key, title, url } = t;
    return {
      ticket_id: id,
      snapshot: { key, ...(title ? { title } : {}), ...(url ? { url } : {}) },
    };
  });
}
export function validateTimelineCatalogReferences(
  s: TimelineSession,
  kind: keyof TimelineCatalogRequestMap,
  objectID: string,
  value: unknown,
): void {
  const r = parseTimelineCatalogRequest(kind, value);
  if (r.expected_session_version !== s.version)
    fail("expected_session_version", "version_conflict");
  if (s.state === "closed" || s.state === "failed")
    fail("state", "session_terminal");
  if (
    r.actor.kind !== "native_producer" ||
    !s.producers.some((p) => p.producer_id === r.actor.producer_id)
  )
    fail("actor", "producer_not_accepted");
  if (
    kind !== "candidate_create" &&
    r.actor.producer_id !== s.foreground_producer_id
  )
    fail("actor", "lead_required");
  switch (kind) {
    case "ticket_create": {
      const q = r as TimelineTicketCreateRequest;
      const existing = s.tickets.find(
        (t) =>
          canonicalTimelineIdentityURL(t.external_scope) ===
            canonicalTimelineIdentityURL(q.external_scope) &&
          t.external_id === q.external_id,
      );
      if (existing) {
        if (
          existing.key !== q.key ||
          (existing.title ?? "") !== (q.title ?? "") ||
          (existing.url ?? "") !== (q.url ?? "")
        )
          fail("ticket_id", "entity_exists");
        return;
      }
      if (
        q.url &&
        s.tickets.some((t) => t.url && locator(t.url) === locator(q.url))
      )
        fail("url", "entity_conflict");
      if (s.tickets.length >= 100) fail("tickets", "limit");
      break;
    }
    case "ticket_patch": {
      const q = r as TimelineTicketPatchRequest;
      if (!s.tickets.some((t) => t.ticket_id === objectID))
        fail("ticket_id", "unknown_ticket");
      if (
        q.url &&
        s.tickets.some(
          (t) =>
            t.ticket_id !== objectID &&
            t.url &&
            locator(t.url) === locator(q.url),
        )
      )
        fail("url", "entity_conflict");
      break;
    }
    case "delivery_create": {
      const q = r as TimelineDeliveryCreateRequest;
      const existing = s.deliveries.find(
        (d) =>
          canonicalTimelineIdentityURL(d.repository_url) ===
            canonicalTimelineIdentityURL(q.repository_url) &&
          d.kind === q.kind &&
          d.external_id === q.external_id,
      );
      if (existing) {
        if (
          existing.label !== q.label ||
          existing.url !== q.url ||
          existing.status !== q.status
        )
          fail("delivery_id", "entity_exists");
        return;
      }
      if (s.deliveries.some((d) => locator(d.url) === locator(q.url)))
        fail("url", "entity_conflict");
      if (s.deliveries.length >= 100) fail("deliveries", "limit");
      break;
    }
    case "delivery_patch": {
      const q = r as TimelineDeliveryPatchRequest;
      if (!s.deliveries.some((d) => d.delivery_id === objectID))
        fail("delivery_id", "unknown_delivery");
      if (
        q.url &&
        s.deliveries.some(
          (d) =>
            d.delivery_id !== objectID && locator(d.url) === locator(q.url),
        )
      )
        fail("url", "entity_conflict");
      break;
    }
    case "candidate_create": {
      const q = r as TimelineCandidateCreateRequest;
      if (
        s.candidates.some(
          (c) =>
            c.branch === q.branch && c.commit === q.commit && c.base === q.base,
        )
      )
        return;
      if (s.candidates.length >= 100) fail("candidates", "limit");
      break;
    }
  }
}
