const KEY = "tinychess:clientId";

export function getOrCreateClientId(): string {
  try {
    const existing = sessionStorage.getItem(KEY);
    if (existing) return existing;
  } catch {
    /* sessionStorage unavailable */
  }
  const id = crypto.randomUUID();
  try {
    sessionStorage.setItem(KEY, id);
  } catch {
    /* ignore */
  }
  return id;
}

export function setClientId(id: string): void {
  try {
    sessionStorage.setItem(KEY, id);
  } catch {
    /* ignore */
  }
}

export function clearClientId(): void {
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    /* ignore */
  }
}
