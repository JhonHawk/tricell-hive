"""Fingerprint what agent-written inline scripts (python3/node/perl heredocs and -c) do; prints tag sets with counts and costs only."""
import pickle, re, statistics, os
from collections import defaultdict, Counter
OUT = os.path.dirname(os.path.abspath(__file__))
S = pickle.load(open(os.path.join(os.environ.get("TMPDIR", "/tmp"), "scriptable-ops-calls.pkl"), "rb"))
TAGS = [
 ("tasks.md-checkboxes", r"tasks\.md|\[ \]|\[x\]"), ("proposal/status", r"proposal\.md|Status:|\*\*Status"), ("openspec-specs", r"openspec/specs|/specs/"),
 ("change-archive", r"changes/archive|archive"), ("design/validation.md", r"design\.md|validation\.md"),
 ("git-log/diff/status", r"git log|git diff|git status|merge-base|rev-parse|ls-files|rev-list|git show|git branch|git worktree|git fetch|cherry"),
 ("gh/github", r"\bgh\b|api\.github|pulls|check-runs"), ("hive-view/viewer", r"hive view|viewer|stage\.start|session\.create"),
 ("transcript-jsonl", r"jsonl|\.claude/projects|\.codex/sessions"), ("engram", r"engram"), ("sqlite", r"sqlite"),
 ("stdin-json-parse", r"sys\.stdin|json\.load\(sys|JSON\.parse\(.*stdin|readFileSync\(0"), ("AGENTS/Hive-settings", r"AGENTS\.md|## Hive"),
 ("hash", r"hashlib|sha256|md5"), ("http", r"urllib|requests\.|http\.client|fetch\("), ("html/svg/css", r"\.html|<svg|\.css|querySelector|getBoundingClientRect|contrast|luminance"),
 ("regex-text-edit", r"re\.sub|\.replace\(|write_text|\.write\("), ("fs-walk/glob", r"os\.walk|glob|pathlib|listdir|rglob|readdirSync"),
 ("skills-files", r"SKILL\.md|/skills/|references/"), ("size/len-count", r"len\(|wc -|\.length|getsize|st_size"), ("yaml/toml/json-config", r"yaml|toml|package\.json|\.json"),
 ("linear/tracker", r"linear|jira|acli"), ("time/date", r"datetime|strftime|date\(|time\.time"),
]
pat = [(n, re.compile(p, re.I)) for n, p in TAGS]
rows = defaultdict(list)
for s in S:
    for c in s["calls"]:
        if not any(o.split(":")[0] in ("python3 inline", "node inline", "perl inline", "python inline", "ruby inline") for o in c["ops"]):
            continue
        txt = c.get("cmd", "")
        tags = [n for n, p in pat if p.search(txt)]
        # keep the 3 most specific (earlier in TAGS = more specific)
        sig = tuple(tags[:3]) or ("(none)",)
        rows[sig].append((s["sid"], s["start"][:10], c["skill"], c["tok"], c["sec"], c["failed"], s["repo"], s["host"]))
tot = sum(len(v) for v in rows.values())
print("inline-script calls:", tot)
for sig, v in sorted(rows.items(), key=lambda kv: -len(kv[1]))[:30]:
    sess = {x[0] for x in v}
    print(f'{len(v):5} calls  {len(sess):3} sess  {len({x[6] for x in v})} repos  tok med={statistics.median(x[3] for x in v):.0f}  sec med={statistics.median(x[4] for x in v):.1f}  fail={sum(x[5] for x in v)/len(v):.2f}  top skills={Counter(x[2] for x in v).most_common(2)}  {" + ".join(sig)}')
