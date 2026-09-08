type Storage = {
  getItem(key: string): Promise<string | null>;
  setItem(key: string, value: string): Promise<void>;
};

export function persistentIdentity(storage: Storage, key: string, generate: () => string): () => Promise<string> {
  let pending: Promise<string> | null = null;
  return () => {
    pending ??= (async () => {
      const saved = await storage.getItem(key);
      if (saved) return saved;
      const id = generate();
      await storage.setItem(key, id);
      return id;
    })().catch((error) => { pending = null; throw error; });
    return pending;
  };
}
