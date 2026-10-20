// Development-only fixture. Vite eliminates this module from production.
import { useEffect, useRef, useState } from "react";
import { type ViewerSession, ViewerShell } from "../shell";

const sessions: ViewerSession[] = [
  {
    id: "preview-navigation-review",
    project: "Hive",
    ticket: "#89 · React viewer",
    title: "Navigation and document reading",
    createdAt: "19 Oct 2026 · 14:32",
    classification: "local",
    pinned: true,
    status: "in_progress",
  },
  {
    id: "preview-mobile-review",
    project: "Hive",
    ticket: "#89 · React viewer",
    title: "Mobile navigation and focus recovery",
    createdAt: "19 Oct 2026 · 13:47",
    classification: "tracked",
    status: "closed",
  },
  {
    id: "preview-retention-review",
    project: "Hive",
    ticket: "#82 · Session retention",
    title: "Retention notices and copied evidence",
    createdAt: "18 Oct 2026 · 17:16",
    classification: "unknown",
    status: "unknown",
  },
];
export function DevelopmentPreview() {
  const [selectedId, setSelectedId] = useState(sessions[0].id);
  const state = new URLSearchParams(location.search).get("state");
  const ref = useRef<HTMLDivElement>(null);
  const showDocument =
    !state &&
    sessions.find((s) => s.id === selectedId)?.classification !== "tracked";
  useEffect(() => {
    let cancelled = false;
    const path = "/assets/walkthrough.mjs";
    void import(/* @vite-ignore */ path).then(({ renderWalkthrough }) => {
      if (cancelled || !ref.current || !showDocument) return;
      renderWalkthrough(
        ref.current,
        {
          kind: "first_view",
          lang: "en",
          title: sessions.find((s) => s.id === selectedId)?.title,
          created_at: "2026-10-19",
          ticket: "#89",
          candidate: {
            branch: "feat/gh-89-react-reui-viewer",
            commit: "development preview",
          },
          session: { id: selectedId, host: "Local preview" },
          result:
            "Review the navigation, reading space and provenance notices before connecting the live library.",
          flows: [],
          checks: [],
          environment: [
            { label: "Surface", value: "Development-only composition" },
          ],
          provenance: [
            "Generated example; no registered repositories or stored session data.",
          ],
        },
        { fileBase: "/" },
      );
    });
    return () => {
      cancelled = true;
    };
  }, [selectedId, showDocument]);
  return (
    <ViewerShell
      sessions={state === "empty" ? [] : sessions}
      selectedId={selectedId}
      onSelect={setSelectedId}
      state={
        state === "error"
          ? "error"
          : state === "loading"
            ? "loading"
            : state === "empty"
              ? "empty"
              : "ready"
      }
      onRetry={() => location.reload()}
      notice="Development preview · generated example data. The live library is not connected yet."
    >
      {showDocument && <div ref={ref} className="walkthrough" />}
    </ViewerShell>
  );
}
