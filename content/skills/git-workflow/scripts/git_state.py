#!/usr/bin/env python3
"""Report a work branch's Git and pull-request state as one JSON object.

Usage:
  python3 -I git_state.py --base BRANCH [--repo DIR] [--remote NAME] [--fetch]
                          [--pr N | --no-pr] [--since ISO-8601] [--timeout SECONDS]

Read-only: it changes nothing but the remote-tracking refs that --fetch
updates. It reports facts and decides nothing: which commits belong to the
work, which reviewer is a bot and whether a thread still applies stay with the
git-workflow skill. Each review and comment carries `on_head` (its commit is
the pull request's head) and `after_since` (later than --since).

Exit status: 0 when every source answered, 1 when one failed (its field holds
`error`; the others are kept), 2 on invalid usage or when DIR is not a Git
repository.
"""

import argparse
import json
import os
import re
import shutil
import signal
import subprocess
import sys
from datetime import datetime, timezone
from urllib.parse import quote

FETCH_TIMEOUT = 60
COMMAND_TIMEOUT = 20
MAX_OUTPUT = 8 * 1024 * 1024
MAX_COMMITS = 200

SAFE_ENV = {
    "GIT_OPTIONAL_LOCKS": "0",
    "GIT_TERMINAL_PROMPT": "0",
    "GCM_INTERACTIVE": "never",
    "GH_PROMPT_DISABLED": "1",
    "GH_NO_UPDATE_NOTIFIER": "1",
    "GH_PAGER": "cat",
    "PAGER": "cat",
    "NO_COLOR": "1",
    "LC_ALL": "C",
}


class Runner:
    """Runs one program per call: no shell, no stdin, no controlling terminal,
    a time limit that kills the whole process group, and a capped output."""

    def __init__(self, cap):
        self.cap = cap
        self.env = None

    def environment(self, git):
        if self.env is None:
            env = dict(os.environ)
            # Drop the variables that would point git at another repository,
            # so --repo always decides which one is read.
            try:
                listed = subprocess.run(
                    [git, "rev-parse", "--local-env-vars"],
                    capture_output=True, text=True, timeout=COMMAND_TIMEOUT,
                    stdin=subprocess.DEVNULL, cwd=os.path.abspath(os.sep),
                ).stdout.split()
            except (OSError, subprocess.SubprocessError):
                listed = []
            for key in listed:
                env.pop(key, None)
            env.update(SAFE_ENV)
            self.env = env
        return self.env

    def run(self, argv, timeout, env):
        limit = min(timeout, self.cap) if self.cap else timeout
        try:
            proc = subprocess.Popen(
                argv, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                stderr=subprocess.PIPE, env=env, start_new_session=True,
            )
        except OSError as exc:
            return None, "", f"cannot run {os.path.basename(argv[0])}: {exc.strerror}"
        try:
            out, err = proc.communicate(timeout=limit)
        except subprocess.TimeoutExpired:
            try:
                os.killpg(proc.pid, signal.SIGKILL)
            except OSError:
                pass
            proc.communicate()
            return None, "", f"{os.path.basename(argv[0])} {argv[1] if len(argv) > 1 else ''} timed out after {limit:g}s".replace("  ", " ")
        out = out[:MAX_OUTPUT].decode("utf-8", "replace")
        err = err[:4096].decode("utf-8", "replace").strip()
        return proc.returncode, out, err


