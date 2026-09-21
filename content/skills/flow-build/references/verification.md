# Verification by risk and surface

Read when planning or implementing changes involving integration, runtime, UI, performance, hardware, test replacement/removal, or costly checks. The build skill owns the essential test-first and completion procedure; this reference selects depth and timing. Follow the project's established commands, environments, and delivery policy. Verification does not grant deployment, publication, destructive effects, or external access.

## Select evidence for the change

Identify the failure being guarded against, affected users and dependencies, reversibility, and boundaries crossed: persistence, permissions, processes, devices, or concurrency. Classify the changed surface, not the entire repository. A text edit in a busy service does not automatically need load testing; a landing page's payment flow is not merely presentation.

Choose the lowest-cost layer that can detect the relevant failure. The following are candidates, not a checklist to run in full for every change:

| Changed surface | Useful evidence | Escalation trigger |
| --- | --- | --- |
| Informational landing or presentation | Build/checks when available; rendered layout, links and actions, keyboard access, relevant viewports and console errors | Forms, authentication, payments or data changes add their functional and security checks |
| Application or API | Behavior tests, changed contracts and persistence integration, a real affected journey; UI review when visible surfaces change | Roles, tenant isolation, migrations, retries or shared middleware widen coverage |
| Mobile | Host/component tests and affected emulator/device journey | Lifecycle, offline, permissions or sensors require corresponding states; release candidates and representative devices when release fidelity matters |
| Driver or firmware | Isolated logic tests, cross-build/static checks, appropriate kernel/RTOS integration or simulation | Physical timing, interrupts, energy, device interaction or recovery require target hardware evidence; simulation alone cannot establish it |
| Transactional or high-throughput path | Invariants, real persistence, targeted concurrency, duplicate/retry/partial-failure checks | Query, transaction, cache, pool, queue or retry changes can require representative load and recovery tests |
| Configuration or agent guidance | Schema, diff, consistency and readback checks available in the project | Runtime effects need corresponding runtime evidence; model pilots require explicit authorization when paused |

When replacing or removing tests, compare the old and new scenarios, fixtures, preconditions, and assertions; justify any retired coverage against the current contract rather than preserving obsolete tests or assuming similar assertions cover the same behavior.

## Runtime and UI gates

Exercise the actual changed path and relevant failure or permission states. A mocked integration or source inspection does not establish behavior in the running system. Resolve authorized local setup, accounts, roles and seed data; access only the user can supply remains a reported prerequisite.

Use dev/HMR for iteration. At the candidate boundary, validate the built artifact in a representative execution mode when dev and delivered behavior differ. Do not invent a build step for a service without one or rebuild every app for each edit.

For UI changes, inspect rendered appearance and interaction separately: layout, content, feedback, error recovery, responsive behavior, keyboard/focus and accessibility relevant to the change. Capture the prior state when comparing an existing interface and practical before editing; missing baseline limits claims of visual improvement, not unrelated functional checks. Automated accessibility or screenshot checks support, but do not replace, inspection of the affected experience.

Use an independent reviewer when consequence or scope warrants it, with concrete criteria and evidence. Reuse existing review roles where available; do not summon every specialist by default. Parallel functional and UI reviews need isolated data/accounts/sessions; otherwise serialize state-changing work.

## Frequency and coverage

- During implementation, run the small checks that give feedback on the current behavior. Inspect transitive impact and use existing test-selection tooling where reliable; file proximity alone may miss affected callers.
- At a coherent increment's close, verify the final affected scope and applicable build/integration/runtime/UI gates. Avoid duplicating every edge case at every layer.
- At integration or promotion, satisfy the project's broader regression and release gates. Confirm what CI actually executes; do not assume it covers a missing check. Local checkpoints, branch publication, merge and release are distinct boundaries, governed by the project's policy and existing authorization.
- After authorized deployment, identify the artifact or revision actually served in the target environment and verify the affected behavior and recovery criteria. Use appropriate readbacks, logs, metrics, traces, or journeys; a successful pipeline or an available shell alone does not establish service health. Check recovery against the data/configuration changes made, not only the previous code version.
- Reuse evidence only for unchanged relevant inputs. Code, configuration, dependency, environment or external-state changes can invalidate it. Performance trends, changing data/traffic and experimental noise can justify scheduled or repeated checks even without a code change.

Parallelize independent checks; isolate or serialize shared state. After a correction, rerun its failing check and affected scope. Broaden for cross-cutting changes or new failure evidence. Bound review/repair loops by material findings and explicit stopping conditions rather than arbitrary repetition for confidence.

## Capacity is separate from correctness

Before load testing, define the workload, dataset scale, arrival pattern, tenant distribution, environment and agreed latency/error/saturation objectives. Validate that thresholds measure meaningful behavior; a green threshold with a poor workload can mislead. Distinguish throughput from transactional integrity and tail latency. Run stress, soak or recovery scenarios when their failure modes matter, not on every commit. Record environment limits rather than extrapolating local results to production.

## Keep the evidence small

Record criterion, command or journey, relevant revision/environment, observed outcome and limitation in the existing plan or closing response. Retain selected evidence only when useful. An unavailable device, missing credential or exhausted time budget leaves its required check unverified; it does not reduce the acceptance bar. If a cheaper substitute observes the same failure mode, explain the substitution and its limits.
