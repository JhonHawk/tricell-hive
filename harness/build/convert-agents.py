#!/usr/bin/env python3
"""Derive per-harness agent definitions from the canonical Claude Code agents.

Source of truth: global/agents/**/*.md (Claude Code format).
Outputs:
  - Codex CLI:  <out>/codex/<name>.toml      (developer_instructions = body)
  - opencode:   <out>/opencode/<name>.md     (mode: subagent, permission map)
  - Grok Build: <out>/grok/<name>.md         (subagent types under ~/.grok/agents/)

Run by /deploy-global before copying to ~/.codex/agents/,
~/.config/opencode/agents/, and ~/.grok/agents/. Never edit generated files by
hand — edit the canonical agent and redeploy.

Usage: convert-agents.py <agents-src-dir> <out-dir>
"""
import re
import sys
from pathlib import Path

GENERATED_NOTE = "Generated from tricell-hive global/agents — do not edit by hand."
CLAUDE_ONLY_FIELDS = {
    "background",
    "disallowedTools",
    "hooks",
    "initialPrompt",
    "isolation",
    "maxTurns",
    "memory",
    "permissionMode",
    "skills",
    "tools",
}
CODEX_REASONING_EFFORTS = {"minimal", "low", "medium", "high", "xhigh", "max", "ultra"}
# Not every model advertises every level. A tier that is not advertised is dropped
# silently by Codex at request time (verified on 0.146.0), so the build rejects it
# instead — keep in sync with `additional_speed_tiers`/`supported_reasoning_levels`
# in ~/.codex/models_cache.json.
CODEX_UNSUPPORTED_EFFORTS = {"gpt-5.6-luna": frozenset({"ultra"})}
# Claude model-alias tiers → (Codex model slug, reasoning effort). Single place to
# update when OpenAI rotates the family. Execution runs on luna at max reasoning;
# judgment roles run on sol. Agents on "inherit" (or an unmapped alias) emit no
# `model` line: they resolve to the [agents] default in config.toml, then to the
# session model.
CODEX_TIER_MAP = {
    "opus": ("gpt-5.6-sol", "high"),
    "sonnet": ("gpt-5.6-luna", "max"),
    "haiku": ("gpt-5.6-luna", "high"),
}


def validate_tier_map():
    for alias, (slug, effort) in CODEX_TIER_MAP.items():
        if effort not in CODEX_REASONING_EFFORTS:
            raise ValueError(f"tier {alias}: unknown reasoning effort '{effort}'")
        if effort in CODEX_UNSUPPORTED_EFFORTS.get(slug, frozenset()):
            raise ValueError(f"tier {alias}: {slug} does not advertise effort '{effort}'")


validate_tier_map()


def parse_agent(path: Path):
    text = path.read_text(encoding="utf-8")
    m = re.match(r"^---\n(.*?)\n---\n(.*)$", text, re.S)
    if not m:
        raise ValueError(f"{path}: no frontmatter")
    fm_text, body = m.groups()

    def field(name):
        # block scalar (>) or plain value, single line continuation aware
        block = re.search(rf"^{name}:\s*>\s*\n((?:[ \t]+.*\n?)+)", fm_text, re.M)
        if block:
            lines = [ln.strip() for ln in block.group(1).splitlines()]
            return " ".join(ln for ln in lines if ln)
        plain = re.search(rf"^{name}:\s*(.+)$", fm_text, re.M)
        return plain.group(1).strip() if plain else None

    field_names = [
        line.split(":", 1)[0]
        for line in fm_text.splitlines()
        if re.match(r"^[A-Za-z][A-Za-z0-9_-]*:", line)
    ]

    name = field("name")
    description = field("description") or ""
    # strip <example> blocks from descriptions — routing aids for Claude only
    description = re.sub(r"<example>.*?</example>", "", description, flags=re.S)
    description = re.sub(r"\s+", " ", description).strip()
    tools_raw = field("tools")
    tools = [t.strip() for t in tools_raw.split(",")] if tools_raw else None
    disallowed_raw = field("disallowedTools")
    disallowed = (
        [t.strip() for t in disallowed_raw.split(",")] if disallowed_raw else []
    )
    color = field("color")
    model = field("model")
    effort = field("effort")
    permission_mode = field("permissionMode")
    max_turns = field("maxTurns")
    skills_raw = field("skills")
    skills = [s.strip() for s in skills_raw.split(",")] if skills_raw else []
    memory = field("memory")
    mcp_servers = field("mcpServers")
    return {
        "name": name,
        "description": description,
        "tools": tools,  # None = inherits all
        "disallowed": disallowed,
        "color": color,
        "model": model,
        "effort": effort,
        "permission_mode": permission_mode,
        "max_turns": max_turns,
        "skills": skills,
        "memory": memory,
        "mcp_servers": mcp_servers,
        "field_names": field_names,
        "body": body.strip(),
    }


