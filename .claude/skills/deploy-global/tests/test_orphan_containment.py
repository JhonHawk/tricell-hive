#!/usr/bin/env python3
"""The orphan sweep is the only destructive path in deploy-global.sh.

Its input is `~/.claude/.deploy-manifest` — a plain file a corrupted run, a bad
merge or a hand edit can poison. An entry like `agents/../../victim.md` maps to
`${HOME}/victim.md`, which is outside every root this deploy manages, and the
`rm -rf` in `step_delete_orphans` had nothing between it and that path.

These tests pin the containment guard by construction (resolved target inside
the entry's managed root), never by pattern-matching one crafted string, and
pin the size-based circuit breaker that was documented but untested.

Everything runs against a throwaway HOME; the deletion cases are the point, so
they run for real inside that fake HOME and nowhere else.
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
# Exit status of a run that completed but refused at least one manifest entry.
ORPHAN_GUARD_EXIT_CODE = 3


def run_sourced(command: str, home: Path, *extra: str) -> subprocess.CompletedProcess[str]:
    """Source the script's definitions (never main) and run one command.

    Keeping the temporary copy beside the real script preserves its REPO_ROOT
    and bundled-filter path calculations.
    """
    source_only = SCRIPT.read_text(encoding="utf-8").split(
        "# ---------------------------------------------------------------------------\n# Main\n",
        1,
    )[0]
    with tempfile.NamedTemporaryFile(
        mode="w",
        encoding="utf-8",
        dir=SCRIPT.parent,
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
                f'source "$1"; REPORT_LOG="$2"; {command}',
                "orphan-guard-test",
                str(source_path),
                str(home / "report.log"),
                # Extra argv ($3 onward) so a test can hand the script bytes no
                # amount of quoting would survive — a bare CR, for one.
                *extra,
            ],
            cwd=REPO_ROOT,
            env={**os.environ, "HOME": str(home)},
            check=False,
            capture_output=True,
            text=True,
        )
    finally:
        source_path.unlink(missing_ok=True)


def run_script(home: Path, *args: str) -> subprocess.CompletedProcess[str]:
    """Run the real script end to end against a fake HOME (dry-run only)."""
    return subprocess.run(
        [str(SCRIPT), *args],
        cwd=REPO_ROOT,
        env={**os.environ, "HOME": str(home), "GIT_CONFIG_GLOBAL": "/dev/null"},
        check=False,
        capture_output=True,
        text=True,
    )


def copy_repo(destination: Path) -> Path:
    """A throwaway copy of the repo, so an --apply run may rebuild harness/.

    An end-to-end `--apply` runs `python3 harness/build.py`, which WRITES to
    the repo it runs in. Copying the three trees the deploy reads keeps every
    such write inside the temporary directory.
    """
    destination.mkdir(parents=True)
    for relative in (".claude/skills/deploy-global", "global", "harness"):
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copytree(
            REPO_ROOT / relative,
            target,
            ignore=shutil.ignore_patterns("__pycache__", "node_modules"),
            symlinks=True,
        )
    return destination


def run_deploy(repo: Path, home: Path, *args: str) -> subprocess.CompletedProcess[str]:
    """Run main() end to end — including --apply — against a copied repo."""
    return subprocess.run(
        [str(repo / ".claude/skills/deploy-global/scripts/deploy-global.sh"), *args],
        cwd=repo,
        env={**os.environ, "HOME": str(home), "GIT_CONFIG_GLOBAL": "/dev/null"},
        check=False,
        capture_output=True,
        text=True,
    )


def repo_entries(repo: Path, limit: int) -> list[str]:
    """Manifest filler from a COPIED repo: real sources, so never orphans.

    `global/agents` joined the pool when the 19 glob-scoped rules left
    `global/rules`: rules + skills alone fell to 97 files, under the 100 the
    breaker tests below ask for. Any root added here must (a) have a manifest
    prefix `manifest_entry_scope` recognizes — an unrecognized entry is skipped
    before MANIFEST_TOTAL counts it, padding the file without padding the
    denominator the percentage breaker divides by — and (b) resolve to a real
    source, which for `agents/*.md` is the GENERATED harness/claude/agents
    tree, so a stale tree would turn this filler into false orphans.
    """
    entries = [
        str(path.relative_to(repo / "global"))
        for root in ("global/rules", "global/skills", "global/agents")
        for path in sorted((repo / root).rglob("*"))
        if path.is_file() and "__pycache__" not in path.parts
    ]
    assert len(entries) >= limit, (
        f"expected at least {limit} non-orphan manifest entries, found {len(entries)} "
        "across global/{rules,skills,agents} — add a root with a recognized "
        "manifest prefix rather than lowering the limit"
    )
    return entries[:limit]


def make_home(temp: str) -> Path:
    home = Path(temp) / "home"
    (home / ".claude/agents").mkdir(parents=True)
    (home / "report.log").touch()
    return home


def write_manifest(home: Path, entries: list[str]) -> None:
    (home / ".claude/.deploy-manifest").write_text(
        "".join(f"{entry}\n" for entry in entries), encoding="utf-8"
    )


def report_of(home: Path) -> str:
    return (home / "report.log").read_text(encoding="utf-8")


def deployable_entries(limit: int) -> list[str]:
    """Real manifest entries whose sources exist, so they are never orphans.

    Same pool and the same two constraints as repo_entries above — read its
    docstring before changing a root here.
    """
    entries = [
        str(path.relative_to(REPO_ROOT / "global"))
        for root in ("global/rules", "global/skills", "global/agents")
        for path in sorted((REPO_ROOT / root).rglob("*"))
        if path.is_file() and "__pycache__" not in path.parts
    ]
    assert len(entries) >= limit, (
        f"expected at least {limit} non-orphan manifest entries, found {len(entries)} "
        "across global/{rules,skills,agents}"
    )
    return entries[:limit]


class ContainmentRejectionTests(unittest.TestCase):
    """Entries that escape their managed root are refused, never deleted."""

    def test_parent_traversal_entry_is_not_reported_as_a_deletable_orphan(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            (home / "victim.md").write_text("precious\n", encoding="utf-8")
            write_manifest(home, ["agents/../../victim.md"])

            result = run_script(home, "--dry-run", "--only", "claude")

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertIn("REFUSING to touch manifest entry", result.stderr)
            self.assertNotIn("Found 1 orphan(s)", result.stderr)
            self.assertTrue((home / "victim.md").exists())

    def test_parent_traversal_entry_survives_an_apply_run(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            (home / "victim.md").write_text("precious\n", encoding="utf-8")
            write_manifest(home, ["agents/../../victim.md"])

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertTrue(
                (home / "victim.md").exists(), "the file outside the managed root was deleted"
            )
            self.assertIn("REJECTED entry", report_of(home))
            self.assertNotIn("PARTIAL DEPLOY", result.stderr)

    def test_symlinked_parent_escaping_the_root_is_refused(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            outside = home / "outside"
            outside.mkdir()
            (outside / "victim.md").write_text("precious\n", encoding="utf-8")
            (home / ".claude/agents/escape").symlink_to(outside)
            write_manifest(home, ["agents/escape/victim.md"])

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertTrue(
                (outside / "victim.md").exists(),
                "a symlinked parent walked the delete outside the root",
            )
            self.assertIn("REJECTED entry", report_of(home))

    def test_the_deletion_step_refuses_an_escaping_entry_even_when_handed_one_directly(self) -> None:
        """Defense in depth: the guard sits at the rm, not only at detection."""
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            (home / "victim.md").write_text("precious\n", encoding="utf-8")
            write_manifest(home, ["agents/../../victim.md"])

            result = run_sourced(
                'APPLY=1; ORPHANS=("agents/../../victim.md"); step_delete_orphans; orphan_guard_exit_status',
                home,
            )

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertTrue((home / "victim.md").exists())
            self.assertIn("orphans: 0 deleted", report_of(home))

    def test_malformed_entries_are_refused_by_shape(self) -> None:
        cases = [
            ("../outside", "'..' segment"),
            ("agents/../../victim.md", "'..' segment"),
            ("/etc/passwd", "absolute path"),
            ("", "empty manifest entry"),
            ("agents/we\rird.md", "control character"),
            ("not-a-managed-prefix/file.md", "no managed root"),
            # A trailing slash survives dirname/basename, so the guard would
            # validate `…/skills/link` while `rm -rf` received `…/skills/link/`
            # — and on BSD that slash makes rm follow the symlink and empty its
            # DESTINATION. Empty and '.' segments ride the same rule.
            ("skills/deploy-global/", "empty or trailing path segment"),
            ("skills//deploy-global", "empty or trailing path segment"),
            ("skills/deploy-global/.", "'.' segment"),
            ("skills/./deploy-global", "'.' segment"),
        ]
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            for entry, reason in cases:
                with self.subTest(entry=entry or "<empty>"):
                    result = run_sourced(
                        'manifest_entry_map "$3"; '
                        'orphan_entry_is_contained "$3" "${MAP_TGT}" && echo CONTAINED || echo REFUSED',
                        home,
                        entry,
                    )
                    self.assertIn("REFUSED", result.stdout, result.stderr[-2000:])
                    self.assertIn(reason, result.stderr)

    def test_an_apply_run_drops_the_refused_entry_from_the_rewritten_manifest(self) -> None:
        """Skipping is not ignoring: the corrupt line does not come back forever.

        A refused entry is never promoted to KEPT_ORPHANS, so the manifest
        rewrite leaves it behind — the run self-heals the manifest without ever
        having acted on the path it named.
        """
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            (home / "victim.md").write_text("precious\n", encoding="utf-8")
            write_manifest(home, ["agents/../../victim.md"])

            result = run_sourced("APPLY=1; step_detect_orphans; step_write_manifest", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertTrue((home / "victim.md").exists())
            manifest = (home / ".claude/.deploy-manifest").read_text(encoding="utf-8")
            self.assertNotIn("agents/../../victim.md", manifest)
            self.assertIn("CLAUDE.md", manifest)

    def test_a_managed_root_itself_is_never_a_deletable_target(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            (home / ".claude/rules").mkdir()
            result = run_sourced(
                'orphan_entry_is_contained "rules/" "${HOME}/.claude/rules/" && echo CONTAINED || echo REFUSED',
                home,
            )
            self.assertIn("REFUSED", result.stdout, result.stderr[-2000:])


class LegitimateOrphanTests(unittest.TestCase):
    """The guard must not break the cleanup it protects."""

    def test_a_legitimate_orphan_inside_the_root_is_deleted(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            retired = home / ".claude/agents/development/retired-agent.md"
            retired.parent.mkdir(parents=True)
            retired.write_text("# retired\n", encoding="utf-8")
            # Filler keeps the single orphan under the percentage breaker.
            write_manifest(home, ["agents/development/retired-agent.md"] + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertFalse(retired.exists(), "the legitimate orphan was not deleted")
            self.assertIn("orphans: 1 deleted", report_of(home))

    def test_a_legitimate_orphan_symlink_is_unlinked_without_following_it(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            destination = home / "keep-me.md"
            destination.write_text("keep\n", encoding="utf-8")
            link = home / ".claude/agents/retired-link.md"
            link.symlink_to(destination)
            write_manifest(home, ["agents/retired-link.md"] + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertFalse(link.is_symlink(), "the deploy-created symlink survived")
            self.assertTrue(destination.exists(), "the delete followed the symlink to its destination")


class PipelineValidationTests(unittest.TestCase):
    """Shape validation belongs to the manifest loop, not only to the orphans.

    Calling the guard directly proves it rejects a string; it says nothing
    about whether the pipeline ever HANDS it that string. A malformed line
    whose scope is unknown, or whose source happens to exist, used to be
    skipped silently: never deleted, but never reported and no exit 3 either.
    """

    def test_malformed_lines_are_reported_even_when_they_are_not_orphans(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            malformed = [
                "-rf",
                "~/victim.md",
                "/etc/passwd",
                "not-a-managed-prefix/file.md",
                # Source EXISTS (global/../../README.md resolves), so the old
                # pipeline never reached the orphan branch for this line.
                "rules/../../README.md",
            ]
            write_manifest(home, malformed + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            report = report_of(home)
            for entry in malformed:
                with self.subTest(entry=entry):
                    self.assertIn(entry, report)
            self.assertEqual(result.stderr.count("REFUSING to touch manifest entry"), len(malformed))

    def test_a_dot_segment_entry_never_reaches_rm_and_never_aborts_the_run(self) -> None:
        """`rm -rf x/.` exits 1 on BSD; under set -e that killed the deploy."""
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            (home / ".claude/skills/deploy-global").mkdir(parents=True)
            write_manifest(home, ["skills/deploy-global/."] + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertIn("'.' segment", result.stderr)
            self.assertTrue((home / ".claude/skills/deploy-global").is_dir())

    def test_an_orphan_in_a_non_claude_scope_is_still_deleted(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            retired = home / ".codex/agents/retired-agent.toml"
            retired.parent.mkdir(parents=True)
            retired.write_text("# retired\n", encoding="utf-8")
            filler = [
                f"codex-agents/{path.stem}.toml"
                for path in sorted((REPO_ROOT / "global/agents").rglob("*.md"))
            ][:9]
            write_manifest(home, ["codex-agents/retired-agent.toml"] + filler)

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertFalse(retired.exists())
            self.assertIn("orphans: 1 deleted", report_of(home))


class SymlinkAndPruneTests(unittest.TestCase):
    def test_a_symlinked_root_still_deletes_its_legitimate_orphans(self) -> None:
        """`~/.claude/skills -> elsewhere` is a real setup; resolve, don't refuse."""
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            elsewhere = home / "elsewhere/skills"
            elsewhere.mkdir(parents=True)
            (elsewhere / "retired").mkdir()
            (elsewhere / "retired/SKILL.md").write_text("# retired\n", encoding="utf-8")
            (home / ".claude/skills").symlink_to(elsewhere)
            write_manifest(home, ["skills/retired/SKILL.md"] + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertFalse((elsewhere / "retired/SKILL.md").exists())
            self.assertIn("orphans: 1 deleted", report_of(home))

    def test_deleting_an_orphan_directory_does_not_follow_a_symlink_inside_it(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            outside = home / "outside"
            outside.mkdir()
            (outside / "decoy.txt").write_text("precious\n", encoding="utf-8")
            retired = home / ".claude/agents/retired-pack"
            retired.mkdir(parents=True)
            (retired / "agent.md").write_text("# retired\n", encoding="utf-8")
            (retired / "link").symlink_to(outside)
            write_manifest(home, ["agents/retired-pack"] + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertFalse(retired.exists())
            self.assertTrue((outside / "decoy.txt").exists(), "rm -rf followed a symlink inside the orphan")

    def test_empty_directory_cleanup_leaves_user_directories_alone(self) -> None:
        """Prune what the deletion emptied — not every empty dir under a root."""
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            user_dir = home / ".claude/rules/my-own-notes"
            user_dir.mkdir(parents=True)
            retired = home / ".claude/agents/retired-pack/agent.md"
            retired.parent.mkdir(parents=True)
            retired.write_text("# retired\n", encoding="utf-8")
            write_manifest(home, ["agents/retired-pack/agent.md"] + deployable_entries(9))

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertFalse(retired.parent.exists(), "the emptied orphan directory was not pruned")
            self.assertTrue(user_dir.is_dir(), "a user-created empty directory was deleted")
            self.assertTrue((home / ".claude/agents").is_dir(), "the scope root itself was pruned")


class EndToEndApplyTests(unittest.TestCase):
    """Through main(), with --apply, against a copied repo and a fake HOME."""

    def test_trailing_slash_on_a_symlink_cannot_destroy_its_destination(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            repo = copy_repo(Path(temp) / "repo")
            home = make_home(temp)
            (home / ".claude/skills").mkdir(parents=True)
            (home / "dev/mine").mkdir(parents=True)
            (home / "dev/mine/decoy.txt").write_text("precious\n", encoding="utf-8")
            (home / ".claude/skills/mine").symlink_to(home / "dev/mine")
            write_manifest(home, ["skills/mine/"] + repo_entries(repo, 9))

            result = run_deploy(repo, home, "--apply", "--only", "claude")

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertTrue(
                (home / "dev/mine/decoy.txt").exists(),
                "rm -rf followed the trailing slash into the symlink destination",
            )
            self.assertTrue((home / ".claude/skills/mine").is_symlink())
            self.assertIn("REFUSING to touch manifest entry", result.stderr)

    def test_a_rejected_entry_does_not_stop_the_deploy_or_the_other_orphan(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            repo = copy_repo(Path(temp) / "repo")
            home = make_home(temp)
            (home / "victim.md").write_text("precious\n", encoding="utf-8")
            retired = home / ".claude/agents/retired-agent.md"
            retired.write_text("# retired\n", encoding="utf-8")
            write_manifest(
                home,
                ["agents/../../victim.md", "agents/retired-agent.md"] + repo_entries(repo, 18),
            )

            result = run_deploy(repo, home, "--apply", "--only", "claude")

            self.assertEqual(result.returncode, ORPHAN_GUARD_EXIT_CODE, result.stderr[-3000:])
            self.assertTrue((home / "victim.md").exists(), "the refused entry was deleted")
            self.assertFalse(retired.exists(), "the legitimate orphan was not deleted")
            # The deploy itself completed: files landed and the manifest was rewritten.
            self.assertTrue((home / ".claude/CLAUDE.md").exists())
            manifest = (home / ".claude/.deploy-manifest").read_text(encoding="utf-8")
            self.assertNotIn("agents/../../victim.md", manifest)
            self.assertNotIn("agents/retired-agent.md", manifest)
            self.assertIn("CLAUDE.md", manifest)


class CircuitBreakerTests(unittest.TestCase):
    """>ORPHAN_MAX_ABS entries or >=ORPHAN_MAX_PCT% of the manifest refuses."""

    def _stage_orphans(self, home: Path, count: int, filler: list[str]) -> list[Path]:
        targets = []
        for index in range(count):
            target = home / f".claude/agents/retired-{index}.md"
            target.write_text("# retired\n", encoding="utf-8")
            targets.append(target)
        write_manifest(home, [f"agents/retired-{index}.md" for index in range(count)] + filler)
        return targets

    def test_percentage_threshold_refuses_deletion(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            targets = self._stage_orphans(home, 3, deployable_entries(9))  # 3/12 = 25%

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertIn("REFUSING to delete", result.stderr)
            self.assertIn("orphans: REFUSED deletion", report_of(home))
            for target in targets:
                self.assertTrue(target.exists())

    def test_absolute_threshold_refuses_deletion_below_the_percentage(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            targets = self._stage_orphans(home, 21, deployable_entries(100))  # 21/121 = 17%

            result = run_sourced("APPLY=1; step_detect_orphans; orphan_guard_exit_status", home)

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertIn("REFUSING to delete", result.stderr)
            for target in targets:
                self.assertTrue(target.exists())

    def test_force_flag_overrides_the_breaker(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            targets = self._stage_orphans(home, 3, deployable_entries(9))

            result = run_sourced(
                "APPLY=1; FORCE_DELETE_ORPHANS=1; step_detect_orphans; orphan_guard_exit_status", home
            )

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertIn("orphans: 3 deleted", report_of(home))
            for target in targets:
                self.assertFalse(target.exists())

    def test_refused_orphans_stay_in_the_manifest_for_a_later_run(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = make_home(temp)
            self._stage_orphans(home, 3, deployable_entries(9))

            result = run_sourced(
                'APPLY=1; step_detect_orphans; printf "%s\\n" "${KEPT_ORPHANS[@]}"', home
            )

            self.assertEqual(result.returncode, 0, result.stderr[-3000:])
            self.assertEqual(
                sorted(line for line in result.stdout.splitlines() if line.startswith("agents/")),
                sorted(f"agents/retired-{index}.md" for index in range(3)),
            )


if __name__ == "__main__":
    unittest.main()
