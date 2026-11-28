# Stored data

Read before a change to stored data. Read it in any flow.

- Before choosing a technique, take a read-only count of the affected records in each environment the change reaches, and record it. Scale the technique to that count: up to 100 records need one targeted update in an authorized window; a count of zero needs no migration.
- Migrate or remove the stored values the new code stops reading, as a task of the same change.
- Before running the change, define its cutover checks, the conditions that abort it and its recovery path, and verify that the backup or restore it relies on exists. Rolling back code or schema does not restore transformed or deleted data; when reversal cannot recover it, choose a validated restore or a fix-forward.
