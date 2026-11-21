#!/usr/bin/env python3
"""Run a check with its whole output in a capped log, then show the log's end.

Usage: run_check.py [--log PATH] [--max-bytes N] [--tail N | --grep REGEX] -- COMMAND [ARGS...]

The command runs without a shell, in its own process group, with stdin from
/dev/null and stdout and stderr in the log. The log is closed before its end
is shown, so nothing this script prints can reach the log. Exit status: the
command's own; 128+n when signal n ended it; 127 when it cannot start; 125
when the cap stopped it; 2 for a usage error.
"""

import argparse
import os
import re
import select
import signal
import subprocess
import sys
import tempfile
import time

DEFAULT_MAX_BYTES = 64 * 1024 * 1024
SHOWN_BYTES = 16 * 1024
GRACE_SECONDS = 5.0
CHUNK = 64 * 1024


def parse(argv):
    if "--" not in argv:
        print("run_check: put the command after --", file=sys.stderr)
        sys.exit(2)
    split = argv.index("--")
    command = argv[split + 1:]
    parser = argparse.ArgumentParser(prog="run_check.py", description=__doc__.splitlines()[0])
    parser.add_argument("--log", help="log path to create; refuses an existing file or link")
    parser.add_argument("--max-bytes", type=int, default=DEFAULT_MAX_BYTES, help="stop the command past this many bytes of output")
    shown = parser.add_mutually_exclusive_group()
    shown.add_argument("--tail", type=int, default=40, help="lines of the log's end to show")
    shown.add_argument("--grep", help="show the log's lines that match this regular expression instead")
    args = parser.parse_args(argv[:split])
    if not command:
        parser.error("no command after --")
    if args.max_bytes <= 0 or args.tail <= 0:
        parser.error("--max-bytes and --tail must be positive")
    if args.grep is not None:
        try:
            args.grep = re.compile(args.grep)
        except re.error as err:
            parser.error(f"--grep: {err}")
    return args, command


def open_log(path):
    if path:
        flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL | getattr(os, "O_NOFOLLOW", 0)
        return os.open(path, flags, 0o600), path
    return tempfile.mkstemp(prefix="hive-check-", suffix=".log")


def kill_group(pgid, sig):
    try:
        os.killpg(pgid, sig)
    except (ProcessLookupError, PermissionError):
        pass


def guard_launcher(alive_fd):
    """Return a preexec hook that forks a guard into the command's new session.

    A helper killed by SIGKILL never runs its cleanup, and the host's group
    kill no longer reaches a command in its own session. The guard blocks on a
    pipe whose only writer is the helper; the kernel closes it when the helper
    dies in any way, and the guard then kills the command's group. It is forked
    twice so that init adopts it: a command that waits for all its children
    must not wait for the guard. Without a guard the check still runs.
    """

    def launch():
        try:
            middle = os.fork()
        except OSError:
            return
        if middle != 0:
            os.waitpid(middle, 0)
            return
        try:
            if os.fork() != 0:
                return
            for fd in sorted(int(name) for name in os.listdir("/dev/fd")):
                if fd > 2 and fd != alive_fd:
                    try:
                        os.close(fd)
                    except OSError:
                        pass
            null = os.open(os.devnull, os.O_RDWR)
            for fd in (0, 1, 2):
                os.dup2(null, fd)
            while os.read(alive_fd, 1):
                pass
            os.killpg(0, signal.SIGKILL)
        finally:
            os._exit(0)

    return launch


def end_of_log(path, lines):
    size = os.path.getsize(path)
    with open(path, "rb") as f:
        f.seek(max(0, size - SHOWN_BYTES))
        data = f.read()
    if size > len(data) and b"\n" in data[:-1]:
        data = data[data.index(b"\n") + 1:]
    return b"".join(data.splitlines(keepends=True)[-lines:])


def matching_lines(path, regex):
    out, total = [], 0
    with open(path, "rb") as f:
        for raw in f:
            if regex.search(raw.decode("utf-8", errors="replace")):
                piece = raw[:SHOWN_BYTES - total]
                out.append(piece)
                total += len(piece)
                if len(out) >= 40 or total >= SHOWN_BYTES:
                    break
    return b"".join(out)


def show(data):
    if data and not data.endswith(b"\n"):
        data += b"\n"
    sys.stdout.buffer.write(data)
    sys.stdout.flush()


def run(args, command):
    try:
        log_fd, log_path = open_log(args.log)
    except OSError as err:
        print(f"run_check: cannot create the log {args.log}: {err.strerror}", file=sys.stderr)
        return 2
    print(f"log: {log_path}", flush=True)

    received = []

    def forward(signum, _frame):
        received.append(signum)
        if proc is not None:
            kill_group(proc.pid, signum)

    proc = None
    for sig in (signal.SIGINT, signal.SIGTERM, signal.SIGHUP):
        signal.signal(sig, forward)

    alive_read, alive_write = os.pipe()
    try:
        try:
            proc = subprocess.Popen(
                command,
                stdin=subprocess.DEVNULL,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                start_new_session=True,
                preexec_fn=guard_launcher(alive_read),
            )
        except (OSError, subprocess.SubprocessError) as err:
            os.close(log_fd)
            reason = getattr(err, "strerror", None) or str(err)
            print(f"run_check: cannot run {command[0]}: {reason}", file=sys.stderr)
            return 127
        finally:
            os.close(alive_read)
        if received:
            kill_group(proc.pid, received[0])

        pipe = proc.stdout.fileno()
        written, capped, term_at, exited_at = 0, False, None, None
        while True:
            ready, _, _ = select.select([pipe], [], [], 0.2)
            if ready:
                data = os.read(pipe, CHUNK)
                if not data:
                    break
                if not capped:
                    room = args.max_bytes - written
                    if len(data) > room:
                        data = data[:room]
                        capped, term_at = True, time.monotonic()
                        kill_group(proc.pid, signal.SIGTERM)
                    os.write(log_fd, data)
                    written += len(data)
            now = time.monotonic()
            if capped and now - term_at > GRACE_SECONDS:
                kill_group(proc.pid, signal.SIGKILL)
            if proc.poll() is not None:
                if exited_at is None:
                    exited_at = now
                elif now - exited_at > GRACE_SECONDS:
                    break
        proc.wait()
    finally:
        if proc is not None:
            kill_group(proc.pid, signal.SIGKILL)
            proc.stdout.close()
        for fd in (log_fd, alive_write):
            try:
                os.close(fd)
            except OSError:
                pass

    if args.grep is not None:
        show(matching_lines(log_path, args.grep))
    else:
        show(end_of_log(log_path, args.tail))

    if capped:
        print(f"run_check: stopped at {args.max_bytes} bytes; log: {log_path}", file=sys.stderr)
        return 125
    if received:
        return 128 + received[0]
    if proc.returncode < 0:
        return 128 - proc.returncode
    return proc.returncode


def main(argv=None):
    args, command = parse(sys.argv[1:] if argv is None else argv)
    return run(args, command)


if __name__ == "__main__":
    sys.exit(main())
