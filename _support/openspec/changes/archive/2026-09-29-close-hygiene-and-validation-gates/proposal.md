# Higiene al cierre, correcciones sin gate intermedio y tests acotados

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · PR [#67](https://github.com/JhonHawk/tricell-hive-private/pull/67) mergeado en `rebuild/harness-engineering` (`343d694`) |
| Tracker · GitHub Issues | Sin issue; nace de la auditoría de sesiones ark/sample-project del 2026-09-29 |
| Git | Automático · PR a `rebuild/harness-engineering`, merge tras T5 · sin parada humana |
| Verificación | `rg` por criterio · `go test ./...` · `go vet ./...` · `update --dry-run` (catálogo y enlaces) |
| Siguiente paso | Ninguno. Fuera: el merge del rebuild a `master` |

## Objetivo

Hoy la higiene de una tarea está repartida en cinco archivos y se ejecuta a destiempo. El plan se commitea y se pushea al quedar listo, en cada hito y al archivar: en ark y sample-project eso son tres o cuatro pushes con bypass de administrador a `master` de specs por ticket, antes incluso de empezar el build. Las reglas de limpieza (ramas, temporales, procesos, capturas, cierre del registro) viven en `global.md`, `flow-build`, `git-workflow`, `change-records.md` y `delivery-decisions.md`, y la auditoría del 2026-09-29 encontró huecos en casi todos los hosts: 80 capturas olvidadas en sample-project, temporales en `/tmp`, hallazgos sin proponer y reportes después de la pregunta de cierre.

Además, cada corrección que el usuario pide durante la validación vuelve a lanzar `hive-review-ux`, `hive-verify-change` y `hive-verify-task`, y los agentes gastan tiempo en suites completas repetidas, en pasadas de verificación de 3 a 6 minutos para menos de 5 s de tests y en esperas de CI de 7 a 8 minutos, sin reportarlo como problema.

Este cambio crea la skill `flow-close`, que concentra toda la higiene y versiona el registro del cambio una sola vez al cierre; generaliza las rondas de corrección sin gate intermedio con un único gate acotado al aceptar; acota los tests a lo afectado; y convierte los atascos de tests y de verificación en hallazgos que el usuario puede abrir como ticket, declinar o silenciar para siempre.

## Alcance y aceptación

Incluye la guía distribuida en `content/` (una skill nueva y cambios en `global.md`, `flow-build`, `verification.md`, `git-workflow`, `flow-plan`, `change-records.md`, `delivery-decisions.md`, `workspace-conventions`), la lista de skills protegidas del piloto y la documentación de arquitectura que describe el cierre.

Queda fuera: los scripts `test:related` de ark y sample-project (son de esos proyectos; se reportan en sus sesiones), la configuración de perfiles de modelo (el problema P2 de verificación con el mismo modelo va en otro cambio) y cualquier regresión nueva en `tests/fixtures/regression/` (AGENTS.md).

- AC1. Existe la skill `flow-close` con dos disparadores (cierre de un ítem de trabajo y fin de sesión) y su descripción la activa cuando el usuario pide higiene. *Falso en la base cuando* `content/skills/flow-close/SKILL.md` no existe.
- AC2. Ninguna guía distribuida pide commitear o pushear la carpeta del cambio al quedar listo el plan ni en los hitos; el estado se actualiza en disco. *Falso en la base cuando* `delivery-decisions.md` dice "commit and push it when the plan is ready, at each delivery milestone" y `flow-plan/SKILL.md` dice "commit and push it now".
- AC3. `flow-close` dice cómo se versiona el registro una sola vez al cierre, siempre después de integrar el código y de que pase el CI posterior al merge: en un repo de specs `direct-base`, un commit y push; con specs dentro del repo de código, un commit a la base por la vía que el repo permita (push directo o un PR pequeño de cierre sin revisión dedicada); con `mv` cuando la carpeta nunca estuvo en Git; una pausa no archiva y un abandono archiva con su razón. *Falso en la base cuando* ninguna guía describe el versionado único al cierre.
- AC4. La limpieza posterior al merge y el cierre del registro tienen un solo hogar: `flow-close`. *Falso en la base cuando* `rg -n "git fetch --prune" content/` encuentra `git-workflow/SKILL.md` y `rg -n "^## Close a change" content/` encuentra `change-records.md`.
- AC5. Durante una pausa de validación, cualquier corrección que pida el usuario (texto, comportamiento o layout) se aplica sin `hive-review-ux`, `hive-verify-change`, `hive-verify-task` ni revisión de código entre rondas, y al aceptarla corre un único gate acotado a la diferencia entre el último candidato revisado y el aceptado. *Falso en la base cuando* `verification.md` limita la excepción a "layout adjustments" y no acota el gate final.
- AC6. La regla de frecuencia de tests pide la selección de tests afectados del proyecto, amplía a la suite del paquete o del repo con disparadores concretos, trata una selección vacía como no verificada y no repite una suite completa sobre un candidato sin cambios salvo como ejercicio declarado. *Falso en la base cuando* `rg -n -i "selects no test|unchanged candidate" content/skills/flow-build/` no encuentra nada.
- AC7. La re-verificación de `hive-verify-task` revisa solo los criterios marcados `not met`. *Falso en la base cuando* el paso 3 de `flow-build` lanza "one re-verification" sin acotar criterios.
- AC8. Los atascos de tests y de verificación son hallazgos incidentales con umbrales observables, separados entre infraestructura del proyecto (tracker del proyecto) y costo del flujo de Hive (se reporta para la sesión de Hive). *Falso en la base cuando* `rg -n "40P01|deadlock|lock timeout" content/` no encuentra nada.
- AC9. Cada hallazgo incidental admite tres respuestas: abrir ticket, declinar (se guarda con conteo y se repropone si reaparece) o no volver a mencionar (se guarda como silenciado en la memoria persistente, por proyecto y tema, y se consulta antes de proponer). *Falso en la base cuando* `global.md` solo ofrece abrir o declinar.

Restricciones:

- `go test ./...` y `go vet ./...` pasan.
- `go run ./tooling/cli update --dry-run --source .` valida el catálogo sin errores de enlaces.
- `content/guidance/global.md` crece como mucho 300 bytes respecto a la base (42 980 bytes en `deb2958`, tope 43 280).
- La autorización básica (qué se puede borrar al cierre, incluidas las ramas mergeadas, y cuándo se mueve un ticket) sigue en `global.md` (AGENTS.md: no depender de que se cargue una skill).

## Entrega

- **Modo (D4-A):** automático. Rama `feat/close-hygiene-and-validation-gates` desde `rebuild/harness-engineering`, commits por tarea, push, PR a `rebuild/harness-engineering` y merge cuando pasen las comprobaciones de T5. El repo no tiene CI (no existe `.github/workflows`), así que T5 es el gate y no hay CI posterior al merge que esperar. Sin parada antes del push; el PR queda para lectura.
- **Revisión (D5-C):** sin revisión dedicada. Cubren `hive-verify-task` por tarea, la inspección del diff y T5.
- **Registro del cambio (D1-A, D6-A):** esta carpeta no se commitea en la rama. Vía de cierre elegida: después del merge, un commit con la carpeta ya archivada va por push directo a `rebuild/harness-engineering`, que no tiene protección de rama y ya recibe commits directos (historial de primer padre del 2026-09-29).
- **Después:** `hive update` desde este checkout (excepción temporal de AGENTS.md); solo cambia `content/` y un test, así que no hace falta recompilar el binario.
- **Punto de cierre:** merge en `rebuild/harness-engineering`, registro archivado y despliegue local refrescado. Fuera: el merge del rebuild a `master`.

## Decisiones del usuario (2026-09-29)

- D1-A: el registro del cambio se versiona una sola vez, al cierre.
- D2-A: la higiene vive en una skill nueva, `flow-close`.
- D3-A: "no volver a mencionar" aplica por proyecto y tema.
- D4-A: entrega automática.
- D5-C: sin revisión dedicada ("es solo texto").
- D6-A: con specs dentro del repo de código, el registro archivado se versiona después del merge.
- Ruta: `flow-plan`.
