#!/usr/bin/env python3
"""Record task progress and the build checkpoint in a plan file.

Usage:
  python3 -I tasks.py <plan.md> done T<n> --commit <sha>|none --unverified "<text>|none"
  python3 -I tasks.py <plan.md> checkpoint --stage <stage> --step <step> --commit <sha>|none
          --next "<action>" [--finding "<ID: decision>"]... [--check "<command> · <result> · <commit>"]...
          [--worktree "<path> · <branch>"] [--unwritten "<text>|none"]
          [--rules "<flow>/<packet>.md[, …][ · extra: <leaf>, …]"|none] [--principles "<name>, …"|none]
  python3 -I tasks.py <plan.md> execution init|check|update|validate-handoff --record FILE ...

<plan.md> is the one plan file of a work, `<YYYY-MM-DD>-<work>.plan.md`, in its folder
under `<support>/workspace/`. A folder is not accepted.

A task is a `### T<n> — <title>` heading; its boxes are the `- [ ]` and `- [x]`
lines up to the next `##` or `###` heading.

- done: ticks only that task's boxes, appends `<date> — T<n> · <sha> · unverified: <text>`
  to the `## ... progress` section unless a line for that task and commit is there,
  and prints `✓ T<n> · <tasks done>/<tasks> · <title>`.
- checkpoint: writes or replaces `## Checkpoint`, just before the progress section,
  with the fixed lines Updated, Stage (stage and step), Last commit, Open findings and
  Next, then `- Worktree: <path> · <branch>` when --worktree was given now or before,
  followed by `### Evidence`. Existing evidence lines and the Worktree line are kept;
  each --check adds one, a new --worktree replaces it and --worktree none removes it.
  --unwritten writes `- Unwritten: <text>` (what no file holds yet, asked at a pause) after
  the Worktree line; later calls keep it, a new value replaces it, `none` removes it.
  --rules writes `- Rules: <flow>/<packet>.md[, <flow>/<packet>.md][ · extra: <leaves>]`, the
  stage packets the flow read and the conditional leaves read besides them: later calls
  keep it, packets and leaves of the same flow join it without duplicates, a packet of
  another flow starts a new line, and `none` removes it.
  --principles writes `- Principles: <names>|none`, the code principles that changed a
  decision: new names join the line's names without duplicates, `none` is written only
  while the line holds no name and a name replaces it, and no value removes the line.
  Boxes under `## Delivery units` belong to no task: `done` never ticks them.
- execution: init, check, update and validate-handoff keep a local execution record (see
  ../references/execution-record.md). Its grant_ref names a section of this plan
  (`<YYYY-MM-DD>-<work>.plan.md#decisiones`), its criteria are the `AC<n>` lines of the plan,
  its plan_ref is this plan's path, and `init` links it from `## Checkpoint`. The record sits
  outside the plan's work folder, inside the repository's `_support/workspace`.

Writes the plan file directly. Exit status: 0 on success, 2 when the arguments are invalid,
the plan cannot be read or written, the task does not exist, or the progress section is
missing for `done`. Standard library only; self-contained.
"""

import argparse
import datetime
import copy
import hashlib
import importlib.util
import json
from contextlib import contextmanager
from types import SimpleNamespace
import os
import re
import sys
import tempfile
from pathlib import Path

TASK_HEADING = re.compile(r"^###\s+(T\d+)\s*(?:[—–:-]\s*)?(.*?)\s*$")
BOX = re.compile(r"^(\s*[-*]\s+\[)([ xX])(\]\s*.*)$")
HEADING = re.compile(r"^#{2,3}\s")
LEVEL2 = re.compile(r"^##\s")
PROGRESS = re.compile(r"^##\s.*(progress|progreso|avance)", re.IGNORECASE)
CHECKPOINT = re.compile(r"^##\s+Checkpoint\s*$", re.IGNORECASE)
EVIDENCE = re.compile(r"^###\s+Evidence\s*$", re.IGNORECASE)
SHA = re.compile(r"^[0-9a-fA-F]{7,40}$")
TASK_ID = re.compile(r"^T\d+$")


class Failure(Exception):
    pass


def single_line(value):
    if "\n" in value or "\r" in value:
        raise argparse.ArgumentTypeError("must be a single line")
    return value.strip()


def sha(value):
    if value == "none":
        return "uncommitted"  # a delivery that makes no commits, such as hold
    if not SHA.match(value):
        raise argparse.ArgumentTypeError("must be a 7 to 40 character hex commit id, or none")
    return value


def task_id(value):
    if not TASK_ID.match(value):
        raise argparse.ArgumentTypeError("must look like T2")
    return value


def read(path):
    try:
        return path.read_text(encoding="utf-8").splitlines()
    except (OSError, UnicodeDecodeError) as error:
        raise Failure(f"{path}: {error}")


