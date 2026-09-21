# Infrastructure naming

Read this reference when planning or implementing a change that creates, chooses, or materially changes an infrastructure resource name. Treat the project's existing naming convention and the relevant provider or platform restrictions as the primary source. Consult current official provider documentation when a restriction affects the proposed name.

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
| Repository or image | The repository can be environment-neutral while tags, registries, or deployment configuration carry the environment. |
| DNS | Public hostnames commonly put an environment label in a provider- or user-facing position, while the production hostname can omit it. |

Do not force the default token order onto an exception. Do not invent abbreviated environments, limits, or naming constraints without checking the applicable current documentation.

## Existing resources

Preserve existing names unless a rename is explicitly within scope. A rename can affect identifiers, data, references, permissions, and deployment configuration; evaluate its compatibility impact, migration path, and recovery or rollback needs before proposing it. Naming guidance does not authorize creating infrastructure, changing repository topology, or creating a separate specifications repository.
