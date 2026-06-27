import type { AnalyticsEvent } from "@/constants/analytics";
import { analyticsAdapters } from "./adapters";
import { hasConsent } from "./consent";

export function track(event: AnalyticsEvent, props?: Record<string, unknown>): void {
  if (!hasConsent()) return;
  
  for (const adapter of analyticsAdapters) {
    try {
      adapter(event, props);
    } catch {
      // Best-effort: failures in one provider don't affect others or the main app
    }
  }
}
