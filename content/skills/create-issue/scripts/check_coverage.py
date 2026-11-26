#!/usr/bin/env python3
"""Requirement coverage of ticket drafts (English and Spanish).

Usage:
  python3 check_coverage.py --prd <prd-file> [--out-of-scope <codes>] <draft-file>...

A draft cites the rules it covers in a line that starts "Covers:" or "Cubre:",
singly or as a range; codes elsewhere in a draft do not count. Reports the rules
no draft cites, those two or more drafts cite, and the cited codes the document
lacks or cancelled. Exits 1 when a rule is uncovered or a cited code is invalid,
2 on a usage error.
"""

import argparse
import json
import pathlib
import re
import sys

RULE_LINE_RE = re.compile(r"^###\s+R\d+\b.*$", re.MULTILINE)
RULE_HEADING_RE = re.compile(r"^###\s+R(\d+)\s*·\s*P[012][^·]*·\s*(.+?)\s*$")
CANCELLED_PREFIXES = ("cancelled", "cancelada", "cancelado")
COVERS_LINE_RE = re.compile(r"^\s*(?:[-*]\s+)?[*_]*(?:Covers|Cubre)[*_]*\s*:(.*)$", re.IGNORECASE | re.MULTILINE)
RANGE_RE = re.compile(r"\bR(\d+)\s*(?:[-–—]|\b(?i:a|al|hasta|to|through|thru)\b)\s*R(\d+)\b")
CODE_RE = re.compile(r"\bR(\d+)\b")
OUT_OF_SCOPE = "--out-of-scope"


def parse_rules(prd_text):
    """Return ({number: cancelled}, [rule headings that do not parse])."""
    rules = {}
    malformed = []
    for line in RULE_LINE_RE.findall(prd_text):
        match = RULE_HEADING_RE.match(line)
        if match is None:
            malformed.append(line.strip())
            continue
        state = match.group(2).strip("*_` ").lower()
        rules[int(match.group(1))] = state.startswith(CANCELLED_PREFIXES)
    return rules, malformed


def cited_numbers(text, rules):
    """Return the numbers a citation text names: (active ones, {number: reason} for invalid ones)."""
    active = set()
    invalid = {}

    def cite(number):
        if number not in rules:
            invalid[number] = "not in the document"
        elif rules[number]:
            invalid[number] = "cancelled"
        else:
            active.add(number)

    for low, high in RANGE_RE.findall(text):
        first, last = sorted((int(low), int(high)))
        cite(first)
        cite(last)
        active.update(n for n, cancelled in rules.items() if first < n < last and not cancelled)
    for number in CODE_RE.findall(RANGE_RE.sub(" ", text)):
        cite(int(number))
    return active, invalid


def coverage(rules, drafts, out_of_scope=""):
    """Compare the document's rules with the drafts, given as {name: text}."""
    cited_by = {}
    invalid = []

    def record(name, bad):
        invalid.extend({"code": f"R{n}", "draft": name, "reason": bad[n]} for n in sorted(bad))

    for name, text in drafts.items():
        active, bad = cited_numbers("\n".join(COVERS_LINE_RE.findall(text)), rules)
        for number in active:
            cited_by.setdefault(number, []).append(name)
        record(name, bad)
    excluded, bad = cited_numbers(out_of_scope, rules)
    record(OUT_OF_SCOPE, bad)

    uncovered = [f"R{n}" for n in sorted(rules) if not rules[n] and n not in cited_by and n not in excluded]
    return {
        "ok": not uncovered and not invalid,
        "uncovered": uncovered,
        "out_of_scope": [f"R{n}" for n in sorted(excluded)],
        "shared": {f"R{n}": cited_by[n] for n in sorted(cited_by) if len(cited_by[n]) > 1},
        "invalid": invalid,
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

    texts = {}
    for path in (args.prd, *args.drafts):
        try:
            texts[path] = pathlib.Path(path).read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError) as error:
            return usage_error(f"{path}: cannot read as UTF-8 text ({error})")
    rules, malformed = parse_rules(texts.pop(args.prd))
    if malformed:
        return usage_error(f"{args.prd}: rule headings not written '### R<n> · P<n> · <state>': {'; '.join(malformed)}")
    if not rules:
        return usage_error(f"{args.prd}: no '### R<n> · P<n> · <state>' rule headings")

    result = coverage(rules, texts, args.out_of_scope)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