def write(path, lines):
    text = "\n".join(lines) + "\n"
    temp = None
    try:
        fd, temp = tempfile.mkstemp(dir=str(path.parent), prefix=".plan-", suffix=".tmp")
        with os.fdopen(fd, "w", encoding="utf-8", newline="") as handle:
            handle.write(text)
        if path.exists():
            os.chmod(temp, path.stat().st_mode & 0o7777)
        os.replace(temp, path)
        temp = None
    except OSError as error:
        raise Failure(f"{path}: {error}")
    finally:
        if temp and os.path.exists(temp):
            os.unlink(temp)


def parse_tasks(lines):
    """Return [{id, title, start, end, boxes}] where boxes are line indexes."""
    tasks = []
    for i, line in enumerate(lines):
        match = TASK_HEADING.match(line)
        if match:
            tasks.append({"id": match.group(1), "title": match.group(2), "start": i})
    for task in tasks:
        end = len(lines)
        for j in range(task["start"] + 1, len(lines)):
            if HEADING.match(lines[j]):
                end = j
                break
        task["end"] = end
        task["boxes"] = [j for j in range(task["start"] + 1, end) if BOX.match(lines[j])]
    return tasks


def ticked(lines, task):
    return sum(1 for j in task["boxes"] if BOX.match(lines[j]).group(2) != " ")


def complete(lines, task):
    return bool(task["boxes"]) and ticked(lines, task) == len(task["boxes"])


def section_bounds(lines, heading):
    """Return (start, end) of the level-2 section whose heading matches, or None."""
    for i, line in enumerate(lines):
        if heading.match(line):
            end = len(lines)
            for j in range(i + 1, len(lines)):
                if LEVEL2.match(lines[j]):
                    end = j
                    break
            return i, end
    return None


def today():
    return datetime.date.today().isoformat()


def cmd_done(lines, args):
    tasks = parse_tasks(lines)
    task = next((t for t in tasks if t["id"] == args.task), None)
    if task is None:
        raise Failure(f"{args.task}: no such task in the plan")
    if not task["boxes"]:
        raise Failure(f"{args.task}: the task has no boxes to tick")
    bounds = section_bounds(lines, PROGRESS)
    if bounds is None:
        raise Failure("the plan has no progress section (a `##` heading with progress, progreso or avance)")
    for j in task["boxes"]:
        match = BOX.match(lines[j])
        lines[j] = f"{match.group(1)}x{match.group(3)}"
    start, end = bounds
    entry = f"{today()} — {args.task} · {args.commit} · unverified: {args.unverified}"
    if args.commit == "uncommitted":
        # Without a commit, only the same task, day and result is a repeat.
        seen = any(line.strip() == entry for line in lines[start + 1:end])
    else:
        known = re.compile(rf"(?<![A-Za-z0-9]){args.task}\s*·\s*{re.escape(args.commit)}(?![0-9A-Za-z])")
        seen = any(known.search(line) for line in lines[start + 1:end])
    if not seen:
        last = end
        while last > start + 1 and not lines[last - 1].strip():
            last -= 1
        if last == start + 1:  # an empty section: the entry goes after one blank line under the heading
            if last < end:
                last += 1
            else:
                lines.insert(last, "")
                last += 1
        lines.insert(last, entry)
    done = sum(1 for t in parse_tasks(lines) if complete(lines, t))
    return lines, f"✓ {task['id']} · {done}/{len(tasks)} · {task['title']}"


PACKET_PATH = r"[a-z0-9]+(?:-[a-z0-9]+)*/stage(?:-[a-z]+)?\.md"
RULES_VALUE = re.compile(rf"^{PACKET_PATH}(?:, {PACKET_PATH})*(?: · extra: [a-z0-9]+(?:-[a-z0-9]+)*(?:, [a-z0-9]+(?:-[a-z0-9]+)*)*)?$")
PRINCIPLE_NAME = r"[a-z0-9]+(?:-[a-z0-9]+)*"
PRINCIPLES_VALUE = re.compile(rf"^(?:none|(?!none(?:,|$)){PRINCIPLE_NAME}(?:, (?!none(?:,|$)){PRINCIPLE_NAME})*)$")


def rules_value(value):
    value = single_line(value)
    if value != "none" and not RULES_VALUE.match(value):
        raise argparse.ArgumentTypeError("expected '<flow>/stage[-<name>].md[, …][ · extra: <leaf>, …]' or 'none'")
    return value


def principles_value(value):
    value = single_line(value)
    if not PRINCIPLES_VALUE.match(value):
        raise argparse.ArgumentTypeError("expected 'none' or principle names joined by ', '")
    return value


def split_rules(value):
    packets, _, extra = value.partition(" · extra: ")
    return packets.split(", "), [e for e in extra.split(", ") if e]


def merge_rules(current, given):
    """Join the given packets and extra leaves to the line's while the flow stays the
    same, without duplicates and in order of arrival; a packet of another flow starts a
    new line, and `none` removes it."""
    if given is None:
        return current
    if given == "none":
        return None
    packets, extra = split_rules(given)
    if current:
        old_packets, old_extra = split_rules(current)
        flows = {p.split("/", 1)[0] for p in old_packets}
        if all(p.split("/", 1)[0] in flows for p in packets):
            packets = old_packets + [p for p in packets if p not in old_packets]
            extra = old_extra + [e for e in extra if e not in old_extra]
    packets = list(dict.fromkeys(packets))
    extra = list(dict.fromkeys(extra))
    return ", ".join(packets) + (" · extra: " + ", ".join(extra) if extra else "")


