// Synthetic development-only data. No tracker or project account is contacted.
import { captureTimelineTicketReferences } from "../timeline-catalog";
import type { TimelineSessionResponse } from "../timeline-contract";
import { previewDocumentMetadata } from "./timeline-documents-fixture";

export const timelineFixture: TimelineSessionResponse = {
  schema_version: 3,
  mode: "timeline_api",
  session: {
    schema_version: 3,
    producers: [
      {
        producer_id: "producer-original",
        source: {
          project_root: "/fixture/demo",
          host: "codex",
          native_session_id: "preview-native-session",
        },
        joined_at: "2026-10-31T17:20:00-06:00",
      },
    ],
    foreground_producer_id: "producer-original",
    pending_join_requests: [],
    deliveries: [
      {
        delivery_id: "delivery-pr",
        repository_url: "https://example.test/repository",
        external_id: "42",
        kind: "pull_request",
        label: "PR #42",
        url: "https://example.test/pull/42",
        status: "open",
      },
    ],
    candidates: [
      {
        candidate_id: "candidate-r1",
        branch: "fix/profile-chat",
        commit: "preview-r1",
        base: "development",
      },
      {
        candidate_id: "candidate-r2",
        branch: "fix/profile-chat",
        commit: "preview-r2",
        base: "development",
      },
    ],
    tasks: [],
    waits: [],
    documents: previewDocumentMetadata,
    current_stage_event_id: "event_plan_r3",
    viewer_session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    project_key: "example.test/demo",
    project_name: "Proyecto demo",
    source: {
      project_root: "/fixture/demo",
      host: "codex",
      native_session_id: "preview-native-session",
    },
    version: 12,
    title: "Etapas y documentos del visor",
    summary:
      "Previsualización con extractos locales del plan R3 y evidencia sintética de perfil y chat.",
    state: "active",
    archived: false,
    pinned: false,
    created_at: "2026-10-31T17:20:00-06:00",
    updated_at: "2026-10-31T21:40:00-06:00",
    validation_count: 2,
    tickets: [
      {
        ticket_id: "ticket-1",
        external_scope: "https://example.test/tracker",
        external_id: "COM-1",
        key: "COM-1",
        title: "Espaciado del perfil",
        url: "https://example.test/tickets/COM-1",
      },
      {
        ticket_id: "ticket-2",
        external_scope: "https://example.test/tracker",
        external_id: "COM-2",
        key: "COM-2",
        title: "Filtros de conversaciones",
        url: "https://example.test/tickets/COM-2",
      },
    ],
    events: [
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        event_id: "event_request",
        sequence: 1,
        version: 1,
        kind: "request",
        title: "Perfil y filtros más claros",
        summary: "Trabajar los dos tickets en una misma sesión y entrega.",
        created_at: "2026-10-31T17:20:00-06:00",
        updated_at: "2026-10-31T17:20:00-06:00",
        origin: "agent_reported",
        ticket_refs: [
          {
            ticket_id: "ticket-1",
            snapshot: {
              key: "COM-1",
              url: "https://example.test/tickets/COM-1",
            },
          },
          {
            ticket_id: "ticket-2",
            snapshot: {
              key: "COM-2",
              url: "https://example.test/tickets/COM-2",
            },
          },
        ],
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        event_id: "event_started",
        sequence: 2,
        version: 1,
        kind: "work_started",
        title: "Trabajo iniciado",
        summary: "Plan acordado y rama preparada para el cambio.",
        created_at: "2026-10-31T17:30:00-06:00",
        updated_at: "2026-10-31T17:30:00-06:00",
        origin: "agent_reported",
        ticket_refs: [
          {
            ticket_id: "ticket-1",
            snapshot: {
              key: "COM-1",
              url: "https://example.test/tickets/COM-1",
            },
          },
          {
            ticket_id: "ticket-2",
            snapshot: {
              key: "COM-2",
              url: "https://example.test/tickets/COM-2",
            },
          },
        ],
        candidate_id: "candidate-r1",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        event_id: "event_pr",
        sequence: 3,
        version: 1,
        kind: "delivery",
        title: "PR de trabajo abierto",
        summary:
          "La entrega permanece abierta mientras se revisan los cambios.",
        created_at: "2026-10-31T18:05:00-06:00",
        updated_at: "2026-10-31T18:05:00-06:00",
        origin: "agent_reported",
        ticket_refs: [
          {
            ticket_id: "ticket-1",
            snapshot: {
              key: "COM-1",
              url: "https://example.test/tickets/COM-1",
            },
          },
          {
            ticket_id: "ticket-2",
            snapshot: {
              key: "COM-2",
              url: "https://example.test/tickets/COM-2",
            },
          },
        ],
        delivery_refs: [
          {
            delivery_id: "delivery-pr",
            snapshot: {
              kind: "pull_request",
              label: "PR #42",
              url: "https://example.test/pull/42",
              status: "open",
            },
          },
        ],
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        event_id: "event_validation",
        sequence: 4,
        version: 1,
        kind: "validation",
        title: "Primera validación",
        summary: "Perfil y filtros listos para revisar en escritorio y móvil.",
        created_at: "2026-10-31T18:30:00-06:00",
        updated_at: "2026-10-31T18:30:00-06:00",
        origin: "agent_reported",
        ticket_refs: [
          {
            ticket_id: "ticket-1",
            snapshot: {
              key: "COM-1",
              url: "https://example.test/tickets/COM-1",
            },
          },
          {
            ticket_id: "ticket-2",
            snapshot: {
              key: "COM-2",
              url: "https://example.test/tickets/COM-2",
            },
          },
        ],
        candidate_id: "candidate-r1",
        evidence_id: "evidence_initial",
        views: [
          {
            view_id: "view_profile",
            title: "Perfil personal",
            ticket_refs: [
              {
                ticket_id: "ticket-1",
                snapshot: {
                  key: "COM-1",
                  url: "https://example.test/tickets/COM-1",
                },
              },
            ],
            steps: [
              "Abrir Configuración y entrar a Perfil personal.",
              "Revisar la separación entre la imagen y el formulario, en escritorio y móvil.",
            ],
            expected_result:
              "Los campos quedan separados del divisor y dentro del ancho disponible.",
            image_refs: [
              {
                artifact_id: "artifact_profile-light",
                caption: "Escritorio · claro",
              },
              {
                artifact_id: "artifact_profile-mobile",
                caption: "390 px · claro",
              },
            ],
          },
          {
            view_id: "view_chat",
            title: "Conversaciones",
            ticket_refs: [
              {
                ticket_id: "ticket-2",
                snapshot: {
                  key: "COM-2",
                  url: "https://example.test/tickets/COM-2",
                },
              },
            ],
            steps: [
              "Abrir Conversaciones con la cuenta de demostración.",
              "Activar un filtro y revisar el buscador a 390 px.",
            ],
            expected_result:
              "Los filtros seleccionados permanecen visibles sin desplazar la lista.",
            image_refs: [
              {
                artifact_id: "artifact_chat-light",
                caption: "Escritorio · claro",
              },
              {
                artifact_id: "artifact_chat-mobile-before",
                caption: "390 px · oscuro",
              },
            ],
          },
        ],
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        event_id: "event_changes",
        sequence: 5,
        version: 1,
        kind: "changes_reported",
        title: "Cambios solicitados",
        summary:
          "Mostrar solo el icono del filtro en móvil y conservar el ancho del buscador.",
        created_at: "2026-10-31T18:55:00-06:00",
        updated_at: "2026-10-31T18:55:00-06:00",
        origin: "agent_reported",
        ticket_refs: [
          {
            ticket_id: "ticket-2",
            snapshot: {
              key: "COM-2",
              url: "https://example.test/tickets/COM-2",
            },
          },
        ],
        target: {
          kind: "evidence",
          event_id: "event_validation",
          evidence_id: "evidence_initial",
          view_ids: ["view_chat"],
        },
        source_reference: "Solicitud de ejemplo en el chat",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        event_id: "event_iteration",
        sequence: 6,
        version: 1,
        kind: "iteration",
        title: "Filtro compacto en móvil",
        summary:
          "Se actualizó Chat móvil. Perfil conserva la evidencia de la primera validación.",
        created_at: "2026-10-31T19:10:00-06:00",
        updated_at: "2026-10-31T19:10:00-06:00",
        origin: "agent_reported",
        ticket_refs: [
          {
            ticket_id: "ticket-2",
            snapshot: {
              key: "COM-2",
              url: "https://example.test/tickets/COM-2",
            },
          },
        ],
        candidate_id: "candidate-r2",
        evidence_id: "evidence_iteration",
        views: [
          {
            view_id: "view_chat",
            title: "Chat móvil",
            ticket_refs: [
              {
                ticket_id: "ticket-2",
                snapshot: {
                  key: "COM-2",
                  url: "https://example.test/tickets/COM-2",
                },
              },
            ],
            steps: [
              "Abrir Conversaciones a 390 px.",
              "Activar un filtro y cerrar el panel.",
              "Comprobar que el icono del filtro comparte fila con el buscador.",
            ],
            expected_result:
              "El filtro muestra solo su icono en móvil; el buscador mantiene su espacio.",
            image_refs: [
              {
                artifact_id: "artifact_chat-mobile-after",
                caption: "390 px · claro",
              },
              {
                artifact_id: "artifact_chat-mobile-dark",
                caption: "390 px · oscuro",
              },
            ],
            supersedes: {
              event_id: "event_validation",
              evidence_id: "evidence_initial",
              view_id: "view_chat",
            },
          },
        ],
        related_event_id: "event_validation",
      },
    ],
    artifacts: [
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        artifact_id: "artifact_profile-light",
        session_id:
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sha256:
          "ef7ba2b659a4f87e4767c98e98a4e6421f6a8fcf1d5f8ce1111789d50debc607",
        size: 40259,
        media_type: "image/png",
        received_at: "2026-10-31T18:20:00-06:00",
        captured_at: "2026-10-31T18:19:00-06:00",
        candidate_id: "candidate-r1",
        viewport: "desktop",
        theme: "light",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        artifact_id: "artifact_profile-mobile",
        session_id:
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sha256:
          "019c548d7295958838c643ac1171df7802005ec484898467336cce1d731c03e1",
        size: 29522,
        media_type: "image/png",
        received_at: "2026-10-31T18:20:00-06:00",
        captured_at: "2026-10-31T18:19:00-06:00",
        candidate_id: "candidate-r1",
        viewport: "mobile",
        theme: "light",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        artifact_id: "artifact_chat-light",
        session_id:
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sha256:
          "f72878dae49325d442bf6e6b80b5a4c43f9ba95eebde6eace29a8386f8b84e28",
        size: 59492,
        media_type: "image/png",
        received_at: "2026-10-31T18:20:00-06:00",
        captured_at: "2026-10-31T18:19:00-06:00",
        candidate_id: "candidate-r1",
        viewport: "desktop",
        theme: "light",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        artifact_id: "artifact_chat-mobile-before",
        session_id:
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sha256:
          "43d8de8e4cbbb0831eaa7cd664fd4a81182713870f4a98a0a2fbd5b5076f882a",
        size: 45834,
        media_type: "image/png",
        received_at: "2026-10-31T18:20:00-06:00",
        captured_at: "2026-10-31T18:19:00-06:00",
        candidate_id: "candidate-r1",
        viewport: "mobile",
        theme: "dark",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        artifact_id: "artifact_chat-mobile-after",
        session_id:
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sha256:
          "0e2d8162b654cc5c5c74b2bb6e742b2d515a900a7a56dcef1eaadf1a8fdb45a7",
        size: 46318,
        media_type: "image/png",
        received_at: "2026-10-31T19:08:00-06:00",
        captured_at: "2026-10-31T19:07:00-06:00",
        candidate_id: "candidate-r2",
        viewport: "mobile",
        theme: "light",
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        artifact_id: "artifact_chat-mobile-dark",
        session_id:
          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        sha256:
          "c6c7321066b2bc5b5482167d73a03da36440dd6e7ef49c06e438e55eba0c02fb",
        size: 46246,
        media_type: "image/png",
        received_at: "2026-10-31T19:08:00-06:00",
        captured_at: "2026-10-31T19:07:00-06:00",
        candidate_id: "candidate-r2",
        viewport: "mobile",
        theme: "dark",
      },
    ],
  },
  artifact_urls: {
    "artifact_profile-light":
      "/files/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/profile-light.png",
    "artifact_profile-mobile":
      "/files/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/profile-mobile.png",
    "artifact_chat-light":
      "/files/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/chat-light.png",
    "artifact_chat-mobile-before":
      "/files/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/chat-mobile-before.png",
    "artifact_chat-mobile-after":
      "/files/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/chat-mobile-after.png",
    "artifact_chat-mobile-dark":
      "/files/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/chat-mobile-dark.png",
  },
};

