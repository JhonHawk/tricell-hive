# Modelos de delegación — guía de onboarding

Cómo se reparte el trabajo entre agentes en este sistema de configuración. Existen **tres
figuras de delegación**, no una — y el error más caro en la práctica nace de confundirlas:
asumir que un subagente "ya sabe" el contexto de la conversación.

## Las tres figuras

### 1. Subagente efímero (el default)

Un `Agent` estándar (Explore, un especialista de dominio, review-refuter) **nace sin
memoria alguna de la conversación**. Su única fuente de verdad es el prompt de invocación:
lo que no está escrito ahí, para él no existe. Vive una misión, entrega su informe final y
se disuelve — solo ese informe regresa al hilo principal.

Consecuencias prácticas:
- El prompt de delegación lleva **la intención completa**: el objetivo mayor, quién consume
  el output, rutas de archivos concretas. "Revisa el fix" no significa nada para alguien que
  no vio la conversación.
- Su informe viaja de vuelta como **claims a verificar, no contexto a confiar** — el
  verificador re-ejecuta contra el estado real (diff, corrida de tests), nunca acepta el
  reporte del implementador como evidencia.

### 2. Agente con memoria persistente

Un agente puede declarar `memory: user | project | local` en su frontmatter: retiene
recuerdos propios **entre invocaciones**, aunque cada invocación siga sin ver tu
conversación. Es la figura intermedia: amnésico respecto al hilo, persistente respecto a su
propio historial. Hoy ningún agente de la flota lo usa — existe como opción de diseño
(referencia: `CLAUDE.md > Agent Frontmatter Reference`).

### 3. Agent Teams (el equipo real)

Workers **persistentes y paralelos** con task list compartida y mensajería entre pares
(`SendMessage`). Aquí sí hay coordinación continua: un teammate puede mandarte resultados a
mitad de tu turno, y tú puedes continuarlo con su contexto intacto. Es la única figura donde
"le pregunto al que ya trabajó en esto" tiene sentido.

## Cuándo usar cada una

La política canónica vive en `global/rules/workflow/agent-routing.md > Workflow Tool vs
Subagents vs Agent Teams` — esta guía no la duplica. La forma corta:

- **Default: subagente efímero** — un `Agent` con el especialista correcto cubre el trabajo
  de dominio de una sola pieza.
- **Agent Teams solo cuando** el trabajo necesita workers paralelos persistentes que se
  coordinen entre sí, o el usuario lo pide.
- **Workflow tool solo con opt-in explícito** del usuario (spawns masivos = costo que el
  usuario autoriza, nunca se infiere).

## Ideas erróneas frecuentes

- **"El subagente ya sabe de qué hablamos."** No. Ver figura 1: solo conoce su prompt.
- **"La compactación de contexto borra todo."** No es amnesia total: el hilo continúa con un
  resumen más el contexto reciente sin resumir, y la memoria nativa, el ledger y git
  persisten por sí mismos. Engram es el registro de trabajo buscable entre sesiones, no el
  único sobreviviente.
- **"Explore es barato siempre."** Su valor garantizado es la **higiene de contexto** (el
  ruido de búsqueda no entra al hilo principal); el costo en tokens depende del modelo que
  hereda (`model: inherit` = el de la sesión).
- **"Si el agente reportó verde, está verde."** El informe de un ejecutor son claims; la
  verificación corre en contexto fresco contra el estado real (`agent-routing.md >
  Verification runs in fresh context`).

---

*Origen: research adversarial 2026-07-17 — dos propuestas independientes examinadas por
`review-refuter`; las tres figuras emergieron como corrección a ambas. Artefacto completo:
`_support/archive/docs/2026-07-17-sistema-magia-isekai.html`.*
