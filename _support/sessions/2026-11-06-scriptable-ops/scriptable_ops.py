"""Bottom-up measurement of recurring agent operations that a deterministic script could replace.

Usage: python3 -I scriptable_ops.py build  [--exclude SESSION_ID ...]   # parse transcripts -> calls.pkl
       python3 -I scriptable_ops.py report                               # rank candidates from calls.pkl

Reuses the method of measure-archive-cost.py: size = (call text + result) chars / 4; carried = size x later model
calls until the next compaction; seconds = from the event before the call to its result (capped at 600 s), made
exclusive (parallel calls do not double count). Shell calls are split into simple commands (quotes, heredocs and
pipes respected); each simple command is one operation key and shares the call's size and seconds equally.
Codex `exec` scripts (code mode) are split into their tools.* calls the same way (lower confidence).
"""
import glob, json, os, pickle, re, shlex, statistics, sys, datetime
from collections import Counter, defaultdict

HOME = os.path.expanduser("~")
OUT = os.path.dirname(os.path.abspath(__file__))
PKL = os.path.join(os.environ.get("TMPDIR", "/tmp"), "scriptable-ops-calls.pkl")
WORKTREE = re.compile(r"/\.(claude|grok|codex)/worktrees/[^/]+.*$")
EXIT = re.compile(r"(?:^exit code|process exited with code|\"exit_code\":)\s*(-?\d+)", re.I | re.M)
NOT_WORK_FAILURE = re.compile(r"user doesn't want|user rejected|was rejected|permission to use|denied by|"
                              r"interrupted by user|request interrupted|hook.*block", re.I)
SKILL_PATH = re.compile(r"/skills/([\w:.-]+)/SKILL\.md")
IMAGE_CHARS = 6000  # an image result counts as ~1.5k tokens, not its base64 length


def ts_parse(t):
    try:
        return datetime.datetime.fromisoformat(t.replace("Z", "+00:00")).timestamp()
    except Exception:
        return None


def text_len(content):
    if isinstance(content, str):
        return len(content)
    n = 0
    for c in content or []:
        if isinstance(c, dict):
            if c.get("type") == "text":
                n += len(c.get("text") or "")
            elif c.get("type") == "image":
                n += IMAGE_CHARS
            else:
                n += len(json.dumps(c))
    return n


def text_of(content):
    if isinstance(content, str):
        return content
    return "\n".join((c.get("text") or "") for c in content or [] if isinstance(c, dict) and c.get("type") == "text")


# ---------------------------------------------------------------- path classes
def path_class(p):
    p = str(p)
    if p.endswith("SKILL.md"):
        return "SKILL.md"
    m = re.search(r"/skills/([\w:.-]+)/references/([\w./-]+)", p)
    if m:
        return f"ref/{m.group(1)}/{os.path.basename(m.group(2))}"
    if "/references/" in p and p.endswith(".md"):
        return "ref/" + os.path.basename(p)
    if "openspec" in p or "/changes/" in p:
        b = os.path.basename(p)
        if "/changes/archive" in p:
            return "openspec/archive"
        if "/specs/" in p and "/changes/" in p:
            return "openspec/delta"
        if "/openspec/specs" in p or p.rstrip("/").endswith("/specs"):
            return "openspec/spec"
        if b in ("tasks.md", "design.md", "proposal.md", "validation.md"):
            return "openspec/" + b
        return "openspec/other"
    if "ACCOUNTS.md" in p:
        return "ACCOUNTS.md"
    b = os.path.basename(p)
    if b in ("AGENTS.md", "CLAUDE.md", "CLAUDE.local.md"):
        return "AGENTS.md"
    if "/memory/" in p and b.endswith(".md") or b == "MEMORY.md":
        return "memory"
    if "_support/" in p:
        m = re.search(r"_support/([^/]+)", p)
        return "_support/" + (m.group(1) if m else "")
    if "/tool-results/" in p:
        return "tool-results"
    if "/.claude/" in p or "/.codex/" in p or "/.agents/" in p:
        return "agent-config"
    ext = os.path.splitext(b)[1].lower()
    if ext in (".md", ".mdx"):
        return "md"
    if ext in (".json", ".yaml", ".yml", ".toml", ".lock"):
        return "config"
    if ext in (".png", ".jpg", ".jpeg", ".webp", ".gif", ".pdf", ".svg"):
        return "media"
    if ext in (".log", ".txt", ".csv", ".jsonl"):
        return "data"
    return "src"


