# Diseño · #192

## Contexto verificado (base `b3edd8f4`, 2026-10-26)

**Auditoría de cam, carvento y facti** (tres `hive-review-harness`; hallazgos principales comprobados en la fuente):
- **cam:** `cam-monorepo/AGENTS.md:24` da un bypass administrativo permanente a `qa` «including … data changes», y un push a `qa` corre `prisma migrate deploy` (`.github/workflows/deploy-backend.yml`). Los `Specs` relativos no llevan URL y se rompen en los worktrees.
- **carvento:** `carvento-monorepo` quedó sin `Delivery`, `Review` ni `UI verification` (382b8dfdb retiró Bugbot). `AGENTS.md:53` mueve el ticket a In Progress «when you pick it». Hay otra frase ambigua sobre la etiqueta `production-approved`.
- **facti:** solo `facti-2020-new` tiene sección, y `Specs: ../facti-specs` no incluye `/openspec`. La raíz manda planear en `facti-specs/sessions/`, sin uso desde el 2026-09-17, mientras `openspec/changes/` está activo. El `CLAUDE.md` de reforge es una plantilla anterior a Hive.
- **En los tres:** casi todas las sesiones arrancan en la raíz del workspace.

**Hive:**
- **Catálogo de claves:** `content/skills/flow-plan/references/project-settings.md:9`, con el ejemplo en `:11-23`, y «Later completion point» en `:25-29`.
- **Valores declarados que autorizan:**
  - `content/guidance/global.md:69` (`Review`/`Delivery` declarados) y `:35` («declared or answered»);
  - `delivery-decisions.md:5,20,22,26,34,36,44`;
  - `unattended-delegation/SKILL.md:3,8`;
  - `flow-build/SKILL.md:24`, `flow-verify/SKILL.md:22,29-33`;
  - `ui-gate.md:3,19,44`, `review-and-gates.md:9,30`, `flow-walkthrough/SKILL.md:34`;
  - `plan-format.md:136`, `flow-close/SKILL.md:78`, `reporting.md:98`;
  - `security-boundaries.md:14`, `verification.md:27`, `flow-plan/SKILL.md:22`.
- **Punto de terminado en prosa:** `global.md:34`, `flow-close/SKILL.md:14`, `flow-deploy/SKILL.md:25`, `reporting.md:102`.
- **Manager:**
  - Tabla única de claves `hiveSettingKeys` en `tooling/cli/doctor_project.go:30-40`.
  - Valores de `Delivery` en `:477` y de UI en `:479`; casos por clave en `hiveFindings` `:318`.
  - `runDoctor` devuelve `nil` siempre (`tooling/cli/main.go:128-143`).
  - El formulario está en `tui_project_view.go:299`, y `validateHiveRequest` en `project_write.go:450` rechaza claves desconocidas al hacer `--set`.
  - Dos pruebas Go leen `project-settings.md` («Declare settings»): `TestHiveSettingKeysAppearInTheGuidance` y `TestHiveSettingKeysCoverTheGuidanceExample` (`doctor_project_test.go:440,460`).
- **`harness-audit`:**
  - Es de solo lectura (`SKILL.md`, «Boundaries»), salvo una migración de layout aprobada.
  - Ya pregunta las decisiones en el hilo principal («Findings format»).
  - HA-IF-17 vive en `references/instruction-files.md:37`, con procedencia en `_support/docs/harness-engineering/harness-audit-rules.md:54` y prueba en `tests/skills/harness_audit_rules_test.py`.
- **Despliegue de skills:** sin manifiesto; `tooling/management/plan.go:230` toma todo `content/skills/*/SKILL.md`.

## Diseño elegido

### Formato de la sección

La sección es una ficha de hechos que el agente puede verificar:

```markdown
## Hive
- Project: ark (workspace `..`)
- Base branch: development
- Environments: development → qa → production
- Tracker: Linear · team sample · project ark · key ARK
- Specs: ../ark-specs/openspec · https://github.com/<org>/ark-specs.git
- Completion: To Deploy on origin/development · Done on origin/qa
```

