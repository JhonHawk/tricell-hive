#!/usr/bin/env python3
"""The Claude agents source is the generated tree, not the canonical one.

Agents that declare `packs:` only carry their rule texts in the BUILT output,
so a deploy reading `global/agents/` would ship the unpacked source and every
packed agent would silently lose its conventions. These tests pin the source
at `harness/claude/agents/` on every path that touches it: preview, diff,
deploy, and the manifest map that decides what counts as an orphan.
"""

from __future__ import annotations

import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[4]
SCRIPT = REPO_ROOT / ".claude/skills/deploy-global/scripts/deploy-global.sh"
GENERATED_AGENTS = REPO_ROOT / "harness/claude/agents"
SAMPLE_AGENT = Path("development/backend-developer.md")


def clone_repo(destination: Path) -> Path:
    """A throwaway copy of the repo, so a test may add a file to global/agents.

    Only the paths the deploy script reads: `cp -R` of the whole tree would
    drag 478 MB (node_modules, .git) for a handful of markdown files.
    """
    destination.mkdir(parents=True)
    for relative in (
        ".claude/skills/deploy-global",
        "global",
        "harness/claude",
        "harness/build",
        "harness/build.py",
    ):
        source = REPO_ROOT / relative
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        if source.is_dir():
            shutil.copytree(source, target, ignore=shutil.ignore_patterns("__pycache__"))
        else:
            shutil.copy2(source, target)
    return destination


def run_sourced(command: str, home: Path, repo: Path | None = None):
    """Source the script's definitions (never main) and run one command."""
    script = (repo or REPO_ROOT) / ".claude/skills/deploy-global/scripts/deploy-global.sh"
    source_only = script.read_text(encoding="utf-8").split(
        "# ---------------------------------------------------------------------------\n# Main\n",
        1,
    )[0]
    with tempfile.NamedTemporaryFile(
        mode="w",
        encoding="utf-8",
        dir=script.parent,
        prefix=".deploy-global-test-",
        suffix=".sh",
        delete=False,
    ) as source_file:
        source_file.write(source_only)
        source_path = Path(source_file.name)
    try:
        return subprocess.run(
            [
                "/bin/bash",
                "-c",
                f'source "$1"; {command}',
                "claude-agents-source-test",
                str(source_path),
            ],
            cwd=repo or REPO_ROOT,
            env={**os.environ, "HOME": str(home)},
            check=False,
            capture_output=True,
            text=True,
        )
    finally:
        source_path.unlink(missing_ok=True)


