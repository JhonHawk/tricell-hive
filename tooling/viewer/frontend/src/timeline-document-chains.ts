import type {
  TimelineDocumentSnapshot,
  TimelineDocumentTab,
  TimelineEvent,
  TimelineSession,
  TimelineStage,
} from "./timeline-contract";
export type DocumentRevision = {
  event: TimelineEvent;
  revision: TimelineDocumentSnapshot;
};
export type DocumentChain = { key: string; revisions: DocumentRevision[] };
export function documentChains(session: TimelineSession): DocumentChain[] {
  const chains = new Map<string, DocumentChain>();
  for (const event of [...session.events].sort(
    (a, b) => a.sequence - b.sequence,
  ))
    for (const revision of event.document_snapshots ?? []) {
      const chain = chains.get(revision.document_key) ?? {
        key: revision.document_key,
        revisions: [],
      };
      chain.revisions.push({ event, revision });
      chains.set(chain.key, chain);
    }
  return [...chains.values()];
}
export function latestSuccessor(
  chain: DocumentChain | undefined,
  revisionID: string,
): DocumentRevision | undefined {
  let id = revisionID,
    last: DocumentRevision | undefined;
  const seen = new Set<string>();
  while (chain && !seen.has(id)) {
    seen.add(id);
    const next = chain.revisions.find(
      (r) => r.revision.predecessor_revision_id === id,
    );
    if (!next) break;
    last = next;
    id = next.revision.document_revision_id;
  }
  return last;
}
export function compareRevisionTabs(
  revision: TimelineDocumentSnapshot,
  previous?: TimelineDocumentSnapshot,
) {
  const before = new Map(
      previous?.tabs.map((t) => [t.label, t.document_id]) ?? [],
    ),
    after = new Map(revision.tabs.map((t) => [t.label, t.document_id]));
  return {
    initial: !previous,
    added: previous
      ? revision.tabs.filter((t) => !before.has(t.label)).map((t) => t.label)
      : [],
    changed: previous
      ? revision.tabs
          .filter(
            (t) => before.has(t.label) && before.get(t.label) !== t.document_id,
          )
          .map((t) => t.label)
      : [],
    removed: previous
      ? previous.tabs.filter((t) => !after.has(t.label)).map((t) => t.label)
      : [],
  };
}
export type RevisionTab =
  | { key: string; kind: "document"; tab: TimelineDocumentTab }
  | { key: string; kind: "specs"; tabs: TimelineDocumentTab[] };
export function revisionTabs(
  revision: TimelineDocumentSnapshot,
  stage?: TimelineStage,
): RevisionTab[] {
  const canonical = (label: string) =>
    (
      ({
        proposal: "proposal",
        propuesta: "proposal",
        design: "design",
        diseño: "design",
        tasks: "tasks",
        tareas: "tasks",
      }) as Record<string, string>
    )[label.toLocaleLowerCase()];
  const structured =
    stage === "plan" &&
    ["proposal", "design", "tasks"].every((key) =>
      revision.tabs.some((t) => canonical(t.label) === key),
    );
  if (!structured)
    return revision.tabs.map((tab) => ({
      key: tab.label,
      kind: "document",
      tab,
    }));
  const specs = revision.tabs.filter((tab) => !canonical(tab.label));
  return [
    ...revision.tabs
      .filter((tab) => canonical(tab.label))
      .map((tab) => ({ key: tab.label, kind: "document" as const, tab })),
    ...(specs.length
      ? [{ key: "__specifications__", kind: "specs" as const, tabs: specs }]
      : []),
  ];
}
