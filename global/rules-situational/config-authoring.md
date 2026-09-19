---
globs:
  - "**/CLAUDE.md"
  - "**/CLAUDE.local.md"
  - "**/AGENTS.md"
  - "**/.claude/{agents,commands,rules}/**"
  - "**/skills/**/*.md"
  - "**/rules-situational/**"
  - "**/core-sections/**"
  - "**/global/agents/**"
  - "**/opencode/commands/**"
---

## Config & Rule Authoring

> How a directive is written into a config or rule surface — held by `rule-delivery` on a write to one (a CLAUDE.md/AGENTS.md, a rule file, an agent prompt, a skill's or command wrapper's prose — never the scripts, tests or fixtures beside them), and read through the `workspace-conventions` skill when the surface is named before any file exists. A hook message is the same authoring act with no file to announce it: apply this there too. The gate line stays in the core.

- **Concise-first — these files are not debates.** Adding or editing any line in a CLAUDE.md/AGENTS.md, rule file, agent prompt, or hook message: write the minimal actionable form on the first pass — one directive per rule, no justification, no provenance notes ("mirrors project X"), no evidence citations (benchmarks, measurements, "v3/v4", version history), no examples unless they disambiguate; evidence lives in the repo's bibliography/README, never in the directive. **Compress the sentence, not only the content:** name the subject once and at its shortest, fold repeated verbs into one ("never run or offer it"), drop qualifiers a default already carries. **Never compress away what makes it applicable** — scope (which environments, which files), the concrete trigger the agent would otherwise take, and precedence when it overrides another loaded source. Shorter than that is not concise, it is incomplete. **Carve-out:** a stat stays only when it IS the operational threshold the rule turns on — test: would removing the number change the decision? Expand only if asked.
- **Name the capability, not the tool.** In anything a second harness can read (`global/CLAUDE.md`, rules, skills, agent prompts): "the question tool", "the task-list tool", "the search tool". A proper tool name expires silently when the harness renames it and is inert where it does not exist — name the fallback when the capability may be absent. A proper name stays only where the behavior itself differs by harness, which IS the information.
- **Don't restate what the agent reads from the repo, nor teach what the floor model already does** — the package manager a lockfile declares, the framework its config declares, a directory listing: no-ops that also go stale. **Test:** delete the line and name what the agent would do differently; nothing → cut it. Calibrate to the floor model, never to yourself — a reminder the weakest model still needs (BSD vs GNU flags) earns its place.
- **Name the enforcement layer.** A rule that states a gate says what enforces it — deterministic (hook, deny permission, allowlist), confirm-gated (the user's explicit confirmation is the key), or prompt-convention. Never phrase a prompt-convention as mechanical impossibility; if a gate must be unbreakable, that's a request for a deterministic backstop, not stronger wording.
