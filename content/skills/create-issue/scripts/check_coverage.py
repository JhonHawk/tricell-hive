#!/usr/bin/env python3
"""Requirement coverage of ticket drafts (English and Spanish).

Usage:
  python3 check_coverage.py --prd <prd-file> [--out-of-scope <codes>] <draft-file>...

A draft cites the rules it covers in a line that starts "Covers:" or "Cubre:":
codes such as R4 and ranges such as R7-R9, separated by commas, on that line or
as a list right under it, or "none". Codes elsewhere in a draft do not count.

The check reads one strict form and reports the rest instead of guessing: text
beside the codes makes the line malformed, a draft without such a line is named,
and a rule heading the document format rejects is a usage error. Exits 1 when a
rule is uncovered, a cited code is invalid, a line is malformed or a draft has no
Covers line; 2 on a usage error.
"""

import argparse
import json
import pathlib
import re
import sys

QUOTE_RE = re.compile(r"^((?:\s{0,3}>)*)(.*)$")
FENCE_RE = re.compile(r"^\s{0,3}(`{3,}|~{3,})(.*)$")
SPACE = r"[^\S\n]"
RULE_LINE_RE = re.compile(
    rf"^(?:{SPACE}{{0,3}}>)*{SPACE}{{0,3}}"
    rf"(?:###{SPACE}+[*_`]*R\d+[*_`]*(?={SPACE}|$)|#{{1,6}}{SPACE}+[*_`]*R\d+\w*[*_`]*{SPACE}*·).*$",
    re.MULTILINE | re.IGNORECASE,
)
RULE_HEADING_RE = re.compile(rf"^###{SPACE}+R(\d+){SPACE}*·{SPACE}*P[012][^·\n]*·{SPACE}*(.+?)\s*$", re.IGNORECASE)
VALID_STATE_PREFIXES = (
    "decided",
    "decidida",
    "decidido",
    "assumption",
    "supuesto",
    "to decide",
    "por decidir",
    "cancelled",
    "cancelada",
    "cancelado",
)
CANCELLED_PREFIXES = ("cancelled", "cancelada", "cancelado")
LINE_PREFIX = r"^(?:\s{0,3}>)*\s{0,3}(?:#{1,6}\s+|[-*]\s+|\d+[.)]\s+)?(?:\[[ xX]\]\s+)?"
COVERS_LINE_RE = re.compile(LINE_PREFIX + r"[*_]*(?:Covers|Cubre)[*_]*\s*:[*_]*(.*)$", re.IGNORECASE)
CONTINUED_LINE_RE = re.compile(LINE_PREFIX + r"(.*)$")
LIST_ITEM_RE = re.compile(r"^\s*(?:>\s*)*([-*]|\d+[.)])\s+(?:\[[ xX]\]\s+)?(.*)$")
CODE = r"[*_`]*[Rr](\d+)[*_`]*"
ITEM_RE = re.compile(rf"\s*{CODE}(?:\s*(?:[-–—]|\b(?i:a|al|hasta|to|through|thru)\b)\s*{CODE})?")
SEPARATOR_RE = re.compile(r"\s*(?:[,;]\s*(?:(?i:and|y|e)\b)?|\b(?i:and|y|e)\b)")
ANY_CODE_RE = re.compile(r"\b[Rr]\d+")
NO_CODES_RE = re.compile(r"(?i:none|ninguna|ninguno)\.?")
OUT_OF_SCOPE = "--out-of-scope"
UNREADABLE = "text it cannot read"
DANGLING = "separator with no code after it"
OPEN_FENCE = "code fence never closed"


