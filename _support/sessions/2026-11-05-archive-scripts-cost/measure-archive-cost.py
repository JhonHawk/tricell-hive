"""Measure the context and time that change-folder archives cost, in Claude Code and Codex sessions.

Usage (main threads only, never subagents):
    cd ~/.claude/projects
    python3 -I <this file> [--scripted] \\
        $(rg -l 'changes/archive' --glob '*.jsonl' --glob '!**/subagents/**' -- $(ls -d -- */ | grep -v scratchpad)) \\
        $(rg -l 'changes/archive' ~/.codex/sessions/2026/1[01])

Without --scripted it measures manual archives: a close window ends at the
first commit after a call that moves a change folder into changes/archive,
starts at the latest flow-close load before it (or 60 calls earlier), and is
skipped when it runs apply_deltas.py or check_change.py. With --scripted it
measures closes made with the scripts instead, anchored on the
`apply_deltas.py ... --apply` call. Each call in the window touching openspec
files gets one category; its size is (call text + result) chars / 4, its
carried tokens are that size times the model calls that follow it until the
next compaction, and its seconds run from the event before the call to its
result (model time plus execution, capped at 600 s).

Baseline from the manual run of 2026-11-05 (83 closes, 2026-10-16 to
2026-11-05): engram:tricell-hive/measurement/archive-scripts-baseline.
It omits model reasoning, subagent closes and moves the regexes miss.
"""
import json
import re
import statistics
import sys
from collections import Counter, defaultdict

MOVE = re.compile(r"(\bmv\b|rename\(|shutil\.move|Move to:)[^\n]{0,400}changes/archive|changes/archive[^\n]{0,200}(rename|\bmv\b)"
                  r"|\bmv\b[^\n]*openspec/changes/[^\n]*archive")
COMMIT = re.compile(r"\bgit\b[^\n]{0,80}\bcommit\b")
SPEC = re.compile(r"openspec/specs/")
DELTA = re.compile(r"openspec/changes/(?!archive/)[^/\s\"'\\]+/specs/")
PROPOSAL = re.compile(r"openspec/changes/(?!archive/)[^/\s\"'\\]+/proposal\.md")
CHANGE = re.compile(r"openspec/changes/(?!archive/)[^/\s\"'\\]+/(tasks|design|validation)\.md")
SCRIPTED = re.compile(r"apply_deltas\.py|check_change\.py")
APPLY = re.compile(r"apply_deltas\.py[^\n]*--apply")
SHELL = {"Bash", "exec", "exec_command", "shell", "local_shell_call"}  # calls that run commands
ARCHIVE_STEP = re.compile(r"check_change\.py[^\n]*--archive|apply_deltas\.py")
# Test, replay or development runs: temporary copies, or the script's source in this repository
# (a real close runs the installed copy under the host's skills folder).
NOT_A_CLOSE = re.compile(r"scratchpad|/tmp/|tempfile|mktemp|TMPDIR|unittest|_test\.py|measure-archive|git archive"
                         r"|content/skills/flow-(close|plan)/")
WRITE_HINT = re.compile(r"write_text|\.write\(|open\([^)]*['\"]w|apply_patch|Update File:|Add File:|sed -i|perl -pi|>>|(?<![-=>])>\s*\S*\.md")


def seconds(a, b):
    import datetime
    try:
        f = lambda t: datetime.datetime.fromisoformat(t.replace("Z", "+00:00"))
        return max((f(b) - f(a)).total_seconds(), 0.0)
    except (ValueError, AttributeError, TypeError):
        return None


def text_len(content):
    if isinstance(content, str):
        return len(content)
    n = 0
    for c in content or []:
        if isinstance(c, dict):
            n += len(c.get("text") or "") if c.get("type") == "text" else len(json.dumps(c))
    return n


def classify(name, blob, path=""):
    if SCRIPTED.search(blob):
        return "script"
    if MOVE.search(blob):
        return "move"
    writes = name in ("Edit", "Write", "MultiEdit") or bool(WRITE_HINT.search(blob))
    if COMMIT.search(blob) and "openspec" in blob and not SPEC.search(blob):
        return "commit"
    target = path or blob
    if SPEC.search(target):
        return "spec write" if writes else "spec read"
    if DELTA.search(target):
        return "delta write" if writes else "delta read"
    if PROPOSAL.search(target):
        if writes:
            return "status edit" if re.search(r"Status|Estado", blob) else "proposal edit"
        return "change read"
    if CHANGE.search(target):
        return "tasks edit" if writes else "change read"
    return None


