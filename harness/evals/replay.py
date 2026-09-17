#!/usr/bin/env python3
"""Re-evaluate read evidence from an immutable, hash-verified pilot baseline."""
from __future__ import annotations

import argparse
import hashlib
import json
import sys
from dataclasses import asdict, dataclass
from pathlib import Path

from runner import DEFAULT_MANIFEST, EvalError, _normalise_trace, load_manifest


@dataclass
class ReadReplay:
    before: str
    after: str
    activation_before: str
    activation_after: str
    missing_before: list[str]
    missing_after: list[str]
    newly_observed_reads: list[str]


@dataclass
class CaseReplay:
    id: str
    repeat: int
    skill: str
    process_status: str
    timed_out: bool
    reads: ReadReplay


@dataclass
class RunReplay:
    revision: str
    harness: str
    raw_sha256: str
    cases: list[CaseReplay]


@dataclass
class ReplayReport:
    mode: str
    baseline_sha256: str
    manifest_sha256: str
    runner_sha256: str
    replay_sha256: str
    runs: list[RunReplay]


def _digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def _matches(read: str, required: str) -> bool:
    return read == required or read.endswith('/' + required)


def replay_case(recorded: dict, original: dict, expected: dict) -> ReadReplay:
    evidence = original['evidence']
    _, _, reads, terminal, failure = _normalise_trace(evidence['stdout'], evidence['stderr'], expected['skill'])
    process = original['process']
    complete = process['status'] == 'passed' and process['exit_code'] == 0 and not process['timed_out'] and terminal and not failure
    target = expected['skill'] + '/SKILL.md'
    loaded = any(_matches(read, target) for read in reads)
    if loaded:
        activation = 'passed' if expected['expected_activation'] else 'failed'
    else:
        activation = 'passed' if not expected['expected_activation'] and complete else 'unknown'
    checks = original['outcome']['checks']
    if set(checks['required_reads']) != set(expected['required_reads']):
        raise EvalError('recorded required reads differ from the manifest')
    missing_before = [key for key, value in checks['required_reads'].items() if value['status'] != 'passed']
    missing_after = [required for required in expected['required_reads'] if not any(_matches(read, required) for read in reads)]
    preserved = [value['status'] for group, values in checks.items() if group != 'required_reads' for value in values.values()]
    preserved.extend(value['status'] for value in evidence['protected_files'].values())
    preserved.append('failed' if recorded['unexpected_writes'] else 'passed')
    statuses = preserved + [activation] + (['unknown'] if missing_after else [])
    if not complete:
        verdict = 'not_verified'
    elif 'failed' in statuses:
        verdict = 'fail'
    elif any(status != 'passed' for status in statuses):
        verdict = 'not_verified'
    else:
        verdict = 'pass'
    relevant = [target, *expected['required_reads']]
    new_reads = sorted(set(reads) - set(evidence['skill_loads']))
    return ReadReplay(
        recorded['verdict'], verdict, recorded['activation'], activation,
        missing_before, missing_after,
        [read for read in new_reads if any(_matches(read, required) for required in relevant)],
    )


def build_replay(baseline_path: Path, manifest_path: Path) -> ReplayReport:
    baseline_bytes = baseline_path.read_bytes()
    baseline = json.loads(baseline_bytes)
    manifest_digest = _digest(manifest_path.read_bytes())
    if manifest_digest != baseline['manifest_sha256']:
        raise EvalError('manifest hash differs from the recorded baseline')
    specs = {case['id']: case for case in load_manifest(manifest_path)['cases']}
    runs = []
    if not baseline['runs']:
        raise EvalError('baseline has no runs')
    for recorded_run in baseline['runs']:
        raw_bytes = Path(recorded_run['raw_report']).read_bytes()
        raw_digest = _digest(raw_bytes)
        if raw_digest != recorded_run['raw_sha256']:
            raise EvalError(f"raw report hash mismatch: {recorded_run['raw_report']}")
        original = json.loads(raw_bytes)
        originals = {(case['id'], case['repeat_index']): case for case in original['cases']}
        keys = [(case['id'], case['repeat']) for case in recorded_run['cases']]
        if len(originals) != len(original['cases']) or len(set(keys)) != len(keys) or set(keys) != set(originals):
            raise EvalError('baseline case identities differ from raw report')
        cases = []
        for recorded in recorded_run['cases']:
            case = originals[(recorded['id'], recorded['repeat'])]
            expected = specs[case['id']]
            if case['skill'] != expected['skill'] or recorded['skill'] != expected['skill']:
                raise EvalError('skill identity differs from manifest')
            process = case['process']
            cases.append(CaseReplay(case['id'], case['repeat_index'], case['skill'], process['status'], process['timed_out'], replay_case(recorded, case, expected)))
        runs.append(RunReplay(recorded_run['revision'], recorded_run['harness'], raw_digest, cases))
    return ReplayReport('read-observability-replay', _digest(baseline_bytes), manifest_digest,
                        _digest(Path(__file__).with_name('runner.py').read_bytes()),
                        _digest(Path(__file__).read_bytes()), runs)


def write_replay(report: ReplayReport, output: Path) -> None:
    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open('x', encoding='utf-8') as stream:
        json.dump(asdict(report), stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    output.chmod(0o600)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--baseline', type=Path, required=True)
    parser.add_argument('--manifest', type=Path, default=DEFAULT_MANIFEST)
    parser.add_argument('--output', type=Path, required=True, help='new output file; existing files are never overwritten')
    args = parser.parse_args()
    try:
        report = build_replay(args.baseline, args.manifest)
        write_replay(report, args.output)
    except (OSError, ValueError, KeyError, TypeError) as exc:
        print(f'error: {exc}', file=sys.stderr)
        return 2
    print(f'Replayed {sum(len(run.cases) for run in report.runs)} cases without model execution: {args.output}')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
