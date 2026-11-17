import importlib.util
import pathlib
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/flow-plan/scripts/business_rules.py"


def load_module():
    spec = importlib.util.spec_from_file_location("business_rules", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class BusinessRulesScriptTests(unittest.TestCase):
    def setUp(self):
        self.mod = load_module()

    def test_first_write_appends_section_preserving_prefix_byte_for_byte(self):
        desc = "# Ticket title\n\nOriginal author text that must not change.\n"
        heading = "Reglas de negocio · export-csv"
        body = "2026-11-16\n\n- **Decididas:**\n  - Solo administradores exportan CSV (2026-11-16)."
        res = self.mod.splice_description(desc, heading, body)
        self.assertTrue(res["ok"])
        self.assertTrue(res["changed"])
        updated = res["description"]
        self.assertTrue(updated.startswith(desc))
        self.assertIn(heading, updated)
        self.assertIn(f"{heading} · end", updated)

    def test_second_write_replaces_only_between_markers_and_skips_when_unchanged_ignoring_date(self):
        prefix = "Intro paragraph.\n\n"
        suffix = "\n\nFooter paragraph after rules.\n"
        heading = "Business rules · export-csv"
        old_section = (
            f"{heading}\n2026-11-10\n\n- **Decided:**\n  - Rule A (2026-11-10).\n{heading} · end"
        )
        desc = prefix + old_section + suffix

        # Same content with newer write date -> skipped (changed == False)
        same_body_new_date = "2026-11-16\n\n- **Decided:**\n  - Rule A (2026-11-10)."
        res_skip = self.mod.splice_description(desc, heading, same_body_new_date)
        self.assertTrue(res_skip["ok"])
        self.assertFalse(res_skip["changed"])
        self.assertEqual(res_skip["description"], desc)

        # Updated rules -> replaces only inside markers, keeping prefix and suffix byte-for-byte
        new_body = "2026-11-16\n\n- **Decided:**\n  - Rule A (2026-11-10).\n  - Rule B (2026-11-16)."
        res_upd = self.mod.splice_description(desc, heading, new_body)
        self.assertTrue(res_upd["ok"])
        self.assertTrue(res_upd["changed"])
        self.assertTrue(res_upd["description"].startswith(prefix + heading + "\n"))
        self.assertTrue(res_upd["description"].endswith(f"\n{heading} · end" + suffix))

    def test_orphan_or_duplicate_markers_refuse_to_edit_description(self):
        heading = "Business rules · export-csv"
        orphan = f"Intro\n\n{heading}\n2026-11-16\nno closing marker\n"
        res_orphan = self.mod.splice_description(orphan, heading, "2026-11-16\n- Rule")
        self.assertFalse(res_orphan["ok"])
        self.assertTrue(res_orphan["fallback_to_comment"])

        dup = (
            f"{heading}\n2026-11-16\n{heading} · end\n\n"
            f"{heading}\n2026-11-16\n{heading} · end\n"
        )
        res_dup = self.mod.splice_description(dup, heading, "2026-11-16\n- Rule")
        self.assertFalse(res_dup["ok"])
        self.assertTrue(res_dup["fallback_to_comment"])


if __name__ == "__main__":
    unittest.main()
