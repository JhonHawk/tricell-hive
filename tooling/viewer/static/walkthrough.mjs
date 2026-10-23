// Walkthrough viewer module. No imports, no side effects on import.
// Everything is built with createElement, textContent and setAttribute:
// manifest text never reaches innerHTML, and no style or on* attribute is set.

const STATUS_ORDER = ["pass", "fail", "warn", "info", "not_run"];

const LABELS = {
  es: {
    kicker: "Recorrido de UI",
    summary: "Resumen",
    index: "Índice",
    checks: "Comprobaciones",
    flows: "Flujos",
    steps: "Pasos",
    captures: "Capturas",
    recording: "Grabación",
    action: "Acción requerida",
    environment: "Entorno",
    provenance: "Procedencia",
    check: "Comprobación",
    result: "Resultado",
    flowOne: "flujo", flowMany: "flujos",
    captureOne: "captura", captureMany: "capturas",
    previous: "Anterior", next: "Siguiente", close: "Cerrar",
    carousel: "Carrusel de capturas",
    notRecorded: "no grabado",
    stale: "obsoleto — la grabación es anterior al candidato; grabada en",
    duration: "Duración",
    recordedOn: "grabada en el candidato",
    ticket: "Ticket", branch: "Rama", candidate: "Candidato", session: "Sesión", host: "Host", date: "Fecha",
    desktop: "Escritorio", mobile: "Móvil", viewSwitch: "Vista de las capturas", strip: "Capturas",
    videoFallback: "Tu navegador no puede reproducir esta grabación; abre el archivo directamente.",
  },
  en: {
    kicker: "UI walkthrough",
    summary: "Summary",
    index: "Index",
    checks: "Checks",
    flows: "Flows",
    steps: "Steps",
    captures: "Captures",
    recording: "Recording",
    action: "Action required",
    environment: "Environment",
    provenance: "Provenance",
    check: "Check",
    result: "Result",
    flowOne: "flow", flowMany: "flows",
    captureOne: "capture", captureMany: "captures",
    previous: "Previous", next: "Next", close: "Close",
    carousel: "Capture carousel",
    notRecorded: "not recorded",
    stale: "stale — the recording predates the candidate; recorded on",
    duration: "Duration",
    recordedOn: "recorded on candidate",
    ticket: "Ticket", branch: "Branch", candidate: "Candidate", session: "Session", host: "Host", date: "Date",
    desktop: "Desktop", mobile: "Mobile", viewSwitch: "Capture view", strip: "Captures",
    videoFallback: "Your browser cannot play this recording; open the file directly.",
  },
};

function el(tag, attrs, ...children) {
  const node = document.createElement(tag);
  for (const [name, value] of Object.entries(attrs || {})) {
    if (value === undefined || value === null || value === false) continue;
    node.setAttribute(name, value === true ? "" : String(value));
  }
  for (const child of children) {
    if (child === undefined || child === null) continue;
    node.append(typeof child === "string" ? document.createTextNode(child) : child);
  }
  return node;
}

function fileUrl(base, path) {
  return base + path.split("/").map(encodeURIComponent).join("/");
}