def can_write(agent):
    if agent["tools"] is None:
        return not ({"Write", "Edit"} & set(agent["disallowed"]))
    return bool({"Write", "Edit"} & set(agent["tools"]))


def has_bash(agent):
    if agent["tools"] is None:
        return "Bash" not in agent["disallowed"]
    return "Bash" in agent["tools"]


def toml_escape(s: str) -> str:
    return s.replace("\\", "\\\\").replace('"', '\\"')


def comment_escape(s: str) -> str:
    return s.replace("\n", " ").replace("\r", " ").strip()


def comma_join(values) -> str:
    return ", ".join(values)


def codex_model(agent):
    """Map the Claude model to Codex: pass through gpt-*/o-series IDs, translate alias tiers."""
    model = (agent["model"] or "").strip()
    if not model or model == "inherit":
        return None
    # o-series needs a digit after the "o" — a bare startswith("o") would swallow
    # Claude aliases like "opus" and stamp them as Codex model slugs.
    if model.startswith("gpt-") or re.match(r"^o\d", model):
        return model
    tier = CODEX_TIER_MAP.get(model)
    return tier[0] if tier else None


def codex_reasoning_effort(agent):
    """The tier's effort wins over the Claude `effort` frontmatter, which is
    calibrated for Claude's models rather than the Codex tier it maps to. An
    explicit gpt-*/o-series model keeps its frontmatter effort, dropped when the
    model does not advertise it."""
    tier = CODEX_TIER_MAP.get((agent["model"] or "").strip())
    if tier:
        return tier[1]
    effort = (agent["effort"] or "").strip()
    if effort not in CODEX_REASONING_EFFORTS:
        return None
    model = codex_model(agent)
    if model and effort in CODEX_UNSUPPORTED_EFFORTS.get(model, frozenset()):
        return None
    return effort


def codex_compatibility_comments(agent):
    comments = [
        "# Harness compatibility:",
        "# - Codex custom agents support session config plus developer_instructions;",
        "#   Claude-only controls below are preserved as comments.",
    ]

    if agent["tools"] is not None:
        comments.append(f"# Claude tools: {comment_escape(comma_join(agent['tools']))}")
    if agent["disallowed"]:
        comments.append(
            f"# Claude disallowedTools: {comment_escape(comma_join(agent['disallowed']))}"
        )
    if agent["model"] and agent["model"] != "inherit" and not codex_model(agent):
        comments.append(f"# Claude model: {comment_escape(agent['model'])}")
    elif agent["model"] in CODEX_TIER_MAP:
        slug, effort = CODEX_TIER_MAP[agent["model"]]
        comments.append(
            f"# Claude model alias: {agent['model']} -> {slug} @ {effort}"
        )
    resolved_effort = codex_reasoning_effort(agent)
    if agent["effort"] and not resolved_effort:
        comments.append(f"# Claude effort: {comment_escape(agent['effort'])}")
    elif agent["effort"] and agent["effort"].strip() != resolved_effort:
        comments.append(
            f"# Claude effort: {comment_escape(agent['effort'])} (tier default wins on Codex)"
        )
    if agent["permission_mode"]:
        comments.append(f"# Claude permissionMode: {comment_escape(agent['permission_mode'])}")
    if agent["max_turns"]:
        comments.append(f"# Claude maxTurns: {comment_escape(agent['max_turns'])}")
    if agent["skills"]:
        comments.append(f"# Claude skills: {comment_escape(comma_join(agent['skills']))}")
    if agent["memory"]:
        comments.append(f"# Claude memory: {comment_escape(agent['memory'])}")
    if agent["mcp_servers"]:
        comments.append(f"# Claude mcpServers: {comment_escape(agent['mcp_servers'])}")

    preserved = [
        field
        for field in agent["field_names"]
        if field in CLAUDE_ONLY_FIELDS and field != "tools"
    ]
    if preserved:
        comments.append(
            "# Review native Codex support before relying on: "
            + comment_escape(comma_join(sorted(set(preserved))))
        )
    return comments


