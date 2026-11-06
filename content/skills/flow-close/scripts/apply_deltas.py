#!/usr/bin/env python3
"""Merge a change folder's spec deltas into the specs and archive the folder.

Usage: python3 apply_deltas.py <change-folder> --specs <Specs> [--detail TEXT] [--apply]

Without --apply it prints what it would do and changes nothing. With --apply it
merges every delta, marks proposal.md completed and moves the folder to
<Specs>/changes/archive/YYYY-MM-DD-HHMM-<change-id>/ with the UTC time, or
changes nothing when anything is refused.

Each <change-folder>/specs/<capability>/spec.md delta merges into
<Specs>/specs/<capability>/spec.md in this order: REMOVED deletes the
requirement (an absent one is a no-op), MODIFIED replaces the whole block in
place and lists each scenario it drops, ADDED appends at the end of
`## Requirements` (one identical to the current block is a no-op). A new
capability gets `# <capability>`, `## Purpose` and `## Requirements`.
Headings inside fenced code are text; another `##` section of a delta is
listed as not applied. The first Status (or Estado) value of proposal.md
becomes Completed (or Completado), followed by ` · TEXT` when --detail gives
one line of TEXT, such as where the work was integrated; its previous details
are dropped, since they describe an earlier phase. Both runs print the new
value as `status: ...`.
Refused, with nothing written: a MODIFIED requirement the spec lacks, an ADDED
one the spec holds with other text, a requirement listed twice in one delta or
in the spec, a new capability without ADDED or with MODIFIED, a delta without
operations, a `## ... Requirements` or RENAMED section other than the three, a
non-requirement `###` heading among requirements, a spec without
`## Requirements`, a proposal without Status, a folder outside
<Specs>/changes/, an existing archive target.

Exit status: 0 done or previewed, 1 refused, 2 the folder, its proposal.md or
<Specs> cannot be read.
"""

import datetime
import os
import re
import sys
from pathlib import Path

OPERATIONS = ("ADDED", "MODIFIED", "REMOVED")
REQUIREMENT = re.compile(r"^### Requirement:\s*(.+?)\s*$")
SCENARIO = re.compile(r"^#### Scenario:\s*(.+?)\s*$")
FENCE_OPEN = re.compile(r"^ {0,3}(`{3,}(?=[^`]*$)|~{3,})")  # a backtick info string holds no backtick
STATUS = re.compile(r"^(\s*\|\s*(?:\*\*)?(Status|Estado)(?:\*\*)?\s*\|)((?:\\\||[^|\n])*)(\|.*)$", re.M)
TABLE_SEPARATOR = re.compile(r"^\s*\|[\s:|-]+\|\s*$")
STATUS_FIELD = re.compile(r"^(\s*(?:\*\*)?(Status|Estado)(?:\*\*)?\s*:(?:\*\*)?)(.*)()$", re.M)


class Refused(Exception):
    pass


def fence_flags(lines):
    """Return, per line, whether it belongs to a fence, and whether a fence is left open."""
    flags, fence = [], None  # fence: the opening run of backticks or tildes
    for line in lines:
        if fence is None:
            match = FENCE_OPEN.match(line)
            fence = match.group(1) if match else None
            flags.append(fence is not None)
            continue
        stripped = line.strip()
        if stripped and set(stripped) == {fence[0]} and len(stripped) >= len(fence) and len(line) - len(line.lstrip(" ")) <= 3:
            fence = None  # a closing fence: the same character, at least as long, nothing else
        flags.append(True)
    return flags, fence is not None


def unfenced(lines):
    """Yield (index, line, inside_fence) so headings in code stay text."""
    flags, _ = fence_flags(lines)
    for i, line in enumerate(lines):
        yield i, line, flags[i]


def fence_left_open(lines):
    return fence_flags(lines)[1]


def normalize(text):
    return " ".join(text.split())


def scenarios(block):
    return [m.group(1) for _, line, inside in unfenced(block.split("\n")) if not inside and (m := SCENARIO.match(line))]


def parse_delta(text):
    """Return {op: [(name, block)]}, the `## Purpose` body and the other sections."""
    ops, purpose, others, op, current = {o: [] for o in OPERATIONS}, [], [], None, None
    in_purpose = seen_op = False
    lines = text.split("\n")
    if fence_left_open(lines):
        raise Refused("has a code fence that is never closed")
    for _, line, inside in unfenced(lines):
        if not inside and line.startswith("# "):
            if seen_op:
                others.append(f"# {line[2:].strip()}")
            op = current = None
            in_purpose = False
            continue
        if not inside and line.startswith("## "):
            title = line[3:].strip()
            op = next((o for o in OPERATIONS if title == f"{o} Requirements"), None)
            seen_op = seen_op or op is not None
            if op is None and title != "Purpose":
                if "requirement" in title.lower() or title.upper().startswith("RENAMED"):
                    raise Refused(f'section "## {title}" is not ADDED, MODIFIED or REMOVED Requirements')
                others.append(f"## {title}")
            in_purpose, current = title == "Purpose", None
            continue
        if not inside and line.startswith("### "):
            match = REQUIREMENT.match(line)
            if op and not match:
                raise Refused(f'"{line.strip()}" inside {op} Requirements is not a requirement heading')
            current = [match.group(1), [line]] if match and op else None
            if current:
                ops[op].append(current)
            continue
        if current is not None:
            current[1].append(line)
        elif in_purpose:
            purpose.append(line)
    blocks = {o: [(name, "\n".join(lines).rstrip()) for name, lines in items] for o, items in ops.items()}
    return blocks, "\n".join(purpose).strip(), others


