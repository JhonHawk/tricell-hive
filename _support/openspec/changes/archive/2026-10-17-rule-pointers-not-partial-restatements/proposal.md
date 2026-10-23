# Punteros en lugar de copias parciales de una regla

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por [PR #64](https://github.com/JhonHawk/tricell-hive/pull/64) (`4dc73c28`), 2026-10-17 |
| Tracker · GitHub Issues | Sin issue: nace de la auditoría de las sesiones de fleetsystem del 2026-10-17 |
| Git | rama de prueba activa `feat/delegation-threshold-global` · commit por tarea y despliegue local (C1-A) · sin push ni merge |
| Verificación | comprobaciones `rg` por criterio · `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` |
| Siguiente paso | Ninguno: cambio cerrado y archivado |

## Petición

> Analizando la sesión /Users/jmartinez/Development/projects/comsys/fleetsystem no esta usando la herramiento de preguntas, no al final

> Aqui tengo una duda, es mejor tener la regla de "usar tarjeta" disperso en varios lugares onde se hace mención el uso o una sola vez la regla con todas las menciones? aunque este es un ejemplo, hablo en general

> R4-A Planear ambos (Recommended)

> Esto debe volverse canon y regla en nuestra forma de redactar reglas, porfa

> D1-B También en harness-audit

> nadamas que harness-audit debe confirmar con el usuario qué regla es la que debe seguir

(Todas del 2026-10-17.) R4-A decía: "Un registro de cambio con flow-plan, en la rama de prueba. (1) El principio entra en AGENTS.md, en 'Guidance structure', para que el mantenimiento futuro lo siga. (2) Las unas 8 copias parciales de la regla de preguntas pasan a ser avisos, y se aclara el orden de global.md:20." D1-B decía: "Además se amplía HA-IF-08 en harness-audit-rules.md y en la skill para que marque las copias parciales en los proyectos que auditas."

## Objetivo

Dos sesiones de Claude Code en fleetsystem (`02891730` y `8168b9a2`, 2026-10-17, claude-opus-5-5) terminaron varias respuestas con una pregunta en texto y la línea de espera, sin la tarjeta nativa. Grok, OpenCode y Codex no muestran el fallo en sus sesiones del 16 y 17 de octubre. Antes, carvento había mostrado lo mismo (auditoría del 2026-10-16).

El procedimiento completo vive en `questions.md:25-37`: primero la línea de espera y después la tarjeta. Pero unas diez partes de la guía lo resumen como "termina con la línea de espera" o remiten a "las reglas globales de elección", que son a su vez un resumen. Un modelo que sigue la versión corta omite la tarjeta.

La evidencia externa apunta en la misma dirección: la guía de OpenAI para GPT-5.6, la documentación de Claude Code, IHEval y ManyIFEval. También el registro de Hive: las contradicciones de la poda del 15 de octubre y el umbral de delegación que no se leía. La regla vive completa en un solo lugar; donde se aplica va solo un puntero que manda leerla y seguirla; y `global.md` repite únicamente la invariante mínima, completa, de las reglas que deben actuar sin que se lea su referencia.

El cambio:
- Convierte ese principio en regla de redacción de Hive (AGENTS.md).
- Lo aplica a la regla de preguntas.
- Hace que `harness-audit` marque las copias parciales en los proyectos que audita (HA-IF-08).

## Alcance y aceptación

Incluye:
- La viñeta "Give each rule one canonical home" de AGENTS.md.
- `content/guidance/global.md:20`.
- Las once menciones parciales de la regla de preguntas en las skills.
- HA-IF-08 en `content/skills/harness-audit/references/instruction-files.md` y su fila en `_support/docs/harness-engineering/harness-audit-rules.md`.

Queda fuera:
- **Las demás reglas de Hive con copias:** se corrigen cuando un cambio las toque, siguiendo el principio nuevo.
- **La regla de glosar IDs**, duplicada completa en `global.md:18` y `reporting.md:19`, que no se ha desfasado.
- **Un hook del host para forzar la tarjeta.**
- **El sitio**, porque `flows/research.md:47` y `flows/close.md:63` ya describen la tarjeta y la línea de resumen.

Restricciones:
- La suite local pasa, incluida la prueba de paridad de IDs de `harness-audit` (`tests/skills/harness_audit_rules_test.py`).
- El texto distribuido se escribe para los modelos del perfil `execution`.
- Cada commit nombra las sesiones de origen y las decisiones R4-A y D1-B.
- El pull request declara cuánto cambia `global.md`.

- AC1. La viñeta de AGENTS.md dice:
  - El hogar canónico es el único que enuncia los pasos completos de una regla, su orden y sus excepciones.
  - Donde la regla se aplica fuera de su hogar va solo un puntero que nombra la regla y cuándo leerla, nunca una parte de sus pasos.
  - `global.md` repite una regla solo cuando debe actuar sin que se lea su hogar, y entonces enuncia completa su invariante mínima.
  - Una excepción local nombra cada elemento de la regla que cambia.
  - Una copia existente se corrige cuando un cambio edita esa copia o la regla que repite.

  *Falso en la base cuando* `rg -c 'never restate some of its steps' AGENTS.md` no encuentra coincidencias.
- AC2. `global.md` dice en la regla de preguntas que el mensaje termina con una línea de espera y que después se pregunta con la tarjeta nativa cuando el host la tiene; si no, las opciones completas van en el texto, antes de esa línea. *Falso en la base cuando* `rg -c 'end the message with one pause line; then ask through a waiting native card' content/guidance/global.md` no encuentra coincidencias.
- AC3. Ninguna skill ni referencia fuera de `questions.md` resume parcialmente la regla de preguntas. Diez de los once sitios remiten a `questions.md`. `verification.md:69` remite a la pregunta de la misma pausa de validación ("in the same way"), que nombra la línea de espera, la tarjeta y la condición de los hijos. *Falso en la base cuando* `rg -n -o 'global choice rules|global pause line|ending with the pause line|end with its pause line|goes in the pause line|with the pause line and the in-progress item|native question tool when it has one, otherwise in text' content` devuelve coincidencias (12 en la base).
- AC4. HA-IF-08 en `instruction-files.md` marca también la copia parcial que repite solo algunos pasos de una regla. No elige qué versión queda: pregunta al usuario cuál es la regla a seguir, una decisión por regla con cada versión citada y su ubicación, y después propone borrar las demás (sujeto a HA-IF-13) o reanclarlas como punteros. La fila del catálogo cita la decisión D1-B. No marca la invariante mínima completa en un archivo que siempre se carga. `harness-audit/SKILL.md:58` incluye esa decisión entre las que pregunta al cerrar. Su fila en `harness-audit-rules.md` refleja el cambio, con la fuente y la fecha nuevas. *Falso en la base cuando* `rg -c 'partial restatement' content/skills/harness-audit/references/instruction-files.md` no encuentra coincidencias.

## Entrega

- **Modo:** C1-A, elegido en esta sesión para este repositorio: un commit por tarea en `feat/delegation-threshold-global` cuando llega a `[x]`, y después despliegue local. Sin push ni merge.
- **Revisión de código:** ninguna dedicada, porque son cambios solo de texto, según la preferencia guardada para Hive.
- **Registro:** esta carpeta queda sin versionar hasta el merge.
