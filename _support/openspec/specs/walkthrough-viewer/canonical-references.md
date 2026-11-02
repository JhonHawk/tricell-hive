# Canonical references — final protocol 3

This contract extends the already authorized build. The user requested one origin for tasks and every other shared entity, explicitly authorizing a schema adjustment now. It supersedes the provisional protocol/storage 2 target. No compatibility reader, adapter, migration, provider fetch, remote Git action or habitual-service cutover is included.

## One mutable origin, explicit historical context

A work owns typed registries. Current projections resolve IDs against those registries; a producer updates the entity once. Do not introduce a generic entity framework, duplicate stage-owned task collections, URL-keyed render caches or a second mutable catalog.

- Tasks remain in `session.tasks`; segments retain their independent stage states and bindings. The toolbar Popover is the live plan catalog. Build projects its six bound segments; Verify and Delivery reference the same T7. Do not add a second live list inside the document Sheet.
- Tickets move to a stable-ID work registry. Deliveries gain a work registry. Header Popovers and current references resolve these registries directly.
- Candidates gain an immutable work registry. Evidence and waits retain their original candidate ID; new publications explicitly select their candidate ID and never retarget older references.
- Producers, artifacts, documents, revision chains and waits already have canonical IDs. Preserve them. Source identity, image placement captions/alt, published Markdown, audit snapshots, external-run observations and captured policies are provenance/context, not independent mutable entities to synchronize. Existing same-byte artifact dedup retains the first immutable provenance/candidate_id; created:false is reuse, not a new capture. Readback must not relabel that original artifact as freshly captured by a later upload.

Current references, including corrected ticket labels/URLs and delivery state, resolve the current entity. Publication snapshots remain available separately as historical context. Never present a current `merged` badge as the original state of an event titled “PR opened”: label historical/current states `Publicado` / `Actual` when both appear. Default current-resource links use the canonical current URL; an original publication link is explicitly historical. No patch rewrites an event's captured reference snapshot, document bytes/hash, image bytes/hash, audit or CI observation.

## Identity and registries

IDs are opaque, server-assigned and work-local. They are never inferred from a display title, visual position or URL. Natural identity prevents repeated registration of the same external resource:

| Registry | Immutable natural identity | Mutable values |
| --- | --- | --- |
| Ticket | `(external_scope, external_id)` | `key`, optional `title`, optional `url` |
| Delivery | `(repository_url, kind, external_id)` | `label`, `url`, `status` |
| Candidate | exact `(branch, commit, base)` | none |

`external_scope` is the authenticated tracker's HTTPS namespace in which its native ID is unique. `repository_url` is the authenticated repository's canonical HTTPS URL. The publisher supplies observed scope/ID; the viewer neither consults providers nor guesses from labels. Two trackers may use the same display key. Correcting a display key or locator preserves the opaque ID; changing external identity is not a metadata patch and never silently rewrites associations.

Canonicalize identity URL scheme/host to lowercase, remove HTTPS port 443 and trailing slash, preserve path case, reject credentials/query/fragment/dot segments. Reuse the safe HTTPS validator for editable links. An editable URL must be unique among tickets, and separately among deliveries, within a work, comparing normalized scheme/host/default HTTPS port while preserving its remaining locator components. Conflicting updates fail `entity_conflict`; do not merge entities.

A repeated create with the same natural identity and identical values returns its existing ID with `created:false`. Different editable values fail `entity_exists`; the caller patches the known ID. Exact operation replay returns its original receipt, including its original created result. Duplicate natural identities in one bootstrap fail atomically. Candidates deduplicate exact reported descriptors; an immutable candidate is reported provenance, not proof of process identity or of the contents of a moving working tree. Never move captured evidence/CI to a newer HEAD implicitly.

Required DTO delta (illustrative field notation; closed Go/TS schemas must agree):

```ts
TimelineTask += { plan_order: number };
TimelineTicket = {
  ticket_id: string;
  external_scope: string;
  external_id: string;
  key: string;
  title?: string;
  url?: string;
};
TimelineDelivery = {
  delivery_id: string;
  repository_url: string;
  external_id: string;
  kind: "pull_request" | "commit";
  label: string;
  url: string;
  status: "open" | "merged" | "local";
};
TimelineCandidateRecord = {
  candidate_id: string;
  branch: string;
  commit: string;
  base: string;
};
TimelineSession += {
  deliveries: TimelineDelivery[];
  candidates: TimelineCandidateRecord[];
};
```

`plan_order` is positive, safe, explicit and unique among nonwithdrawn tasks, including future tasks. It need not be contiguous or equal to task_key. The lead can patch it under existing task/session CAS, reason and plan/decision-reference rules. A collision fails `duplicate_plan_order`; no shifting, bulk reorder endpoint or drag-and-drop. `segment.order` stays independent and stage projections retain that ordering. The Popover sorts by plan_order.

