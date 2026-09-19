---
globs:
  - "**/*.{ts,tsx,js,jsx,mts,cts,mjs,cjs,java,kt,kts}"
---

## Patterns & Anti-patterns (JS/TS, Java, Kotlin)

> Scope is JS/TS and JVM. Examples reference Promise/await, NestJS, Spring, Zod, and class-validator. For Python, Go, and other languages, the broader principles still hold (don't swallow errors, validate env at startup, externalize config) — see `quality/development-principles.md` for language-agnostic guidance.

### Error Handling
- **Never swallow errors.** `catch (e) { console.log(e) }` is almost always a hidden bug. Propagate, re-throw typed, or handle with explicit recovery.
- **Never match error strings.** Use custom error classes (`UserNotFoundError extends DomainError`) with codes. In Java: domain exceptions extending `RuntimeException`.
- **Prefer granular try/catch over function-level wrapping.** Catch per operation when error recovery differs. Function-level wrapping is acceptable for controller/handler entry points or when using a global error boundary (NestJS: `ExceptionFilter`, Spring: `@ControllerAdvice`).
- **Always handle async errors.** No `.then()` chains without `.catch()`, no `async` calls without try/catch. No fire-and-forget without explicit justification.

### Mutability & Side Effects
- **Never mutate function arguments.** Return new copies: `return [...arr, x]`. In Java: `List.copyOf()`, unmodifiable collections.
- **Never use `delete obj.key`.** Use destructuring: `const { key, ...rest } = obj;`.
- **Never share mutable state between requests.** Each request gets its own scope. Use proper lifecycle (`Scope.REQUEST` in Nest, `@RequestScope` in Spring).

### Layer Responsibilities
- **Prefer clean layer separation:** thin controllers (validate + delegate), HTTP-unaware services (throw domain errors, not HTTP exceptions), and pure repositories (data access, no business rules). Deviate when the project's established patterns warrant it — but follow existing conventions, don't invent new layering. **Non-negotiable:** services must never throw HTTP exceptions regardless of project patterns.
- **Never return entities as API responses.** Map to DTOs/ViewModels. Never expose internal fields (`_id`, `password`, `createdBy`).

### Async & Concurrency
- **Parallelize independent async calls.** Sequential `await` on independent operations → use `Promise.all`. Use `Promise.allSettled` when partial results are acceptable.
- **No synchronous I/O in request handlers (Node.js).** No `fs.readFileSync` or synchronous heavy parsing in HTTP handlers. Synchronous APIs are acceptable in scripts, CLIs, and startup code.

### Types & Validation
- **Single source of truth for validation.** If a Zod schema exists, `.parse()` it and trust the resulting type. Don't add redundant manual `if (!x)` checks.
- **Never force-cast.** No `as SomeType` in TS without verification. Use type guards, `satisfies`, or `instanceof`. In Java: pattern matching (`if (obj instanceof User u)`).
- **Centralize enums.** `type Status = "active" | "inactive"` scattered across files → single `as const` object or `enum`. In Java/Kotlin: `enum class`.

### Configuration & Environment
- **Validate env vars at startup.** Use Zod (or `@nestjs/config`) to validate and type all env vars. A missing var must crash at boot, not at runtime in production.
- **Externalize values that vary by environment.** URLs, timeouts, feature flags → config. Constants that are truly fixed (math constants, protocol versions, internal defaults) can be hardcoded with a descriptive name.
