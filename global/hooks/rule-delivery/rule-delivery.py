#!/usr/bin/env python3
"""rule-delivery — read the rule, then write the file.

Claude Code loads a path-scoped rule when a matching file is READ, so creating
the first file of a kind never loads it, and every read-only agent that opens a
`.ts` pays for a rule it will never apply. This hook replaces that mechanism
with the one primitive every harness enforces — a PreToolUse denial carrying a
short reason:

    Write src/a.ts   ->  denied: "Held: read <rule file> first, then re-issue
                                  this call."
    Read <rule file> ->  (PostToolUse) observed; the rule is now KNOWN
    Write src/a.ts   ->  allowed

A rule that no file announces declares a COMMAND prefix instead — `git
commit`, `pnpm add`, `agent-browser` — and a terminal call whose parsed stage
starts with one is held exactly the same way, by the same reason, released by
the same read.

Two events, one job each:

  PreToolUse   gate only. Never marks anything known.
  PostToolUse  observe only. It fires for calls that actually RAN, so a read
               the user refused — the rule files sit outside the project — is
               never recorded as a read.

No rule text travels in any output. No chunking, no transcript, no notion of a
turn: the parallel writes of one assistant message are simply denied too, which
is what makes them safe without any of it.

Fail-open everywhere: any exception, malformed stdin, missing or corrupt
manifest exits 0 with an empty stdout. `HIVE_RULE_DELIVERY=off|0|false|no`
disables it entirely.
"""

import glob as globbing
import hashlib
import json
import os
import re
import shlex
import sys
import time

# Only Grok clips a denial reason (~264 visible chars). Claude and Codex carry
# 24 KB, and a `.tsx` write matches five rules: naming them one at a time would
# cost three denial rounds for nothing.
REASON_BUDGETS = {"grok": 260, "pi": 260, "claude": 8000, "codex": 8000}
MAX_DENIALS = 3           # counted denials before a rule is released
DENIAL_WINDOW = 5.0       # seconds; a burst of denials counts once
STALE_SECONDS = 7 * 24 * 60 * 60
MAX_ALTERNATIVES = 256    # brace expansions per glob before a rule is dropped
STATE_DIR_NAME = "hive-rule-delivery"
KILL_VALUES = {"off", "0", "false", "no"}

REASON_PREFIX = "Held: read "
REASON_JOIN = " and "
REASON_SUFFIX = " first, then re-issue this call."
# Codex has no read tool: a shell reader is its ONLY way out of the gate, so
# the reason hands it the exact command instead of a path to figure out.
CODEX_SUFFIX = " first — run: cat %s — then re-issue this call."

READ_TOOLS = {"Read", "read_file"}
# Kept in step with global/hooks/executor-dispatch-gate (the repo's established
# file-edit surface) plus Grok's `write`. test_rule_delivery.py fails when this
# set and the deployed matcher diverge.
WRITE_TOOLS = {"Write", "Edit", "MultiEdit", "NotebookEdit",
               "search_replace", "write_file", "create_file", "edit_file", "write"}
TERMINAL_TOOLS = {"Bash", "run_terminal_command", "run_terminal_cmd", "shell"}
PATCH_TOOL = "apply_patch"
# `Delete File:` authors nothing and needs no rule.
PATCH_FILE_LINE = re.compile(
    r"^\*\*\* (?:Add File|Update File|Move to): *(.+?)\s*$", re.M)
PATCH_OPENING = "*** Begin Patch"
PATCH_CLOSING = "*** End Patch"
# Nobody reads a rule file through a 64 KB command, and nobody should pay to
# parse one. Past this, a shell command is neither parsed nor scanned.
MAX_COMMAND_CHARS = 64 * 1024
REFERENCE_ROOT = {"claude": "claude", "grok": "claude", "codex": "agents", "pi": "agents"}


class StateUnavailable(Exception):
    """The gate cannot be recorded; never deny under this."""


def window_seconds():
    try:
        return max(0.0, float(os.environ.get("HIVE_RULE_DELIVERY_WINDOW", DENIAL_WINDOW)))
    except (TypeError, ValueError):
        return DENIAL_WINDOW


# --- glob matching -------------------------------------------------------
#
# The manifest carries the globs verbatim from the rule's frontmatter: `**`
# crosses directories, `*` does not, and a brace group is a set of
# alternatives. `[` and `]` are LITERAL — no rule in the manifest uses a
# character class, and `app/[locale]/**` means the directory it looks like.
# Matching is by PATH SEGMENT: a substring match makes `**/_support/**` fire on
# `src/customer_support/…`.
#
# Deliberately not a regex: `**a**a**a**a**a**a**b` sends a backtracking engine
# exponential. The two-pointer wildcard match below is O(pattern x text).

def expand_braces(pattern, budget=MAX_ALTERNATIVES):
    """Every alternative of one glob, or None when it explodes past `budget`."""
    pending = [pattern]
    done = []
    while pending:
        if len(pending) + len(done) > budget:
            return None
        current = pending.pop()
        match = re.search(r"\{([^{}]*)\}", current)
        if not match:
            done.append(current)
            continue
        head, tail = current[:match.start()], current[match.end():]
        for option in match.group(1).split(","):
            pending.append(head + option + tail)
    return done


def compile_globs(globs):
    """Globs as segment lists, or None when one of them is unusable."""
    compiled = []
    for pattern in globs:
        expansions = expand_braces(pattern)
        if expansions is None:
            return None
        for expanded in expansions:
            # `***` and `**` inside a segment are just `*`; only a whole `**`
            # segment crosses directories.
            expanded = re.sub(r"\*{2,}", "**", expanded)
            segments = [segment if segment == "**" else segment.replace("**", "*")
                        for segment in expanded.split("/") if segment]
            if not segments:
                continue
            # Anchor at a segment boundary: a glob matches any tail of the path
            # that starts where a segment starts. This is also what makes a
            # slash-free glob (`turbo.json`) a basename match.
            if segments[0] != "**":
                segments = ["**"] + segments
            compiled.append(segments)
    return compiled


# --- command prefixes ----------------------------------------------------
#
# Some rules have no file that announces them and every reason to be read
# before a particular COMMAND runs: the supply-chain check before an install,
# git mechanics before a commit, the browser CLI before `agent-browser`. The
# manifest carries them as `commands` — command PREFIXES, matched
# TOKEN-FOR-TOKEN against the leading tokens of a stage the shell parser
# already produced. No new parser, and no evaluation of control flow: `git
# commit` matches, `git commit-tree` and `echo git commit` do not.

