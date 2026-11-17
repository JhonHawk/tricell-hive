#!/usr/bin/env python3
"""Plan and optionally execute branch and worktree cleanup for flow-close.

Usage:
  python3 cleanup_branches.py --repo DIR --base BRANCH [--env-branch BRANCH]...
                              [--active-branch BRANCH]... [--open-pr-base BRANCH]...
                              [--merged-pr BRANCH=SHA]... [--apply]

By default runs in dry-run mode and prints a JSON plan of `worktrees` and
`branches` with their action (`remove_worktree`, `delete_branch`, or `keep`)
and reason. With `--apply`, executes only the safe removals and deletions.
"""

import argparse
import filecmp
import fnmatch
import json
import os
import pathlib
import shutil
import subprocess
import sys

DISPOSABLE_BUILD_DIRS = {"node_modules", ".venv", "venv", "dist", "build", ".next", "target", "out", "coverage"}
DISPOSABLE_CACHE_DIRS = {".turbo", "__pycache__", ".pytest_cache", ".mypy_cache", ".ruff_cache"}
DISPOSABLE_FILE_GLOBS = ("*.tsbuildinfo", "next-env.d.ts", "*.pyc", ".DS_Store")


def run_git(repo, *args):
    git = shutil.which("git") or "git"
    proc = subprocess.run(
        [git, "-C", str(repo), *args],
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
    )
    return proc.returncode, proc.stdout, proc.stderr.strip()