def without_fences(text):
    """Drop fenced code blocks, which show formats and are not content.

    A fence opens and closes at up to three spaces inside the same level of
    quote markers, and ends with the quote that holds it. Returns (the remaining
    text, the opening line of a fence left open, or None).
    """
    kept = []
    fence = None
    depth = 0
    opening = None
    for line in text.splitlines():
        quote, rest = QUOTE_RE.match(line).groups()
        level = quote.count(">")
        if fence is not None and level < depth:
            fence = None
        marker = FENCE_RE.match(rest)
        if fence is None:
            if marker and not (marker.group(1)[0] == "`" and "`" in marker.group(2)):
                fence, depth, opening = marker.group(1), level, line.strip()
            else:
                kept.append(line)
        elif marker and level == depth and marker.group(1).startswith(fence) and not marker.group(2).strip():
            fence = None
    return "\n".join(kept), opening if fence else None


def parse_rules(prd_text):
    """Return ({number: cancelled}, [what the document format rejects])."""
    rules = {}
    text, open_fence = without_fences(prd_text)
    problems = [f"{OPEN_FENCE}: {open_fence}"] if open_fence else []
    for line in RULE_LINE_RE.findall(text):
        heading = RULE_HEADING_RE.match(line)
        state = heading.group(2).lower() if heading else ""
        if not state.startswith(VALID_STATE_PREFIXES):
            problems.append(f"not written '### R<n> · P<n> · <state>' with a known state: {line.strip()}")
            continue
        number = int(heading.group(1))
        if number in rules:
            problems.append(f"duplicate code R{number}")
        rules[number] = state.startswith(CANCELLED_PREFIXES)
    return rules, problems


def scan_citation(text):
    """Read codes and ranges from the start of a text.

    Returns (numbers, ranges, the text it could not read, whether it ended on a separator).
    """
    numbers = []
    ranges = []
    position = 0
    open_list = False
    while True:
        item = ITEM_RE.match(text, position)
        if item is None:
            break
        low = int(item.group(1))
        high = int(item.group(2)) if item.group(2) else low
        if high < low:
            break
        numbers.extend({low, high})
        if high > low:
            ranges.append((low, high))
        position = item.end()
        separator = SEPARATOR_RE.match(text, position)
        open_list = separator is not None
        if not open_list:
            break
        position = separator.end()
    return numbers, ranges, text[position:].strip(" \t.*_"), open_list


def draft_citation(draft_text):
    """Return (numbers, ranges, [(malformed line, reason)], whether the draft has a Covers line).

    A line or list item that names a code the scan cannot read is malformed. A
    Covers label followed by text with no code at all is prose, not a citation.
    A list under the label ends where its marker changes from bullets to numbers.
    """
    text, open_fence = without_fences(draft_text)
    lines = text.splitlines()
    numbers, ranges = [], []
    malformed = [(open_fence, OPEN_FENCE)] if open_fence else []
    has_line = False
    index = 0

    def read(text):
        found, spans, rest, open_list = scan_citation(text)
        cites = bool(found) or ANY_CODE_RE.search(rest) is not None
        if cites:
            numbers.extend(found)
            ranges.extend(spans)
        return cites, rest, open_list

    while index < len(lines):
        line = lines[index]
        covers = COVERS_LINE_RE.match(line)
        index += 1
        if covers is None:
            continue
        body = covers.group(1).strip()
        if NO_CODES_RE.fullmatch(body.strip("*_ ")):
            has_line = True
        elif body:
            cites, rest, open_list = read(body)
            if not cites:
                continue
            has_line = True
            while open_list and not rest:
                continued = CONTINUED_LINE_RE.match(lines[index]).group(1) if index < len(lines) else ""
                more, _, more_rest, _ = scan_citation(continued)
                if not more or more_rest:
                    malformed.append((line.strip(), DANGLING))
                    break
                _, _, open_list = read(continued)
                index += 1
            if rest:
                malformed.append((line.strip(), UNREADABLE))
        else:
            numbered = None
            while True:
                following = index
                while following < len(lines) and not lines[following].strip():
                    following += 1
                item = LIST_ITEM_RE.match(lines[following]) if following < len(lines) else None
                if item is None or numbered not in (None, item.group(1)[0].isdigit()):
                    break
                cites, rest, _ = read(item.group(2))
                if not cites:
                    break
                has_line = True
                numbered = item.group(1)[0].isdigit()
                if rest:
                    malformed.append((lines[following].strip(), UNREADABLE))
                index = following + 1
    return numbers, ranges, malformed, has_line


