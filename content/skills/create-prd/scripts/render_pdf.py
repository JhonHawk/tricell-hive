#!/usr/bin/env python3
"""Render a PRD Markdown file to PDF with a browser already installed.

    render_pdf.py <prd.md> [--out FILE.pdf] [--browser PATH]

Converts the Markdown subset the PRD format uses (headings, paragraphs,
lists, tables, block quotes, code, emphasis, links) to a printable A4 page
and prints it with a Chromium-family browser in headless mode. Standard
library only; installs nothing.

The Markdown is untrusted text: it may come from a document someone else
wrote. Every character of it is HTML-escaped, so markup it carries shows as
text; the page emits no image or other remote resource and carries a content
policy that forbids scripts and network loads; the browser runs with a
temporary profile that is removed afterwards.

Exit codes: 0 PDF written (its path on stdout) · 2 usage or unreadable
input · 3 no browser found (the printable page is left beside the Markdown
file) · 4 the browser failed (same page left).
"""

import argparse
import html
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

SAFE_LINK = re.compile(r"(?i)^(https?://|mailto:)")
FENCE = re.compile(r"^ {0,3}(```|~~~)")
HEADING = re.compile(r"^ {0,3}(#{1,6})\s+(.*?)\s*#*\s*$")
RULE = re.compile(r"^ {0,3}([-*_])(\s*\1){2,}\s*$")
QUOTE = re.compile(r"^ {0,3}>\s?")
ITEM = re.compile(r"^(\s*)([-*+]|\d{1,9}[.)])\s+(.*)$")
TABLE_SEPARATOR = re.compile(r"^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$")

STYLE = """
@page { size: A4; margin: 18mm 16mm; }
html { font: 10.5pt/1.5 -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; color: #111; }
body { margin: 0 auto; max-width: 178mm; }
h1 { font-size: 20pt; margin: 0 0 6mm; }
h2 { font-size: 14pt; margin: 8mm 0 3mm; border-bottom: 0.3mm solid #999; padding-bottom: 1mm; }
h3 { font-size: 12pt; margin: 6mm 0 2mm; }
h4, h5, h6 { font-size: 10.5pt; margin: 4mm 0 2mm; }
h1, h2, h3, h4, h5, h6 { break-after: avoid; }
p, ul, ol, table, pre, blockquote { margin: 0 0 3mm; }
ul, ol { padding-left: 6mm; }
table { border-collapse: collapse; width: 100%; }
th, td { border: 0.2mm solid #888; padding: 1.2mm 1.8mm; text-align: left; vertical-align: top; }
th { background: #eee; }
tr { break-inside: avoid; }
code, pre { font: 9.5pt/1.4 ui-monospace, "SF Mono", Menlo, Consolas, monospace; }
pre { background: #f3f3f3; padding: 2mm; white-space: pre-wrap; }
blockquote { border-left: 1mm solid #bbb; margin-left: 0; padding-left: 4mm; color: #333; }
hr { border: 0; border-top: 0.3mm solid #999; margin: 5mm 0; }
a { color: inherit; }
@media print { a[href^="http"]::after { content: " (" attr(href) ")"; font-size: 0.85em; } }
"""


def inline(text):
    """Escape one run of text and apply emphasis, code spans and safe links."""
    tokens = []

    def keep(markup):
        tokens.append(markup)
        return "\x00%d\x00" % (len(tokens) - 1)

    def link(match):
        label, target = match.group(1), match.group(2)
        if not SAFE_LINK.match(target):
            return match.group(0)
        return keep('<a href="%s">' % target) + label + keep("</a>")

    text = html.escape(text.replace("\x00", ""), quote=True)
    text = re.sub(r"`([^`]+)`", lambda m: keep("<code>%s</code>" % m.group(1)), text)
    text = re.sub(r"(?<!!)\[([^\]]+)\]\(([^)\s]+)\)", link, text)
    text = re.sub(r"\*\*(?!\s)(.+?)(?<!\s)\*\*", r"<strong>\1</strong>", text)
    text = re.sub(r"(?<!\w)__(?!\s)(.+?)(?<!\s)__(?!\w)", r"<strong>\1</strong>", text)
    text = re.sub(r"(?<![*\w])\*(?![\s*])(.+?)(?<![\s*])\*(?![*\w])", r"<em>\1</em>", text)
    text = re.sub(r"(?<!\w)_(?![\s_])(.+?)(?<![\s_])_(?!\w)", r"<em>\1</em>", text)
    return re.sub(r"\x00(\d+)\x00", lambda m: tokens[int(m.group(1))], text)


