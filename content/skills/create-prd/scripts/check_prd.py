#!/usr/bin/env python3
"""Static ready-list validator for PRD drafts (English and Spanish).

Usage:
  python3 check_prd.py <prd-file>
"""

import argparse
import json
import pathlib
import re
import sys

EXPECTED_SECTIONS = [
    ("problem", {"problem", "problema"}),
    ("goals", {"goals", "objetivos", "metas"}),
    ("non-goals", {"non-goals", "non goals", "fuera de alcance", "no objetivos"}),
    ("user stories", {"user stories", "historias de usuario"}),
    ("requirements", {"requirements", "requisitos", "reglas"}),
    ("acceptance criteria", {"acceptance criteria", "criterios de aceptacion", "criterios de aceptación"}),
    ("success metrics", {"success metrics", "metricas de exito", "métricas de éxito"}),
    ("open questions", {"open questions", "preguntas abiertas"}),
    ("version history", {"version history", "historial de versiones"}),
]

RULE_HEADING_RE = re.compile(
    r"^###\s+(R\d+)\s*·\s*(P[012])(?:\s*\((?:proposed|propuesta)\))?\s*·\s*(.+?)\s*$",
    re.IGNORECASE,
)
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
SOURCE_LINE_RE = re.compile(r"^\s*[-*]\s+\**(?:Source|Fuente)\*?:\*?\s+\S+", re.IGNORECASE)
VERDICT_LINE_RE = re.compile(
    r"^\s*[-*]\s+\**(?:Verdict|Veredicto)\*?:\*?\s+"
    r"(?:exists|existe|new|nueva|nuevo|contradicts|contradice|blocked|bloqueada|bloqueado|not checked|sin revisar)\b",
    re.IGNORECASE,
)
TABLE_RULE_RE = re.compile(r"^\s*\|\s*R\d+\s*\|")
VAGUE_TERMS_RE = re.compile(
    r"(?:\b(?:fast|rápido|rapido|user-friendly|as needed|tbd|to be defined|por definir)\b|\betc\.)",
    re.IGNORECASE,
)


def _norm_heading(text):
    return re.sub(r"\s+", " ", text.strip().lower().strip(":*"))


def validate_prd(text):
    findings = []
    lines = text.splitlines()

    sec_positions = []
    sec_ranges = {}
    h2_headings = []
    for idx, line in enumerate(lines):
        m = re.match(r"^##\s+(.+?)\s*$", line)
        if m:
            h2_headings.append((idx, _norm_heading(m.group(1)), m.group(1).strip()))

    matched_indices = []
    for key, aliases in EXPECTED_SECTIONS:
        found = None
        for h_idx, (line_no, norm, raw) in enumerate(h2_headings):
            if norm in aliases:
                found = (h_idx, line_no, raw)
                break
        if found is None:
            findings.append(f"missing required section: {key}")
        else:
            matched_indices.append((key, found[0], found[1]))
            start_line = found[1] + 1
            end_line = h2_headings[found[0] + 1][0] if found[0] + 1 < len(h2_headings) else len(lines)
            sec_ranges[key] = (start_line, end_line)
            sec_positions.append(found[0])

    if sec_positions != sorted(sec_positions):
        findings.append("sections are not in the required 1..9 order")

    for key, _ in EXPECTED_SECTIONS:
        if key not in sec_ranges:
            continue
        s, e = sec_ranges[key]
        content = [l.strip() for l in lines[s:e] if l.strip() and not l.strip().startswith("<!--")]
        if not content:
            findings.append(f"{key} section is empty")

    seen_codes = set()
    p0_p1_codes = set()
    if "requirements" in sec_ranges:
        s, e = sec_ranges["requirements"]
        req_lines = lines[s:e]
        for offset, line in enumerate(req_lines):
            if TABLE_RULE_RE.match(line):
                findings.append(f"line {s + offset + 1}: rule written as a table row in Requirements")

        rule_starts = []
        for offset, line in enumerate(req_lines):
            if line.startswith("### "):
                rule_starts.append(offset)

        if not rule_starts:
            findings.append("requirements section contains no ### R<n> rule entries")

        for i, r_off in enumerate(rule_starts):
            r_end = rule_starts[i + 1] if i + 1 < len(rule_starts) else len(req_lines)
            heading_line = req_lines[r_off]
            abs_line = s + r_off + 1
            m = RULE_HEADING_RE.match(heading_line)
            if not m:
                findings.append(
                    f"line {abs_line}: malformed rule heading {heading_line!r}; expected '### R<n> · P0|P1|P2 · <state>'"
                )
                continue
            code, prio, state = m.group(1).upper(), m.group(2).upper(), m.group(3).strip().lower()
            if code in seen_codes:
                findings.append(f"line {abs_line}: duplicate rule code {code}")
            seen_codes.add(code)
            is_cancelled = state.startswith(("cancelled", "cancelada", "cancelado"))
            if prio in ("P0", "P1") and not is_cancelled:
                p0_p1_codes.add(code)
            if not state.startswith(VALID_STATE_PREFIXES):
                findings.append(f"line {abs_line}: invalid state {state!r} for {code}")

            block = req_lines[r_off + 1 : r_end]
            body_lines = [l.strip() for l in block if l.strip() and not l.strip().startswith("<!--")]
            has_statement = any(
                not re.match(r"^(?:[-*]|\|)\s*", l) and not SOURCE_LINE_RE.match(l) and not VERDICT_LINE_RE.match(l)
                for l in body_lines
            )
            if not has_statement:
                findings.append(f"line {abs_line}: {code} is missing a rule statement paragraph")
            has_source = any(SOURCE_LINE_RE.match(l) for l in block)
            has_verdict = any(VERDICT_LINE_RE.match(l) for l in block)
            if not has_source:
                findings.append(f"line {abs_line}: {code} is missing its Source:/Fuente: item")
            if not has_verdict:
                findings.append(f"line {abs_line}: {code} is missing a valid Verdict:/Veredicto: item")
            for b_off, b_line in enumerate(block, 1):
                if VAGUE_TERMS_RE.search(b_line):
                    findings.append(f"line {abs_line + b_off}: vague term in {code}: {b_line.strip()}")

    if "acceptance criteria" in sec_ranges and p0_p1_codes:
        s, e = sec_ranges["acceptance criteria"]
        ac_text = "\n".join(lines[s:e])
        for code in sorted(p0_p1_codes):
            if not re.search(rf"\b{code}\b", ac_text):
                findings.append(f"acceptance criteria section does not reference {code}")
        for offset, line in enumerate(lines[s:e]):
            if VAGUE_TERMS_RE.search(line):
                findings.append(f"line {s + offset + 1}: vague term in acceptance criteria: {line.strip()}")

    return {"ok": len(findings) == 0, "findings": findings, "rules_checked": len(seen_codes)}


def main(argv=None):
    parser = argparse.ArgumentParser(description="Validate a PRD Markdown file against the ready list.")
    parser.add_argument("file", help="Path to the PRD Markdown file")
    args = parser.parse_args(argv)
    path = pathlib.Path(args.file)
    if not path.is_file():
        print(json.dumps({"ok": False, "error": f"{path}: file not found"}), file=sys.stderr)
        return 2
    res = validate_prd(path.read_text(encoding="utf-8"))
    print(json.dumps(res, ensure_ascii=False, indent=2))
    return 0 if res["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
