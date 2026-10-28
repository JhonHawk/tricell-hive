# Language

## When to read

Read before naming an identifier, an i18n key or a route, writing an example or template, contributing to an external or open-source repository, or resuming a record or relaying a notification written in another language. The global guidance keeps the invariants: human-facing text in the session language, source files in English. Reading this reference does not invoke a workflow.

## Human-facing text

Write human-facing findings, progress records, and reports in the session language unless the user requests otherwise. This includes new or updated headings, status labels except the context line's stage, which stays untranslated, status values, option labels of questions, next steps, and the labels the host shows the person while you work, such as a command's description, a task's subject, and a child's label. Apply this when resuming a record written in another language, even when agents also use that report for continuity, and to every message that relays a child's or a tool's notification written in another language.

## Source files

Write everything that goes into source files in English, in a new project and in any session language: identifiers (variables, functions, types, API paths and properties, URL paths and route segments including those of public pages, database objects, event keys), file and directory names, code comments and docstrings, test names, log and developer-facing error messages, and identifiers introduced in specifications. The session language governs the conversation and human-facing records, never source code. Commit messages and branch names follow the repository's convention, and English when it has none. For a contribution to an external or open-source repository, check that repository's language first, in its contributing guide and recent history, and use it for everything that reaches it, such as commits, pull-request and issue text, review replies, and code comments; the conversation and the host's labels keep the session language. Preserve the project's casing and layout conventions and tool-required names, and choose established domain terminology rather than literal translations.

Write placeholders in examples and templates in English, even within session-language prose, such as `<project>-specs/openspec/specs/<capability>/spec.md`.

Separate identifiers from localized content: use English semantic i18n keys; user-facing text follows the product language, and a URL path is an identifier, not user-facing text. Preserve established domain values and externally defined or persisted identifiers when compatibility requires them, keep each domain vocabulary consistent, and rename contracts or migrate values only under an authorized compatibility plan.
