# Reinstalar lo que Hive instaló y se borró a mano

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `development` el 2026-09-30 con el [PR #77](https://github.com/JhonHawk/tricell-hive-private/pull/77) (merge `f95ea11`); especificación `versioned-installation` actualizada |
| Tracker · GitHub Issues | Sin ticket: es el hallazgo H1 del cierre de #65 y #66, y el usuario pidió resolverlo directamente |
| Git | `automatic` · rama `feat/reinstall-missing-managed-files` desde `development`, PR a `development`, `/code-review` antes del push, merge · sin CI |
| Verificación | pruebas por paquete · suite local · `hive-verify-change` con el binario real en un `HOME` temporal |
| Siguiente paso | Ninguno en este cambio |

## Objetivo

Si falta algo que Hive gestiona y no hay copia de seguridad, hoy no hay salida. Instalar, actualizar y quitar se rechazan, y el único consejo es restaurar desde una copia. Pasa con tres cosas borradas a mano: una skill, un agente o un enlace; un bloque de Hive o de voz; o el archivo del usuario que contenía ese bloque.

Decisión del usuario del 2026-09-30: si se borró a mano, se reinstala. Instalar y actualizar reponen lo que falta por la misma ruta que una instalación nueva. Quitar termina sin error cuando lo que tenía que borrar ya no está. Lo editado a mano se sigue rechazando, porque ahí sí hay contenido del usuario que proteger.

## Alcance y aceptación

Incluido:

- **Gestor (`tooling/management`):** los cambios del plan llevan un campo `Gone`, fijado cuando `owned()` informa `ManagedFileMissing` o `ManagedBlockMissing`.
  - Instalar y actualizar reponen lo que falta y lo registran.
  - Quitar no escribe nada y retira el registro.
  - La voz sigue el mismo criterio: `voice set` y `voice off` omiten las rutas sin bloque de Hive en vez de fallar.
- **CLI (`tooling/cli`):** los resúmenes cuentan una reposición como cambio, así que ya no dicen «already up to date» cuando escriben. El texto de ayuda de doctor distingue un archivo editado de uno borrado.
- **Documentación:** `deployment-manager.md` y un escenario nuevo en `versioned-installation/spec.md`.

Excluido:

- Reparar contenido editado: se sigue rechazando.
- Un estado `missing` separado de `drift` (D2-A).
- Mecanismos para no instalar una skill concreta: no existen hoy, y borrar un archivo no lo es.

Restricciones:

- Lo editado a mano se sigue rechazando: skill o agente cambiado, permisos, enlace reapuntado, bloque editado o marcadores rotos.
- Interrumpir una reposición deja el sistema recuperable con `hive recover`.
- Un plan con `Gone` cambia su ID. `state.json` no cambia. Los planes guardados y el diario ganan un campo opcional (`gone`, con `omitempty`): los anteriores siguen siendo válidos, pero un binario viejo rechaza un plan o un diario pendiente que lo lleve.
- `go vet ./...` y `go test ./...` siguen en verde.

Criterios:

- AC1. Con una skill, un agente o un enlace gestionados borrados, `hive update` e `hive install` los vuelven a crear con los mismos bytes y permisos, el resumen cuenta el cambio, y después `hive status` no marca `drift`. *Falso en la base cuando* los dos se rechazan con «Hive installed <ruta> and it is no longer there».
- AC2. Con el bloque de Hive o de voz borrado de un archivo del usuario que sigue existiendo, `hive install` vuelve a añadir el bloque y conserva el texto del usuario. *Falso en la base cuando* se rechaza con «the Hive block in <ruta> is gone».
- AC3. Con el archivo del usuario que contenía el bloque borrado entero, `hive install` crea el archivo con el bloque. *Falso en la base cuando* se rechaza con «<ruta>, which held the Hive block, is no longer there».
- AC4. `hive plan remove` de un host cuyo archivo gestionado ya no está termina sin error y retira el registro. Si otro host sigue usando ese recurso compartido, el archivo sigue sin reponer (D1-A) hasta el próximo `install` o `update`, que lo repone. *Falso en la base cuando* quitar se rechaza.
- AC5. `hive voice set` y `hive voice off` con el bloque de voz borrado reponen o dan por quitado el bloque. Con un archivo sin bloque de Hive, omiten esa ruta, y solo fallan si no queda ninguna, con el mensaje «run hive install first». *Falso en la base cuando* `TestVoiceSetFailsAtPlanTimeWhenHiveBlockMissing` y `TestVoiceOffFailsAtPlanTimeWhenHiveBlockMissing` exigen el rechazo.
- AC6. El texto de ayuda de `drift` en doctor, en `deployment-manager.md` y en la especificación distingue: si se editó, deshacer el cambio; si se borró, correr `hive install` o `hive update`. *Falso en la base cuando* `driftRepairText` (`doctor.go:436`) dice que Hive no puede reparar el archivo y que se restaure desde una copia.

## Entrega

Siguen en vigor las decisiones de esta sesión: entrega automática, `/code-review` de Claude Code antes del push y sin CI. Decisiones del usuario del 2026-09-30:

- **Resolver ya:** «si se borró manual, se reinstala».
- **D1-A:** al quitar un host, una skill compartida que falta no se repone.
- **D2-A:** lo borrado sigue mostrándose como `drift`, con un texto de ayuda nuevo.

Secuencia:

1. Worktree con `feat/reinstall-missing-managed-files` desde `origin/development`.
2. T1 y T3 en paralelo con el contrato fijado; después T2.
3. T4, la verificación de punta a punta.
4. Suite local.
5. `/code-review`.
6. `gh pr create --base development`.
7. Merge.
8. Recompilar `hive`.

El registro se cierra con un PR pequeño hacia `development`, sin revisión dedicada, al que yo hago merge.
