#!/usr/bin/env python3
"""Executable contract for the walkthrough manifest and its generator."""

from __future__ import annotations

import json
import re
import shutil
import struct
import subprocess
import sys
import tempfile
import unittest
import zlib
from html.parser import HTMLParser
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
GENERATOR = ROOT / "content/skills/flow-walkthrough/scripts/render_walkthrough.py"
CHECKER = ROOT / "content/skills/flow-report/scripts/self_check.py"
REFERENCE = ROOT / "content/skills/flow-walkthrough/references/walkthrough-manifest.md"
FIXTURES = ROOT / "tests/fixtures/walkthrough"

# Field each invalid fixture must name in the error message.
INVALID_FIELDS = {
    "unknown-field": "bogus",
    "missing-required": "title",
    "schema-version": "schema_version",
    "duplicate-flow-id": "flows[1].id",
    "bad-flow-id": "flows[0].id",
    "absolute-path": "flows[0].captures[0].path",
    "dotdot-path": "flows[0].captures[0].path",
    "outside-folder": "flows[0].captures[0].path",
    "bad-segment": "flows[0].captures[0].path",
    "symlink": "flows[0].captures[0].path",
    "capture-extension": "flows[0].captures[0].path",
    "recording-extension": "flows[0].recording.path",
    "missing-file": "flows[0].captures[0].path",
    "bad-viewport": "flows[0].captures[0].viewport",
}

DATA_BLOCK = re.compile(
    r'<script type="application/json" id="walkthrough-data">(.*?)</script>', re.DOTALL
)


def tiny_png() -> bytes:
    def chunk(kind: bytes, data: bytes) -> bytes:
        body = struct.pack(">I", len(data)) + kind + data
        return body + struct.pack(">I", zlib.crc32(kind + data) & 0xFFFFFFFF)

    raw = b"\x00" + b"\x20\x80\xc0"
    return (
        b"\x89PNG\r\n\x1a\n"
        + chunk(b"IHDR", struct.pack(">IIBBBBB", 1, 1, 8, 2, 0, 0, 0))
        + chunk(b"IDAT", zlib.compress(raw))
        + chunk(b"IEND", b"")
    )


def run_generator(folder: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(GENERATOR), str(folder)], text=True, capture_output=True, check=False
    )


def run_checker(page: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(CHECKER), str(page)], text=True, capture_output=True, check=False
    )


class EventAttributes(HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=True)
        self.found: list[str] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        self.found.extend(f"{tag}[{name}]" for name, _ in attrs if name.casefold().startswith("on"))


