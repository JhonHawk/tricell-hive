# Secuencia de cierre: limpieza, reporte y pregunta

| Campo | Valor actual |
| --- | --- |
| Estado | En progreso · plan revisado en 2 rondas; implementación autorizada (flow-build, 2026-09-26) |
| Tracker · GitHub Issues | Sin issue; sale del monitoreo de sesiones del 2026-09-25 (O6, G3, G4) |
| Git | `hold` (D14-A) · verificado en el working tree, sin commits · se publica junto con `subagent-effort-profiles` tras tu revisión |
| Verificación | `go test ./...` · `-race` · `go vet` · casos de regresión en `tests/pilot` · piloto acotado Codex y Grok (D13-B) · `/code-review` |
| Siguiente paso | T1 (hilo principal) y T2 (`test-engineer`) en paralelo |

## Objetivo

Con el release 6465f0e9d90c, dos sesiones en hosts que no son Claude cerraron trabajo sin los pasos finales que exige la guía:

- **Codex, sample-project 01a0dab1 (X1):** el reporte de ARK-687 terminó el turno sin la pregunta de cierre.
- **Grok, globex 01a0daf9 (G3 y G4):** el reporte del despliegue a producción terminó el turno sin la pregunta de cierre. Tampoco hubo limpieza: dejó la rama integrada, porque hizo el merge con `--delete-branch=false` y nunca la borró, y no reportó qué pasó con el `git bundle` temporal.

Las dos sesiones tenían la regla cargada, igual que `flow-build`. Hoy los pasos finales están repartidos en tres lugares:

- la limpieza, en "Retention and cleanup" de `global.md`;
- la limpieza de ramas después del merge, en `git-workflow`;
- la pregunta de cierre, en una frase al final de un párrafo largo de Communication (`global.md:22`) y en media oración al final de `flow-build`.

Los modelos de ejecución terminan el turno en cuanto escriben el reporte. OpenCode y la sesión de ark en Codex sí preguntaron, así que el mecanismo funciona cuando el modelo lo tiene presente al final.

El cambio junta los pasos finales en una secuencia explícita, que sea el último paso de `flow-build`, y le da a la pregunta de cierre una viñeta propia en la guía global. Además, lo mide con casos de regresión y con un piloto acotado.

## Alcance y aceptación

Incluye:

- Viñeta propia para la pregunta de cierre en `content/guidance/global.md` (D11-A).
- Paso final explícito en `content/skills/flow-build/SKILL.md`, en este orden: limpiar (temporales, procesos y ramas después del merge según `git-workflow`), reportar con una línea de limpieza, y preguntar como último acto del turno (D11-A, D12-B).
- Criterios deterministas nuevos en `tests/pilot` para la pregunta de cierre y para la limpieza de ramas integradas, con casos de Codex y Grok derivados de X1, G3 y G4.
- Variante de guía por brazo en el runner de `tests/pilot` para los CLIs instalados (`--guidance-source`, `--arm A|B`), con el aislamiento de Engram existente, y un caso de piloto `close-sequence` (D15-A).
- Piloto acotado en Codex y Grok que compare la guía actual con la nueva (D13-B).

Excluye, con su razón:

- Una nota específica para el modo por defecto de Codex (D11-B descartada): con un solo caso de Codex no hay evidencia suficiente.
- Desplegar en los CLIs globales: se decide junto con `subagent-effort-profiles`.

Criterios de aceptación:

- A1. La pregunta de cierre tiene su propia viñeta en `global.md`, después de los elementos de cierre y con su condición definida una sola vez, y las opciones y los criterios de sesión nueva no cambian. La prueba de presupuesto pasa.
- A2. `flow-build` termina con la secuencia de cierre en orden: limpiar (incluida la limpieza de ramas después del merge), reportar con una línea de limpieza y preguntar.
- A3. `close_question_after_report` y `merged_branch_deleted` existen, cada uno con casos de Codex y Grok que fallan y que pasan, derivados de X1, G3 y G4, y sanitizados según el README de regresión.
- A4. El modelo de trazas reconoce la herramienta `question` de OpenCode, y los textos de Codex y OpenCode llevan `Role` (OpenCode también `Message`), sin cambiar el resultado de S1 en los casos actuales. En Codex la pregunta solo se observa como texto, porque `codex exec` 0.157.0 no admite `request_user_input`.
- A5. El runner corre `close-sequence` con `--guidance-source` y `--arm A|B` en Codex y Grok. Registra los hashes de la guía, aísla Engram y limpia los enlaces simbólicos de autenticación, y sus pruebas unitarias lo cubren.
- A6. El piloto hizo sus 8 corridas con Engram aislado y limpio, y hay una tabla que compara A y B. No se afirma una mejora: con dos corridas por celda solo se decide con evidencia.
- A7. `go test ./...`, `go test -race ./...` y `go vet ./...` pasan.
