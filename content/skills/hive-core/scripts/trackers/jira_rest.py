#!/usr/bin/env python3
"""Jira writes acli cannot make: transitions, worklogs and attachments.

The API token is read from Hive's global secret file, which only the user
writes: $XDG_CONFIG_HOME/hive/secrets/jira-api-token, or
~/.config/hive/secrets/jira-api-token, in a 0700 folder with mode 0600. The
account email and the site come from `acli jira auth status`. The token never
reaches an argument, the output or an error message.

Exit codes: 0 success, 1 request failed, 2 bad usage, 3 token file missing,
4 token file or folder with unsafe permissions, 5 acli not signed in.
"""

from __future__ import annotations

import argparse
import base64
import json
import mimetypes
import os
import re
import stat
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid
from datetime import datetime, timezone
from pathlib import Path

TOKEN_FILE = "jira-api-token"


def secrets_dir() -> Path:
    xdg = os.environ.get("XDG_CONFIG_HOME", "")
    base = Path(xdg) if xdg and os.path.isabs(xdg) else Path.home() / ".config"
    return base / "hive" / "secrets"


def read_token() -> str:
    folder = secrets_dir()
    path = folder / TOKEN_FILE
    if not path.is_file():
        print(
            f"no Jira token at {path}: ask the user to create it once with "
            f"`mkdir -p {folder} && chmod 700 {folder}`, then save the token in that file "
            f"and run `chmod 600 {path}`; never write it for them",
            file=sys.stderr,
        )
        sys.exit(3)
    for target, mode in ((folder, 0o700), (path, 0o600)):
        st = target.stat()
        if st.st_uid != os.getuid() or stat.S_IMODE(st.st_mode) & ~mode:
            print(f"{target} must belong to you with mode {mode:o} or stricter", file=sys.stderr)
            sys.exit(4)
    value = path.read_text(encoding="utf-8").strip()
    if not value or "\n" in value:
        print(f"{path} must hold the token alone, on one line", file=sys.stderr)
        sys.exit(4)
    return value


def acli_account() -> tuple[str, str]:
    try:
        out = subprocess.run(
            ["acli", "jira", "auth", "status"],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            timeout=30,
        ).stdout
    except (OSError, subprocess.TimeoutExpired):
        out = ""
    site = re.search(r"^\s*Site:\s*(\S+)\s*$", out, re.MULTILINE)
    email = re.search(r"^\s*Email:\s*(\S+)\s*$", out, re.MULTILINE)
    if not site or not email:
        print("acli is not signed in to Jira: run `acli jira auth login`", file=sys.stderr)
        sys.exit(5)
    return site.group(1), email.group(1)


class Jira:
    def __init__(self, base_url: str, email: str, token: str) -> None:
        self.base = base_url.rstrip("/") + "/rest/api/3/"
        self._auth = "Basic " + base64.b64encode(f"{email}:{token}".encode()).decode()

    def request(self, method: str, path: str, payload=None, body: bytes | None = None, headers=None):
        data = body if body is not None else (None if payload is None else json.dumps(payload).encode())
        hdrs = {"Authorization": self._auth, "Accept": "application/json"}
        if payload is not None:
            hdrs["Content-Type"] = "application/json"
        hdrs.update(headers or {})
        req = urllib.request.Request(self.base + path, data=data, method=method, headers=hdrs)
        try:
            with urllib.request.urlopen(req, timeout=60) as resp:
                raw = resp.read()
        except urllib.error.HTTPError as err:
            print(f"jira {method} {path.partition('?')[0]} returned HTTP {err.code}", file=sys.stderr)
            sys.exit(1)
        except urllib.error.URLError as err:
            print(f"jira {method} {path.partition('?')[0]} failed: {err.reason}", file=sys.stderr)
            sys.exit(1)
        return json.loads(raw) if raw.strip() else None


def issue_path(key: str, suffix: str) -> str:
    if not re.fullmatch(r"[A-Z][A-Z0-9_]*-\d+", key):
        print(f"not an issue key: {key}", file=sys.stderr)
        sys.exit(2)
    return f"issue/{urllib.parse.quote(key)}/{suffix}"


