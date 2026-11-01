// One coherent synthetic work exercises the supported readers; no remote writes.
import type {
  TimelineDocumentSnapshot,
  TimelineEvent,
  TimelineSessionResponse,
  TimelineStage,
} from "../timeline-contract";
import type { TimelineTask, TimelineTaskState } from "../timeline-contract-v2";
import { previewDocumentMetadata } from "./timeline-documents-fixture";
import { timelineFixture } from "./timeline-fixture";
import { workFixture } from "./timeline-work-fixture";

export function completeFixture(): TimelineSessionResponse {
  const data = workFixture("cam"),
    s = data.session;
  const actor = {
    kind: "native_producer" as const,
    producer_id: "producer-original",
  };
  const stamp = "2026-11-01T10:00:00-06:00";
  const stageTitles: Record<TimelineStage, string> = {
    research: "Investigación",
    plan: "Plan",
    build: "Build",
    verify: "Verificación",
    delivery: "Entrega",
  };
  const snapshot = (
    key: string,
    id: string,
    tabs: TimelineDocumentSnapshot["tabs"],
    previous?: string,
  ): TimelineDocumentSnapshot => ({
    document_key: key,
    document_revision_id: id,
    actor,
    status: "complete",
    created_at: stamp,
    tabs,
    ...(previous ? { predecessor_revision_id: previous } : {}),
  });
  const stage = (
    kind: TimelineStage,
    sequence: number,
    state: NonNullable<TimelineEvent["stage_state"]>,
  ): TimelineEvent => ({
    event_id: `${kind}-stage`,
    kind: "stage",
    stage: kind,
    stage_state: state,
    title: stageTitles[kind],
    sequence,
    version: 1,
    actor,
    origin: "agent_reported",
    summary: {
      research:
        "Se identificaron los problemas de navegación, permisos y lectura de evidencia.",
      plan: "Plan del trabajo con revisiones comparables y una cadena de notas independiente.",
      build:
        "Las tareas conservan su estado individual; el acceso pendiente y las revisiones quedan visibles.",
      verify:
        "El tramo de verificación de T7 terminó y conserva sus resultados.",
      delivery:
        "El candidato está publicado; CI espera Pruebas y Build del intento actual.",
    }[kind],
    reason:
      kind === "build"
        ? "Seguimientos de acceso y revisión pendientes mientras avanza la entrega del candidato."
        : undefined,
    ticket_refs: [],
    created_at: stamp,
    updated_at: stamp,
    started_at: stamp,
    ...(state === "completed" ? { finished_at: stamp } : {}),
  });
  const request = structuredClone(timelineFixture.session.events[0]);
  request.title = "Perfil y conversaciones con evidencia completa";
  request.summary =
    "Revisar la composición completa del trabajo: tareas, estados, capturas, galerías, documentos y checks de entrega.";
  request.sequence = 1;
  s.title = "Perfil y conversaciones: implementación, evidencia y entrega";
  s.summary =
    "Un solo trabajo reúne las cinco etapas, tareas independientes y las capturas de la corrección. La entrega espera el intento vigente de CI; los seguimientos de acceso y revisión siguen identificados.\n\n- Evidencia anterior y actual de perfil y conversaciones\n- Lecturas de documentos y tareas sin sustituir lo abierto";
  s.events = [
    request,
    stage("research", 2, "completed"),
    stage("plan", 3, "completed"),
    stage("build", 4, "paused"),
    stage("verify", 5, "completed"),
    stage("delivery", 6, "in_progress"),
  ];
  s.current_stage_event_id = "delivery-stage";
  s.state = "active";
  s.reason = "Entrega en curso; todavía no se ha observado integración.";
  s.version = 40;
  s.validation_count = 0;
  s.documents = structuredClone(previewDocumentMetadata);
  const findStage = (kind: TimelineStage) => {
    const found = s.events.find((e) => e.stage === kind);
    if (!found) throw new Error("preview_stage_missing");
    return found;
  };
  findStage("research").document_snapshots = [
    snapshot("research", "revision-research", [
      { label: "Hallazgos", document_id: "document_research" },
    ]),
  ];
  findStage("plan").document_snapshots = [
    snapshot("plan", "revision-plan-one", [
      { label: "Propuesta", document_id: "document_earlier" },
    ]),
    {
      ...snapshot(
        "plan",
        "revision-plan-two",
        [
          { label: "Propuesta", document_id: "document_proposal" },
          { label: "Diseño", document_id: "document_earlier" },
          { label: "Tareas", document_id: "document_tasks" },
          { label: "Notas retiradas", document_id: "document_research" },
        ],
        "revision-plan-one",
      ),
      created_at: "2026-11-01T11:00:00-06:00",
    },
    {
      ...snapshot(
        "plan",
        "revision-plan-three",
        [
          { label: "Propuesta", document_id: "document_proposal" },
          { label: "Diseño", document_id: "document_design" },
          { label: "Tareas", document_id: "document_tasks" },
          { label: "Criterios", document_id: "document_criteria" },
        ],
        "revision-plan-two",
      ),
      created_at: "2026-11-01T12:00:00-06:00",
    },
    snapshot("risk-review", "revision-risks", [
      { label: "Diseño", document_id: "document_research" },
      { label: "Riesgos", document_id: "document_criteria" },
    ]),
  ];
  findStage("build").document_snapshots = [
    snapshot("implementation", "revision-build", [
      { label: "Implementación", document_id: "document_build" },
    ]),
  ];
  findStage("verify").document_snapshots = [
    snapshot("verification", "revision-verify", [
      { label: "Resultados", document_id: "document_verify" },
    ]),
  ];
  findStage("delivery").document_snapshots = [
    snapshot("delivery", "revision-delivery", [
      { label: "Candidato y CI", document_id: "document_delivery" },
    ]),
  ];
  const titles = [
    "Contrato de consultas y permisos del dashboard",
    "Acceso al entorno para validar el perfil personal",
    "Revisión de búsqueda y filtros de conversaciones en móvil",
    "Comprobación de permisos con el proveedor del entorno",
    "Documentación de ayuda para la configuración de equipos",
    "Pruebas de regresión y resultados del candidato",
  ];
  const summaries = [
    "Contrato terminado y disponible en el documento de implementación.",
    "El equipo del entorno debe habilitar el acceso antes de retomar esta validación.",
    "Se corrigió el filtro móvil; comparar la galería anterior con las capturas actuales.",
    "Comprobación en curso. La captura de esta sesión todavía no está disponible para lectura.",
    "No ha comenzado. Esta tarea todavía no tiene documentos ni imágenes adjuntos.",
    "El alcance de pruebas de esta tarea terminó; los resultados de etapa permanecen en Verificación.",
  ];
  const states: TimelineTaskState[] = [
    "completed",
    "blocked",
    "in_review",
    "in_progress",
    "not_started",
    "completed",
  ];
  s.tasks = titles.map(
    (title, i): TimelineTask => ({
      task_id: `task-${i + 1}`,
      task_key: `T${i + 1}`,
      title,
      summary: summaries[i],
      task_version: 3,
      plan_revision_id: "revision-plan-three",
      state: states[i],
      segments: [
        {
          segment_id: `build-${i + 1}`,
          planned_stage: "build",
          order: i + 1,
          state: states[i],
          assigned_producer_id:
            i === 2 ? "producer-reviewer" : "producer-original",
          display_stage_event_id: "build-stage",
          ...(i === 1
            ? {
                block: {
                  reason:
                    "Falta una cuenta de lectura para el entorno de pruebas.",
                  steps: [
                    "Solicitar acceso al equipo responsable del entorno.",
                    "Guardar la credencial en la ubicación segura del proyecto.",
                    "Ejecutar la prueba de conexión y registrar el resultado.",
                  ],
                  owner: "Equipo del entorno",
                  resume_when:
                    "La prueba de conexión termina correctamente con permisos de lectura.",
                },
              }
            : {}),
        },
      ],
    }),
  );
  s.tasks.push({
    task_id: "task-7",
    task_key: "T7",
    title: "Verificación y entrega del candidato de perfil y conversaciones",
    summary:
      "Verificación finalizó su tramo. Entrega espera los checks del candidato actual y el cierre observado.",
    task_version: 4,
    plan_revision_id: "revision-plan-three",
    state: "in_progress",
    segments: [
      {
        segment_id: "verify-7",
        planned_stage: "verify",
        order: 1,
        state: "completed",
        assigned_producer_id: "producer-original",
        display_stage_event_id: "verify-stage",
      },
      {
        segment_id: "delivery-7",
        planned_stage: "delivery",
        order: 1,
        state: "in_progress",
        assigned_producer_id: "producer-original",
        display_stage_event_id: "delivery-stage",
      },
    ],
  });
  s.tasks.push(
    {
      task_id: "task-8",
      task_key: "T8",
      title: "Investigación de navegación y lectura de evidencia",
      summary: "Hallazgos iniciales publicados.",
      task_version: 1,
      state: "completed",
      segments: [
        {
          segment_id: "research-8",
          planned_stage: "research",
          order: 1,
          state: "completed",
          assigned_producer_id: "producer-original",
          display_stage_event_id: "research-stage",
        },
      ],
    },
    {
      task_id: "task-9",
      task_key: "T9",
      title: "Plan de implementación y revisión de riesgos",
      summary: "Plan acordado con notas de riesgo separadas.",
      task_version: 1,
      state: "completed",
      segments: [
        {
          segment_id: "plan-9",
          planned_stage: "plan",
          order: 1,
          state: "completed",
          assigned_producer_id: "producer-original",
          display_stage_event_id: "plan-stage",
        },
      ],
    },
  );
  s.producers.push({
    producer_id: "producer-reviewer",
    source: {
      project_root: s.source.project_root,
      host: "claude-code",
      native_session_id: "native-reviewer-demo",
    },
    joined_at: stamp,
  });
  const initial = timelineFixture.session.events.find(
      (e) => e.event_id === "event_validation",
    ),
    iteration = timelineFixture.session.events.find(
      (e) => e.event_id === "event_iteration",
    );
  if (!initial || !iteration) throw new Error("preview_evidence_missing");
  const profile = structuredClone(initial);
  profile.event_id = "profile-evidence";
  profile.evidence_id = "evidence-profile";
  profile.sequence = 7;
  profile.stage_event_id = "build-stage";
  profile.task_id = "task-2";
  profile.segment_id = "build-2";
  profile.title = "Perfil: escritorio y móvil antes de la validación de acceso";
  profile.summary =
    "Galería del formulario con capturas horizontales y verticales, caption y texto alternativo.";
  profile.views = profile.views?.filter((v) => v.view_id === "view_profile");
  for (const view of profile.views ?? [])
    for (const ref of view.image_refs) {
      ref.caption = `Perfil · ${ref.caption}`;
      ref.alt = ref.artifact_id.includes("mobile")
        ? "Formulario de perfil en pantalla móvil, con imagen y campos separados."
        : "Formulario de perfil en escritorio, con navegación y datos de la cuenta.";
    }
  s.events.push(profile);
  s.validation_count++;
  const chat = structuredClone(initial);
  chat.event_id = "chat-evidence";
  chat.evidence_id = "evidence-chat-before";
  chat.sequence = 8;
  chat.stage_event_id = "build-stage";
  chat.task_id = "task-3";
  chat.segment_id = "build-3";
  chat.title = "Conversaciones: evidencia anterior de los filtros";
  chat.summary =
    "La vista anterior se conserva para comparar la corrección móvil.";
  chat.views = chat.views?.filter((v) => v.view_id === "view_chat");
  s.events.push(chat);
  s.validation_count++;
  const corrected = structuredClone(iteration);
  corrected.event_id = "chat-correction";
  corrected.sequence = 9;
  corrected.stage_event_id = "build-stage";
  corrected.task_id = "task-3";
  corrected.segment_id = "build-3";
  corrected.related_event_id = "chat-evidence";
  corrected.title = "Conversaciones: filtro móvil corregido";
  for (const view of corrected.views ?? []) {
    view.supersedes = {
      event_id: "chat-evidence",
      evidence_id: "evidence-chat-before",
      view_id: "view_chat",
    };
    for (const ref of view.image_refs) {
      ref.alt = `Filtro de conversaciones corregido en móvil; ${ref.caption}.`;
    }
  }
  s.events.push(corrected);
  const verified = structuredClone(initial);
  verified.event_id = "verification-evidence";
  verified.evidence_id = "evidence-verify";
  verified.sequence = 10;
  verified.stage_event_id = "verify-stage";
  verified.title = "Resumen visual de la validación de conversaciones";
  verified.summary =
    "Evidencia general de etapa con una sola imagen, sin asignación a una tarea.";
  verified.views = verified.views
    ?.filter((v) => v.view_id === "view_chat")
    .map((v) => ({
      ...v,
      image_refs: v.image_refs.filter(
        (r) => r.artifact_id === "artifact_chat-light",
      ),
      title: "Conversaciones verificadas",
    }));
  s.events.push(verified);
  s.validation_count++;
  const unavailableArtifact = {
    ...structuredClone(s.artifacts[0]),
    artifact_id: "artifact-unavailable",
  };
  s.artifacts.push(unavailableArtifact);
  const inaccessible = structuredClone(initial);
  inaccessible.event_id = "unavailable-evidence";
  inaccessible.evidence_id = "evidence-unavailable";
  inaccessible.sequence = 11;
  inaccessible.stage_event_id = "build-stage";
  inaccessible.task_id = "task-4";
  inaccessible.segment_id = "build-4";
  inaccessible.title = "Permisos: captura pendiente de acceso";
  inaccessible.summary =
    "La referencia está registrada; el recurso no está disponible en esta lectura.";
  inaccessible.views = [
    {
      view_id: "view-unavailable",
      title: "Captura del entorno pendiente de acceso",
      ticket_refs: [],
      steps: ["Comprobar el permiso de lectura del recurso."],
      expected_result:
        "La captura estará disponible cuando se complete el acceso.",
      image_refs: [
        {
          artifact_id: "artifact-unavailable",
          caption: "Entorno · captura no disponible",
          alt: "Captura del entorno que todavía no está disponible.",
        },
      ],
    },
  ];
  s.events.push(inaccessible);
  s.validation_count++;
  const access = snapshot("access-notes", "revision-access", [
    { label: "Acceso al entorno", document_id: "document_access" },
  ]);
  access.task_id = "task-2";
  access.segment_id = "build-2";
  findStage("build").document_snapshots?.push(access);
  const blocked = s.tasks[1].segments[0].block;
  s.events.push({
    event_id: "task-block-audit",
    kind: "task_state_transition",
    sequence: 12,
    version: 1,
    title: "Acceso de T2 bloqueado",
    summary:
      "El equipo registró el motivo y los pasos para recuperar el acceso.",
    actor,
    origin: "server",
    task_id: "task-2",
    segment_id: "build-2",
    stage_event_id: "build-stage",
    created_at: stamp,
    updated_at: stamp,
    ticket_refs: [],
    audit: {
      operation: "task_transition",
      reason: "El entorno requiere una cuenta de lectura antes de continuar.",
      previous_state: "in_progress",
      next_state: "blocked",
      block: blocked,
    },
  });
  const wait = s.waits[0];
  wait.candidate = {
    branch: "fix/profile-chat",
    commit: "preview-r2",
    base: "development",
  };
  wait.next_action =
    "Esperar los resultados de Pruebas y Build del intento vigente.";
  wait.state = "waiting";
  wait.wait_version = 4;
  for (const attempt of wait.attempts)
    attempt.external_run.commit = "preview-r2";
  for (const observation of wait.observations) {
    observation.external_run.commit = "preview-r2";
    observation.state =
      observation.check_key === "contracts"
        ? "succeeded"
        : observation.check_key === "tests"
          ? "failed"
          : "cancelled";
    observation.summary =
      observation.check_key === "contracts"
        ? "Contratos correctos"
        : observation.check_key === "tests"
          ? "Fallo anterior conservado en el historial"
          : "Build cancelado antes de la nueva ejecución";
  }
  const attempt = {
    ...structuredClone(wait.attempts[0]),
    attempt_id: "attempt-current",
    generation: 2,
    external_run: {
      run_id: "run-2",
      commit: "preview-r2",
      url: "https://example.test/actions/2",
    },
    source_url: "https://example.test/actions/2",
    observed_at: "2026-11-01T13:00:00-06:00",
  };
  wait.attempts.push(attempt);
  for (const check of wait.policy.required_checks) {
    wait.applicable_attempt_id[check.check_key] = attempt.attempt_id;
    wait.observations.push({
      observation_id: `current-${check.check_key}`,
      attempt_id: attempt.attempt_id,
      check_key: check.check_key,
      external_run: attempt.external_run,
      state: check.check_key === "contracts" ? "succeeded" : "waiting",
      summary:
        check.check_key === "contracts"
          ? "Contratos correctos en el candidato actual"
          : "Resultado pendiente del intento vigente",
      observed_at: attempt.observed_at,
      received_at: attempt.observed_at,
      actor,
    });
  }
  wait.observations.push({
    ...structuredClone(wait.observations[1]),
    observation_id: "late-prior-success",
    state: "succeeded",
    summary:
      "Resultado exitoso atrasado del intento anterior; no reemplaza el intento vigente.",
    observed_at: "2026-11-01T14:00:00-06:00",
    received_at: "2026-11-01T14:00:00-06:00",
  });
  const delivery = timelineFixture.session.events.find(
    (event) => event.kind === "delivery",
  );
  if (delivery)
    s.events.push({
      ...structuredClone(delivery),
      event_id: "delivery-publication",
      sequence: 13,
      stage_event_id: "delivery-stage",
      task_id: "task-7",
      segment_id: "delivery-7",
      title: "Pull request del candidato abierto",
      summary:
        "El candidato está publicado para revisar sus checks. La integración sigue sin confirmar.",
      deliveries: [
        {
          kind: "pull_request",
          label: "PR #42",
          url: "https://example.test/pull/42",
          status: "open",
        },
        {
          kind: "commit",
          label: "Commit preview-r2",
          url: "https://example.test/commit/preview-r2",
          status: "local",
        },
      ],
    });
  // Keep real image examples visible without expanding a stage or task.
  for (const event of s.events) if (event.sequence >= 2) event.sequence++;
  const visualExamples = structuredClone(initial);
  visualExamples.event_id = "visual-examples";
  visualExamples.evidence_id = "evidence-visual-examples";
  visualExamples.sequence = 2;
  visualExamples.title = "Ejemplos de imágenes y galerías";
  visualExamples.summary =
    "Capturas reales del ejemplo: una galería de escritorio y móvil y una imagen individual. Pulsa una miniatura para ampliarla.";
  delete visualExamples.stage_event_id;
  delete visualExamples.task_id;
  delete visualExamples.segment_id;
  visualExamples.views = [
    ...structuredClone(profile.views ?? []).map((view) => ({
      ...view,
      view_id: "view-example-gallery",
      title: "Galería · escritorio y móvil",
    })),
    ...structuredClone(verified.views ?? []).map((view) => ({
      ...view,
      view_id: "view-example-single",
      title: "Captura individual · conversaciones",
    })),
  ];
  s.events.push(visualExamples);
  s.events.sort((a, b) => a.sequence - b.sequence);
  s.validation_count = s.events.filter(
    (e) => e.kind === "validation" || e.kind === "iteration",
  ).length;
  s.summary = s.summary.replaceAll("\n", " ");
  s.updated_at = "2026-11-01T14:30:00-06:00";
  for (const event of s.events) {
    if (event.kind === "stage") event.updated_at = s.updated_at;
  }
  return data;
}