def load_claude(path):
    calls, order, seen, segments, pending = [], [], set(), [0], {}
    prev = ""
    for line in open(path, encoding="utf-8"):
        try:
            e = json.loads(line)
        except ValueError:
            continue
        cur = e.get("timestamp") or prev
        before, prev = prev, cur
        if e.get("type") == "system" and e.get("subtype") == "compact_boundary":
            segments.append(len(order))
        m = e.get("message") or {}
        if e.get("type") == "assistant":
            mid = m.get("id")
            if mid and m.get("usage") and mid not in seen:
                seen.add(mid)
                order.append(mid)
            for c in m.get("content") or []:
                if isinstance(c, dict) and c.get("type") == "tool_use":
                    inp = c.get("input") or {}
                    rec = {"idx": len(order), "name": c.get("name"), "blob": json.dumps(inp, ensure_ascii=False),
                           "path": inp.get("file_path") or "", "ts": e.get("timestamp") or "", "result": 0, "prev": before, "res_ts": ""}
                    rec["close"] = (rec["name"] == "Skill" and "flow-close" in rec["blob"]) or \
                        (rec["name"] == "Read" and "flow-close/SKILL.md" in rec["path"])
                    calls.append(rec)
                    pending[c.get("id")] = rec
        elif e.get("type") == "user":
            content = m.get("content")
            if isinstance(content, str) and "<command-name>/flow-close" in content:
                calls.append({"idx": len(order), "name": "_cmd", "blob": "", "path": "", "ts": e.get("timestamp") or "",
                              "result": 0, "close": True})
            elif isinstance(content, list):
                for c in content:
                    if isinstance(c, dict) and c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                        r = pending.pop(c["tool_use_id"])
                        r["result"] = text_len(c.get("content"))
                        r["res_ts"] = e.get("timestamp") or ""
    return calls, len(order), segments


def load_codex(path):
    calls, n, segments, pending = [], 0, [0], {}
    prev = ""
    for line in open(path, encoding="utf-8"):
        try:
            e = json.loads(line)
        except ValueError:
            continue
        cur = e.get("timestamp") or prev
        before = prev
        p = e.get("payload") or {}
        if not (e.get("type") == "response_item" and p.get("type") in ("function_call", "custom_tool_call", "local_shell_call", "reasoning")):
            prev = cur
        t, pt = e.get("type"), p.get("type")
        if t == "event_msg" and pt == "token_count":
            n += 1
        elif t == "compacted" or (t == "event_msg" and pt == "context_compacted"):
            segments.append(n)
        elif t == "response_item" and pt in ("function_call", "custom_tool_call", "local_shell_call"):
            blob = str(p.get("arguments") or p.get("input") or p.get("action") or "").replace("\\n", "\n")
            rec = {"idx": n, "name": p.get("name") or pt, "blob": blob, "path": "", "ts": e.get("timestamp") or "",
                   "result": 0, "close": bool(re.search(r"flow-close/SKILL\.md", blob)), "prev": before, "res_ts": ""}
            calls.append(rec)
            pending[p.get("call_id")] = rec
        elif t == "response_item" and pt in ("function_call_output", "custom_tool_call_output") and p.get("call_id") in pending:
            out = p.get("output")
            r = pending.pop(p["call_id"])
            r["result"] = len(out if isinstance(out, str) else json.dumps(out))
            r["res_ts"] = e.get("timestamp") or ""
        elif t == "response_item" and pt == "message" and p.get("role") == "user":
            text = json.dumps(p.get("content"))
            if re.search(r"[$/]flow-close\b", text):
                calls.append({"idx": n, "name": "_cmd", "blob": "", "path": "", "ts": e.get("timestamp") or "",
                              "result": 0, "close": True})
    return calls, n, segments


