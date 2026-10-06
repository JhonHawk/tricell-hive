# Tareas, verificación y revisión

Estado: implementación local completa en modo `hold`; pendiente el recorrido humano y la autorización de entrega. Sin commit.

## T1 — Contratos de proveedores e inventario

- [x] (Reducido por D1-B: catálogo manual con fuente oficial y motivo por capacidad.) Resolver matriz verificable para Engram, Context7 y pi-subagents por host/plataforma: versiones de herramientas consultadas, fuente oficial, alcance, efectos, detección y recuperación. Usar documentación y pruebas locales posteriores; no inferir soporte por un nombre de CLI.
- [x] Entregar inventario de instalación/operación/desarrollo en `_support/docs/architecture/tool-dependencies.md`, con fallback y consecuencia de ausencia; catálogo pequeño tipado y fixtures.

Dependencia: ninguna. Ejecución: hijo de integración, propietario de nuevo catálogo/recetas y documento; el principal aprueba la interfaz antes de T4. Fuentes: diseño, #29, skills y scripts distribuidos. Verificación: `go test ./tooling/...` con detección simulada presente/ausente/incompatible; combinaciones no verificadas no ofrecen ejecución. Una receta que no pueda preservar/reconciliar su alcance es manual y se reporta como límite, no se ejecuta a ciegas.

## T2 — Versión de producto y compatibilidad

- [x] Añadir `VERSION`, metadatos de compilación, `--version`, etiqueta en paquetes y estado, sin modificar hashes históricos.
- [x] Añadir índice transaccional de versiones y compatibilidad de esquemas según diseño; migrar legacy solo al aplicar, mantener snapshots/journals antiguos y mostrar estado mixto con honestidad.
- [x] Registrar recibo por consumidor y validar conformidad por bytes, no por procedencia del recurso. Probar etiquetas distintas con payload idéntico, actualización solo de manager y recursos compartidos intactos.
- [x] Congelar fuentes de release y rechazar versión duplicada con contenido distinto dentro del índice; separar builds de desarrollo y releases.

Dependencia: ninguna. Ejecución: principal, por acoplamiento entre `tooling/package`, `distribution`, `management`, `cli` y recuperación. Verificación: `go test ./tooling/distribution ./tooling/package ./tooling/management ./tooling/cli`; fixtures legacy, misma versión/distinto digest, igualdad de hashes anteriores, downgrade incompatible, estado mixto, fallo de escritura y recuperación del índice. Paquetes generados con `go run ./tooling/package --out <scratch>/packages`; sin publicar ni copiar working tree ajeno.

## T3 — Selección y consentimiento

- [x] Selección de hosts, consumidores vinculados, capacidades, requisitos y resumen; preservar `--hosts`, `--dry-run` y `setup` solo lectura.
- [x] Distinguir ausente, presente, registrado, configuración y verificación; cambios de selección nunca equivalen a remove.
- [x] Un lector de terminal persistente evita perder respuestas por buffers; EOF/cancelación/sin TTY no aplican.

Dependencias: T1 (interfaz), T2. Ejecución: principal, propietario de `tooling/cli/install.go`, archivos nuevos del wizard y pruebas. Verificación: `go test ./tooling/cli ./tooling/management` y PTY local con paquete construido y HOME sintético; elegir/deselegir, selección vacía, volver, consumidores compartidos, todas las respuestas en una lectura, falta de CLI, dry-run y estado cambiado antes de confirmar. No cambiar hosts no seleccionados salvo recursos compartidos expresamente consentidos; no afirmar ocultamiento en hosts no registrados.

## T4 — Pasos externos y recuperación

