# release-self-update — Actualización completa desde releases

| Campo | Valor |
| --- | --- |
| Estado | Completado · Listo para validación humana antes del push |
| Tracker | Sin ticket; no se ha solicitado crearlo |
| Git | tricell-hive · development · e2a633e49f9ae59654bd272c574b460696ad3dae · feat/release-self-update · worktree .claude/worktrees/release-self-update |
| Visor | work key `release-self-update` · work ID `059ebc7c344c592434939d3a5e3c91bacb647beea3cb541a67006c554319365a` |

## Decisiones

2026-11-05

- Propuesta aceptada mediante «A»: ampliar `hive update` para actualizar binario e integraciones desde releases, sin Git ni Go; paquetes por versión, comando estable, versiones anteriores retenidas y origen Git explícito para desarrollo.
- Plan solicitado mediante «A»: documentar la implementación; no implementar todavía.
- **D1 A** — ¿Cómo recuperar una interrupción? Antes del commit del motor se recupera el estado previo; después se conservan las integraciones confirmadas y se completa la activación con el binario nuevo retenido.
- **D2 A** — ¿Qué modo de entrega? Interactiva: detenerse antes del push para validar CLI/TUI; después de la validación, push, PR y merge a development.
- **D3 A** — ¿Qué revisión de código? Nativa: Codex `/review`, sobre el candidato antes de la validación humana.
- **D4 A** — ¿Qué hacer si Pi registrado carece de la declaración requerida? Bloquear antes de aplicar, preservar settings y explicar la reparación separada mediante `hive install`; no ejecutar `pi install` durante release update.

## Solicitud

> «Pregunta, si alguien instala el binario una vez, basta con que haga hive update para que se actualicen sus skills, hooks y el mismo binario de go?»
>
> «podemos implementar la función verdad?»

El usuario aceptó primero preparar una propuesta y después preparar este plan. No hay criterios de un ticket. La implementación quedó autorizada al invocar flow-build; release pública y promoción siguen fuera del alcance.

## Objetivo

Hoy `hive update` instala contenido de un commit local y no cambia el ejecutable. El usuario de paquetes debe repetir el instalador. Después de una migración inicial, `hive update` podrá actualizar el binario y las integraciones registradas como una operación recuperable desde una release estable, independientemente del directorio actual.

## Alcance y aceptación

Estimación: 25–35 archivos de distribución, CLI/TUI, gestión de estado, instalador, pruebas y documentación; el cambio de estimación respecto de 20–30 responde al protocolo entre versiones, los planes guardados y el bloqueo de mutaciones durante recuperación. Plataformas: darwin/arm64, linux/arm64, linux/amd64. Sin dependencias nuevas, instalación de runtimes, actualizaciones en segundo plano, firma de artefactos, limpieza de paquetes anteriores ni reinicio automático del visor. Los instaladores offline siguen funcionando sin red. Los proveedores opcionales no se ejecutan implícitamente durante una actualización de release.

- AC1 — Desde una instalación migrada, `hive update` sin Git, Go ni checkout instala una release estable compatible y deja el comando estable y todos los hosts registrados en esa versión. Falso hoy: `planUpdate` requiere Git y no activa ejecutables.
- AC2 — `--version X.Y.Z` selecciona exactamente esa release compatible; la selección predeterminada no instala prereleases ni baja de versión. Una versión sin protocolo compatible se rechaza antes de aplicar y se explica la alternativa manual. Falso hoy: no existe selección de release en update.
- AC3 — Un paquete remoto corrupto, incompatible, con redirección no permitida, rutas peligrosas o límites excedidos se rechaza antes de ejecutar su binario o cambiar la instalación activa. Falso hoy: update no procesa paquetes remotos.
- AC4 — `--dry-run`, cancelación y ejecución sin terminal sin `--out` no alteran la instalación activa, el enlace, el estado ni el almacén persistente de paquetes; los temporales de preview se retiran. Falso hoy en el camino nuevo: no existe preview de una actualización completa.
- AC5 — Un plan de actualización completa guardado con `--out` fija artefacto, versión, plataforma, ejecutor y fingerprints; `hive apply --plan` lo aplica sin red o rechaza bytes, destinos o estado cambiados, sin volver a resolver latest. Falso hoy: los planes no incluyen activación del comando.
- AC6 — El instalador actualizado migra una instalación de paquetes y registra `~/.local/bin/hive` como comando estable sin editar PATH ni sobrescribir un destino no administrado; repetir la migración es idempotente. Falso hoy: get-hive.sh sólo imprime una ruta con versión.
- AC7 — La actualización completa reutiliza las protecciones del motor para preservar ediciones ajenas, configuración y recursos compartidos, y no ejecuta proveedores opcionales implícitamente. Falso hoy en el camino nuevo: no existe aplicación de release y comando en una misma operación coordinada.
- AC8 — Una interrupción anterior al commit del motor permite recuperar sus cambios incompletos y conserva el comando previo. Falso hoy: no hay recuperación que coordine paquete, motor y comando.
- AC9 — Una interrupción posterior al commit del motor conserva ese commit y permite completar, sin red, la activación usando el binario retenido, incluso si el lanzador anterior no entiende el nuevo estado. Falso hoy: no existe el registro padre ni esa delegación de recuperación.
- AC10 — Dos actualizaciones completas, o una mutación del motor que compite con una actualización pendiente, no pueden aplicar operaciones incompatibles; las lecturas explican la recuperación pendiente. Falso hoy: no hay exclusión ni detección del registro padre de update.
- AC11 — `hive update --source DIR [--rev REV]` conserva la actualización desde un commit sin red ni activación de binarios; un build `dev` invocado sin origen explícito no se sustituye por una release. Falso hoy: el origen Git es implícito y no existe el nuevo despacho entre modos.
- AC12 — La TUI permite elegir Release o Source, muestra las versiones anterior y destino, usa la misma preparación/aplicación/recuperación que el comando y comunica la versión del visor aún en ejecución cuando difiere. Falso hoy: Update sólo ofrece Source y Revision y aplica contenido.
- AC13 — Un preview con paquete destino todavía ausente genera los hooks con la ruta final y los bytes esperados, aunque el ejecutor de preparación sea temporal; apply sólo admite ese destino tras verificar el paquete definitivo. Falso hoy: BuildPlan usa el ejecutable del proceso y rechaza rutas temporales.
- AC14 — Si Pi registrado requiere instalar o reparar su declaración de pi-subagents, release update falla antes del commit y de la activación, preserva settings y explica la reparación separada sin ejecutar `pi`. Falso hoy: BuildPlan/Apply pueden programar y ejecutar `pi install` cuando falta la declaración.

