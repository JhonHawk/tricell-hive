#!/usr/bin/env python3
"""Record task progress and the build checkpoint in a change folder's tasks.md.

Usage:
  python3 -I tasks.py <change-folder> status
  python3 -I tasks.py <change-folder> done T<n> --commit <sha> --unverified "<text>|none"
  python3 -I tasks.py <change-folder> checkpoint --stage <stage> --step <step> --commit <sha>
          --next "<action>" [--finding "<ID: decision>"]... [--check "<command> · <result> · <commit>"]...

A task is a `### T<n> — <title>` heading; its boxes are the `- [ ]` and `- [x]`
lines up to the next `##` or `###` heading.

- status: one line per task (`T1 [x] 2/2 · title`), then `<done>/<total> tasks`.
- done: ticks only that task's boxes, appends `<date> — T<n> · <sha> · unverified: <text>`
  to the `## ... progress` section unless a line for that task and commit is there,
  and prints `✓ T<n> · <tasks done>/<tasks> · <title>`.
- checkpoint: writes or replaces `## Checkpoint`, just before the progress section,
  with the fixed lines Updated, Stage (stage and step), Last commit, Open findings and
  Next, followed by `### Evidence`. Existing evidence lines are kept; each --check adds one.

Writes tasks.md directly. Exit status: 0 on success, 2 when the arguments are invalid,
the folder or tasks.md cannot be read or written, the task does not exist, or the
progress section is missing for `done`. Standard library only; self-contained.
"""

import argparse
import datetime
import os
import re
import sys
import tempfile
from pathlib import Path

TASK_HEADING = re.compile(r"^###\s+(T\d+)\s*(?:[—–:-]\s*)?(.*?)\s*$")
BOX = re.compile(r"^(\s*[-*]\s+\[)([ xX])(\]\s*.*)$")
HEADING = re.compile(r"^#{2,3}\s")
LEVEL2 = re.compile(r"^##\s")
PROGRESS = re.compile(r"^##\s.*(progress|progreso|avance)", re.IGNORECASE)
CHECKPOINT = re.compile(r"^##\s+Checkpoint\s*$", re.IGNORECASE)
EVIDENCE = re.compile(r"^###\s+Evidence\s*$", re.IGNORECASE)
SHA = re.compile(r"^[0-9a-fA-F]{7,40}$")
TASK_ID = re.compile(r"^T\d+$")


class Failure(Exception):
    pass


def single_line(value):
    if "\n" in value or "\r" in value:
        raise argparse.ArgumentTypeError("must be a single line")
    return value.strip()


def sha(value):
    if not SHA.match(value):
        raise argparse.ArgumentTypeError("must be a 7 to 40 character hex commit id")
    return value


def task_id(value):
    if not TASK_ID.match(value):
        raise argparse.ArgumentTypeError("must look like T2")
    return value


def read(path):
    try:
        return path.read_text(encoding="utf-8").splitlines()
    except (OSError, UnicodeDecodeError) as error:
        raise Failure(f"{path}: {error}")


def write(path, lines):
    text = "\n".join(lines) + "\n"
    temp = None
    try:
        fd, temp = tempfile.mkstemp(dir=str(path.parent), prefix=".tasks-", suffix=".tmp")
        with os.fdopen(fd, "w", encoding="utf-8", newline="") as handle:
            handle.write(text)
        os.chmod(temp, path.stat().st_mode & 0o7777)
        os.replace(temp, path)
        temp = None
    except OSError as error:
        raise Failure(f"{path}: {error}")
    finally:
        if temp and os.path.exists(temp):
            os.unlink(temp)


def parse_tasks(lines):
    """Return [{id, title, start, end, boxes}] where boxes are line indexes."""
    tasks = []
    for i, line in enumerate(lines):
        match = TASK_HEADING.match(line)
        if match:
            tasks.append({"id": match.group(1), "title": match.group(2), "start": i})
    for task in tasks:
        end = len(lines)
        for j in range(task["start"] + 1, len(lines)):
            if HEADING.match(lines[j]):
                end = j
                break
        task["end"] = end
        task["boxes"] = [j for j in range(task["start"] + 1, end) if BOX.match(lines[j])]
    return tasks


def ticked(lines, task):
    return sum(1 for j in task["boxes"] if BOX.match(lines[j]).group(2) != " ")


def complete(lines, task):
    return bool(task["boxes"]) and ticked(lines, task) == len(task["boxes"])