- [x] (Reemplazada por D1-B: sin recetas ejecutables; las capacidades quedan como pasos `manual` en el journal.) Integrar las tres recetas con intención durable, consentimiento, ejecución acotada, verificación y recibos; proceso interrumpido no se reintenta sin reconciliar.
- [x] Preservar instalaciones y datos anteriores, autenticar mediante proveedor y nunca loguear secretos. Estados parciales/códigos de salida y reanudación definidos en diseño.
- [x] Implementar journal padre y fases de núcleo/proveedores; probar caída entre commit e intención externa, núcleo intacto con capacidad nueva, resultado desconocido y recuperación sin revertir núcleo committed.
- [ ] (Abierta hasta la puerta nativa de Pi; D2-B retiró `Revert` y el código de receta quedó archivado.) Pi cumple aceptación de #29: presencia previa, ausencia, conflicto y recuperación únicamente de la entrada añadida.

Dependencias: T1–T3. Ejecución: hijo implementa recetas en rutas nuevas acordadas; principal integra journal/plan/apply/recover en `tooling/management`; sin escritores simultáneos en esos archivos. Verificación: `go test ./tooling/...` con ejecutables falsos y fixtures de configuración; interrupción antes/después de cada efecto, actualización ajena posterior, auth cancelada, red fallida, exit no cero, doble ejecución y resultado desconocido. Fixtures no prueban proveedores reales: la verificación de integración nativa en entorno desechable es gate separado abajo.

## T5 — Bootstrap online y paquete offline

- [x] Crear `bootstrap.sh`, distribución de binario crudo y generación de índice de releases, con resolución única de versión, download/checksum/extracción segura y delegación al instalador empaquetado.
- [x] Retener manager/paquete verificados después del consentimiento y antes de mutar; probar recuperación desde otra terminal sin red tras salir el bootstrap.
- [x] Mantener `install.sh` offline, manejo de TTY separado, plataformas actuales y fallos sin descargas implícitas de requisitos.

Dependencias: T2 y T3; interfaz estable antes de delegar. Ejecución: hijo de empaquetado, propietario de `bootstrap.sh` y pruebas dedicadas; principal integra cambios del builder para evitar solapamiento con T2. Verificación: `go test ./tooling/package ./tooling/distribution ./tooling/cli`, servidor local de fixtures y PTY; versión inexistente, plataforma errónea, checksum incorrecto, archivo truncado, rutas/enlaces inseguros, sin verificador, sin TTY, EOF, download interrumpido y reejecución. Comprobar que archivos adversos no escriben fuera del scratch ni lanzan procesos, y límites de bytes/entradas, duplicados y enlaces. Medir solicitudes/escrituras antes de confirmar, al cancelar y en dry-run/sin TTY. Probar flujo canalizado con una shell real y stdin ocupado por el script. Offline no realiza solicitudes de red.

## T6 — Verificación integrada y documentación

- [x] (En `tooling/`: `go test -race` y `go vet` pasan. Los fallos de `tests/pilot` son del cambio ajeno `report-readability`.) Ejecutar `go test ./...`, `go test -race ./...`, `go vet ./...`; registrar fallos preexistentes sin atribuirlos al cambio ni arreglarlos fuera de alcance.
- [x] Construir candidato aislado y validar instalación, actualización y recuperación mediante paquete real en HOME sintético. Comparar archivos ajenos antes/después.
- [x] Auditar inventario final de distribución, exclusiones y avisos legales; conservar como gate de publicación cualquier licencia no resuelta.
- [x] Actualizar `README.md`, contrato del instalador/manager/agentes y el inventario; incorporar requisitos de esta capacidad en deltas OpenSpec antes de cerrar e integrar. Revisar referencias si cambia contenido distribuido, siguiendo `instruction-resources.md`.
- [x] (Revisión con `review-code`, `review-refuter`, `review-ux` y dos rondas de `sdd-verify`; recorrido nativo en Linux x86_64; entrega por E1-A en f1272fc.) Revisar implementación con mecanismo acordado y presentar recorrido humano. Solo después, entregar por el modo Git elegido; publicación y HOME real siguen fuera.

Dependencias: T1–T5. Ejecución: principal integra; verificación y revisión UX de terminal por hijos independientes con candidato fijo y fixtures diferentes. No navegador, sitio web, screenshots ni framework UI: la superficie es una terminal. Pruebas de modelo excluidas.

