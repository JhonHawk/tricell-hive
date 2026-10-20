# Visor React con componentes MIT de ReUI

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en development por [PR91](https://github.com/JhonHawk/tricell-hive/pull/91) · 9999f07d |
| Tracker · GitHub Issues | [#89 — migración del visor](https://github.com/JhonHawk/tricell-hive/issues/89) |
| Git | Integrado en `development` · PR91 · 9999f07d; ramas temporales elegibles para retiro |
| Verificación | Pruebas frontend/Go, bundle servido por Go, seguridad y UI local; roles independientes de código/seguridad y UI completa autorizados |
| Siguiente paso | Sin trabajo pendiente de la migración; registro de cierre versionado por PR separado autorizado |

## Solicitud

> Ok, acabo de acquirir el PRO de REUI (https://reui.io/llms.txt)
> Ya tienes el mcp instalado asi como las skills, validalo, y crea una rama nueva para integrar estos cambios (una vez validemos que funcionen), realiza un commit con esto y la idea (abre ticket) es migrar el visualizador a react con reui
> - shell principal es: pnpm dlx shadcn@latest add @reui/app-shell-32 y claro, el contenido es claro que se adaptara a lo que necesitamos (claro que igual todos los componentes), pero esto nos ayudara un monton a implementar nuevos componentes no?
>
> Research intenso de esto

Tras recomendar React + shadcn/ui + componentes gratuitos MIT de ReUI y un shell propio, el usuario respondió:

> ok, vamonos por ese camino entonces

Esa elección reemplaza el bloque Pro y la alternativa de frontend privado. No autoriza por sí sola publicación ni implementación.

## Objetivo y alcance

Reemplazar el shell vanilla del visor por una aplicación React mantenible con componentes existentes y composición propia. Conservar el servidor Go y el renderer compartido de recorridos, sin exigir Node, licencia ReUI ni acceso externo al usuario final.

Incluye frontend, integración de assets, construcción reproducible, pruebas y documentación del desarrollo. Excluye bloques/plantillas/iconos premium, copia o adaptación de shells Pro, asistente IA, búsquedas o vistas nuevas sin requisito, cambio de puerto y rediseño del contenido de los recorridos. Ampliación explícita del usuario: estado de sesión En curso/Cerrada, actualizado por flow-close, con Sin informar para registros sin dato. Esto amplía el manifiesto y el índice derivado, sin migrar copias anteriores ni meta.json.

## Criterios de aceptación

- AC1. El shell servido por el binario usa React y componentes MIT shadcn/ReUI con una composición propia de navegación y lectura. **Falso en la referencia:** el shell actual es `viewer.mjs` vanilla.
- AC2. Un paquete frontend con lockfile produce el bundle estático del visor y permite detectar assets desactualizados respecto de sus fuentes. **Falso en la referencia:** no existe paquete/build frontend del visor ni comprobación de frescura.
- AC3. El shell React conserva apertura autenticada, selección, polling y descarte de respuestas obsoletas; pin y borrado actualizan coherentemente lista y documento. **Falso en la referencia:** estas operaciones todavía no existen en un consumidor React.
- AC4. El binario sirve únicamente los assets declarados del nuevo bundle y ejecuta el visor sin Node ni red de terceros. **Falso en la referencia:** solo se sirven cuatro assets del shell/renderer actuales, sin bundle React.
- AC5. La composición React funciona en escritorio y móvil con teclado, recuperación de errores y CSP vigente, sin romper capturas, videos ni impresión del renderer compartido. **Falso en la referencia:** no hay composición React integrada que permita observarlo.

- AC6. El manifiesto acepta estado explícito opcional `session.status` (`in_progress` o `closed`), el índice informa `session_status` y flow-close publica `closed` antes de limpiar el recorrido; sin dato se informa `unknown`. **Falso en la referencia:** validadores rechazan ese campo y no existe productor de cierre ni confirmación por copia.
- AC7. El cliente React muestra En curso/Cerrada/Sin informar desde el dato explícito, con jerarquía proyecto→ticket→sesión y guías por subnivel, sin derivarlo de `kind` ni de clasificación Git. **Falso en la referencia:** no hay cliente React ni estado de sesión explícito.

## Restricciones de conservación

Mantener contratos API, loopback, Host/Origin, códigos de un solo uso, extracción temprana del fragmento, tokens fuera de URL y capabilities de archivos. Mantener clasificación `tracked`/`local`/`unknown`, exclusiones, avisos, pins, retención y tombstones sin tocar originales. Mantener selección recordada con storage opcional, idioma del shell por navegador e idioma del manifiesto en el recorrido, foco, scroll y protección frente a respuestas antiguas.

No cambiar `content/skills/flow-walkthrough/assets/` ni sus copias byte-idénticas en el visor. No relajar CSP ni introducir recursos CDN para acomodar un componente. No añadir una implementación legacy alternativa al reemplazo final. Conservar los avisos de terceros y revisar licencia de cada fuente: una plantilla marcada Free no equivale a MIT.

## Entrega y decisiones

- Ruta elegida: frontend público propio con componentes MIT; no necesita OEM para esos componentes.
- Historia: preparación local `46526ed4` (reglas, skills y MCP del mantenedor); posteriormente incluida en PR91 bajo la entrega autorizada de la migración.
- Decisión técnica: bundle generado versionado para conservar `go build` sin Node; véase `design.md`. Es reversible cambiando el pipeline, sin migrar datos.
- D2-A (entrega interactiva), elegida por el usuario: commits locales, verificación y pausa de aceptación visual antes de push; tras aceptación, push, PR, revisiones y merge a `development`.
- D3-A (revisión independiente), elegida por el usuario: roles de código y refutación de hallazgos bloqueantes; seguridad por dependencias y nuevo cliente de autenticación.
- D4-A (UI completa), elegida por el usuario: `full`, UX e in-vivo independientes.
- Registro al cierre autorizado por D2-A (entrega interactiva): pequeño PR de cierre integrado por el agente, sin revisión dedicada.
- Sin promoción a `master` ni release. El marcador maintainer, si existe al integrar, activa el refresh local que prescribe `AGENTS.md`; no reiniciar el visor cotidiano por inferencia.

## Estado

Estado final, 19 de octubre de 2026: cinco tareas verificadas, aceptación humana mediante D11-A (candidato final) y D12-A (reinicio coordinado). Integrado mediante PR91 en development9999f07d, incluido fix30d59619 para los dos P2 de revisión. Visor cotidiano18322/50979 y contenido instalado actualizados; preservados2repos,4copias,0pins,28archivos y336h con hashes iguales. Cierre del recorrido confirmado por renderer/CLI real. Sin ActionsCI configurado, master ni release. Historia y límites en tasks.md.

## Criterios del ticket al leerlo

Lectura del 19 de octubre de 2026; transcripción literal. La decisión MIT posterior resuelve el primer criterio sin incorporar premium. Los criterios nuevos del plan califican el consumidor React; las garantías ya existentes se conservan como restricciones.

> - Ruta de licencia resuelta y documentada antes de incorporar materiales premium; conservar avisos MIT de componentes públicos.
> - Build reproducible y assets servidos por el binario Go; instalación/ejecución sin Node, licencia ReUI ni red de terceros.
> - Conservar intercambio de código de un solo uso, limpieza temprana del fragmento, token fuera de URLs, validación Host/Origin y capabilities de imágenes/videos.
> - Conservar separación de sesiones locales/versionadas/desconocidas, exclusiones, pins, TTL, borrado de copias y tombstones sin modificar originales.
> - Conservar selección/foco/teclado, navegación móvil, inglés/español, polling, reconexión y protección contra respuestas obsoletas.
> - Estados vacío/carga/error/reintento, temas claro/oscuro, capturas, video, carrusel e impresión verificados en bundle compilado servido por Go.
> - Resolver compatibilidad con CSP actual (sin scripts/estilos inline autorizados) sin relajarla automáticamente; comprobar overlays, posicionamiento y estilos globales.
> - Mantener igualdad de assets del renderer offline y ejecutar pruebas apropiadas de servidor/seguridad, frontend y verificación visual local.

## Ampliación y corrección de interfaz

D9-A (estado explícito) y aclaración del usuario: «si esta en curso, o cerrada (cuando pasa con flow-close)». Referencia visual actual: shell33 público, sin fuente Pro. Corrección2 añade ramas y sangrías distintas y badges de fixture. Se agregan AC6/AC7 y tarea T5 para el contrato/productores; material para re-revisión de plan en el gate de candidato aceptado, sin lanzar revisores entre correcciones. Historia del plan: esa ampliación requirió revisión antes de ejecutar T5; fue resuelta y verificada, como registra tasks.md.

### Solicitud literal de la ampliación

> Las lineas de la carpeta principal no hace las marcas para los subniveles
> Los registros y los tickets parecen estar al mismo nivel, no hay mucha distinción de jerarquia o nivel
> En el sidebar falta algo (Esto es nuevo) un badge que nos indique cuando una sesión ya termino) o sea, estados

> En el shell-33 viene el ejemplo que me gusta pnpm dlx shadcn@latest add @reui/app-shell-33

> que debe representar? si esta en curso, o cerrada (cuando pasa con flow-close)

D10-A (aceptación de jerarquía y badges) acepta la composición de corrección2 y habilita sus gates; no es aceptación final ni habilita push por sí sola.
