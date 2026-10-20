import {
  BookOpen,
  ChevronRight,
  Folder,
  Pin,
  Ticket,
  Trash2,
} from "lucide-react";
import type { ReactNode } from "react";
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
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
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
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";
export type ViewerSession = {
  id: string;
  project: string;
  projectKey?: string;
  ticket: string;
  title: string;
  createdAt: string;
  classification: "tracked" | "local" | "unknown";
  pinned?: boolean;
  status?: "in_progress" | "closed" | "unknown";
};
export type ViewerShellProps = {
  sessions: ViewerSession[];
  selectedId?: string;
  state: "ready" | "loading" | "empty" | "error";
  notice?: string;
  onSelect?: (id: string) => void;
  onRetry: () => void;
  onPin?: () => void;
  onDelete?: () => void;
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
      pin: "Conservar",
      unpin: "Dejar de conservar",
      remove: "Borrar copia",
      navigation: "Alternar navegación",
      navigationDescription: "Explora proyectos, tickets y sesiones.",
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
      pin: "Keep",
      unpin: "Stop keeping",
      remove: "Delete copy",
      navigation: "Toggle navigation",
      navigationDescription: "Browse projects, tickets and sessions.",
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
function Navigation({
  sessions,
  selectedId,
  state,
  onSelect,
  onRetry,
}: Pick<
  ViewerShellProps,
  "sessions" | "selectedId" | "state" | "onSelect" | "onRetry"
>) {
  const { setOpenMobile, isMobile } = useSidebar();
  const projects = [
    ...new Set(
      sessions.map((session) => session.projectKey ?? session.project),
    ),
  ];
  return (
    <nav aria-label={copy.walkthroughs}>
      <SidebarGroup>
        <SidebarGroupLabel>{copy.projects}</SidebarGroupLabel>
        <SidebarGroupContent>
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
            <div role="status" className="flex flex-col items-start gap-3 p-2">
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
          ) : (
            <SidebarMenu>
              {projects.map((projectKey) => {
                const projectSessions = sessions.filter(
                  (session) =>
                    (session.projectKey ?? session.project) === projectKey,
                );
                const project = projectSessions[0].project;
                const tickets = [
                  ...new Set(projectSessions.map((session) => session.ticket)),
                ];
                return (
                  <Collapsible
                    key={projectKey}
                    defaultOpen
                    className="group/project font-semibold"
                    render={<SidebarMenuItem />}
                  >
                    <CollapsibleTrigger
                      render={
                        <SidebarMenuButton type="button" title={project} />
                      }
                    >
                      <Folder aria-hidden />
                      <span className="min-w-0 flex-1 truncate">{project}</span>
                      <ChevronRight
                        aria-hidden
                        className="group-data-open/project:rotate-90"
                      />
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <SidebarMenuSub className="navigation-branch">
                        {tickets.map((ticket) => {
                          const ticketSessions = projectSessions.filter(
                            (session) => session.ticket === ticket,
                          );
                          return (
                            <Collapsible
                              key={ticket}
                              defaultOpen
                              className="navigation-node group/ticket"
                              render={<SidebarMenuSubItem />}
                            >
                              <CollapsibleTrigger
                                render={
                                  <SidebarMenuButton
                                    type="button"
                                    title={ticket}
                                    className="font-medium"
                                  />
                                }
                              >
                                <Ticket
                                  aria-hidden
                                  className="text-muted-foreground"
                                />
                                <span className="min-w-0 flex-1 truncate">
                                  {ticket}
                                </span>
                                <ChevronRight
                                  aria-hidden
                                  className="text-muted-foreground group-data-open/ticket:rotate-90"
                                />
                              </CollapsibleTrigger>
                              <CollapsibleContent>
                                <SidebarMenuSub className="navigation-branch">
                                  {ticketSessions.map((session) => (
                                    <SidebarMenuSubItem
                                      key={session.id}
                                      className="navigation-node navigation-session"
                                    >
                                      <button
                                        type="button"
                                        data-session-key={session.id}
                                        className="session-navigation-link flex w-full min-w-0 items-start gap-2 text-left text-sm text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:outline-2 focus-visible:outline-sidebar-ring"
                                        aria-current={
                                          selectedId === session.id
                                            ? "page"
                                            : undefined
                                        }
                                        onClick={() => {
                                          onSelect?.(session.id);
                                          setOpenMobile(false);
                                          if (isMobile)
                                            requestAnimationFrame(() =>
                                              document
                                                .getElementById("document")
                                                ?.focus({
                                                  preventScroll: true,
                                                }),
                                            );
                                        }}
                                      >
                                        <BookOpen
                                          aria-hidden
                                          className="size-4 shrink-0"
                                        />
                                        <div className="flex min-w-0 flex-1 flex-col gap-1">
                                          <div className="flex min-w-0 items-start gap-2">
                                            <span className="session-navigation-title min-w-min flex-1 font-normal">
                                              {session.title}
                                            </span>
                                            {session.pinned && (
                                              <Pin
                                                aria-hidden
                                                className="size-4 shrink-0"
                                              />
                                            )}
                                          </div>
                                          <div className="flex flex-wrap items-center gap-2">
                                            <span className="text-xs tabular-nums text-muted-foreground">
                                              {session.createdAt}
                                            </span>
                                            <Badge
                                              variant={
                                                session.status === "closed"
                                                  ? "success-light"
                                                  : session.status ===
                                                      "in_progress"
                                                    ? "primary-light"
                                                    : "outline"
                                              }
                                            >
                                              {
                                                copy.sessionStatus[
                                                  session.status ?? "unknown"
                                                ]
                                              }
                                            </Badge>
                                          </div>
                                        </div>
                                      </button>
                                    </SidebarMenuSubItem>
                                  ))}
                                </SidebarMenuSub>
                              </CollapsibleContent>
                            </Collapsible>
                          );
                        })}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </Collapsible>
                );
              })}
            </SidebarMenu>
          )}
        </SidebarGroupContent>
      </SidebarGroup>
    </nav>
  );
}
export function ViewerShell({
  sessions,
  selectedId,
  state,
  notice,
  onSelect,
  onRetry,
  onPin,
  onDelete,
  children,
  panel,
  libraryFooter,
  notices,
  writing,
  coldLoading = false,
}: ViewerShellProps) {
  const selected = sessions.find((session) => session.id === selectedId);
  return (
    <SidebarProvider
      className="viewer-shell min-h-0 text-sm text-foreground"
      style={{ "--sidebar-width": "20rem" } as React.CSSProperties}
    >
      <a className="skip-link" href="#document">
        {copy.select}
      </a>
      <Sidebar variant="floating">
        <SidebarHeader>
          <div className="flex items-center gap-3 px-2 py-3">
            <BookOpen aria-hidden className="size-5 shrink-0" />
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
            onRetry={onRetry}
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
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div className="flex min-w-0 items-center gap-3">
                  <div className="flex min-w-0 flex-col gap-0.5">
                    <p className="text-xs text-muted-foreground">
                      {selected
                        ? `${selected.project} / ${selected.ticket}`
                        : copy.readOnly}
                    </p>
                    <h1 className="break-words text-lg font-semibold tracking-tight">
                      {selected?.title ?? copy.library}
                    </h1>
                  </div>
                </div>
                {selected && (
                  <div className="flex items-center gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={!onPin || writing}
                      onClick={onPin}
                    >
                      <Pin aria-hidden />
                      {selected.pinned ? copy.unpin : copy.pin}
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      disabled={!onDelete || writing}
                      onClick={onDelete}
                    >
                      <Trash2 aria-hidden />
                      {copy.remove}
                    </Button>
                  </div>
                )}
              </div>
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
