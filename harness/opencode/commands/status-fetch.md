---
description: Fetch live project state (git, declared tracker, PRs, deploys) as compact facts, gathered in an isolated subagent.
---
Execute the skill `status-fetch` now, with these arguments: $ARGUMENTS

1. Read `~/.agents/skills/status-fetch/SKILL.md` and follow its phases exactly.
2. **The isolation is not automatic here.** `context: fork` is Claude Code-only and is
   stripped from this tree, so dispatch the work yourself: send the skill's phases to the
   `state-fetcher` subagent via the task tool and report only what it returns. Running the
   sweep inline defeats the skill's entire purpose — the tracker pages and git output would
   land in this session.
3. Tracker calls go through opencode's MCP clients directly. The authorization rule is
   unchanged: only a tracker declared in the ledger may be queried.
