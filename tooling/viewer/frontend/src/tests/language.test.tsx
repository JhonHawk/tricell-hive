import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { completeFixture } from "../dev/timeline-complete-fixture";
import { browserLanguage, resolveLanguage } from "../language";
import { TimelineContent } from "../timeline-content";
import {
  timelineCopy,
  timelineDate,
  timelineEventTitle,
} from "../timeline-copy";

test("primary browser language accepts Spanish subtags, otherwise English", () => {
  for (const input of ["es", "es-MX", "ES-es", "eS-419"])
    assert.equal(resolveLanguage(input), "es");
  for (const input of [
    undefined,
    null,
    3,
    "",
    "es_ MX",
    "es_419",
    "est",
    "es-",
    " es",
    "fr-FR",
    "en-US",
    "es-!",
    "es-a",
    "es-12",
    "es-123456789",
  ])
    assert.equal(resolveLanguage(input), "en");
  assert.equal(browserLanguage(), resolveLanguage(navigator.language));
});

test("reserved headings localize by stage while context and non-stage titles remain verbatim", () => {
  const data = completeFixture();
  const stage = data.session.events.find((event) => event.stage === "research");
  assert.ok(stage);
  for (const lang of ["en", "es"] as const) {
    assert.equal(
      timelineEventTitle(stage, lang),
      timelineCopy(lang).stageNames.research,
    );
    assert.equal(
      timelineEventTitle(
        { ...stage, title: "Análisis · customer context" },
        lang,
      ),
      `${timelineCopy(lang).stageNames.research} · Análisis · customer context`,
    );
    assert.equal(
      timelineEventTitle(
        { ...stage, kind: "request", title: "Investigación" },
        lang,
      ),
      "Investigación",
    );
    // Old text is preserved too: no heuristic translation stripping.
    assert.equal(
      timelineEventTitle({ ...stage, title: "Investigación" }, lang),
      `${timelineCopy(lang).stageNames.research} · Investigación`,
    );
    assert.equal(timelineDate("invalid", lang), "invalid");
  }
});

test("the same seven tasks and immutable producer content render in both locales", async () => {
  const data = completeFixture();
  const before = JSON.stringify(data);
  const stage = data.session.events.find((event) => event.stage === "research");
  assert.ok(stage);
  stage.title = "Contexto literal · revisión";
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  for (const lang of ["en", "es"] as const) {
    await act(async () =>
      root.render(
        <TimelineContent data={data} lang={lang} onRetry={() => {}} />,
      ),
    );
    const text = container.textContent ?? "";
    assert.ok(
      text.includes(
        `${timelineCopy(lang).stageNames.research} · Contexto literal · revisión`,
      ),
    );
    assert.ok(text.includes(timelineCopy(lang).goCurrent));
    assert.ok(text.includes(data.session.events[0].title));
    assert.equal(data.session.tasks.length, 7);
  }
  await act(async () => root.unmount());
  container.remove();
  stage.title = "";
  assert.equal(JSON.stringify(data), before);
});
