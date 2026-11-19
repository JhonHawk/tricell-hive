import base64
import http.server
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/hive-rule-trackers/scripts/jira_rest.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

SAMPLE_VALUE = "tok-SECRET-0123456789"
EMAIL = "dev@example.com"


class Recorder(http.server.BaseHTTPRequestHandler):
    log = []
    fail = False

    def _reply(self, code, payload):
        body = json.dumps(payload).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _handle(self):
        length = int(self.headers.get("Content-Length") or 0)
        body = self.rfile.read(length) if length else b""
        Recorder.log.append({"method": self.command, "path": self.path, "headers": dict(self.headers), "body": body})
        if Recorder.fail:
            return self._reply(500, {"errorMessages": ["boom"]})
        if self.path.endswith("/myself"):
            return self._reply(200, {"accountId": "me"})
        if self.path.endswith("/transitions"):
            return self._reply(200, {"transitions": [{"id": "21", "to": {"name": "En curso", "statusCategory": {"key": "indeterminate"}}}]})
        if self.path.endswith("/worklog") and self.command == "GET":
            return self._reply(200, {"worklogs": [
                {"id": "1", "timeSpentSeconds": 600, "started": "x", "author": {"accountId": "me"}},
                {"id": "2", "timeSpentSeconds": 900, "started": "y", "author": {"accountId": "someone"}},
            ]})
        if self.path.endswith("/worklog") and self.command == "POST":
            return self._reply(201, {"id": "77", "timeSpentSeconds": json.loads(body)["timeSpentSeconds"]})
        if "/worklog/" in self.path and self.command == "PUT":
            return self._reply(200, {"id": self.path.rsplit("/", 1)[1], "timeSpentSeconds": json.loads(body)["timeSpentSeconds"]})
        if self.path.endswith("/attachments"):
            return self._reply(200, [{"id": "9", "filename": "after.png"}])
        return self._reply(404, {})

    do_GET = do_POST = do_PUT = _handle

    def log_message(self, *args):
        pass


class JiraRestTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Recorder)
        threading.Thread(target=cls.server.serve_forever, daemon=True).start()
        cls.base = f"http://127.0.0.1:{cls.server.server_address[1]}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()

    def setUp(self):
        Recorder.log, Recorder.fail = [], False
        self.tmp = tempfile.TemporaryDirectory()
        root = Path(self.tmp.name)
        self.config = root / "config"
        self.secrets = self.config / "hive" / "secrets"
        self.secrets.mkdir(parents=True)
        os.chmod(self.secrets, 0o700)
        self.stored = self.secrets / "jira-api-token"
        self.stored.write_text(SAMPLE_VALUE + "\n")
        os.chmod(self.stored, 0o600)
        bin_dir = root / "bin"
        bin_dir.mkdir()
        acli = bin_dir / "acli"
        acli.write_text(f"#!/bin/sh\nprintf 'Site: example.atlassian.net\\nEmail: {EMAIL}\\n'\n")
        os.chmod(acli, 0o755)
        self.env = {**os.environ, "XDG_CONFIG_HOME": str(self.config), "PATH": f"{bin_dir}{os.pathsep}{os.environ['PATH']}"}

    def tearDown(self):
        self.tmp.cleanup()

    def run_script(self, *args, base=True):
        cmd = [sys.executable, "-I", str(SCRIPT)] + (["--base-url", self.base] if base else []) + list(args)
        return subprocess.run(cmd, env=self.env, capture_output=True, text=True, timeout=60)

    def assert_no_token(self, result):
        for text in (result.stdout, result.stderr):
            self.assertNotIn(SAMPLE_VALUE, text)
            self.assertNotIn(base64.b64encode(f"{EMAIL}:{SAMPLE_VALUE}".encode()).decode(), text)

    def test_worklog_add_sends_only_the_time(self):
        result = self.run_script("worklog-add", "ABC-1", "--seconds", "4200")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(result.stdout), {"id": "77", "seconds": 4200})
        sent = json.loads(Recorder.log[-1]["body"])
        self.assertEqual(set(sent), {"timeSpentSeconds", "started"})
        self.assertEqual(sent["timeSpentSeconds"], 4200)
        auth = Recorder.log[-1]["headers"]["Authorization"]
        self.assertEqual(auth, "Basic " + base64.b64encode(f"{EMAIL}:{SAMPLE_VALUE}".encode()).decode())
        self.assert_no_token(result)

    def test_worklog_list_keeps_only_this_account(self):
        result = self.run_script("worklog-list", "ABC-1")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(json.loads(result.stdout), [{"id": "1", "seconds": 600, "started": "x"}])

    def test_worklog_update_sets_the_time(self):
        result = self.run_script("worklog-update", "ABC-1", "77", "--seconds", "5400")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(Recorder.log[-1]["method"], "PUT")
        self.assertEqual(json.loads(Recorder.log[-1]["body"]), {"timeSpentSeconds": 5400})

    def test_transitions_list_names_and_categories(self):
        result = self.run_script("transitions", "ABC-1")
        self.assertEqual(json.loads(result.stdout), [{"id": "21", "name": "En curso", "category": "indeterminate"}])

    def test_attach_uploads_multipart_without_xsrf_check(self):
        capture = Path(self.tmp.name) / "after.png"
        capture.write_bytes(b"\x89PNG")
        result = self.run_script("attach", "ABC-1", str(capture))
        self.assertEqual(result.returncode, 0, result.stderr)
        req = Recorder.log[-1]
        self.assertEqual(req["headers"]["X-Atlassian-Token"], "no-check")
        self.assertIn(b'filename="after.png"', req["body"])

    def test_http_error_names_route_and_code_only(self):
        Recorder.fail = True
        result = self.run_script("worklog-add", "ABC-1", "--seconds", "600")
        self.assertEqual(result.returncode, 1)
        self.assertIn("returned HTTP 500", result.stderr)
        self.assertNotIn("Authorization", result.stderr)
        self.assert_no_token(result)

    def test_missing_token_file_is_not_created(self):
        self.stored.unlink()
        result = self.run_script("worklog-list", "ABC-1")
        self.assertEqual(result.returncode, 3)
        self.assertFalse(self.stored.exists())
        self.assertEqual(Recorder.log, [])

    def test_unsafe_permissions_stop(self):
        os.chmod(self.stored, 0o644)
        result = self.run_script("worklog-list", "ABC-1")
        self.assertEqual(result.returncode, 4)
        self.assertEqual(Recorder.log, [])

    def test_base_url_outside_loopback_is_refused(self):
        self.base = "https://elsewhere.example.com"
        result = self.run_script("worklog-list", "ABC-1")
        self.assertEqual(result.returncode, 2)
        self.assert_no_token(result)

    def test_worklog_below_a_minute_is_refused(self):
        result = self.run_script("worklog-add", "ABC-1", "--seconds", "30")
        self.assertEqual(result.returncode, 2)
        self.assertEqual(Recorder.log, [])


if __name__ == "__main__":
    unittest.main()