- **`Delivery: direct-base`** sigue siendo la única forma de `Delivery`. Marca un repositorio de soporte y conserva sus efectos (`delivery-decisions.md:18`). `project-settings.md` dice que es la única clave que autoriza algo, porque describe el tipo de repositorio y no una preferencia.
- **`Completion`** lleva el estado intermedio y el final, cada uno con la ref que lo marca. «Later completion point» cambia solo su condición, que pasa a ser «the section declares `Completion`». Sus acciones y el ancla `#later-completion-point` se conservan, porque `flow-close`, `flow-deploy`, `reporting.md` y la página `deploy` del sitio enlazan ahí.
- **Claves retiradas** (`Review`, `UI verification`, `Delivery: automatic|interactive`): una sección que aún las tenga no autoriza nada. El agente las ignora y pregunta, `hive doctor` avisa y `harness-audit` propone quitarlas.
- **Permisos permanentes en prosa** (bypass, deploys preautorizados): no tienen clave. Valen como instrucción del proyecto mientras no contradigan una regla de efecto crítico, y `harness-audit` los señala para que el usuario decida.

### Reglas de entrega, revisión y UI

La ronda de `delivery-decisions.md` se pregunta siempre que no la resuelva:
- una respuesta anterior de la sesión para el mismo repositorio y rama, según la regla de reutilización de `global.md:29`, que además nombrará `UI verification`;
- un pedido más estrecho, como `hold`;
- un repositorio `direct-base`.

**Se conserva:**
- la línea antes de la primera edición, que ahora resume las respuestas reutilizadas y lleva la oferta de revisión de seguridad cuando el cambio cruza un límite de seguridad, también cuando la ronda no se pregunta porque todo se reutiliza (`delivery-decisions.md:26`);
- la regla que vuelve `automatic` no disponible cuando el proyecto exige validación humana;
- la recomendación de un mecanismo que la prosa de `AGENTS.md` prefiera (`delivery-decisions.md:30,34`), que informa la opción Recommended sin autorizarla;
- que una respuesta de revisión nunca cuenta como pedir la revisión de seguridad.

`unattended-delegation` se activa con los valores respondidos.

### `harness-audit` audita, pregunta y aplica

Procedimiento nuevo, después del reporte:

1. Pregunta en una ronda las decisiones que sus hallazgos necesitan: qué copia de una regla duplicada se conserva, qué permisos en prosa se quitan, los hechos de la sección que no pudo deducir y qué resultados aplicar. Usa [questions](../../../../content/skills/flow-plan/references/questions.md). Cuando lo aplicado va a terminar en una entrega Git en un repositorio de código, la misma ronda incluye la pregunta de entrega y revisión de [delivery decisions](../../../../content/skills/flow-plan/references/delivery-decisions.md), que esa regla exige antes de la primera edición. `hold` (las ediciones quedan sin commit) es una opción.
2. Tras la respuesta, en el proyecto actual y con la edición del host, aplica los resultados aceptados: `delete`, `relocate`, `stale`, `demote` con su puntero, `insert` y la sección `## Hive`. Los `manual` los deja escritos como tarea, y `propose-global` y `CLAUDE.local.md` no se tocan nunca. Lee [change records](../../../../content/skills/flow-plan/references/change-records.md) cuando escribe `Specs`.
3. Valida cada sección escrita con el código de salida de `hive doctor --project <repo>` (0) y la ausencia de avisos de claves retiradas. Si `hive` no está en el `PATH`, relee la sección contra `project-settings.md` y lo reporta.
4. Los commits, pushes y PR siguen el modo elegido en el paso 1 a través de `flow-build`.

Bajo [unattended delegation](../../../../content/skills/unattended-delegation/SKILL.md), las decisiones sobre permisos en prosa (D3-A) y sobre qué copia de una regla se conserva quedan sin aplicar y se reportan, porque son del usuario.

Además cambian el párrafo inicial de `SKILL.md` («Report findings with evidence; do not edit»), «Scope» («a read-only audit needs neither»), «Boundaries» («Stay read-only») y `project-layout.md:3`.

**Límites:**
- Un rol de solo lectura (`hive-review-harness`) devuelve hallazgos y decisiones, y el padre pregunta y aplica.
- Aplicar en otro proyecto exige correr la skill en la sesión de ese proyecto, según la regla de alcance global.
- La migración de layout aprobada sigue su camino actual por `workspace-conventions`.

**Descripción:** añade el disparador «set up or update a repository's `## Hive` section» (onboarding).

**HA-IF-17 suma tres comprobaciones:**
- las claves retiradas son `stale`, con propuesta de quitarlas;
- `Completion` debe tener la forma;
- un punto de terminado en prosa es `relocate` a `Completion`, y los permisos permanentes en prosa salen como decisión del usuario.

