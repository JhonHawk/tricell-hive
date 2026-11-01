import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { isSafeTimelineURL } from "./timeline-contract";

/** Read-only documents: no HTML, remote images, embedded views or execution. */
export function TimelineMarkdown({ content }: { content: string }) {
  return (
    <div className="flex min-w-0 flex-col gap-4 break-words text-sm leading-relaxed">
      <Markdown
        remarkPlugins={[remarkGfm]}
        skipHtml
        disallowedElements={["img"]}
        urlTransform={(url) => (isSafeTimelineURL(url) ? url : undefined)}
        components={{
          a: ({ href, children }) =>
            isSafeTimelineURL(href) ? (
              <a
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                className="underline underline-offset-4"
              >
                {children}
              </a>
            ) : (
              <span>{children}</span>
            ),
          img: () => null,
          h1: ({ children }) => (
            <h1 className="text-xl font-semibold">{children}</h1>
          ),
          h2: ({ children }) => (
            <h2 className="text-lg font-semibold">{children}</h2>
          ),
          h3: ({ children }) => (
            <h3 className="text-base font-semibold">{children}</h3>
          ),
          h4: ({ children }) => (
            <h4 className="text-sm font-semibold">{children}</h4>
          ),
          ul: ({ children }) => <ul className="list-disc pl-5">{children}</ul>,
          ol: ({ children }) => (
            <ol className="list-decimal pl-5">{children}</ol>
          ),
          pre: ({ children }) => (
            <pre className="overflow-x-auto whitespace-pre text-sm">
              {children}
            </pre>
          ),
          // GFM document tables are semantic prose, not an interactive data region.
          table: ({ children }) => (
            <div className="min-w-0 overflow-x-auto">
              <table className="w-full border-collapse text-left text-sm">
                {children}
              </table>
            </div>
          ),
          th: ({ children }) => (
            <th className="border-b border-border font-medium">{children}</th>
          ),
          td: ({ children }) => (
            <td className="border-b border-border">{children}</td>
          ),
          blockquote: ({ children }) => (
            <blockquote className="border-l border-border pl-4 text-muted-foreground">
              {children}
            </blockquote>
          ),
        }}
      >
        {content}
      </Markdown>
    </div>
  );
}
