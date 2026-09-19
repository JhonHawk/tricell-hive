#!/usr/bin/env python3
"""Regression tests for the multi-harness agent converter."""
import contextlib
import importlib.util
import subprocess
import sys
import tempfile
import tomllib
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CONVERTER_PATH = ROOT / "harness" / "build" / "convert-agents.py"
SPEC = importlib.util.spec_from_file_location("convert_agents", CONVERTER_PATH)
CONVERTER = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(CONVERTER)


def source_agent(name):
    return CONVERTER.parse_agent(ROOT / "global" / "agents" / name)


def fixture_agent(frontmatter, body="Fixture body"):
    with tempfile.TemporaryDirectory(prefix="hive-agent-") as tmp:
        path = Path(tmp) / "fixture.md"
        path.write_text(f"---\n{frontmatter}\n---\n{body}\n", encoding="utf-8")
        return CONVERTER.parse_agent(path)


# A body shaped like a real agent's: its own sections plus the Role rules table
# whose rows a pack is meant to replace.
ROLE_RULES_BODY = """\
You are a fixture agent.

## Focus
- Something specific

## Role rules

Read the row matching what you touch.

| When | Read |
|---|---|
| TypeScript | `~/.claude/skills/language-rules/references/typescript-standards.md` |
| Naming fields | `~/.claude/skills/language-rules/references/identifier-language.md` |

## Output
- Working code
"""


@contextlib.contextmanager
def required_pack(value):
    """Swap the mandatory-pack constant.

    `agent-core-gates` is authored in a later milestone, so every fixture here
    would fail the requirement it does not exist to satisfy yet. Tests that are
    not ABOUT the requirement turn it off; the ones that are point it at a rule
    that exists today.
    """
    original = CONVERTER.REQUIRED_PACK
    CONVERTER.REQUIRED_PACK = value
    try:
        yield
    finally:
        CONVERTER.REQUIRED_PACK = original


def every_pack_name():
    """Every rule text a `packs:` entry can resolve to — one flat store."""
    names = {p.stem for p in (ROOT / "global" / "rules-situational").glob("*.md")}
    return sorted(names - {"README"})


def rendered_everywhere(agent):
    """The agent as every harness receives it, keyed by harness name."""
    return {
        "claude": CONVERTER.to_claude(agent),
        "codex": CONVERTER.to_codex(agent),
        "opencode": CONVERTER.to_opencode(agent),
        "grok": CONVERTER.to_grok(agent),
        "pi": CONVERTER.to_pi(agent),
    }


