# La sección `## Hive` guarda solo hechos y `harness-audit` audita, pregunta y aplica

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #198](https://github.com/JhonHawk/tricell-hive/pull/198) (merge `88357e5d`) |
| Tracker · GitHub Issues | • [#192 — La sección ## Hive mezcla datos con decisiones permanentes y nada la crea ni la mantiene](https://github.com/JhonHawk/tricell-hive/issues/192) |
| Git | automatic (declarado hoy en `AGENTS.md`) · PR a `development` y merge tras `/code-review` y la suite local · sin parada humana |
| Verificación | pruebas Go y Python · build y check del sitio · `hive doctor` real sobre repos desechables · `hive-verify-change` sobre el formulario del TUI (functional) · `/code-review` |
| Siguiente paso | ninguno; la release menor que publica el cambio se propone aparte |

## Pedido

> Ayudame con lo siguiente:
>
> Las instrucciones en los proyectos, cam, carvento, facti
>
> - Qué tanto estan ayudando a que Hive funcione correctamente?
> - Esa pleca que configura hive en cada proyecto tiene pre-desiciones, creo que debemos ajustar el formato
>
> - Faltará una skill de onboarding que cree, ajuste o actualice esa pleca

> Review y UI verification de hecho creo que son opciones que siempre deberian preguntarse

> ahora, pregunta, harness-audit sigue valiendo la pena como skill?

> podemos hacer que harness audit haga preguntas el usuario para resolver desiciones y tambien actualice y no solo reporte?

Criterios de aceptación de [#192](https://github.com/JhonHawk/tricell-hive/issues/192), citados textualmente (leído el 2026-10-26):

- `project-settings.md` define el formato con hechos y decisiones separados, el lugar del punto de terminado y de los permisos permanentes, y la sección del workspace; las referencias que leen esos valores (`delivery-decisions.md`, `global.md`, `unattended-delegation`, `harness-audit`) usan el formato nuevo sin contradicciones.
- Existe una skill de onboarding distribuida que, ejecutada en un workspace con repositorios sin sección, propone los hechos con su evidencia, pregunta las decisiones en una ronda y escribe la sección en cada repositorio solo tras la respuesta.
- La misma skill, ante una sección existente con decisiones en prosa, las lista con `archivo:línea` y propone su traslado sin aplicarlo sin autorización.
- Las secciones con el formato actual siguen funcionando, o la skill ofrece migrarlas.
- La página del sitio que describe los ajustes y la prueba de tamaño de skills se actualizan.

Las decisiones de planeación cambian esos criterios así:
- D1 y la nota de 2026-10-26 hacen que la sección no guarde decisiones, así que no hay dos bloques.
- D3 deja los permisos permanentes sin clave.
- D9 hace que la «skill de onboarding» sea `harness-audit` ampliada, no una skill nueva.

Los `AC<n>` de abajo los reemplazan.

## Objetivo

Hoy la sección `## Hive` mezcla hechos verificables (rama base, tracker, specs) con decisiones que autorizan efectos sin preguntar (`Delivery`, `Review`, `UI verification`). Además, las decisiones permanentes que el formato no admite terminan en prosa con fuerza dispar, como el punto de terminado o los bypass. La auditoría de cam, carvento y facti del 2026-10-26 lo mostró:
- secciones ausentes (facti tiene una de cuatro);
- un bypass permanente a `qa` que corre migraciones;
- el punto de terminado duplicado en prosa;
- rutas de `Specs` incompletas.

Este cambio deja la sección como ficha de hechos. La entrega, la revisión y la verificación de UI se preguntan siempre, en la ronda única que ya existe, y se reutilizan durante la sesión. El punto de terminado pasa a ser la clave `Completion`. `harness-audit` deja de ser solo un reporte: audita, pregunta las decisiones y aplica lo aceptado en el proyecto actual, incluida la creación y actualización de la sección, con `hive doctor` como validador de salida.

## Alcance y aceptación

**Incluye:**
- el catálogo de claves en `project-settings.md`;
- las reglas que leen valores declarados;
- el manager en Go (`hive doctor`, `hive project set`, la vista Project del TUI);
- `harness-audit`, con su procedimiento de aplicación y HA-IF-17;
- la sección `## Hive` de este repositorio;
- el sitio, el README, `llms.txt`, la documentación de arquitectura y el CHANGELOG.

**Excluye:**
- Editar cam, carvento o facti: son otros proyectos. Se corrigen corriendo `harness-audit` en la sesión de cada uno.
- Una clave para permisos permanentes (D3-A).
- Una pilot de modelo para medir la skill: las pilots están en pausa.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...`, `go test ./...` y la suite Python pasan.
- Cada `SKILL.md` queda en 16 000 bytes o menos.
- `Delivery: direct-base` conserva sus efectos.
- Las secciones existentes con claves retiradas no bloquean nada.
- El rol `hive-review-harness` sigue siendo de solo lectura.
- El PR declara el cambio de tamaño de `content/guidance/global.md` y lista los lugares revisados.

**Criterios:**

- AC1. `project-settings.md` declara como claves solo hechos. Son `Project`, `Base branch`, `Tracker`, `Specs`, `Environments`, `Delivery` (solo `direct-base`), `Completion` y `Hive guidance`. `Review`, `UI verification` y `Delivery: automatic|interactive` aparecen como retiradas, se preguntan siempre y nunca autorizan nada, y el ejemplo no las muestra. *Falso en la base cuando* `project-settings.md:9` las lista como opcionales y el ejemplo incluye `Review: /code-review`.
- AC2. El punto de terminado posterior se lee de la clave `Completion` (`<estado intermedio> on <ref> · <estado final> on <ref>`). «Later completion point», `global.md`, `flow-close`, `flow-deploy`, `reporting.md`, el README y las páginas `close`/`deploy` del sitio se refieren a esa clave. *Falso en la base cuando* la condición dice «a project's guidance declares a completion point» (`project-settings.md:27`).
- AC3. Ninguna regla distribuida trata un `Review`, `UI verification` o `Delivery: automatic|interactive` declarado como respuesta o autorización. La ronda de entrega, revisión y UI se pregunta salvo que una respuesta de la sesión o el pedido la resuelva, y `unattended-delegation` exige esos valores respondidos. *Falso en la base cuando* `global.md:69`, `delivery-decisions.md:5,26,36` y `unattended-delegation/SKILL.md:8` aceptan valores declarados.
- AC4. `hive doctor` marca `Review`, `UI verification` y `Delivery: automatic|interactive` con un aviso no bloqueante que dice que ya no se leen y se preguntan en cada sesión. `Delivery` acepta solo `direct-base`, y cualquier otro valor bloquea. `Completion` es conocida y avisa si no tiene la forma. *Falso en la base cuando* `Delivery: automatic` sale `Valid` y `Completion` sale como clave desconocida.
- AC5. `hive doctor` sale con un código distinto de 0, con `hive: Project section has blocking findings` en stderr, cuando, en un repositorio Git, el `AGENTS.md` de la raíz existe y no se puede leer como archivo regular dentro del límite de tamaño, o su sección `## Hive` tiene un hallazgo bloqueante: un valor obligatorio ausente o vacío, un `Delivery` o `Hive guidance` inválido, o una sección duplicada. Sin `AGENTS.md` o sin sección, o cuando la ruta no existe, no es un repositorio Git o git no puede ejecutarse, sale distinto de 0 solo cuando se pasa `--project` de forma explícita (D10-A, D11-A). Sale con 0 cuando solo hay avisos, o sin flag en esos casos. *Falso en la base cuando* `runDoctor` devuelve `nil` siempre (`tooling/cli/main.go:128-143`).
- AC6. El formulario de la vista Project y `hive project set` no ofrecen `Review` ni `UI verification`. `Delivery` cicla entre vacío y `direct-base`, y `Completion` es un campo de texto. `--set Review`, `--set UI verification` y `--set "Delivery: automatic"` se rechazan nombrándolas como retiradas y diciendo cómo quitarlas. Una línea retirada que ya existe se conserva y se puede quitar con `--unset` o, en el formulario, con su fila `remove`. *Falso en la base cuando* el formulario muestra nueve campos con `Review` y `UI verification`.
- AC7. `harness-audit`, en el hilo principal, pregunta las decisiones de sus hallazgos, y en la misma ronda la entrega y la revisión cuando lo aplicado terminará en Git. Tras la respuesta aplica los resultados aceptados en el proyecto actual con la edición del host, crea o actualiza la sección `## Hive` y la valida con el código de salida de `hive doctor --project`. Nunca aplica `propose-global` ni toca `CLAUDE.local.md`. El rol `hive-review-harness` devuelve hallazgos y decisiones sin editar. Su descripción se activa con pedidos de configurar Hive en un proyecto. *Falso en la base cuando* «Boundaries» dice «Stay read-only».
- AC8. HA-IF-17 marca las claves retiradas como `stale` con propuesta de quitarlas y valida la forma de `Completion`. Un punto de terminado en prosa sale como `relocate` a `Completion`, y los permisos permanentes en prosa se listan con `archivo:línea` para que el usuario decida si se quitan o se quedan. El catálogo de procedencia registra el cambio, y `tests/skills/harness_audit_rules_test.py` pasa. *Falso en la base cuando* HA-IF-17 no menciona claves retiradas ni `Completion`.
- AC9. La sección `## Hive` de este repositorio no declara `Review` ni `Delivery`. `/code-review` queda como preferencia en prosa, y la línea que trataba los valores declarados como instrucción explícita (`AGENTS.md:59`) se reescribe. El `hive doctor --project .` compilado sale con 0 y sin avisos de claves retiradas. *Falso en la base cuando* `AGENTS.md:13-15` declara `Review: /code-review` y `Delivery: automatic`.
- AC10. El sitio (`getting-started`, `build`, `verify`, `security`, `close`, `deploy`, `skills`, `index`), el README, `llms.txt`, `deployment-manager.md` y `workspace-and-artifacts.md` describen el conjunto nuevo de claves y la aplicación de `harness-audit`. El CHANGELOG `[Unreleased]` trae una nota de migración. *Falso en la base cuando* `getting-started.md:35-39` ofrece `Review`, `Delivery` y `UI verification` como claves opcionales.

## Decisiones en vigor

- **Nota del usuario (2026-10-26):** `Review` y `UI verification` se preguntan siempre.
- **D1-A:** `Delivery: automatic|interactive` también se pregunta siempre; `direct-base` se queda como hecho de tipo de repositorio.
- **D2-A:** el punto de terminado pasa a ser una clave.
- **D3-A:** los permisos permanentes no tienen clave; la auditoría los señala y pregunta por proyecto.
- **D4-A:** la escritura se hace con la edición del host y se valida con `hive doctor`.
- **D5-A:** las claves retiradas producen un aviso que no bloquea.
- **D6-A:** `hive doctor` sale distinto de 0 si hay un hallazgo bloqueante.
- **D7-A:** las reglas de la sección se quedan en `harness-audit`.
- **D8-A:** UI verification `functional`, válida para la sesión en este repo.
- **D9-A:** `harness-audit` audita, pregunta y aplica, y absorbe el onboarding; no se crea `project-setup`.
- **D10-A:** sin `AGENTS.md` o sin sección, `hive doctor` sale distinto de 0 solo con `--project` explícito.
- **D11-A** (tras `/code-review`): con `--project` explícito, una ruta inexistente, fuera de Git o donde git no puede ejecutarse también sale distinto de 0.

**Decisiones menores tomadas sin preguntar:**
- El nombre de la clave es `Completion`, con la forma `<estado> on <ref> · <estado> on <ref>`.
- `harness-audit` conserva su nombre para no tocar el catálogo, el rol, el snapshot de OpenCode ni el sitio.
- `/code-review` pasa a prosa como preferencia del mantenedor.
- Tras la revisión del plan:
  - El formulario muestra una fila `remove` por cada clave retirada presente, y `Completion` lleva un placeholder deducido de `Environments`.
  - Los mensajes de claves retiradas dicen cómo quitarlas.
  - `harness-audit` pregunta la entrega en su ronda de decisiones, como exige `delivery-decisions.md`, y bajo delegación desatendida no aplica las decisiones que D3-A deja al usuario.

## Entrega

| Repositorio | Base | Modo | Revisión | Fin acordado |
| --- | --- | --- | --- | --- |
| tricell-hive | `development` | automatic (declarado hoy; rige hasta que este cambio se integre) | `/code-review` (declarado hoy) sobre el candidato local | merge del PR a `development`; luego un PR pequeño de cierre del registro de cambio, sin revisión dedicada, que se mergea |

**Después del merge**, por el marcador de mantenedor: recompilar el binario (`go build -o "$(command -v hive)" ./tooling/cli`) y correr `hive update`. Las dos cosas están autorizadas por `AGENTS.md`.

**Release:** el cambio de ajustes de la sección sube la versión menor (`CONTRIBUTING.md:121`), pero la release se propone al cierre y no se corta sin pedido.

**Despliegue:** ninguno; el sitio se publica con la release.
