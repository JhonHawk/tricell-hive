#!/usr/bin/env python3
"""Measure how a real agent session delegates, against the delegation criterion
of content/skills/flow-build/references/delegation.md (PR #352).

Read-only. Takes Codex rollouts (~/.codex/sessions/**/rollout-*.jsonl) and
Claude Code main transcripts (~/.claude/projects/<project>/<session>.jsonl),
by path or by a session ID prefix, and prints per session:

- the main thread: model calls, processed input tokens (input + cache reads +
  cache writes per call), cached share, peak context, compactions, and its
  large reads (tool outputs over 100 lines);
- each child: role, calls, first and last input, tokens gathered (last - first),
  processed tokens, duration, report size, main-thread calls after it, and the
  inline-versus-delegated estimate (gathered x later calls, against child
  processed + report x later calls);
- flags against the rules: review roles, reviewer briefs that carry the
  producer's framing (Claude Code only; Codex briefs are encrypted), small
  non-review children (gathered < 10k tokens), writing children that ran at
  the same time.

Usage:
  python3 -I measure-delegation.py <path-or-id-prefix> [...] [--json]
"""
import glob
import json
import os
import re
import sys
from datetime import datetime

CODEX_ROOT = os.path.expanduser("~/.codex/sessions")
CLAUDE_ROOT = os.path.expanduser("~/.claude/projects")
LARGE_LINES = 100
SMALL_GATHERED = 10_000
REVIEW_ROLES = re.compile(r"^hive-(review|verify)-")
WRITE_ROLES = re.compile(r"^hive-(build|write)-")
BIAS = re.compile(
    r"\b(already verified|verified that|tests? pass(es|ed)?|all green|works correctly|"
    r"is correct|I (implemented|fixed|changed|chose|decided)|because I|my reasoning)\b",
    re.I,
)


def ts(value):
    if not value:
        return None
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


def minutes(a, b):
    return round((ts(b) - ts(a)).total_seconds() / 60, 1) if a and b else None


def lines_of(text):
    # Codex nests tool output in JSON, where newlines arrive escaped.
    return text.count("\n") + text.count("\\n") + 1 if text else 0


# ---------- Codex ----------

