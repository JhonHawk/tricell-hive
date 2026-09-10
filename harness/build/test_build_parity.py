#!/usr/bin/env python3
"""Regression tests for read-only generated-tree parity checks."""

import hashlib
import importlib.util
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
                Path("harness/opencode/rules"),
            ),
        )

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
                Path("harness/pi/agents/code-reviewer.md"),
                Path(
                    "harness/agents-skills/language-rules/references/"
                    "python-standards.md"
                ),
                Path("harness/opencode/rules/python-standards.md"),
            )
            for relative in stale_paths:
                target = fixture / relative
                target.write_bytes(target.read_bytes() + b"\nstale\n")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            message = str(raised.exception)
            for relative in stale_paths:
                self.assertIn(f"stale: {relative.as_posix()}", message)

    def test_reports_missing_and_extra_entries_together(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            missing = Path("harness/codex/agents/code-reviewer.toml")
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
