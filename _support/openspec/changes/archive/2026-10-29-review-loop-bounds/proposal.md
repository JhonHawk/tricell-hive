# Límites del bucle de revisión de código

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado por PR #286 (merge 051471ec, 2026-10-29) |
| Tracker · GitHub Issues | Sin issue: el usuario no pidió abrir uno |
| Git | automatic · PR a `development` y merge · sin parada humana |
| Verificación | tests de contenido · suite completa una vez · `/code-review` |
| Siguiente paso | Ninguno; archivado |

## Request

> Evaluando las sesiones de la sultimas horas, creo que lo que más tiempo esta tomando es el CI/CD y las revisiones -> retrabajo -> nueva revisión, qué opinas? haz un research al respecto

Rutas elegidas en la sesión (2026-10-29): D1-A (medir las segundas revisiones), D2-A (planear F2+F3+F4; la elección aprueba las correcciones de F4), D3-A, D4-A, D5-A y D6-A (ver Entrega y `design.md`).

## Objective

La medición de las sesiones posteriores a hive-core mostró que en tricell-hive, que no tiene CI, la revisión y su retrabajo ocupan el 28 % del tiempo del agente y la mitad del tiempo por ticket. La causa principal es la combinación de un revisor nativo orientado a recall, que apunta a 5–10 hallazgos sin verificarlos, con una regla que lo trata como verificador, no exige registrar la severidad cuando el revisor no la da, deja que la segunda ronda arregle P2 y permite una tercera ronda por ediciones posteriores.

El cambio ajusta la regla de la revisión dedicada para que la segunda ronda sea rara, acotada y terminal: los P0/P1 de un revisor que no verifica pasan por el refutador, la severidad queda registrada, la segunda ronda solo arregla P0/P1 del arreglo, y después de ella no hay otra revisión.

## Scope and acceptance

Incluido: la sección «Dedicated code review» de `content/skills/flow-verify/references/review-and-gates.md` y cada texto que repita o dependa de lo que cambia.

Excluido, con destino:
- Cadencia de la suite completa en las rondas (H6): en observación; se reabre si otra sesión con la regla nueva repite más de dos suites completas sobre cambios solo de `content/`.
- Espera del CI tras el merge y re-corrida por base movida en cam (H1, H2): reporte para la sesión de cam (correr el job de integración también en los PR).
- Tope de tamaño al lote final sin revisar: descartado; la cláusula de contrato ya manda esos casos al usuario.
- Revisión de las correcciones de la validación humana (`review-and-gates.md:51`): fuera de alcance; es una etapa distinta, pedida por el usuario.

Búsqueda de requisitos: en `_support/openspec/specs/` (context-delivery, documentation-site, versioned-installation, walkthrough-viewer) se buscaron `re-review`, `refut`, `severity`, `verify each finding`, `verifies its own`, `second pass` y `code-review`; nada coincide. El cambio solo altera redacción de guía, sin requisito declarado: no lleva delta.

Restricciones: `review-and-gates.md` queda en 16 000 bytes o menos (`tests/content/skill_size_test.go`); la regla nombra la capacidad y no un comando de host (AGENTS.md); `content/guidance/global.md` no cambia.

- AC1. La regla define cuándo una revisión verifica sus hallazgos por lo que su salida declara de cada hallazgo, y no nombra ningún comando nativo como verificador. *Falso en la base cuando* `review-and-gates.md:25` dice que `/code-review` en medium o más ya verifica cada hallazgo.
- AC2. La regla exige asignar P0–P3 a cada hallazgo cuando el mecanismo no da severidad, registrar el nivel en el reporte y nombrar el arreglo P0/P1 que inicia la segunda revisión. *Falso en la base cuando* la viñeta «Severity» solo cubre mecanismos con niveles propios.
- AC3. De la segunda revisión solo se arreglan sus P0/P1, refutados o verificados; sus P2/P3 se reportan sin arreglar. *Falso en la base cuando* la viñeta «Re-review» manda arreglar «a finding still open or newly reported» bajo las reglas generales, que incluyen P2.
- AC4. Una edición posterior entra en el diff de la segunda revisión si esta aún no corrió; después de ella no hay otra revisión: se reporta sin revisar si es local, dentro del alcance y sin contrato, o se pregunta al usuario. *Falso en la base cuando* la viñeta «Later edits» manda revisar el diff desde el commit revisado cuando una respuesta de revisión lo cubre, sin límite de rondas.
- AC5. Si el candidato revisado no tenía commit, sin importar el modo, la segunda revisión y las ediciones posteriores se miden contra su identidad de candidato. *Falso en la base cuando* la identidad de candidato solo se nombra «under `hold`».

## Delivery

Ver `tasks.md`, sección Entrega.
