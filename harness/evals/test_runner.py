#!/usr/bin/env python3
"""Tests for the isolated skill evaluation runner."""

import hashlib
import json
import subprocess
import sys
import tempfile
import time
import unittest
from unittest.mock import patch
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(Path(__file__).parent))
from runner import (  # noqa: E402
    EvalError,
    EvalRunner,
    _matches_any,
    _normalise_claude_trace,
    _normalise_trace,
    _read_native_claude_oauth_token,
    _sandbox_argv,
    default_command_builder,
    load_manifest,
)


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

    def test_claude_uses_supplied_oauth_token_without_keychain_lookup(self):
        runner = EvalRunner(repo_root=ROOT, harness="claude", model="test-model", timeout=5)
        env = {"CLAUDE_CODE_OAUTH_TOKEN": "fake-session-token"}
        with patch("runner.subprocess.run") as keychain:
            auth = runner._link_auth(env, {"claude_home": Path("/tmp/isolated-claude")})

        keychain.assert_not_called()
        self.assertEqual(env["CLAUDE_CODE_OAUTH_TOKEN"], "fake-session-token")
        self.assertEqual(auth, {"status": "credential-reused", "source": "environment"})

    def test_claude_reads_native_keychain_once_per_runner_and_keeps_token_out_of_metadata(self):
        runner = EvalRunner(repo_root=ROOT, harness="claude", model="test-model", timeout=5)
        roots = {"claude_home": Path("/tmp/isolated-claude")}
        result = type("CommandResult", (), {
            "returncode": 0,
            "stdout": '{"claudeAiOauth":{"accessToken":"fake-keychain-token","refreshToken":"fake-refresh"}}',
            "stderr": "",
        })()
        with patch("runner.subprocess.run", return_value=result) as keychain:
            first_env = {"USER": "eval-account"}
            second_env = {"USER": "eval-account"}
            first_auth = runner._link_auth(first_env, roots)
            second_auth = runner._link_auth(second_env, roots)

        keychain.assert_called_once()
        self.assertEqual(first_env["CLAUDE_CODE_OAUTH_TOKEN"], "fake-keychain-token")
        self.assertEqual(second_env["CLAUDE_CODE_OAUTH_TOKEN"], "fake-keychain-token")
        self.assertEqual(first_auth, {"status": "credential-reused", "source": "native-keychain"})
        self.assertEqual(second_auth, first_auth)
        self.assertNotIn("fake-keychain-token", json.dumps([first_auth, second_auth]))

    def test_native_claude_keychain_reader_selects_only_access_token(self):
        result = type("CommandResult", (), {
            "returncode": 0,
            "stdout": '{"claudeAiOauth":{"accessToken":"fake-access","refreshToken":"fake-refresh"}}',
            "stderr": "",
        })()
        with patch("runner.subprocess.run", return_value=result) as keychain:
            token = _read_native_claude_oauth_token({"USER": "eval-account"})

        self.assertEqual(token, "fake-access")
        args, kwargs = keychain.call_args
        self.assertEqual(
            args[0],
            ["/usr/bin/security", "find-generic-password", "-a", "eval-account", "-w", "-s", "Claude Code-credentials"],
        )
        self.assertTrue(kwargs["capture_output"])
        self.assertEqual(kwargs["timeout"], 5)

    def test_native_claude_keychain_reader_fails_closed_without_leaking_output(self):
        result = type("CommandResult", (), {
            "returncode": 1,
            "stdout": "fake-access-token-in-command-output",
            "stderr": "fake-access-token-in-error-output",
        })()
        with patch("runner.subprocess.run", return_value=result):
            with self.assertRaises(EvalError) as error:
                _read_native_claude_oauth_token({"USER": "eval-account"})

        self.assertNotIn("fake-access-token", str(error.exception))

    def test_native_claude_keychain_reader_rejects_custom_config_without_guessing_service(self):
        with patch("runner.subprocess.run") as keychain:
            with self.assertRaises(EvalError):
                _read_native_claude_oauth_token({"USER": "eval-account", "CLAUDE_CONFIG_DIR": "/tmp/custom"})

        keychain.assert_not_called()

    def test_claude_version_probe_does_not_receive_oauth_token(self):
        runner = EvalRunner(repo_root=ROOT, harness="claude", model="test-model", timeout=5)
        with patch("runner.subprocess.run") as probe:
            probe.return_value.returncode = 0
            probe.return_value.stdout = "Claude Code 2.1.278"
            probe.return_value.stderr = ""
            runner._version("claude", {"PATH": "/usr/bin", "CLAUDE_CODE_OAUTH_TOKEN": "fake-secret"}, Path("/tmp"))

        self.assertNotIn("CLAUDE_CODE_OAUTH_TOKEN", probe.call_args.kwargs["env"])

    def test_claude_process_output_and_stream_logs_redact_oauth_token(self):
        token = "fake-secret-that-must-not-leak"
        script = (
            "import json, os, sys; "
            "print(os.environ['CLAUDE_CODE_OAUTH_TOKEN']); "
            "print(json.dumps({'type':'result','subtype':'success','result':os.environ['CLAUDE_CODE_OAUTH_TOKEN']})); "
            "sys.stderr.write(os.environ['CLAUDE_CODE_OAUTH_TOKEN'])"
        )
        streamed = []
        runner = EvalRunner(
            repo_root=ROOT,
            harness="claude",
            model="test-model",
            timeout=5,
            stream_log=streamed.append,
        )
        execution = runner._run_process(
            [sys.executable, "-c", script],
            Path("/tmp"),
            {"CLAUDE_CODE_OAUTH_TOKEN": token},
        )

        serialized = json.dumps(execution) + "".join(streamed)
        self.assertNotIn(token, serialized)
        self.assertIn("[REDACTED]", serialized)

    def test_grok_and_opencode_auth_are_symlinked_from_the_native_roots(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-auth-") as tmp:
            root = Path(tmp)
            native_grok = root / "native-grok"
            native_grok.mkdir()
            (native_grok / "auth.json").write_text("{}\n", encoding="utf-8")
            isolated_grok = root / "isolated-grok"
            isolated_grok.mkdir()
            grok = EvalRunner(repo_root=ROOT, harness="grok", model="test-model", timeout=5)
            self.assertEqual(
                grok._link_auth({"GROK_HOME": str(native_grok)}, {"grok_home": isolated_grok}),
                {"status": "symlinked"},
            )
            self.assertTrue((isolated_grok / "auth.json").is_symlink())
            self.assertEqual((isolated_grok / "auth.json").resolve(), (native_grok / "auth.json").resolve())

            native_data = root / "native-data"
            native_opencode = native_data / "opencode"
            native_opencode.mkdir(parents=True)
            (native_opencode / "auth.json").write_text("{}\n", encoding="utf-8")
            isolated_opencode = root / "isolated-data" / "opencode"
            opencode = EvalRunner(repo_root=ROOT, harness="opencode", model="test-model", timeout=5)
            self.assertEqual(
                opencode._link_auth(
                    {"XDG_DATA_HOME": str(native_data)},
                    {"opencode_data_home": isolated_opencode},
                ),
                {"status": "symlinked"},
            )
            self.assertTrue((isolated_opencode / "auth.json").is_symlink())
            self.assertEqual((isolated_opencode / "auth.json").resolve(), (native_opencode / "auth.json").resolve())

    def test_opencode_config_has_isolated_plugins_mcp_and_v2_skill_permissions(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-config-") as tmp:
            _, _, roots = EvalRunner(repo_root=ROOT, harness="opencode", model="test-model", timeout=5)._stage_skills(Path(tmp))
            config = json.loads(roots["opencode_config"].read_text(encoding="utf-8"))

        self.assertEqual(config["plugins"], [])
        self.assertEqual(config["mcp"], {"servers": {}})
        self.assertEqual(
            [rule for rule in config["permissions"] if rule["action"] == "skill"],
            [
                {"action": "skill", "resource": "flow-*", "effect": "ask"},
                {"action": "skill", "resource": "flow-research", "effect": "allow"},
            ],
        )
        self.assertTrue(any(rule["action"] == "external_directory" for rule in config["permissions"]))

    def test_opencode_isolation_marks_listener_address_as_unverified(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-listener-") as tmp:
            run_root = Path(tmp)
            runner = EvalRunner(repo_root=ROOT, harness="opencode", model="test-model", timeout=5)
            _, _, roots = runner._stage_skills(run_root)
            with patch.object(runner, "_link_auth", return_value={"status": "mocked"}):
                _, isolation = runner._environment(run_root, roots)

        self.assertEqual(isolation["network_listener"]["status"], "not_verified")
        self.assertIn("host-owned addresses", isolation["network_listener"]["limitation"])

    def test_source_root_stages_arm_core_and_claude_native_skills(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-source-") as tmp:
            source = Path(tmp) / "source"
            (source / "harness" / "agents-skills" / "workspace-conventions").mkdir(parents=True)
            (source / "global" / "skills" / "workspace-conventions").mkdir(parents=True)
            (source / "harness" / "agents-skills" / "workspace-conventions" / "SKILL.md").write_text(
                "generated skill", encoding="utf-8"
            )
            (source / "global" / "skills" / "workspace-conventions" / "SKILL.md").write_text(
                "canonical skill", encoding="utf-8"
            )
            (source / "harness" / "AGENTS.md").write_text("codex core", encoding="utf-8")
            (source / "global" / "CLAUDE.md").write_text("claude core", encoding="utf-8")
            runner = EvalRunner(source_root=source, harness="claude", model="test-model", timeout=5)
            with tempfile.TemporaryDirectory(prefix="hive-eval-stage-") as run_tmp:
                staged, _, roots = runner._stage_skills(Path(run_tmp))

                self.assertEqual((roots["codex_home"] / "AGENTS.md").read_text(encoding="utf-8"), "codex core")
                self.assertEqual((roots["claude_home"] / "CLAUDE.md").read_text(encoding="utf-8"), "claude core")
                self.assertEqual((roots["pi_core"]).read_text(encoding="utf-8"), "codex core")
                self.assertEqual((roots["workspace_core"]).read_text(encoding="utf-8"), "codex core")
                self.assertTrue(roots["opencode_config"].is_file())
                self.assertEqual(
                    (roots["claude_home"] / "skills" / "workspace-conventions" / "SKILL.md").read_text(encoding="utf-8"),
                    "generated skill",
                )
                self.assertEqual(
                    (staged / "workspace-conventions" / "SKILL.md").read_text(encoding="utf-8"),
                    "generated skill",
                )


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

    def test_arm_selects_expected_skill_without_changing_prompt(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-manifest-") as tmp:
            path = Path(tmp) / "cases.json"
            prompt = "Investiga la evidencia local y dame una conclusión."
            path.write_text(
                json.dumps(
                    {
                        "version": 1,
                        "cases": [
                            {
                                "id": "research",
                                "skill": "flow-research",
                                "skill_by_arm": {"baseline": "task-routing", "candidate": "flow-research"},
                                "category": "implicit",
                                "prompt": prompt,
                                "expected_activation": True,
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )

            baseline = load_manifest(path, arm="baseline")["cases"][0]
            candidate = load_manifest(path, arm="candidate")["cases"][0]

            self.assertEqual(baseline["skill"], "task-routing")
            self.assertEqual(candidate["skill"], "flow-research")
            self.assertEqual(baseline["prompt"], candidate["prompt"])

    def test_rejects_unknown_arm_skill_mapping(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-manifest-") as tmp:
            path = Path(tmp) / "cases.json"
            path.write_text(
                json.dumps(
                    {
                        "version": 1,
                        "cases": [
                            {
                                "id": "research",
                                "skill": "flow-research",
                                "skill_by_arm": {"candidate": "flow-research"},
                                "category": "implicit",
                                "prompt": "Investiga.",
                                "expected_activation": True,
                            }
                        ],
                    }
                ),
                encoding="utf-8",
            )

            with self.assertRaises(EvalError):
                load_manifest(path, arm="baseline")

    def test_screen_accepts_session_slug_without_accepting_other_workspace_paths(self):
        case = load_manifest(ROOT / "harness" / "evals" / "activity-skills-screen.json")["cases"][1]
        accepted = "_support/sessions/2026-09-20-retain-this-conclusion/retain-this-conclusion-findings.md"
        outside_sessions = "_support/workspace/2026-09-20-retain-this-conclusion/retain-this-conclusion-findings.md"

        self.assertTrue(_matches_any(accepted, case["allowed_writes"]))
        self.assertTrue(_matches_any(accepted, case["required_files"]))
        self.assertFalse(_matches_any(outside_sessions, case["allowed_writes"]))
        self.assertFalse(_matches_any(outside_sessions, case["required_files"]))

    def test_standalone_save_requests_repository_markdown_without_naming_placement(self):
        case = load_manifest(ROOT / "harness" / "evals" / "activity-skills-screen.json")["cases"][1]

        self.assertIn("archivo Markdown de este repositorio", case["prompt"])
        self.assertNotIn("workspace-conventions", case["prompt"])
        self.assertNotIn("_support/sessions", case["prompt"])

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

    def test_claude_structured_events_require_completed_tool_results(self):
        events = [
            {"type": "assistant", "message": {"content": [
                {"type": "tool_use", "id": "r1", "name": "Read", "input": {"file_path": "/tmp/workspace/.claude/skills/workspace-conventions/references/project-structure.md"}}
            ]}},
            {"type": "user", "message": {"content": [
                {"type": "tool_result", "tool_use_id": "r1", "content": "fixture rule", "is_error": False}
            ]}},
            {"type": "assistant", "message": {"content": [
                {"type": "tool_use", "id": "w1", "name": "Write", "input": {"file_path": "/tmp/workspace/_support/sessions/2026-09-20-conclusion/conclusion.md", "content": "saved"}}
            ]}},
            {"type": "user", "message": {"content": [
                {"type": "tool_result", "tool_use_id": "w1", "content": "File written successfully", "is_error": False}
            ]}},
            {"type": "result", "subtype": "success", "is_error": False, "result": "Saved.", "modelUsage": {"claude-opus-test": {"inputTokens": 4}}},
        ]

        _, tool_events, skill_loads, terminal, failure = _normalise_claude_trace(
            "\n".join(json.dumps(event) for event in events), "", "workspace-conventions"
        )

        self.assertTrue(terminal)
        self.assertFalse(failure)
        self.assertIn("/tmp/workspace/.claude/skills/workspace-conventions/references/project-structure.md", skill_loads)
        self.assertEqual([event["name"] for event in tool_events], ["Read", "Write"])
        self.assertLess(tool_events[0]["sequence"], tool_events[1]["sequence"])
        self.assertEqual(tool_events[1]["paths"], ["/tmp/workspace/_support/sessions/2026-09-20-conclusion/conclusion.md"])

    def test_claude_failed_or_unmatched_tool_result_does_not_prove_read(self):
        events = [
            {"type": "assistant", "message": {"content": [
                {"type": "tool_use", "id": "r1", "name": "Read", "input": {"file_path": "/tmp/workspace/workspace-conventions/SKILL.md"}},
                {"type": "tool_use", "id": "r2", "name": "Read", "input": {"file_path": "/tmp/workspace/other.md"}},
            ]}},
            {"type": "user", "message": {"content": [
                {"type": "tool_result", "tool_use_id": "r1", "content": "denied", "is_error": True},
            ]}},
        ]

        _, _, skill_loads, terminal, _ = _normalise_claude_trace(
            "\n".join(json.dumps(event) for event in events), "", "workspace-conventions"
        )

        self.assertFalse(terminal)
        self.assertEqual(skill_loads, [])

    def test_claude_parallel_read_and_write_do_not_prove_read_before_write(self):
        events = [
            {"type": "assistant", "message": {"content": [
                {"type": "tool_use", "id": "r1", "name": "Read", "input": {"file_path": "/tmp/workspace/references/project-structure.md"}},
                {"type": "tool_use", "id": "w1", "name": "Write", "input": {"file_path": "/tmp/workspace/_support/sessions/findings.md", "content": "saved"}},
            ]}},
            {"type": "user", "message": {"content": [
                {"type": "tool_result", "tool_use_id": "r1", "content": "reference", "is_error": False},
                {"type": "tool_result", "tool_use_id": "w1", "content": "written", "is_error": False},
            ]}},
        ]

        _, tool_events, _, _, _ = _normalise_claude_trace(
            "\n".join(json.dumps(event) for event in events), "", "workspace-conventions"
        )

        self.assertEqual(tool_events[0]["sequence"], tool_events[1]["sequence"])

    def test_claude_builder_uses_native_json_output_without_injecting_skill_name(self):
        prompt = "Investiga los documentos locales."
        argv = default_command_builder(
            harness="claude",
            model="opus[1m]",
            effort="high",
            prompt=prompt,
            skill="flow-research",
            category="implicit",
            cwd=Path("/tmp/workspace"),
            environment={
                "HOME": "/tmp/eval-home",
                "HIVE_EVAL_SKILLS_DIR": "/tmp/staged-skills",
            },
        )

        self.assertEqual(argv[-1], prompt)
        self.assertIn("--output-format", argv)
        self.assertIn("stream-json", argv)
        self.assertIn("--no-session-persistence", argv)
        self.assertIn("opus[1m]", argv)
        mcp_config = argv[argv.index("--mcp-config") + 1]
        self.assertEqual(json.loads(mcp_config), {"mcpServers": {}})
        claude_skills = Path("/tmp/eval-home/.claude/skills").resolve()
        self.assertEqual(argv[argv.index("--add-dir") + 1], str(claude_skills))
        self.assertIn(
            f"Edit(//{str(claude_skills).lstrip('/')}/**)",
            argv[argv.index("--disallowedTools") + 1],
        )
        self.assertNotIn("Write(", argv[argv.index("--disallowedTools") + 1])
        self.assertNotIn("flow-research", argv)

    def test_pi_builder_does_not_disable_context_files(self):
        argv = default_command_builder(
            harness="pi",
            model="xai/grok-4.6",
            effort="high",
            prompt="Guarda esta conclusión.",
            skill="workspace-conventions",
            category="implicit",
            cwd=Path("/tmp/workspace"),
            environment={"HIVE_EVAL_SKILLS_DIR": "/tmp/staged-skills"},
        )

        self.assertNotIn("--no-context-files", argv)
        self.assertIn("--no-extensions", argv)
        self.assertIn("--no-session", argv)
        self.assertIn("xai/grok-4.6", argv)
        self.assertIn("high", argv)

    def test_grok_builder_uses_headless_json_and_native_model_controls(self):
        prompt = "Investiga las notas locales y resuelve la contradicción."
        argv = default_command_builder(
            harness="grok",
            model="grok-4.6",
            effort="high",
            prompt=prompt,
            skill="flow-research",
            category="implicit",
            cwd=Path("/tmp/workspace"),
            environment={},
        )

        self.assertEqual(argv[0], "grok")
        self.assertEqual(argv[argv.index("--cwd") + 1], "/tmp/workspace")
        self.assertEqual(argv[argv.index("--model") + 1], "grok-4.6")
        self.assertEqual(argv[argv.index("--reasoning-effort") + 1], "high")
        self.assertEqual(argv[argv.index("--output-format") + 1], "streaming-json")
        self.assertIn("--no-memory", argv)
        self.assertIn("--no-subagents", argv)
        self.assertIn("--disable-web-search", argv)
        self.assertEqual(argv[argv.index("--permission-mode") + 1], "dontAsk")
        allow_values = [argv[index + 1] for index, value in enumerate(argv[:-1]) if value == "--allow"]
        self.assertEqual(allow_values, ["Bash", "Read", "Grep", "Glob", "Edit", "Write"])
        self.assertEqual(argv[argv.index("--single") + 1], prompt)
        self.assertNotIn("flow-research", argv)

    def test_opencode_builder_uses_v2_headless_json_and_model_variant(self):
        prompt = "Investiga las notas locales y resuelve la contradicción."
        argv = default_command_builder(
            harness="opencode",
            model="opencode-go/deepseek-flash#max",
            effort="high",
            prompt=prompt,
            skill="flow-research",
            category="implicit",
            cwd=Path("/tmp/workspace"),
            environment={},
        )

        self.assertEqual(argv[0], "opencode")
        self.assertIn("--standalone", argv)
        self.assertEqual(argv[argv.index("--model") + 1], "opencode-go/deepseek-flash#max")
        self.assertEqual(argv[argv.index("--format") + 1], "json")
        self.assertEqual(argv[-1], prompt)
        self.assertNotIn("flow-research", argv)

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

    def test_pi_grok_and_opencode_report_the_core_that_each_harness_receives(self):
        script = "import json; print(json.dumps({'type':'turn.completed'}))"
        expected_core_sources = {
            "pi": ROOT / "harness" / "AGENTS.md",
            "grok": ROOT / "global" / "CLAUDE.md",
            "opencode": ROOT / "harness" / "AGENTS.md",
        }
        case = self._case(
            expected_activation=False,
            required_reads=[],
            required_files=[],
            output_contains=[],
            file_contains={},
        )
        for harness, source in expected_core_sources.items():
            with self.subTest(harness=harness), tempfile.TemporaryDirectory(prefix="hive-eval-core-") as tmp:
                runner = EvalRunner(
                    repo_root=ROOT,
                    harness=harness,
                    model="test-model",
                    timeout=5,
                    command_builder=lambda **kwargs: [sys.executable, "-c", script],
                    stream_log=lambda _: None,
                )
                with patch.object(runner, "_link_auth", return_value={"status": "mocked"}), patch(
                    "runner._sandbox_argv", side_effect=lambda argv, *_: argv
                ):
                    result = runner.run_case(case, Path(tmp))

                expected = hashlib.sha256(source.read_bytes()).hexdigest()
                self.assertEqual(result["hashes"]["cores"][harness], expected)

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

    def test_retained_artifacts_omit_files_containing_claude_oauth_token(self):
        token = "fake-token-must-not-be-retained"
        script = (
            "import json, os, pathlib; "
            "pathlib.Path('secret.txt').write_text(os.environ['CLAUDE_CODE_OAUTH_TOKEN']); "
            "print(json.dumps({'type':'result','subtype':'success','result':'saved'}))"
        )
        case = self._case(
            expected_activation=False,
            required_reads=[],
            allowed_writes=["secret.txt"],
            required_files=[],
            output_contains=[],
            file_contains={},
        )
        with tempfile.TemporaryDirectory(prefix="hive-eval-retain-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT,
                harness="claude",
                model="test-model",
                timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script],
                stream_log=lambda _: None,
            )
            with patch.dict("os.environ", {"CLAUDE_CODE_OAUTH_TOKEN": token}):
                result = runner.run_case(case, Path(tmp), retain_dir=Path(tmp) / "artifacts")

            artifact = result["evidence"]["artifacts"]
            self.assertFalse((Path(artifact["root"]) / "secret.txt").exists())
            self.assertEqual(artifact["status"], "retained-with-security-omissions")
            self.assertEqual(
                artifact["omitted_files"],
                [{"path": "secret.txt", "reason": "contains Claude OAuth token"}],
            )
            self.assertNotIn(token, json.dumps(result))

    def test_completed_claude_write_outside_workspace_is_reported_as_out_of_scope(self):
        script = (
            "import json, os, pathlib; "
            "target = pathlib.Path(os.environ['HOME']) / '.claude' / 'projects' / 'fixture' / 'memory' / 'MEMORY.md'; "
            "target.parent.mkdir(parents=True); target.write_text('saved memory'); "
            "print(json.dumps({'type':'assistant','message':{'content':[{'type':'tool_use','id':'w1','name':'Write','input':{'file_path':str(target),'content':'saved memory'}}]}})); "
            "print(json.dumps({'type':'user','message':{'content':[{'type':'tool_result','tool_use_id':'w1','content':'written','is_error':False}]}})); "
            "print(json.dumps({'type':'result','subtype':'success','result':'Done'}))"
        )
        case = self._case(
            expected_activation=False,
            required_reads=[],
            allowed_writes=[],
            required_files=[],
            output_contains=[],
            file_contains={},
        )
        with tempfile.TemporaryDirectory(prefix="hive-eval-outside-write-") as tmp:
            runner = EvalRunner(
                repo_root=ROOT,
                harness="claude",
                model="test-model",
                timeout=5,
                command_builder=lambda **kwargs: [sys.executable, "-c", script],
                stream_log=lambda _: None,
            )
            with patch.dict("os.environ", {"CLAUDE_CODE_OAUTH_TOKEN": "fake-token"}):
                result = runner.run_case(case, Path(tmp))

        self.assertEqual(result["evidence"]["writes"]["status"], "failed")
        self.assertTrue(any(path.endswith("/.claude/projects/fixture/memory/MEMORY.md") for path in result["evidence"]["unexpected_writes"]))
        self.assertTrue(any(
            path.endswith("/.claude/projects/fixture/memory/MEMORY.md")
            for path in result["evidence"]["operations"]["write_tool_calls"]["out_of_workspace_targets"]
        ))
        self.assertEqual(result["outcome"]["status"], "failed")

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

    def test_pi_auth_is_mutable_in_private_run_and_never_retained(self):
        with tempfile.TemporaryDirectory(prefix="hive-eval-pi-auth-") as tmp:
            root = Path(tmp)
            native_pi = root / "native-pi"
            native_pi.mkdir()
            native_auth = native_pi / "auth.json"
            native_auth.write_text('{"token":"fake-native-secret"}\n', encoding="utf-8")
            native_auth.chmod(0o600)
            retain_dir = root / "retained"
            case = self._case(
                fixtures={}, protected_files=[], expected_activation=False, required_reads=[],
                allowed_writes=[], required_files=[], file_contains={}, output_contains=[],
            )

            def command_builder(**kwargs):
                isolated_auth = Path(kwargs["environment"]["PI_CODING_AGENT_DIR"]) / "auth.json"
                script = (
                    "import json, pathlib, stat; "
                    f"auth=pathlib.Path({str(isolated_auth)!r}); "
                    "assert auth.is_file() and not auth.is_symlink(); "
                    "assert stat.S_IMODE(auth.stat().st_mode)==0o600; "
                    "auth.write_text('mutated'); "
                    "print(json.dumps({'type':'agent_end'}))"
                )
                return [sys.executable, "-c", script]

            runner = EvalRunner(
                repo_root=ROOT, harness="pi", model="test-model", timeout=5,
                command_builder=command_builder, stream_log=lambda _: None,
            )
            with patch.dict("runner.os.environ", {"PI_CODING_AGENT_DIR": str(native_pi)}):
                result = runner.run_case(case, retain_dir=retain_dir)

            self.assertEqual(result["process"]["status"], "passed")
            self.assertEqual(native_auth.read_text(encoding="utf-8"), '{"token":"fake-native-secret"}\n')
            self.assertEqual(native_auth.stat().st_mode & 0o777, 0o600)
            artifact_evidence = result["evidence"]["artifacts"]
            self.assertNotIn("pi-agent/auth.json", artifact_evidence["files"])
            self.assertIn(
                {
                    "path": "pi-agent/auth.json",
                    "reason": "isolated Pi credentials are excluded from retained artifacts",
                },
                artifact_evidence["omitted_files"],
            )
            self.assertNotIn("fake-native-secret", json.dumps(result))


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


class SandboxNetworkTests(unittest.TestCase):
    def _run_sandbox_probe(self, harness: str, script: str, *args: str) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory(prefix="hive-eval-sandbox-", dir="/tmp") as temporary:
            run_root = Path(temporary).resolve()
            workspace = run_root / "workspace"
            workspace.mkdir()
            argv = _sandbox_argv(
                [sys.executable, "-c", script, *args], run_root, workspace, harness=harness
            )
            return subprocess.run(argv, cwd=run_root, capture_output=True, text=True, timeout=5, check=False)

    def test_grok_can_connect_to_its_unix_leader_socket_inside_run_root(self):
        script = (
            "import socket, sys; "
            "server=socket.socket(socket.AF_UNIX); server.bind(sys.argv[1]); server.listen(); "
            "client=socket.socket(socket.AF_UNIX); client.connect(sys.argv[1]); "
            "peer,_=server.accept(); client.sendall(b'x'); "
            "assert peer.recv(1)==b'x'; print('socket-ok')"
        )

        result = self._run_sandbox_probe("grok", script, "leader.sock")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("socket-ok", result.stdout)

    def test_grok_cannot_bind_unix_socket_outside_run_root(self):
        script = "import socket, sys; socket.socket(socket.AF_UNIX).bind(sys.argv[1])"
        with tempfile.TemporaryDirectory(prefix="hive-eval-outside-", dir="/tmp") as outside:
            result = self._run_sandbox_probe("grok", script, str(Path(outside) / "leader.sock"))

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Operation not permitted", result.stderr)

    def test_opencode_can_bind_loopback_tcp_listener(self):
        script = (
            "import socket; server=socket.socket(); server.bind(('127.0.0.1',0)); server.listen(); "
            "client=socket.create_connection(server.getsockname(), timeout=2); "
            "peer,_=server.accept(); client.sendall(b'x'); "
            "assert peer.recv(1)==b'x'; print('listener-ok')"
        )

        result = self._run_sandbox_probe("opencode", script)

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("listener-ok", result.stdout)

    def test_claude_and_pi_do_not_receive_tcp_listener_permission(self):
        script = "import socket; s=socket.socket(); s.bind(('127.0.0.1',0)); s.listen()"

        for harness in ("claude", "pi"):
            with self.subTest(harness=harness):
                result = self._run_sandbox_probe(harness, script)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Operation not permitted", result.stderr)


if __name__ == "__main__":
    unittest.main()