def analyze(path, scripted=False):
    head = open(path, encoding="utf-8").readline()
    calls, total, segments = load_codex(path) if '"session_meta"' in head else load_claude(path)
    closes, used = [], set()
    for i, rec in enumerate(calls):
        anchor = APPLY.search(rec["blob"]) if scripted else MOVE.search(rec["blob"])
        if i in used or rec["name"] == "_cmd" or not anchor:
            continue
        loaded = next((j for j in range(i - 1, max(-1, i - 400), -1) if calls[j]["close"]), None)
        if scripted and (loaded is None or rec["name"] not in SHELL or NOT_A_CLOSE.search(rec["blob"])):
            continue  # a test, replay or measurement run, not a real close
        start = loaded if loaded is not None else max(0, i - 60)
        end = next((k for k in range(i, min(len(calls), i + 80)) if COMMIT.search(calls[k]["blob"])), i)
        if scripted:
            if not re.search(r"archive", calls[end]["blob"]) or not COMMIT.search(calls[end]["blob"]):
                continue  # no archive commit follows: a verification run, not a close
            # the archive part starts at its first check or preview, not at the build's earlier checks
            start = next((j for j in range(start, i + 1) if ARCHIVE_STEP.search(calls[j]["blob"])), i)
        window = calls[start:end + 1]
        if any(SCRIPTED.search(r["blob"]) for r in window) != scripted:
            continue
        used.update(range(start, end + 1))
        cats = defaultdict(lambda: [0, 0.0, 0.0])
        times = {}
        for r in window:
            if r["name"] == "_cmd":
                continue
            cat = classify(r["name"], r["blob"], r["path"])
            if not cat:
                continue
            tok = (len(r["blob"]) + r["result"]) / 4
            stop = next((s for s in segments if s > r["idx"]), total)
            cats[cat][0] += 1
            cats[cat][1] += tok
            cats[cat][2] += tok * max(stop - r["idx"], 0)
            secs = seconds(r.get("prev"), r.get("res_ts") or r["ts"])
            times.setdefault(cat, []).append(secs)
        name = re.search(r"changes/(?!archive)([A-Za-z0-9._-]+)", rec["blob"])
        closes.append({"file": path, "host": "codex" if '"session_meta"' in head else "claude",
                       "ts": rec["ts"][:16], "change": name.group(1) if name else "?", "cats": dict(cats), "times": times,
                       "span": seconds(window[0]["ts"], window[-1].get("res_ts") or window[-1]["ts"])})
    return closes


REPLACEABLE = {"spec read", "spec write", "delta read", "delta write", "status edit", "move", "script"}


def stats(xs, unit=""):
    xs = sorted(x for x in xs if x is not None)
    if not xs:
        return "n=0"
    return (f"n={len(xs)} median={statistics.median(xs):.0f}{unit} mean={statistics.mean(xs):.0f}{unit} "
            f"p90={xs[int(0.9 * (len(xs) - 1))]:.0f}{unit}")


def summary(rows):
    for label, merged in (("with deltas", True), ("without deltas", False)):
        sel = [r for r in rows if r["merge"] == merged]
        ctx = [sum(v[1] for k, v in r["cats"].items() if k in REPLACEABLE) for r in sel]
        carried = [sum(v[2] for k, v in r["cats"].items() if k in REPLACEABLE) for r in sel]
        secs = [sum(min(x, 600) for k, v in r["times"].items() if k in REPLACEABLE for x in v if x is not None)
                for r in sel]
        print(f"== {label}\n  context tokens: {stats(ctx)}\n  carried tokens: {stats(carried)}\n  seconds: {stats(secs, 's')}")
        for host in ("claude", "codex"):
            h = [c for c, r in zip(ctx, sel) if r["host"] == host]
            if h:
                print(f"  {host} context: {stats(h)}")


if __name__ == "__main__":
    args = sys.argv[1:]
    scripted = "--scripted" in args
    rows = [c for p in args if p != "--scripted" for c in analyze(p, scripted)]
    rows.sort(key=lambda c: c["ts"])
    for c in rows:
        c["merge"] = scripted or any(k in c["cats"] for k in ("spec write", "delta read", "delta write"))
        rep = sum(v[1] for k, v in c["cats"].items() if k in REPLACEABLE)
        print(f'{c["ts"]} {c["host"]:6} {"Δ" if c["merge"] else " "} {c["change"][:38]:38} repl {rep:7.0f} '
              + " ".join(f'{k}:{v[0]}/{v[1]:.0f}' for k, v in sorted(c["cats"].items())))
    print()
    summary(rows)
