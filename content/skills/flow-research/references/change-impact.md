# Change impact

## When to read

Read before recommending, planning, or building either of these, unless it is a mechanical change as the global guidance defines it:

- a change that removes, closes, disables, migrates, or renames something that runs in production, such as a dependency, a network rule or port, an endpoint, a scheduled job, or a table or collection, or that moves it to another host or process;
- a behavior change in code that runs in production when no test or CI run covers that behavior.

Reading it does not invoke planning or authorize any effect.

## Scale the check to its signals

Always run step 1. Run each later step only when you observe its signal, and state in one line which signals you found. A change that shows none needs no further step.

1. **Search for the value, not only the symbol.** Search for the literal value the change touches: the host name or IP address, port, connection-string key, table or collection name, environment variable, route, or job name. Include configuration, deployment scripts, container and process-manager files, crontabs, and infrastructure code. Search every repository of the declared workspace when the value is a resource reached over the network, such as a host, an IP address, a port, or a database; this repository is enough for a value only it uses, such as an internal route or a local key. Each hit is a consumer to account for.
2. **Signal: a consumer runs outside this repository's release, or the system runs on several hosts or processes.** Map which host or process runs what, such as an API server and a separate cron server, and which of them read and which write the data the change touches. Use read-only sources: infrastructure state, process lists, configuration, and project guidance.
3. **Signal: the change touches authentication, permissions, or account data; makes reachable code that was disabled, commented out, or unreachable; or creates or migrates records that existing checks read.** Read the code paths the change makes reachable and state what they allow once it ships. List a defect in such a path in the Impact list below as part of this change, since the change makes it reachable.
4. **Signal: the repository cannot rule out a consumer, such as an external client, another server, or a periodic job.** Mark that consumer unverified. Name the read-only observation that would settle it, such as connection or access logs or database counters, a window long enough to cover the system's known periodic jobs, and who can run it, and state how to revert the change. Until that observation runs, do not describe the change as safe or quick.

## Report the impact

Where the finding or plan states the recommendation, add a short Impact list: each consumer or path found, its evidence, what the change does to it, and what remains unverified. In a plan it goes in the verified context, and each unverified item becomes a task or a blocker. When you delegate part of the investigation, pass this reference and name the change in the brief, so the child looks for consumers instead of only answering a narrower question.
