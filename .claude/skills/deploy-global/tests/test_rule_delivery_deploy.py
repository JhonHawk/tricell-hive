#!/usr/bin/env python3
"""The rule-delivery hook ships with its manifest, in both harnesses.

Two things make this hook different from every sibling: it is Python, not
shell, and it is useless without `harness/rule-manifest.json` — the hook reads
the deployed rule texts' paths from it and exits silently when it is missing.
So the deploy has to copy a `*.py` hook (the existing steps only matched
`*.sh`) and land the manifest next to it, on the Claude side and the Codex one,
without the orphan sweep then deleting what it just deployed.
"""

from __future__ import annotations

import importlib.util
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[4]
SCRIPT = REPO_ROOT / ".claude/skills/deploy-global/scripts/deploy-global.sh"
HOOK_SOURCE = REPO_ROOT / "global/hooks/rule-delivery/rule-delivery.py"
RULE_MANIFEST = REPO_ROOT / "harness/rule-manifest.json"


def run_sourced(command: str, home: Path) -> subprocess.CompletedProcess[str]:
    """Source the script's definitions (never main) and run one command."""
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
            ["/bin/bash", "-c", f'source "$1"; {command}', "rule-delivery-deploy-test",
             str(source_path)],
            cwd=REPO_ROOT,
            env={**os.environ, "HOME": str(home)},
            check=False,
            capture_output=True,
            text=True,
        )
    finally:
        source_path.unlink(missing_ok=True)


