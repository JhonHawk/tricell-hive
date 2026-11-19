---
name: hive-principle-type-system-discipline
description: "Hive principle leaf: read when designing a type or a signature, or writing statically typed code. Let the compiler reject impossible states, mixed-up primitives and unhandled variants."
disable-model-invocation: true
---

# Type system discipline

Read when you design a type, shape a function signature or write code in a statically typed language. The goal: a state the code must never hold cannot be written down.

1. When two values share a primitive type but mean different things (two ids, a price and a count), give each its own distinct type, and check the value once, where it is created.
2. Model variants as a tagged union, one tag per valid state, instead of a record of optional fields where contradictory combinations compile.
3. Treat data from outside the process, such as a payload, a file, an environment variable or a stored row, as untyped until a parse step at the boundary turns it into the domain type. Inside, trust that type.
4. When a schema, contract or generated client already owns a shape, derive the type from it; do not hand-write a parallel copy that can drift.
5. When code branches over variants, make the compiler fail when a variant is added, rather than falling into a silent default.

Test: name one invalid state the types still allow. If you can name one cheaply removable, remove it.