COMMAND_ARGUMENT = "+"    # trailing prefix token: demands a further argument


def compile_commands(commands):
    """Prefixes as `(tokens, requires_argument)`; a trailing `+` demands one.

    `npm install` with no package is a lockfile install and a different
    decision from `npm install <pkg>`, so a prefix may insist on at least one
    further non-option argument.
    """
    compiled = []
    for raw in commands:
        tokens = raw.split()
        requires_argument = bool(tokens) and tokens[-1] == COMMAND_ARGUMENT
        if requires_argument:
            tokens = tokens[:-1]
        if tokens:
            compiled.append((tokens, requires_argument))
    return compiled


def skippable(argv, index):
    """A tool's own global option, or the value one takes.

    `git -C <dir> commit`, `git --no-pager commit` and `pnpm --filter x add`
    are the same verbs as their bare forms. Which options take a value is not
    knowable without a table per tool, so a token following an option (unless
    that option carried its value with `=`) counts as one — the expected token
    is always tried FIRST, so `--no-pager commit` never swallows `commit`.
    """
    token = argv[index]
    if token.startswith("-"):
        return True
    previous = argv[index - 1]
    return previous.startswith("-") and "=" not in previous


def prefix_matches(argv, tokens, requires_argument):
    """Does this simple command start with that prefix?"""
    if not argv or os.path.basename(argv[0]) != tokens[0]:
        return False
    index = 1
    for token in tokens[1:]:
        while index < len(argv) and argv[index] != token and skippable(argv, index):
            index += 1
        if index >= len(argv) or argv[index] != token:
            return False
        index += 1
    if not requires_argument:
        return True
    return any(not later.startswith("-") for later in argv[index:])


def segment_match(pattern, text):
    """One path segment against one pattern segment (`*`, `?`)."""
    p = t = 0
    star_p = star_t = -1
    while t < len(text):
        matched = False
        if p < len(pattern):
            char = pattern[p]
            if char == "*":
                star_p, star_t = p, t
                p += 1
                continue
            if char == "?" or char == text[t]:
                matched = True
                p += 1
            if matched:
                t += 1
                continue
        if star_p >= 0:
            star_t += 1
            t = star_t
            p = star_p + 1
            continue
        return False
    while p < len(pattern) and pattern[p] == "*":
        p += 1
    return p == len(pattern)


def segments_match(patterns, parts):
    """Path segments against pattern segments; `**` spans zero or more."""
    p = t = 0
    star_p = star_t = -1
    while t < len(parts):
        if p < len(patterns) and patterns[p] == "**":
            star_p, star_t = p, t
            p += 1
            continue
        if p < len(patterns) and segment_match(patterns[p], parts[t]):
            p += 1
            t += 1
            continue
        if star_p >= 0:
            star_t += 1
            t = star_t
            p = star_p + 1
            continue
        return False
    while p < len(patterns) and patterns[p] == "**":
        p += 1
    return p == len(patterns)


def path_matches(path, compiled):
    parts = [part for part in path.replace(os.sep, "/").split("/") if part]
    if not parts:
        return False
    return any(segments_match(segments, parts) for segments in compiled)


# --- what a glob is matched against --------------------------------------
#
# The ABSOLUTE realpath of the target, always. It already contains the repo's
# own folder name, so a glob anchored on it (`**/*-specs/**`, `**/*-infra/**`)
# fires with no derivation at all.
#
# Deriving a project-relative path instead produced three separate false skips
# — a `cd` into a subdirectory, a `/tmp` vs `/private/tmp` spelling, and a
# Codex payload with an empty `cwd` — because every one of them changed what
# the "project" was believed to be. There is nothing left to get wrong here:
# the cost is over-matching (an ancestor directory can gate a file), which
# costs a read, against a skip, which costs the rule.

# Harness-owned homes and OS temp roots. A file there is configuration, memory
# or scratch — never the project code a language rule speaks about — and its
# absolute path carries the project slug, which makes it match by accident.
HARNESS_ROOTS = ("~/.claude", "~/.codex", "~/.grok", "~/.agents", "~/.pi",
                 "~/.config/opencode")
TEMP_ROOTS = ("/tmp", "/private/tmp", "/var/folders", "/private/var/folders")
# Generated: written by a tool, reviewed by nobody, frequently enormous. Only
# names no one hand-writes source into — `build`, `vendor` and `target` are all
# real source directories somewhere (`harness/build/` in this very repo).
DERIVED_SEGMENTS = frozenset({"node_modules", "dist", ".next", ".turbo", ".venv",
                              "__pycache__", "coverage"})


def real(path):
    """`realpath`, or the normalized path when it cannot be resolved.

    `/tmp/p` and `/private/tmp/p` are ONE directory on macOS; comparing the
    spellings as strings says they are unrelated. Both sides of every
    containment test go through this. It also swallows the `getcwd()` a
    relative path triggers, which raises once the process's own directory is
    deleted.
    """
    try:
        return os.path.realpath(path)
    except (OSError, ValueError):
        return os.path.normpath(path)


def ungated_roots():
    roots = [os.path.expanduser(root) for root in HARNESS_ROOTS]
    roots.extend(TEMP_ROOTS)
    temp = os.environ.get("TMPDIR")
    if temp:
        roots.append(temp)
    return [real(root) for root in roots if root]


def under(path, root):
    return path == root or path.startswith(root.rstrip("/") + "/")


def gate_path(target):
    """The path this file is gated BY, or None when it is never gated."""
    path = real(target) if os.path.isabs(target) else os.path.normpath(target)
    # DIRECTORY components only: a file called `dist` is source. Ancestors do
    # count — none of these names is a plausible ancestor of a real project,
    # and a checkout inside one is vendored or generated by definition.
    if any(part in DERIVED_SEGMENTS for part in os.path.dirname(path).split(os.sep)):
        return None
    if os.path.isabs(path) and any(under(path, root) for root in ungated_roots()):
        return None
    # A path that could not be made absolute is matched as it stands: a glob
    # may still fire on it, and failing toward gating is the whole policy.
    return path


# --- payload -------------------------------------------------------------
#
# Claude: `tool_name`, absolute `tool_input.file_path` (`notebook_path` for
# NotebookEdit), `agent_id`/`agent_type` only inside a subagent. Grok: both key
# spellings, `read_file` carries `target_file`, and paths are RELATIVE to
# `cwd`/`workspaceRoot`. Codex: `apply_patch` with the whole patch in
# `tool_input.command`, the shell as `Bash` (argv sometimes an array), and a
# subagent payload carrying the PARENT's `session_id`.

