# Boundary discipline

Read when you add validation, error handling or an adapter around a framework, a transport or an external service.

1. Validate, narrow types and handle errors where data crosses a system boundary: command arguments, configuration and environment variables, network input, a third-party API, a file, a stored row. Parse it into a domain type there.
2. Inside the boundary, trust the types. Do not re-validate the same value or add a defensive check for a state the types exclude; each one hides the question of who should have stopped it.
3. Keep transport and framework types out of the public surface of a module; convert them at the edge.
4. Put business logic in functions that need no framework, with a thin shell that reads input, calls them and writes output.

Two tests. Is this data crossing a boundary right now? If not, the check is redundant. Could this be a plain function the shell calls? If so, extract it.

This does not weaken input validation required by [security boundaries](../build/security-boundaries.md): untrusted input is always validated where it enters.