function formatDuration(seconds) {
  const total = Math.round(seconds);
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, "0")}`;
}

function plural(count, one, many) {
  return `${count} ${count === 1 ? one : many}`;
}

// ---------------------------------------------------------------- blocks
//
// The page is an ordered list of blocks. Each block is { id, applies, render },
// optionally with `index`:
//   applies(ctx) -> boolean            the block appears only when this is true
//   render(ctx)  -> Node | Node[]      what to append to the root, in list order
//   index(ctx)   -> [{ id, label }]    optional: entries this block adds to the index
// `ctx` carries { root, manifest, t, fileBase, firstView, flows, checks, captureTotal, el },
// and `indexEntries` once the applying blocks are known.
//
// To add a block (for example a new section of the page):
//   1. Write its render function below (build DOM with ctx.el, textContent and
//      setAttribute only; never innerHTML or style attributes).
//   2. Add one entry to BLOCKS at the position where it should appear.
//   3. If it needs data, add the field to the manifest with validation in
//      render_walkthrough.py and in tooling/viewer/manifest.go, and one line in
//      references/walkthrough-manifest.md. Add its labels to LABELS (es and en).
//   4. Copy this file and walkthrough.css to tooling/viewer/static/ (a Go test
//      compares the bytes).
// renderWalkthrough has no sections written by hand: the viewer mounts the page
// through it too.

const prefersReducedMotion = () =>
  typeof matchMedia === "function" && matchMedia("(prefers-reduced-motion: reduce)").matches;

function renderHeader(ctx) {
  const { manifest, t } = ctx;
  const meta = ctx.el("div", { class: "wt-meta" });
  const addMeta = (label, value, title) => {
    if (value) meta.append(ctx.el("span", { title }, `${label}: ${value}`));
  };
  addMeta(t.date, manifest.created_at);
  addMeta(t.ticket, manifest.ticket);
  addMeta(t.branch, manifest.candidate.branch);
  addMeta(t.candidate, manifest.candidate.commit);
  const sessionId = manifest.session.id;
  addMeta(t.session, sessionId.length > 18 ? `${sessionId.slice(0, 10)}…${sessionId.slice(-6)}` : sessionId, sessionId);
  addMeta(t.host, manifest.session.host);
  return ctx.el("header", {}, ctx.el("p", { class: "wt-kicker" }, t.kicker), ctx.el("h1", {}, manifest.title), meta);
}

function renderSummary(ctx) {
  const { manifest, t, checks, flows, firstView, captureTotal } = ctx;
  const summary = ctx.el("div", { class: "wt-summary" });
  if (manifest.result) summary.append(ctx.el("p", { class: "wt-result" }, manifest.result));
  if (checks.length && !firstView) {
    const counts = ctx.el("ul", { class: "wt-counts" });
    for (const status of STATUS_ORDER) {
      const count = checks.filter((check) => check.status === status).length;
      if (count) counts.append(ctx.el("li", {}, ctx.el("span", { class: `wt-chip ${status}` }, `${status} ${count}`)));
    }
    summary.append(counts);
  }
  summary.append(
    ctx.el("p", { class: "wt-totals" }, `${plural(flows.length, t.flowOne, t.flowMany)} · ${plural(captureTotal, t.captureOne, t.captureMany)}`)
  );
  return [ctx.el("h2", {}, t.summary), summary];
}

// The index lists what the other blocks put in it (see `index` on a block).
function renderIndex(ctx) {
  const { root, t } = ctx;
  const list = ctx.el("ol", { class: "wt-index" });
  for (const entry of ctx.indexEntries) {
    const link = ctx.el("a", { href: `#${entry.id}` }, entry.label);
    link.addEventListener("click", (event) => {
      event.preventDefault();
      const target = root.querySelector(`#${entry.id}`);
      if (!target) return;
      target.scrollIntoView({ behavior: prefersReducedMotion() ? "auto" : "smooth", block: "start" });
      const heading = target.querySelector("h3");
      if (heading) heading.focus({ preventScroll: true });
    });
    list.append(ctx.el("li", {}, link));
  }
  return [ctx.el("h2", {}, t.index), ctx.el("nav", { "aria-label": t.index }, list)];
}

function renderChecks(ctx) {
  const { t, checks } = ctx;
  const body = ctx.el("tbody");
  for (const check of checks) {
    body.append(
      ctx.el("tr", {},
        ctx.el("td", {}, check.name),
        ctx.el("td", {}, ctx.el("span", { class: `wt-chip ${check.status}` }, check.status), check.evidence ? ` ${check.evidence}` : "")
      )
    );
  }
  const table = ctx.el("table", {},
    ctx.el("thead", {}, ctx.el("tr", {}, ctx.el("th", { scope: "col" }, t.check), ctx.el("th", { scope: "col" }, t.result))),
    body
  );
  // A native <details>, closed by default: the summary is the section heading plus the per-status counts.
  const counts = ctx.el("span", { class: "wt-sum-counts" });
  for (const status of STATUS_ORDER) {
    const count = checks.filter((check) => check.status === status).length;
    if (count) counts.append(ctx.el("span", { class: `wt-chip ${status}` }, `${status} ${count}`));
  }
  return ctx.el("details", { class: "wt-checks" },
    ctx.el("summary", {}, ctx.el("h2", { class: "wt-checks-title" }, t.checks), counts),
    ctx.el("div", { class: "wt-tablewrap", tabindex: "0", role: "region", "aria-label": ctx.t.checks }, table)
  );
}

