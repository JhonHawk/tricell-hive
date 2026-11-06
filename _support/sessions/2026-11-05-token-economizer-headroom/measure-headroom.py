"""Measure the cost headroom TokenEconomizer's three mechanisms target, on real sessions.

Replicates the "headroom measurements" of TokenEconomizer (arXiv 2610.22732, appendix B)
on this machine's own Claude Code and Codex main-thread sessions, with no model call:

- M1 repeated actions: share of a session's tool calls whose action key already appeared in
  an earlier session of the same repository, against a control (the same share measured
  against each other repository's full history, averaged).
- M2 unused tools (Claude Code only, from the recorded request shapes): the loaded tool
  definitions a session never calls, and the share of the session's cost they carry as
  cached prefix; plus the tools unused across a repository's sessions under TokenEconomizer's
  default criterion (unused over >= 10 prompts across >= 2 sessions, or >= 5 sessions).
- M3 post-repetition spend: share of a session's cost spent after a failure signature (the
  normalized first error line of a failing call) first repeats inside it.
- M4 cross-session failures: failure signatures that recur in >= 3 sessions of one repository,
  the evidence micro-skills needs before it states a fact.

Cost is relative effective tokens: Claude input 1, cache write 1.25 (5 min) or 2 (1 h),
cache read 0.1, output 5; Codex uncached input 1, cached 0.1, output 8. Hosts are not mixed.
All shares are upper bounds of exposure, not achievable savings.

Usage: python3 -I measure-headroom.py [--days 30] [--json out.json]
"""
import argparse
import glob
import json
import os
import re
import statistics
import time
from collections import Counter, defaultdict

HOME = os.path.expanduser("~")
WORKTREE = re.compile(r"/\.(claude|grok|codex)/worktrees/[^/]+.*$")
EXIT = re.compile(r"(?:^exit code|process exited with code|\"exit_code\":)\s*(-?\d+)", re.I | re.M)
BOILER = re.compile(r"^(exit code -?\d+|process exited with code -?\d+|<tool_use_error>|</tool_use_error>|"
                    r"chunk id:.*|wall time:.*|output:|original token count:.*|command:.*|"
                    r"warning: truncated output.*)$", re.I)
NOT_WORK_FAILURE = re.compile(r"user doesn't want|user rejected|was rejected|permission to use|denied by|"
                              r"interrupted by user|request interrupted|hook.*block", re.I)
SHELL = {"Bash", "exec_command", "shell", "local_shell_call", "exec"}
FILE_TOOLS = {"Read", "Edit", "Write", "MultiEdit", "NotebookEdit", "Grep", "Glob"}
W_CLAUDE = {"in": 1.0, "w5": 1.25, "w1h": 2.0, "read": 0.1, "out": 5.0}
W_CODEX = {"in": 1.0, "read": 0.1, "out": 8.0}


def norm_repo(cwd):
    return WORKTREE.sub("", (cwd or "").rstrip("/")) or "?"


def shell_key(cmd):
    if isinstance(cmd, list):
        cmd = " ".join(str(c) for c in cmd)
        cmd = re.sub(r"^(bash|zsh|sh) -l?c ", "", cmd)
    c = re.sub(r"\s+", " ", str(cmd).strip().lower())
    c = re.sub(r"^cd [^&;]+(&&|;)\s*", "", c)
    return "sh:" + " ".join(c.split(" ")[:6])


def action_key(name, inp):
    if name in SHELL:
        cmd = inp.get("command") or inp.get("cmd") or inp.get("argv") or ""
        return shell_key(cmd)
    if name in FILE_TOOLS:
        p = inp.get("file_path") or inp.get("path") or inp.get("notebook_path") or inp.get("pattern") or ""
        return f"{name}:{os.path.basename(str(p))}"
    if name == "apply_patch":
        m = re.search(r"\*\*\* (?:Update|Add|Delete) File: (\S+)", inp.get("patch") or "")
        return f"apply_patch:{os.path.basename(m.group(1)) if m else ''}"
    if name == "Skill":
        return f"Skill:{inp.get('skill')}"
    return name


