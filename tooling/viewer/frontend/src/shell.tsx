import {
  Archive,
  ArchiveRestore,
  BookOpen,
  ChevronRight,
  Folder,
  Pencil,
  Pin,
  Route,
  Trash2,
} from "lucide-react";
import {
  type ReactNode,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/reui/alert";
import { Badge } from "@/components/reui/badge";
import {
  Frame,
  FrameDescription,
  FrameFooter,
  FrameHeader,
  FramePanel,
  FrameTitle,
} from "@/components/reui/frame";
import { Button } from "@/components/ui/button";
import { ButtonGroup } from "@/components/ui/button-group";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Input } from "@/components/ui/input";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { optionalStorage, save, stored } from "./api";
import { messages as t } from "./messages";
export type ViewerSession = {
  id: string;
  segments?: { id: string; title: string; ticket: string }[];
  project: string;
  projectKey?: string;
  ticket: string;
  title: string;
  localTitle?: boolean;
  createdAt: string;
  classification: "tracked" | "local" | "unknown";
  pinned?: boolean;
  archived?: boolean;
  excluded?: boolean;
  status?: "in_progress" | "closed" | "unknown";
};
export type ViewerShellProps = {
  sessions: ViewerSession[];
  selectedId?: string;
  state: "ready" | "loading" | "empty" | "error";
  notice?: string;
  onSelect?: (id: string) => void;
  onJump?: (id: string) => void;
  onRetry: () => void;
  onPin?: () => void;
  onDelete?: () => void;
  onArchive?: () => void;
  onRename?: () => void;
  feedback?: ReactNode;
  showArchived?: boolean;
  onShowArchivedChange?: (show: boolean) => void;
  children?: ReactNode;
  panel?: ReactNode;
  libraryFooter?: ReactNode;
  notices?: ReactNode;
  writing?: boolean;
  coldLoading?: boolean;
};
const spanish = navigator.language.startsWith("es");
const copy = spanish
  ? {
      walkthroughs: "Recorridos",
      library: "Biblioteca de recorridos",
      description: "Evidencia y validación de los cambios de tu proyecto.",
      projects: "Proyectos",
      showArchived: "Mostrar archivados",
      hideArchived: "Ocultar archivados",
      archive: "Archivar sesión",
      unarchive: "Restaurar sesión",
      actions: "Acciones de la sesión",
      statusDescription:
        "Estado declarado del trabajo. Cerrado no significa aprobado.",
      excluded: "Copias apartadas",
      excludedOnly:
        "Solo hay copias apartadas. Selecciona una abajo para conocer el motivo.",
      excludedDescription:
        "El manifiesto fuente está versionado en Git. Estas copias se mantienen fuera del árbol de proyectos; su documento no se abre aquí.",
      empty: "Aún no hay recorridos",
      emptyDescription: "Registra un recorrido con hive view para verlo aquí.",
      retry: "Actualizar",
      error: "No se pudo cargar la biblioteca",
      errorDescription:
        "Comprueba que el visor local sigue en ejecución y vuelve a intentarlo.",
      loading: "Cargando recorridos",
      select: "Selecciona un recorrido",
      selectDescription:
        "Abre una sesión en la navegación para revisar su evidencia.",
      pin: "Conservar sesión",
      unpin: "Dejar de conservar sesión",
      remove: "Eliminar sesión",
      navigation: "Alternar navegación",
      navigationDescription: "Explora proyectos y sesiones.",
      tracked: "Versionado",
      local: "Local",
      unknown: "Origen desconocido",
      readOnly: "Visor local",
      session: "Sesión",
      sessionIdentifier: "Identificador de sesión",
      sessionStatus: {
        in_progress: "En curso",
        closed: "Cerrada",
        unknown: "Sin informar",
      },
    }
  : {
      walkthroughs: "Walkthroughs",
      library: "Walkthrough library",
      description: "Evidence and validation for your project changes.",
      projects: "Projects",
      showArchived: "Show archived",
      hideArchived: "Hide archived",
      archive: "Archive session",
      unarchive: "Restore session",
      actions: "Session actions",
      statusDescription: "Declared work status. Closed does not mean approved.",
      excluded: "Set-aside copies",
      excludedOnly:
        "Only set-aside copies are available. Select one below to see why.",
      excludedDescription:
        "The source manifest is tracked in Git. These copies stay outside the project tree; their documents cannot be opened here.",
      empty: "No walkthroughs yet",
      emptyDescription: "Register a walkthrough with hive view to see it here.",
      retry: "Refresh",
      error: "Could not load the library",
      errorDescription:
        "Check that the local viewer is still running and try again.",
      loading: "Loading walkthroughs",
      select: "Select a walkthrough",
      selectDescription:
        "Open a session in the navigation to review its evidence.",
      pin: "Keep session",
      unpin: "Stop keeping session",
      remove: "Delete session",
      navigation: "Toggle navigation",
      navigationDescription: "Browse projects and sessions.",
      tracked: "Tracked",
      local: "Local",
      unknown: "Unknown origin",
      readOnly: "Local viewer",
      session: "Session",
      sessionIdentifier: "Session identifier",
      sessionStatus: {
        in_progress: "In progress",
        closed: "Closed",
        unknown: "Not reported",
      },
    };
