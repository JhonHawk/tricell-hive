#!/usr/bin/env python3
"""List what already worked on this machine for a project, from session transcripts.

Reads the top-level Claude Code and Codex transcripts of the last N days whose
session ran in a project root or below it, pairs each shell command with its
result, and prints JSON candidates for HIVE.local.md: commands that exited 0,
commands found missing, ports, and cloud profiles and regions, each with how
many sessions show it and when it was last seen. It masks every value that
could be a secret before printing, reads only, and never runs a transcript
command.
"""

import argparse
import datetime
import json
import os
import re
import shlex
import sys
import time
from pathlib import Path

NOT_COVERED = ["grok", "pi", "opencode", "cursor", "antigravity"]
SHELLS = {"sh", "bash", "zsh", "fish", "dash"}
PREFIXES = {"sudo", "env", "command", "exec", "time", "nohup"}
RELEVANT = {
    "docker", "docker-compose", "podman", "colima", "npm", "pnpm", "yarn", "bun", "npx", "node", "deno",
    "make", "just", "task", "go", "python", "python3", "uv", "poetry", "pytest", "cargo", "mvn", "gradle",
    "java", "psql", "pg_isready", "mongosh", "mongo", "mongod", "redis-cli", "mysql", "aws", "gcloud", "az",
    "kubectl", "helm", "terraform", "ssh", "scp", "lsof", "acli", "vercel", "supabase", "prisma", "flyctl",
    "agent-browser", "playwright",
}
KEEP_ENV = {"AWS_PROFILE", "AWS_REGION", "AWS_DEFAULT_REGION", "NODE_ENV", "RAILS_ENV", "APP_ENV"}
PUBLIC_ENV = re.compile(r"^(?:NEXT_PUBLIC_|VITE_|PUBLIC_)|_(?:ENV|PORT|REGION|PROFILE)$")
FLAGS = {"true", "false", "yes", "no", "on", "off"}
INLINE = {"python", "python3", "node", "deno", "bun", "ruby", "perl"}
INLINE_FLAGS = {"-", "-c", "-e", "-p", "--eval", "--print"}
SECRET_FLAG = re.compile(r"pass|pwd|secret|token|key|auth|cred|cookie|bearer|session", re.I)
# Flags whose next argument may be a credential or a payload: always masked.
VALUE_FLAGS = {"-H", "--header", "-u", "--user", "-b", "--cookie", "--data", "--data-raw", "--data-binary",
               "--data-urlencode", "--json", "--form", "-F", "-a", "--body", "--value"}
# -p and -P carry a port only for these programs; for any other, such as mysql, they may carry a password.
PORT_P = {"docker", "docker-compose", "podman", "ssh", "scp", "kubectl"}
PORT_VALUE = re.compile(r"^\d{2,5}(?::\d{2,5})*$")
HTTP_CLIENTS = {"curl", "wget", "http"}
# Long flags whose next argument is kept; after any other long flag a plain word is masked.
KEEP_NEXT = {"--port", "--profile", "--region", "--filter", "--file", "--env-file", "--project", "--project-name",
             "--prefix", "--dir", "--cwd", "--workspace", "--tail", "--since", "--context", "--namespace", "--name",
             "--target", "--stage", "--branch", "--base", "--host", "--username", "--db", "--dbname", "--database"}
NUMERIC = re.compile(r"^\d+(?::\d+)*$")
ENTROPY = re.compile(r"^(?=.*[A-Za-z])(?=.*\d)[^/]{20,}$")


def secret_like(text):
    return bool(ENTROPY.match(text) or MIXED.match(text))


def safe_path(token):
    """A path whose every part looks like a name, not like a key."""
    return ("/" in token and PATHLIKE.match(token) is not None and len(token) <= 160
            and not any(secret_like(part) for part in token.split("/")))
ASSIGN = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*)=(.*)$", re.S)
USERINFO = re.compile(r"^([A-Za-z][A-Za-z0-9+.-]*://)\S*@")  # greedy: up to the last @, so a password may hold / or @
PATHLIKE = re.compile(r"^[\w./~:%@-]+$")
MIXED = re.compile(r"^(?=.*[a-z])(?=.*[A-Z])[A-Za-z0-9+=_-]{12,}$")
SIMPLE = re.compile(r"^[\w.:/%+@,-]+$")
# Whole-line shell messages for a missing program (zsh, bash in English or Spanish, dash).
MISSING = [
    re.compile(r"^\S+: command not found: ([A-Za-z0-9._+-]+)\s*$"),
    re.compile(r"^(?:\S+: )?(?:\S+ \d+: )?([A-Za-z0-9._+-]+): (?:command not found|orden no encontrada)\s*$"),
    re.compile(r"^\S+: \d+: ([A-Za-z0-9._+-]+): not found\s*$"),
]
PORTS = [
    re.compile(r"(?:localhost|127\.0\.0\.1|0\.0\.0\.0):(\d{2,5})\b"),
    re.compile(r"--port[= ](\d{2,5})\b"),
    re.compile(r"(?:^| )-p (\d{2,5})(?::\d{2,5})?\b"),
    re.compile(r"\bPORT=(\d{2,5})\b"),
    re.compile(r"-iTCP:(\d{2,5})\b"),
]
REDACTED = "<redacted>"


