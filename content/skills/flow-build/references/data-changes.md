# Persistent data changes

Read when changing stored schemas, data representations, or migration/backfill procedures. Inspect the actual engine, schema constraints, data volume, access patterns, and deployment topology before choosing a technique.

- Identify readers and writers, versions that may coexist, and the required downtime tolerance. Choose a compatible transition, phased expand/migrate/contract, or an authorized coordinated window according to those constraints; dual-write is not a default requirement.
- Estimate lock duration, transaction size, resource pressure, and impact on concurrent traffic using representative evidence. Consult the engine/version documentation for the actual operation; a small local dataset does not establish production duration or safety.
- Make backfills and transformations bounded and restartable where needed. Define progress, duplicate/retry handling, and integrity reconciliation against source data and constraints, including concurrent writes when applicable.
- Define cutover checks, abort conditions, and a recovery path before execution. Distinguish rollback of code/schema from restoration of transformed or deleted data; choose validated restore or fix-forward when reversal cannot recover it. Verify the required backup or recovery capability rather than assuming its existence.
- Retire the old representation only after affected consumers have moved and integrity checks pass. Record why temporary coexistence can end and what evidence remains unverified.

Use the project's migration mechanism and proportionate integrity, compatibility, and workload checks. Destructive operations and production execution remain within explicit authorization; planning a recovery procedure does not authorize running it.
