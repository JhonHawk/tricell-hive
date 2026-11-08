from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
import json
import os
import shutil
import subprocess
import sys
import tempfile
import textwrap
import threading
import time
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/git-workflow/scripts/git_state.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

GIT = shutil.which("git")
OLD = "1" * 40

FAKE_GH = textwrap.dedent("""\
    #!/usr/bin/env python3
    import json, os, sys, time
    if os.environ.get("FAKE_GH_SLEEP"):
        time.sleep(float(os.environ["FAKE_GH_SLEEP"]))
    with open(os.environ["FAKE_GH_FIXTURE"]) as f:
        fixture = json.load(f)
    path = sys.argv[-1]
    if path not in fixture:
        sys.stderr.write("HTTP 404: Not Found (" + path + ")\\n")
        sys.exit(1)
    reply = fixture[path]
    if isinstance(reply, dict) and "__fail__" in reply:
        sys.stderr.write(reply["__fail__"] + "\\n")
        sys.exit(1)
    if "--slurp" in sys.argv:
        reply = [reply]
    print(json.dumps(reply))
""")


class Repos:
    """A bare origin, a work clone on a feature branch, and a second clone
    that moves the base, all isolated from the user's Git configuration."""

    def __init__(self, tmp):
        self.tmp = Path(tmp)
        self.env = dict(os.environ, GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_NOSYSTEM="1",
                        GIT_AUTHOR_NAME="Ana", GIT_AUTHOR_EMAIL="ana@example.com",
                        GIT_COMMITTER_NAME="Ana", GIT_COMMITTER_EMAIL="ana@example.com")
        self.origin = self.tmp / "origin.git"
        self.work = self.tmp / "work"
        self.other = self.tmp / "other"
        self.git(self.tmp, "init", "--bare", "-b", "main", str(self.origin))
        self.git(self.tmp, "clone", "-q", str(self.origin), str(self.work))
        self.commit(self.work, "base")
        self.git(self.work, "push", "-q", "origin", "main")
        self.git(self.tmp, "clone", "-q", str(self.origin), str(self.other))

    def git(self, cwd, *args):
        return subprocess.run([GIT, *args], cwd=cwd, env=self.env, check=True,
                              capture_output=True, text=True).stdout.strip()

    def commit(self, repo, name):
        (repo / f"{name}.txt").write_text(name)
        self.git(repo, "add", f"{name}.txt")
        self.git(repo, "commit", "-q", "-m", name)
        return self.git(repo, "rev-parse", "HEAD")

    def snapshot(self, repo):
        return (self.git(repo, "status", "--porcelain=v1", "--untracked-files=all"),
                self.git(repo, "for-each-ref", "refs/heads", "refs/tags"),
                self.git(repo, "config", "--local", "--list"),
                (repo / ".git/index").read_bytes())


def run(repo, *args, env=None, path=None):
    env = dict(env or os.environ)
    if path is not None:
        env["PATH"] = path
    started = time.monotonic()
    proc = subprocess.run([sys.executable, "-I", str(SCRIPT), "--repo", str(repo), *args],
                          env=env, capture_output=True, text=True, timeout=60)
    elapsed = time.monotonic() - started
    data = json.loads(proc.stdout) if proc.stdout.strip() else None
    return proc.returncode, data, elapsed, proc.stderr


