import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogCancel,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import type { Delivery, SegmentDocument } from "./controller";
import { loadWalkthroughModule, type WalkthroughModule } from "./document";
import { messages as t } from "./messages";
import type { SessionGroup } from "./sessions";
export function DeliveryLink({ item }: { item: Delivery }) {
  return (
    <a
      className="text-sm underline underline-offset-4 [overflow-wrap:anywhere] hover:text-muted-foreground transition-colors"
      href={item.url}
      target="_blank"
      rel="noopener noreferrer"
    >
      {item.label}
    </a>
  );
}
export function SessionDeliveries({
  group,
  documents,
  onRetry,
  loadModule = loadWalkthroughModule,
}: {
  group: SessionGroup;
  documents: Record<string, SegmentDocument>;
  onRetry: (key: string) => void;
  loadModule?: () => Promise<WalkthroughModule>;
}) {
  const [module, setModule] = useState<WalkthroughModule>();
  const [moduleError, setModuleError] = useState(false);
  const [attempt, setAttempt] = useState(0);
  // biome-ignore lint/correctness/useExhaustiveDependencies: A user retry must restart the read even when the loader identity is unchanged.
  useEffect(() => {
    let cancelled = false;
    setModule(undefined);
    setModuleError(false);
    void loadModule()
      .then((loaded) => {
        if (!cancelled) setModule(loaded);
      })
      .catch(() => {
        if (!cancelled) setModuleError(true);
      });
    return () => {
      cancelled = true;
    };
  }, [loadModule, attempt]);
  const entries = group.segments.map((segment) => ({
    segment,
    entry: documents[segment.id],
  }));
  const deliveries = new Map<string, { item: Delivery; segments: string[] }>();
  for (const { segment, entry } of entries)
    for (const item of module?.validDeliveries(
      entry?.document?.manifest.delivery,
    ) ?? []) {
      const found = deliveries.get(item.url);
      if (found) found.segments.push(segment.title);
      else deliveries.set(item.url, { item, segments: [segment.title] });
    }
  const all = [...deliveries.values()];
  const pending =
    (!module && !moduleError) ||
    entries.some(({ entry }) => !entry || entry.state === "loading");
  const failed = entries.filter(
    ({ entry }) => entry?.state === "error" || entry?.state === "excluded",
  );
  return (
    <section aria-label={t.deliveries} className="flex min-w-0 flex-col gap-3">
      <h2 className="font-medium tracking-tight">{t.deliveries}</h2>
      <div className="flex flex-wrap items-center gap-3">
        {all.slice(0, 2).map(({ item }) => (
          <DeliveryLink key={item.url} item={item} />
        ))}
        {all.length > 2 && (
          <Dialog>
            <DialogTrigger
              render={<Button type="button" size="sm" variant="outline" />}
            >
              {t.viewAll} ({all.length})
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{t.deliveries}</DialogTitle>
                <DialogDescription>{t.deliveryDescription}</DialogDescription>
              </DialogHeader>
              <ul className="flex min-h-0 flex-col gap-4 overflow-auto">
                {all.map(({ item, segments }) => (
                  <li key={item.url} className="flex min-w-0 flex-col gap-1">
                    <DeliveryLink item={item} />
                    <p className="break-words text-xs text-muted-foreground">
                      {segments.join(" · ")}
                    </p>
                  </li>
                ))}
              </ul>
              <DialogCancel type="button">{t.close}</DialogCancel>
            </DialogContent>
          </Dialog>
        )}
        {pending && (
          <div role="status" aria-label={t.loadingDeliveries}>
            <Skeleton className="h-4 w-32" />
          </div>
        )}
        {!pending && !moduleError && !failed.length && !all.length && (
          <p className="text-sm text-muted-foreground">{t.noDelivery}</p>
        )}
      </div>
      {(moduleError || failed.length > 0) && (
        <div role="status" className="flex flex-wrap items-center gap-2">
          <p className="text-sm text-muted-foreground">{t.partialDeliveries}</p>
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={() => {
              if (moduleError) setAttempt((attempt) => attempt + 1);
              failed.forEach(({ segment }) => {
                onRetry(segment.id);
              });
            }}
          >
            {t.retry}
          </Button>
        </div>
      )}
    </section>
  );
}
