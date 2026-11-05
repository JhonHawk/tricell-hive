import {
  Archive,
  ArchiveRestore,
  Check,
  Circle,
  GitPullRequest,
  ListChecks,
  MoreHorizontal,
  Pencil,
  Pin,
  Tickets,
  Trash2,
  X,
} from "lucide-react";
import { type ReactNode, useEffect, useId, useRef, useState } from "react";
import { Alert, AlertAction, AlertDescription } from "@/components/reui/alert";
import { Badge } from "@/components/reui/badge";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { ButtonGroup, ButtonGroupText } from "@/components/ui/button-group";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTitle,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  isSafeTimelineURL,
  type TimelineSessionResponse,
} from "./timeline-contract";
import {
  type TimelineLanguage,
  timelineAge,
  timelineCopy,
} from "./timeline-copy";
import { TimelineParticipants } from "./timeline-participants";
import { timelineStateVariant } from "./timeline-state";
import { TimelineSummary } from "./timeline-summary";

export type TimelineHeaderPatch = {
  title?: string;
  archived?: boolean;
  pinned?: boolean;
};
export type TimelineHeaderProps = {
  data: TimelineSessionResponse;
  disabled?: boolean;
  onUpdate: (
    patch: TimelineHeaderPatch,
    expectedVersion: number,
  ) => Promise<void>;
  onDelete: (expectedVersion: number) => Promise<void>;
  onRetry: () => void | Promise<void>;
  onRecover?: (
    requestID: string,
    reason: string,
    expectedVersion: number,
  ) => Promise<void>;
  lang?: TimelineLanguage;
};

function IconAction({
  label,
  children,
  ...props
}: { label: string; children: ReactNode } & Omit<
  React.ComponentProps<typeof Button>,
  "children"
>) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            type="button"
            variant="outline"
            size="icon-sm"
            aria-label={label}
            {...props}
          />
        }
      >
        {children}
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