// Printing always shows the table: collapsed sections are opened for the print and restored after it.
let printHooked = false;
function hookPrint() {
  if (printHooked || typeof window === "undefined") return;
  printHooked = true;
  const opened = new Set();
  const expand = () => {
    for (const node of document.querySelectorAll("details.wt-checks")) {
      if (!node.open) { node.open = true; opened.add(node); }
    }
  };
  const restore = () => {
    for (const node of opened) node.open = false;
    opened.clear();
  };
  window.addEventListener("beforeprint", expand);
  window.addEventListener("afterprint", restore);
  if (typeof matchMedia === "function") {
    matchMedia("print").addEventListener("change", (event) => (event.matches ? expand() : restore()));
  }
}

function renderFlows(ctx) {
  const { t, flows, firstView, fileBase } = ctx;
  const carousel = buildCarousel(t, fileBase);
  const nodes = carousel.dialog ? [carousel.dialog] : [];
  nodes.push(ctx.el("h2", {}, t.flows));
  flows.forEach((flow, i) => {
    const section = ctx.el("section", { class: "wt-flow", id: `wt-flow-${i}`, "aria-labelledby": `wt-flow-${i}-h` });
    section.append(ctx.el("h3", { id: `wt-flow-${i}-h`, tabindex: "-1" }, flow.title));
    if (!firstView && flow.steps && flow.steps.length) {
      section.append(ctx.el("p", { class: "wt-label" }, t.steps), ctx.el("ol", { class: "wt-steps" }, ...flow.steps.map((step) => ctx.el("li", {}, step))));
    }
    if (flow.captures.length) {
      const tabbed = flow.captures.some((c) => c.viewport === "mobile") && flow.captures.some((c) => c.viewport !== "mobile");
      // With a view switch the tabs carry the counts; otherwise the label does.
      section.append(ctx.el("p", { class: "wt-label" }, tabbed ? t.captures : `${t.captures} · ${flow.captures.length}`), buildGallery(flow, i, carousel, t, fileBase));
    }
    const recording = flow.recording;
    if (!firstView && recording && recording.status !== "none") {
      section.append(ctx.el("p", { class: "wt-label" }, t.recording));
      if (recording.status === "recorded") {
        const poster = flow.captures[0] ? fileUrl(fileBase, flow.captures[0].path) : undefined;
        section.append(
          ctx.el("video", { controls: true, preload: "metadata", poster, src: fileUrl(fileBase, recording.path), "aria-label": `${t.recording}: ${flow.title}` }, t.videoFallback),
          ctx.el("p", { class: "wt-vmeta" }, `${t.duration} ${formatDuration(recording.duration_seconds)} · ${t.recordedOn} ${recording.candidate}`)
        );
      } else if (recording.status === "stale") {
        section.append(ctx.el("p", { class: "wt-mark stale" }, `${t.stale} ${recording.candidate}.`));
      } else if (recording.status === "not_recorded") {
        section.append(ctx.el("p", { class: "wt-mark" }, `${t.notRecorded} — ${recording.reason}`));
      }
    }
    nodes.push(section);
  });
  return nodes;
}

function renderAction(ctx) {
  return [
    ctx.el("h2", {}, ctx.t.action),
    ctx.el("div", { class: "wt-action" }, ctx.el("ul", {}, ...ctx.manifest.action_required.map((item) => ctx.el("li", {}, item)))),
  ];
}

function renderEnvironment(ctx) {
  const list = ctx.el("dl", { class: "wt-env" });
  for (const entry of ctx.manifest.environment) list.append(ctx.el("dt", {}, entry.label), ctx.el("dd", {}, entry.value));
  return [ctx.el("h2", {}, ctx.t.environment), list];
}

function renderProvenance(ctx) {
  return ctx.el("details", { class: "wt-prov", open: true },
    ctx.el("summary", {}, ctx.t.provenance),
    ctx.el("ul", {}, ...ctx.manifest.provenance.map((item) => ctx.el("li", {}, item)))
  );
}

const hasItems = (list) => Boolean(list && list.length);