## Prerrequisitos

Ninguno externo para implementar y verificar: paquetes, red y homes de pruebas se simulan localmente; se compilan dos candidatos de prueba para ejercer el protocolo entre procesos. No hacen falta cuentas ni registros reales.

La disponibilidad pública exige posteriormente una release con este protocolo y la publicación del instalador actualizado. Eso no es un prerrequisito del build local ni está autorizado por este plan.

## Contexto verificado

- Base local y `development` remoto coincidentes en e2a633e4 al iniciar la propuesta; revalidar antes del build.
- Release pública observada: v0.8.0, 2026-11-03, paquetes y checksums para las tres plataformas del alcance.
- `get-hive.sh` resuelve releases y verifica descarga; `tooling/distribution` ya contiene extracción acotada y verificación del paquete.
- `tooling/management/plan.go` resuelve el ejecutable real para los hooks: mantener hooks hacia el binario versionado real y usar el enlace estable sólo para el comando del usuario.
- `tooling/management/apply.go` conserva transacciones committed; el estado actual usa esquema 7. La arquitectura contiene referencias históricas a v6 que deben actualizarse al tocar ese contrato.
- La spec `versioned-installation` describe paquetes, actualización Git, TUI y recuperación; el delta de este cambio se aplica a esa capacidad.
- No hay suite general en CI según CONTRIBUTING.md; los checks locales son la evidencia funcional.
- Durante la revisión HEAD avanzó a 2f39491b14813e19ba11b71e6f1960be47a687a1 con dos commits de flow-close (cb9b3ffb y 2f39491b). Se comprobó el diff: no toca el updater inspeccionado. Este plan mantiene e2a633e4 como snapshot de investigación; antes del build se reconciliará con el HEAD entonces vigente.

## Entrega

**D2 A** fija entrega interactiva y **D3 A** revisión nativa Codex `/review`. Antes del push se presenta el candidato para validación humana de CLI/TUI; la validación aceptada habilita push, PR y merge a development según ese modo. La revisión de plan es independiente y se realiza antes del handoff. No se autorizan implementación, release pública, promoción a master ni modificaciones de la instalación real con la mera aprobación del plan. La regla del marcador de mantenedor se evalúa sólo si posteriormente se integra código y obliga entonces a inspeccionar los commits y efectos que abarca.

## Revisión y progreso

2026-11-05 — Primera redacción, D1 A incorporada. Próximo paso: comprobar coherencia, revisión independiente y decisiones de entrega.
2026-11-05 — Revisión 2 por H1/H2: añadido binding entre ejecutor temporal y destino final (AC13), y preflight de Pi sin efectos externos, elegido en D4 A (AC14). Se revisan T1, T2, T3 y T6 y el delta; ninguna tarea implementada queda invalidada porque todas siguen sin comenzar. D2 A y D3 A registradas.
2026-11-05 — La revisión independiente de la revisión 2 no encontró gaps materiales. El checker de coherencia pasa sin salida. Sólo se han escrito los cuatro documentos de este cambio; no hay implementación ni checks de producto. Próximo paso: aprobación de AC1–AC14 y entrada explícita a flow-build.

2026-11-05 — Usuario inició flow-build y aprobó AC1–AC14. Se trasladó la carpeta al worktree aislado feat/release-self-update desde origin/development e2a633e4, excluyendo commits ajenos del checkout principal.

2026-11-05 — T1–T6 implementados y verificados. Candidato de código f6d8cf91; base integrada a67feac4. Pendiente D5: aceptación humana del flujo antes del push, conforme D2 A. Revisión nativa ejecutada con un follow-up; H12 local posterior lleva Unreviewed y regresión verde.
