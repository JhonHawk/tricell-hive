#!/usr/bin/env python3
"""Tests for the isolated skill evaluation runner."""

import json
import sys
import tempfile
import time
import unittest
from unittest.mock import patch
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(Path(__file__).parent))
from runner import EvalError, EvalRunner, _normalise_trace, load_manifest  # noqa: E402


class IsolatedEnvironmentTests(unittest.TestCase):
    def test_shared_skill_root_is_readable_and_contained_for_both_harnesses(self):
        for harness in ("codex", "pi"):
            with self.subTest(harness=harness), tempfile.TemporaryDirectory(prefix="hive-eval-env-") as tmp:
                run_root = Path(tmp)
                runner = EvalRunner(repo_root=ROOT, harness=harness, model="test-model", timeout=5)
                staged, _, roots = runner._stage_skills(run_root)
                with patch.object(runner, "_link_auth", return_value={"status": "mocked"}) as link_auth:
                    env, isolation = runner._environment(run_root, roots)

                link_auth.assert_called_once_with(env, roots)
                shared_root = Path(env["HOME"]) / ".agents" / "skills"
                self.assertTrue(shared_root.is_symlink())
                skill_file = shared_root / "flow-core" / "SKILL.md"
                self.assertTrue(skill_file.is_file())
                self.assertTrue(skill_file.read_text(encoding="utf-8"))
                self.assertEqual(skill_file.resolve(), (staged / "flow-core" / "SKILL.md").resolve())
                self.assertTrue(skill_file.resolve().is_relative_to(run_root.resolve()))
                self.assertEqual(isolation["auth"], {"status": "mocked"})


