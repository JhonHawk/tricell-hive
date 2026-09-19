---
order: 50
targets: [claude]
---

## Package Manager
- **pnpm is the default, but detect first:** the lockfile on disk decides; absent one, the `packageManager` field in `package.json`; absent both, pnpm.
- **Never mix package managers.** Use only the detected one for all operations.
- **Never delete or regenerate lock files** unless explicitly requested. Deterministic deny in the `bash-policy` hook — which cannot see your request, so it honors exactly one carve-out it can verify itself: a `git rm` of a per-package lockfile that is redundant with a same-named one at the repo root (workspace consolidation), with no install in the same command. Any other authorized deletion is handed to the user as the exact command, never evaded.
- **Never install pnpm via corepack** (deprecated upstream; Node 25+ no longer bundles it). Bootstrap per pnpm's official paths: standalone script (`curl -fsSL https://get.pnpm.io/install.sh | sh -`) on machines/servers/generic CI, the `pnpm/setup` action on GitHub Actions, `ghcr.io/pnpm/pnpm` image in Docker (file-level patterns: `iac-devops.md`, via `language-rules`). Version pinning lives in `packageManager` — pnpm ≥10 honors it natively.
