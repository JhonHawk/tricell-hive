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

- [ ] Human acceptance, push/PR/merge, maintainer refresh and same-settings viewer restart with live verification.

## Checkpoint

Candidate verified; finish content-delta check and commit record, then D7 human acceptance before push. Worktree .claude/worktrees/fix-retired-worktree-grouping, branch fix/retired-worktree-grouping, base084728fa. Native log /private/tmp/hive-retired-worktree-grouping-review.txt. Read-only actual-index probe /private/tmp/hive-retired-worktree-live-result.txt and overlay directory named by /private/tmp/hive-retired-worktree-live-path. Host checkpoint /private/tmp/hive-retired-worktree-grouping-checkpoint.md. Task-owned site/node_modules symlink must be removed without touching its main-checkout target; generated site/dist and .astro are reproducible. Existing viewer PID21909 rev0b69859c, port4483, default state, TTL0s. No restart or publication yet. Viewer work6790dad82652c4d9425412b13498149f920eafba1377376411bdc4aebd5c26ec (source main checkout). Main unrelated three session folders preserved.