class WalkthroughRenderTest(unittest.TestCase):
    def setUp(self) -> None:
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.tmp = Path(self._tmp.name)

    def copy_fixture(self, relative: str) -> Path:
        target = self.tmp / relative.replace("/", "-")
        shutil.copytree(FIXTURES / relative, target, symlinks=True)
        return target

    def make_folder(self, manifest: dict, files: tuple[str, ...] = ("images/a.png",)) -> Path:
        folder = self.tmp / "case"
        shutil.rmtree(folder, ignore_errors=True)
        for name in files:
            path = folder / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(tiny_png())
        (folder / "walkthrough.json").write_text(json.dumps(manifest), encoding="utf-8")
        return folder

    def minimal(self, **overrides: object) -> dict:
        manifest = json.loads((FIXTURES / "valid/first-view/walkthrough.json").read_text("utf-8"))
        manifest["flows"] = [
            {"id": "one", "title": "One", "captures": [{"path": "images/a.png", "caption": "A"}]}
        ]
        manifest.update(overrides)
        return manifest

    def assert_rejected(self, folder: Path, field: str) -> None:
        result = run_generator(folder)
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertIn(field, result.stderr)
        self.assertFalse((folder / "walkthrough.html").exists())
        self.assertEqual(sorted(p.name for p in folder.iterdir() if p.suffix == ".tmp"), [])

    def test_delivery_contract_shared_with_go(self) -> None:
        cases = json.loads((FIXTURES / "delivery-contract.json").read_text("utf-8"))
        for case in cases:
            with self.subTest(case=case["name"]):
                manifest = self.minimal()
                if not case.get("omit"):
                    manifest["delivery"] = case["delivery"]
                folder = self.make_folder(manifest)
                if "field" in case:
                    self.assert_rejected(folder, case["field"])
                else:
                    result = run_generator(folder)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    data = DATA_BLOCK.search((folder / "walkthrough.html").read_text("utf-8"))
                    self.assertIsNotNone(data)
                    self.assertEqual(json.loads(data.group(1)), manifest)

    def test_invalid_delivery_preserves_previous_page_and_source(self) -> None:
        folder = self.make_folder(self.minimal(delivery=[{
            "kind": "commit", "label": "abc1234", "url": "https://example.com/commit/abc1234"
        }]))
        self.assertEqual(run_generator(folder).returncode, 0)
        page = (folder / "walkthrough.html").read_bytes()
        manifest = self.minimal(delivery=[{"kind": "commit", "label": "abc1234", "url": "javascript:alert(1)"}])
        source = json.dumps(manifest).encode()
        (folder / "walkthrough.json").write_bytes(source)
        result = run_generator(folder)
        self.assertEqual(result.returncode, 1)
        self.assertIn("delivery[0].url", result.stderr)
        self.assertEqual((folder / "walkthrough.html").read_bytes(), page)
        self.assertEqual((folder / "walkthrough.json").read_bytes(), source)
        self.assertFalse(list(folder.glob("*.tmp")))

    # AC1
    def test_valid_final_generates_page_that_passes_self_check(self) -> None:
        folder = self.copy_fixture("valid/final")
        result = run_generator(folder)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        page = folder / "walkthrough.html"
        self.assertTrue(page.is_file())
        self.assertIn(str(page), result.stdout)
        checked = run_checker(page)
        self.assertEqual(checked.returncode, 0, checked.stdout + checked.stderr)

    def test_every_valid_fixture_generates(self) -> None:
        names = sorted(p.name for p in (FIXTURES / "valid").iterdir() if p.is_dir())
        self.assertTrue({"final", "first-view", "extreme", "viewports"} <= set(names))
        for name in names:
            with self.subTest(fixture=name):
                folder = self.copy_fixture(f"valid/{name}")
                result = run_generator(folder)
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                checked = run_checker(folder / "walkthrough.html")
                self.assertEqual(checked.returncode, 0, checked.stdout + checked.stderr)

    # AC2
    def test_every_invalid_fixture_names_the_field_and_writes_nothing(self) -> None:
        names = sorted(p.name for p in (FIXTURES / "invalid").iterdir() if p.is_dir())
        self.assertEqual(names, sorted(INVALID_FIELDS))
        for name in names:
            with self.subTest(fixture=name):
                folder = self.copy_fixture(f"invalid/{name}")
                if name == "symlink":
                    (folder / "images/link.png").symlink_to("real.png")
                self.assert_rejected(folder, INVALID_FIELDS[name])

    def test_missing_manifest_is_rejected(self) -> None:
        folder = self.tmp / "empty"
        folder.mkdir()
        self.assert_rejected(folder, "walkthrough.json")

    def test_invalid_json_is_rejected(self) -> None:
        folder = self.tmp / "broken"
        folder.mkdir()
        (folder / "walkthrough.json").write_text("{not json", encoding="utf-8")
        self.assert_rejected(folder, "walkthrough.json")

    def test_wrong_types_are_rejected(self) -> None:
        for field, value, expected in (
            ("schema_version", True, "schema_version"),
            ("kind", "draft", "kind"),
            ("kind", ["final"], "kind"),
            ("checks", [{"name": "x", "status": {"a": 1}}], "checks[0].status"),
            ("flows", [], "flows"),
            ("checks", [{"name": "x", "status": "maybe"}], "checks[0].status"),
            ("created_at", "yesterday", "created_at"),
        ):
            with self.subTest(field=field):
                folder = self.make_folder(self.minimal(**{field: value}))
                self.assert_rejected(folder, expected)

    def test_created_at_accepts_a_trailing_z(self) -> None:
        # datetime.fromisoformat rejects "Z" before Python 3.11.
        for value in ("2026-10-18T22:00:00Z", "2026-10-18T22:00:00z", "2026-10-18T22:00:00.5Z"):
            with self.subTest(created_at=value):
                folder = self.make_folder(self.minimal(created_at=value))
                result = run_generator(folder)
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_overflowing_numbers_are_rejected(self) -> None:
        manifest = self.minimal(
            flows=[
                {
                    "id": "one",
                    "title": "One",
                    "captures": [],
                    "recording": {
                        "status": "recorded",
                        "path": "images/v.mp4",
                        "candidate": "abc",
                        "duration_seconds": "DURATION",
                    },
                }
            ]
        )
        for literal, expected in (
            ("1e999", "duration_seconds"),
            ("-1e999", "duration_seconds"),
            ("1" + "0" * 400, "duration_seconds"),
            ("Infinity", "walkthrough.json"),
            ("NaN", "walkthrough.json"),
        ):
            with self.subTest(literal=literal[:12]):
                folder = self.make_folder(manifest, ("images/v.mp4",))
                text = (folder / "walkthrough.json").read_text(encoding="utf-8")
                (folder / "walkthrough.json").write_text(
                    text.replace('"DURATION"', literal), encoding="utf-8"
                )
                self.assert_rejected(folder, expected)

    def test_recording_status_requirements(self) -> None:
        def flow(recording: dict) -> dict:
            return self.minimal(
                flows=[{"id": "one", "title": "One", "captures": [], "recording": recording}]
            )

        cases = (
            ({"status": "recorded", "path": "images/v.mp4", "candidate": "abc"}, "duration_seconds"),
            ({"status": "not_recorded"}, "reason"),
            ({"status": "stale", "path": "images/v.mp4"}, "candidate"),
            ({"status": "none", "path": "images/v.mp4"}, "path"),
        )
        for recording, expected in cases:
            with self.subTest(recording=recording):
                folder = self.make_folder(flow(recording), ("images/v.mp4",))
                self.assert_rejected(folder, f"recording.{expected}")

    def test_limits(self) -> None:
        many_flows = self.minimal(
            flows=[{"id": f"f{i}", "title": "F", "captures": []} for i in range(51)]
        )
        self.assert_rejected(self.make_folder(many_flows), "flows")
        many_captures = self.minimal(
            flows=[
                {
                    "id": "f",
                    "title": "F",
                    "captures": [{"path": "images/a.png", "caption": "c"}] * 101,
                }
            ]
        )
        self.assert_rejected(self.make_folder(many_captures), "flows[0].captures")
        huge = self.minimal(result="x" * (1024 * 1024 + 1))
        self.assert_rejected(self.make_folder(huge), "walkthrough.json")
        folder = self.make_folder(self.minimal())
        with (folder / "images/a.png").open("r+b") as handle:
            handle.truncate(20 * 1024 * 1024 + 1)
        self.assert_rejected(folder, "flows[0].captures[0].path")

    # AC5, static part
    def test_static_html_is_inert(self) -> None:
        folder = self.copy_fixture("valid/final")
        self.assertEqual(run_generator(folder).returncode, 0)
        html = (folder / "walkthrough.html").read_text(encoding="utf-8")

        match = DATA_BLOCK.search(html)
        self.assertIsNotNone(match, "data block missing")
        block = match.group(1)
        self.assertIn("\\u003c/script\\u003e", block)
        self.assertNotIn("<", block)
        self.assertNotIn(">", block)
        self.assertNotIn("&", block)
        manifest = json.loads((FIXTURES / "valid/final/walkthrough.json").read_text("utf-8"))
        self.assertEqual(json.loads(block), manifest)

        self.assertEqual(html.count("<script"), html.count("</script>"))
        outside = html[: match.start()] + html[match.end() :]
        parser = EventAttributes()
        parser.feed(outside)
        self.assertEqual(parser.found, [])
        self.assertNotIn("onerror", outside)
        self.assertNotIn("https://", html)
        self.assertNotIn("http://", html.replace("http://localhost:3000", ""))

    # AC20, static part
    def test_viewports_fixture_embeds_both_views(self) -> None:
        folder = self.copy_fixture("valid/viewports")
        self.assertEqual(run_generator(folder).returncode, 0)
        block = DATA_BLOCK.search((folder / "walkthrough.html").read_text("utf-8")).group(1)
        captures = json.loads(block)["flows"][0]["captures"]
        self.assertEqual([c["viewport"] for c in captures], ["desktop"] * 3 + ["mobile"] * 2)

    def test_viewport_is_optional_and_restricted(self) -> None:
        manifest = self.minimal()
        self.assertEqual(run_generator(self.make_folder(manifest)).returncode, 0)
        manifest["flows"][0]["captures"][0]["viewport"] = "mobile"
        self.assertEqual(run_generator(self.make_folder(manifest)).returncode, 0)
        manifest["flows"][0]["captures"][0]["viewport"] = 3
        self.assert_rejected(self.make_folder(manifest), "flows[0].captures[0].viewport")

    def test_generation_replaces_previous_page(self) -> None:
        folder = self.copy_fixture("valid/first-view")
        (folder / "walkthrough.html").write_text("old", encoding="utf-8")
        self.assertEqual(run_generator(folder).returncode, 0)
        self.assertIn("walkthrough-data", (folder / "walkthrough.html").read_text("utf-8"))

    # H3: a planted walkthrough.html.tmp link must not be followed.
    def test_planted_temp_symlink_is_not_followed(self) -> None:
        folder = self.copy_fixture("valid/first-view")
        victim = self.tmp / "victim.txt"
        victim.write_text("precious", encoding="utf-8")
        (folder / "walkthrough.html.tmp").symlink_to(victim)
        result = run_generator(folder)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(victim.read_text(encoding="utf-8"), "precious")
        page = folder / "walkthrough.html"
        self.assertFalse(page.is_symlink())
        self.assertIn("walkthrough-data", page.read_text("utf-8"))
        leftovers = [p.name for p in folder.iterdir() if p.name.startswith(".") or p.suffix == ".tmp"]
        self.assertEqual(leftovers, ["walkthrough.html.tmp"], "only the planted link may remain")

    # I1: parity with the Go validator.
    def test_trailing_newline_is_not_accepted_in_patterns(self) -> None:
        manifest = self.minimal()
        manifest["flows"][0]["id"] = "one\n"
        self.assert_rejected(self.make_folder(manifest), "flows[0].id")
        self.assert_rejected(self.make_folder(self.minimal(lang="es\n")), "lang")
        manifest = self.minimal()
        manifest["flows"][0]["captures"][0]["path"] = "images\n/a.png"
        folder = self.make_folder(manifest, ("images\n/a.png",))
        self.assert_rejected(folder, "flows[0].captures[0].path")

    def test_manifest_that_is_a_symlink_is_rejected(self) -> None:
        folder = self.copy_fixture("valid/first-view")
        real = self.tmp / "elsewhere.json"
        real.write_bytes((folder / "walkthrough.json").read_bytes())
        (folder / "walkthrough.json").unlink()
        (folder / "walkthrough.json").symlink_to(real)
        self.assert_rejected(folder, "walkthrough.json")

    def test_non_finite_numbers_are_rejected(self) -> None:
        for token in ("NaN", "Infinity", "-Infinity"):
            with self.subTest(token=token):
                manifest = self.minimal(
                    flows=[
                        {
                            "id": "one",
                            "title": "One",
                            "captures": [],
                            "recording": {
                                "status": "recorded",
                                "path": "images/v.mp4",
                                "duration_seconds": 1,
                                "candidate": "abc",
                            },
                        }
                    ]
                )
                folder = self.make_folder(manifest, ("images/v.mp4",))
                path = folder / "walkthrough.json"
                path.write_text(path.read_text("utf-8").replace('"duration_seconds": 1', f'"duration_seconds": {token}'), "utf-8")
                self.assertIn(token, path.read_text("utf-8"))
                self.assert_rejected(folder, "walkthrough.json")

    # H4: attacker-chosen field names are bounded in the error message.
    def test_unknown_field_name_is_bounded_and_printable(self) -> None:
        manifest = self.minimal()
        manifest["x" * 5000 + "\x1b[31m\nINJECT"] = True
        folder = self.make_folder(manifest)
        result = run_generator(folder)
        self.assertEqual(result.returncode, 1)
        self.assertLess(len(result.stderr), 600)
        self.assertNotIn("\x1b", result.stderr)
        self.assertEqual(result.stderr.count("\n"), 1)

    # Reference example stays valid.
    def test_reference_example_passes_the_validator(self) -> None:
        if not REFERENCE.is_file():
            self.skipTest("walkthrough-manifest.md does not exist yet")
        text = REFERENCE.read_text(encoding="utf-8")
        match = re.search(r"```json\n(.*?)\n```", text, re.DOTALL)
        self.assertIsNotNone(match, "reference has no json block")
        manifest = json.loads(match.group(1))
        folder = self.tmp / "reference"
        folder.mkdir()
        (folder / "walkthrough.json").write_text(match.group(1), encoding="utf-8")
        paths = []
        for flow in manifest.get("flows", []):
            paths.extend(capture["path"] for capture in flow.get("captures", []))
            recording = flow.get("recording") or {}
            if "path" in recording:
                paths.append(recording["path"])
        for relative in paths:
            target = folder / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(tiny_png() if relative.endswith(".png") else b"stub")
        result = run_generator(folder)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
