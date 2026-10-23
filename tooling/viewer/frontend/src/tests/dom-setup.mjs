import { JSDOM } from "jsdom";

const dom = new JSDOM(
  '<!doctype html><html><body><div id="root"></div></body></html>',
  { url: "http://127.0.0.1:5191/" },
);
for (const key of [
  "window",
  "document",
  "navigator",
  "location",
  "history",
  "sessionStorage",
  "HTMLElement",
  "Element",
  "Node",
  "NodeFilter",
  "HTMLInputElement",
  "HTMLButtonElement",
  "HTMLDialogElement",
  "MutationObserver",
  "getComputedStyle",
  "Event",
  "MouseEvent",
  "KeyboardEvent",
])
  Object.defineProperty(globalThis, key, {
    value:
      key === "getComputedStyle"
        ? dom.window.getComputedStyle.bind(dom.window)
        : dom.window[key],
    configurable: true,
  });
globalThis.IS_REACT_ACT_ENVIRONMENT = true;
window.matchMedia = () => ({
  matches: false,
  addEventListener() {},
  removeEventListener() {},
});
globalThis.matchMedia = window.matchMedia;
globalThis.requestAnimationFrame = (callback) =>
  setTimeout(() => callback(0), 0);
globalThis.cancelAnimationFrame = clearTimeout;
window.requestAnimationFrame = globalThis.requestAnimationFrame;
window.cancelAnimationFrame = globalThis.cancelAnimationFrame;
window.ResizeObserver = class {
  observe() {}
  disconnect() {}
  unobserve() {}
};
globalThis.ResizeObserver = window.ResizeObserver;
if (!globalThis.CSS) globalThis.CSS = { escape: (value) => String(value) };
