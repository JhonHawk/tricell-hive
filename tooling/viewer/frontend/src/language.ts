/** Resolve only the primary browser language; unsupported input uses English. */
export function resolveLanguage(value: unknown): "en" | "es" {
  if (typeof value !== "string" || !/^es(?:-|$)/i.test(value)) return "en";
  try {
    return Intl.getCanonicalLocales(value).length === 1 ? "es" : "en";
  } catch {
    return "en";
  }
}
export function browserLanguage(): "en" | "es" {
  return resolveLanguage(
    typeof navigator === "undefined" ? undefined : navigator.language,
  );
}