# ---------------------------------------------------------------- shell splitting
def split_shell(cmd):
    segs, cur, i, n, q = [], [], 0, len(cmd), None
    heredocs, drop, depth = [], False, 0

    def flush():
        nonlocal cur, drop
        s = "".join(cur).strip()
        cur = []
        if s and not drop:
            segs.append(s)
        drop = False

    while i < n:
        ch = cmd[i]
        if q:
            cur.append(ch)
            if ch == "\\" and q == '"' and i + 1 < n:
                cur.append(cmd[i + 1]); i += 2; continue
            if ch == q:
                q = None
            i += 1; continue
        if ch in "'\"":
            q = ch; cur.append(ch); i += 1; continue
        if ch == "\\" and i + 1 < n:
            cur.append(cmd[i:i + 2]); i += 2; continue
        if cmd.startswith("$(", i):
            depth += 1; cur.append("$("); i += 2; continue
        if depth and ch == ")":
            depth -= 1; cur.append(ch); i += 1; continue
        if depth:
            cur.append(ch); i += 1; continue
        if cmd.startswith("<<", i) and not cmd.startswith("<<<", i):
            m = re.match(r"<<-?\s*(['\"]?)([A-Za-z_]\w*)\1", cmd[i:])
            if m:
                heredocs.append(m.group(2)); cur.append(m.group(0)); i += len(m.group(0)); continue
        if ch == "\n":
            flush(); i += 1
            while heredocs:
                d = heredocs.pop(0)
                m = re.search(r"^[ \t]*" + re.escape(d) + r"[ \t]*$", cmd[i:], re.M)
                i = i + m.end() + 1 if m else n
            continue
        if cmd.startswith("&&", i) or cmd.startswith("||", i):
            flush(); i += 2; continue
        if ch == ";":
            flush(); i += 1; continue
        if ch == "|":
            flush(); drop = True; i += 1; continue
        cur.append(ch); i += 1
    flush()
    return segs


NOOP = {"[", "[[", "echo", "printf", "cd", "export", "set", "true", "false", ":", "source", ".", "unset", "exit", "return",
        "trap", "wait", "local", "read", "shopt", "alias", "pushd", "popd", "{", "}", "done", "fi", "esac", "then", "do", "else"}
PREFIX = {"env", "time", "sudo", "nohup", "command", "exec", "builtin", "nice", "timeout"}
READERS = {"cat", "head", "tail", "less", "bat", "nl", "more"}
SUB2 = {"docker", "kubectl", "aws", "ctx7", "engram", "codegraph", "hive", "gh", "vercel", "neon", "supabase", "fly", "terraform", "claude", "codex"}


def first_path(tokens):
    for t in tokens:
        if t.startswith("-") or t in ("&&",):
            continue
        if "/" in t or re.search(r"\.\w{1,5}$", t):
            return t
    return ""