def tool_name(payload):
    name = payload.get("tool_name") or payload.get("toolName") or ""
    return name if isinstance(name, str) else ""


def tool_input(payload):
    fields = payload.get("tool_input") or payload.get("toolInput") or {}
    if isinstance(fields, str):
        try:  # one recorded producer hands the patch over as a JSON string
            fields = json.loads(fields)
        except ValueError:
            return {}
    return fields if isinstance(fields, dict) else {}


def agent_name(payload):
    """The subagent's roster name; absent means the main thread.

    Grok names it `subagentType`, camelCase only, on every event that fires
    inside a child and on none of the main thread's.
    """
    for key in ("agent_type", "agentType", "subagentType", "agent_name"):
        value = payload.get(key)
        if isinstance(value, str) and value:
            return value
    return ""


def detect_harness(payload):
    """The harness's own declaration first; the payload's shape as a fallback.

    Every hook command line this repo owns exports `HIVE_HARNESS`. Grok is the
    exception by construction: it reads the CLAUDE settings block, so it
    inherits `HIVE_HARNESS=claude` and is identified by its camelCase keys —
    which is what keeps it on the clipped reason budget.
    """
    if "toolUseId" in payload or "toolName" in payload or "hookEventName" in payload:
        return "grok"
    declared = os.environ.get("HIVE_HARNESS", "").strip().lower()
    if declared in REFERENCE_ROOT:
        return declared
    if payload.get("harness") == "pi":
        return "pi"
    if tool_name(payload) == PATCH_TOOL or "turn_id" in payload:
        return "codex"
    return "claude"


def workspace(payload):
    """What a RELATIVE path is relative to — the only thing cwd is used for.

    A cwd pointing at a subdirectory is harmless here: a relative path is
    relative to it by definition. Codex sends `cwd` empty and carries the real
    path in `workspace_roots` (the shape `bash-policy.sh` already reads); with
    none of them, the hook's own directory is the last resort — and `getcwd()`
    raises once that directory has been deleted, which must not reach the gate.
    """
    for key in ("cwd", "workspaceRoot", "workspace_root"):
        value = payload.get(key)
        if isinstance(value, str) and value:
            return value
    roots = payload.get("workspace_roots")
    if isinstance(roots, list):
        for value in roots:
            if isinstance(value, str) and value:
                return value
    try:
        return os.getcwd()
    except OSError:
        return ""  # the path stays relative and is matched as it stands


def resolve(path, base):
    # A NUL or a newline is not a path any of these harnesses writes to, and a
    # NUL raises out of every os.path call it reaches: drop it here rather than
    # gate on it.
    if not path or "\x00" in path or "\n" in path:
        return ""
    path = os.path.expanduser(path)
    if not os.path.isabs(path) and base:
        return os.path.join(base, path)
    return path


def patch_targets(command, base):
    """Files a patch body authors.

    A header counts only when it starts a line AND the text opens a patch: a
    commit message quoting `*** Update File:` writes nothing. The substring
    check also keeps the scan off every large command that is not a patch —
    which is why this one has no size limit: a patch too big to LEX still
    authors the files its headers name.
    """
    if not isinstance(command, str) or PATCH_OPENING not in command:
        return []
    found = [resolve(name, base) for name in PATCH_FILE_LINE.findall(command)]
    return [path for path in found if path]


def target_paths(payload):
    """Every file this call would author, absolute where the payload allows."""
    base = workspace(payload)
    fields = tool_input(payload)
    if tool_name(payload) == PATCH_TOOL:
        return patch_targets(fields.get("command"), base)
    path = (fields.get("file_path") or fields.get("notebook_path")
            or fields.get("target_file") or fields.get("filePath")
            or fields.get("path") or "")
    if not isinstance(path, str):
        return []
    resolved = resolve(path, base)
    return [resolved] if resolved else []


def command_text(payload):
    """The shell command as text, whether it arrived as a string or as argv."""
    command = tool_input(payload).get("command")
    if isinstance(command, list):
        return " ".join(str(part) for part in command)
    return command if isinstance(command, str) else ""


def read_starts_at_top(payload):
    """False when the read begins past the first line — it showed no rule."""
    offset = tool_input(payload).get("offset")
    if isinstance(offset, bool) or not isinstance(offset, (int, float)):
        return True
    return offset <= 1


def same_file(left, right):
    """Through `~`, `..`, symlinks and a case-insensitive volume."""
    try:
        if os.path.realpath(left) == os.path.realpath(right):
            return True
        return os.path.samefile(left, right)
    except (OSError, ValueError):
        return False


# --- shell observation ---------------------------------------------------
#
# A rule is observed only when a command actually PUT ITS TEXT in front of the
# model. Substring matching cannot tell `cat rule.md` from `rm rule.md`,
# `cat rule.md.bak` or `tail -f app.log # see rule.md`, so the command is
# parsed instead: unwrapped, tokenized, split into simple commands, and each
# candidate checked as a reader with the rule file among its ARGUMENTS.

READER_COMMANDS = {"cat", "head", "tail", "sed", "less", "bat", "nl"}
SHELL_WRAPPERS = {"bash", "sh", "zsh", "dash", "ksh"}
IGNORED_PREFIXES = {"sudo", "env", "command", "nohup", "time", "nice", "exec"}
LOOKUP_ONLY = {"command"}  # `command -v X` resolves X; it does not run it
PIPELINE_BREAKS = {";", "&&", "||", "&", "\n"}
# With `punctuation_chars=True` an operator is always its OWN token made of
# these characters alone. Testing a substring instead took the de-quoted script
# of `bash -lc 'pnpm add zod 2>&1'` for a redirection and never unwrapped it.
# `|` belongs here for `>|` (the noclobber override) — a token of these
# characters is a redirection only when it also carries `<` or `>`, so `|`,
# `||` and `&&` fall through to the pipeline splits below.
REDIRECT_CHARS = frozenset("<>&|")
ASSIGNMENT = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=")
# Where one command ends and the next begins, for the lenient scan: no shell
# rules, no quoting — `echo don't&&git commit` has to yield both stages.
OPERATOR_RUN = re.compile(r"[;|&]+")
ZERO_COUNT = re.compile(r"^-(?:[nc])?0+$")


