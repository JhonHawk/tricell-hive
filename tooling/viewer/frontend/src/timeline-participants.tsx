import { Users } from "lucide-react";
import { useId, useRef, useState } from "react";
import { Alert, AlertAction, AlertDescription } from "@/components/reui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTitle,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Skeleton } from "@/components/ui/skeleton";
import { HttpError } from "./api";
import type { TimelineSessionResponse } from "./timeline-contract";
import { type TimelineLanguage, timelineCopy } from "./timeline-copy";
export function TimelineParticipants({
  data,
  lang,
  onRetry,
  onRecover,
}: {
  data: TimelineSessionResponse;
  lang: TimelineLanguage;
  onRetry: () => void | Promise<void>;
  onRecover?: (
    requestID: string,
    reason: string,
    expectedVersion: number,
  ) => Promise<void>;
}) {
  const reasonID = useId();
  const [popoverOpen, setPopoverOpen] = useState(false);
  const { session } = data,
    t = timelineCopy(lang);
  const [requestID, setRequestID] = useState<string | null>(null),
    [reason, setReason] = useState(""),
    [pending, setPending] = useState(false),
    [error, setError] = useState<"conflict" | "failure" | null>(null);
  const version = useRef(session.version),
    trigger = useRef<HTMLElement | null>(null);
  const latest = useRef(session);
  latest.current = session;
  const refresh = async () => {
    setPending(true);
    try {
      await onRetry();
      version.current = latest.current.version;
      setError(null);
    } catch {
      setError("failure");
    } finally {
      setPending(false);
    }
  };
  const request = session.pending_join_requests.find(
      (r) => r.request_id === requestID,
    ),
    lead = session.producers.find(
      (p) => p.producer_id === session.foreground_producer_id,
    );
  const recover = async () => {
    if (pending || !request || !reason.trim() || !onRecover || error !== null)
      return;
    setPending(true);
    setError(null);
    try {
      await onRecover(request.request_id, reason.trim(), version.current);
      setRequestID(null);
    } catch (e) {
      setError(
        e instanceof HttpError && e.status === 409 ? "conflict" : "failure",
      );
    } finally {
      setPending(false);
    }
  };
  return (
    <>
      <Popover open={popoverOpen} onOpenChange={setPopoverOpen}>
        <PopoverTrigger
          render={
            <Button
              ref={trigger as React.Ref<HTMLButtonElement>}
              aria-label={`${t.sessions}: ${session.producers.length}`}
              title={`${t.sessions}: ${session.producers.length}`}
              type="button"
              variant="outline"
              size="sm"
            />
          }
        >
          <Users aria-hidden />
          {session.producers.length}
        </PopoverTrigger>
        <PopoverContent align="end">
          <PopoverTitle>{t.sessions}</PopoverTitle>
          {pending && !requestID && (
            <div role="status" aria-label={t.loading}>
              <Skeleton className="h-8 w-full" />
            </div>
          )}
          {error && !requestID && (
            <Alert variant="destructive">
              <AlertDescription>{t.loadFailed}</AlertDescription>
              <AlertAction>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void refresh()}
                >
                  {t.retry}
                </Button>
              </AlertAction>
            </Alert>
          )}
          <ul className="flex flex-col gap-3">
            {session.producers.map((p) => (
              <li key={p.producer_id} className="flex min-w-0 flex-col gap-1">
                <span className="text-sm">
                  {p.source.host}
                  {p.producer_id === session.foreground_producer_id
                    ? ` · ${t.lead}`
                    : ""}
                </span>
                <code className="break-all text-xs">
                  {p.source.native_session_id}
                </code>
              </li>
            ))}
          </ul>
          {session.pending_join_requests.length ? (
            <ul className="flex flex-col gap-3">
              {session.pending_join_requests.map((r) => (
                <li
                  key={r.request_id}
                  className="flex flex-col items-start gap-2"
                >
                  <p className="text-sm">
                    {t.requester}: {r.source.host}
                  </p>
                  <code className="break-all text-xs">
                    {r.source.native_session_id}
                  </code>
                  <p className="text-sm">{r.reason}</p>
                  {onRecover &&
                    !["closed", "failed"].includes(session.state) && (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={async (e) => {
                          e.preventDefault();
                          setPopoverOpen(false);
                          version.current = session.version;
                          setRequestID(r.request_id);
                          setReason("");
                          setError(null);
                          await refresh();
                        }}
                      >
                        {t.reassign}
                      </Button>
                    )}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm">{t.noRequests}</p>
          )}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => void refresh()}
          >
            {t.reload}
          </Button>
        </PopoverContent>
      </Popover>
      <Dialog
        open={requestID !== null}
        onOpenChange={(open) => {
          if (!pending && !open) setRequestID(null);
        }}
      >
        <DialogContent finalFocus={trigger}>
          <DialogHeader>
            <DialogTitle>{t.reassignLead}</DialogTitle>
            <DialogDescription className="break-words [overflow-wrap:anywhere]">
              {t.from}: {lead?.source.host} · {lead?.source.native_session_id}
              <br />
              {t.to}: {request?.source.host} ·{" "}
              {request?.source.native_session_id}
            </DialogDescription>
          </DialogHeader>
          <label htmlFor={reasonID} className="flex flex-col gap-2 text-sm">
            {t.reason}
            <Input
              id={reasonID}
              value={reason}
              maxLength={500}
              disabled={pending}
              onChange={(e) => setReason(e.target.value)}
              aria-invalid={Boolean(error)}
            />
          </label>
          {(!request || error) && (
            <Alert variant="destructive">
              <AlertDescription>
                {error === "conflict" || !request
                  ? t.recoveryConflict
                  : t.recoveryFailed}
              </AlertDescription>
              <AlertAction>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void refresh()}
                >
                  {t.reload}
                </Button>
              </AlertAction>
            </Alert>
          )}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => setRequestID(null)}
            >
              {t.cancel}
            </Button>
            <Button
              type="button"
              disabled={pending || !request || !reason.trim() || error !== null}
              onClick={() => void recover()}
            >
              {pending ? t.saving : t.reassign}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
