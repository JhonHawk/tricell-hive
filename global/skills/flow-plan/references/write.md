# /flow-plan `write` — the executable plan (plan gate)

Crystallizes findings + spec into a plan another harness executes. The plan is the contract
**and** the execution state — write it to the standard in
`~/.claude/skills/flow-core/references/plan-format.md`.

1. **OPEN** per the flow contract (`~/.claude/skills/flow-core/SKILL.md`): read PROJECT.md, the
   `<slug>-findings.md` if `research` ran, and the target epic.
2. **Produce the plan per `flow-core/references/plan-format.md`**: the `Status` header, and one self-contained block
   per task — local ID `T<n>`, exact Files, Interfaces (consume/produce), atomic steps, the
   `in-vivo: yes/no` flag (you decide it: UI/integration → yes, pure logic → no — so the gate is
   never assumed or omitted downstream), a `Verify:` command paired with its expected output, and
   the `Commit: feat(<scope>): T<n> …` tag. Declare the integration semantics (branch, merge
   mechanics, which CI gates each PR) in the header. Routing each task to a specialist is
   `flow-build`'s job at execution time — the plan stays harness-neutral; the optional `Agent:`
   line annotates the routing row where one clearly applies (it is a visibility aid, not a
   lock — `flow-core/references/plan-format.md` owns the rule, including the `in-vivo: yes` →
   `in-vivo-qa-tester` convention and the substitution-reporting duty). When any decision
   blocks task detail, add a **Decisions to close BEFORE executing** table above the tasks
   (`flow-core/references/plan-format.md`): technical rows you confirm with a peer/tool, stakeholder rows folded into
   the approval gate at step 6 — so execution never drips questions mid-task. **This gate is the
   technical gate** (counterpart of `/flow-specs review`, the business gate): the epic's parked
   technical questions must be closed — in its TECH.md, or as rows in the Decisions table — before
   the plan is ready for approval; an open parked question is a plan defect, never something
   execution absorbs silently.
3. **Preflight — resources confirmed at plan time, not at point of use.** The plan's approval is
   the last interruption; a missing credential found mid-execution kills the autonomy. Derive
   from the WHOLE flow (implementation, the in-vivo gate, and what promoting to qa/prod will
   need) and resolve NOW. Check presence, never print values:

   | Resource class | Verify |
   |---|---|
   | Env/config | `.env.<env>` and config files the tasks read exist |
   | CLI auth | `gh auth status`, `aws sts get-caller-identity` / `hcloud` for the accounts touched |
   | Domain tools | CLIs beyond gh/aws/hcloud (tunnels, webhook simulators, provider CLIs) — `which`/`--version`; missing → ask before installing |
   | Services | DB/Redis/queues reachable (or note how they start) |
   | Integrations | tracker access works; external sandbox tokens present — including the in-vivo gate's credentials |
   | Test baseline | the touched suite runs before T1 (affected subset per `testing.md`); a red baseline is a Preflight decision for the user, never absorbed silently |

   What's checkable gets reported `ok`/`missing`; what needs the user goes in a **Preflight
   section** of the plan as explicit asks. The plan is not ready for approval while a known-needed
   resource is unresolved.
4. **Large scope → split into an initiative.** When the scope is too dense for one plan, propose
   splitting into numbered parts and create the **initiative** folder
   (`flow-core/references/specs-structure.md > Session & initiative conventions`): `plan/` with a
   master plan + parts `00-NN`, each part its own `Status`.
   One-off scope stays a single `<slug>-plan.md`.
5. **Write the plan to the session/initiative home** (detection rule in `project-structure.md`):
   the specs repo if present, else `<repo>/_support/sessions/`. Add an `Implements:` line (epic/
   task refs) and an `in-progress` row to the sessions index.
6. **Gate the plan for approval**, by harness mode (`flow-core/references/harness-mechanics.md`):
   - **Native plan mode active**: the plan document is the presentation — finalize and exit via
     ExitPlanMode; do not duplicate it as a summary.
   - **Any other mode**: render the plan IN the conversation FIRST — every task block, the
     Preflight results (`ok`/`missing`), the Decisions table, the integration semantics — then
     gate with the structured-question mechanic. The file path plus a summary, or a synopsis
     embedded in the question text, is NOT a presentation: the user approves what they can
     read on screen, never a plan they would have to open a file to see.
7. **CLOSE**: update the ledger handoff (plan path, next phase). **Offer the next step
   explicitly** — the exact `/flow-build` invocation and the plan path, phrased as an offer to
   run it now, plus the model/harness recommendation (capable model for review; a cheaper harness
   MAY run the build) framed as a cost choice, not a requirement. Unresolved Preflight asks the
   user must clear are the only user to-dos — list them with why they are the user's.
