# Tests in a refactor

Read from [tests](../tests.md) when work restructures code and keeps its behavior.

- Where no test pins the behavior the move touches, add one before moving anything.
- Do not edit a test to make a step pass: a test that must change means the behavior changed, which stops the step.