class GitStateTest(unittest.TestCase):
    def setUp(self):
        self.tmpdir = tempfile.TemporaryDirectory()
        self.r = Repos(self.tmpdir.name)
        self.bin = Path(self.tmpdir.name) / "bin"
        self.bin.mkdir()
        os.symlink(GIT, self.bin / "git")

    def tearDown(self):
        self.tmpdir.cleanup()

    def fake_gh(self, fixture):
        gh = self.bin / "gh"
        gh.write_text(FAKE_GH.replace("/usr/bin/env python3", sys.executable))
        gh.chmod(0o755)
        path = Path(self.tmpdir.name) / "fixture.json"
        path.write_text(json.dumps(fixture))
        return dict(self.r.env, FAKE_GH_FIXTURE=str(path), PATH=str(self.bin))

    def feature_two_ahead_one_behind(self):
        r = self.r
        r.git(r.work, "switch", "-q", "-c", "feature")
        r.commit(r.work, "f1")
        r.commit(r.work, "f2")
        r.git(r.work, "push", "-q", "-u", "origin", "feature")
        r.commit(r.other, "moved")
        r.git(r.other, "push", "-q", "origin", "main")

    def test_ahead_behind_and_base_containment(self):
        r = self.r
        self.feature_two_ahead_one_behind()

        code, data, _, _ = run(r.work, "--base", "main", "--fetch", "--no-pr", env=r.env)
        self.assertEqual(code, 0)
        self.assertEqual((data["ahead"], data["behind"]), (2, 1))
        self.assertEqual([c["subject"] for c in data["commits_ahead"]], ["f2", "f1"])
        self.assertEqual(data["commits_ahead"][0]["author"], "Ana <ana@example.com>")
        self.assertIs(data["base_in_head"], False)
        self.assertIs(data["base_in_pushed_head"], False)

        r.git(r.work, "merge", "-q", "--no-edit", "origin/main")
        code, data, _, _ = run(r.work, "--base", "main", "--no-pr", env=r.env)
        self.assertIs(data["base_in_head"], True)
        self.assertIs(data["base_in_pushed_head"], False, "integrated locally but not pushed")

        r.git(r.work, "push", "-q", "origin", "feature")
        code, data, _, _ = run(r.work, "--base", "main", "--no-pr", env=r.env)
        self.assertIs(data["base_in_pushed_head"], True)

    def test_failed_fetch_reports_no_comparison(self):
        r = self.r
        self.feature_two_ahead_one_behind()
        r.git(r.work, "remote", "set-url", "origin", str(Path(self.tmpdir.name) / "gone.git"))

        code, data, _, _ = run(r.work, "--base", "main", "--fetch", "--no-pr", env=r.env)
        self.assertEqual(code, 1)
        self.assertIs(data["fetch"]["ok"], False)
        self.assertTrue(data["fetch"]["error"])
        for field in ("ahead", "behind", "commits_ahead", "base_in_head", "base_in_pushed_head"):
            self.assertIsNone(data[field], field)
        self.assertEqual(data["branch"], "feature", "the other Git fields are kept")

    def test_http_remote_asking_for_credentials_does_not_wait(self):
        class Deny(BaseHTTPRequestHandler):
            def do_GET(self):
                self.send_response(401)
                self.send_header("WWW-Authenticate", 'Basic realm="x"')
                self.end_headers()

            def log_message(self, *args):
                pass

        server = HTTPServer(("127.0.0.1", 0), Deny)
        threading.Thread(target=server.serve_forever, daemon=True).start()
        self.addCleanup(server.server_close)
        self.addCleanup(server.shutdown)
        r = self.r
        r.git(r.work, "remote", "set-url", "origin", f"http://127.0.0.1:{server.server_port}/repo.git")

        code, data, elapsed, _ = run(r.work, "--base", "main", "--fetch", "--no-pr", env=r.env)
        self.assertEqual(code, 1)
        self.assertIs(data["fetch"]["ok"], False)
        self.assertLess(elapsed, 30, "a credential prompt would block until the fetch timeout")

    def test_reviews_and_comments_carry_facts_not_verdicts(self):
        r = self.r
        self.feature_two_ahead_one_behind()
        r.git(r.work, "remote", "set-url", "origin", "https://github.com/acme/widget.git")
        head = r.git(r.work, "rev-parse", "HEAD")
        repo = "repos/acme/widget"
        bot = {"login": "review-bot[bot]", "type": "Bot"}
        person = {"login": "maria", "type": "User"}
        env = self.fake_gh({
            f"{repo}/pulls/7": {"number": 7, "state": "open", "merged_at": None, "head": {"sha": head}},
            f"{repo}/commits/{head}/check-runs?per_page=100": {"check_runs": [
                {"name": "unit", "status": "completed", "conclusion": "success"}]},
            f"{repo}/commits/{head}/status?per_page=100": {"statuses": [{"context": "legacy", "state": "pending"}]},
            f"{repo}/pulls/7/reviews?per_page=100": [
                {"id": 1, "user": bot, "commit_id": OLD, "state": "COMMENTED", "submitted_at": "2026-11-08T08:00:00Z"},
                {"id": 2, "user": bot, "commit_id": head, "state": "COMMENTED", "submitted_at": "2026-11-08T10:00:00Z"}],
            f"{repo}/pulls/7/comments?per_page=100": [
                {"id": 10, "user": person, "original_commit_id": OLD, "commit_id": head,
                 "created_at": "2026-11-08T07:00:00Z", "updated_at": "2026-11-08T07:00:00Z", "body": "rename this"},
                {"id": 11, "user": bot, "original_commit_id": OLD, "in_reply_to_id": 10,
                 "created_at": "2026-11-08T10:00:00Z", "updated_at": "2026-11-08T10:00:00Z"},
                {"id": 12, "user": bot, "original_commit_id": OLD, "in_reply_to_id": 10,
                 "created_at": "2026-11-08T08:00:00Z", "updated_at": "2026-11-08T11:00:00Z"}],
            f"{repo}/issues/7/comments?per_page=100": [
                {"id": 20, "user": bot, "created_at": "2026-11-08T07:00:00Z", "updated_at": "2026-11-08T07:30:00Z"},
                {"id": 21, "user": bot, "created_at": "2026-11-08T07:00:00Z", "updated_at": "2026-11-08T10:30:00Z"}],
        })

        code, data, _, stderr = run(r.work, "--base", "main", "--pr", "7", "--since", "2026-11-08T09:00:00Z", env=env)
        self.assertEqual(code, 0, stderr)
        pr = data["pr"]
        self.assertEqual(pr["head_ref_oid"], head)
        self.assertTrue(pr["head_matches_local"] and pr["head_matches_pushed"])
        self.assertEqual({c["name"]: c["conclusion"] for c in pr["checks"]}, {"unit": "success", "legacy": "pending"})
        self.assertEqual([(x["on_head"], x["user_type"]) for x in pr["reviews"]], [(False, "Bot"), (True, "Bot")])

        comments = {c["id"]: c for c in pr["review_comments"]}
        human = comments[10]
        self.assertEqual((human["user_type"], human["on_head"]), ("User", False))
        self.assertNotIn("current", human, "the script must not rule out a human thread")
        self.assertIs(comments[11]["after_since"], True, "a reply compares its creation time")
        self.assertIs(comments[12]["after_since"], False, "an edit does not make an old reply new")
        self.assertEqual([c["after_since"] for c in pr["comments"]], [False, True])

        for same_instant in ("2026-11-08t09:00:00z", "2026-11-08T10:00:00+0100"):
            code, other, _, stderr = run(r.work, "--base", "main", "--pr", "7", "--since", same_instant, env=env)
            self.assertEqual(code, 0, stderr)
            self.assertEqual(other["pr"], pr, same_instant)

    def test_without_since_after_since_is_undecided(self):
        r = self.r
        r.git(r.work, "remote", "set-url", "origin", "git@github.com:acme/widget.git")
        head = r.git(r.work, "rev-parse", "HEAD")
        repo = "repos/acme/widget"
        env = self.fake_gh({
            f"{repo}/pulls?state=open&head=acme%3Amain": [{"number": 3, "state": "open", "head": {"sha": head}}],
            f"{repo}/commits/{head}/check-runs?per_page=100": {"check_runs": []},
            f"{repo}/commits/{head}/status?per_page=100": {"statuses": []},
            f"{repo}/pulls/3/reviews?per_page=100": [],
            f"{repo}/pulls/3/comments?per_page=100": [],
            f"{repo}/issues/3/comments?per_page=100": [{"id": 1, "user": {"login": "x", "type": "User"},
                                                       "updated_at": "2026-11-08T07:00:00Z"}],
        })
        code, data, _, stderr = run(r.work, "--base", "main", env=env)
        self.assertEqual(code, 0, stderr)
        self.assertEqual(data["pr"]["number"], 3, "found by the branch's open pull request")
        self.assertIsNone(data["pr"]["comments"][0]["after_since"])

    def test_pull_request_source_failures_keep_git_fields(self):
        r = self.r
        r.git(r.work, "remote", "set-url", "origin", "https://github.com/acme/widget.git")

        code, data, _, _ = run(r.work, "--base", "main", "--pr", "7", path=str(self.bin), env=r.env)
        self.assertEqual(code, 1)
        self.assertEqual(data["pr"], {"available": False, "reason": "gh not found"})
        self.assertEqual(data["branch"], "main")

        env = self.fake_gh({"repos/acme/widget/pulls/7": {"__fail__": "HTTP 502: Bad Gateway"}})
        code, data, _, _ = run(r.work, "--base", "main", "--pr", "7", env=env)
        self.assertEqual(code, 1)
        self.assertIs(data["pr"]["available"], False)
        self.assertIn("502", data["pr"]["reason"])
        self.assertIs(data["base_in_head"], True)

    def test_a_slow_gh_is_cut_at_its_time_limit(self):
        r = self.r
        r.git(r.work, "remote", "set-url", "origin", "https://github.com/acme/widget.git")
        env = dict(self.fake_gh({}), FAKE_GH_SLEEP="30")

        code, data, elapsed, _ = run(r.work, "--base", "main", "--pr", "7", "--timeout", "1", env=env)
        self.assertEqual(code, 1)
        self.assertIn("timed out", data["pr"]["reason"])
        self.assertLess(elapsed, 10)

    def test_reading_changes_only_remote_tracking_refs(self):
        r = self.r
        self.feature_two_ahead_one_behind()
        (r.work / "f1.txt").write_text("edited")
        (r.work / "new.txt").write_text("untracked")
        r.git(r.work, "add", "new.txt")
        before = r.snapshot(r.work)

        code, data, _, _ = run(r.work, "--base", "main", "--fetch", "--no-pr", env=r.env)
        self.assertEqual(code, 0)
        self.assertEqual(data["dirty"], {"staged": 1, "unstaged": 1, "untracked": 0})
        self.assertEqual(r.snapshot(r.work), before)
        self.assertFalse((r.work / ".git/FETCH_HEAD").exists())

    def test_invalid_usage_and_non_repository(self):
        code, _, _, _ = run(self.r.work, "--base", "--upload-pack=x", env=self.r.env)
        self.assertEqual(code, 2)
        code, _, _, _ = run(self.r.work, "--base", "main", "--since", "yesterday", env=self.r.env)
        self.assertEqual(code, 2)
        code, _, _, _ = run(self.r.work, "--base", "main", "--since", "", env=self.r.env)
        self.assertEqual(code, 2, "an empty push time must not pass as no push time")
        code, _, _, _ = run(self.r.work, "--base", "main", "--since", "2026-11-08T10:00:00", "--no-pr", env=self.r.env)
        self.assertEqual(code, 0, "a push time without a zone is read as local time")
        code, _, _, _ = run(self.r.work, "--base", "main", "--pr", "0", env=self.r.env)
        self.assertEqual(code, 2)
        code, _, _, _ = run(Path(self.tmpdir.name) / "bin", "--base", "main", "--no-pr", env=self.r.env)
        self.assertEqual(code, 2)


if __name__ == "__main__":
    unittest.main()
