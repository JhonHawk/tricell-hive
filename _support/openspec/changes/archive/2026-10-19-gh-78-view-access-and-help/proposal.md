# Acceso de un uso y ayuda del visor

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en development |
| Tracker · GitHub Issues | • [#78 — token en el abridor](https://github.com/JhonHawk/tricell-hive/issues/78)<br>• [#82 — ayuda del visor](https://github.com/JhonHawk/tricell-hive/issues/82) |
| Git | Automática a `development` al autorizar implementar |
| Verificación | Pruebas · carreras · revisión de código y seguridad · recorrido funcional aislado |
| Siguiente paso | Ninguno · registro archivado |

## Solicitud

> ataquemos los dos en una sola sesión, haz research al respecto

El usuario eligió después `D1-A Plan conjunto (Recommended)` y, el 2026-10-19, entrega automática, revisión mediante roles instalados y verificación funcional independiente. La elección del plan no autoriza implementar.

Issues leídos el 2026-10-19; no contienen una sección formal de criterios. Sus arreglos sugeridos dicen:

- #78 (token en el abridor): «Un código de un solo uso en el fragmento que la página canjea por el token en una petición al servidor.»
- #82 (ayuda del visor): «Reconocer `--help`/`-h`/`help` en el despachador de `view` e imprimir sus seis subcomandos.»

## Objetivo y alcance

Retirar el token de vida del servidor de la URL entregada al abridor del sistema, sustituyéndolo por un código breve y consumible una sola vez. Hacer que la ayuda del grupo `view` termine satisfactoriamente sin ejecutar operaciones.

Incluye CLI, servidor y módulo propio del visor, pruebas, documentación y deltas. No modifica el renderizador de recorridos ni las reglas distribuidas. No cambia almacenamiento persistente, capacidades de archivos, diseño visual ni autenticación a cookies. No incluye otros issues del visor ni pilotos de modelos, que siguen pausados.

## Criterios de aceptación

- AC1. Cada apertura con servidor activo entrega al abridor `#code=<code>` y nunca el token del servidor; stdout y errores tampoco revelan token o código. Falso en la base: la prueba de apertura exige `#token`.
- AC2. Un código emitido por un cliente autenticado permite un solo canje antes de 60 segundos; es inválido al cumplir 60 segundos y entre canjes concurrentes solo uno tiene éxito. Falso en la base: no existe emisión/canje de códigos.
- AC3. Dentro de un protocolo cuya emisión y canje satisfactorios se verifican, la emisión sin token y el canje sin código, con código inválido o con `Host`/`Origin` ajenos, ausentes o `null` según el contrato responden 403 sin conceder acceso. Falso en la base: las rutas inexistentes responden 404, no los estados del contrato.
- AC4. La página completa el acceso con `#code`, limpia el fragmento antes de esperar la red, conserva acceso al recargar y puede recibir otra apertura sin que un canje anterior sobrescriba el nuevo. Un canje fallido muestra cómo reabrir. Falso en la base: solo procesa `#token`.
- AC5. `hive view --help`, `-h` y `help` enumeran los seis subcomandos, terminan con éxito y no normalizan estado ni lanzan procesos. Falso en la base: se rechazan como subcomandos desconocidos.

Restricciones de regresión: índice, salud, sesiones y escrituras normales conservan autenticación; capacidades de archivos no cambian; servidor permanece en loopback; `open` no inicia servidor; ayudas de subcomandos siguen funcionando y sin argumentos/subcomando desconocido siguen fallando. Los códigos no persisten al reiniciar. No se aceptan URLs nuevas con `#token` como ruta alternativa.

## Entrega acordada

- D2-A (entrega automática): cuando se autorice implementar, rama desde `development`, commits verificados, push, PR y merge por el agente tras gates. No hay pausa de validación humana. Registro se publica una sola vez al cierre mediante pequeño PR sin revisión dedicada que el agente integra.
- D3-A (revisión): `hive-review-code` y `hive-review-security` sobre el candidato local; hallazgos bloqueantes pasan por refutación según el flujo. Consentimiento de ejecución de esos mecanismos concedido para la futura implementación.
- D4-A (interfaz): `functional`, con `hive-verify-change`; sin revisión estética independiente. Rige el resto de la sesión en este repositorio hasta que el usuario lo cambie.
- El repositorio no tiene CI según `CONTRIBUTING.md` y no existe `.github/workflows`. Comprobarlo de nuevo antes de publicar. El marcador `_support/workspace/maintainer.local` existe: integrar cambios de manager exige reconstruir el binario mediante el procedimiento del repositorio. No hay contenido distribuido que desplegar por este cambio. No promover a `master` ni publicar release.
- D5-A autorizó `flow-build` el 2026-10-19. Rama `fix/gh-78-view-access-and-help` en worktree ignorado `_support/workspace/worktrees/gh-78-view-access-and-help`; issues cerrados tras integrar. GitHub Issues no ofrece estado intermedio configurado: sin projectItems ni etiqueta de progreso; no se inventa uno.

Detalle técnico: [diseño](design.md). Ejecución y revisión: [tareas](tasks.md).

## Resultado

Integrado mediante [PR #83 — acceso de un uso y ayuda](https://github.com/JhonHawk/tricell-hive/pull/83), commit `b2c0ecb2d22b804b00e65fa7285cec3114f78643`, el 2026-10-19. Ambos issues cerrados por GitHub. Sin CI configurado ni corridas disparadas; gates locales correctos. Binario local reconstruido desde revisión integrada, ayuda comprobada; `hive view status` informó `running: false`, sin reiniciar ni arrancar el visor cotidiano.
