#!/usr/bin/env python3
"""Derive per-harness agent definitions from the canonical Claude Code agents.

Source of truth: global/agents/**/*.md (Claude Code format).
Outputs:
  - Claude Code: <out>/claude/<role>/<name>.md (source frontmatter + inlined packs)
  - Codex CLI:  <out>/codex/<name>.toml      (developer_instructions = body)
  - opencode:   <out>/opencode/<name>.md     (mode: subagent, permission map)
  - Grok Build: <out>/grok/<name>.md         (subagent types under ~/.grok/agents/)
  - Pi:         <out>/pi/<name>.md          (pi-subagents custom agents)

Run by /deploy-global before copying to ~/.claude/agents/, ~/.codex/agents/,
~/.config/opencode/agents/, ~/.grok/agents/, and the Pi agent directory. Never
edit generated files by hand — edit the canonical agent and redeploy.

Usage: convert-agents.py <agents-src-dir> <out-dir>
"""
import re
import sys
import tomllib
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
    "omitClaudeMd",
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
    "sonnet": ("gpt-5.6-luna", "high"),
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
PI_RESEARCH_READINESS_TOOL = "hive_research_readiness"


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
    packs = parse_packs(name, field("packs"), fm_text)
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
        # Hive-only field. A pack is a rule file's basename: resolution and the
        # inlined text are settled here so an unknown pack fails ONE build, not
        # five renders.
        "packs": packs,
        "pack_texts": [(pack, resolve_pack(name, pack)) for pack in packs],
        "memory": memory,
        "mcp_servers": mcp_servers,
        # The source frontmatter verbatim — only the Claude output re-emits it.
        "frontmatter": fm_text,
        # Claude's nested hook frontmatter is deliberately not converted as a
        # hook object. The reviewer guard is a required PI child extension, so
        # retain only the observable source contract that selects it.
        "has_reviewer_guard": "reviewer-guard.sh" in fm_text,
        "field_names": field_names,
        "body": body.strip(),
    }


def can_write(agent):
    # Either write tool still available => the agent can write. Denying only one
    # of them (sdd-verify denies Edit but keeps Write for its report) must
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


# Characters a TOML basic string forbids RAW: every control character except
# tab and newline, plus DEL. A multi-line basic string also normalizes CRLF to
# LF, so a bare CR is escaped too — otherwise `\r\n` would come back as `\n`
# and the alteration would be invisible. Escaped, every one of them round-trips
# exactly; nothing is ever silently rewritten or dropped.
TOML_FORBIDDEN_RAW = re.compile(r"[\x00-\x08\x0b-\x1f\x7f]")


def toml_multiline_escape(s: str) -> str:
    """Escape a body for a TOML multi-line BASIC string, losslessly.

    `developer_instructions = \"\"\"…\"\"\"` is a BASIC string: TOML interprets
    backslash sequences inside it. Agent prose never carried one, but rule
    texts do — `infra-naming` holds `development\\|qa\\|production` (an
    "Unescaped '\\'" parse error) and `typescript-standards` holds a literal
    `\\n` inside backticks, which TOML turns into a real newline with nothing
    in the rendered diff to show for it. Neither survives a review by reading.

    Backslash first: every escape inserted afterwards starts with one. Then the
    raw-forbidden characters as `\\uXXXX`, then any run of three-or-more quotes
    — escaped whole, so runs of 4, 5, 6 … are covered by the same branch as 3.

    Correctness here is not enough on its own: `verify_codex_round_trip` parses
    the result back on every build, so a gap fails the build instead of
    shipping.
    """
    s = s.replace("\\", "\\\\")
    s = TOML_FORBIDDEN_RAW.sub(lambda m: f"\\u{ord(m.group()):04X}", s)
    return re.sub(r'"{3,}', lambda m: '\\"' * len(m.group()), s)


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


def codex_body(agent) -> str:
    """The instructions Codex is MEANT to receive — before any TOML encoding.

    Named separately because it is the yardstick `verify_codex_round_trip`
    measures the emitted file against: comparing the output to itself would
    prove nothing.
    """
    body = packed_body(agent, rebase=rebase_skill_root)
    extra_instructions = codex_extra_instructions(agent)
    if extra_instructions:
        body = body + "\n\n## Codex compatibility instructions\n\n" + "\n".join(
            f"- {instruction}" for instruction in extra_instructions
        )
    return body


