#!/usr/bin/env python3
"""Check that a change folder stays coherent after its plan is revised.

Usage: python3 check_change.py <change-folder> [--archive]

Reads proposal.md, tasks.md, design.md and specs/**/*.md and prints one
`<file>:<line>: <finding>` line per problem:

- a criterion defined in proposal.md, outside its progress section, that no
  **Closes:** line of tasks.md names (ranges such as AC3–AC5 count, in either
  order), unless it is retired or replaced; markers may be written `**Closes:**`,
  `**Closes**:` or `Closes:`;
- a **Closes:** line that names a criterion proposal.md does not define;
- a mention of an ID listed right after a **Replaces:** marker (D4-A names
  that option only, D4 any option), or of a criterion a **Retired:** marker
  retired, outside the lines that may keep it: the marker lines, the
  proposal.md line that defines the old ID, a line that also names its
  replacement, and the progress section (a `##` heading containing
  "progress", "progreso" or "avance");
- a proposal.md without `- AC<n>` criteria while tasks.md has **Closes:** lines;
- with --archive, each unchecked `- [ ]` box of tasks.md, then their count.

Exit status: 0 when nothing is printed, 1 when something is, 2 when the folder,
its proposal.md or its tasks.md cannot be read. It only reads.
"""

import re
import sys
from pathlib import Path

ID = r"[A-Z]{1,3}\d+(?:[- ][A-Z])?"
ID_TOKEN = re.compile(r"(?<![A-Za-z0-9])([A-Z]{1,3})(\d+)(?:[- ]([A-Z]))?(?![A-Za-z0-9])")
ID_LIST = re.compile(rf"\s*((?<![A-Za-z0-9]){ID}(?:\s*(?:,|;|/|&|\band\b|\by\b|\be\b)\s*{ID})*)")


def marker(word):
    """Match `**Word:**`, `**Word**:` and `Word:`."""
    return re.compile(rf"\*{{0,2}}{word}(?::\*{{0,2}}|\*{{0,2}}:)")


REPLACES = marker("Replaces")
RETIRED = marker("Retired")
CLOSES = re.compile(marker("Closes").pattern + r"(.*)$")
AC_DEFINITION = re.compile(r"^\s*[-*]\s+\**AC(\d+)\b")
AC_REFERENCE = re.compile(r"AC(\d+)(?:\s*[-–—]\s*(?:AC)?(\d+))?")
UNCHECKED = re.compile(r"^\s*[-*]\s+\[ \]\s*(.*)$")
PROGRESS_HEADING = re.compile(r"^##\s.*(progress|progreso|avance)", re.IGNORECASE)
LEVEL2_HEADING = re.compile(r"^##\s")


def id_label(prefix, number, letter):
    return f"{prefix}{number}" + (f"-{letter}" if letter else "")


def id_pattern(prefix, number, letter=""):
    """An ID with an option letter matches only that option; a bare ID matches any."""
    if letter:
        return re.compile(rf"(?<![A-Za-z0-9]){prefix}{number}[- ]?{letter}(?![A-Za-z0-9])")
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
    """Map each replaced or retired ID to its label, origin, patterns and kind."""
    old = {}
    for rel, lines in docs.items():
        for n, line in enumerate(lines, 1):
            found = REPLACES.search(line)
            if found:
                before = ID_TOKEN.findall(line[: found.start()])
                new = id_pattern(*before[0]) if before else None
                listed = ID_LIST.match(line[found.end():])
                for prefix, number, letter in ID_TOKEN.findall(listed.group(1) if listed else ""):
                    old.setdefault((prefix, number, letter), {
                        "label": id_label(prefix, number, letter), "where": f"{rel}:{n}", "how": "replaced",
                        "pattern": id_pattern(prefix, number, letter), "new": new,
                        "definition": re.compile(rf"^\s*[-*]\s+\**{prefix}{number}(?![0-9])"),
                    })
            if RETIRED.search(line):
                match = AC_DEFINITION.match(line)
                if match:
                    number = match.group(1)
                    old.setdefault(("AC", number, ""), {
                        "label": f"AC{number}", "where": f"{rel}:{n}", "how": "retired",
                        "pattern": id_pattern("AC", number), "new": None,
                        "definition": re.compile(rf"^\s*[-*]\s+\**AC{number}(?![0-9])"),
                    })
    return old


def stale_mentions(docs, old):
    findings = []
    for rel, lines in docs.items():
        progress = exempt_progress(lines)
        for i, line in enumerate(lines):
            if i in progress or REPLACES.search(line) or RETIRED.search(line):
                continue
            for item in old.values():
                if not item["pattern"].search(line):
                    continue
                if rel == "proposal.md" and item["definition"].match(line):
                    continue  # the proposal line that defines the old ID keeps it
                if item["new"] is not None and item["new"].search(line):
                    continue  # the line names the replacement too
                findings.append(f"{rel}:{i + 1}: {item['label']} was {item['how']} ({item['where']}) but is still named here")
    return findings


def criteria_findings(proposal, tasks, old):
    progress = exempt_progress(proposal)
    defined = {}
    for i, line in enumerate(proposal):
        match = AC_DEFINITION.match(line)
        if match and i not in progress:
            defined.setdefault(int(match.group(1)), i + 1)
    closes = [(n, m.group(1)) for n, line in enumerate(tasks, 1) if (m := CLOSES.search(line))]
    if not defined:
        if closes:
            return ["proposal.md: no criterion is written as a `- AC<n>` bullet, but tasks.md has **Closes:** lines"]
        return []
    findings, closed = [], set()
    for n, text in closes:
        for start, end in AC_REFERENCE.findall(text):
            first, last = sorted((int(start), int(end or start)))
            for named in sorted({first, last}):
                if named not in defined:
                    findings.append(f"tasks.md:{n}: AC{named} in **Closes:** is not defined in proposal.md")
            closed.update(range(first, min(last, first + 999) + 1))
    gone = {int(number) for prefix, number, _ in old if prefix == "AC"}
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