class RuleDeliveryDeployTests(unittest.TestCase):
    def test_the_claude_deploy_ships_the_python_hook_with_its_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_sourced("APPLY=1; RUN_CLAUDE=1; step_deploy_hooks", home)
            self.assertEqual(result.returncode, 0, result.stderr)

            deployed = home / ".claude/hooks/rule-delivery.py"
            self.assertEqual(deployed.read_bytes(), HOOK_SOURCE.read_bytes())
            self.assertTrue(os.access(deployed, os.X_OK), "a hook that is not executable is inert")

            manifest = home / ".claude/hooks/rule-manifest.json"
            self.assertEqual(manifest.read_bytes(), RULE_MANIFEST.read_bytes())

    def test_the_hooks_own_test_runner_is_never_deployed(self) -> None:
        # D8: it is a *.py under global/hooks/ like the hook itself; only the
        # hook belongs in the user's home.
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            self.assertEqual(
                run_sourced("APPLY=1; RUN_CLAUDE=1; step_deploy_hooks", home).returncode, 0)
            self.assertEqual(
                run_sourced("APPLY=1; RUN_CODEX=1; step_deploy_codex_hooks", home).returncode, 0)
            for hooks in (home / ".claude/hooks", home / ".codex/hooks"):
                self.assertFalse((hooks / "test_rule_delivery.py").exists(), str(hooks))
                self.assertTrue((hooks / "rule-delivery.py").exists(), str(hooks))

    def test_the_codex_deploy_ships_the_python_hook_with_its_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_sourced("APPLY=1; RUN_CODEX=1; step_deploy_codex_hooks", home)
            self.assertEqual(result.returncode, 0, result.stderr)

            deployed = home / ".codex/hooks/rule-delivery.py"
            self.assertEqual(deployed.read_bytes(), HOOK_SOURCE.read_bytes())
            self.assertTrue(os.access(deployed, os.X_OK))
            self.assertEqual(
                (home / ".codex/hooks/rule-manifest.json").read_bytes(),
                RULE_MANIFEST.read_bytes(),
            )

    def test_the_deployed_manifest_maps_back_to_its_source(self) -> None:
        # Unmapped, the orphan sweep reads the file it just deployed as
        # unowned and offers to delete it.
        with tempfile.TemporaryDirectory() as temp:
            for entry, target in (
                ("hooks/rule-manifest.json", f"{temp}/.claude/hooks/rule-manifest.json"),
                ("codex-hooks/rule-manifest.json", f"{temp}/.codex/hooks/rule-manifest.json"),
            ):
                result = run_sourced(
                    f'manifest_entry_map "{entry}"; printf "%s\\n%s\\n" "${{MAP_SRC}}" "${{MAP_TGT}}"',
                    Path(temp),
                )
                self.assertEqual(result.returncode, 0, result.stderr)
                source, resolved = result.stdout.splitlines()[:2]
                self.assertEqual(source, str(RULE_MANIFEST), entry)
                self.assertEqual(resolved, target, entry)

    def test_the_python_hook_maps_back_to_its_source(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            result = run_sourced(
                'manifest_entry_map "hooks/rule-delivery.py"; printf "%s\\n" "${MAP_SRC}"',
                Path(temp),
            )
            self.assertEqual(result.stdout.strip(), str(HOOK_SOURCE))

    def test_the_written_manifest_tracks_the_hook_and_its_rule_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            (home / ".claude").mkdir(parents=True)
            result = run_sourced(
                "APPLY=1; RUN_CLAUDE=1; RUN_CODEX=1; RUN_OPENCODE=0; RUN_GROK=0; RUN_PI=0; "
                "KEPT_ORPHANS=(); step_write_manifest",
                home,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            listed = (home / ".claude/.deploy-manifest").read_text(encoding="utf-8").splitlines()
            for entry in ("hooks/rule-delivery.py", "hooks/rule-manifest.json",
                          "codex-hooks/rule-delivery.py", "codex-hooks/rule-manifest.json"):
                self.assertIn(entry, listed)

            # Everything tracked resolves to a real source — the invariant the
            # orphan sweep depends on.
            for entry in listed:
                if not entry.startswith(("hooks/", "codex-hooks/")):
                    continue
                resolved = run_sourced(
                    f'manifest_entry_map "{entry}"; printf "%s\\n" "${{MAP_SRC}}"', home
                )
                self.assertTrue(Path(resolved.stdout.strip()).is_file(), entry)


def load_pi_deploy():
    spec = importlib.util.spec_from_file_location(
        "hive_pi_deploy", REPO_ROOT / "harness/pi/deploy.py"
    )
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    bytecode = sys.dont_write_bytecode
    sys.dont_write_bytecode = True
    # `@dataclass` resolves its own module out of sys.modules, so the entry has
    # to exist before the body runs.
    sys.modules[spec.name] = module
    try:
        spec.loader.exec_module(module)
    finally:
        sys.dont_write_bytecode = bytecode
    return module


class PiRuleDeliveryDeployTests(unittest.TestCase):
    def test_the_pi_bundle_carries_the_hook_and_the_manifest_beside_it(self) -> None:
        # PI resolves hooks under its own agent directory, so a hook mapped in
        # hooks.ts but absent from the bundle is a path that never runs — and
        # the manifest has to travel with it or the hook exits silently.
        deploy = load_pi_deploy()
        shipped = dict(deploy._source_files(REPO_ROOT))
        self.assertEqual(
            shipped.get("global/hooks/rule-delivery/rule-delivery.py"), HOOK_SOURCE
        )
        self.assertEqual(
            shipped.get("global/hooks/rule-delivery/rule-manifest.json"), RULE_MANIFEST
        )
        # D6: PI silently skips a hook it cannot execute, and a checkout does
        # not have to preserve the executable bit — so the deploy forces it,
        # for a .py exactly as for a .sh. Checked against a NON-executable
        # source: reading it off the repo file would pass either way.
        with tempfile.TemporaryDirectory() as temp:
            plain = Path(temp) / "rule-delivery.py"
            plain.write_text("#!/usr/bin/env python3\n", encoding="utf-8")
            plain.chmod(0o644)
            self.assertEqual(
                oct(deploy._deploy_mode(
                    plain, "global/hooks/rule-delivery/rule-delivery.py")),
                oct(0o755),
            )


if __name__ == "__main__":
    unittest.main()