def verify_codex_round_trip(agent, rendered):
    """Parse the generated TOML back and demand the exact intended text.

    The escaper being correct is a claim; this is the check. It runs on every
    agent on every build, so a rule text with a character the escaper does not
    handle fails the build naming the agent — instead of shipping a Codex file
    that either refuses to load or silently says something else.
    """
    expected = codex_body(agent) + "\n"
    try:
        parsed = tomllib.loads(rendered)
    except tomllib.TOMLDecodeError as exc:
        raise ValueError(
            f"{agent['name']}: the generated Codex TOML does not parse ({exc}). "
            f"Check the line the error names: the body (toml_multiline_escape) "
            f"or an emitted field — description or name (toml_escape)."
        ) from None
    actual = parsed.get("developer_instructions")
    if actual != expected:
        raise ValueError(
            f"{agent['name']}: the generated Codex TOML parses, but its "
            f"developer_instructions differ from the intended body — the "
            f"string encoding altered the text "
            f"({len(expected)} chars intended, "
            f"{len(actual) if actual is not None else 'none'} read back)."
        )


def to_codex(agent) -> str:
    sandbox = "workspace-write" if can_write(agent) else "read-only"
    body = toml_multiline_escape(codex_body(agent))
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


# One flat store for every rule text, always-on ones included: a core section
# `include:`s the text it carries, it does not hold a copy.
RULES_DIR = Path(__file__).resolve().parents[2] / "global" / "rules-situational"
ROLE_RULES_SECTION = re.compile(r"^## Role rules\n.*?(?=^## |\Z)", re.M | re.S)
ROLE_RULE_TARGET = re.compile(r"/references/([A-Za-z0-9_-]+\.md)`")

# --- Packs ---------------------------------------------------------------
#
# A specialized agent declares `packs:` and receives those rule texts inlined
# into its prompt, in EVERY harness. That is the only delivery mechanism all
# five share: no harness but Claude Code can preload a skill into a subagent,
# and a Role rules row is a pointer an agent may or may not follow. Inlining
# makes presence a build-time fact instead of a routing probability — which is
# also why the rows a pack covers are dropped: keeping both would tell the
# agent to go read what it is already holding.
# A pack name IS a rule file's basename, so it is spelled like one: lowercase,
# digits and hyphens. The pattern is a SECURITY boundary, not a style check —
# an unconstrained name reaches the filesystem, and `../../CLAUDE.local`
# resolved to the owner's private, gitignored file and inlined it into all five
# harnesses. Resolution below never builds a glob out of the name either.
PACK_NAME = re.compile(r"^[a-z0-9][a-z0-9-]*$")
# Every packed agent must carry this one. `omitClaudeMd: true` drops the
# always-on corpus on Claude Code — destructive-op confirmation, secrets
# hygiene, the package manager, the output language — and nothing else puts
# them back. A packed agent without it is a subagent with no gates.
REQUIRED_PACK = "agent-core-gates"
CARRIED_RULES_HEADING = "## Carried rules"
CARRIED_RULES_INTRO = (
    "These conventions are already loaded below, complete as written — never "
    "look for them in rule files, skills, or anywhere else."
)
PACK_LIST_FORM_HINT = (
    "use the comma-separated form `packs: a, b` — one top-level line, "
    "unquoted lowercase key, plain value on the same line: no YAML list, no "
    "block scalar (`>`/`|`), no quotes, no inline comment, never empty."
)


def _rule_files():
    """Every file a pack may resolve to. READMEs are not rules.

    Two exclusions that are not style: a DIRECTORY named `x.md` (reading it
    raises a bare IsADirectoryError instead of naming the offending pack), and
    anything whose real path leaves the store — a symlink inside it pointing
    anywhere on the machine would otherwise be inlined into all five harnesses.
    """
    # The store itself must be real: the escape check below compares against
    # the RESOLVED store, so a symlinked store resolves to wherever it points
    # and then confirms every file under it is "inside" — the same vacuity the
    # core-assembly include check guards (`build.py > _resolve_core_include`).
    if RULES_DIR.is_symlink() or not RULES_DIR.is_dir():
        raise ValueError(
            f"{_display_path(RULES_DIR)} is not a real directory (a symlink, a "
            f"file, or missing). Pack resolution anchors its escape check on "
            f"the store, so a symlinked store would inline whatever it points at."
        )
    store = RULES_DIR.resolve()
    for path in sorted(RULES_DIR.glob("*.md")):
        if path.name == "README.md" or not path.is_file():
            continue
        if store != path.resolve() and store not in path.resolve().parents:
            continue
        yield path


def _display_path(path: Path) -> str:
    """Repo-relative when possible — an absolute temp path helps nobody."""
    try:
        return path.relative_to(Path(__file__).resolve().parents[2]).as_posix()
    except ValueError:
        return str(path)


