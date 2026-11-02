import { useState } from "react";
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
  const lengthy = content.length > 160 || content.split("\n").length > lines;
  return (
    <div className="flex min-w-0 flex-col items-start gap-1">
      <div
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
      {lengthy && (
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