def merge_principles(current, given):
    """Join the given names to the line's, without duplicates and in order of arrival;
    `none` stands only while no name is known, and nothing removes the line."""
    if given is None:
        return current
    names = [n for n in (current or "").split(", ") if n and n != "none"]
    for name in given.split(", "):
        if name != "none" and name not in names:
            names.append(name)
    return ", ".join(names) if names else "none"


def cmd_checkpoint(lines, args):
    kept = []
    worktree = args.worktree
    unwritten = args.unwritten
    given_rules = getattr(args, "rules", None)
    rules = None
    principles = None
    execution = None
    existing = section_bounds(lines, CHECKPOINT)
    if existing:
        start, end = existing
        inside = False
        for line in lines[start + 1:end]:
            if not inside and line.startswith("- Execution record: "):
                execution = line
            elif worktree is None and not inside and line.startswith("- Worktree: "):
                worktree = line[len("- Worktree: "):]
            elif unwritten is None and not inside and line.startswith("- Unwritten: "):
                unwritten = line[len("- Unwritten: "):]
            elif not inside and line.startswith("- Rules: "):
                rules = line[len("- Rules: "):]
            elif not inside and line.startswith("- Principles: "):
                principles = line[len("- Principles: "):]
            elif EVIDENCE.match(line):
                inside = True
            elif HEADING.match(line):
                inside = False
            elif inside and line.strip():
                kept.append(line)
    rules = merge_rules(rules, given_rules)
    principles = merge_principles(principles, getattr(args, "principles", None))
    for check in args.check:
        entry = f"- {check}"
        if entry not in kept:
            kept.append(entry)
    findings = "; ".join(args.finding) if args.finding else "none"
    block = [
        "## Checkpoint",
        "",
        f"- Updated: {today()}",
        f"- Stage: {args.stage} · {args.step}",
        f"- Last commit: {args.commit}",
        f"- Open findings: {findings}",
        f"- Next: {args.next}",
        *([f"- Worktree: {worktree}"] if worktree and worktree != "none" else []),
        *([f"- Unwritten: {unwritten}"] if unwritten and unwritten != "none" else []),
        *([f"- Rules: {rules}"] if rules and rules != "none" else []),
        *([f"- Principles: {principles}"] if principles else []),
        *([execution] if execution else []),
        "",
        "### Evidence",
        "",
        *kept,
    ]
    if existing:
        start, end = existing
        lines[start:end] = block + ([""] if end < len(lines) else [])
    else:
        progress = section_bounds(lines, PROGRESS)
        if progress:
            lines[progress[0]:progress[0]] = block + [""]
        else:
            while lines and not lines[-1].strip():
                lines.pop()
            lines.extend([""] + block)
    return lines, f"checkpoint written · {args.stage} · {args.step} · {len(kept)} evidence line(s)"


# Execution records are local declarations, not credentials or a scheduler.
OWNER_ROLES = {"hive-build-owner", "hive-ticket-owner"}
FULL_SHA = re.compile(r"^[0-9a-f]{40}$")
UNIT_STATES = {"assigned", "running", "returned", "verified", "integrated", "blocked"}
TRANSITIONS = {
    "assigned": {"assigned", "running", "blocked"},
    "running": {"running", "returned", "blocked"},
    "returned": {"returned", "running", "verified", "blocked"},
    "verified": {"verified", "running", "integrated", "blocked"},
    "blocked": {"blocked", "running"},
    "integrated": {"integrated"},
}
ROOT_FIELDS = {
    "schema_version", "generation", "coordinator_id", "plan_ref", "plan_revision",
    "grant_ref", "integration_target", "units", "integration_scope", "integration_head",
    "joint_check", "publication",
}
UNIT_FIELDS = {
    "id", "tasks", "criteria", "depends_on", "assignment_id", "owner_id", "role",
    "repo", "worktree", "branch", "base_sha", "allowed_paths", "resources", "state",
    "returned_head", "evidence", "verifier", "limitations", "next_action", "cleanup",
    "history", "closure",
}


def require(condition, message):
    if not condition:
        raise Failure(message)


def nonempty(value):
    return isinstance(value, str) and bool(value.strip()) and not any(c in value for c in "\n\r\0")


def unique_strings(value, label, allow_empty=True):
    require(isinstance(value, list) and all(nonempty(v) for v in value), f"{label}: expected string list")
    require(len(set(value)) == len(value), f"{label}: duplicate values")
    require(allow_empty or bool(value), f"{label}: must not be empty")
    return value


def object_fields(value, required, optional, label):
    require(isinstance(value, dict), f"{label}: expected object")
    require(not (set(value) - required - optional), f"{label}: unknown fields")
    require(required <= set(value), f"{label}: missing fields: {', '.join(sorted(required - set(value)))}")


