#!/usr/bin/env python3
"""Behavior tests for the rule-delivery PreToolUse gate.

Every test feeds the script one payload on stdin — shaped like the payloads
recorded from the real harnesses on 2026-09-18 — and asserts on stdout, stderr
and the exit code. No harness, no network: the hook is a pure stdin/stdout
filter over a manifest.

The manifest is always a fixture here (`HIVE_RULE_MANIFEST`), never the repo's
own: a fixture pins the shapes each test needs — including the one the real
manifest no longer produces, a glob-scoped rule still under `global/rules/`,
which the hook must keep refusing to gate.
"""

import json
import os
import shlex
import shutil
import subprocess
import sys
import tempfile
import threading
import time
import unittest
from pathlib import Path

HOOK = Path(__file__).with_name("rule-delivery.py")
HOOK_DIR = HOOK.parent

# The reason template the gate emits, mirrored here so a change to either side
# shows up as a failing test rather than as a silently different message.
# Only Grok clips a denial reason (~264 visible chars); Claude and Codex carry
# 24 KB, so they name every pending rule in one denial.
GROK_BUDGET = 260
MAX_DENIALS = 3
PREFIX = "Held: read "
SUFFIX = " first, then re-issue this call."
# The tests never wait on the real 5 s valve window; they set it explicitly.
NO_WINDOW = {"HIVE_RULE_DELIVERY_WINDOW": "0"}
WIDE_WINDOW = {"HIVE_RULE_DELIVERY_WINDOW": "600"}


def reason_for(paths, *, harness="claude"):
    """The compact denial text: each directory named once, basenames after it."""
    grouped = {}
    for path in paths:
        directory, name = os.path.split(str(path))
        grouped.setdefault(directory, []).append(name)
    listed = " and ".join(
        f"{', '.join(names)} in {directory.rstrip('/')}/" if directory
        else ", ".join(names)
        for directory, names in grouped.items())
    if harness == "codex":
        return (PREFIX + listed + " first — run: cat "
                + " ".join(str(path) for path in paths)
                + " — then re-issue this call.")
    return PREFIX + listed + SUFFIX


def rule(name, globs, reference, *, source_dir="global/rules-situational",
         always_on=False, readers=False, roots=("claude", "agents"),
         exclusive_with=()):
    """One manifest rule entry, shaped exactly like harness/build.py emits it."""
    return {
        "name": name,
        "source": f"{source_dir}/{name}.md",
        "globs": list(globs),
        "exclusive_with": list(exclusive_with),
        "always_on": always_on,
        "references": {root: str(reference) for root in roots},
        "readers": readers,
    }


# --- payload builders ----------------------------------------------------
#
# Key sets copied from the recordings: lean-probe (Claude Code 2.1.277),
# grok-hook-proto (Grok 1.0.34, both key spellings, RELATIVE paths) and
# codex-hook-proto (Codex 0.155.0, `turn_id`, `apply_patch`, shell as `Bash`).

def claude_write(path, *, session="s1", tool="Write", tool_use_id="toolu_01W", **extra):
    payload = {
        "session_id": session,
        "transcript_path": "",
        "cwd": "/repo",
        "prompt_id": "p-1",
        "permission_mode": "auto",
        "hook_event_name": "PreToolUse",
        "tool_name": tool,
        "tool_input": {"file_path": path, "content": "x"},
        "tool_use_id": tool_use_id,
    }
    payload.update(extra)
    return payload


def claude_read(path, **extra):
    payload = claude_write(path, tool="Read", tool_use_id="toolu_01R", **extra)
    payload["tool_input"] = {"file_path": path}
    return payload


def claude_bash(command, **extra):
    payload = claude_write("", tool="Bash", tool_use_id="toolu_01B", **extra)
    payload["tool_input"] = {"command": command}
    return payload


def observed_read(path, **extra):
    """A read that RAN: observation lives on PostToolUse now."""
    payload = claude_read(path, **extra)
    payload["hook_event_name"] = "PostToolUse"
    payload["tool_response"] = {"type": "text", "file": {"filePath": path}}
    return payload


def observed_bash(command, **extra):
    payload = claude_bash(command, **extra)
    payload["hook_event_name"] = "PostToolUse"
    payload["tool_response"] = {"stdout": "…", "stderr": "", "interrupted": False}
    return payload


def grok_call(tool, tool_input, *, session="g1", cwd="/repo",
              tool_use_id="call-aaaa-bbbb-0", **extra):
    payload = {
        "hookEventName": "pre_tool_use",
        "sessionId": session,
        "cwd": cwd,
        "workspaceRoot": cwd + "/",
        "permissionMode": "bypassPermissions",
        "toolName": tool,
        "toolUseId": tool_use_id,
        "toolInput": dict(tool_input),
        "toolInputTruncated": False,
        "hook_event_name": "PreToolUse",
        "session_id": session,
        "permission_mode": "bypassPermissions",
        "tool_name": tool,
        "tool_input": dict(tool_input),
        "tool_use_id": tool_use_id,
    }
    payload.update(extra)
    return payload


def grok_write(path, **extra):
    return grok_call("write", {"file_path": path, "content": "x"}, **extra)


def grok_read(path, **extra):
    return grok_call("read_file", {"target_file": path}, **extra)


def observed_grok_read(path, **extra):
    payload = grok_read(path, **extra)
    payload["hookEventName"] = "post_tool_use"
    payload["hook_event_name"] = "PostToolUse"
    payload["toolResponse"] = {"content": "…"}
    return payload


def codex_patch(body, *, session="c1", tool_use_id="exec-1", **extra):
    payload = {
        "session_id": session,
        "turn_id": "t-1",
        "transcript_path": "",
        "cwd": "/repo",
        "hook_event_name": "PreToolUse",
        "model": "gpt-6-astra",
        "permission_mode": "bypassPermissions",
        "tool_name": "apply_patch",
        "tool_input": {"command": body},
        "tool_use_id": tool_use_id,
    }
    payload.update(extra)
    return payload


def codex_bash(command, **extra):
    payload = codex_patch("", **extra)
    payload["tool_name"] = "Bash"
    payload["tool_input"] = {"command": command}
    return payload


def observed_codex_bash(command, **extra):
    payload = codex_bash(command, **extra)
    payload["hook_event_name"] = "PostToolUse"
    payload["tool_response"] = "export const x: number = 1;\n"
    return payload


def patch_body(*headers):
    lines = ["*** Begin Patch"]
    lines.extend(headers)
    lines.append("*** End Patch")
    return "\n".join(lines)


def scratch_base():
    """The shortest writable temp base.

    The reason budget is measured in characters, so a fixture rooted under a
    very long ambient `TMPDIR` would stop fitting and every gate assertion
    would flip to "allowed" for a reason that has nothing to do with the hook.
    """
    candidates = [path for path in (tempfile.gettempdir(), "/tmp")
                  if os.path.isdir(path) and os.access(path, os.W_OK)]
    return min(candidates, key=len) if candidates else tempfile.gettempdir()


class HookCase(unittest.TestCase):
    """A scratch HOME/TMPDIR per test: state never leaks between behaviors."""

    def setUp(self):
        scratch = tempfile.mkdtemp(prefix="hive-rd-", dir=scratch_base())
        self.addCleanup(shutil.rmtree, scratch, ignore_errors=True)
        self.root = Path(scratch)
        self.home = self.root / "home"
        self.home.mkdir()
        self.state_root = self.root / "state"
        self.state_root.mkdir()

    # -- fixtures ---------------------------------------------------------

    def write_rule_text(self, name, body="RULE TEXT"):
        path = self.root / f"{name}.md"
        path.write_text(body, encoding="utf-8")
        return path

    def reference_of_length(self, total, *, name):
        """A readable rule file whose manifest reference is exactly `total` chars.

        The reference is relative to the hook's cwd (which `run_hook` sets to
        the scratch root): it is the only way to pin the reason's length without
        inheriting however long the ambient temp directory happens to be.
        """
        stem = f"{name}/"
        self.assertLessEqual(len(stem) + 4, total, "cannot build a path that short")
        filler = total - len(stem) - len(".md")
        (self.root / name).mkdir(exist_ok=True)
        chunks = []
        while filler > 0:
            take = min(filler, 120)
            chunks.append("x" * take)
            filler -= take
            if filler:
                chunks.append("/")
                filler -= 1
        relative = stem + "".join(chunks) + ".md"
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("RULE TEXT", encoding="utf-8")
        self.assertEqual(len(relative), total)
        return relative

    def write_manifest(self, rules, agents=None, read_only_agents=None, name="rule-manifest.json"):
        path = self.root / name
        path.write_text(json.dumps({
            "_generated": "fixture",
            "agents": agents or {},
            "read_only_agents": list(read_only_agents or []),
            "rules": list(rules),
        }), encoding="utf-8")
        return path

    # -- execution --------------------------------------------------------

    def run_hook(self, payload, *, manifest=None, env=None, raw_stdin=None, cwd=None):
        environment = {
            "PATH": os.environ.get("PATH", ""),
            "HOME": str(self.home),
            "TMPDIR": str(self.state_root),
        }
        if manifest is not None:
            environment["HIVE_RULE_MANIFEST"] = str(manifest)
        environment.update(env or {})
        stdin = raw_stdin if raw_stdin is not None else json.dumps(payload)
        return subprocess.run(
            [sys.executable, str(HOOK)], input=stdin, text=True, capture_output=True,
            env=environment, cwd=str(cwd or self.root), check=False,
        )

    # -- assertions -------------------------------------------------------

    def assertAllowed(self, result, message=""):
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), "", message or result.stdout)
        self.assertEqual(result.stderr.strip(), "", message or result.stderr)

    def held(self, result):
        """The denial reason, or a failure naming what came back instead."""
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stderr.strip(), "",
                         "a traceback must never reach the harness")
        self.assertNotEqual(result.stdout.strip(), "", "expected a denial, got an allow")
        output = json.loads(result.stdout)["hookSpecificOutput"]
        self.assertEqual(output.get("hookEventName"), "PreToolUse")
        self.assertEqual(output.get("permissionDecision"), "deny", result.stdout)
        return output["permissionDecisionReason"]


# --- fail open -----------------------------------------------------------

class FailOpenTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.write_rule_text("ts")
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.text)])

    def test_the_kill_switch_exits_before_reading_anything(self):
        for spelling in ("off", "OFF", "0", "false", "False", "no", "NO", " off "):
            with self.subTest(spelling=spelling):
                self.assertAllowed(self.run_hook(
                    claude_write("/repo/a.ts"), manifest=self.manifest,
                    env={"HIVE_RULE_DELIVERY": spelling}))
        # Any other value leaves the gate armed.
        self.assertNotEqual("", self.held(self.run_hook(
            claude_write("/repo/a.ts"), manifest=self.manifest,
            env={"HIVE_RULE_DELIVERY": "on"})))

    def test_hostile_stdin_is_silent(self):
        for raw in ("", "   ", "not json{", "[1, 2]", "null", "true", '"a string"',
                    '{"tool_input": 5, "tool_name": "Write", "session_id": "s"}',
                    '{"tool_name": ["Write"], "session_id": "s"}',
                    '{"tool_name": "Write", "tool_input": {"file_path": 7}, "session_id": "s"}',
                    '{"tool_name": "Write", "tool_input": {"file_path": ["/repo/a.ts"]},'
                    ' "session_id": "s"}',
                    '{"tool_name": "Write", "tool_input": {"file_path": null}, "session_id": "s"}',
                    '{"tool_name": "Write", "tool_input": {"file_path": "/repo/a\\u0000.ts"},'
                    ' "session_id": "s"}',
                    '{"tool_name": "Write", "tool_input": {"file_path": "/repo/a\\nb.ts"},'
                    ' "session_id": "s"}'):
            with self.subTest(raw=raw[:40]):
                self.assertAllowed(self.run_hook(None, manifest=self.manifest, raw_stdin=raw))

    def test_a_broken_manifest_is_silent(self):
        broken = self.root / "broken.json"
        broken.write_text("{ not json", encoding="utf-8")
        directory = self.root / "a-directory.json"
        directory.mkdir()
        shapes = {
            "rules-not-a-list": {"rules": {"ts": {}}},
            "entry-not-a-dict": {"rules": ["ts"]},
            "globs-a-string": {"rules": [dict(rule("ts", [], self.text), globs="**/*.ts")]},
            "references-not-a-dict": {"rules": [dict(rule("ts", ["**/*.ts"], self.text),
                                                     references="~/x.md")]},
            "name-not-a-string": {"rules": [dict(rule("ts", ["**/*.ts"], self.text), name=7)]},
            "top-level-a-list": [],
        }
        candidates = [broken, directory, self.root / "absent.json"]
        for label, payload in shapes.items():
            path = self.root / f"{label}.json"
            path.write_text(json.dumps(payload), encoding="utf-8")
            candidates.append(path)
        for manifest in candidates:
            with self.subTest(manifest=manifest.name):
                self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest))

    def test_state_that_cannot_be_written_allows(self):
        if os.geteuid() == 0:
            self.skipTest("root ignores the mode bits this test depends on")
        self.state_root.chmod(0o555)
        self.addCleanup(self.state_root.chmod, 0o700)
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=self.manifest))

    def test_without_a_session_id_nothing_is_ever_denied(self):
        payload = claude_write("/repo/a.ts")
        payload.pop("session_id")
        self.assertAllowed(self.run_hook(payload, manifest=self.manifest))

    def test_an_unrelated_tool_is_ignored(self):
        payload = claude_write("/repo/a.ts", tool="Glob")
        self.assertAllowed(self.run_hook(payload, manifest=self.manifest))

    def test_an_event_that_is_not_pre_tool_use_never_denies(self):
        # A denial emitted after the tool already ran decides nothing and reads
        # as a malfunction; PI sends no event name at all, which stays gated.
        for event in ("PostToolUse", "post_tool_use", "Stop", "PreCompact"):
            with self.subTest(event=event):
                payload = claude_write("/repo/a.ts", hook_event_name=event)
                payload["tool_response"] = {"filePath": "/repo/a.ts"}
                self.assertAllowed(self.run_hook(payload, manifest=self.manifest))
        # A completed WRITE of the rule file is not a read of it either.
        completed = claude_write(str(self.text), hook_event_name="PostToolUse")
        self.assertAllowed(self.run_hook(completed, manifest=self.manifest))
        self.assertNotEqual("", self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                        manifest=self.manifest)))


# --- what is deliverable at all ------------------------------------------

class DeliverabilityTests(HookCase):
    def test_a_rule_still_under_global_rules_is_never_gated(self):
        # Claude Code loads global/rules/ natively; gating it here would ask the
        # model to read a rule it already has.
        manifest = self.write_manifest([
            rule("ts", ["**/*.ts"], self.write_rule_text("ts"),
                 source_dir="global/rules/languages"),
        ])
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest))

    def test_an_always_on_or_unscoped_rule_is_never_gated(self):
        text = self.write_rule_text("core")
        manifest = self.write_manifest([
            rule("always", ["**/*.ts"], text, always_on=True),
            rule("unscoped", [], text),
        ])
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest))

    def test_a_reference_that_is_missing_unreadable_or_a_directory_is_never_gated(self):
        absent = self.root / "never-deployed.md"
        folder = self.root / "a-folder.md"
        folder.mkdir()
        unreadable = self.write_rule_text("unreadable")
        unreadable.chmod(0o000)
        self.addCleanup(unreadable.chmod, 0o600)
        cases = [("absent", absent), ("directory", folder)]
        if os.geteuid() != 0:
            cases.append(("unreadable", unreadable))
        for label, reference in cases:
            with self.subTest(reference=label):
                manifest = self.write_manifest([rule("ts", ["**/*.ts"], reference)],
                                               name=f"m-{label}.json")
                self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest))

    def test_a_rule_with_no_reference_for_this_harness_is_never_gated(self):
        manifest = self.write_manifest([
            rule("agents-only", ["**/*.ts"], self.write_rule_text("ao"), roots=("agents",)),
        ])
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest))
        # The same rule DOES gate the harness whose root it names.
        self.assertIn("ao.md", self.held(self.run_hook(
            codex_patch(patch_body("*** Add File: /repo/a.ts")), manifest=manifest)))


# --- the gate, per harness -----------------------------------------------

class ClaudeGateTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.write_rule_text("ts")
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.text)])

    def test_the_first_write_is_denied_naming_the_rule_file(self):
        reason = self.held(self.run_hook(claude_write("/repo/src/a.ts"), manifest=self.manifest))
        self.assertEqual(reason, reason_for([self.text]))
        self.assertNotIn("RULE TEXT", reason, "the gate never carries the rule text")

    def test_every_write_tool_is_gated(self):
        for index, tool in enumerate(("Write", "Edit", "MultiEdit")):
            with self.subTest(tool=tool):
                result = self.run_hook(
                    claude_write("/repo/a.ts", tool=tool, session=f"s-{index}"),
                    manifest=self.manifest)
                self.assertEqual(self.held(result), reason_for([self.text]))

    def test_notebook_edit_carries_its_path_under_notebook_path(self):
        # The field NotebookEdit actually sends (executor-dispatch-gate reads
        # the same one); a fixture using file_path would pass against a payload
        # the harness never produces.
        payload = claude_write("/repo/nb.ts", tool="NotebookEdit")
        payload["tool_input"] = {"notebook_path": "/repo/analysis.ipynb",
                                 "new_source": "x"}
        notebook = self.write_rule_text("nb")
        manifest = self.write_manifest([rule("nb", ["**/*.ipynb"], notebook)])
        self.assertEqual(self.held(self.run_hook(payload, manifest=manifest)),
                         reason_for([notebook]))

    def test_a_write_outside_the_globs_is_allowed(self):
        self.assertAllowed(self.run_hook(claude_write("/repo/README.md"), manifest=self.manifest))

    def test_a_read_of_a_matching_file_is_not_gated_for_a_writer(self):
        self.assertAllowed(self.run_hook(claude_read("/repo/a.ts"), manifest=self.manifest))

    def test_the_observed_read_releases_every_later_write(self):
        self.held(self.run_hook(claude_write("/repo/a.ts"), manifest=self.manifest))
        self.assertAllowed(self.run_hook(observed_read(str(self.text)),
                                         manifest=self.manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=self.manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/b.ts"), manifest=self.manifest))

    def test_a_subagent_keeps_its_own_state(self):
        sub = {"agent_id": "a531600b6bed36537", "agent_type": "probe-writer"}
        self.held(self.run_hook(claude_write("/repo/a.ts", **sub), manifest=self.manifest))
        self.assertAllowed(self.run_hook(observed_read(str(self.text), **sub),
                                         manifest=self.manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts", **sub),
                                         manifest=self.manifest))
        # The main thread of the same session never saw that read.
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                 manifest=self.manifest)),
                         reason_for([self.text]))


class GrokGateTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.write_rule_text("ts")
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.text)])

    def test_a_relative_write_path_is_resolved_against_the_workspace(self):
        reason = self.held(self.run_hook(grok_write("src/a.ts"), manifest=self.manifest))
        self.assertEqual(reason, reason_for([self.text]))

    def test_search_replace_is_gated_and_read_file_releases_it(self):
        self.held(self.run_hook(grok_call("search_replace", {"file_path": "src/a.ts"}),
                                manifest=self.manifest))
        self.assertAllowed(self.run_hook(observed_grok_read(str(self.text)),
                                         manifest=self.manifest))
        self.assertAllowed(self.run_hook(grok_call("search_replace", {"file_path": "src/a.ts"}),
                                         manifest=self.manifest))

    def test_the_workspace_root_resolves_the_path_when_cwd_is_absent(self):
        payload = grok_write("src/a.ts")
        payload.pop("cwd")
        self.assertEqual(self.held(self.run_hook(payload, manifest=self.manifest)),
                         reason_for([self.text]))


