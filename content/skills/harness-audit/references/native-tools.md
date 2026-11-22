# Native tools

Read before running a host tool, using its output, or asserting version-sensitive loading behavior. `<Claude home>` and `<Codex home>` are the hosts' configuration directories (by default `~/.claude` and `~/.codex`).

## Authorization (HA-ME-01)

- Run static tools first; they read files and at most write a local cache. An auditor that must stay read-only does not run tools that write a cache (`codex debug prompt-input`, `uv run`): the caller runs them and supplies the output, which you use instead of running the tool again; request a missing one in your result.
- Run a model-invoking evaluation only when the caller explicitly authorizes it, with a cost cap and no publication.
- Cover Claude Code and Codex by default; use tools for other hosts only on explicit request.

## Static tools

| Tool | Host | Checks | Notes |
| --- | --- | --- | --- |
| `(cd <entry> && codex debug prompt-input)` | Codex | The exact instruction chain Codex sends from that entry directory | No model call; writes only a model-list cache. Works outside git repositories. |
| `claude plugin validate <path> --strict --json` | Claude Code | Plugin or marketplace manifests and the skills, agents and commands they declare | Needs a plugin or marketplace layout; elsewhere it succeeds with empty `contents`, which is no evidence. |
| `claude plugin details <name>` | Claude Code | Component inventory and projected token cost of an installed plugin | — |
| `quick_validate.py <skill-directory>` from the `skill-creator` skill's `scripts/` | Claude Code and Codex copies | Frontmatter shape, name and description limits; the Codex copy also flags `[TODO:` placeholders | Needs PyYAML: check the dependency as the global guidance's «Scope and authorization» sets, then run `uv run --no-project --with pyyaml python <script> <skill-directory>`. The Codex copy lives under `<Codex home>/skills/.system/skill-creator/`. |
| `claude doctor`, `codex doctor --json` | Both | Installation, configuration and authentication health | Not a content audit. |

## Human-run views

- **Claude Code `/context`** lists the loaded instruction files under "Memory files" for the session's own directory; neither an agent nor its parent can run it. To confirm loading for an entry point, ask a human to start Claude Code there and run it, stating the hypothesis and the expected entry. A past run is recorded in that session's transcript under `<Claude home>/projects/<entry path with every non-alphanumeric character replaced by ->/`; extract the "Memory Files" table with `jq -r` rather than reading the whole transcript.
- **Claude Code `/memory`** opens the memory locations interactively.

## Model-invoking tools (HA-ME-01)

`claude plugin eval` (real sessions per eval case; use `--max-cost-usd`, `--no-publish`, `--threshold`), the `skill-creator` scripts `run_eval.py` and `run_loop.py`, and Claude Code `/doctor` trim proposals need explicit authorization. `codex review` and `codex exec` review diffs or run tasks and do not audit instruction files.

## Version-sensitive behavior (HA-ME-02)

Loading behavior changes between releases (whether an ancestor `CLAUDE.md` resolves its imports differed between Claude Code releases). Before asserting it, record the installed version (`claude --version`, `codex --version`, `cursor-agent --version`, `agy --version`) and confirm with the host's own view (`/context` or `codex debug prompt-input`; Cursor CLI and Antigravity have none). Without that view, report a hypothesis with the exact check.
