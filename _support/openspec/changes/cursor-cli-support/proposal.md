# Soporte de Cursor CLI en Hive

| Campo | Valor actual |
| --- | --- |
| Estado | En validación · T1–T3 verificados en el working tree; diff pendiente de tu revisión; T4 pendiente de autorización |
| Tracker · GitHub Issues | Sin issue vinculado |
| Git | `hold` · working tree de `rebuild/harness-engineering` · sin commits; el diff queda para tu revisión |
| Verificación | `go test ./...` · `-race` · `go vet` · prueba Python de `harness-audit` · `/code-review` · in vivo (T4, autorización aparte) |
| Siguiente paso | Revisar el diff; luego decidir entrega y T4 |

## Objetivo

El CLI de Cursor ya descubre los skills de Hive, pero no recibe la guía global: no carga `~/.claude/CLAUDE.md` ni `~/.cursor/rules/*.mdc`, y su único canal global documentado (User Rules en Settings) no es un archivo que el gestor pueda mantener. Tampoco lista los agentes de nivel usuario, aunque su documentación los soporta. Una sesión de Cursor trabaja hoy sin las reglas de autorización, secretos, artefactos y comunicación de Hive.

El cambio agrega Cursor como sexto host del gestor en scope usuario: un bloque gestionado con `content/guidance/global.md` en `~/.cursor/AGENTS.md`, el registro como consumidor de los skills compartidos y los roles de Hive en `~/.cursor/agents/`. Cada proyecto que se use con Cursor agrega a su `AGENTS.md` una línea condicional que manda leer ese archivo cuando la guía no está en contexto. Lo único probado hasta ahora es el mecanismo, con otra redacción y otro destino: una línea incondicional que mandaba leer `~/.claude/CLAUDE.md` hizo que Cursor leyera la guía y la aplicara, y el control sin línea no. El texto y el destino actuales se prueban en T4.

## Alcance y aceptación

Incluye:

- Adaptador `cursor` en el gestor, solo scope usuario.
- Roles de Hive renderizados en `~/.cursor/agents/` con el formato de Cursor, y Cursor como sexto host de `integrations/agent-profiles.json` (D7-C).
- R2 (D10-A, 2026-09-25): la tabla "Native role selection hints" sale de `global.md` y queda una regla por capacidad ("A role the tool does not list counts as lacking role selection; read its contract from the installed role file."), probada en los seis hosts en el [#30](https://github.com/JhonHawk/tricell-hive/issues/30). La tabla, con la fila de Cursor, pasa a `agent-delivery.md` como referencia de mantenimiento.
- Texto canónico del puntero y su comprobación en `harness-audit`.
- Documentación del gestor, la distribución y la lista de hosts objetivo.
- Verificación in vivo en Cursor con dos modelos, y comprobación de que el puntero no duplica la guía en Claude Code ni Codex.

Excluye, con su razón:

- Copiar los roles a cada proyecto para esquivar el bug del CLI 2026.09.18, que no lista agentes de nivel usuario: mientras dure, el hijo genérico lee el contrato desde `~/.cursor/agents/`.
- Scope proyecto del gestor para Cursor: D2-A deja el puntero como edición manual en la sesión de cada proyecto.
- Escribir el puntero en otros proyectos (ark, sample-project…): son proyectos externos; se hace en su propia sesión.
- User Rules de Cursor Settings: descartado en D1.

Criterios de aceptación:

- A1. `go run ./tooling/cli plan install --hosts cursor --scope user --home <tmp-home> --state-dir <tmp-state>` propone un bloque en `<tmp-home>/.cursor/AGENTS.md`, los skills en `<tmp-home>/.agents/skills/` y un `[agent]` por rol en `<tmp-home>/.cursor/agents/<role>.md`, con `readonly: true` en los roles de acceso `observe`; con `--scope project --root <dir>` falla con el error de scope no soportado.
- A2. Install, update, remove, status y recover de Cursor pasan en home sintético, incluidos la convivencia con otros consumidores de los skills compartidos, el rechazo de actualizar un skill compartido sin todos sus consumidores y la preservación de un `~/.cursor/` ajeno.
- A3. `go test ./...`, `go test -race ./...` y `go vet ./...` pasan.
- A4. `global.md` coincide con la variante B probada en el #30 (sin tabla por host) y el presupuesto de `tests/content/budget_test.go` baja a su tamaño nuevo.
- A5. `harness-audit` tiene la regla HA-IF-18 con el texto canónico del puntero, su disparador, severidades y `keep` para un puntero existente, activa en el catálogo de reglas. Su conducta en una auditoría real no se prueba mientras sigan pausados los pilotos.
- A6. Comprobación de humo: en Cursor, con la guía desplegada y el puntero, dos modelos distintos leen `~/.cursor/AGENTS.md` y responden según Hive; los controles sin puntero no.
- A7. Comprobación de humo: en las sesiones principales de Claude Code y Codex, con el puntero cargado (canario presente), la sesión no lee `~/.cursor/AGENTS.md`.
- A8. Comprobación de humo: con los roles desplegados, el CLI de Cursor no lista ningún rol de Hive dos veces. Si el CLI sigue sin listar roles de usuario, un hijo genérico recibe el contrato leído de `~/.cursor/agents/<role>.md`.

## Entrega

- Decisiones del usuario (2026-09-25): D1-A puntero condicional y ruta propia de Cursor; D2-A puntero escrito a mano en cada proyecto con aviso de `harness-audit`; D4-A ajustes `## Hive` de este repo; D5-D `hold`; D6-A `/code-review` de Claude Code; D7-C roles desplegados en `~/.cursor/agents/`; D8-C seguir en `hold`; D10-A aplicar R2 dentro de este cambio.
- Repositorio y base: `tricell-hive`, `rebuild/harness-engineering`.
- Modo `hold`: se implementa y verifica en el working tree, sin commits ni push. El diff queda para tu revisión y la entrega sigue pendiente. Este registro del plan tampoco se versiona.
- Revisión de código: `/code-review` de Claude Code sobre el diff al terminar T1–T3, antes de dar por verificado el working tree. Ejecutarla está incluido en la implementación autorizada.
- Fuera de este modo: el despliegue en tu configuración real (solo Cursor o los seis hosts, ver las dos rutas de T4) y las corridas de modelos de T4 necesitan autorización explícita aparte (la pausa de pilotos de `AGENTS.md` sigue vigente), además del diff aceptado y entregado.
- Punto final acordado: working tree verificado con el diff presentado.
