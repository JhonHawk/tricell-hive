// Typed synthetic Markdown and exact immutable metadata for local previews.
import type {
  TimelineDocument,
  TimelineDocumentResponse,
} from "../timeline-contract";
export const previewDocumentMetadata: TimelineDocument[] = [
  {
    actor: {
      kind: "native_producer",
      producer_id: "producer-original",
    },
    document_id: "document_proposal",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "607c24787366fa5717d9a93078f7927c6688b61803aeda727ed702b7d629a423",
    size: 419,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: {
      kind: "native_producer",
      producer_id: "producer-original",
    },
    document_id: "document_design",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "4af3e6ab0b1c73a6f64ed208d0db15217fac5ed03b2e943e631df7cf1215f369",
    size: 1128,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: {
      kind: "native_producer",
      producer_id: "producer-original",
    },
    document_id: "document_tasks",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "b4ad8e698468b4fbf9ba03b1170a41b6097f930cc9aa2258269aab4fca83e9ea",
    size: 581,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: {
      kind: "native_producer",
      producer_id: "producer-original",
    },
    document_id: "document_criteria",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "df5d97e6bb40eddeca16b957a97c4fad0f19e20c57f7d6c5ef6aaf2388ed3bb4",
    size: 429,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: {
      kind: "native_producer",
      producer_id: "producer-original",
    },
    document_id: "document_earlier",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "f70822d448a642ab9ed271113e3ff41f17575f689aa1e87feb7994e1fd86a725",
    size: 152,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: {
      kind: "native_producer",
      producer_id: "producer-original",
    },
    document_id: "document_research",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "0323b121a5a998f3328820c5744cad0a23477a31838de504703beae4c663cc05",
    size: 213,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
];
const content: Record<string, string> = {
  document_proposal:
    "# Propuesta · R3\n\nLeer el trabajo de arriba hacia abajo y conservar las decisiones, las revisiones y la evidencia de cada etapa.\n\n## Resultado esperado\n\n1. Historial ascendente con tareas en su etapa real.\n2. Un Sheet común para investigación, plan, build, verificación y entrega.\n3. Lecturas estables que se actualizan solo por una acción explícita.\n\nLa publicación de este documento no equivale a aprobación.\n",
  document_design:
    "# Diseño · R3\n\n### Native composition and sketches\n\nReutilizar Timeline, Sheet, Tabs, Select, Collapsible, Tooltip y Spinner instalados. El encabezado y las pestañas permanecen fijos; el cuerpo tiene su propio scroll. En móvil el Sheet ocupa el ancho completo y en escritorio dos tercios.\n\n```text\nTrabajo                         6/7 finalizadas\nBuild                           6/6 finalizadas\nVerificación                    1/1 finalizadas\nEntrega                         0/1 finalizadas\n  ◌ T7 — Verificación y entrega\n    Ⅱ CI — Checks del pull request\n```\n\n## Estados independientes\n\n| Tramo | Estado declarado | Acción de lectura | Evidencia de origen |\n| --- | --- | --- | --- |\n| Build | Finalizada | Abrir detalle de implementación | Capturas de escritorio y móvil |\n| Verificación | Finalizada | Revisar resultados y observaciones | Contratos y pruebas completas |\n| Entrega | No iniciada | Consultar el candidato publicado | Checks del commit actual |\n\nLos estados se muestran solamente en el punto de cada fila. El Sheet muestra el global y todos los tramos, sin deducir progreso por la posición.\n",
  document_tasks:
    "# Tareas · R3\n\n- [x] T1 — Contrato tipado de productores, tareas y CI.\n- [x] T2 — Historial ascendente y navegación por trabajo.\n- [ ] T3 — Lecturas estables de documentos, tareas y CI.\n- [ ] T4 — Almacén v2 nuevo y corte limpio.\n- [ ] T5 — Colaboración y autoridad en el servidor.\n- [ ] T6 — Recorrido real de publicación con binario aislado.\n- [ ] T7 — Guía canónica y documentación.\n- [ ] T8 — Verificación integrada y validación humana.\n\nEsta lista es documento del plan. Los indicadores y contadores usan el registro tipado, nunca el texto Markdown.\n",
  document_criteria:
    "# Criterios · R3\n\n## MODIFIED Requirements\n\n### Lectura común\n\nEl panel conserva el documento abierto y ofrece Ver la última cuando llega un sucesor de su misma cadena. Las pestañas nuevas, cambiadas y quitadas se comparan por etiqueta e identidad del documento.\n\n### Tareas y Entrega\n\nT7 conserva un único ID entre Verificación y Entrega. Los checks verdes del candidato actual no prueban integración ni autorizan merge.\n",
  document_earlier:
    "# Plan · revisión R2\n\nRevisión anterior retenida para comparar las pestañas. Esta versión todavía no incluía el detalle completo de tareas y CI.\n",
  document_research:
    "# Investigación\n\n## Hallazgos\n\n- Los eventos asociados a etapas necesitan una jerarquía explícita.\n- Las lecturas largas requieren espacio y pestañas desplazables.\n- Cada productor conserva su identidad real.\n",
};
export async function loadPreviewDocument(
  documentID: string,
): Promise<TimelineDocumentResponse> {
  const document = previewDocumentMetadata.find(
    (d) => d.document_id === documentID,
  );
  if (!document) throw new Error("preview_document_missing");
  return { schema_version: 2, document, content: content[documentID] };
}