def cells(row):
    row = row.strip()
    if row.startswith("|"):
        row = row[1:]
    if row.endswith("|") and not row.endswith("\\|"):
        row = row[:-1]
    return [cell.replace("\\|", "|").strip() for cell in re.split(r"(?<!\\)\|", row)]


def starts_block(lines, i):
    line = lines[i]
    return bool(FENCE.match(line) or HEADING.match(line) or RULE.match(line) or QUOTE.match(line)
                or ITEM.match(line) or starts_table(lines, i))


def starts_table(lines, i):
    return "|" in lines[i] and i + 1 < len(lines) and bool(TABLE_SEPARATOR.match(lines[i + 1])) \
        and "-" in lines[i + 1]


def render_list(block):
    """Render consecutive list lines; deeper-indented lines nest under their item."""
    out, i = [], 0
    while i < len(block):
        base = ITEM.match(block[i])
        indent, ordered = len(base.group(1)), base.group(2)[0].isdigit()
        tag = "ol" if ordered else "ul"
        out.append("<%s>" % tag)
        while i < len(block):
            match = ITEM.match(block[i])
            if not match or len(match.group(1)) > indent + 1 or match.group(2)[0].isdigit() != ordered:
                break
            text, i = [match.group(3)], i + 1
            children = []
            while i < len(block):
                child = ITEM.match(block[i])
                if child and len(child.group(1)) <= indent + 1:
                    break
                if child or children:
                    children.append(block[i])
                else:
                    text.append(block[i].strip())
                i += 1
            nested = ""
            if children:
                nested = "\n" + render_list(children) + "\n"
            out.append("<li>%s%s</li>" % (inline(" ".join(text)), nested))
        out.append("</%s>" % tag)
        if i < len(block) and not ITEM.match(block[i]):
            out.append("<p>%s</p>" % inline(block[i].strip()))
            i += 1
    return "\n".join(out)


def convert(markdown):
    """Convert the Markdown subset to HTML; anything outside it stays as literal text."""
    lines = markdown.replace("\r\n", "\n").replace("\r", "\n").expandtabs(4).split("\n")
    out, i = [], 0
    while i < len(lines):
        line = lines[i]
        if not line.strip():
            i += 1
        elif FENCE.match(line):
            fence, body, i = FENCE.match(line).group(1), [], i + 1
            while i < len(lines) and not lines[i].lstrip().startswith(fence):
                body.append(lines[i])
                i += 1
            i += 1
            out.append("<pre><code>%s\n</code></pre>" % html.escape("\n".join(body), quote=True))
        elif HEADING.match(line):
            level, text = HEADING.match(line).groups()
            out.append("<h%d>%s</h%d>" % (len(level), inline(text), len(level)))
            i += 1
        elif RULE.match(line):
            out.append("<hr>")
            i += 1
        elif QUOTE.match(line):
            body = []
            while i < len(lines) and QUOTE.match(lines[i]):
                body.append(QUOTE.sub("", lines[i], count=1))
                i += 1
            out.append("<blockquote>\n%s\n</blockquote>" % convert("\n".join(body)))
        elif starts_table(lines, i):
            head, i = cells(line), i + 2
            rows = []
            while i < len(lines) and lines[i].strip() and "|" in lines[i]:
                rows.append(cells(lines[i]))
                i += 1
            table = ["<table>", "<thead><tr>%s</tr></thead>" % "".join("<th>%s</th>" % inline(c) for c in head),
                     "<tbody>"]
            for row in rows:
                table.append("<tr>%s</tr>" % "".join("<td>%s</td>" % inline(c) for c in row))
            out.append("\n".join(table + ["</tbody>", "</table>"]))
        elif ITEM.match(line):
            block = []
            while i < len(lines):
                if lines[i].strip():
                    if not (ITEM.match(lines[i]) or lines[i].startswith("  ")):
                        break
                    if block and ITEM.match(lines[i]) and not lines[i].startswith(" ") \
                            and ITEM.match(lines[i]).group(2)[0].isdigit() != ITEM.match(block[0]).group(2)[0].isdigit():
                        break
                    block.append(lines[i])
                    i += 1
                else:
                    nxt = i + 1
                    if nxt < len(lines) and lines[nxt].strip() and (
                            lines[nxt].startswith("  ") or (ITEM.match(lines[nxt]) and
                            ITEM.match(lines[nxt]).group(2)[0].isdigit() == ITEM.match(block[0]).group(2)[0].isdigit())):
                        i += 1
                    else:
                        break
            out.append(render_list(block))
        else:
            body = [line.strip()]
            i += 1
            while i < len(lines) and lines[i].strip() and not starts_block(lines, i):
                body.append(lines[i].strip())
                i += 1
            out.append("<p>%s</p>" % inline(" ".join(body)))
    return "\n".join(out)


