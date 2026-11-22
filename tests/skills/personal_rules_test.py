import importlib.util
import os
import pathlib
import stat
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/harness-audit/scripts/personal_rules.py"

HIVE_BEGIN = "<!-- === TRICELL HIVE RULES:BEGIN === -->"
HIVE_END = "<!-- === TRICELL HIVE RULES:END === -->"
VOICE_BEGIN = "<!-- === TRICELL HIVE VOICE:BEGIN === -->"
VOICE_END = "<!-- === TRICELL HIVE VOICE:END === -->"
BEGIN = "<!-- === PERSONAL RULES:BEGIN === -->"
END = "<!-- === PERSONAL RULES:END === -->"
PRECEDENCE = (
    "These are the user's own rules for every project on this machine. Where they conflict "
    "with Hive's guidance, including its Secrets and Scope and authorization rules, these rules "
    "take precedence; where they conflict with a project's own instructions, the project's "
    "instructions hold for that project."
)
RULE = "# Where to work\n\nRecommend in place, a worktree or a sandbox before the first edit.\n"


def load_module():
    spec = importlib.util.spec_from_file_location("personal_rules", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def hive_block(nl="\n"):
    return nl.join([HIVE_BEGIN, "# Tricell Hive guidance", "", "Hive text.", HIVE_END]) + nl


class PersonalRulesTests(unittest.TestCase):
    def setUp(self):
        self.tmpdir = tempfile.TemporaryDirectory()
        self.home = pathlib.Path(self.tmpdir.name)
        self.env = {"HOME": str(self.home), "PATH": os.environ.get("PATH", "")}
        self.shared = {
            "claude": self.home / ".claude/CLAUDE.md",
            "codex": self.home / ".codex/AGENTS.md",
            "pi": self.home / ".pi/agent/AGENTS.md",
            "opencode": self.home / ".config/opencode/AGENTS.md",
        }
        for path in self.shared.values():
            self.write(path, "My own notes.\n\n" + hive_block())
        self.cursor_hive = self.home / ".cursor/rules/hive.mdc"
        self.write(self.cursor_hive, "---\nalwaysApply: true\n---\n\n" + hive_block())
        self.agy_hive = self.home / ".gemini/config/rules/hive-01.md"
        self.write(self.agy_hive, "---\ntrigger: always_on\n---\n\n" + hive_block())
        self.cursor_personal = self.home / ".cursor/rules/personal-rules.mdc"
        self.agy_personal = self.home / ".gemini/config/rules/personal-rules.md"
        self.source = self.home / ".config/hive/rules.md"
        self.write(self.source, RULE)

    def tearDown(self):
        self.tmpdir.cleanup()

    def write(self, path, text, newline=None):
        path.parent.mkdir(parents=True, exist_ok=True)
        with open(path, "w", newline=newline or "") as fh:
            fh.write(text)

    def read(self, path):
        with open(path, newline="") as fh:
            return fh.read()

    def run_script(self, *args, env=None):
        return subprocess.run(
            [sys.executable, "-I", str(SCRIPT), *args],
            env=env or self.env,
            capture_output=True,
            text=True,
        )

    def section(self, text, nl="\n"):
        body = PRECEDENCE + nl + nl + text.replace("\n", nl)
        if not body.endswith(nl):
            body += nl
        return BEGIN + nl + body + END + nl

    def test_deploy_writes_section_after_hive_block_and_own_files(self):
        result = self.run_script("deploy")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        for path in self.shared.values():
            text = self.read(path)
            self.assertEqual(text.count(BEGIN), 1, path)
            self.assertIn(HIVE_END + "\n" + self.section(RULE), text, path)
            self.assertTrue(text.startswith("My own notes.\n\n" + HIVE_BEGIN), path)
        cursor = self.read(self.cursor_personal)
        self.assertTrue(cursor.startswith("---\n"))
        self.assertIn("alwaysApply: true\n---\n\n" + PRECEDENCE + "\n\n" + RULE, cursor)
        agy = self.read(self.agy_personal)
        self.assertTrue(agy.startswith("---\ntrigger: always_on\n---\n\n" + PRECEDENCE + "\n\n" + RULE))

    def test_section_goes_after_the_voice_block(self):
        path = self.shared["codex"]
        voice = "\n".join([VOICE_BEGIN, "Voice text.", VOICE_END]) + "\n"
        self.write(path, hive_block() + voice + "Trailing user text.\n")
        self.assertEqual(self.run_script("deploy").returncode, 0)
        text = self.read(path)
        self.assertIn(VOICE_END + "\n" + self.section(RULE) + "Trailing user text.\n", text)

    def test_redeploy_replaces_in_place_and_keeps_bytes_permissions_and_line_endings(self):
        path = self.shared["pi"]
        self.write(path, "Before.\r\n" + hive_block("\r\n") + "After.\r\n")
        os.chmod(path, 0o640)
        self.assertEqual(self.run_script("deploy").returncode, 0)
        self.write(self.source, RULE + "\nAlso keep tests green.\n")
        self.assertEqual(self.run_script("deploy").returncode, 0)
        text = self.read(path)
        expected = (
            "Before.\r\n"
            + hive_block("\r\n")
            + self.section(RULE + "\nAlso keep tests green.\n", "\r\n")
            + "After.\r\n"
        )
        self.assertEqual(text, expected)
        self.assertEqual(stat.S_IMODE(os.stat(path).st_mode), 0o640)

    def test_only_hosts_with_hive_receive_rules(self):
        self.write(self.shared["pi"], "No Hive here.\n")
        self.cursor_hive.unlink()
        self.assertEqual(self.run_script("deploy").returncode, 0)
        self.assertEqual(self.read(self.shared["pi"]), "No Hive here.\n")
        self.assertFalse(self.cursor_personal.exists())
        self.assertTrue(self.agy_personal.exists())

    def test_check_passes_after_deploy_and_names_drift_and_position(self):
        self.assertEqual(self.run_script("deploy").returncode, 0)
        self.assertEqual(self.run_script("check").returncode, 0)
        codex = self.shared["codex"]
        self.write(codex, self.read(codex).replace("Recommend", "Always recommend"))
        opencode = self.shared["opencode"]
        text = self.read(opencode)
        start = text.index(BEGIN)
        section = text[start : text.index(END) + len(END) + 1]
        self.write(opencode, section + text.replace(section, ""))
        result = self.run_script("check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("codex", result.stdout)
        self.assertIn("opencode", result.stdout)
        self.assertNotIn("claude", result.stdout)

    def test_check_names_a_copy_in_a_host_without_hive(self):
        self.assertEqual(self.run_script("deploy").returncode, 0)
        self.cursor_hive.unlink()
        result = self.run_script("check")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("personal-rules.mdc", result.stdout)

    def test_deploy_moves_a_section_out_of_position(self):
        self.assertEqual(self.run_script("deploy").returncode, 0)
        path = self.shared["claude"]
        text = self.read(path)
        section = text[text.index(BEGIN) : text.index(END) + len(END) + 1]
        self.write(path, section + text.replace(section, ""))
        self.assertEqual(self.run_script("deploy").returncode, 0)
        text = self.read(path)
        self.assertEqual(text.count(BEGIN), 1)
        self.assertIn(HIVE_END + "\n" + self.section(RULE), text)

    def test_invalid_source_is_refused_without_writes(self):
        cases = {
            "size": "x" * 8001,
            "marker": RULE + HIVE_BEGIN + "\n",
            "own marker": RULE + END + "\n",
            "secret": RULE + "token ghp_" + "a1B2" * 9 + "\n",
        }
        before = {p: self.read(p) for p in self.shared.values()}
        for name, text in cases.items():
            with self.subTest(name):
                self.write(self.source, text)
                result = self.run_script("deploy")
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("rules.md", result.stdout + result.stderr)
                self.assertEqual({p: self.read(p) for p in self.shared.values()}, before)
                self.assertFalse(self.cursor_personal.exists())

    def test_prose_identifiers_are_not_secrets(self):
        self.write(self.source, RULE + "Set autoMemoryDirectory and toolActivation in settings.\n")
        result = self.run_script("deploy")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_remove_deletes_sections_and_files_and_keeps_hive(self):
        self.assertEqual(self.run_script("deploy").returncode, 0)
        result = self.run_script("remove")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        for path in self.shared.values():
            self.assertEqual(self.read(path), "My own notes.\n\n" + hive_block())
        self.assertFalse(self.cursor_personal.exists())
        self.assertFalse(self.agy_personal.exists())
        self.assertTrue(self.cursor_hive.exists())

    def test_host_homes_follow_their_environment_variables(self):
        claude_dir = self.home / "claude-alt"
        codex_dir = self.home / "codex-alt"
        pi_dir = self.home / "pi-alt"
        xdg = self.home / "xdg"
        for path in (claude_dir / "CLAUDE.md", codex_dir / "AGENTS.md", pi_dir / "AGENTS.md", xdg / "opencode/AGENTS.md"):
            self.write(path, hive_block())
        self.write(xdg / "hive/rules.md", RULE)
        env = dict(
            self.env,
            CLAUDE_CONFIG_DIR=str(claude_dir),
            CODEX_HOME=str(codex_dir),
            PI_CODING_AGENT_DIR=str(pi_dir),
            XDG_CONFIG_HOME=str(xdg),
        )
        result = self.run_script("deploy", env=env)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        for path in (claude_dir / "CLAUDE.md", codex_dir / "AGENTS.md", pi_dir / "AGENTS.md", xdg / "opencode/AGENTS.md"):
            self.assertEqual(self.read(path).count(BEGIN), 1, path)
        # Grok always reads ~/.claude/CLAUDE.md, so it still gets the rules.
        self.assertEqual(self.read(self.shared["claude"]).count(BEGIN), 1)
        self.assertNotIn(BEGIN, self.read(self.shared["codex"]))

    def test_codex_override_file_wins_when_not_empty(self):
        override = self.home / ".codex/AGENTS.override.md"
        self.write(override, hive_block())
        self.assertEqual(self.run_script("deploy").returncode, 0)
        self.assertEqual(self.read(override).count(BEGIN), 1)
        self.assertNotIn(BEGIN, self.read(self.shared["codex"]))

    def test_a_file_changed_between_read_and_write_is_skipped(self):
        module = load_module()
        env = dict(self.env)
        plan = module.build_plan(env)
        codex = self.shared["codex"]
        self.write(codex, self.read(codex) + "Edited meanwhile.\n")
        skipped = module.apply_plan(plan)
        self.assertEqual([str(p) for p in skipped], [str(codex)])
        self.assertNotIn(BEGIN, self.read(codex))
        self.assertEqual(self.read(self.shared["pi"]).count(BEGIN), 1)
        result = self.run_script("deploy")
        self.assertEqual(result.returncode, 0)

    def test_deploy_exits_non_zero_when_it_skips_a_file(self):
        module = load_module()
        plan = module.build_plan(dict(self.env))
        codex = self.shared["codex"]
        self.write(codex, self.read(codex) + "Edited meanwhile.\n")
        skipped = module.apply_plan(plan)
        self.assertNotEqual(module.exit_code_for(skipped=skipped, problems=[]), 0)


if __name__ == "__main__":
    unittest.main()