def seg_ops(seg, depth=0):
    seg = re.sub(r"^(?:(?:do|then|else|elif|if|while|until)\s+)+", "", seg)
    if re.match(r"^(for|while|until|case|select)\b", seg) or seg in ("done", "fi", "esac"):
        return []
    try:
        toks = shlex.split(seg, posix=True)
    except ValueError:
        toks = seg.split()
    while toks and (re.match(r"^\w+=", toks[0]) or toks[0] in PREFIX):
        if toks[0] == "command" and len(toks) > 1 and toks[1].startswith("-"):
            return ["command -v"]
        if toks[0] == "timeout" and len(toks) > 1:
            toks = toks[2:]
        else:
            toks = toks[1:]
    if not toks:
        return []
    name = os.path.basename(toks[0])
    if name in NOOP or name.startswith(("#", "!")):
        return []
    if name in ("bash", "sh", "zsh") and depth < 2:
        for i, t in enumerate(toks[1:], 1):
            if t in ("-c", "-lc", "-ic", "-lic") and i + 1 < len(toks):
                out = []
                for s in split_shell(toks[i + 1]):
                    out += seg_ops(s, depth + 1)
                return out
        sc = next((t for t in toks[1:] if not t.startswith("-")), "")
        return ["script:" + os.path.basename(sc)] if sc else ["sh"]
    rest = toks[1:]
    nonopt = [t for t in rest if not t.startswith("-")]
    if name == "git":
        i, args = 0, []
        while i < len(rest):
            t = rest[i]
            if t in ("-C", "-c", "--git-dir", "--work-tree"):
                i += 2; continue
            if t.startswith("-"):
                i += 1; continue
            args.append(t); i += 1
        if not args:
            return ["git"]
        sub = args[0]
        if sub in ("worktree", "stash", "remote", "submodule", "notes", "bisect") and len(args) > 1:
            sub += " " + args[1]
        return ["git " + sub]
    if name in SUB2:
        k = 2 if (name in ("hive", "docker", "kubectl", "ctx7", "gh") and nonopt[:1] in (["view"], ["pr"], ["issue"], ["run"], ["repo"], ["release"], ["compose"], ["workflow"], ["label"])) else 1
        if name == "gh" and nonopt[:1] == ["api"]:
            return ["gh api"]
        if name == "hive" and nonopt[:1] == ["view"]:
            k = 2
        return [" ".join([name] + nonopt[:k])]
    if name in ("go", "cargo", "make", "just", "task", "deno", "bun", "mvn", "gradle", "dotnet"):
        return [" ".join([name] + nonopt[:1])]
    if name in ("npm", "pnpm", "yarn", "npx"):
        if name == "npx":
            return ["npx " + (nonopt[0] if nonopt else "")]
        a = nonopt[:2] if nonopt[:1] in (["run"], ["exec"], ["dlx"]) else nonopt[:1]
        return [" ".join([name] + a)]
    if name in ("node", "python3", "python", "ruby", "perl"):
        if "-m" in rest:
            i = rest.index("-m")
            return [f"{name} -m {rest[i + 1] if i + 1 < len(rest) else ''}"]
        if "-c" in rest or "-e" in rest or "-" in rest or "<<" in seg or "-I" in rest and not nonopt:
            sc = next((t for t in nonopt if re.search(r"\.(py|js|mjs|rb|pl)$", t)), "")
            if sc:
                return ["script:" + os.path.basename(sc)]
            return [name + " inline"]
        sc = next((t for t in nonopt if re.search(r"\.(py|js|mjs|rb|pl)$", t)), "")
        return ["script:" + os.path.basename(sc)] if sc else [name + " inline"]
    if name in READERS:
        return ["read:" + path_class(first_path(rest))]
    if name == "sed":
        if any(t == "-i" or t.startswith("-i") and not t.startswith("--") for t in rest):
            return ["edit:" + path_class(first_path(rest))]
        return ["read:" + path_class(first_path(rest))]
    if name == "mv" and "archive" in seg:
        return ["mv:openspec-archive"]
    if name.startswith(("./", "/")) or re.search(r"\.(sh|py)$", name):
        return ["script:" + os.path.basename(name)]
    if re.match(r"^[./~]", toks[0]) and "/" in toks[0]:
        return ["script:" + os.path.basename(toks[0])]
    return [name]


INLINE_FEATS = [("openspec", r"openspec|changes/archive"), ("jsonl", r"jsonl|\.claude/projects|\.codex/sessions"),
                ("git", r"subprocess|git "), ("http", r"urllib|requests|http\.client"), ("yaml", r"yaml"),
                ("sqlite", r"sqlite"), ("json", r"json\.load|json\.dumps|JSON\.parse"), ("fs", r"pathlib|Path\(|glob|os\.walk|os\.listdir|readFileSync"),
                ("re", r"\bre\.")]


def inline_kind(cmd):
    for k, pat in INLINE_FEATS:
        if re.search(pat, cmd):
            return k
    return "plain"


def shell_ops(cmd):
    ops = []
    for s in split_shell(cmd or ""):
        ops += seg_ops(s)
    ops = [o + ":" + inline_kind(cmd) if o in ("python3 inline", "python inline", "node inline", "perl inline", "ruby inline") else o for o in ops]
    return ops or ["sh:trivial"]