## Reference inputs and captured outputs

Mutation inputs use IDs only:

- Events/views: `ticket_ids: string[]`, replacing inline ticket_refs inputs.
- Delivery events: `delivery_ids: string[]`, replacing inline deliveries.
- Events/artifact metadata/CI waits: `candidate_id`, replacing inline candidate wherever that kind previously permitted or required one.
- CI wait creation: `pull_request_delivery_id`, replacing inline pull_request_url. It must resolve to a work-local delivery of kind pull_request. Its current link comes from the registry; immutable publication context is captured by the server.

Stored/public event and view output reference entries include a server-generated publication snapshot:

```ts
ticket_refs: {
  ticket_id: string;
  snapshot: { key: string; title?: string; url?: string };
}[];
delivery_refs?: {
  delivery_id: string;
  snapshot: {
    kind: "pull_request" | "commit";
    label: string;
    url: string;
    status: "open" | "merged" | "local";
  };
}[];
```

Capture all snapshot values from the catalog under the same publication lock. Callers cannot submit snapshots, assigned IDs or output fields. A stored/public wait has required `pull_request_ref: TimelineDeliveryReference` (the `{delivery_id, snapshot}` shape above); its create input instead has required pull_request_delivery_id. Neither shape carries inline pull_request_url. Its candidate_id resolves an immutable record from the same work. A standalone CI-wait validator must receive a resolved work-local candidate record/session context; it cannot skip commit validation after removing the inline candidate. Do not keep parallel mutable inline fields or accept retired fields as aliases. Artifact/candidate records and all graph references belong to the same work. Preserve the old kind-specific allowance matrix rather than allowing every reference on every kind.

## Writers, authority and versions

Add typed native operations `ticket_create`, `ticket_patch`, `delivery_create`, `delivery_patch`, `candidate_create`. Use session-scoped collection/item routes and matching CLI verbs, each described by the offline catalog: collection POST `/api/timeline/sessions/:id/tickets`, item PATCH `/api/timeline/sessions/:id/tickets/:ticket-id`, equivalent `/deliveries` routes, and collection POST `/api/timeline/sessions/:id/candidates`. These follow the existing native authentication/origin rules; no new browser writer is added. Create results return the relevant ticket_id/delivery_id/candidate_id and created flag; patches return the relevant ID. Keep operation result unions closed.

Ticket/delivery create and patch require the current lead. Candidate creation permits any accepted native producer so collaborators can publish their own evidence. No global current-candidate pointer is introduced: current publications select explicit IDs. Other session_patch permissions stay unchanged; remove full-ticket replacement from session_patch, leaving one writer for that registry. Local-viewer actions gain no entity publication permission.

create_session can bootstrap tickets using an input type without server IDs. It validates the whole bootstrap and assigns IDs in one transaction/receipt. All later catalog changes use dedicated operations. Preserve actual native Source and canonical project membership checks.

Every mutation uses operation_id, accepted actor, reason and expected_session_version. Existing task patches also require expected_task_version. No new per-ticket/delivery version is needed: work CAS serializes those writers. Reject unknown/foreign/wrong-type IDs, missing required references, duplicate reference IDs, stale versions, mixed/retired fields and unsafe values before writing. Current entity patches append typed private audit with object, actor, reason and server time; do not add a generic public audit event or promise complete value reconstruction not actually stored.

A visible shared ticket/delivery change advances each affected task_version once in the same durable commit, traversing that task's event/view/wait references. A task linked through multiple paths advances once. Unrelated tasks, no-ops and retries do not advance. An open task Sheet retains its captured data/catalog and offers manual update. Wait snapshots that display a changed canonical PR link also advance wait_version once; linked task versions advance once, not twice. Historical snapshots themselves remain unchanged. Publishing a new candidate reference does not change fixed evidence/wait candidates or complete any task/stage/work.

Keep mutex/CAS, private request digests, atomic persistence, receipts, restart/replay, uncertainty reconciliation and closure reserves for event count, receipt count, public bytes and private-record bytes. New entity operations consume the normal content budget; they are not lifecycle exemptions. Failed registration/patch/invalidations persist nothing partially.

## Limits

