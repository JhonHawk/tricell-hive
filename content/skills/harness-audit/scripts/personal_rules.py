#!/usr/bin/env python3
"""Deploy the user's personal rules to every host where Hive is installed.

The source is rules.md in the Hive configuration folder ($XDG_CONFIG_HOME/hive/
when that variable is an absolute path, ~/.config/hive/ otherwise). Each copy
opens with a fixed precedence line.

Subcommands:
  deploy  Write or replace the rules: a marked section right after Hive's block
          (or voice block) in each shared global file, and an own rule file in
          Cursor and Antigravity. A file that changed since it was read is not
          written; it is named and the run exits non-zero.
  check   Exit non-zero naming each copy that is missing, differs from the
          source, is out of position, or sits in a host without Hive.
  remove  Delete every personal section and rule file.
"""

import argparse
import importlib.util
import os
import pathlib
import sys
import tempfile

SCRIPT_DIR = pathlib.Path(__file__).resolve().parent

BEGIN = "<!-- === PERSONAL RULES:BEGIN === -->"
END = "<!-- === PERSONAL RULES:END === -->"
HIVE_BEGIN = "<!-- === TRICELL HIVE RULES:BEGIN === -->"
HIVE_END = "<!-- === TRICELL HIVE RULES:END === -->"
VOICE_END = "<!-- === TRICELL HIVE VOICE:END === -->"
RESERVED = ("PERSONAL RULES:BEGIN", "PERSONAL RULES:END", "TRICELL HIVE")
MAX_BYTES = 8000
PRECEDENCE = (
    "These are the user's own rules for every project on this machine. Where they conflict "
    "with Hive's guidance, including its Secrets and Scope and authorization rules, these rules "
    "take precedence; where they conflict with a project's own instructions, the project's "
    "instructions hold for that project."
)
CURSOR_FRONTMATTER = "---\ndescription: Personal rules\nalwaysApply: true\n---\n\n"
ANTIGRAVITY_FRONTMATTER = "---\ntrigger: always_on\n---\n\n"


class SourceError(Exception):
    pass