def codex_extra_instructions(agent):
    instructions = []
    if "Agent" in agent["disallowed"]:
        instructions.append(
            "Do not spawn, delegate to, or coordinate other agents from this agent. "
            "Return findings or changes directly to the parent session."
        )
    for skill in agent["skills"]:
        instructions.append(
            f"When the `{skill}` skill is available and relevant, use it before "
            "performing the specialized workflow manually."
        )
    if agent["permission_mode"] == "plan" and not can_write(agent):
        instructions.append(
            "Operate as read-only: report findings and recommendations without editing files."
        )
    return instructions


def codex_warnings(agent):
    warnings = []
    if agent["model"] and agent["model"] != "inherit" and not codex_model(agent):
        warnings.append(f"model={agent['model']} preserved as comment only")
    if agent["disallowed"]:
        warnings.append(
            f"disallowedTools={comma_join(agent['disallowed'])} has no native Codex per-tool equivalent"
        )
    if agent["skills"]:
        warnings.append(
            f"skills={comma_join(agent['skills'])} translated as developer instruction"
        )
    if agent["memory"]:
        warnings.append(f"memory={agent['memory']} preserved as comment only")
    if agent["permission_mode"]:
        warnings.append(
            f"permissionMode={agent['permission_mode']} preserved as comment/instruction only"
        )
    if agent["max_turns"]:
        warnings.append(f"maxTurns={agent['max_turns']} preserved as comment only")
    return warnings


def to_codex(agent) -> str:
    sandbox = "workspace-write" if can_write(agent) else "read-only"
    body = agent["body"].replace('"""', "'''")
    extra_instructions = codex_extra_instructions(agent)
    if extra_instructions:
        body = body + "\n\n## Codex compatibility instructions\n\n" + "\n".join(
            f"- {instruction}" for instruction in extra_instructions
        )
    lines = [
        f"# {GENERATED_NOTE}",
        f'name = "{agent["name"]}"',
        f'description = "{toml_escape(agent["description"])}"',
        f'sandbox_mode = "{sandbox}"',
        *codex_compatibility_comments(agent),
    ]
    # Always derive nickname_candidates from name so the TUI shows the agent role
    # instead of a random built-in nickname or raw UUID.
    lines.append(f'nickname_candidates = ["{agent["name"]}"]')
    model = codex_model(agent)
    if model:
        lines.append(f'model = "{model}"')
    reasoning_effort = codex_reasoning_effort(agent)
    if reasoning_effort:
        lines.append(f'model_reasoning_effort = "{reasoning_effort}"')
    lines.extend([
        "",
        'developer_instructions = """',
        body,
        '"""',
        "",
    ])
    return "\n".join(lines)


# opencode's config schema only accepts hex (#RRGGBB) or named theme tokens for
# `color` — Claude Code color names are rejected. Map by role semantics (the hub's
# color-by-role scheme): design→primary, review→info, build→success,
# validate→warning, creative→accent, critical/ops→error.
OPENCODE_COLOR = {
    "blue": "primary",
    "cyan": "info",
    "green": "success",
    "yellow": "warning",
    "magenta": "accent",
    "red": "error",
}


def to_opencode(agent) -> str:
    perms = []
    if not can_write(agent):
        perms.append('  edit: "deny"')
        if not has_bash(agent):
            perms.append('  bash: "deny"')
    if "Agent" in agent["disallowed"]:
        perms.append('  task: "deny"')
    fm = [
        "---",
        f"# {GENERATED_NOTE}",
        "description: >",
        f"  {agent['description']}",
        "mode: subagent",
    ]
    if agent["color"] and agent["color"] in OPENCODE_COLOR:
        fm.append(f"color: {OPENCODE_COLOR[agent['color']]}")
    if perms:
        fm.append("permission:")
        fm.extend(perms)
    fm.append("---")
    return "\n".join(fm) + "\n\n" + agent["body"] + "\n"


