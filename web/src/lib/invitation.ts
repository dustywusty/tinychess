export function gameIDFromInput(input: string): string {
  const value = input.trim();
  if (/^[a-zA-Z0-9_-]{1,100}$/.test(value)) return value;
  try {
    const url = new URL(value);
    if (!["https:", "http:", "yourmove:"].includes(url.protocol)) return "";
    const path = url.protocol === "yourmove:" ? `/${url.host}${url.pathname}` : url.pathname;
    return path.match(/^\/g\/([a-zA-Z0-9_-]{1,100})\/?$/)?.[1] ?? "";
  } catch { return ""; }
}
