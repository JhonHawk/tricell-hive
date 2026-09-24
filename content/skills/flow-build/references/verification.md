# Verification by risk and surface

Read when planning or executing integration, runtime, UI, performance, hardware, test replacement/removal, promotion, deployment, or costly checks. The build skill owns the essential test-first and completion procedure; this reference selects depth and timing. Follow the project's established commands, environments, and delivery policy. Verification does not grant deployment, publication, destructive effects, or external access.

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

## Gate selection and timing

Apply the shared authorization rule before substantial gates. Use existing project conventions to discover required versus offered gates, preferred mechanisms, and where availability can be checked. Keep capability, preference, execution authorization, and technical verdict distinct. Do not launch a review to test availability or infer approval from repeated historical choices. Present the concrete scope/candidate, mechanism, environment, relevant cost or effects, and expected evidence when a decision is missing. A plan may carry explicit advance authorization; its mere inclusion of a check is insufficient.

| Boundary | Required decision or evidence |
| --- | --- |
| Research / task entry | Inspect state, risks, project commands and automatic triggers; propose experiments exceeding ordinary authorized investigation before running them. |
| Plan / implementation brief | Select applicable checks, prerequisites and pass criteria; place substantial gates and resolve or record execution authorization. Bounded read-only feedback on a saved implementation plan follows the [plan-review procedure](../../flow-plan/references/plan-review.md) as part of planning; broader review services remain subject to gate authorization. |
| Build iteration | Run targeted tests and routine checks; inspect the implementation's own diff. This is not a dedicated independent code-review gate. |
| Local candidate | Complete applicable local in-vivo and UI verification, using the built artifact for final UI/runtime validation where the stack has a build step; a rendered UI effect adds the required independent UI review and in-vivo children. Propose dedicated code review separately unless already authorized. |
| Publication / integration | Review the identified candidate with the selected mechanism and inspect actual CI coverage. Account for automatic review/deployment effects before the push or PR event that triggers them. |
| Predeploy | Reconcile the complete promoted candidate, affected destinations, migrations, compatibility, recovery and reusable prior evidence. A promotion label does not prove the candidate was reviewed. |
| Postdeploy | Verify active artifact/revision and routine health per affected surface. Offer the functional in-vivo/UI gate in that environment unless its concrete execution is already authorized. |
| Close | Reconcile evidence and remaining gates. Separate published, healthy and functionally accepted; keep pending, failed and blocked checks visible. |

Large integration suites, load/soak, hardware exercises and external-service scenarios need scope and effect assessment, regardless of whether their command starts locally. Their phase follows the required environment and artifact, not a universal final-stage checklist. For authorized re-review, focus on affected changes and invalidate previous evidence when relevant inputs change; do not assume unlimited rounds or a broader mechanism are covered.

When the selected mechanism is `review-code`, or any review that does not verify its own findings, send each P0 or P1 finding, and any finding that would block the merge, to a `review-refuter` child before acting on it. Fix confirmed and plausible findings under the normal rules, drop refuted ones with the refuter's evidence, and report both. Native reviews that already verify each finding, such as Claude Code's `/code-review` at medium effort or above, need no second pass.

A technical verdict does not authorize delivery. A required gate declined or unavailable remains an unmet transition condition; an optional gate may be deferred explicitly. Interpret actual findings and terminal status rather than treating absent comments, a queued run, or a non-failing check as approval.

## Runtime and UI gates

Decide local in-vivo applicability from the actual changed behavior. When needed, include and execute the relevant walk within authorized implementation without asking whether to run it. When judged unnecessary, explain why and ask whether the user wants it anyway, offering Skip (Recommended) and Run with the concrete proposed scope. Reuse an explicit answer for the same scope; do not repeat this question per task or check. While unanswered, continue independent work and record the optional walk as pending, not declined or passed. This does not waive a required gate, override paused model pilots, or grant external effects. Deployed-environment walks retain the explicit authorization boundary above.

For local implementation, always exercise the applicable changed path and relevant failure or permission states. Perform the corresponding deployed-environment exercise only within its explicit gate authorization. A mocked integration or source inspection does not establish behavior in the running system. Resolve authorized local setup, accounts, roles and seed data; access only the user can supply remains a reported prerequisite.

Use dev/HMR for iteration. For the final in-vivo or UI handoff, build and serve the compiled candidate using the project's production-style start or preview command where the stack distinguishes build from dev. Services without a build step use their normal runtime. Build only the affected applications and required dependencies, not every app per edit.

Default the handoff stack to local frontend, backend, database and necessary dependencies unless a different target was explicitly specified. Check effective connections, including build-time frontend URLs and backend persistence targets, without exposing credentials; localhost in the browser does not prove the whole path is local. Prepare authorized local accounts, roles and seed data before the walk. Report missing dependencies and any uncovered integration; never silently substitute a remote environment. External authentication, messaging, payments and similar effects need their own authorized boundary.

For UI changes, local rendered and interactive inspection is part of completion, not an optional gate: layout, content, feedback, error recovery, responsive behavior, keyboard/focus and accessibility relevant to the change. Capture the prior state when comparing an existing interface and practical before editing; missing baseline limits claims of visual improvement, not unrelated functional checks. Automated accessibility or screenshot checks support, but do not replace, inspection of the affected experience.

