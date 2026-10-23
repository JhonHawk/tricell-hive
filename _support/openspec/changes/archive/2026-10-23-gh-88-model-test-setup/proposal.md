# Reducir el coste de preparación de las pruebas de modelos

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` mediante PR #98 el 2026-10-23 |
| Tracker · GitHub Issues | • [#88 — pruebas del CLI superiores a cinco minutos](https://github.com/JhonHawk/tricell-hive/issues/88) |
| Git | PR #98 integrado · sin revisión dedicada · checks y verificación independiente aprobados |
| Verificación | Caracterización existente · medición por prueba · regresión Go con detector de carreras |
| Siguiente paso | Ninguno de implementación; registro archivado por la ruta de cierre autorizada |

## Solicitud

El usuario eligió, el 2026-10-23:

> D1-A Investigar #88 (Recommended)

Después de recibir la medición, eligió:

> D2-A Preparar plan (Recommended)

La segunda opción proponía «Usar flow-plan para optimizar la preparación de pruebas, conservar cobertura y definir mediciones comparables; sin implementar aún».

Resultado esperado del ticket, leído el 2026-10-23:

> Una explicación sustentada por mediciones, con los tests responsables y sus tiempos, y una recomendación: optimización verificable o coste esperado documentado. Este ticket no autoriza eliminar cobertura, debilitar comprobaciones ni iniciar una optimización automáticamente.

## Objetivo y alcance

Reducir el trabajo de instalación que se repite para preparar pruebas de modelos. El código productivo y sus validaciones permanecen iguales. La oportunidad principal es el caso de 22 roles en seis hosts: sus dos escenarios visuales consumen 0,43 segundos, pero su preparación consume aproximadamente 31,78 segundos.

Se dividió únicamente esa preparación en planes de un host y se minimizaron los hosts de diez pruebas que sólo comprueban uno por escenario. Se conservan instalaciones reales, directorios temporales independientes y todos los escenarios funcionales. La corrida local final bajó de 316,680 a 274,062 segundos; no se extrapola ese resultado a otros equipos.

### Criterios de aceptación

- **AC1.** La prueba de catálogo grande prepara los seis hosts mediante seis planes reales de un host; cada uno presenta los 22 roles y conserva scroll, navegación y anchura en 80×24 y 120×40. *Falso en la base:* usa una instalación combinada de seis hosts y no comprueba explícitamente el orden esperado de todos ellos.
- **AC2.** En una corrida local comparable, sin caché de resultados y con detector de carreras, `tooling/cli` termina en menos de 300 segundos de duración de paquete, con el mismo conjunto de escenarios. *Falso en la base medida:* 316,680 segundos. Es un objetivo local, no una garantía para otros equipos ni una afirmación estadística de fiabilidad.
- **AC3.** Los diez casos enumerados en `design.md` preparan sólo el host que ejercen, conservando proveedores, tamaños, resultados y assertions. *Falso en la base:* instalan entre dos y cinco hosts por caso, o cuatro por subcaso de listas.

### Restricciones que deben seguir cumpliéndose

- No modificar `tooling/management`, `integrations/`, dependencias, comportamiento del CLI, validaciones por escritura, permisos ni cobertura de carreras.
- No reducir los 22 roles, los seis hosts ni los dos tamaños del caso grande. No alterar goldens o expectativas para acomodar una diferencia funcional.
- No cambiar globalmente `modelsEditHosts`, los helpers de instalación ni el significado limpio e independiente de `twin`.
- Conservar las pruebas de instalación combinada, onboarding, receipts, equivalencia interfaz/comando y aislamiento entre hosts.
- No clonar instalaciones, reescribir rutas/hashes o fabricar `State`; no introducir una caché o un nuevo sistema de fixtures.
- Si la técnica no mejora el caso aislado o no alcanza AC2 (duración local inferior a cinco minutos), registrar el criterio incumplido y volver a investigación; no ampliar cambios productivos por inferencia.

No hay delta de requisitos: se modifica exclusivamente la preparación de pruebas, sin cambiar contratos del producto.

## Entrega y autorización

El 2026-10-23 el usuario eligió D3-A (entrega automática), D4-A (sin revisión dedicada) y D5-A (iniciar implementación). Autoriza rama, commits, push, PR y merge a `development`, más un PR pequeño de archivo integrado por el agente. Checks y verificación independiente siguen obligatorios. No hay parada humana antes del push para este cambio de pruebas. La regla del checkout mantenedor cubre la reconstrucción local del binario después de integrar; no reiniciar visor ni promover a `master` o publicar release.

Inicio de implementación (histórico): se creó `perf/gh-88-model-test-setup` en el worktree `_support/workspace/worktrees/gh-88-model-test-setup`, necesario por `_support/sessions/` previo sin seguimiento en el checkout principal. El registro se trasladó a ese worktree y se archivó una sola vez al cierre. Durante la implementación, issue #88 (duración del CLI) permaneció OPEN: GitHub no ofrece estado intermedio en este repositorio, no tiene Project asociado ni subissues; no se inventó un estado o label compartido. Su estado final es CLOSED, como consta abajo.

### Resultado de entrega

[PR #98 — preparación de pruebas de modelos](https://github.com/JhonHawk/tricell-hive/pull/98) integrado en `development` mediante `3e5d789e105316b8ce62194a7082ab843b33f11d`. El issue #88 (duración del CLI) pasó a CLOSED automáticamente; no tiene subissues. Sin workflows ni ejecuciones CI registradas; el contexto automático CodeRabbit informó SUCCESS, sin revisión dedicada solicitada o ejecutada. Los checks locales y las verificaciones independientes están en `tasks.md`.

Registro archivado, sin delta de requisitos, por un PR pequeño autorizado que integra el agente. La reconstrucción local del binario desde `3e5d789e` terminó correctamente; no se reinició el visor ni se modificó contenido distribuido. No se promovió a `master` ni se publicó release. La limpieza retira las dos ramas y el worktree sólo después de comprobar merges, ausencia de PR dependientes y árbol limpio; conserva `_support/sessions/` previo y las mediciones locales únicas.