# The ONE accepted spelling. Anything else matching `^packs\s*:` is an error:
# `packs : ts` parsed as no packs at all AND survived verbatim into the Claude
# frontmatter, shipping a hive-only key to Claude Code — a silent no-op is the
# worst outcome for a field whose whole job is delivering conventions.
# The ONE accepted line: unquoted lowercase key at column 0, plain scalar on
# the same line — no block-scalar indicator, no quotes, no inline comment.
# Everything else is rejected by _packs_declarations below, so by the time
# `claude_frontmatter` removes `^packs:` this is provably the only shape that
# can exist: the removal cannot orphan a continuation line under the previous
# key (`tools: Read` + an orphaned `  agent-core-gates` is one tool allowlist
# to a YAML parser, in a file deployed to Claude Code).
PACKS_KEY_CANONICAL = re.compile(r"^packs:[ \t]*[^\s>|'\"#][^#\n]*$")
# A frontmatter line that declares SOME key. The value group tells a block
# scalar apart from a plain one.
FRONTMATTER_KEY = re.compile(
    r"^(?P<indent>[ \t]*)(?P<key>\"[^\"]*\"|'[^']*'|[^:\s]+)[ \t]*:(?P<value>.*)$"
)


def _packs_declarations(frontmatter):
    """Every line declaring a `packs` key, however it is spelled.

    Reads KEYS, never text: a block scalar's continuation lines are prose
    (`description: >` may legitimately discuss packs) and nested YAML belongs
    to its own key, so neither can be mistaken for a declaration. Case and
    surrounding quotes are folded, because `Packs:`/`"packs":` used to exit 0
    with nothing inlined — an authoring typo that silently produced an agent
    with no packs and no required-gates check.
    """
    declarations = []
    block_indent = None
    for line in frontmatter.splitlines():
        if not line.strip():
            continue
        indent = len(line) - len(line.lstrip(" \t"))
        if block_indent is not None:
            if indent > block_indent:
                continue
            block_indent = None
        match = FRONTMATTER_KEY.match(line)
        if not match:
            continue
        if match.group("value").lstrip()[:1] in (">", "|"):
            block_indent = indent
        if match.group("key").strip("\"'").lower() == "packs":
            declarations.append(line)
    return declarations


def parse_packs(agent_name, packs_raw, frontmatter):
    """Validate the declared pack list before anything touches the filesystem."""
    declarations = _packs_declarations(frontmatter)
    if not declarations:
        return []
    if len(declarations) > 1:
        raise ValueError(
            f"{agent_name}: `packs:` is declared {len(declarations)} times — "
            f"only one would be read and the rest would vanish silently. "
            f"Merge them into a single line."
        )
    if not PACKS_KEY_CANONICAL.match(declarations[0]):
        raise ValueError(f"{agent_name}: {PACK_LIST_FORM_HINT}")
    # ASCII spacing only: a NBSP or a stray CR is part of the name, not
    # padding, and a name that merely LOOKS right must fail rather than
    # resolve to nothing or to something else.
    packs = [p.strip(" \t") for p in (packs_raw or "").split(",")]
    # A YAML list reaches `field()` as the first item with its dash attached
    # (`- typescript-standards`), because the plain-value regex walks onto the
    # continuation line — so the leading dash IS the signal, and without this
    # branch the rest of the list would vanish silently.
    if (not packs_raw) or any(not p for p in packs) \
            or re.search(r"[\[\]]", packs_raw) or packs[0].startswith("-"):
        raise ValueError(f"{agent_name}: {PACK_LIST_FORM_HINT}")
    seen = set()
    for pack in packs:
        if not PACK_NAME.match(pack):
            raise ValueError(
                f"{agent_name}: invalid pack name '{pack}' — a pack is a rule "
                f"file's basename: lowercase letters, digits and hyphens only."
            )
        if pack in seen:
            raise ValueError(
                f"{agent_name}: pack '{pack}' is declared twice — each pack's "
                f"text is carried once, so a repeat is a typo, not an emphasis."
            )
        seen.add(pack)
    if REQUIRED_PACK and REQUIRED_PACK not in packs:
        raise ValueError(
            f"{agent_name}: a packed agent must declare the '{REQUIRED_PACK}' "
            f"pack. Its Claude output carries `omitClaudeMd: true`, which drops "
            f"the always-on gates (destructive ops, secrets, package manager, "
            f"output language); '{REQUIRED_PACK}' is what restores them."
        )
    return packs