def mcp_key(server, tool):
    if "engram" in server:
        server = "engram"
    server = re.sub(r"^plugin_", "", server)
    server = re.sub(r"^claude_ai_", "", server)
    return f"mcp:{server}.{tool}"


def claude_ops(name, inp):
    if name == "Bash":
        return shell_ops(inp.get("command"))
    if name == "Read":
        return ["read:" + path_class(inp.get("file_path"))]
    if name in ("Edit", "MultiEdit", "Write", "NotebookEdit"):
        return ["edit:" + path_class(inp.get("file_path") or inp.get("notebook_path"))]
    if name in ("Grep", "Glob"):
        return [name]
    if name == "Skill":
        return ["skill:load"]
    if name in ("Agent", "Task"):
        return ["agent:" + str(inp.get("subagent_type") or inp.get("agent") or "")]
    if name in ("TodoWrite", "TaskCreate", "TaskUpdate", "TaskList", "TaskGet", "TaskOutput", "TaskStop"):
        return ["todo"]
    if name == "AskUserQuestion":
        return ["ask"]
    m = re.match(r"mcp__(.+?)__(.+)$", name or "")
    if m:
        return [mcp_key(m.group(1), m.group(2))]
    return [name or "?"]


# ---------------------------------------------------------------- Codex code-mode parsing
JS_STR = r"""("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|`(?:[^`\\]|\\.)*`)"""
TOOL_CALL = re.compile(r"tools\.(\w+)\(")
CMD_FIELD = re.compile(r"\bcmd\s*:\s*" + JS_STR)
ESC = {"n": "\n", "t": "\t", "r": "\r", '"': '"', "'": "'", "`": "`", "\\": "\\", "$": "$"}


def js_unescape(s):
    s = s[1:-1]
    return re.sub(r"\\(.)", lambda m: ESC.get(m.group(1), m.group(1)), s)


def codex_exec_ops(js):
    """Operations of one code-mode `exec`: one entry per tools.* call, in order."""
    ops, cmds = [], []
    ms = list(TOOL_CALL.finditer(js or ""))
    for k, m in enumerate(ms):
        tool = m.group(1)
        end = ms[k + 1].start() if k + 1 < len(ms) else len(js)
        body = js[m.end():end]
        if tool == "exec_command":
            c = CMD_FIELD.search(body)
            if c:
                cmd = js_unescape(c.group(1))
                cmds.append(cmd)
                ops += shell_ops(cmd)
            else:
                ops.append("sh:dynamic")
        elif tool == "apply_patch":
            ops.append("edit:" + path_class((re.search(r"\*\*\* (?:Update|Add|Delete) File: (\S+)", body) or [0, ""])[1]))
        elif tool.startswith("mcp__"):
            mm = re.match(r"mcp__(.+?)__(.+)$", tool)
            ops.append(mcp_key(mm.group(1), mm.group(2)) if mm else "mcp:" + tool)
        elif tool == "write_stdin":
            ops.append("sh:stdin")
        elif tool in ("view_image",):
            ops.append("view_image")
        else:
            ops.append("codex:" + tool)
    if not ops:
        if "ALL_TOOLS" in (js or ""):
            ops.append("toolsearch")
        else:
            ops.append("exec:js")
    return ops, cmds


CODEX_FN = {"spawn_agent": "agent:spawn", "send_message": "agent:msg", "followup_task": "agent:msg", "wait": "agent:wait",
            "wait_agent": "agent:wait", "list_agents": "agent:msg", "sleep": "sleep", "request_user_input": "ask",
            "request_user_input_async": "ask", "js": "exec:js"}


# ---------------------------------------------------------------- loaders
def norm_repo(cwd):
    return WORKTREE.sub("", (cwd or "").rstrip("/")) or "?"


def failed_text(txt, is_error=False):
    if NOT_WORK_FAILURE.search(txt[:400]):
        return False
    if is_error:
        return True
    m = EXIT.search(txt[:300])
    return bool(m and m.group(1) != "0")


