# Proporcionalidad de tests y TDD

| Campo | Valor actual |
| --- | --- |
| Estado | En validación · T1–T4 hechos y revisados; esperando que el usuario lea la redacción en el diff |
| Tracker · GitHub Issues | Sin issue; sale de la investigación del 2026-09-26 sobre sobreuso de tests |
| Git | `direct-base` · push a `rebuild/harness-engineering` · parada para que el usuario lea la redacción en el diff |
| Verificación | `go test ./tests/content/...` · tests de skills en Python · `go vet` y `go test -race` una vez · lecturas con `rg` · `/code-review` |
| Siguiente paso | Usuario: leer el diff; después commit de estas rutas y push a `rebuild/harness-engineering` |

## Objetivo

Hoy `flow-build` pide TDD red-green "por defecto" para cualquier comportamiento "automáticamente testeable", y la única excepción son los cambios de presentación o documentación pasiva (`content/skills/flow-build/SKILL.md:40`). Si hay tests pero no hay CI, pide además correr la suite completa en local (`:54`). Un modelo de ejecución que lee esto al pie de la letra escribe tests nuevos para renombres, cambios de configuración o cableado que el compilador ya cubre, y repite la suite completa. El Hive legacy tenía una excepción para cambios triviales, un valor `not-applicable` por tarea y reglas de alcance de ejecución, y el rebuild no los trajo.

La evidencia de la investigación del 2026-09-26 apunta en otra dirección:

- **Dónde está el beneficio de TDD:** viene de trabajar en pasos pequeños con ritmo constante más que de escribir el test primero (Fucci et al. 2016).
- **Qué cuestan los tests:** mantenimiento, tiempo de CI y revisión. Los "change detectors" se rompen con cada refactor sin atrapar bugs (Google Testing Blog).
- **Riesgo con agentes:** los agentes tienden a escribir tests que confirman su propio código o a editarlos para que pasen (METR 2025, Kent Beck 2025).
- **Guía oficial:** la [guía oficial de Claude Code](https://code.claude.com/docs/en/best-practices) pide "una forma de verificar" proporcional al cambio. Advierte contra los "tests for cases that can't happen" y sugiere correr tests puntuales en lugar de la suite completa.

El cambio acota TDD a comportamiento cuya falla importa, declara por tarea el enfoque de test y limita las corridas de la suite. Así el agente verifica siempre, pero solo conserva los tests que protegen algo.

## Alcance y aceptación

**Incluye:**
- `content/skills/flow-build/SKILL.md`: criterio de TDD por disparadores (D1-A), cambios mecánicos sin test nuevo, oráculo fijo, eliminación de tests que el cambio vuelve obsoletos, alcance de ejecución de la suite y proporcionalidad de la verificación.
- `content/skills/flow-plan/SKILL.md` y `references/plan-format.md`: enfoque de test por tarea.
- `content/agents/quality/test-engineer.md`: cada test nombra el cambio realista que lo haría fallar.
- `content/guidance/global.md:93`: se quita el ajuste opcional `TDD` (D2-A); `tests/content/budget_test.go` baja el presupuesto.

**Excluye:**
- **Caso de regresión en `tests/pilot`:** `AGENTS.md` (Measurement) lo exige cuando un cambio corrige una falla observada en una sesión real. Este cambio sale de la investigación: en la sesión `c1a83bea`, los ciclos RED/GREEN fueron proporcionales a unas 3,700 líneas con superficies de seguridad. Su desvío fue un salto de alcance sin preguntar, que queda registrado en Engram como incidente único.
- **Pilotos con modelos:** están pausados por instrucción del usuario. La reducción de tiempo de verificación no se afirma sin medirla.
- **`verification.md`:** su fila "Build iteration" ya pide "Run targeted tests" y su tabla ya dice que son candidatos, no un checklist.
- **`review-code`:** ya revisa que los tests puedan fallar (`review-code.md:12`).
- **Despliegue a los hosts globales:** se ofrece aparte al terminar.

**Criterios de aceptación:**
- A1. `flow-build` pide TDD solo para comportamiento que coincide con la lista de disparadores. Nombra los cambios mecánicos que no llevan test nuevo, da una rama residual (check directo o extender los tests del área) para el resto de cambios de comportamiento, y ancla el desempate al comportamiento: si el cambio altera un comportamiento de la lista, aplica TDD.
- A2. `flow-build` corre tests puntuales durante la iteración. La suite completa corre una vez sobre el candidato final cuando ningún CI la correrá, y no se repite sin una edición de por medio.
- A3. `flow-build` prohíbe editar un test que el usuario o el plan dieron como aceptación para hacerlo pasar, y permite eliminar un test que el cambio vuelve obsoleto comparando antes sus escenarios.
- A4. `flow-build` pide no agregar tests para casos que el cambio no puede producir, y reportar en el progreso cuando la verificación necesaria es mucho mayor que el cambio.
- A5. `flow-plan` pide un enfoque de test por tarea que cambia comportamiento (`tdd`, `characterization` o `check` para todo lo que queda fuera del criterio de TDD, nombrando el check) y la plantilla de tarea lo incluye con la misma definición; `flow-build` obedece el enfoque declarado.
- A6. `test-engineer` exige nombrar, antes de escribir cada test, el cambio realista que lo haría fallar.
- A7. `global.md` ya no lista `TDD` y `globalGuidanceBudget` coincide con el nuevo tamaño.
- A8. `go test -count=1 ./tests/content/...`, los tests de skills en Python, `go vet ./...` y `go test -race -count=1 ./...` pasan, salvo fallas preexistentes del trabajo en curso ajeno, que se reportan con su salida.

## Decisiones del usuario (2026-09-26)

- **Ruta:** `flow-plan` con P1–P6 de la investigación.
- **D1-A (disparadores):** TDD solo con lógica con ramas, contratos o datos persistidos, seguridad, permisos o dinero, o un bug que puede volver; un cambio mecánico no lleva test nuevo.
- **D2-A (quitar `TDD`):** se elimina de los ajustes opcionales; un proyecto que quiera TDD estricto lo escribe en su propio `AGENTS.md`.

## Entrega

Respuesta del usuario (2026-09-26):
- **Modo:** `direct-base`. Implementar y verificar, y parar para que el usuario lea la redacción en el diff. Después, un commit con solo las rutas de este cambio y esta carpeta `openspec`, con push a `rebuild/harness-engineering`. El archivo del cambio va en un segundo commit con push.
- **Revisión:** `/code-review` de Claude Code sobre el diff, antes del commit.
- **Despliegue:** fuera de este plan; se ofrece al terminar.

El árbol de trabajo tiene cambios ajenos sin commitear de otra sesión activa (`versioned-installer-onboarding`, `tooling/**`, `VERSION`, `bootstrap.sh`, `_support/docs/architecture/deployment-manager.md`). La entrega los preserva y agrega al commit solo las rutas de este cambio.
