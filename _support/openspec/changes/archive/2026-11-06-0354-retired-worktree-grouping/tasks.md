# Work record

### T7 — Resolve retired-source presentation

- [x] Implement same-project witness resolution and conservative fallback.

**Closes:** AC15, AC16, AC17

Owner HTTP navigation regression reproduced the intended failure in0.397s and passes in0.506s, including contradictory declarations and immutable stored state. Source candidate a1ddf64e. Native Codex review completed with no actionable defects.

### T8 — Verify candidate and document the exception

- [x] Update public viewer description, changelog and isolated requirement delta.
- [x] Verify actual index projection read-only and affected runtime checks.

**Closes:** AC15, AC16, AC17

Live probe over65 actual records: warning origins1→0, both Hive works share tricell-hive, unavailable original source remains preserved. Vet passed. Viewer38.777s, CLI212.839s, Astro17pages passed. The initial base-spec edit triggered archived-delta mismatch; moved the contract exception into this active delta and restored the base spec for the normal close merger. Historical archived records remain unchanged.

### T9 — Integrate and refresh local runtime

- [x] Human acceptance, push/PR/merge, maintainer refresh and same-settings viewer restart with live verification.

## Checkpoint

Integrated PR393 as c0286f86c915d75881e4509b84b70a44f81ca096 after D7 A. No configured development CI or pending reviews at the single read. Local binary rebuilt from clean integrated source; existing viewer restarted with same state, port4483 and TTL0s. Actual health revision matches c0286f86, PID51279,65 previous works and registry hash preserved, original source metadata preserved; warning origins1to0 and both Hive works share one group. Old viewer tabs lost access; offer open viewer on request.

Viewer work6790dad82652c4d9425412b13498149f920eafba1377376411bdc4aebd5c26ec closed and read back at version9; receipt retained with the local evidence. A first post-restart publication read returned work_unavailable transiently; subsequent work show/native session read confirmed work and accepted original producer. Do not restart viewer for cleanup.

User selected close (A). The MODIFIED delta is merged by the deterministic merger and this record archived; own merged branch/worktree cleanup follows archive integration. Unique evidence is retained in the main checkout at _support/workspace/2026-11-05-retired-worktree-grouping. Native review /private/tmp/hive-retired-worktree-grouping-review.txt; read-only actual-data probe /private/tmp/hive-retired-worktree-live-result.txt and temp overlay directory in /private/tmp/hive-retired-worktree-live-path; before/after restart summaries /private/tmp/hive-retired-worktree-before-restart.json and -after-restart.json; checkpoint /private/tmp/hive-retired-worktree-grouping-checkpoint.md. site/node_modules is a task-created symlink to main dependencies (remove only link); site/dist and .astro reproducible. No task test processes remain. Main unrelated3 session folders preserved. No new public release or master promotion authorized.
