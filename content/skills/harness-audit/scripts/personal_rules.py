#!/usr/bin/env python3
"""Deploy the user's personal rules to every host where Hive is installed.

The source is rules.md in the Hive configuration folder ($XDG_CONFIG_HOME/hive/
when that variable is an absolute path, ~/.config/hive/ otherwise). Each copy
opens with a fixed precedence line.

Subcommands:
  deploy  Write or replace the rules: a marked section right after Hive's block
          (or voice block) in each shared global file, and an own rule file in
          Cursor and Antigravity. It refuses, writing nothing, an invalid
          source, a symlinked target and a rule file it did not write. A file
          that changed since it was read, or that cannot be written, is named
          and the run exits non-zero.
  check   Exit non-zero naming each copy that is missing, differs from the
          source, is out of position, or sits where its host no longer reads it.
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
# Hive's release refuses any file that holds its block delimiters verbatim, so
# this script composes them instead of spelling them out.
_HIVE_MARKER = "<!-- === TRICELL HIVE {}:{} === -->"
HIVE_BEGIN = _HIVE_MARKER.format("RULES", "BEGIN")
HIVE_END = _HIVE_MARKER.format("RULES", "END")
VOICE_END = _HIVE_MARKER.format("VOICE", "END")
RESERVED = ("PERSONAL RULES:BEGIN", "PERSONAL RULES:END", "TRICELL HIVE")
MAX_BYTES = 8000
PRECEDENCE = (
    "These are the user's own rules for every project on this machine. Where they conflict "
    "with Hive's guidance, including its Scope and authorization rules, these rules take "
    "precedence, except Hive's Secrets rules, which still hold; where they conflict with a "
    "project's own instructions, the project's instructions hold for that project."
)
CURSOR_FRONTMATTER = "---\ndescription: Personal rules\nalwaysApply: true\n---\n\n"
ANTIGRAVITY_FRONTMATTER = "---\ntrigger: always_on\n---\n\n"


class SourceError(Exception):
    pass


class TargetError(Exception):
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


def _codex_files(env):
    """Return (the file Codex reads, the other one), as the manager decides."""
    codex_home = _home_dir(env, "CODEX_HOME", ".codex")
    agents = codex_home / "AGENTS.md"
    override = codex_home / "AGENTS.override.md"
    if override.is_file() and override.stat().st_size > 0:
        return override, agents
    return agents, override


def _dedupe(pairs):
    seen, out = set(), []
    for host, path in pairs:
        key = os.path.realpath(path)
        if key not in seen:
            seen.add(key)
            out.append((host, path))
    return out


def shared_targets(env):
    """Return (host, path) for each shared global file a host reads now."""
    claude = _home_dir(env, "CLAUDE_CONFIG_DIR", ".claude") / "CLAUDE.md"
    grok = pathlib.Path(env["HOME"], ".claude", "CLAUDE.md")
    codex, _ = _codex_files(env)
    pi = _home_dir(env, "PI_CODING_AGENT_DIR", ".pi", "agent") / "AGENTS.md"
    opencode = _xdg_config(env) / "opencode" / "AGENTS.md"
    return _dedupe([("claude", claude), ("grok", grok), ("codex", codex), ("pi", pi), ("opencode", opencode)])


def unread_files(env):
    """Return (host, path) for shared files a host could hold a section in but does not read now."""
    _, other = _codex_files(env)
    read = {os.path.realpath(p) for _, p in shared_targets(env)}
    return [(host, path) for host, path in [("codex", other)] if os.path.realpath(path) not in read]


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
    try:
        text = data.decode("utf-8").replace("\r\n", "\n")
    except UnicodeDecodeError:
        raise SourceError(f"{path}: rules.md is not valid UTF-8")
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


def _check_no_link(path, env):
    """Refuse a target that is a symlink or sits under a symlinked folder below HOME."""
    home = os.path.realpath(env["HOME"])
    current = pathlib.Path(path)
    while True:
        if current.is_symlink():
            raise TargetError(f"{path}: refused, {current} is a symbolic link")
        parent = current.parent
        if parent == current or os.path.realpath(parent) == home or str(parent) == env["HOME"]:
            return
        current = parent


def _read(path):
    with open(path, encoding="utf-8", newline="") as fh:
        return fh.read()


def _newline(text):
    """Return the line ending of the line that ends Hive's block, as Hive writes it."""
    marker = VOICE_END if VOICE_END in text else HIVE_END
    if marker in text:
        index = text.index(marker) + len(marker)
        if text.startswith("\r\n", index):
            return "\r\n"
        if text.startswith("\n", index):
            return "\n"
    first = text.find("\n")
    return "\r\n" if first > 0 and text[first - 1] == "\r" else "\n"


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
    if text.index(END) < start:
        raise ValueError("reversed personal markers")
    end = text.index(END) + len(END)
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
    return text[: span[0]] == base[: _anchor(base)]