def build_page(markdown, fallback_title):
    """Return the whole printable page for a Markdown document."""
    body = convert(markdown)
    first = re.search(r"<h1>(.*?)</h1>", body)
    title = re.sub(r"<[^>]+>", "", first.group(1)) if first else html.escape(fallback_title, quote=True)
    return ("<!doctype html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n"
            "<meta http-equiv=\"Content-Security-Policy\" content=\"default-src 'none'; style-src 'unsafe-inline'\">\n"
            "<title>%s</title>\n<style>%s</style>\n</head>\n<body>\n%s\n</body>\n</html>\n" % (title, STYLE, body))


def find_browser(override=None):
    """Return the path of a Chromium-family browser, or None. `override` skips the search."""
    if override:
        path = Path(override)
        return str(path) if path.is_file() and os.access(path, os.X_OK) else None
    for name in ("google-chrome", "google-chrome-stable", "chromium", "chromium-browser",
                 "microsoft-edge", "microsoft-edge-stable", "brave-browser", "chrome", "msedge"):
        found = shutil.which(name)
        if found:
            return found
    candidates = []
    for root in (Path("/Applications"), Path.home() / "Applications"):
        for app, binary in (("Google Chrome", "Google Chrome"), ("Chromium", "Chromium"),
                            ("Microsoft Edge", "Microsoft Edge"), ("Brave Browser", "Brave Browser")):
            candidates.append(root / (app + ".app") / "Contents" / "MacOS" / binary)
    for variable in ("PROGRAMFILES", "PROGRAMFILES(X86)", "LOCALAPPDATA"):
        base = os.environ.get(variable)
        if base:
            for parts in (("Google", "Chrome", "Application", "chrome.exe"),
                          ("Microsoft", "Edge", "Application", "msedge.exe"),
                          ("BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
                          ("Chromium", "Application", "chrome.exe")):
                candidates.append(Path(base).joinpath(*parts))
    for path in candidates:
        if path.is_file() and os.access(path, os.X_OK):
            return str(path)
    return None


def print_pdf(browser, page, stem, pdf):
    """Print the page to `pdf` with a temporary browser profile. Return an error text or None."""
    work = Path(tempfile.mkdtemp(prefix="render-pdf-"))
    try:
        source = work / (stem + ".html")
        source.write_text(page, encoding="utf-8")
        profile = work / "profile"
        profile.mkdir()
        command = [browser, "--headless", "--disable-gpu", "--no-first-run", "--no-default-browser-check",
                   "--disable-extensions", "--disable-background-networking", "--no-pdf-header-footer",
                   "--user-data-dir=%s" % profile, "--print-to-pdf=%s" % pdf, source.as_uri()]
        try:
            result = subprocess.run(command, capture_output=True, text=True, timeout=120)
        except (OSError, subprocess.TimeoutExpired) as error:
            return str(error)
        if result.returncode != 0:
            return "the browser exited with code %d" % result.returncode
        if not pdf.is_file() or pdf.stat().st_size == 0:
            return "the browser wrote no PDF"
        return None
    finally:
        shutil.rmtree(work, ignore_errors=True)


def main(argv=None):
    parser = argparse.ArgumentParser(description="Render a PRD Markdown file to PDF with an installed browser.")
    parser.add_argument("markdown", help="the PRD Markdown file")
    parser.add_argument("--out", help="the PDF to write (default: beside the Markdown file)")
    parser.add_argument("--browser", help="browser executable to use; skips the search")
    args = parser.parse_args(argv)

    source = Path(args.markdown)
    try:
        markdown = source.read_text(encoding="utf-8", errors="replace")
    except OSError as error:
        print("render_pdf: cannot read %s: %s" % (source, error.strerror or error), file=sys.stderr)
        return 2
    pdf = Path(args.out) if args.out else source.with_suffix(".pdf")
    page = build_page(markdown, source.stem)
    fallback = source.with_suffix(".html")
    how = ("The printable page is at %s: open it in any browser and print it to PDF "
           "(Cmd+P or Ctrl+P, then Save as PDF)." % fallback)

    browser = find_browser(args.browser)
    if browser is None:
        fallback.write_text(page, encoding="utf-8")
        print("render_pdf: no Chromium-family browser found (Chrome, Edge, Chromium or Brave). " + how,
              file=sys.stderr)
        return 3
    pdf.parent.mkdir(parents=True, exist_ok=True)
    error = print_pdf(browser, page, source.stem, pdf.resolve())
    if error:
        fallback.write_text(page, encoding="utf-8")
        print("render_pdf: %s. %s" % (error, how), file=sys.stderr)
        return 4
    print(pdf)
    return 0


if __name__ == "__main__":
    sys.exit(main())
