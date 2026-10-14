# Diseño: fuente Hive de pi-subagents sin versión

## Verified context

- Constante actual: `integrations/pi/package_settings.go` `HiveSubagentsSource = "npm:pi-subagents@0.74.0"`. `ClassifySubagents` ya trata `npm:pi-subagents` y `npm:pi-subagents@…` como presentes (`package_settings_test.go` caso `unpinned`).
- Pin 0.74.0: D3-B del cambio archivado `2026-10-01-gh-29-install-pi-subagents` (lo que había en esta máquina). Ese cambio aparcó el fallo de hijos asíncronos en Pi 1.0 (H1).
- Observado 2026-10-11, sesión Pi `01a12d85`, Pi 1.1.0: `pi-subagents` 0.74.0 falla al arrancar hijos en segundo plano con «does not provide @earendil-works/pi-agent-core/node». El changelog de 0.75.0 describe ese arreglo.
- Pi 1.1.0 `package-manager.js` `normalizePackageSourceForSettings`: para `npm:` devuelve la fuente sin reescribir. `pi install npm:pi-subagents` deja `npm:pi-subagents` en `settings.json`. `recover` puede usar esa cadena exacta.
- Catálogo opcional: solo Engram, Context7 y pi-subagents. Engram y Context7 siguen manuales. Context7 ya se sugiere con `npx ctx7@latest`. D5-A no añade otros paquetes de Pi.
- `pi update --extensions` puede actualizar un paquete sin pin; Hive no lo ejecuta. Quien ya tiene `@0.74.0` no se actualiza con este cambio.

## Proposed design

Una sola constante `HiveSubagentsSource = "npm:pi-subagents"`. Plan, `execApplyPi` / `syntheticApplyPi` y recover siguen usándola. No se cambia la regla de no reescribir una declaración existente.

El `package.json` sintético de apply no debe fingir la versión 0.74.0 (eso implicaba el pin). Escribir `"name":"pi-subagents"` y una versión claramente ficticia, por ejemplo `"0.0.0-synthetic"`, porque el falso no llama a npm. Las pruebas no leen esa versión; las que comprueban un pin de usuario (`@0.67.0`, `@0.74.0` con `extensions: []`) se quedan.

Goldens `state_without_overrides.golden` que hoy copian `@0.74.0` como fuente Hive pasan a `npm:pi-subagents`.

### Contratos

- Entrada: plan de usuario con host `pi`; settings de Pi ausente / presente / conflicto.
- Salida: `pi install npm:pi-subagents` o omisión; recover `pi remove npm:pi-subagents` solo si esta operación añadió esa cadena y sigue ahí.
- Compatibilidad: un usuario con `@0.74.0` u otro pin sigue en Present; Hive no lo toca. Quien instale Hive de nuevo en Pi vacío obtiene la última de npm ese día, incluido un posible corte de API de `pi-subagents`.
- Dependencia: npm en apply real (igual que hoy). Pruebas: `pi` falso.

### Rechazado

- Pin `@0.77.0`: reproducible, pero vuelve a envejecer (D4, no elegido).
- Instalar también gentle-engram, rpiv y pi-web-access (D5-B): fuera del catálogo y del «no instalar herramientas de más».