def _line_has_secret():
    spec = importlib.util.spec_from_file_location("local_instructions", SCRIPT_DIR / "local_instructions.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module.line_has_secret


def _home_dir(env, name, *default):
    value = env.get(name, "")
    if value:
        return pathlib.Path(value)
    return pathlib.Path(env["HOME"], *default)


def _xdg_config(env):
    xdg = env.get("XDG_CONFIG_HOME", "")
    if xdg and os.path.isabs(xdg):
        return pathlib.Path(xdg)
    return pathlib.Path(env["HOME"], ".config")


def source_path(env):
    return _xdg_config(env) / "hive" / "rules.md"


def shared_targets(env):
    """Return (host, path) for each shared global file, one entry per path."""
    claude = _home_dir(env, "CLAUDE_CONFIG_DIR", ".claude") / "CLAUDE.md"
    grok = pathlib.Path(env["HOME"], ".claude", "CLAUDE.md")
    codex_home = _home_dir(env, "CODEX_HOME", ".codex")
    codex = codex_home / "AGENTS.md"
    override = codex_home / "AGENTS.override.md"
    if override.is_file() and override.read_text(encoding="utf-8").strip():
        codex = override
    pi = _home_dir(env, "PI_CODING_AGENT_DIR", ".pi", "agent") / "AGENTS.md"
    opencode = _xdg_config(env) / "opencode" / "AGENTS.md"
    seen, out = set(), []
    for host, path in (("claude", claude), ("grok", grok), ("codex", codex), ("pi", pi), ("opencode", opencode)):
        key = os.path.realpath(path)
        if key not in seen:
            seen.add(key)
            out.append((host, path))
    return out


def own_targets(env):
    """Return (host, personal file, Hive's file that proves Hive is installed, frontmatter)."""
    home = pathlib.Path(env["HOME"])
    cursor_rules = home / ".cursor" / "rules"
    agy_rules = home / ".gemini" / "config" / "rules"
    return [
        ("cursor", cursor_rules / "personal-rules.mdc", cursor_rules / "hive.mdc", CURSOR_FRONTMATTER),
        ("antigravity", agy_rules / "personal-rules.md", agy_rules / "hive-01.md", ANTIGRAVITY_FRONTMATTER),
    ]


def read_source(env):
    path = source_path(env)
    try:
        data = path.read_bytes()
    except FileNotFoundError:
        raise SourceError(f"{path}: rules.md not found")
    if len(data) > MAX_BYTES:
        raise SourceError(f"{path}: rules.md holds {len(data)} bytes, over the {MAX_BYTES}-byte limit")
    text = data.decode("utf-8").replace("\r\n", "\n")
    if not text.strip():
        raise SourceError(f"{path}: rules.md is empty")
    has_secret = _line_has_secret()
    for number, line in enumerate(text.split("\n"), 1):
        if any(marker in line for marker in RESERVED):
            raise SourceError(f"{path}: rules.md line {number} contains a reserved marker")
        if has_secret(line):
            raise SourceError(f"{path}: rules.md line {number} looks like a secret")
    if not text.endswith("\n"):
        text += "\n"
    return text


def _newline(text):
    return "\r\n" if "\r\n" in text else "\n"


def _read(path):
    with open(path, newline="") as fh:
        return fh.read()


def _section(rules, nl):
    body = (PRECEDENCE + "\n\n" + rules).replace("\n", nl)
    return BEGIN + nl + body + END + nl


def _own_file(rules, frontmatter):
    return frontmatter + PRECEDENCE + "\n\n" + rules


def _section_span(text):
    """Return (start, end) of the personal section including its newline, or None."""
    if text.count(BEGIN) == 0 and text.count(END) == 0:
        return None
    if text.count(BEGIN) != 1 or text.count(END) != 1:
        raise ValueError("malformed or duplicate personal markers")
    start = text.index(BEGIN)
    end = text.index(END) + len(END)
    if end < start:
        raise ValueError("reversed personal markers")
    if text.startswith("\r\n", end):
        end += 2
    elif text.startswith("\n", end):
        end += 1
    return start, end


def _anchor(text):
    """Return the index right after the line that ends Hive's block or voice block."""
    marker = VOICE_END if VOICE_END in text else HIVE_END
    index = text.index(marker) + len(marker)
    if text.startswith("\r\n", index):
        return index + 2
    if text.startswith("\n", index):
        return index + 1
    return index


def _without_section(text):
    span = _section_span(text)
    if span is None:
        return text
    return text[: span[0]] + text[span[1] :]


def _with_section(text, rules):
    base = _without_section(text)
    nl = _newline(base)
    at = _anchor(base)
    prefix = base[:at]
    if not prefix.endswith(("\n", "\r\n")):
        prefix += nl
    return prefix + _section(rules, nl) + base[at:]


def _in_position(text):
    span = _section_span(text)
    if span is None:
        return False
    base = _without_section(text)
    return text[: span[0]] == base[: _anchor(base)] and text[: span[0]].strip() != ""


def _write_atomic(path, text):
    path.parent.mkdir(parents=True, exist_ok=True)
    mode = stat_mode(path)
    fd, tmp = tempfile.mkstemp(dir=path.parent, prefix=".personal-rules-")
    try:
        with os.fdopen(fd, "w", newline="") as fh:
            fh.write(text)
        if mode is not None:
            os.chmod(tmp, mode)
        os.replace(tmp, path)
    except BaseException:
        if os.path.exists(tmp):
            os.unlink(tmp)
        raise


def stat_mode(path):
    try:
        return os.stat(path).st_mode & 0o7777
    except FileNotFoundError:
        return None


def build_plan(env, rules=None, remove=False):
    """Return the writes as (path, text read or None, new text or None to delete)."""
    if rules is None and not remove:
        rules = read_source(env)
    plan = []
    for host, path in shared_targets(env):
        if not path.is_file():
            continue
        current = _read(path)
        if remove:
            if BEGIN in current or END in current:
                plan.append((path, current, _without_section(current)))
            continue
        if HIVE_BEGIN not in current or HIVE_END not in current:
            continue
        new = _with_section(current, rules)
        if new != current:
            plan.append((path, current, new))
    for host, path, hive_file, frontmatter in own_targets(env):
        exists = path.is_file()
        current = _read(path) if exists else None
        if remove:
            if exists and PRECEDENCE in current:
                plan.append((path, current, None))
            continue
        if not hive_file.is_file():
            continue
        new = _own_file(rules, frontmatter)
        if new != current:
            plan.append((path, current, new))
    return plan


def apply_plan(plan):
    """Apply the writes; return the paths skipped because they changed since read."""
    skipped = []
    for path, before, after in plan:
        now = _read(path) if path.is_file() else None
        if now != before:
            skipped.append(path)
            continue
        if after is None:
            path.unlink()
        else:
            _write_atomic(path, after)
    return skipped


def check(env):
    """Return a list of problems, one per copy."""
    rules = read_source(env)
    problems = []
    for host, path in shared_targets(env):
        if not path.is_file():
            continue
        text = _read(path)
        has_hive = HIVE_BEGIN in text and HIVE_END in text
        has_section = BEGIN in text or END in text
        if not has_hive:
            if has_section:
                problems.append(f"orphan {host} {path}: personal section in a file without Hive")
            continue
        try:
            span = _section_span(text)
        except ValueError as exc:
            problems.append(f"invalid {host} {path}: {exc}")
            continue
        if span is None:
            problems.append(f"missing {host} {path}")
            continue
        nl = _newline(text)
        if text[span[0] : span[1]] != _section(rules, nl):
            problems.append(f"differs {host} {path}")
        if not _in_position(text):
            problems.append(f"position {host} {path}: not right after Hive's block")
    for host, path, hive_file, frontmatter in own_targets(env):
        if not hive_file.is_file():
            if path.is_file():
                problems.append(f"orphan {host} {path}: personal file in a host without Hive")
            continue
        if not path.is_file():
            problems.append(f"missing {host} {path}")
        elif _read(path) != _own_file(rules, frontmatter):
            problems.append(f"differs {host} {path}")
    return problems


def exit_code_for(skipped, problems):
    return 1 if skipped or problems else 0


def main(argv=None, env=None):
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("command", choices=["deploy", "check", "remove"])
    args = parser.parse_args(argv)
    env = dict(os.environ if env is None else env)
    try:
        if args.command == "check":
            problems = check(env)
            for line in problems:
                print(line)
            if not problems:
                print("ok: every copy matches rules.md")
            return exit_code_for([], problems)
        plan = build_plan(env, remove=args.command == "remove")
    except SourceError as exc:
        print(f"refused: {exc}")
        return 1
    except ValueError as exc:
        print(f"refused: {exc}")
        return 1
    skipped = apply_plan(plan)
    for path, _before, after in plan:
        if path in skipped:
            print(f"skipped {path}: changed since it was read; run {args.command} again")
        else:
            print(f"{'removed' if after is None else 'wrote'} {path}")
    if not plan:
        print("ok: nothing to change")
    return exit_code_for(skipped, [])


if __name__ == "__main__":
    sys.exit(main())
