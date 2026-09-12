import { existsSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { createHash } from "node:crypto";
const root = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
for (const path of ["web/public/arasan/arasan.js", "web/public/arasan/arasan.wasm", "web/public/arasan/NOTICES.txt", "apps/mobile/assets/engine/arasan.html"]) {
 if (!existsSync(resolve(root, path))) throw new Error(`Missing ${path}. Run pnpm engine:build with Emscripten 4.0.14 available.`);
}
const metadata = JSON.parse(readFileSync(resolve(root, "web/public/arasan/version.json"), "utf8"));
if (metadata.commit !== "361840f407b588252a48b25bb3000c751a470c89") throw new Error("Engine version is not current. Run pnpm engine:build.");
if (!metadata.hashes) throw new Error("Engine checksums are missing. Run pnpm engine:build.");
for (const [path, hash] of Object.entries(metadata.hashes)) {
 if (createHash("sha256").update(readFileSync(resolve(root, path))).digest("hex") !== hash) throw new Error(`Engine file changed: ${path}. Run pnpm engine:build.`);
}
