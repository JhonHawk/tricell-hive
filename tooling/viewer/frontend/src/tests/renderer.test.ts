import "./dom-setup.mjs";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const modulePath = new URL(
  "../../../../../content/skills/flow-walkthrough/assets/walkthrough.mjs",
  import.meta.url,
);
const { renderWalkthrough } = await import(modulePath.href);
const manifest = {
  title: "Review",
  lang: "en",
  created_at: "2026-10-23",
  candidate: { branch: "work", commit: "working tree" },
  session: { id: "fixture", host: "Codex" },
  checks: [{ name: "Check", status: "pass" }],
  flows: [
    {
      title: "Read",
      captures: [
        { path: "desktop.png", caption: "Desktop", viewport: "desktop" },
        { path: "mobile.png", caption: "Mobile", viewport: "mobile" },
      ],
    },
  ],
};
test("renderer validates the shared delivery contract before creating links", () => {
  const fixtures = JSON.parse(
    readFileSync(
      new URL(
        "../../../../../tests/fixtures/walkthrough/delivery-contract.json",
        import.meta.url,
      ),
      "utf8",
    ),
  );
  for (const fixture of fixtures) {
    const root = document.createElement("div");
    const dispose = renderWalkthrough(root, {
      ...manifest,
      ...(fixture.omit ? {} : { delivery: fixture.delivery }),
    });
    const links = root.querySelectorAll("[data-delivery] a");
    assert.equal(root.querySelector("script"), null);
    assert.equal(
      links.length,
      fixture.field ? 0 : (fixture.delivery?.length ?? 0),
      fixture.name,
    );
    for (const link of links) {
      assert.equal(link.getAttribute("target"), "_blank");
      assert.equal(link.getAttribute("rel"), "noopener noreferrer");
      assert.equal(link.textContent, fixture.delivery[0].label);
    }
    dispose?.();
  }
});
test("two mounts isolate IDs, tabs, anchors and print listeners with disposal", (t) => {
  const added = t.mock.method(window, "addEventListener");
  const removed = t.mock.method(window, "removeEventListener");
  const originalShow = HTMLDialogElement.prototype.showModal;
  const originalClose = HTMLDialogElement.prototype.close;
  HTMLDialogElement.prototype.showModal = function () {
    this.open = true;
  };
  HTMLDialogElement.prototype.close = function () {
    this.open = false;
    this.dispatchEvent(new Event("close"));
  };
  const roots = [document.createElement("div"), document.createElement("div")];
  roots.forEach((r) => {
    document.body.append(r);
  });
  const disposers = roots.map((root, i) =>
    renderWalkthrough(root, manifest, {
      idPrefix: `segment-${i}-`,
      embedded: true,
    }),
  );
  const ids = roots.flatMap((r) =>
    [...r.querySelectorAll("[id]")].map((n) => n.id),
  );
  assert.equal(new Set(ids).size, ids.length);
  for (const root of roots)
    for (const tab of root.querySelectorAll("[aria-controls]"))
      assert(root.querySelector(`[id="${tab.getAttribute("aria-controls")}"]`));
  (roots[1].querySelectorAll('[role="tab"]')[1] as HTMLElement).click();
  assert.equal(
    roots[0].querySelector('[role="tab"]')?.getAttribute("aria-selected"),
    "true",
  );
  const scrolled: Element[] = [];
  for (const root of roots) {
    const section = root.querySelector<HTMLElement>(".wt-flow");
    assert(section);
    section.scrollIntoView = () => {
      scrolled.push(section);
    };
  }
  roots[1].querySelector<HTMLAnchorElement>(".wt-index a")?.click();
  assert.equal(scrolled.length, 1);
  assert.equal(scrolled[0], roots[1].querySelector(".wt-flow"));
  assert.equal(document.activeElement, roots[1].querySelector(".wt-flow h3"));
  const dialogs = roots.map((root) => {
    const dialog = root.querySelector<HTMLDialogElement>("dialog");
    assert(dialog);
    return dialog;
  });
  const thumbs = roots.map((root) => {
    const thumb = root.querySelector<HTMLButtonElement>(
      ".wt-strip:not([hidden]) .wt-thumb",
    );
    assert(thumb);
    return thumb;
  });
  thumbs.forEach((thumb) => {
    thumb.click();
  });
  assert(dialogs.every((dialog) => dialog.open));
  dialogs[0].querySelector<HTMLButtonElement>(".wt-close")?.click();
  assert.equal(dialogs[0].open, false);
  assert.equal(dialogs[1].open, true);
  assert.equal(document.activeElement, thumbs[0]);
  dialogs[1].close();
  assert.equal(document.activeElement, thumbs[1]);
  const details = roots.map((r) => {
    const node = r.querySelector("details");
    assert(node);
    return node;
  });
  window.dispatchEvent(new Event("beforeprint"));
  assert(details.every((d) => d.open));
  window.dispatchEvent(new Event("afterprint"));
  assert(details.every((d) => !d.open));
  disposers.forEach((dispose) => {
    assert.equal(typeof dispose, "function");
    dispose();
  });
  window.dispatchEvent(new Event("beforeprint"));
  assert(details.every((d) => !d.open));
  for (const event of ["beforeprint", "afterprint"]) {
    const registrations = added.mock.calls.filter(
      (call) => call.arguments[0] === event,
    );
    const removals = removed.mock.calls.filter(
      (call) => call.arguments[0] === event,
    );
    assert.equal(registrations.length, 2);
    assert.equal(removals.length, registrations.length);
    for (const call of registrations)
      assert(
        removals.some((removal) => removal.arguments[1] === call.arguments[1]),
      );
  }
  roots.forEach((r) => {
    r.remove();
  });
  HTMLDialogElement.prototype.showModal = originalShow;
  HTMLDialogElement.prototype.close = originalClose;
});

test("frontend consumes the served renderer without a bundled canonical-source import", () => {
  const source = readFileSync(
    new URL("../deliveries.tsx", import.meta.url),
    "utf8",
  );
  assert.doesNotMatch(
    source,
    /from\s+["'][^"']*content\/skills\/flow-walkthrough/,
  );
});
