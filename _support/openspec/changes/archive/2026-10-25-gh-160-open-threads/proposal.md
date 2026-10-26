# Temas agregados a mitad del trabajo

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #165 (`1e5732fe`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#160 — Temas agregados a mitad del trabajo se pierden: el agente atiende solo el último en investigación, plan y build](https://github.com/JhonHawk/tricell-hive/issues/160) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana |
| Verificación | `go test ./tests/content/...` · conteo de frases · `go vet ./...` y `go test ./...` sobre el candidato final |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

> Antes de avanzar, esto es interesante, porque ha pasado otras veces:
> - Iniciamos un research
> - Uusario (yo) va añadiendo cosas extras que en teoria amplian el alcances, estando relacionados poco o mucho, no lo sé
> - Lo que hace el agente de la sesión es centrarse en el punto más reciente, ignorando ahora los temas anteriores
> > Esto provoca que varios planes queden a medias (oops) porque oslo consideran lo ultimo

> ojo, aunque dije research, tambien pasa en la etapa de plan y build

Respuesta D8-A «Ticket + flow-plan» (2026-10-25).

Criterios del ticket [#160](https://github.com/JhonHawk/tricell-hive/issues/160), leídos el 2026-10-25:

> - **AC1:** la guía pide mantener la lista de temas abiertos y rige en investigación, plan y build, también cuando ninguna skill está cargada.
> - **AC2:** la guía pide clasificar cada mensaje que trae un tema (aclara, relacionado, aparte) y decir la clasificación al usuario.
> - **AC3:** la guía prohíbe dar por cerrado un tema sin destino y enumera los destinos válidos.
> - **AC4:** los reportes, las preguntas de ruta y de cierre y los traspasos (por ejemplo, tras compactar el contexto o al cambiar de sesión) cubren todos los temas abiertos.
> - **AC5:** `flow-plan/SKILL.md:16` queda como aplicación de esta regla, sin repetirla.
> - **AC6:** las páginas de `site/` afectadas lo describen.

## Objetivo

Cuando el usuario suma temas mientras el trabajo avanza, el modelo responde al último y los anteriores se quedan sin destino; varios planes del usuario quedaron a medias así. La guía solo trata la ampliación de alcance mientras se planea. El cambio agrega una regla que rige en toda etapa: llevar la lista de temas abiertos, clasificar cada tema nuevo, no cerrar ninguno sin destino y cubrirlos todos en cada reporte, pregunta de ruta o cierre y traspaso.

## Alcance y aceptación

Incluye el hogar de la regla (según E8), `content/skills/flow-plan/SKILL.md:16`, los puntos de uso que lo necesiten (`flow-close/references/reporting.md`, la pregunta de ruta de `flow-research`) y las páginas `site/src/content/docs/flows/research.md`, `plan.md` y `build.md`.

Fuera de alcance: un archivo o tablero persistente de temas.

Restricciones: `go test ./tests/content/...` sigue pasando; cada regla vive en un solo hogar; si el hogar es `global.md`, el PR declara el cambio de tamaño.

- AC1. `global.md` pide mantener la lista de temas abiertos que el usuario agrega mientras hay trabajo abierto, en investigación, plan y build, con el alcance de [design.md](design.md) (la respuesta a una pregunta pendiente no es un tema; un hijo no aplica la regla). *Falso en la base cuando* `rg -i -c "open topics" content/guidance/global.md` da 0.
- AC2. La regla pide clasificar en una línea cada mensaje que trae un tema: aclara uno abierto (actualiza el trabajo en curso), abre uno relacionado (se resuelve como alcance antes del trabajo que depende de él) o es aparte (espera su destino). *Falso en la base cuando* esa clasificación no aparece.
- AC3. La regla prohíbe dar por cerrado un tema sin destino y enumera los destinos: ticket, plan o tarea; descartado con el motivo dicho al usuario; en vigilancia con la condición que lo reabre, registrada en el plan o el traspaso. `flow-close` cuenta un tema abierto sin destino como pendiente. *Falso en la base cuando* no aparece.
- AC4. La regla pide que cada reporte de hallazgos, pregunta de ruta o de cierre, traspaso que escribe el agente y primer mensaje tras un resumen de compactación nombre cada tema abierto con su destino propuesto. *Falso en la base cuando* no aparece.
- AC5. `flow-plan/SKILL.md:16` remite a la regla en vez de restatearla. *Falso en la base cuando* la línea no la nombra.
- AC6. `research.md`, `plan.md` y `build.md` del sitio describen el comportamiento enlazando la regla. *Falso en la base cuando* ninguna lo menciona.

## Decisiones

- **E8-A (2026-10-25):** la regla vive en `content/guidance/global.md`, porque un tema nuevo llega en cualquier etapa y sin skill cargada. Costo aceptado: unos 500 bytes sobre 36 990; la revisión del plan mostró que el alcance explícito y la excepción de preguntas necesitan unos 700, desviación avisada al usuario y declarada en el PR.
- **E3-A y E4-A (2026-10-25):** entrega `automatic` y revisión `/code-review`.

## Entrega

`automatic`: rama `feat/gh-160-open-threads` desde `development` en un worktree, commits, push, PR a `development`, `/code-review` y merge cuando pasan las comprobaciones; registro cerrado con un PR pequeño sin revisión; sin despliegues; tras el merge, `hive update`.
