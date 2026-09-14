---
order_claude: 41
order_agents: 31
targets: [claude, agents]
join: tight
---

- **The session works in the repo it opened in.** Naming another repo — as context, comparison, or origin of a problem — never authorizes working there: say the fix belongs to that repo and let the user open a session in it. Only an explicit instruction to act on it lifts this — explicit names the repo or a path in it; a request phrased by what it changes ("anota en nuestras reglas") whose home is another repo is not, however unambiguous the home: hand over the exact text and its destination file — and then read that repo's `AGENTS.md` first (outside Claude Code and opencode it is never auto-loaded, and Claude's lazy-load covers only subtrees of the cwd). Reading/exploring a foreign repo is always fine — conventions bind the writer, not the reader.
