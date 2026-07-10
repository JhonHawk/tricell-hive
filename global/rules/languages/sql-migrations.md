---
paths:
  - "**/*.sql"
  - "**/schema.prisma"
  - "**/migrations/**"
  - "drizzle.config.*"
---

## SQL & Migrations

### Migration Safety
- **Prefer forward-only migrations with a backward-compatible transition window (expand-contract)** over paired down-scripts: add the new shape, dual-write/backfill, migrate readers, then drop the old shape in a later release — never a single in-place breaking change. Provide a down/rollback only where it's genuinely cheap; document why it's irreversible (data backfill, enum removal).
- **Idempotent when possible.** Use `IF NOT EXISTS`, `IF EXISTS`, `OR REPLACE` to make reruns safe.
- **Never `DROP COLUMN` without verifying data impact.** Check if the column has non-null values, foreign key dependencies, or application reads. Ask the user before destructive schema changes.
- **Adding NOT NULL columns** to existing tables requires a `DEFAULT` value or a multi-step migration (add nullable → backfill → set NOT NULL).
- **Index foreign key columns.** Postgres does *not* auto-index them (MySQL/InnoDB does) — every `REFERENCES` on Postgres should have a corresponding index unless the table is trivially small.
- **Timestamp-prefixed naming.** Migration files: `YYYYMMDDHHMMSS_description.sql` or framework-generated equivalents.

### Prisma
- **PascalCase models (singular), camelCase fields.** Map to DB naming with `@map`/`@@map`: `model User { ... @@map("users") }`.
- **Explicit relations.** Define both relation fields (the list side and the FK side); `@relation(fields: [...], references: [...])` goes ONLY on the side holding the foreign key — never on both.
- **`@default(now())`** for `createdAt`. `@updatedAt` for `updatedAt` — Prisma Client sets it on every write (application-level, NOT a DB trigger); raw-SQL writes or other clients bypassing the Client won't update it.
- **Run `npx prisma format`** after schema changes. Run `npx prisma generate` after migration to sync the client.

### Drizzle
- **`drizzle-kit generate`** for codebase-first migrations from TypeScript schema. Never hand-write migrations that drift from the schema source.
- **Schema co-located with feature modules.** Export from a central `schema.ts` that re-exports feature schemas.
- **`drizzle-kit push`** — reserved for development; production uses `drizzle-kit migrate` with generated SQL files for an auditable, version-controlled history. (Drizzle's docs do endorse `push` in production for blue/green + serverless setups; our convention is stricter for auditability.)
