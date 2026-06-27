import { onLCP, onINP, onCLS, onFCP, onTTFB, type Metric } from "web-vitals";
import { track } from "./track";
import { AnalyticsEvent } from "@/constants/analytics";

export function reportWebVitals(): void {
  const send = (m: Metric) =>
    track("web_vital" as AnalyticsEvent, { // Using a loose type cast here for dynamic metric reporting without expanding main events list
      name: m.name,
      value: Math.round(m.name === "CLS" ? m.value * 1000 : m.value),
      rating: m.rating,
      id: m.id,
    });
    
  onLCP(send);
  onINP(send);
  onCLS(send);
  onFCP(send);
  onTTFB(send);
}
