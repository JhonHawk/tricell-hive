# Harness engineering research

This folder holds the research and measurement basis for rebuilding Hive as a small, portable guidance layer for Claude Code, Codex, Grok Build, Pi, and OpenCode V2. It is not a promise that the five hosts load instructions identically, nor an implementation plan.

## Start here

- [Cross-harness findings and design implications](2026-09-20-portable-harness-research.md) — source-backed findings, limits, recommendations, and a bounded evaluation approach.

## Supplied corpus

The synthesis includes the five Markdown texts and inspected figures from `uber-software-factory.zip`, alongside external harness-engineering and official host documentation. The original archive remains at `/path/to/home/Downloads/uber-software-factory.zip`; its SHA-256 is `1f53cb048d8dd298d7cea81fca73933902c8093f8aa4ae03ad9a9f6c82c260bb` (rechecked on 2026-09-20). Filenames, figure interpretations, source limitations, and attribution caveats are recorded in the synthesis. The archive is a local source, not a tracked dependency.

This branch retains the harness-engineering research. Earlier implementation designs, agent packs, audits, pilots, and retirement logs remain in Git history rather than this research index.

## Evidence standard

The dated synthesis records documentation checked on 2026-09-20 and analysis of the supplied corpus. Most host documentation is not pinned to the local CLI versions; recheck the installed version before relying on a loading or permission detail. Documentation describes intended behavior; runtime traces establish only what happened in an observed run. External results do not demonstrate an improvement in Hive, and this research does not certify host state or prove rebuilt guidance works across all five CLIs.

Keep three questions separate: did the host discover/select the guidance, did it read the relevant source before acting, and did the resulting work meet the criterion? Record runtime, model resolution, terminal state, reads, writes, human intervention, and cost/time when measuring a change. Do not treat prompt-cache hits, bytes divided by four, or one successful pair as proof of reduced cost or improved reliability.
