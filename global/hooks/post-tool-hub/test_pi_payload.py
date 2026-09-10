"""PI metadata must not trigger Claude-specific recovery or Git heuristics."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

HOOK = Path(__file__).with_name('post-tool-hub.sh')


class PiPayloadTests(unittest.TestCase):
    def test_pi_skips_claude_heuristics_and_resets_delegation(self):
        with tempfile.TemporaryDirectory() as scratch:
            root = Path(scratch)
            subprocess.run(['git', 'init', '-q', scratch], check=True)
            env = {**os.environ, 'TMPDIR': scratch}
            for tool in ['Write', 'subagent']:
                result = subprocess.run(
                    ['/bin/bash', str(HOOK)], input=json.dumps({
                        'harness': 'pi', 'session_id': 'pi-test',
                        'tool_name': tool, 'cwd': scratch, 'tool_input': {},
                    }), text=True, capture_output=True, env=env, check=True,
                )
                self.assertNotIn('Session git mode', result.stdout)
            self.assertFalse((root / 'claude-flow-plan-recovery-pi-test').exists())
            self.assertFalse((root / 'claude-askq-seen-pi-test').exists())
            self.assertEqual((root / 'claude-delegation-reminder-pi-test').read_text(), 'main|0')


if __name__ == '__main__':
    unittest.main()