class CodexGateTests(HookCase):
    def setUp(self):
        super().setUp()
        self.ts = self.write_rule_text("ts")
        self.sql = self.write_rule_text("sql")
        self.manifest = self.write_manifest([
            rule("ts", ["**/*.ts"], self.ts),
            rule("sql", ["**/*.sql"], self.sql),
        ])

    def test_every_authoring_header_of_a_patch_is_a_target(self):
        for index, header in enumerate((
            "*** Add File: /repo/src/a.ts",
            "*** Update File: /repo/src/a.ts",
            "*** Move to: /repo/src/a.ts",
        )):
            with self.subTest(header=header):
                result = self.run_hook(codex_patch(patch_body(header), session=f"c-{index}"),
                                       manifest=self.manifest)
                self.assertEqual(self.held(result),
                                 reason_for([self.ts], harness="codex"))

    def test_a_delete_only_patch_authors_nothing_and_is_allowed(self):
        self.assertAllowed(self.run_hook(
            codex_patch(patch_body("*** Delete File: /repo/src/a.ts")), manifest=self.manifest))

    def test_a_multi_file_patch_names_every_rule_it_touches(self):
        body = patch_body("*** Add File: /repo/src/a.ts", "*** Update File: /repo/db/x.sql")
        reason = self.held(self.run_hook(codex_patch(body), manifest=self.manifest))
        self.assertEqual(reason, reason_for([self.ts, self.sql], harness="codex"))

    def test_a_rename_carries_both_the_source_and_the_destination(self):
        body = patch_body("*** Update File: /repo/src/a.txt", "*** Move to: /repo/src/a.ts")
        self.assertEqual(self.held(self.run_hook(codex_patch(body), manifest=self.manifest)),
                         reason_for([self.ts], harness="codex"))

    def test_a_shell_read_releases_the_patch_gate(self):
        body = patch_body("*** Add File: /repo/src/a.ts")
        self.held(self.run_hook(codex_patch(body), manifest=self.manifest))
        self.assertAllowed(self.run_hook(observed_codex_bash(f"cat {self.ts}"),
                                         manifest=self.manifest))
        self.assertAllowed(self.run_hook(codex_patch(body), manifest=self.manifest))

    def test_a_commit_message_quoting_a_patch_header_is_not_a_patch(self):
        # A header only authors a file when it starts a line AND the command
        # actually opens a patch.
        for command in (
            "git commit -m 'fix: *** Update File: src/a.ts was renamed'",
            "git commit -m 'summary\n*** Update File: src/a.ts\n'",
            "echo '*** Add File: src/a.ts'",
        ):
            with self.subTest(command=command):
                self.assertAllowed(self.run_hook(codex_bash(command),
                                                 manifest=self.manifest))

    def test_a_patch_written_through_the_shell_is_a_write_and_is_gated(self):
        # `apply_patch` is a shell verb too: the same headers inside a Bash
        # command author the same files, and this is the ONE case in which a
        # terminal call is denied.
        body = patch_body("*** Add File: /repo/src/a.ts")
        command = f"apply_patch <<'PATCH'\n{body}\nPATCH"
        self.assertEqual(self.held(self.run_hook(codex_bash(command),
                                                 manifest=self.manifest)),
                         reason_for([self.ts], harness="codex"))
        # …while an ordinary shell call is never denied.
        self.assertAllowed(self.run_hook(codex_bash("ls /repo/src"), manifest=self.manifest))


# --- observation ---------------------------------------------------------

class ObservationTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.home / "rules" / "ts.md"
        self.text.parent.mkdir(parents=True)
        self.text.write_text("RULE TEXT", encoding="utf-8")
        self.reference = f"~/rules/{self.text.name}"
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.reference)])

    def read_releases(self, build, *, expected=True, session="s1"):
        """Run one candidate observation, then check whether a write got through.

        `build` receives the session id so each case runs against its own state
        instead of re-creating the fixture (which would move the rule file the
        payload names).
        """
        self.assertAllowed(self.run_hook(build(session), manifest=self.manifest,
                                         cwd=self.root))
        write = self.run_hook(claude_write("/repo/a.ts", session=session),
                              manifest=self.manifest, env=WIDE_WINDOW)
        if expected:
            self.assertAllowed(write, "the read should have released the gate")
        else:
            self.assertEqual(self.held(write), reason_for([self.text]),
                             "this read must NOT release the gate")

    def test_the_gate_names_the_expanded_reference(self):
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                 manifest=self.manifest)),
                         reason_for([self.text]))

    def test_observation_happens_on_completion_not_before_the_call(self):
        # A PreToolUse read is an INTENTION: it can still be refused (the rule
        # files live outside the project) and marking it known would lie.
        self.read_releases(lambda session: claude_read(str(self.text), session=session),
                           expected=False, session="pre")
        self.read_releases(lambda session: observed_read(str(self.text), session=session),
                           session="post")

    def test_the_read_tool_releases_through_every_spelling_of_the_path(self):
        symlink = self.root / "linked-ts.md"
        symlink.symlink_to(self.text)
        swapped = str(self.text.parent / self.text.name.upper())
        spellings = {
            "absolute": str(self.text),
            "tilde": self.reference,
            "relative-to-cwd": os.path.relpath(self.text, self.root),
            "symlink": str(symlink),
            "dot-segments": str(self.text.parent / "." / ".." / self.text.parent.name
                                / self.text.name),
        }
        for index, (label, spelling) in enumerate(spellings.items()):
            with self.subTest(spelling=label):
                # A relative read resolves against the session's cwd, which the
                # payload carries — never against the hook process's own.
                self.read_releases(
                    lambda session: observed_read(spelling, session=session,
                                                  cwd=str(self.root)),
                    session=f"s{index}")
        # A case-insensitive volume resolves the swapped spelling to the same
        # file; a case-sensitive one does not, and then it must NOT release.
        self.read_releases(lambda session: observed_read(swapped, session=session),
                           expected=os.path.exists(swapped), session="s-case")

    def test_a_read_that_starts_past_the_top_of_the_file_shows_nothing_of_it(self):
        payload = observed_read(str(self.text))
        payload["tool_input"] = {"file_path": str(self.text), "offset": 9999, "limit": 1}
        self.read_releases(lambda session: dict(payload, session_id=session),
                           expected=False)
        # offset 1 IS the top of the file.
        top = observed_read(str(self.text))
        top["tool_input"] = {"file_path": str(self.text), "offset": 1}
        self.read_releases(lambda session: dict(top, session_id=session), session="top")

    def test_a_reader_command_naming_the_full_path_releases(self):
        directory = self.text.parent
        commands = (f"cat {self.text}",
                    f"/bin/cat {self.text}",
                    f"sed -n '1,40p' {self.text}",
                    f"head -n 50 {self.text}",
                    f"tail -n 5 {self.text}",
                    f"nl {self.text}",
                    f"less {self.text}",
                    f"cat {self.reference}",
                    f"cat /repo/src/a.ts {self.text}",
                    f"cat {self.text} | head -20",
                    # The forms the reviewer measured as false NON-releases.
                    f"bash -lc 'cat {self.text}'",
                    f"sh -c \"cat {self.text}\"",
                    f"FOO=1 BAR=2 cat {self.text}",
                    f"sudo cat {self.text}",
                    f"env cat {self.text}",
                    f"command cat {self.text}",
                    f"cd /repo && cat {self.text}",
                    # `cd` moves the base the LATER stages resolve against: the
                    # payload's cwd is elsewhere, which is the only way this
                    # case proves anything.
                    f"cd {directory} && cat {self.text.name}",
                    f"cd {directory} && cat ./{self.text.name}",
                    f"ls /repo; cat {self.text}",
                    # Codex writes the reason's own command back with a glob or
                    # a brace group — expanded against the filesystem, never run.
                    f"cat {directory}/*.md",
                    f"cat {directory}/{{{self.text.stem},absent}}.md",
                    f"cd {directory} && cat *.md")
        for index, command in enumerate(commands):
            with self.subTest(command=command):
                self.read_releases(
                    lambda session: observed_bash(command, session=session,
                                                  cwd=str(self.root)),
                    session=f"c{index}")

    def test_a_wildcard_outside_the_last_component_is_not_expanded(self):
        # One pattern, one directory listing. A wildcard in a parent component
        # makes the expansion walk the tree — `cat ~/*/*/*/*/*/*/*/*/*` measured
        # 6.3 s against a 15 s hook timeout — and no model needs it to read a
        # rule whose directory the reason just named in full.
        started = time.monotonic()
        self.read_releases(
            lambda session: observed_bash(
                f"cat {self.text.parent.parent}/*/{self.text.name}", session=session),
            expected=False)
        self.read_releases(
            lambda session: observed_bash(f"cat {self.text.parent}/*/*.md",
                                          session=session),
            expected=False, session="deep")
        self.read_releases(
            lambda session: observed_bash(f"cat {self.text.parent}/{{a,b}}*/*.md",
                                          session=session),
            expected=False, session="brace")
        self.assertLess(time.monotonic() - started, 10)
        # Confined to the last component, it still releases.
        self.read_releases(
            lambda session: observed_bash(f"cat {self.text.parent}/*.md",
                                          session=session),
            session="flat")

    def test_a_loop_is_not_an_observation_and_never_pretends_to_be(self):
        # `for f in …; do cat "$f"; done` shows the file, but the hook does not
        # track loop variables: it is not observed, and Codex's reason names a
        # command that IS.
        self.read_releases(
            lambda session: observed_bash(
                f'for f in {self.text}; do cat "$f"; done', session=session),
            expected=False)

    def test_an_argv_list_command_is_normalized_before_it_is_read(self):
        # Codex sends shell argv as an array (see rule-context.sh).
        for index, argv in enumerate((["bash", "-lc", f"cat {self.text}"],
                                      ["cat", str(self.text)],
                                      ["/bin/cat", str(self.text)])):
            with self.subTest(argv=argv):
                self.read_releases(
                    lambda session: observed_bash(list(argv), session=session),
                    session=f"v{index}")

    def test_a_command_that_shows_nothing_of_the_file_does_not_release(self):
        other = self.root / "other.md"
        other.write_text("other", encoding="utf-8")
        decoy = self.root / "decoy"
        decoy.mkdir()
        (decoy / self.text.name).write_text("NOT THE RULE", encoding="utf-8")
        (self.root / f"{self.text.name}.bak").write_text("backup", encoding="utf-8")
        commands = (
            # reader, but the file is an argument to a NON-reader stage
            f"cat {other} | grep -f {self.text}",
            f"cat {other}; rm {self.text}",
            f"sed -n 1p {other} && echo {self.text}",
            f"cat /dev/null && ls {self.text}",
            f"ls -la {self.text}",
            f"echo {self.text}",
            f"git log --oneline -- {self.text}",
            f"grep -n TODO {self.text}",
            f"awk 'BEGIN{{exit}}' {self.text}",
            # reader, but it prints nothing
            f"head -0 {self.text}",
            f"head -c0 {self.text}",
            f"head -n 0 {self.text}",
            f"head -n0 {self.text}",
            f"tail -c 0 {self.text}",
            f"tail -f {self.text}",
            f"sed -i '' s/a/b/ {self.text}",
            # reader, but the output never reaches the model
            f"cat {self.text} > /tmp/copy.md",
            f"cat {self.text} >> /tmp/copy.md",
            f"cat {self.text} | grep -c .",
            f"cat > {self.text} <<EOF",
            # a different file that merely SHARES a substring or a basename
            f"cat {self.text}.bak",
            f"cat /backup{self.text}",
            f"cat {self.text.name}",
            f"cat {decoy / self.text.name}",
            # mentioned in a comment, never opened
            f"tail -f app.log # see {self.text}",
            f"python3 -c 'print(1)' # cat {self.text}",
        )
        for index, command in enumerate(commands):
            with self.subTest(command=command):
                self.read_releases(
                    lambda session: observed_bash(command, session=session),
                    expected=False, session=f"n{index}")

    def test_an_unparsable_command_is_not_an_observation(self):
        self.read_releases(
            lambda session: observed_bash(f"cat '{self.text}", session=session),
            expected=False)

    def test_ordinary_shapes_that_still_show_the_file_release(self):
        commands = (
            # stderr is not the model's channel: redirecting it hides nothing
            f"cat {self.text} 2>/dev/null",
            f"cat {self.text} 2>&1 | head -200",
            f"sed -n '1,80p' {self.text} 2>/dev/null",
            # a multi-line command is a sequence, not one line
            f"echo reading\ncat {self.text}",
            f"cd /repo\ncat {self.text}\necho done",
            # the spellings a model actually types
            "cat \"$HOME/rules/ts.md\"",
            "cat $HOME/rules/ts.md",
            "cat ~/rules/ts.md",
        )
        for index, command in enumerate(commands):
            with self.subTest(command=command):
                self.read_releases(
                    lambda session: observed_bash(command, session=session),
                    session=f"o{index}")

    def test_the_shapes_left_out_by_design_do_not_release(self):
        commands = (f"grep '' {self.text}",
                    f"rg . {self.text}",
                    f"awk '{{print}}' {self.text}",
                    f"cat {self.text} | tee /tmp/copy.md",
                    f"cat < {self.text}",
                    f"echo $(cat {self.text})",
                    f"(cat {self.text})")
        for index, command in enumerate(commands):
            with self.subTest(command=command):
                self.read_releases(
                    lambda session: observed_bash(command, session=session),
                    expected=False, session=f"d{index}")

    def test_a_read_that_failed_is_not_an_observation(self):
        failures = ({"is_error": True},
                    {"isError": True},
                    {"exit_code": 1},
                    {"exitCode": 127},
                    {"interrupted": True},
                    {"stdout": "", "stderr": "No such file", "exit_code": 2})
        for index, response in enumerate(failures):
            with self.subTest(response=response):
                def build(session, response=response):
                    payload = observed_bash(f"cat {self.text}", session=session)
                    payload["tool_response"] = response
                    return payload
                self.read_releases(build, expected=False, session=f"f{index}")
        # A success, or a shape the hook does not recognize, still observes.
        for index, response in enumerate(({"exit_code": 0, "interrupted": False},
                                          {"stdout": "RULE TEXT"},
                                          "plain text response",
                                          None)):
            with self.subTest(ok=response):
                def build(session, response=response):
                    payload = observed_read(str(self.text), session=session)
                    payload["tool_response"] = response
                    return payload
                self.read_releases(build, session=f"k{index}")

    def test_a_read_of_another_file_with_the_same_basename_does_not_release(self):
        decoy = self.root / "decoy" / self.text.name
        decoy.parent.mkdir()
        decoy.write_text("NOT THE RULE", encoding="utf-8")
        self.read_releases(lambda session: observed_read(str(decoy), session=session),
                           expected=False)

    def test_a_terminal_call_that_authors_nothing_is_never_denied(self):
        for command in ("ls /repo",
                        "python3 -c \"open('/repo/a.ts','w')\"",
                        f"cat {self.text}",
                        "cat > /repo/a.ts <<EOF\nx\nEOF"):
            with self.subTest(command=command):
                self.assertAllowed(self.run_hook(claude_bash(command),
                                                 manifest=self.manifest))


