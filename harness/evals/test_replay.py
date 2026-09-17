"""Regression checks for offline read-evidence replay."""
import hashlib
import json
import tempfile
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from replay import build_replay, replay_case, write_replay
from runner import EvalError


class ReplayTests(unittest.TestCase):
    def sample(self):
        events = [
            {'type': 'item.completed', 'item': {'type': 'command_execution', 'command': 'cat language-rules/SKILL.md', 'status': 'completed', 'exit_code': 0}},
            {'type': 'turn.completed'},
        ]
        recorded = {'id': 'sample', 'repeat': 1, 'skill': 'language-rules', 'verdict': 'not_verified', 'activation': 'unknown', 'unexpected_writes': []}
        original = {'id': 'sample', 'skill': 'language-rules', 'repeat_index': 1, 'process': {'status': 'passed', 'exit_code': 0, 'timed_out': False}, 'outcome': {'checks': {'required_reads': {}, 'required_files': {'out.md': {'status': 'passed'}}}}, 'evidence': {'stdout': '\n'.join(map(json.dumps, events)), 'stderr': '', 'skill_loads': [], 'protected_files': {}}}
        expected = {'skill': 'language-rules', 'expected_activation': True, 'required_reads': []}
        return recorded, original, expected

    def test_completed_read_can_resolve_unknown_activation(self):
        result = replay_case(*self.sample())
        self.assertEqual(result.after, 'pass')
        self.assertEqual(result.activation_after, 'passed')
        self.assertEqual(result.newly_observed_reads, ['language-rules/SKILL.md'])

    def test_read_evidence_does_not_erase_timeout(self):
        recorded, original, expected = self.sample()
        original['process'].update(status='failed', timed_out=True, exit_code=-9)
        self.assertEqual(replay_case(recorded, original, expected).after, 'not_verified')

    def test_read_evidence_does_not_erase_artifact_failure(self):
        recorded, original, expected = self.sample()
        recorded['verdict'] = 'fail'
        original['outcome']['checks']['required_files']['out.md']['status'] = 'failed'
        self.assertEqual(replay_case(recorded, original, expected).after, 'fail')

    def test_negative_activation_is_still_failure(self):
        recorded, original, expected = self.sample()
        recorded.update(verdict='pass', activation='passed')
        expected['expected_activation'] = False
        self.assertEqual(replay_case(recorded, original, expected).after, 'fail')

    def test_missing_terminal_event_cannot_pass(self):
        recorded, original, expected = self.sample()
        original['evidence']['stdout'] = original['evidence']['stdout'].splitlines()[0]
        self.assertEqual(replay_case(recorded, original, expected).after, 'not_verified')


class ReplayIntegrityTests(unittest.TestCase):
    def prepare(self, root):
        recorded, original, expected = ReplayTests().sample()
        manifest = {'version': 1, 'cases': [dict(expected, id='sample', category='explicit', prompt='Use language-rules', fixtures={}, protected_files=[], allowed_writes=['out.md'], required_files=['out.md'], output_contains=[])]}
        manifest_path = root / 'cases.json'
        manifest_path.write_text(json.dumps(manifest))
        raw_path = root / 'raw.json'
        raw_path.write_text(json.dumps({'cases': [original]}))
        baseline = {'manifest_sha256': hashlib.sha256(manifest_path.read_bytes()).hexdigest(), 'runs': [{'revision':'A', 'harness':'codex', 'raw_report':str(raw_path), 'raw_sha256':hashlib.sha256(raw_path.read_bytes()).hexdigest(), 'cases':[recorded]}]}
        baseline_path = root / 'baseline.json'
        baseline_path.write_text(json.dumps(baseline))
        return baseline_path, manifest_path, raw_path

    def test_replay_refuses_changed_manifest(self):
        with tempfile.TemporaryDirectory() as temporary:
            baseline, manifest, _ = self.prepare(Path(temporary))
            manifest.write_text(manifest.read_text() + ' ')
            with self.assertRaisesRegex(EvalError, 'manifest hash'):
                build_replay(baseline, manifest)

    def test_replay_refuses_changed_raw_trace(self):
        with tempfile.TemporaryDirectory() as temporary:
            baseline, manifest, raw = self.prepare(Path(temporary))
            raw.write_text(raw.read_text() + ' ')
            with self.assertRaisesRegex(EvalError, 'raw report hash'):
                build_replay(baseline, manifest)

    def test_replay_refuses_duplicate_case_identity(self):
        with tempfile.TemporaryDirectory() as temporary:
            baseline, manifest, _ = self.prepare(Path(temporary))
            data = json.loads(baseline.read_text())
            data['runs'][0]['cases'] *= 2
            baseline.write_text(json.dumps(data))
            with self.assertRaisesRegex(EvalError, 'case identities'):
                build_replay(baseline, manifest)

    def test_output_never_overwrites_an_existing_file(self):
        with tempfile.TemporaryDirectory() as temporary:
            baseline, manifest, _ = self.prepare(Path(temporary))
            report = build_replay(baseline, manifest)
            before = baseline.read_bytes()
            with self.assertRaises(FileExistsError):
                write_replay(report, baseline)
            self.assertEqual(baseline.read_bytes(), before)

    def test_verified_replay_writes_a_separate_comparison(self):
        with tempfile.TemporaryDirectory() as temporary:
            baseline, manifest, raw = self.prepare(Path(temporary))
            before = raw.read_bytes()
            report = build_replay(baseline, manifest)
            output = Path(temporary) / 'comparison.json'
            write_replay(report, output)
            result = json.loads(output.read_text())
            self.assertEqual(result['runs'][0]['cases'][0]['reads']['after'], 'pass')
            self.assertEqual(raw.read_bytes(), before)


if __name__ == '__main__':
    unittest.main()