def json_object(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result, f"duplicate JSON key: {key}")
        result[key] = value
    return result


def load_record(path):
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=json_object)


def plan_revision(plan):
    """Hash the plan, ignoring mutable task progress: the boxes' ticks, `## Checkpoint` and the progress section."""
    kept, inside = [], False
    for line in read(plan):
        if LEVEL2.match(line):
            inside = bool(PROGRESS.match(line) or CHECKPOINT.match(line))
        if not inside:
            kept.append(re.sub(r"^(\s*[-*]\s+\[)[xX](\])", r"\1 \2", line))
    digest = hashlib.sha256()
    digest.update(plan.name.encode())
    digest.update(b"\0")
    digest.update("\n".join(kept).rstrip().encode())
    digest.update(b"\0")
    return digest.hexdigest()


def sibling_script(relative, name):
    source = Path(__file__).resolve().parents[2] / relative
    spec = importlib.util.spec_from_file_location(name, source)
    module = importlib.util.module_from_spec(spec)
    previous = sys.dont_write_bytecode
    try:
        sys.dont_write_bytecode = True
        spec.loader.exec_module(module)
    finally:
        sys.dont_write_bytecode = previous
    return module


def git_reader(repo):
    # Use the existing timeout, safe environment and factual Git query boundary.
    module = sibling_script("git-workflow/scripts/git_state.py", "hive_git_state")
    state = module.State(SimpleNamespace(repo=str(repo), timeout=20))
    require(state.git is not None, "git executable unavailable")
    return state


def git_out(state, *args):
    value, error = state.git_out(*args)
    require(error is None, f"Git check failed: {error}")
    return value


def absolute_directory(value, label):
    require(nonempty(value) and Path(value).is_absolute(), f"{label}: expected absolute directory")
    path = Path(value)
    require(path.is_dir(), f"{label}: directory unavailable")
    return path.resolve()


def repo_identity(repo):
    state = git_reader(repo)
    top = Path(git_out(state, "rev-parse", "--show-toplevel")).resolve()
    require(top == Path(repo).resolve(), "repository/worktree must name its root")
    common = Path(git_out(state, "rev-parse", "--git-common-dir"))
    return state, (Path(repo) / common).resolve()


def execution_location(plan, record):
    # The record belongs to the checkout that holds it, which may be the coordinator's
    # checkout while the plan lies in a worktree or another repository.
    folder = plan.resolve().parent
    require(record.parent.is_dir(), "record directory must already exist")
    state = git_reader(record.parent.resolve())
    root = Path(git_out(state, "rev-parse", "--show-toplevel")).resolve()
    require(record.name == "execution.json", "record must be named execution.json")
    require(not record.is_symlink(), "execution record must not be a symlink")
    workspace = root / "_support/workspace"
    require(not workspace.is_symlink() and workspace.resolve().is_relative_to(root), "workspace must stay inside repository without symlink escape")
    require(record.resolve().is_relative_to(workspace.resolve()), "record must be inside this repository's _support/workspace")
    require(not record.resolve().is_relative_to(folder), "execution record must be outside the plan's work folder")
    # Check both the record and an arbitrary sibling (including the lock) before writing.
    for path in (record, record.parent / ".execution-local-check"):
        code, _, _ = state.git_run("check-ignore", "-q", "--", str(path.resolve()))
        require(code == 0, "execution workspace must already be Git-ignored")
    code, tracked, _ = state.git_run("ls-files", "--", str(record.resolve()))
    require(code == 0 and not tracked, "execution record must not be tracked")
    return root


def local_evidence(base, value, available=True):
    require(nonempty(value), "evidence path missing")
    path = Path(value)
    require(not path.is_absolute() and ".." not in path.parts, "evidence path must stay in record directory")
    resolved = (base / path).resolve()
    require(resolved.is_relative_to(base.resolve()) and (not available or resolved.is_file()), "evidence file unavailable or escapes record directory")
    return resolved


def allowed_path(value):
    require(nonempty(value), "allowed path missing")
    path = Path(value)
    require(not path.is_absolute() and ".." not in path.parts and path.parts,
            "allowed paths must be repository-relative without ..")
    require(value != "." and not any(c in value for c in "*?[]\\"), "allowed paths must be literal files or directory/ prefixes")
    require(path.as_posix() + ("/" if value.endswith("/") else "") == value, "allowed path must be normalized")
    return value


def path_covers(prefix, path):
    return path == prefix or (prefix.endswith("/") and path.startswith(prefix))


def paths_overlap(first, second):
    return any(path_covers(a, b) or path_covers(b, a) for a in first for b in second)


def evidence_check(check, base, head=None, available=True):
    required = {"command", "result", "path"} | ({"head"} if head is not None else set())
    object_fields(check, required, set(), "check")
    require(nonempty(check["command"]) and check["result"] == "pass", "check needs a command and passing result")
    if head is not None:
        require(check["head"] == head, "joint check head differs from integration head")
    local_evidence(base, check["path"], available=available)


