"""Measure the Hive viewer's share of processed tokens in Claude Code sessions.

Usage (main-thread transcripts only, never subagents):
    cd ~/.claude/projects
    python3 -I <this file> $(find . -name '*.jsonl' -not -path '*/subagents/*' \
        -newermt 'YYYY-MM-DD HH:MM' | xargs rg -l 'hive view')

Processed tokens of one API call = input + cache_creation + cache_read + output.

Two metrics, reported per session and in total:
- calls: processed tokens of the API calls that issued a viewer tool call
  (a `hive view` Bash command or a Read of flow-walkthrough files), split by
  the share of viewer tool calls in that response. This is the method behind
  the 5-16 % figure of #328.
- carried: estimated size of each viewer tool call and its result (chars / 4)
  multiplied by the API calls that follow it until the next compaction.
"""
import json
import re
import sys
from collections import Counter


def classify_bash(cmd):
    if "hive view" not in cmd:
        return None
    if "hive view contract" in cmd:
        for flag, name in (("--for", "contract --for"), ("--list", "contract --list"),
                           ("--section", "contract --section"), ("--operation", "contract --operation")):
            if flag in cmd:
                return name
        return "contract (full)"
    if "hive view session get" in cmd:
        return "session get"
    if re.search(r"hive view (status|open|start|stop|register|serve)", cmd):
        return "lifecycle"
    return "publish/other"


HEREDOC = re.compile(r"<<-?\s*['\"]?(\w+)['\"]?([^\n]*)\n(.*?)\n\s*\1\b", re.S)
HEREDOC_WRITE = re.compile(r">\s*([^\s<>|;&]+)\s*<<-?\s*['\"]?(\w+)['\"]?[^\n]*\n(.*?)\n\s*\2\b", re.S)


def strip_heredocs(cmd):
    """Drop heredoc bodies: text written into a file is not a command run."""
    return HEREDOC.sub(lambda m: "<<" + m.group(1) + m.group(2), cmd)


def reference_kind(path):
    return "read viewer-publication.md" if "viewer-publication" in path else "read flow-walkthrough other"


def classify(name, inp, helpers):
    """helpers: basenames of files this session wrote that call `hive view`."""
    if name == "Bash":
        cmd = strip_heredocs(inp.get("command", ""))
        cat = classify_bash(cmd)
        if cat:
            return cat
        if any(h in cmd for h in helpers):
            return "helper script"
        m = re.search(r"\b(cat|head|tail|sed|less|bat|nl)\b[^|;&]*?(\S*flow-walkthrough\S*)", cmd)
        return reference_kind(m.group(2)) if m else None
    path = inp.get("file_path") or ""
    if name == "Read" and "flow-walkthrough" in path:
        return reference_kind(path)
    return None


SCRIPT_EXT = (".py", ".sh", ".bash", ".zsh", ".js", ".mjs")
# A script that runs the viewer CLI, not one that only mentions it in text.
CALLS_VIEWER = re.compile(
    r"""["']hive["']\s*,\s*["']view["']|(subprocess|check_output|\brun\(|execFile|spawn)[^\n]*hive view|^\s*hive view """,
    re.M)


def is_helper(path, text):
    return bool(path) and path.endswith(SCRIPT_EXT) and bool(CALLS_VIEWER.search(text or ""))


def note_helpers(name, inp, helpers):
    """Record scripts this session wrote that run `hive view` (Write, Edit, heredoc)."""
    if name in ("Write", "Edit"):
        path = inp.get("file_path") or ""
        if is_helper(path, inp.get("content") or inp.get("new_string")):
            helpers.add(path.rsplit("/", 1)[-1])
    elif name == "Bash":
        for m in HEREDOC_WRITE.finditer(inp.get("command", "")):
            if is_helper(m.group(1), m.group(3)):
                helpers.add(m.group(1).rsplit("/", 1)[-1])


def text_len(content):
    if isinstance(content, str):
        return len(content)
    n = 0
    for c in content or []:
        if isinstance(c, dict):
            n += len(c.get("text") or "") if c.get("type") == "text" else len(json.dumps(c))
    return n


def analyze(path):
    usage, order, tools = {}, [], {}
    pending, items, segments, counts, helpers = {}, [], [0], Counter(), set()
    first = last = None
    for line in open(path, encoding="utf-8"):
        try:
            e = json.loads(line)
        except ValueError:
            continue
        ts = e.get("timestamp")
        if ts:
            first, last = first or ts, ts
        if e.get("type") == "system" and e.get("subtype") == "compact_boundary":
            segments.append(len(order))
        m = e.get("message") or {}
        if e.get("type") == "assistant":
            mid, u = m.get("id"), m.get("usage")
            if mid and u and mid not in usage:
                usage[mid] = sum(u.get(k, 0) or 0 for k in (
                    "input_tokens", "cache_creation_input_tokens", "cache_read_input_tokens", "output_tokens"))
                order.append(mid)
            for c in m.get("content") or []:
                if not isinstance(c, dict) or c.get("type") != "tool_use":
                    continue
                inp = c.get("input") or {}
                note_helpers(c.get("name"), inp, helpers)
                cat = classify(c.get("name"), inp, helpers)
                if not cat and c.get("name") in ("Write", "Edit") and \
                        is_helper(inp.get("file_path") or "", inp.get("content") or inp.get("new_string")):
                    cat = "helper script"
                tools.setdefault(mid, []).append(bool(cat))
                if cat:
                    pending[c["id"]] = (cat, len(json.dumps(inp)))
        elif e.get("type") == "user" and isinstance(m.get("content"), list):
            for c in m["content"]:
                if isinstance(c, dict) and c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                    cat, chars = pending.pop(c["tool_use_id"])
                    items.append((len(order), (chars + text_len(c.get("content"))) / 4))
                    counts[cat] += 1
    total = sum(usage.values())
    calls = sum(usage.get(mid, 0) * sum(ts) / len(ts) for mid, ts in tools.items() if any(ts))
    carried = 0.0
    for idx, tok in items:
        end = next((s for s in segments if s > idx), len(order))
        carried += tok * max(end - idx, 0)
    return {"session": path.split("/")[-1][:8], "project": path.split("/")[-2].split("-")[-1],
            "start": first, "end": last, "api_calls": len(order), "compactions": len(segments) - 1,
            "processed": total, "viewer_calls_tokens": round(calls), "viewer_carried_tokens": round(carried),
            "counts": dict(counts)}


if __name__ == "__main__":
    T = C = K = 0
    for p in sys.argv[1:]:
        r = analyze(p)
        T, C, K = T + r["processed"], C + r["viewer_calls_tokens"], K + r["viewer_carried_tokens"]
        pt = r["processed"] or 1
        print(f'{(r["start"] or "")[:16]} {r["project"][:9]:9} {r["session"]} calls {r["api_calls"]:5} '
              f'proc {r["processed"] / 1e6:7.1f}M  viewer-calls {100 * r["viewer_calls_tokens"] / pt:5.1f}%  '
              f'carried {100 * r["viewer_carried_tokens"] / pt:4.1f}%  {r["counts"]}')
    if T:
        print(f"TOTAL proc {T / 1e6:.0f}M  viewer-calls {C / 1e6:.1f}M ({100 * C / T:.1f}%)  "
              f"carried {K / 1e6:.1f}M ({100 * K / T:.1f}%)")
