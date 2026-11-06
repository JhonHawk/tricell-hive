#!/usr/bin/env python3
"""Merge a change folder's spec deltas into the specs and archive the folder.

Usage: python3 apply_deltas.py <change-folder> --specs <Specs> [--apply]

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
Headings inside fenced code are text. Refused, with nothing written: a MODIFIED
requirement the spec lacks, an ADDED one the spec holds with other text, a
requirement listed twice in one delta, a MODIFIED in a new capability, a delta
without operations, a spec without `## Requirements`, an existing archive
target.

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
FENCE = re.compile(r"^\s*(```|~~~)")
STATUS = re.compile(r"^(\s*\|\s*(?:\*\*)?(?:Status|Estado)(?:\*\*)?\s*\|)([^|\n]*)(\|.*)$", re.M)
STATUS_FIELD = re.compile(r"^(\s*(?:\*\*)?(?:Status|Estado)(?:\*\*)?\s*:(?:\*\*)?)(.*)$", re.M)


class Refused(Exception):
    pass


def unfenced(lines):
    """Yield (index, line, inside_fence) so headings in code stay text."""
    fence = None
    for i, line in enumerate(lines):
        match = FENCE.match(line)
        inside = fence is not None
        if match:
            fence = None if fence == match.group(1) else (fence or match.group(1))
            inside = True
        yield i, line, inside


def normalize(text):
    return " ".join(text.split())


def scenarios(block):
    return [m.group(1) for _, line, inside in unfenced(block.split("\n")) if not inside and (m := SCENARIO.match(line))]


def parse_delta(text):
    """Return {op: [(name, block)]} and the delta's `## Purpose` body."""
    ops, purpose, op, current = {o: [] for o in OPERATIONS}, [], None, None
    in_purpose = False
    for _, line, inside in unfenced(text.split("\n")):
        if not inside and line.startswith("## "):
            title = line[3:].strip()
            op = next((o for o in OPERATIONS if title == f"{o} Requirements"), None)
            in_purpose, current = title == "Purpose", None
            continue
        if not inside and line.startswith("### "):
            match = REQUIREMENT.match(line)
            current = [match.group(1), [line]] if match and op else None
            if current:
                ops[op].append(current)
            continue
        if current is not None:
            current[1].append(line)
        elif in_purpose:
            purpose.append(line)
    blocks = {o: [(name, "\n".join(lines).rstrip()) for name, lines in items] for o, items in ops.items()}
    return blocks, "\n".join(purpose).strip()


def parse_spec(text):
    """Split a spec into (head, [(name, block)], tail) around `## Requirements`."""
    lines = text.split("\n")
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
    ops, purpose = parse_delta(delta_text)
    if not any(ops.values()):
        raise Refused(f"{capability}: delta has no operations")
    seen = {}
    for op in OPERATIONS:
        for name, _ in ops[op]:
            if name in seen:
                raise Refused(f'{capability}: "{name}" is listed in {seen[name]} and {op}')
            seen[name] = op
    if spec_text is None:
        if ops["MODIFIED"]:
            raise Refused(f"{capability}: new capability has MODIFIED requirements; only ADDED can create it")
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


def mark_completed(text):
    for pattern, build in ((STATUS, lambda m: f"{m.group(1)} Completed {m.group(3)}"), (STATUS_FIELD, lambda m: f"{m.group(1)} Completed")):
        match = pattern.search(text)
        if match:
            return text[: match.start()] + build(match) + text[match.end():]
    return text


def main(argv):
    apply = "--apply" in argv
    args = [a for a in argv if a != "--apply"]
    if len(args) != 3 or args[1] != "--specs":
        print(__doc__.strip().splitlines()[2], file=sys.stderr)
        return 2
    folder, specs = Path(args[0]), Path(args[2])
    proposal = folder / "proposal.md"
    if not proposal.is_file() or not specs.is_dir():
        print(f"apply_deltas: {proposal if not proposal.is_file() else specs}: not found", file=sys.stderr)
        return 2
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
    except Refused as refused:
        print(f"refused: {refused}")
        return 1
    except (OSError, UnicodeDecodeError) as error:
        print(f"apply_deltas: {error}", file=sys.stderr)
        return 2
    for line in report:
        print(line)
    print(f"archive: {target}")
    if not apply:
        return 0
    proposal_old = proposal.read_text(encoding="utf-8")
    done = []
    try:
        for path, (_, new) in writes.items():
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(new, encoding="utf-8")
            done.append(path)
        proposal.write_text(mark_completed(proposal_old), encoding="utf-8")
        target.parent.mkdir(parents=True, exist_ok=True)
        os.rename(folder, target)
    except OSError as error:
        for path in done:
            old = writes[path][0]
            path.unlink() if old is None else path.write_text(old, encoding="utf-8")
        if proposal.exists():
            proposal.write_text(proposal_old, encoding="utf-8")
        print(f"apply_deltas: {error}; restored the specs", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
