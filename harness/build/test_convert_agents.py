#!/usr/bin/env python3
"""Regression tests for the multi-harness agent converter."""
import importlib.util
import subprocess
import sys
import tempfile
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


def fixture_agent(frontmatter):
    with tempfile.TemporaryDirectory(prefix="hive-agent-") as tmp:
        path = Path(tmp) / "fixture.md"
        path.write_text(f"---\n{frontmatter}\n---\nFixture body\n", encoding="utf-8")
        return CONVERTER.parse_agent(path)


class ConverterTests(unittest.TestCase):
    def test_opus_judgment_roles_use_astra_medium_in_codex_and_pi(self):
        expected = {
            "code-reviewer",
            "database-specialist",
            "performance-engineer",
            "product-critic",
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
        self.assertEqual(CONVERTER.codex_reasoning_effort(sonnet), "max")
        self.assertIn("model: openai-codex/gpt-5.6-luna", CONVERTER.to_pi(sonnet))
        self.assertIn("thinking: max", CONVERTER.to_pi(sonnet))

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

    def test_cli_generates_all_25_pi_agents_and_common_policy(self):
        with tempfile.TemporaryDirectory(prefix="hive-agent-output-") as tmp:
            result = subprocess.run(
                [sys.executable, str(CONVERTER_PATH), "global/agents", tmp],
                cwd=ROOT,
                capture_output=True,
                text=True,
                check=True,
            )
            self.assertIn("converted 25 agents", result.stdout)
            pi_dir = Path(tmp) / "pi"
            files = sorted(pi_dir.glob("*.md"))
            self.assertEqual(len(files), 25)
            reviewer_guard_agents = {
                "code-reviewer",
                "code-scout",
                "product-critic",
                "security-reviewer",
                "spec-quality-reviewer",
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
                ROOT / "global" / "agents" / "review" / "code-reviewer.md"
            )
            reviewer_text = (pi_dir / "code-reviewer.md").read_text(encoding="utf-8")
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
            "tools: read, write, edit, web_search, mcp, "
            "mem_save, contact_supervisor",
            pi,
        )
        self.assertIn("excludeTools: write, mcp", pi)
        self.assertIn("subagentOnlyExtensions: __HIVE_PI_ROOT__/extensions/hive-hooks.ts", pi)
        self.assertNotIn("mcp:", pi)

    def test_pi_context7_agents_use_one_guarded_mcp_gateway(self):
        expected_agents = {
            "code-reviewer",
            "code-scout",
            "finding-refuter",
            "product-critic",
            "security-reviewer",
            "spec-quality-reviewer",
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
        refuter = source_agent(Path("review") / "finding-refuter.md")
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
        agent = source_agent(Path("review") / "code-reviewer.md")
        pi = CONVERTER.to_pi(agent)
        self.assertIn("~/.agents/skills/language-rules/references/", pi)
        self.assertNotIn("~/.claude/skills/language-rules/references/", pi)


if __name__ == "__main__":
    unittest.main()
