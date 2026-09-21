# Security boundaries and dependencies

Read when changing untrusted inputs, permissions, process execution, public data exposure, or dependencies. Apply the affected boundaries using existing project controls; this is not a mandate for an unrelated security audit.

- Identify the actor, resource, operation, and tenant or ownership boundary. Enforce authorization at the trusted execution boundary, including denied and cross-owner cases where applicable; authentication alone does not establish permission.
- Validate external data at runtime. Keep data separate from executable syntax through parameterized queries, structured process arguments, and context-appropriate output encoding. Inspect downstream interpretation even when the first API accepts an argument array; avoid interpolating untrusted values into shell, query, template, or path expressions.
- Preserve useful internal diagnostics while keeping secrets and unintended internal fields out of public errors, responses, logs, and retained evidence. Check the affected failure path as well as success.
- Before introducing or updating a dependency, check its source, resolved version, installation/build scripts, and relevant runtime or build exposure. Use existing lockfile, provenance and advisory controls; consult current authoritative advisories when version-specific risk matters. A clean advisory scan is not evidence of trusted origin or absence of unknown vulnerabilities. Review automated remediation before running it: it may execute project-controlled scripts or use configured registries.

Select checks for the actual trust boundary and reachable failure mode. Report findings with preconditions, impact, and evidence rather than treating a severity label or tool exit as the conclusion. This procedure does not authorize exploitation, credential rotation, or unrelated remediation.
