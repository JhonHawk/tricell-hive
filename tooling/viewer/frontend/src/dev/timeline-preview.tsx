// Development-only adapter: the production components receive typed stub data.
import { useEffect, useState } from "react";
import { ViewerShell } from "../shell";
import { TimelineContent } from "../timeline-content";
import type { TimelineSessionResponse } from "../timeline-contract";
import { TimelineHeader } from "../timeline-header";
import { loadPreviewDocument } from "./timeline-documents-fixture";
import { timelineFixture } from "./timeline-fixture";

export function TimelineDevelopmentPreview() {
  const query = new URLSearchParams(location.search);
  const scenario = query.get("state");
  const dark = query.get("theme") !== "light";
  const [data, setData] = useState<TimelineSessionResponse>(() => {
    const fixture = structuredClone(timelineFixture);
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
    ticket: "COM-1 · COM-2",
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
