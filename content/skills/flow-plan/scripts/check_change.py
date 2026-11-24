#!/usr/bin/env python3
"""Check that a plan file stays coherent after it is revised.

Usage: python3 check_change.py <plan.md>

Reads the one plan file `<YYYY-MM-DD>-<work>.plan.md` and prints one
`<plan file>:<line>: <finding>` line per problem:

- a criterion, written as an `AC<n>` line (a bullet or not) before the first
  `### T<n>` task heading and outside the progress section, that no **Closes:**
  line of the tasks names (ranges such as AC3–AC5 count, in either order), unless
  it is retired or replaced; markers may be written `**Closes:**`, `**Closes**:` or
  `Closes:`;
- a **Closes:** line that names a criterion the plan does not define;
- a mention of an ID listed right after a **Replaces:** marker (D4-A names
  that option only, D4 any option), or of a criterion a **Retired:** marker
  retired, outside the lines that may keep it: the marker lines, the line that
  defines the old ID, a line that also names its replacement, and the progress
  section (a `##` heading containing "progress", "progreso" or "avance", or a
  `## Checkpoint` heading with its `###` subsections);
- a **Closes:** line naming a criterion when the plan has no `AC<n>` line;
- a criterion, neither retired nor replaced, none of whose lines, read with the
  lines indented under it up to a nested criterion, names a place after its last
  **Observed in:** (also `**Observed in**:` or `Observed in:`); an empty marker, a
  dash or a `<…>` placeholder that is not a link is no place;
- a malformed unit heading or box under `## Delivery units`: headings
  `### E<n> — <unit> · depends on: <…> · files: <…>` (or `· depende de: … · archivos: …`); boxes
  `unit: <command>`, `live: <scenario> · pass: <predicate>`, `perf: <metric> · base: <value> ·
  fail: <threshold>` (or `falla:`); a predicate written `n/a: <reason>` replaces pass, base and
  threshold;
- a malformed line under `## Business rules`;
- a malformed `- Rules:` or `- Principles:` line in the `## Checkpoint` section: Rules is one
  or more `<flow>/stage[-<name>].md` joined by `, `, optionally followed by ` · extra: <leaf>, <leaf>`;
  Principles is `none` or principle names (`[a-z0-9-]+`) joined by `, `, never `none` with a name;
- a `- Execution record:` pointer in `## Checkpoint` whose record fails `tasks.py` validation.

Exit status: 0 when nothing is printed, 1 when something is, 2 when the arguments are
invalid or the plan cannot be read. It only reads.
"""

import re
import importlib.util
import sys
from pathlib import Path

ID = r"(?<![A-Za-z0-9])[A-Z]{1,3}\d+(?:[- ][A-Z](?![A-Za-z0-9]))?(?![A-Za-z0-9])"
ID_TOKEN = re.compile(r"(?<![A-Za-z0-9])([A-Z]{1,3})(\d+)(?:[- ]([A-Z]))?(?![A-Za-z0-9])")
WRAP = r"[`*\[\]]*"
ID_SEPARATOR = r"\s*(?:,|;|/|&|–|—|-|\band\b|\by\b|\be\b)\s*"
ID_LIST = re.compile(rf"\s*({WRAP}{ID}{WRAP}(?:{ID_SEPARATOR}{WRAP}{ID}{WRAP})*)")
AC_RANGE = re.compile(r"AC(\d+)\s*[-–—]\s*(?:AC)?(\d+)")


def marker(word):
    """Match `**Word:**`, `**Word**:` and `Word:`."""
    return re.compile(rf"\*{{0,2}}{word}(?::\*{{0,2}}|\*{{0,2}}:)")


