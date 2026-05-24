// Date: 2026-05-25
// Author: XinYang Li

type LogLevel = "debug" | "info" | "warn" | "error";

/**
 * Writes a browser-safe structured log for front-end diagnostics.
 * @param level The logical severity level for the event.
 * @param message The human-readable event summary.
 * @param context Optional contextual fields that help correlate UI behavior.
 */
export function uiLog(level: LogLevel, message: string, context?: Record<string, unknown>): void {
  const payload = {
    layer: "frontend",
    level,
    message,
    context,
    at: new Date().toISOString(),
  };

  const writer = level === "error" ? console.error : level === "warn" ? console.warn : console.info;
  writer(payload);
}
