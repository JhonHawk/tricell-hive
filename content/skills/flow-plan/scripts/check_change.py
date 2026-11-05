#!/usr/bin/env python3
"""Check that a change folder stays coherent after its plan is revised.

Usage: python3 check_change.py <change-folder> [--archive]

Reads proposal.md, tasks.md, design.md and specs/**/*.md and prints one
`<file>:<line>: <finding>` line per problem:

- a criterion defined in proposal.md that no **Closes:** line of tasks.md names
  (ranges such as AC3–AC5 count), unless it is retired or replaced;
- a **Closes:** line that names a criterion proposal.md does not define;
- a mention of an ID that a **Replaces:** marker replaced, or of a criterion a
  **Retired:** marker retired, outside the lines that may keep it: the marker
  lines, the line that defines the old ID, a line that also names its
  replacement, and the progress section (a `##` heading containing
  "progress", "progreso" or "avance");
- with --archive, each unchecked `- [ ]` box of tasks.md, then their count.

Exit status: 0 when nothing is printed, 1 when something is, 2 when the folder,
its proposal.md or its tasks.md cannot be read. It only reads.
"""

import re
import sys
from pathlib import Path

ID_TOKEN = re.compile(r"(?<![A-Za-z0-9])([A-Z]{1,3})(\d+)(?:[- ]([A-Z]))?(?![A-Za-z0-9])")
REPLACES = re.compile(r"\*{0,2}Replaces:\*{0,2}")
RETIRED = re.compile(r"\*{0,2}Retired:\*{0,2}")
AC_DEFINITION = re.compile(r"^\s*[-*]\s+\**AC(\d+)\b")
CLOSES = re.compile(r"\*{0,2}Closes:\*{0,2}:?(.*)$")
AC_REFERENCE = re.compile(r"AC(\d+)(?:\s*[-–—]\s*(?:AC)?(\d+))?")
UNCHECKED = re.compile(r"^\s*[-*]\s+\[ \]\s*(.*)$")
PROGRESS_HEADING = re.compile(r"^##\s.*(progress|progreso|avance)", re.IGNORECASE)
LEVEL2_HEADING = re.compile(r"^##\s")


def id_label(prefix, number, letter):
    return f"{prefix}{number}" + (f"-{letter}" if letter else "")


def id_pattern(prefix, number):
    return re.compile(rf"(?<![A-Za-z0-9]){prefix}{number}(?![0-9])")


def read_lines(path):
    return path.read_text(encoding="utf-8").splitlines()


def scanned_files(folder):
    files = [folder / name for name in ("proposal.md", "design.md", "tasks.md")]
    files += sorted((folder / "specs").rglob("*.md")) if (folder / "specs").is_dir() else []
    return [f for f in files if f.is_file()]


def exempt_progress(lines):
    """Return the indexes of the lines inside a progress section."""
    exempt, inside = set(), False
    for i, line in enumerate(lines):
        if LEVEL2_HEADING.match(line):
            inside = bool(PROGRESS_HEADING.match(line))
        elif inside:
            exempt.add(i)
    return exempt


def collect_replacements(docs):
    """Map each replaced or retired ID to (where, replacing ID pattern or None)."""
    old = {}
    for rel, lines in docs.items():
        for n, line in enumerate(lines, 1):
            marker = REPLACES.search(line)
            if marker:
                before = ID_TOKEN.findall(line[: marker.start()])
                new = id_pattern(before[0][0], before[0][1]) if before else None
                for prefix, number, letter in ID_TOKEN.findall(line[marker.end():]):
                    old.setdefault((prefix, number), (id_label(prefix, number, letter), f"{rel}:{n}", new, "replaced"))
            if RETIRED.search(line):
                match = AC_DEFINITION.match(line)
                if match:
                    old.setdefault(("AC", match.group(1)), (f"AC{match.group(1)}", f"{rel}:{n}", None, "retired"))
    return old


def stale_mentions(docs, old):
    findings = []
    for rel, lines in docs.items():
        progress = exempt_progress(lines)
        for i, line in enumerate(lines):
            if i in progress or REPLACES.search(line) or RETIRED.search(line):
                continue
            for (prefix, number), (label, where, new, how) in old.items():
                pattern = id_pattern(prefix, number)
                if not pattern.search(line):
                    continue
                if re.match(rf"^\s*[-*]\s+\**{prefix}{number}(?![0-9])", line):
                    continue  # the line that defines the old ID keeps it
                if new is not None and new.search(line):
                    continue  # the line names the replacement too
                findings.append(f"{rel}:{i + 1}: {label} was {how} ({where}) but is still named here")
    return findings


def criteria_findings(proposal, tasks, old):
    defined = {}
    for n, line in enumerate(proposal, 1):
        match = AC_DEFINITION.match(line)
        if match:
            defined.setdefault(int(match.group(1)), n)
    if not defined:
        return []
    findings, closed = [], set()
    for n, line in enumerate(tasks, 1):
        match = CLOSES.search(line)
        if not match:
            continue
        for start, end in AC_REFERENCE.findall(match.group(1)):
            first, last = int(start), int(end or start)
            for named in {first, last}:
                if named not in defined:
                    findings.append(f"tasks.md:{n}: AC{named} in **Closes:** is not defined in proposal.md")
            closed.update(range(first, min(last, first + 999) + 1))
    gone = {int(number) for prefix, number in old if prefix == "AC"}
    for number, n in sorted(defined.items()):
        if number not in closed and number not in gone:
            findings.append(f"proposal.md:{n}: AC{number} is named by no **Closes:** line in tasks.md")
    return findings


def main(argv):
    args = [a for a in argv if a != "--archive"]
    if len(args) != 1:
        print(__doc__.strip().splitlines()[2], file=sys.stderr)
        return 2
    folder = Path(args[0])
    for required in ("proposal.md", "tasks.md"):
        if not (folder / required).is_file():
            print(f"check_change: {folder / required}: not found", file=sys.stderr)
            return 2
    try:
        docs = {f.relative_to(folder).as_posix(): read_lines(f) for f in scanned_files(folder)}
    except (OSError, UnicodeDecodeError) as error:
        print(f"check_change: {error}", file=sys.stderr)
        return 2

    old = collect_replacements(docs)
    findings = criteria_findings(docs["proposal.md"], docs["tasks.md"], old)
    findings += stale_mentions(docs, old)
    if "--archive" in argv:
        boxes = [
            f"tasks.md:{n}: unchecked box: {m.group(1).strip()}"
            for n, line in enumerate(docs["tasks.md"], 1)
            if (m := UNCHECKED.match(line))
        ]
        if boxes:
            findings += boxes
            findings.append(f"{len(boxes)} unchecked box" + ("" if len(boxes) == 1 else "es"))
    for line in findings:
        print(line)
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