REPLACES = marker("Replaces")
RETIRED = marker("Retired")
CLOSES = re.compile(marker("Closes").pattern + r"(.*)$")
OBSERVED = marker("Observed in")
PLACEHOLDER = re.compile(r"^<(?![A-Za-z][A-Za-z0-9+.-]*:)[^>]*>")
AC_DEFINITION = re.compile(r"^\s*(?:[-*]\s+)?\**AC(\d+)\b")
TASK_HEADING = re.compile(r"^###\s+T\d+\b")
AC_REFERENCE = re.compile(r"AC(\d+)(?:\s*[-–—]\s*(?:AC)?(\d+))?")
PROGRESS_HEADING = re.compile(r"^##\s.*(progress|progreso|avance)|^##\s+checkpoint\s*$", re.IGNORECASE)
BUSINESS_RULES_HEADING = re.compile(r"^##\s+Business rules\s*$")
LEVEL2_HEADING = re.compile(r"^##\s")
UNITS_HEADING = re.compile(r"^##\s+Delivery units\s*$")
UNIT_HEADING = re.compile(r"^###\s")
UNIT_TITLE = re.compile(
    r"^###\s+E\d+\s+[—–-]\s+\S.*?\s+·\s+(?:depende de|depends on):\s+\S.*?\s+·\s+(?:archivos|files):\s+\S.*$"
)
BOX_LINE = re.compile(r"^\s*[-*]\s+\[[ xX]\]\s*(.*)$")
NOT_APPLICABLE = re.compile(r"n/a:\s*\S", re.IGNORECASE)
FIELD = r"\s+·\s+{}:\s+(?![Nn]/[Aa]:)\S"
UNIT_BOX = re.compile(r"^unit:\s+(?![Nn]/[Aa]:)\S")
LIVE_BOX = re.compile(r"^live:\s+\S.*" + FIELD.format("pass") + r".*$")
PERF_BOX = re.compile(
    r"^perf:\s+\S.*" + FIELD.format("base") + r".*" + FIELD.format("(?:falla|fail)") + r".*$"
)

BR_NONE = re.compile(r"^None:\s+\S")
BR_TICKET_SECTION = re.compile(r"^ticket section:\s+[^·]+?\s+·\s+\S.*$")
BR_TICKET_WRITE = re.compile(r"^ticket write:\s+\d{4}-\d{2}-\d{2}\s+·\s+[^·]+?\s+·\s+\S+$")
BR_CARD_STALE = re.compile(r"^card stale:\s+[^·]+?\s+·\s+\S.*$")
BR_FRAME_STALE = re.compile(r"^frame stale:\s+[^·]+?\s+·\s+\S.*$")
BR_RULE = re.compile(
    r"^rule:\s+(.+?)\s+·\s+screen\s+([^·]+?)\s+·\s+(.+?)\s+·\s+found by\s+(drawer|reviewer|research|plan|build|verify)\s*$"
)
BR_STATUS = re.compile(
    r"^(?:proposed|assumption|scope|approved\s+\d{4}-\d{2}-\d{2}:\s+\S.*|cancelled\s+\d{4}-\d{2}-\d{2}:\s+.+?\s+·\s+replaced by\s+\S.*)$"
)


def id_label(prefix, number, letter):
    return f"{prefix}{number}" + (f"-{letter}" if letter else "")


def id_pattern(prefix, number, letter=""):
    """An ID with an option letter matches only that option; a bare ID matches any."""
    if letter:
        return re.compile(rf"(?<![A-Za-z0-9]){prefix}{number}[- ]?{letter}(?![A-Za-z0-9])")
    return re.compile(rf"(?<![A-Za-z0-9]){prefix}{number}(?![0-9])")


def read_lines(path):
    return path.read_text(encoding="utf-8").splitlines()


def task_start(lines):
    """The index of the first `### T<n>` task heading, or the number of lines when there is none."""
    return next((i for i, line in enumerate(lines) if TASK_HEADING.match(line)), len(lines))


ACCEPTANCE_HEADING = re.compile(r"^##\s+(?:Alcance y aceptación|Scope and acceptance)\s*$", re.IGNORECASE)


def acceptance_range(lines):
    """The (start, end) line indexes of the scope-and-acceptance section, before the first task; everything before the first task when the heading is worded otherwise."""
    limit = task_start(lines)
    for i, line in enumerate(lines[:limit]):
        if ACCEPTANCE_HEADING.match(line):
            end = next((j for j in range(i + 1, limit) if LEVEL2_HEADING.match(lines[j])), limit)
            return i + 1, end
    return 0, limit


