# Native tools

Read before running a host tool, using its output, or asserting version-sensitive loading behavior. `<Claude home>` and `<Codex home>` are the hosts' configuration directories (by default `~/.claude` and `~/.codex`).

## Authorization (HA-ME-01)

- Run static tools first. They read files and at most write a local cache. An auditor that must stay read-only does not run tools that write a cache (`codex debug prompt-input`, `uv run`); the caller runs them and supplies the output.
- Run a model-invoking evaluation only when the caller explicitly authorizes it, with a cost cap and no publication.
- Cover Claude Code and Codex by default. Use tools for other hosts only on explicit request.
- When the caller supplies a tool's output, use it instead of running the tool again. If a needed output is missing and you cannot run the tool within your boundary, request it in your result.

## Static tools

| Tool | Host | Checks | Notes |
| --- | --- | --- | --- |
| `(cd <entry> && codex debug prompt-input)` | Codex | The exact instruction chain Codex sends from that entry directory | No model call; writes only a model-list cache. Works outside git repositories. |
| `claude plugin validate <path> --strict --json` | Claude Code | Plugin or marketplace manifests and the skills, agents, and commands they declare | Needs a plugin or marketplace layout; elsewhere it succeeds with empty `contents`, which is no evidence. `--strict` fails on warnings. |
| `claude plugin details <name>` | Claude Code | Component inventory and projected token cost of an installed plugin | — |
| `quick_validate.py <skill-directory>` from the `skill-creator` skill's `scripts/` | Claude Code and Codex copies | Frontmatter shape, name, and description limits; the Codex copy also flags `[TODO:` placeholders | Needs PyYAML: `uv run --no-project --with pyyaml python <script> <skill-directory>`. The Codex copy lives under `<Codex home>/skills/.system/skill-creator/`. |
| `claude doctor`, `codex doctor --json` | Both | Installation, configuration, and authentication health | Not a content audit. |

## Human-run views

- **Claude Code `/context`** lists the loaded instruction files under "Memory files" for the session's own directory. Neither an agent nor its parent can run it. To confirm loading for an entry point, ask a human to start Claude Code there and run `/context`, stating the hypothesis and the expected entry.
- A past `/context` run is recorded in that session's transcript under `<Claude home>/projects/<entry path with every non-alphanumeric character replaced by ->/`. The per-file "Memory Files" table is a separate Markdown record from the collapsed summary; extract it with `jq -r` rather than reading the whole transcript.
- **Claude Code `/memory`** opens the memory locations interactively.

## Model-invoking tools

These need explicit authorization (HA-ME-01):

- **`claude plugin eval`**: runs real sessions per eval case with graders. Use `--max-cost-usd` and `--no-publish`, and `--threshold` for pass or fail.
- **`skill-creator` evaluation and benchmark scripts** (`run_eval.py`, `run_loop.py`): measure triggering and effect of a skill.
- **Claude Code `/doctor` trim proposals** for `CLAUDE.md`: interactive.
- **`codex review` and `codex exec`**: review diffs or run tasks; they do not audit instruction files.

## Version-sensitive behavior (HA-ME-02)

Loading behavior changes between releases. For example, whether an ancestor `CLAUDE.md` resolves its imports differed between Claude Code releases. Before asserting such behavior:

1. Record the installed version (`claude --version`, `codex --version`).
2. Confirm with the host's own view (`/context` or `codex debug prompt-input`).
3. Without that view, report the conclusion as a hypothesis with the exact check.
