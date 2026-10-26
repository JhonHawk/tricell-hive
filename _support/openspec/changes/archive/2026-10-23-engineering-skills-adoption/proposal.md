# Adopción de técnicas de diagnóstico, cuerpo de tickets y créditos (2026-10-23)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-23 por el [PR #121](https://github.com/JhonHawk/tricell-hive/pull/121), merge `5f249c60`; registro archivado el mismo día |
| Tracker · GitHub Issues | Sin issue: el trabajo sale de decisiones del usuario en la revisión de referencias (D14, D15) |
| Git | `automatic` (D9) · integrado por merge del PR #121; rama `feat/engineering-skills-adoption` eliminada tras el merge |
| Verificación | Comprobaciones por tarea con `hive-verify-task` · `go vet ./...`, `go test ./...`, pruebas de skills y build del sitio sobre el candidato final · sin revisión dedicada (D10) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

> Luego, siguen estas:
> https://www.aihero.dev/skills-improve-codebase-architecture
> https://www.aihero.dev/skills-diagnosing-bugs
> https://www.aihero.dev/skills-triage
> https://www.aihero.dev/skills-wizard
> https://www.aihero.dev/skills-implement-spec

Respuestas del usuario (2026-10-23, tarjetas nativas):

- D14 (qué adoptar): «K1–K4: técnicas de diagnóstico, K9: contenido de los tickets». El resto (K5–K8, K10, K11) queda en observación en Engram.
- D15 (nombres de referencias en `agent-delivery.md`): «Quitar los nombres de ese archivo».
- D16 (dónde vive K9): «Referencia nueva».
- D17 (rutas en los tickets): «Como evidencia, con fecha».
- D18 (créditos): «Sí, en este cambio».

## Objetivo

La revisión de cinco skills externas frente a Hive no encontró fallos reales que exijan reglas nuevas, pero el usuario decidió adoptar dos grupos de ideas por decisión explícita, como permite la regla de admisión de `AGENTS.md` (Measurement). El resultado es guía distribuida que (1) da al diagnóstico técnicas concretas para reducir y discriminar una reproducción, (2) define qué lleva el cuerpo de un ticket que Hive abre, (3) deja de nombrar repositorios de referencia en un documento versionado fuera de la página de créditos, y (4) acredita en esa página el origen de las reglas que Hive tomó de él. Además corrige el `CHANGELOG.md`, al que le faltaban las entradas de #118 y #120.

## Alcance y aceptación

Incluido: `content/skills/flow-research/references/diagnosis.md`, una referencia nueva `content/skills/flow-research/references/ticket-body.md` con punteros desde `content/skills/flow-research/SKILL.md` y `content/skills/flow-close/SKILL.md`, `_support/docs/architecture/agent-delivery.md`, `site/src/content/docs/inspiration.md`, `CONTRIBUTING.md` (su enumeración de créditos, conservando «in that order») y `CHANGELOG.md`. Todo el texto de archivos fuente en inglés y redactado con palabras propias, sin copiar el texto de la referencia.

Excluido: los registros archivados que citan textualmente peticiones del usuario con nombres de referencias (la regla de cita literal los exige); `CHANGELOG.md:158` y `CONTRIBUTING.md:78-79`, que nombran solo proyectos acreditados; las páginas `site/src/content/docs/flows/close.md` y `research.md`, que no describen el cuerpo del ticket ni las técnicas de diagnóstico (comprobado en la revisión del plan); `README.md` y `llms.txt`, que no nombran proyectos; las ideas en observación K5–K8, K10 y K11; cualquier cambio de código del gestor.

Restricciones: `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` y `pnpm run build` en `site/` siguen pasando; los enlaces nuevos pasan la validación de enlaces de instrucciones; ninguna regla queda con dos hogares (`AGENTS.md`, «Guidance structure»).

En los comandos, `<diag>` es `content/skills/flow-research/references/diagnosis.md`, `<tb>` es `content/skills/flow-research/references/ticket-body.md` y `<unreleased>` es la salida de `awk '/^## \[Unreleased\]/{f=1;next} /^## \[/{f=0} f' CHANGELOG.md`.

