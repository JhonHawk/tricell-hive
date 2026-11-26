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

FENCE_RE = re.compile(r"^\s{0,3}(`{3,}|~{3,})(.*)$")
RULE_LINE_RE = re.compile(r"^(?:###\s+R\d+\b|#{1,6}\s+R\d+\s*·).*$", re.MULTILINE)
RULE_HEADING_RE = re.compile(r"^###\s+R(\d+)\s*·\s*P[012][^·]*·\s*(.+?)\s*$")
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
LINE_PREFIX = r"^\s*(?:>\s*)*(?:#{1,6}\s+|[-*]\s+|\d+[.)]\s+)?"
COVERS_LINE_RE = re.compile(LINE_PREFIX + r"[*_]*(?:Covers|Cubre)[*_]*\s*:[*_]*(.*)$", re.IGNORECASE)
CONTINUED_LINE_RE = re.compile(LINE_PREFIX + r"(.*)$")
LIST_ITEM_RE = re.compile(r"^\s*(?:>\s*)*[-*]\s+(.*)$")
CODE = r"`?[Rr](\d+)`?"
ITEM_RE = re.compile(rf"\s*{CODE}(?:\s*(?:[-–—]|\b(?i:a|al|hasta|to|through|thru)\b)\s*{CODE})?")
SEPARATOR_RE = re.compile(r"\s*(?:[,;]|\b(?i:and|y|e)\b)")
NO_CODES_RE = re.compile(r"(?i:none|ninguna|ninguno)\.?")
OUT_OF_SCOPE = "--out-of-scope"


def without_fences(text):
    """Drop fenced code blocks, which show formats and are not content."""
    kept = []
    fence = None
    for line in text.splitlines():
        marker = FENCE_RE.match(line)
        if fence is None:
            if marker:
                fence = marker.group(1)
            else:
                kept.append(line)
        elif marker and marker.group(1).startswith(fence) and not marker.group(2).strip():
            fence = None
    return "\n".join(kept)


def parse_rules(prd_text):
    """Return ({number: cancelled}, [rule headings the document format rejects])."""
    rules = {}
    problems = []
    for line in RULE_LINE_RE.findall(without_fences(prd_text)):
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
    """Return (numbers, ranges, malformed lines, whether the draft has a Covers line)."""
    numbers, ranges, malformed = [], [], []
    has_line = False
    lines = without_fences(draft_text).splitlines()
    index = 0

    def take(text, line):
        found, spans, rest, open_list = scan_citation(text)
        numbers.extend(found)
        ranges.extend(spans)
        if found and rest:
            malformed.append(line.strip())
        return bool(found), open_list and not rest

    while index < len(lines):
        covers = COVERS_LINE_RE.match(lines[index])
        index += 1
        if covers is None:
            continue
        body = covers.group(1).strip()
        if NO_CODES_RE.fullmatch(body.strip("*_ ")):
            has_line = True
        elif body:
            cited, open_list = take(body, lines[index - 1])
            has_line = has_line or cited
            while open_list and index < len(lines):
                cited, open_list = take(CONTINUED_LINE_RE.match(lines[index]).group(1), lines[index])
                index += cited
        else:
            while index < len(lines) and not lines[index].strip():
                index += 1
            while index < len(lines):
                item = LIST_ITEM_RE.match(lines[index])
                if item is None or not take(item.group(1), lines[index])[0]:
                    break
                has_line = True
                index += 1
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
        malformed.extend({"draft": name, "line": line} for line in bad_lines)
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
