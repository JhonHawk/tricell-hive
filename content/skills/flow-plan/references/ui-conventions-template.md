# Adaptable project UI guide

Use when creating or restructuring durable UI conventions under [UI planning](ui-planning.md). Preserve an existing equivalent guide. This is agent-facing guidance: write it in English unless the user requests otherwise. Keep human plans and reports in the session language.

Include only established decisions and material open questions; replace or omit placeholders before delivery. Start small, with sections for the affected applications/areas. Link executable examples and canonical token/component sources rather than copying their inventories or API definitions. A proposed pattern is not an accepted one.

```markdown
# UI conventions

## Scope and sources

Applications/areas covered: <paths and user contexts>.
Shared foundations: <links to applicable tokens/components/design sources>.
Application-specific differences: <what intentionally varies and why>.
Maintainer/review path: <existing owner or project process>.

## <Application or area>

Shells: <name, source path, applicable routes/context>.

| Pattern | Status | Use when | Reference and reusable composition | Allowed differences |
| --- | --- | --- | --- | --- |
| <pattern> | <proposed / accepted / deprecated> | <task/context> | <source and rendered/story/design example> | <variants and rationale> |

Interaction conventions: <relevant actions, filtering, pagination, forms,
feedback and state handling; link existing guidance>.
Responsive/accessibility expectations: <applicable behavior and references>.

## Decisions and exceptions

<Consequential departures, affected application/pattern, rationale,
decision status and replacement when deprecated. Omit when none.>

## Maintenance

<Update the relevant convention and example when implementation changes;
identify affected consumers of shared components. Record meaningful source
revision or verification context where freshness matters.>
```

Separate accepted intent from observed implementation when they differ. Screenshots can illustrate an example but do not replace source links, interaction criteria, or verification. Do not require a screenshot archive, Storybook installation, or generated manifest solely to fill this template.