def lex(text):
    """Shell tokens with operators separated; raises ValueError if unparsable.

    Comments are removed by `split_lines`, which can tell a word-initial `#`
    from `foo#bar`; `shlex` cannot, and cutting a line at a mid-word `#` loses
    every stage after it.
    """
    lexer = shlex.shlex(text, posix=True, punctuation_chars=True)
    lexer.whitespace_split = True
    lexer.commenters = ""
    return list(lexer)


def split_lines(text):
    """Lines of a multi-line command, honouring quotes.

    `shlex` swallows a newline as ordinary whitespace, which would glue
    `echo a` and `cat rule.md` into one nonsense command. Three things are
    resolved here, where the quoting state is known:

    - a `\\`+newline **continuation** joins the two lines: it is one command,
      and the half-line left behind is unlexable (`No escaped character`);
    - a **comment** starting a word is dropped to the end of the line, so an
      apostrophe in it never unbalances the quoting;
    - a backslash escape is consumed with the character it escapes.
    """
    lines, current, quote = [], [], None
    index = 0
    while index < len(text):
        char = text[index]
        if quote:
            current.append(char)
            if char == quote:
                quote = None
            elif quote == '"' and char == "\\" and index + 1 < len(text):
                index += 1
                current.append(text[index])
            index += 1
            continue
        if char == "\\" and index + 1 < len(text):
            if text[index + 1] == "\n":
                index += 2  # a continuation: the next line is this same command
                continue
            current.append(char)
            current.append(text[index + 1])
            index += 2
            continue
        if char in "'\"":
            quote = char
            current.append(char)
            index += 1
            continue
        if char == "#" and (not current or current[-1] in " \t"):
            while index < len(text) and text[index] != "\n":
                index += 1
            continue
        if char == "\n":
            lines.append("".join(current))
            current = []
            index += 1
            continue
        current.append(char)
        index += 1
    lines.append("".join(current))
    return [line for line in lines if line.strip()]


def strip_prefixes(argv):
    """The command itself, with what merely introduces it removed.

    A wrapper's own OPTIONS are not the command either — `env -i git commit`,
    `time -p git commit`. The value an option takes is deliberately NOT
    skipped: `env -i git …` and `sudo -u bob git …` are indistinguishable
    without a table per wrapper, and skipping one token too many would hide a
    real command (`sudo -u bob git commit` stays a documented miss).
    """
    index = 0
    while index < len(argv):
        if ASSIGNMENT.match(argv[index]):
            index += 1
            continue
        name = os.path.basename(argv[index])
        if name in IGNORED_PREFIXES:
            index += 1
            # `command -v X` LOOKS X UP instead of running it, and its only
            # options do exactly that — skipping them would hold a presence
            # check for the tool it asked about.
            if name not in LOOKUP_ONLY:
                while index < len(argv) and argv[index].startswith("-"):
                    index += 1
            continue
        break
    return argv[index:]


def wrapped_script(argv):
    """The script of a `bash -c '…'`-style stage, or None."""
    if not argv or os.path.basename(argv[0]) not in SHELL_WRAPPERS:
        return None
    for index in range(1, len(argv)):
        token = argv[index]
        if token.startswith("-") and "c" in token and index + 1 < len(argv):
            return argv[index + 1]
        if not token.startswith("-"):
            return None
    return None


def redirects(token):
    """An operator token, never a word that merely contains one of its chars."""
    return bool(token) and REDIRECT_CHARS.issuperset(token)


def pipelines(command, depth=0, unparsed=None):
    """[[ (argv, redirected), … ], …] — one list per pipeline, stages in order.

    `unparsed` is the HOLD side's channel, and the only asymmetry between the
    two sides of this hook. Pass a list and two things change, both of them
    "rather hold than skip":

    - a line the lexer refuses is collected there instead of failing the whole
      call — one apostrophe in a heredoc body must not drop every other line;
    - a `bash -c` script is unwrapped even when the wrapper's stdout is
      redirected, because a redirection hides output from the model without
      stopping the command from RUNNING.

    The observation side passes None and stays strict: a lenient observation
    would record a rule as read on a call nobody could parse.
    """
    if depth > 2:
        return []
    holding = unparsed is not None
    if isinstance(command, list):
        tokens = [str(part) for part in command if isinstance(part, (str, int, float))]
        raw = [[(tokens, False)]] if tokens else []
    else:
        raw = []
        for line in split_lines(command):
            try:
                tokens = lex(line)
            except ValueError:
                if not holding:
                    raise
                unparsed.append(line)
                continue
            stage, pipeline = [], []
            redirected = False
            skip_next = False
            for token in tokens:
                if skip_next:
                    skip_next = False
                    continue
                if redirects(token) and ">" in token:
                    # `2>` / `2>&1` redirect stderr; the model still sees stdout.
                    if stage and stage[-1] == "2":
                        stage.pop()
                    else:
                        redirected = True
                    skip_next = True
                    continue
                if redirects(token) and "<" in token:
                    skip_next = True  # an input redirection names no argument
                    continue
                if token == "|":
                    pipeline.append((stage, redirected))
                    stage, redirected = [], False
                    continue
                if token in PIPELINE_BREAKS:
                    pipeline.append((stage, redirected))
                    raw.append(pipeline)
                    stage, pipeline, redirected = [], [], False
                    continue
                stage.append(token)
            pipeline.append((stage, redirected))
            raw.append(pipeline)

    resolved = []
    for pipeline in raw:
        expanded = []
        for argv, redirected in pipeline:
            argv = strip_prefixes(argv)
            script = None if redirected and not holding else wrapped_script(argv)
            if script is not None:
                resolved.extend(pipelines(script, depth + 1, unparsed))
                continue
            if argv:
                expanded.append((argv, redirected))
        if expanded:
            resolved.append(expanded)
    return resolved


def shows_nothing(name, args):
    """A reader invoked so that the model sees none of the file."""
    if name == "sed":
        return any(argument.startswith("-") and not argument.startswith("--")
                   and "i" in argument for argument in args)
    if name in {"head", "tail"}:
        for index, argument in enumerate(args):
            if name == "tail" and argument in {"-f", "--follow"}:
                return True
            if name == "tail" and (argument.startswith("-") and not argument.startswith("--")
                                   and "f" in argument):
                return True
            if ZERO_COUNT.match(argument):
                return True
            if argument in {"-n", "-c"} and index + 1 < len(args):
                following = args[index + 1].lstrip("+-")
                if following.isdigit() and int(following) == 0:
                    return True
    return False


