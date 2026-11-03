import { type CSSProperties, useEffect } from "react";
import { Toaster, toast } from "sonner";
import type { ViewerController, ViewerState } from "./controller";
import { messages as t } from "./messages";

export function ArchiveNotification({
  controller,
  undo,
}: {
  controller: ViewerController;
  undo: ViewerState["undoArchive"];
}) {
  useEffect(() => {
    if (!undo) return;
    let active = true;
    let restoring = false;
    let id: string | number | undefined;
    const clear = (notification: { id: string | number }) => {
      if (active && !restoring && notification.id === id)
        controller.clearUndo(undo);
    };
    const show = (failed = false) => {
      id = toast(failed ? t.writeError : t.archived, {
        id,
        description: <span title={undo.title}>{undo.title}</span>,
        duration: 8000,
        closeButton: true,
        onDismiss: clear,
        onAutoClose: clear,
        action: {
          label: failed ? t.retry : t.undo,
          onClick(event) {
            // Keep the action available on failure; successful restoration clears it.
            event.preventDefault();
            if (restoring) return;
            restoring = true;
            const previous = id;
            id = toast.loading(t.unarchive, {
              description: <span title={undo.title}>{undo.title}</span>,
              duration: Infinity,
              dismissible: false,
              closeButton: false,
            });
            if (previous !== undefined) toast.dismiss(previous);
            void controller
              .undoArchive(undo)
              .catch(() => {
                if (active && controller.getSnapshot().undoArchive === undo)
                  show(true);
              })
              .finally(() => {
                restoring = false;
              });
          },
        },
      });
    };
    show();
    return () => {
      active = false;
      if (id !== undefined) toast.dismiss(id);
    };
  }, [controller, undo]);

  return (
    <Toaster
      position="bottom-right"
      containerAriaLabel={t.notifications}
      theme="system"
      style={
        {
          "--normal-bg": "var(--card)",
          "--normal-text": "var(--foreground)",
          "--normal-border": "var(--border)",
        } as CSSProperties
      }
      toastOptions={{
        closeButtonAriaLabel: t.dismiss,
        descriptionClassName: "line-clamp-2 text-muted-foreground!",
        actionButtonStyle: {
          background: "var(--primary)",
          color: "var(--primary-foreground)",
        },
      }}
    />
  );
}
