
## Security — code and dependencies

> The situational half of the security rules, delivered on a code, infrastructure or package-manifest write, or a dependency install. The always-on half — the exposure-gated floor, authentication and secrets — is `security-floor.md`, included in the core.

### Package and workspace manifests
- Review changed lifecycle scripts, dependency sources, overrides and build-script permissions before executing an install or script; never broaden install-script permissions to bypass a failed check.
- Apply the install-time checks below when installing dependencies; a manifest-only metadata edit does not itself install a package.

### Input Validation
- Validate all user input at system boundaries with schema-based validation (Zod, class-validator, Pydantic) — never manual string checks, never trusted external data.
- Sanitize HTML output to prevent XSS: framework auto-escaping; DOMPurify for raw HTML.
- Validation complements, never replaces, parameterized queries (injection) and contextual output encoding (XSS) — a secondary control, not the primary defense for either.

### Injection Prevention
- SQL: always parameterized queries or ORM methods. Never concatenate user input into queries.
- Command injection: never pass user input to shell commands. Use safe APIs (`execFile`, not `exec`).
- SSRF: whitelist allowed domains for outbound HTTP requests with user-provided URLs.

### Error Handling
- Error messages must not leak sensitive data (stack traces, DB schemas, internal paths).
- Log security events (failed logins, permission denials, input validation failures) for audit.

### Supply Chain Security

Unconditional (the floor in `security-floor.md`). The check runs before ANY dependency install; severity decides the path — not a default question:

- **Before installing any dependency version, query OSV.dev:** `curl -s -X POST https://api.osv.dev/v1/query -H "Content-Type: application/json" -d '{"version": "<version>", "package": {"name": "<pkg>", "ecosystem": "<ecosystem>"}}'`. Check the exact version, not just the name — safe in 2.1.0 can be compromised in 2.1.1. Empty `vulns` = no *known* advisories, not a guarantee.
- **Ecosystem mapping:** npm/pnpm/yarn → `npm`, pip/uv → `PyPI`, Maven/Gradle → `Maven`, Go modules → `Go`, Cargo → `crates.io`, NuGet → `NuGet`, RubyGems → `RubyGems`, Pub → `Pub`, Swift → `SwiftURL`, Hex → `Hex`.
- **Resolve by severity — the gate applies to installing the VULNERABLE version; choosing a safe one discharges it:**
  - Compatible `fixed` version exists (patch/minor, or a major verified non-breaking) → install THAT version and report the swap. No question — but a fix available only in a new major is a bump decision, never a silent swap (MEDIUM/LOW → proceed and flag the bump; CRITICAL/HIGH → the no-safe-path ask below).
  - MEDIUM/LOW with no fixed version → proceed; report severity and affected range at close.
  - **CRITICAL/HIGH with no safe path** (no fixed version, or the fixed one is incompatible) → **ask — this gate never relaxes.** The install itself is the exposure event; a compromised postinstall has no rollback.
- **Lockfile-only installs** (`npm install`, `uv sync` with no package argument): run the available audit after install (`npm audit`, `pnpm audit`; `./gradlew dependencyCheckAnalyze` if the OWASP plugin is configured) and resolve findings by the same tiers. For Python, use the OSV.dev query above — pip-based audit tools conflict with the pip ban in `global/CLAUDE.md`.