def load_claude(path):
    sid = os.path.basename(path)[:-6]
    s = {"host": "claude", "sid": sid, "calls": [], "segments": [0], "total": 0, "cwd": None, "start": None, "end": None}
    seen, pending, order, prev = set(), {}, 0, ""
    last_user = ""
    skill = "none"
    for line in open(path, encoding="utf-8", errors="replace"):
        try:
            e = json.loads(line)
        except ValueError:
            continue
        if e.get("isSidechain"):
            continue
        cur = e.get("timestamp") or prev
        before, prev = prev, cur
        s["cwd"] = s["cwd"] or e.get("cwd")
        if e.get("timestamp"):
            s["start"] = s["start"] or e["timestamp"]
            s["end"] = e["timestamp"]
        t = e.get("type")
        if t == "system" and e.get("subtype") == "compact_boundary":
            s["segments"].append(order)
        m = e.get("message") or {}
        if t == "assistant":
            mid = m.get("id")
            if mid and m.get("usage") and mid not in seen:
                seen.add(mid)
                order += 1
            for c in m.get("content") or []:
                if isinstance(c, dict) and c.get("type") == "tool_use":
                    inp = c.get("input") or {}
                    name = c.get("name")
                    if name == "Skill":
                        skill = str(inp.get("skill") or "none")
                    else:
                        blobp = inp.get("file_path") or inp.get("command") or ""
                        mm = SKILL_PATH.search(str(blobp))
                        if mm and (name == "Read" or (name == "Bash" and re.search(r"\b(cat|head|sed|bat|less)\b", str(blobp)))):
                            skill = mm.group(1)
                    rec = {"idx": order, "name": name, "ops": claude_ops(name, inp), "blob": len(json.dumps(inp, ensure_ascii=False)),
                           "result": 0, "ts": e.get("timestamp") or "", "prev": last_user or before, "res_ts": "", "failed": False, "skill": skill,
                           "cmd": str(inp.get("command") or "")[:4000] if name == "Bash" else ""}
                    s["calls"].append(rec)
                    pending[c.get("id")] = rec
        elif t == "user":
            if e.get("timestamp"):
                last_user = e["timestamp"]
            content = m.get("content")
            if isinstance(content, str):
                mm = re.search(r"<command-name>/?([^<\s]+)</command-name>", content)
                if mm and (mm.group(1).startswith("flow-") or ":" in mm.group(1) or mm.group(1) in ("code-review", "simplify")):
                    skill = mm.group(1)
            elif isinstance(content, list):
                for c in content:
                    if isinstance(c, dict) and c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                        r = pending.pop(c["tool_use_id"])
                        r["result"] = text_len(c.get("content"))
                        r["res_ts"] = e.get("timestamp") or ""
                        r["failed"] = failed_text(text_of(c.get("content"))[:400], bool(c.get("is_error")))
    s["total"] = order
    return s