def failure_signature(text):
    if not text or NOT_WORK_FAILURE.search(text[:400]):
        return None
    lines = [l.strip() for l in text.splitlines() if l.strip() and not BOILER.match(l.strip())]
    if lines and lines[0].lower().startswith("traceback"):
        lines = lines[-1:]  # a Python traceback is identified by its exception line
    for line in lines:
        line = re.sub(r"</?tool_use_error>", "", line).strip()
        if not line or BOILER.match(line):
            continue
        line = line.lower()
        line = re.sub(r"/users/[^\s:'\"]+", "<path>", line)
        line = re.sub(r"0x[0-9a-f]+|\b[0-9a-f]{7,40}\b", "<hex>", line)
        line = re.sub(r"\d+", "#", line)
        line = re.sub(r"\s+", " ", line)[:120]
        if len(re.findall(r"[a-z]{2,}", line)) < 2:
            continue  # separators, bare paths: not a signature
        return line
    return None


CODE_CMD = re.compile(r"""tools\.(\w+)\(\s*\{\s*cmd\s*:\s*("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|`[^`]*`)|tools\.(\w+)\(""")
CODE_RESULT = re.compile(r'"exit_code":(-?\d+).*?"output":"((?:[^"\\]|\\.)*)"', re.S)


def code_mode_calls(js):
    """Codex code mode: one `exec` call runs a script that calls several tools."""
    out = []
    for m in CODE_CMD.finditer(js or ""):
        if m.group(1):
            out.append((m.group(1), shell_key(m.group(2)[1:-1].encode().decode("unicode_escape", "replace"))))
        else:
            out.append((m.group(3), m.group(3)))
    return out or [("exec", "exec")]


def code_mode_results(text):
    res = []
    for m in CODE_RESULT.finditer(text):
        try:
            body = json.loads('"' + m.group(2) + '"')
        except ValueError:
            body = m.group(2)
        res.append((m.group(1), body))
    return res


def text_of(content):
    if isinstance(content, str):
        return content
    out = []
    for c in content or []:
        if isinstance(c, dict) and c.get("type") == "text":
            out.append(c.get("text") or "")
    return "\n".join(out)


def load_claude(path):
    s = {"host": "claude", "path": path, "calls": [], "cost": 0.0, "requests": 0, "catalog": {}, "start": None, "cwd": None}
    seen, pending = set(), {}
    for line in open(path, encoding="utf-8", errors="replace"):
        try:
            e = json.loads(line)
        except ValueError:
            continue
        if e.get("isSidechain"):
            continue
        s["start"] = s["start"] or e.get("timestamp")
        s["cwd"] = s["cwd"] or e.get("cwd")
        t = e.get("type")
        if t == "api-request-shape":
            for tool in (e.get("shape") or {}).get("tools") or []:
                if tool.get("defer_loading") or not tool.get("name"):
                    continue
                s["catalog"][tool["name"]] = max(s["catalog"].get(tool["name"], 0), len(json.dumps(tool)))
        m = e.get("message") or {}
        if t == "assistant":
            mid, u = m.get("id"), m.get("usage")
            if mid and u and mid not in seen:
                seen.add(mid)
                cc = u.get("cache_creation") or {}
                w1h = cc.get("ephemeral_1h_input_tokens") or 0
                w5 = cc.get("ephemeral_5m_input_tokens")
                if w5 is None:
                    w5 = (u.get("cache_creation_input_tokens") or 0) - w1h
                s["cost"] += (W_CLAUDE["in"] * (u.get("input_tokens") or 0) + W_CLAUDE["w5"] * w5 + W_CLAUDE["w1h"] * w1h
                              + W_CLAUDE["read"] * (u.get("cache_read_input_tokens") or 0) + W_CLAUDE["out"] * (u.get("output_tokens") or 0))
                s["requests"] += 1
            for c in m.get("content") or []:
                if isinstance(c, dict) and c.get("type") == "tool_use":
                    rec = {"name": c.get("name"), "key": action_key(c.get("name"), c.get("input") or {}),
                           "at": s["cost"], "sig": None}
                    s["calls"].append(rec)
                    pending[c.get("id")] = rec
        elif t == "user" and isinstance(m.get("content"), list):
            for c in m["content"]:
                if isinstance(c, dict) and c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                    r = pending.pop(c["tool_use_id"])
                    txt = text_of(c.get("content"))
                    code = EXIT.search(txt[:300])
                    if c.get("is_error") or (code and code.group(1) != "0"):
                        r["sig"] = failure_signature(txt)
    return s