# --- concurrency ---------------------------------------------------------

class ParallelWriteTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.write_rule_text("ts")
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.text)])

    def fire(self, targets, *, tool="Write", env=None):
        """Eight processes released at once, the way one turn arrives."""
        barrier = threading.Barrier(len(targets))
        results = [None] * len(targets)

        def go(index):
            barrier.wait()
            results[index] = self.run_hook(
                claude_write(targets[index], tool=tool, tool_use_id=f"toolu_{index}"),
                manifest=self.manifest, env=env)

        threads = [threading.Thread(target=go, args=(index,))
                   for index in range(len(targets))]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        return results

    def known_markers(self):
        return [name
                for state in (self.state_root / "hive-rule-delivery").iterdir()
                for name in os.listdir(state) if name.startswith("k")]

    def test_eight_parallel_writes_of_distinct_targets_are_all_denied(self):
        targets = [f"/repo/src/f{index}.ts" for index in range(8)]
        for index, result in enumerate(self.fire(targets)):
            self.assertEqual(self.held(result), reason_for([self.text]), f"sibling {index}")
        state = list((self.state_root / "hive-rule-delivery").iterdir())
        self.assertEqual(len(state), 1, "one directory per session+agent")
        self.assertEqual(self.known_markers(), [],
                         "a burst of denials must not make the rule known")

        self.assertAllowed(self.run_hook(observed_read(str(self.text)),
                                         manifest=self.manifest))
        for index, result in enumerate(self.fire(targets)):
            self.assertAllowed(result, f"sibling {index} after the read")

    def test_eight_parallel_edits_of_the_SAME_target_are_all_denied(self):
        # The reproduced defect: per-target ordinals let one turn burn the whole
        # valve, so 3 calls were denied and 5 wrote the file with no read at all.
        targets = ["/repo/a.ts"] * 8
        for index, result in enumerate(self.fire(targets, tool="Edit")):
            self.assertEqual(self.held(result), reason_for([self.text]), f"sibling {index}")
        self.assertEqual(self.known_markers(), [],
                         "one turn must count as at most one denial")

    def test_a_burst_counts_once_so_the_valve_still_has_room(self):
        self.fire(["/repo/a.ts"] * 8, tool="Edit")
        # Two more counted denials (the window is open) and only then a pass.
        self.assertNotEqual("", self.held(self.run_hook(claude_write("/repo/b.ts"),
                                                        manifest=self.manifest,
                                                        env=NO_WINDOW)))
        self.assertAllowed(self.run_hook(claude_write("/repo/c.ts"),
                                         manifest=self.manifest, env=NO_WINDOW))


# --- the release valve ---------------------------------------------------

class ReleaseValveTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.write_rule_text("ts")
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.text)])

    def test_three_counted_denials_release_the_rule_for_good(self):
        for attempt, target in enumerate(("/repo/a.ts", "/repo/b.ts")):
            with self.subTest(attempt=attempt):
                self.assertEqual(self.held(self.run_hook(claude_write(target),
                                                         manifest=self.manifest,
                                                         env=NO_WINDOW)),
                                 reason_for([self.text]))
        # The third counted denial IS the release: this call goes through.
        self.assertAllowed(self.run_hook(claude_write("/repo/c.ts"),
                                         manifest=self.manifest, env=NO_WINDOW))
        for target in ("/repo/d.ts", "/repo/a.ts"):
            self.assertAllowed(self.run_hook(claude_write(target), manifest=self.manifest),
                               "the valve must stay open, for the whole rule")

    def test_denials_inside_the_window_do_not_count(self):
        # A model retrying in a tight loop, or a turn with many writes, must not
        # spend the valve: nothing counts until the window has passed.
        for attempt in range(12):
            target = f"/repo/f{attempt}.ts"
            with self.subTest(attempt=attempt):
                self.assertEqual(self.held(self.run_hook(claude_write(target),
                                                         manifest=self.manifest,
                                                         env=WIDE_WINDOW)),
                                 reason_for([self.text]))

    def test_the_valve_advances_on_the_real_clock(self):
        # The only test that lets real time pass: with a window of 0.3 s a model
        # retrying every ~70 ms must still reach the release, which it cannot if
        # an uncounted denial refreshes the counter it is waiting on.
        window = 0.3
        env = {"HIVE_RULE_DELIVERY_WINDOW": str(window)}
        started = time.monotonic()
        denials = 0
        while True:
            result = self.run_hook(claude_write(f"/repo/f{denials}.ts"),
                                   manifest=self.manifest, env=env)
            if not result.stdout.strip():
                break
            self.assertEqual(self.held(result), reason_for([self.text]))
            denials += 1
            self.assertLess(time.monotonic() - started, 5.0,
                            "the valve never advanced: a denial inside the window "
                            "must not push the counter it is waiting on")
            time.sleep(0.05)
        elapsed = time.monotonic() - started
        self.assertGreater(denials, MAX_DENIALS,
                           "retries inside the window must be denied without counting")
        self.assertGreater(elapsed, (MAX_DENIALS - 1) * window * 0.8,
                           "the window has to be respected, not skipped")
        self.assertLess(elapsed, 5.0)

    def test_a_marker_dated_in_the_future_does_not_freeze_the_counter(self):
        # A clock that went backwards (NTP, a suspended laptop) leaves a marker
        # dated ahead; a negative age must read as "the window has passed".
        def shift_markers_into_the_future():
            for state in (self.state_root / "hive-rule-delivery").iterdir():
                for name in os.listdir(state):
                    ahead = time.time() + 3600
                    os.utime(os.path.join(state, name), (ahead, ahead))

        self.held(self.run_hook(claude_write("/repo/a.ts"), manifest=self.manifest,
                                env=WIDE_WINDOW))
        shift_markers_into_the_future()
        self.held(self.run_hook(claude_write("/repo/b.ts"), manifest=self.manifest,
                                env=WIDE_WINDOW))
        shift_markers_into_the_future()
        self.assertAllowed(self.run_hook(claude_write("/repo/c.ts"),
                                         manifest=self.manifest, env=WIDE_WINDOW))

    def markers(self):
        state = self.state_root / "hive-rule-delivery"
        return {path for directory in state.iterdir() for path in directory.iterdir()}

    def test_a_rule_the_budget_left_out_still_holds_when_the_named_one_releases(self):
        # The leak: when every rule NAMED in a denial hits its valve, the reason
        # is empty and the write used to pass — while a rule the budget deferred
        # had been neither read nor released. Releasing one rule is not a reason
        # to stop holding for another.
        first = self.reference_of_length(200, name="one")
        second = self.reference_of_length(200, name="two")
        manifest = self.write_manifest([rule("one", ["**/*.ts"], first),
                                        rule("two", ["**/*.ts"], second)])
        for attempt in range(3):
            self.held(self.run_hook(grok_write(f"src/f{attempt}.ts"), manifest=manifest,
                                    env=NO_WINDOW))
        older = self.markers()
        self.held(self.run_hook(grok_write("src/f3.ts"), manifest=manifest, env=NO_WINDOW))
        # Age everything the FOURTH call did not create: `one` may now walk its
        # counter to the end, `two` is still inside its window.
        for marker in older:
            ahead = time.time() - 3600
            os.utime(marker, (ahead, ahead))

        result = self.run_hook(grok_write("src/f4.ts"), manifest=manifest,
                               env=WIDE_WINDOW)
        self.assertEqual(self.held(result), reason_for([second]),
                         "the rule the budget deferred must keep holding")

    def test_a_rule_left_out_of_the_reason_is_never_counted(self):
        # Grok's 260-char reason fits one of these paths at a time; the rule that
        # was not named cannot be counted, so each takes its own three rounds.
        first = self.reference_of_length(200, name="one")
        second = self.reference_of_length(200, name="two")
        manifest = self.write_manifest([
            rule("one", ["**/*.ts"], first),
            rule("two", ["**/*.ts"], second),
        ])
        seen = []
        for attempt in range(6):
            result = self.run_hook(grok_write(f"src/f{attempt}.ts"), manifest=manifest,
                                   env=NO_WINDOW)
            seen.append("ALLOW" if not result.stdout.strip() else self.held(result))
        named = [reason_for([first]), reason_for([second])]
        self.assertEqual(seen, [named[0], named[1], named[0], named[1], "ALLOW", "ALLOW"],
                         "each rule takes three counted denials of its own")


