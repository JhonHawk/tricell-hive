"""Ranks candidate operations/sequences from calls.pkl (built by scriptable_ops.py build)."""
import re, os, pickle, statistics, sys, json, datetime
from collections import Counter, defaultdict

OUT = os.path.dirname(os.path.abspath(__file__)) if "__file__" in globals() else "."
S = pickle.load(open(os.path.join(os.environ.get("TMPDIR", "/tmp"), "scriptable-ops-calls.pkl"), "rb"))
SCRIPT_TOK = (120 + 550) / 4.0   # one script call + a ~550-char summary
SCRIPT_SEC = 4.0
MINSESS = 5

def ts(t):
    return datetime.datetime.fromisoformat(t.replace("Z", "+00:00")).timestamp()
t0 = min(ts(s["start"]) for s in S); t1 = max(ts(s["end"]) for s in S)
WEEKS = (t1 - t0) / 86400 / 7
sdate = {s["sid"]: s["start"][:10] for s in S}

EXPLORE_EXACT = {"rg", "grep", "find", "ls", "tree", "Grep", "Glob", "todo", "ask", "skill:load", "ToolSearch", "toolsearch", "sleep",
                 "sh:stdin", "sh:trivial", "agent-browser", "SendMessage", "exec:js", "ssh", "WebFetch", "WebSearch", "sh:dynamic",
                 "read:src", "read:md", "read:config", "read:data", "read:media", "read:agent-config", "read:tool-results", "wc", "head", "tail",
                 "command -v", "test", "which", "pwd", "open", "view_image", "sh", "xargs", "sort", "uniq", "tee", "awk", "jq", "cut", "tr", "diff", "stat", "file", "date", "mkdir", "rm", "cp", "touch", "chmod", "curl"}
def category(ops):
    ops = ops if isinstance(ops, tuple) else (ops,)
    if any(o.startswith(("read:ref/", "read:SKILL.md", "read:AGENTS.md", "read:memory", "read:agent-config")) for o in ops):
        return "instruction"
    if any(o.startswith("mcp:engram") for o in ops):
        return "memory-mcp"
    if any(o in ("go test", "go vet", "gofmt", "python3 -m unittest", "go build", "pnpm", "npm test", "pnpm test", "go run") or o.startswith(("pnpm ", "npm run", "npx ")) for o in ops):
        return "exec-bound"
    return "candidate"


def excluded(op):
    return op in EXPLORE_EXACT or op.startswith(("edit:", "agent:", "codex:", "playwright", "mcp:playwright", "mcp:chrome", "mcp:computer"))

def stat(xs):
    xs = sorted(xs)
    if not xs: return (0, 0)
    return (statistics.median(xs), xs[int(0.9 * (len(xs) - 1))])

# ------------------------------------------------ flatten
def flat(s):
    out = []
    for ci, c in enumerate(s["calls"]):
        k = len(c["ops"])
        for o in c["ops"]:
            out.append((o, ci, 1.0 / k))
    return out

def call_flags(s):
    """repeat flag per call: same op set appeared within the previous 3 calls."""
    rep = []
    for i, c in enumerate(s["calls"]):
        prev = [set(s["calls"][j]["ops"]) for j in range(max(0, i - 3), i)]
        rep.append(any(set(c["ops"]) & p for p in prev))
    return rep

