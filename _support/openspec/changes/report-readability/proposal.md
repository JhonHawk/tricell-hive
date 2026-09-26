# Legibilidad de reportes: glosa de IDs y convención de formato

| Campo | Valor actual |
| --- | --- |
| Estado | En validación · implementado y verificado; commit, push, despliegue y archivo en curso |
| Tracker · GitHub Issues | Sin issue; sale de la investigación del 2026-09-26 sobre quejas del usuario |
| Git | `direct-base` · push a `rebuild/harness-engineering` · parada: lectura de la redacción |
| Verificación | `go vet ./...` · `go test -race ./...` · casos de regresión en `tests/pilot` · escaneo de fixtures · revisión de código |
| Siguiente paso | `flow-build`: T1–T3 en el hilo principal, T4 delegada |

## Objetivo

El usuario reporta dos problemas al leer las respuestas de los agentes en sesiones largas:

- **IDs sin glosa.** Aparecen preguntas como "¿Resolvemos D7 de una vez o terminamos la sesión?" sin decir qué es D7. La regla que lo evita existe (`content/guidance/global.md:17`, última frase), pero en ~1/3 de las citas a IDs de turnos anteriores revisadas no se aplicó. Los fallos se concentran en la pregunta nativa y en las preguntas de continuación y cierre.
- **Reportes desordenados.** La guía pide "short labeled paragraphs or a compact list" (`global.md:17`) y `flow-report/SKILL.md:8` pide mantener "status updates, and handoffs in prose". Nada orienta sobre cuándo usar tabla, lista o prosa, ni sobre aprovechar lo que cada CLI ofrece (lista de tareas, enlaces con etiqueta, páginas).

El cambio mueve la glosa a una regla propia que cubre todas las superficies y cada cita, y agrega una convención de formato por capacidad, con alternativa en texto, que funciona igual en los seis hosts.

## Alcance y aceptación

**Incluye:**
- `content/guidance/global.md`: regla propia de glosa (D1-A), regla de forma de reportes (D2-A) con estados en palabras (D3-A); ajuste del presupuesto de bytes.
- `content/skills/flow-report/SKILL.md:8`: deja de exigir prosa para los reportes en el chat.
- `_support/docs/architecture/agent-delivery.md`: mapa de capacidades de presentación por host, con fuentes y fecha.
- Dos criterios deterministas en `tests/pilot/regression.go` con fixtures derivados de sesiones reales: `cited_id_glossed` y `no_bare_url`.

**Excluye:**
- Diagramas Mermaid en el chat (se descartó D2-B): el modelo no sabe si su host los renderiza y en 3 de 6 hosts se vería el código.
- Pilotos con modelos: pausados por instrucción del usuario (`AGENTS.md`, Measurement). La mejora subjetiva de legibilidad no se afirma sin revisión humana.
- Despliegue a los hosts globales: se ofrece aparte al terminar.

**Criterios de aceptación:**
- A1. `global.md` exige glosa en cada cita de un ID fuera del mensaje que lo define, en texto, enunciado y opciones de la pregunta nativa, pregunta de cierre y relevos de subagentes; los rangos llevan glosa de grupo.
- A2. `global.md` define, para respuestas en el chat, la forma por contenido (tabla solo para ≥3 elementos × ≥2 atributos con ≤4 columnas, lista para paralelos, prosa para causas; manda el formato que prescriba un skill o plantilla), estados en palabras, todo enlace con etiqueta descriptiva en vez de URL suelta, diagramas solo en archivo o página y oferta de página en una línea cuando el usuario vaya a compartir o conservar el resultado. Nombra capacidades, no herramientas.
- A3. `flow-report/SKILL.md` ya no dice "in prose"; deja las respuestas, reportes de estado y handoffs fuera de HTML y remite a Communication de la guía global.
- A4. `agent-delivery.md` documenta, por host, lista de tareas, pregunta nativa, enlaces, diagramas y página publicable, con la etiqueta de evidencia de cada dato.
- A5. `cited_id_glossed` falla con el fixture derivado de la sesión `876d776d` (S1–S6 sin glosa) y pasa con el de `9d7ef3ee` (D2-A con glosa); `no_bare_url` falla con una URL suelta real y pasa con un enlace etiquetado.
- A6. `go vet ./...` y `go test -race ./...` pasan; el escaneo de fixtures del README no imprime nada; el presupuesto de `global.md` sube solo lo medido y justificado.

## Decisiones del usuario (2026-09-26)

- **D1-A (glosa en cada cita):** regla propia; todas las superficies; cada vez.
- **D2-A (convención por capacidad):** las reglas de forma en `global.md`, con alternativa en texto; el detalle por host va a `agent-delivery.md`. La regla de lista de tareas nativa no se duplica: ya vive en `flow-build/SKILL.md:46`, con el límite de que el plan guardado manda (revisión del plan, B3).
- **D3-A (estados en palabras):** Hecho, En curso, Bloqueado, Pendiente (usuario); sin emoji.
- **Ruta:** `flow-plan`.
- **Enlaces:** en la sesión de Claude Code 2.1.283 del usuario, `[issue #26390](url)` se mostró como etiqueta clicable. Los issues que decían lo contrario (#26390, #37808) están cerrados como "not planned" y duplicado.

## Entrega

Respuesta del usuario (2026-09-26):
- **Modo:** `direct-base`. Implementar y verificar, parar para que el usuario lea la redacción en el diff, y después commit (solo las rutas de este cambio, con su carpeta `openspec`) y push a `rebuild/harness-engineering`. El archivo del cambio va en un segundo commit con push.
- **Revisión:** `/code-review` de Claude Code sobre el diff, antes del commit.
- **Despliegue:** release nuevo a los hosts globales con el manager de Hive después del push, verificando los hashes.

El árbol de trabajo tiene cambios ajenos sin commitear (`versioned-installer-onboarding`, `tooling/**`, `VERSION`, `bootstrap.sh`); la entrega los preserva y commitea solo las rutas de este cambio.
