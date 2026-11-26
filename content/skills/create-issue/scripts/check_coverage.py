#!/usr/bin/env python3
"""Requirement coverage of ticket drafts (English and Spanish).

Usage:
  python3 check_coverage.py --prd <prd-file> [--out-of-scope <codes>] <draft-file>...

A draft cites the rules it covers in a line that starts "Covers:" or "Cubre:",
singly or as a range, on that line or listed right under it; codes elsewhere in
a draft do not count. Reports the rules no draft cites, those two or more drafts
cite, the cited codes the document lacks or cancelled, and the Covers lines that
hold text besides codes. Exits 1 on any of those except shared rules, 2 on a
usage error.
"""

import argparse
import json
import pathlib
import re
import sys

RULE_LINE_RE = re.compile(r"^#{1,6}\s+R\d+\b.*$", re.MULTILINE)
RULE_HEADING_RE = re.compile(r"^###\s+R(\d+)\s*·\s*P[012][^·]*·\s*(.+?)\s*$")
CANCELLED_PREFIXES = ("cancelled", "canceled", "cancelada", "cancelado")
LINE_PREFIX = r"^\s*(?:>\s*)*(?:#{1,6}\s+|[-*]\s+|\d+[.)]\s+)?"
COVERS_LINE_RE = re.compile(LINE_PREFIX + r"[*_]*(?:Covers|Cubre)[*_]*\s*:[*_]*(.*)$", re.IGNORECASE)
LIST_LINE_RE = re.compile(LINE_PREFIX + r"(.*)$")
CONNECTOR = r"(?:[-–—]|\b(?i:a|al|hasta|to|through|thru)\b)"
PART_RE = re.compile(
    rf"\s*(?:(?P<chain>[Rr]\d+(?:\s*{CONNECTOR}\s*[Rr]?\d+)+)|(?P<code>[Rr]\d+\b)|[,;.&*_]|\b(?i:and|y|e)\b)"
)
NO_CODES_RE = re.compile(r"(?i:none|ninguna|ninguno|n/a)?[\s.\-–—]*")
OUT_OF_SCOPE = "--out-of-scope"


def without_fences(text):
    """Drop the lines of fenced code blocks, which show formats and are not content."""
    kept = []
    fenced = False
    for line in text.splitlines():
        if line.lstrip().startswith(("```", "~~~")):
            fenced = not fenced
        elif not fenced:
            kept.append(line)
    return "\n".join(kept)


def parse_rules(prd_text):
    """Return ({number: cancelled}, [problems in the rule headings])."""
    rules = {}
    problems = []
    for line in RULE_LINE_RE.findall(without_fences(prd_text)):
        match = RULE_HEADING_RE.match(line)
        if match is None:
            problems.append(f"not written '### R<n> · P<n> · <state>': {line.strip()}")
            continue
        number = int(match.group(1))
        if number in rules:
            problems.append(f"duplicate code R{number}")
        state = re.sub(r"[*_`~]", "", match.group(2)).strip().lower()
        rules[number] = state.startswith(CANCELLED_PREFIXES)
    return rules, problems


def scan_citation(text):
    """Read the codes at the start of a citation text: (listed numbers, ranges, text left over)."""
    numbers = []
    ranges = []
    position = 0
    while True:
        part = PART_RE.match(text, position)
        if part is None:
            break
        if part.group("chain"):
            chain = [int(n) for n in re.findall(r"\d+", part.group("chain"))]
            numbers.extend(chain)
            ranges.append((min(chain), max(chain)))
        elif part.group("code"):
            numbers.append(int(part.group("code")[1:]))
        position = part.end()
    return numbers, ranges, text[position:].strip()


def draft_citation(draft_text):
    """Return (numbers, ranges, malformed Covers lines, whether the draft has a Covers line)."""
    numbers, ranges, malformed = [], [], []
    has_line = False
    lines = without_fences(draft_text).splitlines()
    index = 0
    while index < len(lines):
        covers = COVERS_LINE_RE.match(lines[index])
        index += 1
        if covers is None:
            continue
        has_line = True
        found, spans, rest = scan_citation(covers.group(1))
        if rest and not NO_CODES_RE.fullmatch(rest):
            malformed.append(lines[index - 1].strip())
        while index < len(lines):
            more, more_spans, more_rest = scan_citation(LIST_LINE_RE.match(lines[index]).group(1))
            if not more or more_rest:
                break
            found.extend(more)
            spans.extend(more_spans)
            index += 1
        numbers.extend(found)
        ranges.extend(spans)
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

    numbers, ranges, _ = scan_citation(out_of_scope)
    excluded, bad = classify(numbers, ranges, rules)
    bad.update({n: "also cited by a draft" for n in excluded if n in cited_by})
    invalid.extend({"code": f"R{n}", "draft": OUT_OF_SCOPE, "reason": bad[n]} for n in sorted(bad))

    uncovered = [f"R{n}" for n in sorted(rules) if not rules[n] and n not in cited_by and n not in excluded]
    return {
        "ok": not uncovered and not invalid and not malformed,
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
