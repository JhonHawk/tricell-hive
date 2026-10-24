# Spec: descargas opcionales de agent-browser

| Campo | Valor actual |
| --- | --- |
| Estado | Validado por el usuario (D15-A) · en PR con `/code-review` |
| Tracker · GitHub Issues | Sin incidencia: hallazgo H2 del cierre de #108, resuelto en el momento por decisión del usuario (D13) |
| Git | Interactivo (D3-A de esta sesión) · PR → `/code-review` (D4-A) → merge a `development` |
| Verificación | `rg` sobre el spec · suite local |
| Siguiente paso | Validación del usuario |

## Solicitud

> Resolverlo ahora

(Respuesta del usuario a D13 en el cierre de #108, 2026-10-23.)

## Objetivo

El requisito «Verified release package installation» de `versioned-installation` decía que solo el paso de pi-subagents puede descargar. El instalador también descarga, tras el consentimiento del operador, con la receta fija de agent-browser (`tooling/providers/run.go:14-18`: `brew install agent-browser`, `npm install -g agent-browser`, `agent-browser install`), como ya documenta `SECURITY.md:28-38`. El requisito ahora lo nombra.

## Alcance y aceptación

- AC1. El requisito «Verified release package installation» del spec vigente nombra las descargas de la receta de agent-browser tras el consentimiento y tiene un escenario para ese caso. *Falso en la base cuando* dice «only the Pi host's pi-subagents step may download».

Restricciones: no cambia código ni `content/`; el resto del spec queda igual.
