from pathlib import Path
import html
import importlib.util
import json
import os
import re
import shutil
import stat
import subprocess
import sys
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/create-prd/scripts/render_pdf.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

PRD = textwrap.dedent("""\
    # PRD: Returns

    Version 2 · iterating

    ## Problem

    Customers **cannot** return an order from the *app*; they call `support`.
    See [the policy](https://example.com/policy).

    ## Requirements

    | Code | Priority | Rule | State | Source | Verdict |
    | --- | --- | --- | --- | --- | --- |
    | R1 | P0 | A customer returns an order within 30 days | decided | Ana, 2026-11-10 | not checked |
    | R2 | P1 | A refund over $500 needs a supervisor | to decide | Ana, 2026-11-12 | not checked |

    ## User stories

    - As a customer I return an order
      - from the order list
    - As a supervisor I approve a refund

    1. Open the order
    2. Tap return

    > Quoted note from the author.

    ```
    status = "returned"
    ```

    ---

    Closing paragraph with a footnote[^1] and an image ![logo](http://tracker.example/logo.png).
""")

FAKE_BROWSER = textwrap.dedent("""\
    #!/usr/bin/env python3
    import json, os, sys
    out = [a.split("=", 1)[1] for a in sys.argv if a.startswith("--print-to-pdf=")][0]
    profile = [a.split("=", 1)[1] for a in sys.argv if a.startswith("--user-data-dir=")]
    with open(os.environ["FAKE_BROWSER_LOG"], "w") as f:
        json.dump({"argv": sys.argv[1:], "profile_existed": bool(profile) and os.path.isdir(profile[0])}, f)
    if not os.environ.get("FAKE_BROWSER_SILENT"):
        with open(out, "wb") as f:
            f.write(b"%PDF-1.4 fake")
""")