def classify(numbers, ranges, rules):
    """Split cited numbers into (active ones, {number: reason} for invalid ones)."""
    active = set()
    invalid = {}
    for number in numbers:
        if number not in rules:
            invalid[number] = "not in the document"
        elif rules[number]:
            invalid[number] = "cancelled"
        else:
            active.add(number)
    for first, last in ranges:
        active.update(n for n, cancelled in rules.items() if first < n < last and not cancelled)
    return active, invalid


def coverage(rules, drafts, out_of_scope=""):
    """Compare the document's rules with the drafts, given as {name: text}."""
    cited_by = {}
    invalid = []
    malformed = []
    without_covers_line = []
    for name, text in drafts.items():
        numbers, ranges, bad_lines, has_line = draft_citation(text)
        active, bad = classify(numbers, ranges, rules)
        for number in active:
            cited_by.setdefault(number, []).append(name)
        invalid.extend({"code": f"R{n}", "draft": name, "reason": bad[n]} for n in sorted(bad))
        malformed.extend({"draft": name, "line": line, "reason": reason} for line, reason in bad_lines)
        if not has_line:
            without_covers_line.append(name)

    numbers, ranges, _, _ = scan_citation(out_of_scope)
    excluded, bad = classify(numbers, ranges, rules)
    bad.update({n: "also cited by a draft" for n in excluded if n in cited_by})
    invalid.extend({"code": f"R{n}", "draft": OUT_OF_SCOPE, "reason": bad[n]} for n in sorted(bad))

    uncovered = [f"R{n}" for n in sorted(rules) if not rules[n] and n not in cited_by and n not in excluded]
    return {
        "ok": not (uncovered or invalid or malformed or without_covers_line),
        "uncovered": uncovered,
        "out_of_scope": [f"R{n}" for n in sorted(excluded)],
        "shared": {f"R{n}": cited_by[n] for n in sorted(cited_by) if len(cited_by[n]) > 1},
        "invalid": invalid,
        "malformed": malformed,
        "without_covers_line": without_covers_line,
    }


def usage_error(message):
    print(json.dumps({"ok": False, "error": message}, ensure_ascii=False), file=sys.stderr)
    return 2


def main(argv=None):
    parser = argparse.ArgumentParser(description="Check that ticket drafts cover a requirement document's rules.")
    parser.add_argument("--prd", required=True, help="Path to the requirement document")
    parser.add_argument(OUT_OF_SCOPE, default="", help="Codes the set leaves out, such as 'R7, R9-R11'")
    parser.add_argument("drafts", nargs="+", help="Paths to the children's drafts in Markdown")
    args = parser.parse_args(argv)

    unread = scan_citation(args.out_of_scope)[2]
    if unread:
        return usage_error(f"{OUT_OF_SCOPE}: codes and ranges separated by commas only; could not read '{unread}'")
    resolved = [pathlib.Path(path).resolve() for path in (args.prd, *args.drafts)]
    if len(set(resolved)) != len(resolved):
        return usage_error("the document and each draft must be different files")
    texts = {}
    for path in (args.prd, *args.drafts):
        try:
            texts[path] = pathlib.Path(path).read_text(encoding="utf-8-sig")
        except (OSError, UnicodeDecodeError) as error:
            return usage_error(f"{path}: cannot read as UTF-8 text ({error})")
    rules, problems = parse_rules(texts.pop(args.prd))
    if problems:
        return usage_error(f"{args.prd}: rule headings: {'; '.join(problems)}")
    if not rules:
        return usage_error(f"{args.prd}: no '### R<n> · P<n> · <state>' rule headings")

    result = coverage(rules, texts, args.out_of_scope)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
