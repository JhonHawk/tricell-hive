# Skills and agents

Apply these rules together with the project's own authoring conventions (its guidance files and documented frontmatter contracts). A project convention that is stricter than a rule here wins; report a conflict instead of choosing silently. Use outcome `fix` for concrete corrections.

## Skills

- **HA-SK-01 — Name and description.** `name` is kebab-case, at most 64 characters, and matches the skill directory. `description` states what the skill does and when to use it, at most 1024 characters, without angle brackets. A passive summary without a use condition is a triggering risk.
- **HA-SK-02 — Procedure and references.** `SKILL.md` holds the essential procedure and stays under 500 lines. Each reference is linked with the condition that requires reading it; references are not preloaded for every task.
- **HA-SK-03 — Focus.** Prefer a focused skill with a few modules over a comprehensive documentation dump. Flag content that restates general knowledge the model already has.
- **HA-SK-04 — Portable frontmatter.** A skill meant for several hosts uses only portable fields (`name`, `description`, and, where the project allows, `license`, `compatibility`, `metadata`, `allowed-tools`). Host-only fields break packaging or are ignored elsewhere.
- **HA-SK-05 — Placement.** Knowledge needed in most sessions belongs in the always-loaded index, because skills are pulled on demand and may not trigger. Skills carry situational procedure.
- **HA-SK-06 — Resources resolve.** Every linked resource resolves relative to the skill directory. No personal installation paths or machine-specific absolute paths.
- **HA-SK-07 — Consequential effects.** A skill whose procedure takes an effect that needs explicit authorization (any commit, push, merge, publication, deployment, deletion, or global-configuration change) states that authorization before those steps.

## Agents

- **HA-AG-01 — Structural restriction.** Restrict tools structurally where the host supports it (tool lists, read-only sandbox or permission mode) and also state the boundary in the agent's instructions. Access settings are not universal sandboxes: parent permission modes and shell access can override them.
- **HA-AG-02 — Narrow purpose.** Each agent has one purpose, and its description says when to use it. Flag broad mandates that overlap another agent without a distinct deliverable or boundary.
- **HA-AG-03 — Valid frontmatter.** Claude Code silently skips an agent with a missing `description`, an invalid `name`, or malformed YAML. Check the host's required fields and the project's own frontmatter contract.
- **HA-AG-04 — Return contract.** The agent states what it returns: findings or changes, evidence, limits, and decisions needed from the parent.
- **HA-AG-05 — Resolvable resources.** An agent that depends on skill resources names them in a form that resolves from where the host installs the agent, such as the project's documented skill locator, never by a path relative to the agent's source file.

## Validators

Run the static validators in [native tools](native-tools.md) before judging format by hand, and report their output as evidence. Validators check shape, not quality: an agent or skill that passes them can still fail these rules.
