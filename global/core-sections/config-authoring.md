---
order: 165
targets: [claude]
---

<!-- trigger: any session (Flag contradictions — two sources disagreeing, no file to announce it); the authoring bullets are scoped to the write and live in rules-situational/config-authoring.md -->

### Config & Rule Authoring
- **Writing a CLAUDE.md/AGENTS.md, rule file, skill, agent prompt or hook message is concise-first** — one directive per rule, no justification or provenance; never compress away scope, the trigger, or precedence over another loaded source. Full text: `config-authoring.md`, held by `rule-delivery` on the write and carried by the `workspace-conventions` skill (a hook message has no file to hold: read it there).
- **Flag contradictions.** A codebase pattern that contradicts a global rule, or two authoritative sources that disagree about the same decision: surface it rather than silently following either — `gap-resolution.md > Divergence Between Sources`, via the `task-routing` skill.
