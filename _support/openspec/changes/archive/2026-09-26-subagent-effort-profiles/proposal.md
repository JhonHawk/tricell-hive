# Esfuerzo explícito y portable por rol

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · T1–T5 completos; integrado en `rebuild/harness-engineering` (`16a7263`) y desplegado |
| Tracker · GitHub Issues | Sin issue propio; hallazgo incidental en • [#32 — dialecto de delegación de OpenCode V2](https://github.com/JhonHawk/tricell-hive-private/issues/32) |
| Git | Commit `16a7263` con push a `rebuild/harness-engineering` (autorizado por el usuario el 2026-09-26) |
| Verificación | `go test ./...` · `-race` · `go vet` · home sintético (T4) · `/code-review` (D7-A) |
| Siguiente paso | Ninguno en este cambio |

## Objetivo

Seis roles de ejecución (`backend-developer`, `frontend-developer`, `kotlin-multiplatform-developer`, `devops-engineer`, `test-engineer`, `sdd-spec-writer`) no declaran esfuerzo en Claude Code y heredan el de la sesión. Cuando alguien baja a `/effort low` para trabajar de cerca con el agente, como recomienda el artículo [Spending your effort](https://claude.dev/blog/spending-your-effort/), esos roles también bajan a `low`, aunque trabajan sin nadie que los corrija sobre la marcha. Además, el esfuerzo por rol (`claude_effort`) solo existe en Claude. Si `review-security` sube a `max`, que es donde el artículo muestra la mayor ganancia (64 % → 87 % en tareas de seguridad), el cambio no tendría efecto en Codex ni en Pi.

Con este cambio, el esfuerzo por defecto de cada rol sale de su perfil en todos los hosts que lo permiten, y solo los roles que se apartan lo declaran, con un campo `effort` que cada host traduce a su escala.

## Alcance y aceptación

Incluye:

- `effort` en lugar de `claude_effort` en el formato de los roles, con su traducción a Claude, Codex y Pi (D3-A, D5-A).
- Esfuerzo `high` en el perfil de ejecución de Claude y excepciones declaradas solo en cuatro roles (D2-A, D4-A).
- Documentación del contrato y del límite para volver a releases anteriores.
- Fusionar `cloud-architect` en `solution-architect` (D8-B). En el conteo de D1-A, `cloud-architect` no tuvo llamadas en ninguno de los seis hosts.

Excluye, con su razón:

- Quitar `kotlin-multiplatform-developer`: no tuvo llamadas en ningún host, pero se mantiene por decisión (D8-B).
- Esfuerzo por rol en Grok, Cursor y OpenCode. Grok no tiene el campo. En Cursor el esfuerzo va unido a un modelo concreto (`<model>[effort=high]`) y Hive usa `inherit` por los planes de suscripción. En OpenCode es una variante fija, `#max`. En esos tres hosts las excepciones no tienen efecto.
- Cambiar el perfil de razonamiento de Codex y Pi (Astra/`medium`). La [documentación de Codex](https://learn.chatgpt.com/docs/agent-configuration/subagents) recomienda empezar Astra en `low`, así que `medium` ya está por encima.
- Desplegar la versión nueva en los CLIs globales: se pregunta al terminar.
- Caso de regresión en `tests/pilot`: H2 se infirió de la configuración y la documentación, no se observó como falla en una sesión real, así que la regla de regresión del `AGENTS.md` no aplica.

Criterios de aceptación:

- A1. Ningún rol renderizado para Claude queda sin `effort`.
- A2. Los niveles renderizados coinciden con la tabla de `design.md` para los 19 roles en los seis hosts: `review-security` `max`, `state-fetcher` `low`, `review-plan` y `sdd-spec-writer` `medium`, el resto según su perfil. En Codex la clave es `model_reasoning_effort` y en Pi `thinking`, con la equivalencia de T1. Grok, Cursor y OpenCode no emiten esfuerzo.
- A3. Un rol con `claude_effort` falla la validación con un error de campo desconocido.
- A4. `go test ./...`, `go test -race ./...` y `go vet ./...` pasan.
- A5. En un home sintético, una instalación hecha desde la base actual con su propio binario se actualiza con el working tree sin conflictos y `status` queda limpio. Con el binario nuevo, un `plan install --release` del release de base falla con `unsupported agent field "claude_effort"`, y ese límite queda documentado.

## Entrega

- Modo `hold` (D6-C, 2026-09-25): implementar y verificar en el working tree, sin commits ni push. Esta carpeta de cambio también queda sin versionar hasta que decidas la entrega.
- Revisión del código: `/code-review` de Claude Code (D7-A) antes de entregar el diff.
- Despliegue global: fuera de este modo; se pregunta al terminar.
