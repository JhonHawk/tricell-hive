# Design

## Approaches

La propuesta aceptada usa paquetes inmutables por versión y un comando estable. Reemplazar `bin/hive` dentro del paquete existente invalidaría su manifest y perdería la unidad entre ejecutable y fuentes; además, hoy los hooks registran la ruta real del ejecutable. No se elige ese enfoque.

Para recuperación se presentaron dos rutas: respetar el commit existente y completar activación, o ampliar el motor para deshacer también operaciones confirmadas. El usuario eligió **D1 A**, la primera, que mantiene el significado del commit actual. No se promete atomicidad de todos los archivos ni reversión de una operación ya confirmada.

## Decisions

- Origen de release predeterminado, fijo y HTTPS; selección estable siguiendo la release latest oficial, no desarrollo ni prereleases. No inferir el modo por cwd.
- `--source` o `--rev` seleccionan modo Git; `--rev` sin `--source` usa `.`. `--version` no se combina con ninguno. El modo Git no recompila ni activa binarios y sigue excluyendo cambios sin commit.
- Un binario `dev` sin origen explícito falla antes de descargar e indica el modo Git. No se convierte una instalación de desarrollo en release por inferencia.
- `--version` permite una selección anterior explícita sólo cuando el protocolo y el esquema del estado son compatibles; se muestra como downgrade. Una versión anterior al protocolo nuevo conserva la vía de instalador manual.
- Se retienen paquetes anteriores. No se agrega un comando general de rollback de operaciones confirmadas.
- El enlace estable es `~/.local/bin/hive`, sin editar PATH, perfiles ni enlaces ajenos. Se registra propiedad y fingerprint. Una colisión no administrada se rechaza y el mensaje da la ruta nueva utilizable; mover o adoptar el ejecutable ajeno exige una acción explícita aparte.
- Mantener los hooks apuntando al ejecutable real del paquete destino. `HiveExecutable` conserva su semántica para source/install normales; la preparación de release incorpora el binding validado entre ejecutor temporal y destino final descrito abajo. El enlace estable sirve para las invocaciones del usuario.
- El paquete offline conserva instalación sin red. La activación del comando es parte de la instalación gestionada del paquete, visible en la confirmación, sin ejecutarse para una instalación desde checkout.
- La actualización de release no instala ni ejecuta Engram, Context7, pi-subagents, agent-browser u otros proveedores; conserva las capacidades existentes. Un requisito nuevo de proveedor se informa para instalación explícita aparte.
- **D4 A**: si Pi está registrado y su declaración de pi-subagents falta o exige reparación, el preflight de release update bloquea la operación antes de core writes y activación y nombra `hive install` como reparación separada. La excepción se refleja explícitamente en el requisito de Pi; install y source update conservan su comportamiento. Los casos con declaración válida no invocan Pi.
- Preview puede descargar a un temporal privado; no mantiene paquetes en el almacén ni altera instalaciones. `--out` sí retiene un paquete verificado y un plan privado para aplicación offline, sin activar nada.

## Proposed shape

### Distribución y protocolo entre versiones

Extender `tooling/distribution` con resolución de release, cliente de descarga y preparación segura. Usar la biblioteca estándar, origen fijo y límites de tiempo, bytes comprimidos, extracción y entradas. Verificar HTTPS y hosts permitidos en cada salto; probar redirecciones y respuestas truncadas. Separar identidad del producto, SHA del archivo, identidad del manifest y release de contenido.

El paquete añade un marcador de protocolo de update versionado al manifest; su bootstrap mínimo queda estable y se incluye en los hashes. El lanzador conoce sólo ese bootstrap, no el esquema de plan o estado de versiones futuras. Primero valida checksum, estructura, plataforma, identidad y hash de binario sin ejecutar el paquete. Después invoca por ruta absoluta el ejecutable verificado mediante un subcomando interno fijo, con entrada/salida estructurada acotada y sin shell; no ejecuta comandos leídos del manifest. El ejecutor nuevo valida el manifest completo y el estado antes de producir un plan. Un protocolo ausente o desconocido se rechaza con instrucciones de instalación manual.