class ConverterTests(unittest.TestCase):
    def test_opus_judgment_roles_use_astra_medium_in_codex_and_pi(self):
        expected = {
            "review-code",
            "database-specialist",
            "performance-engineer",
            "sdd-product-critic",
            "prompt-engineer",
        }
        for name in expected:
            agent = CONVERTER.parse_agent(
                next((ROOT / "global" / "agents").rglob(f"{name}.md"))
            )
            codex = CONVERTER.to_codex(agent)
            pi = CONVERTER.to_pi(agent)
            self.assertIn('model = "gpt-6-astra"', codex, name)
            self.assertIn('model_reasoning_effort = "medium"', codex, name)
            self.assertIn("model: openai-codex/gpt-6-astra", pi, name)
            self.assertIn("thinking: medium", pi, name)

    def test_sonnet_and_inherit_effort_behavior_stays_unchanged(self):
        sonnet = CONVERTER.parse_agent(
            next((ROOT / "global" / "agents").rglob("backend-developer.md"))
        )
        self.assertEqual(CONVERTER.codex_model(sonnet), "gpt-5.6-luna")
        self.assertEqual(CONVERTER.codex_reasoning_effort(sonnet), "high")
        self.assertIn("model: openai-codex/gpt-5.6-luna", CONVERTER.to_pi(sonnet))
        self.assertIn("thinking: high", CONVERTER.to_pi(sonnet))

        inherited = fixture_agent(
            "name: inherited\n"
            "description: Inherit the coordinator model\n"
            "tools: Read\n"
            "model: inherit\n"
            "effort: high"
        )
        self.assertIsNone(CONVERTER.codex_model(inherited))
        self.assertEqual(CONVERTER.codex_reasoning_effort(inherited), "high")
        self.assertNotIn("model = ", CONVERTER.to_codex(inherited))
        self.assertIn("model: inherit", CONVERTER.to_pi(inherited))
        self.assertIn("thinking: high", CONVERTER.to_pi(inherited))

    def test_explicit_model_and_effort_are_preserved(self):
        explicit = fixture_agent(
            "name: explicit\n"
            "description: Explicit OpenAI model\n"
            "tools: Read\n"
            "model: gpt-6-astra\n"
            "effort: high"
        )
        self.assertEqual(CONVERTER.codex_model(explicit), "gpt-6-astra")
        self.assertEqual(CONVERTER.codex_reasoning_effort(explicit), "high")
        codex = CONVERTER.to_codex(explicit)
        pi = CONVERTER.to_pi(explicit)
        self.assertIn('model = "gpt-6-astra"', codex)
        self.assertIn('model_reasoning_effort = "high"', codex)
        self.assertIn("model: openai-codex/gpt-6-astra", pi)
        self.assertIn("thinking: high", pi)

    def test_cli_generates_all_26_pi_agents_and_common_policy(self):
        with tempfile.TemporaryDirectory(prefix="hive-agent-output-") as tmp:
            result = subprocess.run(
                [sys.executable, str(CONVERTER_PATH), "global/agents", tmp],
                cwd=ROOT,
                capture_output=True,
                text=True,
                check=True,
            )
            self.assertIn("converted 26 agents", result.stdout)
            pi_dir = Path(tmp) / "pi"
            files = sorted(pi_dir.glob("*.md"))
            self.assertEqual(len(files), 26)
            reviewer_guard_agents = {
                "review-code",
                "sdd-explore",
                "sdd-product-critic",
                "review-security",
                "sdd-spec-reviewer",
                "workspace-custodian",
            }
            for path in files:
                text = path.read_text(encoding="utf-8")
                self.assertIn(
                    "subagentOnlyExtensions: __HIVE_PI_ROOT__/extensions/hive-hooks.ts",
                    text,
                )
                self.assertIn("async: true", text)
                self.assertIn("defaultContext: fresh", text)
                self.assertIn("systemPromptMode: append", text)
                self.assertIn("inheritProjectContext: true", text)
                self.assertIn("inheritGlobalContext: true", text)
                self.assertIn("inheritSkills: true", text)
                self.assertIn("allowNestedSubagents: false", text)
                self.assertRegex(text, r"(?m)^tools: .*hive_hook_readiness")
                self.assertIn("memory:\n  scope: project", text)
                if path.stem in reviewer_guard_agents:
                    self.assertIn("hive/reviewer-guard.ts", text)
                    self.assertRegex(text, r"(?m)^tools: .*hive_reviewer_readiness")
                else:
                    self.assertNotIn("hive/reviewer-guard.ts", text)
                    self.assertNotIn("hive_reviewer_readiness", text)
            reviewer = CONVERTER.parse_agent(
                ROOT / "global" / "agents" / "review" / "review-code.md"
            )
            reviewer_text = (pi_dir / "review-code.md").read_text(encoding="utf-8")
            reviewer_tools = reviewer_text.split("tools: ", 1)[1].split("\n", 1)[0]
            self.assertEqual(reviewer_tools.split(", ").count("mcp"), 1)
            self.assertNotIn("mcp:", reviewer_tools)
            self.assertIn(
                "mcp({tool:'context7_resolve-library-id',args:{query,libraryName}})",
                reviewer_text,
            )
            self.assertIn(
                "mcp({tool:'context7_query-docs',args:{libraryId,query}})",
                reviewer_text,
            )
            self.assertIn("mem_search, mem_context, mem_get_observation", reviewer_text)
            self.assertIn("contact_supervisor", reviewer_text)
            self.assertIn("hive_git_read", reviewer_text)
            self.assertIn("hive_hook_readiness, hive_reviewer_readiness", reviewer_text)
            self.assertNotIn("requiredTools:", reviewer_text)
            self.assertIn(
                "hive_hook_readiness, hive_reviewer_readiness",
                reviewer_text.split("tools: ", 1)[1].split("\n", 1)[0],
            )
            self.assertFalse(CONVERTER.can_write(reviewer))

            writer = CONVERTER.parse_agent(
                ROOT / "global" / "agents" / "development" / "database-specialist.md"
            )
            writer_text = (pi_dir / "database-specialist.md").read_text(encoding="utf-8")
            self.assertTrue(CONVERTER.can_write(writer))
            self.assertIn("mem_save", writer_text)
            self.assertIn("contact_supervisor", writer_text)
            self.assertIn("hive_hook_readiness", writer_text)
            self.assertNotIn("mem_search", writer_text)

            inherited = CONVERTER.parse_agent(
                ROOT / "global" / "agents" / "quality" / "state-fetcher.md"
            )
            inherited_text = (pi_dir / "state-fetcher.md").read_text(encoding="utf-8")
            self.assertIsNone(inherited["tools"])
            self.assertIn("tools: read, bash, find, ls, grep", inherited_text)
            self.assertIn("hive_hook_readiness", inherited_text)
            self.assertNotIn("write, edit", inherited_text.split("tools: ", 1)[1].split("\n", 1)[0])
            self.assertNotIn("subagent", inherited_text.split("tools: ", 1)[1].split("\n", 1)[0])

    def test_pi_tool_allowlist_and_denylist_preserve_deny_wins(self):
        agent = fixture_agent(
            "name: tools-fixture\n"
            "description: Tool conversion fixture\n"
            "tools: Read, Write, Edit, WebSearch, mcp__context7__query-docs\n"
            "disallowedTools: Write, mcp__context7__resolve-library-id"
        )
        pi = CONVERTER.to_pi(agent)
        self.assertIn(
            "tools: read, write, edit, fetch_content, get_search_content, web_search, "
            "source_check, mcp, "
            "mem_save, contact_supervisor",
            pi,
        )
        self.assertIn("excludeTools: write, mcp", pi)
        self.assertIn("subagentOnlyExtensions: __HIVE_PI_ROOT__/extensions/hive-hooks.ts", pi)
        self.assertNotIn("mcp:", pi)

    def test_pi_research_capabilities_expand_and_add_non_blocking_readiness(self):
        documentation = fixture_agent(
            "name: documentation-research\n"
            "description: Documentation research fixture\n"
            "tools: Read, WebFetch\n"
        )
        documentation_pi = CONVERTER.to_pi(documentation)
        self.assertIn(
            "tools: read, fetch_content, get_search_content, "
            "mem_search, mem_context, mem_get_observation, contact_supervisor, "
            "hive_hook_readiness, hive_research_readiness",
            documentation_pi,
        )
        self.assertIn("profile `documentation`", documentation_pi)
        self.assertNotIn("web_search, source_check", documentation_pi.split("tools: ", 1)[1].split("\n", 1)[0])

        web = fixture_agent(
            "name: web-research\n"
            "description: Web research fixture\n"
            "tools: Read, WebSearch\n"
        )
        web_pi = CONVERTER.to_pi(web)
        self.assertIn(
            "tools: read, fetch_content, get_search_content, web_search, source_check, "
            "mem_search, mem_context, mem_get_observation, contact_supervisor, "
            "hive_hook_readiness, hive_research_readiness",
            web_pi,
        )
        self.assertIn("profile `web`", web_pi)

        local = fixture_agent(
            "name: local-only\n"
            "description: Local-only fixture\n"
            "tools: Read, Bash\n"
        )
        local_pi = CONVERTER.to_pi(local)
        self.assertNotIn("hive_research_readiness", local_pi)
        self.assertNotIn("PI research readiness", local_pi)

        inherited = fixture_agent(
            "name: inherited-research\n"
            "description: Inherited research fixture\n"
        )
        inherited_pi = CONVERTER.to_pi(inherited)
        inherited_tools = inherited_pi.split("tools: ", 1)[1].split("\n", 1)[0]
        self.assertIn("web_search, fetch_content, get_search_content, source_check", inherited_tools)
        self.assertIn("hive_research_readiness", inherited_pi)
        self.assertIn("profile `web`", inherited_pi)

    def test_pi_research_expanded_denies_win_for_explicit_and_inherited_roles(self):
        explicit = fixture_agent(
            "name: denied-web\n"
            "description: Denied web research fixture\n"
            "tools: Read, WebSearch\n"
            "disallowedTools: WebSearch\n"
        )
        explicit_pi = CONVERTER.to_pi(explicit)
        self.assertIn(
            "excludeTools: fetch_content, get_search_content, web_search, source_check",
            explicit_pi,
        )
        self.assertIn("hive_research_readiness", explicit_pi)

        inherited = fixture_agent(
            "name: denied-inherited\n"
            "description: Denied inherited research fixture\n"
            "disallowedTools: WebFetch\n"
        )
        inherited_pi = CONVERTER.to_pi(inherited)
        tools_line = inherited_pi.split("tools: ", 1)[1].split("\n", 1)[0]
        self.assertNotIn("fetch_content", tools_line)
        self.assertNotIn("get_search_content", tools_line)
        self.assertIn("web_search, source_check", tools_line)
        self.assertIn("hive_research_readiness", inherited_pi)

    def test_pi_context7_agents_use_one_guarded_mcp_gateway(self):
        expected_agents = {
            "review-code",
            "sdd-explore",
            "review-refuter",
            "sdd-product-critic",
            "review-security",
            "sdd-spec-reviewer",
        }
        with tempfile.TemporaryDirectory(prefix="hive-agent-output-") as tmp:
            subprocess.run(
                [sys.executable, str(CONVERTER_PATH), "global/agents", tmp],
                cwd=ROOT,
                capture_output=True,
                text=True,
                check=True,
            )
            pi_dir = Path(tmp) / "pi"
            for name in expected_agents:
                text = (pi_dir / f"{name}.md").read_text(encoding="utf-8")
                tools_line = text.split("tools: ", 1)[1].split("\n", 1)[0]
                self.assertEqual(tools_line.split(", ").count("mcp"), 1, name)
                self.assertNotIn("mcp:", tools_line, name)
                self.assertIn(
                    "mcp({tool:'context7_resolve-library-id',args:{query,libraryName}})",
                    text,
                    name,
                )
                self.assertIn(
                    "mcp({tool:'context7_query-docs',args:{libraryId,query}})",
                    text,
                    name,
                )

    def test_pi_structured_git_read_follows_bash_for_non_guarded_refuter(self):
        refuter = source_agent(Path("review") / "review-refuter.md")
        self.assertTrue(CONVERTER.has_bash(refuter))
        self.assertFalse(refuter["has_reviewer_guard"])
        pi = CONVERTER.to_pi(refuter)
        self.assertIn("hive_git_read", pi)
        self.assertNotIn("hive/reviewer-guard.ts", pi)

        denied = fixture_agent(
            "name: denied-git\n"
            "description: Bash-denied read-only agent\n"
            "tools: Read, Glob, Grep, Bash\n"
            "disallowedTools: Bash\n"
            "model: inherit\n"
            "permissionMode: plan"
        )
        self.assertFalse(CONVERTER.has_bash(denied))
        self.assertNotIn("hive_git_read", CONVERTER.to_pi(denied))

        denied_writes = fixture_agent(
            "name: denied-writes\n"
            "description: Read-only agent with denied write tools\n"
            "tools: Read, Write, Edit, Bash\n"
            "disallowedTools: Write, Edit"
        )
        self.assertFalse(CONVERTER.can_write(denied_writes))
        denied_writes_pi = CONVERTER.to_pi(denied_writes)
        self.assertIn("mem_search, mem_context, mem_get_observation", denied_writes_pi)
        self.assertNotIn("mem_save", denied_writes_pi)

    def test_pi_rebases_role_rule_paths_to_shared_skill_root(self):
        agent = source_agent(Path("review") / "review-code.md")
        pi = CONVERTER.to_pi(agent)
        self.assertIn("~/.agents/skills/language-rules/references/", pi)
        self.assertNotIn("~/.claude/skills/language-rules/references/", pi)

    def test_grok_keeps_every_role_rule_row_no_pack_covers(self):
        # Grok used to load the always-on store through flat symlinks into
        # ~/.grok/rules, so rows pointing at those files were dropped as
        # duplicates. That store is gone and so are the symlinks: Grok now
        # reaches a rule through the hook or the router, exactly like Codex —
        # a dropped row is a rule the agent is never told about.
        agent = source_agent(Path("development") / "backend-developer.md")
        grok = CONVERTER.to_grok(agent)
        claude = CONVERTER.to_claude(agent)
        rows = CONVERTER.ROLE_RULE_TARGET.findall
        self.assertEqual(rows(grok), rows(claude))
        for still_listed in ("testing.md", "development-principles.md",
                             "debugging.md", "java-kotlin.md"):
            self.assertIn(f"language-rules/references/{still_listed}", grok,
                          still_listed)

    def test_the_native_always_on_filter_is_gone(self):
        # A filter whose input directory no longer exists returns "nothing is
        # native" forever — protection nobody provides. It was deleted with
        # the store it read.
        self.assertFalse(hasattr(CONVERTER, "native_always_on_rules"))
        self.assertFalse(hasattr(CONVERTER, "_rule_is_always_on"))
        self.assertFalse(hasattr(CONVERTER, "RULES_SITUATIONAL_DIR"))
        self.assertEqual(CONVERTER.RULES_DIR.name, "rules-situational")

    def test_no_harness_loads_a_rule_natively_any_more(self):
        # Claude Code's path-scoping went first, Grok's flat symlinks with M4:
        # every harness now receives a rule through the hook, a router
        # reference or a pack, so every Role rules row survives everywhere.
        agent = source_agent(Path("review") / "review-code.md")
        for rendered in (
            CONVERTER.to_codex(agent),
            CONVERTER.to_opencode(agent),
            CONVERTER.to_pi(agent),
        ):
            self.assertIn("~/.agents/skills/language-rules/references/testing.md", rendered)
        for rendered in (CONVERTER.to_grok(agent), CONVERTER.to_claude(agent)):
            self.assertIn("~/.claude/skills/language-rules/references/testing.md", rendered)
            self.assertIn("## Role rules", rendered)