# Claude tool names → Grok Build tool names. Unknown Claude-only tools are dropped.
GROK_TOOL_MAP = {
    "Read": "read_file",
    "Write": "search_replace",
    "Edit": "search_replace",
    "MultiEdit": "search_replace",
    "Bash": "run_terminal_command",
    "Glob": "list_dir",
    "ListDir": "list_dir",
    "Grep": "grep",
    "WebSearch": "web_search",
    "WebFetch": "web_fetch",
    "Task": "spawn_subagent",
    "Agent": "spawn_subagent",
    "NotebookEdit": None,
}


def grok_tools(agent):
    """Map Claude tools allowlist to Grok names; None = inherit full toolset."""
    if agent["tools"] is None:
        if not agent["disallowed"]:
            return None
        # Full-ish default minus disallowed write tools (in-vivo/ux-flow pattern).
        base = [
            "read_file",
            "list_dir",
            "grep",
            "run_terminal_command",
            "web_search",
            "web_fetch",
            "search_tool",
            "use_tool",
        ]
        if not ({"Write", "Edit", "MultiEdit"} & set(agent["disallowed"])):
            base.insert(4, "search_replace")
        return base

    mapped = []
    seen = set()
    for t in agent["tools"]:
        g = GROK_TOOL_MAP.get(t)
        if g is None:
            continue
        if g not in seen:
            seen.add(g)
            mapped.append(g)
    return mapped or None


def grok_permission_mode(agent):
    if (agent["permission_mode"] or "").strip() == "plan":
        return "plan"
    if not can_write(agent):
        return "plan"
    return "default"


def grok_extra_instructions(agent):
    instructions = []
    if "Agent" in agent["disallowed"]:
        instructions.append(
            "Do not spawn, delegate to, or coordinate other agents from this agent. "
            "Return findings or changes directly to the parent session."
        )
    if grok_permission_mode(agent) == "plan" and not can_write(agent):
        instructions.append(
            "Operate as read-only: report findings and recommendations without editing files."
        )
    for skill in agent["skills"]:
        instructions.append(
            f"When the `{skill}` skill is available and relevant, use it before "
            "performing the specialized workflow manually."
        )
    return instructions


def to_grok(agent) -> str:
    """Grok Build agent definition (.md under ~/.grok/agents/).

    Model is always inherit — Claude opus/sonnet aliases are not Grok slugs.
    """
    body = agent["body"]
    extra = grok_extra_instructions(agent)
    if extra:
        body = body + "\n\n## Grok compatibility instructions\n\n" + "\n".join(
            f"- {instruction}" for instruction in extra
        )

    claude_model = (agent["model"] or "").strip()
    fm = [
        "---",
        f"# {GENERATED_NOTE}",
        f"name: {agent['name']}",
        "description: >",
        f"  {agent['description']}",
        "prompt_mode: full",
        "model: inherit",
        f"permission_mode: {grok_permission_mode(agent)}",
        "agents_md: true",
    ]
    if claude_model and claude_model != "inherit":
        fm.append(f"# Claude model alias (not mapped): {claude_model}")
    tools = grok_tools(agent)
    if tools is not None:
        fm.append(f"tools: {', '.join(tools)}")
    fm.append("---")
    return "\n".join(fm) + "\n\n" + body + "\n"


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    src, out = Path(sys.argv[1]), Path(sys.argv[2])
    codex_dir, oc_dir, grok_dir = out / "codex", out / "opencode", out / "grok"
    codex_dir.mkdir(parents=True, exist_ok=True)
    oc_dir.mkdir(parents=True, exist_ok=True)
    grok_dir.mkdir(parents=True, exist_ok=True)

    count = 0
    for path in sorted(src.rglob("*.md")):
        agent = parse_agent(path)
        (codex_dir / f"{agent['name']}.toml").write_text(
            to_codex(agent), encoding="utf-8"
        )
        (oc_dir / f"{agent['name']}.md").write_text(
            to_opencode(agent), encoding="utf-8"
        )
        (grok_dir / f"{agent['name']}.md").write_text(
            to_grok(agent), encoding="utf-8"
        )
        for warning in codex_warnings(agent):
            print(f"WARNING {agent['name']}: {warning}", file=sys.stderr)
        count += 1
    print(f"converted {count} agents -> {codex_dir} , {oc_dir} , {grok_dir}")


if __name__ == "__main__":
    main()
