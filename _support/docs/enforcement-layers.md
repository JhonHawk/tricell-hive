# Enforcement Layers

How a rule, gate, or constraint in this configuration system is actually enforced. Three
layers exist, with very different strength. Confusing them is the #1 cartographic error when
writing or reasoning about rules: a convention phrased as impossibility reads as safe when it
is not, and a gate described loosely invites an agent to argue past it.

## The three layers

| Layer | Who enforces | Mechanism | If the model ignores it |
|---|---|---|---|
| **Deterministic** | The harness | Hooks, deny permissions, tool allowlists, `InputValidationError` on unloaded deferred tools, `permissionMode: plan` | The action is intercepted — obedience is irrelevant |
| **Confirm-gated** | The user | The rule requires explicit user confirmation before the action; the harness may or may not back it | The model can technically proceed, but doing so violates the gate; the user's yes IS the key |
| **Prompt-convention** | The model itself | The rule is loaded into context; the model follows it because it read it | Nothing intercepts — discipline, review, and audits catch it after the fact |

## Representative classification

- **Deterministic:** hook scripts (`global/hooks/`), permission denies in settings, agent
  frontmatter `tools:` allowlists, `disable-model-invocation` (the harness omits the skill
  from model-triggerable context), deferred-tool schema sealing via `ToolSearch`.
- **Confirm-gated:** production deploys and protected-branch merges, force-push and history
  rewrites, destructive operations (`CLAUDE.md > Destructive Operations`), writing a real
  secret to a file (allowed only on explicit user request after a risk confirmation —
  `security.md`), `/deploy-global`, CRITICAL/HIGH supply-chain installs with no safe path.
- **Prompt-convention:** most `alwaysApply` rules ("never mix package managers", commit
  style, memory routing, delegation gates), agent-prompt constraints ("never mutate") on
  agents whose tool contract does not mechanically prevent it.

Two honest subtleties:

- **A tool allowlist is deterministic but not airtight.** An agent without `Write` but with
  unrestricted `Bash` can still mutate through the shell — that is why read-only review
  agents carry `permissionMode: plan` in addition to the allowlist, and why
  `finding-refuter` (Bash without plan mode, constrained by prompt) is a documented
  exception, not the norm (`AGENTS.md > Agent Design Principles`, "Restricted tools" bullet).
- **Loading is deterministic; obeying is not.** Path-scoped rules are glob-loaded by the
  harness (deterministic), but the loaded content is still prompt-convention.

## The authoring rule that follows

A rule that states a gate names its layer. Never phrase a prompt-convention as mechanical
impossibility ("cannot", "physically blocked") — if a gate must be unbreakable, that is a
request for a deterministic backstop (hook, deny permission), not for stronger wording.
Operationalized in `global/CLAUDE.md > Communication > Name the enforcement layer` and in
`/manage-rules validate` (criterion 8, "Enforcement honesty").

## Related

- `global/skills/flow-core/references/harness-mechanics.md > Enforcement differences to
  respect` — how the deterministic layer maps per harness (Claude Code allowlists, Codex
  `sandbox_mode`, opencode `permission` denies).

---

*Origin: adversarial research session 2026-07-17 (two independent proposals cross-examined
by `finding-refuter` against the rule corpus). Full artifact:
`_support/archive/docs/2026-07-17-sistema-magia-isekai.html`.*
