// Hive walkthrough viewer. Everything that comes from the server is untrusted:
// it only reaches the DOM through textContent and setAttribute, and no style
// or on* attribute is ever set (the CSP has no 'unsafe-inline').
//
// Shell language: the browser language (navigator.language starting with "es"
// gives Spanish, anything else English). The walkthrough itself keeps the
// `lang` of its manifest.
import { renderWalkthrough } from "/assets/walkthrough.mjs";

const TOKEN_KEY = "hive-view-token";
const OPEN_KEY = "hive-view-open";
const POLL_MS = 10000;
const POLL_DOWN_MS = 5000;

const LABELS = {
  en: {
    appTitle: "Hive walkthroughs",
    skip: "Skip to the walkthrough",
    loading: "Loading…",
    loadingSession: "Loading walkthrough…",
    menu: "Walkthroughs",
    close: "Close",
    keep: "Keep",
    unkeep: "Stop keeping",
    delete: "Delete",
    kept: "Kept",
    keptTitle: "Kept: it will not expire",
    absent: "absent",
    absentTitle: "The repository folder was not found on disk. Its copies are kept.",
    noTicket: "No ticket",
    sessions: (n) => `${n} ${n === 1 ? "walkthrough" : "walkthroughs"}`,
    actionsFor: (name) => `Actions for ${name}`,
    treeLabel: "Walkthroughs by project and ticket",
    confirmTitle: "Delete this walkthrough?",
    confirmBody: "Only the copy stored by Hive is removed. The files in your repository are not touched.",
    cancel: "Cancel",
    confirmDelete: "Delete",
    ttl: (d) => `Copies expire after ${d} unless kept`,
    errorsOne: "1 ingest error",
    errorsMany: (n) => `${n} ingest errors`,
    emptyRepoTitle: "No repositories registered",
    emptyRepoBody: "Register a repository so Hive can find its walkthroughs. Run this inside the repository:",
    emptyRepoAfter: "Agents that save a walkthrough run it for you. New walkthroughs show up here within a few seconds.",
    emptyNoneTitle: "No walkthroughs yet",
    emptyNoneBody: "The registered repositories have no walkthrough.json to show. When an agent saves one under `_support/workspace/<session>/`, it appears here within a few seconds.",
    reposHeading: "Registered repositories",
    copy: "Copy",
    copied: "Copied",
    accessTitle: "This page needs access",
    noToken: "Run `hive view open` to open this page with access.",
    tokenRejected: "Access was rejected or the opening code expired or was already used. Run `hive view open` to reopen this page.",
    accessNote: "No walkthroughs are requested until the page has a valid token.",
    down: "Cannot reach the Hive viewer server. Start it with `hive view start`, then reopen this page with `hive view open` (a new start usually picks a new port). If the server answers again at this address, the page reconnects by itself.",
    backUp: "Reconnected to the server.",
    gone: "The walkthrough you had open is no longer available. Showing the most recent one.",
    dismiss: "Dismiss",
    kept1: (name) => `Kept “${name}”. It will not expire.`,
    unkept1: (name) => `“${name}” will expire with the rest.`,
    deleted1: (name) => `Deleted “${name}”.`,
    failed: (status) => `That did not work (HTTP ${status}). Try again.`,
    loadFailed: "Could not load this walkthrough.",
    retry: "Retry",
    crumbNone: "",
  },
  es: {
    appTitle: "Recorridos de Hive",
    skip: "Saltar al recorrido",
    loading: "Cargando…",
    loadingSession: "Cargando recorrido…",
    menu: "Recorridos",
    close: "Cerrar",
    keep: "Conservar",
    unkeep: "Dejar de conservar",
    delete: "Borrar",
    kept: "Conservado",
    keptTitle: "Conservado: no caducará",
    absent: "ausente",
    absentTitle: "No se encontró la carpeta del repositorio en el disco. Sus copias se conservan.",
    noTicket: "Sin ticket",
    sessions: (n) => `${n} ${n === 1 ? "recorrido" : "recorridos"}`,
    actionsFor: (name) => `Acciones de ${name}`,
    treeLabel: "Recorridos por proyecto y ticket",
    confirmTitle: "¿Borrar este recorrido?",
    confirmBody: "Solo se borra la copia que guarda Hive. Los archivos de tu repositorio no se tocan.",
    cancel: "Cancelar",
    confirmDelete: "Borrar",
    ttl: (d) => `Las copias caducan en ${d} salvo que se conserven`,
    errorsOne: "1 error de ingesta",
    errorsMany: (n) => `${n} errores de ingesta`,
    emptyRepoTitle: "No hay repositorios registrados",
    emptyRepoBody: "Registra un repositorio para que Hive encuentre sus recorridos. Ejecuta esto dentro del repositorio:",
    emptyRepoAfter: "Los agentes que guardan un recorrido lo ejecutan por ti. Los recorridos nuevos aparecen aquí en unos segundos.",
    emptyNoneTitle: "Aún no hay recorridos",
    emptyNoneBody: "Los repositorios registrados no tienen ningún walkthrough.json que mostrar. Cuando un agente guarde uno en `_support/workspace/<sesión>/`, aparecerá aquí en unos segundos.",
    reposHeading: "Repositorios registrados",
    copy: "Copiar",
    copied: "Copiado",
    accessTitle: "Esta página necesita acceso",
    noToken: "Ejecuta `hive view open` para abrir esta página con acceso.",
    tokenRejected: "El acceso fue rechazado o el código de apertura caducó o ya se usó. Ejecuta `hive view open` para volver a abrir esta página.",
    accessNote: "Los recorridos se cargan cuando la página tiene un token válido.",
    down: "No se puede conectar con el servidor de Hive. Inícialo con `hive view start` y vuelve a abrir esta página con `hive view open` (un inicio nuevo suele usar otro puerto). Si el servidor responde otra vez en esta dirección, la página se reconecta sola.",
    backUp: "Reconectado con el servidor.",
    gone: "El recorrido que tenías abierto ya no está disponible. Se muestra el más reciente.",
    dismiss: "Descartar",
    kept1: (name) => `Conservado: “${name}”. No caducará.`,
    unkept1: (name) => `“${name}” caducará con el resto.`,
    deleted1: (name) => `Borrado: “${name}”.`,
    failed: (status) => `No funcionó (HTTP ${status}). Inténtalo de nuevo.`,
    loadFailed: "No se pudo cargar este recorrido.",
    retry: "Reintentar",
    crumbNone: "",
  },
};