# --- the reason budget ---------------------------------------------------

class ReasonBudgetTests(HookCase):
    """Only Grok clips, so only Grok pays for it."""

    def five_rules(self):
        """Five realistic references, as a `.tsx` write really matches."""
        names = ("identifier-language", "patterns-antipatterns", "react-nextjs",
                 "tailwind", "typescript-standards")
        directory = self.home / ".claude/skills/language-rules/references"
        directory.mkdir(parents=True)
        references = []
        for name in names:
            path = directory / f"{name}.md"
            path.write_text("RULE TEXT", encoding="utf-8")
            references.append(path)
        return references, self.write_manifest(
            [rule(path.stem, ["**/*.tsx"], path) for path in references])

    def test_claude_names_every_pending_rule_in_one_denial(self):
        references, manifest = self.five_rules()
        reason = self.held(self.run_hook(claude_write("/repo/App.tsx"), manifest=manifest))
        self.assertEqual(reason, reason_for(references),
                         "a 24 KB reason has no excuse for three denial rounds")

    def test_codex_names_every_pending_rule_in_one_denial(self):
        references, manifest = self.five_rules()
        body = patch_body("*** Add File: /repo/App.tsx")
        reason = self.held(self.run_hook(codex_patch(body), manifest=manifest))
        self.assertEqual(reason, reason_for(references, harness="codex"))
        # Codex has no read tool, so the reason hands it the command itself.
        self.assertIn("run: cat " + " ".join(str(path) for path in references), reason)

    def test_the_codex_command_is_quoted_so_it_can_be_pasted(self):
        # A HOME with a space makes the un-quoted form `cat /Users/Jo Smith/…`
        # two arguments: Codex runs it, sees two errors, and the rule it was
        # told to read stays unread for as long as the valve holds.
        directory = self.home / "My Rules"
        directory.mkdir(parents=True)
        text = directory / "ts.md"
        text.write_text("RULE TEXT", encoding="utf-8")
        manifest = self.write_manifest([rule("ts", ["**/*.ts"], text)])
        reason = self.held(self.run_hook(
            codex_patch(patch_body("*** Add File: /repo/a.ts")), manifest=manifest))
        self.assertIn(f"run: {shlex.quote(str(text))}".replace("run: ", "run: cat "),
                      reason)
        # And the quoted command really does name exactly one file.
        command = reason.split("run: ")[1].split(" — then")[0]
        self.assertEqual(shlex.split(command), ["cat", str(text)])

    def test_the_reason_names_a_directory_once_and_the_basenames_after_it(self):
        # Six absolute paths do not fit Grok's ~264 visible characters, so the
        # gate used to spend three to five sequential denial rounds on a single
        # `.tsx`. The rules live in ONE directory: naming it once fits them all.
        references, _ = self.five_rules()
        extra = references[0].with_name("ui-visual-design.md")
        extra.write_text("RULE TEXT", encoding="utf-8")
        references.append(extra)
        manifest = self.write_manifest(
            [rule(path.stem, ["**/*.tsx"], path) for path in references])
        directory = references[0].parent
        expected = (PREFIX + ", ".join(path.name for path in references)
                    + f" in {directory}/" + SUFFIX)

        reason = self.held(self.run_hook(grok_write("src/App.tsx"), manifest=manifest))
        self.assertEqual(reason, expected)
        self.assertLessEqual(len(reason), GROK_BUDGET,
                             "six rules in one directory have to fit Grok's clip")
        # Claude uses the same compact form: one shape to read, one to test.
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/src/App.tsx"),
                                                 manifest=manifest)),
                         expected)

    def test_rules_in_different_directories_each_carry_their_own(self):
        first = self.write_rule_text("ts")
        other_dir = self.root / "elsewhere"
        other_dir.mkdir()
        second = other_dir / "py.md"
        second.write_text("RULE TEXT", encoding="utf-8")
        manifest = self.write_manifest([rule("ts", ["**/*.ts"], first),
                                        rule("py", ["**/*.ts"], second)])
        self.assertEqual(
            self.held(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest)),
            f"{PREFIX}{first.name} in {self.root}/ and {second.name} "
            f"in {other_dir}/{SUFFIX}")

    def test_grok_stays_inside_its_clip(self):
        references, manifest = self.five_rules()
        reason = self.held(self.run_hook(grok_write("src/App.tsx"), manifest=manifest))
        self.assertLessEqual(len(reason), GROK_BUDGET)
        # The directory is named once; every name after it is a whole basename
        # of a pending rule, never a path the clip cut in half.
        self.assertEqual(reason.count(str(references[0].parent)), 1)
        listed = reason[len(PREFIX):-len(SUFFIX)].split(" in ")[0].split(", ")
        self.assertTrue(all(name in [path.name for path in references] for name in listed),
                        f"a clipped path reached the model: {listed}")

    def test_two_rules_share_one_grok_reason_when_both_paths_fit(self):
        first = self.reference_of_length(60, name="one")
        second = self.reference_of_length(60, name="two")
        manifest = self.write_manifest([
            rule("one", ["**/*.ts"], first),
            rule("two", ["**/*.ts"], second),
        ])
        reason = self.held(self.run_hook(grok_write("src/a.ts"), manifest=manifest))
        self.assertEqual(reason, reason_for([first, second]))
        self.assertLessEqual(len(reason), GROK_BUDGET)

    def budget_length(self):
        """The reference length whose rendered reason exactly fills Grok's clip."""
        probe = "d/x.md"
        return GROK_BUDGET - (len(reason_for([probe])) - len(probe))

    def test_a_path_at_grok_s_budget_is_named_whole_and_one_over_is_not_gated(self):
        fits = self.reference_of_length(self.budget_length(), name="fits")
        manifest = self.write_manifest([rule("ts", ["**/*.ts"], fits)], name="fits.json")
        reason = self.held(self.run_hook(grok_write("src/a.ts"), manifest=manifest))
        self.assertEqual(reason, reason_for([fits]))
        self.assertEqual(len(reason), GROK_BUDGET)

        over = self.reference_of_length(self.budget_length() + 1, name="over")
        manifest = self.write_manifest([rule("ts", ["**/*.ts"], over)], name="over.json")
        self.assertAllowed(self.run_hook(grok_write("src/a.ts"), manifest=manifest),
                           "a path Grok cannot show whole is never gated there")
        # Claude has room for it, so there it IS gated — named in full.
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                 manifest=manifest)),
                         reason_for([over]))

    def test_a_reference_far_longer_than_any_budget_is_never_clipped(self):
        # Built explicitly long (~420 chars), not by hoping the temp dir is.
        deep = self.home.joinpath(*[f"segment-{index:02d}-{'x' * 20}" for index in range(12)],
                                  "very-long-rule-name.md")
        deep.parent.mkdir(parents=True)
        deep.write_text("RULE TEXT", encoding="utf-8")
        self.assertGreater(len(str(deep)), 400)
        manifest = self.write_manifest([rule("deep", ["**/*.ts"], deep)])
        self.assertAllowed(self.run_hook(grok_write("src/a.ts"), manifest=manifest),
                           "Grok would clip it, so Grok is not gated on it")
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                 manifest=manifest)),
                         reason_for([deep]))


# --- who is gated --------------------------------------------------------