def exempt_progress(lines):
    """Return the indexes of the lines inside a progress or checkpoint section."""
    exempt, inside = set(), False
    for i, line in enumerate(lines):
        if LEVEL2_HEADING.match(line):
            inside = bool(PROGRESS_HEADING.match(line))
        elif inside:
            exempt.add(i)
    return exempt


def collect_replacements(lines, name):
    """Map each replaced or retired ID to its label, origin, patterns and kind."""
    old = {}
    for n, line in enumerate(lines, 1):
        found = REPLACES.search(line)
        if found:
            before = ID_TOKEN.findall(line[: found.start()])
            listed = ID_LIST.match(line[found.end():])
            listed = listed.group(1) if listed else ""
            ids = ID_TOKEN.findall(listed)
            for start, end in AC_RANGE.findall(listed):
                first, last = sorted((int(start), int(end)))
                ids += [("AC", str(k), "") for k in range(first, min(last, first + 999) + 1)]
            for prefix, number, letter in ids:
                new = None
                if before:
                    # The replacement keeps its letter only when it is another option of the same decision.
                    same = (before[0][0], before[0][1]) == (prefix, number)
                    new = id_pattern(*before[0]) if same else id_pattern(before[0][0], before[0][1])
                old.setdefault((prefix, number, letter), {
                    "label": id_label(prefix, number, letter), "where": f"{name}:{n}", "how": "replaced",
                    "pattern": id_pattern(prefix, number, letter), "new": new,
                    "definition": re.compile(rf"^\s*(?:[-*]\s+)?\**{prefix}{number}(?![0-9])"),
                })
        if RETIRED.search(line):
            match = AC_DEFINITION.match(line)
            if match:
                number = match.group(1)
                old.setdefault(("AC", number, ""), {
                    "label": f"AC{number}", "where": f"{name}:{n}", "how": "retired",
                    "pattern": id_pattern("AC", number), "new": None,
                    "definition": re.compile(rf"^\s*(?:[-*]\s+)?\**AC{number}(?![0-9])"),
                })
    return old


def stale_mentions(lines, old, name):
    findings = []
    progress = exempt_progress(lines)
    first_task = task_start(lines)
    for i, line in enumerate(lines):
        if i in progress or REPLACES.search(line) or RETIRED.search(line):
            continue
        for item in old.values():
            if not item["pattern"].search(line):
                continue
            if i < first_task and item["definition"].match(line):
                continue  # the line that defines the old ID keeps it
            if item["new"] is not None and item["new"].search(line):
                continue  # the line names the replacement too
            findings.append(f"{name}:{i + 1}: {item['label']} was {item['how']} ({item['where']}) but is still named here")
    return findings


def criteria_findings(lines, old, name):
    progress = exempt_progress(lines)
    first_task = task_start(lines)
    defined = {}
    start, end = acceptance_range(lines)
    for i in range(start, end):
        line = lines[i]
        match = AC_DEFINITION.match(line)
        if match and i not in progress:
            defined.setdefault(int(match.group(1)), i + 1)
    closes = [(n, m.group(1)) for n, line in enumerate(lines, 1)
              if n > first_task and (m := CLOSES.search(line))]
    if not defined:
        named = [(n, AC_REFERENCE.search(text).group(0)) for n, text in closes if AC_REFERENCE.search(text)]
        if named:
            n, ac = named[0]
            return [f"{name}:{n}: {ac} in **Closes:**, but the plan writes no criterion as an `AC<n>` line"]
        return []
    findings, closed = [], set()
    for n, text in closes:
        for start, end in AC_REFERENCE.findall(text):
            first, last = sorted((int(start), int(end or start)))
            for named in sorted({first, last}):
                if named not in defined:
                    findings.append(f"{name}:{n}: AC{named} in **Closes:** is not defined in the plan")
            closed.update(range(first, min(last, first + 999) + 1))
    gone = {int(number) for prefix, number, _ in old if prefix == "AC"}
    for number, n in sorted(defined.items()):
        if number not in closed and number not in gone:
            findings.append(f"{name}:{n}: AC{number} is named by no **Closes:** line")
    return findings