class PackInliningTests(unittest.TestCase):
    """`packs:` is the one rule-delivery mechanism every harness shares."""

    def setUp(self):
        self._required = required_pack(None)
        self._required.__enter__()
        self.addCleanup(lambda: self._required.__exit__(None, None, None))

    def packed(self, packs="typescript-standards, git-mechanics", extra=""):
        return fixture_agent(
            "name: packed-fixture\n"
            "description: Fixture agent carrying packs\n"
            "tools: Read, Write, Edit\n"
            f"packs: {packs}\n" + extra,
            body=ROLE_RULES_BODY,
        )

    def test_every_pack_is_labelled_with_its_rule_name(self):
        # A rule text is not obliged to open with a title of its own
        # (`ui-visual-design` opens with a blockquote and section headings), so
        # without a label its sections read as a continuation of the pack above.
        rendered = rendered_everywhere(self.packed())
        for harness, text in rendered.items():
            for pack, heading in (
                ("typescript-standards", "## TypeScript & JS Standards"),
                ("git-mechanics", "## Git Mechanics"),
            ):
                label = f"**Rule `{pack}` — carried in full below:**"
                self.assertEqual(text.count(label), 1, (harness, pack))
                self.assertLess(text.index(label), text.index(heading), (harness, pack))

    def test_packs_parse_as_a_declared_ordered_list(self):
        agent = self.packed()
        self.assertEqual(agent["packs"], ["typescript-standards", "git-mechanics"])
        self.assertEqual(fixture_agent("name: bare\ndescription: No packs")["packs"], [])

    def test_every_harness_inlines_each_pack_text_exactly_once(self):
        rendered = rendered_everywhere(self.packed())
        for harness, text in rendered.items():
            self.assertIn("## Carried rules", text, harness)
            self.assertEqual(text.count("## TypeScript & JS Standards"), 1, harness)
            self.assertEqual(
                text.count("## Git Mechanics — branching, commits, PRs"), 1, harness
            )
            # Declared order, not alphabetical or filesystem order.
            self.assertLess(
                text.index("## TypeScript & JS Standards"),
                text.index("## Git Mechanics"),
                harness,
            )
            # The agent's own sections come first; the carried rules follow.
            self.assertLess(text.index("## Focus"), text.index("## Carried rules"),
                            harness)
            # Frontmatter of the rule file never reaches the agent body.
            self.assertNotIn("paths:\n  - \"**/*.{ts,tsx,js,jsx}\"", text, harness)
            # The carried text is complete — the agent must not go hunting.
            self.assertIn("already loaded", text, harness)

    def test_a_packed_agent_loses_the_role_rule_rows_its_packs_cover(self):
        rendered = rendered_everywhere(self.packed(packs="typescript-standards"))
        for harness, text in rendered.items():
            self.assertNotIn(
                "language-rules/references/typescript-standards.md", text, harness
            )
            self.assertIn("## Role rules", text, harness)
            self.assertIn(
                "language-rules/references/identifier-language.md", text, harness
            )

    def test_the_role_rules_section_goes_when_packs_cover_every_row(self):
        rendered = rendered_everywhere(
            self.packed(packs="typescript-standards, identifier-language")
        )
        for harness, text in rendered.items():
            self.assertNotIn("## Role rules", text, harness)
            self.assertNotIn("| When | Read |", text, harness)
            self.assertIn("## Output", text, harness)

    def test_an_unknown_pack_fails_the_build_naming_the_agent_and_the_pack(self):
        with self.assertRaises(ValueError) as raised:
            self.packed(packs="typescript-standards, no-such-pack")
        message = str(raised.exception)
        self.assertIn("packed-fixture", message)
        self.assertIn("no-such-pack", message)

    def test_packs_never_leak_into_a_generated_frontmatter(self):
        for harness, text in rendered_everywhere(self.packed()).items():
            self.assertNotIn("packs:", text, harness)

    def test_claude_output_omits_the_global_corpus_only_for_packed_agents(self):
        packed = CONVERTER.to_claude(self.packed())
        self.assertIn("omitClaudeMd: true", packed)
        self.assertIn("name: packed-fixture", packed)
        self.assertIn("tools: Read, Write, Edit", packed)
        self.assertIn(CONVERTER.GENERATED_NOTE, packed)

        already_set = CONVERTER.to_claude(
            self.packed(extra="omitClaudeMd: true\n")
        )
        self.assertEqual(already_set.count("omitClaudeMd: true"), 1)

        bare = CONVERTER.to_claude(source_agent(Path("development") / "backend-developer.md"))
        self.assertNotIn("omitClaudeMd", bare)

    def test_claude_output_of_an_unpacked_agent_keeps_the_source_body(self):
        for relative in (
            Path("development") / "backend-developer.md",
            Path("review") / "review-code.md",
        ):
            agent = source_agent(relative)
            rendered = CONVERTER.to_claude(agent)
            body = rendered.split("\n---\n", 1)[1].strip()
            self.assertEqual(body, agent["body"], relative.as_posix())
            # Claude Code keeps native path-scoping, so its Role rules rows stay.
            self.assertIn("~/.claude/skills/", rendered, relative.as_posix())

    def test_every_pack_text_round_trips_through_the_codex_toml_string(self):
        """B1: rule texts carry backslashes; a TOML basic string eats them.

        `infra-naming` holds `development\\|qa\\|production` (invalid TOML) and
        `typescript-standards` holds a literal `\\n` inside backticks (silently
        turned into a real newline). Neither is visible in a rendered diff.
        """
        for pack in every_pack_name():
            agent = self.packed(packs=pack)
            rendered = CONVERTER.to_codex(agent)
            with self.subTest(pack=pack):
                parsed = tomllib.loads(rendered)["developer_instructions"]
                self.assertIn(agent["pack_texts"][0][1], parsed)
                # What TOML reads back is exactly what the converter composed:
                # the rebased agent body plus the carried section, nothing that
                # the string encoding added, dropped, or reinterpreted.
                expected = CONVERTER.packed_body(
                    agent, rebase=CONVERTER.rebase_skill_root
                )
                # A prefix, not equality: to_codex appends its own
                # compatibility instructions after the body.
                self.assertTrue(
                    parsed.strip("\n").startswith(expected.strip("\n")),
                    f"{pack}: the TOML round-trip altered the composed body",
                )

    # Text a rule could plausibly hold, and text that simply must not corrupt
    # the output. Today's rule corpus proves nothing about tomorrow's.
    HOSTILE_BODIES = {
        "trailing-quote": 'A line ending in a quote"',
        "trailing-backslash": "A line ending in a backslash\\",
        "escaped-triple": 'Literal \\""" inside prose',
        "quote-runs": "\n".join(f'{n} quotes: ' + '"' * n for n in range(3, 8)),
        "leading-newline": "\nStarts with a blank line",
        "crlf": "A Windows line\r\nand the next one",
        "lone-cr": "A lone carriage return\rmid-text",
        "unicode": "Acentos, ñ, — em dash, 中文, emoji 🙂",
        "backslash-n": "A literal `\\n` inside backticks",
        "pipe-escape": "A cell with `development\\|qa\\|production`",
        **{f"control-{ord(c):02x}": f"before{c}after"
           for c in "\x00\x01\x08\x0b\x0c\x1b\x1f\x7f"},
    }

    def test_hostile_bodies_round_trip_through_the_codex_toml_string(self):
        """R1: escaping must be total, not just good enough for today's corpus.

        A TOML basic string forbids raw control characters outright, so text
        carrying one produces a file `tomllib` refuses — and nothing noticed,
        because the build never parsed what it wrote.
        """
        for label, hostile in self.HOSTILE_BODIES.items():
            agent = fixture_agent(
                "name: hostile\ndescription: Fixture\ntools: Read, Write",
                body=hostile,
            )
            with self.subTest(case=label):
                parsed = tomllib.loads(CONVERTER.to_codex(agent))
                self.assertEqual(
                    parsed["developer_instructions"],
                    CONVERTER.codex_body(agent) + "\n",
                )

    def test_the_build_parses_back_every_codex_agent_it_writes(self):
        """R1b: a deterministic backstop, not a promise about the escaper."""
        for name in ("backend-developer", "review-code"):
            agent = CONVERTER.parse_agent(
                next((ROOT / "global" / "agents").rglob(f"{name}.md"))
            )
            CONVERTER.verify_codex_round_trip(agent, CONVERTER.to_codex(agent))

        agent = fixture_agent("name: broken\ndescription: Fixture", body="body")
        with self.assertRaises(ValueError) as raised:
            CONVERTER.verify_codex_round_trip(agent, 'developer_instructions = "x"\n')
        self.assertIn("broken", str(raised.exception))

        with self.assertRaises(ValueError) as raised:
            CONVERTER.verify_codex_round_trip(agent, 'developer_instructions = """\n\\q\n"""\n')
        self.assertIn("broken", str(raised.exception))

    def test_every_pack_text_survives_the_markdown_harnesses_verbatim(self):
        """The other four embed the body raw — proven, not assumed."""
        for pack in every_pack_name():
            agent = self.packed(packs=pack)
            text = agent["pack_texts"][0][1]
            with self.subTest(pack=pack):
                for harness in ("claude", "opencode", "grok", "pi"):
                    rendered = rendered_everywhere(agent)[harness]
                    self.assertIn(text, rendered, harness)

    def test_a_pack_name_outside_the_basename_alphabet_is_refused(self):
        # `../../CLAUDE.local` would inline the owner's private gitignored file
        # into all five harnesses; `languages/tailwind` and `*` also resolve
        # through rglob, and `/etc/hosts` used to die on a raw NotImplementedError.
        for pack in ("../../CLAUDE.local", "languages/tailwind", "*",
                     "/etc/hosts", "Tailwind", "with_underscore"):
            with self.subTest(pack=pack):
                with self.assertRaises(ValueError) as raised:
                    self.packed(packs=pack)
                message = str(raised.exception)
                self.assertIn("packed-fixture", message)
                self.assertIn(pack, message)

    def test_a_pack_whose_text_is_empty_is_refused(self):
        with tempfile.TemporaryDirectory(prefix="hive-empty-pack-") as tmp:
            store = Path(tmp) / "rules-situational"
            store.mkdir()
            (store / "hollow.md").write_text("---\nglobs:\n  - \"*\"\n---\n\n",
                                             encoding="utf-8")
            with self.patched_rule_dir(store):
                with self.assertRaises(ValueError) as raised:
                    self.packed(packs="hollow")
            self.assertIn("hollow", str(raised.exception))
            self.assertIn("empty", str(raised.exception))

    def test_a_pack_declared_twice_is_refused(self):
        with self.assertRaises(ValueError) as raised:
            self.packed(packs="typescript-standards, git-mechanics, typescript-standards")
        message = str(raised.exception)
        self.assertIn("packed-fixture", message)
        self.assertIn("typescript-standards", message)

    def test_a_yaml_list_or_an_empty_packs_key_asks_for_the_csv_form(self):
        # `-leading` belongs here, not with the bad names: a value opening on
        # a dash is indistinguishable from a YAML list, and that is the more
        # useful thing to tell the author.
        for raw in ("packs:\n  - typescript-standards\n", "packs:\n",
                    "packs: [typescript-standards]\n", "packs: -leading\n"):
            with self.subTest(raw=raw):
                with self.assertRaises(ValueError) as raised:
                    fixture_agent(
                        "name: listy\ndescription: Fixture\n" + raw.rstrip("\n"),
                        body=ROLE_RULES_BODY,
                    )
                self.assertIn("comma-separated", str(raised.exception))

    def test_a_pack_may_not_escape_the_rule_stores_through_a_symlink(self):
        # A name that passes the alphabet check still reaches the filesystem;
        # a link inside the store is enough to inline anything on the machine.
        with tempfile.TemporaryDirectory(prefix="hive-link-pack-") as tmp:
            outside = Path(tmp) / "outside" / "private.md"
            outside.parent.mkdir(parents=True)
            outside.write_text("PRIVATE NOTES\n", encoding="utf-8")
            store = Path(tmp) / "rules-situational"
            store.mkdir()
            (store / "escapee.md").symlink_to(outside)
            with self.patched_rule_dir(store):
                with self.assertRaises(ValueError) as raised:
                    self.packed(packs="escapee")
            message = str(raised.exception)
            self.assertIn("packed-fixture", message)
            self.assertIn("escapee", message)
            self.assertNotIn("PRIVATE NOTES", message)

    def test_a_symlinked_rule_store_is_refused_before_any_pack_resolves(self):
        # Same vacuity as the core-assembly include check: the escape guard
        # compares against the RESOLVED store, so a symlinked store resolves
        # to wherever it points and then confirms every file under it is
        # "inside". A hard link inside a real store stays fine — that is a
        # copy in the store, not an escape from it.
        with tempfile.TemporaryDirectory(prefix="hive-store-link-") as tmp:
            outside = Path(tmp) / "private-notes"
            outside.mkdir()
            (outside / "leak.md").write_text("PRIVATE NOTES\n", encoding="utf-8")
            store = Path(tmp) / "rules-situational"
            store.symlink_to(outside, target_is_directory=True)
            with self.patched_rule_dir(store):
                with self.assertRaises(ValueError) as raised:
                    self.packed(packs="leak")
            message = str(raised.exception)
            self.assertIn("not a real directory", message)
            self.assertNotIn("PRIVATE NOTES", message)

    def test_a_directory_named_like_a_rule_is_not_a_pack(self):
        with tempfile.TemporaryDirectory(prefix="hive-dir-pack-") as tmp:
            store = Path(tmp) / "rules-situational"
            (store / "folder.md").mkdir(parents=True)
            with self.patched_rule_dir(store):
                with self.assertRaises(ValueError) as raised:
                    self.packed(packs="folder")
            self.assertIn("unknown pack 'folder'", str(raised.exception))

    def test_only_ascii_spacing_is_stripped_around_a_pack_name(self):
        # A NBSP reads as part of the name, not as padding: a name that "looks
        # right" but resolves to nothing must say so instead of being cleaned up.
        # (A stray CR cannot reach here — `read_text` normalizes newlines when
        # the agent file is read, so parse_agent never sees one.)
        with self.assertRaises(ValueError) as raised:
            self.packed(packs="typescript-standards, git-mechanics")
        self.assertIn("packed-fixture", str(raised.exception))

        # Called directly, a CR is refused too — the stripping is ASCII-only by
        # construction, not by accident of how the file was read.
        with required_pack(None):
            with self.assertRaises(ValueError):
                CONVERTER.parse_packs(
                    "direct", "typescript-standards\r", "packs: typescript-standards\r"
                )

    def test_a_second_packs_key_is_an_error_not_a_silent_winner(self):
        with self.assertRaises(ValueError) as raised:
            fixture_agent(
                "name: doubled-key\ndescription: Fixture\n"
                "packs: typescript-standards\npacks: git-mechanics",
                body=ROLE_RULES_BODY,
            )
        message = str(raised.exception)
        self.assertIn("doubled-key", message)
        self.assertIn("packs", message)

    # Every way of writing a `packs` key that is NOT the one canonical line.
    # Each one used to exit 0: cases 1-3 silently produced an agent with no
    # packs (and so no required-gates check), while the block scalars inlined
    # the packs and then orphaned their continuation line under the PREVIOUS
    # key — `tools: Read\n  agent-core-gates` reads as one tool allowlist.
    NON_CANONICAL_PACKS = {
        "capitalized": "Packs: typescript-standards",
        "shouted": "PACKS: typescript-standards",
        "quoted-key": '"packs": typescript-standards',
        "single-quoted-key": "'packs': typescript-standards",
        "folded-scalar": "packs: >\n  typescript-standards",
        "literal-scalar": "packs: |\n  typescript-standards",
        "trailing-comment": "packs: typescript-standards  # the TS pack",
        "quoted-value": 'packs: "typescript-standards"',
        "indented-key": "  packs: typescript-standards",
    }

    def test_only_the_one_canonical_packs_line_is_accepted(self):
        for label, line in self.NON_CANONICAL_PACKS.items():
            with self.subTest(case=label):
                with self.assertRaises(ValueError) as raised:
                    fixture_agent(
                        "name: odd-packs\ndescription: Fixture\n"
                        f"tools: Read\n{line}",
                        body=ROLE_RULES_BODY,
                    )
                message = str(raised.exception)
                self.assertIn("odd-packs", message)
                self.assertIn("packs: a, b", message)

    def test_the_word_packs_in_prose_or_nested_yaml_is_not_a_declaration(self):
        # The validation reads KEYS, not text: a block scalar's continuation
        # lines are prose, and nested YAML belongs to its own key.
        agent = fixture_agent(
            "name: innocent\n"
            "description: >\n"
            "  Mentions packs: typescript-standards in its routing prose, and\n"
            "  the word packs again for good measure.\n"
            "tools: Read, Bash\n"
            "hooks:\n"
            "  PreToolUse:\n"
            "    - matcher: \"Bash\"\n"
            "      hooks:\n"
            "        - type: command\n"
            "          command: \"$HOME/.claude/hooks/reviewer-guard.sh\"\n",
            body=ROLE_RULES_BODY,
        )
        self.assertEqual(agent["packs"], [])
        self.assertNotIn("omitClaudeMd", CONVERTER.to_claude(agent))

    def test_the_generated_claude_frontmatter_stays_parseable(self):
        """Removing the canonical line cannot orphan anything, by construction."""
        with required_pack(None):
            agent = self.packed(packs="typescript-standards")
        rendered = CONVERTER.to_claude(agent)
        frontmatter = rendered.split("---\n", 2)[1]
        try:
            import yaml
        except ImportError:
            yaml = None
        if yaml is not None:
            parsed = yaml.safe_load(frontmatter)
            self.assertEqual(parsed["tools"], "Read, Write, Edit")
            self.assertTrue(parsed["omitClaudeMd"])
            self.assertNotIn("packs", parsed)
        else:
            expected = [f"# {CONVERTER.GENERATED_NOTE}"] + [
                line for line in agent["frontmatter"].splitlines()
                if not line.startswith("packs:")
            ] + ["omitClaudeMd: true"]
            self.assertEqual(frontmatter.splitlines(), expected)

    def test_a_packs_key_in_any_other_spelling_is_an_error(self):
        # `packs : ts` parsed as NO packs and then survived verbatim into the
        # Claude frontmatter — a hive-only key shipped to Claude Code.
        for line in ("packs : typescript-standards", "packs\t: typescript-standards"):
            with self.subTest(line=line):
                with self.assertRaises(ValueError) as raised:
                    fixture_agent(
                        f"name: oddly-spelled\ndescription: Fixture\n{line}",
                        body=ROLE_RULES_BODY,
                    )
                self.assertIn("oddly-spelled", str(raised.exception))

    def patched_rule_dir(self, store):
        @contextlib.contextmanager
        def patch():
            original = CONVERTER.RULES_DIR
            CONVERTER.RULES_DIR = store
            try:
                yield
            finally:
                CONVERTER.RULES_DIR = original
        return patch()

    def test_cli_mirrors_the_source_role_folders_into_the_claude_tree(self):
        with tempfile.TemporaryDirectory(prefix="hive-agent-output-") as tmp:
            subprocess.run(
                [sys.executable, str(CONVERTER_PATH), "global/agents", tmp],
                cwd=ROOT, capture_output=True, text=True, check=True,
            )
            claude_dir = Path(tmp) / "claude"
            generated = sorted(
                p.relative_to(claude_dir).as_posix() for p in claude_dir.rglob("*.md")
            )
            source = sorted(
                p.relative_to(ROOT / "global" / "agents").as_posix()
                for p in (ROOT / "global" / "agents").rglob("*.md")
            )
            self.assertEqual(generated, source)