def validate_verifier(unit, base, available=True):
    verdict = unit["verifier"]
    object_fields(verdict, {"id", "head", "criteria", "checks"}, set(), "verifier")
    owners = {unit["owner_id"], *(prior["owner_id"] for prior in unit.get("history", []))}
    require(nonempty(verdict["id"]) and verdict["id"] not in owners,
            "verifier must differ from current and historical owners")
    require(verdict["head"] == unit["returned_head"], "verifier head differs from returned head")
    unique_strings(verdict["criteria"], "verifier criteria", False)
    require(set(verdict["criteria"]) == set(unit["criteria"]), "verifier must cover exact unit criteria")
    require(isinstance(verdict["checks"], list) and verdict["checks"], "verifier checks missing")
    for check in verdict["checks"]:
        evidence_check(check, base, available=available)


def validate_handoff(unit):
    repo = absolute_directory(unit["repo"], "repo")
    worktree = absolute_directory(unit["worktree"], "worktree")
    _, common = repo_identity(repo)
    state, actual_common = repo_identity(worktree)
    require(common == actual_common, "worktree belongs to another repository")
    require(git_out(state, "symbolic-ref", "--quiet", "--short", "HEAD") == unit["branch"], "handoff branch differs from assignment")
    head = unit["returned_head"]
    require(isinstance(head, str) and FULL_SHA.fullmatch(head), "handoff needs full returned head")
    require(git_out(state, "rev-parse", "--verify", "HEAD") == head, "handoff head differs from worktree HEAD")
    base = unit["base_sha"]
    require(git_out(state, "rev-parse", "--verify", base + "^{commit}") == base, "base is not an exact commit")
    ancestor, error = state.is_ancestor(base, head)
    require(error is None and ancestor, "handoff head does not descend from assigned base")
    code, dirty, _ = state.git_run("status", "--porcelain=v1", "-z", "--untracked-files=all")
    require(code == 0 and not dirty, "handoff worktree must be clean")
    code, changed, _ = state.git_run("diff", "--no-ext-diff", "--no-renames", "--name-only", "-z", base, head, "--")
    require(code == 0 and len(changed.encode("utf-8")) < 8 * 1024 * 1024, "cannot read complete handoff changed paths")
    for path in changed.split("\0"):
        if path:
            require("\ufffd" not in path, "handoff changed paths contain undecodable names")
            require(any(path_covers(p, path) for p in unit["allowed_paths"]), "handoff changed a path outside allowed paths")
            require((worktree / path).resolve().is_relative_to(worktree), "handoff path escapes worktree through symlink")


def validate_integration(data, record):
    target = data["integration_target"]
    state, identity = repo_identity(absolute_directory(target["worktree"], "integration worktree"))
    _, repo = repo_identity(absolute_directory(target["repo"], "integration repo"))
    require(identity == repo, "integration worktree belongs to another repository")
    require(git_out(state, "symbolic-ref", "--quiet", "--short", "HEAD") == target["branch"], "integration branch mismatch")
    head = data.get("integration_head")
    require(isinstance(head, str) and FULL_SHA.fullmatch(head), "integration needs full integration_head")
    require(git_out(state, "rev-parse", "HEAD") == head, "integration head differs from target HEAD")
    require(data.get("integration_scope") in {"local", "remote"}, "integration_scope must be local or remote")
    code, dirty, _ = state.git_run("status", "--porcelain=v1", "-z", "--untracked-files=all")
    require(code == 0 and not dirty, "integration target must be clean")
    evidence_check(data.get("joint_check"), record.parent, head)
    for unit in data["units"]:
        if unit["state"] == "integrated":
            _, unit_repo = repo_identity(unit["repo"])
            require(repo == unit_repo, "integrated unit belongs to another repository")
            ancestor, error = state.is_ancestor(unit["returned_head"], head)
            require(error is None and ancestor, "integration head does not contain verified unit head")


