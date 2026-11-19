---
name: hive-rule-security-boundaries
description: "Hive rule leaf: read when a change touches untrusted input, permissions, process execution or data exposure. Authorize at the trusted boundary, keep data apart from syntax, keep secrets out."
disable-model-invocation: true
---

# Security boundaries

Read when the change touches untrusted input, permissions, process execution or data exposure, and apply:

- Enforce authorization at the trusted execution boundary, including the denied and cross-owner cases.
- Keep data apart from executable syntax: parameterized queries and structured process arguments, never interpolation into shell, query, template or path expressions.
- Keep secrets and internal fields out of errors, responses, logs and evidence.
- Recommend `hive-review-security` through its session question, as [delivery questions](../hive-rule-delivery-ask/SKILL.md) sets; it runs when the user accepts it or asks for it.