Se conservan `relocate` para valores repetidos fuera de la sección y `medium` para depender de la sección del workspace.

### Manager

- **Claves:** `hiveSettingKeys` queda con Project, Base branch, Tracker, Specs (obligatorias), Environments, Delivery, Completion y Hive guidance, más una lista `retiredHiveKeys` de Review y UI verification. `knownHiveKey` las excluye, para que no salgan además como «unknown key» ni en «Known keys:». `Delivery` acepta `direct-base`; `automatic` e `interactive` son valores retirados y dan aviso; cualquier otro valor bloquea. Una prueba acopla `retiredHiveKeys` con la frase de claves retiradas de `project-settings.md`, como las dos de acoplamiento que ya existen.
- **Avisos de doctor:** el aviso de una clave o valor retirado es `<Key>: no longer read; Hive asks it each session — remove it with hive project set --unset <Key> or run harness-audit`. `Completion` avisa, sin bloquear, si no tiene la forma `<state> on <ref> · <state> on <ref>`.
- **Código de salida de doctor:** sale distinto de 0 cuando el directorio está en un repositorio Git y su sección `## Hive` existe con un hallazgo bloqueante:
  - valor obligatorio ausente o vacío;
  - `Delivery` o `Hive guidance` inválidos;
  - sección duplicada;
  - un `AGENTS.md` que existe pero no se puede leer, no es un archivo regular o excede el límite.

  Sin `AGENTS.md` o sin sección, el código depende de cómo se llame (D10-A):
  - con `--project` explícito sale distinto de 0, para que `harness-audit` detecte una sección que no quedó escrita; lo mismo cuando la ruta no existe, no está en un repositorio Git o git no puede ejecutarse (D11-A);
  - sin el flag sale con 0, para que diagnosticar la instalación en cualquier repositorio no falle. También sale con 0 con solo avisos o fuera de un repositorio. La marca `Blocking` viaja de `checkProject` por `collectProject` hasta `doctorSection` (campo nuevo) y `runDoctor`, que devuelve el error `Project section has blocking findings`. `main` lo imprime como `hive: Project section has blocking findings` en stderr, y el informe en stdout no cambia. Se actualiza el comentario de `readOnlyOptions` (`main.go:100`).
- **`hive project set`:** rechaza `--set` de una clave retirada, o de `Delivery` distinto de `direct-base`, con `<Key>: retired; Hive asks it each session — use --unset <Key> to remove it`, y permite `--unset`.
- **Formulario:** ocho campos fijos.
  - `Delivery` cicla entre vacío y `direct-base`; un valor heredado se muestra marcado `(retired)` y se puede llevar a vacío.
  - Cada clave retirada presente en la sección suma una fila de elección {valor actual `(retired)`, `remove`}, y `remove` produce `unset`.
  - `Completion` es texto, con placeholder `<state> on origin/<first environment> · <state> on origin/<last environment>` cuando `Environments` existe y la forma genérica si no.

**Origen de cada valor del formulario:**

| Clave | Origen |
| --- | --- |
| Project | sugerido: remoto o carpeta (`suggestHiveValues`) |
| Base branch | sugerido: `origin/HEAD` si existe |
| Tracker | sugerido: GitHub Issues del remoto; si no, se teclea |
| Specs | sugerido: `_support/openspec` si existe; si no, se teclea con la URL del repo de specs |
| Environments | se teclea; no hay fuente fiable |
| Delivery | se elige: vacío o `direct-base` |
| Completion | se teclea, con placeholder deducido de `Environments`; `harness-audit` lo propone desde la prosa al reubicarlo |
| Hive guidance | se elige: vacío o `required` |
| Clave retirada presente | se elige: conservar o `remove` |

### Requisitos

Se modifican «Read-only diagnostics» y «Project settings section editing» de `versioned-installation` (ver `specs/versioned-installation/spec.md`). Las reglas de guía no tienen capability propia, como en #180.

### Alternativas descartadas

- **Dos bloques, hechos y decisiones, dentro de la sección:** el usuario decidió que las decisiones se pregunten (D1).
- **Una skill nueva, `project-setup`:** duplicaría a `harness-audit` como escritor del mismo archivo (D9).
- **Leer el texto de `doctor` en lugar del código de salida:** es frágil para los modelos de ejecución (D6).