def validate_execution(data, plan, record, check_git=True, revision=True):
    object_fields(data, ROOT_FIELDS - {"integration_scope", "integration_head", "joint_check", "publication"},
                  {"integration_scope", "integration_head", "joint_check", "publication"}, "execution")
    require(type(data["schema_version"]) is int and data["schema_version"] == 1, "unsupported schema_version")
    require(type(data["generation"]) is int and data["generation"] >= 0, "invalid generation")
    require(nonempty(data["coordinator_id"]), "coordinator_id missing")
    require(data["plan_ref"] == str(plan.resolve()), "plan_ref differs from the plan path")
    if revision:
        require(data["plan_revision"] == plan_revision(plan), "plan revision changed; reconcile assignments with --reconcile-plan")
    require(isinstance(data["plan_revision"], str) and re.fullmatch(r"[0-9a-f]{64}", data["plan_revision"]), "invalid plan_revision")
    require(data.get("integration_scope") is None or data["integration_scope"] in {"local", "remote"}, "integration_scope must be local or remote")
    require(data.get("integration_head") is None or isinstance(data["integration_head"], str) and FULL_SHA.fullmatch(data["integration_head"]), "integration_head must be null or full SHA")
    require(nonempty(data["grant_ref"]), "grant_ref missing")
    require(data["grant_ref"].split("#", 1)[0] == plan.name, "grant_ref must point inside the plan: <plan file name>#<section>")
    target = data["integration_target"]
    object_fields(target, {"repo", "worktree", "branch"}, set(), "integration_target")
    for field in target:
        require(nonempty(target[field]), "integration target field missing")
    plan_lines = read(plan)
    task_ids = {task["id"] for task in parse_tasks(plan_lines)}
    criteria = set(re.findall(r"^\s*(?:[-*]\s+)?\**(AC\d+)\b", "\n".join(plan_lines[:next((i for i, line in enumerate(plan_lines) if TASK_HEADING.match(line)), len(plan_lines))]), re.MULTILINE))
    units = data["units"]
    require(isinstance(units, list) and units, "units must be a nonempty list")
    ids, assignments, active = set(), set(), []
    for unit in units:
        object_fields(unit, UNIT_FIELDS - {"history", "closure"}, {"history", "closure"}, "unit")
        for field in ("id", "assignment_id", "owner_id", "role", "repo", "worktree", "branch", "next_action"):
            require(nonempty(unit[field]), f"unit {field} missing")
        require(unit["id"] not in ids, "duplicate unit ID")
        ids.add(unit["id"])
        require(unit["assignment_id"] not in assignments, "duplicate assignment ID")
        assignments.add(unit["assignment_id"])
        require(unit["role"] in OWNER_ROLES, "unit role must be one of " + ", ".join(sorted(OWNER_ROLES)))
        require(unit["state"] in UNIT_STATES, "unknown unit state")
        for field in ("tasks", "criteria", "allowed_paths"):
            unique_strings(unit[field], field, False)
        for field in ("depends_on", "resources", "evidence", "limitations"):
            unique_strings(unit[field], field)
        require(not revision or set(unit["tasks"]) <= task_ids, "unit references unknown task")
        require(not revision or set(unit["criteria"]) <= criteria, "unit references unknown criterion")
        require(isinstance(unit["base_sha"], str) and FULL_SHA.fullmatch(unit["base_sha"]), "base_sha must be full SHA")
        require(unit["returned_head"] is None or isinstance(unit["returned_head"], str) and FULL_SHA.fullmatch(unit["returned_head"]), "returned_head must be null or full SHA")
        for path in unit["allowed_paths"]:
            allowed_path(path)
        require(isinstance(unit.get("history", []), list), "history must be a list")
        for prior in unit.get("history", []):
            require(isinstance(prior, dict) and nonempty(prior.get("assignment_id")), "invalid assignment history")
            require(prior["assignment_id"] not in assignments, "reused assignment ID in history")
            assignments.add(prior["assignment_id"])
        for path in unit["evidence"]:
            local_evidence(record.parent, path, available=check_git)
        if unit["state"] in {"returned", "verified", "integrated"}:
            require(unit["returned_head"] is not None and unit["evidence"], "returned unit needs head and evidence")
        if unit["state"] in {"verified", "integrated"}:
            validate_verifier(unit, record.parent, available=check_git)
            if check_git and unit["state"] == "verified":
                validate_handoff(unit)
        elif unit["verifier"] is not None:
            # A blocked unit can preserve a verdict as historical evidence only.
            require(unit["state"] == "blocked", "unverified unit must not retain a verifier")
        if unit["state"] != "integrated":
            repo = absolute_directory(unit["repo"], "repo")
            worktree = absolute_directory(unit["worktree"], "worktree")
            _, common = repo_identity(repo)
            _, wt_common = repo_identity(worktree)
            require(common == wt_common, "assignment worktree belongs to another repository")
            for old, old_repo, old_wt in active:
                require(worktree != old_wt and not worktree.is_relative_to(old_wt) and not old_wt.is_relative_to(worktree), "active worktrees conflict")
                if common == old_repo:
                    require(unit["branch"] != old["branch"], "active branches conflict")
                    require(not paths_overlap(unit["allowed_paths"], old["allowed_paths"]), "active allowed paths conflict")
                require(not set(unit["resources"]) & set(old["resources"]), "active exclusive resources conflict")
            active.append((unit, common, worktree))
    by_id = {u["id"]: u for u in units}
    dependencies = {uid: set(unit["depends_on"]) for uid, unit in by_id.items()}
    require(all(deps <= set(by_id) for deps in dependencies.values()), "unknown dependency unit")
    ready = [uid for uid, deps in dependencies.items() if not deps]
    visited = set()
    while ready:
        uid = ready.pop()
        visited.add(uid)
        for other, deps in dependencies.items():
            if uid in deps:
                deps.remove(uid)
                if not deps and other not in visited:
                    ready.append(other)
    require(len(visited) == len(by_id), "unit dependencies contain a cycle")
    if any(u["state"] == "integrated" for u in units) and check_git:
        validate_integration(data, record)
    publication = data.get("publication")
    if publication is not None:
        object_fields(publication, {"head", "result", "path"}, set(), "publication")
        require(isinstance(publication["head"], str) and FULL_SHA.fullmatch(publication["head"]), "publication needs full head")
        require(publication["result"] in {"pass", "fail", "blocked"}, "invalid publication result")
        local_evidence(record.parent, publication["path"], available=check_git)