def section_bounds(lines, heading):
    """Return (start, end) of the level-2 section whose heading matches, or None."""
    for i, line in enumerate(lines):
        if heading.match(line):
            end = len(lines)
            for j in range(i + 1, len(lines)):
                if LEVEL2.match(lines[j]):
                    end = j
                    break
            return i, end
    return None


def today():
    return datetime.date.today().isoformat()


def cmd_status(lines, args):
    tasks = parse_tasks(lines)
    for task in tasks:
        mark = "x" if complete(lines, task) else " "
        print(f"{task['id']} [{mark}] {ticked(lines, task)}/{len(task['boxes'])} · {task['title']}")
    print(f"{sum(1 for t in tasks if complete(lines, t))}/{len(tasks)} tasks")
    return None


def cmd_done(lines, args):
    tasks = parse_tasks(lines)
    task = next((t for t in tasks if t["id"] == args.task), None)
    if task is None:
        raise Failure(f"{args.task}: no such task in tasks.md")
    if not task["boxes"]:
        raise Failure(f"{args.task}: the task has no boxes to tick")
    bounds = section_bounds(lines, PROGRESS)
    if bounds is None:
        raise Failure("tasks.md has no progress section (a `##` heading with progress, progreso or avance)")
    for j in task["boxes"]:
        match = BOX.match(lines[j])
        lines[j] = f"{match.group(1)}x{match.group(3)}"
    start, end = bounds
    known = re.compile(rf"(?<![A-Za-z0-9]){args.task}\s*·\s*{re.escape(args.commit)}(?![0-9A-Za-z])")
    if not any(known.search(line) for line in lines[start + 1:end]):
        last = end
        while last > start + 1 and not lines[last - 1].strip():
            last -= 1
        lines.insert(last, f"{today()} — {args.task} · {args.commit} · unverified: {args.unverified}")
    done = sum(1 for t in parse_tasks(lines) if complete(lines, t))
    print(f"✓ {task['id']} · {done}/{len(tasks)} · {task['title']}")
    return lines


def cmd_checkpoint(lines, args):
    kept = []
    existing = section_bounds(lines, CHECKPOINT)
    if existing:
        start, end = existing
        inside = False
        for line in lines[start + 1:end]:
            if EVIDENCE.match(line):
                inside = True
            elif HEADING.match(line):
                inside = False
            elif inside and line.strip():
                kept.append(line)
    for check in args.check:
        entry = f"- {check}"
        if entry not in kept:
            kept.append(entry)
    findings = "; ".join(args.finding) if args.finding else "none"
    block = [
        "## Checkpoint",
        "",
        f"- Updated: {today()}",
        f"- Stage: {args.stage} · {args.step}",
        f"- Last commit: {args.commit}",
        f"- Open findings: {findings}",
        f"- Next: {args.next}",
        "",
        "### Evidence",
        "",
        *kept,
    ]
    if existing:
        start, end = existing
        lines[start:end] = block + ([""] if end < len(lines) else [])
    else:
        progress = section_bounds(lines, PROGRESS)
        if progress:
            lines[progress[0]:progress[0]] = block + [""]
        else:
            while lines and not lines[-1].strip():
                lines.pop()
            lines.extend([""] + block)
    print(f"checkpoint written · {args.stage} · {args.step} · {len(kept)} evidence line(s)")
    return lines


def build_parser():
    parser = argparse.ArgumentParser(prog="tasks.py", description="Record task progress in a change folder's tasks.md.")
    parser.add_argument("folder", type=Path)
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("status")
    done = sub.add_parser("done")
    done.add_argument("task", type=task_id)
    done.add_argument("--commit", required=True, type=sha)
    done.add_argument("--unverified", required=True, type=single_line)
    cp = sub.add_parser("checkpoint")
    cp.add_argument("--stage", required=True, type=single_line)
    cp.add_argument("--step", required=True, type=single_line)
    cp.add_argument("--commit", required=True, type=sha)
    cp.add_argument("--next", required=True, type=single_line)
    cp.add_argument("--finding", action="append", default=[], type=single_line)
    cp.add_argument("--check", action="append", default=[], type=single_line)
    return parser


def main(argv):
    args = build_parser().parse_args(argv)
    path = args.folder / "tasks.md"
    try:
        if not path.is_file():
            raise Failure(f"{path}: not found")
        lines = read(path)
        handler = {"status": cmd_status, "done": cmd_done, "checkpoint": cmd_checkpoint}[args.command]
        result = handler(lines, args)
        if result is not None:
            write(path, result)
    except Failure as error:
        print(f"tasks: {error}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
