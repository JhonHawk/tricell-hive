---
name: hive-principle-migrate-callers-then-delete-legacy-apis
description: "Hive principle leaf: read when adding an internal API while old callers remain. Move the callers and delete the old API in the same change; keep a layer only under a compatibility plan."
disable-model-invocation: true
---

# Migrate callers, then delete the legacy API

Read when a change adds a new internal API that replaces an old one and callers of the old one still exist.

1. Find every caller of the old API, including tests, scripts and generated code, and move them to the new one in the same change.
2. Delete the old API and the tests that only pinned its internals. Update the tests that stay to assert the new contract.
3. Do not leave a compatibility wrapper, alias or fallback for callers inside the repository: it keeps two ways to do one thing.

This applies when every caller is in the repository and the change can land as one coordinated edit. It does not apply to a public or persisted contract, such as a published API, a file format, a stored value or a schema other systems read: that needs a compatibility plan the user authorized, with its own migration and removal date. When you cannot tell who calls the old API, treat it as such a contract and ask.
