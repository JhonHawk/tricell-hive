# Contrato del hijo de research y verificación acotada

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #52](https://github.com/JhonHawk/tricell-hive/pull/52), merge `1307b1fb`; desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): En validación · T1–T4 verificadas, con commit y desplegadas en local; en pausa hasta la auditoría del usuario |
| Tracker · GitHub Issues | Sin issue asociado; no se ha pedido abrir uno |
| Git | Integrado por merge del PR #52; la rama `feat/flash-class-guidance` y su worktree se eliminaron tras el merge |
| Verificación | `go vet ./...` · `go test ./...` · comprobaciones `rg` por tarea |
| Siguiente paso | Ninguno; registro cerrado |

Fecha inicial: 2026-10-13, hora local CST.

## Pedido

> Haz un super deep-research sobre el tema de nuestro (ironicamente) flujo de research, tanto la skill como los agentes, qué les falta o pueden mejorar, como mejorar la valocidad manteniendo un excelente nivel de calidad en la investigación o descubrimiento, dame una propuesta de mejora o dime si no encontramos en un buen punto, puedes lanzar tantos subagentes como necesites

Respuesta a la pregunta de ruta (2026-10-13): «D1-A flow-plan R1–R5 (Recomendado)», que planifica R1–R5 de la investigación:

- R1: contrato de retorno de `hive-research` (estado, afirmación con ubicación y extracto, contraevidencia, lo no revisado, tope de extensión).
- R2: verificación acotada en `flow-research`.
- R3: brief con condición de parada y exclusiones; lanzar juntos los hijos independientes; un solo explorador primero cuando no se sabe dónde está la respuesta.
- R4: conectar `hive-read-state` desde `flow-research` para el estado de tracker, CI y PR.
- R5: un solo vocabulario de etiquetas de evidencia.

La decisión explícita del usuario cumple la barrera de admisión de `AGENTS.md` (Measurement) para reglas nuevas sin dos sesiones con fallo.

## Objetivo

La investigación del 2026-10-13 concluyó que el flujo de research está en buen punto en calidad, pero que su tiempo lo marca el hijo más lento (58 % del tiempo hasta la respuesta, mediana de 215 s y 33 llamadas por hijo) y que el hilo principal repite el 15 % de las llamadas que el hijo ya hizo. El contrato de `hive-research` no pide estado, ubicaciones, contraevidencia ni límite de extensión, y la skill pide comprobar «afirmaciones consecuentes» sin decir cuáles ni cómo.

El cambio fija qué devuelve el hijo y cómo lo comprueba el padre, para que la verificación consista en abrir las líneas citadas en vez de repetir búsquedas; da a cada hijo una condición de parada; ordena lanzar juntos los hijos independientes; y conecta el rol de estado que hoy nadie usa. Todo vive en el rol, la skill y la referencia de delegación: `global.md` no crece.

## Alcance y aceptación

Incluye `content/agents/review/hive-research.md`, `content/skills/flow-research/SKILL.md`, `content/skills/flow-research/references/delegation.md`, `site/src/content/docs/flows/research.md` y `CHANGELOG.md`.

Excluye:

- `content/guidance/global.md`: las reglas nuevas viven donde se aplican y el núcleo global no crece.
- Cambiar perfiles de modelo de los roles: no hay evidencia y requeriría un piloto (pausados).
- Cláusula contra subagentes anidados: 0 casos en Hive; queda como vigilancia.
- `hive-refute-claim` y `hive-read-state`: sus contratos ya fijan su salida; solo se enruta a ellos.
- Tablas numéricas de esfuerzo o investigación siempre volcada a archivo.

Restricciones que deben seguir siendo ciertas:

- `go vet ./...` y `go test ./...` pasan.
- `README.md` y `llms.txt` no cambian, porque nada de lo que dicen cambia.
- `CHANGELOG.md` recibe su entrada bajo `[Unreleased]`, como exige `CONTRIBUTING.md`.
- Ninguna instrucción nueva contradice o duplica otra (barrido de T4).

Criterios (base: `34b4edc9`):

- AC1. `hive-research.md` exige devolver la respuesta en una o dos frases, estado (complete, partial o blocked) con el motivo de parada, cada afirmación que decide la respuesta con su ubicación y un extracto literal corto, la contraevidencia buscada, lo que no se revisó y un tope de extensión. *Falso en la base cuando* `rg -n "partial|counterevidence|did not check|locator" content/agents/review/hive-research.md` no devuelve nada.
- AC2. `flow-research/SKILL.md` limita la comprobación del padre a las afirmaciones que deciden la respuesta, abriendo sus ubicaciones citadas sin repetir las búsquedas del hijo salvo que la ubicación falte o no sostenga la afirmación, y manda las premisas de alto impacto no resueltas, juntas, a un hijo `hive-refute-claim`. *Falso en la base cuando* la skill solo dice «checks consequential claims against their sources» y no menciona `hive-refute-claim`.
- AC3. `flow-research/SKILL.md` pide en el brief de cada hijo de research lo ya sabido, lo excluido y la condición de parada (respuesta nombrable con evidencia convergente, o alcance agotado), indicando cuál ocurrió. *Falso en la base cuando* `rg -n "scope is exhausted|out of scope|evidence converges" content/skills/flow-research/SKILL.md` no devuelve nada.
- AC4. `delegation.md` ordena lanzar juntos los hijos independientes y, cuando no se sabe dónde está la respuesta, lanzar primero un solo hijo y repartir el resto según lo que encuentre. *Falso en la base cuando* la sección «Split and announce work» no dice nada de lanzar juntos ni de un primer hijo.
- AC5. `flow-research/SKILL.md` envía a `hive-read-state` la recolección de estado de tickets, PR, CI o despliegues nombrados, o de una consulta de tracker con proyecto y estados, cuando produce salida larga. *Falso en la base cuando* `rg -n "hive-read-state" content/skills` no devuelve nada.
- AC6. `hive-research.md`, `flow-research/SKILL.md` y `delegation.md` usan las etiquetas de la regla global de evidencia (inspected facts, source claims, inference, recommendations; assumptions para premisas). *Falso en la base cuando* `hive-research.md` dice «source facts, live observations» y la skill «current observed behavior, documented intent».
- AC7. La página `site/.../flows/research.md` describe que los ayudantes informan de dónde sale cada afirmación y qué no revisaron, que el agente abre las líneas citadas de las afirmaciones decisivas y que un ayudante recoge el estado largo de tickets y CI. *Falso en la base cuando* la página no menciona lo no revisado ni la recolección de estado.

## Entrega

Respuestas del usuario (2026-10-13):

- D3-A: `hold` en `feat/flash-class-guidance` (worktree `.claude/worktrees/feat-flash-class-guidance`). Se implementa y verifica en el árbol de trabajo, sin commit ni publicación. Esta carpeta queda sin versionar y pendiente, como el código.
- D4-A: sin revisión dedicada. Basta la inspección del diff, las comprobaciones `rg` de cada tarea y `go vet ./...` con `go test ./...`.
- D5-A: iniciar `flow-build` ahora.

Cambio durante el build (2026-10-13): «al terminar los cambios puedes hacer commit y deploy, después de esto abrire una sesión para que auditemos el comportamiento». Esto sustituye D3-A: un commit por tarea al llegar a `[x]` en `feat/flash-class-guidance`, sin la carpeta de cambio, y después el despliegue local desde esa rama a los hosts registrados. Push, PR y merge siguen sin autorizar. Si se piden commits, sus mensajes nombran la evidencia de sesiones reales (host, sesión y fecha) de `design.md`, como exige `AGENTS.md` (Measurement).
