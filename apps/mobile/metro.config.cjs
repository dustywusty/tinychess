const { getDefaultConfig } = require("expo/metro-config");
const http = require("node:http");
const https = require("node:https");

const config = getDefaultConfig(__dirname);
// The optional web preview uses the same API paths as the native app.
config.server.enhanceMiddleware = (middleware) => (req, res, next) => {
  if (!req.url?.startsWith("/api/")) return middleware(req, res, next);
  const target = new URL(req.url, process.env.EXPO_PUBLIC_API_URL || "http://localhost:8080");
  const upstream = (target.protocol === "https:" ? https : http).request(target, {
    method: req.method, headers: { ...req.headers, host: target.host },
  }, (response) => {
    res.writeHead(response.statusCode || 502, response.headers);
    response.pipe(res);
  });
  upstream.on("error", () => { if (!res.headersSent) res.writeHead(502); res.end(); });
  res.on("close", () => upstream.destroy());
  req.pipe(upstream);
};
module.exports = config;
