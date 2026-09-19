"""Regression coverage for privacy-safe Codex pre-tool probe capture."""

import json
import sys
import tempfile
import unittest
from datetime import date
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from codex_probe_capture import parse_probe_stream, select_exact_rollout


THREAD_ID = "thread-fixture-only"
TARGET = "/workspace/probe-module.mts"
RULE_PATH = "/workspace/.agents/skills/language-rules/references/typescript-standards.md"
HOLD_REASON = f"rule-delivery hold: read {RULE_PATH} before retry"
SYNTHETIC_SECRET = "SYNTHETIC_SECRET_NOT_A_CREDENTIAL_7F31"
READ_COMMAND = f'/bin/zsh -lc "cat {RULE_PATH}"'
PRE_READ_PATHS = [
    "/workspace/.agents/skills/language-rules/references/development-principles.md",
    "/workspace/.agents/skills/language-rules/references/test-gate.md",
    "/workspace/.agents/skills/language-rules/references/typescript-standards.md",
]
HELD_PATHS = [
    "/workspace/.agents/skills/language-rules/references/identifier-language.md",
    "/workspace/.agents/skills/language-rules/references/patterns-antipatterns.md",
    "/workspace/.agents/skills/language-rules/references/security.md",
]
PATCH_INPUT = (
    'const result = await tools.apply_patch("*** Begin Patch\\n'
    '*** Add File: /workspace/probe-module.mts\\n'
    '+export const codexProbe = (value: string): string => value.trim();\\n'
    '*** End Patch"); text(result);'
)


def item_stream(*, include_hold=True):
    denied = {
        "type": "item.completed",
        "item": {
            "id": "write-1",
            "type": "file_change",
            "status": "failed",
            "changes": [{"path": TARGET, "kind": "add"}],
            "error": {
                "permissionDecision": "deny",
                "permissionDecisionReason": HOLD_REASON,
                "private_context": SYNTHETIC_SECRET,
            },
        },
    }
    if not include_hold:
        denied["item"].pop("error")
    events = [
        {"type": "thread.started", "thread_id": THREAD_ID},
        {
            "type": "item.started",
            "item": {
                "id": "write-1",
                "type": "file_change",
                "status": "in_progress",
                "changes": [{"path": TARGET, "kind": "add"}],
                "sensitive_argument": SYNTHETIC_SECRET,
            },
        },
        denied,
        {
            "type": "item.started",
            "item": {
                "id": "read-1",
                "type": "command_execution",
                "command": READ_COMMAND,
                "status": "in_progress",
            },
        },
        {
            "type": "item.completed",
            "item": {
                "id": "read-1",
                "type": "command_execution",
                "command": READ_COMMAND,
                "status": "completed",
                "exit_code": 0,
            },
        },
        {
            "type": "item.started",
            "item": {
                "id": "write-2",
                "type": "file_change",
                "status": "in_progress",
                "changes": [{"path": TARGET, "kind": "add"}],
            },
        },
        {
            "type": "item.completed",
            "item": {
                "id": "write-2",
                "type": "file_change",
                "status": "completed",
                "changes": [{"path": TARGET, "kind": "add"}],
            },
        },
        {"type": "turn.completed"},
    ]
    return "\n".join(json.dumps(event) for event in events)


