import { execFileSync } from "node:child_process";
import { copyFileSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { createHash } from "node:crypto";

// Generated files only. Source edits belong in the checked-in adapter.
const root = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const source = resolve(process.env.ARASAN_SOURCE || "engine/arasan/vendor");
const commit = "361840f407b588252a48b25bb3000c751a470c89";
const network = "arasanv8-20260906.nnue";
const run = (command, args, options = {}) => execFileSync(command, args, { stdio: "inherit", ...options });
const compiler = process.env.EMXX || "em++";
const compilerVersion = execFileSync(compiler, ["--version"], { encoding: "utf8" });
if (!/\b4\.0\.14\b/.test(compilerVersion)) throw new Error("Use Emscripten 4.0.14 for this engine build.");
if (execFileSync("git", ["-C", source, "rev-parse", "HEAD"], { encoding: "utf8" }).trim() !== commit) {
 throw new Error(`Arasan source must match ${commit}`);
}
if (execFileSync("git", ["-C", source, "status", "--porcelain", "--untracked-files=no"], { encoding: "utf8" }).trim()) {
 throw new Error("Arasan source contains changes. Use the pinned, clean source.");
}
const out = resolve(root, "web/public/arasan");
mkdirSync(out, { recursive: true });
const generated = resolve(root, "engine/arasan/generated");
mkdirSync(generated, { recursive: true });
run("patch", ["-o", resolve(generated, "protocol.cpp"), resolve(source, "src/protocol.cpp"), resolve(root, "engine/arasan/no-tablebases.patch")]);
const names = `arasanx attacks bench bhash bitutil bitprobe board boardio bookread bookwrit calctime chess chessio eco epdrec globals hash learn legal material movearr movegen notation options protocol scoring searchc search see stats tester threadc threadp evaluate`.split(" ");
run(compiler, [
 ...names.map(name => name === "protocol" ? resolve(generated, "protocol.cpp") : resolve(source, `src/${name}.cpp`)), resolve(root, "engine/arasan/input.cpp"),
 `-I${source}/src`, `-I${source}/src/nnue`, "-std=c++17", "-O3", "-DNDEBUG",
 "-DARASAN_VERSION=26.0", `-DNETWORK=${network}`, "-D_64BIT", "-DUSE_INTRINSICS",
 "-msimd128", "-msse2", "-DSIMD", "-DSSE2",
 "-sASYNCIFY", "-sASYNCIFY_STACK_SIZE=131072", "-sSTACK_SIZE=8388608",
 "-sINITIAL_MEMORY=134217728", "-sALLOW_MEMORY_GROWTH", "-sMAXIMUM_MEMORY=268435456",
 "-sENVIRONMENT=web,worker,node", "-sMODULARIZE", "-sEXPORT_NAME=createArasan",
 "-sEXPORTED_RUNTIME_METHODS=['callMain','stringToNewUTF8']", "-sINVOKE_RUN=0", "-sEXIT_RUNTIME=1",
 "--embed-file", `${source}/network/${network}@/${network}`,
 "-o", resolve(out, "arasan.js"),
]);
copyFileSync(resolve(source, "LICENSE"), resolve(out, "LICENSE.arasan"));
const emRoot = resolve(process.env.EMSCRIPTEN_ROOT || dirname(execFileSync("which", [compiler], { encoding: "utf8" }).trim()));
const noticeFiles = ["LICENSE", "system/lib/libc/musl/COPYRIGHT", "system/lib/libcxx/LICENSE.TXT", "system/lib/libcxxabi/LICENSE.TXT", "system/lib/compiler-rt/LICENSE.TXT"];
const notices = "Your Move computer opponent: Arasan at " + commit + "\n\n" + readFileSync(resolve(source, "LICENSE"), "utf8") + "\n\n" + noticeFiles.map(file => `Emscripten 4.0.14: ${file}\n\n${readFileSync(resolve(emRoot, file), "utf8")}`).join("\n\n");
writeFileSync(resolve(out, "NOTICES.txt"), notices);
// One offline asset for the mobile WebView. No remote executable downloads.
const wasm = readFileSync(resolve(out, "arasan.wasm")).toString("base64");
const loader = readFileSync(resolve(out, "arasan.js"), "utf8");
const mobile = resolve(root, "apps/mobile/assets/engine");
mkdirSync(mobile, { recursive: true });
const code = `${loader}\ncreateArasan({wasmBinary: Uint8Array.from(atob('${wasm}'), c => c.charCodeAt(0)), commands: window.pendingCommands ||= [], print: line => window.ReactNativeWebView.postMessage(line), printErr: () => {}, onAbort: () => window.ReactNativeWebView.postMessage('engine-error')}).then(module => { window.arasan = module; module.callMain(['-c', '1', '-H', '16M']); }).catch(() => window.ReactNativeWebView.postMessage('engine-error'));`;
writeFileSync(resolve(mobile, "arasan.html"), `<!doctype html><meta name="viewport" content="width=device-width"><script type="text/plain" id="third-party-notices">${notices.replaceAll("</script", "<\\/script")}</script><script>${code.replaceAll("</script", "<\\/script")}</script>`);
const sha256 = path => createHash("sha256").update(readFileSync(resolve(root, path))).digest("hex");
const files = ["engine/arasan/build.mjs", "engine/arasan/input.cpp", "engine/arasan/no-tablebases.patch", "web/public/arasan/arasan.js", "web/public/arasan/arasan.wasm", "web/public/arasan/NOTICES.txt", "apps/mobile/assets/engine/arasan.html"];
writeFileSync(resolve(out, "version.json"), JSON.stringify({ engine: "Arasan", commit, network, emscripten: "4.0.14", hashes:Object.fromEntries(files.map(path => [path, sha256(path)])) }, null, 2) + "\n");
console.log(`Built ${readdirSync(out).join(", ")} and offline mobile HTML.`);
