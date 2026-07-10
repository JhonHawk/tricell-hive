
## Java/Kotlin

### Tooling
- **Format with Spotless** (Gradle/Maven plugin, Java + Kotlin). Kotlin adds **ktlint** for style (standalone or via Spotless) and **detekt** for static analysis. Prefer these over Checkstyle/PMD in new setups.

### Framework Preferences (Spring Boot 3.x)
- DTOs for API responses — Java `record`, Kotlin `data class`, no mutable POJOs; never expose entities.
- **`@HttpExchange`** for declarative REST clients (Spring 6+). Prefer over RestTemplate and `@FeignClient`. For programmatic clients, `RestClient` (Spring 6.1+) is the replacement for `RestTemplate`, which is on a deprecation path (reference docs mark it deprecated as of 7.0; `@Deprecated` annotation in 7.1, removal in 8.0).
- **`@Transactional` lives at the service boundary**, not on repositories. Repositories run inside the transaction the service opens. Avoid `@Transactional` on controllers — they shouldn't own transaction lifetime.
- **`@ConfigurationProperties` over scattered `@Value`** for env-driven config. Bind a typed record once; inject the record, not individual values.
- **Virtual threads (Java 21+)**: Enable via `spring.threads.virtual.enabled=true`. Never size virtual thread pools manually. **Caveat (JDK 21–23 only):** virtual threads block synchronously on JDBC — fine when the DB is the bottleneck, but a `synchronized` block guarding *long-lived, frequent* blocking I/O pins the carrier; swap to `ReentrantLock` in those hot spots only. JEP 491 (JDK 24+) removed this pinning, so on 24+ the workaround is unnecessary — verify the target JDK before applying it.

### Kotlin on JVM
- **Coroutines for I/O-bound concurrency** (Ktor, Spring WebFlux with kotlinx-coroutines). For Spring MVC + virtual threads, plain blocking code is simpler and equivalent in throughput.
- **`data class` for DTOs and value objects.** Never plain classes with manual `equals`/`hashCode`/`toString` unless there's a concrete reason (e.g., inheritance).
- **Sealed classes/interfaces** for closed type hierarchies (state machines, result types). Pair with `when` for exhaustive pattern matching.

### Language-Specific Patterns
- **Never return `null` from public methods.** Use `Optional<T>` in Java, `T?` with safe calls in Kotlin.
- **`Optional` is for return types only.** Never use as method parameter. Consume with `.map()`, `.orElseThrow()`, `.ifPresent()`. Never `.get()` without checking.
- **Wrap checked exceptions at domain boundary.** Don't let `SQLException` propagate to controllers. Catch in infra layer, re-throw as unchecked domain exceptions.