Durante preview el ejecutor reside en el temporal y el paquete final puede no existir. Separar `PreparationExecutor` de `TargetExecutable` en el contrato: el primero fija la ruta temporal y hash ejecutados; el segundo fija la ruta real futura `<package-store>/<label>/bin/hive`, hash y binding al manifest validado. Sólo el protocolo de release verificado puede aportar ese binding; no es un override genérico de la ruta del hook. La preparación verifica la ruta final canónica y sus ancestros, la identidad y los bytes del paquete, y renderiza los hooks con la ruta final, permitiendo que ese archivo aún no exista. El plan congela ambos roles y el destino; no registra el temporal en ningún hook. Antes de aplicar, el paquete se coloca/verifica en el destino definitivo y BuildPlan/validatePlan/generación de hooks exigen de nuevo archivo regular, ruta sin symlinks y digest coincidente. El ejecutor de apply es el binario definitivo. No relajar esas comprobaciones para apply ni para los caminos source/install normales. Probar preview con almacén vacío y hooks de todos los hosts registrados en home sintético.

Los contratos tipados se implementan en una primera tarea: descriptor de artefacto, identidad del ejecutor, prepare/apply/recover, plan padre y fases. Incluyen versión de protocolo, artefacto/digests, plataforma, directorio de paquete, ruta/digest del ejecutor, plan del motor congelado, destinos/fingerprints, cohort de hosts, ID de operación y asociación inequívoca con la transacción del motor. Las rutas no se derivan de texto no validado. La respuesta de éxito distingue prepared, core_committed y activated; un exit code por sí solo no demuestra commit.

### Paquete, comando y migración

Paquetes nuevos viven bajo el almacén existente `${XDG_DATA_HOME:-$HOME/.local/share}/hive/packages/<label>/` con XDG absoluto. Se verifica un paquete existente antes de reutilizarlo y nunca se rebindea una versión a bytes distintos. La preparación usa un directorio privado y activa el directorio validado mediante rename seguro, sin mutar paquetes anteriores.

Instalar desde el bootstrap nuevo registra y activa el comando estable, después de la instalación del core, usando el mismo coordinador de activación. El `install.sh` offline permanece sin descargas. La migración actualiza los hooks con el ejecutable del paquete nuevo. No puede actualizar retrospectivamente el binario viejo: el usuario ejecuta el instalador nuevo una vez. La instalación desde checkout no crea ni adopta el enlace del modo release. La desinstalación sólo retira un enlace propio cuyo fingerprint coincida, preserva paquetes retenidos y no deja un recibo activo falso.

### Coordinación y recuperación

Un registro padre independiente del esquema del motor guarda la operación completa en el state home, bajo un namespace propio de update. Guarda intención y fingerprints antes de cada mutación, la asociación con la transacción hija, el paquete y ejecutor retenidos, la fase y la activación observada. Directorios 0700 y archivos 0600.

Fases conceptuales: prepared → core_pending → core_committed → activation_pending → activated. Antes del commit, recovery usa el ejecutor destino para recuperar el motor y conserva el comando anterior. Después, termina la activación mediante compare-and-swap del enlace y verificación de identidad. Si el usuario cambió el enlace, se conserva el conflicto; no se sobrescribe ni se declara éxito.

La asociación al journal del motor se persiste antes de su primera mutación, mediante un ID de transacción reservado o un ID padre en su metadata. El commit se obtiene de evidencia durable del journal, también si el proceso muere antes de devolver su resultado. Esa evidencia y el paquete no se podan mientras el padre esté pendiente. No se deduce el commit comparando sólo contenido o versiones ni buscando el journal más reciente.

Bloqueo: orden único update → motor. Toda mutación del core detecta un padre pendiente y exige su coordinador; el ejecutor hijo usa una autorización interna limitada a ese ID, sin una excepción genérica que salte fingerprints. Las lecturas siguen disponibles; con esquemas nuevos que un lanzador viejo no entiende, el bootstrap de recuperación se consulta antes de decodificar el estado del motor y deriva al ejecutor retenido. No se sostienen locks en un padre mientras se espera a un hijo que los requiere. No prometer exclusión frente a binarios históricos que no implementan este contrato; conservar sus fallos por esquema y documentar que no deben usarse para mutar una actualización pendiente.

