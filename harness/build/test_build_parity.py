#!/usr/bin/env python3
"""Regression tests for read-only generated-tree parity checks."""

import hashlib
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
BUILD_PATH = ROOT / "harness" / "build.py"
SPEC = importlib.util.spec_from_file_location("harness_build", BUILD_PATH)
BUILD_MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(BUILD_MODULE)


def snapshot(root: Path) -> dict[str, tuple[str, int, str]]:
    entries = {}
    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root).as_posix()
        if path.is_dir():
            entries[relative] = ("directory", path.stat().st_mtime_ns, "")
        else:
            digest = hashlib.sha256(path.read_bytes()).hexdigest()
            entries[relative] = ("file", path.stat().st_mtime_ns, digest)
    return entries


class GeneratedTreeParityTests(unittest.TestCase):
    def test_skill_generation_excludes_python_runtime_cache(self):
        with tempfile.TemporaryDirectory(prefix="hive-skill-cache-") as tmp:
            root = Path(tmp)
            source = root / "source" / "flow-core"
            script = source / "scripts" / "plan.py"
            script.parent.mkdir(parents=True)
            script.write_text("print('validator')\n", encoding="utf-8")
            cache = script.parent / "__pycache__"
            cache.mkdir()
            (cache / "plan.cpython-314.pyc").write_bytes(b"machine-specific cache")
            (source / "SKILL.md").write_text("# Shared flow library\n", encoding="utf-8")
            output = root / "generated"
            subprocess.run(
                [sys.executable, str(ROOT / "harness/build/convert-skills.py"),
                 str(root / "source"), str(output)],
                check=True, capture_output=True, text=True,
            )
            self.assertEqual((output / "flow-core/scripts/plan.py").read_bytes(), script.read_bytes())
            self.assertFalse((output / "flow-core/scripts/__pycache__").exists())

    def create_generated_fixture(self, parent: Path) -> Path:
        fixture = parent / "actual"
        BUILD_MODULE.generate_generated_trees(fixture)
        return fixture

    def test_generated_tree_inventory_covers_every_harness_surface(self):
        self.assertEqual(
            BUILD_MODULE.GENERATED_TREE_RELATIVE_PATHS,
            (
                Path("harness/agents-skills"),
                Path("harness/codex/agents"),
                Path("harness/opencode/agents"),
                Path("harness/grok/agents"),
                Path("harness/pi/agents"),
                # The whole directory, not just agents/: a stray file dropped
                # directly under harness/claude/ is generated-tree drift too.
                Path("harness/claude"),
            ),
        )
        self.assertEqual(
            BUILD_MODULE.GENERATED_FILE_RELATIVE_PATHS,
            (Path("harness/rule-manifest.json"),),
        )
        # The agents tree deploys verbatim into ~/.claude/agents/, so it holds
        # agent definitions only — a README there would land beside them.
        self.assertFalse(
            (ROOT / "harness/claude/agents/README.md").exists(),
            "a README inside the Claude agents tree would be deployed as an agent",
        )

    def test_reports_a_stray_file_directly_under_the_claude_root(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            stray = Path("harness/claude/leftover.md")
            (fixture / stray).write_text("stray\n", encoding="utf-8")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertIn(f"extra: {stray.as_posix()}", str(raised.exception))

    def test_reports_a_hand_edited_generated_claude_agent(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            edited = Path("harness/claude/agents/development/backend-developer.md")
            target = fixture / edited
            target.write_bytes(target.read_bytes() + b"\n- hand edit\n")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertIn(f"stale: {edited.as_posix()}", str(raised.exception))

    def test_clean_check_is_read_only(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            before = snapshot(fixture)

            BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertEqual(snapshot(fixture), before)

    def test_reports_stale_agent_skill_reference_and_rule(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            stale_paths = (
                Path("harness/pi/agents/review-code.md"),
                Path(
                    "harness/agents-skills/language-rules/references/"
                    "python-standards.md"
                ),
            )
            for relative in stale_paths:
                target = fixture / relative
                target.write_bytes(target.read_bytes() + b"\nstale\n")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            message = str(raised.exception)
            for relative in stale_paths:
                self.assertIn(f"stale: {relative.as_posix()}", message)

    def test_rule_manifest_describes_every_rule_text_with_its_delivery_paths(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            manifest = json.loads(
                (fixture / "harness/rule-manifest.json").read_text(encoding="utf-8")
            )
            rules = {entry["name"]: entry for entry in manifest["rules"]}

            # A README is not a rule text; nothing downstream should deliver it.
            self.assertNotIn("README", rules)
            # One entry per rule text under global/rules/** + rules-situational/**.
            self.assertEqual(len(manifest["rules"]), 40)
            # The two pack-only texts carry no globs: the hook never delivers them.
            for pack_only in ("agent-core-gates", "test-gate"):
                self.assertEqual(rules[pack_only]["globs"], [], pack_only)
            # The glob-scoped set the hook delivers by touched file.
            glob_scoped = sorted(n for n, e in rules.items() if e["globs"])
            self.assertEqual(len(glob_scoped), 19, glob_scoped)

            typescript = rules["typescript-standards"]
            self.assertEqual(
                typescript["source"],
                "global/rules-situational/typescript-standards.md",
            )
            self.assertEqual(typescript["globs"], ["**/*.{ts,tsx,js,jsx}"])
            self.assertFalse(typescript["always_on"])
            self.assertFalse(typescript["readers"])
            self.assertEqual(
                typescript["references"],
                {
                    "claude": "~/.claude/skills/language-rules/references/"
                              "typescript-standards.md",
                    "agents": "~/.agents/skills/language-rules/references/"
                              "typescript-standards.md",
                },
            )

            # `*.service.ts` is Angular's and NestJS's alike: each rule names
            # the pack that proves the file belongs to the other framework, so
            # the hook never holds a packed agent on its rival's rule.
            self.assertEqual(rules["angular-patterns"]["exclusive_with"],
                             ["nestjs-patterns"])
            self.assertEqual(rules["nestjs-patterns"]["exclusive_with"],
                             ["angular-patterns"])
            self.assertEqual(typescript["exclusive_with"], [])

            session_capture = rules["session-capture"]
            self.assertEqual(
                session_capture["globs"],
                ["**/_support/**", "**/*-specs/**", "**/_support/sessions/**",
                 "**/*-specs/sessions/**"],
            )
            # Reviewers need these two at read time even when they carry no pack.
            self.assertTrue(session_capture["readers"])
            self.assertTrue(rules["project-structure"]["readers"])
            self.assertFalse(rules["support-artifacts"]["readers"])

            # Always-on: under global/rules/ with no `paths:`.
            self.assertTrue(rules["testing"]["always_on"])
            self.assertEqual(rules["testing"]["globs"], [])
            # rules-situational/ is never always-on — no harness loads it by itself.
            agent_routing = rules["agent-routing"]
            self.assertFalse(agent_routing["always_on"])
            self.assertEqual(agent_routing["globs"], [])
            self.assertEqual(
                agent_routing["references"]["agents"],
                "~/.agents/skills/task-routing/references/agent-routing.md",
            )
            # Injected into no router skill -> no deployed reference path.
            self.assertIsNone(rules["security"]["references"])

    def test_an_exclusive_with_naming_an_unknown_rule_fails_the_build(self):
        # The key is a cross-reference between rule files; a typo would silently
        # gate nothing forever, so the build refuses it rather than emit it.
        with tempfile.TemporaryDirectory(prefix="hive-build-exclusive-") as tmp:
            source = Path(tmp)
            store = source / "global/rules-situational"
            store.mkdir(parents=True)
            (store / "nestjs-patterns.md").write_text(
                '---\nglobs:\n  - "**/*.service.ts"\n---\n', encoding="utf-8")
            (store / "angular-patterns.md").write_text(
                '---\nglobs:\n  - "**/*.service.ts"\nexclusive-with: nestjs-pattern\n'
                "---\n", encoding="utf-8")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.build_rule_manifest(source)

            message = str(raised.exception)
            self.assertIn("angular-patterns.md", message)
            self.assertIn("nestjs-pattern", message)

            # Spelled right, the same pair builds.
            (store / "angular-patterns.md").write_text(
                '---\nglobs:\n  - "**/*.service.ts"\nexclusive-with: nestjs-patterns\n'
                "---\n", encoding="utf-8")
            manifest = BUILD_MODULE.build_rule_manifest(source)
            entries = {entry["name"]: entry for entry in manifest["rules"]}
            self.assertEqual(entries["angular-patterns"]["exclusive_with"],
                             ["nestjs-patterns"])

    def test_a_rule_declaring_paths_fails_the_build_wherever_it_lives(self):
        # `paths:` was Claude Code's native path-scoping key and the mechanism is
        # retired: rules are delivered by the hook, off `globs:`. A file still
        # carrying `paths:` would look scoped while loading unconditionally, so
        # the build refuses it — under global/rules/ and rules-situational/ alike.
        with tempfile.TemporaryDirectory(prefix="hive-build-paths-") as tmp:
            source = Path(tmp)
            for directory in ("global/rules/quality", "global/rules-situational"):
                (source / directory).mkdir(parents=True)
            scoped = source / "global/rules-situational/typescript-standards.md"
            scoped.write_text('---\nglobs:\n  - "**/*.ts"\n---\n', encoding="utf-8")

            for relative in ("global/rules/quality/testing.md",
                             "global/rules-situational/react-nextjs.md"):
                offender = source / relative
                offender.write_text('---\npaths: "**/*.tsx"\n---\n', encoding="utf-8")
                with self.subTest(rule=relative):
                    with self.assertRaises(SystemExit) as raised:
                        BUILD_MODULE.build_rule_manifest(source)
                    message = str(raised.exception)
                    self.assertIn(relative, message)
                    self.assertIn("paths:", message)
                offender.unlink()

            # Without it the same tree builds, and a core rule stays always-on.
            manifest = BUILD_MODULE.build_rule_manifest(source)
            entries = {entry["name"]: entry for entry in manifest["rules"]}
            self.assertEqual(entries["typescript-standards"]["globs"], ["**/*.ts"])
            self.assertFalse(entries["typescript-standards"]["always_on"])

    def test_the_manifest_maps_each_agent_to_the_packs_it_carries(self):
        # The rule-delivery hook must skip what an agent already holds inlined,
        # and `packs:` is stripped from every generated output — the manifest is
        # the only place that survives the build.
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            manifest = json.loads(
                (fixture / "harness/rule-manifest.json").read_text(encoding="utf-8")
            )
            self.assertIn("agents", manifest)
            # The specialized development agents, and only they, carry packs;
            # the generic backend agent stays hook-fed.
            self.assertEqual(
                sorted(manifest["agents"]),
                [
                    "angular-developer",
                    "database-specialist",
                    "kotlin-multiplatform-developer",
                    "react-developer",
                    "ts-backend-developer",
                ],
            )
            for agent, packs in manifest["agents"].items():
                self.assertEqual(packs[0], "agent-core-gates", agent)
            # Built from the source agents, so it cannot silently go missing.
            self.assertEqual(
                BUILD_MODULE.build_rule_manifest()["agents"], manifest["agents"]
            )

    def test_the_manifest_lists_the_agents_that_cannot_write(self):
        # The rule-delivery hook gives a read-only agent only the rules a
        # READER needs. "Read-only" is not a label anyone maintains: it is the
        # converter's own write-capability verdict, so an agent that gains or
        # loses Write/Edit moves lists without anyone remembering to.
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            manifest = json.loads(
                (fixture / "harness/rule-manifest.json").read_text(encoding="utf-8")
            )
            read_only = manifest["read_only_agents"]
            self.assertEqual(read_only, sorted(read_only))
            self.assertIn("review-code", read_only)
            self.assertIn("sdd-explore", read_only)
            # sdd-verify denies Edit but keeps Write for its report: still a writer.
            self.assertNotIn("sdd-verify", read_only)
            self.assertNotIn("backend-developer", read_only)

            converter = BUILD_MODULE._load_agent_converter()
            expected = sorted(
                converter.parse_agent(path)["name"]
                for path in (ROOT / "global/agents").rglob("*.md")
                if not converter.can_write(converter.parse_agent(path))
            )
            self.assertEqual(read_only, expected)

    def test_inline_globs_keep_their_brace_groups_intact(self):
        with tempfile.TemporaryDirectory(prefix="hive-globs-") as tmp:
            source = Path(tmp)
            rules = source / "global/rules-situational"
            rules.mkdir(parents=True)
            (source / "global/rules").mkdir(parents=True)
            (rules / "listed.md").write_text(
                '---\nglobs:\n  - "**/*.{ts,tsx}"\n  - "**/*.vue"\n---\n\ntext\n',
                encoding="utf-8",
            )
            (rules / "inlined.md").write_text(
                '---\nglobs: "**/*.{ts,tsx,mts}"\n---\n\ntext\n', encoding="utf-8"
            )
            rules_by_name = {
                entry["name"]: entry
                for entry in BUILD_MODULE.build_rule_manifest(source)["rules"]
            }
            self.assertEqual(
                rules_by_name["listed"]["globs"], ["**/*.{ts,tsx}", "**/*.vue"]
            )
            # A comma inside braces is part of ONE glob, not a separator.
            self.assertEqual(rules_by_name["inlined"]["globs"], ["**/*.{ts,tsx,mts}"])

    def test_always_on_follows_the_directory_and_globs_is_refused_in_the_core(self):
        with tempfile.TemporaryDirectory(prefix="hive-alwayson-") as tmp:
            source = Path(tmp)
            rules = source / "global/rules/quality"
            rules.mkdir(parents=True)
            situational = source / "global/rules-situational"
            situational.mkdir(parents=True)
            (rules / "gate.md").write_text("---\nalwaysApply: true\n---\n\ntext\n",
                                           encoding="utf-8")
            (situational / "stored.md").write_text(
                '---\nglobs:\n  - "**/*.ts"\n---\n\ntext\n', encoding="utf-8"
            )
            rules_by_name = {
                entry["name"]: entry
                for entry in BUILD_MODULE.build_rule_manifest(source)["rules"]
            }
            self.assertTrue(rules_by_name["gate"]["always_on"])
            # The store's own key: scoped, and never "always-on" anywhere.
            self.assertEqual(rules_by_name["stored"]["globs"], ["**/*.ts"])
            self.assertFalse(rules_by_name["stored"]["always_on"])

            # `globs:` under global/rules/ claims a scope nothing applies:
            # that directory loads unconditionally.
            (rules / "mislabeled.md").write_text(
                '---\nglobs:\n  - "**/*.ts"\n---\n\ntext\n', encoding="utf-8"
            )
            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.build_rule_manifest(source)
            message = str(raised.exception)
            self.assertIn("mislabeled.md", message)
            self.assertIn("globs:", message)

    def test_reports_a_stale_rule_manifest(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            target = fixture / "harness/rule-manifest.json"
            target.write_text(
                target.read_text(encoding="utf-8").replace('"readers"', '"reader"'),
                encoding="utf-8",
            )

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertIn("stale: harness/rule-manifest.json", str(raised.exception))

    def test_reports_missing_and_extra_entries_together(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            missing = Path("harness/codex/agents/review-code.toml")
            extra = Path("harness/grok/agents/orphan.md")
            (fixture / missing).unlink()
            (fixture / extra).write_text("orphan\n", encoding="utf-8")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            message = str(raised.exception)
            self.assertIn(f"missing: {missing.as_posix()}", message)
            self.assertIn(f"extra: {extra.as_posix()}", message)


if __name__ == "__main__":
    unittest.main()
