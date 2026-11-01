// Read-only, development-only extracts; IDs and transport are synthetic.
import type {
  TimelineDocument,
  TimelineDocumentResponse,
} from "../timeline-contract";
export const previewDocumentMetadata: TimelineDocument[] = [
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    document_id: "document_proposal",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "d6af195d8f6367c6f2b6a3373cee428e18c267b2e2d0da8f79d45d10bc45c5a8",
    size: 793,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    document_id: "document_design",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "b45b33fca37e0c65a44b02c360553081d58316b29868d6c0f9514b1ebda6205c",
    size: 2124,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    document_id: "document_tasks",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "fd8a3ad7709aa5d31f2d861992c3b0c500e47676ba79c5a450b353bd3a700b78",
    size: 2054,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    document_id: "document_criteria",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "c7f1dd120954614e8b36e985fb5df549389c718aeb9fa770cb621be21692a457",
    size: 449,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    document_id: "document_earlier",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "ca559af0b06cc0d552428d01a2efd2c5435b69c064cf9e223772f083cf7f4048",
    size: 143,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
  {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    document_id: "document_research",
    session_id:
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    sha256: "8b15fece6d99c0489f97796c199da70c6bdec878a316318a295e0288173423a0",
    size: 191,
    media_type: "text/markdown",
    received_at: "2026-10-31T20:00:00-06:00",
  },
];
const content: Record<string, string> = {
  document_proposal:
    "# Propuesta · R3\n\nExtracto local del plan actual, importado para previsualización. No es una publicación del API.\n\n## Objective\n\nPresent a session as a compact evidence timeline. Let the agent publish bounded, typed events and images to the existing local server through Hive CLI, recovering the same server session from its verified project/declared host/native session ID. Separate storage IDs from native runtime identity and preserve explicit evidence-to-ticket relationships.\n\n## Alcance de esta revisión\n\n1. Mostrar etapas explícitas de investigación, plan, implementación y verificación.\n2. Conservar revisiones inmutables de documentos Markdown.\n3. Mantener el historial hasta una eliminación manual.\n\n[Referencia de composición](https://reui.io/blocks/application/timeline)\n",
  document_design:
    "# Diseño · R3\n\n### Native composition and sketches\n\nKeep the approved timeline-4 centered feed and compact header. Use timeline-1 only for declared active/completed indicator semantics; no owner/avatar card duplication. Reference: https://reui.io/preview/base/timeline-1?ref=mcp. Reuse installed Timeline, Badge outline, Button outline size sm, Dialog/Header/Footer/Cancel. Add native Tabs from the same base-nova registry (no local restyling). Use an installed Spinner if available, otherwise the selected icon library's loading indicator with accessible text and reduced-motion support; stage state must never depend on animation alone. Do not infer process liveness from a spinner.\n\nBefore (R2):\n```text\n  [iteration] Filter correction\n              [thumbnails] [Instructions]\n  [request]   Original request\n```\nAfter (R3, newest first):\n```text\n  [spinner] Plan revision 2       [In progress]\n            Add stages and retained documents\n            [View plan]  21:40\n  [check]   Build                 [Completed]\n            [Evidence]\n  [check]   Plan revision 1       [Completed]\n            [View plan]  20:18\n  [check]   Research              [Completed]\n            [View research]\n```\nDocument dialog:\n```text\n  Plan revision 2                 [Draft]\n  [Proposal][Design][Tasks][Criteria]  (Tabs)\n  --------------------------------------------------\n  Read-only Markdown, scroll inside the dialog\n  Headings / lists / tables / code / verified links\n  Snapshot date + revision                 [Close]\n```\nTabs wrap or scroll within their own region on narrow screens; the page must not overflow. Loading uses Skeleton; empty documents offer Reload; fetch errors stay inline with Retry; unavailable/auth-denied reads retain context and use existing access recovery. Dialog size follows the approved readable instruction dialog; use native spacing/motion. Latest snapshot is default; a native revision Select permits an explicit historical selection. Opening captures an immutable snapshot; newer publication shows Update available without replacing open text. No inline editing or browser feedback controls.\n\n",
  document_tasks:
    "# Tareas · R3\n\n### T14 — Extend contracts for stage lifecycle, documents and introspection\n\n- [ ] Extend T7 Go/TS contracts and shared fixtures for stage state, foreground pointer, immutable document revisions, document approval targets, strict limits and CLI contract output.\n\n**Closes:** AC20, AC21, AC22, AC23\n\n**Depends on:** T7 (completed)\n\n**Locations:** Existing timeline contracts/tests/shared fixtures, new self-describing contract definitions.\n\n**Execution:** Settled-contract backend builder; main integrates; one independent task verifier.\n\n**Test approach:** Exact Go/TS parity and schema/validator fixtures.\n\n**Verification:** Start/complete/resume/failed/re-entry sequences; stale transitions; one foreground pointer; terminal closure handling; stable event order independent of reported occurred_at; foreground withdrawal conflict; awaiting-validation atomic pause/resume; immutable snapshots; foreign/missing document/decision references; UTF-8/byte/count bounds; offline contract examples parse with production validators; legacy fixtures remain valid. No API1 compatibility claim for unshipped draft records.\n\n### T15 — Render stage progress and document Dialog against typed stubs\n\n- [ ] Extend approved R2 composition with declared stage indicators and read-only native Tabs/Markdown snapshot Dialog.\n\n**Closes:** AC21, AC22\n\n**Depends on:** T14\n\n**Locations:** Timeline components/copy/dev fixtures/tests; native Tabs/Select source if missing; Markdown dependencies/lock; generated bundle.\n\n**Execution:** Frontend builder; main checks rendered result.\n\n**Verification:** Typed current plan revision fixture, draft/completed/paused/error states, snapshot isolation after refresh, historic selection, keyboard/focus/Tabs, headings/lists/tables/code and safe links, hostile raw HTML/URLs and remote images remain inert. Check expected publisher/advisories before installs. Focused desktop/narrow layout checks only; reuse R2 evidence outside changed regions. Independent UX and show only new stage/Dialog composition before T9.\n\n",
  document_criteria:
    '# Criterios · R3\n\n| Criterio | Resultado esperado |\n| --- | --- |\n| AC21 | Etapas con estado declarado y orden estable |\n| AC22 | Documentos inmutables, con revisión histórica |\n\n## Verificación proporcional\n\n- Revisar las regiones nuevas del timeline y el diálogo.\n- Conservar la evidencia aprobada de la composición anterior.\n- Evitar repetir recorridos completos cuando no cambian.\n\n```json\n{"stage":"plan","stage_state":"in_progress"}\n```\n',
  document_earlier:
    "# Plan · borrador inicial\n\nEsta revisión histórica permanece disponible.\n\n1. Incorporar las etapas.\n2. Definir las revisiones documentales.\n",
  document_research:
    "# Investigación\n\nSe eligió una composición nativa de timeline y documentos de solo lectura.\n\n- Usar Tabs y Select instalados.\n- Mantener las imágenes de evidencia separadas del Markdown.\n",
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