def resolve_pack(agent_name, pack):
    """Pack name -> rule text, frontmatter stripped. Fails loudly, never quietly.

    A pack IS a rule file's basename without `.md`, matched EXACTLY against the
    rule files — never as a glob, so a name can neither escape the store nor
    match several files by wildcard. `parse_packs` has already vetted the
    spelling; this resolves it.
    """
    candidates = [path for path in _rule_files() if path.stem == pack]
    if not candidates:
        raise ValueError(
            f"{agent_name}: unknown pack '{pack}' — no {pack}.md under "
            f"global/rules-situational/."
        )
    # No ambiguity branch: the store is one flat directory, so a stem resolves
    # to at most one file by construction. It had one while two stores existed.
    path = candidates[0]
    raw = path.read_text(encoding="utf-8")
    match = re.match(r"^---\n.*?\n---\n?", raw, re.S)
    text = (raw[match.end():] if match else raw).strip("\n")
    if not text:
        raise ValueError(
            f"{agent_name}: pack '{pack}' ({_display_path(path)}) has no text "
            f"outside its frontmatter — inlining it would carry nothing."
        )
    return text


def carried_rules_section(agent):
    """The `## Carried rules` block: heading, the do-not-hunt line, pack texts."""
    parts = [CARRIED_RULES_HEADING, "", CARRIED_RULES_INTRO]
    for pack, text in agent["pack_texts"]:
        # A rule text need not open with a title of its own; the label is what
        # keeps its sections from reading as part of the pack above it.
        parts.extend(["", f"**Rule `{pack}` — carried in full below:**", "", text])
    return "\n".join(parts)


def packed_body(agent, rebase=None):
    """Agent body with pack-covered rows dropped and the pack texts appended.

    Only a PACK drops a row now. The other filter — rows a harness loaded by
    itself — died with the always-on store: no harness reads a rule file
    natively any more, so a dropped row is a rule the agent is never told
    about rather than one it already holds.

    `rebase` rewrites the skill root — and applies to the AGENT's body ONLY. A
    Role rules row cites the root its harness reads, so it must be rebased; a
    rule TEXT is documentation about the world, and rewriting it corrupts the
    rule that explains which harness reads which root (`unattended-autonomy-mode`
    became "Claude Code and Grok read it from ~/.agents/skills"). Carried text
    is carried verbatim.
    """
    drop = frozenset(f"{pack}.md" for pack in agent["packs"])
    body = drop_covered_role_rules(agent["body"], drop) if drop else agent["body"]
    if rebase:
        body = rebase(body)
    if not agent["packs"]:
        return body
    return f"{body}\n\n{carried_rules_section(agent)}"


