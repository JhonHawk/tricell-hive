---
title: "«Nombre de la pantalla»"
description: "«Una frase: qué es esta pantalla y para qué sirve. Es obligatoria.»"
---

<!--
  PLANTILLA DE PÁGINA — pantalla del manual de usuario.
  Las 6 secciones H2 de abajo son fijas y van en este orden en TODA página de
  pantalla. El cuerpo empieza en `##` (el H1 lo pone `title`). La primera frase
  de cada sección carga el dato distintivo (front-loaded). Voz: consistente con
  el manual (por defecto "usted" formal, la convención establecida).
  Reemplaza los «placeholders» y borra estos comentarios al instanciar.
-->

Ruta: `/«ruta-en-la-app»`. «Una o dos frases que ubican la pantalla en el
producto y dicen a quién sirve.»

## Propósito

«Por qué existe esta pantalla y qué problema resuelve para el operador.»

## Quién puede acceder

«Tabla de capacidades por rol. `✓` permitido, `✗` negado, `—` no aplica.»

| Capacidad | `org:owner` | `org:admin` | `org:agent` | `org:viewer` |
| --- | :---: | :---: | :---: | :---: |
| «Ver la pantalla» | ✓ | ✓ | ✓ | ✓ |
| «Editar» | ✓ | ✓ | ✗ | ✗ |

> La UI esconde; el backend niega. Un rol sin permiso no ve la acción, y si la invoca, el backend la rechaza.

## Tour de la pantalla

«Áreas de la UI, cada una en **negrita**, seguidas de la captura. El texto debe
bastar por sí solo; la captura confirma, no sustituye.»

- **«Área 1»** — «qué muestra o qué hace».
- **«Área 2»** — «qué muestra o qué hace».

<!-- Captura (agrégala cuando exista el .webp real en src/assets/; el alt describe la función):
     ![«Descripción funcional completa de la captura, en español.»](../../assets/«seccion»/01-«slug».webp) -->

## Flujos paso a paso

### «Nombre del flujo»

1. «Paso con la acción de UI en **negrita** y el resultado observable.»
2. «Siguiente paso.»

## Reglas de negocio

«Reglas que gobiernan la pantalla, nombradas. La fuente de verdad son las épicas
del proyecto; aquí se explican para el operador, no se redefinen.»

- «Regla 1.»
- «Regla 2.»

## Estados y errores

«Cada estado o error con su significado y la recuperación. Usa asides con
título; no los apiles (un muro de cajas cansa) — agrúpalos por significado.»

:::note[«Estado normal»]
«Qué significa y qué hacer.»
:::

:::caution[«Advertencia»]
«Qué la dispara y cómo evitarla.»
:::

:::danger[«Error bloqueante»]
«Qué salió mal y cómo recuperarse.»
:::
