"use client";

import { createContext, useContext, useMemo, useEffect, type ReactNode } from "react";
import { track as trackEvent } from "@/services/analytics";
import type { AnalyticsEvent } from "@/constants/analytics";

interface AnalyticsContextValue {
  track: (event: AnalyticsEvent, props?: Record<string, unknown>) => void;
}
const AnalyticsContext = createContext<AnalyticsContextValue | null>(null);

export function AnalyticsProvider({ children }: { children: ReactNode }) {
  useEffect(() => {
    // Dynamic import to keep web-vitals out of main bundle if possible, or just call directly.
    import("@/services/analytics/web-vitals").then(({ reportWebVitals }) => {
      reportWebVitals();
    });
  }, []);

  const value = useMemo<AnalyticsContextValue>(() => ({ track: trackEvent }), []);
  
  return <AnalyticsContext.Provider value={value}>{children}</AnalyticsContext.Provider>;
}

const NOOP: AnalyticsContextValue = { track: () => undefined };
export function useAnalytics(): AnalyticsContextValue {
  return useContext(AnalyticsContext) ?? NOOP;
}
