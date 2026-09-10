#!/usr/bin/env python3
"""Derive per-harness agent definitions from the canonical Claude Code agents.

Source of truth: global/agents/**/*.md (Claude Code format).
Outputs:
  - Codex CLI:  <out>/codex/<name>.toml      (developer_instructions = body)
  - opencode:   <out>/opencode/<name>.md     (mode: subagent, permission map)
  - Grok Build: <out>/grok/<name>.md         (subagent types under ~/.grok/agents/)
  - Pi:         <out>/pi/<name>.md          (pi-subagents custom agents)

Run by /deploy-global before copying to ~/.codex/agents/,
~/.config/opencode/agents/, ~/.grok/agents/, and the Pi agent directory. Never
edit generated files by hand — edit the canonical agent and redeploy.

Usage: convert-agents.py <agents-src-dir> <out-dir>
"""
import re
import sys
from pathlib import Path

# Deliberately names no source repository or path: an agent that reads its own
# definition and finds one treats it as a place to go verify things, and starts
# wandering out of the repository it was invoked in.
GENERATED_NOTE = "Generated file — do not edit by hand; edit the canonical agent and rebuild."
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
# Claude model-alias tiers → (OpenAI model slug, reasoning effort). Codex and Pi
# share this map so a role keeps the same OpenAI tier in both harnesses. Agents on
# "inherit" (or an unmapped alias) emit no model line in either target.
OPENAI_TIER_MAP = {
    "opus": ("gpt-6-astra", "medium"),
    "sonnet": ("gpt-5.6-luna", "max"),
    "haiku": ("gpt-5.6-luna", "high"),
}
# Keep the existing name as the Codex-facing compatibility surface used by the
# converter and its tests; the value is intentionally shared, not copied.
CODEX_TIER_MAP = OPENAI_TIER_MAP

PI_MODEL_PREFIX = "openai-codex/"
PI_THINKING_LEVELS = {
    "off",
    "minimal",
    "low",
    "medium",
    "high",
    "xhigh",
    "max",
}
PI_ROOT_PLACEHOLDER = "__HIVE_PI_ROOT__"
PI_GENERAL_EXTENSION = f"{PI_ROOT_PLACEHOLDER}/extensions/hive-hooks.ts"
PI_REVIEWER_EXTENSION = f"{PI_ROOT_PLACEHOLDER}/extensions/hive/reviewer-guard.ts"
PI_READ_MEMORY_TOOLS = ("mem_search", "mem_context", "mem_get_observation")
PI_WRITE_MEMORY_TOOLS = ("mem_save",)
PI_CHILD_TOOLS = ("contact_supervisor",)
PI_GIT_READ_TOOL = "hive_git_read"
PI_HOOK_READINESS_TOOL = "hive_hook_readiness"
PI_REVIEWER_READINESS_TOOL = "hive_reviewer_readiness"


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
        # Claude's nested hook frontmatter is deliberately not converted as a
        # hook object. The reviewer guard is a required PI child extension, so
        # retain only the observable source contract that selects it.
        "has_reviewer_guard": "reviewer-guard.sh" in fm_text,
        "field_names": field_names,
        "body": body.strip(),
    }


def can_write(agent):
    # Either write tool still available => the agent can write. Denying only one
    # of them (in-vivo-qa-tester denies Edit but keeps Write for its report) must
    # not zero out write capability in the generated trees.
    available = {"Write", "Edit"} if agent["tools"] is None else set(agent["tools"])
    available.difference_update(agent["disallowed"])
    return bool(available & {"Write", "Edit"})


def denies_agent(agent):
    # Claude Code treats both forms as the same denial: "omit `Agent` from its
    # `tools` list or add it to `disallowedTools`" (sub-agents docs). opencode
    # defaults `task` to allow and Codex has no per-tool control, so the
    # allowlist form has to be translated explicitly for them. Grok needs no
    # case here — grok_tools maps the allowlist directly, so `spawn_subagent`
    # is already absent whenever `Agent` is.
    if "Agent" in agent["disallowed"]:
        return True
    return agent["tools"] is not None and "Agent" not in agent["tools"]


def has_bash(agent):
    if "Bash" in agent["disallowed"]:
        return False
    if agent["tools"] is None:
        return True
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


def openai_opus_effort_override(agent):
    """Return whether the OpenAI tier policy overrides Claude's declared effort."""
    return (
        (agent["model"] or "").strip() == "opus"
        and (agent["effort"] or "").strip() == "high"
    )