def load():
    spec = importlib.util.spec_from_file_location("render_pdf", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def visible_text(page):
    """The text a reader sees, plus each link's destination, which the print style shows."""
    body = re.sub(r"(?s)<(style|head)\b.*?</\1>", " ", page)
    body = re.sub(r'<a href="([^"]*)">', r" \1 ", body)
    return " ".join(html.unescape(re.sub(r"<[^>]+>", " ", body)).split())


def run(*args, env=None):
    return subprocess.run([sys.executable, "-I", str(SCRIPT), *map(str, args)],
                          capture_output=True, text=True, env=env)


class Conversion(unittest.TestCase):
    def setUp(self):
        self.page = load().build_page(PRD, "fallback title")

    def test_each_construct_of_the_prd_format_becomes_its_element(self):
        page = self.page
        self.assertIn("<h1>PRD: Returns</h1>", page)
        self.assertIn("<h2>Requirements</h2>", page)
        self.assertIn("<strong>cannot</strong>", page)
        self.assertIn("<em>app</em>", page)
        self.assertIn("<code>support</code>", page)
        self.assertIn('<a href="https://example.com/policy">the policy</a>', page)
        self.assertRegex(page, r"<th[^>]*>Priority</th>")
        self.assertIn("<td>A refund over $500 needs a supervisor</td>", page)
        self.assertRegex(page, r"(?s)<ul>\s*<li>As a customer I return an order\s*<ul>\s*<li>from the order list</li>")
        self.assertRegex(page, r"(?s)<ol>\s*<li>Open the order</li>\s*<li>Tap return</li>\s*</ol>")
        self.assertRegex(page, r"(?s)<blockquote>\s*<p>Quoted note from the author\.</p>")
        self.assertIn("<pre><code>status = &quot;returned&quot;\n</code></pre>", page)
        self.assertIn("<hr>", page)
        self.assertIn("<title>PRD: Returns</title>", page)

    def test_every_word_of_the_markdown_is_in_the_page(self):
        text = visible_text(self.page)
        for word in set(re.findall(r"[A-Za-z0-9$]{2,}", PRD)):
            self.assertIn(word, text)

    def test_a_construct_outside_the_subset_stays_as_literal_text(self):
        text = visible_text(self.page)
        self.assertIn("footnote[^1]", text)
        self.assertIn("![logo](http://tracker.example/logo.png)", text)


class EdgesOfTheSubset(unittest.TestCase):
    def body(self, markdown):
        return load().convert(textwrap.dedent(markdown))

    def test_a_heading_keeps_a_trailing_hash_that_belongs_to_its_text(self):
        self.assertIn("<h2>Support for C#</h2>", self.body("## Support for C#\n"))
        self.assertIn("<h2>Title</h2>", self.body("## Title ##\n"))

    def test_a_paragraph_line_that_starts_with_a_number_stays_in_the_paragraph(self):
        self.assertIn("<p>Released in 2026. It was good.</p>", self.body("Released in\n2026. It was good.\n"))

    def test_an_ordered_list_keeps_the_number_it_starts_at(self):
        body = self.body("""\
            Before.

            3. third
            4. fourth
        """)
        self.assertRegex(body, r'(?s)<ol start="3">\s*<li>third</li>\s*<li>fourth</li>')

    def test_a_line_with_a_pipe_above_a_rule_is_not_a_table(self):
        body = self.body("a | b\n---\nnext\n")
        self.assertNotIn("<table>", body)
        self.assertIn("a | b", body)
        self.assertIn("<hr>", body)

    def test_a_longer_fence_holds_a_shorter_one(self):
        body = self.body("````\n```\ninner\n```\n````\n")
        self.assertEqual(body.count("<pre>"), 1)
        self.assertIn("```\ninner\n```\n</code></pre>", body)

    def test_a_fenced_block_under_a_list_item_keeps_its_lines(self):
        body = self.body("""\
            - step

              ```
              a | b
              # not a heading
              ```
            - next
        """)
        self.assertIn("<pre><code>a | b\n# not a heading\n</code></pre>", body)
        self.assertEqual(body.count("<ul>"), 1)
        self.assertNotIn("<h1>", body)

    def test_an_unindented_continuation_line_stays_in_its_item(self):
        body = self.body("- item one that wraps\nonto the next line\n- item two\n")
        self.assertEqual(body.count("<ul>"), 1)
        self.assertIn("<li>item one that wraps onto the next line</li>", body)

    def test_a_table_with_a_short_separator_row_is_a_table(self):
        body = self.body("| a | b |\n|--|:-:|\n| 1 | 2 |\n")
        self.assertRegex(body, r"<th[^>]*>a</th>")
        self.assertRegex(body, r"<td[^>]*>2</td>")

    def test_code_in_three_backticks_inside_a_line_does_not_open_a_fence(self):
        body = self.body("```inline``` is code\n\nNext paragraph.\n\n## Heading\n")
        self.assertNotIn("<pre>", body)
        self.assertIn("is code", body)
        self.assertIn("<p>Next paragraph.</p>", body)
        self.assertIn("<h2>Heading</h2>", body)

    def test_a_wrapped_item_line_that_starts_with_a_number_stays_in_the_item(self):
        body = self.body("- released in\n  2024. Then more\n- next\n")
        self.assertIn("<li>released in 2024. Then more</li>", body)
        self.assertNotIn("<ol", body)

    def test_a_table_under_a_list_item_is_a_table(self):
        body = self.body("""\
            - item

              | a | b |
              | --- | --- |
              | 1 | 2 |
        """)
        self.assertRegex(body, r"(?s)<li>item\s*<table>")

    def test_a_link_target_keeps_its_parentheses_and_never_holds_a_code_span(self):
        body = self.body("[a](http://a.b/c_(d)) and [b](https://x/`y`)\n")
        self.assertIn('<a href="http://a.b/c_(d)">a</a>', body)
        self.assertNotIn("\x00", body)
        self.assertIn("<code>y</code>", body)


class ShortCells(unittest.TestCase):
    def test_short_columns_stay_on_one_line_and_the_long_one_wraps(self):
        page = load().build_page(PRD, "t")
        for short in ("R1", "P0", "to decide", "Ana, 2026-11-10", "not checked"):
            self.assertIn('<td class="short">%s</td>' % short, page)
        self.assertIn("<td>A customer returns an order within 30 days</td>", page)
        self.assertRegex(page, r"\.short\s*\{[^}]*white-space:\s*nowrap")

    def test_a_table_of_many_short_columns_is_left_free_to_wrap(self):
        row = "| " + " | ".join(["eighteen chars txt"] * 8) + " |"
        table = "\n".join([row, "|" + " --- |" * 8, row])
        self.assertNotIn('class="short"', load().convert(table))


class UntrustedText(unittest.TestCase):
    def test_html_in_the_markdown_is_shown_as_text_and_nothing_remote_is_loaded(self):
        hostile = textwrap.dedent("""\
            # Title <script>alert(1)</script>

            <img src="http://tracker.example/pixel.png" onerror="alert(2)">

            | Code | Rule |
            | --- | --- |
            | R1 | <iframe src="http://evil.example"></iframe> |

            [click](javascript:alert(3)) and [file](file:///etc/passwd)
        """)
        page = load().build_page(hostile, "t")
        for tag in ("<script", "<img", "<iframe", "<link", "<object", "<embed"):
            self.assertNotIn(tag, page.lower())
        self.assertNotIn('href="javascript:', page)
        self.assertNotIn('href="file:', page)
        self.assertIn("&lt;script&gt;alert(1)&lt;/script&gt;", page)
        self.assertIn("&lt;iframe src=", page)
        policy = re.search(r'<meta http-equiv="Content-Security-Policy" content="([^"]*)">', page)
        self.assertIsNotNone(policy)
        self.assertIn("default-src 'none'", policy.group(1))
        self.assertNotIn("script-src", policy.group(1))


@unittest.skipIf(os.name == "nt", "the fake browser is a POSIX script")
class BrowserProcess(unittest.TestCase):
    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.tmp, True)
        self.md = self.tmp / "prd.md"
        self.md.write_text(PRD)

    def test_without_a_browser_it_leaves_the_page_and_fails_saying_how_to_print(self):
        result = run(self.md, "--browser", self.tmp / "no-such-browser")
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.tmp / "prd.pdf").exists())
        page = self.tmp / "prd.print.html"
        self.assertTrue(page.exists())
        self.assertIn("<h1>PRD: Returns</h1>", page.read_text())
        self.assertIn(str(page), result.stderr)
        self.assertIn("print", result.stderr.lower())

    def test_it_prints_through_the_browser_with_a_temporary_profile(self):
        fake = self.tmp / "fake-browser"
        fake.write_text(FAKE_BROWSER)
        fake.chmod(fake.stat().st_mode | stat.S_IXUSR)
        log = self.tmp / "log.json"
        result = run(self.md, "--browser", fake, env=dict(os.environ, FAKE_BROWSER_LOG=str(log)))
        self.assertEqual(result.returncode, 0, result.stderr)
        pdf = self.tmp / "prd.pdf"
        self.assertTrue(pdf.exists())
        self.assertIn(str(pdf), result.stdout)
        self.assertFalse((self.tmp / "prd.print.html").exists())
        seen = json.loads(log.read_text())
        self.assertTrue(seen["profile_existed"])
        profile = [a.split("=", 1)[1] for a in seen["argv"] if a.startswith("--user-data-dir=")][0]
        self.assertFalse(Path(profile).exists(), "the temporary profile is removed afterwards")
        self.assertTrue(any(a.startswith("--headless") for a in seen["argv"]))
        self.assertTrue(seen["argv"][-1].startswith("file://"))

    def test_a_browser_that_writes_nothing_fails_even_when_an_older_pdf_exists(self):
        fake = self.tmp / "fake-browser"
        fake.write_text(FAKE_BROWSER)
        fake.chmod(fake.stat().st_mode | stat.S_IXUSR)
        (self.tmp / "prd.pdf").write_bytes(b"%PDF-1.4 version one")
        result = run(self.md, "--browser", fake, env=dict(os.environ, FAKE_BROWSER_SILENT="1",
                                                           FAKE_BROWSER_LOG=str(self.tmp / "log.json")))
        self.assertNotEqual(result.returncode, 0)
        self.assertTrue((self.tmp / "prd.print.html").exists())

    def test_a_byte_order_mark_does_not_hide_the_first_heading(self):
        self.md.write_bytes(b"\xef\xbb\xbf" + PRD.encode("utf-8"))
        run(self.md, "--browser", self.tmp / "no-such-browser")
        self.assertIn("<h1>PRD: Returns</h1>", (self.tmp / "prd.print.html").read_text())

    def test_the_printable_page_never_replaces_a_file_that_was_there(self):
        (self.tmp / "prd.html").write_text("mine")
        run(self.md, "--browser", self.tmp / "no-such-browser")
        self.assertEqual((self.tmp / "prd.html").read_text(), "mine")

    def test_out_may_name_a_folder(self):
        fake = self.tmp / "fake-browser"
        fake.write_text(FAKE_BROWSER)
        fake.chmod(fake.stat().st_mode | stat.S_IXUSR)
        folder = self.tmp / "exports"
        folder.mkdir()
        result = run(self.md, "--out", folder, "--browser", fake,
                     env=dict(os.environ, FAKE_BROWSER_LOG=str(self.tmp / "log.json")))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue((folder / "prd.pdf").is_file())
        self.assertIn(str(folder / "prd.pdf"), result.stdout)

    def test_browser_may_be_a_command_on_the_path(self):
        bin_dir = self.tmp / "bin"
        bin_dir.mkdir()
        fake = bin_dir / "my-browser"
        fake.write_text(FAKE_BROWSER)
        fake.chmod(fake.stat().st_mode | stat.S_IXUSR)
        env = dict(os.environ, PATH=str(bin_dir) + os.pathsep + os.environ["PATH"],
                   FAKE_BROWSER_LOG=str(self.tmp / "log.json"))
        result = run(self.md, "--browser", "my-browser", env=env)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_a_folder_that_cannot_be_written_is_reported_without_a_traceback(self):
        locked = self.tmp / "locked"
        locked.mkdir()
        md = locked / "prd.md"
        md.write_text(PRD)
        locked.chmod(0o500)
        self.addCleanup(locked.chmod, 0o700)
        result = run(md, "--browser", self.tmp / "no-such-browser")
        self.assertIn(result.returncode, (3, 4))
        self.assertNotIn("Traceback", result.stderr)

    def test_out_names_the_pdf(self):
        fake = self.tmp / "fake-browser"
        fake.write_text(FAKE_BROWSER)
        fake.chmod(fake.stat().st_mode | stat.S_IXUSR)
        out = self.tmp / "shared" / "returns-v2.pdf"
        result = run(self.md, "--out", out, "--browser", fake,
                     env=dict(os.environ, FAKE_BROWSER_LOG=str(self.tmp / "log.json")))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(out.exists())

    def test_a_missing_markdown_file_is_a_usage_error(self):
        result = run(self.tmp / "absent.md")
        self.assertEqual(result.returncode, 2)
        self.assertIn("absent.md", result.stderr)


