let granted = true; // Hardcoded true for MVP, would hydrate from consent manager later.

export function setConsent(value: boolean): void {
  granted = value;
}

export function hasConsent(): boolean {
  return granted;
}
