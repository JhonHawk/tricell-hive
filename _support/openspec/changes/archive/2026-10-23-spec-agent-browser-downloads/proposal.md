# Spec: descargas opcionales de agent-browser

| Campo | Valor actual |
| --- | --- |
| Estado | En validación · alcance ampliado (D16-A) · delta fusionado en el spec en el PR #116 · esperando la validación del usuario |
| Tracker · GitHub Issues | Sin incidencia: hallazgo H2 del cierre de #108, resuelto en el momento por decisión del usuario (D13) |
| Git | Interactivo (D3-A de esta sesión) · PR → `/code-review` (D4-A) → merge a `development` |
| Verificación | `rg` sobre el spec · suite local |
| Siguiente paso | Validación del usuario; después `/code-review` ya hecho, merge |

## Solicitud

> Resolverlo ahora

(Respuesta del usuario a D13 en el cierre de #108, 2026-10-23.) D16-A: fusionar el delta completo de agent-browser. D17-A: abrir una incidencia por los deltas archivados sin aplicar.

## Objetivo

El spec `versioned-installation` decía que solo el paso de pi-subagents puede descargar, aunque el instalador también ejecuta, tras el consentimiento, la receta fija de agent-browser (`tooling/providers/run.go:14-18`), como documenta `SECURITY.md:28-38`. La causa: el cambio `2026-10-15-agent-browser-optional-install` archivó un delta de tres requisitos que nunca se aplicó al spec vigente (el spec no cambió entre el 2026-10-05 y el 2026-10-23, salvo #104), igual que `2026-10-15-pi-subagents-unversioned-install`, que corrigió #108. Después, #104 modificó dos de esos requisitos sobre el texto sin aplicar.

Este cambio fusiona aquel delta con lo que añadieron #104 (CodeGraph, la tarjeta de preguntas de Codex) y #108 (sin pin), en «Optional operation capabilities», «Verified release package installation» y «Read-only diagnostics». El texto de agent-browser es el aprobado el 2026-10-15 (afirmación de esa fuente; no se reverificó contra el código salvo `run.go`).

## Alcance y aceptación

- AC1. «Verified release package installation» nombra las dos descargas opcionales (pi-subagents; agent-browser tras el consentimiento, incluido el binario que baja el postinstall de npm y Chrome for Testing) y su escenario de consentimiento solo restringe los pasos de agent-browser. *Falso en la base cuando* dice «only the Pi host's pi-subagents step may download».
- AC2. «Optional operation capabilities» contiene la puerta por plataforma de agent-browser y sus cuatro escenarios (plataforma soportada, paso fallido, plataforma o requisito ausente, modos seguros), conservando los de CodeGraph. *Falso en la base cuando* no contiene «only on macOS arm64 with Homebrew».
- AC3. «Read-only diagnostics» contiene la detección de agent-browser por `PATH` y archivos y su escenario, conservando el de CodeGraph. *Falso en la base cuando* no contiene «agent-browser present without a browser».

Restricciones: no cambia código ni `content/`; el resto del spec queda igual.
