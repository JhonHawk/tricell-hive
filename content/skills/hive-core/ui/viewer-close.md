# Close and manage a viewer work

Read when closing a work with captures, or managing a viewer work. Closing is the work's completion phase; managing a viewer work is when the user asks to rename, archive, restore, pin or delete one.

## Close and input hygiene

At the work's completion phase, as [flow-close](../../flow-close/SKILL.md) step 1 sets, and only for a work with captures: run `hive view work close --outcome completed|abandoned --reason TEXT` for integrated or abandoned work. For a work that waits, run `hive view work sync --stage STAGE --state in_review|paused|awaiting_validation --reason TEXT` instead, with the stage of its last capture: `in_review` while its open pull request waits for the user's merge or a human reviewer, `paused` for a pause. Both read the work back; do not close an already closed work again or infer a close from a document's status, since the readback is the record of the work's state. Keep the readback and its receipts, recording their location for a later input cleanup that the close left pending. Remove the task's disposable capture inputs only after that readback confirms the work and no capture reply is still `queued`; what to keep, and for how long, is set by [artifact retention](../../project-layout/references/layout.md), and which captures count as replaced by [Retain evidence](../../project-layout/references/layout.md#retain-evidence).

## Manage a work

Renaming, archiving, restoring, pinning, unpinning and deleting run in the authenticated viewer or through `hive view work rename|archive|restore|pin|unpin|delete`, which act as `local_viewer` and take no identity flags. Run one only when the user asks for that action on a named work, once per request. `delete` removes the work and its evidence: ask it as a critical effect every time, even when the user named the work (the one exception to the global guidance's «unless the user asks for it on a named target»), and pass `--confirm` with the same work ID. Never use `local_viewer` in a publication command, and never expose viewer credentials or access fragments in evidence.