def transition_execution(old, new, record, reconcile=False):
    require(old["coordinator_id"] == new["coordinator_id"], "coordinator identity cannot change")
    require(old["integration_target"] == new["integration_target"], "integration target cannot change during execution")
    by_id = {u["id"]: u for u in new["units"]}
    require(set(by_id) == {u["id"] for u in old["units"]}, "update must preserve unit IDs")
    for previous in old["units"]:
        unit = by_id[previous["id"]]
        require(unit.get("history", []) == previous.get("history", []), "assignment history is immutable")
        replacement = unit["assignment_id"] != previous["assignment_id"]
        if replacement:
            closure = unit.get("closure")
            object_fields(closure, {"assignment_id", "owner_id", "kind", "path"}, set(), "replacement closure")
            require(closure["assignment_id"] == previous["assignment_id"] and closure["owner_id"] == previous["owner_id"], "closure identifies another assignment")
            require(closure["kind"] in {"completed", "stop_ack"}, "replacement needs native completion or stop_ack")
            local_evidence(record.parent, closure["path"])
            require(previous["state"] != "integrated" and unit["state"] == "assigned", "replacement must create assigned unit")
            require(unit["returned_head"] == previous["returned_head"], "replacement must preserve partial head")
            require(unit["evidence"] == previous["evidence"] and unit["limitations"] == previous["limitations"], "replacement must preserve partial evidence and limitations")
            snapshot = copy.deepcopy(previous)
            snapshot.pop("history", None)
            snapshot["closure"] = closure
            unit["history"] = [*previous.get("history", []), snapshot]
            unit["verifier"] = None
            unit["closure"] = None
        else:
            for field in ("owner_id", "role", "repo", "worktree", "branch", "base_sha", "allowed_paths", "resources"):
                require(unit[field] == previous[field], f"assignment {field} changed without replacement")
            for field in ("tasks", "criteria", "depends_on"):
                require(unit[field] == previous[field] or reconcile, f"assignment {field} changed without replacement")
            require(unit["state"] in TRANSITIONS[previous["state"]] or reconcile, "invalid state transition")
            if unit["state"] == "running" or (unit["returned_head"] != previous["returned_head"] and unit["state"] not in {"verified", "integrated"}):
                unit["verifier"] = None
            if unit["returned_head"] != previous["returned_head"]:
                require(unit["state"] not in {"verified", "integrated"}, "changed head must return before independent verification")
            if unit["state"] in {"verified", "integrated"}:
                require(previous["state"] in {"returned", "verified", "integrated"}, "verified requires a returned handoff")
            if unit["state"] == "integrated":
                require(previous["state"] in {"verified", "integrated"}, "integrated requires prior verified state")
                if previous["state"] == "verified":
                    validate_handoff(unit)
        if reconcile:
            require(unit["state"] == "blocked", "plan reconciliation requires every unit blocked")
            unit["verifier"] = None


@contextmanager
def execution_lock(record):
    lock = record.with_name(record.name + ".lock")
    try:
        lock.mkdir(mode=0o700)
    except FileExistsError:
        raise Failure("execution writer lock exists; reconcile its owner, never remove it by age")
    try:
        yield
    finally:
        lock.rmdir()


def atomic_record(record, data):
    fd, temp = tempfile.mkstemp(dir=record.parent, prefix=".execution-", suffix=".tmp")
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            json.dump(data, handle, indent=2, ensure_ascii=False)
            handle.write("\n")
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(temp, record)
    finally:
        if os.path.exists(temp):
            os.unlink(temp)


def link_execution(plan, record):
    lines = read(plan)
    pointer = "- Execution record: " + os.path.relpath(record, plan.resolve().parent)
    existing = [i for i, line in enumerate(lines) if line.startswith("- Execution record: ")]
    require(len(existing) <= 1, "checkpoint has multiple execution record pointers")
    if existing:
        require(lines[existing[0]] == pointer, "checkpoint already points to another execution record")
        return
    bounds = section_bounds(lines, CHECKPOINT)
    if not bounds:
        lines, _ = cmd_checkpoint(lines, SimpleNamespace(stage="build", step="E2E", commit="uncommitted",
                                                       next="dispatch units", finding=[], check=[], worktree=None, unwritten=None,
                                                       rules=None, principles=None))
        bounds = section_bounds(lines, CHECKPOINT)
    lines.insert(bounds[0] + 2, pointer)
    write(plan, lines)