class AgentScopeTests(HookCase):
    def setUp(self):
        super().setUp()
        self.ts = self.write_rule_text("ts")
        self.structure = self.write_rule_text("project-structure")
        self.rules = [
            rule("ts", ["**/*.ts"], self.ts),
            rule("project-structure", ["**/_support/**"], self.structure, readers=True),
        ]

    def test_an_agent_carrying_the_pack_is_never_gated_for_it(self):
        manifest = self.write_manifest(self.rules, agents={"ts-backend-developer": ["ts"]})
        packed = {"agent_id": "a1", "agent_type": "ts-backend-developer"}
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts", **packed), manifest=manifest))
        # An agent that does not carry it still is.
        other = {"agent_id": "a2", "agent_type": "backend-developer"}
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts", **other),
                                                 manifest=manifest)),
                         reason_for([self.ts]))

    def test_a_rule_excluded_by_the_agents_framework_is_never_gated(self):
        # `*.service.ts` belongs to Angular and to NestJS alike. An agent that
        # carries one framework's pack must never be held on the other's rule:
        # `exclusive-with` is the rule file saying so, per framework, once.
        angular = self.write_rule_text("angular-patterns")
        nest = self.write_rule_text("nestjs-patterns")
        manifest = self.write_manifest(
            [rule("angular-patterns", ["**/*.service.ts"], angular,
                  exclusive_with=["nestjs-patterns"]),
             rule("nestjs-patterns", ["**/*.service.ts"], nest,
                  exclusive_with=["angular-patterns"])],
            agents={"ts-backend-developer": ["nestjs-patterns"],
                    "angular-developer": ["angular-patterns"]})
        backend = {"agent_id": "a1", "agent_type": "ts-backend-developer"}
        self.assertAllowed(
            self.run_hook(claude_write("/repo/src/users/users.service.ts", **backend),
                          manifest=manifest),
            "a NestJS agent must not be held on the Angular rule")
        angular_agent = {"agent_id": "a2", "agent_type": "angular-developer"}
        self.assertAllowed(
            self.run_hook(claude_write("/repo/src/app/user.service.ts", **angular_agent),
                          manifest=manifest),
            "an Angular agent must not be held on the NestJS rule")
        # An unpacked caller keeps today's behavior: both rules are pending,
        # which is what native path-scoping used to load.
        reason = self.held(self.run_hook(
            claude_write("/repo/src/users/users.service.ts",
                         agent_id="a3", agent_type="general-purpose"),
            manifest=manifest))
        for text in (angular, nest):
            self.assertIn(text.name, reason)

    def test_a_grok_subagent_is_identified_by_subagent_type(self):
        # Grok carries the child's roster name as `subagentType` only — no
        # snake alias, no `agent_type` (recorded from a 1.0.34 child payload).
        manifest = self.write_manifest(self.rules, agents={"ts-backend-developer": ["ts"]})
        packed = {"session": "child-1", "subagentType": "ts-backend-developer"}
        self.assertAllowed(self.run_hook(grok_write("/repo/a.ts", **packed), manifest=manifest))
        other = {"session": "child-2", "subagentType": "backend-developer"}
        self.assertEqual(self.held(self.run_hook(grok_write("/repo/a.ts", **other),
                                                 manifest=manifest)),
                         reason_for([self.ts]))

    def test_a_read_only_agent_is_never_denied_a_write(self):
        manifest = self.write_manifest(self.rules, read_only_agents=["review-code"])
        reviewer = {"agent_id": "r1", "agent_type": "review-code"}
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts", **reviewer),
                                         manifest=manifest))

    def test_a_read_only_agent_is_gated_on_the_first_read_of_a_readers_rule(self):
        manifest = self.write_manifest(self.rules, read_only_agents=["review-code"])
        reviewer = {"agent_id": "r1", "agent_type": "review-code"}
        self.assertEqual(
            self.held(self.run_hook(claude_read("/repo/_support/plan/x.md", **reviewer),
                                    manifest=manifest)),
            reason_for([self.structure]))
        self.assertAllowed(self.run_hook(observed_read(str(self.structure), **reviewer),
                                         manifest=manifest))
        self.assertAllowed(self.run_hook(claude_read("/repo/_support/plan/x.md", **reviewer),
                                         manifest=manifest))

    def test_an_agent_with_a_type_but_no_id_keeps_its_own_state(self):
        # Codex names the agent without always giving it an id; keying such a
        # call as "main" would let one agent's read release another's gate.
        manifest = self.write_manifest(self.rules)
        writer = {"agent_type": "backend-developer"}
        other = {"agent_type": "database-specialist"}
        self.assertNotEqual("", self.held(self.run_hook(claude_write("/repo/a.ts", **writer),
                                                        manifest=manifest)))
        self.assertAllowed(self.run_hook(observed_read(str(self.ts), **writer),
                                         manifest=manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts", **writer),
                                         manifest=manifest))
        self.assertNotEqual("", self.held(self.run_hook(claude_write("/repo/a.ts", **other),
                                                        manifest=manifest, env=WIDE_WINDOW)))
        self.assertNotEqual("", self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                        manifest=manifest, env=WIDE_WINDOW)))

    def test_reading_the_rule_file_itself_is_never_denied(self):
        # A `readers` rule whose globs also match its own deployed text (any
        # `**/*.md` rule) would otherwise deny the exact read it demands.
        text = self.write_rule_text("docs")
        manifest = self.write_manifest(
            [rule("docs", ["**/*.md"], text, readers=True)],
            read_only_agents=["review-code"])
        reviewer = {"agent_id": "r1", "agent_type": "review-code"}
        self.assertNotEqual("", self.held(self.run_hook(
            claude_read("/repo/notes.md", **reviewer), manifest=manifest)))
        self.assertAllowed(self.run_hook(claude_read(str(text), **reviewer),
                                         manifest=manifest),
                           "the gate must not deny the read it just asked for")
        self.assertAllowed(self.run_hook(observed_read(str(text), **reviewer),
                                         manifest=manifest))
        self.assertAllowed(self.run_hook(claude_read("/repo/notes.md", **reviewer),
                                         manifest=manifest))

    def test_a_read_only_agent_is_not_gated_on_a_rule_that_is_not_for_readers(self):
        manifest = self.write_manifest(self.rules, read_only_agents=["review-code"])
        reviewer = {"agent_id": "r1", "agent_type": "review-code"}
        self.assertAllowed(self.run_hook(claude_read("/repo/a.ts", **reviewer),
                                         manifest=manifest))

    def test_nobody_else_is_gated_on_a_read(self):
        manifest = self.write_manifest(self.rules, read_only_agents=["review-code"])
        self.assertAllowed(self.run_hook(claude_read("/repo/_support/plan/x.md"),
                                         manifest=manifest))
        writer = {"agent_id": "w1", "agent_type": "backend-developer"}
        self.assertAllowed(self.run_hook(claude_read("/repo/_support/plan/x.md", **writer),
                                         manifest=manifest))


# --- session lifecycle ---------------------------------------------------

class SessionStateTests(HookCase):
    def setUp(self):
        super().setUp()
        self.text = self.write_rule_text("ts")
        self.manifest = self.write_manifest([rule("ts", ["**/*.ts"], self.text)])
        self.sub = {"agent_id": "a1", "agent_type": "probe-writer"}

    def arm_and_release(self, session):
        self.held(self.run_hook(claude_write("/repo/a.ts", session=session),
                                manifest=self.manifest))
        self.assertAllowed(self.run_hook(observed_read(str(self.text), session=session),
                                         manifest=self.manifest))
        self.assertAllowed(self.run_hook(
            observed_read(str(self.text), session=session, **self.sub),
            manifest=self.manifest))

    def test_compaction_and_clear_re_arm_the_gate_for_the_session_and_its_subagents(self):
        for source in ("compact", "clear"):
            with self.subTest(source=source):
                self.setUp()
                self.arm_and_release("s1")
                self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"),
                                                 manifest=self.manifest))
                self.assertAllowed(self.run_hook(
                    {"hook_event_name": "SessionStart", "session_id": "s1", "source": source},
                    manifest=self.manifest))
                self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts"),
                                                         manifest=self.manifest)),
                                 reason_for([self.text]))
                self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.ts", **self.sub),
                                                         manifest=self.manifest)),
                                 reason_for([self.text]))

    def test_another_sessions_state_survives_a_compaction(self):
        self.arm_and_release("s1")
        self.arm_and_release("s2")
        self.assertAllowed(self.run_hook(
            {"hook_event_name": "SessionStart", "session_id": "s1", "source": "compact"},
            manifest=self.manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts", session="s2"),
                                         manifest=self.manifest))

    def test_a_startup_session_start_clears_nothing(self):
        self.arm_and_release("s1")
        self.assertAllowed(self.run_hook(
            {"hook_event_name": "SessionStart", "session_id": "s1", "source": "startup"},
            manifest=self.manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=self.manifest))


class MarkerNamingTests(HookCase):
    def test_rule_names_that_share_a_prefix_gate_independently(self):
        short = self.write_rule_text("communication-format")
        long = self.write_rule_text("communication-format-mechanics")
        manifest = self.write_manifest([
            rule("communication-format", ["**/*.ts"], short),
            rule("communication-format-mechanics", ["**/*.tsx"], long),
        ])
        self.assertAllowed(self.run_hook(observed_read(str(short)), manifest=manifest))
        self.assertAllowed(self.run_hook(claude_write("/repo/a.ts"), manifest=manifest))
        self.assertEqual(self.held(self.run_hook(claude_write("/repo/a.tsx"),
                                                 manifest=manifest)),
                         reason_for([long]), "the sibling rule must still be held")


# --- glob matching -------------------------------------------------------

class GlobTests(HookCase):
    def gate(self, globs, path, *, session="s1"):
        manifest = self.write_manifest([rule("r", globs, self.write_rule_text("r"))],
                                       name=f"m-{abs(hash((tuple(globs), path)))}.json")
        return self.run_hook(claude_write(path, session=session), manifest=manifest)

    def test_globs_match_by_segment_not_by_substring(self):
        self.assertAllowed(self.gate(["**/_support/**"], "/repo/src/customer_support/a.md"))
        self.assertNotEqual("", self.held(self.gate(["**/_support/**"],
                                                    "/repo/_support/plan/a.md", session="s2")))

    def test_brace_groups_basenames_and_double_star(self):
        held = [
            (["**/*.{ts,tsx}"], "/repo/src/a.tsx"),
            (["docker-compose*.{yml,yaml}"], "/repo/docker-compose.override.yaml"),
            ([".github/workflows/**/*.{yml,yaml}"], "/repo/.github/workflows/nested/ci.yml"),
            (["turbo.json"], "/repo/apps/web/turbo.json"),
            (["**/*.{java,kt,kts}"], "/repo/src/Main.kt"),
        ]
        for index, (globs, path) in enumerate(held):
            with self.subTest(globs=globs, path=path):
                self.assertNotEqual("", self.held(self.gate(globs, path, session=f"h{index}")))
        allowed = [
            (["**/*.{ts,tsx}"], "/repo/src/a.tsxx"),
            (["Dockerfile*"], "/repo/NotADockerfile.md"),
            ([".github/workflows/**/*.yml"], "/repo/.github/ci.yml"),
        ]
        for index, (globs, path) in enumerate(allowed):
            with self.subTest(globs=globs, path=path):
                self.assertAllowed(self.gate(globs, path, session=f"a{index}"))

    def test_brackets_in_a_glob_are_literal_characters(self):
        # No rule in harness/rule-manifest.json uses a character class, and a
        # Next.js route segment (`app/[locale]/**`) means the directory it looks
        # like — not "one of l, o, c, a, e".
        self.assertNotEqual("", self.held(self.gate(["app/[locale]/**/*.tsx"],
                                                    "/repo/app/[locale]/page.tsx")))
        self.assertAllowed(self.gate(["app/[locale]/**/*.tsx"], "/repo/app/l/page.tsx",
                                     session="s2"))
        for index, glob in enumerate((r"[a\]", "[z-a]", "[[a]", "[!", "[", "[]")):
            with self.subTest(glob=glob):
                result = self.gate([glob], "/repo/a.ts", session=f"g{index}")
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(result.stderr.strip(), "",
                                 "a warning on stderr is output the harness will show")

    def test_a_brace_bomb_makes_its_rule_undeliverable_instead_of_expanding(self):
        started = time.monotonic()
        result = self.gate(["{a,b}" * 18 + ".ts"], "/repo/ababab.ts")
        self.assertAllowed(result, "a rule nobody can match is not worth 2^18 strings")
        self.assertLess(time.monotonic() - started, 2)
        # A rule with a sane brace group still gates.
        self.assertNotEqual("", self.held(self.gate(["**/*.{ts,tsx,mts}"], "/repo/a.mts",
                                                    session="s-ok")))

    def test_a_pathological_glob_does_not_hang(self):
        started = time.monotonic()
        self.gate(["**a**a**a**a**a**a**a**b"], "/repo/" + "a" * 60 + ".ts")
        self.assertLess(time.monotonic() - started, 5)


