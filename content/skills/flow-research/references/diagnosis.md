# Diagnose with discriminating evidence

Read when investigating a failure, regression, or incident with an uncertain cause. Keep diagnostic experiments within the task's authorized effects.

1. Establish expected and observed behavior, affected scope, actual revision/configuration, and recent relevant changes. Reproduce when feasible; otherwise preserve logs, traces, state, or other evidence and state the limits of the reconstruction. Failure to reproduce is not proof that the event did not occur.
2. During an incident, prioritize reducing ongoing impact within existing authority while preserving useful evidence when practical. Mitigation may precede a complete causal explanation; distinguish temporary mitigation from a verified fix.
3. Form plausible hypotheses from the evidence and inspect the relevant component boundaries. Choose the smallest discriminating observation or experiment, name the expected result, and consider its risk before changing the system. Capture only necessary, redacted diagnostic data rather than dumping environment variables or payloads.
4. Compare the result with the prediction and revise the explanation. Reconsider the approach when actions stop producing information or increase impact; fixed counts of fixes or deployments do not establish a useful stopping condition.
5. Verify the affected behavior and relevant regression after correction or mitigation. Separate observed recovery, supported cause, and remaining uncertainty; identify temporary measures that still need retirement.
