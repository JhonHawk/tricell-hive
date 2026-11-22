---
name: hive-typescript
description: "TypeScript rules for reading or editing .ts and .tsx files: tagged unions, parsed boundaries, no casts, exhaustive switches."
paths: ["**/*.ts", "**/*.tsx"]
---

# TypeScript

Applies when you read or edit `.ts` or `.tsx` files. Start from [type system discipline](../hive-core/principles/type-system-discipline.md) and [boundary discipline](../hive-core/principles/boundary-discipline.md); the repository's own compiler settings, linter and conventions win over anything here.

1. **Tagged unions.** Model variants with one literal tag field (`kind`, `type`) whose value is unique per variant. Do not use a record of optional fields or a boolean plus optionals.
2. **Branded primitives.** Give a primitive with its own meaning a brand, such as `string & { readonly __brand: "UserId" }`, and create it only through a function that checks the value. Follow the brand shape the repository already uses.
3. **Build the legal shape.** Make the invalid value impossible to construct: a non-empty tuple `[T, ...T[]]` instead of an array plus a length check, a start and a duration instead of a start and an end. Keep a plain `T[]` while every operation on it is safe; strengthen it only where the loose type forces a `!`, a cast or a throw for a state that cannot happen.
4. **`unknown`, not `any`.** Data from outside is `unknown` until parsed. Use the repository's schema library and infer the type from the schema before you hand-write a property-by-property guard.
5. **No `as` casts on unchecked data.** Narrow in this order of preference: a switch on the tag, the `in` operator, `typeof` or `instanceof`, a type guard that verifies what it claims, and only then a cast. Use `satisfies` to check a value without widening its literal types.
6. **Honest type guards.** A guard (`isX`, `hasX`) must check the claim; a guard that returns true without checking is worse than a cast, because its name says it is safe.
7. **Exhaustive branches.** End a switch over a union with `const _exhaustive: never = value;` in the default arm, so adding a variant fails the build.
8. **Parse once, at the entry.** Follow [boundary discipline](../hive-core/principles/boundary-discipline.md); `Record<string, unknown>` stops where the data enters.
9. **Derive before declaring.** Use `Pick`, `Omit`, `Parameters`, `ReturnType`, `Awaited` or `typeof` before writing a new interface for a shape that already exists.
10. **Object arguments.** Pass an object when a function takes several arguments of similar type, so the call names them. Skip this in hot paths such as per-frame rendering or parsers.
11. **Tests and logs.** Run real code in tests and double only what crosses the system boundary, as [Tests](../hive-core/build/tests.md) sets. Log through the project's structured logger with an identifier that lets you trace the case; leave no `console.log` in shipped code.

A lint or type-checker suppression (`@ts-ignore`, `@ts-expect-error`, `eslint-disable`) is a defect to resolve, not a tool: read [suppression review](../hive-core/review/suppression-review.md) before adding one.