class ClaudeAgentsSourceTests(unittest.TestCase):
    def test_manifest_entry_map_resolves_agents_to_the_generated_tree(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            result = run_sourced(
                f'manifest_entry_map "agents/{SAMPLE_AGENT.as_posix()}"; '
                'printf "%s\\n%s\\n" "${MAP_SRC}" "${MAP_TGT}"',
                Path(temp),
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            src, tgt = result.stdout.splitlines()[:2]
            self.assertEqual(src, str(GENERATED_AGENTS / SAMPLE_AGENT))
            self.assertEqual(tgt, f"{temp}/.claude/agents/{SAMPLE_AGENT.as_posix()}")

    def test_preview_counts_the_generated_claude_agents_tree(self) -> None:
        # Assert the COUNT, not the annotation: the numbers come from the tree
        # that was walked, so they cannot agree with the wrong source by accident.
        expected_files = sum(1 for _ in GENERATED_AGENTS.rglob("*") if _.is_file())
        expected_lines = sum(
            path.read_text(encoding="utf-8").count("\n")
            for path in GENERATED_AGENTS.rglob("*") if path.is_file()
        )
        source_lines = sum(
            path.read_text(encoding="utf-8").count("\n")
            for path in (REPO_ROOT / "global/agents").rglob("*.md")
        )
        self.assertNotEqual(expected_lines, source_lines, "the counts cannot discriminate")

        with tempfile.TemporaryDirectory() as temp:
            result = run_sourced(
                "RUN_CODEX=0; RUN_OPENCODE=0; RUN_GROK=0; RUN_PI=0; step_preview",
                Path(temp),
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            agents_line = next(
                line for line in result.stderr.splitlines()
                if line.strip().startswith("agents:")
            )
            self.assertIn(f"{expected_files} files, {expected_lines} lines", agents_line)

    def test_the_written_manifest_lists_the_generated_claude_agents(self) -> None:
        # step_detect_orphans resolves `agents/*` against the generated tree, so
        # a manifest enumerated from global/agents would list entries whose
        # source "does not exist" — and an --apply would delete them from the
        # user's home. The two enumerations have to name the same files.
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            (home / ".claude").mkdir(parents=True)
            result = run_sourced(
                'APPLY=1; RUN_CODEX=0; RUN_OPENCODE=0; RUN_GROK=0; RUN_PI=0; '
                'KEPT_ORPHANS=(); step_write_manifest',
                home,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            listed = {
                line for line in
                (home / ".claude/.deploy-manifest").read_text(encoding="utf-8").splitlines()
                if line.startswith("agents/")
            }
            expected = {
                f"agents/{path.relative_to(GENERATED_AGENTS).as_posix()}"
                for path in GENERATED_AGENTS.rglob("*") if path.is_file()
            }
            self.assertEqual(listed, expected)

    def test_an_old_manifest_entry_for_a_non_agent_file_is_not_an_orphan(self) -> None:
        """R3: the fix for new manifests does nothing for the ones already out there.

        The previous script enumerated `global/agents -type f`, so a user's
        manifest can already list a non-`.md` file. Resolved against the
        generated tree — which only ever holds `*.md` — its source reads as
        missing, orphan detection flags it, and `--apply` deletes a file from
        the user's home. It is not generated: it lives in `global/agents/`.
        """
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            repo = clone_repo(root / "repo")
            stray_rel = "agents/development/team-notes.txt"
            (repo / "global" / stray_rel).write_text("not an agent\n", encoding="utf-8")
            home = root / "home"
            (home / ".claude/agents/development").mkdir(parents=True)
            (home / ".claude" / stray_rel).write_text("not an agent\n", encoding="utf-8")
            (home / ".claude/.deploy-manifest").write_text(
                f"{stray_rel}\nagents/development/backend-developer.md\n",
                encoding="utf-8",
            )

            resolved = run_sourced(
                f'manifest_entry_map "{stray_rel}"; printf "%s\\n" "${{MAP_SRC}}"',
                home, repo=repo,
            )
            self.assertEqual(resolved.returncode, 0, resolved.stderr)
            self.assertEqual(
                resolved.stdout.strip(), str(repo / "global" / stray_rel)
            )

            orphans = run_sourced(
                'RUN_CODEX=0; RUN_OPENCODE=0; RUN_GROK=0; RUN_PI=0; '
                'step_detect_orphans; printf "%s\\n" "${ORPHANS[@]+${ORPHANS[@]}}"',
                home, repo=repo,
            )
            self.assertEqual(orphans.returncode, 0, orphans.stderr)
            self.assertNotIn(stray_rel, orphans.stdout)
            self.assertIn("No orphans found", orphans.stderr)

    def test_a_non_markdown_file_in_the_agent_source_never_enters_the_manifest(self) -> None:
        """S5: what the manifest lists, orphan detection must be able to resolve.

        The converter only renders `*.md`, so anything else under
        `global/agents/` exists in the canonical tree and in NO generated one.
        Enumerated from `global/agents`, it lands in the manifest, resolves to a
        missing source, and an `--apply` deletes it from the user's home.
        """
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            repo = clone_repo(root / "repo")
            stray = repo / "global/agents/development/team-notes.txt"
            stray.write_text("not an agent\n", encoding="utf-8")
            home = root / "home"
            (home / ".claude").mkdir(parents=True)

            result = run_sourced(
                'APPLY=1; RUN_CODEX=0; RUN_OPENCODE=0; RUN_GROK=0; RUN_PI=0; '
                'KEPT_ORPHANS=(); step_write_manifest',
                home,
                repo=repo,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            manifest = (home / ".claude/.deploy-manifest").read_text(encoding="utf-8")
            self.assertNotIn("agents/development/team-notes.txt", manifest)

            # And every agent entry that IS listed resolves to a real file.
            for entry in sorted(
                line for line in manifest.splitlines() if line.startswith("agents/")
            ):
                resolved = run_sourced(
                    f'manifest_entry_map "{entry}"; printf "%s\\n" "${{MAP_SRC}}"',
                    home,
                    repo=repo,
                )
                self.assertTrue(Path(resolved.stdout.strip()).is_file(), entry)

    def test_diff_compares_the_deployed_agents_against_the_generated_tree(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            shutil.copytree(GENERATED_AGENTS, home / ".claude/agents")
            result = run_sourced(
                "RUN_CODEX=0; RUN_OPENCODE=0; RUN_GROK=0; RUN_PI=0; step_diff",
                home,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            agents_line = next(
                line for line in result.stderr.splitlines()
                if line.strip().startswith("agents:")
            )
            # Every file already matches the generated tree byte-for-byte; a
            # diff against global/agents would report them all as modified.
            self.assertIn("0 new, 0 modified", agents_line)

    def test_deploy_copies_the_generated_agents_with_their_build_provenance(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_sourced("APPLY=1; step_deploy_claude", home)
            self.assertEqual(result.returncode, 0, result.stderr)
            deployed = home / ".claude/agents" / SAMPLE_AGENT
            self.assertEqual(
                deployed.read_bytes(), (GENERATED_AGENTS / SAMPLE_AGENT).read_bytes()
            )
            self.assertIn("do not edit by hand", deployed.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