def rollout_stream(
    *, include_hook_marker=True, wrong_retry_target=False, wrong_retry_patch_target=False
):
    hook_text = "blocked by PreToolUse hook; read " + " ".join(HELD_PATHS)
    if not include_hook_marker:
        hook_text = "Please read " + " ".join(HELD_PATHS)
    retry_path = "/workspace/unrelated-module.mts" if wrong_retry_target else TARGET
    retry_patch_input = (
        PATCH_INPUT.replace(TARGET, "/other/probe-module.mts")
        if wrong_retry_patch_target
        else PATCH_INPUT
    )
    events = [
        {
            "type": "event_msg",
            "payload": {
                "type": "exec_command_end",
                "item": {
                    "type": "CommandExecution",
                    "command": ["cat", *PRE_READ_PATHS],
                    "status": "completed",
                    "exit_code": 0,
                },
            },
        },
        {
            "type": "response_item",
            "payload": {
                "type": "custom_tool_call",
                "call_id": "exec-hold",
                "name": "exec",
                "status": "completed",
                "input": PATCH_INPUT,
            },
        },
        {
            "type": "response_item",
            "payload": {
                "type": "custom_tool_call_output",
                "call_id": "exec-hold",
                "output": [
                    {"type": "input_text", "text": "tool result"},
                    {
                        "type": "input_text",
                        "text": hook_text + " " + SYNTHETIC_SECRET,
                    },
                ],
            },
        },
        {
            "type": "event_msg",
            "payload": {
                "type": "exec_command_end",
                "item": {
                    "type": "CommandExecution",
                    "command": ["cat", *HELD_PATHS],
                    "status": "completed",
                    "exit_code": 0,
                },
            },
        },
        {
            "type": "response_item",
            "payload": {
                "type": "custom_tool_call",
                "call_id": "exec-retry",
                "name": "exec",
                "status": "completed",
                "input": retry_patch_input,
            },
        },
        {
            "type": "event_msg",
            "payload": {
                "type": "file_change_end",
                "item": {
                    "type": "FileChange",
                    "status": "completed",
                    "changes": {retry_path: {"type": "add", "content": SYNTHETIC_SECRET}},
                },
            },
        },
        {
            "type": "response_item",
            "payload": {
                "type": "custom_tool_call_output",
                "call_id": "exec-retry",
                "output": [{"type": "input_text", "text": "applied"}],
            },
        },
    ]
    return "\n".join(json.dumps(event) for event in events)


def hook_output_text():
    return "blocked by PreToolUse hook; read " + " ".join(HELD_PATHS) + " " + SYNTHETIC_SECRET


