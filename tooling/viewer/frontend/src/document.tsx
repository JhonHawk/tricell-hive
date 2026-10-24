import { useLayoutEffect, useRef } from "react";
import type { Delivery, DocumentData } from "./controller";
export type WalkthroughRenderer = (
  root: HTMLElement,
  manifest: DocumentData["manifest"],
  options: { fileBase: string; idPrefix: string; embedded: boolean },
) => (() => void) | undefined;
export type WalkthroughModule = {
  renderWalkthrough: WalkthroughRenderer;
  validDeliveries: (value: unknown) => Delivery[];
};
let modulePromise: Promise<WalkthroughModule> | undefined;
let rendererPromise: Promise<WalkthroughRenderer> | undefined;
let rendererAttempt = 0;
export function loadWalkthroughModule() {
  // Some browsers retain a failed module URL. Change only failed attempts;
  // a successful import and concurrent callers keep one module instance.
  const path =
    rendererAttempt === 0
      ? "/assets/walkthrough.mjs"
      : `/assets/walkthrough.mjs?retry=${rendererAttempt}`;
  modulePromise ??= import(/* @vite-ignore */ path).catch((error) => {
    modulePromise = undefined;
    rendererAttempt++;
    throw error;
  });
  return modulePromise;
}
export function loadWalkthroughRenderer() {
  rendererPromise ??= loadWalkthroughModule()
    .then((module) => module.renderWalkthrough)
    .catch((error) => {
      rendererPromise = undefined;
      throw error;
    });
  return rendererPromise;
}
export function clearWalkthrough(root: HTMLElement) {
  for (const video of root.querySelectorAll("video")) {
    video.pause();
    video.removeAttribute("src");
    video.load();
  }
  for (const dialog of root.querySelectorAll<HTMLDialogElement>("dialog[open]"))
    dialog.close();
  root.replaceChildren();
}
export function WalkthroughDocument({
  data,
  onError,
  onReady,
  loadRenderer = loadWalkthroughRenderer,
}: {
  data: DocumentData;
  onError: () => void;
  onReady?: (data: DocumentData) => void;
  loadRenderer?: () => Promise<WalkthroughRenderer>;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const savedFocus = useRef<{ key: string; index: number } | null>(null);
  useLayoutEffect(() => {
    const root = ref.current;
    if (!root) return;
    let cancelled = false;
    let dispose: (() => void) | undefined;
    void loadRenderer()
      .then((render) => {
        if (cancelled) return;
        try {
          clearWalkthrough(root);
          dispose = render(root, data.manifest, {
            fileBase: data.fileBase,
            idPrefix: `segment-${data.key}-`,
            embedded: true,
          });
          if (
            savedFocus.current?.key === data.key &&
            savedFocus.current.index >= 0
          )
            root
              .querySelectorAll<HTMLElement>(
                "button,a,input,summary,[tabindex]",
              )
              [savedFocus.current.index]?.focus({ preventScroll: true });
          onReady?.(data);
        } catch {
          clearWalkthrough(root);
          onError();
        }
      })
      .catch(() => {
        if (!cancelled) onError();
      });
    return () => {
      cancelled = true;
      const controls = [
        ...root.querySelectorAll<HTMLElement>(
          "button,a,input,summary,[tabindex]",
        ),
      ];
      const index = controls.indexOf(document.activeElement as HTMLElement);
      savedFocus.current = { key: data.key, index };
      const focused = index >= 0;
      dispose?.();
      clearWalkthrough(root);
      if (focused)
        document.getElementById("document")?.focus({ preventScroll: true });
    };
  }, [data, loadRenderer, onError, onReady]);
  return <div className="walkthrough walkthrough-embedded" ref={ref} />;
}