## Gates

| Gate | Evidencia y momento | Autorización |
| --- | --- | --- |
| Revisión de este plan | Integración/datos y distribución/seguridad/terminal, sobre los tres archivos guardados | Completada en dos dominios; ver detalle abajo |
| Tests ordinarios y PTY | Comandos T1–T6, macOS local, HOME sintético, antes de entregar | Dentro de futura implementación autorizada |
| Revisión UX e in-vivo de terminal | `review-ux` y `sdd-verify` sobre paquete local: selección, cancelación, compartidos, fallos y resumen; sin instalaciones reales de terceros | Dentro de futura implementación local |
| Proveedores reales | Entorno desechable por plataforma, pin y host; no HOME cotidiano, no tokens/logs, red e instaladores oficiales; no modelos | Ejecución pendiente de acordar alcance/entorno; necesaria para declarar receta verificada |
| Linux ARM64/AMD64 | Paquete ejecutado nativamente en entorno disponible y autorizado; cross-build solo no cuenta | Entorno y ejecución remota pendientes; no distribuir plataforma sin gate |
| Revisión de implementación | Diff exacto respecto a base aislada, después de tests y antes de entrega | D5-A: `review-code` + `review-refuter`, cuando se autorice implementación |
| Publicación / smoke público | Origen público definitivo, índice y paquetes revisados, versión inmutable | Fuera del cambio local; requiere autorización específica |
| Licencias del paquete | El paquete no incluye `LICENSE` (MIT) ni el aviso de la licencia de Go (BSD-3 exige reproducirlo al distribuir binarios) | Pendiente antes de publicar; no bloquea el cambio local |

## Recorrido humano propuesto

Sobre el candidato construido y HOME sintético, arrancar el asistente, seleccionar dos hosts y luego desmarcar uno; comprobar resumen. Repetir con recursos compartidos que obligan a incorporar consumidores, y verificar que no se amplía alcance sin consentimiento. Seleccionar una capacidad sin requisito, omitirla y comprobar estado final. Cancelar antes de aplicar y contrastar ausencia de cambios. Repetir instalación y comprobar idempotencia y versiones. Los comandos exactos y paths del candidato se entregan al terminar implementación.

## Revisión y progreso

- 2026-09-25: fuentes inspeccionadas; decisiones D1-A/D2-A/D3-A recibidas. No se ejecutaron tests ni instaladores.
- Exploración delegada: `sdd-explore` nativo para ownership compartido; perfil anunciado por herramienta `gpt-5.6-terra/high`, controles efectivos no verificados. Resultado confirmado contra ownership y shared_test: no versiones independientes por host; D1-A preserva el contrato.
- Primera revisión: dos dominios detectaron seis huecos. Se incorporaron recibos por consumidor, journal padre, protocolo previo a Go, consentimiento de descarga, retención para recuperación y allowlist/licencias. La segunda y última revisión confirmó las correcciones sin nuevos hallazgos bloqueantes en ambos dominios.
- D4-A: `hold`; D5-A: `review-code`/`review-refuter`. No implementación autorizada.
- Escrituras de esta etapa limitadas a esta carpeta. Sin scratch ni procesos nuevos de instalación que limpiar.

### Evidencia de revisión final

Dos hijos `review-plan` seleccionados por ID nativo: integración/datos y distribución/seguridad/terminal. La herramienta anuncia `gpt-6-astra/medium`; no se verificó de forma independiente el modelo efectivo ni aislamiento adicional. Ambos trabajaron bajo contrato de solo lectura, sin pruebas ni escrituras. Exploración y revisión del principal contrastaron los hallazgos con `plan.go`, `apply.go`, `shared_test.go`, el builder y el instalador.