// R3 adds document publication to the approved R2 evidence fixture.
const stageTime = "2026-10-31T21:30:00-06:00";
timelineFixture.session.events.push(
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    event_id: "event_research",
    sequence: 7,
    version: 2,
    kind: "stage",
    stage: "research",
    stage_state: "completed",
    title: "Investigación",
    summary: "Composición nativa y documentos de solo lectura.",
    origin: "agent_reported",
    ticket_refs: [],
    created_at: "2026-10-31T20:00:00-06:00",
    updated_at: "2026-10-31T20:15:00-06:00",
    started_at: "2026-10-31T20:00:00-06:00",
    finished_at: "2026-10-31T20:15:00-06:00",
    document_snapshots: [
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        document_key: "research",
        document_revision_id: "revision_research",
        status: "complete",
        created_at: "2026-10-31T20:15:00-06:00",
        tabs: [{ label: "Investigación", document_id: "document_research" }],
      },
    ],
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    event_id: "event_plan_r2",
    sequence: 8,
    version: 2,
    kind: "stage",
    stage: "plan",
    stage_state: "completed",
    title: "Plan anterior",
    summary: "Primera revisión del plan, conservada como documento completo.",
    origin: "agent_reported",
    ticket_refs: [],
    created_at: "2026-10-31T20:18:00-06:00",
    updated_at: "2026-10-31T20:30:00-06:00",
    started_at: "2026-10-31T20:18:00-06:00",
    finished_at: "2026-10-31T20:30:00-06:00",
    document_snapshots: [
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        document_key: "plan",
        document_revision_id: "revision_plan_r2",
        status: "complete",
        created_at: "2026-10-31T20:30:00-06:00",
        tabs: [{ label: "Propuesta", document_id: "document_earlier" }],
      },
    ],
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    event_id: "event_build_r2",
    sequence: 9,
    version: 2,
    kind: "stage",
    stage: "build",
    stage_state: "completed",
    title: "Implementación de la composición",
    summary:
      "La evidencia visual de perfil y chat permanece en los nodos anteriores.",
    origin: "agent_reported",
    ticket_refs: [],
    created_at: "2026-10-31T20:35:00-06:00",
    updated_at: "2026-10-31T21:00:00-06:00",
    started_at: "2026-10-31T20:35:00-06:00",
    finished_at: "2026-10-31T21:00:00-06:00",
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    event_id: "event_plan_r3",
    sequence: 10,
    version: 3,
    kind: "stage",
    stage: "plan",
    stage_state: "in_progress",
    title: "Plan · revisión R3",
    summary:
      "Etapas declaradas, revisiones de documentos y conservación manual. Extractos locales del plan; datos de previsualización.",
    origin: "agent_reported",
    ticket_refs: [],
    created_at: stageTime,
    updated_at: "2026-10-31T21:40:00-06:00",
    started_at: stageTime,
    related_event_id: "event_plan_r2",
    document_snapshots: [
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        document_key: "plan",
        predecessor_revision_id: "revision_plan_r2",
        document_revision_id: "revision_plan_r3_initial",
        status: "draft",
        created_at: "2026-10-31T21:35:00-06:00",
        tabs: [{ label: "Propuesta", document_id: "document_earlier" }],
      },
      {
        actor: { kind: "native_producer", producer_id: "producer-original" },
        document_key: "plan",
        predecessor_revision_id: "revision_plan_r3_initial",
        document_revision_id: "revision_plan_r3_current",
        status: "draft",
        created_at: "2026-10-31T21:40:00-06:00",
        tabs: [
          { label: "Propuesta", document_id: "document_proposal" },
          { label: "Diseño", document_id: "document_design" },
          { label: "Tareas", document_id: "document_tasks" },
          { label: "Criterios", document_id: "document_criteria" },
        ],
      },
    ],
  },
);

// Emulate server publication capture once, before previews can patch the catalog.
for (const event of timelineFixture.session.events) {
  event.ticket_refs = captureTimelineTicketReferences(
    timelineFixture.session,
    event.ticket_refs.map((ref) => ref.ticket_id),
  );
  for (const view of event.views ?? []) {
    view.ticket_refs = captureTimelineTicketReferences(
      timelineFixture.session,
      view.ticket_refs.map((ref) => ref.ticket_id),
    );
  }
}
