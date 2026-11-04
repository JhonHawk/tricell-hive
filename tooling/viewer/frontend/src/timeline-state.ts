import type { TimelineState } from "./timeline-contract";

/** Installed ReUI semantic variants shared by navigation and selected work. */
export const timelineStateVariant = {
  active: "info-light",
  in_review: "focus-light",
  awaiting_validation: "warning-light",
  paused: "secondary",
  closed: "success-light",
  failed: "destructive-light",
} as const satisfies Record<TimelineState, string>;