def mask_assignment(name, value):
    """Keep an environment value only when it cannot be a secret."""
    if SECRET_FLAG.search(name):
        return REDACTED
    if value.isdigit() or value.lower() in FLAGS:
        return value
    if value.startswith(("/", "~/", "./")) and safe_path(value):
        return value
    if (name in KEEP_ENV or PUBLIC_ENV.search(name)) and SIMPLE.match(value) and len(value) <= 40:
        return value
    return REDACTED


def mask_url(token):
    token = USERINFO.sub(r"\g<1>" + REDACTED + "@", token)
    base, query, _ = token.partition("?")
    base = base.split("#", 1)[0]
    if any(c in base for c in " \t\n'\"`{}$\\") or len(base) > 200:
        return REDACTED
    scheme, _, rest = base.partition("://")
    parts = [REDACTED if secret_like(part) else part for part in rest.split("/")]
    return scheme + "://" + "/".join(parts) + ("?" + REDACTED if query else "")


def mask_token(token):
    if "://" in token and not ASSIGN.match(token):
        return mask_url(token)
    match = ASSIGN.match(token)
    if match:
        name, value = match.groups()
        return name + "=" + mask_assignment(name, value)
    if token.startswith("-") and "=" in token:
        flag, value = token.split("=", 1)
        if flag in VALUE_FLAGS or SECRET_FLAG.search(flag):
            return flag + "=" + REDACTED
        if flag in KEEP_NEXT or NUMERIC.match(value) or safe_path(value):
            return flag + "=" + mask_token(value)
        return flag + "=" + REDACTED
    if any(c in token for c in " \t\n'\"`{}$\\"):
        return REDACTED
    if safe_path(token):
        return token
    if len(token) > 40 or secret_like(token):
        return REDACTED
    return token


def mask_tokens(tokens):
    """Keep the program, words, flag names, numbers and paths; mask what a flag may carry."""
    program = os.path.basename(tokens[0]) if tokens else ""
    out, pending = [], None
    for token in tokens:
        if pending == "hide":
            out.append(REDACTED)
            pending = None
            continue
        if pending == "port":
            out.append(token if PORT_VALUE.match(token) else REDACTED)
            pending = None
            continue
        if pending == "long" and not (token.startswith("-") or NUMERIC.match(token) or safe_path(token)):
            out.append(REDACTED)
            pending = None
            continue
        pending = None
        if token.startswith("--") and "=" not in token:
            out.append(token)
            if token in VALUE_FLAGS or SECRET_FLAG.search(token):
                pending = "hide"
            elif token not in KEEP_NEXT:
                pending = "long"
        elif re.fullmatch(r"-[A-Za-z]", token):
            out.append(token)
            if token in ("-p", "-P"):
                pending = "port" if program in PORT_P else "hide"
            elif token in VALUE_FLAGS or (token == "-d" and program in HTTP_CLIENTS):
                pending = "hide"
        elif token.startswith("-") and "=" not in token and len(token) > 2:
            # A cluster such as -it is kept; -pS3cret or -uroot carries a value.
            keep = re.fullmatch(r"-\d+", token) or (re.fullmatch(r"-[a-z]{2,3}", token) and token[1] not in "pu")
            out.append(token if keep else token[:2] + REDACTED)
        else:
            out.append(mask_token(token))
            if out[-1] == token and SECRET_FLAG.search(token) and len(out) > 1:
                pending = "hide"  # a word that names a secret, such as aws_secret_access_key, precedes its value
    return out


def segments(command):
    """Split the first line of a shell command into masked simple commands."""
    first = command.strip().splitlines()[0] if command.strip() else ""
    if "<<" in first:
        first = first.split("<<", 1)[0]
    lexer = shlex.shlex(first, posix=True, punctuation_chars=True)
    lexer.whitespace_split = True
    try:
        tokens = list(lexer)
    except ValueError:
        return []
    parts, current = [], []
    for token in tokens:
        if token and set(token) <= set("&|;()<>"):
            if current:
                parts.append(current)
            current = []
        else:
            current.append(token)
    if current:
        parts.append(current)
    result = []
    for part in parts:
        env, rest = [], list(part)
        while rest and (ASSIGN.match(rest[0]) or rest[0] in PREFIXES):
            token = rest.pop(0)
            if ASSIGN.match(token):
                env.append(mask_token(token))
        if not rest:
            continue
        if os.path.basename(rest[0]) in INLINE and INLINE_FLAGS & set(rest[1:3]):
            continue  # inline code written for one task, not a way to run the project
        result.append({"env": env, "tokens": mask_tokens(rest)})
    return result


