#!/usr/bin/env python3
"""Parity between the harness-audit skill and its provenance catalog."""

from __future__ import annotations

import re
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
CATALOG = ROOT / "_support/docs/harness-engineering/harness-audit-rules.md"
DEPLOYED = [ROOT / "content/skills/harness-audit", ROOT / "content/agents/review/hive-review-harness.md"]
RULE_ID = re.compile(r"HA-[A-Z]+-[0-9]+")
ACTIVE_ROW = re.compile(r"^\| (HA-[A-Z]+-[0-9]+) \|.*\| active \|$", re.MULTILINE)


def active_ids(catalog: str) -> set[str]:
    return set(ACTIVE_ROW.findall(catalog))


def deployed_ids(paths: list[Path]) -> set[str]:
    files = [f for p in paths for f in ([p] if p.is_file() else sorted(p.rglob("*.md")))]
    return {m for f in files for m in RULE_ID.findall(f.read_text(encoding="utf-8"))}


class HarnessAuditRulesTest(unittest.TestCase):
    def test_active_catalog_ids_match_deployed_ids(self) -> None:
        catalog, deployed = active_ids(CATALOG.read_text(encoding="utf-8")), deployed_ids(DEPLOYED)
        self.assertTrue(catalog, "catalog has no active rules")
        self.assertEqual(sorted(catalog - deployed), [], "active in catalog but absent from the skill")
        self.assertEqual(sorted(deployed - catalog), [], "used by the skill but not active in the catalog")

    def test_retired_and_referenced_ids_are_not_active(self) -> None:
        catalog = "| HA-IF-01 | rule (see HA-IF-09) | x | x | x | 2026-09-22 | active |\n" \
                  "| HA-IF-02 | old rule | x | x | x | 2026-09-22 | retired |\n"
        self.assertEqual(active_ids(catalog), {"HA-IF-01"})


if __name__ == "__main__":
    unittest.main()