export function TimelineHeader({
  data,
  disabled = false,
  onUpdate,
  onDelete,
  onRetry,
  onRecover,
  lang = "en",
}: TimelineHeaderProps) {
  const { session } = data;
  const titleHintID = useId();
  const t = timelineCopy(lang);
  const hostActivityAt = session.host_activity_at;
  const hostActive = hostActivityAt !== undefined;
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!hostActive) return;
    setNow(Date.now());
    const timer = setInterval(() => setNow(Date.now()), 60000);
    return () => clearInterval(timer);
  }, [hostActive]);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(session.title);
  const [saving, setPending] = useState(false);
  const pending = saving || disabled;
  const [error, setError] = useState<"required" | "save" | null>(null);
  const [deleting, setDeleting] = useState(false);
  const editVersion = useRef(session.version);
  const editSession = useRef(session.viewer_session_id);
  const deleteVersion = useRef(session.version);
  // Leaving title edit returns focus to the edit button, so keyboard users
  // keep their place instead of landing on the page body.
  const titleRow = useRef<HTMLDivElement>(null);
  const returnFocus = useRef(false);
  const closeEdit = () => {
    returnFocus.current = true;
    setEditing(false);
  };
  useEffect(() => {
    if (!editing && returnFocus.current) {
      returnFocus.current = false;
      // Only when focus was lost with the input, never taken from where the
      // user moved it while a save was pending.
      const active = document.activeElement;
      const header = titleRow.current?.closest('[data-slot="timeline-header"]');
      if (active === document.body || header?.contains(active)) {
        titleRow.current
          ?.querySelector<HTMLButtonElement>(`button[aria-label="${t.edit}"]`)
          ?.focus();
      }
    }
  }, [editing, t.edit]);
  const titleWraps = /\s/u.test(session.title);
  const deliveries = session.deliveries;
  const validations =
    t.validations.charAt(0).toUpperCase() + t.validations.slice(1);
  const statusVariant = timelineStateVariant[session.state];
  const update = async (
    patch: TimelineHeaderPatch,
    version = session.version,
  ) => {
    setPending(true);
    setError(null);
    try {
      await onUpdate(patch, version);
      if (patch.title !== undefined) closeEdit();
    } catch {
      setError("save");
    } finally {
      setPending(false);
    }
  };
  const saveTitle = () => {
    if (
      !draft.trim() ||
      [...draft.trim()].length > 200 ||
      /\p{Cc}/u.test(draft)
    ) {
      setError("required");
      return;
    }
    if (editSession.current !== session.viewer_session_id) {
      setError("save");
      return;
    }
    void update({ title: draft.trim() }, editVersion.current);
  };
  const remove = async () => {
    setPending(true);
    setError(null);
    try {
      await onDelete(deleteVersion.current);
      setDeleting(false);
    } catch {
      setError("save");
    } finally {
      setPending(false);
    }
  };
  return (
    <div
      data-slot="timeline-header"
      className="flex min-w-0 flex-col gap-3 border-b border-border px-4 py-3 md:px-6"
    >
      <div className="flex min-w-0 flex-col gap-1">
        {editing ? (
          <div className="flex min-w-0 items-center gap-1">
            <Input
              autoFocus
              aria-label={t.title}
              aria-invalid={Boolean(error)}
              value={draft}
              disabled={pending}
              onChange={(e) => setDraft(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  saveTitle();
                }
                if (e.key === "Escape") {
                  e.preventDefault();
                  closeEdit();
                  setError(null);
                }
              }}
            />
            <IconAction label={t.save} disabled={pending} onClick={saveTitle}>
              <Check aria-hidden />
            </IconAction>
            <IconAction
              label={t.cancel}
              disabled={pending}
              onClick={() => {
                closeEdit();
                setError(null);
              }}
            >
              <X aria-hidden />
            </IconAction>
          </div>
        ) : (
          <div ref={titleRow} className="flex min-w-0 items-center gap-2">
            {titleWraps ? (
              // A title with spaces wraps, long tokens included, and is clamped
              // to two lines; screen readers still read the whole text.
              <h1
                className="line-clamp-2 min-w-0 whitespace-normal text-lg font-semibold tracking-tight [overflow-wrap:anywhere]"
                title={session.title}
              >
                {session.title}
              </h1>
            ) : (
              <h1
                className="min-w-0 overflow-x-auto whitespace-nowrap break-normal text-lg font-semibold tracking-tight focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                aria-describedby={titleHintID}
                // biome-ignore lint/a11y/noNoninteractiveTabindex: An unbroken title is a deliberate keyboard-accessible horizontal reading region.
                tabIndex={0}
                title={session.title}
              >
                {session.title}
              </h1>
            )}
            {!titleWraps && (
              <span id={titleHintID} className="sr-only">
                {lang === "es"
                  ? "Si el nombre no cabe, usa las flechas izquierda y derecha para leerlo completo."
                  : "If the name does not fit, use the left and right arrow keys to read it in full."}
              </span>
            )}
            <IconAction
              label={t.edit}
              disabled={pending}
              onClick={() => {
                setDraft(session.title);
                editVersion.current = session.version;
                editSession.current = session.viewer_session_id;
                setEditing(true);
                setError(null);
              }}
            >
              <Pencil aria-hidden />
            </IconAction>
          </div>
        )}
      </div>
      <div className="flex min-w-0 flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0 flex-1">
          <TimelineSummary content={session.summary} lines={2} lang={lang} />
        </div>
        <div
          className="flex max-w-full flex-wrap items-center justify-end gap-2 self-end lg:shrink-0 lg:self-auto"
          aria-busy={pending}
        >
          {hostActive && (
            <span
              className="text-xs text-muted-foreground"
              data-slot="last-activity"
            >
              {t.lastActivity(timelineAge(hostActivityAt, now))}
            </span>
          )}
          <ButtonGroup aria-label={t.states[session.state]}>
            <ButtonGroupText>
              <Badge variant={statusVariant}>
                <Circle aria-hidden className="size-2 fill-current" />
                {t.states[session.state]}
              </Badge>
            </ButtonGroupText>
          </ButtonGroup>
          <ButtonGroup aria-label={t.actions}>
            <Popover>
              <PopoverTrigger
                render={
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    aria-label={`${t.tickets}: ${session.tickets.length}`}
                    title={`${t.tickets}: ${session.tickets.length}`}
                  />
                }
              >
                <Tickets aria-hidden />
                {session.tickets.length}
              </PopoverTrigger>
              <PopoverContent align="end">
                <PopoverTitle>{t.tickets}</PopoverTitle>
                {session.tickets.length ? (
                  <ul className="flex flex-col gap-2">
                    {session.tickets.map((ticket) => (
                      <li
                        key={ticket.ticket_id}
                        className="flex min-w-0 flex-col gap-0.5"
                      >
                        {isSafeTimelineURL(ticket.url) ? (
                          <a
                            className="text-sm underline underline-offset-4 transition-colors hover:text-muted-foreground"
                            href={ticket.url}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            {ticket.key}
                          </a>
                        ) : (
                          <span className="text-sm">
                            {ticket.key} · {t.ticketUnavailable}
                          </span>
                        )}
                        {ticket.title && (
                          <span className="text-xs text-muted-foreground">
                            {ticket.title}
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">{t.noTickets}</p>
                )}
              </PopoverContent>
            </Popover>
            <Popover>
              <PopoverTrigger
                render={
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    aria-label={`${t.git}: ${deliveries.length}`}
                    title={`${t.git}: ${deliveries.length}`}
                  />
                }
              >
                <GitPullRequest aria-hidden />
                {deliveries.length}
              </PopoverTrigger>
              <PopoverContent align="end">
                <PopoverTitle>{t.git}</PopoverTitle>
                {deliveries.length ? (
                  <ul className="flex flex-col gap-2">
                    {deliveries.map((delivery) => (
                      <li
                        key={delivery.delivery_id}
                        className="flex min-w-0 items-center justify-between gap-2"
                      >
                        {isSafeTimelineURL(delivery.url) ? (
                          <a
                            className="min-w-0 truncate text-sm underline underline-offset-4 transition-colors hover:text-muted-foreground"
                            title={delivery.label}
                            href={delivery.url}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            {delivery.label}
                          </a>
                        ) : (
                          <span className="text-sm">{delivery.label}</span>
                        )}
                        <Badge variant="outline">{t[delivery.status]}</Badge>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    {t.noDeliveries}
                  </p>
                )}
              </PopoverContent>
            </Popover>
            {/* A count, not an action: it stays out of the tab order. */}
            <ButtonGroupText
              title={`${validations}: ${session.validation_count}`}
            >
              <ListChecks aria-hidden />
              <span className="sr-only">{validations}: </span>
              {session.validation_count}
            </ButtonGroupText>
          </ButtonGroup>
          <TimelineParticipants
            data={data}
            lang={lang}
            onRetry={onRetry}
            onRecover={onRecover}
          />
          <ButtonGroup aria-label={t.more}>
            <IconAction
              label={session.archived ? t.unarchive : t.archive}
              disabled={pending}
              onClick={() => void update({ archived: !session.archived })}
            >
              {session.archived ? (
                <ArchiveRestore aria-hidden />
              ) : (
                <Archive aria-hidden />
              )}
            </IconAction>
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    type="button"
                    variant="outline"
                    size="icon-sm"
                    aria-label={t.more}
                    disabled={pending}
                  />
                }
              >
                <MoreHorizontal aria-hidden />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-max">
                <DropdownMenuItem
                  className="whitespace-nowrap"
                  onClick={() => void update({ pinned: !session.pinned })}
                >
                  <Pin aria-hidden />
                  {session.pinned ? t.unpin : t.pin}
                </DropdownMenuItem>
                <DropdownMenuItem
                  className="whitespace-nowrap"
                  variant="destructive"
                  onClick={() => {
                    deleteVersion.current = session.version;
                    setDeleting(true);
                    setError(null);
                  }}
                >
                  <Trash2 aria-hidden />
                  {t.remove}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </ButtonGroup>
          {pending && (
            <span role="status" className="text-xs text-muted-foreground">
              {t.saving}
            </span>
          )}
        </div>
      </div>
      {session.archived && (
        <Alert variant="default" role="status">
          <AlertDescription>
            {session.archive_origin === "automatic"
              ? t.autoarchived
              : t.archivedExplanation}
          </AlertDescription>
          <AlertAction>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={pending}
              onClick={() => void update({ archived: false })}
            >
              {t.restore}
            </Button>
          </AlertAction>
        </Alert>
      )}
      {session.reason && <p className="text-sm">{session.reason}</p>}
      {error && !deleting && (
        <Alert variant="destructive">
          <AlertDescription>
            {error === "required" ? t.titleRequired : t.saveFailed}
          </AlertDescription>
          {error === "save" && (
            <AlertAction>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={onRetry}
              >
                {t.reload}
              </Button>
            </AlertAction>
          )}
        </Alert>
      )}
      <AlertDialog
        open={deleting}
        onOpenChange={(open) => {
          if (!pending) setDeleting(open);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t.remove}</AlertDialogTitle>
            <AlertDialogDescription>{t.removeBody}</AlertDialogDescription>
          </AlertDialogHeader>
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{t.saveFailed}</AlertDescription>
              <AlertAction>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={onRetry}
                >
                  {t.reload}
                </Button>
              </AlertAction>
            </Alert>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={pending}>{t.cancel}</AlertDialogCancel>
            <Button
              type="button"
              variant="destructive"
              disabled={pending}
              onClick={() => void remove()}
            >
              {pending ? t.saving : t.removeConfirm}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
