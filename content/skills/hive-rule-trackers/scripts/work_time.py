#!/usr/bin/env python3
"""Time marks for a host whose transcript `hive time` cannot read.

`mark <file> <event>` appends one mark: start (first message on the ticket),
wait (a stop that waits for the user), resume (the user's reply), ci-start and
ci-end (around a remote CI or deployment wait). `sum <file>` prints the same
fields as `hive time`: active time is the span from the first start to the last
mark minus the user waits; converted_seconds is (active - ci) x factor + ci,
rounded up to the next whole minute.

Exit codes: 0 success, 2 bad usage or an unreadable marks file.
"""

from __future__ import annotations

import argparse
import json
import math
import sys
from datetime import datetime, timezone
from pathlib import Path

EVENTS = ("start", "wait", "resume", "ci-start", "ci-end")


def parse_time(value: str) -> datetime:
    moment = datetime.fromisoformat(value.replace("Z", "+00:00"))
    return moment if moment.tzinfo else moment.replace(tzinfo=timezone.utc)


def factor_arg(value: str) -> float:
    factor = float(value)
    if not 1 <= factor <= 3:
        raise argparse.ArgumentTypeError("the factor must be between 1 and 3")
    return factor


def mark(path: Path, event: str, at: str | None) -> dict:
    moment = parse_time(at) if at else datetime.now(timezone.utc)
    row = {"event": event, "at": moment.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")}
    with path.open("a", encoding="utf-8") as fh:
        fh.write(json.dumps(row) + "\n")
    return row


def load(path: Path) -> list[tuple[str, datetime]]:
    marks = []
    for n, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if not line.strip():
            continue
        try:
            row = json.loads(line)
            event, moment = row["event"], parse_time(row["at"])
        except (ValueError, KeyError, TypeError):
            raise SystemExit(f"{path}:{n}: not a mark") from None
        if event not in EVENTS:
            raise SystemExit(f"{path}:{n}: unknown event {event!r}")
        marks.append((event, moment))
    return sorted(marks, key=lambda m: m[1])


def span_length(spans, first, last) -> float:
    """Seconds covered by the spans, clipped to [first, last], overlaps merged."""
    clipped = sorted((max(a, first), min(b, last)) for a, b in spans if min(b, last) > max(a, first))
    total, end = 0.0, None
    for a, b in clipped:
        if end is not None and a < end:
            if b > end:
                total += (b - end).total_seconds()
                end = b
            continue
        total += (b - a).total_seconds()
        end = b
    return total


def total(path: Path, factor: float, since: str | None) -> dict:
    marks = load(path)
    if since:
        floor = parse_time(since)
        marks = [m for m in marks if m[1] >= floor]
    starts = [moment for event, moment in marks if event == "start"]
    if not starts:
        raise SystemExit(f"{path}: no start mark in the window")
    first, last = starts[0], marks[-1][1]
    waits, ci = [], []
    wait_at = ci_at = None
    for event, moment in marks:
        if moment < first:
            continue
        if event == "wait":
            wait_at = wait_at or moment
        elif event == "resume" and wait_at:
            waits.append((wait_at, moment))
            wait_at = None
        elif event == "ci-start":
            ci_at = ci_at or moment
        elif event == "ci-end" and ci_at:
            ci.append((ci_at, moment))
            ci_at = None
    if wait_at:  # a wait still open at the last mark ends the window there
        last = wait_at
    # As in hive time: CI time is the CI spans' union with the waits, minus the waits.
    user = span_length(waits, first, last)
    ci_seconds = round(span_length(waits + ci, first, last) - user)
    active = max(0, round((last - first).total_seconds() - user))
    converted = math.ceil(((active - ci_seconds) * factor + ci_seconds) / 60 - 1e-9) * 60
    return {
        "result": "ok",
        "since": first.isoformat().replace("+00:00", "Z"),
        "until": last.isoformat().replace("+00:00", "Z"),
        "active_seconds": active,
        "user_wait_seconds": round(user),
        "ci_wait_seconds": ci_seconds,
        "factor": factor,
        "converted_seconds": converted,
    }


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    sub = parser.add_subparsers(dest="command", required=True)
    p = sub.add_parser("mark", help="append one mark")
    p.add_argument("file", type=Path)
    p.add_argument("event", choices=EVENTS)
    p.add_argument("--at", help="ISO 8601 time; defaults to now")
    p = sub.add_parser("sum", help="print the active and converted time")
    p.add_argument("file", type=Path)
    p.add_argument("--factor", type=factor_arg, default=2.0)
    p.add_argument("--since", help="ISO 8601 start of this ticket's window")
    args = parser.parse_args(argv)
    try:
        if args.command == "mark":
            out = mark(args.file, args.event, args.at)
        else:
            out = total(args.file, args.factor, args.since)
    except SystemExit as err:
        print(err, file=sys.stderr)
        return 2
    except (OSError, ValueError) as err:
        print(f"{args.file}: {err}", file=sys.stderr)
        return 2
    print(json.dumps(out))
    return 0


if __name__ == "__main__":
    sys.exit(main())
