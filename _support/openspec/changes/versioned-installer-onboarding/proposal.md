# Instalación guiada y versiones públicas de Hive

| Campo | Valor actual |
| --- | --- |
| Estado | Implementación local completa (`hold`) · pendiente recorrido humano y autorización de entrega |
| Tracker · GitHub Issues | [#29 — instalación de pi-subagents](https://github.com/JhonHawk/tricell-hive/issues/29), para esa parte; sin issue general |
| Git | `hold` (D4-A) · base `rebuild/harness-engineering` · plan y código locales, sin commit ni publicación |
| Verificación | Tests Go · terminal y paquetes en hogares sintéticos · integración de proveedores por plataforma · revisión del plan completada |
| Siguiente paso | Recorrido humano del candidato local; después, entrega según D4-A si se autoriza |

## Objetivo

Instalar y actualizar Hive mediante un asistente que permita elegir CLIs y capacidades operativas, explique los requisitos ausentes y muestre versiones legibles. Ofrecer una entrada online de descarga pública sin acceso al repositorio fuente, manteniendo el paquete offline y el motor existente de planificación, aplicación y recuperación.

## Decisiones del usuario · 2026-09-25

- D1-A: conservar recursos compartidos; cuando cambien, solicitar seleccionar juntos todos los consumidores afectados. No prometer versiones independientes por CLI ni aislamiento de descubrimiento.
- D2-A (reemplazada en su parte automática por D1-B del 2026-09-26: las tres capacidades se ofrecen como instrucciones manuales hasta su puerta nativa): primera entrega con instalación opcional de Engram, Context7 y `pi-subagents`. Para hosts ausentes, navegador, Python, Git y GitHub CLI: detectar, explicar e indicar instalación oficial; no instalarlos automáticamente.
- D3-A: diseñar descarga pública de paquetes, independiente de la visibilidad del código. Esta decisión no autoriza crear infraestructura pública, cambiar visibilidad ni publicar archivos.
- D4-A y D5-A (2026-09-25): entrega `hold` y revisión de implementación `review-code`/`review-refuter`; no iniciar implementación con estas respuestas.
- Selección explícita de destinos: desmarcar no significa desinstalar. Conservar configuración, historial, credenciales y paquetes preexistentes.
- Versión pública de Hive para el usuario; hashes internos para identidad e integridad. Versiones de proveedores separadas.

## Decisiones del usuario · 2026-09-26

- Idioma (D1-A de esa ronda): todo texto del instalador visible para el usuario queda en inglés, como estaba antes de este cambio.
- D1-B: se retiran las tres recetas ejecutables; las capacidades se ofrecen como instrucciones manuales hasta que cada una pase su puerta nativa. A5 se cumple mostrando soporte no verificado como manual.
- D2-B: `Revert` sale de la interfaz de pasos externos; #29 sigue abierta hasta la puerta nativa de Pi.
- D3-B: se aplican ahora todas las simplificaciones de la auditoría de mantenibilidad.
- D4-B: el journal padre de onboarding se conserva para recetas futuras; las capacidades seleccionadas quedan como pasos `manual` y el resultado es parcial.

## Alcance y aceptación

- A1. Paquetes nombrados por versión y plataforma; el binario y el estado muestran versión pública, identidad del contenido y límites de compatibilidad. Un checkout sin versión de release se identifica como desarrollo, no como release publicada.
- A2. Asistente con selección de los seis hosts actuales, detección de ejecutables e instalaciones previas, selección vacía/cancelación sin escrituras y lista explícita de consumidores compartidos afectados. Un host ausente no se declara operativo.
- A3. Catálogo que separa instalación, operación, capacidades condicionales y desarrollo. Para cada dependencia: propósito, alcance, detección, alternativa y consecuencia de omitirla. Solo tres proveedores instalables en esta entrega.
- A4. Antes de cambios persistentes o ejecución de proveedores, resumen de herramientas, versiones resueltas, hosts, destinos, red y configuración que cambiarán. Consentimiento explícito, sin actualización automática de herramientas existentes ni instalación implícita de Node/npm, Python, gestores o hosts.
- A5. Ejecución de recetas oficiales verificadas para la combinación host/plataforma; preservar configuraciones y datos preexistentes. Soporte desconocido se muestra como manual/no verificado, nunca como éxito. Pi cumple el contrato de #29.
- A6. Descarga online de una versión explícita o resolución única de la estable; validación antes de ejecutar; terminal independiente del stdin del script. Sin terminal, EOF o cancelación: ninguna aplicación. Offline sigue funcionando sin red ni Go/Git/gh.
- A7. Resultado por componente: instalado, configurado, verificado, omitido, fallido o pendiente de autenticación. Una herramienta opcional fallida no se confunde con fallo del núcleo ni con instalación completa de la capacidad.
- A8. Reejecución idempotente, detección de cambio de estado entre selección y aplicación, recuperación de interrupciones y conservación de la historia anterior. Versiones previas sin etiqueta aparecen como legacy/sin versión, conservando su hash.
- A9. Pruebas ordinarias, interacción con terminal y recorridos de paquetes cubren éxito y fallos. No se lanzan modelos. Cada plataforma publicada deberá tener evidencia de ejecución nativa, no solo compilación cruzada.
- A10. Documentación de usuario e inventario de dependencias separados del mantenimiento; contrato de versionado, actualización, recuperación y publicación explícito.

## Fuera de alcance

Publicar releases o bootstrap, crear repositorios/buckets/dominios, cambiar visibilidad, instalar en el HOME cotidiano, gestionar suscripciones/modelos, instalar automáticamente otros proveedores, migrar helpers Python a Go, añadir Windows o macOS Intel, reanudar pilotos y limpiar trabajo ajeno. La primera entrega no introduce autoactualizaciones en segundo plano, una TUI de terceros ni un gestor universal de paquetes.

## Contexto y continuidad

Fuentes actuales: [instalador](../../../docs/architecture/installer.md), [manager](../../../docs/architecture/deployment-manager.md), [agentes](../../../docs/architecture/agent-delivery.md). Describen comportamiento vigente; este cambio propone reemplazar las fronteras de onboarding correspondientes al integrarse. El inventario exploratorio permanece en conversación; no se crea otro reporte de investigación.

El checkout tiene cambios ajenos en agentes, perfiles, guía y pruebas. Solo esta carpeta pertenece a la planificación. La implementación debe reconciliar ese estado y trabajar en un checkout aislado o sobre una base acordada, sin incorporar ni revertir esos cambios. El empaquetador copia el working tree: no distribuirlo accidentalmente como release.

## Entrega

Plan local autorizado. D4-A fija `hold`: cuando se autorice implementar, verificar en aislamiento y entregar diff y recorrido humano, sin commits. D5-A autoriza entonces revisión con `review-code` y `review-refuter` para hallazgos bloqueantes tras las pruebas. Implementación aún pendiente. Publicación y HOME real quedan separados. El plan permanece sin versionar conforme al modo elegido.

Destino público definitivo y versión de primera publicación se deciden al publicar. Durante implementación, un servidor de fixtures local valida el protocolo; no inventar una URL pública ni mostrarla como operativa. La configuración del origen público forma parte del artefacto de distribución, no de la detección del remoto Git del usuario.

## Inicio de ejecución

El usuario invocó `flow-build`. Base reconciliada `bacf5d5`; solo la carpeta del plan estaba sin seguimiento, sin modificaciones ajenas. Se trabaja en el checkout actual con ownership por archivos, sin aislamiento adicional innecesario. Tres hijos implementan proveedores, distribución y CLI; principal integra management y verifica. Esta actualización reemplaza la condición histórica de implementación pendiente; `hold` y exclusiones siguen vigentes.