- AC1. `diagnosis.md` pide, una vez reproducido el fallo, reducir la reproducción a la más pequeña que aún muestra el síntoma. *Falso en la base cuando* `rg -c -i 'smallest reproduction' <diag>` no encuentra nada.
- AC2. `diagnosis.md` pide comparar la reproducción con el caso más cercano que funciona. *Falso en la base cuando* `rg -c -i 'closest (case|variant) that works' <diag>` no encuentra nada.
- AC3. `diagnosis.md` pide cambiar una sola variable por experimento. *Falso en la base cuando* `rg -c -i 'one variable' <diag>` no encuentra nada.
- AC4. `diagnosis.md` pide, ante un fallo intermitente, subir la tasa de reproducción antes de formular hipótesis e informar la tasa observada. *Falso en la base cuando* `rg -c -i 'intermittent' <diag>` no encuentra nada.
- AC5. `diagnosis.md` pide marcar la instrumentación temporal con una etiqueta única y retirarla buscando esa etiqueta, como parte de retirar las medidas temporales. *Falso en la base cuando* `rg -c -i 'unique tag' <diag>` no encuentra nada.
- AC6. `diagnosis.md` dice que forzar el fallo cambiando código, un fixture o datos sirve para construir la reproducción y no es un arreglo, y pide confirmar con un diff contra la copia sin cambiar que el cambio se aplicó antes de confiar en el resultado. *Falso en la base cuando* `rg -c -i 'forced failure' <diag>` no encuentra nada.
- AC7. `ticket-body.md` existe y define el comportamiento actual y el esperado, la evidencia con `path:line` y el commit o la fecha en que se leyó, criterios observables y lo que queda fuera de alcance. *Falso en la base cuando* el archivo no existe; tras el cambio, cada una de `rg -c -i 'current behavior' <tb>`, `rg -c -i 'expected behavior' <tb>`, `rg -c 'path:line' <tb>`, `rg -c -i 'commit or (the )?date' <tb>` y `rg -c -i 'out of scope' <tb>` encuentra al menos una línea.
- AC8. El párrafo del ticket de `flow-research/SKILL.md` y el paso 5 de `flow-close/SKILL.md` enlazan a `ticket-body.md`. *Falso en la base cuando* `rg -c 'ticket-body.md' content/skills/flow-research/SKILL.md content/skills/flow-close/SKILL.md` no encuentra nada.
- AC9. `agent-delivery.md` no nombra repositorios de referencia y su frase sobre los flujos de referencia inspeccionados no lleva el paréntesis de nombres. *Falso en la base cuando* `rg -c -i 'gentle-ai|superpowers|mattpocock|OpenSpec' _support/docs/architecture/agent-delivery.md` encuentra la línea 9.
- AC10. `inspiration.md` tiene la entrada «Skills For Real Engineers» enlazada a `https://github.com/mattpocock/skills`, y la enumeración de `CONTRIBUTING.md` la incluye. *Falso en la base cuando* `rg -c 'Skills For Real Engineers' site/src/content/docs/inspiration.md`, `rg -c 'github.com/mattpocock/skills' site/src/content/docs/inspiration.md` y `rg -c 'Skills For Real Engineers' CONTRIBUTING.md` no encuentran nada.
- AC11. `<unreleased>` tiene la entrada de #118 (stash compartido entre worktrees). *Falso en la base cuando* `<unreleased> | rg -c '#118'` no encuentra nada.
- AC12. `<unreleased>` tiene la entrada de #120 (lo aún no especificado en los planes). *Falso en la base cuando* `<unreleased> | rg -c '#120'` no encuentra nada.
- AC13. `<unreleased>` tiene la entrada de este cambio, que describe las técnicas de diagnóstico y el cuerpo de los tickets sin número, porque no tiene ticket (las entradas citan tickets o, a falta de ticket, su PR ya existente). *Falso en la base cuando* `<unreleased> | rg -c -i 'diagnos'` y `<unreleased> | rg -c -i 'ticket'` no encuentran nada.

## Entrega

- Modo `automatic` (D9, elegido en esta sesión para este repositorio): rama `feat/engineering-skills-adoption` desde `development`, un commit por tarea verificada, push, PR a `development` y merge cuando pasen las pruebas locales (el repositorio no tiene CI). Luego `hive update` desde el checkout, por la marca de mantenedor.
- El PR lista los lugares revisados y el cambio de tamaño de `content/guidance/global.md` (0: no se edita), como pide `AGENTS.md` («Guidance structure»).
- Revisión: sin revisión dedicada (D10).
- Cierre del registro: un PR pequeño de cierre sin revisión dedicada, que fusiono yo.
- No incluye release ni promoción a `master`.
