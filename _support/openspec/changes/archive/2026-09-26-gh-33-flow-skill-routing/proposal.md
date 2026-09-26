# Ruta a flow-build en pedidos conversacionales

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · T1–T3 completos; integrado en `rebuild/harness-engineering` (`de3235e`) y desplegado |
| Tracker · GitHub Issues | • [#34 — límites del modelo de trazas](https://github.com/JhonHawk/tricell-hive/issues/34)<br>• [#33 — Despliegue sin recorrido funcional: producción en blanco (Grok, globex)](https://github.com/JhonHawk/tricell-hive/issues/33) |
| Git | `direct-base` (D21-A) · commit `de3235e` con push a `rebuild/harness-engineering` |
| Verificación | `go test ./...` · `-race` · `go vet` · caso de regresión nuevo · `/code-review` |
| Siguiente paso | Ninguno; los límites quedan en #34 |

## Objetivo

La sesión de Grok en globex (`01a0daf9`, release 6465f0e9d90c) cambió código, abrió e integró PR y desplegó a producción con pedidos conversacionales: "haz un pr", "despliega en el server prod". Leyó `git-workflow` (línea 306 de la traza) pero nunca `flow-build` ni su referencia `verification.md`. Por eso:

- **G5:** no hizo el recorrido funcional después de desplegar ni revisó la URL del API compilada en el frontend, y la página de alta quedó en blanco en producción.
- **G6:** integró el hotfix sin esperar CI ni revisión, con `git-workflow` ya leído. Esto es incumplir una regla cargada y queda en observación: este cambio no lo corrige.

La regla del recorrido existe, pero solo en `flow-build`. `global.md:77` manda usar `flow-research` para investigar, y no tiene una regla equivalente para implementar y entregar.

El cambio agrega esa regla de ruta en la guía global y un criterio determinista que detecta entregas de Git hechas sin leer `flow-build` antes.

## Alcance y aceptación

Incluye:

- Una viñeta en `content/guidance/global.md`, junto a la de `flow-research`, con este texto: "Before a commit, pull request, merge, promotion, or deployment, including on a conversational request, read and use `flow-build`. Mechanical edits alone do not require it; subagents follow their brief." Sin la mención a `git-workflow`, porque `flow-build` ya envía allí (hallazgo 3 de `/code-review`). (D18-A)
- El criterio `flow_skill_read_before_delivery` en `tests/pilot/regression.go`, con un caso que falla (globex G6) y otro que pasa (D19-A).

Excluye:

- Copiar a `global.md` las reglas del recorrido tras desplegar (D18-B descartada: duplicaría una regla).
- Piloto con modelos (D20-A): el efecto se mide con el monitoreo de sesiones reales.

Criterios de aceptación:

- A1. `global.md` tiene la viñeta de ruta a `flow-build` y `git-workflow` junto a la de `flow-research`. La prueba de presupuesto pasa con el aumento medido con `wc -c`.
- A2. `flow_skill_read_before_delivery` falla cuando un intento de acción de Git (`git commit`, `git push`, `gh pr create`, `gh pr merge`) no tiene antes una lectura de `flow-build/SKILL.md` con contenido. Cuenta como lectura una de tipo `read`, un `cat` en el shell (Codex) o una carga nativa con contenido, igual que `observeSkill`. Pasa cuando la hay, y queda como no aplicable cuando no hay acciones de Git. Límites declarados: no ve las skills cargadas con un comando escrito (`/flow-build` en Claude, `$flow-build` en Codex), ni despliegues que no pasen por Git (G5), y no se conecta a `assessFlows`.
- A3. Hay casos que fallan y que pasan, sanitizados y con su procedencia, y la reversión del criterio hace fallar la prueba.
- A4. `go test ./...`, `-race` y `go vet` pasan, y `/code-review` queda sin hallazgos abiertos.
