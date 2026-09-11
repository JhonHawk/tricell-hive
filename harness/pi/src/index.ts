import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { registerGeneralHiveHooks } from "./hooks.ts";

export * from "./child-registry.ts";
export * from "./git-read.ts";
export * from "./hook-runner.ts";
export * from "./hooks.ts";
export * from "./reviewer.ts";
export * from "./types.ts";

export default function hivePiExtension(pi: ExtensionAPI): void {
  registerGeneralHiveHooks(pi);
}