const BLOCKS = [
  { id: "header", applies: () => true, render: renderHeader },
  { id: "summary", applies: () => true, render: renderSummary },
  { id: "index", applies: (ctx) => ctx.indexEntries.length > 0, render: renderIndex },
  { id: "checks", applies: (ctx) => ctx.checks.length > 0 && !ctx.firstView, render: renderChecks },
  {
    id: "flows",
    applies: () => true,
    render: renderFlows,
    index: (ctx) => ctx.flows.map((flow, i) => ({ id: `wt-flow-${i}`, label: flow.title })),
  },
  { id: "action", applies: (ctx) => hasItems(ctx.manifest.action_required), render: renderAction },
  { id: "environment", applies: (ctx) => hasItems(ctx.manifest.environment), render: renderEnvironment },
  { id: "provenance", applies: (ctx) => hasItems(ctx.manifest.provenance), render: renderProvenance },
];

export function renderWalkthrough(root, manifest, { fileBase = "" } = {}) {
  const flows = manifest.flows;
  const ctx = {
    root, manifest, fileBase, flows, el,
    t: LABELS[manifest.lang] || LABELS.en,
    firstView: manifest.kind === "first_view",
    checks: manifest.checks || [],
    captureTotal: flows.reduce((sum, flow) => sum + flow.captures.length, 0),
    indexEntries: [],
  };
  ctx.indexEntries = BLOCKS.filter((block) => block.index && block.applies(ctx)).flatMap((block) => block.index(ctx));

  hookPrint();
  root.replaceChildren();
  for (const block of BLOCKS) {
    if (!block.applies(ctx)) continue;
    const out = block.render(ctx);
    root.append(...(Array.isArray(out) ? out : [out]));
  }
}


const VIEWS = ["desktop", "mobile"];

// Capture strip of one flow. A flow with desktop and mobile captures gets a tablist above
// the strip instead of two flows; the carousel only walks the active view.
function buildGallery(flow, flowIndex, carousel, t, fileBase) {
  const groups = { desktop: [], mobile: [] };
  for (const capture of flow.captures) groups[capture.viewport === "mobile" ? "mobile" : "desktop"].push(capture);
  const views = VIEWS.filter((view) => groups[view].length);
  const gallery = el("div", { class: "wt-gallery" });
  const tabs = {};
  const panels = {};
  const thumbsByView = {};

  for (const view of views) {
    const captures = groups[view];
    const list = el("ul", { class: "wt-strip-list" });
    const thumbs = (thumbsByView[view] = []);
    captures.forEach((capture, c) => {
      const alt = capture.alt || capture.caption;
      const thumb = el("button", { type: "button", class: "wt-thumb", "aria-label": `${capture.caption} (${c + 1} / ${captures.length})` },
        el("img", {
          src: fileUrl(fileBase, capture.path), alt, loading: "lazy",
          width: view === "mobile" ? "112" : "320", height: "200",
        })
      );
      thumbs.push(thumb);
      thumb.addEventListener("click", () => carousel.open(captures, c, thumbs));
      list.append(el("li", { class: `wt-item ${view}` }, el("figure", {}, thumb, el("figcaption", { title: capture.caption }, capture.caption))));
    });
    // The scroller is focusable so the arrow keys scroll it; its label announces the count.
    panels[view] = el("div", {
      class: "wt-strip", role: "region", tabindex: "0",
      "aria-label": `${t.strip}: ${captures.length}`,
      id: `wt-panel-${flowIndex}-${view}`,
    }, list);
  }

  if (views.length > 1) {
    const tablist = el("div", { class: "wt-tabs", role: "tablist", "aria-label": t.viewSwitch });
    const select = (view, focus) => {
      for (const other of views) {
        const on = other === view;
        tabs[other].setAttribute("aria-selected", on ? "true" : "false");
        tabs[other].setAttribute("tabindex", on ? "0" : "-1");
        if (on) panels[other].removeAttribute("hidden"); else panels[other].setAttribute("hidden", "");
      }
      if (focus) tabs[view].focus();
    };
    for (const view of views) {
      const tab = el("button", {
        type: "button", role: "tab", class: "wt-tab", id: `wt-tab-${flowIndex}-${view}`,
        "aria-controls": `wt-panel-${flowIndex}-${view}`,
      }, `${t[view]} · ${groups[view].length}`);
      tab.addEventListener("click", () => select(view, false));
      tab.addEventListener("keydown", (event) => {
        const at = views.indexOf(view);
        let target = null;
        if (event.key === "ArrowRight") target = views[(at + 1) % views.length];
        else if (event.key === "ArrowLeft") target = views[(at - 1 + views.length) % views.length];
        else if (event.key === "Home") target = views[0];
        else if (event.key === "End") target = views[views.length - 1];
        if (target) { event.preventDefault(); select(target, true); }
      });
      tabs[view] = tab;
      tablist.append(tab);
    }
    for (const view of views) panels[view].setAttribute("aria-labelledby", `wt-tab-${flowIndex}-${view}`);
    gallery.append(tablist);
    select(views[0], false);
  }
  for (const view of views) gallery.append(panels[view]);
  return gallery;
}