class State:
    def __init__(self, args):
        self.args = args
        self.runner = Runner(args.timeout)
        self.git = shutil.which("git")
        self.gh = shutil.which("gh")
        self.failed = False

    # -- git ---------------------------------------------------------------

    def git_run(self, *args, timeout=COMMAND_TIMEOUT):
        env = self.runner.environment(self.git)
        argv = [self.git, "-C", self.args.repo, "-c", "gc.auto=0", "-c", "maintenance.auto=false", *args]
        return self.runner.run(argv, timeout, env)

    def git_out(self, *args):
        code, out, err = self.git_run(*args)
        if code != 0:
            return None, err or f"git {args[0]} exited {code}"
        return out.strip(), None

    def is_ancestor(self, ancestor, ref):
        code, _, err = self.git_run("merge-base", "--is-ancestor", ancestor, ref)
        if code == 0:
            return True, None
        if code == 1:
            return False, None
        return None, err or "git merge-base failed"

    def git_state(self):
        a = self.args
        state = {"repo": None, "branch": None, "head": None, "upstream": None, "dirty": None,
                 "base": f"{a.remote}/{a.base}", "fetch": {"ran": False, "ok": None, "error": None},
                 "ahead": None, "behind": None, "commits_ahead": None, "commits_ahead_truncated": False,
                 "base_in_head": None, "pushed_head": None, "base_in_pushed_head": None, "errors": []}
        top, err = self.git_out("rev-parse", "--show-toplevel")
        if top is None:
            return None, err
        state["repo"] = top
        code, out, _ = self.git_run("symbolic-ref", "--quiet", "--short", "HEAD")
        state["branch"] = out.strip() if code == 0 else None
        state["head"], err = self.git_out("rev-parse", "--verify", "--quiet", "HEAD")
        if err:
            state["errors"].append(f"head: {err}")
        code, out, _ = self.git_run("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
        state["upstream"] = out.strip() if code == 0 else None
        state["dirty"] = self.dirty(state)

        compare = True
        if a.fetch:
            state["fetch"]["ran"] = True
            code, _, err = self.git_run("fetch", "--quiet", "--no-tags", "--no-write-fetch-head",
                                        "--no-recurse-submodules", a.remote, timeout=FETCH_TIMEOUT)
            state["fetch"]["ok"] = code == 0
            if code != 0:
                state["fetch"]["error"] = err or "git fetch failed"
                compare = False

        base_ref = f"refs/remotes/{a.remote}/{a.base}"
        if compare and state["head"]:
            exists, _ = self.git_out("rev-parse", "--verify", "--quiet", base_ref)
            if exists is None:
                state["errors"].append(f"compare: {state['base']} not found; fetch it or check --base")
            else:
                self.compare(state, base_ref)

        if state["branch"]:
            pushed_ref = f"refs/remotes/{a.remote}/{state['branch']}"
            state["pushed_head"], _ = self.git_out("rev-parse", "--verify", "--quiet", pushed_ref)
            if compare and state["pushed_head"] and state["behind"] is not None:
                state["base_in_pushed_head"], err = self.is_ancestor(base_ref, pushed_ref)
                if err:
                    state["errors"].append(f"base_in_pushed_head: {err}")
        if state["errors"] or state["fetch"]["ok"] is False:
            self.failed = True
        return state, None

    def dirty(self, state):
        code, out, err = self.git_run("status", "--porcelain=v1", "-z", "--untracked-files=all")
        if code != 0:
            state["errors"].append(f"dirty: {err or 'git status failed'}")
            return None
        staged = unstaged = untracked = 0
        tokens = out.split("\0")
        i = 0
        while i < len(tokens):
            entry = tokens[i]
            i += 1
            if len(entry) < 3:
                continue
            x, y = entry[0], entry[1]
            if x == "?" and y == "?":
                untracked += 1
                continue
            if x not in " ?":
                staged += 1
            if y not in " ?":
                unstaged += 1
            if x in "RC":
                i += 1  # the rename or copy source follows as its own token
        return {"staged": staged, "unstaged": unstaged, "untracked": untracked}

    def compare(self, state, base_ref):
        counts, err = self.git_out("rev-list", "--left-right", "--count", f"{base_ref}...HEAD")
        if counts is None:
            state["errors"].append(f"ahead/behind: {err}")
            return
        behind, ahead = (int(n) for n in counts.split())
        state["behind"], state["ahead"] = behind, ahead
        log, err = self.git_out("log", f"--max-count={MAX_COMMITS}", "--format=%H%x1f%s%x1f%an <%ae>%x1f%cI%x1e",
                                f"{base_ref}..HEAD")
        if log is None:
            state["errors"].append(f"commits_ahead: {err}")
        else:
            commits = []
            for record in log.split("\x1e"):
                record = record.strip("\n")
                if not record:
                    continue
                sha, subject, author, date = record.split("\x1f")
                commits.append({"sha": sha, "subject": subject, "author": author, "date": date})
            state["commits_ahead"] = commits
            state["commits_ahead_truncated"] = ahead > len(commits)
        state["base_in_head"], err = self.is_ancestor(base_ref, "HEAD")
        if err:
            state["errors"].append(f"base_in_head: {err}")

    # -- pull request ----------------------------------------------------

    def gh_api(self, path, paginate=False):
        env = self.runner.environment(self.git) if self.git else dict(os.environ, **SAFE_ENV)
        argv = [self.gh, "api"]
        if self.host and self.host != "github.com":
            argv += ["--hostname", self.host]
        if paginate:
            argv += ["--paginate", "--slurp"]
        argv.append(path)
        code, out, err = self.runner.run(argv, COMMAND_TIMEOUT, env)
        if code != 0:
            raise ApiError(err or f"gh api {path} exited {code}")
        try:
            return json.loads(out)
        except ValueError:
            raise ApiError(f"gh api {path} returned no JSON")

    def pr_state(self, git):
        pr = {"available": False, "reason": None}
        if not self.gh:
            pr["reason"] = "gh not found"
            self.failed = True
            return pr
        code, url, err = self.git_run("remote", "get-url", self.args.remote)
        slug = parse_remote(url.strip()) if code == 0 else None
        if not slug:
            pr["reason"] = f"remote {self.args.remote} is not a GitHub repository" if code == 0 else (err or "no remote")
            self.failed = True
            return pr
        self.host, owner, name = slug
        repo = f"repos/{owner}/{name}"
        try:
            if self.args.pr is not None:
                data = self.gh_api(f"{repo}/pulls/{self.args.pr}")
            else:
                if not git["branch"]:
                    pr.update(available=True, reason="detached HEAD: pass --pr", number=None)
                    return pr
                found = self.gh_api(f"{repo}/pulls?state=open&head={quote(owner + ':' + git['branch'], safe='')}")
                if not found:
                    pr.update(available=True, reason="no open pull request for this branch", number=None)
                    return pr
                data = found[0]
        except ApiError as exc:
            pr["reason"] = str(exc)
            self.failed = True
            return pr

        head = data.get("head", {}).get("sha")
        pr.update(available=True, number=data.get("number"), state=data.get("state"),
                  merged=bool(data.get("merged_at")), head_ref_oid=head,
                  head_matches_local=head is not None and head == git["head"],
                  head_matches_pushed=head is not None and head == git["pushed_head"])
        number = data.get("number")
        since = self.args.since
        pr["checks"] = self.collect(lambda: self.checks(repo, head))
        pr["reviews"] = self.collect(lambda: [
            dict(item_facts(r), state=r.get("state"), commit_id=r.get("commit_id"),
                 on_head=r.get("commit_id") == head, after_since=after(r.get("submitted_at"), since))
            for r in flatten(self.gh_api(f"{repo}/pulls/{number}/reviews?per_page=100", paginate=True))])
        pr["review_comments"] = self.collect(lambda: [
            dict(item_facts(c), path=c.get("path"), commit_id=c.get("commit_id"),
                 original_commit_id=c.get("original_commit_id"), in_reply_to_id=c.get("in_reply_to_id"),
                 on_head=c.get("original_commit_id") == head,
                 after_since=after(c.get("created_at") if c.get("in_reply_to_id") else c.get("updated_at"), since))
            for c in flatten(self.gh_api(f"{repo}/pulls/{number}/comments?per_page=100", paginate=True))])
        pr["comments"] = self.collect(lambda: [
            dict(item_facts(c), on_head=None, after_since=after(c.get("updated_at"), since))
            for c in flatten(self.gh_api(f"{repo}/issues/{number}/comments?per_page=100", paginate=True))])
        return pr

    def checks(self, repo, head):
        if not head:
            return []
        runs = []
        for page in self.gh_api(f"{repo}/commits/{head}/check-runs?per_page=100", paginate=True):
            runs += [{"name": r.get("name"), "status": r.get("status"), "conclusion": r.get("conclusion")}
                     for r in page.get("check_runs", [])]
        for page in self.gh_api(f"{repo}/commits/{head}/status?per_page=100", paginate=True):
            runs += [{"name": s.get("context"), "status": "completed" if s.get("state") != "pending" else "pending",
                      "conclusion": s.get("state")} for s in page.get("statuses", [])]
        return runs

    def collect(self, read):
        try:
            return read()
        except ApiError as exc:
            self.failed = True
            return {"error": str(exc)}


class ApiError(Exception):
    pass


REMOTE = re.compile(r"^(?:[a-z+]+://)?(?:[^@/]+@)?([^:/]+)(?::\d+)?[:/]([^/]+)/([^/]+?)(?:\.git)?/?$")


def parse_remote(url):
    m = REMOTE.match(url)
    if not m:
        return None
    host = m.group(1)
    if "github" not in host:
        return None
    return host, m.group(2), m.group(3)


def flatten(pages):
    items = []
    for page in pages:
        items += page if isinstance(page, list) else [page]
    return items


def item_facts(item):
    user = item.get("user") or {}
    return {"id": item.get("id"), "user": user.get("login"), "user_type": user.get("type"),
            "created_at": item.get("created_at") or item.get("submitted_at"),
            "updated_at": item.get("updated_at"), "url": item.get("html_url"), "body": item.get("body")}


def parse_time(text):
    if text is None:
        return None
    value = datetime.fromisoformat(text.replace("Z", "+00:00"))
    return value if value.tzinfo else value.replace(tzinfo=timezone.utc)


def after(stamp, since):
    if since is None or stamp is None:
        return None
    return parse_time(stamp) > since


def valid_name(text):
    return bool(text) and not text.startswith("-") and "\0" not in text and ".." not in text


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--base", required=True, help="base branch name, such as development")
    parser.add_argument("--repo", default=".", help="repository directory (default: current)")
    parser.add_argument("--remote", default="origin")
    parser.add_argument("--fetch", action="store_true", help="fetch the remote before comparing")
    group = parser.add_mutually_exclusive_group()
    group.add_argument("--pr", type=int, help="pull request number (default: the branch's open one)")
    group.add_argument("--no-pr", action="store_true", help="skip the pull-request read")
    parser.add_argument("--since", help="push time, ISO-8601 with a time zone, for after_since")
    parser.add_argument("--timeout", type=float, help="cap every command at this many seconds")
    try:
        args = parser.parse_args(argv)
    except SystemExit as exc:
        return 2 if exc.code else 0
    if not valid_name(args.base) or not valid_name(args.remote):
        print("git_state.py: --base and --remote must be plain names", file=sys.stderr)
        return 2
    if args.pr is not None and args.pr <= 0:
        print("git_state.py: --pr must be a positive number", file=sys.stderr)
        return 2
    if args.timeout is not None and args.timeout <= 0:
        print("git_state.py: --timeout must be positive", file=sys.stderr)
        return 2
    try:
        since = datetime.fromisoformat(args.since.replace("Z", "+00:00")) if args.since else None
    except ValueError:
        print(f"git_state.py: --since is not ISO-8601: {args.since}", file=sys.stderr)
        return 2
    if since is not None and since.tzinfo is None:
        # A local push time read as UTC would shift after_since by the offset.
        print("git_state.py: --since needs a time zone, such as Z or -06:00", file=sys.stderr)
        return 2
    args.since = since

    state = State(args)
    state.host = None
    if not state.git:
        print("git_state.py: git not found", file=sys.stderr)
        return 2
    git, err = state.git_state()
    if git is None:
        print(f"git_state.py: {args.repo} is not a Git repository: {err}", file=sys.stderr)
        return 2
    result = dict(git)
    if not args.no_pr:
        result["pr"] = state.pr_state(git)
    print(json.dumps(result, separators=(",", ":"), ensure_ascii=False))
    return 1 if state.failed else 0


if __name__ == "__main__":
    sys.exit(main())
