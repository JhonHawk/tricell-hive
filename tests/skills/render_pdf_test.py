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

    | Code | Priority | Rule | State |
    | --- | --- | --- | --- |
    | R1 | P0 | A customer returns an order within 30 days | decided |
    | R2 | P1 | A refund over $500 needs a supervisor | to decide |

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
        self.assertIn("<th>Priority</th>", page)
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
        page = self.tmp / "prd.html"
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
        self.assertFalse((self.tmp / "prd.html").exists())
        seen = json.loads(log.read_text())
        self.assertTrue(seen["profile_existed"])
        profile = [a.split("=", 1)[1] for a in seen["argv"] if a.startswith("--user-data-dir=")][0]
        self.assertFalse(Path(profile).exists(), "the temporary profile is removed afterwards")
        self.assertTrue(any(a.startswith("--headless") for a in seen["argv"]))
        self.assertTrue(seen["argv"][-1].startswith("file://"))

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
            text = " ".join(subprocess.run([pdftotext, "-layout", str(tmp / "prd.pdf"), "-"],
                                           capture_output=True, text=True).stdout.split())
            for rule in ("A customer returns an order within 30 days", "A refund over $500 needs a supervisor"):
                self.assertIn(rule, text)


if __name__ == "__main__":
    unittest.main()
