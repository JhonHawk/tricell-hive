import type { LibrarySession } from "./api";

export type SessionGroup = {
  id: string;
  projectKey: string;
  project: string;
  title: string;
  tickets: string[];
  segments: LibrarySession[];
  activity: number;
  inProgress: boolean;
};
export function sessionIdentity(segment: LibrarySession): string {
  return JSON.stringify([
    segment.projectKey,
    segment.raw.host ?? null,
    segment.raw.session_id,
  ]);
}
export function segmentTime(segment: LibrarySession): number {
  const created = Date.parse(segment.raw.created_at);
  return Number.isNaN(created)
    ? Date.parse(segment.raw.ingested_at) || 0
    : created;
}
export function groupSessions(segments: LibrarySession[]): SessionGroup[] {
  const groups = new Map<string, LibrarySession[]>();
  for (const segment of segments) {
    if (segment.excluded || segment.classification === "tracked") continue;
    const id = sessionIdentity(segment);
    const siblings = groups.get(id) ?? [];
    siblings.push(segment);
    groups.set(id, siblings);
  }
  return [...groups]
    .map(([id, siblings]) => {
      const segments = siblings.sort(
        (a, b) => segmentTime(a) - segmentTime(b) || a.id.localeCompare(b.id),
      );
      const first = segments[0];
      return {
        id,
        projectKey: first.projectKey ?? "",
        project: first.project,
        title:
          segments.find((segment) => segment.raw.session_title)?.raw
            .session_title || first.title,
        tickets: [...new Set(segments.map((segment) => segment.ticket))],
        segments,
        activity: Math.max(...segments.map(segmentTime)),
        inProgress: segments.some(
          (segment) => segment.status === "in_progress",
        ),
      };
    })
    .sort((a, b) => b.activity - a.activity || a.id.localeCompare(b.id));
}
export function matchingSegments(
  group: SessionGroup,
  query: string,
): LibrarySession[] {
  const search = query.trim().toLocaleLowerCase();
  if (group.title.toLocaleLowerCase().includes(search)) return group.segments;
  return group.segments.filter((segment) =>
    `${segment.title} ${segment.ticket}`.toLocaleLowerCase().includes(search),
  );
}
