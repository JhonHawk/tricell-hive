# Árbol compartido, esperas y hooks de Git

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por [PR #64](https://github.com/JhonHawk/tricell-hive/pull/64) (`4dc73c28`), 2026-10-17 |
| Tracker · GitHub Issues | Sin issue: nace de la auditoría de sesiones del 2026-10-16 |
| Git | rama de prueba activa `feat/delegation-threshold-global` · sin commits hasta que los pidas · despliegue local solo a pedido |
| Verificación | comprobaciones `rg` por criterio · `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` |
| Siguiente paso | Ninguno: cambio cerrado y archivado |

## Petición

> R1-B Planificar con flow-plan

(2026-10-16, respuesta a R1: qué hacer con U1, subagente que ejecuta git sobre todo el árbol compartido, y U2, espera cuya condición de fin no refleja el estado real.)

> este y el del --no-verify tambien me preocupa

(2026-10-16, mismo turno.)

Decisiones del usuario en la misma sesión:
- **D1-A:** adoptar U3, saltarse hooks de Git sin permiso, por decisión explícita. Solo hay un caso real.
- **D2-A:** la cláusula de U1 va en el pie de contrato de los roles que escriben y en `delegation.md`.
- **D3-A:** U3 cubre también los chequeos que un hook omite en silencio.
- **D4** (respuesta propia, tras la revisión del plan): "Antes de empujar confirma que hay CI que lo pueda atrapar, si no queda pendiente, en el reporte se incluye la propuesta de instalar las herramienta".
- **D5-A:** la aprobación para saltar un hook vale el resto de la sesión en ese repositorio.

## Objetivo

La auditoría de siete sesiones reales del 2026-10-16 encontró tres fallos de proceso.

- **U1, `git stash` de un subagente en el árbol compartido.** Un subagente que escribía en un árbol compartido con el hilo principal ejecutó `git stash` para ver si un test ya fallaba antes. Así guardó también los cambios del hilo principal. Ya había pasado en carvento `039a9a2b` (2026-10-13) y ahora en pnpm `e5edd628`. El pie de contrato de los roles ya dice "Modify or delete only the paths the brief names", pero el subagente no entendió `git stash` como una modificación de rutas.
- **U2, señales de fin que no reflejan el estado real.** En pnpm `e5edd628`, un `grep` anclado nunca vio la línea `    Finished` ni un error distinto del esperado, y la sesión quedó parada 15 y 27 minutos hasta que el usuario preguntó. En fleetsystem `b6644329`, `gh pr checks --watch` terminó con "no checks reported" antes de que CI registrara los checks. `delegation.md` exige una señal de fin, pero no que cubra todos los resultados.
- **U3, hooks saltados y chequeos omitidos.** En pnpm `e5edd628` los 8 commits se hicieron con `--no-verify` sin preguntar. Además, el hook de pre-push se saltaba los chequeos de Rust en silencio porque faltaba cargo, y Dylint falló en CI cuatro rondas seguidas. Ninguna regla distribuida de Hive menciona los hooks.

El cambio hace explícitas esas tres reglas donde las lee quien actúa. Así, un hijo no reescribe el árbol de otros. Toda espera termina con éxito, con fallo o con "aún no existe". Y un commit no se salta los hooks ni deja chequeos sin correr sin decirlo.

## Alcance y aceptación

Incluye:
- El pie de los 10 roles que escriben.
- `content/skills/flow-research/references/delegation.md`.
- `content/skills/git-workflow/SKILL.md`.
- La página `site/src/content/docs/flows/build.md`.

Queda fuera:
- **Worktrees obligatorios para hijos en paralelo:** se descartaron en D2.
- **Cambios en `content/guidance/global.md`:** la regla de delegación de la línea 79 ya remite a la ruta asignada.
- **Los hallazgos de conteo 1 de la auditoría que no son U3:** siguen como puntos a vigilar en Engram.

Restricciones:
- `global.md` no cambia de tamaño.
- La suite local pasa.
- El texto se escribe para los modelos del perfil `execution`.
- El commit nombra las sesiones de origen (host, ID y fecha), como pide AGENTS.md.

- AC1. El pie de los 10 roles que llevan "Modify or delete only the paths the brief names" dice que `git stash`, `reset`, `checkout` o `restore` de rutas, `clean` y `switch` modifican todas las rutas que tocan. También dice que, en un árbol que otros pueden estar escribiendo, la comparación con la base se hace con Git de solo lectura o en un worktree aparte. *Falso en la base cuando* `rg -l 'git stash' content/agents` no lista ningún archivo.
- AC2. La sección del brief en `delegation.md` dice lo mismo desde el lado del hilo principal, y lo aplica también al hilo principal mientras un hijo trabaja en ese árbol. *Falso en la base cuando* `rg -c 'stash' content/skills/flow-research/references/delegation.md` no encuentra coincidencias.
- AC3. La sección de esperas de `delegation.md` exige que la señal de fin termine con el éxito, con el fallo y con un objetivo que aún no existe. Pide preferir el estado de salida a la coincidencia de texto, comprobar el patrón contra la salida real, y tratar "no checks reported" como no terminado. *Falso en la base cuando* `rg -c 'does not exist yet|no checks reported' content/skills/flow-research/references/delegation.md` no encuentra coincidencias.
- AC4. `git-workflow` prohíbe saltar un hook (`--no-verify`, `git commit -n`, `HUSKY=0`, `LEFTHOOK=0`, un `core.hooksPath` sobrescrito) sin aprobación explícita del usuario para ese repositorio. Esa aprobación vale el resto de la sesión en ese repositorio y se nombra en el reporte de cada commit que cubre. *Falso en la base cuando* `rg -c -i 'no-verify' content/skills/git-workflow/SKILL.md` no encuentra coincidencias.
- AC5. `git-workflow` pide leer el hook que falla o se salta chequeos por falta de una herramienta, y correr esos chequeos con las herramientas ya disponibles. Los que no se pueden correr se reportan como no ejecutados, con la herramienta que falta y la propuesta de instalarla, y se pregunta antes del push. Si nadie puede responder, solo se empuja tras confirmar que el CI del repositorio corre esos mismos chequeos; si no, el push queda pendiente. *Falso en la base cuando* `rg -c -i 'skips checks|skipping checks' content/skills/git-workflow/SKILL.md` no encuentra coincidencias.
- AC6. La sección "Who commits and merges?" de `site/src/content/docs/flows/build.md` dice que el agente corre los hooks del repositorio, que solo salta uno con tu aprobación, y que avisa de los chequeos que un hook no pudo correr. *Falso en la base cuando* `rg -c -i 'hook' site/src/content/docs/flows/build.md` no encuentra coincidencias.

## Entrega

- **Modo:** sigue la preferencia guardada para Hive, que es una rama de prueba activa (N2-A). Los cambios se hacen en `feat/delegation-threshold-global`, sin commits hasta que los pidas, y el despliegue local solo ocurre a pedido. Tú pruebas en sesiones reales y decides el merge a `development`.
- **Revisión de código:** ninguna dedicada, porque son cambios solo de texto, según la misma preferencia.
- **Registro:** este registro de cambio queda sin versionar hasta el merge, como dice `flow-close`.