def expanded_arguments(argument, base):
    """The existing files one argument names.

    `{a,b}` and a literal `*`/`?` glob are what a model writes back when the
    reason hands it a directory and six basenames. Both are resolved by looking
    at the FILESYSTEM — nothing is ever executed, and a pattern matching no
    existing file simply names none.

    The pattern must be confined to the LAST component, so expanding it costs
    exactly one directory listing: a wildcard in a parent component walks the
    tree instead (`cat ~/*/*/*/*/*/*/*/*/*` measured 6.3 s against a 15 s hook
    timeout), and no model needs it to read a file whose directory the denial
    reason just named in full. Anything wider is simply not observed.
    """
    head = argument.rpartition("/")[0]
    if any(char in head for char in "*?{"):
        return []
    alternatives = expand_braces(argument) if "{" in argument else [argument]
    if alternatives is None:
        return []  # a brace bomb expands to nothing rather than to 2^n paths
    resolved = []
    for alternative in alternatives:
        # `cat "$HOME/…"` and `cat ~/…` are what a model actually types.
        target = resolve(os.path.expandvars(alternative), base)
        if not target:
            continue
        if "*" in alternative or "?" in alternative:
            resolved.extend(sorted(globbing.glob(target))[:MAX_ALTERNATIVES])
        else:
            resolved.append(target)
    return resolved


def stage_arguments(stage, base):
    """The files this simple command prints, or [] when it prints none."""
    argv, redirected = stage
    if redirected or not argv:
        return []
    name = os.path.basename(argv[0])
    if name not in READER_COMMANDS:
        return []
    args = argv[1:]
    if shows_nothing(name, args):
        return []
    shown = []
    for argument in args:
        if argument.startswith("-"):
            continue
        for target in expanded_arguments(argument, base):
            try:
                shown.append((os.path.realpath(target), target))
            except (OSError, ValueError):
                continue
    return shown


def moved_base(pipeline, base):
    """Where a `cd` stage leaves the shell, or None when this is not one.

    `cd <dir> && cat a.md b.md` is the shape a model writes when the reason
    names a directory once — the names after it resolve against that directory,
    not against the payload's cwd.
    """
    if len(pipeline) != 1:
        return None
    argv, redirected = pipeline[0]
    if redirected or not argv or os.path.basename(argv[0]) != "cd":
        return None
    moved = [argument for argument in argv[1:] if not argument.startswith("-")]
    if len(moved) != 1:
        return None  # bare `cd`, `cd -`, `cd a b`: not a destination we know
    target = resolve(os.path.expandvars(moved[0]), base)
    return target if target and os.path.isdir(target) else None


def command_size(command):
    """Characters to parse, or more than the limit when this is not a command."""
    if isinstance(command, list):
        return sum(len(str(part)) for part in command)
    if isinstance(command, str):
        return len(command)
    return MAX_COMMAND_CHARS + 1


def parse_command(command):
    """The pipelines of a shell call — parsed ONCE per call, never per rule.

    The previous shape re-lexed the whole command for each rule in the
    manifest, which put a 1 MB heredoc at 5.4 s on every terminal call. Past
    MAX_COMMAND_CHARS nothing is parsed at all, and an unparsable command
    yields nothing rather than a guess — this is the OBSERVATION side, where
    guessing would be a false release.
    """
    if command_size(command) > MAX_COMMAND_CHARS:
        return []
    try:
        return pipelines(command)
    except ValueError:
        return []


def displayed_files(command, base):
    """Every file this shell call showed the model."""
    shown = []
    for pipeline in parse_command(command):
        destination = moved_base(pipeline, base)
        if destination is not None:
            base = destination  # every later stage resolves from there
            continue
        for index, stage in enumerate(pipeline):
            # Whatever the reader feeds must also be something the model reads:
            # `cat rule.md | grep -c .` shows a number, not the rule.
            rest = pipeline[index + 1:]
            if not all(not redirected and argv
                       and os.path.basename(argv[0]) in READER_COMMANDS
                       for argv, redirected in rest):
                continue
            shown.extend(stage_arguments(stage, base))
    return shown


def lenient_stages(chunk):
    """Stages of a chunk read as PHYSICAL LINES — whitespace tokens only.

    A quote the lexer could not balance is the model's prose (`fix: don't
    crash`), not a different command: splitting on whitespace and stripping the
    quote characters recovers `git commit` from `git commit -m 'x`. Each
    physical line is its own command, cut on `;`, `|` and `&` wherever they
    appear — an unbalanced quote swallows every line after it, and an EVEN
    number of stray quotes welds two lines into one the lexer then parses
    happily, hiding whatever sat between them.

    HOLD side only: a false hold costs one read, a false skip costs the rule.
    """
    stages = []
    for line in chunk.split("\n"):
        for piece in OPERATOR_RUN.split(line):
            stage = [token.strip("'\"") for token in piece.split()]
            stages.append(strip_prefixes(stage))
    return [argv for argv in stages if argv]


def without_patch_bodies(command):
    """The command with every patch it carries removed — that part is DATA.

    Applies to an argv array element by element: Codex sends the whole call as
    `["bash", "-lc", "apply_patch <<'P' … P"]`, and a body that is data in the
    string form is data there too.

    A patch body is what the call WRITES, not what it runs, and its context
    lines read exactly like commands (` pnpm add zod` in a patch to a doc or a
    workflow). Removing them before the command parse drops that false hold and
    keeps the parse off a body that can be 60 KB of source. The write path is
    unaffected: it reads the same headers straight out of the raw text.

    An unterminated patch cuts nothing — losing a real stage is the worse error.
    """
    if isinstance(command, list):
        return [without_patch_bodies(part) if isinstance(part, str) else part
                for part in command]
    if not isinstance(command, str):
        return command
    while True:
        start = command.find(PATCH_OPENING)
        if start < 0:
            return command
        end = command.find(PATCH_CLOSING, start)
        if end < 0:
            return command
        command = command[:start] + command[end + len(PATCH_CLOSING):]


def multiline_chunks(command):
    """The pieces of a call whose PHYSICAL LINES are scanned leniently too.

    The strict parse succeeding on a chunk says nothing about the lines it
    welded together: an even number of stray quotes (a prose heredoc either
    side of the command) glues them into one line the lexer is happy with,
    and whatever sat between them is gone. So the lenient scan is not a
    fallback for a failed parse — it is a second reading of every multi-line
    chunk, unioned with the first.

    A single-line chunk needs no second reading: nothing can be hidden by a
    weld that did not happen, and the lexer either parsed it or said so.
    """
    if isinstance(command, list):
        return [part for part in command if isinstance(part, str) and "\n" in part]
    if isinstance(command, str) and "\n" in command:
        return [command]
    return []