class ScopeOfMatchingTests(HookCase):
    """WHAT a glob is matched against: the absolute realpath of the target.

    Three separate false skips came out of deriving a *relative* path to match
    (`basename(cwd)/…`, then `basename(root)/…`): a `cd` into a subdirectory, a
    `/tmp` vs `/private/tmp` spelling, and a Codex payload with an empty `cwd`.
    The absolute path already contains the repo's folder name, so the whole
    derivation is gone — and with it that class of defect. What remains is a
    matrix over the shapes that produced those bugs, each case judged by the
    oracle below rather than by a hand-written expectation.
    """

    HARNESS_ROOTS = ("/.claude", "/.codex", "/.grok", "/.agents", "/.pi",
                     "/.config/opencode")
    DERIVED = ("node_modules", "dist", ".next", ".turbo", ".venv",
               "__pycache__", "coverage")

    def setUp(self):
        super().setUp()
        self.specs = self.write_rule_text("session-capture")
        self.python = self.write_rule_text("python-standards")
        self.rules = {
            "session-capture": ["**/_support/**", "**/*-specs/**",
                                "**/_support/sessions/**", "**/*-specs/sessions/**"],
            "python-standards": ["**/*.py"],
        }
        self.manifest = self.write_manifest([
            rule("session-capture", self.rules["session-capture"], self.specs,
                 readers=True),
            rule("python-standards", self.rules["python-standards"], self.python),
        ])

    # -- the oracle -------------------------------------------------------

    def expected(self, absolute):
        """Held iff the absolute realpath matches a glob and nothing exempts it.

        Deliberately re-derived from the rule text rather than from the hook:
        an oracle that shares the hook's code proves only that it agrees with
        itself.
        """
        resolved = os.path.realpath(absolute)
        parts = os.path.dirname(resolved).split(os.sep)
        if any(part in self.DERIVED for part in parts):
            return []
        exempt = [os.path.realpath(self.MATRIX_HOME + suffix)
                  for suffix in self.HARNESS_ROOTS]
        exempt += [os.path.realpath(root) for root in
                   ("/tmp", "/private/tmp", "/var/folders", "/private/var/folders",
                    str(self.state_root))]
        if any(resolved == root or resolved.startswith(root.rstrip("/") + "/")
               for root in exempt):
            return []
        held = []
        for name, path in (("session-capture", self.specs),
                           ("python-standards", self.python)):
            if self.matches(resolved, self.rules[name]):
                held.append(path)
        return held

    @staticmethod
    def matches(path, globs):
        """A plain fnmatch-per-segment oracle, independent of the hook's matcher."""
        import fnmatch
        parts = [part for part in path.split("/") if part]
        for pattern in globs:
            for expanded in ScopeOfMatchingTests.braces(pattern):
                segments = [s for s in expanded.split("/") if s]
                if ScopeOfMatchingTests.walk(segments, parts):
                    return True
        return False

    @staticmethod
    def braces(pattern):
        import re as regex
        found = regex.search(r"\{([^{}]*)\}", pattern)
        if not found:
            return [pattern]
        out = []
        for option in found.group(1).split(","):
            out.extend(ScopeOfMatchingTests.braces(
                pattern[:found.start()] + option + pattern[found.end():]))
        return out

    @staticmethod
    def walk(segments, parts):
        import fnmatch
        if not segments:
            return not parts
        if segments[0] == "**":
            for index in range(len(parts) + 1):
                if ScopeOfMatchingTests.walk(segments[1:], parts[index:]):
                    return True
            return False
        if not parts or not fnmatch.fnmatchcase(parts[0], segments[0]):
            return False
        return ScopeOfMatchingTests.walk(segments[1:], parts[1:])

    # -- the matrix -------------------------------------------------------

    # HOME and the projects live OUTSIDE the scratch tree on purpose: the
    # scratch root is under /tmp, which is itself an exempt temp root, and a
    # matrix rooted there would assert "allowed" for reasons unrelated to it.
    # Nothing here needs to exist on disk — the gate stats no target.
    MATRIX_HOME = "/Users/hive-test"

    def locations(self):
        """One path per LOCATION kind of the matrix."""
        normal = Path("/ws/projects/ark-specs")
        temp_project = Path(tempfile.mkdtemp(prefix="hive-tempproj-", dir="/tmp"))
        self.addCleanup(shutil.rmtree, temp_project, ignore_errors=True)
        (temp_project / "_support").mkdir()
        return {
            "normal project": (normal, "domains/a.py"),
            "project under a temp root": (temp_project, "_support/a.py"),
            "harness root": (Path(self.MATRIX_HOME) / ".claude",
                             "projects/-Users-x-ark-specs/memory/x.py"),
            "derived dir below a project":
                (normal, "apps/web/node_modules/pkg/_support/a.py"),
            "derived-named ancestor": (Path("/ws/dist/inner-specs"), "domains/a.py"),
        }

    def test_every_cwd_and_path_shape_agrees_with_the_absolute_path_oracle(self):
        # cwd shape x target spelling x location x declared-root field. Each of
        # the three historical false skips is one cell of this table; enumerating
        # the class is what keeps the fourth from being found in production.
        cases = 0
        for location, (project, inside) in self.locations().items():
            absolute = project / inside
            spellings = {"absolute": str(absolute)}
            if str(project).startswith("/tmp/"):
                # The same file, spelled the way macOS also accepts.
                spellings["symlinked-spelling"] = "/private" + str(absolute)
            spellings["relative"] = inside
            cwds = {"project root": str(project),
                    "subdirectory": str(absolute.parent),
                    "empty": "",
                    "absent": None}
            roots = {"none": {}, "workspaceRoot": {"workspaceRoot": str(project)},
                     "workspace_roots[]": {"workspace_roots": [str(project)]}}
            for spelling, target in spellings.items():
                for cwd_label, cwd in cwds.items():
                    for root_label, extra in roots.items():
                        if spelling == "relative" and cwd in ("", None) and not extra:
                            continue  # nothing can make this absolute; covered below
                        cases += 1
                        payload = claude_write(target, session=f"m{cases}", **extra)
                        if cwd is None:
                            payload.pop("cwd", None)
                        else:
                            payload["cwd"] = cwd
                        # A relative target resolves against cwd first, then the
                        # declared root — that is the ONLY use of either.
                        base = cwd or (extra.get("workspaceRoot")
                                       or (extra.get("workspace_roots") or [""])[0])
                        resolved = (target if os.path.isabs(target)
                                    else os.path.join(base, target))
                        held = self.expected(resolved)
                        with self.subTest(location=location, spelling=spelling,
                                          cwd=cwd_label, root=root_label):
                            result = self.run_hook(payload, manifest=self.manifest,
                                                   env={"HOME": self.MATRIX_HOME})
                            if held:
                                self.assertEqual(self.held(result), reason_for(held))
                            else:
                                self.assertAllowed(result)
        self.assertGreater(cases, 60, "the matrix must actually enumerate the class")

    def test_the_codex_shape_that_skipped_every_root_anchored_glob(self):
        # Verbatim from the re-verification: `cwd` empty, a RELATIVE patch path,
        # the root only in `workspace_roots`. The relative path used to be
        # matched bare (`domains/a.md`), so every glob anchored on the repo's own
        # folder name — `**/*-specs/**`, `**/*-infra/**` — was skipped.
        project = "/ws/projects/ark-specs"
        payload = codex_patch(patch_body("*** Add File: domains/a.md"),
                              cwd="", workspace_roots=[project])
        self.assertEqual(self.held(self.run_hook(payload, manifest=self.manifest,
                                                 env={"HOME": self.MATRIX_HOME})),
                         reason_for([self.specs], harness="codex"))
        # The same patch with the root only in `cwd` was never the broken case,
        # and stays held.
        self.assertEqual(self.held(self.run_hook(
            codex_patch(patch_body("*** Add File: domains/a.md"), cwd=project,
                        session="c2"),
            manifest=self.manifest, env={"HOME": self.MATRIX_HOME})),
            reason_for([self.specs], harness="codex"))

    def test_a_file_named_like_a_derived_directory_is_still_source(self):
        # The exemption is about DIRECTORIES. Widening it to the basename would
        # skip every rule for a file someone named `dist` or `coverage` — both
        # are ordinary names for a script or a note — and no other case in this
        # suite distinguishes the two.
        project = "/ws/projects/ark-specs"
        for index, name in enumerate(("dist", "coverage", "build.dist")):
            with self.subTest(name=name):
                self.assertEqual(
                    self.held(self.run_hook(
                        claude_write(f"{project}/_support/{name}", session=f"n{index}"),
                        manifest=self.manifest, env={"HOME": self.MATRIX_HOME})),
                    reason_for([self.specs]),
                    f"a file called {name} is source, not a generated tree")
        # …while the same name as a directory component still exempts.
        self.assertAllowed(self.run_hook(
            claude_write(f"{project}/_support/dist/a.py", session="dir"),
            manifest=self.manifest, env={"HOME": self.MATRIX_HOME}))

    def test_a_deleted_process_cwd_never_raises_out_of_the_hook(self):
        # `os.getcwd()` raises FileNotFoundError when the directory the process
        # started in has been removed; the gate must not fail open on it, and
        # must not crash.
        doomed = self.root / "doomed"
        doomed.mkdir()
        payload = claude_write("_support/sessions/n.py", session="gone")
        payload["cwd"] = ""
        result = subprocess.run(
            [sys.executable, "-c",
             "import os, shutil, subprocess, sys, json\n"
             "os.chdir(sys.argv[2]); shutil.rmtree(sys.argv[2])\n"
             "print(subprocess.run([sys.executable, sys.argv[1]], input=sys.stdin.read(),"
             " text=True, capture_output=True).stdout, end='')",
             str(HOOK), str(doomed)],
            input=json.dumps(payload), text=True, capture_output=True,
            env={"PATH": os.environ.get("PATH", ""), "HOME": str(self.home),
                 "TMPDIR": str(self.state_root),
                 "HIVE_RULE_MANIFEST": str(self.manifest)}, check=False)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("session-capture", result.stdout,
                      "a gone cwd must not turn the gate off")