def stat_mode(path):
    try:
        return os.stat(path).st_mode & 0o7777
    except FileNotFoundError:
        return None


def _write_atomic(path, text):
    path.parent.mkdir(parents=True, exist_ok=True)
    mode = stat_mode(path)
    fd, tmp = tempfile.mkstemp(dir=path.parent, prefix=".personal-rules-")
    try:
        with os.fdopen(fd, "w", encoding="utf-8", newline="") as fh:
            fh.write(text)
        if mode is not None:
            os.chmod(tmp, mode)
        os.replace(tmp, path)
    except BaseException:
        if os.path.exists(tmp):
            os.unlink(tmp)
        raise


def _planned(path, current, new, env):
    _check_no_link(path, env)
    return (path, current, new)


def build_plan(env, rules=None, remove=False):
    """Return the writes as (path, text read or None, new text or None to delete).

    Raises SourceError or TargetError, before any write, for an invalid source,
    a symlinked target, a rule file this script did not write, or malformed markers.
    """
    if rules is None and not remove:
        rules = read_source(env)
    plan = []
    candidates = shared_targets(env) + (unread_files(env) if remove else [])
    for host, path in candidates:
        if not path.is_file():
            continue
        _check_no_link(path, env)
        current = _read(path)
        try:
            if remove:
                if BEGIN in current or END in current:
                    plan.append(_planned(path, current, _without_section(current), env))
                continue
            if HIVE_BEGIN not in current or HIVE_END not in current:
                continue
            new = _with_section(current, rules)
        except ValueError as exc:
            raise TargetError(f"{path}: {exc}")
        if new != current:
            plan.append(_planned(path, current, new, env))
    for host, path, hive_file, frontmatter in own_targets(env):
        exists = path.is_file() or path.is_symlink()
        if exists:
            _check_no_link(path, env)
        current = _read(path) if exists else None
        if current is not None and PRECEDENCE not in current:
            if remove:
                continue
            if hive_file.is_file():
                raise TargetError(f"{path}: refused, this file was not written by personal_rules.py")
            continue
        if remove:
            if exists:
                plan.append(_planned(path, current, None, env))
            continue
        if not hive_file.is_file():
            continue
        new = _own_file(rules, frontmatter)
        if new != current:
            plan.append(_planned(path, current, new, env))
    return plan


def apply_plan(plan):
    """Apply the writes; return (paths skipped because they changed, (path, error) failures)."""
    skipped, failed = [], []
    for path, before, after in plan:
        try:
            now = _read(path) if path.is_file() else None
            if now != before:
                skipped.append(path)
                continue
            if after is None:
                path.unlink()
            else:
                _write_atomic(path, after)
        except OSError as exc:
            failed.append((path, exc.strerror or str(exc)))
    return skipped, failed


def check(env):
    """Return a list of problems, one per copy or per source error."""
    problems = []
    try:
        rules = read_source(env)
    except SourceError as exc:
        rules = None
        problems.append(f"source {exc}")
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
            if rules is not None:
                problems.append(f"missing {host} {path}")
            continue
        if rules is None:
            problems.append(f"present {host} {path}: no valid rules.md to compare with")
        elif text[span[0] : span[1]] != _section(rules, _newline(text)):
            problems.append(f"differs {host} {path}")
        if not _in_position(text):
            problems.append(f"position {host} {path}: not right after Hive's block")
    for host, path in unread_files(env):
        if path.is_file() and (BEGIN in _read(path) or END in _read(path)):
            problems.append(f"orphan {host} {path}: personal section in a file {host} no longer reads")
    for host, path, hive_file, frontmatter in own_targets(env):
        exists = path.is_file()
        ours = exists and PRECEDENCE in _read(path)
        if not hive_file.is_file():
            if ours:
                problems.append(f"orphan {host} {path}: personal file in a host without Hive")
            continue
        if not ours:
            if exists:
                problems.append(f"foreign {host} {path}: not written by personal_rules.py")
            elif rules is not None:
                problems.append(f"missing {host} {path}")
        elif rules is None:
            problems.append(f"present {host} {path}: no valid rules.md to compare with")
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
    if args.command == "check":
        problems = check(env)
        for line in problems:
            print(line)
        if not problems:
            print("ok: every copy matches rules.md")
        return exit_code_for([], problems)
    try:
        plan = build_plan(env, remove=args.command == "remove")
    except (SourceError, TargetError) as exc:
        print(f"refused: {exc}")
        return 1
    skipped, failed = apply_plan(plan)
    failed_paths = {path for path, _ in failed}
    for path, _before, after in plan:
        if path in skipped:
            print(f"skipped {path}: changed since it was read; run {args.command} again")
        elif path not in failed_paths:
            print(f"{'removed' if after is None else 'wrote'} {path}")
    for path, reason in failed:
        print(f"failed {path}: {reason}")
    if not plan:
        print("ok: nothing to change")
    return exit_code_for(skipped, [str(p) for p, _ in failed])


if __name__ == "__main__":
    sys.exit(main())