const NAVIGATION_KEY = "hive-view-navigation";
type NavigationPreferences = {
  sidebarOpen: boolean;
  projects: Record<string, boolean>;
};
function readNavigationPreferences(): NavigationPreferences {
  try {
    const value = JSON.parse(stored(optionalStorage(), NAVIGATION_KEY));
    return {
      sidebarOpen: value.sidebarOpen !== false,
      projects: Object.fromEntries(
        Object.entries(value.projects ?? {}).filter(
          ([, open]) => typeof open === "boolean",
        ),
      ) as Record<string, boolean>,
    };
  } catch {
    return { sidebarOpen: true, projects: {} };
  }
}
export function ActionButton({
  label,
  children,
  description,
  ...props
}: React.ComponentProps<typeof Button> & {
  label: string;
  description?: string;
}) {
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
      <TooltipContent>
        {label}
        {description ? `. ${description}` : ""}
      </TooltipContent>
    </Tooltip>
  );
}
function sessionDisplayTitle(session: ViewerSession) {
  if (session.localTitle) return session.title;
  if (!/^(?:#\d+|[a-z][a-z0-9]*-\d+)$/i.test(session.ticket))
    return session.title;
  const prefix = session.title.slice(0, session.ticket.length);
  if (prefix.toLowerCase() !== session.ticket.toLowerCase())
    return session.title;
  const remainder = session.title.slice(session.ticket.length);
  const match = remainder.match(/^\s*(?:[·:|–—]|-(?=\s))\s+(.+)$/);
  return match?.[1].trim() || session.title;
}
function SessionNavigationButton({
  session,
  selectedId,
  onSelect,
}: {
  session: ViewerSession;
  selectedId?: string;
  onSelect?: (id: string) => void;
}) {
  const { setOpenMobile, isMobile } = useSidebar();
  return (
    <button
      type="button"
      data-session-key={session.id}
      title={
        session.segments
          ? session.title
          : `${session.title} — ${copy.sessionStatus[session.status ?? "unknown"]}. ${copy.statusDescription}`
      }
      className="session-navigation-link flex w-full min-w-0 items-start gap-2 text-left text-sm text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:outline-2 focus-visible:outline-sidebar-ring"
      aria-current={selectedId === session.id ? "page" : undefined}
      onClick={() => {
        onSelect?.(session.id);
        setOpenMobile(false);
        if (isMobile)
          requestAnimationFrame(() =>
            document.getElementById("document")?.focus({
              preventScroll: true,
            }),
          );
      }}
    >
      {!session.segments && (
        <span
          aria-hidden="true"
          data-session-status={session.status ?? "unknown"}
          className="session-status-dot mt-1.5 size-1.5 shrink-0 rounded-full"
        />
      )}
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex min-w-0 items-start gap-2">
          <span className="session-navigation-title min-w-0 flex-1 font-normal">
            {sessionDisplayTitle(session)}
          </span>
          {session.pinned && <Pin aria-hidden className="size-3 shrink-0" />}
        </div>
        <div className="flex min-w-0 items-center gap-1 text-xs text-muted-foreground">
          <span className="truncate">{session.ticket}</span>
          {session.segments && (
            <span className="shrink-0">
              · {session.segments.length}{" "}
              {session.segments.length === 1 ? t.segment : t.segments}
            </span>
          )}
          <span aria-hidden="true">·</span>
          <span className="shrink-0 tabular-nums">{session.createdAt}</span>
          <span className="sr-only">
            {!session.segments &&
              copy.sessionStatus[session.status ?? "unknown"]}
            . {!session.segments && copy.statusDescription}
          </span>
        </div>
      </div>
    </button>
  );
}
function Navigation({
  sessions,
  selectedId,
  state,
  onSelect,
  onJump,
  onRetry,
  showArchived,
  onShowArchivedChange,
  preferences,
  onProjectOpenChange,
}: Pick<
  ViewerShellProps,
  | "sessions"
  | "selectedId"
  | "state"
  | "onSelect"
  | "onJump"
  | "onRetry"
  | "showArchived"
  | "onShowArchivedChange"
> & {
  preferences: NavigationPreferences;
  onProjectOpenChange: (key: string, open: boolean) => void;
}) {
  const { setOpenMobile } = useSidebar();
  const navigationRef = useRef<HTMLDivElement>(null);
  const focusKey = useRef<string | undefined>(undefined);
  const [query, setQuery] = useState("");
  const matches = (session: ViewerSession) =>
    !query.trim() ||
    session.title
      .toLocaleLowerCase()
      .includes(query.trim().toLocaleLowerCase()) ||
    (session.segments ?? [session]).some((segment) =>
      `${segment.title} ${segment.ticket}`
        .toLocaleLowerCase()
        .includes(query.trim().toLocaleLowerCase()),
    );
  const normal = sessions.filter(
    (session) =>
      !session.excluded &&
      session.classification !== "tracked" &&
      matches(session),
  );
  const excluded = sessions.filter(
    (session) => session.excluded || session.classification === "tracked",
  );
  useLayoutEffect(() => {
    if (
      focusKey.current &&
      sessions.some((session) => session.id === focusKey.current) &&
      document.activeElement === document.body
    ) {
      const target = [
        ...(navigationRef.current?.querySelectorAll<HTMLElement>(
          "[data-session-key]",
        ) ?? []),
      ].find((node) => node.dataset.sessionKey === focusKey.current);
      target?.focus({ preventScroll: true });
    }
  }, [sessions]);
  const projects = [
    ...new Set(normal.map((session) => session.projectKey ?? session.project)),
  ];
  return (
    <div
      ref={navigationRef}
      onFocusCapture={(event) => {
        focusKey.current = (event.target as HTMLElement).closest<HTMLElement>(
          "[data-session-key]",
        )?.dataset.sessionKey;
      }}
      onBlurCapture={(event) => {
        if (
          (event.relatedTarget instanceof Node &&
            !event.currentTarget.contains(event.relatedTarget)) ||
          (!event.relatedTarget && event.target.isConnected)
        )
          focusKey.current = undefined;
      }}
    >
      <nav aria-label={copy.walkthroughs}>
        <SidebarGroup>
          <div className="flex items-center justify-between gap-2">
            <SidebarGroupLabel>{copy.projects}</SidebarGroupLabel>
            <ActionButton
              label={showArchived ? copy.hideArchived : copy.showArchived}
              variant="ghost"
              aria-pressed={Boolean(showArchived)}
              disabled={!onShowArchivedChange}
              onClick={() => onShowArchivedChange?.(!showArchived)}
            >
              <Archive aria-hidden />
            </ActionButton>
          </div>
          <SidebarGroupContent>
            <div className="flex flex-col gap-2 p-2">
              <Input
                type="search"
                className="border-ring"
                aria-label={t.search}
                placeholder={t.search}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
              {query && !normal.length && (
                <div role="status" className="flex flex-col items-start gap-2">
                  <p className="text-sm text-muted-foreground">{t.noMatches}</p>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    onClick={() => setQuery("")}
                  >
                    {t.clearSearch}
                  </Button>
                </div>
              )}
            </div>
            {state === "loading" ? (
              <div
                role="status"
                aria-busy="true"
                className="flex flex-col gap-3 p-2"
              >
                <span className="sr-only">{copy.loading}</span>
                <Skeleton className="h-8 w-full" />
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
              </div>
            ) : state === "error" || sessions.length === 0 ? (
              <div
                role="status"
                className="flex flex-col items-start gap-3 p-2"
              >
                <p className="text-sm text-muted-foreground">
                  {state === "error" ? copy.error : copy.empty}
                </p>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={onRetry}
                >
                  {copy.retry}
                </Button>
              </div>
            ) : normal.length === 0 ? (
              <p className="p-2 text-sm text-muted-foreground">
                {copy.excludedOnly}
              </p>
            ) : (
              <SidebarMenu>
                {projects.map((projectKey) => {
                  const projectSessions = normal.filter(
                    (session) =>
                      (session.projectKey ?? session.project) === projectKey,
                  );
                  const project = projectSessions[0].project;
                  return (
                    <Collapsible
                      key={projectKey}
                      open={
                        (query ? true : preferences.projects[projectKey]) ??
                        projectSessions.some(
                          (session) => session.status === "in_progress",
                        )
                      }
                      onOpenChange={(open) =>
                        onProjectOpenChange(projectKey, open)
                      }
                      className="group/project font-semibold"
                      render={<SidebarMenuItem />}
                    >
                      <CollapsibleTrigger
                        render={
                          <SidebarMenuButton type="button" title={project} />
                        }
                      >
                        <Folder aria-hidden />
                        <span className="min-w-0 flex-1 truncate">
                          {project}
                        </span>
                        <span className="text-xs tabular-nums text-muted-foreground">
                          {projectSessions.length}
                        </span>
                        <ChevronRight
                          aria-hidden
                          className="text-muted-foreground group-data-open/project:rotate-90"
                        />
                      </CollapsibleTrigger>
                      <CollapsibleContent>
                        <SidebarMenu>
                          {projectSessions.map((session) => (
                            <SidebarMenuItem key={session.id}>
                              <SessionNavigationButton
                                session={session}
                                selectedId={selectedId}
                                onSelect={onSelect}
                              />
                              {query &&
                                session.segments
                                  ?.filter((segment) =>
                                    `${segment.title} ${segment.ticket}`
                                      .toLocaleLowerCase()
                                      .includes(
                                        query.trim().toLocaleLowerCase(),
                                      ),
                                  )
                                  .map((segment) => (
                                    <Button
                                      key={segment.id}
                                      type="button"
                                      size="sm"
                                      variant="ghost"
                                      className="max-w-full justify-start"
                                      onClick={() => {
                                        onJump?.(segment.id);
                                        setOpenMobile(false);
                                      }}
                                    >
                                      <span className="truncate">
                                        {segment.ticket} · {segment.title}
                                      </span>
                                    </Button>
                                  ))}
                            </SidebarMenuItem>
                          ))}
                        </SidebarMenu>
                      </CollapsibleContent>
                    </Collapsible>
                  );
                })}
              </SidebarMenu>
            )}
          </SidebarGroupContent>
        </SidebarGroup>
      </nav>
      {excluded.length > 0 && (
        <section aria-label={copy.excluded}>
          <SidebarGroup>
            <SidebarGroupLabel>
              {copy.excluded} ({excluded.length})
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <p className="p-2 text-xs text-muted-foreground">
                {copy.excludedDescription}
              </p>
              <SidebarMenu>
                {excluded.map((session) => (
                  <SidebarMenuItem key={session.id}>
                    <SessionNavigationButton
                      session={session}
                      selectedId={selectedId}
                      onSelect={onSelect}
                    />
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </section>
      )}
    </div>
  );
}
export function ViewerShell({
  sessions,
  selectedId,
  state,
  notice,
  onSelect,
  onJump,
  onRetry,
  onPin,
  onDelete,
  onArchive,
  onRename,
  feedback,
  showArchived = false,
  onShowArchivedChange,
  children,
  panel,
  libraryFooter,
  notices,
  writing,
  coldLoading = false,
}: ViewerShellProps) {
  const selected = sessions.find((session) => session.id === selectedId);
  const [preferences, setPreferences] = useState(readNavigationPreferences);
  useEffect(() => {
    save(optionalStorage(), NAVIGATION_KEY, JSON.stringify(preferences));
  }, [preferences]);
  return (
    <SidebarProvider
      open={preferences.sidebarOpen}
      onOpenChange={(sidebarOpen) =>
        setPreferences((previous) => ({ ...previous, sidebarOpen }))
      }
      className="viewer-shell min-h-0 text-sm text-foreground"
      style={{ "--sidebar-width": "20rem" } as React.CSSProperties}
    >
      <a className="skip-link" href="#document">
        {copy.select}
      </a>
      <Sidebar variant="floating">
        <SidebarHeader>
          <div className="flex items-center gap-3 px-2 py-3">
            <Route aria-hidden className="size-5 shrink-0" />
            <div className="flex min-w-0 flex-1 flex-col gap-0.5">
              <span className="font-semibold tracking-tight">hive</span>
              <span className="text-xs text-muted-foreground">
                {copy.walkthroughs}
              </span>
            </div>
            <SidebarTrigger
              type="button"
              className="md:hidden"
              aria-label={copy.navigation}
            />
          </div>
        </SidebarHeader>
        <SidebarSeparator />
        <SidebarContent tabIndex={0} aria-label={copy.walkthroughs}>
          <Navigation
            sessions={sessions}
            selectedId={selectedId}
            state={state}
            onSelect={onSelect}
            onJump={onJump}
            onRetry={onRetry}
            showArchived={showArchived}
            onShowArchivedChange={onShowArchivedChange}
            preferences={preferences}
            onProjectOpenChange={(key, open) =>
              setPreferences((previous) => ({
                ...previous,
                projects: { ...previous.projects, [key]: open },
              }))
            }
          />
        </SidebarContent>
        <SidebarSeparator />
        <SidebarFooter>
          <div className="flex flex-col gap-1 p-2">
            <span className="text-xs font-medium">{copy.readOnly}</span>
            <p className="text-xs text-muted-foreground">{copy.description}</p>
            {libraryFooter}
          </div>
        </SidebarFooter>
      </Sidebar>
      <SidebarInset className="min-h-0 min-w-0">
        <header className="flex shrink-0 items-center justify-between gap-4 border-b border-border px-4 py-3 md:px-6">
          <div className="flex min-w-0 items-center gap-3">
            <SidebarTrigger type="button" aria-label={copy.navigation} />
            <span className="font-medium">{copy.walkthroughs}</span>
          </div>
          <Badge variant="outline">{copy.readOnly}</Badge>
        </header>
        {coldLoading ? (
          <div
            data-slot="session-toolbar"
            role="status"
            className="shrink-0 border-b border-border p-4"
          >
            <span className="sr-only">{copy.loading}</span>
            <Skeleton className="h-8 w-2/3" />
          </div>
        ) : (
          <div
            data-slot="session-toolbar"
            className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3 md:px-6"
          >
            <div className="flex min-w-0 items-center gap-3">
              <div className="flex min-w-0 flex-col gap-0.5">
                <p className="text-xs text-muted-foreground">
                  {selected
                    ? `${selected.project} / ${selected.ticket}`
                    : copy.readOnly}
                </p>
                <h1
                  title={selected?.title}
                  // biome-ignore lint/a11y/noNoninteractiveTabindex: The name is a deliberate keyboard-accessible horizontal reading region.
                  tabIndex={0}
                  aria-describedby="session-name-reading-hint"
                  style={{
                    whiteSpace: /\s/u.test(selected?.title ?? copy.library)
                      ? "normal"
                      : "nowrap",
                  }}
                  className="min-w-0 overflow-x-auto whitespace-normal break-normal text-lg font-semibold tracking-tight focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                >
                  {selected?.title ?? copy.library}
                </h1>
                <span id="session-name-reading-hint" className="sr-only">
                  {t.nameReadingHint}
                </span>
                {selected && (
                  <p className="text-xs text-muted-foreground">
                    {selected.createdAt} · {selected.segments?.length ?? 1}{" "}
                    {t.segments}
                  </p>
                )}
              </div>
            </div>
            {selected && (onPin || onArchive || onDelete) && (
              <ButtonGroup aria-label={copy.actions} aria-busy={writing}>
                {onRename && (
                  <ActionButton
                    label={t.rename}
                    disabled={writing}
                    onClick={onRename}
                  >
                    <Pencil aria-hidden />
                  </ActionButton>
                )}
                <ActionButton
                  label={selected.archived ? copy.unarchive : copy.archive}
                  disabled={!onArchive || writing}
                  onClick={onArchive}
                >
                  {selected.archived ? (
                    <ArchiveRestore aria-hidden />
                  ) : (
                    <Archive aria-hidden />
                  )}
                </ActionButton>
                <ActionButton
                  label={selected.pinned ? copy.unpin : copy.pin}
                  description={t.keepHelp}
                  aria-pressed={Boolean(selected.pinned)}
                  disabled={!onPin || writing}
                  onClick={onPin}
                >
                  <Pin aria-hidden />
                </ActionButton>
                <ActionButton
                  label={copy.remove}
                  disabled={!onDelete || writing}
                  onClick={onDelete}
                >
                  <Trash2 aria-hidden />
                </ActionButton>
              </ButtonGroup>
            )}
          </div>
        )}
        {feedback && (
          <div
            data-slot="session-feedback"
            className="shrink-0 px-4 py-2 md:px-6"
          >
            {feedback}
          </div>
        )}
        <div
          data-slot="viewer-canvas"
          aria-busy={coldLoading}
          className="flex min-h-0 min-w-0 flex-1 flex-col gap-6 overflow-auto p-4 md:p-6"
        >
          {coldLoading ? (
            <div
              role="status"
              aria-label={copy.loading}
              className="flex min-h-0 flex-1 flex-col"
            >
              <span className="sr-only">{copy.loading}</span>
              <Skeleton aria-hidden="true" className="min-h-0 w-full flex-1" />
            </div>
          ) : (
            <>
              {notice && (
                <Alert variant="warning" role="status">
                  <AlertTitle>{copy.readOnly}</AlertTitle>
                  <AlertDescription className="text-foreground">
                    {notice}
                  </AlertDescription>
                </Alert>
              )}
              {notices}
              <Frame spacing="default">
                <FrameHeader>
                  <FrameTitle>
                    {selected ? copy.session : copy.walkthroughs}
                  </FrameTitle>
                  <FrameDescription>
                    {selected?.createdAt ?? copy.description}
                  </FrameDescription>
                </FrameHeader>
                <FramePanel>
                  {panel ??
                    (state === "loading" ? (
                      <div
                        role="status"
                        aria-busy="true"
                        className="flex flex-col gap-4"
                      >
                        <span className="sr-only">{copy.loading}</span>
                        <Skeleton className="h-5 w-2/3" />
                        <Skeleton className="h-4 w-1/2" />
                        <Skeleton className="h-24 w-full" />
                      </div>
                    ) : state === "error" ? (
                      <Alert variant="destructive">
                        <AlertTitle>{copy.error}</AlertTitle>
                        <AlertDescription>
                          {copy.errorDescription}
                        </AlertDescription>
                        <AlertAction>
                          <Button
                            type="button"
                            size="sm"
                            variant="outline"
                            onClick={onRetry}
                          >
                            {copy.retry}
                          </Button>
                        </AlertAction>
                      </Alert>
                    ) : state === "empty" ? (
                      <div
                        role="status"
                        className="flex flex-col items-start gap-3"
                      >
                        <BookOpen
                          aria-hidden
                          className="size-5 text-muted-foreground"
                        />
                        <h2 className="font-medium">{copy.empty}</h2>
                        <p className="text-muted-foreground">
                          {copy.emptyDescription}
                        </p>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={onRetry}
                        >
                          {copy.retry}
                        </Button>
                      </div>
                    ) : selected ? (
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge
                          variant={
                            selected.classification === "unknown"
                              ? "warning-light"
                              : "outline"
                          }
                        >
                          {copy[selected.classification]}
                        </Badge>
                        <section
                          aria-label={copy.sessionIdentifier}
                          // biome-ignore lint/a11y/noNoninteractiveTabindex: This scroll region exposes the complete identifier to keyboard users.
                          tabIndex={0}
                          className="min-w-0 w-full overflow-x-auto whitespace-nowrap text-xs text-muted-foreground"
                        >
                          <code>{selected.id}</code>
                        </section>
                      </div>
                    ) : (
                      <div className="flex flex-col gap-3">
                        <p>{copy.selectDescription}</p>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={onRetry}
                        >
                          {copy.retry}
                        </Button>
                      </div>
                    ))}
                </FramePanel>
                {!selected && (
                  <FrameFooter>
                    <p className="text-xs text-muted-foreground">
                      {copy.selectDescription}
                    </p>
                  </FrameFooter>
                )}
              </Frame>
              <section
                id="document"
                tabIndex={-1}
                aria-label={selected?.title ?? copy.select}
                className="document-region min-w-0"
              >
                {children}
              </section>
            </>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