def relevant(tokens):
    program = os.path.basename(tokens[0])
    if tokens[0].startswith("./") or program in RELEVANT:
        return True
    return program == "curl" and any(h in " ".join(tokens) for h in ("localhost", "127.0.0.1", "0.0.0.0"))


def missing_programs(text):
    found = set()
    for line in (text or "").splitlines():
        for pattern in MISSING:  # the zsh form names the shell first, so the first match wins
            match = pattern.match(line.strip())
            if match:
                found.add(match.group(1))
                break
    return found


class Collector:
    def __init__(self, limit):
        self.limit = limit
        self.commands, self.missing, self.ports, self.cloud = {}, {}, {}, {}

    @staticmethod
    def _bump(table, key, session, date, extra=None):
        entry = table.setdefault(key, {"sessions": set(), "last_seen": "", "last_session": "", "where": set()})
        entry["sessions"].add(session)
        entry["where"].add(session[2])
        if date >= entry["last_seen"]:
            entry["last_seen"], entry["last_session"] = date, session[1]
        if extra:
            entry.setdefault("env", set()).update(extra)

    def worked(self, command, session, date):
        for seg in segments(command):
            line = " ".join(seg["tokens"])
            joined = " ".join(seg["env"] + seg["tokens"])
            for pattern in PORTS:
                for port in pattern.findall(joined):
                    if 1024 <= int(port) <= 65535:  # below 1024 the matches were PIDs and line numbers
                        self._bump(self.ports, int(port), session, date)
            profile = region = ""
            for i, token in enumerate(seg["tokens"]):
                nxt = seg["tokens"][i + 1] if i + 1 < len(seg["tokens"]) else ""
                if token == "--profile":
                    profile = nxt
                elif token.startswith("--profile="):
                    profile = token.split("=", 1)[1]
                elif token == "--region":
                    region = nxt
                elif token.startswith("--region="):
                    region = token.split("=", 1)[1]
            for assignment in seg["env"]:
                name, _, value = assignment.partition("=")
                if name == "AWS_PROFILE" and value != REDACTED:
                    profile = value
                if name in ("AWS_REGION", "AWS_DEFAULT_REGION") and value != REDACTED:
                    region = value
            if (profile or region) and REDACTED not in (profile, region):
                self._bump(self.cloud, (profile, region), session, date)
            if relevant(seg["tokens"]):
                self._bump(self.commands, line, session, date, extra=seg["env"])

    def failed(self, command, output, code, session, date):
        names = missing_programs(output)
        if not names and code == 127:
            segs = segments(command)
            if segs:
                names = {os.path.basename(segs[0]["tokens"][0])}
        for name in names:
            self._bump(self.missing, name, session, date)

    def _rows(self, table, key_name):
        rows = []
        for key, entry in table.items():
            row = {key_name: key} if not isinstance(key, tuple) else {"profile": key[0], "region": key[1]}
            row.update(sessions=len(entry["sessions"]), last_seen=entry["last_seen"], last_session=entry["last_session"],
                       where=sorted(entry["where"])[:5])
            if entry.get("env"):
                row["env"] = sorted(entry["env"])
            rows.append(row)
        rows.sort(key=lambda r: (-r["sessions"], r["last_seen"]), reverse=False)
        rows.sort(key=lambda r: r["last_seen"], reverse=True)
        rows.sort(key=lambda r: r["sessions"], reverse=True)
        return rows[: self.limit]

    def report(self):
        return {
            "commands": self._rows(self.commands, "command"),
            "missing": self._rows(self.missing, "program"),
            "ports": self._rows(self.ports, "port"),
            "cloud": self._rows(self.cloud, "cloud"),
        }


def in_roots(cwd, roots):
    return bool(cwd) and any(cwd == r or cwd.startswith(r + "/") for r in roots)


def where(cwd, roots):
    """The first folder below the matching root, or "." for the root itself and hidden worktree folders."""
    root = max((r for r in roots if cwd == r or cwd.startswith(r + "/")), key=len)
    first = os.path.relpath(cwd, root).split(os.sep)[0]
    return "." if first.startswith(".") else first


def day(stamp, path):
    if isinstance(stamp, str) and len(stamp) >= 10:
        return stamp[:10]
    return datetime.date.fromtimestamp(path.stat().st_mtime).isoformat()