class RequiredPackTests(unittest.TestCase):
    """`omitClaudeMd: true` drops the always-on gates — something must replace them."""

    def declare(self, packs):
        return fixture_agent(
            f"name: gated-fixture\ndescription: Fixture\npacks: {packs}",
            body=ROLE_RULES_BODY,
        )

    def test_a_packed_agent_without_the_core_gates_pack_fails_the_build(self):
        with required_pack("security"):
            with self.assertRaises(ValueError) as raised:
                self.declare("typescript-standards")
            message = str(raised.exception)
            self.assertIn("gated-fixture", message)
            self.assertIn("security", message)

            # Present -> builds, and its text is carried like any other pack.
            agent = self.declare("security, typescript-standards")
            self.assertIn("security", agent["packs"])
            self.assertIn("## Security", CONVERTER.to_claude(agent))

    def test_an_agent_with_no_packs_is_untouched_by_the_requirement(self):
        with required_pack("security"):
            agent = fixture_agent("name: plain\ndescription: Fixture")
            self.assertEqual(agent["packs"], [])
            self.assertNotIn("omitClaudeMd", CONVERTER.to_claude(agent))

    def test_the_production_requirement_names_the_core_gates_pack(self):
        self.assertEqual(CONVERTER.REQUIRED_PACK, "agent-core-gates")


if __name__ == "__main__":
    unittest.main()
