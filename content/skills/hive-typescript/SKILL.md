---
name: hive-typescript
description: "TypeScript rules for reading or editing .ts and .tsx files: tagged unions, parsed boundaries, no casts, exhaustive switches."
paths: ["**/*.ts", "**/*.tsx"]
---

# TypeScript

Applies when you read or edit `.ts` or `.tsx` files. Start from [type system discipline](../hive-core/principles/type-system-discipline.md) and [boundary discipline](../hive-core/principles/boundary-discipline.md); the repository's own compiler settings, linter and conventions win over anything here.

1. **Tagged unions.** Model variants with one literal tag field (`kind`, `type`) whose value is unique per variant, never a record of optional fields or a boolean plus optionals. End a switch over a union with `const _exhaustive: never = value;` in the default arm, so adding a variant fails the build.
2. **Branded primitives.** Give a primitive with its own meaning a brand, such as `string & { readonly __brand: "UserId" }`, created only through a function that checks the value; follow the brand shape the repository already uses.
3. **Build the legal shape.** Make the invalid value impossible to construct (a non-empty tuple `[T, ...T[]]` instead of an array plus a length check, a start and a duration instead of a start and an end). Keep a plain `T[]` while every operation on it is safe; strengthen it only where the loose type forces a `!`, a cast or a throw for a state that cannot happen.
4. **No `as` casts on unchecked data.** Data from outside is `unknown` until parsed, once at the entry, with the repository's schema library. Narrow in this order: a switch on the tag, `in`, `typeof` or `instanceof`, a type guard that verifies what it claims, and only then a cast; use `satisfies` to check a value without widening its literals.
5. **Honest type guards.** A guard (`isX`, `hasX`) must check the claim: one that returns true without checking is worse than a cast, because its name says it is safe.
6. **Tests and logs.** Write tests as [Tests](../hive-core/build/tests.md) sets; log through the project's structured logger with an identifier that lets you trace the case, leaving no `console.log` in shipped code.

A lint or type-checker suppression (`@ts-ignore`, `@ts-expect-error`, `eslint-disable`) is a defect to resolve, not a tool: read [suppression review](../hive-core/review/suppression-review.md) before adding one.