El modo de planificación del core distingue explícitamente release update de install/source para Pi. Release preflight evalúa la declaración sin programar pasos externos; si el motor requeriría instalar o reparar pi-subagents, devuelve un conflicto accionable antes de aplicar. También revalida esa condición al aplicar un plan guardado: no puede colarse un paso Pi editando el plan. No basta omitir el catálogo de capacidades opcionales, porque el requisito de Pi se resuelve dentro de BuildPlan y Apply.

Un fallo previo al core puede dejar sólo un paquete validado retenido, sin cambiar la instalación activa. Un fallo tras el commit mantiene las integraciones destino y muestra una ruta absoluta al ejecutor retenido para completar recuperación sin red. Borrar o corromper manualmente ese paquete no tiene recuperación offline garantizada: se detecta y exige restaurar el artefacto exacto.

### Planes guardados

El plan padre usa un discriminador y una versión propios, e incluye la huella del plan del motor y del paquete retenido. `hive apply --plan` reconoce el padre antes de cargar el plan del motor y despacha al ejecutor fijado, que repite todas las verificaciones. No vuelve a consultar latest ni permite sustituir el ejecutor. El mismo plan alterado o con estado/destino obsoleto falla antes de core writes; una aplicación ya completada se reconoce por su operación y no se duplica.

### CLI y TUI

Sin terminal, update de release sin `--dry-run` ni `--out` falla antes de descargas. La confirmación cubre core y comando en una sola operación. Un resultado parcial nunca imprime “Hive updated” como éxito completo. El comando source conserva la oferta existente del Codex wait floor; no se añade esa oferta al modo release.

TUI: mantener los inputs, teclas, spinner y confirmación existentes. Cambiar sólo Update, sin rediseño del sistema visual. Los errores de red permiten volver a preparar y nunca reutilizan bytes incompletos. Escape cancela antes de apply; durante la aplicación muestra estado y recuperación en lugar de sugerir rollback completo.

Before:

```text
Update
> Source    .
  Revision  HEAD
Enter resolves the revision in a Git checkout...
```

After:

```text
Update
> Mode      Release
  Version   latest stable
Current: <installed>   Target: <resolved>
Enter checks the release and previews changes.
```

```text
Update
> Mode      Source
  Source    .
  Revision  HEAD
Enter resolves the committed source and previews changes.
```

Estados: sin hosts → instrucción de instalación; sin actualización → versión actual y sin activación redundante; preparando → spinner; listo → resumen/confirmación; conflicto o red fallida → error y reintento; recuperación pendiente → ejecutor retenido y acción de recuperación. El modo inicial de un build dev es Source. La TUI usa el mismo coordinador y no instala desde el cwd por inferencia.

## Copy

Mensajes del producto en inglés, concretos y con rutas seguras; redactar y probar los casos: versión incompatible, descarga fallida, enlace ocupado por otro archivo, actualización completa, operación pendiente y downgrade explícito. Tras éxito informar que las sesiones abiertas deben reiniciarse. Consultar el visor existente sin mostrar credenciales; si continúa con otra revisión, informar que sigue en esa revisión y requiere un reinicio aparte. No arrancarlo ni reiniciarlo como efecto de update.

## Impact

`tooling/cli/update.go`, `tui_update_view.go`, `main.go`, install/apply/recover dispatch, `tooling/distribution`, `tooling/management` (coordinación, bloqueo, recibos de comando y retención de journals), `tooling/package`, `get-hive.sh`, `install.sh` según el contrato final, tests de esos módulos, README, llms.txt, CHANGELOG, AGENTS de mantenimiento, arquitectura y getting-started del sitio. No se toca runtime ni assets del visor. No hay promoción ni publicación en producción.

Actualizar los ejemplos de mantenimiento que hoy usan update sin origen a `hive update --source <checkout>`, y la ayuda/TUI y spec del modo Git. No cambiar las reglas distribuidas de autorización o de revisión por esta función. Registrar el tamaño antes/después de `content/guidance/global.md` en la entrega: se espera diferencia cero.

## Recovery

La recuperación operativa es **D1 A**, descrita arriba; la versión anterior se conserva como evidencia y herramienta, no como garantía de rollback de estado migrado. Para revertir el código de la función antes de publicarla basta el diff Git. Después de una migración de estado, una versión previa sólo se ejecuta si reconoce ese esquema. Release pública, promoción a master y reinicio de servicios quedan fuera de este plan.
