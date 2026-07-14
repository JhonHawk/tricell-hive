---
paths:
  - "**/*.module.ts"
  - "**/*.controller.ts"
  - "**/*.service.ts"
  - "**/*.guard.ts"
  - "**/*.interceptor.ts"
  - "**/*.filter.ts"
  - "**/*.pipe.ts"
  - "**/*.decorator.ts"
  - "nest-cli.json"
---

> **Applies when:** `@nestjs/core` is in `package.json` dependencies. Skip if working on Angular or non-NestJS TypeScript.

## NestJS

### Version Detection
- **Check `package.json` → `@nestjs/core` version BEFORE generating code.** v10 and v11 differ materially: v11 defaults to Express v5, whose path matching changes routes/middleware — named wildcards (`{*splat}`) replace bare `*`, and `(.*)` is no longer supported.

### Architecture
- **Request lifecycle order:** middleware → guards → interceptors (pre) → pipes → controller → service → interceptors (post) → exception filters → response. Place logic in the correct layer.
- **Feature modules** encapsulate related controllers, services, and providers. One module per bounded context — never dump everything in `AppModule`.
- **Implement an `ExceptionFilter`** to translate domain errors (`UserNotFoundError`, `InsufficientBalanceError`) to HTTP responses. Services stay HTTP-unaware (see `patterns-antipatterns.md`).
- **Guards** for auth/authorization only. **Interceptors** for cross-cutting (logging, caching, response mapping). **Middleware** for raw request preprocessing (CORS, body parsing). **Pipes** for validation and transformation.

### DTOs & Validation
- **`class-validator`** for incoming HTTP request validation via `ValidationPipe`. Use `zod` for config, env vars, and non-NestJS contexts.
- **`@ApiProperty()`** on every DTO field — `@nestjs/swagger` generates docs from decorators, not inference.
- **Separate Create/Update DTOs.** Use `PartialType()`, `PickType()`, `OmitType()` from `@nestjs/mapped-types` instead of duplicating fields.

### Configuration
- **`@nestjs/config` + `zod`** for env var validation at startup. Never use raw `process.env.X` without validation.
- **Typed `ConfigService`**: use `ConfigService<EnvSchema, true>` generic to get type-safe `.get()` calls.

### Testing
- **`Test.createTestingModule()`** for unit and integration tests. Override specific providers with `.overrideProvider()` — don't mock the entire DI container.
- **Service logic → unit tests on the service.** Cover the business rules at the layer that owns them.
- **Request/response contracts → integration tests through the controller** (using `supertest` or `@nestjs/testing`). Validates pipes, guards, interceptors, and serialization. Don't unit-test pure service logic through the controller layer — slower and harder to debug.
