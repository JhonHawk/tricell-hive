#!/usr/bin/env python3
"""Isolated tests for the PI-only deploy route."""

from __future__ import annotations

import importlib.util
import hashlib
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock


REPO_ROOT = Path(__file__).resolve().parents[4]
HELPER = REPO_ROOT / "harness" / "pi" / "deploy.py"
SCRIPT = REPO_ROOT / ".claude" / "skills" / "deploy-global" / "scripts" / "deploy-global.sh"
FIXTURE_ROOT = Path(__file__).resolve().parent / "fixtures" / "pi-source"
PACKAGE_FIXTURE_ROOT = Path(__file__).resolve().parent / "fixtures" / "pi-package" / "pi-subagents"
PI_HOOKS = (
    "bash-policy/bash-policy.sh",
    "flow-context/flow-context.sh",
    "flow-plan-capture/flow-plan-capture.sh",
    "flow-session-context/flow-session-context.sh",
    "post-tool-hub/post-tool-hub.sh",
    "reviewer-guard/reviewer-guard.sh",
    "rule-context/rule-context.sh",
    "session-hygiene-report/session-hygiene-report.sh",
)
PATCH_FILES = (
    "harness/pi/patches/pi-subagents-0.67.0.patch",
    "harness/pi/patches/pi-subagents-0.67.0.json",
)
PI_PACKAGE_PINS = (
    ("pi-subagents", "0.67.0"),
    ("gentle-engram", "0.1.12"),
    ("pi-mcp-adapter", "2.32.1"),
    ("@juicesharp/rpiv-ask-user-question", "2.9.0"),
    ("pi-web-access", "0.29.0"),
)
PI_PACKAGE_SOURCES = tuple(
    f"npm:{name}@{version}" for name, version in PI_PACKAGE_PINS
)


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def seed_installed_package(pi_dir: Path) -> None:
    package_root = pi_dir / "npm/node_modules/pi-subagents"
    if (package_root / "package.json").is_file():
        return
    source_root = PACKAGE_FIXTURE_ROOT
    package_root.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source_root / "package.json", package_root / "package.json")
    metadata = json.loads((FIXTURE_ROOT / PATCH_FILES[1]).read_text(encoding="utf-8"))
    for target in metadata["targets"]:
        source = source_root / target["path"]
        destination = package_root / target["path"]
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source, destination)


def seed_full_pi_prerequisites(pi_dir: Path) -> None:
    """Create the exact installed package/settings surface for preflight tests."""

    seed_installed_package(pi_dir)
    for package_name, version in PI_PACKAGE_PINS[1:]:
        package_root = pi_dir / "npm/node_modules" / package_name
        package_root.mkdir(parents=True, exist_ok=True)
        (package_root / "package.json").write_text(
            json.dumps({"name": package_name, "version": version}) + "\n",
            encoding="utf-8",
        )
    settings = pi_dir / "settings.json"
    settings.parent.mkdir(parents=True, exist_ok=True)
    settings.write_text(json.dumps({"packages": list(PI_PACKAGE_SOURCES)}) + "\n", encoding="utf-8")


def seed_minimal_source(source: Path) -> None:
    (source / "harness/pi/agents").mkdir(parents=True)
    (source / "harness/pi/src").mkdir(parents=True)
    (source / "harness/pi/extensions/hive").mkdir(parents=True)
    (source / "harness/AGENTS.md").write_text("core\n", encoding="utf-8")
    (source / "harness/pi/agents/demo.md").write_text("agent\n", encoding="utf-8")
    (source / "harness/pi/src/index.ts").write_text("runtime\n", encoding="utf-8")
    (source / "harness/pi/extensions/hive-hooks.ts").write_text("extension\n", encoding="utf-8")
    (source / "harness/pi/extensions/hive/reviewer-guard.ts").write_text("reviewer\n", encoding="utf-8")
    for relative in PATCH_FILES:
        destination = source / relative
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(FIXTURE_ROOT / relative, destination)
    for relative in PI_HOOKS:
        hook = source / "global/hooks" / relative
        hook.parent.mkdir(parents=True, exist_ok=True)
        hook.write_text("#!/bin/sh\n", encoding="utf-8")
        hook.chmod(0o755)


def run_helper(
    *arguments: str, check: bool = True, seed_package: bool = True
) -> subprocess.CompletedProcess[str]:
    if seed_package:
        for index, argument in enumerate(arguments):
            if argument == "--pi-dir" and index + 1 < len(arguments):
                seed_installed_package(Path(arguments[index + 1]))
            elif argument.startswith("--pi-dir="):
                seed_installed_package(Path(argument.split("=", 1)[1]))
    return subprocess.run(
        ["python3", str(HELPER), *arguments],
        cwd=REPO_ROOT,
        check=check,
        text=True,
        capture_output=True,
    )


def run_global_script(*arguments: str, home: Path, pi_dir: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [str(SCRIPT), *arguments],
        cwd=REPO_ROOT,
        env={**os.environ, "HOME": str(home), "PI_CODING_AGENT_DIR": str(pi_dir)},
        check=False,
        text=True,
        capture_output=True,
    )


def load_deploy_module():
    module_name = "pi_deploy_test_module"
    spec = importlib.util.spec_from_file_location(module_name, HELPER)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"Unable to load {HELPER}")
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    spec.loader.exec_module(module)
    return module