def codex_reasoning_effort(agent):
    """Resolve Codex effort, including the shared Astra medium policy.

    Explicit effort remains authoritative for every role except the deliberate
    ``opus``/``high`` mapping: those judgment roles move from the old Sol/high
    tier to Astra/medium in Codex and PI together.
    """
    model = codex_model(agent)
    declared = (agent["effort"] or "").strip()
    if openai_opus_effort_override(agent):
        return OPENAI_TIER_MAP["opus"][1]
    if declared in CODEX_REASONING_EFFORTS and declared not in CODEX_UNSUPPORTED_EFFORTS.get(
        model, frozenset()
    ):
        return declared
    tier = CODEX_TIER_MAP.get((agent["model"] or "").strip())
    if tier:
        return tier[1]
    return None


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
    resolved_effort = codex_reasoning_effort(agent)
    if agent["model"] and agent["model"] != "inherit" and not codex_model(agent):
        comments.append(f"# Claude model: {comment_escape(agent['model'])}")
    elif agent["model"] in CODEX_TIER_MAP:
        slug, tier_effort = CODEX_TIER_MAP[agent["model"]]
        note = f"# Claude model alias: {agent['model']} -> {slug} @ {tier_effort}"
        if resolved_effort and resolved_effort != tier_effort:
            note += f" (frontmatter effort {resolved_effort} wins)"
        comments.append(note)
    if agent["effort"] and not resolved_effort:
        comments.append(f"# Claude effort: {comment_escape(agent['effort'])}")
    elif openai_opus_effort_override(agent):
        comments.append(
            "# Claude effort: high (OpenAI Astra tier policy uses medium on Codex)"
        )
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
    if denies_agent(agent):
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
    body = rebase_skill_root(agent["body"]).replace('"""', "'''")
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


# Agent bodies cite role rules by absolute path so the agent can Read them
# without a skill tool — the only mechanism all four harnesses share. There are
# exactly two roots, and which one is correct depends on the harness:
#   ~/.claude/skills  — Claude Code (deployed there) and Grok (scans it)
#   ~/.agents/skills  — Codex, opencode, and PI (shared universal skills)
# Sources are written with the Claude root; each converter rebases it for its
# target harness.
CLAUDE_SKILL_ROOT = "~/.claude/skills/"
AGENTS_SKILL_ROOT = "~/.agents/skills/"
PI_SKILL_ROOT = AGENTS_SKILL_ROOT


def rebase_skill_root(body: str) -> str:
    return body.replace(CLAUDE_SKILL_ROOT, AGENTS_SKILL_ROOT)


def rebase_pi_skill_root(body: str) -> str:
    return body.replace(CLAUDE_SKILL_ROOT, PI_SKILL_ROOT)


def opencode_extra_instructions(agent):
    """opencode cannot preload a skill into an agent.

    Its own `permission.skill` and `tools.skill` keys restrict which skills an
    agent may reach — neither attaches one. Codex and Grok translate `skills:`
    into a prose instruction; opencode was the only harness where the field
    vanished with no warning at all.

    Agent-spawn and write restrictions are deliberately NOT repeated here:
    opencode enforces those deterministically through the `permission` block,
    and an instruction restating a hard denial only invites arguing with it.
    """
    return [
        f"When the `{skill}` skill is available and relevant, use it before "
        "performing the specialized workflow manually."
        for skill in agent["skills"]
    ]


def to_opencode(agent) -> str:
    perms = []
    if not can_write(agent):
        perms.append('  edit: "deny"')
        if not has_bash(agent):
            perms.append('  bash: "deny"')
    if denies_agent(agent):
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
    if not perms:
        # opencode 2.x (beta) silently drops an agent whose frontmatter ends on
        # `color:`. `edit: "allow"` is opencode's own default, so this keeps the
        # key off the last line without changing behavior on any harness.
        # Remove once v2 parses a trailing `color:` correctly.
        perms.append('  edit: "allow"')
    fm.append("permission:")
    fm.extend(perms)
    fm.append("---")
    body = rebase_skill_root(agent["body"])
    extra = opencode_extra_instructions(agent)
    if extra:
        body = body + "\n\n## opencode compatibility instructions\n\n" + "\n".join(
            f"- {instruction}" for instruction in extra
        )
    return "\n".join(fm) + "\n\n" + body + "\n"