Candidato de segunda revisión (SHA-256):
- `design.md`: `c2f9ea51fdd98f0fb9b36802ecae16e057a6704e0e2f33a78d3b9670d30e6d0d`.
- `proposal.md`: `acb62ab95061e712170f9bff200c26e0a099c76d00e9f269f0f55e94c49c5b4b`.
- `tasks.md`: `256723e1556315e2c870de18bde4556e3eae0248b4d222c1162770d557d4369c`.

Resultado: los seis hallazgos de la primera ronda quedaron resueltos en el diseño; ninguna nueva decisión bloqueante dentro de las coberturas. Tras esa revisión solo se actualizaron estado, siguiente paso y esta evidencia, sin cambios de contratos/tareas. Enlaces locales y espacios finales comprobados. No se ejecutaron tests de implementación ni instaladores; los gates de proveedores reales y plataformas permanecen pendientes para su etapa, con autorización específica.

Continuación propuesta: `flow-build` empieza por T1/T2 después de autorización, en checkout aislado para preservar trabajo ajeno. El resultado local se detiene para recorrido humano y revisión del diff; no commit/push/PR/merge/publicación/despliegue. No se creó scratch que limpiar; se retienen únicamente los tres archivos del plan.

### Ejecución en curso

Pruebas base `go test ./tooling/...` pasaron. RED observado para recibos de versión, colisión de versión y recuperación de metadatos; GREEN en pruebas enfocadas y suite management. RED/GREEN adicional para onboarding: fallo opcional, núcleo intacto y caída tras efectos. Evidencia en `_support/evidence/2026-09-25-versioned-installer-onboarding/`. Tres hijos `backend-developer` nativos (perfil anunciado gpt-5.6-terra/high, modelo efectivo no comprobado) escriben rutas disjuntas: providers, distribution/package/version/bootstrap y cli. Principal es propietario de management, docs y este registro. No se ejecutaron proveedores reales ni modelos.

- 2026-09-26: la sesión de Codex se cortó por límite de uso con dos hijos de implementación sin reporte final; `go test ./tooling/providers` quedó con tres fallos. Corregidos en Claude Code:
  - Las recetas de Context7 y Pi exigen que el ejecutable sea `npm` o `pi` (ruta absoluta y limpia); antes aceptaban `/bin/sh` con los argumentos de la receta. RED/GREEN observado, con prueba nueva para Pi.
  - El helper de archivos tar de las pruebas escribía cuerpo en un symlink; ahora la prueba del symlink ejerce el rechazo real del instalador.
  - `InspectEngram` comparaba el SHA-256 del binario contra el del `.tar.gz`, así que nunca verificaba. Ahora el readback exige archivo regular ejecutable en el destino con la versión exacta; el checksum del archivo sigue verificándose en `RunEngram` antes de instalar.
  
  `go vet ./tooling/...` y `go test ./...` pasan. T4 sigue abierta: los dos hijos cortados no reportaron su estado, así que las tareas no se marcan completas.