class CodexProbeCaptureTests(unittest.TestCase):
    def test_item_events_capture_hold_read_retry_without_exposing_payloads(self):
        summary, thread_id = parse_probe_stream(item_stream(), target_path=TARGET)

        self.assertEqual(summary["write_attempts"], 2)
        self.assertEqual(summary["hold_rounds"], 1)
        self.assertEqual(summary["rule_reads"], 1)
        self.assertTrue(summary["retry_succeeded"])
        self.assertEqual(summary["reason_lengths_chars"], [len(HOLD_REASON)])
        self.assertEqual(summary["rule_basenames"], ["typescript-standards.md"])
        self.assertTrue(summary["terminal_event"])
        self.assertEqual(thread_id, THREAD_ID)

        safe_output = json.dumps(summary, sort_keys=True)
        self.assertNotIn(SYNTHETIC_SECRET, safe_output)
        self.assertNotIn(THREAD_ID, safe_output)
        self.assertNotIn(RULE_PATH, safe_output)

    def test_failed_file_change_without_explicit_hold_stays_unknown(self):
        summary, _ = parse_probe_stream(item_stream(include_hold=False), target_path=TARGET)

        self.assertIsNone(summary["hold_rounds"])
        self.assertIsNone(summary["retry_succeeded"])

    def test_rollout_hold_correlates_by_call_id_without_exposing_output(self):
        rollout = "\n".join(
            json.dumps(event)
            for event in (
                {
                    "type": "response_item",
                    "payload": {
                        "type": "custom_tool_call",
                        "call_id": "write-1",
                        "name": "apply_patch",
                        "arguments": SYNTHETIC_SECRET,
                    },
                },
                {
                    "type": "response_item",
                    "payload": {
                        "type": "custom_tool_call_output",
                        "call_id": "write-1",
                        "output": {
                            "permissionDecision": "deny",
                            "permissionDecisionReason": HOLD_REASON,
                            "private_context": SYNTHETIC_SECRET,
                        },
                    },
                },
            )
        )
        summary, _ = parse_probe_stream(
            item_stream(include_hold=False), target_path=TARGET, rollout_text=rollout
        )

        self.assertEqual(summary["hold_rounds"], 1)
        self.assertTrue(summary["retry_succeeded"])
        safe_output = json.dumps(summary, sort_keys=True)
        self.assertNotIn(SYNTHETIC_SECRET, safe_output)
        self.assertNotIn(RULE_PATH, safe_output)

    def test_retry_requires_exact_target_and_held_reference_read(self):
        unrelated_read_events = [json.loads(line) for line in item_stream().splitlines()]
        for event in unrelated_read_events:
            item = event.get("item")
            if isinstance(item, dict) and item.get("id") == "read-1":
                item["command"] = (
                    '/bin/zsh -lc "cat '
                    "/workspace/references/SYNTHETIC_SECRET_NOT_A_CREDENTIAL_7F31.md"
                    '"'
                )
        summary, _ = parse_probe_stream(
            "\n".join(json.dumps(event) for event in unrelated_read_events),
            target_path=TARGET,
        )
        self.assertFalse(summary["retry_succeeded"])
        self.assertNotIn(SYNTHETIC_SECRET, json.dumps(summary))

        unrelated_write_events = [json.loads(line) for line in item_stream().splitlines()]
        for event in unrelated_write_events:
            item = event.get("item")
            if isinstance(item, dict) and item.get("id") == "write-2":
                item["changes"] = [{"path": "/workspace/unrelated-module.mts", "kind": "add"}]
        summary, _ = parse_probe_stream(
            "\n".join(json.dumps(event) for event in unrelated_write_events),
            target_path=TARGET,
        )
        self.assertFalse(summary["retry_succeeded"])

    def test_exact_rollout_selector_checks_date_and_session_metadata(self):
        session_date = date(2026, 9, 19)
        with tempfile.TemporaryDirectory() as temp_dir:
            date_dir = Path(temp_dir) / "2026" / "09" / "19"
            date_dir.mkdir(parents=True)
            rollout = date_dir / f"rollout-{THREAD_ID}.jsonl"
            rollout.write_text(
                json.dumps({"type": "session_meta", "payload": {"id": THREAD_ID}}) + "\n",
                encoding="utf-8",
            )
            self.assertEqual(
                select_exact_rollout(temp_dir, THREAD_ID, session_date), str(rollout.resolve())
            )
            rollout.write_text(
                json.dumps({"type": "session_meta", "payload": {"id": "other-thread"}})
                + "\n",
                encoding="utf-8",
            )
            self.assertIsNone(select_exact_rollout(temp_dir, THREAD_ID, session_date))

    def test_codex_exec_wrapper_captures_hook_read_retry_and_multi_file_cats(self):
        summary, _ = parse_probe_stream(
            item_stream(include_hold=False), target_path=TARGET, rollout_text=rollout_stream()
        )

        self.assertEqual(summary["write_attempts"], 2)
        self.assertEqual(summary["hold_rounds"], 1)
        self.assertEqual(summary["rule_reads"], 6)
        self.assertEqual(summary["hook_output_block_lengths_chars"], [len(hook_output_text())])
        self.assertIsNone(summary["reason_lengths_chars"])
        self.assertIsNone(summary["unclassified_error_outputs"])
        self.assertTrue(summary["retry_succeeded"])
        self.assertTrue(summary["retry_input_matches"])
        self.assertEqual(summary["write_mechanism"], "apply_patch")
        self.assertEqual(
            summary["reads_before_first_write"],
            ["development-principles.md", "test-gate.md", "typescript-standards.md"],
        )
        self.assertEqual(
            summary["required_rule_basenames"],
            ["identifier-language.md", "patterns-antipatterns.md", "security.md"],
        )
        self.assertNotIn(SYNTHETIC_SECRET, json.dumps(summary, sort_keys=True))

    def test_exec_wrapper_text_without_explicit_hook_marker_is_not_a_hold(self):
        summary, _ = parse_probe_stream(
            item_stream(include_hold=False),
            target_path=TARGET,
            rollout_text=rollout_stream(include_hook_marker=False),
        )

        self.assertIsNone(summary["hold_rounds"])
        self.assertIsNone(summary["retry_succeeded"])

    def test_exec_wrapper_retry_must_change_the_held_target(self):
        summary, _ = parse_probe_stream(
            item_stream(include_hold=False),
            target_path=TARGET,
            rollout_text=rollout_stream(wrong_retry_target=True),
        )

        self.assertEqual(summary["hold_rounds"], 1)
        self.assertFalse(summary["retry_succeeded"])

    def test_exec_wrapper_rejects_same_basename_in_other_patch_directory(self):
        summary, _ = parse_probe_stream(
            item_stream(include_hold=False),
            target_path=TARGET,
            rollout_text=rollout_stream(wrong_retry_patch_target=True),
        )

        self.assertEqual(summary["hold_rounds"], 1)
        self.assertEqual(summary["write_attempts"], 1)
        self.assertTrue(summary["target_change_completed"])
        self.assertFalse(summary["retry_succeeded"])


if __name__ == "__main__":
    unittest.main()
