// Development-only generated fixture; no repository data or writes.
import { useState } from "react";
import { AccessClient } from "../api";
import { ViewerApplication } from "../application";
import { ViewerController } from "../controller";

const keys = ["1", "2", "3"].map((n) => n.repeat(64));
const segment = (i: number) => ({
  key: keys[i],
  source: "/generated/preview",
  source_status: "local",
  content_hash: "preview",
  created_at: `2026-10-${i === 2 ? 20 : 22 + i}T12:00:00Z`,
  ingested_at: "2026-10-23T12:00:00Z",
  session_id: i < 2 ? "preview-session" : "preview-other-session",
  host: "Codex",
  title:
    i === 0
      ? "Navigation and document reading"
      : i === 1
        ? "Delivery links and continuation reading"
        : "Retention notices and copied evidence",
  session_status: i === 1 ? "in_progress" : "closed",
  fileBase: `/files/${keys[i]}/${"a".repeat(64)}/`,
});
const manifests = keys.map((_, i) => ({
  kind: "first_view",
  lang: "en",
  title: segment(i).title,
  created_at: segment(i).created_at,
  ticket: i === 0 ? "#89" : "#90",
  candidate: {
    branch: "feat/viewer-session-segments",
    commit: "development preview",
  },
  session: { id: segment(i).session_id, host: "Codex" },
  result:
    i === 0
      ? "Review the navigation, reading space and provenance notices."
      : "Read the continuation in this session and review its recorded delivery.",
  delivery: (i === 0 ? [1, 2] : [2, 3, 4]).map((n) => ({
    kind: n === 4 ? "commit" : "pull_request",
    label: n === 4 ? "hive · abc1234" : `hive #${89 + n}`,
    url: `https://example.com/hive/${n === 4 ? "commit/abc1234" : `pull/${89 + n}`}`,
  })),
  flows: [],
  checks: [],
  environment: [{ label: "Surface", value: "Development-only composition" }],
  provenance: [
    "Generated example; no registered repositories or stored session data.",
  ],
}));
export function DevelopmentPreview() {
  const [controller] = useState(() => {
    const state = new URLSearchParams(location.search).get("state");
    const fixture = {
      projects:
        state === "empty"
          ? []
          : [
              {
                key: "hive",
                name: "Hive",
                tickets: [
                  { ticket: "#89", sessions: [segment(0)] },
                  { ticket: "#90", sessions: [segment(1), segment(2)] },
                ],
              },
            ],
      excluded_sessions: [],
      repositories: [],
      errors: [],
      ttl_seconds: 86400,
    };
    const client = new AccessClient(
      async (url) => {
        if (state === "loading") return new Promise<Response>(() => {});
        if (state === "error") return new Response(null, { status: 500 });
        return new Response(
          JSON.stringify(
            String(url) === "/api/index"
              ? fixture
              : manifests[keys.indexOf(String(url).split("/").at(-1) ?? "")],
          ),
        );
      },
      { getItem: () => "preview", setItem() {}, removeItem() {} },
    );
    const controller = new ViewerController(client, () => false);
    void controller.open(null);
    return controller;
  });
  return <ViewerApplication controller={controller} />;
}
