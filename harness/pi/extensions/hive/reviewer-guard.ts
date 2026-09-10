// Reviewer-only entrypoint. The parent agent selects this path explicitly;
// the hive/ directory intentionally has no index or package manifest.
export { default } from "../../src/reviewer.ts";