def read_jsonl(path):
    with open(path, encoding="utf-8", errors="replace") as handle:
        for line in handle:
            try:
                yield json.loads(line)
            except ValueError:
                continue


def text_of(content):
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        return "\n".join(i.get("text", "") for i in content if isinstance(i, dict))
    return ""


def first_cwd(path, field):
    for n, record in enumerate(read_jsonl(path)):
        cwd = field(record)
        if cwd:
            return cwd
        if n > 200:
            return ""
    return ""


def scan_claude(home, roots, cutoff, collector, stats):
    projects = home / "projects"
    if not projects.is_dir():
        return
    for folder in projects.iterdir():
        if not folder.is_dir():
            continue
        for path in folder.glob("*.jsonl"):
            try:
                if path.stat().st_mtime < cutoff:
                    continue
                if not in_roots(first_cwd(path, lambda r: r.get("cwd")), roots):
                    continue
            except OSError:
                stats["unreadable"] += 1
                continue
            stats["sessions"] += 1
            session = ("claude-code", path.stem, where(first_cwd(path, lambda r: r.get("cwd")), roots))
            uses = {}
            for record in read_jsonl(path):
                content = (record.get("message") or {}).get("content")
                if not isinstance(content, list):
                    continue
                for item in content:
                    if not isinstance(item, dict):
                        continue
                    if item.get("type") == "tool_use" and item.get("name") == "Bash":
                        tool_input = item.get("input") or {}
                        uses[item.get("id")] = (tool_input.get("command", ""), bool(tool_input.get("run_in_background")),
                                                day(record.get("timestamp"), path))
                    elif item.get("type") == "tool_result" and item.get("tool_use_id") in uses:
                        command, background, date = uses.pop(item["tool_use_id"])
                        if item.get("is_error"):
                            collector.failed(command, text_of(item.get("content")), None, session, date)
                        elif not background:  # a background command reports success when it starts
                            collector.worked(command, session, date)


def scan_codex(home, roots, cutoff, collector, stats):
    sessions = home / "sessions"
    if not sessions.is_dir():
        return
    for path in sessions.rglob("rollout-*.jsonl"):
        try:
            if path.stat().st_mtime < cutoff:
                continue
            meta = next((r.get("payload") or {} for r in read_jsonl(path) if r.get("type") == "session_meta"), {})
        except OSError:
            stats["unreadable"] += 1
            continue
        source = meta.get("source")
        if not in_roots(meta.get("cwd", ""), roots) or (isinstance(source, dict) and "subagent" in source):
            continue
        stats["sessions"] += 1
        session = ("codex", path.stem, where(meta["cwd"], roots))
        for record in read_jsonl(path):
            payload = record.get("payload") or {}
            item = payload.get("item") or {}
            if payload.get("type") != "item_completed" or item.get("type") != "CommandExecution":
                continue
            argv = item.get("command") or []
            if isinstance(argv, str):
                command = argv
            elif len(argv) >= 3 and os.path.basename(argv[0]) in SHELLS and argv[1] in ("-c", "-lc", "-ic"):
                command = argv[2]
            else:
                command = shlex.join(str(a) for a in argv)
            date = day(record.get("timestamp"), path)
            code = item.get("exit_code")
            if code == 0:
                collector.worked(command, session, date)
            elif code == 127:
                collector.failed(command, item.get("aggregated_output", ""), code, session, date)


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--root", action="append", required=True, help="project root; repeat for a workspace root")
    parser.add_argument("--days", type=int, default=30)
    parser.add_argument("--limit", type=int, default=60, help="most rows per list")
    parser.add_argument("--claude-home", default=os.environ.get("CLAUDE_CONFIG_DIR") or str(Path.home() / ".claude"))
    parser.add_argument("--codex-home", default=os.environ.get("CODEX_HOME") or str(Path.home() / ".codex"))
    args = parser.parse_args()

    roots = set()
    for root in args.root:
        clean = os.path.normpath(os.path.expanduser(root))
        roots.update({clean, os.path.realpath(clean)})
    cutoff = time.time() - args.days * 86400
    collector = Collector(args.limit)
    hosts = {"claude-code": {"sessions": 0, "unreadable": 0}, "codex": {"sessions": 0, "unreadable": 0}}
    scan_claude(Path(args.claude_home), roots, cutoff, collector, hosts["claude-code"])
    scan_codex(Path(args.codex_home), roots, cutoff, collector, hosts["codex"])
    out = {"roots": sorted(set(args.root)), "window_days": args.days, "hosts": hosts, "not_covered": NOT_COVERED}
    out.update(collector.report())
    json.dump(out, sys.stdout, indent=1, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