class HarnessDetectionTests(HookCase):
    """Which reference root a call gets, and how the hook decides."""

    def setUp(self):
        super().setUp()
        self.claude_text = self.write_rule_text("claude-copy")
        self.agents_text = self.write_rule_text("agents-copy")
        self.manifest = self.write_manifest([{
            "name": "ts",
            "source": "global/rules-situational/ts.md",
            "globs": ["**/*.ts"],
            "exclusive_with": [],
            "always_on": False,
            "references": {"claude": str(self.claude_text), "agents": str(self.agents_text)},
            "readers": False,
        }])

    def test_the_harness_env_var_decides_before_any_payload_sniffing(self):
        # Each harness sets it on the hook's own command line; sniffing is the
        # fallback, not the contract.
        cases = {"claude": self.claude_text, "codex": self.agents_text,
                 "pi": self.agents_text}
        for harness, expected in cases.items():
            with self.subTest(harness=harness):
                result = self.run_hook(claude_write("/repo/a.ts", session=harness),
                                       manifest=self.manifest,
                                       env={"HIVE_HARNESS": harness})
                self.assertEqual(self.held(result),
                                 reason_for([expected], harness=harness))

    def test_an_unknown_or_absent_value_falls_back_to_sniffing(self):
        for value in ("", "nonsense"):
            with self.subTest(value=value):
                result = self.run_hook(codex_patch(patch_body("*** Add File: /repo/a.ts"),
                                                   session=f"c{value}"),
                                       manifest=self.manifest,
                                       env={"HIVE_HARNESS": value} if value else None)
                self.assertEqual(self.held(result),
                                 reason_for([self.agents_text], harness="codex"))

    def test_grok_is_sniffed_even_when_the_env_says_claude(self):
        # Grok arrives through the Claude settings block, so it inherits
        # HIVE_HARNESS=claude — its camelCase keys are what identify it, and
        # what it must keep is the 260-char reason budget.
        long_reference = self.reference_of_length(240, name="long")
        manifest = self.write_manifest([rule("long", ["**/*.ts"], long_reference)],
                                       name="long.json")
        self.assertAllowed(self.run_hook(grok_write("src/a.ts"), manifest=manifest,
                                         env={"HIVE_HARNESS": "claude"}),
                           "Grok must not inherit Claude's wide reason budget")


# --- cost ----------------------------------------------------------------

class PerformanceTests(HookCase):
    def forty_rules(self):
        text = self.write_rule_text("ts")
        rules = [rule(f"r{index}", [f"**/*.e{index}"], text) for index in range(39)]
        rules.append(rule("ts", ["**/*.ts"], text))
        return self.write_manifest(rules)

    def test_a_huge_shell_command_is_not_parsed_once_per_rule(self):
        # Measured before the fix: 38 gateable rules x a 1 MB command = 5.4 s,
        # on EVERY terminal call, because each rule re-lexed the whole thing.
        manifest = self.forty_rules()
        heredoc = "cat <<'EOF'\n" + ("x" * 1024 + "\n") * 1024 + "EOF"
        self.assertGreater(len(heredoc), 1_000_000)
        for label, payload in (("post", observed_bash(heredoc)),
                               ("pre", claude_bash(heredoc))):
            with self.subTest(event=label):
                started = time.monotonic()
                result = self.run_hook(payload, manifest=manifest)
                elapsed = time.monotonic() - started
                self.assertAllowed(result)
                self.assertLess(elapsed, 1.5, f"{elapsed * 1000:.0f} ms on a 1 MB command")

    def test_a_huge_patch_shaped_command_is_scanned_once(self):
        manifest = self.forty_rules()
        noise = "\n".join(f"line {index} *** Update File: src/a.ts" for index in range(11000))
        command = f"git commit -m 'summary\n{noise}'"
        self.assertGreater(len(command), 300_000)
        started = time.monotonic()
        result = self.run_hook(claude_bash(command), manifest=manifest)
        self.assertAllowed(result, "a commit message is not a patch")
        self.assertLess(time.monotonic() - started, 1.5)

    def test_a_five_hundred_rule_manifest_stays_cheap(self):
        text = self.write_rule_text("ts")
        rules = [rule(f"r{index}", [f"**/*.ext{index}", "**/*.md"], text)
                 for index in range(499)]
        rules.append(rule("ts", ["**/*.ts"], text))
        manifest = self.write_manifest(rules)
        started = time.monotonic()
        result = self.run_hook(claude_write("/repo/a.ts"), manifest=manifest)
        elapsed = time.monotonic() - started
        self.assertEqual(self.held(result), reason_for([text]))
        # The budget is 100 ms; the bound is generous so a loaded machine or a
        # cold interpreter start never makes this flake.
        self.assertLess(elapsed, 2.0, f"{elapsed * 1000:.0f} ms per invocation")


# --- wiring --------------------------------------------------------------

def load_hook_module():
    import importlib.util

    spec = importlib.util.spec_from_file_location("rule_delivery_under_test", HOOK)
    module = importlib.util.module_from_spec(spec)
    bytecode = sys.dont_write_bytecode
    sys.dont_write_bytecode = True
    try:
        spec.loader.exec_module(module)
    finally:
        sys.dont_write_bytecode = bytecode
    return module


class WiringTests(unittest.TestCase):
    def matcher(self, path, event="PreToolUse"):
        config = json.loads(Path(path).read_text(encoding="utf-8"))
        return set(config["hooks"][event][0]["matcher"].split("|"))

    def command(self, path, event="PreToolUse"):
        config = json.loads(Path(path).read_text(encoding="utf-8"))
        return config["hooks"][event][0]["hooks"][0]["command"]

    def test_the_pre_tool_use_matcher_carries_every_gated_tool(self):
        module = load_hook_module()
        matcher = self.matcher(HOOK_DIR / "settings-config.json")
        # Terminal tools stay on PreToolUse for one reason only: a patch written
        # through the shell authors files and must be gated like `apply_patch`.
        expected = (module.READ_TOOLS | module.WRITE_TOOLS | module.TERMINAL_TOOLS
                    | {module.PATCH_TOOL})
        self.assertEqual(expected - matcher, set(),
                         "a tool the hook handles but the matcher never routes is dead code")

    def test_observation_is_wired_on_completion(self):
        matcher = self.matcher(HOOK_DIR / "settings-config.json", event="PostToolUse")
        module = load_hook_module()
        self.assertEqual((module.READ_TOOLS | module.TERMINAL_TOOLS) - matcher, set())
        self.assertEqual(matcher & module.WRITE_TOOLS, set(),
                         "a write is never an observation")

    def test_the_claude_matcher_covers_the_repos_file_edit_surface(self):
        # executor-dispatch-gate is the established file-edit surface; a tool it
        # gates and this hook does not is a write that slips past the rule.
        gate = self.matcher(HOOK_DIR.parent / "executor-dispatch-gate/settings-config.json")
        matcher = self.matcher(HOOK_DIR / "settings-config.json")
        self.assertEqual((gate - {"Agent", "Task"}) - matcher, set())
        module = load_hook_module()
        self.assertEqual((gate - {"Agent", "Task"}) - module.WRITE_TOOLS, set())

    def test_the_session_start_block_still_clears_the_state(self):
        matcher = self.matcher(HOOK_DIR / "settings-config.json", event="SessionStart")
        self.assertEqual(matcher, {"compact", "clear"})

    def test_every_wired_command_declares_its_harness(self):
        for config, event in ((HOOK_DIR / "settings-config.json", "PreToolUse"),
                              (HOOK_DIR / "settings-config.json", "PostToolUse"),
                              (HOOK_DIR / "settings-config.json", "SessionStart"),
                              (HOOK_DIR / "codex-hooks.json", "PreToolUse"),
                              (HOOK_DIR / "codex-hooks.json", "PostToolUse")):
            with self.subTest(config=config.name, event=event):
                command = self.command(config, event)
                expected = "claude" if config.name == "settings-config.json" else "codex"
                self.assertIn(f"HIVE_HARNESS={expected}", command)

    def test_the_codex_matcher_carries_the_patch_tool_and_the_shell(self):
        pre = self.matcher(HOOK_DIR / "codex-hooks.json")
        self.assertIn("apply_patch", pre)
        self.assertIn("Bash", pre, "a patch written through the shell is still a write")
        post = self.matcher(HOOK_DIR / "codex-hooks.json", event="PostToolUse")
        self.assertIn("Bash", post, "reads reach Codex as shell calls; without it the "
                                    "gate has no release")


# --- armed against the real manifest -------------------------------------

REPO_ROOT = HOOK_DIR.parent.parent.parent
REAL_MANIFEST = REPO_ROOT / "harness/rule-manifest.json"
INJECTED_REFERENCES = REPO_ROOT / "harness/agents-skills"


class RealManifestTests(HookCase):
    """The one place the repo's OWN manifest is the input.

    Every other test pins a fixture. This one asks whether the gate is armed
    at all: while the glob-scoped rules lived under `global/rules/`, the hook
    skipped every one of them by design and shipped gating nothing.
    """

    def stage_references(self, manifest):
        """Put each injected reference where the manifest says claude reads it.

        The gate refuses to ask for a read it cannot point at, so a rule is
        only gateable once its reference file exists under this HOME.
        """
        staged = 0
        for entry in manifest["rules"]:
            raw = (entry.get("references") or {}).get("claude")
            if not raw:
                continue
            # ~/.claude/skills/<skill>/references/<rule>.md — the tail after
            # `skills/` is exactly the path inside the generated tree.
            parts = Path(raw).parts
            source = INJECTED_REFERENCES.joinpath(*parts[parts.index("skills") + 1:])
            if not source.is_file():
                continue
            target = self.home / Path(*parts[1:])
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
            staged += 1
        return staged

    def test_the_real_manifest_arms_the_gate_for_a_glob_scoped_rule(self):
        manifest = json.loads(REAL_MANIFEST.read_text(encoding="utf-8"))
        armed = [
            entry for entry in manifest["rules"]
            if entry["globs"] and not entry["always_on"]
            and entry["source"].startswith("global/rules-situational/")
        ]
        self.assertTrue(
            armed,
            "no glob-scoped rule lives in the store the hook owns — the gate "
            "ships disarmed",
        )
        self.assertTrue(self.stage_references(manifest), "no reference staged")

        result = self.run_hook(
            claude_write("/repo/src/app.ts"), manifest=REAL_MANIFEST, env=NO_WINDOW,
        )

        self.assertIn(
            "typescript-standards.md", self.held(result),
            "a TypeScript write was not held for the TypeScript rule",
        )


if __name__ == "__main__":
    unittest.main()
