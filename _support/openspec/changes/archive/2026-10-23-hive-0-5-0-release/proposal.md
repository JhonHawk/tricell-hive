# Publicación de Hive 0.5.0

| Campo | Valor |
| --- | --- |
| Estado | Completado · publicación verificada |
| Tracker | GitHub Issues · sin ticket asignado a esta publicación |
| Git | PR a development, promoción a master y tag v0.5.0; seguimiento VERSION=dev |
| Verificación | Suites locales, favicon HTTP/navegador, paquetes en macOS y Linux ARM64/AMD64, sitio y artefactos publicados |
| Siguiente paso | Conservar el registro archivado; higiene de la entrega |

## Solicitud

> - Resolvamos el tema del favicon
> - Generemos nueva versión de hive para todo lo que tenemos acumulado, actualicemos la doc y publiquemos

## Objetivo y alcance

Publicar los cambios acumulados desde v0.4.0, incluidos el visor ya aprobado, archivo/persistencia y favicon, con documentación y paquetes verificables. La última publicación y master corresponden a c92992485b3d0ad5f5b50e12760aa081cd78fa06; development remoto y base local son 6210efb9d27594dfa0a3130a0413162714b82088, con 160 commits acumulados. No hay CI configurado. CONTRIBUTING exige versión menor por comandos, skills, roles y opciones nuevos: 0.5.0.

- AC1. El favicon del visor responde a GET/HEAD con imagen Hive válida, mismo origen y MIME correcto; el HTML y Vite usan el asset. Falso en la base: favicon404.
- AC2. VERSION, README y el changelog canónico presentan 0.5.0 y documentan los cambios acumulados; Unreleased queda vacío. Falso en la base: VERSION=dev e historial termina0.4.0.
- AC3. Los tres paquetes 0.5.0 pasan instalación, verificación y recuperación en runtime compatible, con inventario/checksums íntegros. Falso en la base: no existen paquetes 0.5.0.
- AC4. La release pública v0.5.0 contiene tres tar.gz y tres sha256 construidos desde el tag de master; los bytes descargados coinciden con los verificados. Falso en la base: última release0.4.0.
- AC5. El sitio publicado desde ese tag muestra0.5.0 y sirve sus recursos; la revisión desplegada corresponde al candidato. Falso en la base: sitio no publica0.5.0.

## Restricciones y contratos

Preservar _support/sessions/ ajeno y cambios del usuario. No repetir pilotos de modelos. Mantener autenticación, CSP, copias y retención del visor. El estado es informativo; archivar sólo cambia metadata de la copia y no pin/TTL/estado/fuente. Metadata archived ausente equivale false; no backfill; downgrade a binarios de visor anteriores necesita restaurar copias archivadas para omitir el campo. Los 8 registros habituales no se migran ni se borran.

El visor fue aceptado visualmente y pasó UI full en esta sesión, árbol cb52ef579375a8e8903056c2e8b6df923e138390. Esta publicación sólo añade el favicon; no cambia composición. Se conserva la decisión sin revisión de código dedicada y UI full ya satisfecha para el visor; se comprueba el delta del icono en navegador. El sitio conserva su diseño y requiere comprobación de las páginas documentales/versionadas afectadas.

## Entrega y recuperación

La instrucción de publicar autoriza commits/push/PR/merge a development, promoción master, tag/release GitHub y despliegue del sitio, además del seguimiento VERSION=dev. Método merge commit según historial. El registro se versionará al cierre por PR pequeño sin revisión dedicada a development, fusionado por el agente. Maintainer marker exige refresh de binario/contenido al integrar development; preservar el viewer habitual y no confundir binario reemplazado con proceso reiniciado. Coordinar el reinicio existente si hace falta cargar el nuevo binario, conservando estado/puerto/TTL.

Preflight construye los tres paquetes juntos en salida nueva antes de publicar. El builder usa -buildvcs=false e identidad de archivos: reconstruir desde checkout limpio del tag debe dar los mismos bytes si no cambian insumos. Si los bytes difieren, no publicar hasta explicar y repetir los runtimes afectados. Ejecutar instalación y recuperación macOS ARM64 y Linux ARM64/AMD64 mediante Docker aislado, sin hosts reales. Después publicar archivos y comprobar descarga. No mover ni reescribir tags publicados; corregir hacia delante si ya se publicó una versión defectuosa. La promoción sigue development→master; no commits directos a master.

## Referencias verificadas

- CONTRIBUTING.md: ramas, versionado, suites, build reproducible y publicación de sitio desde tag.
- _support/docs/architecture/installer.md: instalador empaquetado y recuperación en runtimes compatibles.
- tooling/package/main.go: tres targets, inventario congelado, buildvcs=false y checksums.
- site/wrangler.jsonc: Worker hive-docs, ruta hive.tricell.tech/*.

## Progreso

Favicon implementado en el checkout original y trasladado a este worktree; pendiente verificación independiente. Plan revisado en entrega y ejecutado; recuperación real concretada según hallazgo P2. PR #100 (visor y preparación) integró visor/release en development 04f64f6b; PR #101 (promoción de release) promovió master 8c2882bd y tag v0.5.0; PR #102 (seguimiento de versión) devolvió development a dev en b85f2c4b. El trabajo ajeno quedó preservado. El registro se archiva y versiona por PR de cierre sin revisión dedicada, autorizado en esta entrega.

## Resultado verificado

- Release pública: https://github.com/JhonHawk/tricell-hive/releases/tag/v0.5.0; latest, no draft/prerelease. Seis assets descargados iguales al preflight;105 hashes internos por paquete correctos.
- Fuente publicada: tag v0.5.0 en 8c2882bd172d723581c98602c8a62862cc6ec3a4; source ID fc68a8ceb1075a7088226288f0bf8c376a6fbaae0a9c6326fae83063871c7cae.
- Sitio: https://hive.tricell.tech, Worker hive-docs, deployment ec9c53d9-c2e7-45c3-a5f6-1a761c4d4609, versión fced1547-6884-4f1a-8b93-b824bec29565 al 100%; tag/mensaje del commit coinciden. Portada/changelog/build/favicon/CSS200 y bytes idénticos al tag; instalador302→GitHub 200 idéntico.
- Manager local reconstruido desde04f64f6b;562 recursos instalados y verificados en6 hosts. Su versión de desarrollo es dev. No se reinició el visor habitual, PID 18322/puerto 50979/TTL 336h; reiniciarlo requiere petición explícita.
- Límites: Linux AMD64 validado mediante emulación; consumidor de paquete codex en hogares sintéticos. Cuatro skips de plataforma/capturas/helper en Go. Sin recorrido funcional público: salud/identidad solamente. urllib con agente predeterminado recibió403; curl obtuvo200 y bytes correctos. Las advertencias Astro de head-inject/i18n/404 no mostraron fallo adicional en páginas/metadatos revisados.
- Evidencia de sesión curada en _support/workspace/2026-10-23-hive-0-5-0-release/{images,evidence}/ del checkout principal. Temporal de entrega y checkouts propios se retiran al cierre; _support/sessions ajeno se conserva.
