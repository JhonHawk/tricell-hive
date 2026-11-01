import { Badge } from "@/components/reui/badge";
import {
  isSafeTimelineURL,
  type TimelineDeliveryReference,
  type TimelineSession,
  type TimelineTicketRef,
} from "./timeline-contract";
import { type TimelineLanguage, timelineCopy } from "./timeline-copy";

/** Resolve only work-local opaque IDs; publication snapshots are never fallbacks. */
export function resolveTicket(
  session: TimelineSession,
  ref: TimelineTicketRef,
) {
  return session.tickets.find((ticket) => ticket.ticket_id === ref.ticket_id);
}
export function resolveDelivery(
  session: TimelineSession,
  ref: TimelineDeliveryReference,
) {
  return session.deliveries.find(
    (delivery) => delivery.delivery_id === ref.delivery_id,
  );
}
function ResourceLink({ url, label }: { url?: string; label: string }) {
  return isSafeTimelineURL(url) ? (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="break-words underline underline-offset-4 transition-colors hover:text-muted-foreground"
    >
      {label}
    </a>
  ) : (
    <span>{label}</span>
  );
}
export function TimelineTicketLinks({
  session,
  refs,
  lang,
}: {
  session: TimelineSession;
  refs: TimelineTicketRef[];
  lang: TimelineLanguage;
}) {
  const t = timelineCopy(lang);
  return (
    <div className="flex flex-wrap items-center gap-2">
      {refs.map((ref) => {
        const current = resolveTicket(session, ref);
        const changed =
          !current ||
          current.key !== ref.snapshot.key ||
          current.title !== ref.snapshot.title ||
          current.url !== ref.snapshot.url;
        return (
          <span
            key={ref.ticket_id}
            className="flex min-w-0 flex-wrap items-center gap-2 text-xs"
            data-ticket-id={ref.ticket_id}
            title={current?.title}
          >
            {changed && (
              <span className="text-muted-foreground">
                {t.published}:{" "}
                {current ? (
                  <ResourceLink
                    url={ref.snapshot.url}
                    label={ref.snapshot.key}
                  />
                ) : (
                  ref.snapshot.key
                )}
              </span>
            )}
            {changed && <span>{t.current}:</span>}
            {current ? (
              <ResourceLink url={current.url} label={current.key} />
            ) : (
              <span>{t.ticketUnavailable}</span>
            )}
          </span>
        );
      })}
    </div>
  );
}
export function TimelineDeliveryLinks({
  session,
  refs,
  lang,
}: {
  session: TimelineSession;
  refs: TimelineDeliveryReference[];
  lang: TimelineLanguage;
}) {
  const t = timelineCopy(lang);
  return (
    <div className="flex min-w-0 flex-col gap-2">
      {refs.map((ref) => {
        const current = resolveDelivery(session, ref);
        const changed =
          !current ||
          current.label !== ref.snapshot.label ||
          current.url !== ref.snapshot.url ||
          current.status !== ref.snapshot.status;
        return (
          <div
            key={ref.delivery_id}
            data-delivery-id={ref.delivery_id}
            className="flex min-w-0 flex-wrap items-center gap-2 text-xs"
          >
            {changed && (
              <span className="text-muted-foreground">
                {t.published}:{" "}
                {current ? (
                  <ResourceLink
                    url={ref.snapshot.url}
                    label={ref.snapshot.label}
                  />
                ) : (
                  ref.snapshot.label
                )}{" "}
                · {t[ref.snapshot.status]}
              </span>
            )}
            {changed && <span>{t.current}:</span>}
            {current ? (
              <>
                <ResourceLink url={current.url} label={current.label} />
                <Badge
                  variant={
                    current.status === "merged" ? "success-light" : "outline"
                  }
                >
                  {t[current.status]}
                </Badge>
              </>
            ) : (
              <span>{t.unavailableDetail}</span>
            )}
          </div>
        );
      })}
    </div>
  );
}
