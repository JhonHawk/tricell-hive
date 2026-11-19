---
name: hive-principle-minimize-reader-load
description: "Hive principle leaf: read when shaping or reviewing code that is hard to trace. Count the layers between a question and its answer and the mutable state a reader must hold; cut both."
disable-model-invocation: true
---

# Minimize reader load

Read when you shape code a person must follow, and when a review asks why a piece is hard to trace. Reader load has two axes: the layers between a question and its answer, and the hidden or changeable state the reader must remember.

- Inline layers that only forward: a wrapper with one caller, an adapter with one implementation, a method that passes its arguments on.
- Shrink the scope of state: return a value instead of mutating, prefer a local to a field, a field to module state, module state to a global. Derive a value instead of keeping a copy in step.
- State an invariant once, at the boundary, not in each consumer.

Test: can a new reader answer "where does X come from?" and "what can change X?" within half a minute? If not, remove a layer or narrow a state.

A layer that hides a real second implementation, a boundary or a test seam the code needs stays.
