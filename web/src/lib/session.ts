const KEY = "tinychess:clientId";
let memoryID: string | undefined;

export function getOrCreateClientId(): string {
  try {
    const existing = localStorage.getItem(KEY);
    if (existing) return (memoryID = existing);
  } catch {
    /* Storage can be blocked. Keep a stable identity for this page. */
  }
  if (memoryID) return memoryID;
  let previous: string | null = null;
  try { previous = sessionStorage.getItem(KEY); } catch { /* Legacy storage unavailable. */ }
  const id = previous || crypto.randomUUID();
  setClientId(id);
  return id;
}

export function setClientId(id: string): void {
  memoryID = id;
  try {
    localStorage.setItem(KEY, id);
  } catch {
    /* Keep the in-memory identity when persistent storage is unavailable. */
  }
  try { sessionStorage.removeItem(KEY); } catch { /* Legacy storage unavailable. */ }
}

export function clearClientId(): void {
  memoryID = undefined;
  try { localStorage.removeItem(KEY); } catch { /* Storage unavailable. */ }
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    /* ignore */
  }
}