def parse_spec(text):
    """Split a spec into (head, [(name, block)], tail) around `## Requirements`."""
    lines = text.split("\n")
    if fence_left_open(lines):
        raise Refused("has a code fence that is never closed")
    start = end = None
    for i, line, inside in unfenced(lines):
        if inside or not line.startswith("## "):
            continue
        if start is None and line.strip() == "## Requirements":
            start = i
        elif start is not None:
            end = i
            break
    if start is None:
        raise Refused("has no `## Requirements` section")
    end = len(lines) if end is None else end
    body, blocks, head_extra = lines[start + 1 : end], [], []
    for _, line, inside in unfenced(body):
        match = None if inside else REQUIREMENT.match(line)
        if not inside and not match and line.startswith("### "):
            raise Refused(f'has "{line.strip()}" among its requirements')
        if match and any(name == match.group(1) for name, _ in blocks):
            raise Refused(f'lists "{match.group(1)}" twice')
        if match:
            blocks.append([match.group(1), [line]])
        elif blocks:
            blocks[-1][1].append(line)
        else:
            head_extra.append(line)
    head = "\n".join(lines[: start + 1] + head_extra).rstrip() + "\n\n"
    tail = "\n".join(lines[end:]).strip()
    return head, [(name, "\n".join(block).rstrip()) for name, block in blocks], tail


def render(head, blocks, tail):
    body = "\n\n".join(block for _, block in blocks)
    return head + body + "\n" + (f"\n{tail}\n" if tail else "")


def merge(capability, delta_text, spec_text, change_id, report):
    """Return the new spec text, or None when nothing changes."""
    try:
        ops, purpose, others = parse_delta(delta_text)
    except Refused as refused:
        raise Refused(f"{capability}: {refused}") from None
    if spec_text is not None and purpose:
        others = ["## Purpose"] + others
    for title in others:
        report.append(f'{capability}: section "{title}" is not applied')
    if not any(ops.values()):
        raise Refused(f"{capability}: delta has no operations")
    seen = {}
    for op in OPERATIONS:
        for name, _ in ops[op]:
            if name in seen:
                raise Refused(f'{capability}: "{name}" is listed in {seen[name]} and {op}')
            seen[name] = op
    if spec_text is None:
        if ops["MODIFIED"] or not ops["ADDED"]:
            raise Refused(f"{capability}: specs/{capability}/spec.md does not exist; only ADDED requirements can create it")
        for name, _ in ops["REMOVED"]:
            report.append(f'{capability}: REMOVED "{name}" is absent (no-op)')
        purpose = purpose or f"TBD: created by archiving {change_id}; update the purpose."
        head, blocks, tail = f"# {capability}\n\n## Purpose\n\n{purpose}\n\n## Requirements\n\n", [], ""
    else:
        try:
            head, blocks, tail = parse_spec(spec_text)
        except Refused as refused:
            raise Refused(f"specs/{capability}/spec.md {refused}") from None
    order = [name for name, _ in blocks]
    current = dict(blocks)
    for name, _ in ops["REMOVED"] if spec_text is not None else []:
        if name in current:
            del current[name]
            report.append(f'{capability}: REMOVED "{name}"')
        else:
            report.append(f'{capability}: REMOVED "{name}" is absent (no-op)')
    for name, block in ops["MODIFIED"]:
        if name not in current:
            raise Refused(f'{capability}: MODIFIED "{name}" not found in specs/{capability}/spec.md')
        for dropped in (s for s in scenarios(current[name]) if s not in scenarios(block)):
            report.append(f'{capability}: drops scenario "{dropped}" from "{name}"')
        current[name] = block
        report.append(f'{capability}: MODIFIED "{name}"')
    for name, block in ops["ADDED"]:
        if name in current:
            if normalize(current[name]) != normalize(block):
                raise Refused(f'{capability}: ADDED "{name}" already exists with other text')
            report.append(f'{capability}: ADDED "{name}" is already present (no-op)')
            continue
        current[name] = block
        order.append(name)
        report.append(f'{capability}: ADDED "{name}"')
    merged = [(name, current[name]) for name in order if name in current]
    text = render(head, merged, tail)
    check = dict(parse_spec(text)[1])
    for op in ("ADDED", "MODIFIED"):
        for name, block in ops[op]:
            if normalize(check.get(name, "")) != normalize(block):
                raise Refused(f'{capability}: "{name}" does not match its delta after the merge')
    if any(name in check for name, _ in ops["REMOVED"]):
        raise Refused(f"{capability}: a REMOVED requirement survived the merge")
    if spec_text is not None and normalize(text) == normalize(spec_text):
        return None
    return text


