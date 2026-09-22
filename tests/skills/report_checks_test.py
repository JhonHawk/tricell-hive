#!/usr/bin/env python3
"""Behavioral invariants for the flow-report checker."""

from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CHECKER = ROOT / "content/skills/flow-report/scripts/self_check.py"


class ReportChecksTest(unittest.TestCase):
    def check(self, html: str) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "report.html"
            path.write_text(html, encoding="utf-8")
            return subprocess.run(
                [sys.executable, str(CHECKER), str(path)], text=True, capture_output=True, check=False
            )

    def test_offline_report_passes(self) -> None:
        result = self.check("<!doctype html><title>Report</title><p>Readable offline.</p>")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_remote_asset_requires_declaration(self) -> None:
        result = self.check('<img src="https://example.test/chart.svg">')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("flow-report-assets", result.stdout)

    def test_declared_remote_asset_passes(self) -> None:
        result = self.check(
            '<meta name="flow-report-assets" content="external"><img src="https://example.test/chart.svg">'
        )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_inline_script_in_svg_fails(self) -> None:
        result = self.check("<svg><script>alert(1)</script></svg>")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("inside an <svg>", result.stdout)


if __name__ == "__main__":
    unittest.main()
