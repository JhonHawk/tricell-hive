import type { TimelineTask } from "../timeline-contract";
// One typed plan defines both canonical task entities and their initial,
// immutable Markdown snapshot. Subsequent runtime states never rewrite it.
export const completeTaskPlan: ReadonlyArray<
  Readonly<
    Pick<
      TimelineTask,
      "task_key" | "plan_order" | "title" | "summary" | "state"
    >
  >
> = [
  {
    task_key: "T1",
    plan_order: 1,
    title: "Contrato de consultas y permisos del dashboard",
    summary:
      "Contrato terminado y disponible en el documento de implementación.",
    state: "completed",
  },
  {
    task_key: "T2",
    plan_order: 2,
    title: "Acceso al entorno para validar el perfil personal",
    summary:
      "El equipo del entorno debe habilitar el acceso antes de retomar esta validación.",
    state: "blocked",
  },
  {
    task_key: "T3",
    plan_order: 3,
    title: "Revisión de búsqueda y filtros de conversaciones en móvil",
    summary:
      "Se corrigió el filtro móvil; comparar la galería anterior con las capturas actuales.",
    state: "in_review",
  },
  {
    task_key: "T4",
    plan_order: 4,
    title: "Comprobación de permisos con el proveedor del entorno",
    summary:
      "Comprobación en curso. La captura de esta sesión todavía no está disponible para lectura.",
    state: "in_progress",
  },
  {
    task_key: "T5",
    plan_order: 5,
    title: "Documentación de ayuda para la configuración de equipos",
    summary:
      "No ha comenzado. Esta tarea todavía no tiene documentos ni imágenes adjuntos.",
    state: "not_started",
  },
  {
    task_key: "T6",
    plan_order: 6,
    title: "Pruebas de regresión y resultados del candidato",
    summary:
      "El alcance de pruebas de esta tarea terminó; los resultados de etapa permanecen en Verificación.",
    state: "completed",
  },
  {
    task_key: "T7",
    plan_order: 7,
    title: "Verificación y entrega del candidato de perfil y conversaciones",
    summary:
      "Verificación finalizó su tramo. Entrega espera los checks del candidato actual y el cierre observado.",
    state: "in_progress",
  },
];
export const completeTaskMarkdown =
  "# Tareas · R3\n\n" +
  completeTaskPlan
    .map(
      (t) =>
        `- [${t.state === "completed" ? "x" : " "}] ${t.task_key} — ${t.title}`,
    )
    .join("\n") +
  "\n\nInstantánea inicial publicada. El catálogo de tareas muestra su estado actual.\n";
