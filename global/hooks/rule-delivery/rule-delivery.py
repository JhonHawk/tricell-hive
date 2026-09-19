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
    """The subagent's roster name; absent means the main thread."""
    for key in ("agent_type", "agentType", "agent_name"):
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
    for key in ("cwd", "workspaceRoot", "workspace_root"):
        value = payload.get(key)
        if isinstance(value, str) and value:
            return value
    return ""


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


def patch_targets(command, base, limit=None):
    """Files a patch body authors.

    A header counts only when it starts a line AND the text opens a patch: a
    commit message quoting `*** Update File:` writes nothing. The substring
    check also keeps the scan off every large command that is not a patch.
    """
    if not isinstance(command, str) or PATCH_OPENING not in command:
        return []
    if limit is not None and len(command) > limit:
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
IGNORED_PREFIXES = {"sudo", "env", "command", "nohup", "time", "exec"}
PIPELINE_BREAKS = {";", "&&", "||", "&", "\n"}
ASSIGNMENT = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*=")
ZERO_COUNT = re.compile(r"^-(?:[nc])?0+$")


def lex(text):
    """Shell tokens with operators separated; raises ValueError if unparsable."""
    lexer = shlex.shlex(text, posix=True, punctuation_chars=True)
    lexer.whitespace_split = True
    return list(lexer)


def split_lines(text):
    """Lines of a multi-line command, honouring quotes.

    `shlex` swallows a newline as ordinary whitespace, which would glue
    `echo a` and `cat rule.md` into one nonsense command.
    """
    lines, current, quote = [], [], None
    for char in text:
        if quote:
            current.append(char)
            if char == quote:
                quote = None
            continue
        if char in "'\"":
            quote = char
            current.append(char)
            continue
        if char == "\n":
            lines.append("".join(current))
            current = []
            continue
        current.append(char)
    lines.append("".join(current))
    return [line for line in lines if line.strip()]


def strip_prefixes(argv):
    index = 0
    while index < len(argv) and (ASSIGNMENT.match(argv[index])
                                 or os.path.basename(argv[index]) in IGNORED_PREFIXES):
        index += 1
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


def pipelines(command, depth=0):
    """[[ (argv, redirected), … ], …] — one list per pipeline, stages in order."""
    if depth > 2:
        return []
    if isinstance(command, list):
        tokens = [str(part) for part in command if isinstance(part, (str, int, float))]
        raw = [[(tokens, False)]] if tokens else []
    else:
        raw = []
        for line in split_lines(command):
            stage, pipeline = [], []
            redirected = False
            skip_next = False
            for token in lex(line):
                if skip_next:
                    skip_next = False
                    continue
                if ">" in token:
                    # `2>` / `2>&1` redirect stderr; the model still sees stdout.
                    if stage and stage[-1] == "2":
                        stage.pop()
                    else:
                        redirected = True
                    skip_next = True
                    continue
                if "<" in token:
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
            script = None if redirected else wrapped_script(argv)
            if script is not None:
                resolved.extend(pipelines(script, depth + 1))
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
        # `cat "$HOME/…"` and `cat ~/…` are what a model actually types.
        target = resolve(os.path.expandvars(argument), base)
        if not target:
            continue
        try:
            shown.append((os.path.realpath(target), target))
        except (OSError, ValueError):
            continue
    return shown


def displayed_files(command, base):
    """Every file this shell call showed the model — parsed ONCE per call.

    The previous shape re-lexed the whole command for each rule in the
    manifest, which put a 1 MB heredoc at 5.4 s on every terminal call.
    """
    if isinstance(command, list):
        size = sum(len(str(part)) for part in command)
    elif isinstance(command, str):
        size = len(command)
    else:
        return []
    if size > MAX_COMMAND_CHARS:
        return []  # not parsed, so not observed
    try:
        parsed = pipelines(command)
    except ValueError:
        return []  # unparsable: never assume it was a read
    shown = []
    for pipeline in parsed:
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


def deliverable(manifest, root, agent):
    """(gateable rules in manifest order, is this agent read-only).

    A rule is gateable when its globs are a usable non-empty list of strings,
    it is not always-on, its text lives in the one store this hook owns, the
    agent does not already carry it as a pack, and the reference deployed for
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
        globs = entry.get("globs")
        if not isinstance(globs, list) or not globs:
            continue
        if not all(isinstance(item, str) and item for item in globs):
            continue
        if entry.get("always_on"):
            continue
        # A rule still under global/rules/ is loaded natively by Claude Code;
        # gating it would ask for a read of what the session already holds.
        if not str(entry.get("source", "")).startswith("global/rules-situational/"):
            continue
        if entry["name"] in packs:
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
        compiled = compile_globs(globs)
        if not compiled:
            continue  # a brace bomb: nobody can afford to match it
        gateable.append({
            "name": entry["name"],
            "globs": compiled,
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

def reason_for(references):
    return REASON_PREFIX + REASON_JOIN.join(references) + REASON_SUFFIX


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


def gate(directory, rules, targets, budget):
    """The denial reason for this call, or None to let it through."""
    pending = []
    for entry in rules:
        if not any(path_matches(target, entry["globs"]) for target in targets):
            continue
        if has_marker(directory, known_marker(entry["name"])):
            continue
        pending.append(entry)
    if not pending:
        return None

    # Least-counted first (manifest order breaks ties): under a budget too
    # tight to name every rule, the one left out last time goes first now.
    pending.sort(key=lambda entry: denials_so_far(directory, entry["name"]))

    named = []
    for entry in pending:
        references = [chosen["reference"] for chosen in named] + [entry["reference"]]
        if len(reason_for(references)) > budget:
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
    return reason_for(held) if held else None


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
    """(gateable rules, read_only, state directory) or None to do nothing."""
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
    return rules, read_only, directory, REASON_BUDGETS[harness]


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
    rules, _read_only, directory, _budget = prepared
    for entry in observed(payload, rules, terminal):
        claim(directory, known_marker(entry["name"]))


def hold(payload):
    """PreToolUse: deny a write of a file whose rule this agent has not read."""
    tool = tool_name(payload)
    writing = tool in WRITE_TOOLS or tool == PATCH_TOOL
    reading = tool in READ_TOOLS
    terminal = tool in TERMINAL_TOOLS and not writing
    if not (writing or reading or terminal):
        return

    targets = []
    if terminal:
        # `apply_patch` is a shell verb too: the same headers inside a shell
        # command author the same files. This is the ONE case in which a
        # terminal call is denied.
        targets = patch_targets(command_text(payload), workspace(payload),
                                limit=MAX_COMMAND_CHARS)
        if not targets:
            return
        writing = True

    prepared = rules_for(payload)
    if prepared is None:
        return
    rules, read_only, directory, budget = prepared

    # Writers are gated on writes; an agent that cannot write is gated on its
    # first READ of a file a `readers` rule scopes, and on nothing else.
    if (reading and not read_only) or (writing and read_only):
        return

    targets = targets or target_paths(payload)
    if not targets:
        return
    # Never deny the read the gate itself asked for: a `readers` rule whose
    # globs also match its own deployed text would otherwise wedge the agent.
    if reading and any(same_file(targets[0], entry["reference"]) for entry in rules):
        return
    reason = gate(directory, rules, targets, budget)
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