def load_codex(path):
    s = {"host": "codex", "path": path, "calls": [], "cost": 0.0, "requests": 0, "catalog": {}, "start": None, "cwd": None}
    pending, last_total = {}, None
    with open(path, encoding="utf-8", errors="replace") as f:
        meta = json.loads(f.readline()).get("payload") or {}
        if meta.get("thread_source") != "user" or meta.get("source") not in ("vscode", "cli"):
            return None
        s["start"], s["cwd"] = meta.get("timestamp"), meta.get("cwd")
        for line in f:
            try:
                e = json.loads(line)
            except ValueError:
                continue
            p = e.get("payload") or {}
            t, pt = e.get("type"), p.get("type")
            if t == "event_msg" and pt == "token_count":
                info = p.get("info") or {}
                total = json.dumps(info.get("total_token_usage"))
                u = info.get("last_token_usage")
                if not u or total == last_total:
                    continue
                last_total = total
                cached = u.get("cached_input_tokens") or 0
                s["cost"] += (W_CODEX["in"] * max((u.get("input_tokens") or 0) - cached, 0) + W_CODEX["read"] * cached
                              + W_CODEX["out"] * (u.get("output_tokens") or 0))
                s["requests"] += 1
            elif t == "response_item" and pt in ("function_call", "custom_tool_call", "local_shell_call"):
                raw = p.get("arguments") or p.get("input") or p.get("action") or {}
                if isinstance(raw, str):
                    try:
                        inp = json.loads(raw)
                    except ValueError:
                        inp = {"patch": raw}
                else:
                    inp = raw
                if not isinstance(inp, dict):
                    inp = {"cmd": inp}
                name = p.get("name") or pt
                if name == "exec" and pt == "custom_tool_call":
                    recs = [{"name": n, "key": k, "at": s["cost"], "sig": None} for n, k in code_mode_calls(raw)]
                else:
                    recs = [{"name": name, "key": action_key(name, inp), "at": s["cost"], "sig": None}]
                s["calls"].extend(recs)
                pending[p.get("call_id")] = recs
            elif t == "response_item" and pt in ("function_call_output", "custom_tool_call_output") and p.get("call_id") in pending:
                recs = pending.pop(p["call_id"])
                out = p.get("output")
                txt = out if isinstance(out, str) else "\n".join(
                    (c.get("text") or "") for c in out if isinstance(c, dict)) if isinstance(out, list) else json.dumps(out)
                results = code_mode_results(txt)
                if results:
                    shells = [r for r in recs if r["name"] == "exec_command"] or recs
                    for r, (code, body) in zip(shells, results):
                        if code != "0":
                            r["sig"] = failure_signature(body)
                else:
                    code = EXIT.search(txt[:400])
                    if code and code.group(1) != "0":
                        body = re.sub(r"(?s)^.*?\nOutput:\n", "", txt, count=1)
                        recs[0]["sig"] = failure_signature(body)
    return s


def sessions(days):
    cutoff = time.time() - days * 86400
    out = []
    for f in glob.glob(f"{HOME}/.claude/projects/*/*.jsonl"):
        if "scratchpad" in f or os.path.getmtime(f) < cutoff:
            continue
        out.append(load_claude(f))
    for f in glob.glob(f"{HOME}/.codex/sessions/*/*/*/*.jsonl"):
        if os.path.getmtime(f) < cutoff:
            continue
        s = load_codex(f)
        if s:
            out.append(s)
    out = [s for s in out if len(s["calls"]) >= 3 and s["cost"] > 0 and s["start"]]
    for s in out:
        s["repo"] = norm_repo(s["cwd"])
    return out


def med(xs):
    return round(statistics.median(xs), 3) if xs else None


def pooled(num, den):
    return round(num / den, 3) if den else None


def m1(group):
    by_repo = defaultdict(list)
    for s in group:
        by_repo[s["repo"]].append(s)
    hist = {r: {c["key"] for s in ss for c in s["calls"]} for r, ss in by_repo.items()}
    within, control, n_calls, n_rep, n_ctrl = [], [], 0, 0, 0
    for r, ss in by_repo.items():
        if len(ss) < 3:
            continue
        seen = set()
        others = [hist[o] for o in hist if o != r and len(by_repo[o]) >= 3]
        for s in sorted(ss, key=lambda x: x["start"]):
            keys = [c["key"] for c in s["calls"]]
            if seen:
                rep = sum(k in seen for k in keys)
                within.append(rep / len(keys))
                n_calls += len(keys)
                n_rep += rep
                if others:
                    ctrl = statistics.mean(sum(k in h for k in keys) / len(keys) for h in others)
                    control.append(ctrl)
                    n_ctrl += ctrl * len(keys)
            seen.update(keys)
    return {"sessions": len(within), "median_within": med(within), "median_control": med(control),
            "pooled_within": pooled(n_rep, n_calls), "pooled_control": pooled(n_ctrl, n_calls)}


