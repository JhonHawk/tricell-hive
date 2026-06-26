---
alwaysApply: true
---

## Security

> Secure by default. These checks apply to all code, not just security-critical modules.

### Exposure-gated security floor

- **The floor is gated on exposure, not on a project stage.** Auth defaults (route authentication below) and the CVE gates (Supply Chain Security below) apply whenever the system touches real user data, real production systems, or the public network.
- **Injection/SSRF prevention is part of the same floor** — it reapplies the moment the system touches something real, even in throwaway/experimental work on synthetic local data.
- **Secrets hygiene and destructive-op confirmation are absolute** — they never relax, regardless of exposure.

### Input Validation
- Validate all user input at system boundaries. Never trust external data.
- Use schema-based validation (Zod, class-validator, Pydantic) — never manual string checks.
- Sanitize HTML output to prevent XSS. Use framework auto-escaping; add DOMPurify for raw HTML.
- Validation complements, never replaces, parameterized queries (injection) and contextual output encoding (XSS) — it is a secondary control, not the primary defense for either.

### Injection Prevention
- SQL: always use parameterized queries or ORM methods. Never concatenate user input into queries.
- Command injection: never pass user input to shell commands. Use safe APIs (`execFile`, not `exec`).
- SSRF: whitelist allowed domains for outbound HTTP requests with user-provided URLs.

### Authentication & Secrets
- **Never write real secret values into any file** — source code, documentation, markdown, YAML, JSON, .env, scripts, workspace notes, or any other file type. Use env vars or secret managers.
- When documenting infrastructure or deployments:
  - **Templates/configs:** use full placeholders (`<EC2_HOST>`, `$DB_PASSWORD`).
  - **Reference docs** where partial identification helps: semi-obfuscate — show just enough to identify the resource, mask the rest (e.g., `AKIA****TBQX`, `13.222.***.**2`, `sg-0978****f49`, `9302-XXXX-1741`).
  - Always reference where the full values are stored (e.g., "see GitHub Secrets in repo X").
- If the user explicitly requests writing a full, unmasked secret to a file, confirm the risk and suggest semi-obfuscation or placeholders before proceeding.
- For startup validation of secret-bearing env vars, see `patterns-antipatterns.md > Configuration & Environment`.
- Hash passwords with Argon2id (preferred) or bcrypt for legacy systems where Argon2/scrypt are unavailable. Never compare plaintext.
- **Default every route to authenticated.** When creating or reviewing a new endpoint, verify that auth middleware/guard is applied (decorator, middleware, or explicit check). Public endpoints are legitimate but limited to these patterns: landing/marketing pages, signed webhook receivers (signature verification IS the auth), OAuth/SSO callbacks, orchestrator health probes, and public read APIs. Each public endpoint must declare its status in a code comment or route metadata explaining which pattern it fits. "Internal use only" and "it's a health check" are not by themselves justifications — they map (or don't) to one of the patterns above.

### Error Handling
- Error messages must not leak sensitive data (stack traces, DB schemas, internal paths).
- Log security events (failed logins, permission denials, input validation failures) for audit.

### Supply Chain Security
- **Before installing any dependency version, check OSV.dev for known vulnerabilities.** Query via Bash: `curl -s -X POST https://api.osv.dev/v1/query -H "Content-Type: application/json" -d '{"version": "<version>", "package": {"name": "<pkg>", "ecosystem": "<ecosystem>"}}'`. Empty `vulns` means no *known* advisories — not a guarantee of safety.
- **Ecosystem mapping:** npm/pnpm/yarn → `npm`, pip/uv → `PyPI`, Maven/Gradle → `Maven`, Go modules → `Go`, Cargo → `crates.io`, NuGet → `NuGet`, RubyGems → `RubyGems`, Pub → `Pub`, Swift → `SwiftURL`, Hex → `Hex`.
- **Always check the exact version being installed**, not just the package name. A package may be safe in 2.1.0 and compromised in 2.1.1.
- **If vulnerabilities are found:** show severity (CRITICAL/HIGH/MEDIUM/LOW) and affected version range. If a `fixed` version exists in the response, suggest it; otherwise state that no known fix is available. Ask the user whether to install anyway, pin a safe version, or skip.
- **Never silently install a package with known CRITICAL or HIGH CVEs.** Confirmation is mandatory.
- **Lockfile-only installs** (`npm install`, `uv sync` with no package argument): suggest running an audit command after install if available (`npm audit`, `pnpm audit`; `./gradlew dependencyCheckAnalyze` if the OWASP plugin is configured). For Python, prefer the OSV.dev query above — pip-based audit tools conflict with the pip ban in `global/CLAUDE.md`.
