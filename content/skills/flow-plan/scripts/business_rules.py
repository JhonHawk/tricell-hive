#!/usr/bin/env python3
"""Render and splice the marker-bounded business rules section in a ticket description.

Subcommands:
  splice-description  Insert at the end on first write, or replace only between
                      `<heading>` and `<heading> · end`, verifying that all text
                      outside the markers remains byte-for-byte identical.
"""

import argparse
import json
import pathlib
import re
import sys

LEADING_DATE_LINE = re.compile(r"^\d{4}-\d{2}-\d{2}(?:\s+·\s+(.*))?$")


def _normalize_ignoring_date(section_body):
    lines = [line.rstrip() for line in section_body.strip().splitlines()]
    if lines:
        m = LEADING_DATE_LINE.match(lines[0])
        if m:
            lines[0] = f"<DATE> · {m.group(1)}" if m.group(1) else "<DATE>"
    return "\n".join(lines)


def _find_marker_lines(lines, marker):
    return [i for i, line in enumerate(lines) if line.strip() == marker]


def splice_description(description, heading, section_body):
    heading = heading.strip()
    end_marker = f"{heading} · end"
    body = section_body.strip("\r\n")
    block = f"{heading}\n{body}\n{end_marker}"

    lines = description.splitlines(keepends=True)
    start_idxs = _find_marker_lines(lines, heading)
    end_idxs = _find_marker_lines(lines, end_marker)

    if not start_idxs and not end_idxs:
        if not description:
            new_desc = block + "\n"
            prefix = ""
        elif description.endswith("\n\n"):
            prefix = description
            new_desc = prefix + block + "\n"
        elif description.endswith("\n"):
            prefix = description + "\n"
            new_desc = prefix + block + "\n"
        else:
            prefix = description + "\n\n"
            new_desc = prefix + block + "\n"
        if not new_desc.encode("utf-8").startswith(description.encode("utf-8")):
            return {
                "ok": False,
                "changed": False,
                "fallback_to_comment": True,
                "reason": "prefix byte verification failed on first write",
            }
        return {
            "ok": True,
            "changed": True,
            "first_write": True,
            "fallback_to_comment": False,
            "description": new_desc,
        }

    if len(start_idxs) != 1 or len(end_idxs) != 1 or start_idxs[0] >= end_idxs[0]:
        return {
            "ok": False,
            "changed": False,
            "fallback_to_comment": True,
            "reason": "missing or duplicate business-rules section markers in description",
        }

    s_idx = start_idxs[0]
    e_idx = end_idxs[0]
    prefix = "".join(lines[:s_idx])
    suffix = "".join(lines[e_idx + 1 :])
    existing_inner = "".join(lines[s_idx + 1 : e_idx]).strip("\r\n")

    if _normalize_ignoring_date(existing_inner) == _normalize_ignoring_date(body):
        return {
            "ok": True,
            "changed": False,
            "first_write": False,
            "fallback_to_comment": False,
            "description": description,
        }

    trailing_nl = "\n" if lines[e_idx].endswith("\n") else ""
    new_desc = prefix + block + trailing_nl + suffix
    if not (
        new_desc.encode("utf-8").startswith(prefix.encode("utf-8"))
        and new_desc.encode("utf-8").endswith(suffix.encode("utf-8"))
    ):
        return {
            "ok": False,
            "changed": False,
            "fallback_to_comment": True,
            "reason": "outer text byte-for-byte verification failed",
        }

    return {
        "ok": True,
        "changed": True,
        "first_write": False,
        "fallback_to_comment": False,
        "description": new_desc,
    }


def main(argv=None):
    parser = argparse.ArgumentParser(description="Splice business rules section between markers in a ticket description.")
    sub = parser.add_subparsers(dest="cmd", required=True)

    p_splice = sub.add_parser("splice-description")
    p_splice.add_argument("--description-file", required=True)
    p_splice.add_argument("--heading", required=True)
    p_splice.add_argument("--body-file", required=True)
    p_splice.add_argument("--out", default=None)

    args = parser.parse_args(argv)
    if args.cmd == "splice-description":
        desc = pathlib.Path(args.description_file).read_text(encoding="utf-8")
        body = pathlib.Path(args.body_file).read_text(encoding="utf-8")
        res = splice_description(desc, args.heading, body)
        if res.get("ok") and args.out:
            pathlib.Path(args.out).write_text(res["description"], encoding="utf-8")
        print(json.dumps(res, ensure_ascii=False, indent=2))
        return 0 if res.get("ok") else 1
    return 2


if __name__ == "__main__":
    sys.exit(main())