class RealBrowser(unittest.TestCase):
    def test_the_pdf_is_a4_and_carries_every_rule(self):
        browser = load().find_browser()
        if browser is None:
            self.skipTest("no Chromium-family browser on this machine")
        tmp = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, tmp, True)
        md = tmp / "prd.md"
        md.write_text(PRD)
        result = run(md)
        self.assertEqual(result.returncode, 0, result.stderr)
        data = (tmp / "prd.pdf").read_bytes()
        boxes = re.findall(rb"/MediaBox\s*\[\s*0\s+0\s+([\d.]+)\s+([\d.]+)\s*\]", data)
        self.assertTrue(boxes)
        for width, height in boxes:
            self.assertAlmostEqual(float(width), 595.3, delta=1)
            self.assertAlmostEqual(float(height), 841.9, delta=1)
        pdftotext = shutil.which("pdftotext")
        if pdftotext:
            # Reading order, not -layout: a rule that wraps inside its cell must still read whole.
            text = " ".join(subprocess.run([pdftotext, str(tmp / "prd.pdf"), "-"],
                                           capture_output=True, text=True).stdout.split())
            for rule in ("A customer returns an order within 30 days", "A refund over $500 needs a supervisor"):
                self.assertIn(rule, text)
            # With -layout a value broken across two lines reads apart: short cells must not break.
            laid_out = subprocess.run([pdftotext, "-layout", str(tmp / "prd.pdf"), "-"],
                                      capture_output=True, text=True).stdout
            for value in ("Ana, 2026-11-10", "to decide", "not checked"):
                self.assertIn(value, laid_out)


if __name__ == "__main__":
    unittest.main()