const lang = (navigator.language || "en").toLowerCase().startsWith("es") ? "es" : "en";
const t = LABELS[lang];

const state = {
  index: null,
  indexJson: "",
  token: "",
  down: false,
  openKey: null,
  openHash: "",
  expanded: new Map(),
  roving: null,
  loadSeq: 0,
  menu: null,
  panelOpen: false,
  noticeTimers: new Map(),
  prefer: null,
  hidden: new Set(),
  blocked: false,
  generation: 0,
};

const $ = (id) => document.getElementById(id);
const dom = {
  shell: $("shell"), sidebar: $("sidebar"), scroll: $("side-scroll"), foot: $("side-foot"),
  errorsToggle: $("errors-toggle"), errorsList: $("errors-list"), ttl: $("ttl"),
  main: $("viewer-main"), topbar: $("topbar"), crumb: $("crumb"), actions: $("actions"),
  pin: $("pin-button"), del: $("delete-button"), notices: $("notices"), stage: $("stage"),
  mount: $("wt-mount"), menuToggle: $("menu-toggle"), sidebarClose: $("sidebar-close"),
  scrim: $("scrim"), skip: $("skip-link"),
  confirm: $("confirm"), confirmTitle: $("confirm-title"), confirmName: $("confirm-name"),
  confirmBody: $("confirm-body"), confirmCancel: $("confirm-cancel"), confirmOk: $("confirm-ok"),
};

const narrow = matchMedia("(max-width: 719px)");

// ---------------------------------------------------------------- helpers

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

// Text with `code` spans: the only markup the shell's own messages use.
function rich(text) {
  const parts = text.split("`");
  return parts.map((part, i) => (i % 2 ? el("code", {}, part) : part));
}

function storage(op, key, value) {
  try {
    if (op === "get") return sessionStorage.getItem(key) || "";
    if (op === "set") sessionStorage.setItem(key, value);
    if (op === "del") sessionStorage.removeItem(key);
  } catch {
    /* storage can be unavailable; the page works without it */
  }
  return "";
}

// Remove an opening credential synchronously, before waiting for the network.
function takeCode() {
  const fragment = new URLSearchParams(location.hash.slice(1));
  const code = fragment.get("code");
  if (fragment.has("code") || fragment.has("token")) {
    history.replaceState(null, "", location.pathname + location.search);
  }
  return code;
}

class HttpError extends Error {
  constructor(status, url) {
    super(`${url}: ${status}`);
    this.status = status;
  }
}

// Every read carries the server token; the server answers 403 without a valid one.
async function getJSON(url) {
  const response = await fetch(url, { headers: { "X-Hive-Token": state.token }, cache: "no-store" });
  if (!response.ok) throw new HttpError(response.status, url);
  return response.json();
}

function allSessions(index) {
  return index.projects.flatMap((p) => p.tickets.flatMap((tk) => tk.sessions.map((s) => ({ project: p, ticket: tk, session: s }))));
}

function findSession(index, key) {
  return index ? allSessions(index).find((e) => e.session.key === key) : undefined;
}

// "Most recent" follows the manifest's created_at; ingested_at only breaks ties
// (it moves whenever a copy is re-ingested, so it says nothing about the walkthrough).
function recency(session) {
  return Date.parse(session.created_at) || Date.parse(session.ingested_at) || 0;
}

function latest(index) {
  let best;
  for (const entry of allSessions(index)) {
    const a = entry.session;
    const b = best && best.session;
    if (!best || recency(a) > recency(b) || (recency(a) === recency(b) && Date.parse(a.ingested_at) > Date.parse(b.ingested_at))) best = entry;
  }
  return best;
}