function buildCarousel(t, fileBase) {
  const dialog = el("dialog", { class: "wt-carousel", "aria-label": t.carousel });
  if (typeof dialog.showModal !== "function") return { dialog: null, open() {} };
  const counter = el("span", { class: "wt-counter", "aria-live": "polite" });
  const closeButton = el("button", { type: "button", class: "wt-close" }, t.close);
  const prev = el("button", { type: "button", class: "wt-prev", "aria-label": t.previous }, "‹");
  const next = el("button", { type: "button", class: "wt-next", "aria-label": t.next }, "›");
  const image = el("img", { alt: "" });
  // The bar and the caption stay fixed; only the body scrolls (tall mobile captures).
  const body = el("div", { class: "wt-body", tabindex: "0" }, image);
  const caption = el("p", { class: "wt-caption" });
  dialog.append(el("div", { class: "wt-bar" }, prev, counter, next, closeButton), body, caption);

  let captures = [];
  let position = 0;
  let thumbs = null;

  function show(index) {
    const focused = [prev, next].find((button) => button === dialog.ownerDocument.activeElement);
    position = Math.max(0, Math.min(captures.length - 1, index));
    const capture = captures[position];
    dialog.setAttribute("data-view", capture.viewport === "mobile" ? "mobile" : "desktop");
    image.setAttribute("src", fileUrl(fileBase, capture.path));
    image.setAttribute("alt", capture.alt || capture.caption);
    body.setAttribute("aria-label", capture.alt || capture.caption);
    body.scrollTop = 0;
    caption.textContent = capture.caption;
    counter.textContent = `${position + 1} / ${captures.length}`;
    prev.disabled = position === 0;
    next.disabled = position === captures.length - 1;
    // A disabled button cannot keep focus: move it to a usable control.
    if (focused && focused.disabled) {
      (prev.disabled ? (next.disabled ? closeButton : next) : prev).focus();
    }
  }

  prev.addEventListener("click", () => show(position - 1));
  next.addEventListener("click", () => show(position + 1));
  closeButton.addEventListener("click", () => dialog.close());
  dialog.addEventListener("click", (event) => {
    if (event.target === dialog) dialog.close();
  });
  dialog.addEventListener("keydown", (event) => {
    if (event.key === "ArrowLeft") { event.preventDefault(); show(position - 1); }
    else if (event.key === "ArrowRight") { event.preventDefault(); show(position + 1); }
    else if (event.target !== body) {
      // The body scrolls natively when focused; from a button, scroll it by hand.
      const step = { ArrowUp: -80, ArrowDown: 80, PageUp: -body.clientHeight * 0.9, PageDown: body.clientHeight * 0.9 }[event.key];
      if (step) { event.preventDefault(); body.scrollBy({ top: step }); }
    }
  });
  dialog.addEventListener("close", () => {
    // Focus goes back to the thumbnail of the capture that was showing, not the one that opened the carousel.
    const shown = thumbs && thumbs[position];
    if (shown && shown.isConnected) shown.focus();
    thumbs = null;
  });

  return {
    dialog,
    open(list, index, thumbList) {
      captures = list;
      thumbs = thumbList;
      show(index);
      dialog.showModal();
      body.focus({ preventScroll: true });
    },
  };
}