Compare changed UI with the selected application/area’s accepted shell and screen pattern, including justified differences. Distinguish pattern conformity, functional behavior, and user suitability. A screenshot diff against a prior baseline does not establish that a new screen follows a different reference screen’s pattern. When shared UI code changes, inspect affected consumers across applications according to impact.

Any change with a rendered UI effect requires two independent children on the built local candidate before completion is reported: a `review-ux` review and an `sdd-verify` in-vivo check of the changed journeys. They are part of authorized local implementation and need no separate approval. The implementer's own walk remains required but satisfies neither. Brief the reviewer with [UI review criteria](ui-review-criteria.md), the affected screens, the base revision for comparison, realistic and extreme data, and the viewports and themes; do not pre-rate findings or say what to ignore. Resolve its findings before completion: fix introduced or worsened Blocker and High findings and re-review, and dispose of pre-existing ones under the shared incidental-findings rule. Run both in parallel with isolated data, accounts and browser sessions; otherwise serialize state-changing work. If a child, the browser or the stack is unavailable, report the gate as not passed. Backend or infrastructure work without a UI effect needs neither; add other specialists by actual risk rather than by default.

Before driving a browser, read [browser automation](browser-automation.md) and pass it to children that drive one.

## Human UI handoff

During the existing UI walk, capture a small useful selection of changed screens and relevant states; do not repeat the walk solely to produce images. The assigned reviewer returns screenshots, observed state and reproduction steps to the orchestrator without editing the application. Grant a specific temporary evidence destination when delegating; an observe role does not imply general filesystem write permission. If the host cannot capture there, return observations and have the orchestrator capture within its available tools. Avoid credentials and sensitive user data in images.

Use the global UI-image placement and hygiene rule: put review captures inside the established session folder and link or display them in the handoff. Preserve them while human review is pending; hygiene removes the task's review images except those the user selected to keep. Screenshots do not replace interactive checks.

For UI work awaiting human validation, leave the same verified built stack running, confirm its URLs are reachable, and name the candidate and relevant local targets. Provide clickable application URLs and the required account/role without secrets. Write the walkthrough as one block per flow, headed by the ticket and a short description (`ARK-123 - <description>`), followed by bullet steps from a named starting point: preconditions such as the role and signed-in state in parentheses on the first step, navigation with `>` between screens, the action, and what the user should see. Condense obvious steps. When several tickets share one flow or module, cite the ticket in the step it concerns; separate flows get separate blocks. Link the selected captures, state what the agent verified and what remains for the human, and explain how to stop task-owned processes. Do not claim availability beyond the host's process lifetime; disclose inability to keep the stack alive and give the exact restart instructions.

## Frequency and coverage

- During implementation, run the small checks that give feedback on the current behavior. Inspect transitive impact and use existing test-selection tooling where reliable; file proximity alone may miss affected callers.
- At a coherent increment's close, verify the final affected scope and applicable build/integration/runtime/UI gates. Avoid duplicating every edge case at every layer.
- At integration or promotion, satisfy the project's broader regression and release gates. Confirm what CI actually executes; do not assume it covers a missing check. Local checkpoints, branch publication, merge and release are distinct boundaries, governed by the project's policy and existing authorization.
- After authorized deployment, identify the artifact or revision actually served in each affected target and check routine health: health endpoints and the project's smoke checks. An HTTP smoke proves liveness, not functional behavior. Then send the functional walk to an `sdd-verify` child against the environment's URLs, plus `review-ux` when a UI changed, each with its own account or session; the orchestrator does not script its own walk. Run it without asking when the project declares so or the user asked to validate the deployment; otherwise offer it and record acceptance as pending. Recovery exercises stay within their authorized scope. Use appropriate readbacks, logs, metrics, traces, or journeys; a successful pipeline or an available shell alone does not establish service health. Check recovery against the data/configuration changes made, not only the previous code version.
- Reuse evidence only for unchanged relevant inputs. Code, configuration, dependency, environment or external-state changes can invalidate it. Performance trends, changing data/traffic and experimental noise can justify scheduled or repeated checks even without a code change.

Parallelize independent checks; isolate or serialize shared state. After a correction, rerun its failing check and affected scope. Broaden for cross-cutting changes or new failure evidence. Bound review/repair loops by material findings and explicit stopping conditions rather than arbitrary repetition for confidence.

## Capacity is separate from correctness

Before load testing, define the workload, dataset scale, arrival pattern, tenant distribution, environment and agreed latency/error/saturation objectives. Validate that thresholds measure meaningful behavior; a green threshold with a poor workload can mislead. Distinguish throughput from transactional integrity and tail latency. Run stress, soak or recovery scenarios when their failure modes matter, not on every commit. Record environment limits rather than extrapolating local results to production.

## Keep the evidence small

Record criterion, command or journey, relevant revision/environment, observed outcome and limitation in the existing plan or closing response. Retain selected evidence only when useful. An unavailable device, missing credential or exhausted time budget leaves its required check unverified; it does not reduce the acceptance bar. If a cheaper substitute observes the same failure mode, explain the substitution and its limits.