def hold_stages(command):
    """Every simple command of a shell call, as argv — the HOLD side's parse.

    Same parser as the observation side (`bash -lc`, argv arrays, a leading
    `VAR=v`/`sudo`/`env`/`command`, split on `;`, `&&`, `||`, `|`), unioned
    with a lenient reading of every multi-line chunk and of anything the lexer
    refused. Operators are split on, never evaluated — see *Known limits*.

    Patch bodies come out FIRST, before the size check: a 69 KB patch is 69 KB
    of data around a handful of commands, and measuring the whole thing made
    "too big to lex" a way to run a declared command unheld.
    """
    command = without_patch_bodies(command)
    stages, unparsed = [], []
    if command_size(command) <= MAX_COMMAND_CHARS:
        stages = [argv for pipeline in pipelines(command, unparsed=unparsed)
                  for argv, _redirected in pipeline]
    chunks = multiline_chunks(command) or unparsed
    if not stages and not chunks:
        # Over the limit and single-line: the lexer is out of reach, so the
        # whole thing is read leniently rather than waved through.
        chunks = [part for part in (command if isinstance(command, list) else [command])
                  if isinstance(part, str)]
    for chunk in chunks:
        stages.extend(lenient_stages(chunk))
    return stages


def shows_reference(shown, reference):
    """Was this rule's text among what the command displayed?"""
    try:
        target = os.path.realpath(reference)
    except (OSError, ValueError):
        return False
    lowered = target.lower()
    for real, token in shown:
        if real == target:
            return True
        # A case-insensitive volume resolves both spellings to one file; confirm
        # before trusting it, and only for the handful that could collide.
        if real.lower() == lowered and same_file(token, reference):
            return True
    return False


# --- manifest ------------------------------------------------------------

def manifest_path():
    """HIVE_RULE_MANIFEST, else the deployed sibling, else the repo source."""
    override = os.environ.get("HIVE_RULE_MANIFEST")
    if override:
        return override
    here = os.path.dirname(os.path.abspath(__file__))
    deployed = os.path.join(here, "rule-manifest.json")
    if os.path.exists(deployed):
        return deployed
    return os.path.join(here, "..", "..", "..", "harness", "rule-manifest.json")


def load_manifest():
    with open(manifest_path(), encoding="utf-8") as handle:
        data = json.load(handle)
    if not isinstance(data, dict):
        raise ValueError("manifest is not an object")
    return data


def string_list(value):
    """A usable list of non-empty strings, or None when the entry is malformed.

    `[]` and an absent key are both "no trigger of this kind"; anything else
    that is not a list of non-empty strings makes the whole rule undeliverable
    rather than half-matched.
    """
    if value is None:
        return []
    if not isinstance(value, list):
        return None
    if not all(isinstance(item, str) and item for item in value):
        return None
    return value


def deliverable(manifest, root, agent):
    """(gateable rules in manifest order, is this agent read-only).

    A rule is gateable when it carries a usable trigger — globs, command
    prefixes, or both — it is not always-on, its text lives in the one store
    this hook owns, the agent does not already carry it as a pack nor a pack
    the rule declares itself exclusive with, and the reference deployed for
    this harness is a readable regular file.
    """
    agents = manifest.get("agents")
    packs = set()
    if agent and isinstance(agents, dict) and isinstance(agents.get(agent), list):
        packs = {name for name in agents[agent] if isinstance(name, str)}
    read_only_list = manifest.get("read_only_agents")
    read_only = bool(agent) and isinstance(read_only_list, list) and agent in read_only_list

    rules = manifest.get("rules")
    gateable = []
    for entry in rules if isinstance(rules, list) else []:
        if not isinstance(entry, dict) or not isinstance(entry.get("name"), str):
            continue
        globs = string_list(entry.get("globs"))
        commands = string_list(entry.get("commands"))
        if globs is None or commands is None:
            continue
        commands = compile_commands(commands)
        if not globs and not commands:
            continue
        if entry.get("always_on"):
            continue
        # `global/rules-situational/` is the ONE store this hook delivers from:
        # every rule text lives there, and the always-on ones are already
        # inlined in the core (skipped just above). A source anywhere else is
        # not a rule this gate can point at, so it is never gated on.
        if not str(entry.get("source", "")).startswith("global/rules-situational/"):
            continue
        if entry["name"] in packs:
            continue
        # Two frameworks claiming the same file extension (`*.service.ts` is
        # Angular's and NestJS's alike) would hold each other's agent on a rule
        # it will never apply. The rule file names the packs it is exclusive
        # with; carrying one of them is proof of which framework this is.
        exclusive = entry.get("exclusive_with")
        if isinstance(exclusive, list) and packs.intersection(
                name for name in exclusive if isinstance(name, str)):
            continue
        if read_only and not entry.get("readers"):
            continue
        references = entry.get("references")
        raw = references.get(root) if isinstance(references, dict) else None
        if not isinstance(raw, str) or not raw:
            continue
        reference = os.path.expanduser(raw)
        if not os.path.isfile(reference) or not os.access(reference, os.R_OK):
            continue
        compiled = compile_globs(globs) if globs else []
        if globs and not compiled:
            continue  # a brace bomb: nobody can afford to match it
        gateable.append({
            "name": entry["name"],
            "globs": compiled,
            "commands": commands,
            "reference": reference,
        })
    return gateable, read_only


# --- state ---------------------------------------------------------------
#
# One 0700 directory per (session, agent) holding empty marker files, each
# created with O_CREAT|O_EXCL so the parallel calls of one turn cannot corrupt
# it. Marker names are a HASH of what they key, never a prefix of one another
# (`communication-format` and `communication-format-mechanics` both exist).
# A marker that cannot be written at all means the gate cannot be recorded, and
# an unrecordable denial would repeat forever — so that case ALLOWS.

def slug(value):
    return re.sub(r"[^A-Za-z0-9_.-]", "-", str(value))[:120]


def digest(*parts):
    raw = "\x00".join(parts).encode("utf-8", "replace")
    return hashlib.sha1(raw).hexdigest()[:16]


def known_marker(name):
    return "k" + digest(name)


def denial_marker(name, ordinal):
    return "d%s-%d" % (digest(name), ordinal)


def state_root():
    return os.path.join(os.environ.get("TMPDIR", "/tmp"), STATE_DIR_NAME)


