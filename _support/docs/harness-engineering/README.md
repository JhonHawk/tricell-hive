# Harness engineering research

This folder holds the research and measurement basis for rebuilding Hive as a small, portable guidance layer for Claude Code, Codex, Grok Build, Pi, and OpenCode V2. It is not a promise that the five hosts load instructions identically, nor an implementation plan.

## Start here

- [Cross-harness findings and design implications](2026-09-20-portable-harness-research.md) — source-backed findings, limits, recommendations, and a bounded evaluation approach.
- Activity-skills research (historical evidence omitted from public history) — skill scope, progressive disclosure, and the earlier cross-harness comparison.
- Activity-skills pilot record (historical evidence omitted from public history) — native runs, deployment, cleanup, and per-criterion status. Its results did not qualify the candidate or demonstrate quality improvement.
- Installed Hive inventory (historical evidence omitted from public history) — ownership and preservation boundaries observed before the rebuild.
- Rebuild research checkpoint (historical evidence omitted from public history) — prior branch and scope notes; read its status as a dated snapshot.
- Hive retirement findings (historical evidence omitted from public history) — later, live-scoped retirement results; use this for current cleanup status rather than the earlier inventory.

## Evidence standard

The dated synthesis records live documentation checked on 2026-09-20 and observations from the retained pilot. Most host documentation is not pinned to the local CLI versions; recheck the installed version before relying on a loading or permission detail. Documentation describes intended behavior; runtime traces establish only what happened in the observed run. The pilot's exploratory model grading is not human validation, and its mixed or contaminated cells cannot support a general quality claim. The later retirement report records a scoped cleanup; it does not certify all host state or prove the rebuilt guidance works across all five CLIs.

Keep three questions separate: did the host discover/select the guidance, did it read the relevant source before acting, and did the resulting work meet the criterion? Record runtime, model resolution, terminal state, reads, writes, human intervention, and cost/time when measuring a change. Do not treat prompt-cache hits, bytes divided by four, or one successful pair as proof of reduced cost or improved reliability.
