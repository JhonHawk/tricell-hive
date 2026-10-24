# Tools and dependencies

## When to read

Read before you install a missing tool or add a dependency during a build.

## Install a missing tool

Before installing a missing tool, check the project-local and available system installations and the required version. Prefer the project’s existing setup. Distinguish task-local dependency setup from system-wide installs or shell/global configuration changes; the latter require explicit authorization, which may already be present. Verify the actual OS and shell before using platform-specific commands rather than assuming them from the harness name.

## Add a dependency

Before adding a dependency, assess whether an existing project dependency, standard library, or native platform capability meets the requirement, including established design-system conventions. Choose based on compatibility, maintainability, and fit rather than a universal library preference. Explain a new dependency’s purpose and any overlap; consolidating existing usages remains separate unless authorized.