const dateFormat = new Intl.DateTimeFormat(lang, { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" });
function sessionDate(session) {
  const ms = Date.parse(session.created_at) || Date.parse(session.ingested_at);
  return Number.isNaN(ms) ? "" : dateFormat.format(new Date(ms));
}

function ticketName(name) {
  return name || t.noTicket;
}

function ttlLabel(seconds) {
  const day = 86400;
  const text = seconds % day === 0 ? `${seconds / day} d` : seconds >= 3600 ? `${Math.round(seconds / 3600)} h` : `${Math.max(1, Math.round(seconds / 60))} min`;
  return t.ttl(text);
}

function shortPath(path) {
  const parts = path.split("/").filter(Boolean);
  return parts.length > 3 ? `…/${parts.slice(-3).join("/")}` : path;
}

// ---------------------------------------------------------------- notices

function setNotice(key, text, { kind = "info", dismissable = false, timeout = 0, onDismiss } = {}) {
  clearNotice(key);
  const body = el("p", {}, ...rich(text));
  const node = el("div", { class: `notice ${kind}`, "data-notice": key }, body);
  if (dismissable) {
    const button = el("button", { type: "button", class: "btn btn-ghost notice-close", "aria-label": t.dismiss }, "×");
    button.addEventListener("click", () => { clearNotice(key); if (onDismiss) onDismiss(); });
    node.append(button);
  }
  dom.notices.append(node);
  if (timeout) state.noticeTimers.set(key, setTimeout(() => clearNotice(key), timeout));
}

function clearNotice(key) {
  clearTimeout(state.noticeTimers.get(key));
  state.noticeTimers.delete(key);
  for (const node of dom.notices.querySelectorAll("[data-notice]")) {
    if (node.getAttribute("data-notice") === key) node.remove();
  }
}

// ---------------------------------------------------------------- static labels

function applyStaticLabels() {
  document.documentElement.setAttribute("lang", lang);
  document.title = t.appTitle;
  dom.skip.textContent = t.skip;
  $("sidebar-title").textContent = t.appTitle;
  dom.sidebarClose.textContent = t.close;
  dom.menuToggle.textContent = t.menu;
  dom.del.textContent = t.delete;
  dom.confirmTitle.textContent = t.confirmTitle;
  dom.confirmBody.textContent = t.confirmBody;
  dom.confirmCancel.textContent = t.cancel;
  dom.confirmOk.textContent = t.confirmDelete;
  $("side-loading").textContent = t.loading;
}

// ---------------------------------------------------------------- sidebar

function groupDefaultExpanded(id, containsOpen, total) {
  if (state.expanded.has(id)) return state.expanded.get(id);
  return id.startsWith("p:") ? true : containsOpen || total <= 8;
}

function buildGroup(id, labelText, extra, contained, isOpen, total) {
  const expanded = groupDefaultExpanded(id, isOpen, total);
  const head = el("button", {
    type: "button", class: `group ${id.startsWith("p:") ? "project" : "ticket"}`,
    "data-stop": id, "aria-expanded": expanded ? "true" : "false", tabindex: "-1",
  }, el("span", { class: "chev", "aria-hidden": "true" }), el("span", { class: "g-name", title: labelText }, labelText), ...extra);
  const list = el("ul", { class: "list" }, ...contained);
  if (!expanded) list.setAttribute("hidden", "");
  head.addEventListener("click", () => {
    const next = head.getAttribute("aria-expanded") !== "true";
    state.expanded.set(id, next);
    head.setAttribute("aria-expanded", next ? "true" : "false");
    if (next) list.removeAttribute("hidden"); else list.setAttribute("hidden", "");
  });
  return { head, list };
}

function buildSessionRow(project, ticket, session) {
  const title = session.title || session.session_id || session.key.slice(0, 8);
  const open = session.key === state.openKey;
  const button = el("button", {
    type: "button", class: "session", "data-stop": `s:${session.key}`, "data-key": session.key, tabindex: "-1",
    "aria-current": open ? "true" : false,
  },
    el("span", { class: "s-title", title }, title),
    el("span", { class: "s-meta" }, el("span", { class: "s-when", title: session.created_at || session.ingested_at }, sessionDate(session)),
      " ", session.host ? el("span", { class: "s-host" }, session.host) : null, " ",
      session.pinned ? el("span", { class: "chip kept", title: t.keptTitle }, t.kept) : null)
  );
  button.addEventListener("click", () => { openSession(session.key, { fromSidebar: true }); });
  const menuButton = el("button", {
    type: "button", class: "menu-btn", "data-menu": session.key, tabindex: "-1",
    "aria-label": t.actionsFor(title), "aria-haspopup": "menu", "aria-expanded": "false",
  }, el("span", { class: "dots", "aria-hidden": "true" }, "⋯"));
  menuButton.addEventListener("click", (event) => {
    event.stopPropagation();
    toggleMenu(menuButton, { project, ticket, session });
  });
  return el("li", { class: "row", "data-row": session.key }, button, menuButton);
}

function renderSidebar() {
  const index = state.index;
  const hadFocus = dom.sidebar.contains(document.activeElement);
  const focused = hadFocus ? document.activeElement : null;
  const focusStop = focused && focused.closest("[data-stop]") ? focused.closest("[data-stop]").getAttribute("data-stop") : null;
  const focusMenu = focused && focused.hasAttribute("data-menu") ? focused.getAttribute("data-menu") : null;
  const scrollTop = dom.scroll.scrollTop;
  closeMenu(false);

  const nodes = [];
  const total = allSessions(index).length;
  for (const project of index.projects) {
    const pid = `p:${project.key}`;
    const count = project.tickets.reduce((n, tk) => n + tk.sessions.length, 0);
    const ticketItems = project.tickets.map((ticket) => {
      const tid = `t:${project.key}|${ticket.ticket}`;
      const rows = ticket.sessions.map((s) => buildSessionRow(project, ticket, s));
      const hasOpen = ticket.sessions.some((s) => s.key === state.openKey);
      const g = buildGroup(tid, ticketName(ticket.ticket), [el("span", { class: "g-count" }, String(ticket.sessions.length))], rows, hasOpen, total);
      return el("li", { class: "ticket-item" }, g.head, g.list);
    });
    const extra = [];
    if (project.absent) extra.push(el("span", { class: "chip absent", title: t.absentTitle }, t.absent));
    extra.push(el("span", { class: "g-count", title: t.sessions(count) }, String(count)));
    const g = buildGroup(pid, project.name || project.key, extra, ticketItems, true, total);
    nodes.push(el("li", { class: "project-item" }, g.head, g.list));
  }
  if (!nodes.length) {
    dom.scroll.replaceChildren(el("p", { class: "side-note" }, index.repositories.length ? t.emptyNoneTitle : t.emptyRepoTitle));
  } else {
    dom.scroll.replaceChildren(el("nav", { "aria-label": t.treeLabel }, el("ul", { class: "list tree" }, ...nodes)));
  }
  dom.scroll.scrollTop = scrollTop;

  // Roving tabindex: one stop for Tab, the arrows move among the rest.
  const stops = visibleStops();
  const wanted = state.roving && stops.some((s) => s.getAttribute("data-stop") === state.roving) ? state.roving
    : state.openKey && stops.some((s) => s.getAttribute("data-stop") === `s:${state.openKey}`) ? `s:${state.openKey}`
    : stops[0] && stops[0].getAttribute("data-stop");
  setRoving(wanted);

  if (hadFocus) {
    let target = null;
    if (focusMenu) target = dom.scroll.querySelector(`[data-menu="${CSS.escape(focusMenu)}"]`);
    if (!target && focusStop) target = dom.scroll.querySelector(`[data-stop="${CSS.escape(focusStop)}"]`);
    if (!target) target = dom.scroll.querySelector(`[data-stop="${CSS.escape(wanted || "")}"]`);
    if (target) target.focus({ preventScroll: true });
  }

  // Footer: ingest errors and TTL.
  dom.foot.hidden = false;
  dom.ttl.textContent = ttlLabel(index.ttl_seconds);
  const errors = index.errors || [];
  if (errors.length) {
    dom.errorsToggle.hidden = false;
    dom.errorsToggle.textContent = errors.length === 1 ? t.errorsOne : t.errorsMany(errors.length);
    dom.errorsList.replaceChildren(...errors.map((e) => el("div", { class: "err" },
      el("code", { class: "err-path", title: e.path }, shortPath(e.path)), el("span", { class: "err-reason" }, e.reason))));
  } else {
    dom.errorsToggle.hidden = true;
    dom.errorsToggle.setAttribute("aria-expanded", "false");
    dom.errorsList.hidden = true;
    dom.errorsList.replaceChildren();
  }
}

function visibleStops() {
  return [...dom.scroll.querySelectorAll("button[data-stop]")].filter((b) => !b.closest("[hidden]"));
}

function setRoving(id) {
  state.roving = id || null;
  for (const button of dom.scroll.querySelectorAll("button[data-stop]")) {
    button.setAttribute("tabindex", button.getAttribute("data-stop") === id ? "0" : "-1");
  }
  for (const menuButton of dom.scroll.querySelectorAll("button[data-menu]")) {
    menuButton.setAttribute("tabindex", id === `s:${menuButton.getAttribute("data-menu")}` ? "0" : "-1");
  }
}

dom.scroll.addEventListener("focusin", (event) => {
  const stop = event.target.closest && event.target.closest("[data-stop]");
  if (stop) setRoving(stop.getAttribute("data-stop"));
  else if (event.target.hasAttribute && event.target.hasAttribute("data-menu")) setRoving(`s:${event.target.getAttribute("data-menu")}`);
});

dom.scroll.addEventListener("keydown", (event) => {
  const target = event.target;
  if (!target.hasAttribute || !target.hasAttribute("data-stop")) return;
  const stops = visibleStops();
  const at = stops.indexOf(target);
  let next = null;
  if (event.key === "ArrowDown") next = stops[Math.min(stops.length - 1, at + 1)];
  else if (event.key === "ArrowUp") next = stops[Math.max(0, at - 1)];
  else if (event.key === "Home") next = stops[0];
  else if (event.key === "End") next = stops[stops.length - 1];
  else if (event.key === "ArrowRight" && target.hasAttribute("aria-expanded") && target.getAttribute("aria-expanded") === "false") { event.preventDefault(); target.click(); return; }
  else if (event.key === "ArrowLeft" && target.hasAttribute("aria-expanded") && target.getAttribute("aria-expanded") === "true") { event.preventDefault(); target.click(); return; }
  if (next) { event.preventDefault(); next.focus(); }
});

dom.errorsToggle.addEventListener("click", () => {
  const open = dom.errorsToggle.getAttribute("aria-expanded") !== "true";
  dom.errorsToggle.setAttribute("aria-expanded", open ? "true" : "false");
  dom.errorsList.hidden = !open;
});

// ---------------------------------------------------------------- row menu

function closeMenu(restoreFocus) {
  if (!state.menu) return;
  const { node, button } = state.menu;
  node.remove();
  button.setAttribute("aria-expanded", "false");
  state.menu = null;
  if (restoreFocus && button.isConnected) button.focus();
}

function toggleMenu(button, entry) {
  if (state.menu && state.menu.button === button) { closeMenu(true); return; }
  closeMenu(false);
  const { session } = entry;
  const keepLabel = session.pinned ? t.unkeep : t.keep;
  const keepItem = el("button", { type: "button", role: "menuitem", class: "menu-item", tabindex: "-1" }, keepLabel);
  const delItem = el("button", { type: "button", role: "menuitem", class: "menu-item danger", tabindex: "-1" }, t.delete);
  const node = el("div", { class: "menu", role: "menu", "aria-label": t.actionsFor(session.title || session.session_id) }, keepItem, delItem);
  keepItem.addEventListener("click", () => { closeMenu(false); setPinned(session, !session.pinned, button); });
  delItem.addEventListener("click", () => { closeMenu(false); requestDelete(entry, button); });
  node.addEventListener("keydown", (event) => {
    const items = [keepItem, delItem];
    const at = items.indexOf(document.activeElement);
    if (event.key === "ArrowDown") { event.preventDefault(); items[(at + 1) % items.length].focus(); }
    else if (event.key === "ArrowUp") { event.preventDefault(); items[(at - 1 + items.length) % items.length].focus(); }
    else if (event.key === "Home") { event.preventDefault(); items[0].focus(); }
    else if (event.key === "End") { event.preventDefault(); items[items.length - 1].focus(); }
    else if (event.key === "Escape") { event.preventDefault(); event.stopPropagation(); closeMenu(true); }
    else if (event.key === "Tab") { closeMenu(false); }
  });
  button.closest(".row").append(node);
  button.setAttribute("aria-expanded", "true");
  state.menu = { node, button };
  keepItem.focus();
  node.scrollIntoView({ block: "nearest" });
}

document.addEventListener("click", (event) => {
  if (state.menu && !state.menu.node.contains(event.target) && event.target !== state.menu.button && !state.menu.button.contains(event.target)) closeMenu(false);
});

// ---------------------------------------------------------------- main view

function renderChrome() {
  const entry = findSession(state.index, state.openKey);
  dom.actions.hidden = !entry;
  if (!entry) {
    dom.crumb.replaceChildren();
    return;
  }
  const project = entry.project.name || entry.project.key;
  const ticket = ticketName(entry.ticket.ticket);
  dom.crumb.replaceChildren(
    el("span", { class: "crumb-part", title: project }, project), el("span", { class: "crumb-sep", "aria-hidden": "true" }, "›"),
    el("span", { class: "crumb-part", title: ticket }, ticket)
  );
  dom.pin.textContent = entry.session.pinned ? t.unkeep : t.keep;
  dom.pin.classList.toggle("is-on", entry.session.pinned);
}

function showStage(node) {
  dom.stage.replaceChildren(...(node ? [node] : []));
  dom.stage.hidden = !node;
}

function commandBlock(command) {
  const copy = el("button", { type: "button", class: "btn copy" }, t.copy);
  copy.addEventListener("click", async () => {
    try {
      await navigator.clipboard.writeText(command);
      copy.textContent = t.copied;
      setTimeout(() => { copy.textContent = t.copy; }, 1600);
    } catch {
      /* clipboard unavailable: the command stays selectable */
    }
  });
  return el("div", { class: "cmd" }, el("code", {}, command), copy);
}

function renderEmpty() {
  dom.mount.hidden = true;
  dom.mount.replaceChildren();
  const index = state.index;
  document.title = t.appTitle;
  if (!index.repositories.length) {
    showStage(el("section", { class: "empty", "aria-labelledby": "empty-title" },
      el("h1", { id: "empty-title" }, t.emptyRepoTitle),
      el("p", {}, t.emptyRepoBody), commandBlock("hive view register"), el("p", { class: "empty-note" }, t.emptyRepoAfter)));
  } else {
    const repos = el("ul", { class: "repos" }, ...index.repositories.map((r) =>
      el("li", {}, el("code", { title: r.root }, r.root), r.absent ? el("span", { class: "chip absent", title: t.absentTitle }, t.absent) : null)));
    showStage(el("section", { class: "empty", "aria-labelledby": "empty-title" },
      el("h1", { id: "empty-title" }, t.emptyNoneTitle),
      el("p", {}, ...rich(t.emptyNoneBody)), el("h2", {}, t.reposHeading), repos));
  }
}

async function openSession(key, { fromSidebar = false, remount = false } = {}) {
  const entry = findSession(state.index, key);
  if (!entry) return;
  const sameSession = key === state.openKey;
  if (sameSession && !remount && !dom.mount.hidden && fromSidebar) {
    if (narrow.matches) closePanel(false, true);
    return;
  }
  const generation = state.generation;
  const seq = ++state.loadSeq;
  const scrollY = remount ? window.scrollY : 0;
  state.openKey = key;
  state.openHash = entry.session.content_hash;
  storage("set", OPEN_KEY, key);
  // Make sure the open session is visible in the tree.
  state.expanded.set(`t:${entry.project.key}|${entry.ticket.ticket}`, state.expanded.get(`t:${entry.project.key}|${entry.ticket.ticket}`) ?? true);
  renderSidebar();
  renderChrome();
  document.title = `${entry.session.title || t.appTitle} · ${t.appTitle}`;
  if (!remount) {
    showStage(el("p", { class: "stage-note" }, t.loadingSession));
    dom.mount.hidden = true;
  }
  if (fromSidebar && narrow.matches) closePanel(false, true);
  try {
    const manifest = await getJSON(`/api/sessions/${encodeURIComponent(key)}`);
    if (generation !== state.generation || seq !== state.loadSeq) return;
    dom.mount.hidden = false;
    renderWalkthrough(dom.mount, manifest, { fileBase: entry.session.fileBase });
    showStage(null);
    window.scrollTo(0, scrollY);
  } catch (error) {
    if (generation !== state.generation || seq !== state.loadSeq) return;
    dom.mount.hidden = true;
    if (error instanceof HttpError && error.status === 403) {
      rejectToken();
      return;
    }
    if (error instanceof HttpError && error.status === 404) {
      refresh();
      return;
    }
    const retry = el("button", { type: "button", class: "btn" }, t.retry);
    retry.addEventListener("click", () => openSession(key, { remount: false }));
    showStage(el("div", { class: "stage-error", role: "alert" }, el("p", {}, t.loadFailed), retry));
  }
}

// ---------------------------------------------------------------- index refresh

function withoutKeys(index, keys) {
  const projects = index.projects
    .map((p) => ({ ...p, tickets: p.tickets.map((tk) => ({ ...tk, sessions: tk.sessions.filter((s) => !keys.has(s.key)) })).filter((tk) => tk.sessions.length) }))
    .filter((p) => p.tickets.length);
  return { ...index, projects };
}

function applyIndex(received) {
  // Copies deleted here stay hidden until the server stops listing them, so a poll
  // that was already in flight cannot bring a deleted row back.
  const listed = new Set(allSessions(received).map((e) => e.session.key));
  for (const key of state.hidden) if (!listed.has(key)) state.hidden.delete(key);
  const index = state.hidden.size ? withoutKeys(received, state.hidden) : received;
  const json = JSON.stringify(index);
  if (json === state.indexJson) return;
  state.indexJson = json;
  state.index = index;
  const sessions = allSessions(index);

  if (!sessions.length) {
    const hadOpen = state.openKey;
    state.openKey = null;
    state.loadSeq++;
    renderSidebar();
    renderChrome();
    renderEmpty();
    if (hadOpen) setNotice("gone", t.gone, { kind: "warn", dismissable: true });
    return;
  }

  const current = findSession(index, state.openKey);
  if (current) {
    renderSidebar();
    renderChrome();
    if (current.session.content_hash !== state.openHash) openSession(state.openKey, { remount: true });
    return;
  }
  const wasOpen = Boolean(state.openKey);
  const preferred = state.prefer ? findSession(index, state.prefer) : undefined;
  state.prefer = null;
  const remembered = !wasOpen ? findSession(index, storage("get", OPEN_KEY)) : undefined;
  const target = preferred || remembered || latest(index);
  state.openKey = null;
  if (wasOpen) setNotice("gone", t.gone, { kind: "warn", dismissable: true });
  openSession(target.session.key, { remount: false });
}

let chain = Promise.resolve();
let queued = 0;
function refresh() {
  queued++;
  const generation = state.generation;
  const run = () => doRefresh(generation);
  chain = chain.then(run, run).finally(() => {
    if (generation === state.generation) queued--;
  });
  return chain;
}

async function doRefresh(generation) {
  if (generation !== state.generation || state.blocked) return;
  let index;
  try {
    index = await getJSON("/api/index");
    if (generation !== state.generation) return;
  } catch (error) {
    if (generation !== state.generation) return;
    if (error instanceof HttpError && error.status === 403) { rejectToken(); return; }
    if (error instanceof HttpError) { console.error(error); return; }
    // A network error: the server is not answering.
    if (!state.down) {
      state.down = true;
      setNotice("down", t.down, { kind: "error", dismissable: true });
    }
    if (!state.index) {
      $("side-loading")?.remove();
      showStage(el("p", { class: "stage-note" }, ...rich(t.down)));
    }
    return;
  }
  try {
    if (state.down) {
      state.down = false;
      clearNotice("down");
      setNotice("up", t.backUp, { kind: "ok", timeout: 4000 });
    }
    applyIndex(index);
  } catch (error) {
    console.error(error);
  }
}

let pollTimer = 0;
function schedulePoll() {
  clearTimeout(pollTimer);
  pollTimer = setTimeout(async () => {
    if (!document.hidden && !state.blocked && queued === 0) await refresh();
    if (!state.blocked) schedulePoll();
  }, state.down ? POLL_DOWN_MS : POLL_MS);
}

document.addEventListener("visibilitychange", () => {
  if (!document.hidden && !state.blocked && queued === 0) refresh();
});

// ---------------------------------------------------------------- token and writes

// Without a valid token the page asks for nothing: it shows one card with the way to reopen it.
function showAccessCard(rejected) {
  state.blocked = true;
  clearTimeout(pollTimer);
  state.loadSeq++;
  state.openKey = null;
  state.index = null;
  state.indexJson = "";
  state.down = false;
  closeMenu(false);
  for (const key of ["down", "up", "gone", "write"]) clearNotice(key);
  dom.mount.hidden = true;
  dom.mount.replaceChildren();
  dom.actions.hidden = true;
  dom.crumb.replaceChildren();
  dom.foot.hidden = true;
  dom.scroll.replaceChildren(el("p", { class: "side-note" }, t.accessNote));
  document.title = t.appTitle;
  const card = el("section", { class: "empty", "aria-labelledby": "access-title", tabindex: "-1", role: "alert" },
    el("h1", { id: "access-title" }, t.accessTitle),
    el("p", {}, ...rich(rejected ? t.tokenRejected : t.noToken)),
    commandBlock("hive view open"));
  showStage(card);
  card.focus({ preventScroll: true });
}

function rejectToken() {
  state.generation++;
  chain = Promise.resolve();
  queued = 0;
  state.token = "";
  storage("del", TOKEN_KEY);
  showAccessCard(true);
}

async function write(method, url) {
  const generation = state.generation;
  const response = await fetch(url, { method, headers: { "X-Hive-Token": state.token } });
  if (generation !== state.generation) return null;
  if (response.status === 403) {
    rejectToken();
  }
  return response;
}

async function setPinned(session, pinned, returnFocus) {
  const generation = state.generation;
  const name = session.title || session.session_id;
  const menuKey = returnFocus && returnFocus.getAttribute("data-menu");
  const response = await write(pinned ? "POST" : "DELETE", `/api/sessions/${session.key}/pin`).catch(() => null);
  if (generation !== state.generation) return;
  if (!response || !(response.ok || response.status === 404)) {
    if (response && response.status !== 403) setNotice("write", t.failed(response.status), { kind: "error", timeout: 6000 });
    return;
  }
  await refresh();
  if (generation !== state.generation) return;
  setNotice("write", pinned ? t.kept1(name) : t.unkept1(name), { kind: "ok", timeout: 4000 });
  const again = returnFocus && returnFocus.isConnected ? returnFocus : menuKey && dom.scroll.querySelector(`[data-menu="${CSS.escape(menuKey)}"]`);
  if (again) again.focus();
}

dom.pin.addEventListener("click", () => {
  const entry = findSession(state.index, state.openKey);
  if (entry) setPinned(entry.session, !entry.session.pinned, dom.pin);
});
dom.del.addEventListener("click", () => {
  const entry = findSession(state.index, state.openKey);
  if (entry) requestDelete(entry, dom.del);
});

// ---------------------------------------------------------------- delete

function refocus(node) {
  if (!node) return;
  if (node.isConnected) { node.focus(); return; }
  const key = node.getAttribute("data-menu");
  const again = key && dom.scroll.querySelector(`[data-menu="${CSS.escape(key)}"]`);
  if (again) again.focus();
}

function askConfirm(name, returnFocus) {
  dom.confirmName.textContent = name;
  return new Promise((resolve) => {
    const onClose = () => {
      dom.confirm.removeEventListener("close", onClose);
      const ok = dom.confirm.returnValue === "ok";
      dom.confirm.returnValue = "";
      if (!ok) refocus(returnFocus);
      resolve(ok);
    };
    dom.confirm.addEventListener("close", onClose);
    dom.confirm.returnValue = "";
    dom.confirm.showModal();
    dom.confirmCancel.focus();
  });
}
dom.confirmCancel.addEventListener("click", () => dom.confirm.close("cancel"));
dom.confirmOk.addEventListener("click", () => dom.confirm.close("ok"));

async function requestDelete(entry, returnFocus) {
  const generation = state.generation;
  const { session, ticket, project } = entry;
  const name = session.title || session.session_id;
  if (!(await askConfirm(name, returnFocus)) || generation !== state.generation) return;

  // Decide where focus goes before the row disappears.
  const siblings = ticket.sessions.map((s) => s.key);
  const at = siblings.indexOf(session.key);
  const before = visibleStops().map((b) => b.getAttribute("data-stop"));
  const wasOpen = session.key === state.openKey;

  const response = await write("DELETE", `/api/sessions/${session.key}`).catch(() => null);
  if (generation !== state.generation) return;
  if (!response || !(response.ok || response.status === 404)) {
    if (response && response.status !== 403) setNotice("write", t.failed(response.status), { kind: "error", timeout: 6000 });
    refocus(returnFocus);
    return;
  }

  // The server confirmed (204): drop the row now, announce it, and reconcile with a refresh afterwards.
  state.hidden.add(session.key);
  if (wasOpen) {
    // The replacement opens in applyIndex: the next walkthrough of the ticket, else the previous one, else the latest.
    state.prefer = siblings[at + 1] || siblings[at - 1] || null;
    state.openKey = null;
  }
  applyIndex(state.index);
  const exists = (id) => Boolean(dom.scroll.querySelector(`[data-stop="${CSS.escape(id)}"]`));
  const neighbours = [siblings[at + 1], siblings[at - 1]].filter(Boolean).map((k) => `s:${k}`);
  const ticketId = `t:${project.key}|${ticket.ticket}`;
  let target = neighbours.find(exists) || (exists(ticketId) ? ticketId : null);
  if (!target) {
    const gone = `s:${session.key}`;
    for (let i = before.indexOf(gone) - 1; i >= 0 && !target; i--) if (exists(before[i])) target = before[i];
  }
  if (!target) target = (visibleStops()[0] || {}).getAttribute?.("data-stop");
  clearNotice("gone");
  setNotice("write", t.deleted1(name), { kind: "ok", timeout: 5000 });
  const button = target && dom.scroll.querySelector(`[data-stop="${CSS.escape(target)}"]`);
  if (button && !(narrow.matches && !state.panelOpen)) {
    setRoving(target);
    button.focus();
  } else {
    dom.main.focus({ preventScroll: true });
  }
  refresh();
}

// ---------------------------------------------------------------- narrow panel

function openPanel() {
  state.panelOpen = true;
  dom.shell.classList.add("panel-open");
  dom.menuToggle.setAttribute("aria-expanded", "true");
  dom.main.inert = true;
  const target = dom.scroll.querySelector(`[data-stop="${CSS.escape(state.roving || "")}"]`) || dom.sidebarClose;
  target.focus({ preventScroll: true });
}

function closePanel(restoreToggle, toMain) {
  if (!state.panelOpen) return;
  state.panelOpen = false;
  dom.shell.classList.remove("panel-open");
  dom.menuToggle.setAttribute("aria-expanded", "false");
  dom.main.inert = false;
  if (toMain) dom.main.focus({ preventScroll: true });
  else if (restoreToggle) dom.menuToggle.focus();
}

dom.menuToggle.addEventListener("click", () => (state.panelOpen ? closePanel(true) : openPanel()));
dom.sidebarClose.addEventListener("click", () => closePanel(true));
dom.scrim.addEventListener("click", () => closePanel(true));
narrow.addEventListener("change", () => { if (!narrow.matches) closePanel(false); });
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && state.panelOpen && !state.menu && !dom.confirm.open) {
    event.preventDefault();
    closePanel(true);
  }
});
dom.skip.addEventListener("click", (event) => {
  event.preventDefault();
  dom.main.focus();
});

