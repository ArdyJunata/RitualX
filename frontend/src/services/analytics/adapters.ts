import type { AnalyticsEvent } from "@/constants/analytics";

export type AnalyticsAdapter = (event: AnalyticsEvent, props?: Record<string, unknown>) => void;

export const consoleAdapter: AnalyticsAdapter = (event, props) => {
  if (process.env.NODE_ENV !== "production") {
    console.log(`[Analytics Track] ${event}`, props || {});
  }
};

// Add actual provider adapters (Firebase, PostHog, etc) here later
export const analyticsAdapters: AnalyticsAdapter[] = [
  consoleAdapter,
];
