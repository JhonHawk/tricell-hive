// Development-only API boundary: real application/controller with typed illustrative work.
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { AccessClient } from "../api";
import { ViewerApplication } from "../application";
import { ViewerController } from "../controller";
import { loadPreviewDocument } from "./timeline-documents-fixture";
import {
  navigationFixture,
  navigationIndexEntry,
} from "./timeline-navigation-fixture";

export function TimelineNavigationPreview() {
  const [{ controller, works, mismatch }, setPreview] = useState(() => {
    const works = navigationFixture();
    const mismatch = {
      value: new URLSearchParams(location.search).get("mismatch") === "initial",
    };
    const restored = new Set<string>();
    const response = (body: unknown, status = 200) =>
      new Response(JSON.stringify(body), {
        status,
        headers: { "X-Hive-Timeline-Revision": "3" },
      });
    const controller = new ViewerController(
      new AccessClient(
        async (path, options) => {
          if (mismatch.value) return response({ code: "reload_required" }, 409);
          if (path === "/api/health") return response({});
          if (path === "/api/index?navigation=grouped")
            return response({
              protocol_version: 3,
              retention: "manual",
              ttl_seconds: 0,
              repositories: [],
              errors: [],
              timeline_sessions: works.map(navigationIndexEntry),
            });
          if (String(path).includes("/documents/"))
            return response(
              await loadPreviewDocument(String(path).split("/").at(-1) ?? ""),
            );
          const work = works.find(({ session }) =>
            String(path).endsWith(session.viewer_session_id),
          );
          if (!work) return response({ code: "not_found" }, 404);
          if (options?.method === "PATCH") {
            const body = JSON.parse(String(options.body));
            if (body.expected_version !== work.session.version)
              return response({ code: "version_conflict" }, 409);
            Object.assign(work.session, {
              ...("title" in body ? { title: body.title } : {}),
              ...("pinned" in body ? { pinned: body.pinned } : {}),
              ...("archived" in body
                ? {
                    archived: body.archived,
                    archive_origin: body.archived ? "manual" : null,
                  }
                : {}),
              version: work.session.version + 1,
            });
            if (body.archived === false)
              restored.add(work.session.viewer_session_id);
            work.session.autoarchive_at =
              work.session.state === "closed" &&
              !work.session.archived &&
              !work.session.pinned &&
              !restored.has(work.session.viewer_session_id) &&
              work.session.closed_at
                ? new Date(
                    Date.parse(work.session.closed_at) + 7_200_000,
                  ).toISOString()
                : null;
            return response({
              schema_version: 3,
              kind: "session_patch",
              operation_id: body.operation_id,
              viewer_session_id: work.session.viewer_session_id,
              session_version: work.session.version,
              current_stage_event_id: work.session.current_stage_event_id,
            });
          }
          return response(work);
        },
        {
          getItem: (key) =>
            key === "hive-view-token" ? "development-only" : null,
          setItem() {},
          removeItem() {},
        },
      ),
      () => false,
    );
    return { controller, works, mismatch };
  });
  useEffect(() => {
    document.documentElement.classList.toggle(
      "dark",
      new URLSearchParams(location.search).get("theme") !== "light",
    );
    void controller.open(null);
  }, [controller]);
  return (
    <div className="flex h-dvh min-h-0 flex-col">
      <div className="flex shrink-0 flex-wrap gap-2">
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => {
            const work = works.find(
              ({ session }) =>
                session.viewer_session_id ===
                controller.getSnapshot().selectedId,
            );
            if (!work) return;
            const stage = work.session.events.find(
              (event) => event.event_id === work.session.current_stage_event_id,
            );
            if (stage) {
              stage.stage_state = "completed";
              stage.finished_at = stage.updated_at;
            }
            work.session.state = "closed";
            work.session.current_stage_event_id = null;
            work.session.closed_at = work.session.last_activity_at;
            work.session.autoarchive_at = null;
            work.session.archived = true;
            work.session.archive_origin = "automatic";
            work.session.version++;
            void controller.refresh();
          }}
        >
          Simulate automatic archive
        </Button>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => {
            mismatch.value = true;
            void controller.refresh();
            setPreview((previous) => ({ ...previous }));
          }}
        >
          Simulate revision mismatch
        </Button>
      </div>
      <ViewerApplication controller={controller} />
    </div>
  );
}