def default_lsof_runner():
    lsof = shutil.which("lsof")
    if not lsof:
        return 1, "lsof not found"
    proc = subprocess.run(
        [lsof, "-d", "cwd", "-Fn"],
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0 and not proc.stdout:
        return 1, proc.stderr.strip() or "lsof failed"
    return 0, proc.stdout


def parse_worktrees(repo):
    code, out, err = run_git(repo, "worktree", "list", "--porcelain")
    if code != 0:
        raise RuntimeError(err or "git worktree list failed")
    items = []
    current = {}
    for line in out.splitlines():
        if not line.strip():
            if current:
                items.append(current)
                current = {}
            continue
        if line.startswith("worktree "):
            current["path"] = line[len("worktree ") :].strip()
        elif line.startswith("HEAD "):
            current["head"] = line[len("HEAD ") :].strip()
        elif line.startswith("branch "):
            ref = line[len("branch ") :].strip()
            if ref.startswith("refs/heads/"):
                current["branch"] = ref[len("refs/heads/") :]
            else:
                current["branch"] = ref
        elif line == "bare":
            current["bare"] = True
    if current:
        items.append(current)
    return items


def is_disposable_ignored_path(rel_path):
    parts = pathlib.PurePosixPath(rel_path.strip("/")).parts
    if not parts:
        return False
    if any(p in DISPOSABLE_BUILD_DIRS or p in DISPOSABLE_CACHE_DIRS for p in parts):
        return True
    name = parts[-1]
    return any(fnmatch.fnmatch(name, pat) for pat in DISPOSABLE_FILE_GLOBS)


def dir_matches_main(main_dir, wt_dir):
    if not main_dir.is_dir() or not wt_dir.is_dir():
        return False
    for root, dirs, files in os.walk(wt_dir):
        rel = pathlib.Path(root).relative_to(wt_dir)
        target_root = main_dir / rel
        if not target_root.is_dir():
            return False
        for fname in files:
            wfile = pathlib.Path(root) / fname
            mfile = target_root / fname
            if not mfile.is_file() or not filecmp.cmp(str(mfile), str(wfile), shallow=False):
                return False
    return True


def matches_main_checkout(main_path, wt_path, rel_path):
    if not main_path:
        return False
    m_target = pathlib.Path(main_path) / rel_path.rstrip("/")
    w_target = pathlib.Path(wt_path) / rel_path.rstrip("/")
    if w_target.is_file():
        return m_target.is_file() and filecmp.cmp(str(m_target), str(w_target), shallow=False)
    if w_target.is_dir():
        return dir_matches_main(m_target, w_target)
    return False


def inspect_worktree_cleanliness(main_path, wt_path):
    code, out, err = run_git(wt_path, "status", "--ignored", "--porcelain=v1")
    if code != 0:
        return False, f"git status failed: {err}"
    for line in out.splitlines():
        if len(line) < 4:
            continue
        status = line[:2]
        rel = line[3:].strip()
        if status == "??":
            return False, f"untracked entry: {rel}"
        if status != "!!":
            return False, f"uncommitted change ({status.strip()}): {rel}"
        if is_disposable_ignored_path(rel):
            continue
        if matches_main_checkout(main_path, wt_path, rel):
            continue
        return False, f"non-disposable ignored entry differs from main checkout: {rel}"
    return True, None


def worktree_has_active_cwd(wt_path, lsof_code, lsof_out):
    if lsof_code != 0:
        return True, "process check could not run (lsof unavailable or failed)"
    real_wt = os.path.realpath(wt_path)
    prefix = real_wt.rstrip("/") + "/"
    for line in lsof_out.splitlines():
        if not line.startswith("n"):
            continue
        cwd = os.path.realpath(line[1:].strip())
        if cwd == real_wt or cwd.startswith(prefix):
            return True, f"running process has working directory in worktree ({cwd})"
    return False, None


def is_branch_merged(repo, branch, base, merged_prs):
    code, local_sha, _ = run_git(repo, "rev-parse", "--verify", "--quiet", f"refs/heads/{branch}")
    local_sha = local_sha.strip() if code == 0 else None
    code_r, remote_sha, _ = run_git(repo, "rev-parse", "--verify", "--quiet", f"refs/remotes/origin/{branch}")
    remote_sha = remote_sha.strip() if code_r == 0 else None

    code_anc, _, _ = run_git(repo, "merge-base", "--is-ancestor", f"refs/heads/{branch}", base) if local_sha else (1, "", "")
    git_merged = code_anc == 0

    pr_head = merged_prs.get(branch)
    if pr_head:
        if local_sha and local_sha != pr_head:
            return False, False, f"local branch head ({local_sha[:8]}) moved past merged PR headRefOid ({pr_head[:8]})"
        if remote_sha and remote_sha != pr_head:
            return False, False, f"remote branch head ({remote_sha[:8]}) moved past merged PR headRefOid ({pr_head[:8]})"
        return True, not git_merged, None

    if git_merged:
        if remote_sha and local_sha and remote_sha != local_sha:
            code_r_anc, _, _ = run_git(repo, "merge-base", "--is-ancestor", f"refs/remotes/origin/{branch}", base)
            if code_r_anc != 0:
                return False, False, f"remote branch origin/{branch} has unmerged commits"
        return True, False, None

    return False, False, f"branch {branch} has commits not merged into {base}"


def plan_cleanup(
    repo,
    base="development",
    env_branches=None,
    active_branches=None,
    open_pr_bases=None,
    merged_prs=None,
    lsof_runner=None,
):
    env_branches = set(env_branches or [])
    active_branches = set(active_branches or [])
    open_pr_bases = set(open_pr_bases or [])
    merged_prs = dict(merged_prs or {})
    lsof_runner = lsof_runner or default_lsof_runner

    worktrees = parse_worktrees(repo)
    main_checkout = worktrees[0]["path"] if worktrees and not worktrees[0].get("bare") else None
    lsof_code, lsof_out = lsof_runner()

    protected = {base, "master", "main"} | env_branches
    wt_plans = []
    kept_branches = set()
    removable_wt_branches = set()
    checked_out_branches = {}

    for idx, wt in enumerate(worktrees):
        branch = wt.get("branch")
        path = wt.get("path")
        if branch:
            checked_out_branches[branch] = path
        if idx == 0:
            continue
        if not branch:
            wt_plans.append({"path": path, "branch": None, "action": "keep", "reason": "detached HEAD worktree"})
            continue
        if branch in protected:
            reason = f"protected branch ({branch})"
            wt_plans.append({"path": path, "branch": branch, "action": "keep", "reason": reason})
            kept_branches.add((branch, reason))
            continue
        if branch in open_pr_bases:
            reason = f"used as base of an open pull request ({branch})"
            wt_plans.append({"path": path, "branch": branch, "action": "keep", "reason": reason})
            kept_branches.add((branch, reason))
            continue
        if branch in active_branches:
            reason = f"named in active work/checkpoint ({branch})"
            wt_plans.append({"path": path, "branch": branch, "action": "keep", "reason": reason})
            kept_branches.add((branch, reason))
            continue

        busy, busy_reason = worktree_has_active_cwd(path, lsof_code, lsof_out)
        if busy:
            wt_plans.append({"path": path, "branch": branch, "action": "keep", "reason": busy_reason})
            kept_branches.add((branch, busy_reason))
            continue

        clean, clean_reason = inspect_worktree_cleanliness(main_checkout, path)
        if not clean:
            wt_plans.append({"path": path, "branch": branch, "action": "keep", "reason": clean_reason})
            kept_branches.add((branch, clean_reason))
            continue

        merged, _, merge_reason = is_branch_merged(repo, branch, base, merged_prs)
        if not merged:
            wt_plans.append({"path": path, "branch": branch, "action": "keep", "reason": merge_reason})
            kept_branches.add((branch, merge_reason))
            continue

        wt_plans.append({"path": path, "branch": branch, "action": "remove_worktree", "reason": "clean merged worktree"})
        removable_wt_branches.add(branch)

    code, out, _ = run_git(repo, "for-each-ref", "--format=%(refname:short)", "refs/heads")
    local_branches = [b.strip() for b in out.splitlines() if b.strip()] if code == 0 else []
    candidates = sorted(set(local_branches) | set(merged_prs.keys()))

    kept_map = dict(kept_branches)
    branch_plans = []
    for branch in candidates:
        if branch in protected:
            continue
        if branch in kept_map:
            branch_plans.append({"branch": branch, "action": "keep", "reason": kept_map[branch]})
            continue
        if branch in open_pr_bases:
            branch_plans.append({"branch": branch, "action": "keep", "reason": "base of an open pull request"})
            continue
        if branch in active_branches:
            branch_plans.append({"branch": branch, "action": "keep", "reason": "named in active work/checkpoint"})
            continue
        if branch in checked_out_branches and branch not in removable_wt_branches:
            branch_plans.append(
                {"branch": branch, "action": "keep", "reason": f"checked out in {checked_out_branches[branch]}"}
            )
            continue

        merged, squash_or_rebase, merge_reason = is_branch_merged(repo, branch, base, merged_prs)
        if not merged:
            branch_plans.append({"branch": branch, "action": "keep", "reason": merge_reason})
            continue

        code_r, _, _ = run_git(repo, "rev-parse", "--verify", "--quiet", f"refs/remotes/origin/{branch}")
        has_remote = code_r == 0
        branch_plans.append(
            {
                "branch": branch,
                "action": "delete_branch",
                "delete_flag": "-D" if squash_or_rebase else "-d",
                "delete_remote": has_remote,
                "reason": "merged into base",
            }
        )

    return {
        "repo": str(repo),
        "base": base,
        "main_checkout": main_checkout,
        "worktrees": wt_plans,
        "branches": branch_plans,
    }


def remove_disposable_ignored(wt_path):
    code, out, _ = run_git(wt_path, "status", "--ignored", "--porcelain=v1")
    if code != 0:
        return
    for line in out.splitlines():
        if not line.startswith("!! "):
            continue
        rel = line[3:].strip().rstrip("/")
        target = pathlib.Path(wt_path) / rel
        if target.is_symlink() or target.is_file():
            target.unlink(missing_ok=True)
        elif target.is_dir():
            shutil.rmtree(target, ignore_errors=True)


def apply_cleanup(repo, plan):
    errors = []
    removed_worktrees = []
    deleted_branches = []

    for wt in plan.get("worktrees", []):
        if wt.get("action") != "remove_worktree":
            continue
        remove_disposable_ignored(wt["path"])
        code, _, err = run_git(repo, "worktree", "remove", wt["path"])
        if code != 0:
            errors.append(f"worktree remove {wt['path']}: {err}")
        else:
            removed_worktrees.append(wt["path"])

    for br in plan.get("branches", []):
        if br.get("action") != "delete_branch":
            continue
        branch = br["branch"]
        if br.get("delete_remote"):
            code_r, _, err_r = run_git(repo, "push", "origin", "--delete", branch)
            if code_r != 0:
                errors.append(f"remote delete {branch}: {err_r}")
                continue
        flag = br.get("delete_flag", "-d")
        code_l, _, err_l = run_git(repo, "branch", flag, branch)
        if code_l != 0:
            errors.append(f"branch {flag} {branch}: {err_l}")
        else:
            deleted_branches.append(branch)

    return {
        "ok": len(errors) == 0,
        "removed_worktrees": removed_worktrees,
        "deleted_branches": deleted_branches,
        "errors": errors,
    }


def parse_merged_pr_arg(items):
    result = {}
    for item in items or []:
        if "=" not in item:
            raise ValueError(f"invalid --merged-pr {item!r}, expected BRANCH=HEAD_SHA")
        branch, sha = item.split("=", 1)
        result[branch.strip()] = sha.strip()
    return result


def main(argv=None):
    parser = argparse.ArgumentParser(description="Plan and execute branch and worktree cleanup.")
    parser.add_argument("--repo", default=".", help="Repository path")
    parser.add_argument("--base", required=True, help="Base branch name")
    parser.add_argument("--env-branch", action="append", default=[], help="Environment branch to exclude")
    parser.add_argument("--active-branch", action="append", default=[], help="Branch named in active plan/checkpoint")
    parser.add_argument("--open-pr-base", action="append", default=[], help="Branch used as base of an open PR")
    parser.add_argument("--merged-pr", action="append", default=[], help="Merged PR mapping BRANCH=HEAD_SHA")
    parser.add_argument("--apply", action="store_true", help="Execute removals and branch deletions")
    args = parser.parse_args(argv)

    try:
        merged_prs = parse_merged_pr_arg(args.merged_pr)
        plan = plan_cleanup(
            args.repo,
            base=args.base,
            env_branches=args.env_branch,
            active_branches=args.active_branch,
            open_pr_bases=args.open_pr_base,
            merged_prs=merged_prs,
        )
    except (ValueError, RuntimeError) as exc:
        print(json.dumps({"ok": False, "error": str(exc)}), file=sys.stderr)
        return 2

    if args.apply:
        applied = apply_cleanup(args.repo, plan)
        plan["applied"] = applied
        print(json.dumps(plan, indent=2))
        return 0 if applied["ok"] else 1

    print(json.dumps(plan, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
