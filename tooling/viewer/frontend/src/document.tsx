import { useLayoutEffect, useRef } from "react";
import type { DocumentData } from "./controller";
export type WalkthroughRenderer = (
  root: HTMLElement,
  manifest: DocumentData["manifest"],
  options: { fileBase: string },
) => void;
let rendererPromise: Promise<WalkthroughRenderer> | undefined;
export function loadWalkthroughRenderer() {
  const path = "/assets/walkthrough.mjs";
  rendererPromise ??= import(/* @vite-ignore */ path).then(
    (module) => module.renderWalkthrough,
  );
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
  loadRenderer = loadWalkthroughRenderer,
}: {
  data: DocumentData;
  onError: () => void;
  loadRenderer?: () => Promise<WalkthroughRenderer>;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const saved = useRef<{ key: string; scroll: number; focus: number } | null>(
    null,
  );
  useLayoutEffect(() => {
    const root = ref.current;
    if (!root) return;
    let cancelled = false;
    const scrollParent = document.getElementById("document")?.parentElement;
    void loadRenderer()
      .then((render) => {
        if (cancelled) return;
        try {
          clearWalkthrough(root);
          render(root, data.manifest, { fileBase: data.fileBase });
          if (saved.current?.key === data.key) {
            if (scrollParent) scrollParent.scrollTop = saved.current.scroll;
            const controls = root.querySelectorAll<HTMLElement>(
              "button,a,input,summary,[tabindex]",
            );
            controls[saved.current.focus]?.focus({ preventScroll: true });
          } else if (scrollParent) scrollParent.scrollTop = 0;
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
      const focus = controls.indexOf(document.activeElement as HTMLElement);
      saved.current = {
        key: data.key,
        scroll: scrollParent?.scrollTop ?? 0,
        focus,
      };
      clearWalkthrough(root);
      if (focus >= 0)
        document.getElementById("document")?.focus({ preventScroll: true });
    };
  }, [data, loadRenderer, onError]);
  return <div className="walkthrough" ref={ref} />;
}
