# Enforcement Layers

How a rule, gate, or constraint in this configuration system is actually enforced. Three
layers exist, with very different strength. Confusing them is the #1 cartographic error when
writing or reasoning about rules: a convention phrased as impossibility reads as safe when it
is not, and a gate described loosely invites an agent to argue past it.

## The three layers

| Layer | Who enforces | Mechanism | If the model ignores it |
|---|---|---|---|
| **Deterministic** | The harness | Hooks, deny permissions, tool allowlists, `InputValidationError` on unloaded deferred tools | The action is intercepted — obedience is irrelevant |
| **Confirm-gated** | The user | The rule requires explicit user confirmation before the action; the harness may or may not back it | The model can technically proceed, but doing so violates the gate; the user's yes IS the key |
| **Prompt-convention** | The model itself | The rule is loaded into context; the model follows it because it read it | Nothing intercepts — discipline, review, and audits catch it after the fact |

## Representative classification

- **Deterministic:** hook scripts (`global/hooks/`), permission denies in settings, agent
  frontmatter `tools:` allowlists, `disable-model-invocation` (the harness omits the skill
  from model-triggerable context), deferred-tool schema sealing via `ToolSearch`; since F4
  also the **protected-branch deny** (`bash-policy.sh` (d): `git commit` on
  `production`/`qa`, and `git push` targeting them or bare-pushed from them — the
  deterministic backstop of the promotion confirm-gate; `master`/`main` stay out: trunk-
  direct is a declared workflow, prompt-convention by design) and the **reviewer guard**
  (`reviewer-guard.sh`, agent-scoped via frontmatter `hooks:` — see below).
- **Confirm-gated:** production deploys and protected-branch merges, force-push and history
  rewrites, destructive operations (`CLAUDE.md > Destructive Operations`), writing a real
  secret to a file (allowed only on explicit user request after a risk confirmation —
  `security.md`), `/deploy-global`, CRITICAL/HIGH supply-chain installs with no safe path.
- **Prompt-convention:** most `alwaysApply` rules ("never mix package managers", commit
  style, memory routing, delegation gates), agent-prompt constraints ("never mutate") on
  agents whose tool contract does not mechanically prevent it.

Two honest subtleties:

- **A tool allowlist is deterministic but not airtight.** An agent without `Write` but with
  unrestricted `Bash` can still mutate through the shell. Review agents carry
  `permissionMode: plan` against this, but **it is not a layer to count on**: when the parent
  session runs in auto mode — the default on Pro/Max/Team unless `permissions.defaultMode`
  says otherwise — a subagent inherits auto mode and its frontmatter `permissionMode` is
  ignored. Verified empirically 2026-08-15: a `code-reviewer` subagent had no `ExitPlanMode`
  and created a file via `touch` unblocked. **Since F4 the agent-scoped `reviewer-guard`
  hook closes the common Bash-mutation paths** (git mutations, deleters, in-place edits,
  installs, non-temp writes) for the strict roster — `code-reviewer`, `security-reviewer`,
  `product-critic`, `spec-quality-reviewer`, `code-scout`, `workspace-custodian` — via a
  frontmatter `hooks:` PreToolUse block on `Bash`. **Exempt by doctrine
  (execute-to-observe):** `finding-refuter` (runs tests/repro commands), `in-vivo-qa-tester`
  and `ui-reviewer` (drive a browser), `state-fetcher` (executes approved tracker writes) —
  for them the deterministic layer remains the `tools:` allowlist, with never-mutate as
  prompt-convention. Residual gap even under the guard: an interpreter one-liner can still
  write — the hook is a guardrail, not a sandbox. Harness scope: `hooks:` is Claude-only
  frontmatter (`convert-agents.py` strips it; `codex exec` runs no hooks), so on
  Codex/opencode/Grok the whole review-agent read-only doctrine stays allowlist +
  prompt-convention.
- **Loading is deterministic; obeying is not.** Path-scoped rules are glob-loaded by the
  harness (deterministic), but the loaded content is still prompt-convention.

## The authoring rule that follows

A rule that states a gate names its layer. Never phrase a prompt-convention as mechanical
impossibility ("cannot", "physically blocked") — if a gate must be unbreakable, that is a
request for a deterministic backstop (hook, deny permission), not for stronger wording.
Operationalized in `global/CLAUDE.md > Communication > Name the enforcement layer` and in
`/manage-rules validate` (criterion 8, "Enforcement honesty").

## Portable planning

- The shared plan validator deterministically checks structure, contract digest and declared
  action/target scope when invoked. It is read-only and does not launch or intercept tools.
- Plan approval is confirm-gated. Recorded conversational evidence and a matching digest
  are not a cryptographic user signature.
- Invoking the validator, checking live evidence and honoring its result are prompt
  conventions. Progress and delivery records are claims to reconcile, not evidence by themselves.
- Draft planning has no ambient parent-write blockade. General shell, MCP and reviewer
  guards keep their own coverage; native plan-mode transitions grant no Hive authority.
- Recovery structure and budget checks are deterministic when invoked, over recorded entries.
  Recording attempts, preserving history and reconciling real outcomes are prompt-conventions;
  an editable plan is not an immutable history or an idempotency mechanism.
- The verification matrix is prompt-convention. Independent execution and review supply
  evidence; neither a selected profile nor a green check grants publication authority.
- PI research readiness deterministically compares required names with active tools in the
  current agent. Invoking it is prompt-convention; it does not check provider credentials,
  network reachability or evidence quality.

## Related

- `global/skills/flow-core/references/harness-mechanics.md > Enforcement differences to
  respect` — how the deterministic layer maps per harness (Claude Code allowlists, Codex
  `sandbox_mode`, opencode `permission` denies).

---

*Origin: adversarial research session 2026-07-17 (two independent proposals cross-examined
by `finding-refuter` against the rule corpus). Full artifact:
`_support/archive/docs/2026-07-17-sistema-magia-isekai.html`.*