def criterion_lines(plan, start):
    """Return a criterion's line with the lines indented under it, up to a nested criterion."""
    indent = len(plan[start]) - len(plan[start].lstrip())
    lines = [plan[start]]
    for following in plan[start + 1:]:
        if not following.strip():
            continue
        if len(following) - len(following.lstrip()) <= indent or AC_DEFINITION.match(following):
            break
        lines.append(following)
    return lines


def names_place(lines):
    """True when text follows the criterion's last marker, on its line or wrapped below it, and is not a placeholder."""
    text = " ".join(part.strip() for part in lines)
    found = list(OBSERVED.finditer(text))
    place = text[found[-1].end():].strip(" *`.·—–-") if found else ""
    return bool(place) and not PLACEHOLDER.match(place)


def observed_findings(plan, old, name):
    """Report each active criterion none of whose lines names a place where it is observed."""
    progress = exempt_progress(plan)
    gone = {int(number) for prefix, number, _ in old if prefix == "AC"}
    named, last = {}, {}
    start, end = acceptance_range(plan)
    for i in range(start, end):
        line = plan[i]
        match = AC_DEFINITION.match(line)
        if not match or i in progress or int(match.group(1)) in gone:
            continue
        number = int(match.group(1))
        last[number] = i + 1
        named[number] = named.get(number, False) or names_place(criterion_lines(plan, i))
    return [
        f"{name}:{last[number]}: AC{number} names no **Observed in:** place"
        for number in sorted(named) if not named[number]
    ]


def delivery_units(lines):
    """Return [(line number, text)] of the lines inside `## Delivery units`."""
    inside, found = False, []
    for n, line in enumerate(lines, 1):
        if LEVEL2_HEADING.match(line):
            inside = bool(UNITS_HEADING.match(line))
        elif inside:
            found.append((n, line))
    return found


def box_is_not_applicable(text):
    """A `live:` or `perf:` box whose whole predicate is `n/a: <reason>`."""
    kind, _, rest = text.partition(":")
    return kind.strip() in ("live", "perf") and bool(NOT_APPLICABLE.match(rest.strip()))


def delivery_unit_findings(plan, name):
    findings = []
    for n, line in delivery_units(plan):
        if UNIT_HEADING.match(line):
            if not UNIT_TITLE.match(line.strip()):
                findings.append(
                    f"{name}:{n}: invalid delivery unit heading, expected "
                    f"`### E<n> — <unit> · depends on: <…> · files: <…>`: {line.strip()}"
                )
            continue
        box = BOX_LINE.match(line)
        if not box:
            continue
        text = box.group(1).strip()
        if box_is_not_applicable(text):
            continue
        if text.startswith("unit:"):
            ok, expected = UNIT_BOX.match(text), "unit: <command>"
        elif text.startswith("live:"):
            ok, expected = LIVE_BOX.match(text), "live: <scenario> · pass: <predicate>, or live: n/a: <reason>"
        elif text.startswith("perf:"):
            ok = PERF_BOX.match(text)
            expected = "perf: <metric> · base: <value> · fail: <threshold>, or perf: n/a: <reason>"
        else:
            findings.append(f"{name}:{n}: delivery unit box must start with unit:, live: or perf:: {text}")
            continue
        if not ok:
            findings.append(f"{name}:{n}: invalid delivery unit box, expected `{expected}`: {text}")
    return findings


def valid_business_rule_line(trimmed):
    if BR_NONE.match(trimmed) or BR_TICKET_SECTION.match(trimmed) or BR_TICKET_WRITE.match(trimmed):
        return True, ""
    if BR_CARD_STALE.match(trimmed) or BR_FRAME_STALE.match(trimmed):
        return True, ""
    m = BR_RULE.match(trimmed)
    if not m:
        return False, "invalid line in ## Business rules"
    status = m.group(3).strip()
    if status.startswith("stated:"):
        return False, "stated: status is only valid in examples.md, not in a plan's ## Business rules"
    if not BR_STATUS.match(status):
        return False, f"invalid rule status in ## Business rules: {status}"
    return True, ""


