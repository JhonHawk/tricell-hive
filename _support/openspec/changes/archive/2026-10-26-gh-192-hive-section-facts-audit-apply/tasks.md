# Tareas · #192

Commit base: `b3edd8f4` (`origin/development`, 2026-10-26). Rama `feat/gh-192-hive-section-facts` en el worktree `../tricell-hive-worktrees/gh-192`.

## Tareas

### T1 — El catálogo de claves describe solo hechos

- [x] `project-settings.md` define el conjunto nuevo de claves, nombra las retiradas, actualiza el ejemplo y pone la condición de «Later completion point» sobre `Completion`.

**Closes:** AC1, AC2

**Checkpoint:** C1

**Depends on:** none

**Locations:** `content/skills/flow-plan/references/project-settings.md` («Declare settings» `:9-23`, «Later completion point» `:25-29`, «Resolve a missing value» `:37-41`).

**Execution:** hilo principal. Es redacción de guía, y T2 a T6 se apoyan en su texto exacto.

**Test approach:** check. `rg -n '`Review`|`UI verification`|automatic|interactive' content/skills/flow-plan/references/project-settings.md` solo debe encontrar la frase de claves retiradas. La lectura del ejemplo no debe mostrar claves retiradas.

**Changes:**
- Aplicar «Formato de la sección» de [design.md](design.md#formato-de-la-sección).
- Añadir un puntero en «Resolve a missing value» hacia `harness-audit`, para crear o actualizar la sección completa.
- Conservar el ancla `#later-completion-point`.

**Verification:** el `rg` anterior, y `grep -n 'Completion' …/project-settings.md` muestra la clave entre comillas invertidas en la prosa y en el ejemplo. Las dos pruebas Go que leen este archivo corren en T2.

### T2 — El manager valida y edita el conjunto nuevo de claves

- [x] `hive doctor`, `hive project set` y la vista Project aplican «Manager» de [design.md](design.md#manager).

**Closes:** AC4, AC5, AC6

**Checkpoint:** C2

**Depends on:** T1

**Locations:**
- `tooling/cli/doctor_project.go`: `hiveSettingKeys` `:30-40`, `hiveFindings` `:318`, `hiveDeliveryValues` `:477`, `hiveUIVerificationValues` `:479`.
- `tooling/cli/doctor_project.go`: `checkProject` `:142`, `collectProject` `:116`, `knownHiveKey` `:431`, `validateHiveSection` `:382`.
- `tooling/cli/doctor.go`: `doctorSection` y `collectDoctor` (`:170,209`).
- `tooling/cli/main.go`: `runDoctor` `:128` y el comentario de `readOnlyOptions` `:100`.
- `tooling/cli/project_write.go`: `validateHiveRequest` `:450`.
- `tooling/cli/tui_project_view.go`: `newProjectForm` `:299` (incluida la rama de opciones de UI `:323-324`), `request()` `:448`, `formView` `:535`, la etiqueta especial `:567` y `projectLabelCol` `:267`.
- Las pruebas de `doctor_project_test.go`, `doctor_command_test.go`, `project_write_test.go`, `project_command_test.go` y `tui_project_view_test.go`:
  - de doctor: `TestHiveSettingKeysRequiredSet`, `TestHiveSectionUnknownKeyAndOptionalValues`, `TestDoctorProjectAcceptsDeliveryModes`, `TestDoctorProjectChecksUIVerificationValues`, `TestHiveTextSeparatesBlockingFromWarnings`, `TestDoctorCommandProjectFlagChecksThatRepository` (`doctor_command_test.go:63-88`);
  - de `project set` y edición: `TestProjectCommandUsageErrors` (`project_command_test.go:114,130`), `TestHiveSectionEditRejections`, `project_write_test.go:48,129-137`;
  - del formulario: `formRows` y las pruebas que dependen del índice de fila.
- `_support/docs/architecture/deployment-manager.md` (`:71`, `:77`, `:110`, `:230`).

**Execution:** delegada a `hive-build-backend`. La interfaz quedó fijada en el diseño, los archivos son disjuntos de T3, y corre en paralelo con T3.

**Test approach:** tdd. Cambia un contrato público (código de salida, claves aceptadas) y la lógica de validación. Comportamientos a cubrir:
- un aviso por cada clave retirada;
- `Delivery: direct-base` válido, `automatic` con aviso y `pr` bloqueante;
- `Completion` conocido, con aviso cuando no tiene la forma;
- `runDoctor` sale distinto de 0 con una pleca existente con hallazgo bloqueante (obligatorio ausente, `Delivery` inválido, sección duplicada) y devuelve el error `Project section has blocking findings`;
- sin `AGENTS.md` o sin sección, `runDoctor` sale distinto de 0 con `--project` explícito y con 0 sin el flag (D10-A). Con solo avisos o fuera de un repositorio sale con 0. Se ajusta `TestDoctorCommandProjectFlagChecksThatRepository`, que pasa `--project` a un repositorio sin `AGENTS.md`, porque su resultado esperado cambia por esta decisión;
- las claves retiradas no salen como «unknown key», y su frase en `project-settings.md` coincide con `retiredHiveKeys`, verificado por una prueba de acoplamiento;
- `--set Review` y `--set "Delivery: automatic"` se rechazan como retirados y `--unset` se permite. `validateHiveRequest` comprueba las retiradas antes que las desconocidas, para que el mensaje diga «retired»;
- sección duplicada y `AGENTS.md` de más de 1 MiB salen distinto de 0, porque el exceso de tamaño pasa de aviso a bloqueo;
- el formulario tiene ocho campos fijos, `Delivery` ∈ {vacío, `direct-base`} con el heredado marcado `(retired)`, una fila `remove` por cada clave retirada presente que produce `unset`, y el placeholder de `Completion` deducido de `Environments`.

Las pruebas que solo verificaban valores retirados se eliminan tras comparar sus escenarios.

**Changes:** según el diseño. Actualizar `deployment-manager.md` en las mismas líneas.

**Verification:** `go test ./tooling/cli` pasa, incluidas `TestHiveSettingKeysAppearInTheGuidance` y `TestHiveSettingKeysCoverTheGuidanceExample` contra el `project-settings.md` de T1. Luego, con el binario compilado en un repositorio desechable del scratchpad:
- una sección con `Delivery: automatic` y `Review: x` muestra los dos avisos y sale con 0;
- una sección sin `Tracker` sale distinto de 0, con `hive: Project section has blocking findings` en stderr;
- en un repositorio sin `AGENTS.md`, `hive doctor --project <repo>` sale distinto de 0, y `hive doctor` sin flag desde ese repositorio sale con 0;
- fuera de un repositorio sale con 0;
- `hive project set --set "Delivery: automatic" --dry-run` se rechaza nombrándolo retirado.

### T3 — Las reglas preguntan siempre la entrega, la revisión y la UI, y leen `Completion`

- [x] Ninguna regla distribuida trata un valor declarado retirado como respuesta o autorización, y el punto de terminado se cita como la clave.

**Closes:** AC2, AC3

**Checkpoint:** C3

**Depends on:** T1

**Locations:**
- `content/guidance/global.md` `:29` (añadir `UI verification` a la reutilización), `:34`, `:35`, `:69`.
- `content/skills/flow-plan/references/delivery-decisions.md` `:5,18,20,22,26,34,36,44`.
- `content/skills/flow-plan/SKILL.md:22`.
- `content/skills/flow-plan/references/plan-format.md:136` y `plan-review.md:16`.
- `content/skills/flow-build/SKILL.md:24`; `references/verification.md:27`; `references/security-boundaries.md:14`.
- `content/skills/flow-verify/SKILL.md:22,29-33,37,41`; `references/ui-gate.md:3,19,44`; `references/review-and-gates.md:9,30`.
- `content/skills/flow-walkthrough/SKILL.md:34`.
- `content/skills/flow-close/SKILL.md:14,78`; `references/reporting.md:98,102`.
- `content/skills/flow-deploy/SKILL.md:25`.
- `content/skills/unattended-delegation/SKILL.md:3,8`.

**Execution:** hilo principal. Es redacción acoplada entre reglas con un solo hogar cada una, y la regla de AGENTS.md exige revisar todo el comportamiento.

**Test approach:** check.
- `rg -n 'declared `(Review|Delivery)`|`UI verification` value declared|declared in its repository|declared or answered|declares (them|a later completion)|completion point a project declares|values declared in|declared settings|repository.s settings|whose value it does not declare' content` solo deja menciones a `direct-base`.
- Leer cada línea cambiada contra el diseño (se conserva la línea previa a la primera edición, la regla de `automatic` no disponible, la preferencia en prosa y la exclusión de la revisión de seguridad).

**Changes:** según «Reglas de entrega, revisión y UI» de [design.md](design.md#reglas-de-entrega-revisión-y-ui). Medir el tamaño de `global.md` antes y después para el PR.

**Verification:** el `rg` anterior, más `go test ./tests/content`, que cubre tamaños, frontmatter, enlaces de `global.md` y la huella de lectura.

### T4 — `harness-audit` audita, pregunta y aplica, incluida la sección `## Hive`

- [x] La skill pregunta las decisiones de sus hallazgos, aplica lo aceptado en el proyecto actual, valida la sección con `hive doctor` y HA-IF-17 cubre las claves retiradas, `Completion` y las decisiones en prosa.

**Closes:** AC7, AC8

**Checkpoint:** C4

**Depends on:** T1, T2

**Locations:**
- `content/skills/harness-audit/SKILL.md` (descripción, párrafo inicial `:8`, «Scope» `:12`, «Procedure», «Findings format», «Boundaries»).
- `content/skills/harness-audit/references/instruction-files.md:37` (HA-IF-17).
- `content/skills/harness-audit/references/project-layout.md:3`.
- `content/agents/review/hive-review-harness.md` (punto 4: devolver las decisiones; el padre aplica).
- `_support/docs/harness-engineering/harness-audit-rules.md:54`.

**Execution:** hilo principal. Redacción de skill con límites de autoridad.

**Test approach:** check.
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.
- `go test ./tests/content ./tooling/management` pasa (tamaño ≤ 16 000 bytes, frontmatter, enlaces de recursos).
- Leer «Boundaries» confirma que el rol sigue sin escribir y que `propose-global` y `CLAUDE.local.md` quedan excluidos.

**Changes:** según «`harness-audit` audita, pregunta y aplica» de [design.md](design.md#harness-audit-audita-pregunta-y-aplica). No se agregan IDs HA nuevos, para no tocar la prueba del catálogo más allá de la fila de procedencia.

**Verification:**
- las pruebas de arriba;
- `wc -c content/skills/harness-audit/SKILL.md` menor que 16 000;
- `rg -n 'Review|UI verification|Completion|relocate' content/skills/harness-audit/references/instruction-files.md` muestra en HA-IF-17 las claves retiradas como `stale`, la forma de `Completion` y la prosa del punto de terminado como `relocate`;
- `rg -n 'hive doctor --project|propose-global|delivery|unattended|read-only' content/skills/harness-audit/SKILL.md` muestra la validación por código de salida, la exclusión de `propose-global` y `CLAUDE.local.md`, la ronda de entrega, la regla bajo delegación y el rol de solo lectura;
- la descripción nombra la configuración de la sección `## Hive`;
- `rg -n 'do not edit|read-only audit needs neither|Stay read-only' content/skills/harness-audit` no encuentra nada.

El comportamiento con modelo no se verifica (pilots en pausa); queda registrado como límite.

### T5 — La sección de este repositorio declara solo hechos

- [x] `AGENTS.md` quita `Review` y `Delivery` de `## Hive`, deja `/code-review` como preferencia en prosa y reescribe `:59`.

**Closes:** AC9

**Checkpoint:** C5

**Depends on:** T2, T3

**Locations:** `AGENTS.md:7-15` (`## Hive`), `:59` («Scope and preservation»).

**Execution:** hilo principal. Son dos líneas y una frase.

**Test approach:** check, con el binario de T2.

**Changes:**
- Añadir en «Scope and preservation» que la revisión de código preferida es `/code-review`, como preferencia que informa la opción Recommended.
- Reescribir `:59` para que solo `Delivery: direct-base` cuente como instrucción.

**Verification:**
- `go run ./tooling/cli doctor --project . ; echo $?` imprime `Valid` en la sección Project, sin avisos de claves retiradas, y sale con 0.
- `rg -n '^- (Review|Delivery):' AGENTS.md` no encuentra nada.
- `rg -n '/code-review|counts as the user.s explicit instruction' AGENTS.md` muestra la preferencia en prosa, y la frase de `:59` nombra solo `Delivery: direct-base`.

### T6 — La documentación describe las claves nuevas y la aplicación de `harness-audit`

- [x] El sitio, el README, `llms.txt`, `workspace-and-artifacts.md` y el CHANGELOG quedan al día. (Se agregó `flows/plan.md:69`, que el implementador encontró fuera de la lista.)
  ronda 1: AC10 not met — `flows/build.md:45` aún decía «declared review setting»; corregido en `d8432ca6` junto con `build.md:48`, `verify.md:3`, `llms.txt:35` y la descripción de `flow-verify`.
  ronda 2: AC10 not met — `README.md:347` decía «the project's settings»; corrección local única aplicada como ronda 3 (`388599ff`).
  ronda 3: AC10 met.

**Closes:** AC10, AC2

**Checkpoint:** C5

**Depends on:** T1, T2, T3, T4

**Locations:**
- `site/src/content/docs/getting-started.md:35-39`; `flows/build.md:48,76`; `flows/verify.md:22,27,57`; `flows/security.md:19`; `flows/close.md:17,43,62,72`; `flows/deploy.md:6`; `skills.md:25,30`; `index.mdx:92,97`.
- `README.md:354,358,376`; `llms.txt:18,38,43`.
- `_support/docs/architecture/workspace-and-artifacts.md:23`.
- `CHANGELOG.md` `[Unreleased]`, con una nota de migración: quitar las claves retiradas o correr `harness-audit`; `hive doctor` sale distinto de 0 con una sección bloqueante.

**Execution:** delegada a `hive-write-docs`. Las fuentes ya están verificadas (T1 a T4, el inventario del 2026-10-26), los archivos son disjuntos, y las páginas describen el comportamiento sin copiar reglas.

**Test approach:** check.

**Verification:**
- `pnpm --dir site build` y `pnpm --dir site check` pasan.
- `rg -n '`Review`, `Delivery`|`UI verification`' site README.md llms.txt` solo deja usos como respuesta, no como clave declarable.
- `rg -n 'later completion point|Completion' README.md site/src/content/docs/flows/close.md site/src/content/docs/flows/deploy.md` muestra que cada mención del punto de terminado nombra la clave `Completion`.
- `workspace-and-artifacts.md:23` atribuye la sección a `project-settings.md`.
- `CHANGELOG.md` `[Unreleased]` trae la nota de migración (claves retiradas, `Completion`, código de salida de `hive doctor`, `harness-audit` aplica).

## Verificación y revisión humana

**Puntos de control:**
- C1 — T1 — T2, T3, T4 y T6 dependen de su texto.
- C2 — T2 — cambia un contrato público del manager (código de salida, claves).
- C3 — T3 — cambia las reglas de autorización de la guía, que son un contrato.
- C4 — T4 — cambia el límite de autoridad de `harness-audit`: pasa de solo lectura a escribir.
- C5 — T5, T6 — el resto, cerrado por T6.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | candidato final, local | `go vet ./...` y `go test ./...` pasan; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa | parte de la implementación |
| Sitio | candidato final | `pnpm --dir site build` y `pnpm --dir site check` pasan | parte de la implementación |
| In vivo del manager | tras C2, scratchpad | el binario compilado sobre repos desechables: avisos de claves retiradas, códigos de salida, `project set --set Review` rechazado y `--unset` permitido | parte de la implementación |
| UI `functional` (D8-A) | candidato final | `hive-verify-change` abre la vista Project del TUI en un repositorio desechable (identidad: el binario compilado del candidato; base `b3edd8f4`; terminal de 100×30): ocho campos, `Delivery` cicla entre vacío y `direct-base`, `Completion` acepta texto, guardar escribe lo mismo que `hive project set`, y una línea `Review` heredada se muestra y se puede quitar | respondida en D8 |
| Revisión dedicada | candidato local, antes del push | `/code-review` sobre el diff completo, según «Dedicated review timing» de `delivery-decisions.md` para `automatic` | declarada en `AGENTS.md` (vigente hasta este cambio) |

La skill con modelo (`harness-audit` aplicando) no se ejerce: las pilots están en pausa. Queda como límite declarado. El usuario puede probarla en la sesión de cam, carvento o facti.

## Estado de la revisión y progreso

**Revisión del plan** sobre la revisión `acfe2d01e1c7`, con tres `hive-review-plan` en paralelo:

| Dominio | Hallazgos | Costo |
| --- | --- | --- |
| Consistencia de guía | 5 bloqueantes, 3 no bloqueantes | 155k tokens, 175 s |
| Manager en Go, con el chequeo de orden | 4 bloqueantes, 3 no bloqueantes | 135k tokens, 144 s |
| Experiencia del formulario | 2 bloqueantes, 4 no bloqueantes | 110k tokens, 76 s |

Todos los hallazgos se aceptaron, como propuestas de los revisores o como decisiones menores (registradas en `proposal.md`):
- **Fila `remove` para claves retiradas.** El formulario solo recorría `hiveSettingKeys` (`tui_project_view.go:310`).
- **Registro del origen de cada valor del formulario.**
- **`--set Delivery` distinto de `direct-base` rechazado.**
- **Conjunto bloqueante y transporte de la marca hasta `runDoctor`.** Corregí mi propia extensión: sin `AGENTS.md` o sin sección sale con 0, según D6.
- **Mensajes con la acción para quitar claves retiradas, y el error en stderr.**
- **`knownHiveKey` excluye las retiradas, con una prueba de acoplamiento.**
- **Locations y pruebas faltantes de T2:** `doctor.go`, `doctor_command_test.go`, `deployment-manager.md:71`.
- **Locations de T3 y T4.** T3 suma `delivery-decisions.md:18` y amplía su `rg`. T4 suma el párrafo inicial y «Scope» de `harness-audit`.
- **`close.md:17` en T6.**
- **Verificaciones de AC2, AC7, AC8, AC9 y AC10.**
- **T4 cierra C4 y T5/T6 pasan a C5.**
- **`harness-audit` pregunta la entrega en su ronda.** Lo exige `delivery-decisions.md:5`.
- **Bajo delegación desatendida no aplica las decisiones del usuario.**
- **La oferta de revisión de seguridad se conserva en la línea previa a la primera edición.**

**Re-revisión del manager** (una ronda, 149k tokens, 39 s):
- Confirmó las correcciones: los llamadores de `doctorSection` no se rompen y `TestDoctorCommandProjectFlagChecksThatRepository` se mantiene.
- Pidió alinear AC5 con el spec, agregar un escenario para sección duplicada o archivo grande, y que las claves retiradas se comprueben antes que las desconocidas. Aplicado como propuesta suya.
- Dejó un hueco bloqueante: si doctor sale con 0 sin sección, `harness-audit` no puede validar con el código de salida que la sección quedó escrita. Fue al usuario como D10 (código de salida cuando falta la sección), y la respuesta fue D10-A: el error solo aplica con `--project` explícito. Lo apliqué como decisión del usuario, sin otra ronda.

**Plan listo** en la revisión final. No queda ninguna cobertura sin revisar. Límite que se mantiene: el comportamiento con modelo de `harness-audit` no se ejerce.

**Build** (2026-10-26, worktree `gh-192`):
- T1 a T6 verificadas por un `hive-verify-task` cada una, todas en `[x]`.
  - T3, T4 y T6 tuvieron rondas de corrección: T6 necesitó tres rondas, por `build.md:45` y `README.md:347`.
  - T2 se reverificó dos veces tras los arreglos de la revisión.
- Implementador de T2 y de los arreglos: `hive-build-backend`. T6: `hive-write-docs`. Ambos modelos configurados, no observados.

**Verify:**
- UI `functional`: `hive-verify-change` recorrió el formulario sobre `d8432ca6` y `37bd63b3`; todo pasó.
- `/code-review` (medium) sobre `388599ff`: 8 hallazgos.
  - 6 corregidos en `97d8b917`; uno de ellos con la decisión D11-A.
  - 2 solo reportados.
- Re-revisión acotada de `388599ff..97d8b917`: 8 hallazgos, todos locales, corregidos en `37bd63b3` sin otra ronda, con pruebas.
- Suite completa sobre `997e8b54`, tras dos merges de `origin/development`: exit 0.

**Entrega:** [PR #198](https://github.com/JhonHawk/tricell-hive/pull/198), merge `88357e5d`; #192 cerrado.

**Incidente:** un `hive-verify-change` corrió `tmux kill-server` en su primera pasada; se informó al usuario.

**Temas abiertos:**
- Los arreglos de cam, carvento y facti. Destino: la sesión de cada proyecto, con `harness-audit`.
- La release menor. Destino: se propone en el reporte de cierre.
