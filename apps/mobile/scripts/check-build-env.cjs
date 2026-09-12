// Run on EAS workers before dependency installation, not during local previews.
const value = process.env.EXPO_PUBLIC_API_URL;
let url;
try { url = new URL(value); } catch { /* Report a clear build error below. */ }
if (!url || url.protocol !== "https:" || url.username || url.password || url.search || url.hash || url.pathname !== "/" || ["localhost", "127.0.0.1", "[::1]"].includes(url.hostname)) {
  console.error("Set EXPO_PUBLIC_API_URL to your public HTTPS origin in the EAS build environment. Do not include /api or credentials.");
  process.exit(1);
}