# Claude tool names → pi-subagents tool names. Unknown Claude-only tools are
# omitted; MCP entries use the adapter's stable gateway proxy.
PI_TOOL_MAP = {
    "Read": "read",
    "Write": "write",
    "Edit": "edit",
    "MultiEdit": "edit",
    "Bash": "bash",
    "Glob": "find",
    "ListDir": "ls",
    "Grep": "grep",
    "WebSearch": "web_search",
    "WebFetch": "fetch_content",
    "Task": "subagent",
    "Agent": "subagent",
    "NotebookEdit": None,
}
# A missing Claude `tools:` field means the role inherits the full Claude
# surface. PI requires an explicit list for extension-tool readiness checks, so
# retain the corresponding PI builtins and installed research/MCP tools before
# applying source disallowedTools. The role-specific additions below then add
# memory, supervision, Git, and Hive readiness sentinels.
PI_INHERITED_TOOL_BASELINE = (
    "read",
    "write",
    "edit",
    "bash",
    "find",
    "ls",
    "grep",
    "web_search",
    "fetch_content",
    "get_search_content",
    "source_check",
    "mcp",
    "subagent",
)


def pi_mcp_server(tool):
    """Map Claude's mcp__server__tool spelling to PI's stable gateway tool."""
    if not tool.startswith("mcp__"):
        return None
    parts = tool.split("__", 2)
    if len(parts) != 3 or not parts[1] or not parts[2]:
        return None
    return "mcp"


def pi_mcp_exclusion_names(tool):
    """Map a denied MCP operation to the stable gateway's deny entry."""
    if not tool.startswith("mcp__"):
        return []
    parts = tool.split("__", 2)
    if len(parts) != 3 or not parts[1] or not parts[2]:
        return []
    return ["mcp"]


def pi_tool_name(tool):
    return pi_mcp_server(tool) or PI_TOOL_MAP.get(tool)


def pi_tools(agent, field):
    """Map an allow/deny field while preserving order and removing duplicates."""
    values = agent[field]
    mapped = []
    seen = set()
    for tool in values or []:
        resolved_names = (
            pi_mcp_exclusion_names(tool)
            if field == "disallowed"
            else [pi_tool_name(tool)]
        )
        if field == "disallowed" and not resolved_names:
            resolved_names = [pi_tool_name(tool)]
        for resolved in resolved_names:
            if resolved and resolved not in seen:
                seen.add(resolved)
                mapped.append(resolved)
    return mapped


def pi_allowlist(agent):
    """Build a strict PI list, including the inherited source-tool surface."""
    if agent["tools"] is None:
        denied = set(pi_tools(agent, "disallowed"))
        return [tool for tool in PI_INHERITED_TOOL_BASELINE if tool not in denied]
    return pi_tools(agent, "tools")


def pi_model(agent):
    """Resolve a source model to a fully-qualified OpenAI provider model."""
    model = (agent["model"] or "").strip()
    if not model:
        return None
    if model == "inherit":
        return "inherit"
    if "/" in model:
        return model
    tier = OPENAI_TIER_MAP.get(model)
    slug = tier[0] if tier else None
    if not slug and (model.startswith("gpt-") or re.match(r"^o\d", model)):
        slug = model
    return f"{PI_MODEL_PREFIX}{slug}" if slug else None


def pi_thinking(agent):
    """Resolve PI thinking while applying the same Astra exception as Codex."""
    declared = (agent["effort"] or "").strip()
    if openai_opus_effort_override(agent):
        return OPENAI_TIER_MAP["opus"][1]
    if declared in PI_THINKING_LEVELS:
        return declared
    tier = OPENAI_TIER_MAP.get((agent["model"] or "").strip())
    return tier[1] if tier and tier[1] in PI_THINKING_LEVELS else None


def pi_additional_tools(agent):
    """Return native PI capabilities required by the role contract."""
    memory = PI_WRITE_MEMORY_TOOLS if can_write(agent) else PI_READ_MEMORY_TOOLS
    tools = [*memory, *PI_CHILD_TOOLS]
    # PI's plan-mode Bash policy can remove the shell before a child runs. The
    # structured read-only Git tool therefore follows the canonical Bash
    # capability, including non-reviewer verifiers such as finding-refuter.
    # `has_bash` applies source deny-wins semantics for both allowlists and
    # disallowedTools; reviewer-guard selection is independent of Git access.
    if has_bash(agent):
        tools.append(PI_GIT_READ_TOOL)
    return tools


