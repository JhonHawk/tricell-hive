// Typed synthetic works; no existing account, tracker or viewer state is read.
import type {
  TimelineEvent,
  TimelineSessionResponse,
} from "../timeline-contract";
import type { TimelineTaskState } from "../timeline-contract-v2";
import { timelineFixture } from "./timeline-fixture";
export function workFixture(scenario = "cam"): TimelineSessionResponse {
  const data = structuredClone(timelineFixture),
    s = data.session;
  const actor = {
    kind: "native_producer" as const,
    producer_id: "producer-original",
  };
  const stamp = s.created_at;
  const stage = (
    id: string,
    kind: NonNullable<TimelineEvent["stage"]>,
    sequence: number,
    state: NonNullable<TimelineEvent["stage_state"]>,
  ): TimelineEvent => ({
    event_id: id,
    kind: "stage",
    stage: kind,
    stage_state: state,
    title: {
      research: "Investigación",
      plan: "Plan",
      build: "Build",
      verify: "Verificación",
      delivery: "Entrega",
    }[kind],
    sequence,
    version: 1,
    summary: "Alcance publicado con responsables y evidencia de cada etapa.",
    created_at: stamp,
    updated_at: stamp,
    started_at: stamp,
    ...(state === "completed" ? { finished_at: stamp } : {}),
    origin: "agent_reported",
    actor,
    ticket_refs: [],
  });
  s.events = [
    s.events[0],
    stage("build-stage", "build", 2, "completed"),
    stage("verify-stage", "verify", 3, "completed"),
  ];
  if (scenario !== "future")
    s.events.push(stage("delivery-stage", "delivery", 4, "in_progress"));
  s.title = "Dashboard: consultas, permisos y revisión de la entrega";
  s.summary =
    "La implementación terminó. Verificación conservó su evidencia y Entrega espera los checks del candidato actual.\n\n- Consultas y permisos revisados\n- Historial y filtros listos para validar";
  s.current_stage_event_id = scenario === "future" ? null : "delivery-stage";
  s.state = "active";
  s.validation_count = 0;
  s.events[0].sequence = 1;
  const titles = [
    "Contrato de consultas y permisos del dashboard",
    "Composición de pantalla y filtros de actividad",
    "Consultas con límites y permisos por equipo",
    "Revisión de actividad reciente y navegación",
    "Integración con el proveedor de resultados",
    "Pruebas de regresión y evidencia del dashboard",
    "Verificación y entrega del candidato",
  ];
  const count = scenario === "volume" ? 30 : 7;
  const states: TimelineTaskState[] = [
    "completed",
    "blocked",
    "completed",
    "in_review",
    "not_started",
    "in_progress",
  ];
  s.tasks = Array.from({ length: count }, (_, i) => ({
    task_id: `task-${i + 1}`,
    task_key: `T${i + 1}`,
    title: titles[i % titles.length],
    summary:
      "Revisar los resultados con permisos reales, documentar las diferencias y conservar la evidencia de escritorio y móvil.",
    task_version: 1,
    state:
      scenario === "volume"
        ? states[i % states.length]
        : i < 6
          ? "completed"
          : "in_progress",
    segments:
      i === 6 && scenario !== "volume"
        ? [
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
              state: "not_started",
              assigned_producer_id: "producer-original",
              ...(scenario !== "future"
                ? { display_stage_event_id: "delivery-stage" }
                : {}),
            },
          ]
        : [
            {
              segment_id: `build-${i + 1}`,
              planned_stage: "build",
              order: i + 1,
              state:
                scenario === "volume" ? states[i % states.length] : "completed",
              assigned_producer_id: "producer-original",
              display_stage_event_id: "build-stage",
              ...(scenario === "volume" &&
              states[i % states.length] === "blocked"
                ? {
                    block: {
                      reason: "Falta acceso al entorno de pruebas.",
                      steps: [
                        "Solicitar acceso al responsable del entorno.",
                        "Configurar la credencial en su ubicación segura.",
                        "Ejecutar la prueba de conexión.",
                      ],
                      owner: "Equipo del entorno",
                      resume_when:
                        "La prueba de conexión termina correctamente.",
                    },
                  }
                : {}),
            },
          ],
  }));
  if (scenario === "volume") {
    s.events[1].stage_state = "in_progress";
    s.events[1].finished_at = undefined;
    s.events = s.events.slice(0, 2);
    s.current_stage_event_id = "build-stage";
    const evidence = timelineFixture.session.events.find(
      (e) => e.views?.length,
    );
    const revision = timelineFixture.session.events
      .find((e) => e.event_id === "event_plan_r3")
      ?.document_snapshots?.at(-1);
    if (evidence) {
      s.events.push({
        ...structuredClone(evidence),
        event_id: "task-review-evidence",
        task_id: "task-2",
        segment_id: "build-2",
        stage_event_id: "build-stage",
        sequence: 3,
        title: "Revisión de perfil y conversaciones",
        summary: "Capturas pertinentes al acceso y navegación de esta tarea.",
      });
      s.validation_count = 1;
    }
    if (revision)
      s.events[1].document_snapshots = [
        {
          ...structuredClone(revision),
          document_key: "implementation-notes",
          document_revision_id: "revision-task-notes",
          predecessor_revision_id: undefined,
          task_id: "task-2",
          segment_id: "build-2",
        },
      ];
  } else if (scenario !== "future")
    s.waits = [
      {
        wait_id: "ci-wait",
        wait_version: 1,
        stage_event_id: "delivery-stage",
        task_id: "task-7",
        segment_id: "delivery-7",
        candidate: {
          branch: "fix/dashboard",
          commit: "abc1234",
          base: "development",
        },
        pull_request_url: "https://example.test/pull/42",
        policy: {
          policy_id: "required-v1",
          confirmed: true,
          authority: "Repository branch policy",
          source_url: "https://example.test/policy",
          observed_at: stamp,
          required_checks: [
            { check_key: "contracts", label: "Contratos" },
            { check_key: "tests", label: "Pruebas" },
            { check_key: "build", label: "Build" },
          ],
        },
        state: "waiting",
        applicable_attempt_id: {
          contracts: "attempt-1",
          tests: "attempt-1",
          build: "attempt-1",
        },
        attempts: [
          {
            attempt_id: "attempt-1",
            generation: 1,
            check_keys: ["contracts", "tests", "build"],
            external_run: {
              run_id: "run-1",
              commit: "abc1234",
              url: "https://example.test/actions/1",
            },
            source_url: "https://example.test/actions/1",
            observed_at: stamp,
            actor,
          },
        ],
        observations: ["contracts", "tests", "build"].map((check_key, i) => ({
          observation_id: `observation-${i}`,
          attempt_id: "attempt-1",
          check_key,
          external_run: {
            run_id: "run-1",
            commit: "abc1234",
            url: "https://example.test/actions/1",
          },
          state: i === 0 ? "succeeded" : "waiting",
          summary: i === 0 ? "Contratos correctos" : "Esperando resultado",
          observed_at: stamp,
          received_at: stamp,
          actor,
        })),
      },
    ];
  s.pending_join_requests = [
    {
      request_id: "join-codex",
      source: {
        project_root: "/fixture/demo-worktree",
        host: "claude-code",
        native_session_id: "native-contributor-preview",
      },
      requested_at: stamp,
      reason: "Continuar la revisión del dashboard desde una sesión nueva.",
    },
  ];
  return data;
}
