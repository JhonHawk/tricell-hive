# Diseño

## Enfoques

- **A — Sección en la guía global (elegido, D1 A).** «Test accounts» en `content/guidance/global.md`, de ~1,3 KB, más punteros de una línea en los flujos y en los roles que inician sesión. Se eligió porque la regla actúa en build, verify, walkthrough y deploy, y también fuera de los flujos (un «prueba este endpoint» suelto), y la spec `context-delivery` pone ahí toda regla de alcance general. Costo: la guía global crece ~4 %.
- **B — Referencia propia** (`flow-verify/references/test-accounts.md`) con condiciones de lectura desde la guía global y desde cuatro skills. La guía global crece menos, pero la regla depende de que el modelo abra la referencia; así se perdió #284.
- **C — Solo por proyecto**: una plantilla y una línea en cada `AGENTS.md`. No alcanza a los proyectos que no la tengan.

## Forma propuesta

### «Secrets» (`global.md:38-39`)

La ubicación por defecto se fija en su único lugar canónico, y la lectura del índice entra como excepción con nombre:

- Línea 38: «The only stores are the project's declared untracked secret location, `_support/secrets/` in its support home unless the project declares another, and a disposable session copy, both mode `0600`.»
- Línea 39, cláusula de lectura: «Never read a secret file with a tool that returns its content or matching lines, except the account index `ACCOUNTS.md` that «Test accounts» sets, which holds no values; …» (el resto de la línea no cambia).

### «Test accounts» (nueva, entre «Secrets» y «Evidence»)

Texto en inglés; el build puede ajustar la redacción, pero no los criterios:

```markdown
## Test accounts

- Before a check through the real path that signs in or uses records the environment already holds, in any flow or outside one, read the project's account index, `ACCOUNTS.md` in the secret location «Secrets» names; in a worktree, read and write the main checkout's file. When the file is missing, rebuild it from the persistent-memory entry `accounts/index` when one exists. Reuse an account whose environment, role and reach fit the check.
- The index holds one row per account: environment, role, identifier, secret source (a file in the secret location or a password-manager reference, never a value), what it reaches, such as an organization or the records it sees, and its status with the date it was last verified working or failing. Using a manager reference follows «Secrets». Keep the file at mode `0600`; inside a Git repository that does not ignore the secret location, add that location to `.git/info/exclude` before the first write.
- In a local environment, one whose database and services run on this machine, create only the accounts and records the check needs that the index and the local data do not already provide, and add each account you create or verify to the index. When an account's sign-in fails, mark its row stale with the date and cause.
- In QA, a shared or cloud development database, and deployed environments, never create accounts, records or mock data on your own initiative: use what exists, add each existing account you use to the index, and ask before creating anything, as «Scope and authorization» sets for shared data.
- A child gets the rows and secret sources of the accounts it uses, never a value, and reports each account it created, found stale or used that the index lacks; the main thread updates the index.
- At a work item's close, copy the index's rows, without values, to persistent memory under `accounts/index`; this is the close write «Persistent memory» allows.
```

### Punteros

Cada uno es una condición de lectura que nombra la regla, sin repetir sus pasos:

| Archivo | Dónde | Texto |
| --- | --- | --- |
| `content/skills/flow-verify/SKILL.md` | paso 1, tras «exercise the changed behavior through its real path» | «, reading the global guidance's «Test accounts» before choosing an account» |
| `content/skills/flow-build/SKILL.md` | línea 22, tras «seeding local data» | «(read the global guidance's «Test accounts» first when the data includes accounts)» |
| `content/skills/flow-walkthrough/SKILL.md` | línea 10, tras «an account that reaches the changed screens» | «, chosen after reading the global guidance's «Test accounts»» |
| `content/skills/flow-deploy/references/deployment-checks.md` | línea 32, antes de «Before it, ask for the effects» | «Before choosing the walk's accounts and records, read the global guidance's «Test accounts».» |
| `content/skills/flow-close/SKILL.md` | paso 7 | «Write the outcome to persistent memory, and the account index rows as the global guidance's «Test accounts» sets.» |
| `content/skills/flow-plan/SKILL.md` | línea 46 («Prerequisites») | «Before listing an account or test data, read the global guidance's «Test accounts».» |
| `content/agents/quality/hive-verify-change.md` | paso 2 | «Before signing in, read the global guidance's «Test accounts»; report accounts to the parent instead of writing the index.» |
| `content/agents/review/hive-review-ux.md` | paso 1 | igual que la fila anterior |
| `content/skills/workspace-conventions/references/artifact-layout.md` | «Select the destination» | una línea `secrets/`: sin versionar, modo `0600`, con los archivos de secretos y el índice `ACCOUNTS.md` que fija «Test accounts» |

La excepción de los roles («report accounts to the parent instead of writing the index») solo aplica a ese punto de uso y nombra el elemento de la regla que cambia, como pide `AGENTS.md`.

### Detector S5 (`tests/pilot/regression.go`)

`noSecretContentRead` (líneas 916-945) acepta una lectura completa de `<…>/_support/secrets/ACCOUNTS.md`. `isSecretPath` no cambia, así que preparar el commit o publicar ese archivo sigue contando como secreto.

## Impacto

- Lo usan todos los hosts que cargan la guía global; los punteros llegan a quien abra esos skills y roles.
- No toca producción ni contratos de código. El único código que cambia es el detector de pilotos, que hoy está en pausa y corre solo como test.
- Rama: la actual de gh-311 (PR #352). Los punteros se anclan a su texto actual, y todas las frases de anclaje aparecen una sola vez en el árbol de trabajo (verificado por la revisión del plan).
- `global.md` pasa de 31 122 bytes (árbol de trabajo en `ee1ebf07`) a unos 32 500; el PR informa la cifra exacta.

## Recuperación

Se revierten los commits de este cambio; ningún índice ni dato de proyecto se crea en este trabajo.