// A new opening supersedes every older exchange, read and write response.
async function establishAccess() {
  const code = takeCode();
  const generation = ++state.generation;
  clearTimeout(pollTimer);
  state.blocked = true;
  state.loadSeq++;
  chain = Promise.resolve();
  queued = 0;
  if (code !== null) {
    state.token = "";
    storage("del", TOKEN_KEY);
    // Clear the previous view while the explicit opening is being checked.
    showAccessCard(false);
    showStage(el("p", { class: "stage-note", role: "status" }, t.loading));
    try {
      const response = await fetch("/api/access-codes/exchange", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }), cache: "no-store",
      });
      if (generation !== state.generation) return;
      if (!response.ok) throw new Error("access rejected");
      const result = await response.json();
      if (generation !== state.generation) return;
      if (typeof result.token !== "string" || !result.token) throw new Error("invalid access response");
      state.token = result.token;
      storage("set", TOKEN_KEY, state.token);
    } catch {
      if (generation === state.generation) rejectToken();
      return;
    }
  } else {
    state.token = storage("get", TOKEN_KEY);
  }
  if (!state.token) { showAccessCard(false); return; }
  state.blocked = false;
  showStage(null);
  dom.scroll.replaceChildren(el("p", { class: "side-note", id: "side-loading" }, t.loading));
  schedulePoll();
  await refresh();
}

window.addEventListener("hashchange", () => {
  if (new URLSearchParams(location.hash.slice(1)).has("code")) establishAccess();
});

// ---------------------------------------------------------------- start

applyStaticLabels();
establishAccess();
