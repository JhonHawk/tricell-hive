# Retain navigation grouping after a worktree is removed

Status: Implemented; verified candidate awaiting interactive acceptance.

## Problem and scope

The closed Hive0.9.0 work retained the original release worktree as provenance. Cleanup removed that directory, so grouped navigation emitted an unavailable diagnostic and split it from the main checkout despite identical stored project keys. Change only read-only presentation resolution; preserve storage, source metadata, membership and authority.

## Accepted behavior

For an absent registered source, use an unambiguous valid explicit name witnessed by another available registered source in the index with the exact persisted project key. Equal display names alone never establish identity. No witness, conflicting names, invalid active config and unregistered retargeted sources retain their fallbacks.

## Acceptance criteria

- AC15 — A removed registered worktree shares the explicit group of an available registered source with the same stored project key.
- AC16 — Missing or conflicting witnesses keep diagnostics; active invalid and unregistered retargeted sources do not gain fallback authority.
- AC17 — Group resolution preserves stored works, documents, identities and original source metadata.

## Delivery

User chose to fix the diagnosed defect. Existing session decisions D2 A interactive and D3 A native Codex review apply. No release publication requested for this fix. Existing maintainer marker covers binary refresh and viewer restart after integrated runtime changes. Restart preserves port, state and TTL; previous viewer tabs need a new access token.