def jira_time(moment: datetime) -> str:
    return moment.astimezone(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.000+0000")


def seconds_arg(value: str) -> int:
    n = int(value)
    if n < 60:
        raise argparse.ArgumentTypeError("a worklog needs at least 60 seconds")
    return n


def cmd_transitions(jira: Jira, args) -> object:
    data = jira.request("GET", issue_path(args.key, "transitions"))
    return [
        {
            "id": t.get("id"),
            "name": (t.get("to") or {}).get("name"),
            "category": ((t.get("to") or {}).get("statusCategory") or {}).get("key"),
        }
        for t in (data or {}).get("transitions", [])
    ]


def cmd_worklog_list(jira: Jira, args) -> object:
    me = (jira.request("GET", "myself") or {}).get("accountId")
    data = jira.request("GET", issue_path(args.key, "worklog"))
    return [
        {"id": w.get("id"), "seconds": w.get("timeSpentSeconds"), "started": w.get("started")}
        for w in (data or {}).get("worklogs", [])
        if (w.get("author") or {}).get("accountId") == me
    ]


def cmd_worklog_add(jira: Jira, args) -> object:
    started = datetime.fromisoformat(args.started) if args.started else datetime.now(timezone.utc)
    data = jira.request(
        "POST",
        issue_path(args.key, "worklog"),
        {"timeSpentSeconds": args.seconds, "started": jira_time(started)},
    )
    return {"id": (data or {}).get("id"), "seconds": (data or {}).get("timeSpentSeconds", args.seconds)}


def cmd_worklog_update(jira: Jira, args) -> object:
    if not re.fullmatch(r"\d+", args.worklog_id):
        print(f"not a worklog ID: {args.worklog_id}", file=sys.stderr)
        sys.exit(2)
    data = jira.request("PUT", issue_path(args.key, f"worklog/{args.worklog_id}"), {"timeSpentSeconds": args.seconds})
    return {"id": (data or {}).get("id", args.worklog_id), "seconds": (data or {}).get("timeSpentSeconds", args.seconds)}


def cmd_attach(jira: Jira, args) -> object:
    boundary = uuid.uuid4().hex
    parts = []
    for name in args.files:
        path = Path(name)
        if not path.is_file():
            print(f"not a file: {name}", file=sys.stderr)
            sys.exit(2)
        ctype = mimetypes.guess_type(path.name)[0] or "application/octet-stream"
        parts.append(
            f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{path.name}"\r\n'
            f"Content-Type: {ctype}\r\n\r\n".encode() + path.read_bytes() + b"\r\n"
        )
    body = b"".join(parts) + f"--{boundary}--\r\n".encode()
    data = jira.request(
        "POST",
        issue_path(args.key, "attachments"),
        body=body,
        headers={"Content-Type": f"multipart/form-data; boundary={boundary}", "X-Atlassian-Token": "no-check"},
    )
    return [{"id": a.get("id"), "filename": a.get("filename")} for a in (data or [])]


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--base-url", help="a loopback test server in place of https://<site acli reports>")
    sub = parser.add_subparsers(dest="command", required=True)
    p = sub.add_parser("transitions", help="list the transitions a ticket can take")
    p.add_argument("key")
    p = sub.add_parser("worklog-list", help="list your worklogs on a ticket")
    p.add_argument("key")
    p = sub.add_parser("worklog-add", help="add a worklog with only its time")
    p.add_argument("key")
    p.add_argument("--seconds", type=seconds_arg, required=True)
    p.add_argument("--started", help="ISO 8601 start; defaults to now")
    p = sub.add_parser("worklog-update", help="set the time of one of your worklogs")
    p.add_argument("key")
    p.add_argument("worklog_id")
    p.add_argument("--seconds", type=seconds_arg, required=True)
    p = sub.add_parser("attach", help="upload files to a ticket")
    p.add_argument("key")
    p.add_argument("files", nargs="+")
    args = parser.parse_args(argv)

    auth_value = read_token()
    site, email = acli_account()
    base = f"https://{site}"
    if args.base_url:
        # A test server only: the token never goes to a host acli did not name.
        if urllib.parse.urlsplit(args.base_url).hostname not in ("127.0.0.1", "localhost", "::1"):
            print("--base-url accepts only a loopback test server", file=sys.stderr)
            return 2
        base = args.base_url
    jira = Jira(base, email, auth_value)
    handlers = {
        "transitions": cmd_transitions,
        "worklog-list": cmd_worklog_list,
        "worklog-add": cmd_worklog_add,
        "worklog-update": cmd_worklog_update,
        "attach": cmd_attach,
    }
    print(json.dumps(handlers[args.command](jira, args), ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
