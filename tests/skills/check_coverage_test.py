import contextlib
import importlib.util
import io
import json
import pathlib
import tempfile
import textwrap
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/create-issue/scripts/check_coverage.py"
PRD_SCRIPT = ROOT / "content/skills/create-prd/scripts/check_prd.py"


def load_module(script=SCRIPT):
    spec = importlib.util.spec_from_file_location(script.stem, script)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {script}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


PRD = textwrap.dedent("""\
    # Refunds — v2

    ## Requirements

    ### R1 · P0 · decided

    A refund over $500 needs a supervisor's approval.

    ### R2 · P1 (proposed) · to decide (Ana answers)

    The requester is notified when the refund is approved.

    ### R3 · P0 · cancelada (v2; la reemplaza R4)

    Refunds are paid within one day.

    ### R4 · P0 · decided

    Refunds are paid within three business days.
    """)


class CheckCoverageTest(unittest.TestCase):
    def setUp(self):
        self.module = load_module()
        self.dir = tempfile.TemporaryDirectory()
        self.addCleanup(self.dir.cleanup)

    def write(self, name, text):
        path = pathlib.Path(self.dir.name) / name
        path.write_text(text, encoding="utf-8")
        return str(path)

    def main_quiet(self, argv):
        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = self.module.main(argv)
        return code, out.getvalue()

    def run_main(self, prd_text, drafts, *options):
        prd = self.write("prd.md", prd_text)
        paths = [self.write(name, text) for name, text in drafts.items()]
        code, out = self.main_quiet(["--prd", prd, *options, *paths])
        return code, out, paths

    def test_reports_uncovered_shared_and_cancelled_codes(self):
        code, out, paths = self.run_main(
            PRD,
            {
                "approval.md": "Covers: R1",
                "notice.md": "- **Covers:** R1–R2",
                "payout.md": "Covers: R3",
            },
        )
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], ["R4"])
        self.assertEqual(result["shared"], {"R1": [paths[0], paths[1]]})
        self.assertEqual(result["invalid"], [{"code": "R3", "draft": paths[2], "reason": "cancelled"}])

    def test_range_skips_the_cancelled_code_it_crosses(self):
        for wording in ("Cubre: R1 a R4", "Cubre: R1 hasta R4", "Cubre: R1 al R4", "Covers: R1 through R4", "Covers: R1 to R4"):
            with self.subTest(wording=wording):
                code, out, _ = self.run_main(PRD, {"todo.md": wording})
                result = json.loads(out)
                self.assertEqual(code, 0)
                self.assertEqual((result["ok"], result["uncovered"], result["invalid"]), (True, [], []))

    def test_accepted_shapes_of_the_covers_line(self):
        drafts = (
            "Covers:\n- R1\n- R2\n- R4\n\nOut of scope: nothing.",
            "Covers:\n\n- R1\n- R2\n- R4",
            "> Cubre: R1, R2,\n> R4\n",
            "1. **Covers:** r1, r2 and r4",
            "Covers: `R1`, `R2`, `R4`.",
            "Covers: **R1**, R2, and R4",
            "Covers:\n- R1\n\n- R2\n\n- R4\n",
            "Covers:\n1. R1\n2. R2\n3. R4\n",
            "Covers:\n- [x] R1\n- [ ] R2\n- [x] R4\n",
            "- [x] Covers: R1, R2, R4",
        )
        for draft in drafts:
            with self.subTest(draft=draft):
                code, out, _ = self.run_main(PRD, {"todo.md": draft})
                result = json.loads(out)
                self.assertEqual(code, 0)
                self.assertEqual((result["uncovered"], result["malformed"]), ([], []))

    def test_only_the_covers_line_cites(self):
        draft = "Covers: R1\n\nOut of scope: R4, in a later ticket.\nFiles go to Cloudflare R2; see r4.\n"
        code, out, _ = self.run_main(PRD, {"todo.md": draft})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], ["R2", "R4"])
        self.assertEqual(result["invalid"], [])

    def test_text_beside_the_codes_is_malformed_and_its_codes_do_not_count(self):
        cases = {
            "Covers: R1, R2 (R4 goes to the billing child)": ["R4"],
            "Covers: R1-4": ["R2", "R4"],
            "Covers: R4-1": ["R1", "R2"],
            "Covers: R1–R2–R4": ["R4"],
            "Covers: R4, on 2026-11-26": ["R1", "R2"],
            "Covers: R4–R1": ["R1", "R2", "R4"],
            "Covers: [R1](https://example.test), R2, R4": ["R1", "R2", "R4"],
        }
        for draft, uncovered in cases.items():
            with self.subTest(draft=draft):
                code, out, paths = self.run_main(PRD, {"todo.md": draft})
                result = json.loads(out)
                self.assertEqual(code, 1)
                self.assertEqual(result["uncovered"], uncovered)
                self.assertEqual(result["malformed"], [{"draft": paths[0], "line": draft, "reason": "text it cannot read"}])
                self.assertEqual(result["without_covers_line"], [])

    def test_line_after_a_trailing_separator_must_hold_only_codes(self):
        code, out, paths = self.run_main(PRD, {"todo.md": "Covers: R1, R2,\nR4 is the main one.\n"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], ["R4"])
        self.assertEqual(
            result["malformed"],
            [{"draft": paths[0], "line": "Covers: R1, R2,", "reason": "separator with no code after it"}],
        )

    def test_fenced_example_in_a_draft_does_not_cite(self):
        drafts = (
            "Covers: R1, R2\n\n> ```\n> Covers: R4\n> ```\n",
            "Covers: R1, R2\n\n```md\n    ```\n    Covers: R4\n    ```\n```\n",
            "Covers: R1, R2\n```\nx\n> ```\nCovers: R4\n```\n",
        )
        for draft in drafts:
            with self.subTest(draft=draft):
                code, out, _ = self.run_main(PRD, {"todo.md": draft})
                result = json.loads(out)
                self.assertEqual((code, result["uncovered"], result["malformed"]), (1, ["R4"], []))

    def test_another_list_after_the_covers_list_does_not_cite(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers:\n- R1\n- R2\n\n1. R4\n2. R9\n"})
        result = json.loads(out)
        self.assertEqual((code, result["uncovered"], result["invalid"]), (1, ["R4"], []))

    def test_unclosed_fence_in_a_draft_is_malformed(self):
        code, out, paths = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4\n\n```sh\nmake test\n"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["malformed"], [{"draft": paths[0], "line": "```sh", "reason": "code fence never closed"}])

    def test_annotated_list_item_is_malformed_and_still_checked(self):
        code, out, paths = self.run_main(PRD, {"todo.md": "Covers:\n- R1\n- R2\n- R4\n- R3 (old)\n"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["invalid"], [{"code": "R3", "draft": paths[0], "reason": "cancelled"}])
        self.assertEqual(result["malformed"], [{"draft": paths[0], "line": "- R3 (old)", "reason": "text it cannot read"}])

    def test_covers_label_without_any_code_is_not_a_citation(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4\n\n- Covers: both locales\n"})
        result = json.loads(out)
        self.assertEqual(code, 0)
        self.assertEqual(result["malformed"], [])

    def test_draft_without_a_covers_line_fails_and_is_named(self):
        code, out, paths = self.run_main(PRD, {"a.md": "Covers: R1, R2, R4", "b.md": "## Covers\n\n- R1\n"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["without_covers_line"], [paths[1]])

    def test_draft_that_covers_no_rule_says_none(self):
        code, out, _ = self.run_main(PRD, {"a.md": "Covers: R1, R2, R4", "b.md": "Cubre: ninguna"})
        result = json.loads(out)
        self.assertEqual(code, 0)
        self.assertEqual(result["without_covers_line"], [])

    def test_out_of_scope_rule_is_not_uncovered(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2"}, "--out-of-scope", "R4")
        result = json.loads(out)
        self.assertEqual(code, 0)
        self.assertEqual((result["uncovered"], result["out_of_scope"]), ([], ["R4"]))

    def test_out_of_scope_code_the_document_cancelled_is_invalid(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4"}, "--out-of-scope", "R3")
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["invalid"], [{"code": "R3", "draft": "--out-of-scope", "reason": "cancelled"}])

    def test_code_cited_and_out_of_scope_is_invalid(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4"}, "--out-of-scope", "R4")
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["invalid"], [{"code": "R4", "draft": "--out-of-scope", "reason": "also cited by a draft"}])

    def test_out_of_scope_value_with_other_text_is_a_usage_error(self):
        for value in ("R4 (later), R9", "R2/R4", "[R2, R4]"):
            with self.subTest(value=value):
                code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1"}, "--out-of-scope", value)
                self.assertEqual((code, out), (2, ""))

    def test_code_the_document_lacks_is_invalid(self):
        code, out, paths = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4 and R9"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], [])
        self.assertEqual(result["invalid"], [{"code": "R9", "draft": paths[0], "reason": "not in the document"}])

    def test_fenced_examples_are_not_rules_and_do_not_hide_later_rules(self):
        nested = PRD + "\n````text\n```\n### R99 · P0 · decided\n```\n````\n"
        code, out, _ = self.run_main(nested, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual((code, json.loads(out)["uncovered"]), (0, []))
        mixed = PRD + "\n~~~\n```\n### R98 · P0\n~~~\n\n### R5 · P0 · decided\n\nA later rule.\n"
        code, out, _ = self.run_main(mixed, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual((code, json.loads(out)["uncovered"]), (1, ["R5"]))

    def test_headings_that_are_not_rule_headings_are_ignored(self):
        prd = "# R1 refunds requirements\n\n" + PRD.replace("A refund over", "#### R4 examples\n\nA refund over")
        prd += "\n###\nR9 · P1 · decided\n\nInline ```py``` is not a fence.\n```py``` at a line start is not one either.\n"
        prd += "\n### R2D2 integration\n\n### r5a instance sizing\n\n    ```\n    an indented block that shows a marker\n"
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual((code, json.loads(out)["uncovered"]), (0, []))

    def test_lower_case_or_no_break_space_rule_heading_is_a_rule(self):
        for heading in ("### r4 · p0 · Decided", "### R4\u00a0· P0 · decided"):
            with self.subTest(heading=heading):
                prd = PRD.replace("### R4 · P0 · decided", heading)
                code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2"})
                self.assertEqual((code, json.loads(out)["uncovered"]), (1, ["R4"]))

    def test_unclosed_fence_in_the_document_is_a_usage_error(self):
        prd = PRD.replace("### R4 · P0 · decided", "```text\nexample\n\n### R4 · P0 · decided")
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2"})
        self.assertEqual((code, out), (2, ""))

    def test_rule_heading_the_format_rejects_is_a_usage_error(self):
        replacements = (
            ("### R1 · P0 · decided", "### R1 · P0"),
            ("### R4 · P0 · decided", "#### R4 · P0 · decided"),
            ("### R4 · P0 · decided", " ### R4 · P0 · decided"),
            ("### R4 · P0 · decided", "### **R4** · P0 · decided"),
            ("### R4 · P0 · decided", "### R4a · P0 · decided"),
            ("### R4 · P0 · decided", "> ### R4 · P0 · decided"),
            ("### R2 · P1 (proposed) · to decide (Ana answers)", "### R2 - P1 - to decide"),
            ("cancelada (v2; la reemplaza R4)", "**cancelled** (v2)"),
            ("· decided\n\nA refund", "· canceled in v2\n\nA refund"),
            ("Refunds are paid within three business days.", "Refunds.\n\n### R1 · P0 · cancelled (v3)\n\nRecap."),
        )
        for old, new in replacements:
            with self.subTest(new=new):
                self.assertIn(old, PRD)
                code, out, _ = self.run_main(PRD.replace(old, new), {"todo.md": "Covers: R1, R2, R4"})
                self.assertEqual((code, out), (2, ""))

    def test_states_match_the_requirement_document_check(self):
        self.assertEqual(self.module.VALID_STATE_PREFIXES, load_module(PRD_SCRIPT).VALID_STATE_PREFIXES)

    def test_document_without_rule_headings_is_a_usage_error(self):
        code, out, _ = self.run_main("# Notes\n\nNo rules here.\n", {"todo.md": "Covers: R1"})
        self.assertEqual((code, out), (2, ""))

    def test_document_passed_as_a_draft_is_a_usage_error(self):
        prd = self.write("prd.md", PRD)
        self.assertEqual(self.main_quiet(["--prd", prd, prd]), (2, ""))

    def test_draft_that_is_not_utf8_is_a_usage_error(self):
        prd = self.write("prd.md", PRD)
        draft = pathlib.Path(self.dir.name) / "todo.md"
        draft.write_bytes("Covers: R1".encode("utf-16"))
        self.assertEqual(self.main_quiet(["--prd", prd, str(draft)]), (2, ""))

    def test_byte_order_mark_does_not_hide_the_first_line(self):
        prd = self.write("prd.md", "\ufeff" + PRD.split("## Requirements\n\n", 1)[1])
        draft = self.write("todo.md", "\ufeffCovers: R1, R2, R4")
        code, out = self.main_quiet(["--prd", prd, draft])
        self.assertEqual((code, json.loads(out)["uncovered"]), (0, []))


if __name__ == "__main__":
    unittest.main()
