"""Git-mode advisory is main-thread only: a subagent never owns the session's git mode."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

HOOK = Path(__file__).with_name('post-tool-hub.sh')
ADVISORY = 'Session git mode'


def run_hook(scratch: str, payload: dict) -> str:
    env = {**os.environ, 'TMPDIR': scratch}
    result = subprocess.run(
        ['/bin/bash', str(HOOK)], input=json.dumps(payload),
        text=True, capture_output=True, env=env, check=True,
    )
    return result.stdout


def write_payload(scratch: str, session: str, **identity) -> dict:
    return {
        'session_id': session, 'tool_name': 'Write', 'cwd': scratch,
        'tool_input': {'file_path': f'{scratch}/x.md'}, **identity,
    }


class GitModeSubagentTests(unittest.TestCase):
    def test_main_thread_still_gets_the_advisory(self):
        with tempfile.TemporaryDirectory() as scratch:
            subprocess.run(['git', 'init', '-q', scratch], check=True)
            out = run_hook(scratch, write_payload(scratch, 'main-thread'))
            self.assertIn(ADVISORY, out)

    def test_subagent_identity_silences_the_advisory(self):
        cases = {
            'claude-agent-id': {'agent_id': 'agent_abc123'},
            'claude-agent-id-camel': {'agentId': 'agent_abc123'},
            'claude-agent-type': {'agent_type': 'react-developer'},
            'claude-agent-type-camel': {'agentType': 'react-developer'},
            'grok-subagent-type': {'subagentType': 'backend-developer'},
            'agent-name': {'agent_name': 'sdd-explore'},
        }
        with tempfile.TemporaryDirectory() as scratch:
            subprocess.run(['git', 'init', '-q', scratch], check=True)
            for session, identity in cases.items():
                with self.subTest(session=session):
                    out = run_hook(scratch, write_payload(scratch, session, **identity))
                    self.assertNotIn(ADVISORY, out)

    def test_other_sections_still_fire_inside_a_subagent(self):
        """Only the git-mode section is main-thread-only; the hub is otherwise unchanged."""
        with tempfile.TemporaryDirectory() as scratch:
            subprocess.run(['git', 'init', '-q', scratch], check=True)
            out = run_hook(scratch, {
                'session_id': 'zsh-in-subagent', 'tool_name': 'Bash',
                'cwd': scratch, 'agent_type': 'react-developer',
                'tool_input': {'command': 'echo hi'},
                'tool_response': {'stdout': 'zsh:4: read-only variable: status'},
            })
            self.assertIn('zsh failure signature', out)


if __name__ == '__main__':
    unittest.main()