def drop_covered_role_rules(body: str, carried: frozenset) -> str:
    """Remove Role rules rows the agent already carries inlined; drop an emptied section."""

    def filtered(section):
        lines = section.group(0).splitlines(keepends=True)
        kept = [
            ln
            for ln in lines
            if not (
                (target := ROLE_RULE_TARGET.search(ln)) and target.group(1) in carried
            )
        ]
        if not any(ROLE_RULE_TARGET.search(ln) for ln in kept):
            return ""
        return "".join(kept)

    trailing = body[len(body.rstrip("\n")) :]
    return ROLE_RULES_SECTION.sub(filtered, body).rstrip("\n") + trailing


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
    body = packed_body(agent, rebase=rebase_skill_root)
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
# Claude's research tools are capability groups in Pi. A source WebFetch grant
# needs both retrieval operations, while WebSearch also needs source checking.
# The expansion is used for allowlists and deny entries so an explicit source
# deny remains authoritative over every mapped Pi tool.
PI_TOOL_EXPANSIONS = {
    "WebFetch": ("fetch_content", "get_search_content"),
    "WebSearch": ("fetch_content", "get_search_content", "web_search", "source_check"),
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


def pi_tool_names(tool):
    expanded = PI_TOOL_EXPANSIONS.get(tool)
    if expanded:
        return list(expanded)
    mapped = pi_mcp_server(tool) or PI_TOOL_MAP.get(tool)
    return [mapped] if mapped else []


def pi_tool_name(tool):
    names = pi_tool_names(tool)
    return names[0] if names else None


def pi_tools(agent, field):
    """Map an allow/deny field while preserving order and removing duplicates."""
    values = agent[field]
    mapped = []
    seen = set()
    for tool in values or []:
        resolved_names = (
            pi_mcp_exclusion_names(tool)
            if field == "disallowed"
            else pi_tool_names(tool)
        )
        if field == "disallowed" and not resolved_names:
            resolved_names = pi_tool_names(tool)
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
    # Structured read-only Git follows the canonical Bash capability, including
    # non-reviewer verifiers such as review-refuter. It is independent of
    # planning authorization and remains available under reviewer restrictions.
    # `has_bash` applies source deny-wins semantics for both allowlists and
    # disallowedTools; reviewer-guard selection is independent of Git access.
    if has_bash(agent):
        tools.append(PI_GIT_READ_TOOL)
    return tools


def pi_research_profile(agent):
    """Return the research profile implied by the source tool surface."""
    if agent["tools"] is None:
        return "web"
    source_tools = set(agent["tools"])
    if "WebSearch" in source_tools:
        return "web"
    if "WebFetch" in source_tools:
        return "documentation"
    return None


def pi_readiness_tools(agent):
    """Return extension sentinels required before the agent's first turn."""
    tools = [PI_HOOK_READINESS_TOOL]
    if agent["has_reviewer_guard"]:
        tools.append(PI_REVIEWER_READINESS_TOOL)
    if pi_research_profile(agent) is not None:
        tools.append(PI_RESEARCH_READINESS_TOOL)
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


def pi_research_instructions(agent):
    """Add a non-blocking readiness instruction to research-capable roles."""
    profile = pi_research_profile(agent)
    if profile is None:
        return ""
    return (
        "\n\n## PI research readiness\n\n"
        f"Before external research, call `hive_research_readiness` with profile `{profile}`. "
        "It inspects this agent's active tools and reports `available`, `missing`, and `ready`. "
        "Treat a missing research tool as informational: continue local tasks, but do not "
        "pretend an unavailable tool or provider is ready."
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
    body = (
        packed_body(agent, rebase=rebase_pi_skill_root)
        + pi_context7_instructions(agent)
        + pi_research_instructions(agent)
    )
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
    body = packed_body(agent)
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


def claude_frontmatter(agent):
    """Source frontmatter, minus `packs:`, plus `omitClaudeMd:` when packed.

    Everything else is re-emitted verbatim: this is the only target whose
    frontmatter dialect IS the source dialect, so translating it would be
    inventing differences. `packs:` is hive-only and would be an unknown key
    to Claude Code; `omitClaudeMd: true` is what makes a pack the agent's rule
    corpus instead of a second copy stacked on the global one.
    """
    lines = [
        line for line in agent["frontmatter"].splitlines()
        if not re.match(r"^packs:", line)
    ]
    if agent["packs"] and not any(
        re.match(r"^omitClaudeMd:", line) for line in lines
    ):
        lines.append("omitClaudeMd: true")
    return lines


def to_claude(agent) -> str:
    """Claude Code agent definition (.md under ~/.claude/agents/<role>/).

    No harness loads a rule file by itself any more, so Role rules rows
    survive here unless a pack covers them — and an agent with no packs
    renders its source body unchanged.
    """
    fm = ["---", f"# {GENERATED_NOTE}", *claude_frontmatter(agent), "---"]
    return "\n".join(fm) + "\n\n" + packed_body(agent) + "\n"


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    src, out = Path(sys.argv[1]), Path(sys.argv[2])
    claude_dir, codex_dir, oc_dir, grok_dir, pi_dir = (
        out / "claude",
        out / "codex",
        out / "opencode",
        out / "grok",
        out / "pi",
    )
    claude_dir.mkdir(parents=True, exist_ok=True)
    codex_dir.mkdir(parents=True, exist_ok=True)
    oc_dir.mkdir(parents=True, exist_ok=True)
    grok_dir.mkdir(parents=True, exist_ok=True)
    pi_dir.mkdir(parents=True, exist_ok=True)

    count = 0
    for path in sorted(src.rglob("*.md")):
        agent = parse_agent(path)
        codex_text = to_codex(agent)
        verify_codex_round_trip(agent, codex_text)
        # Claude keeps the source layout: the role subfolders are the human
        # namespace of the tree /deploy-global copies into ~/.claude/agents/.
        claude_target = claude_dir / path.relative_to(src)
        claude_target.parent.mkdir(parents=True, exist_ok=True)
        claude_target.write_text(to_claude(agent), encoding="utf-8")
        (codex_dir / f"{agent['name']}.toml").write_text(
            codex_text, encoding="utf-8"
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
        f"converted {count} agents -> {claude_dir} , {codex_dir} , {oc_dir} , "
        f"{grok_dir} , {pi_dir}"
    )


if __name__ == "__main__":
    main()
