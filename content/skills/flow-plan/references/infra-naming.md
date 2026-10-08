# Infrastructure naming

Read this reference when planning or implementing a change that creates, chooses, or materially changes an infrastructure resource name, including resources whose names existing infrastructure code generates. Treat the project's existing naming convention and the relevant provider or platform restrictions as the primary source. Consult current official provider documentation when a restriction affects the proposed name.

## Default convention

When the project has no applicable convention or provider-imposed shape, use:

```text
<project>-<component>-<environment>
```

Use kebab-case. The default environment vocabulary is `development`, `qa`, and `production`. This is a fallback convention, not a requirement to rename existing resources, create a naming document, or impose environment branches.

## Justified exceptions

Use a different shape only when the resource's role, the platform, or an established project convention justifies it. Record the reason in the plan or implementation evidence when it affects a durable interface. Common cases include:

| Resource or surface | Usual consideration |
| --- | --- |
| Shared resource | It may intentionally have no environment segment. |
| Secret store | The platform may use a hierarchical path or a service-specific identifier. |
| Database | The engine or existing schema may require or favor a different separator or identifier form. |
| Repository or image | `<project>-<component>` with no environment segment, such as `acme-marketplace-backend`; tags, registries, or deployment configuration carry the environment. Before proposing a repository name, list the organization's repositories and follow the pattern most comparable ones share, including suffixes such as `-monorepo`. For a copy or variant of an existing product, such as a fork for one client, the comparable repositories are the organization's other variants, such as `<product>-<client>`, not only the product's own repositories. |
| DNS | Public hostnames commonly put an environment label in a provider- or user-facing position, while the production hostname can omit it. |

Do not force the default token order onto an exception. Do not invent abbreviated environments, limits, or naming constraints without checking the applicable current documentation.

## Names the project already defines

Look for the project's recorded naming convention, such as a naming table in its specifications documentation (for example `docs/conventions/naming.md`), before choosing a name. When the work touches infrastructure code that already generates names, such as Terraform modules or deployment configuration, compare the names that code produces with that convention rather than copying them into the plan or the change:

- When a recorded convention exists, report each mismatch as a finding that names the resource, the generated name, and the conventional one, and add each new name the change introduces to that record.
- When no recorded convention exists, compare with this reference's default and present a mismatch as a recommendation, not a defect. Recommend recording the chosen names; create that record only when the user agrees.

Whether to rename, keep, or record a mismatch as a project exception is the user's decision. Renaming a resource that does not exist yet costs an edit; renaming an existing one follows the next section.

## Existing resources

Preserve existing names unless a rename is explicitly within scope. A rename can affect identifiers, data, references, permissions, and deployment configuration; evaluate its compatibility impact, migration path, and recovery or rollback needs before proposing it. Naming guidance does not authorize creating infrastructure, changing repository topology, or creating a separate specifications repository.
