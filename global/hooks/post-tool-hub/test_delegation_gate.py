"""Delegation counter: advisory at each multiple of 20, self-sufficient directive from the third firing on."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

HOOK = Path(__file__).with_name('post-tool-hub.sh')
DIRECTIVE = 'rule may be out of context'


def run_hook(scratch: str, tool: str) -> str:
    env = {**os.environ, 'TMPDIR': scratch}
    result = subprocess.run(
        ['/bin/bash', str(HOOK)], input=json.dumps({
            'session_id': 'gate-test', 'tool_name': tool, 'cwd': scratch,
            'tool_input': {'file_path': f'{scratch}/x.md'},
        }), text=True, capture_output=True, env=env, check=True,
    )
    return result.stdout


class DelegationGateTests(unittest.TestCase):
    def test_directive_appears_from_third_firing(self):
        with tempfile.TemporaryDirectory() as scratch:
            subprocess.run(['git', 'init', '-q', scratch], check=True)
            outputs = {}
            for i in range(1, 61):
                outputs[i] = run_hook(scratch, 'Read')
            self.assertIn('~20 main-thread tool calls', outputs[20])
            self.assertNotIn(DIRECTIVE, outputs[20])
            self.assertNotIn(DIRECTIVE, outputs[40])
            self.assertIn('~60 main-thread tool calls', outputs[60])
            self.assertIn(DIRECTIVE, outputs[60])
            self.assertIn('sdd-explore', outputs[60])
            self.assertEqual(outputs[59], '')


if __name__ == '__main__':
    unittest.main()
