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


def load_module():
    spec = importlib.util.spec_from_file_location("check_coverage", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
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

    def run_main(self, prd_text, drafts, *options):
        prd = self.write("prd.md", prd_text)
        paths = [self.write(name, text) for name, text in drafts.items()]
        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = self.module.main(["--prd", prd, *options, *paths])
        return code, out.getvalue(), paths

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

    def test_spanish_range_skips_the_cancelled_code_it_crosses(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Cubre: R1 a R4"})
        result = json.loads(out)
        self.assertEqual(code, 0)
        self.assertEqual(result["ok"], True)
        self.assertEqual((result["uncovered"], result["shared"], result["invalid"]), ([], {}, []))

    def test_range_words_of_both_languages_expand(self):
        for wording in ("Cubre: R1 hasta R4", "Cubre: R1 al R4", "Covers: R1 through R4", "Covers: R1 to R4"):
            with self.subTest(wording=wording):
                code, out, _ = self.run_main(PRD, {"todo.md": wording})
                self.assertEqual(code, 0)
                self.assertEqual(json.loads(out)["uncovered"], [])

    def test_only_the_covers_line_cites(self):
        draft = "Covers: R1\n\nOut of scope: R4, in a later ticket.\nFiles go to Cloudflare R2; see r4.\n"
        code, out, _ = self.run_main(PRD, {"todo.md": draft})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], ["R2", "R4"])
        self.assertEqual(result["invalid"], [])

    def test_out_of_scope_rule_is_not_uncovered(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2"}, "--out-of-scope", "R4")
        result = json.loads(out)
        self.assertEqual(code, 0)
        self.assertEqual(result["uncovered"], [])
        self.assertEqual(result["out_of_scope"], ["R4"])

    def test_out_of_scope_code_the_document_cancelled_is_invalid(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4"}, "--out-of-scope", "R3")
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["invalid"], [{"code": "R3", "draft": "--out-of-scope", "reason": "cancelled"}])

    def test_formatted_cancelled_state_is_cancelled(self):
        prd = PRD.replace("cancelada (v2; la reemplaza R4)", "**cancelled** (v2; replaced by R4)")
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual(code, 0)
        self.assertEqual(json.loads(out)["uncovered"], [])

    def test_rule_heading_without_its_state_is_a_usage_error(self):
        prd = PRD.replace("### R1 · P0 · decided", "### R1 · P0")
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual(code, 2)
        self.assertEqual(out, "")

    def test_draft_that_is_not_utf8_is_a_usage_error(self):
        prd = self.write("prd.md", PRD)
        draft = pathlib.Path(self.dir.name) / "todo.md"
        draft.write_bytes("Covers: R1".encode("utf-16"))
        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = self.module.main(["--prd", prd, str(draft)])
        self.assertEqual(code, 2)
        self.assertEqual(out.getvalue(), "")

    def test_text_beside_the_codes_makes_the_covers_line_malformed(self):
        code, out, paths = self.run_main(PRD, {"todo.md": "Covers: R1, R2 (R4 goes to the billing child)"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], ["R4"])
        self.assertEqual(len(result["malformed"]), 1)
        self.assertEqual(result["malformed"][0]["draft"], paths[0])

    def test_codes_listed_under_the_covers_line_count(self):
        for draft in ("Covers:\n- R1\n- R2\n- R4\n\nOut of scope: nothing.", "> Cubre: R1, R2,\n> R4\n", "1. **Covers:** r1, r2 and r4"):
            with self.subTest(draft=draft):
                code, out, _ = self.run_main(PRD, {"todo.md": draft})
                result = json.loads(out)
                self.assertEqual(code, 0)
                self.assertEqual(result["uncovered"], [])
                self.assertEqual(result["malformed"], [])

    def test_range_with_a_bare_or_chained_end_expands(self):
        for wording in ("Covers: R1-4", "Covers: R1 through 4", "Covers: R1–R2–R4"):
            with self.subTest(wording=wording):
                code, out, _ = self.run_main(PRD, {"todo.md": wording})
                self.assertEqual(code, 0)
                self.assertEqual(json.loads(out)["uncovered"], [])

    def test_struck_or_us_spelled_cancelled_state_is_cancelled(self):
        for state in ("~~Cancelada~~ (v2)", "canceled in v2"):
            with self.subTest(state=state):
                prd = PRD.replace("cancelada (v2; la reemplaza R4)", state)
                code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2, R4"})
                self.assertEqual(code, 0)

    def test_rule_heading_inside_a_code_fence_is_not_a_rule(self):
        prd = PRD + "\n```text\n### R99 · P0 · decided\n### R98 · P0\n```\n"
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual(code, 0)
        self.assertEqual(json.loads(out)["uncovered"], [])

    def test_rule_heading_at_another_level_is_a_usage_error(self):
        prd = PRD.replace("### R4 · P0 · decided", "#### R4 · P0 · decided")
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2"})
        self.assertEqual(code, 2)
        self.assertEqual(out, "")

    def test_duplicate_rule_code_is_a_usage_error(self):
        prd = PRD + "\n### R1 · P0 · cancelled (v3)\n\nRecap.\n"
        code, out, _ = self.run_main(prd, {"todo.md": "Covers: R1, R2, R4"})
        self.assertEqual(code, 2)
        self.assertEqual(out, "")

    def test_draft_without_a_covers_line_is_named(self):
        code, out, paths = self.run_main(PRD, {"a.md": "Covers: R1, R2, R4", "b.md": "No codes here."})
        result = json.loads(out)
        self.assertEqual(code, 0)
        self.assertEqual(result["without_covers_line"], [paths[1]])

    def test_code_cited_and_out_of_scope_is_invalid(self):
        code, out, _ = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4"}, "--out-of-scope", "R4")
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["invalid"], [{"code": "R4", "draft": "--out-of-scope", "reason": "also cited by a draft"}])

    def test_document_passed_as_a_draft_is_a_usage_error(self):
        prd = self.write("prd.md", PRD)
        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = self.module.main(["--prd", prd, prd])
        self.assertEqual(code, 2)
        self.assertEqual(out.getvalue(), "")

    def test_byte_order_mark_does_not_hide_the_first_line(self):
        prd = self.write("prd.md", "\ufeff" + PRD.split("## Requirements\n\n", 1)[1])
        draft = self.write("todo.md", "\ufeffCovers: R1, R2, R4")
        out, err = io.StringIO(), io.StringIO()
        with contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            code = self.module.main(["--prd", prd, draft])
        self.assertEqual(code, 0)
        self.assertEqual(json.loads(out.getvalue())["uncovered"], [])

    def test_code_the_document_lacks_is_invalid(self):
        code, out, paths = self.run_main(PRD, {"todo.md": "Covers: R1, R2, R4 and R9"})
        result = json.loads(out)
        self.assertEqual(code, 1)
        self.assertEqual(result["uncovered"], [])
        self.assertEqual(result["invalid"], [{"code": "R9", "draft": paths[0], "reason": "not in the document"}])

    def test_document_without_rule_headings_is_a_usage_error(self):
        code, out, _ = self.run_main("# Notes\n\nNo rules here.\n", {"todo.md": "Covers: R1"})
        self.assertEqual(code, 2)
        self.assertEqual(out, "")


if __name__ == "__main__":
    unittest.main()
