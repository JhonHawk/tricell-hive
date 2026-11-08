from pathlib import Path
import json
import os
import subprocess
import sys
import tempfile
import time
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/local-instructions/scripts/extract_candidates.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

SECRETS = ["hunter2pass", "s3cr3tUserPw", "abcTOKENxyz123456789", "pa55word!", "S3cretPw", "dbpw77",
           "qtok99", "731942", "redispw", "vercelpw", "wJalrXUtnFEMI", "84736251", "98765432", "55512345",
           "zq/x7k", "x7kq", "p@ss", "secretvalue123"]


def claude_session(cwd, session, commands):
    """One Claude Code transcript: each command is (text, is_error, result_text[, background])."""
    lines = []
    for i, (command, is_error, result, *rest) in enumerate(commands):
        tool_id = f"toolu_{session}_{i}"
        tool_input = {"command": command, "description": "run"}
        if rest and rest[0]:
            tool_input["run_in_background"] = True
        lines.append({"type": "assistant", "cwd": cwd, "sessionId": session, "timestamp": "2026-11-05T10:00:00Z",
                      "message": {"content": [{"type": "tool_use", "id": tool_id, "name": "Bash",
                                               "input": tool_input}]}})
        lines.append({"type": "user", "cwd": cwd, "sessionId": session, "timestamp": "2026-11-05T10:00:01Z",
                      "message": {"content": [{"type": "tool_result", "tool_use_id": tool_id, "is_error": is_error,
                                               "content": result}]}})
    return "\n".join(json.dumps(line) for line in lines) + "\n"


def codex_session(cwd, commands, source="cli"):
    """One Codex rollout: each command is (argv, exit_code, output)."""
    lines = [{"timestamp": "2026-11-06T09:00:00Z", "type": "session_meta",
              "payload": {"id": "cx1", "cwd": cwd, "source": source}}]
    for i, (argv, code, output) in enumerate(commands):
        lines.append({"timestamp": "2026-11-06T09:00:05Z", "type": "event_msg",
                      "payload": {"type": "item_completed", "item": {"type": "CommandExecution", "id": f"c{i}",
                                  "command": argv, "cwd": cwd, "exit_code": code, "aggregated_output": output}}})
    return "\n".join(json.dumps(line) for line in lines) + "\n"


class ExtractCandidatesTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        base = Path(self.tmp.name)
        self.root = "/work/shop"
        self.claude = base / "claude"
        self.codex = base / "codex"
        projects = self.claude / "projects"

        main = projects / "-work-shop"
        main.mkdir(parents=True)
        (main / "s1.jsonl").write_text(claude_session(self.root, "s1", [
            ("PORT=3091 DATABASE_URL=postgres://app:hunter2pass@localhost:5432/shop pnpm start", False, "listening"),
            ("cd apps/api && pnpm run dev --port 3001", False, "ready"),
            ("curl -H 'Authorization: Bearer abcTOKENxyz123456789' http://localhost:3001/health", False, "ok"),
            ("timeout 5 pnpm test", True, "bash: line 1: timeout: command not found"),
            ("aws logs tail /ecs/api --profile shop-qa --region mx-central-1", False, "log"),
            ("pnpm run broken", True, "exit 1"),
            ("MONGOMS_SYSTEM_BINARY=/home/u/.cache/mongod MONGOMS_RUNTIME_DOWNLOAD=false pnpm test:int", False, "ok"),
            ("API_SECRET=\"hunter2pass and more\" NEXT_PUBLIC_APP_ENV=development pnpm dev", False, "ok"),
            ("python3 - <<'EOF'\nprint('inline')\nEOF", False, "inline"),
            ("node -e 'console.log(1)'", False, "1"),
            ("mysql -uroot -pS3cretPw shop", False, "ok"),
            ("docker run -e DB_PASSWORD=dbpw77 -p 5432:5432 postgres", False, "ok"),
            ("curl http://localhost:3001/api?token=qtok99", False, "ok"),
            ("PGPASSWORD=731942 psql -h localhost -U app shop", False, "ok"),
            ("redis-cli -a redispw ping", False, "PONG"),
            ("vercel env add API_KEY production <<< vercelpw", False, "ok"),
            ("pnpm dev:watch", False, "started", True),
            ("rg 'command not found' docs || true", False, "docs/x.md: bash: foo: command not found"),
            ("aws configure set aws_secret_access_key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", False, ""),
            ("mysql -u root -p 84736251 shop", False, "ok"),
            ("redis-cli -a 98765432 ping", False, "PONG"),
            ("docker login --password 55512345 registry.local", False, "ok"),
            ("psql postgres://app:zq/x7kq@localhost:5432/db", False, "ok"),
            ("mongosh mongodb://app:p@ss@db.local/x", False, "ok"),
            ("curl -d secretvalue123 http://localhost:3000/x", False, "ok"),
            ("ssh -p 2222 deploy@qa.local", False, "ok"),
        ]))
        # A second session repeats the start command, so it counts twice.
        (main / "s2.jsonl").write_text(claude_session(self.root, "s2", [
            ("pnpm start", False, "listening"),
            ("mongosh mongodb://admin:s3cr3tUserPw@127.0.0.1:27017/shop --eval 'db.stats()'", False, "{}"),
        ]))
        # Subagent transcripts are excluded.
        sub = main / "s2" / "subagents"
        sub.mkdir(parents=True)
        (sub / "agent-a.jsonl").write_text(claude_session(self.root, "a", [("make subagent-only", False, "")]))
        # A worktree of the project has its own folder and is included.
        wt = projects / "-work-shop--claude-worktrees-w1"
        wt.mkdir()
        (wt / "s3.jsonl").write_text(claude_session(self.root + "/.claude/worktrees/w1", "s3", [
            ("docker compose up -d db", False, "started"),
        ]))
        # A neighbouring project that shares the path prefix is excluded.
        other = projects / "-work-shop-private"
        other.mkdir()
        (other / "s4.jsonl").write_text(claude_session("/work/shop-private", "s4", [("make neighbour-only", False, "")]))
        # A session outside the window is excluded.
        old = main / "s5.jsonl"
        old.write_text(claude_session(self.root, "s5", [("make too-old", False, "")]))
        past = time.time() - 60 * 86400
        os.utime(old, (past, past))

        day = self.codex / "sessions" / "2026" / "11" / "06"
        day.mkdir(parents=True)
        (day / "rollout-2026-11-06T09-00-00-cx1.jsonl").write_text(codex_session(self.root, [
            (["/bin/zsh", "-lc", "mysql --password=pa55word! -h 127.0.0.1 shop -e 'select 1'"], 0, ""),
            (["/bin/zsh", "-lc", "actionlint .github/workflows/ci.yml"], 127, "zsh: command not found: actionlint"),
            (["/bin/zsh", "-lc", "docker compose up -d db"], 0, "started"),
        ]))
        (day / "rollout-2026-11-06T09-30-00-cx2.jsonl").write_text(codex_session("/elsewhere", [
            (["/bin/zsh", "-lc", "make elsewhere-only"], 0, ""),
        ]))
        # A Codex subagent run shares the cwd and is excluded.
        (day / "rollout-2026-11-06T10-00-00-cx3.jsonl").write_text(codex_session(self.root, [
            (["/bin/zsh", "-lc", "make codex-subagent-only"], 0, ""),
        ], source={"subagent": {"thread_spawn": {}}}))
        # A sibling repository of the same workspace counts only with the workspace root.
        api = projects / "-work-api"
        api.mkdir()
        (api / "s6.jsonl").write_text(claude_session("/work/api", "s6", [("pnpm start", False, "listening")]))
        # A session in shop that moves to the sibling repository: its command belongs to api, not shop.
        (main / "s7.jsonl").write_text(claude_session(self.root, "s7", [("cd /work/api && make api-after-cd", False, "")]))
        # A session started in the parent folder that works inside shop counts for shop.
        parent = projects / "-work"
        parent.mkdir()
        (parent / "s8.jsonl").write_text(claude_session("/work", "s8", [
            ("cd shop && make shop-from-parent", False, ""),
            ("make parent-only", False, ""),
        ]))

    def tearDown(self):
        self.tmp.cleanup()

    def run_script(self, *roots):
        roots = roots or (self.root,)
        result = subprocess.run(
            [sys.executable, "-I", str(SCRIPT), *[a for r in roots for a in ("--root", r)], "--days", "30",
             "--claude-home", str(self.claude), "--codex-home", str(self.codex)],
            capture_output=True, text=True, check=False)
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout, json.loads(result.stdout)

    def test_output_holds_no_secret_value(self):
        raw, _ = self.run_script()
        for secret in SECRETS:
            self.assertNotIn(secret, raw)

    def test_counts_commands_that_worked_and_commands_missing(self):
        _, out = self.run_script()
        commands = {c["command"]: c for c in out["commands"]}
        self.assertEqual(commands["pnpm start"]["sessions"], 2)
        self.assertIn("docker compose up -d db", commands)
        self.assertEqual(commands["docker compose up -d db"]["sessions"], 2)
        self.assertIn("pnpm run dev --port 3001", commands)
        self.assertIn("ssh -p 2222 deploy@qa.local", commands)
        self.assertIn("docker run -e DB_PASSWORD=<redacted> -p 5432:5432 postgres", commands)
        self.assertNotIn("pnpm run broken", commands)
        self.assertIn("MONGOMS_SYSTEM_BINARY=/home/u/.cache/mongod", commands["pnpm test:int"]["env"])
        self.assertIn("MONGOMS_RUNTIME_DOWNLOAD=false", commands["pnpm test:int"]["env"])
        self.assertIn("API_SECRET=<redacted>", commands["pnpm dev"]["env"])
        self.assertIn("NEXT_PUBLIC_APP_ENV=development", commands["pnpm dev"]["env"])
        self.assertFalse([c for c in commands if c.startswith(("python3 -", "node -e"))])
        missing = {m["program"] for m in out["missing"]}
        self.assertEqual(missing, {"timeout", "actionlint"})
        self.assertIn(3001, {p["port"] for p in out["ports"]})
        self.assertIn({"profile": "shop-qa", "region": "mx-central-1"},
                      [{"profile": c["profile"], "region": c["region"]} for c in out["cloud"]])

    def test_excludes_neighbours_subagents_and_old_sessions(self):
        raw, out = self.run_script()
        for marker in ("subagent-only", "neighbour-only", "too-old", "elsewhere-only", "codex-subagent-only", "dev:watch"):
            self.assertNotIn(marker, raw)
        self.assertEqual(out["hosts"]["claude-code"]["sessions"], 5)
        self.assertEqual(out["hosts"]["codex"]["sessions"], 1)
        self.assertEqual(set(out["not_covered"]), {"grok", "pi", "opencode", "cursor", "antigravity"})
        self.assertNotIn("foo", {m["program"] for m in out["missing"]})

    def test_each_command_counts_where_it_ran(self):
        raw, out = self.run_script()
        commands = {c["command"]: c for c in out["commands"]}
        self.assertNotIn("api-after-cd", raw)
        self.assertNotIn("parent-only", raw)
        self.assertEqual(commands["make shop-from-parent"]["where"], ["."])
        _, workspace = self.run_script("/work")
        commands = {c["command"]: c for c in workspace["commands"]}
        self.assertEqual(commands["make api-after-cd"]["where"], ["api"])
        self.assertEqual(commands["make shop-from-parent"]["where"], ["shop"])
        self.assertEqual(commands["make parent-only"]["where"], ["."])

    def test_workspace_root_adds_sibling_repositories_and_says_where(self):
        _, out = self.run_script("/work")
        start = {c["command"]: c for c in out["commands"]}["pnpm start"]
        self.assertEqual(start["sessions"], 3)
        self.assertEqual(set(start["where"]), {"shop", "api"})
        _, alone = self.run_script()
        self.assertEqual({c["command"]: c for c in alone["commands"]}["pnpm start"]["where"], ["."])


if __name__ == "__main__":
    unittest.main()