class PiDeployTests(unittest.TestCase):
    def test_dry_run_is_read_only_and_does_not_create_global_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp) / "home"
            pi_dir = Path(temp) / "pi-agent"
            home.mkdir()
            claude_home = home / ".claude"
            claude_home.mkdir()
            global_manifest = claude_home / ".deploy-manifest"
            global_manifest.write_text("global-owned\n", encoding="utf-8")
            environment = {**os.environ, "HOME": str(home), "PI_CODING_AGENT_DIR": str(pi_dir)}
            result = subprocess.run(
                [str(SCRIPT), "--only", "pi"],
                cwd=REPO_ROOT,
                env=environment,
                check=False,
                text=True,
                capture_output=True,
            )

            self.assertEqual(result.returncode, 2)
            self.assertIn("prerequisite missing", result.stderr)
            self.assertEqual(global_manifest.read_text(encoding="utf-8"), "global-owned\n")
            self.assertFalse(pi_dir.exists())

    def test_pi_scope_accepts_mixed_selection_but_fails_preflight_before_writes(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp) / "home"
            home.mkdir()
            result = subprocess.run(
                [str(SCRIPT), "--only", "pi,claude"],
                cwd=REPO_ROOT,
                env={**os.environ, "HOME": str(home), "PI_CODING_AGENT_DIR": str(Path(temp) / "pi-agent")},
                text=True,
                capture_output=True,
            )

            self.assertEqual(result.returncode, 2)
            self.assertIn("prerequisite missing", result.stderr)
            self.assertNotIn("isolated", result.stderr)
            self.assertFalse((home / ".claude").exists())

    def test_pi_preflight_rejects_missing_or_incompatible_package_before_writes(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            home = root / "home"
            home.mkdir()
            pi_dir = root / "pi-agent"
            result = run_global_script("--only", "pi", "--apply", home=home, pi_dir=pi_dir)
            self.assertEqual(result.returncode, 2)
            self.assertIn("prerequisite missing", result.stderr)
            self.assertFalse((home / ".agents").exists())
            self.assertFalse((home / ".claude").exists())

            seed_full_pi_prerequisites(pi_dir)
            bad_package = pi_dir / "npm/node_modules/gentle-engram/package.json"
            bad_package.write_text(json.dumps({"name": "gentle-engram", "version": "9.9.9"}) + "\n", encoding="utf-8")
            result = run_global_script("--only", "pi", "--apply", home=home, pi_dir=pi_dir)
            self.assertEqual(result.returncode, 2)
            self.assertIn("must be gentle-engram@0.1.12", result.stderr)
            self.assertFalse((home / ".agents/.hive-deploy-manifest.json").exists())
            self.assertFalse((pi_dir / ".hive-deploy-backups").exists())

    def test_pi_preflight_accepts_exact_installed_pins_and_enabled_settings(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            seed_full_pi_prerequisites(pi_dir)
            result = run_helper(
                "preflight", "--repo-root", str(FIXTURE_ROOT),
                "--pi-dir", str(pi_dir),
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("PI preflight", result.stdout)
            self.assertFalse((pi_dir / ".hive-deploy-manifest.json").exists())
            self.assertFalse((pi_dir / ".hive-deploy-backups").exists())

    def test_mixed_preflight_rejects_selected_codex_symlink_before_pi_or_shared_writes(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            home = root / "home"
            codex = home / ".codex"
            codex.mkdir(parents=True)
            outside = root / "operator-owned.md"
            outside.write_text("operator\n", encoding="utf-8")
            (codex / "AGENTS.md").symlink_to(outside)

            result = run_global_script(
                "--only", "pi,codex", "--apply",
                home=home, pi_dir=root / "pi-agent",
            )

            self.assertEqual(result.returncode, 1)
            self.assertIn("Codex AGENTS.md target is a symlink", result.stderr)
            self.assertFalse((home / ".agents").exists())
            self.assertFalse((root / "pi-agent").exists())
            self.assertEqual(outside.read_text(encoding="utf-8"), "operator\n")

    def test_pi_only_apply_writes_pi_and_neutral_shared_owner_preserving_legacy_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            seed_minimal_source(source)
            shared_source = source / "harness/agents-skills/demo/SKILL.md"
            shared_source.parent.mkdir(parents=True)
            shared_source.write_text("generated\n", encoding="utf-8")
            home = root / "home"
            claude = home / ".claude"
            claude.mkdir(parents=True)
            legacy = claude / ".deploy-manifest"
            legacy_bytes = b"# deploy-global manifest\n# deployed_at: fixture\n"
            legacy.write_bytes(legacy_bytes)
            pi_dir = root / "pi-agent"
            deploy_module = load_deploy_module()
            seed_installed_package(pi_dir)
            self.assertEqual(deploy_module.deploy(source, pi_dir, True), 0)
            self.assertEqual(
                deploy_module.deploy_shared(
                    source, home / ".agents", legacy, True
                ),
                0,
            )
            self.assertEqual(legacy.read_bytes(), legacy_bytes)
            self.assertTrue((pi_dir / "AGENTS.md").exists())
            neutral_manifest = home / ".agents/.hive-deploy-manifest.json"
            self.assertTrue(neutral_manifest.exists())
            self.assertEqual(json.loads(neutral_manifest.read_text(encoding="utf-8"))["scope"], "shared-skills")
            self.assertTrue((home / ".agents/skills/demo/SKILL.md").exists())
            self.assertFalse((home / ".codex").exists())
            self.assertFalse((home / ".config").exists())

    def test_shared_owner_preserves_foreign_files_and_is_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            skill = source / "harness/agents-skills/demo/SKILL.md"
            skill.parent.mkdir(parents=True)
            skill.write_text("generated\n", encoding="utf-8")
            shared_root = root / "home/.agents"
            foreign = shared_root / "skills/foreign/SKILL.md"
            foreign.parent.mkdir(parents=True)
            foreign.write_text("operator\n", encoding="utf-8")
            arguments = [
                "shared", "deploy", "--repo-root", str(source),
                "--shared-root", str(shared_root),
                "--legacy-manifest", str(root / "home/.claude/.deploy-manifest"),
                "--apply",
            ]

            first = run_helper(*arguments)
            self.assertIn("shared skills deploy", first.stdout)
            backups = sorted((shared_root / ".hive-deploy-backups").iterdir())
            self.assertEqual((shared_root / "skills/demo/SKILL.md").read_text(encoding="utf-8"), "generated\n")
            self.assertEqual(foreign.read_text(encoding="utf-8"), "operator\n")
            manifest = json.loads((shared_root / ".hive-deploy-manifest.json").read_text(encoding="utf-8"))
            self.assertEqual(manifest["scope"], "shared-skills")
            self.assertIn("skills/demo/SKILL.md", manifest["managedFiles"])

            second = run_helper(*arguments)
            self.assertIn("files: 0 write, 0 delete, 0 conflict", second.stdout)
            self.assertEqual(backups, sorted((shared_root / ".hive-deploy-backups").iterdir()))

    def test_shared_rollback_restores_only_the_shared_owner(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            skill = source / "harness/agents-skills/demo/SKILL.md"
            skill.parent.mkdir(parents=True)
            skill.write_text("generated\n", encoding="utf-8")
            shared_root = root / "home/.agents"
            foreign = shared_root / "skills/foreign/SKILL.md"
            foreign.parent.mkdir(parents=True)
            foreign.write_text("operator\n", encoding="utf-8")
            arguments = [
                "shared", "deploy", "--repo-root", str(source),
                "--shared-root", str(shared_root),
                "--legacy-manifest", str(root / "home/.claude/.deploy-manifest"),
                "--apply",
            ]
            self.assertEqual(run_helper(*arguments).returncode, 0)
            backup = sorted((shared_root / ".hive-deploy-backups").iterdir())[-1]
            rollback = run_helper(
                "shared", "rollback", "--shared-root", str(shared_root),
                "--backup-dir", str(backup), "--apply",
            )
            self.assertEqual(rollback.returncode, 0)
            self.assertFalse((shared_root / "skills/demo/SKILL.md").exists())
            self.assertFalse((shared_root / ".hive-deploy-manifest.json").exists())
            self.assertEqual(foreign.read_text(encoding="utf-8"), "operator\n")

    def test_shared_legacy_adoption_uses_recorded_source_commit_without_mutating_legacy_manifest(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            skill = source / "harness/agents-skills/demo/SKILL.md"
            skill.parent.mkdir(parents=True)
            skill.write_text("old generated\n", encoding="utf-8")
            subprocess.run(["git", "init", "-q"], cwd=source, check=True)
            subprocess.run(["git", "config", "user.email", "test@example.invalid"], cwd=source, check=True)
            subprocess.run(["git", "config", "user.name", "Hive Test"], cwd=source, check=True)
            subprocess.run(["git", "add", "."], cwd=source, check=True)
            subprocess.run(["git", "commit", "-qm", "fixture"], cwd=source, check=True)
            commit = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=source, text=True).strip()
            skill.write_text("new generated\n", encoding="utf-8")
            shared_root = root / "home/.agents"
            target = shared_root / "skills/demo/SKILL.md"
            target.parent.mkdir(parents=True)
            target.write_text("old generated\n", encoding="utf-8")
            legacy = root / "home/.claude/.deploy-manifest"
            legacy.parent.mkdir(parents=True)
            legacy_bytes = f"# source_commit: {commit}\nagents-skills/demo/SKILL.md\n".encode()
            legacy.write_bytes(legacy_bytes)

            result = run_helper(
                "shared", "deploy", "--repo-root", str(source),
                "--shared-root", str(shared_root),
                "--legacy-manifest", str(legacy), "--apply",
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(target.read_text(encoding="utf-8"), "new generated\n")
            self.assertEqual(legacy.read_bytes(), legacy_bytes)
            self.assertTrue((shared_root / ".hive-deploy-manifest.json").exists())

    def test_shared_unknown_edit_and_symlink_root_fail_before_any_write(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            skill = source / "harness/agents-skills/demo/SKILL.md"
            skill.parent.mkdir(parents=True)
            skill.write_text("generated\n", encoding="utf-8")
            shared_root = root / "home/.agents"
            target = shared_root / "skills/demo/SKILL.md"
            target.parent.mkdir(parents=True)
            target.write_text("operator edit\n", encoding="utf-8")
            legacy = root / "home/.claude/.deploy-manifest"
            legacy.parent.mkdir(parents=True)
            legacy.write_text("# source_commit: deadbeef\nagents-skills/demo/SKILL.md\n", encoding="utf-8")
            result = run_helper(
                "shared", "deploy", "--repo-root", str(source),
                "--shared-root", str(shared_root),
                "--legacy-manifest", str(legacy), "--apply", check=False,
            )
            self.assertEqual(result.returncode, 2)
            self.assertIn("unknown edits", result.stderr)
            self.assertEqual(target.read_text(encoding="utf-8"), "operator edit\n")
            self.assertFalse((shared_root / ".hive-deploy-manifest.json").exists())
            self.assertFalse((shared_root / ".hive-deploy-backups").exists())

            link_parent = root / "link-parent"
            link_parent.mkdir()
            link = root / "linked-agents"
            link.symlink_to(shared_root, target_is_directory=True)
            result = run_helper(
                "shared", "deploy", "--repo-root", str(source),
                "--shared-root", str(link),
                "--legacy-manifest", str(root / "missing-manifest"), check=False,
            )
            self.assertEqual(result.returncode, 2)
            self.assertIn("root is a symlink", result.stderr)

    def test_rollback_rejects_backup_from_another_owner(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "source"
            seed_minimal_source(source)
            shared_skill = source / "harness/agents-skills/demo/SKILL.md"
            shared_skill.parent.mkdir(parents=True)
            shared_skill.write_text("generated\n", encoding="utf-8")
            pi_dir = root / "pi-agent"
            seed_installed_package(pi_dir)
            self.assertEqual(run_helper("deploy", "--repo-root", str(source), "--pi-dir", str(pi_dir), "--apply").returncode, 0)
            pi_backup = sorted((pi_dir / ".hive-deploy-backups").iterdir())[-1]
            shared_root = root / "home/.agents"
            shared_result = run_helper(
                "shared", "deploy", "--repo-root", str(source),
                "--shared-root", str(shared_root),
                "--legacy-manifest", str(root / "home/.claude/.deploy-manifest"),
                "--apply",
            )
            self.assertEqual(shared_result.returncode, 0)
            shared_backup = sorted((shared_root / ".hive-deploy-backups").iterdir())[-1]

            wrong_shared = run_helper(
                "shared", "rollback", "--shared-root", str(shared_root),
                "--backup-dir", str(pi_backup), "--apply", check=False,
            )
            self.assertEqual(wrong_shared.returncode, 2)
            self.assertIn("Backup scope", wrong_shared.stderr)
            wrong_pi = run_helper(
                "rollback", "--pi-dir", str(pi_dir),
                "--backup-dir", str(shared_backup), "--apply", check=False,
            )
            self.assertEqual(wrong_pi.returncode, 2)
            self.assertIn("Backup scope", wrong_pi.stderr)

    def test_missing_required_source_fails_before_any_target_write(self) -> None:
        missing_paths = (
            "global/hooks/flow-plan-capture/flow-plan-capture.sh",
            "harness/pi/src/index.ts",
            "harness/pi/extensions/hive/reviewer-guard.ts",
        )
        for missing_path in missing_paths:
            with self.subTest(missing_path=missing_path), tempfile.TemporaryDirectory() as temp:
                source = Path(temp) / "source"
                pi_dir = Path(temp) / "pi-agent"
                seed_minimal_source(source)
                (source / missing_path).unlink()
                pi_dir.mkdir(parents=True)
                marker = pi_dir / "operator-owned.txt"
                marker.write_text("leave me\n", encoding="utf-8")
                settings = pi_dir / "settings.json"
                settings.write_text('{"defaultProvider":"xai"}\n', encoding="utf-8")

                result = run_helper(
                    "deploy",
                    "--repo-root",
                    str(source),
                    "--pi-dir",
                    str(pi_dir),
                    "--apply",
                    check=False,
                    seed_package=False,
                )

                self.assertEqual(result.returncode, 2)
                self.assertIn("Missing", result.stderr)
                self.assertEqual(marker.read_text(encoding="utf-8"), "leave me\n")
                self.assertEqual(settings.read_text(encoding="utf-8"), '{"defaultProvider":"xai"}\n')
                self.assertFalse((pi_dir / ".hive-deploy-backups").exists())
                self.assertFalse((pi_dir / ".hive-deploy-manifest.json").exists())

    def test_mid_apply_failure_leaves_incremental_rollback_journal(self) -> None:
        deploy_module = load_deploy_module()
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            seed_minimal_source(source)
            # Resolve this path because macOS may pass /private/var/folders to
            # the helper even when the test constructed the /var/folders form.
            pi_dir = (Path(temp) / "pi-agent").resolve()
            seed_installed_package(pi_dir)
            first_result = deploy_module.deploy(source, pi_dir, True)
            self.assertEqual(first_result, 0)

            core_target = pi_dir / "AGENTS.md"
            core_before = core_target.read_bytes()
            core_mode_before = core_target.stat().st_mode & 0o777
            fail_target = pi_dir / "src/index.ts"
            runtime_before = fail_target.read_bytes()
            runtime_mode_before = fail_target.stat().st_mode & 0o777
            source_core = source / "harness/AGENTS.md"
            source_core.write_bytes(source_core.read_bytes() + b"changed\n")
            source_core.chmod(0o600)
            (source / "harness/pi/src/index.ts").write_text(
                "export const changed = true;\n", encoding="utf-8"
            )

            user_target = pi_dir / "agents/angular-developer.md"
            user_target.parent.mkdir(parents=True, exist_ok=True)
            user_target.write_text("operator-owned\n", encoding="utf-8")
            original_atomic_write = deploy_module._atomic_write

            def fail_on_runtime(path: Path, data: bytes, mode: int) -> None:
                if path == fail_target:
                    raise OSError("injected mid-apply failure")
                original_atomic_write(path, data, mode)

            with mock.patch.object(deploy_module, "_atomic_write", side_effect=fail_on_runtime):
                with self.assertRaises(deploy_module.DeployError) as failure:
                    deploy_module.deploy(source, pi_dir, True)

            self.assertIn("PI apply failed after backup", str(failure.exception))
            backups = sorted((pi_dir / ".hive-deploy-backups").iterdir())
            backup = backups[-1]
            metadata = json.loads((backup / "metadata.json").read_text(encoding="utf-8"))
            self.assertTrue(metadata["files"]["AGENTS.md"]["applied"])
            self.assertIn("pendingSha256After", metadata["files"]["src/index.ts"])

            rollback_result = deploy_module._rollback(backup, pi_dir, True)

            self.assertEqual(rollback_result, 0)
            self.assertEqual(core_target.read_bytes(), core_before)
            self.assertEqual(core_target.stat().st_mode & 0o777, core_mode_before)
            self.assertEqual(fail_target.read_bytes(), runtime_before)
            self.assertEqual(fail_target.stat().st_mode & 0o777, runtime_mode_before)
            self.assertEqual(user_target.read_text(encoding="utf-8"), "operator-owned\n")
            self.assertEqual(
                (pi_dir / "extensions/hive-hooks.ts").read_bytes(),
                (source / "harness/pi/extensions/hive-hooks.ts").read_bytes(),
            )

    def test_change_after_backup_is_not_overwritten(self) -> None:
        deploy_module = load_deploy_module()
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            seed_minimal_source(source)
            pi_dir = (Path(temp) / "pi-agent").resolve()
            seed_installed_package(pi_dir)
            self.assertEqual(deploy_module.deploy(source, pi_dir, True), 0)

            target = pi_dir / "AGENTS.md"
            (source / "harness/AGENTS.md").write_text("new core\n", encoding="utf-8")
            original_create_backup = deploy_module._create_backup

            def create_backup_then_race(
                backup_pi_dir: Path, manifest_path: Path, relatives: tuple[str, ...]
            ) -> Path:
                backup = original_create_backup(backup_pi_dir, manifest_path, relatives)
                target.write_text("operator race\n", encoding="utf-8")
                return backup

            with mock.patch.object(
                deploy_module, "_create_backup", side_effect=create_backup_then_race
            ):
                with self.assertRaises(deploy_module.DeployError):
                    deploy_module.deploy(source, pi_dir, True)

            self.assertEqual(target.read_text(encoding="utf-8"), "operator race\n")

    def test_apply_preserves_unrelated_config_and_is_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            pi_dir.mkdir(parents=True)
            (pi_dir / "settings.json").write_text(
                json.dumps(
                    {
                        "defaultProvider": "xai",
                        "packages": [
                            "npm:other@1.2.3",
                            {
                                "source": "npm:pi-subagents@0.67.0",
                                "autoload": False,
                                "extensions": ["operator-extension"],
                            },
                        ],
                    }
                ),
                encoding="utf-8",
            )
            (pi_dir / "mcp.json").write_text(
                json.dumps({"mcpServers": {"other": {"url": "https://other"}}, "settings": {"keep": True}}),
                encoding="utf-8",
            )
            (pi_dir / "web-search.json").write_text(
                json.dumps({"xaiApiKey": "$XAI", "provider": "xai"}),
                encoding="utf-8",
            )

            arguments = ["deploy", "--repo-root", str(FIXTURE_ROOT), "--pi-dir", str(pi_dir), "--apply"]
            first = run_helper(*arguments)
            self.assertIn("configs: 4 write", first.stdout)
            backups = sorted((pi_dir / ".hive-deploy-backups").iterdir())

            settings = json.loads((pi_dir / "settings.json").read_text(encoding="utf-8"))
            self.assertEqual(settings["defaultProvider"], "xai")
            self.assertEqual(settings["packages"][0], "npm:other@1.2.3")
            managed_object = next(
                item
                for item in settings["packages"]
                if isinstance(item, dict) and item.get("source") == "npm:pi-subagents@0.67.0"
            )
            self.assertFalse(managed_object["autoload"])
            self.assertEqual(managed_object["extensions"], ["operator-extension"])
            package_sources = {
                item if isinstance(item, str) else item.get("source")
                for item in settings["packages"]
            }
            self.assertTrue({
                "npm:pi-subagents@0.67.0",
                "npm:gentle-engram@0.1.12",
                "npm:pi-mcp-adapter@2.32.1",
                "npm:@juicesharp/rpiv-ask-user-question@2.9.0",
                "npm:pi-web-access@0.29.0",
            }.issubset(package_sources))
            subagent = json.loads((pi_dir / "extensions/subagent/config.json").read_text(encoding="utf-8"))
            self.assertTrue(subagent["forceTopLevelAsync"])

            mcp = json.loads((pi_dir / "mcp.json").read_text(encoding="utf-8"))
            self.assertIn("other", mcp["mcpServers"])
            self.assertEqual(mcp["mcpServers"]["context7"]["protocolVersion"], "auto")
            self.assertFalse(mcp["mcpServers"]["context7"]["directTools"])
            self.assertEqual(mcp["mcpServers"]["context7"]["includeTools"], ["resolve-library-id", "query-docs"])
            self.assertEqual(mcp["mcpServers"]["context7"]["lifecycle"], "lazy")
            self.assertTrue(mcp["settings"]["keep"])

            web = json.loads((pi_dir / "web-search.json").read_text(encoding="utf-8"))
            self.assertEqual(web["xaiApiKey"], "$XAI")
            self.assertEqual(web["provider"], "openai")
            self.assertEqual(web["openaiSearchProviders"], ["openai-codex"])
            self.assertEqual(web["workflow"], "none")
            role_files = list((pi_dir / "agents").glob("*.md"))
            if role_files:
                role_text = role_files[0].read_text(encoding="utf-8")
                self.assertNotIn("__HIVE_PI_ROOT__", role_text)
                self.assertIn(str(pi_dir), role_text)
            self.assertFalse((pi_dir / "agents/README.md").exists())
            self.assertTrue((pi_dir / "global/hooks/flow-plan-capture/flow-plan-capture.sh").exists())
            self.assertTrue((pi_dir / "global/hooks/post-tool-hub/post-tool-hub.sh").exists())
            self.assertTrue((pi_dir / "global/hooks/reviewer-guard/reviewer-guard.sh").stat().st_mode & 0o111)
            self.assertFalse((pi_dir / "global/hooks/instructions-audit/instructions-audit.sh").exists())
            self.assertTrue((pi_dir / "src/index.ts").exists())

            second = run_helper(*arguments)
            self.assertIn("files: 0 write, 0 delete, 0 conflict", second.stdout)
            self.assertEqual(backups, sorted((pi_dir / ".hive-deploy-backups").iterdir()))

    def test_package_patch_adopts_pristine_files_and_is_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            seed_minimal_source(source)
            arguments = ["deploy", "--repo-root", str(source), "--pi-dir", str(pi_dir), "--apply"]
            metadata = json.loads((source / PATCH_FILES[1]).read_text(encoding="utf-8"))

            first = run_helper(*arguments)
            self.assertIn("files: 15 write, 0 delete, 0 conflict", first.stdout)
            manifest = json.loads((pi_dir / ".hive-deploy-manifest.json").read_text(encoding="utf-8"))
            for record in metadata["targets"]:
                target = pi_dir / "npm/node_modules/pi-subagents" / record["path"]
                managed = manifest["managedFiles"][
                    f"npm/node_modules/pi-subagents/{record['path']}"
                ]
                self.assertEqual(sha256_file(target), record["afterSha256"])
                self.assertEqual(managed["kind"], "package-patch")
                self.assertEqual(managed["sha256Before"], record["beforeSha256"])
                self.assertEqual(managed["sha256After"], record["afterSha256"])

            second = run_helper(*arguments)
            self.assertIn("files: 0 write, 0 delete, 0 conflict", second.stdout)

    def test_pristine_package_reinstall_is_reapplied(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            seed_minimal_source(source)
            arguments = ["deploy", "--repo-root", str(source), "--pi-dir", str(pi_dir), "--apply"]
            metadata = json.loads((source / PATCH_FILES[1]).read_text(encoding="utf-8"))
            run_helper(*arguments)

            for record in metadata["targets"]:
                target = pi_dir / "npm/node_modules/pi-subagents" / record["path"]
                pristine = PACKAGE_FIXTURE_ROOT / record["path"]
                shutil.copy2(pristine, target)

            result = run_helper(*arguments)
            self.assertIn("files: 2 write, 0 delete, 0 conflict", result.stdout)
            for record in metadata["targets"]:
                target = pi_dir / "npm/node_modules/pi-subagents" / record["path"]
                self.assertEqual(sha256_file(target), record["afterSha256"])

    def test_unknown_package_patch_state_fails_before_other_writes(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            seed_minimal_source(source)
            arguments = ["deploy", "--repo-root", str(source), "--pi-dir", str(pi_dir), "--apply"]
            run_helper(*arguments)
            core = pi_dir / "AGENTS.md"
            core_before = core.read_bytes()
            target = pi_dir / "npm/node_modules/pi-subagents/src/runs/shared/child-tool-plan.ts"
            target.write_text("operator package edit\n", encoding="utf-8")
            (source / "harness/AGENTS.md").write_text("new core\n", encoding="utf-8")

            result = run_helper(*arguments, check=False)

            self.assertEqual(result.returncode, 2)
            self.assertIn("neither pristine nor already patched", result.stderr)
            self.assertEqual(target.read_text(encoding="utf-8"), "operator package edit\n")
            self.assertEqual(core.read_bytes(), core_before)

    def test_package_patch_rejects_ancestor_symlink_escape(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            seed_minimal_source(source)
            seed_installed_package(pi_dir)
            package_root = pi_dir / "npm/node_modules/pi-subagents"
            other_root = pi_dir / "npm/node_modules/other-package"
            other_shared = other_root / "shared"
            other_shared.mkdir(parents=True)
            metadata = json.loads((source / PATCH_FILES[1]).read_text(encoding="utf-8"))
            for record in metadata["targets"]:
                source_target = PACKAGE_FIXTURE_ROOT / record["path"]
                target = other_shared / Path(record["path"]).name
                shutil.copy2(source_target, target)
            package_shared = package_root / "src/runs/shared"
            shutil.rmtree(package_shared)
            package_shared.symlink_to(other_shared, target_is_directory=True)
            marker = pi_dir / "operator-owned.txt"
            marker.write_text("leave me\n", encoding="utf-8")

            result = run_helper(
                "deploy",
                "--repo-root",
                str(source),
                "--pi-dir",
                str(pi_dir),
                "--apply",
                check=False,
            )

            self.assertEqual(result.returncode, 2)
            self.assertIn("outside installed package root", result.stderr)
            for record in metadata["targets"]:
                target = other_shared / Path(record["path"]).name
                self.assertEqual(target.read_bytes(), (PACKAGE_FIXTURE_ROOT / record["path"]).read_bytes())
            self.assertEqual(marker.read_text(encoding="utf-8"), "leave me\n")
            self.assertFalse((pi_dir / ".hive-deploy-manifest.json").exists())
            self.assertFalse((pi_dir / ".hive-deploy-backups").exists())

    def test_wrong_package_version_fails_before_activation(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            seed_minimal_source(source)
            seed_installed_package(pi_dir)
            package_json = pi_dir / "npm/node_modules/pi-subagents/package.json"
            package_json.write_text('{"name":"pi-subagents","version":"9.9.9"}\n', encoding="utf-8")
            marker = pi_dir / "operator-owned.txt"
            marker.write_text("leave me\n", encoding="utf-8")

            result = run_helper(
                "deploy",
                "--repo-root",
                str(source),
                "--pi-dir",
                str(pi_dir),
                "--apply",
                check=False,
            )

            self.assertEqual(result.returncode, 2)
            self.assertIn("must be pi-subagents@0.67.0", result.stderr)
            self.assertEqual(marker.read_text(encoding="utf-8"), "leave me\n")
            self.assertFalse((pi_dir / ".hive-deploy-manifest.json").exists())

    def test_modified_managed_file_is_preserved_and_unmodified_orphan_is_removed(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            seed_minimal_source(source)
            hook = source / "global/hooks/flow-context/flow-context.sh"
            hook.write_text("#!/usr/bin/env bash\nprintf demo\n", encoding="utf-8")
            hook.chmod(0o755)
            (source / "harness/pi/agents/keep.md").write_text("keep\n", encoding="utf-8")

            arguments = ["deploy", "--repo-root", str(source), "--pi-dir", str(pi_dir), "--apply"]
            run_helper(*arguments)
            target = pi_dir / "global/hooks/flow-context/flow-context.sh"
            target.write_text("user edit\n", encoding="utf-8")
            (source / "harness/pi/agents/demo.md").unlink()

            result = run_helper(*arguments)
            self.assertIn("CONFLICT", result.stdout)
            self.assertEqual(target.read_text(encoding="utf-8"), "user edit\n")
            self.assertFalse((pi_dir / "agents/demo.md").exists())

    def test_user_config_edits_are_preserved_as_conflicts(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            arguments = ["deploy", "--repo-root", str(FIXTURE_ROOT), "--pi-dir", str(pi_dir), "--apply"]
            run_helper(*arguments)

            settings = json.loads((pi_dir / "settings.json").read_text(encoding="utf-8"))
            settings["packages"][1] = {
                "source": "npm:gentle-engram@9.9.9",
                "autoload": False,
                "skills": ["operator-skill"],
            }
            (pi_dir / "settings.json").write_text(json.dumps(settings), encoding="utf-8")
            subagent = json.loads((pi_dir / "extensions/subagent/config.json").read_text(encoding="utf-8"))
            subagent["forceTopLevelAsync"] = False
            (pi_dir / "extensions/subagent/config.json").write_text(json.dumps(subagent), encoding="utf-8")
            mcp = json.loads((pi_dir / "mcp.json").read_text(encoding="utf-8"))
            mcp["mcpServers"]["context7"]["url"] = "https://user.example/mcp"
            mcp["mcpServers"]["context7"]["directTools"] = True
            mcp["mcpServers"]["context7"]["lifecycle"] = "eager"
            (pi_dir / "mcp.json").write_text(json.dumps(mcp), encoding="utf-8")
            web = json.loads((pi_dir / "web-search.json").read_text(encoding="utf-8"))
            web["provider"] = "xai"
            (pi_dir / "web-search.json").write_text(json.dumps(web), encoding="utf-8")

            result = run_helper(*arguments)
            self.assertIn("CONFLICT", result.stdout)
            settings_after = json.loads((pi_dir / "settings.json").read_text(encoding="utf-8"))
            self.assertEqual(settings_after["packages"][1]["source"], "npm:gentle-engram@9.9.9")
            self.assertFalse(settings_after["packages"][1]["autoload"])
            self.assertEqual(settings_after["packages"][1]["skills"], ["operator-skill"])
            subagent_after = json.loads((pi_dir / "extensions/subagent/config.json").read_text(encoding="utf-8"))
            self.assertFalse(subagent_after["forceTopLevelAsync"])
            mcp_after = json.loads((pi_dir / "mcp.json").read_text(encoding="utf-8"))
            self.assertEqual(mcp_after["mcpServers"]["context7"]["url"], "https://user.example/mcp")
            self.assertTrue(mcp_after["mcpServers"]["context7"]["directTools"])
            self.assertEqual(mcp_after["mcpServers"]["context7"]["lifecycle"], "eager")
            web_after = json.loads((pi_dir / "web-search.json").read_text(encoding="utf-8"))
            self.assertEqual(web_after["provider"], "xai")

    def test_symlink_targets_are_preserved_without_following_them(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            source = Path(temp) / "source"
            pi_dir = Path(temp) / "pi-agent"
            outside = Path(temp) / "outside.txt"
            seed_minimal_source(source)
            (source / "harness/pi/src/demo.ts").write_text("source\n", encoding="utf-8")
            (source / "harness/pi/extensions/demo.ts").write_text("extension\n", encoding="utf-8")
            pi_dir.mkdir()
            outside.write_text("keep\n", encoding="utf-8")
            (pi_dir / "AGENTS.md").symlink_to(outside)
            for name in ("settings.json", "mcp.json", "web-search.json"):
                (pi_dir / name).symlink_to(outside)

            result = run_helper(
                "deploy",
                "--repo-root",
                str(source),
                "--pi-dir",
                str(pi_dir),
            )

            self.assertEqual(result.returncode, 0)
            self.assertGreaterEqual(result.stdout.count("symlink; preserved"), 4)
            self.assertEqual(outside.read_text(encoding="utf-8"), "keep\n")

    def test_removed_managed_config_is_preserved_as_a_conflict(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            arguments = ["deploy", "--repo-root", str(FIXTURE_ROOT), "--pi-dir", str(pi_dir), "--apply"]
            run_helper(*arguments)

            settings = json.loads((pi_dir / "settings.json").read_text(encoding="utf-8"))
            settings["packages"] = [item for item in settings["packages"] if not item.startswith("npm:gentle-engram@")]
            (pi_dir / "settings.json").write_text(json.dumps(settings), encoding="utf-8")
            subagent = json.loads((pi_dir / "extensions/subagent/config.json").read_text(encoding="utf-8"))
            subagent.pop("forceTopLevelAsync")
            (pi_dir / "extensions/subagent/config.json").write_text(json.dumps(subagent), encoding="utf-8")
            mcp = json.loads((pi_dir / "mcp.json").read_text(encoding="utf-8"))
            mcp["mcpServers"].pop("context7")
            (pi_dir / "mcp.json").write_text(json.dumps(mcp), encoding="utf-8")
            web = json.loads((pi_dir / "web-search.json").read_text(encoding="utf-8"))
            web.pop("workflow")
            (pi_dir / "web-search.json").write_text(json.dumps(web), encoding="utf-8")

            result = run_helper(*arguments)

            self.assertIn("user removed", result.stdout)
            settings_after = json.loads((pi_dir / "settings.json").read_text(encoding="utf-8"))
            self.assertFalse(any(item.startswith("npm:gentle-engram@") for item in settings_after["packages"]))
            subagent_after = json.loads((pi_dir / "extensions/subagent/config.json").read_text(encoding="utf-8"))
            self.assertNotIn("forceTopLevelAsync", subagent_after)
            mcp_after = json.loads((pi_dir / "mcp.json").read_text(encoding="utf-8"))
            self.assertNotIn("context7", mcp_after["mcpServers"])
            web_after = json.loads((pi_dir / "web-search.json").read_text(encoding="utf-8"))
            self.assertNotIn("workflow", web_after)

    def test_backup_payload_is_private_and_preserves_original_bytes(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            pi_dir.mkdir(parents=True)
            web_search = pi_dir / "web-search.json"
            original = b'{"provider":"openai","apiKey":"secret-placeholder"}\n'
            web_search.write_bytes(original)
            web_search.chmod(0o644)

            run_helper("deploy", "--repo-root", str(FIXTURE_ROOT), "--pi-dir", str(pi_dir), "--apply")

            backup_root = pi_dir / ".hive-deploy-backups"
            backup_dir = sorted(backup_root.iterdir())[-1]
            payload = backup_dir / "payload"
            backup_file = payload / "web-search.json"
            self.assertEqual(backup_root.stat().st_mode & 0o777, 0o700)
            self.assertEqual(backup_dir.stat().st_mode & 0o777, 0o700)
            self.assertEqual(payload.stat().st_mode & 0o777, 0o700)
            self.assertEqual(backup_file.stat().st_mode & 0o777, 0o600)
            self.assertEqual(backup_file.read_bytes(), original)

    def test_rollback_restores_first_apply(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            arguments = ["deploy", "--repo-root", str(FIXTURE_ROOT), "--pi-dir", str(pi_dir), "--apply"]
            metadata = json.loads((FIXTURE_ROOT / PATCH_FILES[1]).read_text(encoding="utf-8"))
            run_helper(*arguments)
            backup = sorted((pi_dir / ".hive-deploy-backups").iterdir())[-1]
            for record in metadata["targets"]:
                target = pi_dir / "npm/node_modules/pi-subagents" / record["path"]
                self.assertEqual(sha256_file(target), record["afterSha256"])
            result = run_helper("rollback", "--pi-dir", str(pi_dir), "--backup-dir", str(backup), "--apply")

            self.assertIn("PI rollback", result.stdout)
            self.assertFalse((pi_dir / "AGENTS.md").exists())
            self.assertFalse((pi_dir / "settings.json").exists())
            for record in metadata["targets"]:
                target = pi_dir / "npm/node_modules/pi-subagents" / record["path"]
                pristine = PACKAGE_FIXTURE_ROOT / record["path"]
                self.assertEqual(target.read_bytes(), pristine.read_bytes())
            self.assertTrue((pi_dir / ".hive-deploy-backups").exists())

    def test_rollback_preserves_post_deploy_edits(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            pi_dir = Path(temp) / "pi-agent"
            run_helper("deploy", "--repo-root", str(FIXTURE_ROOT), "--pi-dir", str(pi_dir), "--apply")
            backup = sorted((pi_dir / ".hive-deploy-backups").iterdir())[-1]
            target = pi_dir / "AGENTS.md"
            target.write_text("user edit\n", encoding="utf-8")

            result = run_helper(
                "rollback",
                "--pi-dir",
                str(pi_dir),
                "--backup-dir",
                str(backup),
                "--apply",
                check=False,
            )

            self.assertEqual(result.returncode, 2)
            self.assertIn("CONFLICT", result.stdout)
            self.assertEqual(target.read_text(encoding="utf-8"), "user edit\n")


if __name__ == "__main__":
    unittest.main()