def pi_readiness_tools(agent):
    """Return extension sentinels required before the agent's first turn."""
    tools = [PI_HOOK_READINESS_TOOL]
    if agent["has_reviewer_guard"]:
        tools.append(PI_REVIEWER_READINESS_TOOL)
    return tools


def pi_context7_instructions(agent):
    """Give PI agents with Context7 source tools the guarded gateway calls."""
    if not agent["tools"] or not any(tool.startswith("mcp__") for tool in agent["tools"]):
        return ""
    return (
        "\n\n## PI Context7 usage\n\n"
        "For version-sensitive claims, use the shared `mcp` gateway in this order:\n"
        "1. `mcp({tool:'context7_resolve-library-id',args:{query,libraryName}})`\n"
        "2. `mcp({tool:'context7_query-docs',args:{libraryId,query}})`"
    )


def to_pi(agent) -> str:
    """Render a pi-subagents v0.67 custom agent definition."""
    tools = pi_allowlist(agent)
    seen = set(tools)
    for tool in [*pi_additional_tools(agent), *pi_readiness_tools(agent)]:
        if tool not in seen:
            seen.add(tool)
            tools.append(tool)
    excluded = pi_tools(agent, "disallowed")
    extensions = [PI_GENERAL_EXTENSION]
    if agent["has_reviewer_guard"]:
        extensions.append(PI_REVIEWER_EXTENSION)

    fm = [
        "---",
        f"# {GENERATED_NOTE}",
        f"name: {agent['name']}",
        "description: >",
        f"  {agent['description']}",
    ]
    model = pi_model(agent)
    if model:
        fm.append(f"model: {model}")
    thinking = pi_thinking(agent)
    if thinking:
        fm.append(f"thinking: {thinking}")
    if tools:
        fm.append(f"tools: {', '.join(tools)}")
    if excluded:
        fm.append(f"excludeTools: {', '.join(excluded)}")
    fm.extend([
        f"subagentOnlyExtensions: {', '.join(extensions)}",
        "async: true",
        "defaultContext: fresh",
        "systemPromptMode: append",
        "inheritProjectContext: true",
        "inheritGlobalContext: true",
        "inheritSkills: true",
        "allowNestedSubagents: false",
        "memory:",
        "  scope: project",
        f"  path: hive/{agent['name']}",
        "---",
    ])
    body = rebase_pi_skill_root(agent["body"]) + pi_context7_instructions(agent)
    return "\n".join(fm) + "\n\n" + body + "\n"


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
        # search_replace is Grok's only write tool (Write/Edit/MultiEdit all map
        # to it), so it follows can_write: present whenever either write tool
        # survives the denylist.
        if can_write(agent):
            base.insert(4, "search_replace")
        return base

    mapped = []
    seen = set()
    # MCP tools (mcp__server__tool) have no direct Grok names — grant tool
    # discovery (search_tool + use_tool) so the agent can reach the MCP server.
    if any(t.startswith("mcp__") for t in agent["tools"]):
        mapped.extend(["search_tool", "use_tool"])
        seen.update(mapped)
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
    Do not emit capability_mode (removed as a spawn argument in grok 1.0.6;
    tools come from the agent type / this allowlist). The runtime withholds
    the workflow tool from subagents (1.0.8+); no frontmatter needed.
    """
    # NOT rebased: Grok scans ~/.claude/skills and never ~/.agents/skills.
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
    codex_dir, oc_dir, grok_dir, pi_dir = (
        out / "codex",
        out / "opencode",
        out / "grok",
        out / "pi",
    )
    codex_dir.mkdir(parents=True, exist_ok=True)
    oc_dir.mkdir(parents=True, exist_ok=True)
    grok_dir.mkdir(parents=True, exist_ok=True)
    pi_dir.mkdir(parents=True, exist_ok=True)

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
        (pi_dir / f"{agent['name']}.md").write_text(
            to_pi(agent), encoding="utf-8"
        )
        for warning in codex_warnings(agent):
            print(f"WARNING {agent['name']}: {warning}", file=sys.stderr)
        count += 1
    print(
        f"converted {count} agents -> {codex_dir} , {oc_dir} , {grok_dir} , {pi_dir}"
    )


if __name__ == "__main__":
    main()
