import { useLayoutEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { type TimelineLanguage, timelineCopy } from "./timeline-copy";
import { TimelineMarkdown } from "./timeline-markdown";
export function TimelineSummary({
  content,
  lines,
  lang,
}: {
  content: string;
  lines: 2 | 3;
  lang: TimelineLanguage;
}) {
  const [expanded, setExpanded] = useState(false);
  const t = timelineCopy(lang);
  // Offer "show more" only when the clamp actually hides text, measured on the
  // rendered box and rechecked when its width changes.
  const box = useRef<HTMLDivElement>(null);
  const [clipped, setClipped] = useState(false);
  // biome-ignore lint/correctness/useExhaustiveDependencies: new content can change the clipped height without resizing the box.
  useLayoutEffect(() => {
    const element = box.current;
    if (!element || expanded) return;
    const measure = () =>
      setClipped(element.scrollHeight > element.clientHeight + 1);
    measure();
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    return () => observer.disconnect();
  }, [content, expanded]);
  return (
    <div className="flex min-w-0 flex-col items-start gap-1">
      <div
        ref={box}
        className={
          !expanded
            ? // The markdown root is a flex column, which line-clamp cannot
              // count through; render it as a block while collapsed.
              lines === 2
              ? "line-clamp-2 [&>div]:block"
              : "line-clamp-3 [&>div]:block"
            : undefined
        }
      >
        <TimelineMarkdown content={content} />
      </div>
      {(expanded || clipped) && (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          aria-expanded={expanded}
          onClick={() => setExpanded(!expanded)}
        >
          {expanded ? t.showLess : t.showMore}
        </Button>
      )}
    </div>
  );
}
