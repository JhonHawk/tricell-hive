import { useState } from "react";
import { Button } from "@/components/ui/button";
export function CaptureImage({
  src,
  alt,
  onRetry,
  label,
  className,
  retryLabel,
}: {
  src: string;
  alt: string;
  onRetry: () => void;
  label: string;
  retryLabel: string;
  className?: string;
}) {
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  if (failed)
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-xs text-muted-foreground">{label}</p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          aria-label={retryLabel}
          onClick={() => {
            setFailed(false);
            setAttempt((n) => n + 1);
            onRetry();
          }}
        >
          ↻
        </Button>
      </div>
    );
  return (
    <img
      key={`${src}:${attempt}`}
      src={src}
      alt={alt}
      className={className}
      onError={() => setFailed(true)}
    />
  );
}

export function CaptureThumbnail({
  url,
  alt,
  label,
  onOpen,
  onRetry,
  retryLabel,
  failureLabel,
}: {
  url: string;
  alt: string;
  label: string;
  onOpen: (trigger: HTMLElement) => void;
  onRetry: () => void;
  retryLabel: string;
  failureLabel: string;
}) {
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  if (failed)
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-xs text-muted-foreground">{failureLabel}</p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => {
            setFailed(false);
            setAttempt((n) => n + 1);
            onRetry();
          }}
        >
          {retryLabel}
        </Button>
      </div>
    );
  return (
    <button
      type="button"
      className="group block w-full min-w-0 overflow-hidden border border-border text-left transition-colors hover:border-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      aria-label={label}
      onClick={(e) => onOpen(e.currentTarget)}
    >
      <img
        key={`${url}:${attempt}`}
        src={url}
        alt={alt}
        loading="lazy"
        onError={() => setFailed(true)}
        className="block aspect-[4/3] w-full object-cover object-top"
      />
    </button>
  );
}