def m2(group):
    shares, unused_tok, by_repo = [], Counter(), defaultdict(list)
    total_cost = total_unused = 0.0
    for s in group:
        if not s["catalog"]:
            continue
        used = {c["name"] for c in s["calls"]}
        unused_chars = sum(v for k, v in s["catalog"].items() if k not in used)
        carried = unused_chars / 4 * s["requests"] * W_CLAUDE["read"]
        shares.append(carried / s["cost"])
        total_cost += s["cost"]
        total_unused += carried
        for k, v in s["catalog"].items():
            if k not in used:
                unused_tok[k] += 1
        by_repo[s["repo"]].append(s)
    n = len(shares)
    prunable = {}
    for r, ss in by_repo.items():
        ss = sorted(ss, key=lambda x: x["start"])
        tools = {k for s in ss for k in s["catalog"]}
        for k in tools:
            unused_sessions = [s for s in ss if k in s["catalog"] and k not in {c["name"] for c in s["calls"]}]
            ever_used = any(k in {c["name"] for c in s["calls"]} for s in ss)
            prompts = sum(s["requests"] for s in unused_sessions)
            if not ever_used and (len(unused_sessions) >= 5 or (prompts >= 10 and len(unused_sessions) >= 2)):
                prunable.setdefault(k, []).append(r)
    sizes = {}
    for s in group:
        for k, v in s["catalog"].items():
            sizes[k] = max(sizes.get(k, 0), v)
    return {"sessions": n, "median_share": med(shares), "pooled_share": pooled(total_unused, total_cost),
            "unused_rate": {k: round(v / n, 2) for k, v in unused_tok.most_common()} if n else {},
            "tool_tokens": {k: v // 4 for k, v in sorted(sizes.items(), key=lambda x: -x[1])},
            "never_used_in_repo": {k: len(v) for k, v in sorted(prunable.items(), key=lambda x: -len(x[1]))},
            "repos": len(by_repo)}


def m3(group):
    shares, with_rep, after, total = [], 0, 0.0, 0.0
    for s in group:
        seen, first = set(), None
        for c in s["calls"]:
            if c["sig"]:
                if c["sig"] in seen:
                    first = c["at"]
                    break
                seen.add(c["sig"])
        share = (s["cost"] - first) / s["cost"] if first is not None else 0.0
        with_rep += first is not None
        shares.append(share)
        after += s["cost"] * share
        total += s["cost"]
    fails = sum(1 for s in group for c in s["calls"] if c["sig"])
    calls = sum(len(s["calls"]) for s in group)
    return {"sessions": len(group), "with_repetition": with_rep, "median_share_after": med(shares),
            "pooled_share_after": pooled(after, total), "failing_call_rate": pooled(fails, calls)}


def m4(group):
    by_repo = defaultdict(lambda: defaultdict(set))
    calls_of = Counter()
    example = {}
    for i, s in enumerate(group):
        for c in s["calls"]:
            if c["sig"]:
                by_repo[s["repo"]][c["sig"]].add(i)
                calls_of[(s["repo"], c["sig"])] += 1
                example.setdefault((s["repo"], c["sig"]), c["key"])
    rows = []
    for r, sigs in by_repo.items():
        for sig, ss in sigs.items():
            if len(ss) >= 3:
                rows.append({"repo": os.path.basename(r), "sessions": len(ss), "calls": calls_of[(r, sig)],
                             "signature": sig, "example_key": example[(r, sig)][:80]})
    rows.sort(key=lambda x: (-x["sessions"], -x["calls"]))
    fails = sum(calls_of.values())
    return {"recurring_signatures": len(rows), "calls_in_recurring": sum(x["calls"] for x in rows),
            "all_failing_calls": fails, "top": rows[:25]}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--days", type=int, default=30)
    ap.add_argument("--json")
    a = ap.parse_args()
    ss = sessions(a.days)
    report = {"days": a.days, "generated": time.strftime("%Y-%m-%d %H:%M %z")}
    for host in ("claude", "codex"):
        g = [s for s in ss if s["host"] == host]
        report[host] = {"sessions": len(g), "repos": len({s["repo"] for s in g}),
                        "calls": sum(len(s["calls"]) for s in g), "cost": round(sum(s["cost"] for s in g)),
                        "M1": m1(g), "M3": m3(g), "M4": m4(g)}
        if host == "claude":
            report[host]["M2"] = m2(g)
    text = json.dumps(report, indent=1, ensure_ascii=False)
    if a.json:
        open(a.json, "w", encoding="utf-8").write(text)
    print(text)


if __name__ == "__main__":
    main()
