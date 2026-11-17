#!/usr/bin/env python3
"""Static ready-list validator for tracker ticket drafts (English and Spanish).

Usage:
  python3 check_issue.py --type <story|bug|task|spike> [--finding-group] <ticket-file>
"""

import argparse
import json
import pathlib
import re
import sys

BASE_SECTIONS = {
    "outcome": {"outcome", "resultado", "objetivo"},
    "done when": {"done when", "hecho cuando", "criterios de aceptacion", "criterios de aceptación"},
    "out of scope": {"out of scope", "fuera de alcance"},
    "open questions and assumptions": {
        "open questions and assumptions",
        "preguntas abiertas y supuestos",
        "open questions",
        "assumptions",
    },
    "dependencies": {"dependencies", "dependencias"},
}

TYPE_SECTIONS = {
    "story": [
        ("for whom and why", {"for whom and why", "for whom", "para quien y por que", "para quién y por qué", "para quién"}),
    ],
    "bug": [
        ("steps to reproduce", {"steps to reproduce", "pasos para reproducir"}),
        ("expected and observed", {"expected and observed", "esperado y observado", "resultado esperado y observado"}),
        ("environment and version", {"environment and version", "entorno y version", "entorno y versión", "environment"}),
        ("evidence", {"evidence", "evidencia"}),
    ],
    "task": [
        ("the change", {"the change", "el cambio"}),
        ("verify with", {"verify with", "verificar con", "verificacion", "verificación"}),
    ],
    "spike": [
        ("question", {"question", "pregunta"}),
        ("timebox", {"timebox", "limite de tiempo", "límite de tiempo"}),
        ("deliverable", {"deliverable", "entregable"}),
        ("decision it feeds", {"decision it feeds", "decisión que alimenta", "decision que alimenta"}),
    ],
}

HINT_ALIASES = {
    "hint, not binding",
    "hint (not binding)",
    "pista, no vinculante",
    "pista (no vinculante)",
    "solution hint, not binding",
}

VAGUE_TERMS_RE = re.compile(
    r"(?:\b(?:fast|rápido|rapido|user-friendly|as needed|tbd|to be defined|por definir)\b|\betc\.)",
    re.IGNORECASE,
)


def _norm_heading(text):
    return re.sub(r"\s+", " ", text.strip().lower().strip(":*"))


def validate_issue(text, ticket_type="story", finding_group=False):
    ticket_type = ticket_type.lower().strip()
    if ticket_type not in TYPE_SECTIONS:
        raise ValueError(f"unsupported ticket type: {ticket_type!r}")

    findings = []
    lines = text.splitlines()
    if not any(line.startswith("# ") and line[2:].strip() for line in lines):
        findings.append("missing top-level '# <Title>' heading")

    h2_headings = []
    for idx, line in enumerate(lines):
        m = re.match(r"^##\s+(.+?)\s*$", line)
        if m:
            h2_headings.append((idx, _norm_heading(m.group(1)), m.group(1).strip()))

    sec_ranges = {}
    for h_idx, (line_no, norm, raw) in enumerate(h2_headings):
        start = line_no + 1
        end = h2_headings[h_idx + 1][0] if h_idx + 1 < len(h2_headings) else len(lines)
        sec_ranges[norm] = (start, end, raw)

    def find_section(aliases):
        for alias in aliases:
            if alias in sec_ranges:
                return sec_ranges[alias]
        return None

    for key, aliases in BASE_SECTIONS.items():
        if find_section(aliases) is None:
            findings.append(f"missing required base section: {key}")

    for key, aliases in TYPE_SECTIONS[ticket_type]:
        if find_section(aliases) is None:
            findings.append(f"missing required {ticket_type} section: {key}")

    done_sec = find_section(BASE_SECTIONS["done when"])
    criteria_count = 0
    if done_sec is not None:
        s, e, _ = done_sec
        for offset, line in enumerate(lines[s:e]):
            stripped = line.strip()
            if re.match(r"^(?:[-*]|\d+\.)\s+\S+", stripped):
                criteria_count += 1
                if VAGUE_TERMS_RE.search(stripped):
                    findings.append(f"line {s + offset + 1}: vague term in acceptance criterion: {stripped}")
        if criteria_count == 0:
            findings.append("Done when section must contain at least 1 acceptance criterion")
        elif criteria_count > 7 and not finding_group:
            findings.append(
                f"Done when section has {criteria_count} criteria (> 7 limit; break down or pass --finding-group)"
            )

    out_sec = find_section(BASE_SECTIONS["out of scope"])
    if out_sec is not None:
        s, e, _ = out_sec
        if not any(l.strip() for l in lines[s:e] if not l.strip().startswith("<!--")):
            findings.append("Out of scope section must name at least one item")

    for norm, (s, e, raw) in sec_ranges.items():
        if "hint" in norm or "pista" in norm:
            if norm not in HINT_ALIASES:
                findings.append(f"hint section {raw!r} must be explicitly marked as not binding / no vinculante")
            hint_lines = [l.strip() for l in lines[s:e] if l.strip()]
            hint_text = " ".join(hint_lines)
            locations = re.findall(r"`[^`]+`", hint_text)
            if len(locations) > 5:
                findings.append(f"hint names {len(locations)} locations (> 5 maximum)")
            sentences = [seg.strip() for seg in re.split(r"(?<=[.!?])\s+", hint_text) if seg.strip()]
            if len(sentences) > 1 or any(l.startswith(("- ", "* ", "1. ")) for l in hint_lines):
                findings.append("hint must be at most one sentence of direction without implementation steps")

    return {"ok": len(findings) == 0, "type": ticket_type, "criteria_count": criteria_count, "findings": findings}


def main(argv=None):
    parser = argparse.ArgumentParser(description="Validate a tracker ticket draft against the ready list.")
    parser.add_argument("--type", required=True, choices=sorted(TYPE_SECTIONS), help="Ticket type")
    parser.add_argument("--finding-group", action="store_true", help="Allow > 7 criteria for a grouped finding item")
    parser.add_argument("file", help="Path to the ticket Markdown file")
    args = parser.parse_args(argv)
    path = pathlib.Path(args.file)
    if not path.is_file():
        print(json.dumps({"ok": False, "error": f"{path}: file not found"}), file=sys.stderr)
        return 2
    res = validate_issue(path.read_text(encoding="utf-8"), ticket_type=args.type, finding_group=args.finding_group)
    print(json.dumps(res, ensure_ascii=False, indent=2))
    return 0 if res["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
