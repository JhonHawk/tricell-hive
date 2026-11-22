# Security boundaries

Read when a change touches untrusted input, permissions, process execution, data exposure, authentication, secrets, cryptography or dependencies. Apply:

- Enforce authorization at the trusted execution boundary, including the denied and cross-owner cases.
- Keep data apart from executable syntax: parameterized queries and structured process arguments, never interpolation into shell, query, template or path expressions.
- Keep secrets and internal fields out of errors, responses, logs and evidence.
- Recommend `hive-review-security` through its session question, as [delivery questions](../git/delivery-ask.md) sets.
