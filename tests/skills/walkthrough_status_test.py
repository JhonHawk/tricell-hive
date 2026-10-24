"""Explicit work status and safe viewer confirmation in disposable folders."""

import importlib.util
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/flow-walkthrough/scripts/render_walkthrough.py"
spec = importlib.util.spec_from_file_location("walkthrough_status_renderer", SCRIPT)
renderer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(renderer)


class WalkthroughStatusTest(unittest.TestCase):
    def setUp(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        self.root = Path(temporary.name)
        self.empty_path = self.root / "empty-path"
        self.empty_path.mkdir()
        unavailable = mock.patch.object(shutil, "which", return_value=None)
        unavailable.start()
        self.addCleanup(unavailable.stop)
        self.folder = self.root / "current work"
        self.folder.mkdir()
        self.manifest = {
            "schema_version": 1, "lang": "en", "kind": "final", "title": "Current work",
            "session": {"id": "current"}, "candidate": {"branch": "test", "commit": "working tree"},
            "created_at": "2026-10-19T00:00:00Z",
            "flows": [{"id": "one", "title": "One", "captures": []}],
        }
        self.manifest_path = self.folder / "walkthrough.json"
        self.manifest_path.write_text(json.dumps(self.manifest))

    def run_renderer(self, *flags, env=None):
        return subprocess.run([sys.executable, str(SCRIPT), str(self.folder), *flags],
                              capture_output=True, text=True,
                              env=env if env is not None else {**os.environ, "PATH": str(self.empty_path)},
                              timeout=20)

    def fake_cli(self, body):
        binary = self.root / "bin"
        binary.mkdir(exist_ok=True)
        cli = binary / "hive"
        cli.write_text(f"#!{sys.executable}\nimport json, sys\nfrom pathlib import Path\n" + body)
        cli.chmod(0o700)
        return {**os.environ, "PATH": str(binary)}

    def test_binary_present_uses_json_even_when_daemon_is_stopped(self):
        marker = self.root / "cli-called"
        environment = self.fake_cli(f"Path({str(marker)!r}).write_text('called')\nsys.exit(1)\n")
        for previous_page in (None, b"historical HTML\x00\xff"):
            with self.subTest(previous_page=previous_page):
                page = self.folder / "walkthrough.html"
                if previous_page is not None:
                    page.write_bytes(previous_page)
                before = self.manifest_path.read_bytes()
                result = self.run_renderer(env=environment)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(self.manifest_path.read_bytes(), before)
                self.assertEqual(result.stdout.strip(), str(self.manifest_path))
                self.assertFalse(marker.exists(), "publication must not query daemon availability")
                if previous_page is None:
                    self.assertFalse(page.exists())
                else:
                    self.assertEqual(page.read_bytes(), previous_page)

    def test_binary_present_never_builds_page_and_preserves_metadata(self):
        delivery = [{"kind": "commit", "label": "abc1234", "url": "https://example.com/commit/abc1234"}]
        original = {**self.manifest, "delivery": delivery}
        self.manifest_path.write_text(json.dumps(original))
        old_page = self.folder / "walkthrough.html"
        old_page.write_bytes(b"prior evidence")
        with mock.patch.object(shutil, "which", return_value="/fake/hive") as which, mock.patch.object(renderer, "build_page", side_effect=AssertionError("must not build")) as build, mock.patch.object(renderer, "ASSETS", self.root / "missing-assets"):
            result = renderer.render(self.folder, "closed")
        which.assert_called_once_with("hive")
        build.assert_not_called()
        self.assertEqual(result, self.manifest_path)
        written = json.loads(self.manifest_path.read_text())
        self.assertEqual(written["session"]["status"], "closed")
        self.assertEqual(written["candidate"], original["candidate"])
        self.assertEqual(written["delivery"], delivery)
        self.assertEqual(old_page.read_bytes(), b"prior evidence")
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_binary_present_invalid_manifest_does_not_publish(self):
        self.manifest["delivery"] = [{"kind": "commit", "label": "abc", "url": "javascript:alert(1)"}]
        self.manifest_path.write_text(json.dumps(self.manifest))
        before = self.manifest_path.read_bytes()
        result = self.run_renderer("--session-status", "closed", env=self.fake_cli("sys.exit(1)\n"))
        self.assertEqual(result.returncode, 1)
        self.assertIn("delivery[0].url", result.stderr)
        self.assertEqual(self.manifest_path.read_bytes(), before)
        self.assertFalse((self.folder / "walkthrough.html").exists())
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_binary_present_concurrent_change_is_preserved(self):
        stage = renderer.stage_atomic
        def concurrent_update(path, content):
            staged = stage(path, content)
            self.manifest_path.write_text('{"concurrent":"work"}')
            return staged
        with mock.patch.object(shutil, "which", return_value="/fake/hive"), mock.patch.object(renderer, "stage_atomic", side_effect=concurrent_update), mock.patch.object(renderer, "build_page", side_effect=AssertionError("must not build")):
            with self.assertRaises(renderer.ManifestError):
                renderer.render(self.folder, "closed")
        self.assertEqual(self.manifest_path.read_text(), '{"concurrent":"work"}')
        self.assertFalse((self.folder / "walkthrough.html").exists())
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_binary_absent_returns_html_fallback(self):
        before = self.manifest_path.read_bytes()
        result = self.run_renderer()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), str(self.folder / "walkthrough.html"))
        self.assertTrue((self.folder / "walkthrough.html").is_file())
        self.assertEqual(self.manifest_path.read_bytes(), before)

    def test_absent_binary_confirmation_keeps_html_fallback_and_closed_source(self):
        result = self.run_renderer("--session-status", "closed", "--confirm-viewer")
        self.assertEqual(result.returncode, 2)
        self.assertIn("retain", result.stderr)
        self.assertEqual(json.loads(self.manifest_path.read_text())["session"]["status"], "closed")
        self.assertTrue((self.folder / "walkthrough.html").is_file())
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_optional_status_preserves_old_manifest(self):
        before = self.manifest_path.read_bytes()
        result = self.run_renderer()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.manifest_path.read_bytes(), before)
        self.assertNotIn("status", json.loads(before)["session"])

    def test_status_validation(self):
        for status in ("in_progress", "closed"):
            value = {**self.manifest, "session": {"id": "current", "status": status}}
            renderer.validate(self.folder, value)
        for status in (None, "unknown", "final", True, [], {}):
            with self.subTest(status=status), self.assertRaises(renderer.ManifestError):
                renderer.validate(self.folder, {**self.manifest, "session": {"id": "current", "status": status}})

    def test_close_reopen_idempotent_and_other_work_untouched(self):
        other = self.root / "other.json"
        other.write_text('{"session":{"id":"other","status":"in_progress"}}')
        original_other = other.read_bytes()
        for status in ("closed", "closed", "in_progress"):
            result = self.run_renderer("--session-status", status)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(json.loads(self.manifest_path.read_text())["session"]["status"], status)
            self.assertIn(f'"status":"{status}"', (self.folder / "walkthrough.html").read_text())
        self.assertEqual(other.read_bytes(), original_other)
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_delivery_survives_closure_without_changing_candidate_or_other_work(self):
        delivery = [{"kind": "pull_request", "label": "example #123", "url": "https://example.com/repository/pull/123"}]
        current = {**self.manifest, "delivery": delivery}
        self.manifest_path.write_text(json.dumps(current))
        other = self.root / "other-work"
        other.mkdir()
        other_manifest = other / "walkthrough.json"
        other_manifest.write_text(json.dumps({**self.manifest, "session": {"id": "other"}}))
        previous_other = other_manifest.read_bytes()
        for status in ("closed", "in_progress"):
            result = self.run_renderer("--session-status", status)
            self.assertEqual(result.returncode, 0, result.stderr)
            written = json.loads(self.manifest_path.read_text())
            self.assertEqual(written["delivery"], delivery)
            self.assertEqual(written["candidate"], self.manifest["candidate"])
            self.assertEqual(written["session"]["status"], status)
            self.assertIn('"delivery":', (self.folder / "walkthrough.html").read_text())
            self.assertEqual(other_manifest.read_bytes(), previous_other)

    def test_confirmation_checks_current_closed_source_with_structured_arguments(self):
        environment = self.fake_cli(
            "assert sys.argv[1:4] == ['view', 'status', '--dir']\n"
            "folder = Path(sys.argv[4])\n"
            "assert folder.is_absolute()\n"
            "assert json.loads((folder / 'walkthrough.json').read_text())['session']['status'] == 'closed'\n"
            "assert not (folder / 'walkthrough.html').exists()\n"
            "print(json.dumps({'running': True, 'found': True, 'session_status': 'closed'}))\n")
        result = self.run_renderer("--session-status", "closed", "--confirm-viewer", env=environment)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_offline_confirmation_keeps_unit_and_sanitizes_cli_output(self):
        environment = self.fake_cli("print('PRIVATE_CLI_OUTPUT', file=sys.stderr)\nsys.exit(1)\n")
        image = self.folder / "evidence.png"
        image.write_bytes(b"evidence")
        result = self.run_renderer("--session-status", "closed", "--confirm-viewer", env=environment)
        self.assertEqual(result.returncode, 2)
        self.assertNotIn("PRIVATE_CLI_OUTPUT", result.stdout + result.stderr)
        self.assertIn("retain", result.stderr)
        self.assertEqual(image.read_bytes(), b"evidence")
        self.assertEqual(json.loads(self.manifest_path.read_text())["session"]["status"], "closed")
        self.assertFalse((self.folder / "walkthrough.html").exists())

    def test_timeout_and_stale_copy_do_not_confirm(self):
        for payload in ({"found": False, "session_status": "unknown"},
                        {"found": True, "session_status": "in_progress"},
                        {"found": "true", "session_status": "closed"}, {}):
            with self.subTest(payload=payload), mock.patch.object(renderer.subprocess, "run", return_value=subprocess.CompletedProcess([], 0, json.dumps(payload), "")), mock.patch.object(renderer.time, "monotonic", side_effect=[0, 0, 16]):
                self.assertFalse(renderer.confirm_closed(self.folder))
        with mock.patch.object(renderer.subprocess, "run", side_effect=subprocess.TimeoutExpired("hive", 15)):
            self.assertFalse(renderer.confirm_closed(self.folder))

    def test_prepare_failure_leaves_manifest_and_page_unchanged(self):
        before = self.manifest_path.read_bytes()
        page = self.folder / "walkthrough.html"
        page.write_text("old page")
        with mock.patch.object(renderer, "build_page", side_effect=OSError("cannot build")):
            with self.assertRaises(OSError):
                renderer.render(self.folder, "closed")
        self.assertEqual(self.manifest_path.read_bytes(), before)
        self.assertEqual(page.read_text(), "old page")

    def test_page_publish_failure_restores_manifest(self):
        before = self.manifest_path.read_bytes()
        page = self.folder / "walkthrough.html"
        page.write_text("old page")
        replace = renderer.os.replace
        def fail_page(source, target):
            if Path(target) == page:
                raise OSError("cannot publish page")
            return replace(source, target)
        with mock.patch.object(renderer.os, "replace", side_effect=fail_page):
            with self.assertRaises(OSError):
                renderer.render(self.folder, "closed")
        self.assertEqual(self.manifest_path.read_bytes(), before)
        self.assertEqual(page.read_text(), "old page")
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_page_stage_failure_does_not_publish_manifest(self):
        before = self.manifest_path.read_bytes()
        page = self.folder / "walkthrough.html"
        page.write_text("old page")
        stage = renderer.stage_atomic
        def fail_page(path, content):
            if path == page:
                raise OSError("cannot stage page")
            return stage(path, content)
        with mock.patch.object(renderer, "stage_atomic", side_effect=fail_page):
            with self.assertRaises(OSError):
                renderer.render(self.folder, "closed")
        self.assertEqual(self.manifest_path.read_bytes(), before)
        self.assertEqual(page.read_text(), "old page")
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_concurrent_manifest_change_is_preserved(self):
        def concurrent_update(manifest):
            self.manifest_path.write_text('{"concurrent":"work"}')
            return "candidate page"
        with mock.patch.object(renderer, "build_page", side_effect=concurrent_update):
            with self.assertRaises(renderer.ManifestError):
                renderer.render(self.folder, "closed")
        self.assertEqual(self.manifest_path.read_text(), '{"concurrent":"work"}')
        self.assertFalse((self.folder / "walkthrough.html").exists())
        self.assertFalse(list(self.folder.glob(".walkthrough-*.tmp")))

    def test_confirmation_requires_explicit_close(self):
        before = self.manifest_path.read_bytes()
        result = self.run_renderer("--session-status", "in_progress", "--confirm-viewer")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(self.manifest_path.read_bytes(), before)

    def test_status_update_preserves_manifest_permissions(self):
        self.manifest_path.chmod(0o600)
        result = self.run_renderer("--session-status", "closed")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.manifest_path.stat().st_mode & 0o777, 0o600)


if __name__ == "__main__":
    unittest.main()