def load_codex(path):
    s = {"host": "codex", "sid": None, "calls": [], "segments": [0], "total": 0, "cwd": None, "start": None, "end": None}
    pending, n, prev, last_total, skill = {}, 0, "", None, "none"
    with open(path, encoding="utf-8", errors="replace") as f:
        try:
            meta = json.loads(f.readline()).get("payload") or {}
        except ValueError:
            return None
        if meta.get("thread_source") != "user" or str(meta.get("source")) not in ("vscode", "cli", "exec"):
            return None
        s["sid"] = meta.get("id") or os.path.basename(path)
        s["cwd"], s["start"] = meta.get("cwd"), meta.get("timestamp")
        for line in f:
            try:
                e = json.loads(line)
            except ValueError:
                continue
            cur = e.get("timestamp") or prev
            before = prev
            t = e.get("type")
            p = e.get("payload") or {}
            pt = p.get("type")
            if (t == "response_item" and pt in ("function_call_output", "custom_tool_call_output")) or \
               (t == "response_item" and pt == "message" and p.get("role") == "user") or (t == "event_msg" and pt == "task_started"):
                prev = cur
            if e.get("timestamp"):
                s["end"] = e["timestamp"]
            if t == "event_msg" and pt == "token_count":
                info = p.get("info") or {}
                tot = json.dumps(info.get("total_token_usage"))
                if info.get("last_token_usage") and tot != last_total:
                    last_total = tot
                    n += 1
            elif t == "compacted" or (t == "event_msg" and pt == "context_compacted"):
                s["segments"].append(n)
            elif t == "response_item" and pt == "message" and p.get("role") == "user":
                txt = json.dumps(p.get("content"), ensure_ascii=False)
                mm = re.search(r"<skill>\\n<name>([^<]+)</name>", txt)
                if mm:
                    skill = mm.group(1)
            elif t == "response_item" and pt in ("function_call", "custom_tool_call", "local_shell_call"):
                raw = p.get("arguments") or p.get("input") or p.get("action") or ""
                raw = raw if isinstance(raw, str) else json.dumps(raw)
                name = p.get("name") or pt
                if name == "exec":
                    ops, cmds = codex_exec_ops(raw)
                    for c in cmds:
                        mm = SKILL_PATH.search(c)
                        if mm and re.search(r"\b(cat|head|sed|bat)\b", c):
                            skill = mm.group(1)
                else:
                    ops = [CODEX_FN.get(name, "codex:" + str(name))]
                rec = {"idx": n, "name": name, "ops": ops, "blob": len(raw), "result": 0, "ts": e.get("timestamp") or "",
                       "prev": before, "res_ts": "", "failed": False, "skill": skill,
                       "cmd": ("\n;;\n".join(cmds)[:4000] if name == "exec" else "")}
                s["calls"].append(rec)
                pending[p.get("call_id")] = rec
            elif t == "response_item" and pt in ("function_call_output", "custom_tool_call_output") and p.get("call_id") in pending:
                r = pending.pop(p["call_id"])
                out = p.get("output")
                txt = out if isinstance(out, str) else "\n".join((c.get("text") or "") for c in out if isinstance(c, dict)) if isinstance(out, list) else json.dumps(out)
                r["result"] = len(txt)
                r["res_ts"] = e.get("timestamp") or ""
                r["failed"] = (not NOT_WORK_FAILURE.search(txt[:400])) and (bool(re.search(r"\"exit_code\":\s*[1-9-]", txt)) or txt.startswith("Script failed")
                                                                     or bool(re.match(r"\s*Script (failed|error)", txt)))
    s["total"] = n
    return s


def finalize(s):
    """Exclusive seconds per call (no double counting of parallel calls), sizes, carried tokens."""
    calls = s["calls"]
    order = sorted((c for c in calls if c["res_ts"]), key=lambda c: ts_parse(c["res_ts"]) or 0)
    last = None
    for c in calls:
        c["sec"] = 0.0
    for c in calls:
        c["exec"] = 0.0
    for c in order:
        a, b, m = ts_parse(c["prev"]), ts_parse(c["res_ts"]), ts_parse(c["ts"])
        if a is None or b is None:
            continue
        start = a if last is None else max(a, last)
        c["sec"] = min(max(b - start, 0.0), 600.0)
        c["exec"] = min(max(b - (m if m is not None else a), 0.0), c["sec"])
        last = max(last or b, b)
    for c in calls:
        c["tok"] = (c["blob"] + c["result"]) / 4.0
        stop = next((x for x in s["segments"] if x > c["idx"]), s["total"])
        c["carried"] = c["tok"] * max(stop - c["idx"], 0)
    s["repo"] = norm_repo(s["cwd"])
    s["calls"] = calls
    return s


def build(exclude):
    sessions = []
    for f in sorted(glob.glob(f"{HOME}/.claude/projects/*/*.jsonl")):
        if "scratchpad" in f or any(x in f for x in exclude):
            continue
        sessions.append(finalize(load_claude(f)))
    for f in sorted(glob.glob(f"{HOME}/.codex/sessions/2026/1[01]/*/*.jsonl")):
        s = load_codex(f)
        if s and not any(x in (s["sid"] or "") for x in exclude):
            sessions.append(finalize(s))
    sessions = [s for s in sessions if s["calls"]]
    pickle.dump(sessions, open(PKL, "wb"))
    for h in ("claude", "codex"):
        g = [s for s in sessions if s["host"] == h]
        print(h, "sessions", len(g), "calls", sum(len(s["calls"]) for s in g), "from", min(s["start"] for s in g)[:10], "to", max(s["end"] for s in g)[:10])


if __name__ == "__main__":
    a = sys.argv[1:]
    if a and a[0] == "build":
        build([x for x in a[1:] if not x.startswith("--")])
    else:
        import runpy
        runpy.run_path(os.path.join(OUT, "report.py"))