def mark_completed(text, detail=None):
    """Replace the first Status or Estado value with the completed phase and the given detail.

    Return the new text and the new value.
    """
    lines = text.split("\n")
    def is_header(match):
        row = text.count("\n", 0, match.start())
        return match.re is STATUS and row + 1 < len(lines) and TABLE_SEPARATOR.match(lines[row + 1])
    found = sorted((m for pattern in (STATUS, STATUS_FIELD) for m in pattern.finditer(text) if not is_header(m)), key=lambda m: m.start())
    if found:
        match = found[0]
        value = "Completado" if match.group(2) == "Estado" else "Completed"
        if detail:
            value += " · " + (re.sub(r"(?<!\\)\|", r"\\|", detail) if match.re is STATUS else detail)
        closing = f" {match.group(4)}" if match.group(4) else ""
        return text[: match.start()] + f"{match.group(1)} {value}{closing}" + text[match.end():], value
    raise Refused("proposal.md has no Status or Estado field to mark completed")


def take_detail(args):
    """Remove `--detail TEXT` from args and return TEXT, None when absent, or raise ValueError."""
    if "--detail" not in args:
        return None
    index = args.index("--detail")
    if index + 1 >= len(args) or "--detail" in args[index + 1:]:
        raise ValueError("--detail takes one value, given once")
    detail = args.pop(index + 1).strip()
    args.pop(index)
    if not detail or "\n" in detail or "\r" in detail:
        raise ValueError("--detail takes one non-empty line")
    return detail


def main(argv):
    apply = "--apply" in argv
    args = [a for a in argv if a != "--apply"]
    try:
        detail = take_detail(args)
    except ValueError as error:
        print(f"apply_deltas: {error}", file=sys.stderr)
        return 2
    if len(args) != 3 or args[1] != "--specs":
        print(__doc__.strip().splitlines()[2], file=sys.stderr)
        return 2
    folder, specs = Path(args[0]), Path(args[2])
    proposal = folder / "proposal.md"
    if not proposal.is_file() or not specs.is_dir():
        print(f"apply_deltas: {proposal if not proposal.is_file() else specs}: not found", file=sys.stderr)
        return 2
    if folder.resolve().parent != (specs / "changes").resolve():
        print(f"refused: {folder} is not a change folder under {specs / 'changes'}")
        return 1
    stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%d-%H%M")
    target = specs / "changes" / "archive" / f"{stamp}-{folder.name}"
    report, writes = [], {}
    try:
        for delta in sorted((folder / "specs").glob("*/spec.md")) if (folder / "specs").is_dir() else []:
            capability = delta.parent.name
            main_spec = specs / "specs" / capability / "spec.md"
            old = main_spec.read_text(encoding="utf-8") if main_spec.is_file() else None
            new = merge(capability, delta.read_text(encoding="utf-8"), old, folder.name, report)
            if new is not None:
                writes[main_spec] = (old, new)
        if target.exists():
            raise Refused(f"archive target {target} already exists")
        proposal_old = proposal.read_text(encoding="utf-8")
        proposal_new, status = mark_completed(proposal_old, detail)
    except Refused as refused:
        print(f"refused: {refused}")
        return 1
    except (OSError, UnicodeDecodeError) as error:
        print(f"apply_deltas: {error}", file=sys.stderr)
        return 2
    for line in report:
        print(line)
    print(f"status: {status}")
    print(f"archive: {target}")
    if not apply:
        return 0
    done, created = [], []
    try:
        for path, (_, new) in writes.items():
            if not path.parent.exists():
                created.append(path.parent)
                path.parent.mkdir(parents=True)
            replace(path, new)
            done.append(path)
        replace(proposal, proposal_new)
        done.append(proposal)
        if not target.parent.exists():
            created.append(target.parent)
            target.parent.mkdir(parents=True)
        os.rename(folder, target)
    except OSError as error:
        failed = []
        for path in reversed(done):
            old = proposal_old if path == proposal else writes[path][0]
            try:
                path.unlink() if old is None else replace(path, old)
            except OSError:
                failed.append(str(path))
        for directory in reversed(created):
            try:
                directory.rmdir()
            except OSError:
                pass
        state = f"could not restore {', '.join(failed)}" if failed else "restored every file"
        print(f"apply_deltas: {error}; {state}", file=sys.stderr)
        return 2
    return 0


def replace(path, text):
    """Write through a temporary file so a failed write leaves the old file whole."""
    temporary = path.with_name(f".{path.name}.apply-deltas")
    try:
        temporary.write_text(text, encoding="utf-8")
        os.replace(temporary, path)
    finally:
        if temporary.exists():
            temporary.unlink()


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
