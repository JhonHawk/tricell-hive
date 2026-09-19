#!/usr/bin/env python3
"""The pre-commit hook stages everything it regenerates.

Its invariant is that a generated output never travels in a different commit
than the canonical source it derives from. A tree the hook rebuilds but does
not `git add` breaks exactly that: the commit lands with a stale generated
file, and the next `build.py --check` — on someone else's machine — fails on a
change they did not make.

The hook is exercised in a THROWAWAY copy of the repo: it runs `git add`, so it
must never touch the real working tree.
"""

from __future__ import annotations

import os
import re
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
HOOK = REPO_ROOT / ".githooks/pre-commit"
GENERATED_ARRAY = re.compile(r"^readonly -a GENERATED=\(\n(.*?)^\)", re.M | re.S)


def generated_entries() -> list[str]:
    match = GENERATED_ARRAY.search(HOOK.read_text(encoding="utf-8"))
    assert match is not None, "the hook no longer declares a GENERATED array"
    return [line.strip() for line in match.group(1).splitlines() if line.strip()]


def clone_repo(destination: Path) -> Path:
    """The paths build.py reads plus the hook, as a fresh git repo."""
    destination.mkdir(parents=True)
    for relative in ("global", "harness", ".githooks", "AGENTS.md"):
        source = REPO_ROOT / relative
        target = destination / relative
        if source.is_dir():
            shutil.copytree(
                source,
                target,
                ignore=shutil.ignore_patterns("__pycache__", "node_modules"),
            )
        else:
            shutil.copy2(source, target)
    env = {**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_CONFIG_SYSTEM": "/dev/null"}
    for command in (
        ["git", "init", "-q", "-b", "master"],
        ["git", "config", "user.email", "test@example.invalid"],
        ["git", "config", "user.name", "Hook Test"],
        ["git", "add", "-A"],
        # The baseline commit is made BEFORE the hook is enabled: build.py
        # compares a core against HEAD, and there is no HEAD yet.
        ["git", "commit", "-qm", "baseline"],
        ["git", "config", "core.hooksPath", ".githooks"],
    ):
        subprocess.run(command, cwd=destination, check=True, env=env,
                       capture_output=True, text=True)
    return destination


class PreCommitHookTests(unittest.TestCase):
    def test_the_hook_lists_every_tree_the_build_generates(self) -> None:
        entries = set(generated_entries())
        import importlib.util

        spec = importlib.util.spec_from_file_location(
            "harness_build", REPO_ROOT / "harness/build.py"
        )
        build = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(build)

        expected = {
            path.as_posix()
            for path in (*build.GENERATED_TREE_RELATIVE_PATHS,
                         *build.GENERATED_FILE_RELATIVE_PATHS)
        }
        self.assertTrue(
            expected <= entries,
            f"the hook regenerates but never stages: {sorted(expected - entries)}",
        )

    def test_a_source_edit_stages_the_regenerated_claude_tree_and_manifest(self) -> None:
        with tempfile.TemporaryDirectory(prefix="hive-hook-") as temp:
            repo = clone_repo(Path(temp) / "repo")
            env = {**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null",
                   "GIT_CONFIG_SYSTEM": "/dev/null"}

            # Simulate the drift the hook exists to prevent: a canonical source
            # is edited and staged while its generated outputs stay behind.
            agent = repo / "global/agents/development/backend-developer.md"
            agent.write_text(
                agent.read_text(encoding="utf-8") + "\n- An added focus line.\n",
                encoding="utf-8",
            )
            # A rule's scope, so the manifest genuinely changes: an agent edit
            # alone leaves it byte-identical and it would rightly be absent.
            rule = repo / "global/rules/languages/shell-standards.md"
            rule.write_text(
                rule.read_text(encoding="utf-8").replace(
                    '  - "**/*.bash"', '  - "**/*.bash"\n  - "**/*.zsh"'
                ),
                encoding="utf-8",
            )
            subprocess.run(["git", "add", "global/"], cwd=repo, check=True,
                           env=env, capture_output=True, text=True)

            result = subprocess.run(
                ["git", "commit", "-m", "feat(agents): edit a canonical source"],
                cwd=repo, env=env, capture_output=True, text=True,
            )
            self.assertEqual(result.returncode, 0, result.stderr)

            committed = subprocess.run(
                ["git", "show", "--name-only", "--format=", "HEAD"],
                cwd=repo, check=True, env=env, capture_output=True, text=True,
            ).stdout.split()
            self.assertIn(
                "harness/claude/agents/development/backend-developer.md", committed
            )
            self.assertIn("harness/rule-manifest.json", committed)

            # Nothing regenerated is left behind in the working tree.
            dirty = subprocess.run(
                ["git", "status", "--porcelain", "harness/"],
                cwd=repo, check=True, env=env, capture_output=True, text=True,
            ).stdout
            self.assertEqual(dirty.strip(), "")


if __name__ == "__main__":
    unittest.main()
