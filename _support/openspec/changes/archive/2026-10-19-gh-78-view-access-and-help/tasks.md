# Tareas y verificación

Base de investigación: `0d2e51446476fc04b1664a6f7b4f507694bc7869`. Base efectiva de build verificada igual a la inspeccionada; rama fix/gh-78-view-access-and-help en worktree ignorado.

## T1 — Ayuda del grupo view

- [x] Proveer ayuda satisfactoria sin efectos.

**Closes:** AC5.

**Dependencias:** ninguna. **Ubicaciones:** `tooling/cli/main.go`, `view_command.go`, `view_command_test.go`, pruebas de ayuda general.

**Ejecución:** hilo principal; cambio menor y acoplado al archivo CLI que después integra T2 (canje de acceso).

**Enfoque:** `tdd`. Primero fallar aliases y comprobar ausencia de `launch/open` y creación de estado; después helper compartido y aliases. Mantener pruebas sin argumentos/desconocido y ayudas de subcomandos.

**Verificación:** `go test ./tooling/cli -run 'Test(View.*Help|ViewRequiresKnownSubcommand|Help)' -count=1`; confirmar número de casos ejecutados. Binario candidato: `view --help`, `view -h`, `view help` salida 0, seis comandos, sin directorio de estado creado.

## T2 — Acceso de un uso de extremo a extremo

- [x] Implementar C1 y C2 (emisión/canje y consumidor web), con pruebas y documentación del contrato.

**Closes:** AC1, AC2, AC3, AC4.

**Dependencias:** T1 (ayuda, para evitar escritores simultáneos en CLI).

**Ubicaciones:** `tooling/cli/view_command.go`, `view_command_test.go`; `tooling/viewer/server.go`, pruebas del servidor y `static/viewer.mjs`; README, referencia de arquitectura y delta de la capacidad. Solo el módulo propio del visor, no assets duplicados del renderizador.

**Ejecución:** delegado a un implementador backend con propiedad de esas rutas; unidad de extremo a extremo para no separar el contrato del cliente. Orquestador integra y edita exclusivamente este registro. Un verificador independiente comprueba tarea según flow-build.

**Enfoque:** `tdd`. Pruebas Go primero: argumento del abridor sin token y sin credenciales en salidas/errores; emisión autorizada y denegada; canje satisfactorio, vencimiento con reloj controlado a 60s, doble uso y carrera con un éxito; host/origin/método/tipo/cuerpo, JSON sobrante, límite de pendientes; reinicio elimina pendientes. Para AC3 (rechazos), comprobar los 403 del contrato junto a emisión/canje satisfactorios. Servidor anterior sintético responde 404 a emisión: error explica stop/start/open, sin abrir navegador ni enviar token en URL. Preservar regresiones de autorización actuales. Probar frontend por comportamiento en navegador real, no solo búsqueda de cadenas.

**Verificación:** `go test ./tooling/cli -run '^TestView' -count=1` y `go test ./tooling/viewer -count=1`, con nuevos casos del protocolo. Para AC4 (inicialización web), recorrido funcional en candidato compilado: retener respuesta del primer canje y comprobar, mientras está pendiente, fragmento vacío, exactamente una solicitud de canje y ninguna a `/api/index`; liberar respuesta y comprobar acceso y recarga. Reapertura con un segundo código funciona; rechazo de código vencido/reutilizado muestra recuperación; respuesta demorada de primer canje no reemplaza la segunda apertura. Entregar un 403 de una solicitud autenticada antigua después del nuevo canje satisfactorio y comprobar que conserva acceso nuevo; un índice antiguo tampoco reemplaza datos actuales. Para retener respuestas, usar un wrapper HTTP temporal sobre el `http.Handler` real del visor, en el mismo listener y origen, con compuertas de prueba controlables; `Server` ya implementa `ServeHTTP`. El hijo maneja esas compuertas y el navegador. El wrapper solo demora respuestas, sin reemplazar autenticación ni protocolo; no se añade un hook de producción. Así la observación no depende de interceptación de red del navegador ni de inspección de cadenas. Observar solo estado/longitudes y nombres de cabeceras, nunca valores sensibles. El veredicto de T2 permanece pendiente hasta cubrir el navegador.

## Gates compartidos

