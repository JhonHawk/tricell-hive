// Development-only adapter: the production components receive typed stub data.
import { useEffect, useState } from "react";
import { HttpError } from "../api";
import { ViewerShell } from "../shell";
import { TimelineContent } from "../timeline-content";
import type { TimelineSessionResponse } from "../timeline-contract";
import { TimelineHeader } from "../timeline-header";
import { completeFixture } from "./timeline-complete-fixture";
import { loadPreviewDocument } from "./timeline-documents-fixture";
import { timelineFixture } from "./timeline-fixture";
import { workFixture } from "./timeline-work-fixture";

export function TimelineDevelopmentPreview() {
  const query = new URLSearchParams(location.search);
  const scenario = query.get("state");
  const dark = query.get("theme") !== "light";
  const [data, setData] = useState<TimelineSessionResponse>(() => {
    const fixture = ["complete", "canonical-changed"].includes(scenario ?? "")
      ? completeFixture()
      : ["cam", "volume", "future", "recovery-conflict"].includes(
            scenario ?? "",
          )
        ? workFixture(scenario ?? "cam")
        : structuredClone(timelineFixture);
    if (scenario === "canonical-changed") {
      fixture.session.tickets[0].key = "PERFIL-7";
      fixture.session.tickets[0].title =
        "Perfil y conversaciones: referencia corregida";
      fixture.session.tickets[0].url = "https://example.test/issues/profile-7";
      fixture.session.deliveries[0].label = "PR #42 · integración observada";
      fixture.session.deliveries[0].url =
        "https://example.test/repository/pull/42";
      fixture.session.deliveries[0].status = "merged";
    }
    if (scenario === "chains") {
      const stage = fixture.session.events.find(
        (e) => e.event_id === "event_plan_r3",
      );
      const revision = stage?.document_snapshots?.at(-1);
      if (stage && revision)
        stage.document_snapshots?.push({
          ...structuredClone(revision),
          document_key: "review-notes",
          document_revision_id: "revision-review-notes",
          predecessor_revision_id: undefined,
          tabs: [
            { label: "Diseño", document_id: "document_research" },
            { label: "Riesgos", document_id: "document_criteria" },
          ],
        });
    }
    if (scenario === "empty") {
      fixture.session.events = [];
      fixture.session.current_stage_event_id = null;
      fixture.session.validation_count = 0;
    }
    if (scenario === "no-links") {
      fixture.session.tickets = [];
      for (const event of fixture.session.events) {
        event.ticket_refs = [];
        for (const view of event.views ?? []) view.ticket_refs = [];
      }
    }
    if (scenario === "long") {
      fixture.session.title =
        "Actualización de perfil e invitaciones y organización de los filtros en conversaciones con nombres extensos";
    }
    if (
      scenario === "paused" ||
      scenario === "failed" ||
      scenario === "completed"
    ) {
      const stage = fixture.session.events.find(
        (e) => e.event_id === "event_plan_r3",
      );
      if (stage) {
        stage.stage_state = scenario;
        if (scenario !== "paused") stage.finished_at = stage.updated_at;
      }
      fixture.session.current_stage_event_id = null;
      fixture.session.state = scenario === "paused" ? "paused" : "active";
    }
    return fixture;
  });
  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    document.documentElement.lang = "es";
  }, [dark]);
  const session = data.session;
  const selected = {
    id: session.viewer_session_id,
    project: session.project_name,
    projectKey: session.project_key,
    ticket: session.tickets.map((ticket) => ticket.key).join(" · "),
    title: session.title,
    createdAt: "31 oct 2026",
    classification: "local" as const,
    status: "in_progress" as const,
    pinned: session.pinned,
    archived: session.archived,
  };
  return (
    <ViewerShell
      sessions={[selected]}
      selectedId={selected.id}
      state="ready"
      onRetry={() => {}}
      sessionHeader={
        <TimelineHeader
          data={data}
          lang="es"
          onRetry={() => {}}
          onUpdate={async (patch, expectedVersion) => {
            if (scenario === "fail-save")
              throw new Error("fixture_write_failure");
            if (expectedVersion !== session.version)
              throw new Error("fixture_conflict");
            setData((previous) => ({
              ...previous,
              session: {
                ...previous.session,
                ...patch,
                version: previous.session.version + 1,
              },
            }));
          }}
          onRecover={async (requestID, reason, expectedVersion) => {
            if (
              scenario === "recovery-conflict" ||
              expectedVersion !== session.version
            )
              throw new HttpError(409, { code: "version_conflict" });
            const request = session.pending_join_requests.find(
              (r) => r.request_id === requestID,
            );
            if (!request) throw new Error("fixture_request_unavailable");
            setData((previous) => ({
              ...previous,
              session: {
                ...previous.session,
                version: previous.session.version + 1,
                pending_join_requests:
                  previous.session.pending_join_requests.filter(
                    (r) => r.request_id !== requestID,
                  ),
                producers: [
                  ...previous.session.producers,
                  {
                    producer_id: "producer-recovered",
                    source: request.source,
                    joined_at: previous.session.updated_at,
                  },
                ],
                foreground_producer_id: "producer-recovered",
                reason,
              },
            }));
          }}
          onDelete={async () => {
            setData((previous) => ({
              ...previous,
              session: { ...previous.session, events: [], validation_count: 0 },
            }));
          }}
        />
      }
    >
      <TimelineContent
        data={data}
        defaultExpandedTaskIDs={
          ["complete", "canonical-changed"].includes(scenario ?? "")
            ? ["task-2", "task-3", "task-4", "task-7"]
            : undefined
        }
        lang="es"
        loadDocument={async (documentID) => {
          if (scenario === "document-error")
            throw new Error("preview_document_failure");
          if (scenario === "document-loading") return new Promise(() => {});
          const response = await loadPreviewDocument(documentID);
          return scenario === "document-empty"
            ? { ...response, content: "" }
            : response;
        }}
        loading={scenario === "loading"}
        error={scenario === "error"}
        onRetry={() => {}}
      />
    </ViewerShell>
  );
}