def business_rules_findings(plan, name):
    findings = []
    inside = False
    for i, line in enumerate(plan):
        if LEVEL2_HEADING.match(line):
            inside = bool(BUSINESS_RULES_HEADING.match(line))
            continue
        if not inside:
            continue
        trimmed = line.strip()
        if not trimmed or trimmed.startswith("<!--"):
            continue
        ok, reason = valid_business_rule_line(trimmed)
        if not ok:
            findings.append(f"{name}:{i + 1}: {reason}: {trimmed}")
    return findings


def execution_findings(plan_path, plan):
    """An optional checkpoint pointer uses the owning validator."""
    pointers = [(n, line[len("- Execution record: "):].strip())
                for n, line in enumerate(plan, 1) if line.startswith("- Execution record: ")]
    if not pointers:
        return []
    name = plan_path.name
    if len(pointers) != 1:
        return [f"{name}:1: multiple execution record pointers"]
    n, pointer = pointers[0]
    spec = importlib.util.spec_from_file_location("hive_tasks", Path(__file__).with_name("tasks.py"))
    module = importlib.util.module_from_spec(spec)
    previous = sys.dont_write_bytecode
    try:
        sys.dont_write_bytecode = True
        spec.loader.exec_module(module)
        record = plan_path.parent / pointer
        module.execution_location(plan_path, record)
        module.validate_execution(module.load_record(record), plan_path.resolve(), record)
    except (module.Failure, OSError, UnicodeError, ValueError, TypeError, KeyError) as error:
        return [f"{name}:{n}: execution record: {error}"]
    finally:
        sys.dont_write_bytecode = previous
    return []


CHECKPOINT_HEADING = re.compile(r"^##\s+Checkpoint\s*$", re.IGNORECASE)
LEAF_NAME = r"[a-z0-9]+(?:-[a-z0-9]+)*"
PACKET_PATH = rf"{LEAF_NAME}/stage(?:-[a-z]+)?\.md"
RULES_LINE = re.compile(rf"^{PACKET_PATH}(?:, {PACKET_PATH})*(?: · extra: {LEAF_NAME}(?:, {LEAF_NAME})*)?$")
PRINCIPLES_LINE = re.compile(rf"^(?:none|(?!none(?:,|$)){LEAF_NAME}(?:, (?!none(?:,|$)){LEAF_NAME})*)$")


def checkpoint_line_findings(plan, name):
    """The Rules and Principles lines of the checkpoint, which a resumed session reads."""
    findings, inside = [], False
    for n, line in enumerate(plan, 1):
        if line.startswith("## "):
            inside = bool(CHECKPOINT_HEADING.match(line))
            continue
        if not inside:
            continue
        for label, shape in (("Rules", RULES_LINE), ("Principles", PRINCIPLES_LINE)):
            prefix = f"- {label}:"
            if line.startswith(prefix) and not shape.match(line[len(prefix):].strip()):
                findings.append(f"{name}:{n}: malformed checkpoint line: {line.strip()}")
    return findings


def main(argv):
    if len(argv) != 1 or argv[0].startswith("--"):
        print(__doc__.strip().splitlines()[2], file=sys.stderr)
        return 2
    plan_path = Path(argv[0])
    if not plan_path.is_file():
        print(f"check_change: {plan_path}: not a plan file (pass the path of <YYYY-MM-DD>-<work>.plan.md, not a folder)", file=sys.stderr)
        return 2
    try:
        plan = read_lines(plan_path)
    except (OSError, UnicodeDecodeError) as error:
        print(f"check_change: {error}", file=sys.stderr)
        return 2

    name = plan_path.name
    old = collect_replacements(plan, name)
    findings = criteria_findings(plan, old, name)
    findings += observed_findings(plan, old, name)
    findings += business_rules_findings(plan, name)
    findings += execution_findings(plan_path, plan)
    findings += delivery_unit_findings(plan, name)
    findings += checkpoint_line_findings(plan, name)
    findings += stale_mentions(plan, old, name)
    for line in findings:
        print(line)
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
