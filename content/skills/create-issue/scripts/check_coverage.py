#!/usr/bin/env python3
"""Requirement coverage of ticket drafts (English and Spanish).

Usage:
  python3 check_coverage.py --prd <prd-file> <draft-file>...

Reports the requirement codes no draft cites, those two or more drafts cite,
and the codes a draft cites that the document lacks or cancelled. Exits 1 when
a requirement is uncovered or a cited code is invalid, 2 on a usage error.
"""

import argparse
import json
import pathlib
import re
import sys

RULE_HEADING_RE = re.compile(r"^###\s+R(\d+)\s*·\s*P[012][^·]*·\s*(.+?)\s*$", re.IGNORECASE | re.MULTILINE)
CANCELLED_PREFIXES = ("cancelled", "cancelada", "cancelado")
RANGE_RE = re.compile(r"\bR(\d+)\s*(?:[-–—]|\ba\b|\bto\b)\s*R(\d+)\b", re.IGNORECASE)
CODE_RE = re.compile(r"\bR(\d+)\b", re.IGNORECASE)


def parse_rules(prd_text):
    """Return {number: cancelled} for each rule heading of the document."""
    return {
        int(number): state.strip().lower().startswith(CANCELLED_PREFIXES)
        for number, state in RULE_HEADING_RE.findall(prd_text)
    }


def cited_numbers(draft_text, rules):
    """Return the numbers a draft cites: (active ones, {number: reason} for invalid ones)."""
    active = set()
    invalid = {}

    def cite(number):
        if number not in rules:
            invalid[number] = "not in the document"
        elif rules[number]:
            invalid[number] = "cancelled"
        else:
            active.add(number)

    for low, high in RANGE_RE.findall(draft_text):
        first, last = sorted((int(low), int(high)))
        cite(first)
        cite(last)
        active.update(n for n, cancelled in rules.items() if first < n < last and not cancelled)
    for number in CODE_RE.findall(RANGE_RE.sub(" ", draft_text)):
        cite(int(number))
    return active, invalid


def coverage(rules, drafts):
    """Compare the document's rules with the drafts, given as {name: text}."""
    cited_by = {}
    invalid = []
    for name, text in drafts.items():
        active, bad = cited_numbers(text, rules)
        for number in active:
            cited_by.setdefault(number, []).append(name)
        invalid.extend({"code": f"R{n}", "draft": name, "reason": bad[n]} for n in sorted(bad))
    uncovered = [f"R{n}" for n in sorted(rules) if not rules[n] and n not in cited_by]
    shared = {f"R{n}": cited_by[n] for n in sorted(cited_by) if len(cited_by[n]) > 1}
    return {"ok": not uncovered and not invalid, "uncovered": uncovered, "shared": shared, "invalid": invalid}


def main(argv=None):
    parser = argparse.ArgumentParser(description="Check that ticket drafts cover a requirement document's rules.")
    parser.add_argument("--prd", required=True, help="Path to the requirement document")
    parser.add_argument("drafts", nargs="+", help="Paths to the ticket drafts in Markdown")
    args = parser.parse_args(argv)

    paths = [pathlib.Path(args.prd), *(pathlib.Path(d) for d in args.drafts)]
    missing = [str(p) for p in paths if not p.is_file()]
    if missing:
        print(json.dumps({"ok": False, "error": f"file not found: {', '.join(missing)}"}), file=sys.stderr)
        return 2
    rules = parse_rules(paths[0].read_text(encoding="utf-8"))
    if not rules:
        print(json.dumps({"ok": False, "error": f"{paths[0]}: no '### R<n> · P<n> · <state>' rule headings"}), file=sys.stderr)
        return 2

    result = coverage(rules, {str(p): p.read_text(encoding="utf-8") for p in paths[1:]})
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