- 2026-09-26: conclusión del trabajo cortado, en Claude Code. Reconstrucción del transcript de Codex y auditoría de T1–T5 por dos `sdd-explore` de solo lectura. Hallazgo principal: la CLI usaba `coreOnlyAdapterFactory` y el adaptador nativo no tenía llamador.
  - Huecos del plan cerrados con RED/GREEN: adaptador nativo cableado en `install` y `recover`, detalle por componente del resultado parcial, cambio de estado antes de confirmar, failpoint `core_applied` con su prueba, Node < 18 como paso manual, pruebas de `bootstrap.sh` con shell real (sh, bash, dash) que destaparon un defecto de TTY.
  - Revisión D5-A (`review-code`) sobre el diff: H1 faltaba `hive bootstrap` (implementado); H2 `recover` fallaba con onboarding pendiente; H3 journal padre huérfano; H4 el builder sobrescribía una versión; H5 retención no tolerante a caídas; H6 recuperación de Engram atascada; H7 validación del instalador por archivo y bloqueando `Recover`; H8 salida silenciosa en dash; H9 paso manual reportado como `auth_pending`; H10 documentación contradictoria; H11 código muerto. Todos corregidos con prueba; H4 con RED de compilación.
  - Verificación in-vivo (`sdd-verify`) en HOME sintético con PTY real: 8 escenarios aprobados; halló un panic con `--source` sin versión (corregido). Revisión UX (`review-ux`): reintento ante entrada inválida, resumen contradictorio, motivo en el resultado final, líneas de más de 80 columnas y recurso compartido sin nombre (corregidos). Por decisión del usuario, los mensajes que ya estaban en inglés se quedan en inglés.
  - Seguimiento: [#37](https://github.com/JhonHawk/tricell-hive-private/issues/37) para ver el estado de cada capacidad después de instalar.
  - `go vet ./tooling/...` y `go test -race ./tooling/...` pasan. Los fallos de `tests/pilot` pertenecen al cambio `report-readability`, ajeno a este.
- 2026-09-26: segunda ronda de revisión y simplificación, en Claude Code.
  - `review-refuter` no halló defectos de seguridad ni de datos; sí pruebas que no podían fallar (digest del paquete, vínculo del manager, versión del índice y `BindInstaller` en `hive bootstrap`; ramas que conservan el journal padre y rechazan colisiones de retención). Se añadieron pruebas verificadas por mutación y se corrigió que la retención borrara un instalador registrado ante un fallo transitorio.
  - Idioma (D1-A): todo texto de usuario del instalador en inglés, como en `HEAD`; `design.md` corregido.
  - Auditoría de mantenibilidad y decisiones D1-B, D2-B, D3-B, D4-B: se retiraron las recetas de Engram, Context7 y Pi (archivadas fuera de Git en `_support/workspace/2026-09-26-versioned-installer-onboarding/`), `Revert` y la API sin uso; `Apply` y `Onboard` comparten un único `preflight` con un solo indicador `nested`; `management.Pending` es el punto único de pendientes; constantes de estado únicas; el seam de prueba del bootstrap salió de los artefactos (flag del linker y copia sustituida del script); funciones largas divididas. Producción añadida sobre `HEAD`: de unas 3.900 a unas 1.900 líneas; `tooling/providers` pasó de 1.342 a 107.
  - Límites conocidos, no defectos: `recover` con `--home` necesita `--state-dir`; directorios `.staging-*` de una caída no se limpian; una versión queda reservada si falla la compilación de una plataforma (usar otro `--out`); con varios instaladores retenidos, el mensaje de recuperación online elige el de identificador mayor.
- 2026-09-26: verificación in-vivo final (`sdd-verify`) sobre el candidato simplificado: 8 de 9 escenarios aprobados (instalar, capacidad manual con salida no cero, reintento ante entrada inválida, cancelación sin escrituras, idempotencia, dry-run y sin TTY, recuperación tras `SIGKILL`, archivos ajenos intactos). Corregido después, con pruebas: errores en español preexistentes en `install.sh` y `manifest.go` pasados a inglés (D1-A), líneas del resultado parcial ajustadas a 80 columnas, recuento engañoso en una reinstalación sin cambios, y temporales `.hive-write-*` que dejaba un proceso matado, ahora eliminados por `Recover`. `go test -race ./tooling/...` y `go vet ./tooling/...` pasan.
- 2026-09-26: recorrido del candidato en Linux nativo (bb1, Linux x86_64, paquete `linux/amd64`, HOME sintético bajo `/tmp`, entorno limpio con `env -i`): instalación nueva, capacidad manual (salida 1, motivo y siguiente acción, sin procesos de proveedor), reinstalación sin cambios, entrada inválida con reintento y cancelación sin escrituras, `--version`, dry-run, rechazo sin TTY, `recover` sin pendientes y `status` (80 registros `installed`/`verified`). Tres mensajes de texto fijo superaban 80 columnas; se corrigieron con prueba y se repitieron en bb1. No se ejercitó en Linux una interrupción real con `recover`. El directorio temporal de bb1 se eliminó.