def cmd_execution(args):
    plan, record = args.plan.resolve(), args.record.absolute()
    execution_location(plan, record)
    if args.operation in {"check", "validate-handoff"}:
        data = load_record(record)
        validate_execution(data, plan, record)
        if args.operation == "validate-handoff":
            unit = next((u for u in data["units"] if u["id"] == args.unit), None)
            require(unit is not None, "unknown unit")
            validate_handoff(unit)
        print(f"execution {args.operation} · generation {data['generation']} · metadata checked; inspect evidence independently")
        return
    with execution_lock(record):
        operation_revision = plan_revision(plan)
        data = load_record(args.input)
        require(isinstance(data, dict), "execution input must be an object")
        if args.operation == "init":
            require(not record.exists(), "execution record already exists; refusing overwrite")
            for field, value in {"schema_version": 1, "generation": 0, "coordinator_id": args.coordinator,
                                 "plan_ref": str(plan), "plan_revision": operation_revision}.items():
                require(field not in data or data[field] == value, f"init {field} mismatch")
                data[field] = value
            data.setdefault("publication", None)
            for unit in data.get("units", []):
                require(isinstance(unit, dict) and unit.get("state") == "assigned", "init units must be assigned")
                require(not unit.get("history") and unit.get("verifier") is None, "init cannot invent assignment history or verdict")
            validate_execution(data, plan, record)
            # Link first: a crash can leave a missing pointer target, never an orphaned authoritative record.
            link_execution(plan, record)
        else:
            old = load_record(record)
            validate_execution(old, plan, record, check_git=False, revision=False)
            require(args.coordinator == old["coordinator_id"] == data.get("coordinator_id"), "only recorded coordinator may update")
            require(args.expected_generation == old["generation"] == data.get("generation"), "stale generation; read and reconcile current record")
            require(data.get("plan_revision") == old["plan_revision"], "input plan revision differs from recorded revision")
            drift = old["plan_revision"] != operation_revision
            require(not drift or args.reconcile_plan, "plan revision changed; reconcile assignments with --reconcile-plan")
            require(not args.reconcile_plan or drift, "plan revision has not changed")
            transition_execution(old, data, record, reconcile=args.reconcile_plan)
            data["generation"] += 1
            data["plan_revision"] = operation_revision
            validate_execution(data, plan, record)
        require(plan_revision(plan) == operation_revision, "plan revision changed during operation; read and reconcile before retrying")
        atomic_record(record, data)
    print(f"execution {args.operation} · generation {data['generation']}")


def build_parser():
    parser = argparse.ArgumentParser(prog="tasks.py", description="Record task progress in a plan file.")
    parser.add_argument("plan", type=Path, help="the plan file, <YYYY-MM-DD>-<work>.plan.md")
    sub = parser.add_subparsers(dest="command", required=True)
    done = sub.add_parser("done")
    done.add_argument("task", type=task_id)
    done.add_argument("--commit", required=True, type=sha)
    done.add_argument("--unverified", required=True, type=single_line)
    cp = sub.add_parser("checkpoint")
    cp.add_argument("--stage", required=True, type=single_line)
    cp.add_argument("--step", required=True, type=single_line)
    cp.add_argument("--commit", required=True, type=sha)
    cp.add_argument("--next", required=True, type=single_line)
    cp.add_argument("--finding", action="append", default=[], type=single_line)
    cp.add_argument("--check", action="append", default=[], type=single_line)
    cp.add_argument("--worktree", type=single_line, help="'<path> · <branch>' of the work's worktree, kept by later calls; 'none' removes it")
    cp.add_argument("--unwritten", type=single_line, help="what no file holds yet, asked at a pause, kept by later calls; 'none' removes it")
    cp.add_argument("--rules", type=rules_value, help="'<flow>/<packet>.md[, …][ · extra: <leaf>, …]' the flow read; joins the line while the flow stays the same, starts a new line for another flow; 'none' removes it")
    cp.add_argument("--principles", type=principles_value, help="principle names that changed a decision, joined by ', ', merged into the line; 'none' only while it has none")
    execution = sub.add_parser("execution", help="coordinator-owned local execution record")
    operations = execution.add_subparsers(dest="operation", required=True)
    for name in ("init", "check", "update", "validate-handoff"):
        command = operations.add_parser(name)
        command.add_argument("--record", required=True, type=Path)
        if name in ("init", "update"):
            command.add_argument("--input", required=True, type=Path)
            command.add_argument("--coordinator", required=True, type=single_line)
        if name == "update":
            command.add_argument("--expected-generation", required=True, type=int)
            command.add_argument("--reconcile-plan", action="store_true")
        if name == "validate-handoff":
            command.add_argument("--unit", required=True)
    return parser


def main(argv):
    args = build_parser().parse_args(argv)
    path = args.plan
    try:
        if not path.is_file():
            raise Failure(f"{path}: not a plan file (pass the path of <YYYY-MM-DD>-<work>.plan.md, not a folder)")
        lines = read(path)
        if args.command == "execution":
            cmd_execution(args)
            return 0
        handler = {"done": cmd_done, "checkpoint": cmd_checkpoint}[args.command]
        result = handler(lines, args)
        if result is not None:
            new_lines, message = result
            write(path, new_lines)
            print(message)  # only once the plan holds the change
    except (Failure, OSError, UnicodeError, ValueError, TypeError, KeyError) as error:
        print(f"tasks: {error}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
