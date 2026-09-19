"""Tests for convert-skills.py — the description budget the strictest harness enforces."""
import importlib.util
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
CONVERTER_PATH = ROOT / "harness" / "build" / "convert-skills.py"
spec = importlib.util.spec_from_file_location("convert_skills", CONVERTER_PATH)
CONVERTER = importlib.util.module_from_spec(spec)
spec.loader.exec_module(CONVERTER)


def skill(description_block):
    return f"---\nname: demo\ndescription: >\n{description_block}\n---\n\nBody.\n"


class DescriptionBudgetTests(unittest.TestCase):
    def test_a_folded_description_is_measured_as_the_harness_reads_it(self):
        text = skill("  one two\n  three   four")
        self.assertEqual(CONVERTER.description_length(text), len("one two three four"))
        self.assertEqual(
            CONVERTER.description_length("---\nname: x\ndescription: plain one\n---\n"),
            len("plain one"),
        )

    def test_a_description_over_the_budget_fails_the_build_naming_the_skill(self):
        # PI refuses to auto-load a skill whose description exceeds 1024 characters
        # and says so only in its startup banner — a router nobody loads is a rule
        # nobody has.
        with tempfile.TemporaryDirectory(prefix="hive-skills-") as tmp:
            src = Path(tmp) / "skills" / "too-long"
            src.mkdir(parents=True)
            (src / "SKILL.md").write_text(skill("  " + "x" * 1025), encoding="utf-8")
            result = subprocess.run(
                [sys.executable, str(CONVERTER_PATH), str(src.parent), str(Path(tmp) / "out")],
                capture_output=True, text=True,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("too-long", result.stderr + result.stdout)
            self.assertIn("1024", result.stderr + result.stdout)

    def test_every_real_skill_fits_the_budget(self):
        over = {
            path.parent.name: CONVERTER.description_length(path.read_text(encoding="utf-8"))
            for path in sorted((ROOT / "global" / "skills").glob("*/SKILL.md"))
        }
        over = {name: size for name, size in over.items()
                if size > CONVERTER.MAX_DESCRIPTION_CHARS}
        self.assertEqual(over, {})


if __name__ == "__main__":
    unittest.main()