def session_prefix(payload):
    session = payload.get("session_id") or payload.get("sessionId")
    return "%s__" % slug(session) if session else None


def state_dir(payload):
    prefix = session_prefix(payload)
    if prefix is None:
        return None  # no session id: no state, and so never a denial
    # A Codex subagent may carry a name but no id; keying it as the main thread
    # would let one agent's read release another agent's gate.
    agent = (payload.get("agent_id") or payload.get("agentId")
             or agent_name(payload) or "main")
    return os.path.join(state_root(), prefix + slug(agent))


def ensure_state_dir(directory):
    """True when this call created it — the moment to prune the old ones."""
    if os.path.isdir(directory):
        return False
    try:
        os.makedirs(directory, mode=0o700, exist_ok=True)
        os.chmod(directory, 0o700)
    except OSError as error:
        raise StateUnavailable(str(error))
    return True


def claim(directory, name):
    """Create a marker exactly once. True = this process created it."""
    try:
        handle = os.open(os.path.join(directory, name),
                         os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    except FileExistsError:
        return False
    except OSError as error:
        raise StateUnavailable(str(error))
    os.close(handle)
    return True


def has_marker(directory, name):
    return os.path.exists(os.path.join(directory, name))


def marker_age(directory, name):
    """Seconds since the marker was written, or None when it is absent.

    Signed on purpose: a clock that went backwards (NTP, a suspended laptop)
    leaves a marker dated ahead, and a negative age must read as "the window
    has passed" — never as a counter frozen forever.
    """
    try:
        return time.time() - os.path.getmtime(os.path.join(directory, name))
    except OSError:
        return None


def drop_dir(path):
    for name in os.listdir(path):
        try:
            os.remove(os.path.join(path, name))
        except OSError:
            continue
    os.rmdir(path)


def clear_session(payload):
    """Drop the state of a session AND of every subagent under it."""
    prefix = session_prefix(payload)
    root = state_root()
    if prefix is None or not os.path.isdir(root):
        return
    for name in os.listdir(root):
        if name.startswith(prefix):
            try:
                drop_dir(os.path.join(root, name))
            except OSError:
                continue


def prune_stale_state():
    cutoff = time.time() - STALE_SECONDS
    try:
        names = os.listdir(state_root())
    except OSError:
        return
    for name in names:
        path = os.path.join(state_root(), name)
        try:
            if os.path.isdir(path) and os.path.getmtime(path) < cutoff:
                drop_dir(path)
        except OSError:
            continue


# --- the gate ------------------------------------------------------------

def reason_for(references, harness="claude"):
    """The denial text: each directory named ONCE, its basenames after it.

    Six absolute paths to one directory spent Grok's ~264-character clip three
    times over, and every clipped round is another denial the model has to sit
    through. The directory is the part that repeats.
    """
    grouped = {}
    for reference in references:
        directory, name = os.path.split(reference)
        grouped.setdefault(directory, []).append(name)
    listed = REASON_JOIN.join(
        "%s in %s/" % (", ".join(names), directory.rstrip("/")) if directory
        else ", ".join(names)
        for directory, names in grouped.items())
    if harness == "codex":
        # Quoted: a HOME with a space otherwise makes `cat /Users/Jo Smith/x.md`
        # two arguments, and the rule stays unread until the valve gives up.
        return REASON_PREFIX + listed + CODEX_SUFFIX % " ".join(
            shlex.quote(reference) for reference in references)
    return REASON_PREFIX + listed + REASON_SUFFIX


def denials_so_far(directory, name):
    return sum(1 for ordinal in range(1, MAX_DENIALS + 1)
               if has_marker(directory, denial_marker(name, ordinal)))


def count_denial(directory, name):
    """Record this denial against the rule's one counter.

    Ordinal k is claimable only once ordinal k-1 has aged past the window, so a
    parallel burst — or a model re-issuing in a tight loop — counts ONCE, while
    a model that keeps writing new files over time still walks the counter up.
    Returns True when the counter is full: the rule is released instead.
    """
    window = window_seconds()
    previous_age = 0.0
    for ordinal in range(1, MAX_DENIALS + 1):
        age = marker_age(directory, denial_marker(name, ordinal))
        if age is None:
            if ordinal == 1 or previous_age >= window or previous_age < 0:
                claim(directory, denial_marker(name, ordinal))
                # Claiming the last ordinal IS the release.
                return ordinal >= MAX_DENIALS
            return False  # inside the window: denied, and not counted
        previous_age = age
    return True


def reads_only(stages):
    """Does this call do nothing but display files (and move between them)?"""
    return bool(stages) and all(
        argv and os.path.basename(argv[0]) in READER_COMMANDS | {"cd"}
        for argv in stages)


def triggered(rules, targets, invocations):
    """The rules this call fires: by a path it authors or a command it runs."""
    fired = []
    for entry in rules:
        if targets and entry["globs"] and any(
                path_matches(target, entry["globs"]) for target in targets):
            fired.append(entry)
        elif invocations and entry["commands"] and any(
                prefix_matches(argv, tokens, requires_argument)
                for argv in invocations
                for tokens, requires_argument in entry["commands"]):
            fired.append(entry)
    return fired


def gate(directory, fired, harness):
    """The denial reason for this call, or None to let it through."""
    budget = REASON_BUDGETS[harness]
    pending = [entry for entry in fired
               if not has_marker(directory, known_marker(entry["name"]))]
    if not pending:
        return None

    # Least-counted first: under a budget too tight to name every rule, the one
    # left out last time goes first now. Directory breaks the tie, so rules
    # sharing one get packed into the same round — the reason names a directory
    # once, and a group that straddles two pays for both.
    pending.sort(key=lambda entry: (denials_so_far(directory, entry["name"]),
                                    os.path.dirname(entry["reference"])))

    # One group per round: as many rules as the budget can name whole. When
    # every rule in a group hits its valve the reason comes back empty — and
    # the groups the budget deferred are still UNREAD, so the call keeps being
    # held for them rather than passing on somebody else's release.
    while pending:
        named = []
        for entry in pending:
            references = [chosen["reference"] for chosen in named] + [entry["reference"]]
            if len(reason_for(references, harness)) > budget:
                continue  # named on a later denial; a clipped path is never emitted
            named.append(entry)
        if not named:
            return None  # not even one path fits: this rule is delivered elsewhere

        held = []
        for entry in named:
            if count_denial(directory, entry["name"]):
                claim(directory, known_marker(entry["name"]))  # release valve
            else:
                held.append(entry["reference"])
        if held:
            return reason_for(held, harness)
        chosen = {id(entry) for entry in named}
        pending = [entry for entry in pending if id(entry) not in chosen]
    return None


def emit_deny(reason):
    sys.stdout.write(json.dumps({"hookSpecificOutput": {
        "hookEventName": "PreToolUse",
        "permissionDecision": "deny",
        "permissionDecisionReason": reason,
    }}))


def call_failed(payload):
    """An explicit failure signal in the tool's response.

    A read that errored, was interrupted, or exited non-zero showed the model
    nothing; recording it as known would be the same lie PreToolUse told.
    Unknown response shapes stay observed.
    """
    response = payload.get("tool_response")
    if not isinstance(response, dict):
        response = payload.get("toolResponse")
    if not isinstance(response, dict):
        return False
    for key in ("is_error", "isError", "interrupted"):
        if response.get(key) is True:
            return True
    for key in ("exit_code", "exitCode"):
        code = response.get(key)
        if isinstance(code, int) and not isinstance(code, bool) and code != 0:
            return True
    return False


def observed(payload, rules, terminal):
    """The rules whose text this completed call put in front of the model."""
    base = workspace(payload)
    if call_failed(payload):
        return []
    if terminal:
        shown = displayed_files(tool_input(payload).get("command"), base)
        if not shown:
            return []
        return [entry for entry in rules
                if shows_reference(shown, entry["reference"])]
    if not read_starts_at_top(payload):
        return []
    targets = target_paths(payload)
    if not targets:
        return []
    return [entry for entry in rules if same_file(targets[0], entry["reference"])]


def rules_for(payload):
    """(gateable rules, read_only, state directory, harness) or None."""
    manifest = load_manifest()
    harness = detect_harness(payload)
    rules, read_only = deliverable(manifest, REFERENCE_ROOT[harness], agent_name(payload))
    if not rules:
        return None
    directory = state_dir(payload)
    if directory is None:
        return None
    if ensure_state_dir(directory):
        prune_stale_state()
    return rules, read_only, directory, harness


def observe(payload):
    """PostToolUse: record what the call that just ran actually showed."""
    tool = tool_name(payload)
    reading = tool in READ_TOOLS
    terminal = tool in TERMINAL_TOOLS
    if not (reading or terminal):
        return
    prepared = rules_for(payload)
    if prepared is None:
        return
    rules, _read_only, directory, _harness = prepared
    for entry in observed(payload, rules, terminal):
        claim(directory, known_marker(entry["name"]))


def hold(payload):
    """PreToolUse: deny a call whose rule this agent has not read.

    A write of a file a rule scopes, or a command a rule declares — both hold
    the same way, name the same reason and are released by the same read.
    """
    tool = tool_name(payload)
    writing = tool in WRITE_TOOLS or tool == PATCH_TOOL
    reading = tool in READ_TOOLS
    terminal = tool in TERMINAL_TOOLS and not writing
    if not (writing or reading or terminal):
        return

    targets = []
    command = ""
    if terminal:
        # `apply_patch` is a shell verb too: the same headers inside a shell
        # command author the same files. A call is therefore BOTH — `apply_patch
        # <<'PATCH' … PATCH` followed by `git commit` authors a file AND runs a
        # declared command — so both paths are evaluated and either one holds
        # it. Calling such a call "a write, not a command" let one call through
        # the command gate entirely.
        # No size limit here: the scan short-circuits on a substring test when
        # the text opens no patch, and a patch too big to LEX still authors the
        # files its headers name.
        targets = patch_targets(command_text(payload), workspace(payload))
        writing = bool(targets)
        # The RAW value, never the flattened text: joining an argv array erases
        # the boundary `bash -lc '<script>'` depends on.
        command = tool_input(payload).get("command")

    prepared = rules_for(payload)
    if prepared is None:
        return
    rules, read_only, directory, harness = prepared

    # Writers are gated on writes; an agent that cannot write is gated on its
    # first READ of a file a `readers` rule scopes, and on nothing else. A
    # COMMAND is gated for both — a reviewer drives `agent-browser` too — and
    # a rule it does not match gates nothing, so `git log` is never held.
    if reading and not read_only:
        return
    if writing and read_only and not terminal:
        return

    invocations = []
    if terminal:
        if any(entry["commands"] for entry in rules):
            invocations = hold_stages(command)
        if read_only:
            targets = []  # a reviewer is not gated on what a call AUTHORS
    else:
        targets = target_paths(payload)
        if not targets:
            return
        # Never deny the read the gate itself asked for: a `readers` rule whose
        # globs also match its own deployed text would otherwise wedge the agent.
        if reading and any(same_file(targets[0], entry["reference"]) for entry in rules):
            return
    targets = [path for path in (gate_path(target) for target in targets) if path]
    if not targets and not invocations:
        return

    fired = triggered(rules, targets, invocations)
    if fired and terminal and reads_only(invocations):
        # A call that ONLY reads is the observation the gate asked for. A mixed
        # `cat <rule> && git commit` is not: the commit runs before the model
        # has seen a byte, and denying it wedges nothing — `cat <rule>` alone
        # matches no prefix and was never denied — so the model re-issues the
        # two halves as two calls.
        shown = displayed_files(command, workspace(payload))
        if shown:
            fired = [entry for entry in fired
                     if not shows_reference(shown, entry["reference"])]
    if not fired:
        return
    reason = gate(directory, fired, harness)
    if reason:
        emit_deny(reason)


def main():
    if os.environ.get("HIVE_RULE_DELIVERY", "").strip().lower() in KILL_VALUES:
        return
    payload = json.load(sys.stdin)
    if not isinstance(payload, dict):
        return

    raw_event = payload.get("hook_event_name") or payload.get("hookEventName") or ""
    event = re.sub(r"[^a-z]", "", str(raw_event).lower())
    if event == "sessionstart":
        # A compaction drops what the session had read, so the gate re-arms.
        if str(payload.get("source") or "") in {"compact", "clear"}:
            clear_session(payload)
        prune_stale_state()
        return
    if event == "posttooluse":
        observe(payload)
        return
    # PI sends no event name at pre-tool; anything else that names one and is
    # neither of the above (a Stop, a PreCompact) decides nothing.
    if event and event != "pretooluse":
        return
    hold(payload)


if __name__ == "__main__":
    try:
        main()
    except StateUnavailable:
        sys.exit(0)  # never deny what cannot be recorded
    except Exception:  # noqa: BLE001 — fail open, never leak a traceback
        sys.exit(0)