def occ_stats(rows):
    toks = [r["tok"] for r in rows]; car = [r["carried"] for r in rows]; sec = [r["sec"] for r in rows]
    net_t = [max(r["tok"] - SCRIPT_TOK, 0) for r in rows]
    net_s = [max(r["sec"] - max(SCRIPT_SEC, r["exec"]), 0) for r in rows]
    n = len(rows)
    sess = {r["sid"] for r in rows}
    ev = {}
    for r in sorted(rows, key=lambda r: r["date"]):
        ev.setdefault(r["sid"], r["date"])
    keys = list(ev.items())
    pick = [keys[0], keys[len(keys) // 2], keys[-1]] if len(keys) >= 3 else keys
    skills = Counter(r["skill"] for r in rows).most_common(3)
    return {"n": n, "sessions": len(sess), "repos": len({r["repo"] for r in rows}),
            "codex_share": round(sum(r["host"] == "codex" for r in rows) / n, 2),
            "tok_med": stat(toks)[0], "tok_p90": stat(toks)[1], "carried_med": stat(car)[0],
            "sec_med": stat(sec)[0], "sec_p90": stat(sec)[1],
            "fail": round(sum(r["failed"] for r in rows) / n, 2), "repeat": round(sum(r["repeat"] for r in rows) / n, 2),
            "net_tok_occ": statistics.median(net_t), "net_sec_occ": statistics.median(net_s),
            "week_tok": sum(net_t) / WEEKS, "week_sec": sum(net_s) / WEEKS, "week_n": n / WEEKS,
            "week_carried": sum(car) / WEEKS,
            "skills": skills, "evidence": [(a[:8], b) for a, b in pick]}

# ------------------------------------------------ single ops
single = defaultdict(list)
ngram_rows = defaultdict(list)
for s in S:
    f = flat(s); rep = call_flags(s); calls = s["calls"]
    for o, ci, share in f:
        c = calls[ci]
        single[o].append({"sid": s["sid"], "date": s["start"][:10], "repo": s["repo"], "host": s["host"], "skill": c["skill"],
                          "tok": c["tok"] * share, "carried": c["carried"] * share, "sec": c["sec"] * share, "exec": c["exec"] * share,
                          "failed": c["failed"], "repeat": rep[ci]})
    # n-grams
    L = len(f)
    for n in range(2, 7):
        last_end = {}
        for i in range(L - n + 1):
            seg = f[i:i + n]
            ops = tuple(x[0] for x in seg)
            if any(excluded(o) for o in ops) or len(set(ops)) == 1:
                continue
            sk = {calls[x[1]]["skill"] for x in seg}
            if len(sk) != 1:
                continue
            if last_end.get(ops, -1) >= i:
                continue
            last_end[ops] = i + n - 1
            cis = sorted({x[1] for x in seg})
            # prorate: sum of call share over entries in the n-gram
            tok = sum(calls[x[1]]["tok"] * x[2] for x in seg)
            car = sum(calls[x[1]]["carried"] * x[2] for x in seg)
            sec = sum(calls[x[1]]["sec"] * x[2] for x in seg)
            ngram_rows[ops].append({"sid": s["sid"], "date": s["start"][:10], "repo": s["repo"], "host": s["host"], "skill": sk.pop(),
                                    "tok": tok, "carried": car, "sec": sec, "exec": sum(calls[x[1]]["exec"] * x[2] for x in seg),
                                    "failed": any(calls[j]["failed"] for j in cis), "repeat": any(rep[j] for j in cis)})

def sub(a, b):
    n, m = len(a), len(b)
    return any(a[i:i + m] == b for i in range(n - m + 1))

seqs = []
for ops, rows in ngram_rows.items():
    if len({r["sid"] for r in rows}) >= MINSESS:
        st = occ_stats(rows); st["ops"] = ops; seqs.append(st)

def dedupe(cands, key):
    acc = []
    for c in sorted(cands, key=lambda c: -c[key]):
        if any(sub(c["ops"], a["ops"]) or sub(a["ops"], c["ops"]) for a in acc):
            continue
        acc.append(c)
    return acc

def fmt(c, sec=False):
    return (f'{" > ".join(c["ops"]) if isinstance(c["ops"], tuple) else c["ops"]}\n'
            f'   occ={c["n"]} ({c["week_n"]:.1f}/wk) sess={c["sessions"]} repos={c["repos"]} codex={c["codex_share"]} skills={c["skills"]}\n'
            f'   ctx med={c["tok_med"]:.0f} p90={c["tok_p90"]:.0f} carried med={c["carried_med"]:.0f} | sec med={c["sec_med"]:.1f} p90={c["sec_p90"]:.1f} | fail={c["fail"]} repeat={c["repeat"]}\n'
            f'   net/occ tok={c["net_tok_occ"]:.0f} sec={c["net_sec_occ"]:.1f} | WEEK tok={c["week_tok"]:.0f} sec={c["week_sec"]:.0f} carried={c["week_carried"]:.0f}\n'
            f'   evidence={c["evidence"]}')

singles = []
for o, rows in single.items():
    if len({r["sid"] for r in rows}) >= MINSESS:
        st = occ_stats(rows); st["ops"] = o; st["excluded"] = excluded(o); singles.append(st)

def line(i, c):
    ops = " > ".join(c["ops"]) if isinstance(c["ops"], tuple) else c["ops"]
    sk = ",".join(f"{a}:{b}" for a, b in c["skills"])
    ev = ";".join(f"{a}@{b[5:]}" for a, b in c["evidence"])
    return (f'{i:>2}. {ops}\n    occ={c["n"]}({c["week_n"]:.0f}/wk) sess={c["sessions"]} repos={c["repos"]} cx={c["codex_share"]} | ctx med={c["tok_med"]:.0f} p90={c["tok_p90"]:.0f} car={c["carried_med"]:.0f} '
            f'| s med={c["sec_med"]:.1f} p90={c["sec_p90"]:.1f} | fail={c["fail"]} rep={c["repeat"]} | net/occ {c["net_tok_occ"]:.0f}t {c["net_sec_occ"]:.1f}s | WK {c["week_tok"]:.0f}t {c["week_sec"]:.0f}s\n    {sk} | {ev}')


ARGS = sys.argv[2:]
TOPN = 18


def main():
    for h in ("claude", "codex"):
        g = [s for s in S if s["host"] == h]
        print(f"{h}: sessions={len(g)} calls={sum(len(s['calls']) for s in g)} ops={sum(len(c['ops']) for s in g for c in s['calls'])} "
              f"repos={len({s['repo'] for s in g})} {min(s['start'] for s in g)[:10]}..{max(s['end'] for s in g)[:10]}")
    print(f"weeks={WEEKS:.2f} script_cost={SCRIPT_TOK:.0f} tok / {SCRIPT_SEC}s")
    print(f"total tool tokens={sum(c['tok'] for s in S for c in s['calls']):.0f} total tool sec={sum(c['sec'] for s in S for c in s['calls']):.0f}")
    pool = [dict(c, cat=category(c["ops"]), kind="seq") for c in dedupe(seqs, "week_tok")] + \
           [dict(c, cat=category(c["ops"]), kind="single") for c in singles if not c["excluded"]]
    for cat in ("candidate", "exec-bound", "memory-mcp", "instruction"):
        P = [c for c in pool if c["cat"] == cat]
        print(f"\n######## {cat}: top by weekly context saving")
        for i, c in enumerate(sorted(P, key=lambda c: -c["week_tok"])[:TOPN], 1): print(line(i, c) + f'  [{c["kind"]}]')
        print(f"\n######## {cat}: top by weekly time saving")
        for i, c in enumerate(sorted(P, key=lambda c: -c["week_sec"])[:TOPN], 1): print(line(i, c) + f'  [{c["kind"]}]')
    print("\n######## script reference")
    for c in singles:
        if c["ops"].startswith(("script:apply_deltas", "script:check_change", "script:archive_sessions", "script:init_workspace", "script:self_check",
                                "script:verify_geometry", "script:contrast-check", "script:measure")) or c["ops"].startswith("hive "):
            print(line(0, c))


def seq_view(n):
    SQ = [c for c in dedupe(seqs, "week_tok") if category(c["ops"]) == "candidate"]
    print("#### SEQ candidates by weekly tok")
    for i, c in enumerate(sorted(SQ, key=lambda c: -c["week_tok"])[:n], 1): print(line(i, c))
    SQ2 = [c for c in dedupe(seqs, "week_sec") if category(c["ops"]) == "candidate"]
    print("#### SEQ candidates by weekly sec")
    for i, c in enumerate(sorted(SQ2, key=lambda c: -c["week_sec"])[:n], 1): print(line(i, c))



GIT_W = {"git add", "git commit", "git push", "git switch", "git checkout", "git merge", "git rebase", "git cherry-pick", "git worktree add", "git worktree remove",
         "git worktree prune", "git stash push", "git stash pop", "git reset", "git restore", "git rm", "git mv", "git tag", "git pull", "git clean", "git branch -d", "git commit-tree", "git update-ref", "git apply", "git am"}
GH_W = {"gh pr merge", "gh pr create", "gh pr ready", "gh pr comment", "gh pr edit", "gh pr close", "gh pr review", "gh issue create", "gh issue comment", "gh issue close",
        "gh issue edit", "gh release create", "gh run rerun", "gh repo create", "gh label"}
TESTB = {"go test", "go vet", "gofmt", "go build", "python3 -m unittest", "go run", "pnpm", "npm test"}


GIT_CONTENT = {"git diff", "git show", "git grep", "git blame", "git cat-file", "git ls-tree", "git diff-tree", "git difftool"}
PURPOSE = [("transcripts", r"jsonl|\.claude/projects|\.codex/sessions"), ("viewer", r"hive view|viewer|stage\.start|session\.create"),
           ("tasks.md", r"tasks\.md|\[ \]|\[x\]"), ("openspec-other", r"proposal\.md|design\.md|validation\.md|openspec/specs|changes/archive"),
           ("git-gh-state", r"git |\bgh\b|api\.github|merge-base|rev-parse"), ("html-browser", r"\.html|<svg|\.css|querySelector|getBoundingClientRect|contrast|luminance|page\."),
           ("fs-regex", r"re\.sub|\.replace\(|write_text|glob|pathlib|os\.walk|listdir|readdirSync"), ("json-config", r"yaml|toml|json")]
PURPOSE = [(n, re.compile(p, re.I)) for n, p in PURPOSE]


def purpose(cmd):
    for n, p in PURPOSE:
        if p.search(cmd or ""):
            return n
    return "other"


def fam(op, cmd=""):
    if op.startswith("git "): return "git-write" if op in GIT_W else ("git-content" if op in GIT_CONTENT else "git-probe")
    if op == "git": return "git-probe"
    if op.startswith("gh "): return "gh-write" if op in GH_W else "gh-read"
    if op.startswith("hive view"): return "hive-view"
    if op.startswith("hive"): return "hive-other"
    if op.split(":")[0] in ("python3 inline", "node inline", "perl inline", "ruby inline", "python inline"): return "inline:" + purpose(cmd)
    if op.startswith("script:"): return "named-script"
    if op.startswith(("mcp:Linear", "acli")): return "tracker"
    if op.startswith("mcp:engram"): return "memory-mcp"
    if op.startswith(("read:openspec", "mv:openspec")): return "openspec-read"
    if op.startswith("edit:openspec"): return "openspec-edit"
    if op.startswith(("read:ref/", "read:SKILL.md", "read:AGENTS.md", "read:memory")): return "instruction-read"
    if op in TESTB or op.startswith(("pnpm ", "npm run", "npm test", "npx ", "go ")): return "test-build"
    if op in ("rg", "grep", "find", "ls", "Grep", "Glob", "tree", "wc") or op.startswith(("read:", "codegraph", "mcp:codegraph", "ctx7", "npx ctx7")): return "explore-read"
    if op.startswith(("agent-browser", "mcp:playwright", "mcp:chrome")): return "browser"
    if op.startswith("edit:"): return "edit"
    if op.startswith("agent:"): return "delegation"
    return "other"


def retry_flags(s):
    """strict retry: an earlier call of the previous 3 failed and shares an op key, or the command text is identical."""
    fl = []
    for i, c in enumerate(s["calls"]):
        r = False
        for j in range(max(0, i - 3), i):
            p = s["calls"][j]
            if (p["failed"] and set(p["ops"]) & set(c["ops"])) or (c.get("cmd") and c["cmd"] == p.get("cmd")):
                r = True
        fl.append(r)
    return fl


def windows():
    out = []
    for s in S:
        cur, last_end, key = None, None, None
        rf = retry_flags(s)
        for ci, c in enumerate(s["calls"]):
            a = ts(c["prev"]) if c["prev"] else None
            gap = (a - last_end) if (a is not None and last_end is not None) else 0
            if cur is None or c["skill"] != cur["skill"] or gap > 1800:
                cur = {"sid": s["sid"], "date": s["start"][:10], "repo": s["repo"], "host": s["host"], "skill": c["skill"], "calls": [], "fam": defaultdict(lambda: {"n": 0, "tok": 0.0, "car": 0.0, "sec": 0.0, "exec": 0.0, "failed": 0, "retry": 0, "calls": set()})}
                out.append(cur)
            cur["calls"].append(c)
            k = len(c["ops"])
            for o in c["ops"]:
                f = cur["fam"][fam(o, c.get("cmd", ""))]
                f["n"] += 1; f["tok"] += c["tok"] / k; f["car"] += c["carried"] / k; f["sec"] += c["sec"] / k; f["exec"] += c["exec"] / k
                f["failed"] += c["failed"]; f["retry"] += rf[ci]; f["calls"].add(ci)
            r = ts(c["res_ts"]) if c["res_ts"] else None
            if r is not None: last_end = r
    return out


def fam_view(n, skills_only=True):
    W = windows()
    agg = defaultdict(list)
    for w in W:
        for f, v in w["fam"].items():
            agg[(w["skill"], f)].append((w, v))
    rows = []
    for (sk, f), items in agg.items():
        if f in ("explore-read", "edit", "delegation", "other", "browser", "instruction-read", "openspec-edit", "memory-mcp", "git-write", "gh-write", "test-build", "tracker"): continue
        if len({w["sid"] for w, _ in items}) < MINSESS: continue
        toks = [v["tok"] for _, v in items]; secs = [v["sec"] for _, v in items]; ex = [v["exec"] for _, v in items]
        calls = [len(v["calls"]) for _, v in items]
        net_t = [max(t - SCRIPT_TOK, 0) for t in toks]; net_s = [max(sc - max(SCRIPT_SEC, e), 0) for sc, e in zip(secs, ex)]
        ev = {}
        for w, _ in sorted(items, key=lambda x: x[0]["date"]): ev.setdefault(w["sid"], w["date"])
        keys = list(ev.items()); pick = [keys[0], keys[len(keys) // 2], keys[-1]] if len(keys) >= 3 else keys
        rows.append({"skill": sk, "fam": f, "windows": len(items), "sessions": len({w["sid"] for w, _ in items}), "repos": len({w["repo"] for w, _ in items}),
                     "codex": round(sum(w["host"] == "codex" for w, _ in items) / len(items), 2),
                     "calls_med": statistics.median(calls), "tok_med": stat(toks)[0], "tok_p90": stat(toks)[1], "car_med": stat([v["car"] for _, v in items])[0],
                     "sec_med": stat(secs)[0], "sec_p90": stat(secs)[1], "exec_med": stat(ex)[0],
                     "fail": round(sum(v["failed"] > 0 for _, v in items) / len(items), 2),
                     "net_tok": statistics.median(net_t), "net_sec": statistics.median(net_s), "wk_tok": sum(net_t) / WEEKS, "wk_sec": sum(net_s) / WEEKS,
                     "ev": ";".join(f"{a[:8]}@{b[5:]}" for a, b in pick)})
    def pr(r, i):
        print(f'{i:>2}. {r["skill"]} / {r["fam"]}: windows={r["windows"]}({r["windows"]/WEEKS:.1f}/wk) sess={r["sessions"]} repos={r["repos"]} cx={r["codex"]} calls/win med={r["calls_med"]:.0f} | '
              f'ctx med={r["tok_med"]:.0f} p90={r["tok_p90"]:.0f} car={r["car_med"]:.0f} | s med={r["sec_med"]:.0f} p90={r["sec_p90"]:.0f} exec med={r["exec_med"]:.0f} | fail={r["fail"]} | '
              f'net/win {r["net_tok"]:.0f}t {r["net_sec"]:.0f}s | WK {r["wk_tok"]:.0f}t {r["wk_sec"]:.0f}s | {r["ev"]}')
    print("#### skill x family windows by weekly tok")
    for i, r in enumerate(sorted(rows, key=lambda r: -r["wk_tok"])[:n], 1): pr(r, i)
    print("#### skill x family windows by weekly sec")
    for i, r in enumerate(sorted(rows, key=lambda r: -r["wk_sec"])[:n], 1): pr(r, i)


def bundle_view(n):
    """Unordered bundles: consecutive windows of 3..6 operations compared as sorted multisets (order-insensitive)."""
    B = defaultdict(list)
    for s in S:
        f = flat(s); rep = call_flags(s); calls = s["calls"]; L = len(f)
        for k in range(3, 7):
            seen_end = {}
            for i in range(L - k + 1):
                seg = f[i:i + k]
                ops = tuple(sorted(x[0] for x in seg))
                if any(excluded(o) for o in ops) or len(set(ops)) < 3: continue
                sk = {calls[x[1]]["skill"] for x in seg}
                if len(sk) != 1: continue
                if seen_end.get(ops, -1) >= i: continue
                seen_end[ops] = i + k - 1
                cis = sorted({x[1] for x in seg})
                B[ops].append({"sid": s["sid"], "date": s["start"][:10], "repo": s["repo"], "host": s["host"], "skill": sk.pop(),
                               "tok": sum(calls[x[1]]["tok"] * x[2] for x in seg), "carried": sum(calls[x[1]]["carried"] * x[2] for x in seg),
                               "sec": sum(calls[x[1]]["sec"] * x[2] for x in seg), "exec": sum(calls[x[1]]["exec"] * x[2] for x in seg),
                               "failed": any(calls[j]["failed"] for j in cis), "repeat": any(rep[j] for j in cis)})
    rows = []
    for ops, r in B.items():
        if len({x["sid"] for x in r}) >= MINSESS:
            st = occ_stats(r); st["ops"] = ops; rows.append(st)
    rows = [c for c in dedupe(rows, "week_tok") if category(c["ops"]) == "candidate"]
    print("#### unordered bundles by weekly tok")
    for i, c in enumerate(sorted(rows, key=lambda c: -c["week_tok"])[:n], 1): print(line(i, c))
    print("#### unordered bundles by weekly sec")
    for i, c in enumerate(sorted(rows, key=lambda c: -c["week_sec"])[:n], 1): print(line(i, c))


def famall(n):
    W = windows()
    tot = defaultdict(lambda: {"calls": 0, "tok": 0.0, "sec": 0.0, "exec": 0.0})
    for w in W:
        for f, v in w["fam"].items():
            t = tot[f]; t["calls"] += len(v["calls"]); t["tok"] += v["tok"]; t["sec"] += v["sec"]; t["exec"] += v["exec"]
    print("#### family totals over all windows (per week): calls, tokens, hours of wall time")
    for f, t in sorted(tot.items(), key=lambda kv: -kv[1]["tok"]):
        print(f'{f:24} calls/wk={t["calls"]/WEEKS:8.0f}  tok/wk={t["tok"]/WEEKS:10.0f}  sec/wk={t["sec"]/WEEKS:8.0f}  exec/wk={t["exec"]/WEEKS:7.0f}')
    # per family across skills
    agg = defaultdict(list)
    for w in W:
        for f, v in w["fam"].items():
            agg[f].append((w, v))
    print("\n#### candidate families (all skills pooled, one script call per window)")
    for f, items in sorted(agg.items()):
        if f in ("explore-read", "edit", "delegation", "other", "browser", "instruction-read", "openspec-edit", "memory-mcp", "git-write", "gh-write", "test-build", "tracker"): continue
        net_t = [max(v["tok"] - SCRIPT_TOK, 0) for _, v in items]; net_s = [max(v["sec"] - max(SCRIPT_SEC, v["exec"]), 0) for _, v in items]
        toks = [v["tok"] for _, v in items]; secs = [v["sec"] for _, v in items]
        by = defaultdict(lambda: [0, 0.0, 0.0])
        for (w, _), a, b in zip(items, net_t, net_s):
            by[w["skill"]][0] += 1; by[w["skill"]][1] += a; by[w["skill"]][2] += b
        top = sorted(by.items(), key=lambda kv: -kv[1][1])[:4]
        c3 = [(a_, b_) for (w, v), a_, b_ in zip(items, net_t, net_s) if len(v["calls"]) >= 3]
        ev = {}
        for (w, v), a_ in sorted(zip(items, net_t), key=lambda x: -x[1]):
            ev.setdefault(w["sid"], w["date"])
        evs = ";".join(f"{k[:8]}@{d[5:]}" for k, d in list(ev.items())[:3])
        print(f'{f:22} windows={len(items)}({len(items)/WEEKS:.0f}/wk) sess={len({w["sid"] for w,_ in items})} repos={len({w["repo"] for w,_ in items})} cx={sum(w["host"]=="codex" for w,_ in items)/len(items):.2f} '
              f'calls/win med={statistics.median(len(v["calls"]) for _, v in items):.0f} ctx med={stat(toks)[0]:.0f} p90={stat(toks)[1]:.0f} s med={stat(secs)[0]:.0f} p90={stat(secs)[1]:.0f} '
              f'fail={sum(v["failed"]>0 for _,v in items)/len(items):.2f} retry={sum(v["retry"]>0 for _,v in items)/len(items):.2f} '
              f'car med={stat([v["car"] for _, v in items])[0]:.0f} WKcar={sum(v["car"] for _, v in items)/WEEKS:.0f} WK tok={sum(net_t)/WEEKS:.0f} sec={sum(net_s)/WEEKS:.0f} | >=3calls: {len(c3)}w WK tok={sum(x for x,_ in c3)/WEEKS:.0f} sec={sum(y for _,y in c3)/WEEKS:.0f}\n'
              f'    skills: ' + "; ".join(f"{k}:{v[0]}w,{v[1]/WEEKS:.0f}t,{v[2]/WEEKS:.0f}s" for k, v in top) + f'\n    evidence: {evs}')


if ARGS[:1] == ["famall"]:
    famall(30)
elif ARGS[:1] == ["seq"]:
    seq_view(int(ARGS[1]) if len(ARGS) > 1 else 25)
elif ARGS[:1] == ["fam"]:
    fam_view(int(ARGS[1]) if len(ARGS) > 1 else 30)
elif ARGS[:1] == ["bundle"]:
    bundle_view(int(ARGS[1]) if len(ARGS) > 1 else 20)
else:
    main()
