import { execFileSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
const root = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const source = resolve(process.env.ARASAN_SOURCE || resolve(root, "engine/arasan/vendor"));
if (!existsSync(source)) {
 execFileSync("git", ["clone", "--no-checkout", "--filter=blob:none", "https://github.com/jdart1/arasan-chess.git", source], { stdio: "inherit" });
 execFileSync("git", ["-C", source, "checkout", "--detach", "361840f407b588252a48b25bb3000c751a470c89"], { stdio: "inherit" });
}
execFileSync(process.execPath, [resolve(root, "engine/arasan/build.mjs")], { stdio: "inherit", env: { ...process.env, ARASAN_SOURCE: source } });