class ManifestTests(unittest.TestCase):
    def test_accepts_case_contract_and_normalizes_short_file_assertions(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-manifest-") as tmp:
            path = Path(tmp) / "cases.json"
            path.write_text(
                json.dumps(
                    {
                        "version": 1,
                        "cases": [
                            {
                                "id": "positive",
                                "skill": "language-rules",
                                "category": "activation",
                                "prompt": "Read the Python guidance.",
                                "fixtures": {"notes.txt": "fixture"},
                                "protected_files": ["notes.txt"],
                                "expected_activation": True,
                                "required_reads": [
                                    "language-rules/references/python-standards.md"
                                ],
                                "allowed_writes": [],
                                "required_files": ["notes.txt"],
                                "output_contains": ["Python"],
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )
            manifest = load_manifest(path)
            self.assertEqual(manifest["version"], 1)
            self.assertEqual(manifest["cases"][0]["required_files"][0], "notes.txt")

    def test_rejects_manifest_path_traversal(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-manifest-") as tmp:
            path = Path(tmp) / "cases.json"
            path.write_text(
                json.dumps(
                    {
                        "version": 1,
                        "cases": [
                            {
                                "id": "unsafe",
                                "skill": "language-rules",
                                "category": "negative",
                                "prompt": "No-op",
                                "fixtures": {"../outside.txt": "escape"},
                                "protected_files": [],
                                "expected_activation": False,
                                "required_reads": [],
                                "allowed_writes": [],
                                "required_files": [],
                                "output_contains": [],
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )
            with self.assertRaises(EvalError):
                load_manifest(path)


class NativeCommandTests(unittest.TestCase):
    def loads(self, command, exit_code=0):
        event = {"type": "item.completed", "item": {"type": "command_execution", "command": command, "status": "completed", "exit_code": exit_code}}
        return _normalise_trace(json.dumps(event), "", "language-rules")[2]

    def test_compound_command_cannot_claim_a_read_from_echo(self):
        self.assertEqual(self.loads("cat input.txt; echo language-rules/SKILL.md"), [])

    def test_failed_command_cannot_prove_read(self):
        self.assertEqual(self.loads("cat language-rules/SKILL.md", 1), [])

    def test_shell_wrapped_sed_proves_read(self):
        self.assertEqual(self.loads("/bin/zsh -lc \"sed -n '1,240p' /tmp/language-rules/SKILL.md\""), ["/tmp/language-rules/SKILL.md"])

    def test_pi_turn_end_does_not_complete_run(self):
        event = {"type": "turn_end", "message": {"role": "assistant", "stopReason": "toolUse"}, "toolResults": []}
        self.assertFalse(_normalise_trace(json.dumps(event), "", "language-rules")[3])

    def test_plain_cat_proves_read(self):
        self.assertEqual(self.loads("cat language-rules/SKILL.md"), ["language-rules/SKILL.md"])

    def test_top_level_codex_command_execution_proves_successful_read(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"sed -n '1,240p' staged-skills/language-rules/SKILL.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(
            _normalise_trace(json.dumps(event), "", "language-rules")[2],
            ["staged-skills/language-rules/SKILL.md"],
        )

    def test_top_level_failed_command_cannot_prove_read(self):
        event = {
            "type": "command_execution",
            "command": "cat language-rules/SKILL.md",
            "status": "completed",
            "exit_code": 1,
        }
        self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], [])

    def test_aggregate_shell_command_does_not_prove_each_read(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"sed -n '1,20p' one.md && echo two.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], [])

    def test_all_successful_and_chain_proves_each_simple_read(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"sed -n '1,20p' one.md && cat two.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(
            _normalise_trace(json.dumps(event), "", "language-rules")[2],
            ["one.md", "two.md"],
        )

    def test_semicolon_chain_only_proves_last_simple_read(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"cat missing.md; sed -n '1,20p' last.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], ["last.md"])

    def test_semicolon_chain_with_nonread_segment_proves_nothing(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"cat first.md; git status --short; sed -n '1,20p' last.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], [])

    def test_newline_chain_only_proves_last_simple_read(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"cat first.md\nsed -n '1,20p' last.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], ["last.md"])

    def test_cat_help_and_version_flags_do_not_claim_operands(self):
        for command in ("cat --help language-rules/SKILL.md", "cat --version language-rules/SKILL.md"):
            event = {
                "type": "command_execution",
                "command": command,
                "status": "completed",
                "exit_code": 0,
            }
            with self.subTest(command=command):
                self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], [])

    def test_unquoted_shell_comment_cannot_claim_commented_skill_path(self):
        event = {
            "type": "command_execution",
            "command": "/bin/zsh -lc \"cat input.txt && cat input.txt # language-rules/SKILL.md\"",
            "status": "completed",
            "exit_code": 0,
        }
        self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], [])

    def test_output_only_commands_do_not_prove_a_read(self):
        for command in (
            "echo language-rules/SKILL.md",
            "rg -n language-rules/SKILL.md .",
        ):
            event = {
                "type": "command_execution",
                "command": command,
                "status": "completed",
                "exit_code": 0,
            }
            with self.subTest(command=command):
                self.assertEqual(_normalise_trace(json.dumps(event), "", "language-rules")[2], [])


class RunnerTests(unittest.TestCase):
    def _case(self, **overrides):
        case = {
            "id": "mock-success",
            "skill": "language-rules",
            "category": "activation",
            "prompt": "Read the Python guidance.",
            "fixtures": {"input.txt": "unchanged"},
            "protected_files": ["input.txt"],
            "expected_activation": True,
            "required_reads": [
                "language-rules/references/python-standards.md"
            ],
            "allowed_writes": ["result.txt"],
            "required_files": ["result.txt"],
            "file_contains": {"result.txt": ["ok"]},
            "output_contains": ["finished"],
        }
        case.update(overrides)
        return case

    def test_mock_run_requires_terminal_event_and_records_tool_and_file_evidence(self):
        script = (
            "import json, pathlib; "
            "print(json.dumps({'type':'tool_call.started','name':'read_file',"
            "'path':'language-rules/references/python-standards.md'})); "
            "print(json.dumps({'type':'tool_call.completed','name':'read_file',"
            "'path':'language-rules/references/python-standards.md','success':True})); "
            "print(json.dumps({'type':'tool_call.completed','name':'read_file',"
            "'path':'language-rules/SKILL.md','success':True})); "
            "pathlib.Path('result.txt').write_text('ok', encoding='utf-8'); "
            "print('finished'); "
            "print(json.dumps({'type':'turn.completed'}))"
        )

        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT,
                harness="codex",
                model="test-model",
                effort="high",
                timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script],
                stream_log=lambda _: None,
            )
            result = runner.run_case(self._case(), Path(tmp))

        self.assertEqual(result["outcome"]["status"], "passed")
        self.assertEqual(result["process"]["status"], "passed")
        self.assertEqual(result["activation"]["status"], "passed")
        self.assertTrue(result["evidence"]["skill_loads"])
        self.assertEqual(result["evidence"]["unexpected_writes"], [])
        self.assertEqual(result["evidence"]["protected_files"]["input.txt"]["status"], "passed")

    def test_fixture_has_committed_clean_baseline(self):
        script = (
            "import json, subprocess; "
            "subprocess.run(['git', 'rev-parse', '--verify', 'HEAD'], check=True, capture_output=True); "
            "assert not subprocess.check_output(['git', 'status', '--porcelain']); "
            "print(json.dumps({'type':'turn.completed'}))"
        )
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT, harness="codex", model="test-model", timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script], stream_log=lambda _: None,
            )
            result = runner.run_case(self._case(expected_activation=False, required_reads=[], required_files=[], output_contains=[], file_contains={}), Path(tmp))
        self.assertEqual(result["outcome"]["status"], "passed")

    def test_allowed_parent_directory_is_not_scope_failure_when_output_missing(self):
        script = "import json, pathlib; pathlib.Path('out').mkdir(); print(json.dumps({'type':'turn.completed'}))"
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT, harness="codex", model="test-model", timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script], stream_log=lambda _: None,
            )
            result = runner.run_case(self._case(expected_activation=False, required_reads=[], allowed_writes=['out/report.md'], required_files=['out/report.md'], output_contains=[], file_contains={}), Path(tmp))
        self.assertEqual(result["evidence"]["writes"]["status"], "passed")
        self.assertEqual(result["outcome"]["checks"]["required_files"]["out/report.md"]["status"], "failed")

    def test_allowed_parent_does_not_allow_replacing_existing_file(self):
        script = "import json, pathlib; pathlib.Path('out').unlink(); pathlib.Path('out').mkdir(); print(json.dumps({'type':'turn.completed'}))"
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT, harness="codex", model="test-model", timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script], stream_log=lambda _: None,
            )
            result = runner.run_case(self._case(fixtures={'out':'preserve'}, protected_files=[], expected_activation=False, required_reads=[], allowed_writes=['out/report.md'], required_files=[], output_contains=[], file_contains={}), Path(tmp))
        self.assertEqual(result["evidence"]["writes"]["status"], "failed")

    def test_missing_terminal_event_is_unknown_even_when_exit_is_zero(self):
        script = "print('finished')"
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT,
                harness="codex",
                model="test-model",
                effort="high",
                timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script],
                stream_log=lambda _: None,
            )
            result = runner.run_case(
                self._case(expected_activation=False, required_files=[], output_contains=[], file_contains={}),
                Path(tmp),
            )

        self.assertEqual(result["process"]["status"], "unknown")
        self.assertEqual(result["outcome"]["status"], "unknown")

    def test_reference_read_does_not_prove_skill_activation(self):
        script = (
            "import json; "
            "print(json.dumps({'type':'tool_execution_start','toolName':'read',"
            "'path':'flow-plan/references/plan-format.md'})); "
            "print(json.dumps({'type':'tool_execution_end','toolName':'read',"
            "'path':'flow-plan/references/plan-format.md','isError':False})); "
            "print(json.dumps({'type':'agent_end'}))"
        )
        case = self._case(
            skill="flow-plan",
            expected_activation=True,
            required_reads=["flow-plan/references/plan-format.md"],
            required_files=[],
            output_contains=[],
            file_contains={},
        )
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT, harness="pi", model="test-model", effort="high", timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script], stream_log=lambda _: None,
            )
            result = runner.run_case(case, Path(tmp))
        self.assertEqual(
            result["outcome"]["checks"]["required_reads"]["flow-plan/references/plan-format.md"]["status"],
            "passed",
        )
        self.assertEqual(result["activation"]["status"], "unknown")
        self.assertEqual(result["outcome"]["status"], "unknown")

    def test_timeout_kills_process_group_after_parent_exits(self):
        script = (
            "import subprocess, sys, time; "
            "subprocess.Popen([sys.executable, '-c', 'import time; time.sleep(2)'])"
        )
        case = self._case(expected_activation=False, required_files=[], output_contains=[], file_contains={})
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT, harness="codex", model="test-model", effort="high", timeout=0.15,
                command_builder=lambda **kwargs: [sys.executable, "-c", script], stream_log=lambda _: None,
            )
            started = time.monotonic()
            result = runner.run_case(case, Path(tmp))
            elapsed = time.monotonic() - started
        self.assertTrue(result["process"]["timed_out"])
        self.assertLess(elapsed, 1.0)

    def test_pi_sandbox_rejects_write_outside_disposable_run(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-runner-") as tmp:
            marker = Path(tmp) / "outside.txt"
            script = f"from pathlib import Path; Path({str(marker)!r}).write_text('escaped')"
            case = self._case(expected_activation=False, required_files=[], output_contains=[], file_contains={})
            runner = EvalRunner(
                repo_root=ROOT, harness="pi", model="test-model", effort="high", timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script], stream_log=lambda _: None,
            )
            result = runner.run_case(case, Path(tmp))
            self.assertFalse(marker.exists())
        self.assertEqual(result["process"]["status"], "failed")

    def test_native_event_shapes_distinguish_pi_tool_lifecycle_and_failure(self):
        stdout = "\n".join(
            [
                json.dumps({"type": "tool_execution_start", "toolCallId": "1", "toolName": "read", "args": {"path": "language-rules/SKILL.md"}}),
                json.dumps({"type": "tool_execution_end", "toolCallId": "1", "toolName": "read", "result": "body", "isError": False}),
                json.dumps({"type": "message_end", "message": {"stopReason": "error"}}),
                json.dumps({"type": "agent_end"}),
            ]
        )
        _, tools, loads, terminal, failed = _normalise_trace(stdout, "", "language-rules")
        self.assertEqual([item["status"] for item in tools], ["started", "completed"])
        self.assertEqual(loads, ["language-rules/SKILL.md"])
        self.assertTrue(terminal)
        self.assertTrue(failed)


if __name__ == "__main__":
    unittest.main()