- Ejecutar en candidato final `go vet ./...`, `go test ./...`, `go test -race ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, según CONTRIBUTING. Suite completa una vez por candidato relevante; repetir solo por fallos o cambios que invaliden evidencia.
- Compilar `go build -o <task-temp>/hive ./tooling/cli`; estado sintético propio y fixtures de `tests/fixtures/walkthrough/`. No registrar repositorios cotidianos, no usar el token del servidor real, no abrirlo ni reiniciarlo. El test/binario puede arrancar y detener su propio servidor efímero aislado; no se modifica la instalación cotidiana hasta integrar según entrega.
- `hive-verify-change` sobre binario compilado: recorrido de AC4 (inicialización web) y prueba de apertura sin servidor; comparar con base inspeccionada. `agent-browser` instalado en `/opt/homebrew/bin/agent-browser`, sesión propia sin perfil, estados sintéticos independientes; viewport 1440×900 y 390×844, tema oscuro existente, idioma es/en en los avisos. Cargar skill `agent-browser` y referencias de browser automation/UI gate antes de usar navegador; comandos acotados y capturas sin credenciales. Interacciones funcionales y accesibilidad afectada; no evaluación estética ni nueva composición. Revisión funcional autorizada por D4-A.
- Revisión local independiente `hive-review-code` y `hive-review-security` del candidato final y diff frente a base. Seguridad cubre emisión/canje, concurrencia, procesos, redirecciones/errores, secretos y conservación de límites. Autorizadas por D3-A. Refutar hallazgos bloqueantes; aplicar el límite de correcciones/revisión del flujo.
- Bajo entrega automática no hay pausa humana ni walkthrough de validación obligatorio; publicar evidencia breve de comprobaciones, sin afirmar mejora estética. No hay pilotos ni servicios externos en estos gates.
- Evidencia conserva revisión/candidato, comandos, resultados y límites. Al finalizar se cierran únicamente browser y servidor de prueba propios y se eliminan sus temporales reproducibles.

## Revisión del plan

Revisión inicial completada sobre los cuatro documentos, antes de estas correcciones, por dos despachos nativos `hive-review-plan` sin historial heredado y con encargo de solo lectura: backend/CLI/autenticación (`/root/plan_backend`) y frontend/inicialización (`/root/plan_frontend`). Sin nueva infraestructura. Duraciones y tokens no disponibles en el host; contrato del prompt no demuestra aislamiento técnico.

Se aceptaron las propuestas de ambos revisores: AC3 exige 403 y caso satisfactorio de referencia, recuperación explícita ante servidor anterior activo, observación de fragmento antes de respuesta, y prueba de 403/índice antiguos tras reapertura. Cambios en proposal/design/tasks reconciliados por el orquestador aplicando esas propuestas tal como se formularon; no requieren otra ronda según plan-review. Delta permanece consistente: invalidez 403 y aperturas solapadas ya estaban especificadas. No quedan hallazgos materiales pendientes dentro de la cobertura; no equivale a código revisado o comportamiento ejecutado.

## Progreso

- 2026-10-19: investigación convergente; D1-A (plan conjunto), D2-A (entrega automática), D3-A (código y seguridad), D4-A (funcional) resueltas. Registro sin commit; implementación no autorizada.
- 2026-10-19: revisiones backend/frontend reconciliadas. Listo para implementar, pendiente elección de inicio. Ningún proceso de prueba propio activo ni artefactos temporales por limpiar; material previo preservado.

- 2026-10-19: D5-A autorizó implementar. Estado intermedio no disponible en tracker (sin projectItems ni etiqueta de progreso). T1 en curso.

- T1: RED observado en los tres aliases (unknown view subcommand), GREEN tras helper compartido; pendiente verificador independiente.

- T1 verificado: AC5 cumplido por hive-verify-task independiente. Binario: 3 aliases éxito, seis comandos y sin crear estado; 12 ayudas por subcomando correctas. Implementador principal modelo de sesión no observado; verificador gpt-6-astra/high configurado por rol, no observado. T2 en curso.

- T2 implementado sin commit: RED CLI token/servidor previo y emisión 404; GREEN ^TestView (4.067 s), viewer (6.499 s), node --check y diff --check según implementador. Pendientes verificación independiente/UI y gates completos. Implementador gpt-6.1-sol/medium configurado, no observado.

- Verificación T2: AC1–AC3 (abridor, canje y rechazos) cumplidos por hive-verify-task; AC4 (inicialización web) pendiente de evidencia browser. Code review completo sin hallazgos P0–P3. go vet y 43 Python correctos; go test ./... correcto (CLI 109.732 s); race global y seguridad en curso. Verificador de tarea gpt-6-astra/high configurado, no observado.

- T2 verificado: AC4 (inicialización web) met tras evidencia funcional independiente revisada por hive-verify-task; SHA256 binario 341c320ed61fda7f48fee21612023cc113d6174e624f6e76132d9b9648540755. Viewports 1440×900 y 390×844, es/en, Chromium 151.0.7922.77 con agent-browser 0.39.1; fragmento vacío/1 canje/0 índices antes de respuesta, recarga/reapertura, códigos rechazados y 61s de expiración, canje/índice/403 antiguos descartados. Sin errores de página ni overflow. Review seguridad sin hallazgos. Gitleaks diff y archivos nuevos sin hallazgos. Race global aún en curso; publicación espera ese gate. Harness y sesiones browser propios eliminados; resultado breve temporal conservado hasta incorporar cierre.

- Gates finales correctos: go vet, go test ./..., go test -race ./... (CLI 322.420 s; sin carreras) y Python 43. Revisión código y seguridad sin hallazgos. Commits T1 f4c8a16a y T2 7e090b06. Hallazgo incidental de costo: suite race supera 5 minutos, una ejecución; no bloquea entrega ni se optimiza dentro del cambio. Pendiente preguntar si registrar investigación de rendimiento al cierre.

- Cierre: PR #83 integrado en b2c0ecb2; issues #78 (token en abridor) y #82 (ayuda) cerrados; sin CI. Binario local reconstruido desde revisión integrada y ayuda verificada; visor cotidiano detenido. Registro se publica con PR separado; limpieza al confirmar merge.
