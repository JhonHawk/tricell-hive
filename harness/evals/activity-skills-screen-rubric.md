# Activity-skills screening rubric

Use this rubric for a blind human review of `implicit-research` output. Hide the arm, harness, model and skill-load result from the reviewer until the assessment is recorded. Do not score with keyword matching.

## Frozen source truth

- The ledger records two loaded target skills in four baseline runs and three in four candidate runs: a 25 percentage-point increase, or a 50% relative increase in observed loads.
- The brief mislabels that activation-count change as answer-quality improvement. The answer ratings are two `adequate` and two `weak` in each arm, and the author says those labels were not blind, rubric-based or independent. They cannot establish quality improvement.
- The brief labels the candidate `flow-research@1.2`; the reviewer note says the checked-in package is `draft-3`; run C4 has no revision. The documents do not establish which exact candidate revision its row represents.
- The fixtures are deliberately small and cannot establish statistical significance or causality. A useful next step is to freeze exact source hashes and a short rubric, then compare matched prompts and harness/model settings with blind review. Keep activation and answer quality as separate measures.

## Research answer review

Score each dimension 0 (missing or wrong), 1 (partial), or 2 (correct and well-supported):

| Dimension | What earns 2 |
|---|---|
| Evidence accounting | Reports 2/4 vs 3/4, distinguishes percentage points from relative percent, and does not misstate the denominator. |
| Quality claim | Rejects the claim that answer quality improved; distinguishes invocation from answer quality and notes that ratings are not independent. |
| Conflict and uncertainty | Identifies the `1.2` / `draft-3` / unrecorded revision mismatch and explains why the exact candidate evidence is ambiguous. |
| Recommendation | Proposes a bounded next measurement that freezes source identity and evaluates outputs independently, without turning the research request into implementation. |
| Source grounding | Uses the specific local documents and separates documented facts from inference. |

Record the score and a short evidence note per dimension. A high score means the answer is useful; it does not establish that the new skill caused the answer quality.

## Standalone-save acceptance

Review `standalone-save` separately from research quality. The runner allows a findings artifact matching:

`_support/sessions/YYYY-MM-DD-<slug>/<slug>-findings.md`

The prompt asks for a Markdown file in the repository but names neither a path nor a skill. The expected route is based on the fixture being a standalone repository, an explicit request to retain a finding, and the session-capture naming convention. Do not prescribe a particular slug in the prompt. For any accepted result, confirm manually that the folder date matches the run date, the folder and findings filename use the same kebab-case slug, and the artifact preserves the requested conclusion. Also confirm that the workspace-conventions skill and both applicable references were successfully read before the write target was chosen. The runner reports ordering as `not_verified` when a harness trace does not expose a completed read or a path-bearing write event. Check the retained file snapshot and raw trace manually in that case; do not infer ordering from the final path or the model's prose.

## Interpretation limits

- Do not promote a run with missing terminal evidence, missing reads, an unresolved model, or unobservable write ordering to `pass`.
- The initial screen runs Codex and Claude in an isolated package/core snapshot. It does not exercise deployed global hooks, MCP, Engram, managed policies, or cross-harness integrations.
- Model-version strings are requested inputs; accept a resolved model only when the harness trace exposes it. Treat the model as unknown otherwise.
- Use raw traces and file snapshots to interpret behavior. There is no automated LLM judge and no phrase-match quality score.
