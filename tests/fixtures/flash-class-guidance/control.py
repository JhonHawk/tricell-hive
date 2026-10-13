#!/usr/bin/env python3
"""Coordinate the existing runner; never retry roots or substitute models."""
import argparse
import concurrent.futures
import json
import os
import pathlib
import signal
import subprocess
import threading


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--runner", required=True, help="Previously built pilot executable")
    parser.add_argument("--source", required=True, help="Candidate checkout")
    parser.add_argument("--baseline", required=True, help="d5164f71 checkout")
    parser.add_argument("--out", required=True, help="New batch directory")
    parser.add_argument("--execute", action="store_true", help="Launch the authorized schedule; omitted prints it only")
    args = parser.parse_args()
    schedule = json.loads(subprocess.check_output([args.runner, "--print-flash-schedule"], text=True))
    if len(schedule) != 48:
        raise SystemExit("Unexpected schedule size; no roots launched")
    if not args.execute:
        print(json.dumps(schedule, indent=2))
        return
    output = pathlib.Path(args.out).resolve()
    output.mkdir(mode=0o700, parents=True, exist_ok=False)
    (output / "schedule.json").write_text(json.dumps(schedule, indent=2))
    stopped = threading.Event()
    processes = set()
    lock = threading.Lock()

    def stop(*_):
        stopped.set()
        with lock:
            for process in processes:
                try:
                    os.killpg(process.pid, signal.SIGTERM)
                except ProcessLookupError:
                    pass

    signal.signal(signal.SIGINT, stop)
    signal.signal(signal.SIGTERM, stop)

    def run(index):
        if stopped.is_set():
            return False
        entry = schedule[index]
        destination = output / f"run-{index + 1:02d}"
        command = [args.runner, "--suite", "flash-class-guidance", "--host", "opencode",
                   "--case", entry["Case"], "--arm", entry["Arm"], "--model", entry["Model"],
                   "--source", args.source, "--guidance-source", args.baseline if entry["Arm"] == "A" else args.source,
                   "--timeout", "3m", "--out", str(destination)]
        with lock:
            if stopped.is_set():
                return False
            process = subprocess.Popen(command, start_new_session=True,
                                       stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            processes.add(process)
            with (output / "launch-order.jsonl").open("a") as journal:
                journal.write(json.dumps({"index": index + 1, **entry}) + "\n")
        try:
            try:
                code = process.wait(timeout=175)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGTERM)
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                process.wait()
                stopped.set()
                return False
            if code != 0 or not (destination / "run.json").is_file():
                stopped.set()
                return False
            result = json.loads((destination / "run.json").read_text())
            flash = result.get("FlashClass") or {}
            if result.get("Terminal") != "completed" or result.get("OutsideChanges") or flash.get("OperationSafety") != "pass":
                stopped.set()
                return False
            return True
        finally:
            with lock:
                processes.discard(process)

    # Submit in pairs rather than queueing 48 roots; a failure forbids the next
    # pair while the two already-started runners finish their own cleanup.
    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
        for start in range(0, len(schedule), 2):
            futures = [pool.submit(run, index) for index in range(start, min(start + 2, len(schedule)))]
            if not all(future.result() for future in futures):
                break
    if stopped.is_set():
        raise SystemExit("Batch stopped; preserve failed runs, do not retry")


if __name__ == "__main__":
    main()