def codex_load(path):
    meta, calls, outs, final, compactions = None, [], [], "", []
    first = last = None
    for raw in open(path, encoding="utf-8"):
        d = json.loads(raw)
        t, p, when = d.get("type"), d.get("payload", {}), d.get("timestamp")
        first, last = first or when, when or last
        if t == "session_meta":
            meta = p
        elif t == "compacted":
            compactions.append(when)
        elif t == "event_msg" and p.get("type") == "token_count" and p.get("info"):
            u = p["info"]["last_token_usage"]
            calls.append((when, u.get("input_tokens", 0), u.get("cached_input_tokens", 0)))
        elif t == "response_item" and p.get("type") in ("custom_tool_call_output", "function_call_output"):
            out = p.get("output", "")
            text = out if isinstance(out, str) else json.dumps(out)
            if "data:image" not in text:
                outs.append((when, lines_of(text), len(text) // 4))
        elif t == "response_item" and p.get("type") == "message" and p.get("role") == "assistant":
            final = " ".join(x.get("text", "") for x in p.get("content", []) if isinstance(x, dict))
    return dict(meta=meta or {}, calls=calls, outs=outs, final=final, first=first, last=last,
                compactions=compactions, briefs={})


def codex_children(meta):
    sid, kids = meta.get("id"), []
    if not sid:
        return kids
    for path in glob.glob(os.path.join(CODEX_ROOT, "**", "rollout-*.jsonl"), recursive=True):
        try:
            with open(path, encoding="utf-8") as fh:
                m = json.loads(fh.readline()).get("payload", {})
        except (OSError, ValueError):
            continue
        if m.get("parent_thread_id") == sid:
            s = codex_load(path)
            s["role"] = m.get("agent_role") or "?"
            s["name"] = m.get("agent_path") or os.path.basename(path)
            s["brief"] = None  # encrypted in Codex rollouts
            kids.append(s)
    return kids


# ---------- Claude Code ----------

def claude_load(path):
    calls, outs, final, compactions, briefs = [], [], "", [], {}
    first = last = None
    counted = set()  # one transcript line per content block repeats the call's usage
    for raw in open(path, encoding="utf-8"):
        d = json.loads(raw)
        when = d.get("timestamp")
        first, last = first or when, when or last
        if d.get("type") == "system" and "compact" in str(d.get("subtype", "")):
            compactions.append(when)
        msg = d.get("message") or {}
        content = msg.get("content")
        if d.get("type") == "assistant":
            u = msg.get("usage") or {}
            total = u.get("input_tokens", 0) + u.get("cache_read_input_tokens", 0) + u.get("cache_creation_input_tokens", 0)
            call_id = msg.get("id")
            if total and (call_id is None or call_id not in counted):
                counted.add(call_id)
                calls.append((when, total, u.get("cache_read_input_tokens", 0)))
            for block in content if isinstance(content, list) else []:
                if block.get("type") == "text":
                    final = block.get("text", "")
                if block.get("type") == "tool_use" and block.get("name") in ("Agent", "Task"):
                    inp = block.get("input", {})
                    briefs[block.get("id")] = (inp.get("subagent_type", "?"), inp.get("prompt", ""))
        elif d.get("type") == "user" and isinstance(content, list):
            for block in content:
                if block.get("type") == "tool_result":
                    c = block.get("content")
                    text = c if isinstance(c, str) else " ".join(x.get("text", "") for x in c or [] if isinstance(x, dict))
                    outs.append((when, lines_of(text), len(text) // 4))
    return dict(meta={}, calls=calls, outs=outs, final=final, first=first, last=last,
                compactions=compactions, briefs=briefs)


def claude_children(path, main):
    kids = []
    folder = os.path.join(os.path.splitext(path)[0], "subagents")
    for sub in sorted(glob.glob(os.path.join(folder, "agent-*.jsonl"))):
        s = claude_load(sub)
        meta_path = sub[: -len(".jsonl")] + ".meta.json"
        role, tool_id = "?", None
        skill_path = sub[: -len(".jsonl")] + ".forked-skill.json"
        if os.path.exists(skill_path):
            role = "skill:" + json.load(open(skill_path, encoding="utf-8")).get("skillName", "?")
        elif os.path.exists(meta_path):
            m = json.load(open(meta_path, encoding="utf-8"))
            role, tool_id = m.get("agentType", "?"), m.get("toolUseId")
        s["role"], s["name"] = role, os.path.basename(sub)[:24]
        s["brief"] = main["briefs"].get(tool_id, (None, None))[1] if tool_id else None
        kids.append(s)
    return kids


# ---------- report ----------

def summarize(label, host, main, kids):
    calls = main["calls"]
    processed = sum(c[1] for c in calls)
    cached = sum(c[2] for c in calls)
    large = [o for o in main["outs"] if o[1] > LARGE_LINES]
    rows, flags = [], []
    for k in kids:
        kc = k["calls"]
        if not kc:
            continue
        gathered = kc[-1][1] - kc[0][1]
        kp = sum(c[1] for c in kc)
        report = len(k["final"]) // 4
        cut = min((x for x in main["compactions"] if x and k["last"] and x > k["last"]), default=None)
        after = sum(1 for c in calls if c[0] and k["last"] and c[0] > k["last"] and (cut is None or c[0] < cut))
        rows.append(dict(name=k["name"], role=k["role"], calls=len(kc), first_in=kc[0][1], gathered=gathered,
                         processed=kp, minutes=minutes(k["first"], k["last"]), report=report, main_calls_after=after,
                         inline_estimate=gathered * after, delegated_estimate=kp + report * after,
                         start=k["first"], end=k["last"]))
        review = bool(REVIEW_ROLES.match(k["role"])) or k["role"].startswith("skill:code-review")
        if not review and gathered < SMALL_GATHERED:
            flags.append(f"small non-review child: {k['name']} ({k['role']}) gathered {gathered} tokens")
        if REVIEW_ROLES.match(k["role"]) and k.get("brief"):
            hits = sorted({m.group(0) for m in BIAS.finditer(k["brief"])})
            if hits:
                flags.append(f"reviewer brief framing in {k['name']} ({k['role']}): {', '.join(hits[:5])}")
    writers = [r for r in rows if WRITE_ROLES.match(r["role"])]
    for i, a in enumerate(writers):
        for b in writers[i + 1:]:
            if a["start"] and b["start"] and a["start"] < b["end"] and b["start"] < a["end"]:
                flags.append(f"writing children overlapped: {a['name']} and {b['name']} (check contract and integration check)")
    reviews = sorted({r["role"] for r in rows if REVIEW_ROLES.match(r["role"]) or r["role"].startswith("skill:")})
    return dict(session=label, host=host, main=dict(calls=len(calls), processed=processed,
                cached_share=round(cached / processed, 3) if processed else 0,
                peak_context=max((c[1] for c in calls), default=0), compactions=len(main["compactions"]),
                large_reads=len(large), large_read_tokens=sum(o[2] for o in large),
                minutes=minutes(main["first"], main["last"])),
                children=rows, review_roles=reviews, flags=flags)


def resolve(arg):
    if os.path.exists(arg):
        return arg
    hits = glob.glob(os.path.join(CODEX_ROOT, "**", f"rollout-*{arg}*.jsonl"), recursive=True)
    hits += glob.glob(os.path.join(CLAUDE_ROOT, "*", f"{arg}*.jsonl"))
    return sorted(hits)[0] if hits else None


def measure(path):
    if CODEX_ROOT in os.path.abspath(path):
        main = codex_load(path)
        return summarize(main["meta"].get("id", path), "codex", main, codex_children(main["meta"]))
    main = claude_load(path)
    return summarize(os.path.basename(path)[:-6], "claude-code", main, claude_children(path, main))


def show(r):
    m = r["main"]
    print(f"\n## {r['host']} {r['session']}")
    print(f"main: calls={m['calls']} processed={m['processed']/1e6:.2f}M cached={m['cached_share']:.0%} "
          f"peak={m['peak_context']/1e3:.0f}k compactions={m['compactions']} minutes={m['minutes']} "
          f"large_reads={m['large_reads']} ({m['large_read_tokens']/1e3:.0f}k tokens)")
    for c in r["children"]:
        print(f"  {c['role']:22} {c['name']:26} calls={c['calls']:3} first={c['first_in']/1e3:4.0f}k "
              f"gathered={c['gathered']/1e3:5.0f}k processed={c['processed']/1e6:5.2f}M min={c['minutes']} "
              f"after={c['main_calls_after']:3} inline~{c['inline_estimate']/1e6:5.2f}M delegated~{c['delegated_estimate']/1e6:5.2f}M")
    print("review roles:", ", ".join(r["review_roles"]) or "none")
    for f in r["flags"]:
        print("FLAG", f)


def main(argv):
    as_json = "--json" in argv
    results = []
    for arg in [a for a in argv if a != "--json"]:
        path = resolve(arg)
        if not path:
            print(f"not found: {arg}", file=sys.stderr)
            continue
        results.append(measure(path))
    if as_json:
        print(json.dumps(results, indent=1, default=str))
    else:
        for r in results:
            show(r)
    return 0 if results else 1


if __name__ == "__main__":
    if len(sys.argv) < 2 or sys.argv[1] in ("-h", "--help"):
        print(__doc__)
        sys.exit(0)
    sys.exit(main(sys.argv[1:]))