- At most 100 tickets, 100 deliveries and 100 immutable candidates per work, counting retained records. These are ceilings, not guaranteed capacity.
- At most 100 unique ticket IDs per event/view. Delivery events require 1–100 unique delivery IDs; other kinds omit them under the existing allowance matrix.
- Identity/display HTTPS URLs: existing safe URL rules, at most 4096 characters. Natural external_id: 1–256 characters, no whitespace/control characters; not the internal-ID regex.
- Preserve key 1–100, optional title 0–300, delivery label 1–200, candidate branch/commit/base 1–256 and existing delivery enums. Unsupported PR closed-without-merge is not translated to merged.
- plan_order: 1–2147483647. Preserve other existing task/segment/ID/version bounds.
- Existing body 2 MiB, content budget 768 KiB, private record 4 MiB, content receipts 992/total 1000 and terminal headroom remain binding.
- Empty optional title/url patch values clear those optional display fields; omitted values preserve them. Null remains forbidden outside existing explicitly nullable fields. Immutable identity fields are forbidden in patch payloads.

## Protocol, storage and cutover

Use protocol_version/schema_version/storage_version 3 consistently, including health, index, offline schemas/catalog, operation results and receipts. This avoids an extra subrevision handshake: an installed provisional CLI already advertises version 2, while incompatible pairs must fail before writes. CLI/new server and old clients must reject incompatible versions explicitly.

Use a new positive version-3 state marker and namespace. Preflight rejects version-1/version-2 state, like any name outside the v3 inventory, with invalid_storage before mkdir, lock, log, cleanup, signaling, registry mutation or startup. Unknown/corrupt state remains an explicit error. Do not load/convert an old record to discover compatible data. No compatibility reader, migration, auto-import, preference copy or deletion of old state. Build and verify only in new task-owned directories; preserve the installed binary, habitual viewer, registry, old copies and CAM runtime. The real cutover remains separately authorized.

## Coherent example and UI

The complete example has exactly seven task entities, in one shared typed definition. Its initial Markdown task snapshot is generated from that definition at fixture construction, then stays immutable. Research/Plan can contain documents and activity without inventing T8/T9. Build projects T1–T6; Verify/Delivery share T7. The separate all-complete CAM scenario may still prove 6/7 then 7/7, while the mixed-state complete example derives its honest current total from its seven records.

Preserve the explicit Inicio marker, horizontal Tareas/Ir al actual toolbar, compact Popover, existing content affordances, contextual point colors, inline descriptions for simple tasks, immutable/manual Sheets, image galleries and icon+counter header. Current tickets/Git labels and links use canonical selectors, never last-URL-wins maps. Published context is clearly distinguished from current values without adding another editable source.

## Promotion and historical retention

T7 promotes the verified final content to `_support/openspec/specs/walkthrough-viewer/canonical-references.md`. Both normative requirement blocks use the identical reference link `[canonical references][canonical-references]`. Their definitions live outside normative blocks: the current capability preamble points to the adjacent permanent file and the archived delta preamble retains its own ../../canonical-references.md target. The archived change retains this file as a frozen design/publication record, with its own internal links; it is not another editable source. Verify both resolved links and exact promoted content at close.

## Reserved labels and browser language

The viewer supports English (default/fallback) and Spanish only. Select Spanish when the browser's primary language is es or an es-* locale, case-insensitively; every other, absent or malformed language selects English. Apply the same resolver to production and development previews, document language and all system-owned UI copy, accessible names, dates and status labels. Do not translate producer-authored titles, summaries, task descriptions, Markdown, captions or immutable publication history.

The offline protocol catalog exposes localization metadata: default_language en, supported_languages [en, es], and reserved title keys start, research, plan, build, verify, delivery, tasks, ci_wait. These are semantic keys, never caller-supplied translation dictionaries. Start/tasks/CI are derived UI regions, not additional API event kinds.

The existing stage field research|plan|build|verify|delivery is the canonical reserved stage key; do not add a duplicate title_key field. Stage headings always render the translated stage name once. Their existing title field is optional contextual text: retain the JSON field as a string, allow empty for stage events, and show nonempty context after the stage label. Producers must omit repeated stage names from that context and use an empty title when no extra context is needed. Other event kinds keep their required free title, without a reserved stage key.

Stage title patches may clear context using the existing empty-string title; omission preserves it. Go/TS/schema/CLI must agree on this kind-dependent rule and reject empty non-stage titles before writes. Preserve existing CAS, authority, receipts and snapshot history. Never guess a key from title text or translate stored producer context. The reserved metadata documents the existing stage enum plus derived start/tasks/CI UI regions, without adding event kinds or duplicate key writers. This is part of the unreleased final protocol3, with no compatibility reader, installation or habitual-runtime change.

The task Popover keeps installed component chrome and bounded scrolling for genuinely long content. Seven rows should fit without a tiny residual scrollbar where the viewport permits. Add space through content structure/parent gaps, not overriding installed chrome. At the real end of the timeline retain visible bottom breathing room inside the scroll canvas, after the last node/footer and with no fake node or activity count.
